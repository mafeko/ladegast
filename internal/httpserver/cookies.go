package httpserver

import (
	"net/http"
	"time"

	"github.com/mafeko/ladegast/internal/authtoken"
)

const (
	friendCookieName = "ladegast_friend_session"
	adminCookieName  = "ladegast_admin_session"
	sessionTTL       = 7 * 24 * time.Hour
)

func (s *Server) setSessionCookie(w http.ResponseWriter, name, subject string) error {
	token, err := authtoken.New(s.secret, subject, sessionTTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	return nil
}

func (s *Server) clearSessionCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookies,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

// sessionSubject returns the verified subject carried by the named cookie,
// or "" if absent/invalid/expired.
func (s *Server) sessionSubject(r *http.Request, cookieName string) string {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	subject, err := authtoken.Verify(s.secret, c.Value)
	if err != nil {
		return ""
	}
	return subject
}
