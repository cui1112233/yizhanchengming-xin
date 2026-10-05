package httpapi

import (
	"net/http"
	"net/url"
	"strings"
)

// requireSameOrigin protects browser cookie-authenticated mutation endpoints.
// Safe/read-only routes are registered without this wrapper. Service-to-service
// endpoints must use their own service authentication boundary instead of
// bypassing this browser policy.
func (h handler) requireSameOrigin(next http.Handler) http.Handler {
	if h.deps.Auth == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.csrfSourceAllowed(r) {
			writeJSON(w, http.StatusForbidden, map[string]any{"code": "CSRF_REJECTED", "message": "请求来源校验失败"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h handler) csrfSourceAllowed(r *http.Request) bool {
	source := strings.TrimSpace(r.Header.Get("Origin"))
	if source == "" {
		source = strings.TrimSpace(r.Header.Get("Referer"))
	}
	if source == "" || strings.EqualFold(source, "null") {
		return false
	}
	parsed, err := url.Parse(source)
	if err != nil || parsed.Host == "" {
		return false
	}

	// Host equality is the default same-site policy. Scheme is intentionally
	// not inferred from the backend connection because production may terminate
	// HTTPS at a reverse proxy while preserving the public Host header.
	if strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	candidate := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
	for _, allowed := range h.deps.AllowedOrigins {
		value := strings.TrimSpace(strings.TrimRight(allowed, "/"))
		if strings.EqualFold(candidate, value) {
			return true
		}
	}
	return false
}
