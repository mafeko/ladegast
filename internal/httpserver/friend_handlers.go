package httpserver

import (
	"net"
	"net/http"
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/store"
)

func (s *Server) registerPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+s.publicPath("/l/{uuid}"), s.handleFriendStatus)
	mux.HandleFunc("POST "+s.publicPath("/l/{uuid}/login"), s.handleFriendLogin)
	mux.HandleFunc("POST "+s.publicPath("/l/{uuid}/logout"), s.requireFriend(s.handleFriendLogout))
	mux.HandleFunc("GET "+s.publicPath("/l/{uuid}/dashboard"), s.requireFriend(s.handleFriendDashboard))
	mux.HandleFunc("GET "+s.publicPath("/l/{uuid}/sessions"), s.requireFriend(s.handleFriendSessions))
	mux.HandleFunc("POST "+s.publicPath("/l/{uuid}/mode"), s.requireFriend(s.handleFriendSetMode))
}

// requireFriend wraps a handler so it only runs when the friend session
// cookie's subject matches the {uuid} in the request path.
func (s *Server) requireFriend(next func(http.ResponseWriter, *http.Request, store.Friend)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("uuid")
		subject := s.sessionSubject(r, friendCookieName)
		if subject == "" || subject != uuid {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		friend, err := s.store.GetFriend(uuid)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		next(w, r, friend)
	}
}

func (s *Server) handleFriendStatus(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if _, err := s.store.GetFriend(uuid); err != nil {
		writeError(w, http.StatusNotFound, "unknown link")
		return
	}
	authenticated := s.sessionSubject(r, friendCookieName) == uuid
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": authenticated})
}

func (s *Server) handleFriendLogin(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var body struct {
		PIN string `json:"pin"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}

	rateLimitKey := clientIP(r) + ":" + uuid
	now := time.Now()
	count, err := s.store.CountRecentFailures(rateLimitKey, now, s.cfg.RateLimitWindow)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if count >= s.cfg.RateLimitMaxAttempts {
		writeError(w, http.StatusTooManyRequests, "too many attempts, try again later")
		return
	}

	friend, err := s.store.GetFriend(uuid)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(friend.PINHash), []byte(body.PIN)) != nil {
		_ = s.store.RecordFailedAttempt(rateLimitKey, now, s.cfg.RateLimitWindow)
		writeError(w, http.StatusUnauthorized, "invalid PIN")
		return
	}

	_ = s.store.ResetAttempts(rateLimitKey)
	if err := s.setSessionCookie(w, friendCookieName, friend.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"name": friend.Name})
}

func (s *Server) handleFriendLogout(w http.ResponseWriter, r *http.Request, _ store.Friend) {
	s.clearSessionCookie(w, friendCookieName)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleFriendDashboard(w http.ResponseWriter, r *http.Request, friend store.Friend) {
	summary, err := s.dashboardSummary(r, friend)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

type dashboard struct {
	Friend    friendInfo   `json:"friend"`
	Today     evcc.Summary `json:"today"`
	ThisWeek  evcc.Summary `json:"thisWeek"`
	ThisMonth evcc.Summary `json:"thisMonth"`
	ThisYear  evcc.Summary `json:"thisYear"`
}

type friendInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *Server) dashboardSummary(r *http.Request, friend store.Friend) (dashboard, error) {
	now := time.Now()
	sessions, err := s.evcc.Sessions(r.Context(), now.Year(), 0)
	if err != nil {
		return dashboard{}, err
	}
	own := evcc.FilterByVehicle(sessions, friend.VehicleName)

	summarize := func(from time.Time) evcc.Summary {
		var in []evcc.Session
		for _, sess := range own {
			if !sess.Created.Before(from) {
				in = append(in, sess)
			}
		}
		return evcc.Summarize(in, friend.RateSolarPerKWh, friend.RateGridPerKWh)
	}

	return dashboard{
		Friend:    friendInfo{ID: friend.ID, Name: friend.Name},
		Today:     summarize(startOfDay(now)),
		ThisWeek:  summarize(startOfWeek(now)),
		ThisMonth: summarize(startOfMonth(now)),
		ThisYear:  summarize(startOfYear(now)),
	}, nil
}

func (s *Server) handleFriendSessions(w http.ResponseWriter, r *http.Request, friend store.Friend) {
	now := time.Now()
	sessions, err := s.evcc.Sessions(r.Context(), now.Year(), 0)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	own := evcc.FilterByVehicle(sessions, friend.VehicleName)
	sort.Slice(own, func(i, j int) bool { return own[i].Created.After(own[j].Created) })

	type sessionView struct {
		Created    time.Time `json:"created"`
		Finished   time.Time `json:"finished"`
		ChargedKWh float64   `json:"chargedKWh"`
		SolarKWh   float64   `json:"solarKWh"`
		GridKWh    float64   `json:"gridKWh"`
		CostEUR    float64   `json:"costEUR"`
	}
	views := make([]sessionView, 0, len(own))
	for _, sess := range own {
		solar, grid := sess.SolarKWh(), sess.GridKWh()
		views = append(views, sessionView{
			Created:    sess.Created,
			Finished:   sess.Finished,
			ChargedKWh: sess.ChargedEnergy,
			SolarKWh:   solar,
			GridKWh:    grid,
			CostEUR:    solar*friend.RateSolarPerKWh + grid*friend.RateGridPerKWh,
		})
	}
	writeJSON(w, http.StatusOK, views)
}

func (s *Server) handleFriendSetMode(w http.ResponseWriter, r *http.Request, friend store.Friend) {
	var body struct {
		Mode string `json:"mode"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	mode := evcc.Mode(body.Mode)
	if mode != evcc.ModeSmart && mode != evcc.ModeNow {
		writeError(w, http.StatusBadRequest, "mode must be smart or now")
		return
	}
	if err := s.evcc.SetVehicleMode(r.Context(), friend.VehicleName, mode); err != nil {
		writeError(w, http.StatusBadGateway, "could not set mode")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mode": string(mode)})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
