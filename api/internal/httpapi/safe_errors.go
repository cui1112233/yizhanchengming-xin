package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

type apiRouteErrorWriter struct{ http.ResponseWriter }

func (w *apiRouteErrorWriter) WriteHeader(status int) {
	writeJSON(w.ResponseWriter, status, map[string]string{"message": http.StatusText(status)})
}

func (w *apiRouteErrorWriter) Write(body []byte) (int, error) { return len(body), nil }

// errorEnvelope augments existing error DTOs without removing fields consumed
// by older clients. Callers own safe public copy; never pass raw internal errors.
// Correlation is owned exclusively by withObservability and its response header.
func errorEnvelope(w http.ResponseWriter, status int, value any) any {
	fields := map[string]json.RawMessage{}
	encoded, err := json.Marshal(value)
	if err == nil {
		_ = json.Unmarshal(encoded, &fields)
	}
	if fields == nil {
		fields = map[string]json.RawMessage{}
	}
	put := func(key, value string) { fields[key], _ = json.Marshal(value) }
	if len(fields["code"]) == 0 {
		code := "REQUEST_FAILED"
		switch status {
		case 400:
			code = "INVALID_REQUEST"
		case 401:
			code = "AUTH_UNAUTHENTICATED"
		case 403:
			code = "AUTH_FORBIDDEN"
		case 404:
			code = "NOT_FOUND"
		case 405:
			code = "METHOD_NOT_ALLOWED"
		case 409:
			code = "CONFLICT"
		case 422:
			code = "INVALID_REQUEST"
		case 500:
			code = "INTERNAL_ERROR"
		case 503:
			code = "SERVICE_UNAVAILABLE"
		}
		put("code", code)
	}
	if len(fields["message"]) == 0 {
		var message string
		// Only the reviewed legacy map writers use error as public copy.
		// Domain result DTOs may carry diagnostic error metadata instead.
		switch value.(type) {
		case map[string]string, map[string]any:
			_ = json.Unmarshal(fields["error"], &message)
		}
		if message == "" {
			message = http.StatusText(status)
		}
		put("message", message)
	}
	put("request_id", w.Header().Get(observability.RequestIDHeader))
	return fields
}

func requestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return observability.RequestID(r.Context())
}

func (h handler) writeServiceError(w http.ResponseWriter, r *http.Request, status int, code, message, subsystem, operation string, err error) {
	requestID := requestIDFromRequest(r)
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
		"error":      message,
		"message":    message,
		"request_id": requestID,
	})
}
