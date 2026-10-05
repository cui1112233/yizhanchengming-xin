package httpapi

import (
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

func (h handler) writeServiceError(w http.ResponseWriter, r *http.Request, status int, code, message, subsystem, operation string, err error) {
	requestID := observability.RequestID(r.Context())
	args := []any{
		"request_id", requestID,
		"error_code", code,
		"subsystem", subsystem,
		"operation", operation,
	}
	if err != nil {
		args = append(args, "safe_error", observability.SafeError(err))
	}
	h.logger().Error("request operation failed", args...)
	writeJSON(w, status, map[string]any{
		"code":       code,
		"message":    message,
		"request_id": requestID,
	})
}
