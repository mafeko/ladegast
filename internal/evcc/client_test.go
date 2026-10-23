package evcc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization header = %q", got)
		}
		if r.URL.Path != "/api/sessions" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("year") != "2026" {
			t.Errorf("year query = %q", r.URL.Query().Get("year"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1,"created":"2026-10-17T10:00:00Z","finished":"2026-10-17T12:00:00Z","loadpoint":"warp4","vehicle":"vehicle_1","chargedEnergy":10,"solarPercentage":50}]`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key")
	sessions, err := c.Sessions(context.Background(), 2026, 0)
	if err != nil {
		t.Fatalf("Sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
	if sessions[0].Vehicle != "vehicle_1" {
		t.Errorf("Vehicle = %q, want vehicle_1", sessions[0].Vehicle)
	}
	if sessions[0].SolarKWh() != 5 {
		t.Errorf("SolarKWh = %v, want 5", sessions[0].SolarKWh())
	}
	if sessions[0].GridKWh() != 5 {
		t.Errorf("GridKWh = %v, want 5", sessions[0].GridKWh())
	}
}

func TestSessions_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad-key")
	if _, err := c.Sessions(context.Background(), 0, 0); err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestSetVehicleMode(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Write([]byte(`{"mode":"smart"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key")
	if err := c.SetVehicleMode(context.Background(), "vehicle_1", ModeSmart); err != nil {
		t.Fatalf("SetVehicleMode: %v", err)
	}
	if gotPath != "/api/vehicles/vehicle_1/mode/smart" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestSetVehicleMode_RejectsUnknownMode(t *testing.T) {
	c := NewClient("http://example.invalid", "test-key")
	if err := c.SetVehicleMode(context.Background(), "vehicle_1", Mode("off")); err == nil {
		t.Fatal("expected error for disallowed mode, got nil")
	}
}

func TestSetVehicleMode_RequiresVehicleName(t *testing.T) {
	c := NewClient("http://example.invalid", "test-key")
	if err := c.SetVehicleMode(context.Background(), "", ModeNow); err == nil {
		t.Fatal("expected error for empty vehicle name, got nil")
	}
}
