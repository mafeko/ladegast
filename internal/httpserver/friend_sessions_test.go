package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFriendSessions_RequestsFullHistoryFromEVCC(t *testing.T) {
	var query string
	evccSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte("[]"))
	}))
	defer evccSrv.Close()

	srv, st, _ := testServer(t, evccSrv.URL)
	friend := createFriendDirect(t, st, "123456")
	httpSrv := httptest.NewServer(srv.Handler())
	defer httpSrv.Close()
	client := newClientWithJar(t)
	mustPostJSON(t, client, httpSrv.URL+"/public/l/"+friend.ID+"/login", map[string]string{"pin": "123456"})

	resp, err := client.Get(httpSrv.URL + "/public/l/" + friend.ID + "/sessions")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if query != "" {
		t.Errorf("EVCC query = %q, want none (full history, filtered client-side)", query)
	}
}
