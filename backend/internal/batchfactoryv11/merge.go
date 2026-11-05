package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type MergeState string

const (
	MergeQueued    MergeState = "queued"
	MergeRunning   MergeState = "running"
	MergeSucceeded MergeState = "succeeded"
	MergeFailed    MergeState = "failed"
)

type MergeMedia struct {
	ProductionJobID string `json:"productionJobId"`
	VideoID         string `json:"videoId"`
	MediaURL        string `json:"mediaUrl"`
	// URL is an internal compatibility alias for older V11 consumers. It is
	// deliberately excluded from the provider JSON contract.
	URL                      string  `json:"-"`
	Order                    int     `json:"order"`
	TargetDurationSeconds    float64 `json:"targetDurationSeconds,omitempty"`
	RequestedDurationSeconds float64 `json:"requestedDurationSeconds,omitempty"`
	ActualDurationSeconds    float64 `json:"actualDurationSeconds,omitempty"`
}

type MergeOptions struct {
	TimingMode           string  `json:"timingMode,omitempty"`
	AudioSource          string  `json:"audioSource,omitempty"`
	AudioArtifactID      string  `json:"audioArtifactId,omitempty"`
	Speed                float64 `json:"speed,omitempty"`
	TTSSpeed             float64 `json:"ttsSpeed,omitempty"`
	AudioDurationSeconds float64 `json:"audioDurationSeconds,omitempty"`
	AspectRatio          string  `json:"-"`
}

type MergeJob struct {
	ID                  string       `json:"id"`
	Owner               string       `json:"-"`
	BatchID             string       `json:"batchId"`
	BookID              string       `json:"bookId,omitempty"`
	OpeningVariantIndex int          `json:"openingVariantIndex,omitempty"`
	RequestID           string       `json:"requestId"`
	TimingMode          string       `json:"timingMode,omitempty"`
	AudioSource         string       `json:"audioSource,omitempty"`
	AudioArtifactID     string       `json:"audioArtifactId,omitempty"`
	Speed               float64      `json:"speed,omitempty"`
	ProviderTaskID      string       `json:"providerTaskId,omitempty"`
	Status              MergeState   `json:"status"`
	ProgressPhase       string       `json:"progressPhase,omitempty"`
	ProgressCurrent     int          `json:"progressCurrent,omitempty"`
	ProgressTotal       int          `json:"progressTotal,omitempty"`
	Sources             []MergeMedia `json:"sources"`
	OutputURL           string       `json:"outputUrl,omitempty"`
	ErrorMessage        string       `json:"errorMessage,omitempty"`
	CreatedAt           time.Time    `json:"createdAt"`
	UpdatedAt           time.Time    `json:"updatedAt"`
}

type MergeAdapter interface {
	Submit(context.Context, string, []MergeMedia, MergeOptions) (MergeJob, error)
}

type MergePoller interface {
	Poll(context.Context, string, MergeJob) (MergeJob, error)
}

// VideoDurationProbe reads the completed media artifact. Providers may return
// a usable URL without a duration, but audio-following merge must use the
// artifact's measured duration rather than the requested duration.
type VideoDurationProbe interface {
	DurationSeconds(context.Context, string) (float64, error)
}

type MergeRepository interface {
	FindMergeJob(context.Context, string, string, string) (MergeJob, error)
	CreateMergeJob(context.Context, MergeJob) (MergeJob, error)
	CreateUploadedMergeJob(context.Context, MergeJob) (MergeJob, error)
	UpdateMergeJob(context.Context, string, string, MergeJob) (MergeJob, error)
	ListMergeJobs(context.Context, string, string) ([]MergeJob, error)
}

type MergeService struct {
	Store         Store
	Adapter       MergeAdapter
	Poller        MergePoller
	DurationProbe VideoDurationProbe
	Enabled       bool
}

func normalizeMergeOptions(options MergeOptions) (MergeOptions, error) {
	options.AudioSource = strings.ToLower(strings.TrimSpace(options.AudioSource))
	if options.AudioSource == "" {
		options.AudioSource = "original"
	}
	if options.AudioSource != "tts" && options.AudioSource != "original" {
		return MergeOptions{}, fmt.Errorf("%w: unsupported audio source", ErrInvalid)
	}
	if options.AudioSource == "tts" && strings.TrimSpace(options.AudioArtifactID) == "" {
		return MergeOptions{}, fmt.Errorf("%w: TTS audio artifact is required", ErrInvalid)
	}
	options.TimingMode = strings.TrimSpace(options.TimingMode)
	if options.TimingMode == "" {
		options.TimingMode = "speed"
	}
	if options.TimingMode != "speed" && options.TimingMode != "audio" {
		return MergeOptions{}, fmt.Errorf("%w: unsupported merge timing mode", ErrInvalid)
	}
	if options.Speed == 0 {
		options.Speed = 1
	}
	if options.Speed < 0.5 || options.Speed > 4 {
		return MergeOptions{}, fmt.Errorf("%w: merge speed must be between 0.5 and 4", ErrInvalid)
	}
	if options.TimingMode == "audio" {
		if options.AudioDurationSeconds <= 0 {
			return MergeOptions{}, fmt.Errorf("%w: actual audio duration is required when following audio", ErrInvalid)
		}
		// Follow-audio speed is computed from recorded output duration below.
		// TTSSpeed is retained only as a compatibility input and must not be
		// multiplied into this value.
		options.TTSSpeed = 0
	}
	return options, nil
}

func (s *MergeService) repository() (MergeRepository, error) {
	if s == nil || s.Store == nil {
		return nil, ErrUnavailable
	}
	repository, ok := s.Store.(MergeRepository)
	if !ok {
		return nil, ErrUnavailable
	}
	return repository, nil
}

func normalizeMergeState(value MergeState) MergeState {
	switch value {
	case MergeQueued, MergeRunning, MergeSucceeded, MergeFailed:
		return value
	default:
		return MergeRunning
	}
}

func (s *MergeService) SubmitBatchMerge(ctx context.Context, owner, batchID, requestID string, options MergeOptions) (MergeJob, error) {
	return s.submitMerge(ctx, owner, batchID, "", requestID, options, 0)
}

// SubmitBookMerge produces one final video per opening variant for exactly one
// novel. It shares the same source-selection contract as batch merge but never
// mixes another book's storyboard media into this book's deliverable. Without
// successful opening variants it returns exactly one job, matching the legacy
// behavior.
func (s *MergeService) SubmitBookMerge(ctx context.Context, owner, batchID, bookID, requestID string, options MergeOptions) ([]MergeJob, error) {
	if strings.TrimSpace(bookID) == "" {
		return nil, fmt.Errorf("%w: book id is required", ErrInvalid)
	}
	bookID = strings.TrimSpace(bookID)
	if s == nil || !s.Enabled {
		return nil, fmt.Errorf("%w: merge is not enabled", ErrUnavailable)
	}
	if s.Store == nil {
		return nil, ErrUnavailable
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	var book Book
	found := false
	for _, candidate := range batch.Books {
		if candidate.ID == bookID {
			book = candidate
			found = true
			break
		}
	}
	indices := []int{0}
	if found && len(book.Videos) >= 2 && rawBool(ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch), "openingEnabled", false) {
		for _, variant := range successfulOpeningVariants(book.Videos[0]) {
			indices = append(indices, variant.Index)
		}
	}
	if len(indices) > 1 {
		// A half-ready variant set must not leave partial merge jobs behind:
		// validate every variant's opening media before creating any job.
		if err := s.validateOpeningVariantMedia(ctx, owner, batchID, book, indices); err != nil {
			return nil, err
		}
	}
	jobs := make([]MergeJob, 0, len(indices))
	for _, variantIndex := range indices {
		roundRequestID := requestID
		if variantIndex > 0 {
			roundRequestID = requestID + "-v" + strconv.Itoa(variantIndex)
		}
		job, submitErr := s.submitMerge(ctx, owner, batchID, bookID, roundRequestID, options, variantIndex)
		if submitErr != nil {
			return nil, submitErr
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// validateOpeningVariantMedia proves every opening variant of VIDEO01 has a
// completed production task before SubmitBookMerge creates any merge job.
func (s *MergeService) validateOpeningVariantMedia(ctx context.Context, owner, batchID string, book Book, indices []int) error {
	productionRepository, ok := s.Store.(ProductionRepository)
	if !ok {
		return ErrUnavailable
	}
	if book.DirectorRevision == nil {
		return fmt.Errorf("%w: book %s is not ready for merge", ErrConflict, book.ID)
	}
	productionJobs, err := productionRepository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return err
	}
	tasks := productionTaskSelections(productionJobs, book.ID, book.DirectorRevision.ID)
	allTasks := productionTaskSelectionsAllRevisions(productionJobs, book.ID)
	firstVideo := book.Videos[0]
	for _, variantIndex := range indices {
		candidates := openingVariantTasks(storyboardTaskCandidates(firstVideo, book, tasks, allTasks), variantIndex)
		if _, selectable := selectedProductionTask(firstVideo, candidates); !selectable {
			return fmt.Errorf("%w: book %s has VIDEOs without completed media", ErrConflict, book.ID)
		}
	}
	return nil
}

func openingVariantTasks(candidates []productionTaskSelection, variantIndex int) []productionTaskSelection {
	out := make([]productionTaskSelection, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Task.OpeningVariantIndex == variantIndex {
			out = append(out, candidate)
		}
	}
	return out
}

// RegisterUploadedMerge records an already uploaded HTTPS MP4 as a finished
// final-merge candidate for one novel. Unlike storyboard merges it has no
// VIDEO sources; the uploaded file is the whole deliverable.
func (s *MergeService) RegisterUploadedMerge(ctx context.Context, owner, batchID, bookID, requestID, mediaURL string) (MergeJob, error) {
	if s == nil || s.Store == nil {
		return MergeJob{}, ErrUnavailable
	}
	if strings.TrimSpace(requestID) == "" {
		return MergeJob{}, fmt.Errorf("%w: request id is required", ErrInvalid)
	}
	parsed, err := url.Parse(strings.TrimSpace(mediaURL))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return MergeJob{}, fmt.Errorf("%w: uploaded media URL must be HTTPS", ErrInvalid)
	}
	repository, err := s.repository()
	if err != nil {
		return MergeJob{}, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return MergeJob{}, err
	}
	found := false
	for _, book := range batch.Books {
		if book.ID == bookID {
			found = true
			break
		}
	}
	if !found {
		return MergeJob{}, ErrNotFound
	}
	now := time.Now().UTC()
	return repository.CreateUploadedMergeJob(ctx, MergeJob{Owner: owner, BatchID: batchID, BookID: bookID, RequestID: requestID, Status: MergeSucceeded, OutputURL: parsed.String(), CreatedAt: now, UpdatedAt: now})
}

func sameMergeVersion(left, right MergeJob) bool {
	if left.BookID != right.BookID || left.OpeningVariantIndex != right.OpeningVariantIndex {
		return false
	}
	if left.AudioSource != right.AudioSource || left.AudioArtifactID != right.AudioArtifactID {
		return false
	}
	return left.Speed-right.Speed < 0.000001 && right.Speed-left.Speed < 0.000001
}

func (s *MergeService) mergeVersion(ctx context.Context, repository MergeRepository, owner, batchID string, candidate MergeJob) (MergeJob, bool, error) {
	jobs, err := repository.ListMergeJobs(ctx, owner, batchID)
	if err != nil {
		return MergeJob{}, false, err
	}
	var failed MergeJob
	for _, job := range jobs {
		if !sameMergeVersion(job, candidate) {
			continue
		}
		if job.Status == MergeSucceeded && strings.TrimSpace(job.OutputURL) != "" {
			return job, true, nil
		}
		if job.Status == MergeQueued || job.Status == MergeRunning {
			return job, true, nil
		}
		if job.Status == MergeFailed {
			failed = job
		}
	}
	if failed.ID == "" {
		return MergeJob{}, false, nil
	}
	failed.TimingMode = candidate.TimingMode
	failed.AudioSource = candidate.AudioSource
	failed.AudioArtifactID = candidate.AudioArtifactID
	failed.Speed = candidate.Speed
	failed.ProviderTaskID = ""
	failed.Status = MergeQueued
	failed.ProgressPhase = "queued"
	failed.ProgressCurrent = 0
	failed.ProgressTotal = candidate.ProgressTotal
	failed.Sources = candidate.Sources
	failed.OutputURL = ""
	failed.ErrorMessage = ""
	overwritten, err := repository.UpdateMergeJob(ctx, owner, failed.ID, failed)
	if err != nil {
		return MergeJob{}, false, err
	}
	return overwritten, true, nil
}

func (s *MergeService) submitMerge(ctx context.Context, owner, batchID, onlyBookID, requestID string, options MergeOptions, variantIndex int) (MergeJob, error) {
	if s == nil || !s.Enabled {
		return MergeJob{}, fmt.Errorf("%w: merge is not enabled", ErrUnavailable)
	}
	if s.Adapter == nil {
		return MergeJob{}, ErrUnavailable
	}
	if strings.TrimSpace(requestID) == "" {
		return MergeJob{}, fmt.Errorf("%w: request id is required", ErrInvalid)
	}
	options, err := normalizeMergeOptions(options)
	if err != nil {
		return MergeJob{}, err
	}
	repository, err := s.repository()
	if err != nil {
		return MergeJob{}, err
	}
	if existing, findErr := repository.FindMergeJob(ctx, owner, batchID, requestID); findErr == nil {
		return existing, nil
	} else if !errors.Is(findErr, ErrNotFound) {
		return MergeJob{}, findErr
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return MergeJob{}, err
	}
	productionRepository, ok := s.Store.(ProductionRepository)
	if !ok {
		return MergeJob{}, ErrUnavailable
	}
	productionJobs, err := productionRepository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return MergeJob{}, err
	}
	books := batch.Books
	if onlyBookID != "" {
		books = nil
		for _, book := range batch.Books {
			if book.ID == onlyBookID {
				books = append(books, book)
				break
			}
		}
	}
	if len(books) == 0 {
		return MergeJob{}, fmt.Errorf("%w: book is not ready for merge", ErrConflict)
	}
	sources := []MergeMedia{}
	for _, book := range books {
		if book.DirectorRevision == nil || len(book.Videos) == 0 {
			return MergeJob{}, fmt.Errorf("%w: book %s is not ready for merge", ErrConflict, book.ID)
		}
		tasks := productionTaskSelections(productionJobs, book.ID, book.DirectorRevision.ID)
		// 历史轮次（orphaned 记录）可能是旧 director revision，按当前 revision 过滤会
		// 把它们整批排除；这里另建一份不限 revision 的任务表，只用于同 ordinal 的
		// 「上一版」候选。
		allTasks := productionTaskSelectionsAllRevisions(productionJobs, book.ID)
		mediaVideos := book.Videos
		// Fixed opening keeps all Director cards for review, but merges VIDEO01 only.
		if snapshot, snapshotErr := snapshotForBook(batch, book); snapshotErr != nil {
			return MergeJob{}, snapshotErr
		} else {
			if options.AspectRatio == "" {
				options.AspectRatio = snapshot.AspectRatio
			} else if options.AspectRatio != snapshot.AspectRatio {
				return MergeJob{}, fmt.Errorf("%w: batch merge cannot combine books with different aspect ratios", ErrConflict)
			}
			if snapshot.FixedSingleVideo && len(mediaVideos) > 1 {
				mediaVideos = mediaVideos[:1]
			}
		}
		for _, video := range mediaVideos {
			candidates := storyboardTaskCandidates(video, book, tasks, allTasks)
			// Book merge picks the media produced for this exact opening
			// variant of VIDEO01; every other storyboard keeps the main
			// version selected by the user or the latest success.
			if onlyBookID != "" && video.ID == book.Videos[0].ID {
				candidates = openingVariantTasks(candidates, variantIndex)
			}
			selection, found := selectedProductionTask(video, candidates)
			if !found || selection.Task.Status != ProductionSucceeded || strings.TrimSpace(selection.Task.MediaURL) == "" {
				return MergeJob{}, fmt.Errorf("%w: book %s has VIDEOs without completed media", ErrConflict, book.ID)
			}
			mediaURL := strings.TrimSpace(selection.Task.MediaURL)
			if options.TimingMode == "audio" && selection.Task.ActualDurationSeconds <= 0 {
				if s.DurationProbe == nil {
					return MergeJob{}, fmt.Errorf("%w: book %s VIDEO %s has no actual duration yet", ErrConflict, book.ID, video.ID)
				}
				actualDuration, probeErr := s.DurationProbe.DurationSeconds(ctx, mediaURL)
				if probeErr != nil || actualDuration <= 0 {
					if probeErr == nil {
						probeErr = errors.New("duration probe returned zero")
					}
					return MergeJob{}, fmt.Errorf("%w: book %s VIDEO %s media duration probe failed: %v", ErrUpstream, book.ID, video.ID, probeErr)
				}
				selection.Task.ActualDurationSeconds = actualDuration
				if _, updateErr := productionRepository.UpdateProductionTask(ctx, owner, selection.JobID, selection.Task.ID, selection.Task); updateErr != nil {
					return MergeJob{}, updateErr
				}
			}
			sources = append(sources, MergeMedia{ProductionJobID: selection.JobID, VideoID: video.ID, MediaURL: mediaURL, URL: mediaURL, Order: len(sources), TargetDurationSeconds: selection.Task.TargetDurationSeconds, RequestedDurationSeconds: selection.Task.RequestedDurationSeconds, ActualDurationSeconds: selection.Task.ActualDurationSeconds})
		}
	}
	if options.TimingMode == "audio" {
		actualTotal := 0.0
		for _, source := range sources {
			actualTotal += source.ActualDurationSeconds
		}
		if actualTotal <= 0 {
			return MergeJob{}, fmt.Errorf("%w: actual VIDEO duration is required when following audio", ErrConflict)
		}
		options.Speed = actualTotal / options.AudioDurationSeconds
		if options.Speed < 0.5 || options.Speed > 4 {
			return MergeJob{}, fmt.Errorf("%w: follow-audio speed %.3f is outside 0.5-4", ErrInvalid, options.Speed)
		}
	}
	now := time.Now().UTC()
	candidate := MergeJob{Owner: owner, BatchID: batchID, BookID: onlyBookID, OpeningVariantIndex: variantIndex, RequestID: requestID, TimingMode: options.TimingMode, AudioSource: options.AudioSource, AudioArtifactID: options.AudioArtifactID, Speed: options.Speed, Status: MergeQueued, ProgressPhase: "queued", ProgressTotal: len(sources), Sources: sources, CreatedAt: now, UpdatedAt: now}
	job, reused, err := s.mergeVersion(ctx, repository, owner, batchID, candidate)
	if err != nil {
		return MergeJob{}, err
	}
	if reused && job.Status != MergeQueued {
		return job, nil
	}
	if !reused {
		job, err = repository.CreateMergeJob(ctx, candidate)
		if err != nil {
			return MergeJob{}, err
		}
	}
	result, submitErr := s.Adapter.Submit(ctx, batchID, append([]MergeMedia(nil), sources...), options)
	updated := job
	applyMergeResult(&updated, result)
	if submitErr != nil {
		updated.Status, updated.ProgressPhase, updated.ErrorMessage = MergeFailed, "failed", productionError(submitErr)
	}
	updated, err = repository.UpdateMergeJob(ctx, owner, job.ID, updated)
	if err != nil {
		return MergeJob{}, err
	}
	return updated, nil
}

// GetBatchStatus returns merge jobs owned by the caller. It deliberately
// checks the parent batch first so a missing or cross-owner batch never leaks
// merge history.
func (s *MergeService) GetBatchStatus(ctx context.Context, owner, batchID string) ([]MergeJob, error) {
	// Reading persisted merge jobs does not require a merge provider. This is
	// intentionally available to the staged read-only V12 status endpoint;
	// submit/poll operations still enforce Enabled in their write paths.
	if s == nil {
		return nil, fmt.Errorf("%w: merge is not enabled", ErrUnavailable)
	}
	repository, err := s.repository()
	if err != nil {
		return nil, err
	}
	if err := ensureBatchOwnership(ctx, s.Store, owner, batchID); err != nil {
		return nil, err
	}
	jobs, err := repository.ListMergeJobs(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	if s.Poller == nil {
		return jobs, nil
	}
	for index, job := range jobs {
		if job.ProviderTaskID == "" || (job.Status != MergeQueued && job.Status != MergeRunning) {
			continue
		}
		result, pollErr := s.Poller.Poll(ctx, batchID, job)
		if pollErr != nil {
			updated := job
			updated.Status = MergeFailed
			updated.ProgressPhase = "failed"
			updated.ErrorMessage = productionError(pollErr)
			updated.OutputURL = ""
			if _, updateErr := repository.UpdateMergeJob(ctx, owner, job.ID, updated); updateErr != nil {
				return nil, updateErr
			}
			jobs[index] = updated
			continue
		}
		updated := job
		applyMergeResult(&updated, result)
		if _, updateErr := repository.UpdateMergeJob(ctx, owner, job.ID, updated); updateErr != nil {
			return nil, updateErr
		}
		jobs[index] = updated
	}
	return jobs, nil
}

func applyMergeResult(target *MergeJob, result MergeJob) {
	if target == nil {
		return
	}
	target.Status = normalizeMergeState(result.Status)
	target.ProviderTaskID = strings.TrimSpace(result.ProviderTaskID)
	target.OutputURL = strings.TrimSpace(result.OutputURL)
	target.ErrorMessage = strings.TrimSpace(result.ErrorMessage)
	if strings.TrimSpace(result.ProgressPhase) != "" {
		target.ProgressPhase = strings.TrimSpace(result.ProgressPhase)
	}
	if result.ProgressTotal > 0 {
		target.ProgressTotal = result.ProgressTotal
	}
	if result.ProgressCurrent > 0 || result.ProgressTotal > 0 {
		target.ProgressCurrent = result.ProgressCurrent
	}
	switch target.Status {
	case MergeSucceeded:
		target.ProgressPhase = "completed"
		if target.ProgressTotal > 0 {
			target.ProgressCurrent = target.ProgressTotal
		}
	case MergeFailed:
		target.ProgressPhase = "failed"
	case MergeQueued:
		if target.ProgressPhase == "" {
			target.ProgressPhase = "queued"
		}
	case MergeRunning:
		if target.ProgressPhase == "" || target.ProgressPhase == "queued" {
			target.ProgressPhase = "running"
		}
	}
}

type productionTaskSelection struct {
	JobID string
	Task  ProductionTask
}

func productionTaskSelections(jobs []ProductionJob, bookID, directorRevisionID string) map[string][]productionTaskSelection {
	out := map[string][]productionTaskSelection{}
	for _, job := range jobs {
		if job.BookID != bookID || job.DirectorRevisionID != directorRevisionID {
			continue
		}
		for _, task := range job.Tasks {
			out[task.VideoID] = append(out[task.VideoID], productionTaskSelection{JobID: job.ID, Task: task})
		}
	}
	return out
}

// productionTaskSelectionsAllRevisions groups every task of the book regardless
// of the director revision. Historical (orphaned) records were produced under an
// older revision, so a revision-filtered lookup would hide media that still
// exists and is still playable.
func productionTaskSelectionsAllRevisions(jobs []ProductionJob, bookID string) map[string][]productionTaskSelection {
	out := map[string][]productionTaskSelection{}
	for _, job := range jobs {
		if job.BookID != bookID {
			continue
		}
		for _, task := range job.Tasks {
			out[task.VideoID] = append(out[task.VideoID], productionTaskSelection{JobID: job.ID, Task: task})
		}
	}
	return out
}

// storyboardTaskCandidates returns the media candidates for one active
// storyboard: its own tasks plus tasks that belong to historical video records
// with the same ordinal ("上一版"). Candidates are ordered by update time so the
// latest success remains the default when the user has not pinned a version.
func storyboardTaskCandidates(video Video, book Book, current, all map[string][]productionTaskSelection) []productionTaskSelection {
	candidates := append([]productionTaskSelection{}, current[video.ID]...)
	if len(book.VideoHistory) == 0 {
		sortProductionTaskSelections(candidates)
		return candidates
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		seen[candidate.Task.ID] = true
	}
	for _, history := range book.VideoHistory {
		if history.Ordinal != video.Ordinal {
			continue
		}
		for _, candidate := range all[history.ID] {
			if seen[candidate.Task.ID] {
				continue
			}
			seen[candidate.Task.ID] = true
			candidates = append(candidates, candidate)
		}
	}
	sortProductionTaskSelections(candidates)
	return candidates
}

func sortProductionTaskSelections(items []productionTaskSelection) {
	sort.SliceStable(items, func(left, right int) bool {
		return items[left].Task.UpdatedAt.Before(items[right].Task.UpdatedAt)
	})
}

// selectedProductionTask resolves the user-confirmed main media version. A
// saved primaryMediaTaskId wins; without an explicit choice, the latest
// successful task is the initial main version. Failed/running tasks are never
// eligible for merge, even when an old preference points at one.
func selectedProductionTask(video Video, candidates []productionTaskSelection) (productionTaskSelection, bool) {
	primaryID := rawString(video.SettingsState.Patch, "primaryMediaTaskId", "")
	if primaryID != "" {
		for _, candidate := range candidates {
			if candidate.Task.ID == primaryID && candidate.Task.Status == ProductionSucceeded && strings.TrimSpace(candidate.Task.MediaURL) != "" {
				return candidate, true
			}
		}
	}
	for index := len(candidates) - 1; index >= 0; index-- {
		candidate := candidates[index]
		if candidate.Task.Status == ProductionSucceeded && strings.TrimSpace(candidate.Task.MediaURL) != "" {
			return candidate, true
		}
	}
	return productionTaskSelection{}, false
}
