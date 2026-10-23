package evccmock

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func generated(t *testing.T) Config {
	t.Helper()
	now := time.Date(2026, 10, 23, 20, 0, 0, 0, time.UTC)
	var buf bytes.Buffer
	if err := WriteYAML(&buf, Generate(now, 1)); err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfig(buf.Bytes())
	if err != nil {
		t.Fatalf("generated YAML does not parse: %v", err)
	}
	return cfg
}

func TestGeneratedDataShape(t *testing.T) {
	cfg := generated(t)
	if len(cfg.Users) != 3 {
		t.Fatalf("users = %d, want 3", len(cfg.Users))
	}
	for _, u := range cfg.Users {
		v := u.Vehicle
		if len(v.Sessions) < 30 {
			t.Errorf("%s: only %d sessions", u.Name, len(v.Sessions))
		}
		var kwh, solarKWh float64
		longestGap := time.Duration(0)
		first, last := v.Sessions[0].Created, v.Sessions[len(v.Sessions)-1].Created
		for i, s := range v.Sessions {
			kwh += s.ChargedEnergy
			solarKWh += s.ChargedEnergy * s.SolarPercentage / 100
			if i > 0 {
				if g := s.Created.Sub(v.Sessions[i-1].Created); g > longestGap {
					longestGap = g
				}
			}
		}
		if share := solarKWh / kwh; share < 0.6 {
			t.Errorf("%s: solar share %.2f, want mostly solar (>= 0.6)", u.Name, share)
		}
		if longestGap < 7*24*time.Hour {
			t.Errorf("%s: longest pause %v, want at least a week without charging", u.Name, longestGap)
		}
		if span := last.Sub(first); span < 330*24*time.Hour {
			t.Errorf("%s: data spans only %v", u.Name, span)
		}
	}
}

func TestRebaseToNowShiftsByWholeWeeks(t *testing.T) {
	cfg := generated(t)
	later := cfg.ReferenceTime.Add(30 * 24 * time.Hour)
	r, err := cfg.resolve(later)
	if err != nil {
		t.Fatal(err)
	}
	newest := r.sessions[len(r.sessions)-1].Created
	if newest.After(later) || later.Sub(newest) > 14*24*time.Hour {
		t.Errorf("newest session %v not just before %v", newest, later)
	}
	orig := cfg.Users[0].Vehicle.Sessions[0].Created
	var got time.Time
	for _, s := range r.sessions {
		if s.Vehicle == cfg.Users[0].Vehicle.Name {
			got = s.Created
			break
		}
	}
	if got.Weekday() != orig.Weekday() {
		t.Errorf("weekday changed: %v -> %v", orig.Weekday(), got.Weekday())
	}
	if got.Hour() != orig.In(got.Location()).Hour() {
		t.Errorf("local hour changed across rebase")
	}
}

func newTestServer(t *testing.T, apiKey string) http.Handler {
	t.Helper()
	cfg, err := ParseConfig([]byte(`
apiKey: ` + apiKey + `
timezone: Europe/Berlin
users:
  - name: A
    vehicle:
      name: vehicle_1
      loadpoint: Garage
      sessions:
        - {created: 2026-01-10T10:00:00+01:00, finished: 2026-01-10T12:00:00+01:00, chargedEnergy: 10, solarPercentage: 50, price: 2, pricePerKWh: 0.2}
        - {created: 2026-02-10T10:00:00+01:00, finished: 2026-02-10T11:00:00+01:00, chargedEnergy: 5, solarPercentage: 80, price: 1, pricePerKWh: 0.2}
        - {created: 2025-02-10T10:00:00+01:00, finished: 2025-02-10T11:00:00+01:00, chargedEnergy: 7, solarPercentage: 80, price: 1, pricePerKWh: 0.2}
`))
	if err != nil {
		t.Fatal(err)
	}
	srv, err := NewServer(cfg, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return srv.Handler()
}

func do(h http.Handler, method, url, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSessionsFilterAndShape(t *testing.T) {
	h := newTestServer(t, `""`)
	cases := map[string]int{
		"/api/sessions":                   3,
		"/api/sessions?year=2026":         2,
		"/api/sessions?year=2026&month=2": 1,
		"/api/sessions?year=2024":         0,
	}
	for url, want := range cases {
		rec := do(h, "GET", url, "")
		var got []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("%s: %v", url, err)
		}
		if len(got) != want {
			t.Errorf("%s: %d sessions, want %d", url, len(got), want)
		}
	}
	var got []map[string]any
	_ = json.Unmarshal(do(h, "GET", "/api/sessions?year=2025", "").Body.Bytes(), &got)
	for _, k := range []string{"id", "created", "finished", "loadpoint", "vehicle", "chargedEnergy", "solarPercentage", "price", "pricePerKWh"} {
		if _, ok := got[0][k]; !ok {
			t.Errorf("missing field %q", k)
		}
	}
	if rec := do(h, "GET", "/api/sessions?year=abc", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("invalid year: status %d", rec.Code)
	}
}

func TestAuth(t *testing.T) {
	h := newTestServer(t, "secret")
	if rec := do(h, "GET", "/api/sessions", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("no token: %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/sessions", "wrong"); rec.Code != http.StatusUnauthorized {
		t.Errorf("wrong token: %d", rec.Code)
	}
	if rec := do(h, "GET", "/api/sessions", "secret"); rec.Code != http.StatusOK {
		t.Errorf("right token: %d", rec.Code)
	}
}

func TestSetMode(t *testing.T) {
	h := newTestServer(t, `""`)
	if rec := do(h, "POST", "/api/vehicles/vehicle_1/mode/now", ""); rec.Code != http.StatusOK {
		t.Fatalf("set mode: %d", rec.Code)
	}
	if body := do(h, "GET", "/api/vehicles/vehicle_1/mode", "").Body.String(); !strings.Contains(body, `"now"`) {
		t.Errorf("mode not stored: %s", body)
	}
	if rec := do(h, "POST", "/api/vehicles/nope/mode/now", ""); rec.Code != http.StatusNotFound {
		t.Errorf("unknown vehicle: %d", rec.Code)
	}
	if rec := do(h, "POST", "/api/vehicles/vehicle_1/mode/turbo", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("invalid mode: %d", rec.Code)
	}
}

func TestParseConfigRejectsBadInput(t *testing.T) {
	bad := map[string]string{
		"unknown key":    "users: []\nbogus: 1\n",
		"no users":       "timezone: Europe/Berlin\n",
		"bad timezone":   "timezone: Mars/Base\nusers:\n  - {name: A, vehicle: {name: v}}\n",
		"rebase w/o ref": "rebaseToNow: true\nusers:\n  - {name: A, vehicle: {name: v}}\n",
		"finished first": "users:\n  - name: A\n    vehicle:\n      name: v\n      sessions:\n        - {created: 2026-01-01T12:00:00Z, finished: 2026-01-01T10:00:00Z}\n",
	}
	for name, y := range bad {
		if _, err := ParseConfig([]byte(y)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}
