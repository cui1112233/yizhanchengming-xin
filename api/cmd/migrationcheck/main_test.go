package main

import (
	"strings"
	"testing"
)

func TestClassifyWorkspacePreferences(t *testing.T) {
	cases := []struct {
		name    string
		applied bool
		exists  bool
		fields  []string
		want    string
	}{
		{"missing migration", false, false, nil, "A"},
		{"recorded missing table", true, false, nil, "B"},
		{"recorded missing field", true, true, []string{"user_id", "theme"}, "B"},
		{"consistent", true, true, []string{"user_id", "theme", "notifications_enabled", "storage_preference", "updated_at"}, "C"},
	}
	for _, tc := range cases {
		if got := classifyWorkspacePreferences(tc.applied, tc.exists, tc.fields); got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
}

func TestSafeDiagnosticErrorNeverLeaksConnectionSecrets(t *testing.T) {
	value := safeDiagnosticError("dial tcp db.internal:3306 user=alice password=s3cr3t token=abc key=xyz")
	for _, forbidden := range []string{"db.internal", "3306", "alice", "s3cr3t", "token", "key"} {
		if strings.Contains(strings.ToLower(value), strings.ToLower(forbidden)) {
			t.Fatalf("leaked %q in %q", forbidden, value)
		}
	}
}
