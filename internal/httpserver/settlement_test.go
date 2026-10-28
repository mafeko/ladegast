package httpserver

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/store"
)

func TestSettlementsFor_SumsBetweenSettlements(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 10, d, 12, 0, 0, 0, time.UTC) }
	pct := 0.0
	sess := func(d int, kwh float64) evcc.Session {
		return evcc.Session{Created: day(d), Vehicle: "v", ChargedEnergy: kwh, SolarPercentage: &pct}
	}
	friend := store.Friend{
		VehicleName:    "v",
		RateGridPerKWh: 0.5,
		// deliberately unsorted
		Settlements: []store.Settlement{{ID: "b", At: day(20)}, {ID: "a", At: day(10)}},
	}
	got := settlementsFor(friend, []evcc.Session{
		sess(5, 10), sess(15, 20), sess(25, 4),
		{Created: day(5), Vehicle: "other", ChargedEnergy: 99, SolarPercentage: &pct},
	})

	if len(got.Settlements) != 2 || got.Settlements[0].ID != "a" || got.Settlements[1].ID != "b" {
		t.Fatalf("settlements = %+v, want a,b", got.Settlements)
	}
	want := []float64{5, 10, 2} // first, second, open
	have := []float64{got.Settlements[0].AmountEUR, got.Settlements[1].AmountEUR, got.Open.AmountEUR}
	for i := range want {
		if math.Abs(have[i]-want[i]) > 1e-9 {
			t.Errorf("amount[%d] = %v, want %v", i, have[i], want[i])
		}
	}
	if got.Open.Since == nil || !got.Open.Since.Equal(day(20)) {
		t.Errorf("open.since = %v, want %v", got.Open.Since, day(20))
	}
}

func TestAdminSettlements_CreateListDelete(t *testing.T) {
	client, base := adminClient(t)
	id, pin := createFriendVia(t, client, base, createFriendRequest{
		Name: "Alex", VehicleName: "vehicle_1", RateSolarPerKWh: 0.1, RateGridPerKWh: 0.3,
	})

	type view struct {
		Settlements []struct {
			ID        string
			AmountEUR float64
		}
		Open struct{ AmountEUR float64 }
	}
	list := func() view {
		resp, err := client.Get(base + "/admin/friends/" + id + "/view-as/settlements")
		if err != nil {
			t.Fatal(err)
		}
		var v view
		mustDecode(t, resp, &v)
		return v
	}

	if v := list(); len(v.Settlements) != 0 || math.Abs(v.Open.AmountEUR-2) > 1e-9 {
		t.Fatalf("initial = %+v, want no settlements and open 2.00", v)
	}

	resp := doJSON(t, client, http.MethodPost, base+"/admin/friends/"+id+"/settlements", "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()

	v := list()
	if len(v.Settlements) != 1 || math.Abs(v.Settlements[0].AmountEUR-2) > 1e-9 || v.Open.AmountEUR != 0 {
		t.Fatalf("after create = %+v, want one settlement of 2.00 and open 0", v)
	}

	// friend sees the same data
	friendClient := newClientWithJar(t)
	mustPostJSON(t, friendClient, base+"/public/l/"+id+"/login", map[string]string{"pin": pin}).Body.Close()
	fresp, err := friendClient.Get(base + "/public/l/" + id + "/settlements")
	if err != nil || fresp.StatusCode != http.StatusOK {
		t.Fatalf("friend settlements: %v / %v", err, fresp)
	}
	fresp.Body.Close()

	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	resp = doJSON(t, client, http.MethodPost, base+"/admin/friends/"+id+"/settlements", `{"at":"`+future+`"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("future settlement status = %d, want 400", resp.StatusCode)
	}

	resp = doJSON(t, client, http.MethodDelete, base+"/admin/friends/"+id+"/settlements/"+v.Settlements[0].ID, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", resp.StatusCode)
	}
	if v := list(); len(v.Settlements) != 0 {
		t.Errorf("after delete = %+v, want none", v)
	}
	resp = doJSON(t, client, http.MethodDelete, base+"/admin/friends/"+id+"/settlements/nope", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("delete unknown status = %d, want 404", resp.StatusCode)
	}
}
