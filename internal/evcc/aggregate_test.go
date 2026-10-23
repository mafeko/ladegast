package evcc

import (
	"testing"
	"time"
)

func pct(v float64) *float64 { return &v }

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return tm
}

func TestFilterByVehicle(t *testing.T) {
	sessions := []Session{
		{Vehicle: "vehicle_1", ChargedEnergy: 1},
		{Vehicle: "vehicle_2", ChargedEnergy: 2},
		{Vehicle: "vehicle_1", ChargedEnergy: 3},
	}
	got := FilterByVehicle(sessions, "vehicle_1")
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
}

func TestSummarize(t *testing.T) {
	sessions := []Session{
		{ChargedEnergy: 10, SolarPercentage: pct(50)}, // 5 solar, 5 grid
		{ChargedEnergy: 4, SolarPercentage: pct(100)}, // 4 solar, 0 grid
	}
	sum := Summarize(sessions, 0.10, 0.35)
	if sum.ChargedKWh != 14 {
		t.Errorf("ChargedKWh = %v, want 14", sum.ChargedKWh)
	}
	if sum.SolarKWh != 9 {
		t.Errorf("SolarKWh = %v, want 9", sum.SolarKWh)
	}
	if sum.GridKWh != 5 {
		t.Errorf("GridKWh = %v, want 5", sum.GridKWh)
	}
	wantCost := 9*0.10 + 5*0.35
	if diff := sum.CostEUR - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("CostEUR = %v, want %v", sum.CostEUR, wantCost)
	}
}

func TestGroupBy_Day(t *testing.T) {
	sessions := []Session{
		{Created: mustParse(t, "2026-10-17T10:00:00Z"), ChargedEnergy: 5, SolarPercentage: pct(0)},
		{Created: mustParse(t, "2026-10-17T22:00:00Z"), ChargedEnergy: 3, SolarPercentage: pct(0)},
		{Created: mustParse(t, "2026-10-18T10:00:00Z"), ChargedEnergy: 2, SolarPercentage: pct(0)},
	}
	groups := GroupBy(sessions, PeriodDay, 0.10, 0.35)
	if len(groups) != 2 {
		t.Fatalf("len(groups) = %d, want 2", len(groups))
	}
	if got := groups["2026-10-17"].ChargedKWh; got != 8 {
		t.Errorf("2026-10-17 ChargedKWh = %v, want 8", got)
	}
	if got := groups["2026-10-18"].ChargedKWh; got != 2 {
		t.Errorf("2026-10-18 ChargedKWh = %v, want 2", got)
	}
}

func TestGroupBy_Year(t *testing.T) {
	sessions := []Session{
		{Created: mustParse(t, "2026-01-01T00:00:00Z"), ChargedEnergy: 1, SolarPercentage: pct(0)},
		{Created: mustParse(t, "2026-12-31T00:00:00Z"), ChargedEnergy: 1, SolarPercentage: pct(0)},
	}
	groups := GroupBy(sessions, PeriodYear, 0, 0)
	if len(groups) != 1 {
		t.Fatalf("len(groups) = %d, want 1", len(groups))
	}
	if got := groups["2026"].ChargedKWh; got != 2 {
		t.Errorf("2026 ChargedKWh = %v, want 2", got)
	}
}
