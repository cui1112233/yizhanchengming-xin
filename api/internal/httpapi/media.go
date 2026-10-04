package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
)

type MediaResolver interface {
	Resolve(context.Context, string, string) (media.ResolvedAsset, error)
}

type mediaHandler struct {
	resolver MediaResolver
	owner OwnerResolver
}

func NewMediaHandler(resolver MediaResolver, owner OwnerResolver) http.Handler {
	return &mediaHandler{resolver: resolver, owner: owner}
}

func (h *mediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.resolver == nil || h.owner == nil {
		http.Error(w, "media service unavailable", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	owner, err := h.owner(r)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/media/assets/"), "/")
	if id == "" || strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}
	resolved, err := h.resolver.Resolve(r.Context(), strings.TrimSpace(owner), id)
	if errors.Is(err, media.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "media preview unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resolved)
}
