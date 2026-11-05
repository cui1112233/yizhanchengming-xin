package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

type ProductionState string

const (
	ProductionQueued    ProductionState = "queued"
	ProductionRunning   ProductionState = "running"
	ProductionSucceeded ProductionState = "succeeded"
	ProductionFailed    ProductionState = "failed"
	ProductionCancelled ProductionState = "cancelled"
)

// ResubmitWithoutReceiptAfter is the grace period a queued/running task may stay
// without a provider task id before it is treated as an interrupted submission.
const ResubmitWithoutReceiptAfter = 2 * time.Minute

// SupersededWithoutReceiptMarker is the durable marker written on the cancelled
// task. Its presence is the "this video was already re-submitted once" guard.
const SupersededWithoutReceiptMarker = "提交超时未取得 provider 任务号，已自动重新投递"

// supersededProductionTask records one task whose submission never returned a
// provider id, so the retry can cancel it with an auditable reason.
type supersededProductionTask struct {
	jobID  string
	task   ProductionTask
	waited time.Duration
}

type FrozenVideoModel struct {
	ID                 string `json:"id"`
	MaxDuration        int    `json:"maxDuration"`
	MaxReferenceImages int    `json:"maxReferenceImages"`
}

type ProviderTaskRef struct {
	ProviderTaskID           string          `json:"providerTaskId,omitempty"`
	State                    ProductionState `json:"state"`
	MediaURL                 string          `json:"mediaUrl,omitempty"`
	ErrorMessage             string          `json:"errorMessage,omitempty"`
	RequestedDurationSeconds float64         `json:"requestedDurationSeconds,omitempty"`
	ActualDurationSeconds    float64         `json:"actualDurationSeconds,omitempty"`
}

type ProductionTask struct {
	ID                       string          `json:"id"`
	VideoID                  string          `json:"videoId"`
	OpeningVariantIndex      int             `json:"openingVariantIndex,omitempty"`
	Provider                 string          `json:"provider,omitempty"`
	Status                   ProductionState `json:"status"`
	Attempt                  int             `json:"attempt"`
	FinalPromptHash          string          `json:"finalPromptHash"`
	CompiledPrompt           string          `json:"-"`
	CompilationID            string          `json:"compilationId,omitempty"`
	CompilationSegmentKey    string          `json:"compilationSegmentKey,omitempty"`
	CompileTrace             *H3CompileTrace `json:"compileTrace,omitempty"`
	ReferenceImageURLs       []string        `json:"referenceImageUrls,omitempty"`
	DowngradedAssetIDs       []string        `json:"downgradedAssetIds,omitempty"`
	TargetDurationSeconds    float64         `json:"targetDurationSeconds,omitempty"`
	RequestedDurationSeconds float64         `json:"requestedDurationSeconds,omitempty"`
	ActualDurationSeconds    float64         `json:"actualDurationSeconds,omitempty"`
	ProviderTaskID           string          `json:"providerTaskId,omitempty"`
	MediaURL                 string          `json:"mediaUrl,omitempty"`
	ErrorMessage             string          `json:"errorMessage,omitempty"`
	CreatedAt                time.Time       `json:"createdAt"`
	UpdatedAt                time.Time       `json:"updatedAt"`
}

type ProductionJob struct {
	ID                 string           `json:"id"`
	Owner              string           `json:"-"`
	BatchID            string           `json:"batchId"`
	BookID             string           `json:"bookId"`
	RequestID          string           `json:"requestId"`
	DirectorRevisionID string           `json:"directorRevisionId"`
	Status             ProductionState  `json:"status"`
	Tasks              []ProductionTask `json:"tasks"`
	CreatedAt          time.Time        `json:"createdAt"`
	UpdatedAt          time.Time        `json:"updatedAt"`
}

type BatchStatus struct {
	BatchID string          `json:"batchId"`
	Jobs    []ProductionJob `json:"jobs"`
}

type ProductionCancelSkip struct {
	TaskID string `json:"taskId"`
	Reason string `json:"reason"`
}

// BatchCancellation is an honest cancellation receipt. Only tasks whose
// provider accepted an actual cancellation move to cancelled; other active
// tasks remain unchanged and are returned with their reason.
type BatchCancellation struct {
	BatchID          string                 `json:"batchId"`
	RequestID        string                 `json:"requestId"`
	CancelledTaskIDs []string               `json:"cancelledTaskIds"`
	Skipped          []ProductionCancelSkip `json:"skipped,omitempty"`
}

type PromptResolver interface {
	Compile(context.Context, string, string, string, string) (FinalPrompt, error)
}

type referenceImageLimitPromptResolver interface {
	CompileWithReferenceImageLimit(context.Context, string, string, string, string, int) (FinalPrompt, error)
}

type strictPromptResolver interface {
	CompileForProduction(context.Context, string, string, string, string) (FinalPrompt, error)
}

type ProductionAdapter interface {
	Submit(context.Context, FrozenVideoModel, FinalPrompt) (ProviderTaskRef, error)
}

// ProductionRepository is separate from the general Store so that previous V11
// slices do not accidentally gain mutable production authority.
type ProductionRepository interface {
	FindProductionJob(context.Context, string, string, string, string) (ProductionJob, error)
	CreateProductionJob(context.Context, ProductionJob) (ProductionJob, error)
	UpdateProductionTask(context.Context, string, string, string, ProductionTask) (ProductionJob, error)
	// HideProductionTask removes a completed candidate from the book library
	// without deleting the provider artifact or its durable production audit.
	HideProductionTask(context.Context, string, string, string, string, string) error
	ListProductionJobs(context.Context, string, string) ([]ProductionJob, error)
}

type ProductionOptions struct {
	Force         bool
	VideoID       string
	CompilationID string
	// StripReferenceImages drops reference images for one explicit retry. It is
	// only requested by the moderation-rejection retry path and is gated by an
	// opt-in switch, because a text-only rerun has not been confirmed against
	// every provider yet.
	StripReferenceImages bool
}

type h3CompilationPromptResolver interface {
	CompileH3ForProduction(context.Context, string, string, string, string, string, int) (FinalPrompt, error)
}

type openingVariantPromptResolver interface {
	CompileForOpeningVariant(context.Context, string, string, string, string, int, int) (FinalPrompt, error)
}

type ProductionService struct {
	Store            Store
	Compiler         PromptResolver
	Adapter          ProductionAdapter
	Poller           ProductionPoller
	ProviderRegistry VideoProviderRegistry
	LocalExecutor    *LocalExecutorVideoAdapter
	Enabled          bool
	Model            FrozenVideoModel
}

func (s *ProductionService) resolveProvider(ctx context.Context, owner, provider string) (ProductionAdapter, FrozenVideoModel, error) {
	provider = normalizeVideoProvider(provider)
	if provider == VideoProviderDoubaoLocal {
		if s.LocalExecutor == nil || s.LocalExecutor.Client == nil {
			return nil, FrozenVideoModel{}, fmt.Errorf("%w: Doubao local executor is offline", ErrUnavailable)
		}
		if err := s.LocalExecutor.EnsureAvailable(ctx, owner); err != nil {
			return nil, FrozenVideoModel{}, err
		}
		model := s.Model
		// The local executor has its own Doubao model identity; never inherit
		// the personal Yadi model from static server configuration.
		model.ID = "doubao-seedance"
		if model.MaxDuration <= 0 {
			model.MaxDuration = 15
		}
		return nil, model, nil
	}
	if provider == VideoProviderAutoDLH3 {
		if s.ProviderRegistry == nil {
			return nil, FrozenVideoModel{}, wrapNotConfigured(fmt.Errorf("%w: AutoDL H3 provider registry is unavailable", ErrUnavailable))
		}
		cfg, err := s.ProviderRegistry.Resolve(ctx, owner, provider)
		if err != nil {
			return nil, FrozenVideoModel{}, err
		}
		model := s.Model
		model.ID = cfg.Model
		if model.ID == "" {
			model.ID = AutoDLH3Model
		}
		if model.MaxDuration <= 0 {
			model.MaxDuration = 15
		}
		adapter := &AutoDLH3VideoAdapter{
			CreateURL: cfg.CreateURL, TasksURL: cfg.TasksURL,
			APIKey: cfg.APIKey, Model: model.ID, WorkflowID: cfg.WorkflowID,
			FallbackReferenceImageURL: AutoDLH3FallbackImageURL(),
		}
		if err := adapter.Validate(); err != nil {
			return nil, FrozenVideoModel{}, err
		}
		return adapter, model, nil
	}
	if provider == VideoProviderYFAISeedance {
		if s.ProviderRegistry == nil {
			return nil, FrozenVideoModel{}, wrapNotConfigured(fmt.Errorf("%w: YFAI Seedance provider registry is unavailable", ErrUnavailable))
		}
		cfg, err := s.ProviderRegistry.Resolve(ctx, owner, provider)
		if err != nil {
			return nil, FrozenVideoModel{}, err
		}
		model := s.Model
		model.ID = cfg.Model
		model.MaxDuration = 15
		adapter := &YFAISeedanceAdapter{BaseURL: cfg.CreateURL, APIKey: cfg.APIKey, Model: cfg.Model}
		if err := adapter.Validate(); err != nil {
			return nil, FrozenVideoModel{}, err
		}
		return adapter, model, nil
	}
	if provider != VideoProviderPersonalAPI {
		return nil, FrozenVideoModel{}, fmt.Errorf("%w: unsupported video provider", ErrInvalid)
	}
	if s.ProviderRegistry != nil {
		cfg, err := s.ProviderRegistry.Resolve(ctx, owner, provider)
		if err == nil {
			model := s.Model
			model.ID = cfg.Model
			if model.MaxDuration <= 0 {
				model.MaxDuration = 15
			}
			adapter := &YadiVideoAdapter{
				CreateURL: cfg.CreateURL, TasksURL: cfg.TasksURL, ResultURL: cfg.ResultURL,
				APIKey: cfg.APIKey, Model: cfg.Model,
			}
			if err := adapter.Validate(); err != nil {
				return nil, FrozenVideoModel{}, err
			}
			return adapter, model, nil
		}
		if s.Adapter == nil {
			return nil, FrozenVideoModel{}, err
		}
	}
	if s.Adapter == nil {
		return nil, FrozenVideoModel{}, fmt.Errorf("%w: personal video provider is unavailable", ErrUnavailable)
	}
	return s.Adapter, s.Model, nil
}

func (s *ProductionService) repository() (ProductionRepository, error) {
	repository, ok := s.Store.(ProductionRepository)
	if !ok {
		return nil, ErrUnavailable
	}
	return repository, nil
}

func productionError(err error) string {
	if err == nil {
		return ""
	}
	message := stripProductionErrorPrefixes(strings.TrimSpace(err.Error()))
	// The column is VARCHAR(255). Truncate well below the byte limit so UTF-8
	// multibyte characters never exceed it and the INSERT never fails with
	// Error 1406 (Data too long), which used to cause an infinite retry loop.
	if len(message) > 200 {
		// Byte slicing can land inside a multibyte rune, which stores invalid
		// UTF-8 and shows up as broken characters in the UI. Back off to the
		// closest rune boundary while keeping the same 200-byte budget.
		limit := 200
		for limit > 0 && !utf8.RuneStart(message[limit]) {
			limit--
		}
		return message[:limit] + "..."
	}
	return message
}

// stripProductionErrorPrefixes removes internal sentinel text before a stage
// error reaches the UI. The stage panel must show the real reason ("换开头变体
// 未全部生成") instead of a fake "revision conflict" wrapper; the API layer
// already does the same for JSON responses.
func stripProductionErrorPrefixes(message string) string {
	prefixes := []string{
		ErrConflict.Error() + ": ",
		ErrConflict.Error() + ":",
		ErrUnavailable.Error() + ": ",
		ErrConfigRequired.Error() + ": ",
		ErrInvalid.Error() + ": ",
		ErrUpstream.Error() + ": ",
		ErrNotFound.Error() + ": ",
	}
	for {
		trimmed := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(message, prefix) {
				message = strings.TrimSpace(strings.TrimPrefix(message, prefix))
				trimmed = true
			}
		}
		if !trimmed {
			return message
		}
	}
}

func normalizeProductionState(value ProductionState) ProductionState {
	switch value {
	case ProductionQueued, ProductionRunning, ProductionSucceeded, ProductionFailed, ProductionCancelled:
		return value
	default:
		return ProductionRunning
	}
}

func (s *ProductionService) SubmitBookProduction(ctx context.Context, owner, batchID, bookID, requestID string) (ProductionJob, error) {
	return s.SubmitBookProductionWithOptions(ctx, owner, batchID, bookID, requestID, VideoProviderPersonalAPI, ProductionOptions{})
}

func (s *ProductionService) SubmitBookProductionWithProvider(ctx context.Context, owner, batchID, bookID, requestID, provider string) (ProductionJob, error) {
	return s.SubmitBookProductionWithOptions(ctx, owner, batchID, bookID, requestID, provider, ProductionOptions{})
}

// stripReferenceImagesFromPrompt removes every carrier a provider adapter could
// read a reference image from, so a moderation retry is genuinely text-only.
func stripReferenceImagesFromPrompt(prompt FinalPrompt) FinalPrompt {
	prompt.ReferenceImageURLs = nil
	prompt.DowngradedAssetIDs = nil
	delete(prompt.EffectiveSettings.Values, "imageUrls")
	delete(prompt.EffectiveSettings.Values, "referenceImages")
	return prompt
}

// SubmitBookProductionWithOptions keeps normal generation idempotent while an
// explicit retry may target one failed VIDEO without changing its primary media.
func (s *ProductionService) SubmitBookProductionWithOptions(ctx context.Context, owner, batchID, bookID, requestID, provider string, options ProductionOptions) (ProductionJob, error) {
	// The gate deliberately comes before repository/compiler/provider work.
	if s == nil || !s.Enabled {
		return ProductionJob{}, fmt.Errorf("%w: production is not enabled", ErrUnavailable)
	}
	if s.Store == nil || s.Compiler == nil {
		return ProductionJob{}, ErrUnavailable
	}
	if strings.TrimSpace(requestID) == "" {
		return ProductionJob{}, fmt.Errorf("%w: request id is required", ErrInvalid)
	}
	provider = normalizeVideoProvider(provider)
	repository, err := s.repository()
	if err != nil {
		return ProductionJob{}, err
	}
	if existing, findErr := repository.FindProductionJob(ctx, owner, batchID, bookID, requestID); findErr == nil {
		return existing, nil
	} else if findErr != ErrNotFound {
		return ProductionJob{}, findErr
	}

	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ProductionJob{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return ProductionJob{}, err
	}
	// The effective per-book selection is authoritative at submission time:
	// a caller may still pass the batch engine (or a preset default), but the
	// book's own videoModelId/videoProvider must decide which provider submits.
	// This is the seam that used to let a book configured for YFAI Seedance be
	// submitted through AutoDL H3 because the batch patch won.
	effectiveSettings := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	selectedModel := strings.TrimSpace(rawString(effectiveSettings, "videoModelId", ""))
	batchModel := strings.TrimSpace(rawString(batch.SettingsState.Patch, "videoModelId", ""))
	if selectedModel != "" {
		selectedProvider := normalizeVideoProvider(VideoProviderForModel(selectedModel, rawString(effectiveSettings, "videoProvider", provider)))
		if selectedProvider != provider || selectedModel != batchModel {
			log.Printf("batch factory video submit: book=%s provider=%s model=%s batchModel=%s bookOverride=true", bookID, selectedProvider, selectedModel, batchModel)
		}
		provider = selectedProvider
	}

	// A retried VIDEO keeps its identity while the engine selected for the book
	// may have changed. Record the switch (no credentials) so a release trace
	// can prove the stored provider was not silently reused.
	if current, statusErr := s.GetBatchStatus(ctx, owner, batchID); statusErr == nil {
		var storedProvider, storedVideoID string
		var latest time.Time
		for _, job := range current.Jobs {
			for _, task := range job.Tasks {
				if job.BookID != bookID || strings.TrimSpace(task.Provider) == "" {
					continue
				}
				if storedProvider == "" || task.UpdatedAt.After(latest) {
					latest, storedProvider, storedVideoID = task.UpdatedAt, normalizeVideoProvider(task.Provider), task.VideoID
				}
			}
		}
		if storedProvider != "" && storedProvider != provider {
			log.Printf("video resubmit: video=%s storedProvider=%s effectiveProvider=%s model=%s switched=true", storedVideoID, storedProvider, provider, selectedModel)
		}
	}

	adapter, model, providerErr := s.resolveProvider(ctx, owner, provider)
	if providerErr != nil {
		return ProductionJob{}, providerErr
	}
	if provider != VideoProviderDoubaoLocal && adapter == nil {
		return ProductionJob{}, ErrUnavailable
	}
	if strings.TrimSpace(model.ID) == "" {
		return ProductionJob{}, fmt.Errorf("%w: frozen video model is required", ErrInvalid)
	}
	if book.DirectorRevision == nil || book.DirectorRevision.ID == "" {
		return ProductionJob{}, fmt.Errorf("%w: active Director revision is required", ErrConflict)
	}
	if len(book.Videos) == 0 {
		return ProductionJob{}, fmt.Errorf("%w: no active VIDEOs to produce", ErrConflict)
	}
	if book.DirectorRevision.Output.H3Director != nil && strings.TrimSpace(options.CompilationID) == "" {
		repository, ok := s.Store.(H3Repository)
		if !ok {
			return ProductionJob{}, fmt.Errorf("%w: H3 compilation repository is unavailable", ErrUnavailable)
		}
		latest, latestErr := repository.LatestH3VideoCompilation(ctx, owner, batchID, bookID, book.DirectorRevision.ID)
		if latestErr != nil {
			return ProductionJob{}, fmt.Errorf("%w: active H3 compilation is required", ErrConflict)
		}
		options.CompilationID = latest.ID
	}
	priorJobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return ProductionJob{}, err
	}
	blockedVideos := map[string]bool{}
	resubmittedVideos := map[string]bool{}
	supersededTasks := make([]supersededProductionTask, 0)
	now := time.Now().UTC()
	// Pass 1: the marker is durable evidence that this video already consumed
	// its single no-receipt re-submission, even when the marker sits on a task
	// from an older job.
	for _, prior := range priorJobs {
		if prior.BookID != bookID || prior.DirectorRevisionID != book.DirectorRevision.ID {
			continue
		}
		for _, task := range prior.Tasks {
			if task.ErrorMessage == SupersededWithoutReceiptMarker {
				resubmittedVideos[task.VideoID] = true
			}
		}
	}
	for _, prior := range priorJobs {
		if prior.BookID != bookID || prior.DirectorRevisionID != book.DirectorRevision.ID {
			continue
		}
		for _, task := range prior.Tasks {
			if options.Force {
				continue
			}
			key := fmt.Sprintf("%s#v%d", task.VideoID, task.OpeningVariantIndex)
			if task.Status == ProductionSucceeded || strings.TrimSpace(task.ProviderTaskID) != "" {
				blockedVideos[key] = true
				continue
			}
			if task.Status != ProductionQueued && task.Status != ProductionRunning {
				continue
			}
			// A queued/running task without a provider id means submission was
			// interrupted before the receipt was stored. Older local builds
			// blocked it forever; allow exactly one re-submission per video once
			// the grace period has passed and no earlier re-submission exists.
			waited := now.Sub(task.UpdatedAt)
			if waited < ResubmitWithoutReceiptAfter || resubmittedVideos[task.VideoID] || task.ErrorMessage == SupersededWithoutReceiptMarker {
				blockedVideos[key] = true
				continue
			}
			resubmittedVideos[task.VideoID] = true
			supersededTasks = append(supersededTasks, supersededProductionTask{jobID: prior.ID, task: task, waited: waited})
		}
	}
	mediaVideos := book.Videos
	// Fixed opening keeps all Director cards for review, but only sends VIDEO01
	// into the media-production stage.
	if snapshot, snapshotErr := snapshotForBook(batch, book); snapshotErr != nil {
		return ProductionJob{}, snapshotErr
	} else if snapshot.FixedSingleVideo && len(mediaVideos) > 1 {
		mediaVideos = mediaVideos[:1]
	}
	openingOn := rawBool(ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch), "openingEnabled", false)
	if openingOn && len(book.Videos) >= 2 {
		count := rawInt(ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch), "openingCount", 4)
		if count < 1 || count > 8 {
			return ProductionJob{}, fmt.Errorf("%w: openingCount must be between 1 and 8", ErrInvalid)
		}
		// Missing opening variants are a retryable per-step gap, not a reason to
		// reject the whole book. The task loop below already submits exactly the
		// successful variants; when none exist it falls back to the original
		// opening, so a manual "generate video" action can still produce VIDEO01.
		available := successfulOpeningVariants(book.Videos[0])
		if expected := count - 1; len(available) < expected {
			log.Printf("production: opening variants incomplete book=%s available=%d expected=%d; submitting available variants only", bookID, len(available), expected)
		}
	}
	pendingVideos := make([]Video, 0, len(mediaVideos))
	for _, video := range mediaVideos {
		if options.VideoID != "" && video.ID != options.VideoID {
			continue
		}
		pendingVideos = append(pendingVideos, video)
	}
	if len(pendingVideos) == 0 {
		return ProductionJob{}, fmt.Errorf("%w: no pending VIDEOs to produce", ErrConflict)
	}

	job := ProductionJob{ID: "", Owner: owner, BatchID: batchID, BookID: bookID, RequestID: requestID, DirectorRevisionID: book.DirectorRevision.ID, Status: ProductionQueued, Tasks: []ProductionTask{}, CreatedAt: now, UpdatedAt: now}
	prompts := make(map[string]FinalPrompt, len(pendingVideos))
	// buildTask compiles one VIDEO at one opening-variant index. Every compile
	// happens before the job is persisted, so a failure aborts the whole
	// submission instead of leaving a partial job behind.
	buildTask := func(video Video, variantIndex int) error {
		key := fmt.Sprintf("%s#v%d", video.ID, variantIndex)
		if blockedVideos[key] {
			return nil
		}
		limit := model.MaxReferenceImages
		if limit <= 0 {
			limit = 3
		}
		var prompt FinalPrompt
		var compileErr error
		if variantIndex > 0 {
			compiler, ok := s.Compiler.(openingVariantPromptResolver)
			if !ok {
				return fmt.Errorf("%w: opening variant compiler is unavailable", ErrUnavailable)
			}
			prompt, compileErr = compiler.CompileForOpeningVariant(ctx, owner, batchID, bookID, video.ID, variantIndex, limit)
		} else if options.CompilationID != "" {
			compiler, ok := s.Compiler.(h3CompilationPromptResolver)
			if !ok {
				return fmt.Errorf("%w: H3 production compiler is unavailable", ErrUnavailable)
			}
			prompt, compileErr = compiler.CompileH3ForProduction(ctx, owner, batchID, bookID, video.ID, options.CompilationID, limit)
		} else if compiler, ok := s.Compiler.(referenceImageLimitPromptResolver); ok {
			prompt, compileErr = compiler.CompileWithReferenceImageLimit(ctx, owner, batchID, bookID, video.ID, limit)
		} else if compiler, ok := s.Compiler.(strictPromptResolver); ok {
			prompt, compileErr = compiler.CompileForProduction(ctx, owner, batchID, bookID, video.ID)
		} else {
			prompt, compileErr = s.Compiler.Compile(ctx, owner, batchID, bookID, video.ID)
		}
		if compileErr != nil {
			return compileErr
		}
		if options.StripReferenceImages {
			prompt = stripReferenceImagesFromPrompt(prompt)
		}
		selectedModel := rawString(prompt.EffectiveSettings.Values, "videoModelId", "")
		if !VideoModelMatchesProviderModel(selectedModel, model.ID, provider) {
			return fmt.Errorf("%w: selected video model %q is not available for provider %s", ErrConflict, selectedModel, provider)
		}
		prompts[key] = prompt
		job.Tasks = append(job.Tasks, ProductionTask{VideoID: video.ID, OpeningVariantIndex: variantIndex, Provider: provider, Status: ProductionQueued, Attempt: 1, FinalPromptHash: prompt.SnapshotHash, CompiledPrompt: prompt.CompiledPrompt, CompilationID: prompt.CompilationID, CompilationSegmentKey: prompt.CompilationSegmentKey, CompileTrace: cloneH3CompileTrace(prompt.CompileTrace), ReferenceImageURLs: append([]string(nil), prompt.ReferenceImageURLs...), DowngradedAssetIDs: append([]string(nil), prompt.DowngradedAssetIDs...), TargetDurationSeconds: float64(prompt.DurationSeconds), RequestedDurationSeconds: float64(prompt.DurationSeconds), CreatedAt: now, UpdatedAt: now})
		return nil
	}
	for _, video := range pendingVideos {
		indices := []int{0}
		if openingOn && video.ID == book.Videos[0].ID {
			for _, variant := range successfulOpeningVariants(video) {
				indices = append(indices, variant.Index)
			}
		}
		for _, variantIndex := range indices {
			if err := buildTask(video, variantIndex); err != nil {
				return ProductionJob{}, err
			}
		}
	}
	if len(job.Tasks) == 0 {
		return ProductionJob{}, fmt.Errorf("%w: no pending VIDEOs to produce", ErrConflict)
	}
	job, err = repository.CreateProductionJob(ctx, job)
	if err != nil {
		return ProductionJob{}, err
	}
	for _, superseded := range supersededTasks {
		task := superseded.task
		task.Status = ProductionCancelled
		task.ErrorMessage = SupersededWithoutReceiptMarker
		task.UpdatedAt = time.Now().UTC()
		if _, updateErr := repository.UpdateProductionTask(ctx, owner, superseded.jobID, task.ID, task); updateErr != nil {
			return ProductionJob{}, updateErr
		}
		log.Printf("batch factory resubmitted queued video without provider receipt video=%s variant=%d waited=%s job=%s", task.VideoID, task.OpeningVariantIndex, superseded.waited.Round(time.Second), job.ID)
	}

	for _, task := range job.Tasks {
		prompt, ok := prompts[fmt.Sprintf("%s#v%d", task.VideoID, task.OpeningVariantIndex)]
		if !ok {
			return ProductionJob{}, fmt.Errorf("%w: persisted task prompt is missing", ErrConflict)
		}
		var ref ProviderTaskRef
		var submitErr error
		if provider == VideoProviderDoubaoLocal {
			sourceTaskID := "bf11:" + batchID + ":" + bookID + ":" + task.VideoID + ":" + task.ID
			if task.OpeningVariantIndex > 0 {
				sourceTaskID = fmt.Sprintf("%s#v%d", sourceTaskID, task.OpeningVariantIndex)
			}
			localModel := model.ID
			if selectedModel := rawString(prompt.EffectiveSettings.Values, "videoModelId", ""); strings.TrimSpace(selectedModel) != "" {
				localModel = strings.TrimSpace(selectedModel)
			}
			ref, submitErr = s.LocalExecutor.Submit(ctx, owner, LocalVideoJobInput{
				SourceTaskID: sourceTaskID,
				BatchID:      batchID, BookID: bookID, VideoID: task.VideoID,
				Model: localModel, Prompt: prompt.CompiledPrompt,
				Duration:           prompt.DurationSeconds,
				AspectRatio:        EffectiveVideoAspectRatio(prompt.EffectiveSettings.Values),
				Resolution:         EffectiveVideoResolution(prompt.EffectiveSettings.Values),
				ReferenceImageURLs: append([]string(nil), prompt.ReferenceImageURLs...),
			})
		} else {
			submitAdapter := adapter
			submitModel := model
			// 梵客视频按书选的型号优先于账号目录里的默认型号；旧值（无后缀）继续用默认型号。
			if provider == VideoProviderYFAISeedance && s.ProviderRegistry != nil {
				if fanKeModel := FanKeModelFromSelection(rawString(prompt.EffectiveSettings.Values, "videoModelId", "")); fanKeModel != "" {
					cfg, resolveErr := s.ProviderRegistry.Resolve(ctx, owner, provider)
					if resolveErr != nil {
						submitErr = resolveErr
					} else {
						nextAdapter := &YFAISeedanceAdapter{BaseURL: cfg.CreateURL, APIKey: cfg.APIKey, Model: fanKeModel}
						if validateErr := nextAdapter.Validate(); validateErr != nil {
							submitErr = validateErr
						} else {
							submitModel.ID = fanKeModel
							submitAdapter = nextAdapter
						}
					}
				}
			}
			if submitErr == nil {
				ref, submitErr = submitAdapter.Submit(ctx, submitModel, prompt)
			}
		}
		task.UpdatedAt = time.Now().UTC()
		if submitErr != nil {
			task.Status, task.ErrorMessage = ProductionFailed, productionError(submitErr)
		} else {
			task.Status, task.ProviderTaskID, task.MediaURL, task.ActualDurationSeconds = normalizeProductionState(ref.State), strings.TrimSpace(ref.ProviderTaskID), strings.TrimSpace(ref.MediaURL), ref.ActualDurationSeconds
			if ref.RequestedDurationSeconds > 0 {
				task.RequestedDurationSeconds = ref.RequestedDurationSeconds
			}
		}
		updated, updateErr := repository.UpdateProductionTask(ctx, owner, job.ID, task.ID, task)
		if updateErr != nil {
			return ProductionJob{}, updateErr
		}
		job = updated
	}
	return job, nil
}

// SubmitBatchProduction is the "generate pending" operation. It submits only
// current Director VIDEO identities that have no queued, running, or succeeded
// task, so a repeated batch-button click cannot duplicate successful work.
func (s *ProductionService) SubmitBatchProduction(ctx context.Context, owner, batchID, requestID string) (BatchStatus, error) {
	return s.SubmitBatchProductionWithProvider(ctx, owner, batchID, requestID, VideoProviderPersonalAPI)
}

func (s *ProductionService) SubmitBatchProductionWithProvider(ctx context.Context, owner, batchID, requestID, provider string) (BatchStatus, error) {
	if s == nil || !s.Enabled {
		return BatchStatus{}, fmt.Errorf("%w: production is not enabled", ErrUnavailable)
	}
	if strings.TrimSpace(requestID) == "" {
		return BatchStatus{}, fmt.Errorf("%w: request id is required", ErrInvalid)
	}
	if s.Store == nil {
		return BatchStatus{}, ErrUnavailable
	}
	provider = normalizeVideoProvider(provider)
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return BatchStatus{}, err
	}
	submitted := false
	for _, book := range batch.Books {
		if book.DirectorRevision == nil || len(book.Videos) == 0 {
			continue
		}
		_, submitErr := s.SubmitBookProductionWithProvider(ctx, owner, batchID, book.ID, requestID+":"+book.ID, provider)
		if errors.Is(submitErr, ErrConflict) {
			continue
		}
		if submitErr != nil {
			return BatchStatus{}, submitErr
		}
		submitted = true
	}
	status, err := s.GetBatchStatus(ctx, owner, batchID)
	if err != nil {
		return BatchStatus{}, err
	}
	if !submitted && len(status.Jobs) == 0 {
		return BatchStatus{}, fmt.Errorf("%w: no pending VIDEOs to produce", ErrConflict)
	}
	return status, nil
}

// RemoveBookProductionCandidate hides one completed non-primary candidate from
// this book's clip library. It intentionally preserves provider media, task
// events and past merge receipts so production and publishing remain auditable.
func (s *ProductionService) RemoveBookProductionCandidate(ctx context.Context, owner, batchID, bookID, videoID, taskID string) error {
	if s == nil || s.Store == nil {
		return ErrUnavailable
	}
	repository, err := s.repository()
	if err != nil {
		return err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return err
	}
	videoFound := false
	for _, video := range book.Videos {
		if video.ID == videoID {
			videoFound = true
			break
		}
	}
	if !videoFound {
		return ErrNotFound
	}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return err
	}
	activeDirectorRevisionID := ""
	if book.DirectorRevision != nil {
		activeDirectorRevisionID = book.DirectorRevision.ID
	}
	selections := productionTaskSelections(jobs, bookID, activeDirectorRevisionID)[videoID]
	primary, hasPrimary := selectedProductionTask(Video{ID: videoID, SettingsState: videoSettingsFor(book, videoID)}, selections)
	if hasPrimary && primary.Task.ID == taskID {
		return fmt.Errorf("%w: 当前分镜主视频不能删除，请先选择其他候选版本", ErrConflict)
	}
	if isSelectedUploadTask(book, videoID, taskID) {
		return fmt.Errorf("%w: 当前上传主视频不能删除，请先替换上传主视频", ErrConflict)
	}
	return repository.HideProductionTask(ctx, owner, batchID, bookID, videoID, taskID)
}

func videoSettingsFor(book Book, videoID string) SettingsState {
	for _, video := range book.Videos {
		if video.ID == videoID {
			return video.SettingsState
		}
	}
	return SettingsState{}
}

func isSelectedUploadTask(book Book, videoID, taskID string) bool {
	raw, ok := book.SettingsState.Patch["primaryUploadSource"]
	if !ok {
		return false
	}
	var source struct {
		Kind    string `json:"kind"`
		VideoID string `json:"videoId"`
		TaskID  string `json:"taskId"`
	}
	if json.Unmarshal(raw, &source) != nil {
		return false
	}
	return source.Kind == "video" && source.VideoID == videoID && source.TaskID == taskID
}

// RegisterUploadedVideo records an already uploaded HTTPS MP4 as a successful
// manual candidate and makes it the selected media for its VIDEO.
func (s *ProductionService) RegisterUploadedVideo(ctx context.Context, owner, batchID, bookID, videoID, uploadID, mediaURL string) (ProductionTask, error) {
	if s == nil || s.Store == nil {
		return ProductionTask{}, ErrUnavailable
	}
	if strings.TrimSpace(uploadID) == "" {
		return ProductionTask{}, fmt.Errorf("%w: upload id is required", ErrInvalid)
	}
	parsed, err := url.Parse(strings.TrimSpace(mediaURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return ProductionTask{}, fmt.Errorf("%w: uploaded media URL must be HTTPS", ErrInvalid)
	}
	repository, err := s.repository()
	if err != nil {
		return ProductionTask{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return ProductionTask{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return ProductionTask{}, err
	}
	if book.DirectorRevision == nil || strings.TrimSpace(book.DirectorRevision.ID) == "" {
		return ProductionTask{}, fmt.Errorf("%w: active Director revision is required", ErrConflict)
	}
	var video Video
	found := false
	for _, candidate := range book.Videos {
		if candidate.ID == videoID {
			video, found = candidate, true
			break
		}
	}
	if !found {
		return ProductionTask{}, ErrNotFound
	}
	now := time.Now().UTC()
	job, err := repository.CreateProductionJob(ctx, ProductionJob{Owner: owner, BatchID: batchID, BookID: bookID, RequestID: "manual-upload:" + uploadID, DirectorRevisionID: book.DirectorRevision.ID, Status: ProductionSucceeded, Tasks: []ProductionTask{{VideoID: videoID, Provider: "manual_upload", Status: ProductionSucceeded, Attempt: 1, MediaURL: parsed.String(), CreatedAt: now, UpdatedAt: now}}, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return ProductionTask{}, err
	}
	if len(job.Tasks) != 1 || job.Tasks[0].VideoID != videoID {
		return ProductionTask{}, ErrConflict
	}
	task := job.Tasks[0]
	encoded, err := json.Marshal(task.ID)
	if err != nil {
		return ProductionTask{}, err
	}
	if _, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: videoID}, SettingsUpdate{Patch: SettingsPatch{"primaryMediaTaskId": encoded}, ExpectedRevision: video.Revision}); err != nil {
		return ProductionTask{}, err
	}
	return task, nil
}

func (s *ProductionService) GetBatchStatus(ctx context.Context, owner, batchID string) (BatchStatus, error) {
	if s == nil || s.Store == nil {
		return BatchStatus{}, ErrUnavailable
	}
	repository, err := s.repository()
	if err != nil {
		return BatchStatus{}, err
	}
	if err := ensureBatchOwnership(ctx, s.Store, owner, batchID); err != nil {
		return BatchStatus{}, err
	}
	if err := s.reconcileBatch(ctx, repository, owner, batchID); err != nil {
		return BatchStatus{}, err
	}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return BatchStatus{}, err
	}
	return BatchStatus{BatchID: batchID, Jobs: jobs}, nil
}

func isOperationRequest(jobRequestID, operationRequestID string) bool {
	jobRequestID = strings.TrimSpace(jobRequestID)
	operationRequestID = strings.TrimSpace(operationRequestID)
	return operationRequestID != "" && (jobRequestID == operationRequestID || strings.HasPrefix(jobRequestID, operationRequestID+":"))
}

func (s *ProductionService) CancelBatch(ctx context.Context, owner, batchID, operationRequestID string) (BatchCancellation, error) {
	if s == nil || s.Store == nil {
		return BatchCancellation{}, ErrUnavailable
	}
	operationRequestID = strings.TrimSpace(operationRequestID)
	if operationRequestID == "" {
		return BatchCancellation{}, fmt.Errorf("%w: production operation request id is required", ErrInvalid)
	}
	repository, err := s.repository()
	if err != nil {
		return BatchCancellation{}, err
	}
	if _, err := s.Store.GetBatch(ctx, owner, batchID); err != nil {
		return BatchCancellation{}, err
	}
	result := BatchCancellation{BatchID: batchID, RequestID: operationRequestID, CancelledTaskIDs: []string{}, Skipped: []ProductionCancelSkip{}}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return BatchCancellation{}, err
	}
	for _, job := range jobs {
		if !isOperationRequest(job.RequestID, operationRequestID) {
			continue
		}
		for _, task := range job.Tasks {
			if task.Status != ProductionQueued && task.Status != ProductionRunning {
				continue
			}
			if normalizeVideoProvider(task.Provider) != VideoProviderDoubaoLocal {
				result.Skipped = append(result.Skipped, ProductionCancelSkip{TaskID: task.ID, Reason: "当前视频供应商未提供真实取消接口"})
				continue
			}
			if s.LocalExecutor == nil || strings.TrimSpace(task.ProviderTaskID) == "" {
				result.Skipped = append(result.Skipped, ProductionCancelSkip{TaskID: task.ID, Reason: "本地执行器任务不可取消"})
				continue
			}
			ref, cancelErr := s.LocalExecutor.Cancel(ctx, owner, task.ProviderTaskID)
			if cancelErr != nil {
				result.Skipped = append(result.Skipped, ProductionCancelSkip{TaskID: task.ID, Reason: productionError(cancelErr)})
				continue
			}
			if ref.State != ProductionCancelled {
				result.Skipped = append(result.Skipped, ProductionCancelSkip{TaskID: task.ID, Reason: "本地执行器未确认取消"})
				continue
			}
			updated := task
			updated.Status = ProductionCancelled
			updated.ErrorMessage = "已由用户取消"
			updated.UpdatedAt = time.Now().UTC()
			if _, err := repository.UpdateProductionTask(ctx, owner, job.ID, task.ID, updated); err != nil {
				return BatchCancellation{}, err
			}
			result.CancelledTaskIDs = append(result.CancelledTaskIDs, task.ID)
		}
	}
	return result, nil
}

func (s *ProductionService) reconcileBatch(ctx context.Context, repository ProductionRepository, owner, batchID string) error {
	if s.Poller == nil && s.LocalExecutor == nil && s.ProviderRegistry == nil {
		return nil
	}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		for _, task := range job.Tasks {
			if task.ProviderTaskID == "" || (task.Status != ProductionQueued && task.Status != ProductionRunning) {
				continue
			}
			var ref ProviderTaskRef
			var pollErr error
			provider := normalizeVideoProvider(task.Provider)
			if provider == VideoProviderDoubaoLocal {
				if s.LocalExecutor == nil {
					pollErr = fmt.Errorf("%w: Doubao local executor is offline", ErrUnavailable)
				} else {
					ref, pollErr = s.LocalExecutor.Poll(ctx, owner, task.ProviderTaskID)
				}
			} else if s.ProviderRegistry != nil {
				adapter, model, resolveErr := s.resolveProvider(ctx, owner, provider)
				if resolveErr != nil && s.Poller != nil {
					ref, pollErr = s.Poller.Poll(ctx, s.Model, ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, State: task.Status, MediaURL: task.MediaURL, RequestedDurationSeconds: task.RequestedDurationSeconds})
				} else if resolveErr != nil {
					pollErr = resolveErr
				} else if poller, ok := adapter.(ProductionPoller); !ok {
					pollErr = fmt.Errorf("%w: provider does not support polling", ErrUnavailable)
				} else {
					ref, pollErr = poller.Poll(ctx, model, ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, State: task.Status, MediaURL: task.MediaURL, RequestedDurationSeconds: task.RequestedDurationSeconds})
				}
			} else if s.Poller != nil {
				ref, pollErr = s.Poller.Poll(ctx, s.Model, ProviderTaskRef{ProviderTaskID: task.ProviderTaskID, State: task.Status, MediaURL: task.MediaURL, RequestedDurationSeconds: task.RequestedDurationSeconds})
			} else {
				continue
			}
			updated := task
			if pollErr != nil {
				// A transport failure is retryable; persist a bounded diagnostic while
				// keeping the task running so a later status read can recover it.
				updated.ErrorMessage = productionError(pollErr)
			} else {
				updated.Status = normalizeProductionState(ref.State)
				updated.ProviderTaskID = strings.TrimSpace(ref.ProviderTaskID)
				updated.MediaURL = strings.TrimSpace(ref.MediaURL)
				if updated.Status == ProductionFailed {
					updated.ErrorMessage = productionError(errors.New(strings.TrimSpace(ref.ErrorMessage)))
				}
				if ref.ActualDurationSeconds > 0 {
					updated.ActualDurationSeconds = ref.ActualDurationSeconds
				}
				if updated.Status == ProductionSucceeded {
					updated.ErrorMessage = ""
				}
			}
			if updated.Status == task.Status && updated.ProviderTaskID == task.ProviderTaskID && updated.MediaURL == task.MediaURL && updated.ErrorMessage == task.ErrorMessage {
				continue
			}
			if _, err := repository.UpdateProductionTask(ctx, owner, job.ID, task.ID, updated); err != nil {
				return err
			}
		}
	}
	return nil
}
