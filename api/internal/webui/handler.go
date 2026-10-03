package webui

import (
	"errors"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

func newSPAHandler(root fs.FS) (http.Handler, error) {
	if root == nil {
		return nil, errors.New("web ui filesystem is required")
	}
	index, err := fs.ReadFile(root, "index.html")
	if err != nil {
		return nil, errors.New("web ui index.html is required")
	}
	files := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requestPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if requestPath == "." || requestPath == "" {
			serveIndex(w, r, index)
			return
		}

		if info, statErr := fs.Stat(root, requestPath); statErr == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}

		// Asset requests must remain real 404s. Returning index.html for a missing
		// JS/CSS file causes browsers to fail later with a misleading MIME error.
		if strings.HasPrefix(requestPath, "assets/") || path.Ext(requestPath) != "" {
			http.NotFound(w, r)
			return
		}
		serveIndex(w, r, index)
	}), nil
}

func serveIndex(w http.ResponseWriter, r *http.Request, index []byte) {
	contentType := mime.TypeByExtension(".html")
	if contentType == "" {
		contentType = "text/html; charset=utf-8"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(index)
	}
}
