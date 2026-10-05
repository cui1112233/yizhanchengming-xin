package generation

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeProvider struct {
	responses []string
	errors    map[int]error
	calls     []TextRequest
}

func (p *fakeProvider) Complete(_ context.Context, req TextRequest) (string, error) {
	index := len(p.calls)
	p.calls = append(p.calls, req)
	if err := p.errors[index]; err != nil {
		return "", err
	}
	if index >= len(p.responses) {
		return "ok", nil
	}
	return p.responses[index], nil
}

type memoryStore struct {
	books       map[int64]intake.Book
	projectBook map[int64][]int64
	prompts     map[string]Prompt
	bookRuns    []BookRun
	stageRuns   []StageRun
	now         time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		books: map[int64]intake.Book{
			11: {ID: 11, IntakeID: 7, Title: "甲书", OriginalText: "她推开门，发现桌上的信不见了。"},
			12: {ID: 12, IntakeID: 7, Title: "乙书", OriginalText: "他转身离开医院。"},
			13: {ID: 13, IntakeID: 7, Title: "丙书", OriginalText: "雨夜里她攥紧手机。"},
		},
		projectBook: map[int64][]int64{3: {11, 12, 13}},
		prompts: map[string]Prompt{},
		now: time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC),
	}
}

func (s *memoryStore) seedPrompts() {
	for _, prompt := range DefaultPrompts() {
		s.prompts[prompt.Key] = prompt
	}
}

func (s *memoryStore) GetBookForProject(_ context.Context, projectID, bookID int64) (intake.Book, error) {
	for _, id := range s.projectBook[projectID] {
		if id == bookID {
			return s.books[id], nil
		}
	}
	return intake.Book{}, ErrNotFound
}

func (s *memoryStore) ListBooksForProject(_ context.Context, projectID int64) ([]intake.Book, error) {
	ids := s.projectBook[projectID]
	books := make([]intake.Book, 0, len(ids))
	for _, id := range ids {
		books = append(books, s.books[id])
	}
	return books, nil
}

func (s *memoryStore) ResolvePrompt(_ context.Context, key string) (Prompt, error) {
	prompt, ok := s.prompts[key]
	if !ok || !prompt.Enabled {
		return Prompt{}, ErrNotFound
	}
	return prompt, nil
}

func (s *memoryStore) ListPrompts(_ context.Context) ([]Prompt, error) {
	out := make([]Prompt, 0, len(s.prompts))
	for _, value := range s.prompts {
		out = append(out, value)
	}
	return out, nil
}

func (s *memoryStore) CreateBookRun(_ context.Context, value BookRun) (BookRun, error) {
	value.ID = int64(len(s.bookRuns) + 1)
	value.CreatedAt = s.now
	value.UpdatedAt = s.now
	s.bookRuns = append(s.bookRuns, value)
	return value, nil
}

func (s *memoryStore) UpdateBookRun(_ context.Context, value BookRun) (BookRun, error) {
	for i := range s.bookRuns {
		if s.bookRuns[i].ID == value.ID {
			value.UpdatedAt = s.now
			s.bookRuns[i] = value
			return value, nil
		}
	}
	return BookRun{}, ErrNotFound
}

func (s *memoryStore) LatestBookRun(_ context.Context, projectID, bookID int64) (BookRun, error) {
	for i := len(s.bookRuns) - 1; i >= 0; i-- {
		if s.bookRuns[i].BatchProjectID == projectID && s.bookRuns[i].BookID == bookID {
			return s.bookRuns[i], nil
		}
	}
	return BookRun{}, ErrNotFound
}

func (s *memoryStore) ListBookRunsByProject(_ context.Context, projectID int64) ([]BookRun, error) {
	out := []BookRun{}
	for _, value := range s.bookRuns {
		if value.BatchProjectID == projectID {
			out = append(out, value)
		}
	}
	return out, nil
}

func (s *memoryStore) CreateStageRun(_ context.Context, value StageRun) (StageRun, error) {
	value.ID = int64(len(s.stageRuns) + 1)
	value.CreatedAt = s.now
	value.UpdatedAt = s.now
	s.stageRuns = append(s.stageRuns, value)
	return value, nil
}

func (s *memoryStore) UpdateStageRun(_ context.Context, value StageRun) (StageRun, error) {
	for i := range s.stageRuns {
		if s.stageRuns[i].ID == value.ID {
			value.UpdatedAt = s.now
			s.stageRuns[i] = value
			return value, nil
		}
	}
	return StageRun{}, ErrNotFound
}

func (s *memoryStore) ListStageRuns(_ context.Context, bookRunID int64) ([]StageRun, error) {
	out := []StageRun{}
	for _, value := range s.stageRuns {
		if value.BookRunID == bookRunID {
			out = append(out, value)
		}
	}
	return out, nil
}

func (s *memoryStore) LatestStageRun(_ context.Context, bookRunID int64, stage Stage) (StageRun, error) {
	for i := len(s.stageRuns) - 1; i >= 0; i-- {
		if s.stageRuns[i].BookRunID == bookRunID && s.stageRuns[i].Stage == stage {
			return s.stageRuns[i], nil
		}
	}
	return StageRun{}, ErrNotFound
}

func serviceFixture() (*Service, *memoryStore, *fakeProvider) {
	store := newMemoryStore()
	store.seedPrompts()
	provider := &fakeProvider{responses: []string{"SCRIPT", "HOOK", `{"cards":[{"shot":"推门"}]}`}, errors: map[int]error{}}
	service := NewService(store, provider, func() time.Time { return store.now })
	return service, store, provider
}

func TestRunBookCompletesScriptHookDirectorAndFinalPrompt(t *testing.T) {
	service, store, _ := serviceFixture()
	result, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "req-1"})
	if err != nil {
		t.Fatalf("RunBook: %v", err)
	}
	if result.Run.Status != StatusCompleted {
		t.Fatalf("status = %s", result.Run.Status)
	}
	if len(result.Stages) != 4 {
		t.Fatalf("stages = %d", len(result.Stages))
	}
	for _, stage := range []Stage{StageScript, StageHook, StageDirector, StageFinalPrompt} {
		latest, err := store.LatestStageRun(context.Background(), result.Run.ID, stage)
		if err != nil || latest.Status != StatusCompleted {
			t.Fatalf("stage %s = %#v err=%v", stage, latest, err)
		}
		if latest.PromptVersion <= 0 {
			t.Fatalf("stage %s missing prompt version", stage)
		}
	}
	final, _ := store.LatestStageRun(context.Background(), result.Run.ID, StageFinalPrompt)
	if !strings.Contains(final.OutputText, "SCRIPT") || !strings.Contains(final.OutputText, "HOOK") || !strings.Contains(final.OutputText, "推门") {
		t.Fatalf("final prompt did not include upstream outputs: %s", final.OutputText)
	}
}

func TestHookDisabledIsSkippedNotFailed(t *testing.T) {
	service, store, provider := serviceFixture()
	provider.responses = []string{"SCRIPT", `{"cards":[]}`}
	result, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: false, DirectorMode: DirectorNormal, RequestID: "req-skip"})
	if err != nil {
		t.Fatalf("RunBook: %v", err)
	}
	hook, err := store.LatestStageRun(context.Background(), result.Run.ID, StageHook)
	if err != nil {
		t.Fatal(err)
	}
	if hook.Status != StatusSkipped {
		t.Fatalf("hook status = %s", hook.Status)
	}
	if result.Run.Status != StatusCompleted {
		t.Fatalf("run status = %s", result.Run.Status)
	}
}

func TestH3DirectorUsesDedicatedPromptAndAuthoritativeAudio(t *testing.T) {
	service, store, provider := serviceFixture()
	service.audioProber = &fakeAudioProber{durationMS: 28250}
	if _, err := service.MeasureAudio(context.Background(), AudioMeasurementRequest{BatchProjectID: 3, BookID: 11, AudioAsset: "/audio/11.mp3"}); err != nil {
		t.Fatal(err)
	}
	provider.responses = []string{"SCRIPT", "HOOK", `{"schema_version":"h3-director/v1","director_cards":[{"shot":"a","start":0.00,"end":14.00},{"shot":"b","start":14.00,"end":28.25}]}`}
	result, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorH3, MatchAudio: true, AudioDurationSec: 999, ShotDurationLimitSec: 15, RequestID: "h3-1"})
	if err != nil {
		t.Fatalf("RunBook: %v", err)
	}
	director, _ := store.LatestStageRun(context.Background(), result.Run.ID, StageDirector)
	if director.PromptKey != PromptDirectorH3 {
		t.Fatalf("prompt key = %s", director.PromptKey)
	}
	if !strings.Contains(director.InputSnapshot, `"matchAudio":true`) || !strings.Contains(director.InputSnapshot, `"audioDurationSec":28.25`) {
		t.Fatalf("authoritative audio fields missing: %s", director.InputSnapshot)
	}
	if director.PromptVersion != 2 || !strings.Contains(director.ValidationResult, `"valid":true`) {
		t.Fatalf("prompt/validation not recorded: %#v", director)
	}
	if len(provider.calls) != 3 {
		t.Fatalf("provider calls = %d", len(provider.calls))
	}
}

func TestPlotModeUsesGenericPlotPrompt(t *testing.T) {
	service, store, _ := serviceFixture()
	result, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: false, PlotMode: true, DirectorMode: DirectorNormal, RequestID: "plot-1"})
	if err != nil {
		t.Fatal(err)
	}
	script, _ := store.LatestStageRun(context.Background(), result.Run.ID, StageScript)
	if script.PromptKey != PromptScriptPlotMode {
		t.Fatalf("prompt key = %s", script.PromptKey)
	}
	prompt := store.prompts[PromptScriptPlotMode].Content
	for _, required := range []string{"首镜头", "空间", "动作", "电影", "未成年人", "主剧情"} {
		if !strings.Contains(prompt, required) {
			t.Fatalf("plot prompt missing %q", required)
		}
	}
}

func TestRunBookFailurePersistsSafeErrorAndStopsOnlyThatBook(t *testing.T) {
	service, store, provider := serviceFixture()
	provider.errors[0] = errors.New("Authorization: Bearer super-secret-token upstream exploded")
	result, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "fail-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Run.Status != StatusFailed {
		t.Fatalf("status = %s", result.Run.Status)
	}
	script, _ := store.LatestStageRun(context.Background(), result.Run.ID, StageScript)
	if script.Status != StatusFailed {
		t.Fatalf("script status = %s", script.Status)
	}
	if strings.Contains(strings.ToLower(script.ErrorMessage), "secret") || strings.Contains(strings.ToLower(script.ErrorMessage), "bearer") {
		t.Fatalf("sensitive provider error leaked: %s", script.ErrorMessage)
	}
}

func TestBatchContinuesAfterOneBookFails(t *testing.T) {
	store := newMemoryStore()
	store.seedPrompts()
	provider := &fakeProvider{responses: []string{"", "SCRIPT2", "HOOK2", `{"cards":[]}`, "SCRIPT3", "HOOK3", `{"cards":[]}`}, errors: map[int]error{0: errors.New("provider down")}}
	service := NewService(store, provider, func() time.Time { return store.now })
	result, err := service.RunBatch(context.Background(), RunBatchRequest{BatchProjectID: 3, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "batch-1"})
	if err == nil {
		t.Fatal("batch should report partial failure")
	}
	if len(result.Books) != 3 {
		t.Fatalf("books = %d", len(result.Books))
	}
	if result.Failed != 1 || result.Completed != 2 {
		t.Fatalf("completed=%d failed=%d", result.Completed, result.Failed)
	}
}

func TestRunBookIsIdempotentAfterCompletion(t *testing.T) {
	service, _, provider := serviceFixture()
	first, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "same"})
	if err != nil {
		t.Fatal(err)
	}
	calls := len(provider.calls)
	second, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "same-again"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Run.ID != first.Run.ID {
		t.Fatalf("created duplicate run %d != %d", second.Run.ID, first.Run.ID)
	}
	if len(provider.calls) != calls {
		t.Fatalf("provider called again: %d -> %d", calls, len(provider.calls))
	}
}

func TestRetryOnlyFailedTargetStageAndReusesPrerequisites(t *testing.T) {
	service, store, provider := serviceFixture()
	provider.errors[2] = errors.New("malformed director reply")
	failed, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "retry-1"})
	if err == nil {
		t.Fatal("expected director failure")
	}
	provider.errors = map[int]error{}
	provider.responses = append(provider.responses, `{"cards":[{"shot":"ok"}]}`)
	callsBefore := len(provider.calls)
	result, err := service.RetryStage(context.Background(), RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: StageDirector, RequestID: "retry-2"})
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if result.Run.ID != failed.Run.ID {
		t.Fatalf("retry changed book run")
	}
	if len(provider.calls) != callsBefore+1 {
		t.Fatalf("retry should call only target provider once, calls %d -> %d", callsBefore, len(provider.calls))
	}
	scriptRuns := 0
	hookRuns := 0
	directorRuns := 0
	for _, stage := range store.stageRuns {
		if stage.BookRunID != result.Run.ID {
			continue
		}
		switch stage.Stage {
		case StageScript:
			scriptRuns++
		case StageHook:
			hookRuns++
		case StageDirector:
			directorRuns++
		}
	}
	if scriptRuns != 1 || hookRuns != 1 || directorRuns != 2 {
		t.Fatalf("attempts script=%d hook=%d director=%d", scriptRuns, hookRuns, directorRuns)
	}
}

func TestRetryRejectsCompletedStage(t *testing.T) {
	service, _, _ := serviceFixture()
	_, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, DirectorMode: DirectorNormal, RequestID: "done"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.RetryStage(context.Background(), RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: StageScript, RequestID: "nope"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestProjectSummaryAggregatesLatestRuns(t *testing.T) {
	service, _, _ := serviceFixture()
	for _, bookID := range []int64{11, 12} {
		_, err := service.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: bookID, HookEnabled: false, DirectorMode: DirectorNormal, RequestID: "summary"})
		if err != nil {
			t.Fatal(err)
		}
	}
	summary, err := service.ProjectSummary(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Completed != 2 || len(summary.Books) != 3 {
		t.Fatalf("summary = %#v", summary)
	}
}
