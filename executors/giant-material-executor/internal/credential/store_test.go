package credential

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileStoreRoundTripAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "executor.json")
	store := NewFileStore(path)

	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty store error=%v, want ErrNotFound", err)
	}

	want := Record{ExecutorID: "executor-1", Token: "secret-token"}
	if err := store.Save(want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.ExecutorID != want.ExecutorID || got.Token != want.Token {
		t.Fatalf("record=%+v, want %+v", got, want)
	}
	if err := store.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, err := store.Load(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after clear error=%v, want ErrNotFound", err)
	}
}

func TestFileStoreRejectsMalformedDataWithoutLeakingToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "executor.json")
	if err := os.WriteFile(path, []byte(`{"token":"secret-token"`), 0o600); err != nil {
		t.Fatalf("write malformed record: %v", err)
	}

	_, err := NewFileStore(path).Load()
	if err == nil {
		t.Fatal("expected malformed record error")
	}
	if contains := err.Error(); contains == "" || contains == "secret-token" {
		t.Fatalf("error leaks credential or is empty: %q", contains)
	}
}

func TestFileStoreRejectsBlankRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "executor.json")
	store := NewFileStore(path)
	if err := store.Save(Record{ExecutorID: "", Token: "token"}); err == nil {
		t.Fatal("expected blank executor ID to fail")
	}
	if err := store.Save(Record{ExecutorID: "executor", Token: ""}); err == nil {
		t.Fatal("expected blank token to fail")
	}
}

func TestProtectedStoreRoundTripDoesNotWritePlaintextJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "executor.protected")
	store := NewProtectedStore(path)
	want := Record{ExecutorID: "executor-1", Token: "secret-token"}

	if err := store.Save(want); err != nil {
		t.Fatalf("save protected record: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read protected record: %v", err)
	}
	if runtime.GOOS == "windows" && bytes.Contains(raw, []byte("secret-token")) {
		t.Fatalf("credential was written in plaintext: %q", raw)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load protected record: %v", err)
	}
	if got.ExecutorID != want.ExecutorID || got.Token != want.Token {
		t.Fatalf("record=%+v, want %+v", got, want)
	}
}
