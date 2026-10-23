package evcc

import "time"

// Session mirrors a single charging session as returned by EVCC's
// GET /api/sessions. Fields not needed by Ladegast are omitted even
// though EVCC's response contains more.
type Session struct {
	ID              int       `json:"id"`
	Created         time.Time `json:"created"`
	Finished        time.Time `json:"finished"`
	Loadpoint       string    `json:"loadpoint"`
	Vehicle         string    `json:"vehicle"`
	ChargedEnergy   float64   `json:"chargedEnergy"`
	SolarPercentage *float64  `json:"solarPercentage"`
}

// SolarKWh returns the portion of ChargedEnergy attributed to solar
// production. EVCC reports no separate grid field, so the remainder is the
// grid share (see GridKWh).
func (s Session) SolarKWh() float64 {
	if s.SolarPercentage == nil {
		return 0
	}
	return s.ChargedEnergy * (*s.SolarPercentage) / 100
}

// GridKWh returns the portion of ChargedEnergy drawn from the grid.
func (s Session) GridKWh() float64 {
	return s.ChargedEnergy - s.SolarKWh()
}
