package httpserver

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/idgen"
	"github.com/mafeko/ladegast/internal/store"
)

const pinLength = 6

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST "+s.adminPath("/login"), s.handleAdminLogin)
	mux.HandleFunc("POST "+s.adminPath("/logout"), s.requireAdmin(s.handleAdminLogout))
	mux.HandleFunc("GET "+s.adminPath("/friends"), s.requireAdmin(s.handleAdminListFriends))
	mux.HandleFunc("POST "+s.adminPath("/friends"), s.requireAdmin(s.handleAdminCreateFriend))
	mux.HandleFunc("PUT "+s.adminPath("/friends/{id}"), s.requireAdmin(s.handleAdminUpdateFriend))
	mux.HandleFunc("POST "+s.adminPath("/friends/{id}/regenerate-link"), s.requireAdmin(s.handleAdminRegenerateLink))
	mux.HandleFunc("POST "+s.adminPath("/friends/{id}/reset-pin"), s.requireAdmin(s.handleAdminResetPIN))
	mux.HandleFunc("GET "+s.adminPath("/friends/{id}/view-as/dashboard"), s.requireAdmin(s.handleAdminViewAsDashboard))
	mux.HandleFunc("GET "+s.adminPath("/friends/{id}/view-as/sessions"), s.requireAdmin(s.handleAdminViewAsSessions))
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

func (s *Server) handleAdminListFriends(w http.ResponseWriter, r *http.Request) {
	friends, err := s.store.ListFriends()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, friends)
}

type createFriendRequest struct {
	Name            string  `json:"name"`
	VehicleName     string  `json:"vehicleName"`
	RateSolarPerKWh float64 `json:"rateSolarPerKWh"`
	RateGridPerKWh  float64 `json:"rateGridPerKWh"`
}

func (s *Server) handleAdminCreateFriend(w http.ResponseWriter, r *http.Request) {
	var body createFriendRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Name == "" || body.VehicleName == "" {
		writeError(w, http.StatusBadRequest, "name and vehicleName are required")
		return
	}

	pin := idgen.PIN(pinLength)
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	friend := store.Friend{
		ID:                idgen.UUID(),
		Name:              body.Name,
		VehicleName:       body.VehicleName,
		PINHash:           string(hash),
		RateSolarPerKWh:   body.RateSolarPerKWh,
		RateGridPerKWh:    body.RateGridPerKWh,
		CreatedAt:         time.Now(),
		LinkRegeneratedAt: time.Now(),
	}
	if err := s.store.CreateFriend(friend); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":   friend.ID,
		"pin":  pin,
		"link": s.cfg.PublicBaseURL + s.publicPath("/l/"+friend.ID),
	})
}

func (s *Server) handleAdminUpdateFriend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	friend, err := s.store.GetFriend(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}

	var body createFriendRequest
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.Name != "" {
		friend.Name = body.Name
	}
	if body.VehicleName != "" {
		friend.VehicleName = body.VehicleName
	}
	friend.RateSolarPerKWh = body.RateSolarPerKWh
	friend.RateGridPerKWh = body.RateGridPerKWh

	if err := s.store.UpdateFriend(friend); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, friend)
}

// handleAdminRegenerateLink invalidates a friend's current link by moving
// their record under a freshly generated UUID; the old link stops working
// immediately because its ID no longer exists in the store.
func (s *Server) handleAdminRegenerateLink(w http.ResponseWriter, r *http.Request) {
	oldID := r.PathValue("id")
	friend, err := s.store.GetFriend(oldID)
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}

	friend.ID = idgen.UUID()
	friend.LinkRegeneratedAt = time.Now()
	if err := s.store.CreateFriend(friend); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := s.store.DeleteFriend(oldID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"id":   friend.ID,
		"link": s.cfg.PublicBaseURL + s.publicPath("/l/"+friend.ID),
	})
}

func (s *Server) handleAdminResetPIN(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	friend, err := s.store.GetFriend(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}

	pin := idgen.PIN(pinLength)
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	friend.PINHash = string(hash)
	if err := s.store.UpdateFriend(friend); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"pin": pin})
}

func (s *Server) handleAdminViewAsDashboard(w http.ResponseWriter, r *http.Request) {
	friend, err := s.store.GetFriend(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}
	summary, err := s.dashboardSummary(r, friend)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleAdminViewAsSessions(w http.ResponseWriter, r *http.Request) {
	friend, err := s.store.GetFriend(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}
	s.handleFriendSessions(w, r, friend)
}
