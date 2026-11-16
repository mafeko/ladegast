package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminLocalOnly(t *testing.T) {
	cases := []struct {
		name        string
		allowRemote bool
		remote      string
		header      [2]string
		path        string
		want        int
	}{
		{"loopback", false, "127.0.0.1:1234", [2]string{}, "/admin/guests", http.StatusUnauthorized},
		{"lan", false, "192.168.1.20:1234", [2]string{}, "/admin/guests", http.StatusUnauthorized},
		{"tailnet", false, "100.101.1.1:1", [2]string{}, "/admin/guests", http.StatusUnauthorized},
		{"public ip", false, "8.8.8.8:1", [2]string{}, "/admin/guests", http.StatusNotFound},
		{"tunnel on localhost", false, "127.0.0.1:1", [2]string{"X-Forwarded-For", "8.8.8.8"}, "/admin/guests", http.StatusNotFound},
		{"cloudflare header", false, "172.18.0.3:1", [2]string{"Cf-Connecting-Ip", "8.8.8.8"}, "/admin", http.StatusNotFound},
		{"admin spa page remote", false, "8.8.8.8:1", [2]string{}, "/admin", http.StatusNotFound},
		{"public path unaffected", false, "8.8.8.8:1", [2]string{"X-Forwarded-For", "8.8.8.8"}, "/public/l/x/status", http.StatusNotFound},
		{"similar prefix unaffected", false, "8.8.8.8:1", [2]string{}, "/administrator", http.StatusNotFound},
		{"allow remote", true, "8.8.8.8:1", [2]string{"X-Forwarded-For", "8.8.8.8"}, "/admin/guests", http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, _, _ := testServer(t, "http://127.0.0.1:1")
			srv.cfg.AdminAllowRemote = c.allowRemote
			req := httptest.NewRequest(http.MethodGet, c.path, nil)
			req.RemoteAddr = c.remote
			if c.header[0] != "" {
				req.Header.Set(c.header[0], c.header[1])
			}
			rec := httptest.NewRecorder()
			srv.Handler().ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("got %d, want %d", rec.Code, c.want)
			}
		})
	}
}
