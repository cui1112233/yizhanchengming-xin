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
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	candidate := strings.ToLower(parsed.Scheme + "://" + parsed.Host)
	expected := strings.ToLower(requestScheme(r) + "://" + r.Host)
	if candidate == expected {
		return true
	}
	for _, allowed := range h.deps.AllowedOrigins {
		value := strings.ToLower(strings.TrimSpace(strings.TrimRight(allowed, "/")))
		if candidate == value {
			return true
		}
	}
	return false
}

func requestScheme(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		if index := strings.IndexByte(forwarded, ','); index >= 0 {
			forwarded = forwarded[:index]
		}
		if forwarded = strings.ToLower(strings.TrimSpace(forwarded)); forwarded == "http" || forwarded == "https" {
			return forwarded
		}
	}
	if r.TLS != nil {
		return "https"
	}
	if r.URL != nil {
		if scheme := strings.ToLower(strings.TrimSpace(r.URL.Scheme)); scheme == "http" || scheme == "https" {
			return scheme
		}
	}
	return "http"
}
