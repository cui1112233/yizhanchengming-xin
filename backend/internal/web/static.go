package web

import (
	"io/fs"
	"net/http"
)

// StaticHandler serves the built React directory. The filesystem is injected
// so tests can use a temporary in-memory tree and production can use embed.FS.
func StaticHandler(files fs.FS) http.Handler {
	if files == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend assets are unavailable", http.StatusServiceUnavailable)
		})
	}
	return http.FileServer(http.FS(files))
}
