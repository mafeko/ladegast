package evcc

import "fmt"

// Summary totals a set of sessions in Ladegast's own currency terms, using
// per-friend rates rather than EVCC's own (possibly dynamic) tariff.
type Summary struct {
	ChargedKWh float64
	SolarKWh   float64
	GridKWh    float64
	CostEUR    float64
}

// FilterByVehicle returns only the sessions belonging to the given EVCC
// vehicle name, so one friend never sees another friend's charging data.
func FilterByVehicle(sessions []Session, vehicleName string) []Session {
	var out []Session
	for _, s := range sessions {
		if s.Vehicle == vehicleName {
			out = append(out, s)
		}
	}
	return out
}

// Summarize totals charged/solar/grid energy and cost for sessions, using
// the given per-kWh rates.
func Summarize(sessions []Session, rateSolarPerKWh, rateGridPerKWh float64) Summary {
	var sum Summary
	for _, s := range sessions {
		solar := s.SolarKWh()
		grid := s.GridKWh()
		sum.ChargedKWh += s.ChargedEnergy
		sum.SolarKWh += solar
		sum.GridKWh += grid
		sum.CostEUR += solar*rateSolarPerKWh + grid*rateGridPerKWh
	}
	return sum
}

// Period is a granularity for grouping sessions in the dashboard.
type Period string

const (
	PeriodDay   Period = "day"
	PeriodWeek  Period = "week"
	PeriodMonth Period = "month"
	PeriodYear  Period = "year"
)

// Bucket returns the group key for a single session under the given
// period, e.g. "2026-10-17" (day), "2026-W42" (week), "2026-10" (month),
// "2026" (year).
func Bucket(period Period, s Session) string {
	t := s.Created
	switch period {
	case PeriodDay:
		return t.Format("2006-01-02")
	case PeriodWeek:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case PeriodMonth:
		return t.Format("2006-01")
	case PeriodYear:
		return t.Format("2006")
	default:
		return t.Format("2006-01-02")
	}
}

// GroupBy buckets sessions by period and returns one Summary per bucket key.
func GroupBy(sessions []Session, period Period, rateSolarPerKWh, rateGridPerKWh float64) map[string]Summary {
	byKey := make(map[string][]Session)
	for _, s := range sessions {
		key := Bucket(period, s)
		byKey[key] = append(byKey[key], s)
	}
	out := make(map[string]Summary, len(byKey))
	for key, group := range byKey {
		out[key] = Summarize(group, rateSolarPerKWh, rateGridPerKWh)
	}
	return out
}
