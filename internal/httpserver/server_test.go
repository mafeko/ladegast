package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/config"
	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/store"
)

func fakeEVCC(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/sessions":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[
				{"id":1,"created":"2026-10-17T10:00:00Z","finished":"2026-10-17T12:00:00Z","loadpoint":"warp4","vehicle":"vehicle_1","chargedEnergy":10,"solarPercentage":50},
				{"id":2,"created":"2026-10-17T10:00:00Z","finished":"2026-10-17T12:00:00Z","loadpoint":"warp4","vehicle":"other_vehicle","chargedEnergy":100,"solarPercentage":0}
			]`))
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/vehicles/"):
			w.Write([]byte(`{"mode":"smart"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func testServer(t *testing.T, evccURL string) (*Server, *store.Store, config.Config) {
	t.Helper()
	cfg := config.Config{
		PublicBaseURL:        "http://localhost",
		PublicPathPrefix:     "/public",
		AdminPathPrefix:      "/admin",
		SessionSecret:        "test-secret",
		RateLimitMaxAttempts: 3,
		RateLimitWindow:      15 * time.Minute,
	}
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	client := evcc.NewClient(evccURL, "test-key")
	return New(cfg, st, client), st, cfg
}

func newClientWithJar(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}
	return &http.Client{Jar: jar}
}

func createFriendDirect(t *testing.T, st *store.Store, pin string) store.Friend {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	f := store.Friend{
		ID:              "11111111-1111-4111-8111-111111111111",
		Name:            "Alex",
		VehicleName:     "vehicle_1",
		PINHash:         string(hash),
		RateSolarPerKWh: 0.10,
		RateGridPerKWh:  0.35,
		CreatedAt:       time.Now(),
	}
	if err := st.CreateFriend(f); err != nil {
		t.Fatalf("CreateFriend: %v", err)
	}
	return f
}

func TestFriendFlow(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	// Status before login.
	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID)
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	var status struct{ Authenticated bool }
	mustDecode(t, resp, &status)
	if status.Authenticated {
		t.Fatal("expected authenticated=false before login")
	}

	// Wrong PIN.
	resp = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "000000"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong PIN status = %d, want 401", resp.StatusCode)
	}

	// Correct PIN.
	resp = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("correct PIN status = %d, want 200", resp.StatusCode)
	}

	// Dashboard only contains this friend's vehicle (10 kWh @ 50% solar).
	resp, err = client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/dashboard")
	if err != nil {
		t.Fatalf("GET dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dashboard status = %d, want 200", resp.StatusCode)
	}
	var dash dashboard
	mustDecode(t, resp, &dash)
	if dash.ThisYear.ChargedKWh != 10 {
		t.Errorf("ThisYear.ChargedKWh = %v, want 10 (other_vehicle must be excluded)", dash.ThisYear.ChargedKWh)
	}
	wantCost := 5*0.10 + 5*0.35
	if diff := dash.ThisYear.CostEUR - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("ThisYear.CostEUR = %v, want %v", dash.ThisYear.CostEUR, wantCost)
	}

	// Sessions list.
	resp, err = client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/sessions")
	if err != nil {
		t.Fatalf("GET sessions: %v", err)
	}
	var sessions []map[string]any
	mustDecode(t, resp, &sessions)
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}

	// Mode switch.
	resp = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/mode", map[string]string{"mode": "smart"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("mode switch status = %d, want 200", resp.StatusCode)
	}

	// Invalid mode rejected.
	resp = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/mode", map[string]string{"mode": "off"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("mode=off status = %d, want 400", resp.StatusCode)
	}
}

func TestFriendLogin_RateLimited(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	var last *http.Response
	for i := 0; i < 3; i++ {
		last = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "000000"})
	}
	if last.StatusCode != http.StatusUnauthorized {
		t.Fatalf("3rd wrong attempt status = %d, want 401", last.StatusCode)
	}

	locked := mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})
	if locked.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("4th attempt (correct PIN) status = %d, want 429", locked.StatusCode)
	}
}

func TestAdminFlow(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, cfg := testServer(t, evccSrv.URL)
	cfg.AdminUsername = "admin"
	cfg.AdminPassword = "s3cret-password"
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	resp := mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "wrong"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong admin password status = %d, want 401", resp.StatusCode)
	}

	resp = mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "s3cret-password"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin login status = %d, want 200", resp.StatusCode)
	}

	resp = mustPostJSON(t, client, httpSrv.URL+"/admin/friends", createFriendRequest{
		Name: "Alex", VehicleName: "vehicle_1", RateSolarPerKWh: 0.10, RateGridPerKWh: 0.35,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create friend status = %d, want 201", resp.StatusCode)
	}
	var created struct{ ID, PIN, Link string }
	mustDecode(t, resp, &created)
	if created.ID == "" || created.PIN == "" {
		t.Fatalf("created = %+v, missing id/pin", created)
	}

	resp, err := client.Get(httpSrv.URL + "/admin/friends")
	if err != nil {
		t.Fatalf("GET friends: %v", err)
	}
	var friends []store.Friend
	mustDecode(t, resp, &friends)
	if len(friends) != 1 {
		t.Fatalf("len(friends) = %d, want 1", len(friends))
	}

	resp, err = client.Get(httpSrv.URL + "/admin/friends/" + created.ID + "/view-as/dashboard")
	if err != nil {
		t.Fatalf("GET view-as dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("view-as dashboard status = %d, want 200", resp.StatusCode)
	}

	// Regenerate link: old id gone, new id works.
	resp = mustPostJSON(t, client, httpSrv.URL+"/admin/friends/"+created.ID+"/regenerate-link", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("regenerate-link status = %d, want 200", resp.StatusCode)
	}
	var regenerated struct{ ID, Link string }
	mustDecode(t, resp, &regenerated)
	if regenerated.ID == created.ID {
		t.Fatal("regenerated ID should differ from original")
	}

	oldStatusResp, _ := client.Get(httpSrv.URL + "/public/l/" + created.ID)
	if oldStatusResp.StatusCode != http.StatusNotFound {
		t.Fatalf("old link status = %d, want 404 after regeneration", oldStatusResp.StatusCode)
	}

	// Reset PIN.
	resp = mustPostJSON(t, client, httpSrv.URL+"/admin/friends/"+regenerated.ID+"/reset-pin", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset-pin status = %d, want 200", resp.StatusCode)
	}
}

func mustPostJSON(t *testing.T, client *http.Client, url string, body any) *http.Response {
	t.Helper()
	var reader *strings.Reader
	if body == nil {
		reader = strings.NewReader("{}")
	} else {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = strings.NewReader(string(data))
	}
	resp, err := client.Post(url, "application/json", reader)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func mustDecode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
