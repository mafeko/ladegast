package httpserver

import (
	"io/fs"
	"net/http"
)

// registerFrontendRoutes serves the embedded SPA build. Static files
// (/assets/*, anything vite copied from web/public) are served as-is;
// the known page routes serve index.html so the SPA's own router can take
// over client-side. These routes deliberately live outside
// PUBLIC_PATH_PREFIX/ADMIN_PATH_PREFIX's API sub-paths - see
// DEPLOYMENT.md for what a reverse proxy must expose for this to work.
func (s *Server) registerFrontendRoutes(mux *http.ServeMux, assets fs.FS) {
	fileServer := http.FileServerFS(assets)
	mux.Handle("GET /assets/", fileServer)
	mux.Handle("GET /favicon.svg", fileServer)

	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(assets, "index.html")
		if err != nil {
			http.Error(w, "frontend not built", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}

	mux.HandleFunc("GET /{$}", serveIndex)
	mux.HandleFunc("GET "+s.publicPath("/l/{uuid}"), serveIndex)
	mux.HandleFunc("GET "+s.adminPath(""), serveIndex)
	mux.HandleFunc("GET "+s.adminPath("/view-as/{id}"), serveIndex)
}
