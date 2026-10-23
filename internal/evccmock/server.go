package evccmock

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// sessionJSON mirrors the fields of EVCC's GET /api/sessions response that
// matter for Ladegast, plus the ones real EVCC adds around them.
type sessionJSON struct {
	ID              int       `json:"id"`
	Created         time.Time `json:"created"`
	Finished        time.Time `json:"finished"`
	Loadpoint       string    `json:"loadpoint"`
	Vehicle         string    `json:"vehicle"`
	ChargedEnergy   float64   `json:"chargedEnergy"`
	ChargeDuration  int64     `json:"chargeDuration"` // nanoseconds, like EVCC
	SolarPercentage float64   `json:"solarPercentage"`
	Price           float64   `json:"price"`
	PricePerKWh     float64   `json:"pricePerKWh"`
}

// Server implements the mocked EVCC endpoints:
//
//	GET  /api/sessions?year=&month=
//	POST /api/vehicles/{name}/mode/{mode}
//	GET  /api/vehicles/{name}/mode   (mock-only helper to inspect the mode)
type Server struct {
	data *resolved
	mu   sync.Mutex
}

// NewServer builds a server from cfg; now anchors RebaseToNow.
func NewServer(cfg Config, now time.Time) (*Server, error) {
	data, err := cfg.resolve(now)
	if err != nil {
		return nil, err
	}
	return &Server{data: data}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/sessions", s.auth(s.handleSessions))
	mux.HandleFunc("POST /api/vehicles/{name}/mode/{mode}", s.auth(s.handleSetMode))
	mux.HandleFunc("GET /api/vehicles/{name}/mode", s.auth(s.handleGetMode))
	return mux
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.data.apiKey != "" {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") ||
				subtle.ConstantTimeCompare([]byte(token), []byte(s.data.apiKey)) != 1 {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	year, err := intParam(r, "year")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid year"})
		return
	}
	month, err := intParam(r, "month")
	if err != nil || month > 12 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid month"})
		return
	}
	out := make([]sessionJSON, 0, len(s.data.sessions))
	for _, sess := range s.data.sessions {
		created := sess.Created.In(s.data.loc)
		if year != 0 && created.Year() != year {
			continue
		}
		if month != 0 && int(created.Month()) != month {
			continue
		}
		out = append(out, sess)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSetMode(w http.ResponseWriter, r *http.Request) {
	name, mode := r.PathValue("name"), r.PathValue("mode")
	v, ok := s.data.vehicles[name]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "vehicle not found"})
		return
	}
	if !validMode(mode) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid mode"})
		return
	}
	s.mu.Lock()
	v.mode = mode
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"result": mode})
}

func (s *Server) handleGetMode(w http.ResponseWriter, r *http.Request) {
	v, ok := s.data.vehicles[r.PathValue("name")]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "vehicle not found"})
		return
	}
	s.mu.Lock()
	mode := v.mode
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]string{"result": mode})
}

// intParam parses an optional non-negative integer query parameter; 0 means absent.
func intParam(r *http.Request, key string) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, strconv.ErrSyntax
	}
	return n, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
