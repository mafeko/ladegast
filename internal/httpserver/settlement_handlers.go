package httpserver

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/idgen"
	"github.com/mafeko/ladegast/internal/store"
)

type settlementView struct {
	ID         string    `json:"id"`
	At         time.Time `json:"at"`
	ChargedKWh float64   `json:"chargedKWh"`
	SolarKWh   float64   `json:"solarKWh"`
	GridKWh    float64   `json:"gridKWh"`
	AmountEUR  float64   `json:"amountEUR"`
}

// openAmount is what has been charged since the last settlement (or ever,
// if there is none) and is therefore still unpaid.
type openAmount struct {
	Since      *time.Time `json:"since"`
	ChargedKWh float64    `json:"chargedKWh"`
	SolarKWh   float64    `json:"solarKWh"`
	GridKWh    float64    `json:"gridKWh"`
	AmountEUR  float64    `json:"amountEUR"`
}

type settlementsResponse struct {
	Settlements []settlementView `json:"settlements"`
	Open        openAmount       `json:"open"`
}

// settlementsFor derives each settlement's amount from the friend's sessions
// between the previous settlement (inclusive) and this one (exclusive), priced
// with the friend's current rates. Settlements are returned oldest first.
func settlementsFor(friend store.Friend, sessions []evcc.Session) settlementsResponse {
	sts := append([]store.Settlement(nil), friend.Settlements...)
	sort.Slice(sts, func(i, j int) bool { return sts[i].At.Before(sts[j].At) })

	own := evcc.FilterByVehicle(sessions, friend.VehicleName)
	between := func(from, to time.Time) evcc.Summary {
		var in []evcc.Session
		for _, sess := range own {
			if !sess.Created.Before(from) && (to.IsZero() || sess.Created.Before(to)) {
				in = append(in, sess)
			}
		}
		return evcc.Summarize(in, friend.RateSolarPerKWh, friend.RateGridPerKWh)
	}

	resp := settlementsResponse{Settlements: make([]settlementView, 0, len(sts))}
	var prev time.Time
	for _, st := range sts {
		sum := between(prev, st.At)
		resp.Settlements = append(resp.Settlements, settlementView{
			ID: st.ID, At: st.At, ChargedKWh: sum.ChargedKWh,
			SolarKWh: sum.SolarKWh, GridKWh: sum.GridKWh, AmountEUR: sum.CostEUR,
		})
		prev = st.At
	}
	open := between(prev, time.Time{})
	resp.Open = openAmount{ChargedKWh: open.ChargedKWh, SolarKWh: open.SolarKWh, GridKWh: open.GridKWh, AmountEUR: open.CostEUR}
	if !prev.IsZero() {
		resp.Open.Since = &prev
	}
	return resp
}

func (s *Server) writeSettlements(w http.ResponseWriter, r *http.Request, friend store.Friend) {
	sessions, err := s.evcc.Sessions(r.Context(), 0, 0)
	if err != nil {
		writeError(w, http.StatusBadGateway, "could not reach EVCC")
		return
	}
	writeJSON(w, http.StatusOK, settlementsFor(friend, sessions))
}

func (s *Server) handleFriendSettlements(w http.ResponseWriter, r *http.Request, friend store.Friend) {
	s.writeSettlements(w, r, friend)
}

func (s *Server) handleAdminViewAsSettlements(w http.ResponseWriter, r *http.Request) {
	friend, err := s.store.GetFriend(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "friend not found")
		return
	}
	s.writeSettlements(w, r, friend)
}

// handleAdminCreateSettlement records that the friend has paid up to now, or
// up to the optional {"at": RFC3339} time in the body.
func (s *Server) handleAdminCreateSettlement(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		At time.Time `json:"at"`
	}
	if err := decodeOptionalJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	if body.At.IsZero() {
		body.At = now
	}
	if body.At.After(now.Add(time.Minute)) {
		writeError(w, http.StatusBadRequest, "settlement time must not be in the future")
		return
	}
	st := store.Settlement{ID: idgen.UUID(), At: body.At}
	if err := s.store.AddSettlement(id, st); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "friend not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (s *Server) handleAdminDeleteSettlement(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteSettlement(r.PathValue("id"), r.PathValue("sid")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "settlement not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
