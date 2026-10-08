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
	api     http.Handler
	ui      fs.FS
	adminUI fs.FS
	info    BuildInfo
}

// NewHandler keeps API ownership with the Go API handler while serving the
// compiled React single-page application from the same embedded binary.
func NewHandler(api http.Handler, ui fs.FS, info BuildInfo) http.Handler {
	return NewHandlerWithAdmin(api, ui, nil, info)
}

// NewHandlerWithAdmin serves the user and administrator React applications
// from separate embedded file systems. API routes remain owned by the Go
// handler, so an /api/v1/admin request can never be mistaken for an SPA
// fallback.
func NewHandlerWithAdmin(api http.Handler, ui, adminUI fs.FS, info BuildInfo) http.Handler {
	return handler{api: api, ui: ui, adminUI: adminUI, info: info}
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

	if r.URL.Path == "/admin" || strings.HasPrefix(r.URL.Path, "/admin/") {
		h.serveStatic(w, r, h.adminUI, "/admin", "embedded admin interface is unavailable")
		return
	}
	h.serveStatic(w, r, h.ui, "", "embedded user interface is unavailable")
}

func (h handler) serveStatic(w http.ResponseWriter, r *http.Request, ui fs.FS, prefix, unavailableMessage string) {
	if ui == nil {
		http.Error(w, unavailableMessage, http.StatusServiceUnavailable)
		return
	}
	relativePath := r.URL.Path
	if prefix != "" {
		relativePath = strings.TrimPrefix(relativePath, prefix)
	}
	name := strings.TrimPrefix(path.Clean("/"+relativePath), "/")
	if name == "" || path.Ext(name) == "" {
		name = "index.html"
	}
	contents, err := fs.ReadFile(ui, name)
	if err != nil {
		if name != "index.html" {
			http.NotFound(w, r)
			return
		}
		http.Error(w, unavailableMessage, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("X-YCM-Static-Source", "go-embed")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(contents))
}
