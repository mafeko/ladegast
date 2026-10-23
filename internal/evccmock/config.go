// Package evccmock is a stand-in for the subset of EVCC's REST API that
// Ladegast uses (see internal/evcc), serving charging sessions from a YAML
// file. It exists for local development and end-to-end demos - never for
// production. DESIGN.md section 3 documents the real endpoints.
package evccmock

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the content of the mock's YAML file.
type Config struct {
	// Listen is the bind address; the -listen flag overrides it.
	Listen string `yaml:"listen"`
	// APIKey is the expected Bearer token. Empty disables authentication.
	APIKey string `yaml:"apiKey"`
	// Timezone (IANA name) used for the year/month filter of /api/sessions.
	Timezone string `yaml:"timezone"`
	// ReferenceTime is the moment the data was generated for ("now" at
	// generation time).
	ReferenceTime time.Time `yaml:"referenceTime"`
	// RebaseToNow shifts all timestamps forward by whole weeks so the data
	// ends just before the actual start time of the server. Whole weeks keep
	// weekday patterns intact.
	RebaseToNow bool   `yaml:"rebaseToNow"`
	Users       []User `yaml:"users"`
}

// User is a person owning exactly one vehicle.
type User struct {
	Name    string  `yaml:"name"`
	Vehicle Vehicle `yaml:"vehicle"`
}

// Vehicle is an EVCC vehicle with its recorded charging sessions.
type Vehicle struct {
	// Name is the EVCC config name (what Ladegast stores as vehicleName and
	// what the mode endpoint addresses).
	Name        string  `yaml:"name"`
	Title       string  `yaml:"title"`
	CapacityKWh float64 `yaml:"capacityKWh"`
	Loadpoint   string  `yaml:"loadpoint"`
	// Mode is the initial charge mode; defaults to "pv".
	Mode     string         `yaml:"mode"`
	Sessions []SessionEntry `yaml:"sessions"`
}

// SessionEntry is one recorded charging session.
type SessionEntry struct {
	Created         time.Time `yaml:"created"`
	Finished        time.Time `yaml:"finished"`
	ChargedEnergy   float64   `yaml:"chargedEnergy"` // kWh
	SolarPercentage float64   `yaml:"solarPercentage"`
	// Price and PricePerKWh are EVCC's own tariff values; Ladegast ignores
	// them but real EVCC delivers them, so the mock does too.
	Price       float64 `yaml:"price"`
	PricePerKWh float64 `yaml:"pricePerKWh"`
}

// LoadConfig reads and validates a YAML file. Unknown keys are an error so
// typos do not silently drop data.
func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return ParseConfig(raw)
}

func ParseConfig(raw []byte) (Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Listen == "" {
		cfg.Listen = ":7070"
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "Europe/Berlin"
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("timezone: %w", err)
	}
	if c.RebaseToNow && c.ReferenceTime.IsZero() {
		return errors.New("rebaseToNow requires referenceTime")
	}
	if len(c.Users) == 0 {
		return errors.New("no users configured")
	}
	seen := map[string]bool{}
	for _, u := range c.Users {
		v := u.Vehicle
		if v.Name == "" {
			return fmt.Errorf("user %q: vehicle.name is required", u.Name)
		}
		if seen[v.Name] {
			return fmt.Errorf("duplicate vehicle name %q", v.Name)
		}
		seen[v.Name] = true
		if v.Mode != "" && !validMode(v.Mode) {
			return fmt.Errorf("vehicle %q: invalid mode %q", v.Name, v.Mode)
		}
		for i, s := range v.Sessions {
			switch {
			case s.Created.IsZero() || s.Finished.IsZero():
				return fmt.Errorf("vehicle %q session %d: created and finished are required", v.Name, i)
			case s.Finished.Before(s.Created):
				return fmt.Errorf("vehicle %q session %d: finished before created", v.Name, i)
			case s.ChargedEnergy < 0:
				return fmt.Errorf("vehicle %q session %d: negative chargedEnergy", v.Name, i)
			case s.SolarPercentage < 0 || s.SolarPercentage > 100:
				return fmt.Errorf("vehicle %q session %d: solarPercentage out of range", v.Name, i)
			}
		}
	}
	return nil
}

// resolved is the in-memory form the server works with: rebased, with
// globally unique session ids assigned in chronological order.
type resolved struct {
	loc      *time.Location
	apiKey   string
	vehicles map[string]*vehicleState
	sessions []sessionJSON
}

type vehicleState struct {
	vehicle Vehicle
	mode    string
}

func (c Config) resolve(now time.Time) (*resolved, error) {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, err
	}
	shiftDays := 0
	if c.RebaseToNow {
		shiftDays = int(now.Sub(c.ReferenceTime)/(7*24*time.Hour)) * 7
	}
	shift := func(t time.Time) time.Time { return t.In(loc).AddDate(0, 0, shiftDays) }

	r := &resolved{loc: loc, apiKey: c.APIKey, vehicles: map[string]*vehicleState{}}
	for _, u := range c.Users {
		v := u.Vehicle
		mode := v.Mode
		if mode == "" {
			mode = "pv"
		}
		r.vehicles[v.Name] = &vehicleState{vehicle: v, mode: mode}
		for _, s := range v.Sessions {
			created, finished := shift(s.Created), shift(s.Finished)
			r.sessions = append(r.sessions, sessionJSON{
				Created:         created,
				Finished:        finished,
				Loadpoint:       v.Loadpoint,
				Vehicle:         v.Name,
				ChargedEnergy:   s.ChargedEnergy,
				ChargeDuration:  int64(finished.Sub(created)),
				SolarPercentage: s.SolarPercentage,
				Price:           s.Price,
				PricePerKWh:     s.PricePerKWh,
			})
		}
	}
	sort.SliceStable(r.sessions, func(i, j int) bool { return r.sessions[i].Created.Before(r.sessions[j].Created) })
	for i := range r.sessions {
		r.sessions[i].ID = i + 1
	}
	return r, nil
}

func validMode(m string) bool {
	switch m {
	case "off", "now", "minpv", "pv", "smart":
		return true
	}
	return false
}
