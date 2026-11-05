package batchfactoryv11

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

type recordingProductionAdapter struct {
	calls   int
	ref     ProviderTaskRef
	prompts []FinalPrompt
}

func (a *recordingProductionAdapter) Submit(_ context.Context, _ FrozenVideoModel, prompt FinalPrompt) (ProviderTaskRef, error) {
	a.calls++
	a.prompts = append(a.prompts, prompt)
	return a.ref, nil
}

type countingPromptResolver struct {
	delegate PromptResolver
	calls    int
}

type recordingProductionPoller struct {
	calls int
	ref   ProviderTaskRef
}

type recordingLocalRetryClient struct {
	inputs []LocalVideoJobInput
}

func (c *recordingLocalRetryClient) CreateVideoJob(_ context.Context, _ string, input LocalVideoJobInput) (LocalVideoJob, error) {
	c.inputs = append(c.inputs, input)
	if len(c.inputs) == 1 {
		return LocalVideoJob{ID: "lej_first", State: "failed", ErrorMessage: "first attempt failed"}, nil
	}
	return LocalVideoJob{ID: "lej_retry", State: "queued"}, nil
}

func (c *recordingLocalRetryClient) GetVideoJob(context.Context, string, string) (LocalVideoJob, error) {
	return LocalVideoJob{}, ErrNotFound
}

func (p *recordingProductionPoller) Poll(context.Context, FrozenVideoModel, ProviderTaskRef) (ProviderTaskRef, error) {
	p.calls++
	return p.ref, nil
}

func (r *countingPromptResolver) Compile(ctx context.Context, owner, batchID, bookID, videoID string) (FinalPrompt, error) {
	r.calls++
	return r.delegate.Compile(ctx, owner, batchID, bookID, videoID)
}

func TestProductionGateBlocksBeforeCompilerAndAdapter(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: "succeeded", MediaURL: "https://media.example/video.mp4"}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: false, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	if _, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-1"); err == nil {
		t.Fatal("expected disabled production gate error")
	}
	if adapter.calls != 0 {
		t.Fatalf("disabled gate called adapter %d times", adapter.calls)
	}
}

func TestRepeatedProductionRequestReturnsSameDurableJob(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: "succeeded", MediaURL: "https://media.example/video.mp4"}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	first, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == "" || first.ID != second.ID {
		t.Fatalf("jobs differ: first=%+v second=%+v", first, second)
	}
	if adapter.calls != 1 {
		t.Fatalf("duplicate request submitted %d provider tasks", adapter.calls)
	}
}

func TestProductionStatusPersistsCompilerHashAndMedia(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: "succeeded", ProviderTaskID: "provider-1", MediaURL: "https://media.example/video.mp4"}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	job, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != ProductionSucceeded || len(job.Tasks) != 1 {
		t.Fatalf("job=%+v", job)
	}
	task := job.Tasks[0]
	if task.FinalPromptHash == "" || task.MediaURL != "https://media.example/video.mp4" || task.ProviderTaskID != "provider-1" {
		t.Fatalf("task=%+v", task)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Jobs) != 1 || status.Jobs[0].ID != job.ID || status.Jobs[0].Status != ProductionSucceeded {
		t.Fatalf("status=%+v", status)
	}
}

func TestProductionSubmitsExactlyThePromptItPersisted(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionSucceeded}}
	compiler := &countingPromptResolver{delegate: &PromptCompilerService{Store: store}}
	service := &ProductionService{Store: store, Compiler: compiler, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	job, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-frozen-prompt")
	if err != nil {
		t.Fatal(err)
	}
	if compiler.calls != len(job.Tasks) {
		t.Fatalf("compiled %d times for %d persisted tasks", compiler.calls, len(job.Tasks))
	}
	if len(adapter.prompts) != 1 || adapter.prompts[0].SnapshotHash != job.Tasks[0].FinalPromptHash || adapter.prompts[0].CompiledPrompt != job.Tasks[0].CompiledPrompt {
		t.Fatalf("provider prompt diverged from durable task: prompt=%+v task=%+v", adapter.prompts, job.Tasks[0])
	}
}

func TestFixedSingleProductionOnlySubmitsVideoOne(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", true)
	provider := &queuedDirectorProvider{values: []string{twoDirectorVideosJSON(t)}}
	if _, err := (&DirectorService{Store: store, Provider: provider}).RunDirector(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil || len(latest.Books[0].Videos) != 2 {
		t.Fatalf("director videos=%+v err=%v", latest.Books, err)
	}
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionSucceeded}}
	job, err := (&ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}).SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "fixed-single")
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 1 || job.Tasks[0].VideoID != latest.Books[0].Videos[0].ID || adapter.calls != 1 {
		t.Fatalf("fixed single submitted %#v adapter=%d", job.Tasks, adapter.calls)
	}
}

func TestProductionStatusReconcilesRunningProviderTask(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{ProviderTaskID: "provider-running", State: ProductionRunning}}
	poller := &recordingProductionPoller{ref: ProviderTaskRef{ProviderTaskID: "provider-running", State: ProductionSucceeded, MediaURL: "https://media.example/video.mp4"}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Poller: poller, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	if _, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-running"); err != nil {
		t.Fatal(err)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if poller.calls != 1 || len(status.Jobs) != 1 || status.Jobs[0].Status != ProductionSucceeded || status.Jobs[0].Tasks[0].MediaURL == "" {
		t.Fatalf("poller=%+v status=%+v", poller, status)
	}
}

func TestProductionStatusPersistsTerminalProviderFailureReason(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{ProviderTaskID: "provider-failed", State: ProductionRunning}}
	poller := &recordingProductionPoller{ref: ProviderTaskRef{ProviderTaskID: "provider-failed", State: ProductionFailed, ErrorMessage: "insufficient balance"}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Poller: poller, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	if _, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "request-provider-failure"); err != nil {
		t.Fatal(err)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Jobs) != 1 || len(status.Jobs[0].Tasks) != 1 {
		t.Fatalf("status=%+v", status)
	}
	task := status.Jobs[0].Tasks[0]
	if task.Status != ProductionFailed || task.ErrorMessage != "insufficient balance" {
		t.Fatalf("task=%+v", task)
	}
}

func TestLocalExecutorRetryUsesDistinctSourceTaskIdentity(t *testing.T) {
	ctx := context.Background()
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"videoModelId": rawSetting(t, "local-doubao-executor-video")}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	batch, _ = store.GetBatch(ctx, "alice", batch.ID)
	book, video = batch.Books[0], batch.Books[0].Videos[0]
	client := &recordingLocalRetryClient{}
	service := &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store}, Enabled: true,
		LocalExecutor: NewLocalExecutorVideoAdapter(client, ""), Model: FrozenVideoModel{MaxDuration: 15},
	}
	if _, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, "first", VideoProviderDoubaoLocal, ProductionOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, "retry", VideoProviderDoubaoLocal, ProductionOptions{Force: true, VideoID: video.ID}); err != nil {
		t.Fatal(err)
	}
	if len(client.inputs) != 2 || client.inputs[0].SourceTaskID == client.inputs[1].SourceTaskID {
		t.Fatalf("source task ids must be unique per production attempt: %+v", client.inputs)
	}
}

func TestCancelBatchCancelsOnlyOwnedLocalExecutorTasks(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	job, err := store.CreateProductionJob(context.Background(), ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "cancel-local", DirectorRevisionID: book.DirectorRevision.ID,
		Tasks: []ProductionTask{{VideoID: video.ID, Provider: VideoProviderDoubaoLocal, ProviderTaskID: "lej_cancel", Status: ProductionRunning, Attempt: 1, FinalPromptHash: "snapshot", CompiledPrompt: "prompt"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	other, err := store.CreateProductionJob(context.Background(), ProductionJob{
		Owner:              "alice",
		BatchID:            batch.ID,
		BookID:             book.ID,
		RequestID:          "another-operation",
		DirectorRevisionID: book.DirectorRevision.ID,
		Tasks:              []ProductionTask{{VideoID: video.ID, Provider: VideoProviderDoubaoLocal, ProviderTaskID: "lej_other", Status: ProductionRunning, Attempt: 2, FinalPromptHash: "snapshot-2", CompiledPrompt: "prompt"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeLocalVideoClient{createdOwner: "alice", job: LocalVideoJob{ID: "lej_cancel", State: "running"}}
	service := &ProductionService{Store: store, LocalExecutor: NewLocalExecutorVideoAdapter(client, "https://platform.example")}
	result, err := service.CancelBatch(context.Background(), "alice", batch.ID, "cancel-local")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.CancelledTaskIDs) != 1 || result.CancelledTaskIDs[0] != job.Tasks[0].ID || client.cancelledID != "lej_cancel" || len(client.cancelAttempts) != 1 {
		t.Fatalf("result=%+v client=%+v", result, client)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Jobs[0].Tasks[0].Status != ProductionCancelled || status.Jobs[0].Status != ProductionCancelled || status.Jobs[1].ID != other.ID || status.Jobs[1].Tasks[0].Status != ProductionRunning {
		t.Fatalf("status=%+v", status)
	}
}

func TestProductionPersistsAndSubmitsFrozenAssetReferenceImages(t *testing.T) {
	store, batch, book, _ := seedCompiledVideo(t)
	var character BookAsset
	for _, asset := range book.AssetRecords {
		if asset.Kind == "character" && asset.Name == "林晚" {
			character = asset
			break
		}
	}
	if character.ID == "" {
		t.Fatalf("assets=%+v", book.AssetRecords)
	}
	if _, err := store.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, character.ID, CreateBookAssetImageInput{URL: "https://images.example/lin.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionSucceeded}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15, MaxReferenceImages: 1}}
	job, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "references")
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 1 || len(job.Tasks[0].ReferenceImageURLs) != 1 || job.Tasks[0].ReferenceImageURLs[0] != "https://images.example/lin.png" {
		t.Fatalf("task=%+v", job.Tasks)
	}
	if len(adapter.prompts) != 1 || len(adapter.prompts[0].ReferenceImageURLs) != 1 || adapter.prompts[0].ReferenceImageURLs[0] != job.Tasks[0].ReferenceImageURLs[0] {
		t.Fatalf("adapter=%+v task=%+v", adapter.prompts, job.Tasks[0])
	}
}

func TestProductionPersistsTargetRequestedAndActualDurations(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{
		State:                    ProductionSucceeded,
		ProviderTaskID:           "duration-provider-task",
		MediaURL:                 "https://media.example/duration.mp4",
		RequestedDurationSeconds: 7,
		ActualDurationSeconds:    6.72,
	}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	job, err := service.SubmitBookProduction(context.Background(), "alice", batch.ID, book.ID, "duration-contract")
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 1 {
		t.Fatalf("tasks=%+v", job.Tasks)
	}
	task := job.Tasks[0]
	if task.VideoID != video.ID || task.TargetDurationSeconds <= 0 || task.RequestedDurationSeconds != 7 || task.ActualDurationSeconds != 6.72 {
		t.Fatalf("duration task=%+v", task)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	stored := status.Jobs[0].Tasks[0]
	if stored.TargetDurationSeconds != task.TargetDurationSeconds || stored.RequestedDurationSeconds != task.RequestedDurationSeconds || stored.ActualDurationSeconds != task.ActualDurationSeconds {
		t.Fatalf("duration values were not durable: stored=%+v task=%+v", stored, task)
	}
}

func TestRemoveBookProductionCandidateKeepsMainVersion(t *testing.T) {
	store, batch, book, video := seedCompiledVideo(t)
	first, err := store.CreateProductionJob(context.Background(), ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "candidate-first", DirectorRevisionID: book.DirectorRevision.ID,
		Tasks: []ProductionTask{{VideoID: video.ID, Status: ProductionSucceeded, Attempt: 1, FinalPromptHash: "first", CompiledPrompt: "first", MediaURL: "https://media.example/first.mp4"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateProductionJob(context.Background(), ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "candidate-second", DirectorRevisionID: book.DirectorRevision.ID,
		Tasks: []ProductionTask{{VideoID: video.ID, Status: ProductionSucceeded, Attempt: 1, FinalPromptHash: "second", CompiledPrompt: "second", MediaURL: "https://media.example/second.mp4"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	service := &ProductionService{Store: store}
	if err := service.RemoveBookProductionCandidate(context.Background(), "alice", batch.ID, book.ID, video.ID, second.Tasks[0].ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("implicit latest main must be protected, got %v", err)
	}
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"primaryMediaTaskId": raw(first.Tasks[0].ID)}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	if err := service.RemoveBookProductionCandidate(context.Background(), "alice", batch.ID, book.ID, video.ID, first.Tasks[0].ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("selected main must be protected, got %v", err)
	}
	if err := service.RemoveBookProductionCandidate(context.Background(), "alice", batch.ID, book.ID, video.ID, second.Tasks[0].ID); err != nil {
		t.Fatal(err)
	}
	status, err := service.GetBatchStatus(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Jobs) != 2 || len(status.Jobs[1].Tasks) != 0 || len(status.Jobs[0].Tasks) != 1 || status.Jobs[0].Tasks[0].ID != first.Tasks[0].ID {
		t.Fatalf("removed candidate was still exposed: %+v", status.Jobs)
	}
}

// The error_message column is VARCHAR(255); truncation must stay byte-bounded
// for storage safety but never split a multibyte rune, otherwise the UI shows
// broken characters and MySQL can reject the insert.
func TestProductionErrorTruncatesOnRuneBoundary(t *testing.T) {
	// Byte 200 falls inside the first "中" (bytes 199-201 of the string).
	message := strings.Repeat("a", 199) + strings.Repeat("中", 2)
	if len(message) <= 200 || utf8.ValidString(message[:200]) {
		t.Fatalf("fixture must exceed 200 bytes and split a rune at byte 200: %d", len(message))
	}
	got := productionError(errors.New(message))
	if !utf8.ValidString(got) {
		t.Fatalf("truncated message is invalid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("truncated message must keep the ellipsis suffix: %q", got)
	}
	if len(got) > 200+len("...") {
		t.Fatalf("truncated message exceeds the byte budget: %d", len(got))
	}
	if strings.ContainsRune(got, utf8.RuneError) {
		t.Fatalf("truncated message contains a replacement rune: %q", got)
	}
}

// recordingVideoProviderRegistry captures which provider the executor asks for
// and refuses to resolve it, so the test never reaches a real provider host.
type recordingVideoProviderRegistry struct {
	requested []string
}

func (r *recordingVideoProviderRegistry) Put(context.Context, string, VideoProviderConfig) error {
	return nil
}

func (r *recordingVideoProviderRegistry) Resolve(_ context.Context, _ string, provider string) (VideoProviderConfig, error) {
	r.requested = append(r.requested, provider)
	return VideoProviderConfig{}, ErrUnavailable
}

func (r *recordingVideoProviderRegistry) View(context.Context, string, string) (VideoProviderConfigView, error) {
	return VideoProviderConfigView{}, ErrUnavailable
}

// Live incident (batch_0a9b163b83c04cf6da460439c6912fc3): the batch unified
// configuration was AutoDL H3 while the book card displayed
// seedance-2-0-official, and the retry submitted through AutoDL H3. The
// automation still passes the batch engine as its provider argument, so
// SubmitBookProductionWithOptions must derive the provider from the book's
// effective settings at submission time instead of trusting the caller.
// 自动化再次触发时走的是 Missing 语义：已经有 succeeded 成片的分镜直接复用，
// 不得重新提交任务，也不得轮换成新的 orphaned 槽位。
func TestBookProductionMissingModeNeverResubmitsSucceededVideos(t *testing.T) {
	ctx := context.Background()
	store, batch, book, _ := seedCompiledVideo(t)
	tasks := make([]ProductionTask, 0, len(book.Videos))
	for _, video := range book.Videos {
		tasks = append(tasks, ProductionTask{
			VideoID: video.ID, OpeningVariantIndex: 0, Status: ProductionSucceeded, Attempt: 1,
			FinalPromptHash: "hash", CompiledPrompt: "prompt", MediaURL: "https://media.example/" + video.ID + ".mp4",
		})
	}
	if _, err := store.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "already-done",
		DirectorRevisionID: book.DirectorRevision.ID, Tasks: tasks,
	}); err != nil {
		t.Fatal(err)
	}
	before, err := store.ListProductionJobs(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	service := &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store},
		LocalExecutor: NewLocalExecutorVideoAdapter(&recordingLocalRetryClient{}, ""),
		Enabled:       true,
		Model:         FrozenVideoModel{MaxDuration: 15},
	}
	if _, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, "again", VideoProviderDoubaoLocal, ProductionOptions{}); !errors.Is(err, ErrConflict) {
		t.Fatalf("missing mode must stop with a conflict instead of resubmitting: %v", err)
	}
	after, err := store.ListProductionJobs(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("production jobs grew from %d to %d; completed videos must never be resubmitted", len(before), len(after))
	}
}

func TestVideoSubmitHonorsTheBookOverrideInsteadOfTheBatchEngine(t *testing.T) {
	ctx := context.Background()
	store, batch, book, _ := seedCompiledVideo(t)
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"videoProvider": raw(VideoProviderAutoDLH3),
			"videoModelId":  raw(AutoDLH3Model),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"videoProvider": raw("yfai"),
			"videoModelId":  raw("seedance-2-0-official"),
		},
		ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	latest, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	latestBook := latest.Books[0]
	effective := ResolveSettings(latest.SettingsState.Patch, latestBook.SettingsState.Patch)
	if got := rawString(effective, "videoModelId", ""); got != "seedance-2-0-official" {
		t.Fatalf("fixture must resolve the book model, got %q", got)
	}

	registry := &recordingVideoProviderRegistry{}
	service := &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store},
		ProviderRegistry: registry, Enabled: true,
		Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15},
	}
	// Exactly what the automation path passes: the batch engine value.
	_, _ = service.SubmitBookProductionWithOptions(ctx, "alice", latest.ID, latestBook.ID, "book-override-video", VideoProviderAutoDLH3, ProductionOptions{})

	if len(registry.requested) != 1 {
		t.Fatalf("resolved providers=%v, want exactly one resolution", registry.requested)
	}
	if registry.requested[0] != VideoProviderYFAISeedance {
		t.Fatalf("submission used provider %q, want the book override %q (batch value was %q)",
			registry.requested[0], VideoProviderYFAISeedance, VideoProviderAutoDLH3)
	}
}

// Live incident follow-up: the retry reused the SAME automation request id, and
// FindProductionJob returned the previous failed job (stored AutoDL provider and
// stored error) without submitting again. A reopened run now uses a retry-scoped
// video request id, which must resolve the book's current engine.
func TestRetriedVideoUsesTheBookEngineInsteadOfTheStoredProvider(t *testing.T) {
	ctx := context.Background()
	store, batch, book, video := seedCompiledVideo(t)
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"videoProvider": raw(VideoProviderAutoDLH3),
			"videoModelId":  raw(AutoDLH3Model),
		},
		ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{
			"videoProvider": raw("yfai"),
			"videoModelId":  raw("seedance-2-0-official"),
		},
		ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	oldRequestID := "run-1-video-" + book.ID
	old, err := store.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: oldRequestID,
		DirectorRevisionID: book.DirectorRevision.ID,
		Tasks: []ProductionTask{{
			VideoID: video.ID, Provider: VideoProviderAutoDLH3, Status: ProductionFailed,
			ErrorMessage: "AutoDL H3 rejected submission: code=RequestParameterIsWrong message=缺少必填参数：ref_image_0 request_id=6dc434c43ebac55aec5cc247756c8c57",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	registry := &recordingVideoProviderRegistry{}
	service := &ProductionService{
		Store: store, Compiler: &PromptCompilerService{Store: store},
		ProviderRegistry: registry, Enabled: true,
		Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15},
	}

	// The old request id replays the stored failed job and never resolves a
	// provider: this is the root cause the retry-scoped id fixes.
	stale, err := service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, oldRequestID, VideoProviderAutoDLH3, ProductionOptions{})
	if err != nil || stale.ID != old.ID {
		t.Fatalf("stale request id must replay the failed job, got id=%q err=%v", stale.ID, err)
	}
	if len(registry.requested) != 0 {
		t.Fatalf("stale request id resolved %v, want no provider resolution", registry.requested)
	}

	// A reopened run (Retry sets retryAttempt) submits under a new request id and
	// must switch from the stored AutoDL task to the book's YFAI Seedance engine.
	_, _ = service.SubmitBookProductionWithOptions(ctx, "alice", batch.ID, book.ID, oldRequestID+"-r1", VideoProviderAutoDLH3, ProductionOptions{})
	if len(registry.requested) != 1 {
		t.Fatalf("retry resolved providers=%v, want exactly one", registry.requested)
	}
	if registry.requested[0] != VideoProviderYFAISeedance {
		t.Fatalf("retry submitted through %q, want the book engine %q (stored provider was %q)",
			registry.requested[0], VideoProviderYFAISeedance, VideoProviderAutoDLH3)
	}
}

// 槽位级保留之后，未变化槽位沿用旧 video 身份，它的成功任务挂在旧修订的
// job 上。投递去重按当前 active 槽位过滤而不是按修订号过滤：保留槽位的历史
// 成功任务必须继续阻止重复投递，只有真正轮换的槽位才会重新生成。
func TestProductionBlocksSucceededTasksFromPreviousRevisionForKeptSlots(t *testing.T) {
	ctx := context.Background()
	store, batch, book := seedOpeningBook(t, false, 2, []DirectorVideo{
		{DurationSec: 10, FinalPrompt: "保留分镜"},
		{DurationSec: 10, FinalPrompt: "新分镜"},
	})
	keptVideo := book.Videos[0]
	rotatedVideo := book.Videos[1]
	if _, err := store.CreateProductionJob(ctx, ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "old-revision-job",
		DirectorRevisionID: "director-old-revision",
		Tasks: []ProductionTask{{
			VideoID: keptVideo.ID, Status: ProductionSucceeded, Attempt: 1,
			FinalPromptHash: "old-hash", CompiledPrompt: "old", MediaURL: "https://media.example/kept.mp4",
		}},
	}); err != nil {
		t.Fatal(err)
	}
	adapter := &recordingProductionAdapter{ref: ProviderTaskRef{State: ProductionSucceeded}}
	service := &ProductionService{Store: store, Compiler: &PromptCompilerService{Store: store}, Adapter: adapter, Enabled: true, Model: FrozenVideoModel{ID: "video-model-a", MaxDuration: 15}}
	job, err := service.SubmitBookProduction(ctx, "alice", batch.ID, book.ID, "request-retention")
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Tasks) != 1 || job.Tasks[0].VideoID != rotatedVideo.ID {
		t.Fatalf("tasks=%+v, want only the rotated slot to be submitted", job.Tasks)
	}
	if adapter.calls != 1 {
		t.Fatalf("adapter calls=%d, want 1: the kept slot's old-revision success must keep blocking", adapter.calls)
	}
}
