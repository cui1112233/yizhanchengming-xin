package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type memoryOwned[T any] struct {
	Owner string
	Value T
}
type MemoryStore struct {
	mu                    sync.Mutex
	seq                   int64
	batches               map[string]memoryOwned[Batch]
	intakes               map[string]memoryOwned[Intake]
	patches               map[string]SettingsPatch
	configVersions        map[string]memoryOwned[ConfigVersion]
	prompts               map[string][]Prompt
	drafts                map[string]Draft
	bookAssets            map[string]memoryOwned[BookAsset]
	bookAssetImages       map[string]memoryOwned[BookAssetImage]
	shuihuoConfigs        map[string]map[string]any
	hooks                 map[string][]HookRevision
	directors             map[string][]DirectorRevision
	productionJobs        map[string]memoryOwned[ProductionJob]
	productionRequests    map[string]string
	hiddenProductionTasks map[string]bool
	bookStageRuns         map[string]memoryOwned[BookStageRun]
	mergeJobs             map[string]memoryOwned[MergeJob]
	mergeRequests         map[string]string
	h3AudioMeasurements   map[string]memoryOwned[H3AudioMeasurementRevision]
	h3Timelines           map[string]memoryOwned[H3CanonicalTimelineRevision]
	h3Compilations        map[string]memoryOwned[H3VideoCompilationRevision]
}

func NewMemoryStore() *MemoryStore {
	systemDefault := ConfigVersion{ID: "system-default-v1", Name: "System Default", Config: json.RawMessage(`{"source":"system"}`)}
	return &MemoryStore{
		batches:               map[string]memoryOwned[Batch]{},
		intakes:               map[string]memoryOwned[Intake]{},
		patches:               map[string]SettingsPatch{},
		configVersions:        map[string]memoryOwned[ConfigVersion]{systemDefault.ID: {Owner: "", Value: systemDefault}},
		prompts:               map[string][]Prompt{},
		drafts:                map[string]Draft{},
		bookAssets:            map[string]memoryOwned[BookAsset]{},
		bookAssetImages:       map[string]memoryOwned[BookAssetImage]{},
		shuihuoConfigs:        map[string]map[string]any{},
		hooks:                 map[string][]HookRevision{},
		directors:             map[string][]DirectorRevision{},
		productionJobs:        map[string]memoryOwned[ProductionJob]{},
		productionRequests:    map[string]string{},
		hiddenProductionTasks: map[string]bool{},
		bookStageRuns:         map[string]memoryOwned[BookStageRun]{},
		mergeJobs:             map[string]memoryOwned[MergeJob]{},
		mergeRequests:         map[string]string{},
		h3AudioMeasurements:   map[string]memoryOwned[H3AudioMeasurementRevision]{},
		h3Timelines:           map[string]memoryOwned[H3CanonicalTimelineRevision]{},
		h3Compilations:        map[string]memoryOwned[H3VideoCompilationRevision]{},
	}
}

func productionRequestKey(owner, batchID, bookID, requestID string) string {
	return owner + ":" + batchID + ":" + bookID + ":" + requestID
}

func cloneProductionJob(value ProductionJob) ProductionJob {
	value.Tasks = append([]ProductionTask(nil), value.Tasks...)
	for index := range value.Tasks {
		value.Tasks[index].ReferenceImageURLs = append([]string(nil), value.Tasks[index].ReferenceImageURLs...)
		value.Tasks[index].DowngradedAssetIDs = append([]string(nil), value.Tasks[index].DowngradedAssetIDs...)
		value.Tasks[index].CompileTrace = cloneH3CompileTrace(value.Tasks[index].CompileTrace)
	}
	return value
}

func productionJobState(tasks []ProductionTask) ProductionState {
	if len(tasks) == 0 {
		return ProductionQueued
	}
	allQueued, allSucceeded, allCancelled, hasPending, hasFailed, hasCancelled := true, true, true, false, false, false
	for _, task := range tasks {
		switch task.Status {
		case ProductionQueued:
			hasPending, allSucceeded, allCancelled = true, false, false
		case ProductionRunning:
			hasPending, allQueued, allSucceeded, allCancelled = true, false, false, false
		case ProductionSucceeded:
			allQueued, allCancelled = false, false
		case ProductionFailed:
			allQueued, allSucceeded, allCancelled, hasFailed = false, false, false, true
		case ProductionCancelled:
			allQueued, allSucceeded, hasCancelled = false, false, true
		default:
			allQueued, allSucceeded, allCancelled, hasPending = false, false, false, true
		}
	}
	if allQueued {
		return ProductionQueued
	}
	if hasPending {
		return ProductionRunning
	}
	if allSucceeded {
		return ProductionSucceeded
	}
	if allCancelled {
		return ProductionCancelled
	}
	if hasFailed {
		return ProductionFailed
	}
	if hasCancelled {
		return ProductionCancelled
	}
	return ProductionRunning
}

func (s *MemoryStore) FindProductionJob(_ context.Context, owner, batchID, bookID, requestID string) (ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.productionRequests[productionRequestKey(owner, batchID, bookID, requestID)]
	if !ok {
		return ProductionJob{}, ErrNotFound
	}
	job, ok := s.productionJobs[id]
	if !ok || job.Owner != owner {
		return ProductionJob{}, ErrNotFound
	}
	return cloneProductionJob(job.Value), nil
}

func (s *MemoryStore) CreateProductionJob(_ context.Context, value ProductionJob) (ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[value.BatchID]
	if !ok || batch.Owner != value.Owner {
		return ProductionJob{}, ErrNotFound
	}
	key := productionRequestKey(value.Owner, value.BatchID, value.BookID, value.RequestID)
	if id, exists := s.productionRequests[key]; exists {
		return cloneProductionJob(s.productionJobs[id].Value), nil
	}
	value.ID = s.id("production")
	now := time.Now().UTC()
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
	for index := range value.Tasks {
		value.Tasks[index].ID = s.id("production-task")
	}
	value.Status = productionJobState(value.Tasks)
	s.productionJobs[value.ID] = memoryOwned[ProductionJob]{Owner: value.Owner, Value: cloneProductionJob(value)}
	s.productionRequests[key] = value.ID
	return cloneProductionJob(value), nil
}

func (s *MemoryStore) UpdateProductionTask(_ context.Context, owner, jobID, taskID string, task ProductionTask) (ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.productionJobs[jobID]
	if !ok || owned.Owner != owner {
		return ProductionJob{}, ErrNotFound
	}
	job := owned.Value
	found := false
	for index := range job.Tasks {
		if job.Tasks[index].ID == taskID {
			job.Tasks[index] = task
			found = true
			break
		}
	}
	if !found {
		return ProductionJob{}, ErrNotFound
	}
	job.Status, job.UpdatedAt = productionJobState(job.Tasks), time.Now().UTC()
	s.productionJobs[jobID] = memoryOwned[ProductionJob]{Owner: owner, Value: cloneProductionJob(job)}
	return cloneProductionJob(job), nil
}

func (s *MemoryStore) HideProductionTask(_ context.Context, owner, batchID, bookID, videoID, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, owned := range s.productionJobs {
		job := owned.Value
		if owned.Owner != owner || job.BatchID != batchID || job.BookID != bookID {
			continue
		}
		for _, task := range job.Tasks {
			if task.ID != taskID || task.VideoID != videoID {
				continue
			}
			if task.Status != ProductionSucceeded || strings.TrimSpace(task.MediaURL) == "" {
				return ErrConflict
			}
			s.hiddenProductionTasks[taskID] = true
			return nil
		}
	}
	return ErrNotFound
}

func (s *MemoryStore) ListProductionJobs(_ context.Context, owner, batchID string) ([]ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []ProductionJob{}
	for _, owned := range s.productionJobs {
		if owned.Owner == owner && owned.Value.BatchID == batchID {
			job := cloneProductionJob(owned.Value)
			job.Tasks = slices.DeleteFunc(job.Tasks, func(task ProductionTask) bool { return s.hiddenProductionTasks[task.ID] })
			out = append(out, job)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		left, leftOK := memoryIDSequence(out[i].ID)
		right, rightOK := memoryIDSequence(out[j].ID)
		if leftOK && rightOK && left != right {
			return left < right
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func memoryIDSequence(id string) (int64, bool) {
	separator := strings.LastIndex(id, "-")
	if separator < 0 || separator == len(id)-1 {
		return 0, false
	}
	sequence, err := strconv.ParseInt(id[separator+1:], 10, 64)
	return sequence, err == nil
}

func (s *MemoryStore) CreateBookStageRun(_ context.Context, value BookStageRun) (BookStageRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, err := normalizeBookStageRun(value)
	if err != nil {
		return BookStageRun{}, err
	}
	batch, ok := s.batches[value.BatchID]
	if !ok || batch.Owner != value.Owner {
		return BookStageRun{}, ErrNotFound
	}
	found := false
	for _, book := range batch.Value.Books {
		if book.ID == value.BookID {
			found = true
			break
		}
	}
	if !found {
		return BookStageRun{}, ErrNotFound
	}
	value.ID = s.id("book-stage")
	now := time.Now().UTC()
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = value.CreatedAt
	}
	s.bookStageRuns[value.ID] = memoryOwned[BookStageRun]{Owner: value.Owner, Value: value}
	return value, nil
}

func (s *MemoryStore) UpdateBookStageRun(_ context.Context, owner, id string, value BookStageRun) (BookStageRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.bookStageRuns[id]
	if !ok || owned.Owner != owner {
		return BookStageRun{}, ErrNotFound
	}
	next := owned.Value
	if value.Status != "" {
		next.Status = normalizeProductionState(value.Status)
	}
	if value.ErrorMessage != "" || next.Status == ProductionSucceeded {
		next.ErrorMessage = productionError(errText(value.ErrorMessage))
	}
	if value.InputRevision != "" {
		next.InputRevision = value.InputRevision
	}
	next.UpdatedAt = time.Now().UTC()
	s.bookStageRuns[id] = memoryOwned[BookStageRun]{Owner: owner, Value: next}
	return next, nil
}

func (s *MemoryStore) ListBookStageRuns(_ context.Context, owner, batchID, bookID string) ([]BookStageRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[batchID]
	if !ok || batch.Owner != owner {
		return nil, ErrNotFound
	}
	found := false
	for _, book := range batch.Value.Books {
		if book.ID == bookID {
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNotFound
	}
	out := []BookStageRun{}
	for _, owned := range s.bookStageRuns {
		if owned.Owner == owner && owned.Value.BatchID == batchID && owned.Value.BookID == bookID {
			out = append(out, owned.Value)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].UpdatedAt.Before(out[j].UpdatedAt) })
	return out, nil
}

func cloneMergeJob(value MergeJob) MergeJob {
	value.Sources = append([]MergeMedia(nil), value.Sources...)
	return value
}

func mergeRequestKey(owner, batchID, requestID string) string {
	return owner + ":" + batchID + ":" + requestID
}

func (s *MemoryStore) FindMergeJob(_ context.Context, owner, batchID, requestID string) (MergeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.mergeRequests[mergeRequestKey(owner, batchID, requestID)]
	if !ok {
		return MergeJob{}, ErrNotFound
	}
	job, ok := s.mergeJobs[id]
	if !ok || job.Owner != owner {
		return MergeJob{}, ErrNotFound
	}
	return cloneMergeJob(job.Value), nil
}

func (s *MemoryStore) CreateMergeJob(_ context.Context, value MergeJob) (MergeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.batches[value.BatchID]
	if !ok || batch.Owner != value.Owner {
		return MergeJob{}, ErrNotFound
	}
	key := mergeRequestKey(value.Owner, value.BatchID, value.RequestID)
	if id, exists := s.mergeRequests[key]; exists {
		return cloneMergeJob(s.mergeJobs[id].Value), nil
	}
	value.ID = s.id("merge")
	value.Status = normalizeMergeState(value.Status)
	s.mergeJobs[value.ID] = memoryOwned[MergeJob]{Owner: value.Owner, Value: cloneMergeJob(value)}
	s.mergeRequests[key] = value.ID
	return cloneMergeJob(value), nil
}

func (s *MemoryStore) CreateUploadedMergeJob(ctx context.Context, value MergeJob) (MergeJob, error) {
	return s.CreateMergeJob(ctx, value)
}

func (s *MemoryStore) UpdateMergeJob(_ context.Context, owner, jobID string, value MergeJob) (MergeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.mergeJobs[jobID]
	if !ok || owned.Owner != owner {
		return MergeJob{}, ErrNotFound
	}
	value.ID, value.Owner = jobID, owner
	value.CreatedAt = owned.Value.CreatedAt
	value.UpdatedAt = time.Now().UTC()
	s.mergeJobs[jobID] = memoryOwned[MergeJob]{Owner: owner, Value: cloneMergeJob(value)}
	return cloneMergeJob(value), nil
}

func (s *MemoryStore) ListMergeJobs(_ context.Context, owner, batchID string) ([]MergeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []MergeJob{}
	for _, owned := range s.mergeJobs {
		if owned.Owner == owner && owned.Value.BatchID == batchID {
			out = append(out, cloneMergeJob(owned.Value))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (s *MemoryStore) id(prefix string) string { s.seq++; return fmt.Sprintf("%s-%d", prefix, s.seq) }
func scopeKey(r ScopeRef) string {
	return string(r.Kind) + ":" + r.BatchID + ":" + r.BookID + ":" + r.VideoID
}
func (s *MemoryStore) DebugPatch(r ScopeRef) SettingsPatch {
	s.mu.Lock()
	defer s.mu.Unlock()
	return clonePatch(s.patches[scopeKey(r)])
}

func hydrateMemoryBatchSettingsState(b Batch, patches map[string]SettingsPatch, hooks map[string][]HookRevision, directors map[string][]DirectorRevision, bookAssets map[string]memoryOwned[BookAsset]) Batch {
	b.SettingsState = SettingsState{
		Patch:    clonePatch(patches[scopeKey(ScopeRef{Kind: ScopeBatch, BatchID: b.ID})]),
		Revision: b.Revision,
	}
	books := make([]Book, len(b.Books))
	for i := range b.Books {
		book := b.Books[i]
		book.SettingsState = SettingsState{
			Patch:    clonePatch(patches[scopeKey(ScopeRef{Kind: ScopeBook, BatchID: b.ID, BookID: book.ID})]),
			Revision: book.Revision,
		}
		effective := ResolveSettings(b.SettingsState.Patch, book.SettingsState.Patch)
		book.Mode = rawString(effective, "productionMode", rawString(effective, "mode", "original"))
		if book.Mode == "original_direct" {
			book.Mode = "original"
		}
		if book.Mode == "viral_hook" {
			book.Mode = "viral"
		}
		bookKey := b.ID + ":" + book.ID
		if revisions := hooks[bookKey]; len(revisions) > 0 {
			latest := revisions[len(revisions)-1]
			book.Hook = &latest
		}
		if revisions := directors[bookKey]; len(revisions) > 0 {
			latest := revisions[len(revisions)-1]
			book.DirectorRevision = &latest
			book.Assets = DirectorAssets{Characters: latest.Output.Characters, Scenes: latest.Output.Scenes, Props: latest.Output.Props}
		}
		book.AssetRecords = memoryBookAssets(bookAssets, b.ID, book.ID)
		if len(book.AssetRecords) > 0 {
			book.Assets = directorAssetsFromRecords(book.AssetRecords)
		}
		videos := make([]Video, len(book.Videos))
		for j := range book.Videos {
			video := book.Videos[j]
			video.SettingsState = SettingsState{
				Patch:    clonePatch(patches[scopeKey(ScopeRef{Kind: ScopeVideo, BatchID: b.ID, BookID: book.ID, VideoID: video.ID})]),
				Revision: video.Revision,
			}
			applyVideoPromptOverrides(&video)
			if _, overridden := optionalPromptValue(video.SettingsState.Patch, "videoPrompt"); !overridden && book.DirectorRevision != nil && j < len(book.DirectorRevision.Output.Storyboard) {
				video.VideoPrompt = storyboardVideoPrompt(book.DirectorRevision.Output.Storyboard[j])
			}
			videos[j] = video
		}
		book.Videos = videos
		books[i] = book
	}
	b.Books = books
	return b
}

func memoryBookAssets(all map[string]memoryOwned[BookAsset], batchID, bookID string) []BookAsset {
	assets := make([]BookAsset, 0)
	for _, owned := range all {
		asset := owned.Value
		if asset.BatchID == batchID && asset.BookID == bookID {
			assets = append(assets, asset)
		}
	}
	sort.Slice(assets, func(i, j int) bool {
		if assets[i].Kind == assets[j].Kind {
			return assets[i].Name < assets[j].Name
		}
		return assets[i].Kind < assets[j].Kind
	})
	return assets
}

func directorAssetsFromRecords(records []BookAsset) DirectorAssets {
	assets := DirectorAssets{Characters: []NamedPrompt{}, Scenes: []NamedPrompt{}, Props: []NamedPrompt{}}
	for _, asset := range records {
		value := NamedPrompt{Name: asset.Name, Prompt: asset.Prompt}
		switch asset.Kind {
		case "character":
			assets.Characters = append(assets.Characters, value)
		case "scene":
			assets.Scenes = append(assets.Scenes, value)
		case "prop":
			assets.Props = append(assets.Props, value)
		}
	}
	return assets
}

func (s *MemoryStore) CreateIntake(_ context.Context, owner string, input NovelFetchIntakeInput) (Intake, error) {
	input = normalizeNovelFetchIntake(input)
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.Marshal(input)
	now := time.Now().UTC()
	v := Intake{ID: s.id("intake"), Owner: owner, Payload: b, CreatedAt: now}
	s.intakes[v.ID] = memoryOwned[Intake]{owner, v}
	return v, nil
}
func (s *MemoryStore) GetIntake(_ context.Context, owner, id string) (Intake, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.intakes[id]
	if !ok || v.Owner != owner {
		return Intake{}, ErrNotFound
	}
	return v.Value, nil
}
func (s *MemoryStore) CreateBatchFromIntake(ctx context.Context, owner, intakeID string, input CreateBatchInput) (Batch, error) {
	s.mu.Lock()
	owned, ok := s.intakes[intakeID]
	if !ok || owned.Owner != owner {
		s.mu.Unlock()
		return Batch{}, ErrNotFound
	}
	if owned.Value.ConsumedAt != nil {
		s.mu.Unlock()
		return Batch{}, ErrConflict
	}
	var intake NovelFetchIntakeInput
	_ = json.Unmarshal(owned.Value.Payload, &intake)
	if len(input.Books) == 0 {
		input.Books = intake.Books
	}
	if input.Title == "" {
		input.Title = "Novel Fetch Batch"
	}
	s.mu.Unlock()
	batch, err := s.CreateBatch(ctx, owner, input)
	if err != nil {
		return Batch{}, err
	}
	s.mu.Lock()
	now := time.Now().UTC()
	v := owned.Value
	v.ConsumedAt = &now
	s.intakes[intakeID] = memoryOwned[Intake]{owner, v}
	b := s.batches[batch.ID]
	bv := b.Value
	bv.SourceIntakeID = intakeID
	s.batches[batch.ID] = memoryOwned[Batch]{owner, bv}
	batch = hydrateMemoryBatchSettingsState(bv, s.patches, s.hooks, s.directors, s.bookAssets)
	s.mu.Unlock()
	return batch, nil
}

// AppendBooksFromIntake adds only new source books to the selected batch. The
// intake remains single-use even when every incoming book was already present.
func (s *MemoryStore) AppendBooksFromIntake(_ context.Context, owner, batchID, intakeID string, _ bool) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ownedIntake, ok := s.intakes[intakeID]
	if !ok || ownedIntake.Owner != owner {
		return Batch{}, ErrNotFound
	}
	if ownedIntake.Value.ConsumedAt != nil {
		return Batch{}, ErrConflict
	}
	ownedBatch, ok := s.batches[batchID]
	if !ok || ownedBatch.Owner != owner {
		return Batch{}, ErrNotFound
	}
	var intake NovelFetchIntakeInput
	if err := json.Unmarshal(ownedIntake.Value.Payload, &intake); err != nil {
		return Batch{}, ErrInvalid
	}
	batch := ownedBatch.Value
	existing := map[string]bool{}
	for _, book := range batch.Books {
		if key := batchBookKey(CreateBookInput{BookID: book.BookID, Platform: book.Platform}); key != "" {
			existing[key] = true
		}
	}
	books := uniqueBatchBooks(intake.Books, existing)
	now := time.Now().UTC()
	for _, bookInput := range books {
		bookID := s.id("book")
		sourceID := sourceBookID(bookInput)
		if sourceID == "" {
			sourceID = bookID
		}
		batch.Books = append(batch.Books, Book{ID: bookID, BatchID: batch.ID, BookID: sourceID, Title: bookInput.Title, SourceText: bookInput.SourceText, SourceTaskID: bookInput.SourceTaskID, Platform: bookInput.Platform, TxtText: bookInput.TxtText, TxtFileName: bookInput.TxtFileName, SourceMetadata: bookInput.SourceMetadata, Revision: 1, Videos: []Video{}})
	}
	if len(books) > 0 {
		batch.Revision++
		batch.UpdatedAt = now
	}
	s.batches[batchID] = memoryOwned[Batch]{Owner: owner, Value: batch}
	intakeRecord := ownedIntake.Value
	intakeRecord.ConsumedAt = &now
	s.intakes[intakeID] = memoryOwned[Intake]{Owner: owner, Value: intakeRecord}
	return hydrateMemoryBatchSettingsState(batch, s.patches, s.hooks, s.directors, s.bookAssets), nil
}
func (s *MemoryStore) CreateBatch(_ context.Context, owner string, input CreateBatchInput) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	b := Batch{ID: s.id("batch"), Title: input.Title, Revision: 1, Books: []Book{}, CreatedAt: now, UpdatedAt: now}
	if b.Title == "" {
		b.Title = "Untitled Batch"
	}
	for _, bi := range uniqueBatchBooks(input.Books, map[string]bool{}) {
		if len(input.GiantAutomationPlan) > 0 {
			bi.SourceMetadata = copySourceMetadataWithGiantPlan(bi.SourceMetadata, input.GiantAutomationPlan)
		}
		internalBookID := s.id("book")
		sourceID := sourceBookID(bi)
		if sourceID == "" {
			sourceID = internalBookID
		}
		book := Book{
			ID: internalBookID, BatchID: b.ID, BookID: sourceID, Title: bi.Title,
			SourceText: bi.SourceText, SourceTaskID: bi.SourceTaskID, Platform: bi.Platform,
			TxtText: bi.TxtText, TxtFileName: bi.TxtFileName, SourceMetadata: bi.SourceMetadata,
			Revision: 1, Videos: []Video{},
		}
		for _, vi := range bi.Videos {
			video := Video{ID: s.id("video"), BatchID: b.ID, BookID: book.ID, Label: vi.Label, VideoPrompt: vi.VideoPrompt, VisualPrompt: vi.VisualPrompt, DurationSeconds: vi.DurationSeconds, CompatibilityState: "active", Revision: 1}
			book.Videos = append(book.Videos, video)
		}
		b.Books = append(b.Books, book)
	}
	if len(input.InitialBatchSettings) > 0 {
		b.Revision++
		s.patches[scopeKey(ScopeRef{Kind: ScopeBatch, BatchID: b.ID})] = cloneSettingsPatch(input.InitialBatchSettings)
	}
	s.batches[b.ID] = memoryOwned[Batch]{owner, b}
	return hydrateMemoryBatchSettingsState(b, s.patches, s.hooks, s.directors, s.bookAssets), nil
}
func (s *MemoryStore) ListBatches(_ context.Context, owner string) ([]Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Batch{}
	for _, v := range s.batches {
		if v.Owner == owner {
			out = append(out, hydrateMemoryBatchSettingsState(v.Value, s.patches, s.hooks, s.directors, s.bookAssets))
		}
	}
	return out, nil
}
func (s *MemoryStore) UpdateBookMetadata(_ context.Context, owner, batchID, bookID string, input UpdateBookMetadataInput) (Book, error) {
	if input.ExpectedRevision < 1 {
		return Book{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return Book{}, ErrNotFound
	}
	batch := owned.Value
	for index := range batch.Books {
		book := &batch.Books[index]
		if book.ID != bookID {
			continue
		}
		if book.Revision != input.ExpectedRevision {
			return Book{}, ErrConflict
		}
		book.SourceMetadata = map[string]any{}
		for key, value := range input.Metadata {
			book.SourceMetadata[key] = value
		}
		book.Revision++
		batch.UpdatedAt = time.Now().UTC()
		s.batches[batchID] = memoryOwned[Batch]{owner, batch}
		return *book, nil
	}
	return Book{}, ErrNotFound
}

func (s *MemoryStore) CaptureBookSource(_ context.Context, owner, batchID, bookID string, input CaptureBookSourceInput) (Book, error) {
	if input.ExpectedRevision < 1 || strings.TrimSpace(input.SourceText) == "" {
		return Book{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return Book{}, ErrNotFound
	}
	batch := owned.Value
	for index := range batch.Books {
		book := &batch.Books[index]
		if book.ID != bookID {
			continue
		}
		if book.Revision != input.ExpectedRevision || (strings.TrimSpace(book.SourceText) != "" && !input.ReplaceSource) {
			return Book{}, ErrConflict
		}
		book.SourceText = strings.TrimSpace(input.SourceText)
		book.TxtText = book.SourceText
		if shouldReplaceGeneratedBookTitle(book.Title, book.BookID, input.SourceTitle) {
			book.Title = strings.TrimSpace(input.SourceTitle)
		}
		if book.SourceMetadata == nil {
			book.SourceMetadata = map[string]any{}
		}
		for key, value := range input.SourceMetadata {
			book.SourceMetadata[key] = value
		}
		book.Revision++
		batch.UpdatedAt = time.Now().UTC()
		s.batches[batchID] = memoryOwned[Batch]{owner, batch}
		return *book, nil
	}
	return Book{}, ErrNotFound
}

func (s *MemoryStore) GetBatch(_ context.Context, owner, id string) (Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.batches[id]
	if !ok || v.Owner != owner {
		return Batch{}, ErrNotFound
	}
	return hydrateMemoryBatchSettingsState(v.Value, s.patches, s.hooks, s.directors, s.bookAssets), nil
}

func (s *MemoryStore) DeleteBook(_ context.Context, owner, batchID, bookID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ErrNotFound
	}
	batch := owned.Value
	found := false
	books := make([]Book, 0, len(batch.Books))
	for _, book := range batch.Books {
		if book.ID == bookID {
			found = true
			continue
		}
		books = append(books, book)
	}
	if !found {
		return ErrNotFound
	}
	batch.Books = books
	batch.Revision++
	batch.UpdatedAt = time.Now().UTC()
	s.batches[batchID] = memoryOwned[Batch]{Owner: owner, Value: batch}
	s.deleteMemoryBookData(owner, batchID, bookID)
	return nil
}

func (s *MemoryStore) DeleteBatch(_ context.Context, owner, batchID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return ErrNotFound
	}
	for _, book := range owned.Value.Books {
		s.deleteMemoryBookData(owner, batchID, book.ID)
	}
	delete(s.batches, batchID)
	for key := range s.patches {
		if strings.Contains(key, ":"+batchID+":") {
			delete(s.patches, key)
		}
	}
	for key, draft := range s.drafts {
		if draft.Scope == batchID {
			delete(s.drafts, key)
		}
	}
	for id, job := range s.mergeJobs {
		if job.Owner == owner && job.Value.BatchID == batchID {
			delete(s.mergeJobs, id)
		}
	}
	return nil
}

func (s *MemoryStore) deleteMemoryBookData(owner, batchID, bookID string) {
	for key := range s.patches {
		if strings.Contains(key, ":"+batchID+":"+bookID+":") {
			delete(s.patches, key)
		}
	}
	for key, assets := range s.bookAssets {
		if assets.Owner == owner && assets.Value.BatchID == batchID && assets.Value.BookID == bookID {
			delete(s.bookAssets, key)
		}
	}
	for key, hooks := range s.hooks {
		if len(hooks) > 0 && hooks[0].BatchID == batchID && hooks[0].BookID == bookID {
			delete(s.hooks, key)
		}
	}
	for key, directors := range s.directors {
		if len(directors) > 0 && directors[0].BatchID == batchID && directors[0].BookID == bookID {
			delete(s.directors, key)
		}
	}
	for id, job := range s.productionJobs {
		if job.Owner == owner && job.Value.BatchID == batchID && job.Value.BookID == bookID {
			delete(s.productionJobs, id)
		}
	}
	for key := range s.productionRequests {
		if strings.Contains(key, ":"+batchID+":"+bookID+":") {
			delete(s.productionRequests, key)
		}
	}
	for id, run := range s.bookStageRuns {
		if run.Owner == owner && run.Value.BatchID == batchID && run.Value.BookID == bookID {
			delete(s.bookStageRuns, id)
		}
	}
}
func (s *MemoryStore) SaveSettings(_ context.Context, owner string, ref ScopeRef, update SettingsUpdate) (SettingsResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[ref.BatchID]
	if !ok || owned.Owner != owner {
		return SettingsResult{}, ErrNotFound
	}
	b := owned.Value
	var rev *int64
	switch ref.Kind {
	case ScopeBatch:
		rev = &b.Revision
	case ScopeBook:
		for i := range b.Books {
			if b.Books[i].ID == ref.BookID {
				rev = &b.Books[i].Revision
				break
			}
		}
	case ScopeVideo:
		for i := range b.Books {
			if b.Books[i].ID != ref.BookID {
				continue
			}
			for j := range b.Books[i].Videos {
				if b.Books[i].Videos[j].ID == ref.VideoID {
					rev = &b.Books[i].Videos[j].Revision
					break
				}
			}
		}
	}
	if rev == nil {
		return SettingsResult{}, ErrNotFound
	}
	if update.ExpectedRevision != *rev {
		return SettingsResult{}, ErrConflict
	}
	versionID, selectsVersion, err := configVersionIDFromUpdate(ref, update)
	if err != nil {
		return SettingsResult{}, err
	}
	if selectsVersion {
		version, exists := s.configVersions[versionID]
		if !exists || (version.Owner != "" && version.Owner != owner) {
			return SettingsResult{}, ErrNotFound
		}
		if isAIReasoningPresetConfig(version.Value.Config) {
			return SettingsResult{}, ErrInvalid
		}
	}
	(*rev)++
	patch := ApplySparseUpdate(s.patches[scopeKey(ref)], update)
	s.patches[scopeKey(ref)] = patch
	b.UpdatedAt = time.Now().UTC()
	s.batches[b.ID] = memoryOwned[Batch]{owner, b}
	effective := SettingsPatch{}
	batchPatch := s.patches[scopeKey(ScopeRef{Kind: ScopeBatch, BatchID: b.ID})]
	effective = ResolveSettings(effective, batchPatch)
	if ref.BookID != "" {
		effective = ResolveSettings(effective, s.patches[scopeKey(ScopeRef{Kind: ScopeBook, BatchID: b.ID, BookID: ref.BookID})])
	}
	if ref.VideoID != "" {
		effective = ResolveSettings(effective, s.patches[scopeKey(ScopeRef{Kind: ScopeVideo, BatchID: b.ID, BookID: ref.BookID, VideoID: ref.VideoID})])
	}
	snap := ConfigSnapshot{ID: s.id("snapshot"), BatchID: b.ID, BookID: ref.BookID, VideoID: ref.VideoID, Effective: effective, CreatedAt: time.Now().UTC()}
	return SettingsResult{Scope: ref, Patch: clonePatch(patch), Revision: *rev, Snapshot: &snap}, nil
}
func (s *MemoryStore) ConfigVersions(_ context.Context, owner string) ([]ConfigVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []ConfigVersion{}
	for _, version := range s.configVersions {
		if version.Owner == "" || version.Owner == owner {
			out = append(out, version.Value)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}
func (s *MemoryStore) CreateConfigVersion(_ context.Context, owner string, version ConfigVersion) (ConfigVersion, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(version.Name) == "" || !validConfigVersionJSON(version.Config) {
		return ConfigVersion{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	version.ID = s.id("config")
	version.Name = strings.TrimSpace(version.Name)
	version.Config = append([]byte(nil), version.Config...)
	version.CreatedAt = time.Now().UTC()
	s.configVersions[version.ID] = memoryOwned[ConfigVersion]{Owner: owner, Value: version}
	return version, nil
}
func (s *MemoryStore) RenameConfigVersion(_ context.Context, owner, id, name string) (ConfigVersion, error) {
	if strings.TrimSpace(name) == "" {
		return ConfigVersion{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.configVersions[id]
	if !ok || owned.Owner != owner {
		return ConfigVersion{}, ErrNotFound
	}
	owned.Value.Name = strings.TrimSpace(name)
	s.configVersions[id] = owned
	return owned.Value, nil
}
func (s *MemoryStore) ChangeImpact(_ context.Context, owner, batchID string, u SettingsUpdate) (ChangeImpact, error) {
	b, err := s.GetBatch(context.Background(), owner, batchID)
	if err != nil {
		return ChangeImpact{}, err
	}
	videos := 0
	for _, book := range b.Books {
		videos += len(book.Videos)
	}
	return ChangeImpact{AffectedBooks: len(b.Books), AffectedVideos: videos, InvalidatesDirector: false, PreservesOverrides: true, ChangedKeys: ChangedKeys(u)}, nil
}
func (s *MemoryStore) ListPrompts(_ context.Context, owner, kind string) ([]Prompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Prompt{}
	for _, p := range s.prompts[owner] {
		if kind == "" || p.Kind == kind {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *MemoryStore) CreatePrompt(_ context.Context, owner string, p Prompt) (Prompt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.ID = s.id("prompt")
	p.VersionID = s.id("prompt-version")
	p.Revision = 1
	p.CreatedAt = time.Now().UTC()
	s.prompts[owner] = append(s.prompts[owner], p)
	return p, nil
}
func draftKey(owner, key, kind, scope string) string {
	return owner + ":" + key + ":" + kind + ":" + scope
}
func (s *MemoryStore) GetDraft(_ context.Context, owner, key, kind, scope string) (Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.drafts[draftKey(owner, key, kind, scope)]
	if !ok {
		return Draft{}, ErrNotFound
	}
	return v, nil
}
func (s *MemoryStore) SaveDraft(_ context.Context, owner string, d Draft) (Draft, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := draftKey(owner, d.Key, d.Kind, d.Scope)
	old := s.drafts[k]
	d.Revision = old.Revision + 1
	if d.Revision == 0 {
		d.Revision = 1
	}
	d.UpdatedAt = time.Now().UTC()
	s.drafts[k] = d
	return d, nil
}

func validBookAssetKind(kind string) bool {
	switch kind {
	case "character", "scene", "prop":
		return true
	default:
		return false
	}
}

func normalizeBookAssetInput(input CreateBookAssetInput) (CreateBookAssetInput, error) {
	input.Kind = strings.TrimSpace(input.Kind)
	input.Name = strings.TrimSpace(input.Name)
	input.Prompt = strings.TrimSpace(input.Prompt)
	if !validBookAssetKind(input.Kind) || input.Name == "" || input.Prompt == "" || len(input.Name) > 255 {
		return CreateBookAssetInput{}, ErrInvalid
	}
	return input, nil
}

func (s *MemoryStore) hasBookLocked(owner, batchID, bookID string) bool {
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return false
	}
	for _, book := range owned.Value.Books {
		if book.ID == bookID {
			return true
		}
	}
	return false
}

func (s *MemoryStore) ListBookAssets(_ context.Context, owner, batchID, bookID string) ([]BookAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) {
		return nil, ErrNotFound
	}
	return memoryBookAssets(s.bookAssets, batchID, bookID), nil
}

func (s *MemoryStore) CreateBookAsset(_ context.Context, owner, batchID, bookID string, input CreateBookAssetInput) (BookAsset, error) {
	input, err := normalizeBookAssetInput(input)
	if err != nil {
		return BookAsset{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) {
		return BookAsset{}, ErrNotFound
	}
	now := time.Now().UTC()
	asset := BookAsset{ID: s.id("asset"), BatchID: batchID, BookID: bookID, Kind: input.Kind, Name: input.Name, Prompt: input.Prompt, Source: "manual", Revision: 1, CreatedAt: now, UpdatedAt: now}
	s.bookAssets[asset.ID] = memoryOwned[BookAsset]{Owner: owner, Value: asset}
	return asset, nil
}

func (s *MemoryStore) UpdateBookAsset(_ context.Context, owner, batchID, bookID, assetID string, input UpdateBookAssetInput) (BookAsset, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Name == "" || input.Prompt == "" || len(input.Name) > 255 {
		return BookAsset{}, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) {
		return BookAsset{}, ErrNotFound
	}
	owned, ok := s.bookAssets[assetID]
	if !ok || owned.Owner != owner || owned.Value.BatchID != batchID || owned.Value.BookID != bookID {
		return BookAsset{}, ErrNotFound
	}
	asset := owned.Value
	if input.ExpectedRevision != asset.Revision {
		return BookAsset{}, ErrConflict
	}
	asset.Name, asset.Prompt, asset.Source, asset.Revision, asset.UpdatedAt = input.Name, input.Prompt, "manual", asset.Revision+1, time.Now().UTC()
	s.bookAssets[assetID] = memoryOwned[BookAsset]{Owner: owner, Value: asset}
	return asset, nil
}

func (s *MemoryStore) DeleteBookAsset(_ context.Context, owner, batchID, bookID, assetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) {
		return ErrNotFound
	}
	owned, ok := s.bookAssets[assetID]
	if !ok || owned.Owner != owner || owned.Value.BatchID != batchID || owned.Value.BookID != bookID {
		return ErrNotFound
	}
	delete(s.bookAssets, assetID)
	for imageID, image := range s.bookAssetImages {
		if image.Owner == owner && image.Value.AssetID == assetID {
			delete(s.bookAssetImages, imageID)
		}
	}
	return nil
}

func validBookAssetImageInput(input CreateBookAssetImageInput) (CreateBookAssetImageInput, error) {
	input.URL, input.StorageRef, input.MediaType, input.Source = strings.TrimSpace(input.URL), strings.TrimSpace(input.StorageRef), strings.TrimSpace(input.MediaType), strings.TrimSpace(input.Source)
	if input.Source == "" {
		input.Source = "provider"
	}
	if input.MediaType == "" {
		input.MediaType = "image/png"
	}
	if input.Source != "provider" && input.Source != "upload" {
		return CreateBookAssetImageInput{}, ErrInvalid
	}
	if input.MediaType == "" || (input.URL == "" && input.StorageRef == "") || (input.URL != "" && input.StorageRef != "") {
		return CreateBookAssetImageInput{}, ErrInvalid
	}
	if input.URL != "" {
		parsed, err := url.Parse(input.URL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https" && !strings.HasPrefix(input.URL, "/api/")) {
			return CreateBookAssetImageInput{}, ErrInvalid
		}
	}
	if input.StorageRef != "" && (filepath.Base(input.StorageRef) != input.StorageRef || strings.Contains(input.StorageRef, "..")) {
		return CreateBookAssetImageInput{}, ErrInvalid
	}
	if !strings.HasPrefix(input.MediaType, "image/") {
		return CreateBookAssetImageInput{}, ErrInvalid
	}
	return input, nil
}

func (s *MemoryStore) hasAssetLocked(owner, batchID, bookID, assetID string) bool {
	owned, ok := s.bookAssets[assetID]
	return ok && owned.Owner == owner && owned.Value.BatchID == batchID && owned.Value.BookID == bookID
}

func (s *MemoryStore) ListBookAssetImages(_ context.Context, owner, batchID, bookID, assetID string) ([]BookAssetImage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) || !s.hasAssetLocked(owner, batchID, bookID, assetID) {
		return nil, ErrNotFound
	}
	images := []BookAssetImage{}
	for _, owned := range s.bookAssetImages {
		if owned.Owner == owner && owned.Value.AssetID == assetID {
			images = append(images, owned.Value)
		}
	}
	sort.Slice(images, func(i, j int) bool { return images[i].CreatedAt.Before(images[j].CreatedAt) })
	return images, nil
}

func (s *MemoryStore) CreateBookAssetImage(_ context.Context, owner, batchID, bookID, assetID string, input CreateBookAssetImageInput) (BookAssetImage, error) {
	input, err := validBookAssetImageInput(input)
	if err != nil {
		return BookAssetImage{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) || !s.hasAssetLocked(owner, batchID, bookID, assetID) {
		return BookAssetImage{}, ErrNotFound
	}
	hasPrimary := false
	for _, owned := range s.bookAssetImages {
		if owned.Owner == owner && owned.Value.AssetID == assetID && owned.Value.IsPrimary {
			hasPrimary = true
			break
		}
	}
	image := BookAssetImage{ID: s.id("asset-image"), AssetID: assetID, URL: input.URL, StorageRef: input.StorageRef, MediaType: input.MediaType, Source: input.Source, IsPrimary: !hasPrimary, Revision: 1, CreatedAt: time.Now().UTC()}
	s.bookAssetImages[image.ID] = memoryOwned[BookAssetImage]{Owner: owner, Value: image}
	return image, nil
}

func (s *MemoryStore) SetPrimaryBookAssetImage(_ context.Context, owner, batchID, bookID, assetID, imageID string) (BookAssetImage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasBookLocked(owner, batchID, bookID) || !s.hasAssetLocked(owner, batchID, bookID, assetID) {
		return BookAssetImage{}, ErrNotFound
	}
	target, ok := s.bookAssetImages[imageID]
	if !ok || target.Owner != owner || target.Value.AssetID != assetID {
		return BookAssetImage{}, ErrNotFound
	}
	for id, owned := range s.bookAssetImages {
		if owned.Owner != owner || owned.Value.AssetID != assetID || !owned.Value.IsPrimary {
			continue
		}
		image := owned.Value
		image.IsPrimary = false
		image.Revision++
		s.bookAssetImages[id] = memoryOwned[BookAssetImage]{Owner: owner, Value: image}
	}
	image := target.Value
	if !image.IsPrimary {
		image.IsPrimary = true
		image.Revision++
	}
	s.bookAssetImages[imageID] = memoryOwned[BookAssetImage]{Owner: owner, Value: image}
	return image, nil
}

func memoryBookKey(batchID, bookID string) string { return batchID + ":" + bookID }

func (s *MemoryStore) CreateHookRevision(_ context.Context, owner, batchID, bookID, text, digest string) (HookRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return HookRevision{}, ErrNotFound
	}
	found := false
	for _, book := range owned.Value.Books {
		if book.ID == bookID {
			found = true
			break
		}
	}
	if !found {
		return HookRevision{}, ErrNotFound
	}
	key := memoryBookKey(batchID, bookID)
	now := time.Now().UTC()
	revision := HookRevision{ID: s.id("hook"), BatchID: batchID, BookID: bookID, Revision: int64(len(s.hooks[key]) + 1), Status: "draft", Text: text, SourceDigest: digest, CreatedAt: now}
	s.hooks[key] = append(s.hooks[key], revision)
	return revision, nil
}

func (s *MemoryStore) ApproveHookRevision(_ context.Context, owner, batchID, bookID, hookID string) (HookRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return HookRevision{}, ErrNotFound
	}
	key := memoryBookKey(batchID, bookID)
	for i := range s.hooks[key] {
		if s.hooks[key][i].ID != hookID {
			continue
		}
		now := time.Now().UTC()
		s.hooks[key][i].Status = "approved"
		s.hooks[key][i].ApprovedAt = &now
		return s.hooks[key][i], nil
	}
	return HookRevision{}, ErrNotFound
}

func (s *MemoryStore) LatestHookRevision(_ context.Context, owner, batchID, bookID string) (HookRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[batchID]
	if !ok || owned.Owner != owner {
		return HookRevision{}, ErrNotFound
	}
	revisions := s.hooks[memoryBookKey(batchID, bookID)]
	if len(revisions) == 0 {
		return HookRevision{}, ErrNotFound
	}
	return revisions[len(revisions)-1], nil
}

func (s *MemoryStore) PersistDirectorRevision(_ context.Context, owner string, book Book, snapshot DirectorSnapshot, digest, hookID string, output DirectorResult) (DirectorRevision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[book.BatchID]
	if !ok || owned.Owner != owner {
		return DirectorRevision{}, ErrNotFound
	}
	bookIndex := -1
	for i := range owned.Value.Books {
		if owned.Value.Books[i].ID == book.ID {
			bookIndex = i
			break
		}
	}
	if bookIndex < 0 {
		return DirectorRevision{}, ErrNotFound
	}
	b := owned.Value
	orphaned := []OrphanedOverride{}
	previousSelections := map[int]VideoAssetSelection{}
	previousSelectionExists := map[int]bool{}
	for vi := range b.Books[bookIndex].Videos {
		old := &b.Books[bookIndex].Videos[vi]
		old.CompatibilityState = "orphaned"
		patch := clonePatch(s.patches[scopeKey(ScopeRef{Kind: ScopeVideo, BatchID: b.ID, BookID: book.ID, VideoID: old.ID})])
		if selection, ok := decodeVideoAssetSelection(patch); ok {
			previousSelections[vi], previousSelectionExists[vi] = selection, true
		}
		if len(patch) > 0 {
			orphaned = append(orphaned, OrphanedOverride{VideoID: old.ID, Patch: patch, State: "orphaned"})
		}
	}
	key := memoryBookKey(b.ID, book.ID)
	now := time.Now().UTC()
	revision := DirectorRevision{ID: s.id("director"), BatchID: b.ID, BookID: book.ID, Revision: int64(len(s.directors[key]) + 1), Mode: snapshot.Mode, SnapshotID: s.id("snapshot"), SourceDigest: digest, HookRevisionID: hookID, Output: output, OrphanedOverrides: orphaned, CreatedAt: now}
	newVideos := make([]Video, 0, len(output.Storyboard))
	for i, draft := range output.Storyboard {
		video := Video{ID: s.id("video"), BatchID: b.ID, BookID: book.ID, Label: fmt.Sprintf("VIDEO %02d", i+1), VideoPrompt: storyboardVideoPrompt(draft), VisualPrompt: draft.VisualPrompt, DurationSeconds: float64(draft.DurationSec), CompatibilityState: "active", Revision: 1}
		newVideos = append(newVideos, video)
	}
	revision.Videos = append([]Video(nil), newVideos...)
	b.Books[bookIndex].Videos = newVideos
	b.Books[bookIndex].Revision++
	b.UpdatedAt = now
	s.batches[b.ID] = memoryOwned[Batch]{Owner: owner, Value: b}
	s.directors[key] = append(s.directors[key], revision)
	desiredDirectorAssets := directorBookAssets(book, snapshot, output)
	// Empty structured output (SD plain-text mode): never wipe extracted assets.
	if len(desiredDirectorAssets) > 0 {
		existingRows := []existingAssetRow{}
		hasImage := map[string]bool{}
		for _, ownedImage := range s.bookAssetImages {
			if ownedImage.Owner == owner {
				hasImage[ownedImage.Value.AssetID] = true
			}
		}
		for assetID, ownedAsset := range s.bookAssets {
			asset := ownedAsset.Value
			if ownedAsset.Owner != owner || asset.BatchID != book.BatchID || asset.BookID != book.ID {
				continue
			}
			existingRows = append(existingRows, existingAssetRow{ID: assetID, Kind: asset.Kind, Name: asset.Name, Source: asset.Source})
		}
		assetPlan := reconcileDirectorAssets(existingRows, desiredDirectorAssets, hasImage)
		for _, staleID := range assetPlan.Deletes {
			delete(s.bookAssets, staleID)
		}
		// Matched rows are frozen: no prompt overwrite, no revision bump.
		for _, seed := range assetPlan.Inserts {
			seed.ID, seed.Revision, seed.CreatedAt, seed.UpdatedAt = s.id("asset"), 1, now, now
			s.bookAssets[seed.ID] = memoryOwned[BookAsset]{Owner: owner, Value: seed}
		}
	}
	assetIDs := map[string]string{}
	nameRowsByKind := map[string][][2]string{}
	for _, ownedAsset := range s.bookAssets {
		asset := ownedAsset.Value
		if ownedAsset.Owner == owner && asset.BatchID == book.BatchID && asset.BookID == book.ID {
			assetIDs[asset.Kind+"\x00"+asset.Name] = asset.ID
			nameRowsByKind[asset.Kind] = append(nameRowsByKind[asset.Kind], [2]string{asset.Name, asset.ID})
		}
	}
	for kind, rows := range nameRowsByKind {
		addCoreNameLookups(assetIDs, kind, rows)
	}
	for ordinal, video := range newVideos {
		selection := reconcileVideoAssetSelection(previousSelections[ordinal], previousSelectionExists[ordinal], automaticAssetIDsForDirectorVideo(output.Storyboard[ordinal], assetIDs))
		patchKey := scopeKey(ScopeRef{Kind: ScopeVideo, BatchID: b.ID, BookID: book.ID, VideoID: video.ID})
		if selection.effectivelyEmpty() {
			// 空选择不落库：否则会被当成“用户一个都没选”，把默认资产全过滤掉。
			delete(s.patches, patchKey)
			continue
		}
		encoded, err := json.Marshal(selection)
		if err != nil {
			return DirectorRevision{}, err
		}
		s.patches[patchKey] = SettingsPatch{videoAssetSelectionKey: encoded}
	}
	return revision, nil
}

func (s *MemoryStore) PersistExtractedBookAssets(_ context.Context, owner string, book Book, snapshot DirectorSnapshot, assets DirectorAssets) ([]BookAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, ok := s.batches[book.BatchID]
	if !ok || owned.Owner != owner {
		return nil, ErrNotFound
	}
	if _, err := bookFromBatch(owned.Value, book.ID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for assetID, ownedAsset := range s.bookAssets {
		asset := ownedAsset.Value
		if ownedAsset.Owner == owner && asset.BatchID == book.BatchID && asset.BookID == book.ID && asset.Source == "director" {
			delete(s.bookAssets, assetID)
		}
	}
	output := DirectorResult{Characters: assets.Characters, Scenes: assets.Scenes, Props: assets.Props}
	for _, seed := range directorBookAssets(book, snapshot, output) {
		found := false
		for assetID, ownedAsset := range s.bookAssets {
			asset := ownedAsset.Value
			if ownedAsset.Owner != owner || asset.BatchID != book.BatchID || asset.BookID != book.ID || asset.Kind != seed.Kind || asset.Name != seed.Name {
				continue
			}
			found = true
			if asset.Source != "manual" {
				asset.Prompt, asset.Source, asset.ExtractionPresetID, asset.ExtractionPresetVersion = seed.Prompt, seed.Source, seed.ExtractionPresetID, seed.ExtractionPresetVersion
				asset.Revision++
				asset.UpdatedAt = now
				s.bookAssets[assetID] = memoryOwned[BookAsset]{Owner: owner, Value: asset}
			}
			break
		}
		if !found {
			seed.ID, seed.Revision, seed.CreatedAt, seed.UpdatedAt = s.id("asset"), 1, now, now
			s.bookAssets[seed.ID] = memoryOwned[BookAsset]{Owner: owner, Value: seed}
		}
	}
	return memoryBookAssets(s.bookAssets, book.BatchID, book.ID), nil
}
