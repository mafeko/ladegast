// Package httpserver wires Ladegast's public (friend-facing) and admin
// HTTP APIs. Network-level separation between the two (e.g. only exposing
// the public prefix through a reverse proxy) is the operator's
// responsibility - see DESIGN.md section 1.
package httpserver

import (
	"net/http"
	"strings"

	"github.com/mafeko/ladegast/internal/config"
	"github.com/mafeko/ladegast/internal/evcc"
	"github.com/mafeko/ladegast/internal/store"
)

type Server struct {
	cfg           config.Config
	store         *store.Store
	evcc          *evcc.Client
	secret        []byte
	secureCookies bool
}

func New(cfg config.Config, st *store.Store, evccClient *evcc.Client) *Server {
	return &Server{
		cfg:           cfg,
		store:         st,
		evcc:          evccClient,
		secret:        []byte(cfg.SessionSecret),
		secureCookies: strings.HasPrefix(cfg.PublicBaseURL, "https://"),
	}
}

// Handler returns the complete HTTP handler, routing both the public and
// admin path prefixes from cfg.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerPublicRoutes(mux)
	s.registerAdminRoutes(mux)
	return mux
}

func (s *Server) publicPath(suffix string) string {
	return s.cfg.PublicPathPrefix + suffix
}

func (s *Server) adminPath(suffix string) string {
	return s.cfg.AdminPathPrefix + suffix
}
