package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Dependencies struct {
	Frontend http.Handler
}

func New(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"service": "magic-cut-4040",
		})
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"code":    "API_ROUTE_NOT_FOUND",
			"message": "接口不存在",
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusNotFound, map[string]any{"code": "API_ROUTE_NOT_FOUND"})
			return
		}
		if deps.Frontend == nil {
			http.NotFound(w, r)
			return
		}
		deps.Frontend.ServeHTTP(w, r)
	})
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
