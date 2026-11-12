package aicut

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

type memoryProject struct {
	project       Project
	timeline      Timeline
	history       []Timeline
	historyCursor int
}

type MemoryStore struct {
	assetFolders           map[string]AssetFolder
	assetFolderAssignments map[string]string
	mu                     sync.Mutex
	projects               map[string]memoryProject
	assets                 map[string]Asset
	asrJobs                map[string]ASRJob
	exportJobs             map[string]ExportJob
	editPlans              map[string]EditPlan
	editPlanRequests       map[string]string
	editPlanSummaries      map[string]EditPlanApplySummary
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects: make(map[string]memoryProject), assets: make(map[string]Asset), asrJobs: make(map[string]ASRJob), exportJobs: make(map[string]ExportJob),
		editPlans: make(map[string]EditPlan), editPlanRequests: make(map[string]string), editPlanSummaries: make(map[string]EditPlanApplySummary),
	}
}

func aicutExportJobKey(owner, projectID, id string) string {
	return owner + "\x00" + projectID + "\x00" + id
}

func cloneExportJob(job ExportJob) ExportJob {
	raw, _ := json.Marshal(job)
	var cloned ExportJob
	_ = json.Unmarshal(raw, &cloned)
	cloned.Owner = job.Owner
	cloned.OutputPath = job.OutputPath
	cloned.ProjectSnapshot = job.ProjectSnapshot
	cloned.TimelineSnapshot = cloneTimeline(job.TimelineSnapshot)
	return cloned
}

func (s *MemoryStore) CreateExportJob(_ context.Context, owner string, job ExportJob) error {
	job.Owner = strings.TrimSpace(owner)
	if err := validateExportJob(job); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if item, ok := s.projects[aicutProjectKey(owner, job.ProjectID)]; !ok || item.project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	key := aicutExportJobKey(owner, job.ProjectID, job.ID)
	if _, exists := s.exportJobs[key]; exists {
		return ErrInvalidExportJob
	}
	s.exportJobs[key] = cloneExportJob(job)
	return nil
}

func (s *MemoryStore) ExportJob(_ context.Context, owner, projectID, id string) (ExportJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.exportJobs[aicutExportJobKey(owner, projectID, id)]
	if !ok {
		return ExportJob{}, ErrExportJobNotFound
	}
	return cloneExportJob(job), nil
}

func (s *MemoryStore) ListExportJobs(_ context.Context, owner, projectID string) ([]ExportJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]ExportJob, 0)
	prefix := owner + "\x00" + projectID + "\x00"
	for key, job := range s.exportJobs {
		if strings.HasPrefix(key, prefix) {
			items = append(items, cloneExportJob(job))
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, nil
}

func (s *MemoryStore) ClaimNextExportJob(_ context.Context, now time.Time) (ExportJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var key string
	var selected ExportJob
	for candidateKey, job := range s.exportJobs {
		if job.State != ExportQueued {
			continue
		}
		if key == "" || job.CreatedAt.Before(selected.CreatedAt) || (job.CreatedAt.Equal(selected.CreatedAt) && job.ID < selected.ID) {
			key, selected = candidateKey, job
		}
	}
	if key == "" {
		return ExportJob{}, ErrExportQueueEmpty
	}
	selected = selected.WithState(ExportRunning, 5, "", "", now)
	s.exportJobs[key] = selected
	return cloneExportJob(selected), nil
}

func (s *MemoryStore) UpdateExportJob(_ context.Context, job ExportJob) error {
	if err := validateExportJob(job); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutExportJobKey(job.Owner, job.ProjectID, job.ID)
	if _, ok := s.exportJobs[key]; !ok {
		return ErrExportJobNotFound
	}
	s.exportJobs[key] = cloneExportJob(job)
	return nil
}

func (s *MemoryStore) RequestExportCancel(_ context.Context, owner, projectID, id string, now time.Time) (ExportJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutExportJobKey(owner, projectID, id)
	job, ok := s.exportJobs[key]
	if !ok {
		return ExportJob{}, ErrExportJobNotFound
	}
	switch job.State {
	case ExportQueued:
		job = job.WithState(ExportCanceled, 100, "", "", now)
	case ExportRunning:
		job = job.WithState(ExportCancelRequested, job.Progress, "", "", now)
	}
	s.exportJobs[key] = job
	return cloneExportJob(job), nil
}

func aicutASRJobKey(owner, projectID, id string) string {
	return owner + "\x00" + projectID + "\x00" + id
}

func (s *MemoryStore) CreateASRJob(_ context.Context, owner string, job ASRJob) error {
	if err := validateASRJob(job); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if item, ok := s.projects[aicutProjectKey(owner, job.ProjectID)]; !ok || item.project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	job.Owner = owner
	key := aicutASRJobKey(owner, job.ProjectID, job.ID)
	if _, exists := s.asrJobs[key]; exists {
		return ErrInvalidASRJob
	}
	s.asrJobs[key] = job
	return nil
}

func (s *MemoryStore) ASRJob(_ context.Context, owner, projectID, id string) (ASRJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.asrJobs[aicutASRJobKey(owner, projectID, id)]
	if !ok {
		return ASRJob{}, ErrASRJobNotFound
	}
	return job, nil
}

func (s *MemoryStore) ListASRJobs(_ context.Context, owner, projectID string) ([]ASRJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]ASRJob, 0)
	prefix := owner + "\x00" + projectID + "\x00"
	for key, job := range s.asrJobs {
		if strings.HasPrefix(key, prefix) {
			items = append(items, job)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (s *MemoryStore) ClaimNextASRJob(_ context.Context, now time.Time) (ASRJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var key string
	var selected ASRJob
	for candidateKey, job := range s.asrJobs {
		if job.State != ASRQueued {
			continue
		}
		if key == "" || job.CreatedAt.Before(selected.CreatedAt) {
			key, selected = candidateKey, job
		}
	}
	if key == "" {
		return ASRJob{}, ErrASRQueueEmpty
	}
	selected = selected.WithState(ASRRunning, 5, "", now)
	s.asrJobs[key] = selected
	return selected, nil
}

func (s *MemoryStore) UpdateASRJob(_ context.Context, job ASRJob) error {
	if err := validateASRJob(job); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutASRJobKey(job.Owner, job.ProjectID, job.ID)
	if existing, ok := s.asrJobs[key]; ok && canonicalASRTimeline(existing.TimelineID) != canonicalASRTimeline(job.TimelineID) {
		return ErrInvalidASRJob
	}
	if _, ok := s.asrJobs[key]; !ok {
		return ErrASRJobNotFound
	}
	s.asrJobs[key] = job
	return nil
}

func (s *MemoryStore) RequestASRCancel(_ context.Context, owner, projectID, id string, now time.Time) (ASRJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutASRJobKey(owner, projectID, id)
	job, ok := s.asrJobs[key]
	if !ok {
		return ASRJob{}, ErrASRJobNotFound
	}
	switch job.State {
	case ASRQueued:
		job = job.WithState(ASRCanceled, 100, "", now)
	case ASRRunning:
		job = job.WithState(ASRCancelRequested, job.Progress, "", now)
	}
	s.asrJobs[key] = job
	return job, nil
}

func aicutProjectKey(owner, id string) string { return owner + "\x00" + id }
func aicutAssetKey(owner, projectID, id string) string {
	return owner + "\x00" + projectID + "\x00" + id
}

func cloneTimeline(timeline Timeline) Timeline {
	raw, _ := json.Marshal(timeline)
	var cloned Timeline
	_ = json.Unmarshal(raw, &cloned)
	if cloned.Tracks == nil {
		cloned.Tracks = []Track{}
	}
	return cloned
}

func (s *MemoryStore) ListProjects(_ context.Context, owner string) ([]Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Project, 0)
	for key, item := range s.projects {
		if item.project.DeletedAt == nil && strings.HasPrefix(key, owner+"\x00") {
			out = append(out, item.project)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (s *MemoryStore) CreateProject(_ context.Context, owner string, project Project) error {
	project, err := normalizeProject(project)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, project.ID)
	if _, exists := s.projects[key]; exists {
		return ErrProjectExists
	}
	empty := Timeline{Tracks: []Track{}}
	s.projects[key] = memoryProject{project: project, timeline: empty, history: []Timeline{cloneTimeline(empty)}}
	return nil
}

func (s *MemoryStore) Project(_ context.Context, owner, id string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.projects[aicutProjectKey(owner, id)]
	if !ok || item.project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	return item.project, nil
}

func (s *MemoryStore) RenameProject(_ context.Context, owner, id, title string, now time.Time) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	item.project.Title = strings.TrimSpace(title)
	if item.project.Title == "" {
		return Project{}, ErrInvalidProject
	}
	item.project.UpdatedAt = now
	s.projects[key] = item
	return item.project, nil
}

func (s *MemoryStore) UpdateProjectAppearance(_ context.Context, owner, id, mode, color string, now time.Time) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	candidate := item.project
	candidate.BackgroundMode = mode
	candidate.BackgroundColor = color
	normalized, err := normalizeProject(candidate)
	if err != nil {
		return Project{}, err
	}
	normalized.UpdatedAt = now
	item.project = normalized
	s.projects[key] = item
	return item.project, nil
}

func (s *MemoryStore) DeleteProject(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	item.project.DeletedAt = &now
	item.project.UpdatedAt = now
	s.projects[key] = item
	return nil
}

func (s *MemoryStore) Timeline(_ context.Context, owner, id string) (Timeline, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.projects[aicutProjectKey(owner, id)]
	if !ok || item.project.DeletedAt != nil {
		return Timeline{}, ErrProjectNotFound
	}
	return cloneTimeline(item.timeline), nil
}

func (s *MemoryStore) SaveTimeline(_ context.Context, owner, id string, timeline Timeline, expectedRevision int64, now time.Time) (Timeline, error) {
	if err := validateTimeline(timeline); err != nil {
		return Timeline{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return Timeline{}, ErrProjectNotFound
	}
	if item.timeline.Revision != expectedRevision {
		return Timeline{}, ErrRevisionConflict
	}
	timeline.Revision = expectedRevision + 1
	item.timeline = cloneTimeline(timeline)
	if item.history == nil {
		item.history = []Timeline{{Revision: expectedRevision, Tracks: []Track{}}}
		item.historyCursor = 0
	}
	item.history = append(item.history[:item.historyCursor+1], cloneTimeline(timeline))
	item.historyCursor++
	item.project.UpdatedAt = now
	s.projects[key] = item
	return cloneTimeline(item.timeline), nil
}

func timelineHistoryState(item memoryProject) TimelineHistoryState {
	return TimelineHistoryState{
		CanUndo: item.historyCursor > 0,
		CanRedo: item.historyCursor+1 < len(item.history),
	}
}

func (s *MemoryStore) TimelineHistory(_ context.Context, owner, id string) (TimelineHistoryState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.projects[aicutProjectKey(owner, id)]
	if !ok || item.project.DeletedAt != nil {
		return TimelineHistoryState{}, ErrProjectNotFound
	}
	return timelineHistoryState(item), nil
}

func (s *MemoryStore) moveTimelineHistory(owner, id string, expectedRevision int64, delta int, now time.Time) (Timeline, TimelineHistoryState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := aicutProjectKey(owner, id)
	item, ok := s.projects[key]
	if !ok || item.project.DeletedAt != nil {
		return Timeline{}, TimelineHistoryState{}, ErrProjectNotFound
	}
	if item.timeline.Revision != expectedRevision {
		return Timeline{}, TimelineHistoryState{}, ErrRevisionConflict
	}
	target := item.historyCursor + delta
	if target < 0 || target >= len(item.history) {
		return Timeline{}, timelineHistoryState(item), ErrTimelineHistoryBoundary
	}
	item.historyCursor = target
	item.timeline = cloneTimeline(item.history[target])
	item.timeline.Revision = expectedRevision + 1
	item.project.UpdatedAt = now
	s.projects[key] = item
	return cloneTimeline(item.timeline), timelineHistoryState(item), nil
}

func (s *MemoryStore) UndoTimeline(_ context.Context, owner, id string, expectedRevision int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	return s.moveTimelineHistory(owner, id, expectedRevision, -1, now)
}

func (s *MemoryStore) RedoTimeline(_ context.Context, owner, id string, expectedRevision int64, now time.Time) (Timeline, TimelineHistoryState, error) {
	return s.moveTimelineHistory(owner, id, expectedRevision, 1, now)
}

func (s *MemoryStore) ListAssets(_ context.Context, owner, projectID string) ([]Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[aicutProjectKey(owner, projectID)]
	if !ok || project.project.DeletedAt != nil {
		return nil, ErrProjectNotFound
	}
	items := make([]Asset, 0)
	prefix := owner + "\x00" + projectID + "\x00"
	for key, asset := range s.assets {
		if strings.HasPrefix(key, prefix) {
			items = append(items, asset)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func (s *MemoryStore) CreateAsset(_ context.Context, owner string, asset Asset) error {
	if err := validateAsset(asset); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[aicutProjectKey(owner, asset.ProjectID)]
	if !ok || project.project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	key := aicutAssetKey(owner, asset.ProjectID, asset.ID)
	if _, exists := s.assets[key]; exists {
		return ErrAssetExists
	}
	s.assets[key] = asset
	return nil
}

func (s *MemoryStore) Asset(_ context.Context, owner, projectID, id string) (Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.assets[aicutAssetKey(owner, projectID, id)]
	if !ok {
		return Asset{}, ErrAssetNotFound
	}
	return asset, nil
}
