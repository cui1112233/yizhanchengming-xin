package observability

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRedactHeadersMapsAndNestedSecrets(t *testing.T) {
	headers := http.Header{
		"Authorization": []string{"Bearer abc.def.secret"},
		"Cookie":        []string{"ycm_access=secret"},
		"Set-Cookie":    []string{"ycm_refresh=secret"},
		"X-Safe":        []string{"ok"},
	}
	redactedHeaders := Redact(headers).(map[string]any)
	for _, key := range []string{"Authorization", "Cookie", "Set-Cookie"} {
		if redactedHeaders[key] != "[REDACTED]" {
			t.Fatalf("header %s was not redacted: %#v", key, redactedHeaders[key])
		}
	}

	input := map[string]any{
		"provider": "personal_api",
		"nested": map[string]any{
			"clientSecret": "nested-secret",
			"safe":         "visible",
		},
		"API_KEY": "api-secret",
	}
	got := Redact(input).(map[string]any)
	nested := got["nested"].(map[string]any)
	if nested["clientSecret"] != "[REDACTED]" || got["API_KEY"] != "[REDACTED]" {
		t.Fatalf("nested/map secret redaction failed: %#v", got)
	}
	if nested["safe"] != "visible" {
		t.Fatalf("safe nested field changed: %#v", nested)
	}
}

func TestSanitizeQueryLikeAndErrorStrings(t *testing.T) {
	input := "upstream failed?token=abc123&safe=ok authorization: Bearer topsecret password=hunter2"
	got := SanitizeString(input)
	for _, secret := range []string{"abc123", "topsecret", "hunter2"} {
		if strings.Contains(got, secret) {
			t.Fatalf("sanitized string leaked %q: %s", secret, got)
		}
	}

	err := errors.New("dial mysql app:dbpass@tcp(127.0.0.1:3306) dsn=user:anotherpass@tcp(localhost:3306)")
	safe := SafeError(err)
	if strings.Contains(safe, "dbpass") || strings.Contains(safe, "anotherpass") {
		t.Fatalf("safe error leaked DSN credential: %s", safe)
	}
}

func TestJSONLoggerRedactsSensitiveAttributesAndControls(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(&buf)
	logger.Info("request", "authorization", "Bearer secret", "query", "api_key=supersecret\nforged=true", "safe", "ok")
	out := buf.String()
	if strings.Contains(out, "secret") || strings.Contains(out, "supersecret") || strings.Contains(out, "\nforged=true") {
		t.Fatalf("logger leaked sensitive/control content: %s", out)
	}
	if !strings.Contains(out, `"safe":"ok"`) {
		t.Fatalf("safe structured field missing: %s", out)
	}
}

func TestRequestIDValidation(t *testing.T) {
	if got, ok := ValidRequestID([]string{"abc_123:run.4"}); !ok || got != "abc_123:run.4" {
		t.Fatalf("safe request id rejected: %q %v", got, ok)
	}
	for _, values := range [][]string{
		{"bad/value"},
		{"bad\nvalue"},
		{strings.Repeat("a", MaxRequestIDLength+1)},
		{"one", "two"},
	} {
		if got, ok := ValidRequestID(values); ok || got != "" {
			t.Fatalf("unsafe request id accepted: %#v -> %q", values, got)
		}
	}
}

func TestStuckDetectorReportsOnlyWithoutMutation(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	if !IsStuck("running", now.Add(-31*time.Minute), now, 30*time.Minute) {
		t.Fatal("long-running job was not reported stuck")
	}
	if IsStuck("succeeded", now.Add(-24*time.Hour), now, 30*time.Minute) {
		t.Fatal("terminal job must not be reported stuck")
	}
	if IsStuck("running", now.Add(-5*time.Minute), now, 30*time.Minute) {
		t.Fatal("fresh running job incorrectly reported stuck")
	}
}
