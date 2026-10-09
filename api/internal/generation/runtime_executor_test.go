package generation

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

func TestRuntimeExecutorRunsFencedPipelineAndKeepsSnapshotsSafe(t *testing.T) {
	store := runtimeStoreFixture()
	store.input.Request.ProcessingRules = "RULES CONFIG"
	store.input.Request.KnowledgeBase = "KNOWLEDGE CONFIG"
	store.input.Request.ProjectConfig = `{"characters":"PROJECT CONFIG"}`
	store.input.Request.UserConfig = `{"profile":"USER CONFIG"}`
	store.input.Request.ModelConfig = "MODEL CONFIG"
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT", StageHook: "HOOK", StageDirector: "DIRECTOR"}}
	executor := NewRuntimeExecutor(store, provider, time.Now)
	if err := executor.Execute(context.Background(), store.execution); err != nil {
		t.Fatal(err)
	}
	if got := provider.stages(); strings.Join(got, ",") != "SCRIPT,HOOK,DIRECTOR" {
		t.Fatalf("provider stages=%v", got)
	}
	latest := store.latest()
	if latest[StageScript].PromptVersion != 3 || latest[StageHook].PromptVersion != 4 || latest[StageDirector].PromptVersion != 5 || latest[StageFinalPrompt].PromptVersion != 6 {
		t.Fatalf("prompt versions=%+v", latest)
	}
	if !strings.Contains(latest[StageFinalPrompt].OutputText, "FINAL SYSTEM") || !strings.Contains(latest[StageFinalPrompt].OutputText, "DIRECTOR") {
		t.Fatalf("compiled output=%q", latest[StageFinalPrompt].OutputText)
	}
	for _, expected := range []string{"RULES CONFIG", "KNOWLEDGE CONFIG", "PROJECT CONFIG", "USER CONFIG", "MODEL CONFIG"} {
		if !strings.Contains(latest[StageFinalPrompt].OutputText, expected) {
			t.Fatalf("compiled output lost frozen config %q: %s", expected, latest[StageFinalPrompt].OutputText)
		}
	}
	for _, stage := range store.stages {
		for _, forbidden := range []string{"ORIGINAL SECRET TEXT", "SCRIPT SYSTEM", "HOOK SYSTEM", "DIRECTOR SYSTEM", "FINAL SYSTEM", "RULES CONFIG", "KNOWLEDGE CONFIG", "PROJECT CONFIG", "USER CONFIG", "MODEL CONFIG", "systemPrompt", "userPrompt"} {
			if strings.Contains(stage.InputSnapshot, forbidden) {
				t.Fatalf("%s snapshot leaked %q: %s", stage.Stage, forbidden, stage.InputSnapshot)
			}
		}
		var value map[string]any
		if stage.InputSnapshot != "" && json.Unmarshal([]byte(stage.InputSnapshot), &value) != nil {
			t.Fatalf("%s snapshot is not structured JSON: %q", stage.Stage, stage.InputSnapshot)
		}
	}
}

func TestRuntimeExecutorH3ValidationFailureIsSafeAndStopsPipeline(t *testing.T) {
	store := runtimeStoreFixture()
	store.input.Request.DirectorMode = DirectorH3
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT", StageHook: "HOOK", StageDirector: `{"director_cards":[]}`}}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution)
	var safe task9runtime.SafeExecutionError
	if !errors.As(err, &safe) || safe.SafeCode() != "GENERATION_UNAVAILABLE" || len(provider.calls) != 3 {
		t.Fatalf("err=%T %v calls=%d", err, err, len(provider.calls))
	}
	if director := store.latest()[StageDirector]; director.Status != StatusFailed || director.ErrorMessage != "生成服务暂不可用，请稍后重试" {
		t.Fatalf("director=%+v", director)
	}
	if _, exists := store.latest()[StageFinalPrompt]; exists {
		t.Fatal("final prompt persisted after invalid H3 output")
	}
}

func TestRuntimeExecutorSkipsDisabledHookWithoutProviderCall(t *testing.T) {
	store := runtimeStoreFixture()
	store.input.Request.HookEnabled = false
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT", StageDirector: "DIRECTOR"}}
	if err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(provider.stages(), ","); got != "SCRIPT,DIRECTOR" {
		t.Fatalf("provider stages=%s", got)
	}
	if hook := store.latest()[StageHook]; hook.Status != StatusSkipped || hook.PromptVersion != 0 {
		t.Fatalf("hook=%+v", hook)
	}
}

func TestRuntimeExecutorUsesAuthoritativeAudioMeasurement(t *testing.T) {
	store := runtimeStoreFixture()
	store.input.Request.MatchAudio = true
	store.input.Request.AudioDurationSec = 999
	store.input.Request.ShotDurationLimitSec = 10
	store.measurement = AudioMeasurement{ID: 31, BatchProjectID: 3, BookID: 11, DurationMS: 28000}
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT", StageHook: "HOOK", StageDirector: `{"schema_version":"h3-director/v1","director_cards":[{"start":0,"end":10},{"start":10,"end":20},{"start":20,"end":28}]}`}}
	store.input.Request.DirectorMode = DirectorH3
	if err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution); err != nil {
		t.Fatal(err)
	}
	var director TextRequest
	for _, call := range provider.calls {
		if call.Stage == StageDirector {
			director = call
		}
	}
	if director.AudioDurationSec != 28 || director.ShotDurationLimitSec != 10 {
		t.Fatalf("director request=%+v", director)
	}
	if snapshot := store.latest()[StageDirector].InputSnapshot; !strings.Contains(snapshot, `"audio_measurement_id":31`) || !strings.Contains(snapshot, `"audio_duration_ms":28000`) {
		t.Fatalf("director snapshot=%s", snapshot)
	}
}

func TestRuntimeExecutorStageRetryCopiesOnlyUpstreamAndDoesNotFullRerun(t *testing.T) {
	store := runtimeStoreFixture()
	store.input.Action = task9runtime.GenerationActionStageRetry
	store.input.RetryStage = StageDirector
	store.input.SourceBookRunID = 70
	store.sourceStages = []StageRun{
		{ID: 701, BookRunID: 70, BookID: 11, Stage: StageScript, Status: StatusCompleted, PromptKey: PromptScript, PromptVersion: 2, OutputText: "OLD SCRIPT"},
		{ID: 702, BookRunID: 70, BookID: 11, Stage: StageHook, Status: StatusCompleted, PromptKey: PromptHook, PromptVersion: 2, OutputText: "OLD HOOK"},
		{ID: 703, BookRunID: 70, BookID: 11, Stage: StageDirector, Status: StatusFailed, PromptKey: PromptDirector, PromptVersion: 2},
	}
	provider := &runtimeProvider{responses: map[Stage]string{StageDirector: "NEW DIRECTOR"}}
	if err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(provider.stages(), ","); got != "DIRECTOR" {
		t.Fatalf("stage retry degraded to full rerun: %s", got)
	}
	latest := store.latest()
	if latest[StageScript].OutputText != "OLD SCRIPT" || latest[StageHook].OutputText != "OLD HOOK" || latest[StageDirector].OutputText != "NEW DIRECTOR" {
		t.Fatalf("retry stages=%+v", latest)
	}
	if !strings.Contains(latest[StageScript].InputSnapshot, `"source_stage_run_id":701`) {
		t.Fatalf("copied snapshot=%s", latest[StageScript].InputSnapshot)
	}
}

func TestRuntimeExecutorRejectsInvalidRetryBeforeProvider(t *testing.T) {
	for _, mutate := range []func(*runtimeExecutorStore){
		func(s *runtimeExecutorStore) { s.input.RetryStage = Stage("UNKNOWN") },
		func(s *runtimeExecutorStore) { s.input.SourceBookRunID = 0 },
		func(s *runtimeExecutorStore) { s.sourceStages = nil },
	} {
		store := runtimeStoreFixture()
		store.input.Action = task9runtime.GenerationActionStageRetry
		store.input.RetryStage = StageDirector
		store.input.SourceBookRunID = 70
		store.sourceStages = []StageRun{{ID: 701, BookRunID: 70, BookID: 11, Stage: StageScript, Status: StatusCompleted, OutputText: "SCRIPT"}, {ID: 702, BookRunID: 70, BookID: 11, Stage: StageHook, Status: StatusSkipped}, {ID: 703, BookRunID: 70, BookID: 11, Stage: StageDirector, Status: StatusFailed}}
		mutate(store)
		provider := &runtimeProvider{}
		if err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid retry err=%v", err)
		}
		if len(provider.calls) != 0 {
			t.Fatalf("provider called before validation: %d", len(provider.calls))
		}
	}
}

func TestRuntimeExecutorRejectsInconsistentRuntimeInputBeforeProvider(t *testing.T) {
	for _, mutate := range []func(*runtimeExecutorStore){
		func(s *runtimeExecutorStore) {
			s.input.Action = task9runtime.GenerationActionFull
			s.input.RetryStage = StageScript
		},
		func(s *runtimeExecutorStore) { s.input.Request.DirectorMode = DirectorMode("unknown") },
		func(s *runtimeExecutorStore) {
			s.input.Action = task9runtime.GenerationActionStageRetry
			s.input.RetryStage = StageDirector
			s.input.SourceBookRunID = 70
			s.sourceStages = []StageRun{{ID: 701, BookRunID: 70, BookID: 11, Stage: StageScript, Status: StatusCompleted, OutputText: "SCRIPT"}, {ID: 702, BookRunID: 70, BookID: 11, Stage: StageHook, Status: StatusSkipped}, {ID: 703, BookRunID: 70, BookID: 11, Stage: StageDirector, Status: StatusFailed}}
		},
	} {
		store := runtimeStoreFixture()
		mutate(store)
		provider := &runtimeProvider{}
		if err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution); !errors.Is(err, ErrInvalid) {
			t.Fatalf("invalid runtime input err=%v input=%+v", err, store.input)
		}
		if len(provider.calls) != 0 {
			t.Fatalf("provider called before validation: %d", len(provider.calls))
		}
	}
}

func TestRuntimeExecutorPersistsSafeProviderFailure(t *testing.T) {
	store := runtimeStoreFixture()
	providerErr := errors.New("Authorization: Bearer secret-provider-token")
	provider := &runtimeProvider{errStage: StageScript, err: providerErr}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution)
	var safe task9runtime.SafeExecutionError
	if !errors.As(err, &safe) || safe.SafeCode() != "GENERATION_FAILED" || strings.Contains(safe.SafeMessage(), "secret") || !errors.Is(err, providerErr) {
		t.Fatalf("safe error=%T %v", err, err)
	}
	stage := store.latest()[StageScript]
	if stage.Status != StatusFailed || stage.ErrorMessage != outcomeFailureMessage || strings.Contains(stage.ErrorMessage, "secret") {
		t.Fatalf("failed stage=%+v", stage)
	}
}

func TestRuntimeExecutorStopsAfterStaleStageWrite(t *testing.T) {
	store := runtimeStoreFixture()
	store.staleOnUpdate = StageScript
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT"}}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution)
	if !errors.Is(err, task9runtime.ErrStaleExecution) || len(provider.calls) != 1 {
		t.Fatalf("err=%v calls=%d", err, len(provider.calls))
	}
}

func TestRuntimeExecutorChecksFenceBeforeNextPaidStage(t *testing.T) {
	store := runtimeStoreFixture()
	store.staleOnCreate = StageHook
	provider := &runtimeProvider{responses: map[Stage]string{StageScript: "SCRIPT", StageHook: "HOOK"}}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(context.Background(), store.execution)
	if !errors.Is(err, task9runtime.ErrStaleExecution) || strings.Join(provider.stages(), ",") != "SCRIPT" {
		t.Fatalf("err=%v provider stages=%v", err, provider.stages())
	}
}

func TestRuntimeExecutorContextCancellationIsNotPersistedAsBusinessFailure(t *testing.T) {
	store := runtimeStoreFixture()
	ctx, cancel := context.WithCancel(context.Background())
	provider := &runtimeProvider{callback: cancel, errStage: StageScript, err: context.Canceled}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(ctx, store.execution)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if stage := store.latest()[StageScript]; stage.Status != StatusRunning || stage.ErrorMessage != "" {
		t.Fatalf("cancellation persisted as business failure: %+v", stage)
	}
}

func TestRuntimeExecutorExpiredContextIsNotPersistedAsBusinessFailure(t *testing.T) {
	store := runtimeStoreFixture()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	provider := &runtimeProvider{errStage: StageScript, err: context.DeadlineExceeded}
	err := NewRuntimeExecutor(store, provider, time.Now).Execute(ctx, store.execution)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if len(store.stages) != 0 {
		t.Fatalf("expired context persisted stages: %+v", store.stages)
	}
}

type runtimeProvider struct {
	responses map[Stage]string
	errStage  Stage
	err       error
	callback  func()
	calls     []TextRequest
}

func (p *runtimeProvider) Complete(_ context.Context, req TextRequest) (string, error) {
	p.calls = append(p.calls, req)
	if p.callback != nil {
		p.callback()
	}
	if req.Stage == p.errStage {
		return "", p.err
	}
	return p.responses[req.Stage], nil
}
func (p *runtimeProvider) stages() []string {
	out := make([]string, len(p.calls))
	for i := range p.calls {
		out[i] = string(p.calls[i].Stage)
	}
	return out
}

type runtimeExecutorStore struct {
	execution     task9runtime.Execution
	run           BookRun
	input         RuntimeExecutionInput
	book          intake.Book
	measurement   AudioMeasurement
	prompts       map[string]Prompt
	stages        []StageRun
	sourceStages  []StageRun
	staleOnUpdate Stage
	staleOnCreate Stage
	nextID        int64
}

func runtimeStoreFixture() *runtimeExecutorStore {
	return &runtimeExecutorStore{
		execution: task9runtime.Execution{BookRunID: 90, Attempt: 1, FencingToken: 17, Owner: "worker-a"},
		run:       BookRun{ID: 90, BatchProjectID: 3, BookID: 11, Status: StatusRunning, RequestID: "runtime"},
		input:     RuntimeExecutionInput{Action: task9runtime.GenerationActionFull, Request: RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, ShotDurationLimitSec: 15}},
		book:      intake.Book{ID: 11, OriginalText: "ORIGINAL SECRET TEXT"}, nextID: 100,
		prompts: map[string]Prompt{
			PromptScript:     {Key: PromptScript, Version: 3, Enabled: true, Content: "SCRIPT SYSTEM"},
			PromptHook:       {Key: PromptHook, Version: 4, Enabled: true, Content: "HOOK SYSTEM"},
			PromptDirector:   {Key: PromptDirector, Version: 5, Enabled: true, Content: "DIRECTOR SYSTEM"},
			PromptDirectorH3: {Key: PromptDirectorH3, Version: 5, Enabled: true, Content: "DIRECTOR H3 SYSTEM"},
			PromptFinal:      {Key: PromptFinal, Version: 6, Enabled: true, Content: "FINAL SYSTEM"},
		},
	}
}

func (s *runtimeExecutorStore) RuntimeBookRun(_ context.Context, e task9runtime.Execution) (BookRun, RuntimeExecutionInput, error) {
	if e != s.execution {
		return BookRun{}, RuntimeExecutionInput{}, task9runtime.ErrStaleExecution
	}
	return s.run, s.input, nil
}
func (s *runtimeExecutorStore) GetBookForProject(context.Context, int64, int64) (intake.Book, error) {
	return s.book, nil
}
func (s *runtimeExecutorStore) LatestAudioMeasurement(context.Context, int64, int64) (AudioMeasurement, error) {
	if s.measurement.ID == 0 {
		return AudioMeasurement{}, ErrNotFound
	}
	return s.measurement, nil
}
func (s *runtimeExecutorStore) ResolvePrompt(_ context.Context, key string) (Prompt, error) {
	p, ok := s.prompts[key]
	if !ok {
		return Prompt{}, ErrNotFound
	}
	return p, nil
}
func (s *runtimeExecutorStore) ListStageRuns(_ context.Context, bookRunID int64) ([]StageRun, error) {
	if bookRunID == s.input.SourceBookRunID {
		return append([]StageRun{}, s.sourceStages...), nil
	}
	return append([]StageRun{}, s.stages...), nil
}
func (s *runtimeExecutorStore) CreateStageRunFenced(_ context.Context, e task9runtime.Execution, v StageRun) (StageRun, error) {
	if e != s.execution || v.Stage == s.staleOnCreate {
		return StageRun{}, task9runtime.ErrStaleExecution
	}
	s.nextID++
	v.ID, v.BookRunID, v.BookID = s.nextID, s.run.ID, s.run.BookID
	v.Attempt = 1
	for _, existing := range s.stages {
		if existing.Stage == v.Stage && existing.Attempt >= v.Attempt {
			v.Attempt = existing.Attempt + 1
		}
	}
	s.stages = append(s.stages, v)
	return v, nil
}
func (s *runtimeExecutorStore) UpdateStageRunFenced(_ context.Context, e task9runtime.Execution, v StageRun) (StageRun, error) {
	if e != s.execution || v.Stage == s.staleOnUpdate {
		return StageRun{}, task9runtime.ErrStaleExecution
	}
	for i := range s.stages {
		if s.stages[i].ID == v.ID {
			s.stages[i] = v
			return v, nil
		}
	}
	return StageRun{}, ErrNotFound
}
func (s *runtimeExecutorStore) latest() map[Stage]StageRun {
	out := map[Stage]StageRun{}
	for _, stage := range s.stages {
		if current, ok := out[stage.Stage]; !ok || stage.Attempt >= current.Attempt {
			out[stage.Stage] = stage
		}
	}
	return out
}
