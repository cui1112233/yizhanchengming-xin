package generation

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
)

type RuntimeBookRunStore interface {
	RuntimeBookRun(context.Context, task9runtime.Execution) (BookRun, RuntimeExecutionInput, error)
	GetBookForProject(context.Context, int64, int64) (intake.Book, error)
	LatestAudioMeasurement(context.Context, int64, int64) (AudioMeasurement, error)
	ResolvePrompt(context.Context, string) (Prompt, error)
	ListStageRuns(context.Context, int64) ([]StageRun, error)
	CreateStageRunFenced(context.Context, task9runtime.Execution, StageRun) (StageRun, error)
	UpdateStageRunFenced(context.Context, task9runtime.Execution, StageRun) (StageRun, error)
}

type RuntimeExecutionInput struct {
	Request         RunBookRequest
	Action          string
	RetryStage      Stage
	SourceBookRunID int64
}

type RuntimeExecutor struct {
	store    RuntimeBookRunStore
	provider Provider
	resolver *PromptResolver
	compiler FinalPromptCompiler
	now      Clock
}

func NewRuntimeExecutor(store RuntimeBookRunStore, provider Provider, now Clock) *RuntimeExecutor {
	if now == nil {
		now = time.Now
	}
	return &RuntimeExecutor{store: store, provider: provider, resolver: NewPromptResolver(store), compiler: FinalPromptCompiler{}, now: now}
}

func (e *RuntimeExecutor) Execute(ctx context.Context, execution task9runtime.Execution) error {
	if e == nil || e.store == nil || e.provider == nil || e.resolver == nil {
		return NewRuntimeOutcomeError(ErrUnavailable)
	}
	run, input, err := e.store.RuntimeBookRun(ctx, execution)
	if err != nil {
		return runtimeExecutionError(ctx, err)
	}
	if run.ID != execution.BookRunID || run.BatchProjectID <= 0 || run.BookID <= 0 || run.Status != StatusRunning {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	if input.Request.BatchProjectID != run.BatchProjectID || input.Request.BookID != run.BookID {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	if input.Action == "" {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	if input.Action != task9runtime.GenerationActionFull && input.Action != task9runtime.GenerationActionStageRetry {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	if input.Request.DirectorMode != DirectorNormal && input.Request.DirectorMode != DirectorH3 {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	if input.Action == task9runtime.GenerationActionFull && (input.RetryStage != "" || input.SourceBookRunID != 0) {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	book, err := e.store.GetBookForProject(ctx, run.BatchProjectID, run.BookID)
	if err != nil {
		return runtimeExecutionError(ctx, err)
	}
	if book.ID != run.BookID {
		return NewRuntimeOutcomeError(ErrInvalid)
	}
	measurement, err := e.authoritativeAudio(ctx, &input.Request)
	if err != nil {
		return runtimeExecutionError(ctx, err)
	}

	completed := map[Stage]StageRun{}
	start := StageScript
	if input.Action == task9runtime.GenerationActionStageRetry {
		if !validRuntimeStage(input.RetryStage) || input.SourceBookRunID <= 0 || input.SourceBookRunID == run.ID || (input.RetryStage == StageHook && !input.Request.HookEnabled) {
			return NewRuntimeOutcomeError(ErrInvalid)
		}
		completed, err = e.copyRetryPrerequisites(ctx, execution, run, input)
		if err != nil {
			return runtimeExecutionError(ctx, err)
		}
		start = input.RetryStage
	}

	stages := []Stage{StageScript, StageHook, StageDirector, StageFinalPrompt}
	startIndex := 0
	for i, stage := range stages {
		if stage == start {
			startIndex = i
			break
		}
	}
	for _, stage := range stages[startIndex:] {
		if err := ctx.Err(); err != nil {
			return err
		}
		var value StageRun
		switch stage {
		case StageScript:
			key := PromptScript
			if input.Request.PlotMode {
				key = PromptScriptPlotMode
			}
			value, err = e.providerStage(ctx, execution, run, book, stage, key, TextRequest{BookID: book.ID, Stage: stage, UserPrompt: book.OriginalText}, map[string]any{"book_id": book.ID})
		case StageHook:
			if !input.Request.HookEnabled {
				value, err = e.skippedStage(ctx, execution, run, book, stage, PromptHook)
			} else {
				script, ok := completed[StageScript]
				if !ok || script.Status != StatusCompleted {
					return NewRuntimeOutcomeError(ErrInvalid)
				}
				value, err = e.providerStage(ctx, execution, run, book, stage, PromptHook, TextRequest{BookID: book.ID, Stage: stage, UserPrompt: script.OutputText}, map[string]any{"upstream_stage_run_id": script.ID})
			}
		case StageDirector:
			script, ok := completed[StageScript]
			if !ok || script.Status != StatusCompleted {
				return NewRuntimeOutcomeError(ErrInvalid)
			}
			hook := completed[StageHook]
			hookText := ""
			if hook.Status == StatusCompleted {
				hookText = hook.OutputText
			} else if hook.Status != StatusSkipped {
				return NewRuntimeOutcomeError(ErrInvalid)
			}
			key := PromptDirector
			if input.Request.DirectorMode == DirectorH3 {
				key = PromptDirectorH3
			}
			payload, _ := json.Marshal(map[string]any{"script": script.OutputText, "hook": hookText, "directorMode": input.Request.DirectorMode, "matchAudio": input.Request.MatchAudio, "audioDurationSec": input.Request.AudioDurationSec, "shotDurationLimitSec": input.Request.ShotDurationLimitSec})
			snapshot := map[string]any{"script_stage_run_id": script.ID, "hook_stage_run_id": hook.ID, "director_mode": input.Request.DirectorMode, "match_audio": input.Request.MatchAudio, "shot_duration_limit_sec": input.Request.ShotDurationLimitSec}
			if measurement.ID > 0 {
				snapshot["audio_measurement_id"] = measurement.ID
				snapshot["audio_duration_ms"] = measurement.DurationMS
			}
			value, err = e.providerStage(ctx, execution, run, book, stage, key, TextRequest{BookID: book.ID, Stage: stage, UserPrompt: string(payload), DirectorMode: input.Request.DirectorMode, MatchAudio: input.Request.MatchAudio, AudioDurationSec: input.Request.AudioDurationSec, ShotDurationLimitSec: input.Request.ShotDurationLimitSec}, snapshot)
		case StageFinalPrompt:
			value, err = e.finalPromptStage(ctx, execution, run, book, input.Request, completed)
		}
		if err != nil {
			return runtimeExecutionError(ctx, err)
		}
		completed[stage] = value
	}
	return nil
}

func runtimeExecutionError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, task9runtime.ErrStaleExecution) || errors.Is(err, taskruntime.ErrLeaseNotOwner) {
		return err
	}
	if ctx != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
	}
	var safe *RuntimeOutcomeError
	if errors.As(err, &safe) {
		return err
	}
	return NewRuntimeOutcomeError(err)
}

func validRuntimeStage(stage Stage) bool {
	switch stage {
	case StageScript, StageHook, StageDirector, StageFinalPrompt:
		return true
	default:
		return false
	}
}

func (e *RuntimeExecutor) authoritativeAudio(ctx context.Context, req *RunBookRequest) (AudioMeasurement, error) {
	if !req.MatchAudio {
		return AudioMeasurement{}, nil
	}
	limit, err := normalizeShotDurationLimit(req.ShotDurationLimitSec)
	if err != nil {
		return AudioMeasurement{}, err
	}
	measurement, err := e.store.LatestAudioMeasurement(ctx, req.BatchProjectID, req.BookID)
	if err != nil || measurement.ID <= 0 || measurement.DurationMS <= 0 {
		if err != nil && !errors.Is(err, ErrNotFound) {
			return AudioMeasurement{}, err
		}
		return AudioMeasurement{}, ErrAudioMeasurementRequired
	}
	req.AudioDurationSec = float64(measurement.DurationMS) / 1000
	req.ShotDurationLimitSec = limit
	return measurement, nil
}

func (e *RuntimeExecutor) copyRetryPrerequisites(ctx context.Context, execution task9runtime.Execution, run BookRun, input RuntimeExecutionInput) (map[Stage]StageRun, error) {
	values, err := e.store.ListStageRuns(ctx, input.SourceBookRunID)
	if err != nil {
		return nil, runtimeExecutionError(ctx, err)
	}
	latest := map[Stage]StageRun{}
	for _, value := range values {
		if value.BookRunID != input.SourceBookRunID || value.BookID != run.BookID || !validRuntimeStage(value.Stage) {
			continue
		}
		if current, ok := latest[value.Stage]; !ok || value.Attempt > current.Attempt || (value.Attempt == current.Attempt && value.ID > current.ID) {
			latest[value.Stage] = value
		}
	}
	target, ok := latest[input.RetryStage]
	if !ok || target.Status != StatusFailed {
		return nil, NewRuntimeOutcomeError(ErrInvalid)
	}
	out := map[Stage]StageRun{}
	for _, stage := range []Stage{StageScript, StageHook, StageDirector, StageFinalPrompt} {
		if stage == input.RetryStage {
			break
		}
		source, ok := latest[stage]
		if !ok || (source.Status != StatusCompleted && source.Status != StatusSkipped) {
			return nil, NewRuntimeOutcomeError(ErrInvalid)
		}
		if stage == StageHook && ((input.Request.HookEnabled && source.Status != StatusCompleted) || (!input.Request.HookEnabled && source.Status != StatusSkipped)) {
			return nil, NewRuntimeOutcomeError(ErrInvalid)
		}
		now := e.now().UTC()
		snapshot, _ := json.Marshal(map[string]any{"source_stage_run_id": source.ID})
		copied, err := e.store.CreateStageRunFenced(ctx, execution, StageRun{BookRunID: run.ID, Stage: stage, Status: source.Status, RequestID: run.RequestID, PromptKey: source.PromptKey, PromptVersion: source.PromptVersion, InputSnapshot: string(snapshot), OutputText: source.OutputText, ValidationResult: source.ValidationResult, StartedAt: &now, FinishedAt: &now})
		if err != nil {
			return nil, runtimeExecutionError(ctx, err)
		}
		out[stage] = copied
	}
	return out, nil
}

func (e *RuntimeExecutor) providerStage(ctx context.Context, execution task9runtime.Execution, run BookRun, book intake.Book, stage Stage, promptKey string, req TextRequest, snapshotFields map[string]any) (StageRun, error) {
	prompt, err := e.resolver.Resolve(ctx, promptKey)
	if err != nil {
		return StageRun{}, runtimeExecutionError(ctx, err)
	}
	started := e.now().UTC()
	snapshotFields["prompt_key"] = prompt.Key
	snapshotFields["prompt_version"] = prompt.Version
	snapshot, _ := json.Marshal(snapshotFields)
	stageRun, err := e.store.CreateStageRunFenced(ctx, execution, StageRun{BookRunID: run.ID, Stage: stage, Status: StatusRunning, RequestID: run.RequestID, PromptKey: prompt.Key, PromptVersion: prompt.Version, InputSnapshot: string(snapshot), StartedAt: &started})
	if err != nil {
		return StageRun{}, runtimeExecutionError(ctx, err)
	}
	req.SystemPrompt = prompt.Content
	output, callErr := e.provider.Complete(ctx, req)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return StageRun{}, ctxErr
	}
	if callErr == nil {
		callErr = validateProviderOutput(req, output)
	}
	if callErr == nil && stage == StageDirector && req.MatchAudio {
		targetMS := int64(math.Round(req.AudioDurationSec * 1000))
		limit, limitErr := normalizeShotDurationLimit(req.ShotDurationLimitSec)
		if limitErr != nil {
			callErr = limitErr
		} else {
			var result TimelineValidationResult
			output, result, callErr = validateAndRepairDirectorOutput(output, targetMS, limit*1000)
			stageRun.ValidationResult = validationJSON(callErr == nil, req, result, callErr)
		}
	}
	if callErr != nil && ctx.Err() != nil {
		return StageRun{}, ctx.Err()
	}
	if callErr != nil && errors.Is(callErr, context.Canceled) {
		return StageRun{}, context.Canceled
	}
	finished := e.now().UTC()
	stageRun.FinishedAt = &finished
	if callErr != nil {
		outcome := OutcomeForError(callErr)
		stageRun.Status, stageRun.ErrorMessage, stageRun.ErrorCode = StatusFailed, outcome.Message, outcome.Code
		if stage == StageDirector && req.MatchAudio && stageRun.ValidationResult == "" {
			stageRun.ValidationResult = validationJSON(false, req, TimelineValidationResult{}, callErr)
		}
		if _, err := e.store.UpdateStageRunFenced(ctx, execution, stageRun); err != nil {
			return StageRun{}, runtimeExecutionError(ctx, err)
		}
		return StageRun{}, runtimeExecutionError(ctx, callErr)
	}
	stageRun.Status, stageRun.OutputText, stageRun.ErrorMessage, stageRun.ErrorCode = StatusCompleted, strings.TrimSpace(output), "", ""
	updated, err := e.store.UpdateStageRunFenced(ctx, execution, stageRun)
	return updated, runtimeExecutionError(ctx, err)
}

func (e *RuntimeExecutor) skippedStage(ctx context.Context, execution task9runtime.Execution, run BookRun, book intake.Book, stage Stage, promptKey string) (StageRun, error) {
	now := e.now().UTC()
	snapshot, _ := json.Marshal(map[string]any{"book_id": book.ID, "disabled": true})
	created, err := e.store.CreateStageRunFenced(ctx, execution, StageRun{BookRunID: run.ID, Stage: stage, Status: StatusSkipped, RequestID: run.RequestID, PromptKey: promptKey, InputSnapshot: string(snapshot), StartedAt: &now, FinishedAt: &now})
	return created, runtimeExecutionError(ctx, err)
}

func (e *RuntimeExecutor) finalPromptStage(ctx context.Context, execution task9runtime.Execution, run BookRun, book intake.Book, req RunBookRequest, completed map[Stage]StageRun) (StageRun, error) {
	script, scriptOK := completed[StageScript]
	hook, hookOK := completed[StageHook]
	director, directorOK := completed[StageDirector]
	if !scriptOK || script.Status != StatusCompleted || !hookOK || (hook.Status != StatusCompleted && hook.Status != StatusSkipped) || !directorOK || director.Status != StatusCompleted {
		return StageRun{}, NewRuntimeOutcomeError(ErrInvalid)
	}
	prompt, err := e.resolver.Resolve(ctx, PromptFinal)
	if err != nil {
		return StageRun{}, runtimeExecutionError(ctx, err)
	}
	processingRules, err := e.resolveOptionalPrompt(ctx, req.ProcessingRules)
	if err != nil {
		return StageRun{}, runtimeExecutionError(ctx, err)
	}
	knowledge, err := e.resolveOptionalPrompt(ctx, req.KnowledgeBase)
	if err != nil {
		return StageRun{}, runtimeExecutionError(ctx, err)
	}
	hookText := ""
	if hook.Status == StatusCompleted {
		hookText = hook.OutputText
	}
	compiled := e.compiler.Compile(FinalPromptInput{SystemPreset: prompt.Content, Script: script.OutputText, Hook: hookText, Director: director.OutputText, ProcessingRules: processingRules.Content, KnowledgeBase: knowledge.Content, ProjectConfig: req.ProjectConfig, UserConfig: req.UserConfig, ModelConfig: req.ModelConfig})
	now := e.now().UTC()
	snapshotFields := map[string]any{"script_stage_run_id": script.ID, "hook_stage_run_id": hook.ID, "director_stage_run_id": director.ID, "prompt_key": prompt.Key, "prompt_version": prompt.Version, "book_id": book.ID}
	if processingRules.Key != "" {
		snapshotFields["processing_rules_prompt_key"] = processingRules.Key
		snapshotFields["processing_rules_prompt_version"] = processingRules.Version
	}
	if knowledge.Key != "" {
		snapshotFields["knowledge_prompt_key"] = knowledge.Key
		snapshotFields["knowledge_prompt_version"] = knowledge.Version
	}
	snapshot, _ := json.Marshal(snapshotFields)
	created, err := e.store.CreateStageRunFenced(ctx, execution, StageRun{BookRunID: run.ID, Stage: StageFinalPrompt, Status: StatusCompleted, RequestID: run.RequestID, PromptKey: prompt.Key, PromptVersion: prompt.Version, InputSnapshot: string(snapshot), OutputText: compiled, StartedAt: &now, FinishedAt: &now})
	return created, runtimeExecutionError(ctx, err)
}

func (e *RuntimeExecutor) resolveOptionalPrompt(ctx context.Context, key string) (Prompt, error) {
	if strings.TrimSpace(key) == "" {
		return Prompt{}, nil
	}
	return e.resolver.Resolve(ctx, key)
}
