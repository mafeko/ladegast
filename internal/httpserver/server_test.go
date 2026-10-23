package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/authtoken"
	"github.com/mafeko/ladegast/internal/config"
	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/store"
)

// testAssets is a minimal stand-in for the embedded frontend build, kept
// independent of internal/webui so these tests don't depend on whether a
// real frontend has been built.
func testAssets() fstest.MapFS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>test</body></html>")},
	}
}

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
	return New(cfg, st, client, testAssets()), st, cfg
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
	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/status")
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	var status struct{ Authenticated bool }
	mustDecode(t, resp, &status)
	if status.Authenticated {
		t.Fatal("expected authenticated=false before login")
	}

	// The bare link itself serves the embedded SPA shell, regardless of
	// login state - the frontend router decides what to show.
	resp, err = client.Get(httpSrv.URL + "/public/l/" + friend.ID)
	if err != nil {
		t.Fatalf("GET page: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page status = %d, want 200", resp.StatusCode)
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

	oldStatusResp, _ := client.Get(httpSrv.URL + "/public/l/" + created.ID + "/status")
	if oldStatusResp.StatusCode != http.StatusNotFound {
		t.Fatalf("old link status = %d, want 404 after regeneration", oldStatusResp.StatusCode)
	}

	// Reset PIN.
	resp = mustPostJSON(t, client, httpSrv.URL+"/admin/friends/"+regenerated.ID+"/reset-pin", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset-pin status = %d, want 200", resp.StatusCode)
	}
}

func TestFriendLogin_UnknownUUID(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, _, _ := testServer(t, evccSrv.URL)
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	resp := mustPostJSON(t, client, httpSrv.URL+"/public/l/00000000-0000-4000-8000-000000000000/login", map[string]string{"pin": "123456"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("login for unknown uuid status = %d, want 401", resp.StatusCode)
	}
}

func TestFriendLogin_InvalidBody(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	resp, err := http.Post(httpSrv.URL+"/public/l/"+friend.ID+"/login", "application/json", strings.NewReader("not json"))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestFriendLogout(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})

	resp := mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/logout", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", resp.StatusCode)
	}

	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/dashboard")
	if err != nil {
		t.Fatalf("GET dashboard after logout: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("dashboard after logout status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireFriend_CookieForDifferentFriend(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friendA := createFriendDirect(t, st, "111111")
	friendB := store.Friend{ID: "22222222-2222-4222-8222-222222222222", Name: "Bea", VehicleName: "vehicle_2"}
	hash, _ := bcrypt.GenerateFromPassword([]byte("222222"), bcrypt.DefaultCost)
	friendB.PINHash = string(hash)
	if err := st.CreateFriend(friendB); err != nil {
		t.Fatalf("CreateFriend: %v", err)
	}

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	// Log in as friendA, then try to use that cookie against friendB's path.
	mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friendA.ID+"/login", map[string]string{"pin": "111111"})
	resp, err := client.Get(httpSrv.URL + "/public/l/" + friendB.ID + "/dashboard")
	if err != nil {
		t.Fatalf("GET dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("cross-friend dashboard access status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireFriend_DeletedFriendStillHasValidCookie(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})
	if err := st.DeleteFriend(friend.ID); err != nil {
		t.Fatalf("DeleteFriend: %v", err)
	}

	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/dashboard")
	if err != nil {
		t.Fatalf("GET dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("dashboard for deleted friend status = %d, want 401", resp.StatusCode)
	}
}

func TestFriendHandlers_EVCCUnreachable(t *testing.T) {
	deadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadSrv.Close() // closed before use: connections will be refused

	srv, st, _ := testServer(t, deadSrv.URL)
	friend := createFriendDirect(t, st, "123456")
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)

	mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})

	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/dashboard")
	if err != nil {
		t.Fatalf("GET dashboard: %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("dashboard with EVCC down status = %d, want 502", resp.StatusCode)
	}

	resp, err = client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/sessions")
	if err != nil {
		t.Fatalf("GET sessions: %v", err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("sessions with EVCC down status = %d, want 502", resp.StatusCode)
	}

	resp = mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/mode", map[string]string{"mode": "smart"})
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("mode switch with EVCC down status = %d, want 502", resp.StatusCode)
	}
}

func TestAdminUpdateFriend(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, cfg := testServer(t, evccSrv.URL)
	cfg.AdminUsername, cfg.AdminPassword = "admin", "s3cret-password"
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}
	friend := createFriendDirect(t, st, "123456")

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)
	mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "s3cret-password"})

	req, _ := http.NewRequest(http.MethodPut, httpSrv.URL+"/admin/friends/"+friend.ID, strings.NewReader(`{"name":"Alexandra","vehicleName":"vehicle_1","rateSolarPerKWh":0.2,"rateGridPerKWh":0.4}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("PUT friend: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update friend status = %d, want 200", resp.StatusCode)
	}
	var updated store.Friend
	mustDecode(t, resp, &updated)
	if updated.Name != "Alexandra" || updated.RateSolarPerKWh != 0.2 {
		t.Errorf("updated = %+v, want Name=Alexandra RateSolarPerKWh=0.2", updated)
	}

	// Unknown friend id.
	req2, _ := http.NewRequest(http.MethodPut, httpSrv.URL+"/admin/friends/does-not-exist", strings.NewReader(`{"name":"X","vehicleName":"Y"}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatalf("PUT unknown friend: %v", err)
	}
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("update unknown friend status = %d, want 404", resp2.StatusCode)
	}
}

func TestAdminViewAsSessions(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, cfg := testServer(t, evccSrv.URL)
	cfg.AdminUsername, cfg.AdminPassword = "admin", "s3cret-password"
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}
	friend := createFriendDirect(t, st, "123456")

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)
	mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "s3cret-password"})

	resp, err := client.Get(httpSrv.URL + "/admin/friends/" + friend.ID + "/view-as/sessions")
	if err != nil {
		t.Fatalf("GET view-as sessions: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("view-as sessions status = %d, want 200", resp.StatusCode)
	}
	var sessions []map[string]any
	mustDecode(t, resp, &sessions)
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}

	resp, err = client.Get(httpSrv.URL + "/admin/friends/does-not-exist/view-as/sessions")
	if err != nil {
		t.Fatalf("GET view-as sessions (unknown): %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("view-as sessions (unknown) status = %d, want 404", resp.StatusCode)
	}
}

func TestAdminLogout(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, st, cfg := testServer(t, evccSrv.URL)
	cfg.AdminUsername, cfg.AdminPassword = "admin", "s3cret-password"
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)
	mustPostJSON(t, client, httpSrv.URL+"/admin/login", map[string]string{"username": "admin", "password": "s3cret-password"})

	resp := mustPostJSON(t, client, httpSrv.URL+"/admin/logout", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d, want 200", resp.StatusCode)
	}

	resp, err := client.Get(httpSrv.URL + "/admin/friends")
	if err != nil {
		t.Fatalf("GET friends after logout: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("friends after logout status = %d, want 401", resp.StatusCode)
	}
}

func TestRequireAdmin_ValidTokenForNonexistentAdmin(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, _, _ := testServer(t, evccSrv.URL)
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	// A correctly signed token for a username that was never created (e.g.
	// the account was since removed): must be rejected like any other
	// unauthenticated request, not crash or be trusted blindly.
	token, err := authtoken.New(srv.secret, "ghost-admin", time.Hour)
	if err != nil {
		t.Fatalf("authtoken.New: %v", err)
	}
	req, _ := http.NewRequest(http.MethodGet, httpSrv.URL+"/admin/friends", nil)
	req.AddCookie(&http.Cookie{Name: adminCookieName, Value: token})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET friends: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEnsureAdminBootstrap_NoOpWhenAdminExists(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	if err := st.CreateAdmin(store.Admin{Username: "existing", PasswordHash: "hash"}); err != nil {
		t.Fatalf("CreateAdmin: %v", err)
	}

	// No credentials supplied, but an admin already exists: must not error.
	cfg := config.Config{}
	if err := EnsureAdminBootstrap(st, cfg); err != nil {
		t.Fatalf("EnsureAdminBootstrap: %v", err)
	}
	a, err := st.GetAdmin("existing")
	if err != nil || a.PasswordHash != "hash" {
		t.Errorf("existing admin was modified: %+v, err=%v", a, err)
	}
}

func TestEnsureAdminBootstrap_MissingCredentials(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer st.Close()

	if err := EnsureAdminBootstrap(st, config.Config{}); err == nil {
		t.Fatal("expected error when no admin exists and no credentials are set, got nil")
	}
}

func TestFrontendRoutes(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

	srv, _, _ := testServer(t, evccSrv.URL)
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	pageRoutes := []string{
		"/",
		"/public/l/11111111-1111-4111-8111-111111111111",
		"/admin",
		"/admin/view-as/11111111-1111-4111-8111-111111111111",
	}
	for _, path := range pageRoutes {
		resp, err := http.Get(httpSrv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
		ct := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "text/html") {
			t.Errorf("GET %s content-type = %q, want text/html", path, ct)
		}
	}
}

func TestFrontendRoutes_NilAssetsDisablesThem(t *testing.T) {
	evccSrv := fakeEVCC(t)
	defer evccSrv.Close()

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
	defer st.Close()
	srv := New(cfg, st, evcc.NewClient(evccSrv.URL, "test-key"), nil)

	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()

	resp, err := http.Get(httpSrv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET / status = %d, want 404 when no assets are configured", resp.StatusCode)
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		remoteAddr string
		want       string
	}{
		{"203.0.113.5:54321", "203.0.113.5"},
		{"not-a-host-port", "not-a-host-port"},
	}
	for _, c := range cases {
		req := &http.Request{RemoteAddr: c.remoteAddr}
		if got := clientIP(req); got != c.want {
			t.Errorf("clientIP(%q) = %q, want %q", c.remoteAddr, got, c.want)
		}
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
