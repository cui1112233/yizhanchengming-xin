package platform

import (
	"path/filepath"
	"testing"
)

func TestWorkerCommandUsesNativeVisionWorkerOnDarwin(t *testing.T) {
	dir := t.TempDir()
	got, err := WorkerCommand("darwin", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != filepath.Join(dir, "GiantMaterialOCRWorker") {
		t.Fatalf("darwin command = %q", got)
	}
}

func TestWorkerCommandPreservesWindowsPythonWorker(t *testing.T) {
	dir := t.TempDir()
	got, err := WorkerCommand("windows", dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "python" || got[1] != filepath.Join(dir, "worker", "ocr_worker.py") {
		t.Fatalf("windows command = %q", got)
	}
}

func TestUnsupportedPlatformFailsBeforeLaunchingWorker(t *testing.T) {
	if _, err := WorkerCommand("linux", t.TempDir()); err == nil {
		t.Fatal("expected unsupported platform error")
	}
}
