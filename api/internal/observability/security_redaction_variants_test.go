package observability

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSecurityRedactionBypassVariants(t *testing.T) {
	input := map[string]any{
		"Secret":       "synthetic-secret-a",
		"SECRET":       "synthetic-secret-b",
		"clientSecret": "synthetic-secret-c",
		"client_secret": "synthetic-secret-d",
		"accessToken":  "synthetic-token-a",
		"refreshToken": "synthetic-token-b",
		"nested": map[string]any{
			"credential": "synthetic-credential",
		},
	}
	got := Redact(input)
	var buf bytes.Buffer
	NewJSONLogger(&buf).Info("redaction variants", "payload", got)
	out := buf.String()
	for _, forbidden := range []string{
		"synthetic-secret-a", "synthetic-secret-b", "synthetic-secret-c", "synthetic-secret-d",
		"synthetic-token-a", "synthetic-token-b", "synthetic-credential",
	} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("redaction bypass leaked %q: %s", forbidden, out)
		}
	}
}

func TestSecuritySanitizesURLBearerQueryAndErrorCredentialForms(t *testing.T) {
	cases := []string{
		"https://user:synthetic-basic-pass@example.invalid/path",
		"Authorization: Bearer synthetic-bearer-token",
		"https://example.invalid/callback?clientSecret=synthetic-client-secret&safe=1",
		"accessToken=synthetic-access-token refreshToken=synthetic-refresh-token",
	}
	for _, raw := range cases {
		got := SanitizeString(raw)
		for _, forbidden := range []string{
			"synthetic-basic-pass", "synthetic-bearer-token", "synthetic-client-secret",
			"synthetic-access-token", "synthetic-refresh-token",
		} {
			if strings.Contains(got, forbidden) {
				t.Fatalf("SanitizeString leaked %q from %q: %s", forbidden, raw, got)
			}
		}
	}

	got := SafeError(errors.New("dsn=user:synthetic-dsn-pass@tcp(db.internal:3306) password=synthetic-password"))
	for _, forbidden := range []string{"synthetic-dsn-pass", "synthetic-password"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("SafeError leaked %q: %s", forbidden, got)
		}
	}
}

func TestSecurityRequestIDRejectsCRLFControlOverlongAndMultipleValues(t *testing.T) {
	bad := [][]string{
		{"safe\r\nforged"},
		{"safe\tforged"},
		{strings.Repeat("a", MaxRequestIDLength+1)},
		{"first", "second"},
	}
	for _, values := range bad {
		if got, ok := ValidRequestID(values); ok || got != "" {
			t.Fatalf("ValidRequestID(%q) accepted unsafe value %q", values, got)
		}
	}
}
