package generation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type Clock func() time.Time

type Service struct {
	store Store
	provider Provider
	resolver *PromptResolver
	now Clock
	compiler FinalPromptCompiler
}

func NewService(store Store, provider Provider, now Clock) *Service {
	if now == nil { now = time.Now }
	return &Service{store: store, provider: provider, resolver: NewPromptResolver(store), now: now, compiler: FinalPromptCompiler{}}
}

func safeError(err error) string {
	if err == nil { return "" }
	text := strings.ToLower(err.Error())
	for _, marker := range []string{"authorization", "bearer", "token", "api key", "apikey", "password", "secret", "dsn"} {
		if strings.Contains(text, marker) { return "上游生成服务请求失败，请稍后重试" }
	}
	if len(err.Error()) > 300 { return "生成阶段执行失败，请查看服务日志" }
	return err.Error()
}

func (s *Service) validate() error {
	if s == nil || s.store == nil || s.provider == nil || s.resolver == nil { return ErrUnavailable }
	return nil
}

func (s *Service) result(ctx context.Context, run BookRun) (BookGenerationResult, error) {
	stages, err := s.store.ListStageRuns(ctx, run.ID)
	if err != nil { return BookGenerationResult{}, err }
	latest := map[Stage]StageRun{}
	for _, value := range stages {
		current, ok := latest[value.Stage]
		if !ok || value.Attempt > current.Attempt || (value.Attempt == current.Attempt && value.ID > current.ID) { latest[value.Stage] = value }
	}
	return BookGenerationResult{Run: run, Stages: stages, Latest: latest}, nil
}

func (s *Service) RunBook(ctx context.Context, req RunBookRequest) (BookGenerationResult, error) {
	if err := s.validate(); err != nil { return BookGenerationResult{}, err }
	if req.BatchProjectID <= 0 || req.BookID <= 0 { return BookGenerationResult{}, ErrInvalid }
	if req.DirectorMode == "" { req.DirectorMode = DirectorNormal }
	if req.DirectorMode != DirectorNormal && req.DirectorMode != DirectorH3 { return BookGenerationResult{}, ErrInvalid }

	if existing, err := s.store.LatestBookRun(ctx, req.BatchProjectID, req.BookID); err == nil {
		if existing.Status == StatusCompleted || existing.Status == StatusRunning {
			return s.result(ctx, existing)
		}
	} else if !errors.Is(err, ErrNotFound) { return BookGenerationResult{}, err }

	book, err := s.store.GetBookForProject(ctx, req.BatchProjectID, req.BookID)
	if err != nil { return BookGenerationResult{}, err }
	now := s.now().UTC()
	run, err := s.store.CreateBookRun(ctx, BookRun{BatchProjectID: req.BatchProjectID, BookID: req.BookID, Status: StatusRunning, RequestID: strings.TrimSpace(req.RequestID), StartedAt: &now})
	if err != nil { return BookGenerationResult{}, err }

	scriptKey := PromptScript
	if req.PlotMode { scriptKey = PromptScriptPlotMode }
	script, err := s.executeProviderStage(ctx, run, book, StageScript, scriptKey, TextRequest{BookID: book.ID, Stage: StageScript, UserPrompt: book.OriginalText})
	if err != nil { return s.failRun(ctx, run, err) }

	hookText := ""
	if req.HookEnabled {
		hook, hookErr := s.executeProviderStage(ctx, run, book, StageHook, PromptHook, TextRequest{BookID: book.ID, Stage: StageHook, UserPrompt: script.OutputText})
		if hookErr != nil { return s.failRun(ctx, run, hookErr) }
		hookText = hook.OutputText
	} else {
		if _, skipErr := s.createSkippedStage(ctx, run, book, StageHook, PromptHook); skipErr != nil { return s.failRun(ctx, run, skipErr) }
	}

	directorKey := PromptDirector
	if req.DirectorMode == DirectorH3 { directorKey = PromptDirectorH3 }
	directorInput, _ := json.Marshal(map[string]any{
		"script": script.OutputText, "hook": hookText, "directorMode": req.DirectorMode,
		"matchAudio": req.MatchAudio, "audioDurationSec": req.AudioDurationSec,
	})
	director, err := s.executeProviderStage(ctx, run, book, StageDirector, directorKey, TextRequest{
		BookID: book.ID, Stage: StageDirector, UserPrompt: string(directorInput), DirectorMode: req.DirectorMode,
		MatchAudio: req.MatchAudio, AudioDurationSec: req.AudioDurationSec,
	})
	if err != nil { return s.failRun(ctx, run, err) }

	finalPrompt, err := s.resolver.Resolve(ctx, PromptFinal)
	if err != nil { return s.failRun(ctx, run, err) }
	compiled := s.compiler.Compile(FinalPromptInput{
		SystemPreset: finalPrompt.Content, Script: script.OutputText, Hook: hookText, Director: director.OutputText,
		ProcessingRules: req.ProcessingRules, KnowledgeBase: req.KnowledgeBase, ProjectConfig: req.ProjectConfig,
		UserConfig: req.UserConfig, ModelConfig: req.ModelConfig,
	})
	if _, err = s.completeLocalStage(ctx, run, book, StageFinalPrompt, finalPrompt, compiled, ""); err != nil { return s.failRun(ctx, run, err) }

	finished := s.now().UTC()
	run.Status, run.ErrorMessage, run.FinishedAt = StatusCompleted, "", &finished
	run, err = s.store.UpdateBookRun(ctx, run)
	if err != nil { return BookGenerationResult{}, err }
	return s.result(ctx, run)
}

func (s *Service) failRun(ctx context.Context, run BookRun, cause error) (BookGenerationResult, error) {
	finished := s.now().UTC()
	run.Status, run.ErrorMessage, run.FinishedAt = StatusFailed, safeError(cause), &finished
	updated, updateErr := s.store.UpdateBookRun(ctx, run)
	if updateErr != nil { return BookGenerationResult{}, updateErr }
	result, resultErr := s.result(ctx, updated)
	if resultErr != nil { return BookGenerationResult{}, resultErr }
	result.Error = safeError(cause)
	return result, cause
}

func (s *Service) nextAttempt(ctx context.Context, bookRunID int64, stage Stage) int {
	values, err := s.store.ListStageRuns(ctx, bookRunID)
	if err != nil { return 1 }
	max := 0
	for _, value := range values { if value.Stage == stage && value.Attempt > max { max = value.Attempt } }
	return max + 1
}

func (s *Service) executeProviderStage(ctx context.Context, run BookRun, book intake.Book, stage Stage, promptKey string, req TextRequest) (StageRun, error) {
	prompt, err := s.resolver.Resolve(ctx, promptKey)
	if err != nil { return StageRun{}, err }
	started := s.now().UTC()
	var snapshot string
	if stage == StageDirector {
		// Director retry must be able to read matchAudio/audioDurationSec directly.
		// Persist the frozen Director input itself instead of wrapping it as an
		// escaped string inside another request envelope.
		snapshot = req.UserPrompt
	} else {
		encoded, _ := json.Marshal(map[string]any{"bookId": book.ID, "source": book.OriginalText, "request": req})
		snapshot = string(encoded)
	}
	stageRun, err := s.store.CreateStageRun(ctx, StageRun{BookRunID: run.ID, BookID: book.ID, Stage: stage, Status: StatusRunning, Attempt: s.nextAttempt(ctx, run.ID, stage), RequestID: run.RequestID, PromptKey: prompt.Key, PromptVersion: prompt.Version, InputSnapshot: snapshot, StartedAt: &started})
	if err != nil { return StageRun{}, err }
	req.SystemPrompt = prompt.Content
	output, callErr := s.provider.Complete(ctx, req)
	finished := s.now().UTC()
	stageRun.FinishedAt = &finished
	if callErr != nil {
		stageRun.Status, stageRun.ErrorMessage = StatusFailed, safeError(callErr)
		stageRun, err = s.store.UpdateStageRun(ctx, stageRun)
		if err != nil { return StageRun{}, err }
		return stageRun, callErr
	}
	stageRun.Status, stageRun.OutputText, stageRun.ErrorMessage = StatusCompleted, strings.TrimSpace(output), ""
	stageRun, err = s.store.UpdateStageRun(ctx, stageRun)
	return stageRun, err
}

func (s *Service) createSkippedStage(ctx context.Context, run BookRun, book intake.Book, stage Stage, promptKey string) (StageRun, error) {
	prompt, err := s.resolver.Resolve(ctx, promptKey)
	if err != nil { return StageRun{}, err }
	now := s.now().UTC()
	return s.store.CreateStageRun(ctx, StageRun{BookRunID: run.ID, BookID: book.ID, Stage: stage, Status: StatusSkipped, Attempt: s.nextAttempt(ctx, run.ID, stage), RequestID: run.RequestID, PromptKey: prompt.Key, PromptVersion: prompt.Version, StartedAt: &now, FinishedAt: &now})
}

func (s *Service) completeLocalStage(ctx context.Context, run BookRun, book intake.Book, stage Stage, prompt Prompt, output, snapshot string) (StageRun, error) {
	now := s.now().UTC()
	return s.store.CreateStageRun(ctx, StageRun{BookRunID: run.ID, BookID: book.ID, Stage: stage, Status: StatusCompleted, Attempt: s.nextAttempt(ctx, run.ID, stage), RequestID: run.RequestID, PromptKey: prompt.Key, PromptVersion: prompt.Version, InputSnapshot: snapshot, OutputText: output, StartedAt: &now, FinishedAt: &now})
}

func (s *Service) RunBatch(ctx context.Context, req RunBatchRequest) (BatchGenerationResult, error) {
	books, err := s.store.ListBooksForProject(ctx, req.BatchProjectID)
	if err != nil { return BatchGenerationResult{}, err }
	result := BatchGenerationResult{BatchProjectID: req.BatchProjectID, Books: make([]BatchBookResult, 0, len(books))}
	var failures []string
	for _, book := range books {
		bookReq := RunBookRequest{BatchProjectID: req.BatchProjectID, BookID: book.ID, HookEnabled: req.HookEnabled, PlotMode: req.PlotMode, DirectorMode: req.DirectorMode, MatchAudio: req.MatchAudio, AudioDurationSec: req.AudioDurationSec, RequestID: req.RequestID}
		generated, runErr := s.RunBook(ctx, bookReq)
		item := BatchBookResult{BookID: book.ID, Run: &generated.Run}
		if runErr != nil { item.Error = safeError(runErr); result.Failed++; failures = append(failures, fmt.Sprintf("book %d: %s", book.ID, item.Error)) } else { result.Completed++ }
		result.Books = append(result.Books, item)
	}
	if len(failures) > 0 { return result, fmt.Errorf("%w: %s", ErrUnavailable, strings.Join(failures, "; ")) }
	return result, nil
}

func (s *Service) RetryStage(ctx context.Context, req RetryStageRequest) (BookGenerationResult, error) {
	if req.Stage != StageScript && req.Stage != StageHook && req.Stage != StageDirector && req.Stage != StageFinalPrompt { return BookGenerationResult{}, ErrInvalid }
	run, err := s.store.LatestBookRun(ctx, req.BatchProjectID, req.BookID)
	if err != nil { return BookGenerationResult{}, err }
	latest, err := s.store.LatestStageRun(ctx, run.ID, req.Stage)
	if err != nil { return BookGenerationResult{}, err }
	if latest.Status != StatusFailed { return BookGenerationResult{}, ErrConflict }
	book, err := s.store.GetBookForProject(ctx, req.BatchProjectID, req.BookID)
	if err != nil { return BookGenerationResult{}, err }
	run.Status, run.ErrorMessage, run.FinishedAt = StatusRunning, "", nil
	run.RequestID = strings.TrimSpace(req.RequestID)
	if run, err = s.store.UpdateBookRun(ctx, run); err != nil { return BookGenerationResult{}, err }

	var retryErr error
	switch req.Stage {
	case StageScript:
		_, retryErr = s.executeProviderStage(ctx, run, book, StageScript, latest.PromptKey, TextRequest{BookID: book.ID, Stage: StageScript, UserPrompt: book.OriginalText})
	case StageHook:
		script, e := s.store.LatestStageRun(ctx, run.ID, StageScript); if e != nil { retryErr = e; break }
		_, retryErr = s.executeProviderStage(ctx, run, book, StageHook, latest.PromptKey, TextRequest{BookID: book.ID, Stage: StageHook, UserPrompt: script.OutputText})
	case StageDirector:
		script, e := s.store.LatestStageRun(ctx, run.ID, StageScript); if e != nil { retryErr = e; break }
		hook, hookErr := s.store.LatestStageRun(ctx, run.ID, StageHook); hookText := ""; if hookErr == nil && hook.Status == StatusCompleted { hookText = hook.OutputText }
		var payload map[string]any
		_ = json.Unmarshal([]byte(latest.InputSnapshot), &payload)
		matchAudio, _ := payload["matchAudio"].(bool); audio, _ := payload["audioDurationSec"].(float64)
		mode := DirectorNormal; if latest.PromptKey == PromptDirectorH3 { mode = DirectorH3 }
		body, _ := json.Marshal(map[string]any{"script": script.OutputText, "hook": hookText, "directorMode": mode, "matchAudio": matchAudio, "audioDurationSec": audio})
		_, retryErr = s.executeProviderStage(ctx, run, book, StageDirector, latest.PromptKey, TextRequest{BookID: book.ID, Stage: StageDirector, UserPrompt: string(body), DirectorMode: mode, MatchAudio: matchAudio, AudioDurationSec: audio})
	case StageFinalPrompt:
		prompt, e := s.resolver.Resolve(ctx, latest.PromptKey); if e != nil { retryErr = e; break }
		script, e := s.store.LatestStageRun(ctx, run.ID, StageScript); if e != nil { retryErr = e; break }
		director, e := s.store.LatestStageRun(ctx, run.ID, StageDirector); if e != nil { retryErr = e; break }
		hook, hookErr := s.store.LatestStageRun(ctx, run.ID, StageHook); hookText := ""; if hookErr == nil && hook.Status == StatusCompleted { hookText = hook.OutputText }
		_, retryErr = s.completeLocalStage(ctx, run, book, StageFinalPrompt, prompt, s.compiler.Compile(FinalPromptInput{SystemPreset: prompt.Content, Script: script.OutputText, Hook: hookText, Director: director.OutputText}), latest.InputSnapshot)
	}
	if retryErr != nil { return s.failRun(ctx, run, retryErr) }

	// Retry is intentionally stage-scoped. Mark completed only when all pipeline stages are now terminal-success/skipped.
	allDone := true
	for _, stage := range []Stage{StageScript, StageHook, StageDirector, StageFinalPrompt} {
		value, e := s.store.LatestStageRun(ctx, run.ID, stage)
		if e != nil || (value.Status != StatusCompleted && value.Status != StatusSkipped) { allDone = false; break }
	}
	if allDone { finished := s.now().UTC(); run.Status, run.FinishedAt = StatusCompleted, &finished }
	run, err = s.store.UpdateBookRun(ctx, run)
	if err != nil { return BookGenerationResult{}, err }
	return s.result(ctx, run)
}

func (s *Service) ProjectSummary(ctx context.Context, projectID int64) (ProjectSummary, error) {
	books, err := s.store.ListBooksForProject(ctx, projectID)
	if err != nil { return ProjectSummary{}, err }
	out := ProjectSummary{BatchProjectID: projectID, Books: make([]BookGenerationSummary, 0, len(books))}
	for _, book := range books {
		item := BookGenerationSummary{BookID: book.ID, Stages: map[Stage]StageRun{}}
		run, runErr := s.store.LatestBookRun(ctx, projectID, book.ID)
		if runErr != nil {
			if errors.Is(runErr, ErrNotFound) { out.Pending++; out.Books = append(out.Books, item); continue }
			return ProjectSummary{}, runErr
		}
		item.Run = &run
		stages, stageErr := s.store.ListStageRuns(ctx, run.ID); if stageErr != nil { return ProjectSummary{}, stageErr }
		for _, value := range stages { current, ok := item.Stages[value.Stage]; if !ok || value.Attempt >= current.Attempt { item.Stages[value.Stage] = value } }
		switch run.Status { case StatusCompleted: out.Completed++; case StatusFailed: out.Failed++; case StatusRunning: out.Running++; default: out.Pending++ }
		out.Books = append(out.Books, item)
	}
	return out, nil
}
