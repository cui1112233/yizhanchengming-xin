package main

import (
	"os"
	"strings"
	"testing"
)

func TestMainWiresWindowsExecutorUI(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	source := string(data)
	for _, want := range []string{"internal/ui", "ui.Config", "ui.Run", "PublicURL:", "SetPublicURL:", "Snapshot: snapshot"} {
		if !strings.Contains(source, want) {
			t.Fatalf("main.go missing GUI integration marker %q", want)
		}
	}
}
