// Package webui embeds the built frontend (web/, built via Vite into
// dist/ here) into the Go binary, so a single binary can serve both the
// API and the UI. The committed dist/index.html is a placeholder kept
// under version control so the module always builds, even before
// `task web:build` has produced a real frontend; running that task
// overwrites it with the real build output (gitignored, see .gitignore).
package webui

import (
	"embed"
	"io/fs"
)

//go:embed dist
var distFS embed.FS

// Assets returns the embedded frontend build output, rooted at dist/.
func Assets() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
