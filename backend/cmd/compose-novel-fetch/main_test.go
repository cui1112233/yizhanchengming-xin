package main

import (
	"strings"
	"testing"
)

func TestInjectDeduplicatesAndAddsCanonicalScripts(t *testing.T) {
	source := `<html><body><script src="/batch-rewrite/app.js?v=old"></script><script src="/batch-rewrite/app.js?v=old"></script><script src="./interaction-feedback.js"></script></body></html>`
	got := inject(source)
	if strings.Count(got, `src="/batch-rewrite/app.js?v=20260922-retry-progress-r1"`) != 1 {
		t.Fatalf("app script was not normalized once: %s", got)
	}
	if strings.Count(got, feedbackPath) != 1 {
		t.Fatalf("feedback script was not deduplicated: %s", got)
	}
	for _, path := range []string{loginPath, taskPath, runtimePath, selectedPath, clientPath, datePath, configPath, layoutPath, quickPath, guardPath, controlsPath, versionPath} {
		if !strings.Contains(got, path) {
			t.Fatalf("missing composed script %s", path)
		}
	}
}

func TestInjectLeavesDocumentsWithoutBodyUntouched(t *testing.T) {
	const source = `<html><p>plain</p></html>`
	if got := inject(source); got != source {
		t.Fatalf("source without body changed: %s", got)
	}
}
