package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"
)

const (
	RequestIDHeader    = "X-Request-ID"
	MaxRequestIDLength = 128
)

type contextKey string

const requestIDKey contextKey = "request_id"

var safeRequestID = regexp.MustCompile(`^[A-Za-z0-9._:-]+$`)
var sensitiveAssignment = regexp.MustCompile(`(?i)\b(authorization|cookie|set-cookie|password|passwd|token|access[_-]?token|refresh[_-]?token|api[_-]?key|apikey|secret|credential|client[_-]?secret|ciphertext|nonce|dsn)\s*[:=]\s*([^&\s,;]+)`)
var bearerValue = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]+`)
var urlUserInfo = regexp.MustCompile(`://[^/@\s]+@`)
var mysqlLikeCredential = regexp.MustCompile(`\b[^:\s]+:[^@\s]+@tcp\(`)
var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

func ValidRequestID(values []string) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	value := strings.TrimSpace(values[0])
	if value == "" || len(value) > MaxRequestIDLength || !safeRequestID.MatchString(value) {
		return "", false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", false
		}
	}
	return value, true
}

func NewRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return "req_" + hex.EncodeToString(buf)
	}
	return "req_" + strings.ReplaceAll(time.Now().UTC().Format("20060102T150405.000000000"), ".", "")
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return WithUserCorrelation(context.WithValue(ctx, requestIDKey, requestID))
}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func IsSensitiveKey(key string) bool {
	normalized := nonAlphaNumeric.ReplaceAllString(strings.ToLower(strings.TrimSpace(key)), "")
	for _, marker := range []string{
		"authorization", "cookie", "setcookie", "password", "passwd", "token",
		"accesstoken", "refreshtoken", "apikey", "secret", "credential",
		"clientsecret", "ciphertext", "nonce", "dsn",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func SanitizeString(value string) string {
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
	value = bearerValue.ReplaceAllString(value, "Bearer [REDACTED]")
	value = sensitiveAssignment.ReplaceAllString(value, "$1=[REDACTED]")
	value = urlUserInfo.ReplaceAllString(value, "://[REDACTED]@")
	value = mysqlLikeCredential.ReplaceAllString(value, "[REDACTED]@tcp(")
	if len(value) > 1024 {
		value = value[:1024] + "…"
	}
	return value
}

func SafeError(err error) string {
	if err == nil {
		return ""
	}
	return SanitizeString(err.Error())
}

func Redact(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		return SanitizeString(v)
	case error:
		return SafeError(v)
	case http.Header:
		out := make(map[string]any, len(v))
		for key, values := range v {
			if IsSensitiveKey(key) {
				out[key] = "[REDACTED]"
				continue
			}
			clean := make([]string, len(values))
			for i, item := range values {
				clean[i] = SanitizeString(item)
			}
			out[key] = clean
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if IsSensitiveKey(key) {
				out[key] = "[REDACTED]"
			} else {
				out[key] = Redact(item)
			}
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(v))
		for key, item := range v {
			if IsSensitiveKey(key) {
				out[key] = "[REDACTED]"
			} else {
				out[key] = SanitizeString(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = Redact(item)
		}
		return out
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = SanitizeString(item)
		}
		return out
	default:
		return v
	}
}

func NewJSONLogger(w io.Writer) *slog.Logger {
	if w == nil {
		w = io.Discard
	}
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
		if IsSensitiveKey(attr.Key) {
			return slog.String(attr.Key, "[REDACTED]")
		}
		switch attr.Value.Kind() {
		case slog.KindString:
			return slog.String(attr.Key, SanitizeString(attr.Value.String()))
		case slog.KindAny:
			return slog.Any(attr.Key, Redact(attr.Value.Any()))
		default:
			return attr
		}
	}})
	return slog.New(handler)
}

func ErrorAttrs(code, subsystem, operation string, err error) []any {
	return []any{
		"error_code", code,
		"subsystem", subsystem,
		"operation", operation,
		"safe_error", SafeError(err),
	}
}

type ProcessStats struct {
	Goroutines int    `json:"goroutines"`
	HeapAlloc  uint64 `json:"heapAlloc"`
	HeapInuse  uint64 `json:"heapInuse"`
	Sys        uint64 `json:"sys"`
	UptimeSec  int64  `json:"uptimeSec"`
}

func ReadProcessStats(startedAt time.Time, now time.Time) ProcessStats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	if startedAt.IsZero() {
		startedAt = now
	}
	return ProcessStats{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  mem.HeapAlloc,
		HeapInuse:  mem.HeapInuse,
		Sys:        mem.Sys,
		UptimeSec:  int64(now.Sub(startedAt).Seconds()),
	}
}

func IsStuck(status string, updatedAt, now time.Time, threshold time.Duration) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	if threshold <= 0 || updatedAt.IsZero() {
		return false
	}
	if status != "running" && status != "queued" && status != "pending" {
		return false
	}
	return now.Sub(updatedAt) > threshold
}
