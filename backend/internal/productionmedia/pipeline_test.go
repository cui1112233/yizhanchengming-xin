package productionmedia

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
)

type memoryRepository struct {
	uploads      []Object
	byRef        map[string]Object
	verified     []Object
	failed       []int64
	rehydrated   []int64
	purges       []Object
	localDeleted []int64
	purgeRetried []int64
}

func (r *memoryRepository) ClaimUpload(context.Context) (Object, bool, error) {
	if len(r.uploads) == 0 {
		return Object{}, false, nil
	}
	item := r.uploads[0]
	r.uploads = r.uploads[1:]
	return item, true, nil
}
func (r *memoryRepository) MarkVerified(_ context.Context, item Object) error {
	r.verified = append(r.verified, item)
	return nil
}
func (r *memoryRepository) MarkUploadFailed(_ context.Context, id int64, _ error) error {
	r.failed = append(r.failed, id)
	return nil
}
func (r *memoryRepository) FindVerifiedByStorageRef(_ context.Context, ref string) (Object, bool, error) {
	item, ok := r.byRef[ref]
	return item, ok, nil
}
func (r *memoryRepository) MarkRehydrated(_ context.Context, id int64, _ time.Time) error {
	r.rehydrated = append(r.rehydrated, id)
	return nil
}
func (r *memoryRepository) ClaimDuePurge(context.Context, bool) (Object, bool, error) {
	if len(r.purges) == 0 {
		return Object{}, false, nil
	}
	item := r.purges[0]
	r.purges = r.purges[1:]
	return item, true, nil
}
func (r *memoryRepository) MarkLocalDeleted(_ context.Context, id int64) error {
	r.localDeleted = append(r.localDeleted, id)
	return nil
}
func (r *memoryRepository) MarkPurgeRetry(_ context.Context, id int64, _ error) error {
	r.purgeRetried = append(r.purgeRetried, id)
	return nil
}

type memoryRemote struct {
	bodies      map[string][]byte
	metadata    map[string]RemoteMetadata
	corruptHead bool
}

func (r *memoryRemote) PutFile(_ context.Context, key, path, mediaType, sha string, size int64) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if r.bodies == nil {
		r.bodies = map[string][]byte{}
	}
	if r.metadata == nil {
		r.metadata = map[string]RemoteMetadata{}
	}
	r.bodies[key] = body
	r.metadata[key] = RemoteMetadata{ByteSize: size, SHA256: sha, MediaType: mediaType}
	return nil
}
func (r *memoryRemote) Head(_ context.Context, key string) (RemoteMetadata, error) {
	meta, ok := r.metadata[key]
	if !ok {
		return RemoteMetadata{}, os.ErrNotExist
	}
	if r.corruptHead {
		meta.SHA256 = strings.Repeat("0", 64)
	}
	return meta, nil
}
func (r *memoryRemote) Open(_ context.Context, key string) (io.ReadCloser, error) {
	body, ok := r.bodies[key]
	if !ok {
		return nil, os.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(body)), nil
}

func TestMigratorMarksVerifiedOnlyAfterRemoteSizeAndChecksumMatch(t *testing.T) {
	root := t.TempDir()
	files := localartifact.NewStore(root, 1<<20)
	artifact, err := files.SaveMedia("audio_123", "audio/mpeg", "mp3", strings.NewReader("durable-audio"))
	if err != nil {
		t.Fatal(err)
	}
	repo := &memoryRepository{uploads: []Object{{ID: 7, Owner: "alice", MediaKind: KindAudio, MediaType: artifact.MediaType, LocalStorageRef: artifact.StorageRef, ObjectKey: "production-media/audio/audio_123.mp3", RetentionDays: 7}}}
	remote := &memoryRemote{}
	now := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	summary, err := (Migrator{Repository: repo, Local: files, Remote: remote, Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Verified != 1 || len(repo.verified) != 1 || len(repo.failed) != 0 {
		t.Fatalf("summary=%+v verified=%v failed=%v", summary, repo.verified, repo.failed)
	}
	got := repo.verified[0]
	if got.ByteSize != int64(len("durable-audio")) || len(got.SHA256) != 64 || !got.RemoteVerifiedAt.Equal(now) || !got.PurgeAfter.Equal(now.Add(7*24*time.Hour)) {
		t.Fatalf("verified object=%+v", got)
	}
}

func TestMigratorKeepsLocalFileWhenRemoteVerificationDiffers(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifact, err := files.SaveMedia("video_123", "video/mp4", "mp4", strings.NewReader("video"))
	if err != nil {
		t.Fatal(err)
	}
	repo := &memoryRepository{uploads: []Object{{ID: 8, MediaKind: KindVideo, MediaType: artifact.MediaType, LocalStorageRef: artifact.StorageRef, ObjectKey: "production-media/video/video_123.mp4", RetentionDays: 7}}}
	remote := &memoryRemote{corruptHead: true}
	summary, err := (Migrator{Repository: repo, Local: files, Remote: remote}).RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Failed != 1 || len(repo.verified) != 0 || len(repo.failed) != 1 {
		t.Fatalf("summary=%+v verified=%v failed=%v", summary, repo.verified, repo.failed)
	}
	if _, err := files.Path(artifact.StorageRef); err != nil {
		t.Fatalf("local safety copy was removed: %v", err)
	}
}

func TestResolverRehydratesVerifiedRemoteObject(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	body := []byte("remote-image")
	remote := &memoryRemote{bodies: map[string][]byte{"production-media/comic/image.png": body}}
	sha := SHA256Bytes(body)
	repo := &memoryRepository{byRef: map[string]Object{"image_123.png": {ID: 9, LocalStorageRef: "image_123.png", ObjectKey: "production-media/comic/image.png", ByteSize: int64(len(body)), SHA256: sha, RetentionDays: 7, State: StateCleaned}}}
	resolver := &Resolver{Repository: repo, Local: files, Remote: remote, Now: func() time.Time { return time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC) }}
	files.SetMissingRestorer(resolver)
	file, err := files.Open("image_123.png")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	got, _ := io.ReadAll(file)
	if string(got) != string(body) || len(repo.rehydrated) != 1 {
		t.Fatalf("body=%q rehydrated=%v", got, repo.rehydrated)
	}
}

func TestResolverRefusesPendingRemoteObject(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	repo := &memoryRepository{byRef: map[string]Object{}}
	resolver := &Resolver{Repository: repo, Local: files, Remote: &memoryRemote{}}
	err := resolver.RestoreStorageRef(context.Background(), "missing.mp4")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error=%v, want os.ErrNotExist", err)
	}
}

func TestPurgerDeletesOnlyRepositoryVerifiedObject(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifact, err := files.SaveMedia("audio_due", "audio/mpeg", "mp3", strings.NewReader("audio"))
	if err != nil {
		t.Fatal(err)
	}
	repo := &memoryRepository{purges: []Object{{ID: 10, LocalStorageRef: artifact.StorageRef, State: StateVerified}}}
	summary, err := (Purger{Repository: repo, Local: files}).RunOnce(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Deleted != 1 || len(repo.localDeleted) != 1 {
		t.Fatalf("summary=%+v deleted=%v", summary, repo.localDeleted)
	}
	if _, err := files.Path(artifact.StorageRef); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file still present: %v", err)
	}
}
