package httpserver

import (
	"net"
	"net/http"
	"strings"
)

// proxyHeaders are set by reverse proxies and tunnels (Cloudflare Tunnel,
// Tailscale Funnel, Caddy, nginx, ...). A request carrying any of them did
// not originate on the local network, even if the proxy itself runs there.
var proxyHeaders = []string{
	"Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto",
	"X-Real-Ip", "Cf-Connecting-Ip", "True-Client-Ip",
}

// tailnet is Tailscale's CGNAT range; the VPN counts as trusted.
var tailnet = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// adminLocalOnly answers requests to the admin path prefix with 404 unless
// they come directly from a trusted address (loopback, private LAN,
// link-local, Tailscale) without proxy headers. Disabled by
// LADEGAST_ADMIN_ALLOW_REMOTE=true.
func (s *Server) adminLocalOnly(next http.Handler) http.Handler {
	if s.cfg.AdminAllowRemote {
		return next
	}
	prefix := s.cfg.AdminPathPrefix
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if (p == prefix || strings.HasPrefix(p, prefix+"/")) && !isLocalRequest(r) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isLocalRequest(r *http.Request) bool {
	for _, h := range proxyHeaders {
		if r.Header.Get(h) != "" {
			return false
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || tailnet.Contains(ip)
}
