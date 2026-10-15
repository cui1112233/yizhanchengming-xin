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
		if name == "novel-panel/workbench" || strings.HasPrefix(name, "novel-panel/workbench/") {
			http.NotFound(w, req)
			return
		}
		// Keep the recovered V78 workbench available under its explicit legacy
		// route. The public /novel-fetch entry is the React workbench so it keeps
		// the shared application shell and sidebar.
		if name == "batch-rewrite/v2" {
			serveEmbeddedIndex(w, files, "batch-rewrite/index.html")
			return
		}
		if name == "novel-fetch-react" {
			serveEmbeddedIndex(w, files, "index.html")
			return
		}
		// The admin app is built as a separate entry so it can keep its own
		// navigation and permission gate. Any /admin/* URL must therefore use
		// admin.html instead of falling back to the user-facing index.html.
		if name == "admin" || strings.HasPrefix(name, "admin/") {
			serveEmbeddedIndex(w, files, "admin.html")
			return
		}
		if name != "" {
			if info, err := fs.Stat(files, name); err == nil && info.IsDir() {
				indexName := path.Join(name, "index.html")
				if _, indexErr := fs.Stat(files, indexName); indexErr == nil {
					serveEmbeddedIndex(w, files, indexName)
					return
				}
				// A directory used only to hold a nested page must not turn the
				// SPA route itself into a directory listing.
				name = ""
			} else if err != nil && path.Ext(name) == "" {
				indexName := path.Join(name, "index.html")
				if _, indexErr := fs.Stat(files, indexName); indexErr == nil {
					serveEmbeddedIndex(w, files, indexName)
					return
				}
				index, readErr := fs.ReadFile(files, "index.html")
				if readErr != nil {
					http.Error(w, "frontend index is unavailable", http.StatusServiceUnavailable)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write(index)
				return
			}
			if name == "" {
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

func serveEmbeddedIndex(w http.ResponseWriter, files fs.FS, name string) {
	index, err := fs.ReadFile(files, name)
	if err != nil {
		http.Error(w, "frontend page is unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(index)
}
