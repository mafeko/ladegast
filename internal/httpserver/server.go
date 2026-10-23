// Package httpserver wires Ladegast's public (friend-facing) and admin
// HTTP APIs, plus the embedded frontend. Network-level separation between
// the admin surface and everything else (e.g. only exposing
// ADMIN_PATH_PREFIX internally) is the operator's responsibility - see
// DESIGN.md section 1 and DEPLOYMENT.md.
package httpserver

import (
	"io/fs"
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
	assets        fs.FS
	secret        []byte
	secureCookies bool
}

func New(cfg config.Config, st *store.Store, evccClient *evcc.Client, assets fs.FS) *Server {
	return &Server{
		cfg:           cfg,
		store:         st,
		evcc:          evccClient,
		assets:        assets,
		secret:        []byte(cfg.SessionSecret),
		secureCookies: strings.HasPrefix(cfg.PublicBaseURL, "https://"),
	}
}

// Handler returns the complete HTTP handler: the public and admin API
// routes, plus the embedded frontend if assets was provided.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerPublicRoutes(mux)
	s.registerAdminRoutes(mux)
	if s.assets != nil {
		s.registerFrontendRoutes(mux, s.assets)
	}
	return mux
}

func (s *Server) publicPath(suffix string) string {
	return s.cfg.PublicPathPrefix + suffix
}

func (s *Server) adminPath(suffix string) string {
	return s.cfg.AdminPathPrefix + suffix
}
