package webui

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type BuildInfo struct {
	GitSHA string `json:"gitSha"`
}

type handler struct {
	api  http.Handler
	ui   fs.FS
	info BuildInfo
}

// NewHandler keeps API ownership with the Go API handler while serving the
// compiled React single-page application from the same embedded binary.
func NewHandler(api http.Handler, ui fs.FS, info BuildInfo) http.Handler {
	return handler{api: api, ui: ui, info: info}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/build-info" && r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(h.info)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
		h.api.ServeHTTP(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		h.api.ServeHTTP(w, r)
		return
	}

	name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if name == "" || path.Ext(name) == "" {
		name = "index.html"
	}
	contents, err := fs.ReadFile(h.ui, name)
	if err != nil {
		if name != "index.html" {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "embedded user interface is unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("X-YCM-Static-Source", "go-embed")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(contents))
}
