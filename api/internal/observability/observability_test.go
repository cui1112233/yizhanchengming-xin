package observability

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestJSONLoggerQuotedSecretsAndCompleteCookieValues(t *testing.T) {
	for _, input := range []string{
		`upstream failed {"provider_api_key":"provider-canary","password":"pass-canary","reason":"diagnostic-visible"}`,
		`upstream failed {"password":"pass-canary with spaces, semicolons; and \"quotes\"","reason":"diagnostic-visible"}`,
		`upstream failed {"provider_api_key":{"key":"provider-canary","extra":"pass-canary"},"reason":"diagnostic-visible"}`,
		"upstream failed\nCookie: locale=zh; session=cookie-canary; refresh=refresh-canary\ndiagnostic-visible",
		"upstream failed\nCookie: [REDACTED]; session=cookie-canary\ndiagnostic-visible",
		"upstream failed\nSet-Cookie: locale=zh; session=cookie-canary; Expires=Wed, 21 Oct 2026 07:28:00 GMT; refresh=refresh-canary\ndiagnostic-visible",
		`upstream failed {"Set-Cookie":"locale=zh; session=cookie-canary; refresh=refresh-canary","reason":"diagnostic-visible"}`,
	} {
		t.Run(input, func(t *testing.T) {
			var buf bytes.Buffer
			NewJSONLogger(&buf).Error("request failed", "cause", errors.New(input), "detail", input, "safe", "unchanged")
			for _, secret := range []string{"provider-canary", "pass-canary", "cookie-canary", "refresh-canary"} {
				if strings.Contains(buf.String(), secret) {
					t.Errorf("logger leaked %s", secret)
				}
			}
			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"cause", "detail"} {
				value, _ := entry[key].(string)
				if !strings.Contains(value, "upstream failed") || !strings.Contains(value, "diagnostic-visible") {
					t.Errorf("diagnostic text lost: %s", value)
				}
			}
			if entry["safe"] != "unchanged" {
				t.Fatal("safe attribute changed")
			}
		})
	}
}

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
