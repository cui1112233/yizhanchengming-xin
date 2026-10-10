package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/mergeworker"
)

type mergeOutputFunc func(context.Context, string, string) (string, error)

func (fn mergeOutputFunc) PutMerged(ctx context.Context, taskID, path string) (string, error) {
	return fn(ctx, taskID, path)
}

var _ mergeworker.ObjectStore = mergeOutputFunc(nil)

type productionHTTPAdapter struct{ calls int }

func (a *productionHTTPAdapter) Submit(context.Context, batchfactoryv11.FrozenVideoModel, batchfactoryv11.FinalPrompt) (batchfactoryv11.ProviderTaskRef, error) {
	a.calls++
	return batchfactoryv11.ProviderTaskRef{ProviderTaskID: "provider-http-1", State: batchfactoryv11.ProductionSucceeded, MediaURL: "https://media.example/generated.mp4"}, nil
}

type mergeHTTPAdapter struct{ calls int }

func (a *mergeHTTPAdapter) Submit(_ context.Context, _ string, sources []batchfactoryv11.MergeMedia, _ batchfactoryv11.MergeOptions) (batchfactoryv11.MergeJob, error) {
	a.calls++
	if len(sources) != 1 || sources[0].URL == "" {
		return batchfactoryv11.MergeJob{}, batchfactoryv11.ErrInvalid
	}
	return batchfactoryv11.MergeJob{Status: batchfactoryv11.MergeSucceeded, OutputURL: "https://media.example/merged.mp4"}, nil
}

func TestSliceFourProductionRoutesSubmitAndReadDurableStatus(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	if err != nil {
		t.Fatal(err)
	}
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
	if _, err := director.RunDirector(context.Background(), "alice", batch.ID, batch.Books[0].ID); err != nil {
		t.Fatal(err)
	}
	adapter := &productionHTTPAdapter{}
	production := &batchfactoryv11.ProductionService{Store: store, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: batchfactoryv11.FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 4, Store: store, Director: director, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Production: production})
	base := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/production", map[string]any{"requestId": "submit-http-1"})
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), "provider-http-1") || adapter.calls != 1 {
		t.Fatalf("submit status=%d body=%s calls=%d", rec.Code, rec.Body.String(), adapter.calls)
	}
	status := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v11/batches/"+batch.ID+"/status", nil)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), "succeeded") {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
	batchRun := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/production", map[string]any{"requestId": "submit-batch-1"})
	if batchRun.Code != http.StatusCreated || !strings.Contains(batchRun.Body.String(), batch.ID) {
		t.Fatalf("batch status=%d body=%s", batchRun.Code, batchRun.Body.String())
	}
}

func TestSliceThreeDoesNotExposeProductionMutation(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 3, Store: store})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v11/batches/x/production", map[string]any{"requestId": "nope"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestEarlySliceExposesReadOnlyV12MergeStatus(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "merge-read-1", Status: batchfactoryv11.MergeRunning, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/merge-status"
	owner := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if owner.Code != http.StatusOK || !strings.Contains(owner.Body.String(), "merge-read-1") {
		t.Fatalf("owner status=%d body=%s", owner.Code, owner.Body.String())
	}
	other := signedJSONRequest(t, api, now, "bob", http.MethodGet, path, nil)
	if other.Code != http.StatusNotFound {
		t.Fatalf("cross-owner status=%d body=%s", other.Code, other.Body.String())
	}
}

func TestSliceFiveMergeRouteIsIdempotentAndOwnerScoped(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "林晚推门进入客厅。"}}})
	if err != nil {
		t.Fatal(err)
	}
	director := &batchfactoryv11.DirectorService{Store: store, Provider: &directorHTTPProvider{output: directorHTTPJSON}}
	if _, err := director.RunDirector(context.Background(), "alice", batch.ID, batch.Books[0].ID); err != nil {
		t.Fatal(err)
	}
	production := &batchfactoryv11.ProductionService{Store: store, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Adapter: &productionHTTPAdapter{}, Enabled: true, Model: batchfactoryv11.FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	if _, err := production.SubmitBookProduction(context.Background(), "alice", batch.ID, batch.Books[0].ID, "production-for-merge"); err != nil {
		t.Fatal(err)
	}
	mergeAdapter := &mergeHTTPAdapter{}
	merge := &batchfactoryv11.MergeService{Store: store, Adapter: mergeAdapter, Enabled: true}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Director: director, Compiler: &batchfactoryv11.PromptCompilerService{Store: store}, Production: production, Merge: merge})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/merge"
	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"requestId": "merge-http-1", "timingMode": "speed", "speed": 1.5})
	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{"requestId": "merge-http-1", "timingMode": "speed", "speed": 1.5})
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated || !strings.Contains(first.Body.String(), "merged.mp4") || first.Body.String() != second.Body.String() || mergeAdapter.calls != 1 {
		t.Fatalf("first=%d %s second=%d %s calls=%d", first.Code, first.Body.String(), second.Code, second.Body.String(), mergeAdapter.calls)
	}
	bookPath := "/api/batch-factory/v11/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/merge"
	bookMerge := signedJSONRequest(t, api, now, "alice", http.MethodPost, bookPath, map[string]any{"requestId": "book-merge-http-1", "timingMode": "speed", "speed": 1})
	if bookMerge.Code != http.StatusCreated || !strings.Contains(bookMerge.Body.String(), `"bookId":"`+batch.Books[0].ID+`"`) || mergeAdapter.calls != 2 {
		t.Fatalf("book merge=%d %s calls=%d", bookMerge.Code, bookMerge.Body.String(), mergeAdapter.calls)
	}
	status := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v11/batches/"+batch.ID+"/merge-status", nil)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), "merged.mp4") {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
	other := signedJSONRequest(t, api, now, "bob", http.MethodGet, "/api/batch-factory/v11/batches/"+batch.ID+"/merge-status", nil)
	if other.Code != http.StatusNotFound {
		t.Fatalf("cross-owner status=%d body=%s", other.Code, other.Body.String())
	}
}

func TestSliceFiveServesOnlyTheOwnersLocalMergedArtifact(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifactID := "merge_local_media_1"
	if _, err := files.SaveMP4(artifactID, strings.NewReader("0000ftypisom-local-merged-video")); err != nil {
		t.Fatal(err)
	}
	output := "/api/batch-factory/v11/batches/" + batch.ID + "/merge-media/" + artifactID
	if _, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "local-media", Status: batchfactoryv11.MergeSucceeded, OutputURL: output, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	merge := &batchfactoryv11.MergeService{Store: store, Enabled: true}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Merge: merge, LocalArtifacts: files})
	path := output
	owner := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if owner.Code != http.StatusOK || !strings.Contains(owner.Body.String(), "ftyp") {
		t.Fatalf("owner status=%d body=%q", owner.Code, owner.Body.String())
	}
	other := signedJSONRequest(t, api, now, "bob", http.MethodGet, path, nil)
	if other.Code != http.StatusNotFound {
		t.Fatalf("other status=%d body=%s", other.Code, other.Body.String())
	}
}

func TestSliceFiveServesCachedMergeCoverOnlyToItsOwner(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifactID := "merge_local_cover_1"
	if _, err := files.SaveMP4(artifactID, strings.NewReader("0000ftypisom-local-merged-video")); err != nil {
		t.Fatal(err)
	}
	coverBytes := append([]byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x01, 0x00, 0x00, 0x01}, bytes.Repeat([]byte{0}, 32)...)
	if _, err := files.SaveImage(artifactID, "image/jpeg", bytes.NewReader(coverBytes)); err != nil {
		t.Fatal(err)
	}
	output := "/api/batch-factory/v11/batches/" + batch.ID + "/merge-media/" + artifactID
	if _, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "local-cover", Status: batchfactoryv11.MergeSucceeded, OutputURL: output, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	merge := &batchfactoryv11.MergeService{Store: store, Enabled: true}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Merge: merge, LocalArtifacts: files})
	coverPath := "/api/batch-factory/v11/batches/" + batch.ID + "/merge-cover/" + artifactID
	owner := signedJSONRequest(t, api, now, "alice", http.MethodGet, coverPath, nil)
	if owner.Code != http.StatusOK || owner.Header().Get("Content-Type") != "image/jpeg" || owner.Body.Len() == 0 {
		t.Fatalf("owner cover status=%d type=%q bytes=%d", owner.Code, owner.Header().Get("Content-Type"), owner.Body.Len())
	}
	other := signedJSONRequest(t, api, now, "bob", http.MethodGet, coverPath, nil)
	if other.Code != http.StatusNotFound {
		t.Fatalf("cross-owner cover status=%d body=%s", other.Code, other.Body.String())
	}
}

func TestSliceFiveMigratesLocalMergeOnlyAfterRemoteStoreAcceptsIt(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "b", Books: []batchfactoryv11.CreateBookInput{{Title: "k", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	artifactID := "merge_local_media_2"
	if _, err := files.SaveMP4(artifactID, strings.NewReader("0000ftypisom-local-merged-video")); err != nil {
		t.Fatal(err)
	}
	output := "/api/batch-factory/v11/batches/" + batch.ID + "/merge-media/" + artifactID
	job, err := store.CreateMergeJob(context.Background(), batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, RequestID: "local-media-migrate", Status: batchfactoryv11.MergeSucceeded, OutputURL: output, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	uploaded := false
	remote := mergeOutputFunc(func(_ context.Context, taskID, path string) (string, error) {
		if taskID != job.ID {
			t.Fatalf("task=%q want=%q", taskID, job.ID)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("source artifact missing before upload: %v", err)
		}
		uploaded = true
		return "tos://bucket-a/batch-merged/" + taskID + ".mp4", nil
	})
	merge := &batchfactoryv11.MergeService{Store: store, Enabled: true}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 5, Store: store, Merge: merge, LocalArtifacts: files, MergeOutput: remote})
	path := "/api/batch-factory/v11/batches/" + batch.ID + "/merge-artifacts/migrate-to-tos"
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"migrated":1`) || !uploaded {
		t.Fatalf("status=%d body=%s uploaded=%v", rec.Code, rec.Body.String(), uploaded)
	}
	if _, err := files.Open(artifactID + ".mp4"); !os.IsNotExist(err) {
		t.Fatalf("local artifact should be deleted only after upload, err=%v", err)
	}
	jobs, err := store.ListMergeJobs(context.Background(), "alice", batch.ID)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("jobs=%v err=%v", jobs, err)
	}
	if got, want := jobs[0].OutputURL, "/api/batch-factory/v11/batches/"+batch.ID+"/merge-media/"+job.ID; got != want {
		t.Fatalf("output=%q want=%q", got, want)
	}
}
