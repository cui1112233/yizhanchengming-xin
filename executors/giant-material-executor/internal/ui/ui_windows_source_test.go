package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWindowsImplementationLaunchesBrowserSetupPage(t *testing.T) {
	path := filepath.Join("ui_windows.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Windows UI source: %v", err)
	}
	source := string(data)
	for _, want := range []string{"127.0.0.1:17861/setup", "url.dll", "FileProtocolHandler", "exec.Command"} {
		if !strings.Contains(source, want) {
			t.Fatalf("Windows UI source missing %q", want)
		}
	}
}
