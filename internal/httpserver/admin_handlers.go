package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/idgen"
	"github.com/mafeko/ladegast/internal/store"
)

const (
	pinLength    = 6 // generated PINs
	minPINLength = 4 // admin-chosen PINs
	maxPINLength = 12
)

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+s.adminPath("/login"), s.handleAdminLogin)
	mux.HandleFunc("POST "+s.adminPath("/logout"), s.requireAdmin(s.handleAdminLogout))
	mux.HandleFunc("GET "+s.adminPath("/guests"), s.requireAdmin(s.handleAdminListGuests))
	mux.HandleFunc("POST "+s.adminPath("/guests"), s.requireAdmin(s.handleAdminCreateGuest))
	mux.HandleFunc("GET "+s.adminPath("/vehicles"), s.requireAdmin(s.handleAdminListVehicles))
	mux.HandleFunc("PUT "+s.adminPath("/guests/{id}"), s.requireAdmin(s.handleAdminUpdateGuest))
	mux.HandleFunc("DELETE "+s.adminPath("/guests/{id}"), s.requireAdmin(s.handleAdminDeleteGuest))
	mux.HandleFunc("POST "+s.adminPath("/guests/{id}/regenerate-link"), s.requireAdmin(s.handleAdminRegenerateLink))
	mux.HandleFunc("POST "+s.adminPath("/guests/{id}/reset-pin"), s.requireAdmin(s.handleAdminResetPIN))
	mux.HandleFunc("GET "+s.adminPath("/guests/{id}/view-as/dashboard"), s.requireAdmin(s.handleAdminViewAsDashboard))
	mux.HandleFunc("GET "+s.adminPath("/guests/{id}/view-as/sessions"), s.requireAdmin(s.handleAdminViewAsSessions))
	mux.HandleFunc("GET "+s.adminPath("/guests/{id}/view-as/settlements"), s.requireAdmin(s.handleAdminViewAsSettlements))
	mux.HandleFunc("POST "+s.adminPath("/guests/{id}/settlements"), s.requireAdmin(s.handleAdminCreateSettlement))
	mux.HandleFunc("DELETE "+s.adminPath("/guests/{id}/settlements/{sid}"), s.requireAdmin(s.handleAdminDeleteSettlement))
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := s.sessionSubject(r, adminCookieName)
		if username == "" {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		if _, err := s.store.GetAdmin(username); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		next(w, r)
	}
}

func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	admin, err := s.store.GetAdmin(body.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(body.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := s.setSessionCookie(w, adminCookieName, admin.Username); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"username": admin.Username})
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	s.clearSessionCookie(w, adminCookieName)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAdminListGuests(w http.ResponseWriter, r *http.Request) {
	guests, err := s.store.ListGuests()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	views := make([]adminGuest, 0, len(guests))
	for _, f := range guests {
		views = append(views, s.adminGuestView(f))
	}
	writeJSON(w, http.StatusOK, views)
}

// handleAdminListVehicles returns the EVCC vehicle names seen in charging
// sessions that no guest is assigned to yet. EVCC has no cheap way to list
// configured vehicles, so a vehicle only shows up after its first session.
func (s *Server) handleAdminListVehicles(w http.ResponseWriter, r *http.Request) {
	guests, err := s.store.ListGuests()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	sessions, err := s.evcc.Sessions(r.Context(), 0, 0)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	assigned := make(map[string]bool, len(guests))
	for _, f := range guests {
		assigned[f.VehicleName] = true
	}
	seen := map[string]bool{}
	names := []string{}
	for _, sess := range sessions {
		if sess.Vehicle == "" || assigned[sess.Vehicle] || seen[sess.Vehicle] {
			continue
		}
		seen[sess.Vehicle] = true
		names = append(names, sess.Vehicle)
	}
	sort.Strings(names)
	writeJSON(w, http.StatusOK, names)
}

// adminGuest is the admin API's view of a guest: everything the admin UI
// needs, including the shareable link, but never the PIN hash.
type adminGuest struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	VehicleName       string    `json:"vehicleName"`
	RateSolarPerKWh   float64   `json:"rateSolarPerKWh"`
	RateGridPerKWh    float64   `json:"rateGridPerKWh"`
	CreatedAt         time.Time `json:"createdAt"`
	LinkRegeneratedAt time.Time `json:"linkRegeneratedAt"`
	Link              string    `json:"link"`
}

func (s *Server) adminGuestView(f store.Guest) adminGuest {
	return adminGuest{
		ID:                f.ID,
		Name:              f.Name,
		VehicleName:       f.VehicleName,
		RateSolarPerKWh:   f.RateSolarPerKWh,
		RateGridPerKWh:    f.RateGridPerKWh,
		CreatedAt:         f.CreatedAt,
		LinkRegeneratedAt: f.LinkRegeneratedAt,
		Link:              s.guestLink(f.ID),
	}
}

func (s *Server) guestLink(id string) string {
	return s.cfg.PublicBaseURL + s.publicPath("/l/"+id)
}

// validPIN reports whether pin is an acceptable admin-chosen PIN: only
// digits, between minPINLength and maxPINLength long.
func validPIN(pin string) bool {
	if len(pin) < minPINLength || len(pin) > maxPINLength {
		return false
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// pinOrGenerate returns the admin-chosen PIN, or a random one if none was
// given. ok is false (and the response already written) for an invalid PIN.
func pinOrGenerate(w http.ResponseWriter, chosen string) (pin string, ok bool) {
	if chosen == "" {
		return idgen.PIN(pinLength), true
	}
	if !validPIN(chosen) {
		writeError(w, http.StatusBadRequest, "pin must be 4 to 12 digits")
		return "", false
	}
	return chosen, true
}

type createGuestRequest struct {
	Name            string  `json:"name"`
	VehicleName     string  `json:"vehicleName"`
	RateSolarPerKWh float64 `json:"rateSolarPerKWh"`
	RateGridPerKWh  float64 `json:"rateGridPerKWh"`
	// PIN is optional: on create an empty value means "generate one".
	PIN string `json:"pin"`
}

func validRates(b createGuestRequest) bool {
	return b.RateSolarPerKWh >= 0 && b.RateGridPerKWh >= 0
}

func (s *Server) handleAdminCreateGuest(w http.ResponseWriter, r *http.Request) {
	var body createGuestRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Name == "" || body.VehicleName == "" {
		writeError(w, http.StatusBadRequest, "name and vehicleName are required")
		return
	}

	if !validRates(body) {
		writeError(w, http.StatusBadRequest, "rates must not be negative")
		return
	}
	pin, ok := pinOrGenerate(w, body.PIN)
	if !ok {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	guest := store.Guest{
		ID:                idgen.UUID(),
		Name:              body.Name,
		VehicleName:       body.VehicleName,
		PINHash:           string(hash),
		RateSolarPerKWh:   body.RateSolarPerKWh,
		RateGridPerKWh:    body.RateGridPerKWh,
		CreatedAt:         time.Now(),
		LinkRegeneratedAt: time.Now(),
	}
	if err := s.store.CreateGuest(guest); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":   guest.ID,
		"pin":  pin,
		"link": s.guestLink(guest.ID),
	})
}

func (s *Server) handleAdminUpdateGuest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	guest, err := s.store.GetGuest(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "guest not found")
		return
	}

	var body createGuestRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if !validRates(body) {
		writeError(w, http.StatusBadRequest, "rates must not be negative")
		return
	}
	if body.Name != "" {
		guest.Name = body.Name
	}
	if body.VehicleName != "" {
		guest.VehicleName = body.VehicleName
	}
	guest.RateSolarPerKWh = body.RateSolarPerKWh
	guest.RateGridPerKWh = body.RateGridPerKWh

	if err := s.store.UpdateGuest(guest); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, s.adminGuestView(guest))
}

// handleAdminDeleteGuest removes a guest; their link and any session
// cookie stop working immediately because requireGuest looks the guest up.
func (s *Server) handleAdminDeleteGuest(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteGuest(r.PathValue("id")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "guest not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAdminRegenerateLink invalidates a guest's current link by moving
// their record under a freshly generated UUID; the old link stops working
// immediately because its ID no longer exists in the store.
func (s *Server) handleAdminRegenerateLink(w http.ResponseWriter, r *http.Request) {
	oldID := r.PathValue("id")
	guest, err := s.store.GetGuest(oldID)
	if err != nil {
		writeError(w, http.StatusNotFound, "guest not found")
		return
	}

	guest.ID = idgen.UUID()
	guest.LinkRegeneratedAt = time.Now()
	if err := s.store.CreateGuest(guest); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.store.DeleteGuest(oldID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":   guest.ID,
		"link": s.guestLink(guest.ID),
	})
}

func (s *Server) handleAdminResetPIN(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	guest, err := s.store.GetGuest(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "guest not found")
		return
	}

	// The body is optional: {"pin": "..."} sets a chosen PIN, anything else
	// (no body, null, {}) generates a random one.
	var body struct {
		PIN string `json:"pin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	pin, ok := pinOrGenerate(w, body.PIN)
	if !ok {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	guest.PINHash = string(hash)
	if err := s.store.UpdateGuest(guest); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pin": pin})
}

func (s *Server) handleAdminViewAsDashboard(w http.ResponseWriter, r *http.Request) {
	guest, err := s.store.GetGuest(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "guest not found")
		return
	}
	summary, err := s.dashboardSummary(r, guest)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleAdminViewAsSessions(w http.ResponseWriter, r *http.Request) {
	guest, err := s.store.GetGuest(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "guest not found")
		return
	}
	s.handleGuestSessions(w, r, guest)
}
