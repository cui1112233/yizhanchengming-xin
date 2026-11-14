package productionmedia

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

	"qiantie/backend/internal/localartifact"
)

type countingRepository struct {
	memoryRepository
	failMark  bool
	markCalls int
}

func (r *countingRepository) MarkRehydrated(ctx context.Context, id int64, at time.Time) error {
	r.markCalls++
	if r.failMark {
		return errors.New("db unavailable")
	}
	return r.memoryRepository.MarkRehydrated(ctx, id, at)
}

type countingRemote struct {
	memoryRemote
	openCalls int
}

func (r *countingRemote) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	r.openCalls++
	return r.memoryRemote.Open(ctx, key)
}

func newResolverFixture(t *testing.T, failMark bool) (*Resolver, *countingRepository, *countingRemote, string, []byte) {
	t.Helper()
	body := []byte("stage3-media")
	ref := "stage3_audio.mp3"
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	repo := &countingRepository{
		memoryRepository: memoryRepository{byRef: map[string]Object{
			ref: {
				ID: 11, Owner: "alice", MediaKind: KindAudio, LocalStorageRef: ref,
				ObjectKey: "production-media/audio/" + ref, ByteSize: int64(len(body)),
				SHA256: digest, State: StateCleaned, RetentionDays: 7,
			},
		}},
		failMark: failMark,
	}
	remote := &countingRemote{memoryRemote: memoryRemote{bodies: map[string][]byte{"production-media/audio/" + ref: body}}}
	return &Resolver{Repository: repo, Local: files, Remote: remote}, repo, remote, ref, body
}

func TestResolverMarkFailureKeepsReadWorking(t *testing.T) {
	resolver, repo, remote, ref, _ := newResolverFixture(t, true)
	if err := resolver.RestoreStorageRef(context.Background(), ref); err != nil {
		t.Fatalf("restore should succeed when local file is ready: %v", err)
	}
	path, err := resolver.Local.LocalPath(ref)
	if err != nil {
		t.Fatalf("local path: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("local file missing: %v", err)
	}
	if remote.openCalls != 1 {
		t.Fatalf("first restore remote opens=%d, want 1", remote.openCalls)
	}
	if err := resolver.RestoreStorageRef(context.Background(), ref); err != nil {
		t.Fatalf("second restore should keep working: %v", err)
	}
	if remote.openCalls != 1 {
		t.Fatalf("second restore re-downloaded: remote opens=%d, want 1", remote.openCalls)
	}
	if repo.markCalls != 2 {
		t.Fatalf("mark retries=%d, want 2", repo.markCalls)
	}
}

func TestResolverSkipsDownloadWhenLocalCopyMatchesSize(t *testing.T) {
	resolver, repo, remote, ref, body := newResolverFixture(t, false)
	digest := fmt.Sprintf("%x", sha256.Sum256(body))
	if err := resolver.Local.Restore(ref, strings.NewReader(string(body)), int64(len(body)), digest); err != nil {
		t.Fatal(err)
	}
	if err := resolver.RestoreStorageRef(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	if remote.openCalls != 0 {
		t.Fatalf("remote opens=%d, want 0 for an existing local copy", remote.openCalls)
	}
	if repo.markCalls != 1 {
		t.Fatalf("mark calls=%d, want 1", repo.markCalls)
	}
}

func TestResolverRedownloadsWhenLocalCopyMissing(t *testing.T) {
	resolver, _, remote, ref, _ := newResolverFixture(t, false)
	if err := resolver.RestoreStorageRef(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
	if remote.openCalls != 1 {
		t.Fatalf("remote opens=%d, want 1", remote.openCalls)
	}
}

func TestResolverRejectsUnverifiedState(t *testing.T) {
	resolver, _, remote, ref, _ := newResolverFixture(t, false)
	item := resolver.Repository.(*countingRepository).byRef[ref]
	item.State = StateUploading
	resolver.Repository.(*countingRepository).byRef[ref] = item
	if err := resolver.RestoreStorageRef(context.Background(), ref); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error=%v, want os.ErrNotExist", err)
	}
	if remote.openCalls != 0 {
		t.Fatalf("remote opens=%d, want 0", remote.openCalls)
	}
}
