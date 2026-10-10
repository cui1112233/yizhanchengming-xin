package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// StaticHandler serves the built React directory. The filesystem is injected
// so tests can use a temporary in-memory tree and production can use embed.FS.
func StaticHandler(files fs.FS) http.Handler {
	if files == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend assets are unavailable", http.StatusServiceUnavailable)
		})
	}
	fileServer := http.FileServer(http.FS(files))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		name := strings.TrimPrefix(path.Clean(req.URL.Path), "/")
		if name != "" {
			if _, err := fs.Stat(files, name); err != nil && path.Ext(name) == "" {
				index, readErr := fs.ReadFile(files, "index.html")
				if readErr != nil {
					http.Error(w, "frontend index is unavailable", http.StatusServiceUnavailable)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write(index)
				return
			}
		}
		fileServer.ServeHTTP(w, req)
	})
}
