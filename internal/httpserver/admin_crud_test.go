package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// adminClient returns a logged-in admin client against a fresh server.
func adminClient(t *testing.T) (*http.Client, string) {
	t.Helper()
	evccSrv := fakeEVCC(t)
	t.Cleanup(evccSrv.Close)
	srv, st, cfg := testServer(t, evccSrv.URL)
	cfg.AdminUsername, cfg.AdminPassword = "admin", "s3cret-password"
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}
	httpSrv := httptest.NewServer(srv.Handler())
	t.Cleanup(httpSrv.Close)
	client := newClientWithJar(t)
	mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "s3cret-password"})
	return client, httpSrv.URL
}

func doJSON(t *testing.T, client *http.Client, method, url, body string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func createFriendVia(t *testing.T, client *http.Client, base string, req createFriendRequest) (id, pin string) {
	t.Helper()
	resp := mustPostJSON(t, client, base+"/admin/friends", req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	var created struct{ ID, PIN, Link string }
	mustDecode(t, resp, &created)
	return created.ID, created.PIN
}

func loginStatus(t *testing.T, base, id, pin string) int {
	t.Helper()
	resp := mustPostJSON(t, newClientWithJar(t), base+"/public/l/"+id+"/login", map[string]string{"pin": pin})
	resp.Body.Close()
	return resp.StatusCode
}

func TestAdminCreateFriend_ChosenPIN(t *testing.T) {
	client, base := adminClient(t)

	id, pin := createFriendVia(t, client, base, createFriendRequest{Name: "Alex", VehicleName: "vehicle_1", PIN: "424242"})
	if pin != "424242" {
		t.Errorf("returned pin = %q, want the chosen one", pin)
	}
	if got := loginStatus(t, base, id, "424242"); got != http.StatusOK {
		t.Errorf("login with chosen PIN = %d, want 200", got)
	}
}

func TestAdminCreateFriend_InvalidInput(t *testing.T) {
	client, base := adminClient(t)
	cases := map[string]createFriendRequest{
		"short pin":     {Name: "A", VehicleName: "v", PIN: "123"},
		"non-digit pin": {Name: "A", VehicleName: "v", PIN: "12ab56"},
		"long pin":      {Name: "A", VehicleName: "v", PIN: "1234567890123"},
		"negative rate": {Name: "A", VehicleName: "v", RateSolarPerKWh: -1},
	}
	for name, req := range cases {
		resp := mustPostJSON(t, client, base+"/admin/friends", req)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, resp.StatusCode)
		}
	}
}

func TestAdminListFriends_HasLinkAndHidesPINHash(t *testing.T) {
	client, base := adminClient(t)
	id, _ := createFriendVia(t, client, base, createFriendRequest{Name: "Alex", VehicleName: "vehicle_1"})

	resp, err := client.Get(base + "/admin/friends")
	if err != nil {
		t.Fatal(err)
	}
	var raw []map[string]any
	mustDecode(t, resp, &raw)
	if len(raw) != 1 {
		t.Fatalf("len = %d, want 1", len(raw))
	}
	if _, leaked := raw[0]["pinHash"]; leaked {
		t.Error("list response leaks pinHash")
	}
	if want := "/public/l/" + id; !strings.HasSuffix(raw[0]["link"].(string), want) {
		t.Errorf("link = %v, want suffix %s", raw[0]["link"], want)
	}
}

func TestAdminResetPIN_ChosenAndInvalid(t *testing.T) {
	client, base := adminClient(t)
	id, oldPIN := createFriendVia(t, client, base, createFriendRequest{Name: "Alex", VehicleName: "vehicle_1"})

	resp := doJSON(t, client, http.MethodPost, base+"/admin/friends/"+id+"/reset-pin", `{"pin":"7788"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset with chosen PIN = %d, want 200", resp.StatusCode)
	}
	var out struct{ PIN string }
	mustDecode(t, resp, &out)
	if out.PIN != "7788" {
		t.Errorf("pin = %q, want 7788", out.PIN)
	}
	if got := loginStatus(t, base, id, "7788"); got != http.StatusOK {
		t.Errorf("login with new PIN = %d, want 200", got)
	}
	if got := loginStatus(t, base, id, oldPIN); got != http.StatusUnauthorized {
		t.Errorf("login with old PIN = %d, want 401", got)
	}

	resp = doJSON(t, client, http.MethodPost, base+"/admin/friends/"+id+"/reset-pin", `{"pin":"x"}`)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid chosen PIN = %d, want 400", resp.StatusCode)
	}

	// No body at all still generates a random PIN.
	resp = doJSON(t, client, http.MethodPost, base+"/admin/friends/"+id+"/reset-pin", "")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("reset without body = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAdminDeleteFriend(t *testing.T) {
	client, base := adminClient(t)
	id, pin := createFriendVia(t, client, base, createFriendRequest{Name: "Alex", VehicleName: "vehicle_1"})

	resp := doJSON(t, client, http.MethodDelete, base+"/admin/friends/"+id, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", resp.StatusCode)
	}
	if got := loginStatus(t, base, id, pin); got != http.StatusUnauthorized {
		t.Errorf("login after delete = %d, want 401", got)
	}
	status, _ := client.Get(base + "/public/l/" + id + "/status")
	status.Body.Close()
	if status.StatusCode != http.StatusNotFound {
		t.Errorf("link status after delete = %d, want 404", status.StatusCode)
	}

	resp = doJSON(t, client, http.MethodDelete, base+"/admin/friends/"+id, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("second delete = %d, want 404", resp.StatusCode)
	}
}

func TestAdminDeleteFriend_RequiresAdmin(t *testing.T) {
	_, base := adminClient(t)
	resp := doJSON(t, newClientWithJar(t), http.MethodDelete, base+"/admin/friends/whatever", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unauthenticated delete = %d, want 401", resp.StatusCode)
	}
}

func TestAdminListVehicles_OnlyUnassignedFromSessions(t *testing.T) {
	client, base := adminClient(t)

	resp := doJSON(t, client, http.MethodGet, base+"/admin/vehicles", "")
	var names []string
	mustDecode(t, resp, &names)
	if len(names) != 2 || names[0] != "other_vehicle" || names[1] != "vehicle_1" {
		t.Fatalf("vehicles = %v, want [other_vehicle vehicle_1]", names)
	}

	createFriendVia(t, client, base, createFriendRequest{Name: "Alex", VehicleName: "vehicle_1"})
	resp = doJSON(t, client, http.MethodGet, base+"/admin/vehicles", "")
	names = nil
	mustDecode(t, resp, &names)
	if len(names) != 1 || names[0] != "other_vehicle" {
		t.Fatalf("vehicles after assign = %v, want [other_vehicle]", names)
	}
}

func TestAdminListVehicles_RequiresAdmin(t *testing.T) {
	_, base := adminClient(t)
	resp := doJSON(t, newClientWithJar(t), http.MethodGet, base+"/admin/vehicles", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
