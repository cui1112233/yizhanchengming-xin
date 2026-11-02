package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type BookStageAvailability struct {
	Stage     BookStage `json:"stage"`
	Available bool      `json:"available"`
	Reason    string    `json:"reason,omitempty"`
}

type BookStageSummary struct {
	BookID       string                  `json:"bookId"`
	Runs         []BookStageRun          `json:"runs"`
	Availability []BookStageAvailability `json:"availability"`
	LastFailed   *BookStageRun           `json:"lastFailed,omitempty"`
}

type BookStageService struct {
	Store      Store
	Director   *DirectorService
	Production *ProductionService
	// ImageReady and ImageGenerator are supplied by the host application so
	// the stage service stays independent from a particular image provider.
	ImageReady func(context.Context, string) (bool, string)
	// VideoReady is supplied by the host application so availability reflects
	// the account's selected video model, not only the global production flag.
	VideoReady        func(context.Context, string) (bool, string)
	ImageGenerator    func(context.Context, string, string, string) error
	SmartUnifiedStyle string
	OpeningMeta       PresetSnapshot
	H3Director        bool
	ProviderOverride  string
}

func (s *BookStageService) RunComicBootstrap(ctx context.Context, owner, batchID, bookID string) (ComicBootstrapResult, error) {
	if s == nil || s.Director == nil {
		return ComicBootstrapResult{}, fmt.Errorf("%w: 漫画首次统一生成服务未配置", ErrUnavailable)
	}
	return s.Director.RunComicBootstrap(ctx, owner, batchID, bookID)
}

func (s *BookStageService) RunMissingComicPromptExtraction(ctx context.Context, owner, batchID, bookID string) ([]Video, error) {
	if s == nil || s.Director == nil {
		return nil, fmt.Errorf("%w: 漫画缺页提示词恢复服务未配置", ErrUnavailable)
	}
	return s.Director.RunMissingComicPromptExtraction(ctx, owner, batchID, bookID)
}

func (s *BookStageService) repository() (BookStageRunRepository, error) {
	if s == nil || s.Store == nil {
		return nil, ErrUnavailable
	}
	repository, ok := s.Store.(BookStageRunRepository)
	if !ok {
		return nil, ErrUnavailable
	}
	return repository, nil
}

func (s *BookStageService) availability(ctx context.Context, owner string, stage BookStage) BookStageAvailability {
	switch stage {
	case BookStageAssets:
		if s != nil && s.Director != nil && s.Director.Store != nil && s.Director.Provider != nil {
			return BookStageAvailability{Stage: stage, Available: true}
		}
		return BookStageAvailability{Stage: stage, Reason: "文本模型未配置"}
	case BookStageDirector, BookStageVisual, BookStageComic:
		if s != nil && s.Director != nil && s.Director.Store != nil && s.Director.Provider != nil {
			return BookStageAvailability{Stage: stage, Available: true}
		}
		return BookStageAvailability{Stage: stage, Reason: "文本模型未配置"}
	case BookStageOpening:
		if s.Director != nil && s.Director.Provider != nil {
			return BookStageAvailability{Stage: stage, Available: true}
		}
		return BookStageAvailability{Stage: stage, Reason: "文本模型未配置"}
	case BookStageImage:
		if s != nil && s.ImageReady != nil {
			ready, reason := s.ImageReady(ctx, owner)
			if ready {
				return BookStageAvailability{Stage: stage, Available: true}
			}
			if strings.TrimSpace(reason) != "" {
				return BookStageAvailability{Stage: stage, Reason: reason}
			}
		}
		return BookStageAvailability{Stage: stage, Reason: "图片模型或图片生成服务未配置"}
	case BookStageVideo:
		if s != nil && s.VideoReady != nil {
			ready, reason := s.VideoReady(ctx, owner)
			if ready {
				return BookStageAvailability{Stage: stage, Available: true}
			}
			if strings.TrimSpace(reason) != "" {
				return BookStageAvailability{Stage: stage, Reason: reason}
			}
		}
		if s != nil && s.Production != nil && s.Production.Enabled {
			return BookStageAvailability{Stage: stage, Available: true}
		}
		return BookStageAvailability{Stage: stage, Reason: "视频生成服务未启用"}
	default:
		return BookStageAvailability{Stage: stage, Reason: "未知阶段"}
	}
}

func (s *BookStageService) Summary(ctx context.Context, owner, batchID, bookID string) (BookStageSummary, error) {
	repository, err := s.repository()
	if err != nil {
		return BookStageSummary{}, err
	}
	runs, err := repository.ListBookStageRuns(ctx, owner, batchID, bookID)
	if err != nil {
		return BookStageSummary{}, err
	}
	availability := []BookStageAvailability{s.availability(ctx, owner, BookStageAssets), s.availability(ctx, owner, BookStageDirector), s.availability(ctx, owner, BookStageOpening), s.availability(ctx, owner, BookStageVisual), s.availability(ctx, owner, BookStageComic), s.availability(ctx, owner, BookStageImage), s.availability(ctx, owner, BookStageVideo)}
	return BookStageSummary{BookID: bookID, Runs: runs, Availability: availability, LastFailed: LatestFailedBookStageRun(runs)}, nil
}

func (s *BookStageService) nextAttempt(runs []BookStageRun, stage BookStage) int {
	highest := 0
	for _, run := range runs {
		if run.Stage == stage && run.Attempt > highest {
			highest = run.Attempt
		}
	}
	return highest + 1
}

func (s *BookStageService) createRun(ctx context.Context, owner, batchID, bookID string, stage BookStage, requestID, inputRevision string) (BookStageRun, error) {
	repository, err := s.repository()
	if err != nil {
		return BookStageRun{}, err
	}
	runs, err := repository.ListBookStageRuns(ctx, owner, batchID, bookID)
	if err != nil {
		return BookStageRun{}, err
	}
	return repository.CreateBookStageRun(ctx, BookStageRun{Owner: owner, BatchID: batchID, BookID: bookID, Stage: stage, Status: ProductionRunning, Attempt: s.nextAttempt(runs, stage), RequestID: requestID, InputRevision: inputRevision})
}

func (s *BookStageService) finish(ctx context.Context, owner string, run BookStageRun, runErr error) error {
	repository, err := s.repository()
	if err != nil {
		return err
	}
	status := ProductionSucceeded
	text := ""
	if runErr != nil {
		status, text = ProductionFailed, productionError(runErr)
	}
	_, err = repository.UpdateBookStageRun(ctx, owner, run.ID, BookStageRun{Status: status, ErrorMessage: text, InputRevision: run.InputRevision})
	return err
}

func normalizeStageExecutionError(stage BookStage, err error) error {
	if err == nil || (stage != BookStageAssets && stage != BookStageDirector && stage != BookStageVisual && stage != BookStageComic && stage != BookStageOpening) {
		return err
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrConflict) || errors.Is(err, ErrInvalid) || errors.Is(err, ErrUnavailable) {
		return err
	}
	return fmt.Errorf("%w: 文本模型请求失败：%v", ErrUnavailable, err)
}

// immediateVideoProviderFailure turns a provider-side rejection that was
// persisted on a production task into a failed book stage.  A queued/running
// task is a valid asynchronous submission, but a task already marked failed
// must not leave the UI saying that the stage completed successfully.
func immediateVideoProviderFailure(job ProductionJob) error {
	for _, task := range job.Tasks {
		if task.Status != ProductionFailed && task.Status != ProductionCancelled {
			continue
		}
		message := strings.TrimSpace(task.ErrorMessage)
		if message == "" {
			message = "视频模型未返回可执行任务或成片地址"
		}
		return fmt.Errorf("%w: 视频模型提交失败（%s）：%s", ErrUnavailable, task.VideoID, message)
	}
	return nil
}

func (s *BookStageService) Run(ctx context.Context, owner, batchID, bookID string, stage BookStage, mode StageMode, requestID, videoID string) (BookStageSummary, error) {
	return s.run(ctx, owner, batchID, bookID, stage, mode, requestID, videoID, "")
}

func (s *BookStageService) run(ctx context.Context, owner, batchID, bookID string, stage BookStage, mode StageMode, requestID, videoID, frozenCompilationID string) (BookStageSummary, error) {
	if !validBookStage(stage) {
		return BookStageSummary{}, ErrInvalid
	}
	if !validStageMode(stage, mode) {
		return BookStageSummary{}, ErrInvalid
	}
	if strings.TrimSpace(requestID) == "" {
		return BookStageSummary{}, fmt.Errorf("%w: request id is required", ErrInvalid)
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return BookStageSummary{}, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return BookStageSummary{}, err
	}
	effectiveMode := normalizeProductionMode(ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch))
	if mode == StageModeMissing && effectiveMode == "comic" {
		if stage == BookStageDirector {
			if segmentStore, ok := s.Store.(ComicSourceSegmentStore); ok {
				if segmentation, segmentErr := segmentStore.GetComicSourceSegmentation(ctx, owner, batchID, bookID); segmentErr == nil && !segmentation.Stale {
					return s.Summary(ctx, owner, batchID, bookID)
				}
			}
		} else if stage == BookStageComic {
			// Existing page prompts are not enough to skip the comic stage: legacy
			// batches may have prompts but no durable continuous-source segments.
			// Let the stage run once so the source is segmented and attached to pages.
			segmentationReady := false
			if segmentStore, ok := s.Store.(ComicSourceSegmentStore); ok {
				if segmentation, segmentErr := segmentStore.GetComicSourceSegmentation(ctx, owner, batchID, bookID); segmentErr == nil && !segmentation.Stale {
					segmentationReady = true
				}
			}
			if segmentationReady && comicStageAlreadyComplete(book, stage, ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)) {
				return s.Summary(ctx, owner, batchID, bookID)
			}
		} else if stage == BookStageAssets && comicAssetStageCanReuse(book) {
			return s.Summary(ctx, owner, batchID, bookID)
		} else if comicStageAlreadyComplete(book, stage, ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)) {
			return s.Summary(ctx, owner, batchID, bookID)
		}
	}
	if stage == BookStageVideo && effectiveMode == "comic" {
		return BookStageSummary{}, fmt.Errorf("%w: 漫画生成已开启，不能生成视频", ErrConflict)
	}
	if stage == BookStageComic && effectiveMode != "comic" {
		return BookStageSummary{}, fmt.Errorf("%w: 请先打开漫画生成", ErrConflict)
	}
	// Missing-mode automation is idempotent: once a director revision exists,
	// do not create another stage run or ask the provider for the same book
	// again. Explicit force mode remains the user-controlled regenerate path.
	if stage == BookStageDirector && mode == StageModeMissing && book.DirectorRevision != nil {
		return s.Summary(ctx, owner, batchID, bookID)
	}
	availability := s.availability(ctx, owner, stage)
	inputRevision := ""
	if stage == BookStageVideo {
		inputRevision = strings.TrimSpace(videoID)
	}
	run, createErr := s.createRun(ctx, owner, batchID, bookID, stage, requestID, inputRevision)
	if createErr != nil {
		return BookStageSummary{}, createErr
	}
	if !availability.Available {
		err = fmt.Errorf("%w: %s", ErrUnavailable, availability.Reason)
		_ = s.finish(ctx, owner, run, err)
		summary, _ := s.Summary(ctx, owner, batchID, bookID)
		return summary, err
	}
	switch stage {
	case BookStageAssets:
		if effectiveMode == "comic" {
			_, err = s.Director.RunComicAssetExtraction(ctx, owner, batchID, bookID)
		} else {
			_, err = s.Director.RunAssetExtraction(ctx, owner, batchID, bookID)
		}
	case BookStageOpening:
		_, err = s.Director.RunOpeningVariants(ctx, owner, batchID, bookID, s.OpeningMeta)
		if err == nil && book.DirectorRevision != nil {
			// 换开头绑定当前导演分镜版本：重新生成导演分镜后旧 revision 的成功
			// 记录不再满足编排器门禁，变体会随新 revision 重新生成。
			run.InputRevision = book.DirectorRevision.ID
		}
	case BookStageVisual:
		_, err = s.Director.RunVisualPromptExtraction(ctx, owner, batchID, bookID)
	case BookStageComic:
		if mode == StageModeBootstrap {
			_, err = s.Director.RunComicBootstrap(ctx, owner, batchID, bookID)
			break
		}
		if segmentStore, ok := s.Store.(ComicSourceSegmentStore); ok {
			segmentation, segmentErr := segmentStore.GetComicSourceSegmentation(ctx, owner, batchID, bookID)
			if errors.Is(segmentErr, ErrNotFound) || (segmentErr == nil && segmentation.Stale) {
				_, segmentErr = s.Director.RunComicSourceSegmentation(ctx, owner, batchID, bookID)
			}
			if segmentErr != nil {
				err = segmentErr
				break
			}
		}
		if mode == StageModeMissing && comicStageAlreadyComplete(book, BookStageComic, ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)) {
			// Source segmentation was the missing part; preserve already accepted
			// page prompts and avoid requiring a second asset-extraction pass.
			err = nil
		} else if mode == StageModeMissing {
			_, err = s.Director.RunMissingComicPromptExtraction(ctx, owner, batchID, bookID)
		} else {
			_, err = s.Director.RunComicPromptExtraction(ctx, owner, batchID, bookID)
		}
	case BookStageImage:
		if s.ImageGenerator == nil {
			err = fmt.Errorf("%w: 图片模型或图片生成服务未配置", ErrUnavailable)
		} else {
			err = s.ImageGenerator(ctx, owner, batchID, bookID)
		}
	case BookStageDirector:
		if effectiveMode == "comic" {
			var segmentation ComicSourceSegmentation
			segmentation, err = s.Director.RunComicSourceSegmentation(ctx, owner, batchID, bookID)
			if err == nil {
				run.InputRevision = fmt.Sprintf("comic-source-v%d", segmentation.Version)
			}
		} else if mode == StageModeMissing && book.DirectorRevision != nil {
			err = fmt.Errorf("%w: 导演分镜已存在；请使用重新生成导演分镜", ErrConflict)
		} else {
			var revision DirectorRevision
			if s.usesUnifiedDirectorFlow(batch, book) {
				// H3 video prompt presets use the structured director document and
				// its duration/compilation trace. The caller selects this explicitly;
				// SD presets use the SD text contract below instead.
				revision, err = s.Director.RunConfiguredH3Director(ctx, owner, batchID, bookID, s.SmartUnifiedStyle)
			} else {
				// SD is the default video prompt contract. Historical records without
				// a saved selection remain readable through this same compatible path.
				revision, err = s.Director.RunDirectorWithSmartUnifiedStyle(ctx, owner, batchID, bookID, s.SmartUnifiedStyle)
			}
			if err == nil {
				run.InputRevision = revision.ID
			}
		}
	case BookStageVideo:
		effective := ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
		provider := strings.TrimSpace(s.ProviderOverride)
		if provider == "" {
			provider = VideoProviderForModel(rawString(effective, "videoModelId", ""), rawString(effective, "videoProvider", VideoProviderPersonalAPI))
		}
		if mode == StageModeMissing {
			var job ProductionJob
			job, err = s.Production.SubmitBookProductionWithOptions(ctx, owner, batchID, bookID, requestID, provider, ProductionOptions{VideoID: videoID, CompilationID: frozenCompilationID})
			if err == nil {
				err = immediateVideoProviderFailure(job)
			}
		} else {
			var job ProductionJob
			job, err = s.Production.SubmitBookProductionWithOptions(ctx, owner, batchID, bookID, requestID, provider, ProductionOptions{Force: true, VideoID: videoID, CompilationID: frozenCompilationID})
			if err == nil {
				err = immediateVideoProviderFailure(job)
			}
		}
	}
	err = normalizeStageExecutionError(stage, err)
	if finishErr := s.finish(ctx, owner, run, err); finishErr != nil {
		return BookStageSummary{}, finishErr
	}
	summary, summaryErr := s.Summary(ctx, owner, batchID, bookID)
	if summaryErr != nil {
		return BookStageSummary{}, summaryErr
	}
	return summary, err
}

func comicStageAlreadyComplete(book Book, stage BookStage, effective SettingsPatch) bool {
	switch stage {
	case BookStageAssets:
		for _, asset := range book.AssetRecords {
			if strings.TrimSpace(asset.Source) == "comic" {
				return true
			}
		}
		return false
	case BookStageDirector:
		return len(book.Videos) > 0
	case BookStageComic:
		if len(book.Videos) == 0 {
			return false
		}
		light := aiReasoningPromptConfig(effective).Comic.ID == "batch-comic-page-light"
		for _, video := range book.Videos {
			if strings.TrimSpace(video.ComicPrompt) == "" || (!light && strings.TrimSpace(video.ComicImagePrompt) == "") {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (s *BookStageService) usesUnifiedDirectorFlow(batch Batch, book Book) bool {
	return s != nil && s.H3Director
}

// RetryLastFailed reruns only the most recently failed stage recorded for this book.
func (s *BookStageService) RetryLastFailed(ctx context.Context, owner, batchID, bookID, requestID, videoID string) (BookStageSummary, error) {
	summary, err := s.Summary(ctx, owner, batchID, bookID)
	if err != nil {
		return BookStageSummary{}, err
	}
	if summary.LastFailed == nil {
		failedTask, found := s.latestFailedProductionTask(ctx, owner, batchID, bookID, videoID)
		if !found {
			return BookStageSummary{}, fmt.Errorf("%w: 当前小说没有失败步骤", ErrConflict)
		}
		return s.run(ctx, owner, batchID, bookID, BookStageVideo, StageModeForce, requestID, failedTask.VideoID, failedTask.CompilationID)
	}
	if strings.TrimSpace(videoID) == "" && summary.LastFailed.Stage == BookStageVideo {
		videoID = summary.LastFailed.InputRevision
	}
	frozenCompilationID := ""
	if summary.LastFailed.Stage == BookStageVideo {
		frozenCompilationID = s.latestFailedH3CompilationID(ctx, owner, batchID, bookID, videoID)
	}
	return s.run(ctx, owner, batchID, bookID, summary.LastFailed.Stage, StageModeForce, requestID, videoID, frozenCompilationID)
}

func (s *BookStageService) latestFailedH3CompilationID(ctx context.Context, owner, batchID, bookID, videoID string) string {
	failedTask, found := s.latestFailedProductionTask(ctx, owner, batchID, bookID, videoID)
	if !found {
		return ""
	}
	return failedTask.CompilationID
}

// latestFailedProductionTask covers provider failures that arrive after the
// video submission stage completed. Those failures are real retry targets even
// though there is no failed stage-run row to report them.
func (s *BookStageService) latestFailedProductionTask(ctx context.Context, owner, batchID, bookID, videoID string) (ProductionTask, bool) {
	repository, ok := s.Store.(ProductionRepository)
	if !ok {
		return ProductionTask{}, false
	}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return ProductionTask{}, false
	}
	var latest ProductionTask
	found := false
	for _, job := range jobs {
		if job.BookID != bookID {
			continue
		}
		for _, task := range job.Tasks {
			if task.Status != ProductionFailed || (strings.TrimSpace(videoID) != "" && task.VideoID != videoID) {
				continue
			}
			if !found || task.UpdatedAt.After(latest.UpdatedAt) || (task.UpdatedAt.Equal(latest.UpdatedAt) && task.ID > latest.ID) {
				latest, found = task, true
			}
		}
	}
	return latest, found
}
