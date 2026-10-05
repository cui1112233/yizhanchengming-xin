package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// requireVideoExecutorBootstrap protects the one-time executor registration
// endpoint with a service bootstrap credential. It is intentionally separate
// from browser session authentication and from the per-executor Bearer token
// issued after registration.
func (h handler) requireVideoExecutorBootstrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := strings.TrimSpace(h.deps.VideoExecutorBootstrapToken)
		if expected == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"code":    "VIDEO_EXECUTOR_BOOTSTRAP_UNAVAILABLE",
				"message": "local executor registration is not configured",
			})
			return
		}
		provided := executorBearerToken(r)
		if len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"code":    "VIDEO_EXECUTOR_UNAUTHENTICATED",
				"message": "local executor bootstrap credential is invalid",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
