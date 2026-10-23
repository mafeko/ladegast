package httpserver

import (
	"testing"
	"time"
)

func TestStartOfWeek_Monday(t *testing.T) {
	// 2026-10-17 is a Saturday; the ISO week should start on 2026-10-12 (Monday).
	sat := time.Date(2026, 10, 17, 15, 30, 0, 0, time.UTC)
	got := startOfWeek(sat)
	want := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("startOfWeek(%v) = %v, want %v", sat, got, want)
	}
}

func TestStartOfWeek_AlreadyMonday(t *testing.T) {
	mon := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	got := startOfWeek(mon)
	if !got.Equal(startOfDay(mon)) {
		t.Errorf("startOfWeek(Monday) = %v, want %v", got, startOfDay(mon))
	}
}

func TestStartOfMonthAndYear(t *testing.T) {
	ts := time.Date(2026, 10, 17, 15, 30, 0, 0, time.UTC)
	if got, want := startOfMonth(ts), time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("startOfMonth = %v, want %v", got, want)
	}
	if got, want := startOfYear(ts), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("startOfYear = %v, want %v", got, want)
	}
}
