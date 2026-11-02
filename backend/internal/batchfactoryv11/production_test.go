package batchfactoryv11

import (
	"context"
	"errors"
	"testing"
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
