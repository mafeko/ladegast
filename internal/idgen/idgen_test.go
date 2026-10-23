package idgen

import (
	"regexp"
	"testing"
)

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestUUID_FormatAndUniqueness(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := UUID()
		if !uuidRE.MatchString(id) {
			t.Fatalf("UUID() = %q, does not match v4 pattern", id)
		}
		if seen[id] {
			t.Fatalf("UUID() produced duplicate %q", id)
		}
		seen[id] = true
	}
}

func TestPIN_LengthAndRange(t *testing.T) {
	for i := 0; i < 1000; i++ {
		pin := PIN(6)
		if len(pin) != 6 {
			t.Fatalf("PIN(6) = %q, want length 6", pin)
		}
		for _, c := range pin {
			if c < '0' || c > '9' {
				t.Fatalf("PIN(6) = %q, contains non-digit", pin)
			}
		}
	}
}
