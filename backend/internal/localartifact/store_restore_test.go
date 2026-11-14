package localartifact

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

type testMissingRestorer struct {
	store *Store
	body  string
	calls int
}

func (r *testMissingRestorer) RestoreStorageRef(_ context.Context, storageRef string) error {
	r.calls++
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(r.body)))
	return r.store.Restore(storageRef, strings.NewReader(r.body), int64(len(r.body)), digest)
}

func TestOpenRestoresMissingVerifiedArtifactIntoLocalCache(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	restorer := &testMissingRestorer{store: store, body: "remote-media"}
	store.SetMissingRestorer(restorer)

	file, err := store.Open("clip_123.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	body, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "remote-media" || restorer.calls != 1 {
		t.Fatalf("body=%q restore calls=%d", body, restorer.calls)
	}
	if _, err := os.Stat(store.root + string(os.PathSeparator) + "clip_123.mp4"); err != nil {
		t.Fatalf("restored cache file missing: %v", err)
	}
}

func TestRestoreRejectsChecksumMismatchWithoutVisiblePartialFile(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	err := store.Restore("clip_456.mp4", strings.NewReader("corrupt"), 7, strings.Repeat("0", 64))
	if err == nil {
		t.Fatal("expected checksum mismatch")
	}
	if _, statErr := os.Stat(store.root + string(os.PathSeparator) + "clip_456.mp4"); !os.IsNotExist(statErr) {
		t.Fatalf("partial restore became visible: %v", statErr)
	}
}

func TestPathRestoresMissingVerifiedArtifactForFileBasedWorkers(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	restorer := &testMissingRestorer{store: store, body: "remote-worker-media"}
	store.SetMissingRestorer(restorer)
	path, err := store.Path("worker_123.mp4")
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil || string(body) != "remote-worker-media" || restorer.calls != 1 {
		t.Fatalf("body=%q calls=%d err=%v", body, restorer.calls, err)
	}
}

func TestSaveMediaStopsBeforeWritingWhenDiskGuardRejects(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	store.SetWriteGuard(func() error { return ErrDiskPressure })
	_, err := store.SaveMedia("audio_123", "audio/mpeg", "mp3", strings.NewReader("audio"))
	if err != ErrDiskPressure {
		t.Fatalf("SaveMedia error=%v, want ErrDiskPressure", err)
	}
}

func TestRestoreStopsBeforeWritingWhenDiskGuardRejects(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	store.SetWriteGuard(func() error { return ErrDiskPressure })
	body := "remote-media"
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(body)))
	err := store.Restore("clip_789.mp4", strings.NewReader(body), int64(len(body)), digest)
	if err != ErrDiskPressure {
		t.Fatalf("Restore error=%v, want ErrDiskPressure", err)
	}
	if _, statErr := os.Stat(store.root + string(os.PathSeparator) + "clip_789.mp4"); !os.IsNotExist(statErr) {
		t.Fatalf("restore wrote a file despite disk pressure: %v", statErr)
	}
}

type ctxAwareRestorer struct {
	store *Store
	body  string
	delay time.Duration
}

func (r *ctxAwareRestorer) RestoreStorageRef(ctx context.Context, storageRef string) error {
	if r.delay > 0 {
		select {
		case <-time.After(r.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	body := []byte(r.body)
	return r.store.Restore(storageRef, strings.NewReader(r.body), int64(len(body)), fmt.Sprintf("%x", sha256.Sum256(body)))
}

func TestOpenContextStopsWhenCallerCancels(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	store.SetMissingRestorer(&ctxAwareRestorer{store: store, body: "remote-media", delay: 5 * time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	file, err := store.OpenContext(ctx, "clip_cancel.mp4")
	elapsed := time.Since(start)
	if file != nil {
		_ = file.Close()
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("OpenContext error=%v, want context.Canceled", err)
	}
	if elapsed > time.Second {
		t.Fatalf("OpenContext waited %v after caller cancellation", elapsed)
	}
	if _, statErr := os.Stat(store.root + string(os.PathSeparator) + "clip_cancel.mp4"); !os.IsNotExist(statErr) {
		t.Fatalf("cancelled restore left a visible file: %v", statErr)
	}
}

func TestOpenAppliesRestoreTimeout(t *testing.T) {
	store := NewStore(t.TempDir(), 1<<20)
	store.SetRestoreTimeout(80 * time.Millisecond)
	store.SetMissingRestorer(&ctxAwareRestorer{store: store, body: "remote-media", delay: 5 * time.Second})
	file, err := store.Open("clip_timeout.mp4")
	if file != nil {
		_ = file.Close()
	}
	if !errors.Is(err, ErrRestoreTimeout) {
		t.Fatalf("Open error=%v, want ErrRestoreTimeout", err)
	}
	if _, statErr := os.Stat(store.root + string(os.PathSeparator) + "clip_timeout.mp4"); !os.IsNotExist(statErr) {
		t.Fatalf("timed out restore left a visible file: %v", statErr)
	}
}
