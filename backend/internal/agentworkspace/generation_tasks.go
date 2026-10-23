package agentworkspace

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
)

var (
	ErrGenerationNotFound    = errors.New("agent generation task not found")
	ErrGenerationInvalid     = errors.New("invalid agent generation task")
	ErrGenerationState       = errors.New("invalid agent generation task state")
	ErrGenerationNodeChanged = errors.New("agent generation node changed")
)

type GenerationState string

const (
	GenerationDraft                GenerationState = "draft"
	GenerationQueued               GenerationState = "queued"
	GenerationRunning              GenerationState = "running"
	GenerationAwaitingConfirmation GenerationState = "awaiting_confirmation"
	GenerationCompleted            GenerationState = "completed"
	GenerationFailed               GenerationState = "failed"
	GenerationCancelled            GenerationState = "cancelled"
)

type GenerationResult struct {
	ID        string         `json:"id"`
	AssetID   string         `json:"assetId"`
	Kind      string         `json:"kind"`
	Asset     map[string]any `json:"asset,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

type GenerationTask struct {
	ID                string             `json:"id"`
	ProjectID         string             `json:"projectId"`
	OwnerUsername     string             `json:"-"`
	NodeID            string             `json:"nodeId"`
	NodeVersion       int64              `json:"nodeVersion"`
	Kind              string             `json:"kind"`
	Provider          string             `json:"provider"`
	ModelID           string             `json:"modelId"`
	Mode              string             `json:"mode,omitempty"`
	State             GenerationState    `json:"state"`
	Settings          map[string]any     `json:"settings"`
	ReferenceAssetIDs []string           `json:"referenceAssetIds"`
	Results           []GenerationResult `json:"results"`
	SelectedResultID  string             `json:"selectedResultId,omitempty"`
	IdempotencyKey    string             `json:"idempotencyKey"`
	RetryOfTaskID     string             `json:"retryOfTaskId,omitempty"`
	ErrorMessage      string             `json:"errorMessage,omitempty"`
	CreatedAt         time.Time          `json:"createdAt"`
	UpdatedAt         time.Time          `json:"updatedAt"`
}

type CreateGenerationTaskInput struct {
	ID                string
	ProjectID         string
	NodeID            string
	NodeVersion       int64
	Kind              string
	Provider          string
	ModelID           string
	Mode              string
	Settings          map[string]any
	ReferenceAssetIDs []string
	IdempotencyKey    string
	RetryOfTaskID     string
	Now               time.Time
}

type GenerationTaskTransition struct {
	State             GenerationState
	Results           []GenerationResult
	SelectedResultID  string
	ErrorMessage      string
	ExpectedUpdatedAt time.Time
	Now               time.Time
}

type RetryGenerationTaskInput struct {
	ID             string
	IdempotencyKey string
	Now            time.Time
}

type GenerationTaskStore interface {
	CreateGenerationTask(context.Context, string, CreateGenerationTaskInput) (GenerationTask, error)
	GenerationTask(context.Context, string, string, string) (GenerationTask, error)
	ListGenerationTasks(context.Context, string, string) ([]GenerationTask, error)
	ListRunnableGenerationTasks(context.Context, string, int) ([]GenerationTask, error)
	ListRunnableGenerationTasksByModes(context.Context, []string, int) ([]GenerationTask, error)
	TransitionGenerationTask(context.Context, string, string, string, GenerationTaskTransition) (GenerationTask, error)
	ConfirmGenerationTask(context.Context, string, string, string, string, time.Time) (Canvas, GenerationTask, error)
	ActivateGenerationResult(context.Context, string, string, string, string, int64, time.Time) (Canvas, GenerationTask, error)
	CancelGenerationTask(context.Context, string, string, string, time.Time) (GenerationTask, error)
	RetryGenerationTask(context.Context, string, string, string, RetryGenerationTaskInput) (GenerationTask, error)
}

func validGenerationState(state GenerationState) bool {
	switch state {
	case GenerationDraft, GenerationQueued, GenerationRunning, GenerationAwaitingConfirmation, GenerationCompleted, GenerationFailed, GenerationCancelled:
		return true
	default:
		return false
	}
}

func validGenerationTransition(from, to GenerationState) bool {
	switch from {
	case GenerationDraft:
		return to == GenerationQueued || to == GenerationCancelled
	case GenerationQueued:
		return to == GenerationRunning || to == GenerationFailed || to == GenerationCancelled
	case GenerationRunning:
		return to == GenerationAwaitingConfirmation || to == GenerationCompleted || to == GenerationFailed || to == GenerationCancelled
	case GenerationAwaitingConfirmation:
		return to == GenerationCompleted || to == GenerationFailed || to == GenerationCancelled
	default:
		return false
	}
}

func normalizeGenerationTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func nextGenerationUpdatedAt(current, requested time.Time) time.Time {
	current = normalizeGenerationTime(current)
	requested = normalizeGenerationTime(requested)
	if !requested.After(current) {
		return current.Add(time.Microsecond)
	}
	return requested
}

func normalizeCreateGenerationTask(input CreateGenerationTaskInput) (CreateGenerationTaskInput, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.NodeID = strings.TrimSpace(input.NodeID)
	input.Kind = strings.TrimSpace(input.Kind)
	input.Provider = strings.TrimSpace(input.Provider)
	input.ModelID = strings.TrimSpace(input.ModelID)
	input.Mode = strings.TrimSpace(input.Mode)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.RetryOfTaskID = strings.TrimSpace(input.RetryOfTaskID)
	if input.ID == "" || input.ProjectID == "" || input.NodeID == "" || input.NodeVersion < 1 || input.Kind == "" || input.Provider == "" || input.ModelID == "" || input.IdempotencyKey == "" || input.Now.IsZero() {
		return CreateGenerationTaskInput{}, ErrGenerationInvalid
	}
	if input.Settings == nil {
		input.Settings = map[string]any{}
	}
	input.ReferenceAssetIDs = uniqueNonEmpty(input.ReferenceAssetIDs)
	input.Now = normalizeGenerationTime(input.Now)
	return input, nil
}

func normalizeGenerationResults(results []GenerationResult, kind string, now time.Time) ([]GenerationResult, error) {
	seen := make(map[string]struct{}, len(results))
	out := make([]GenerationResult, 0, len(results))
	for _, result := range results {
		result.ID = strings.TrimSpace(result.ID)
		result.AssetID = strings.TrimSpace(result.AssetID)
		result.Kind = strings.TrimSpace(result.Kind)
		if result.Kind == "" {
			result.Kind = kind
		}
		if result.ID == "" || result.AssetID == "" || result.Kind != kind {
			return nil, ErrGenerationInvalid
		}
		if _, exists := seen[result.ID]; exists {
			return nil, ErrGenerationInvalid
		}
		seen[result.ID] = struct{}{}
		if result.CreatedAt.IsZero() {
			result.CreatedAt = now
		}
		result.CreatedAt = normalizeGenerationTime(result.CreatedAt)
		out = append(out, result)
	}
	return out, nil
}

func cloneGenerationTask(task GenerationTask) GenerationTask {
	raw, _ := json.Marshal(task)
	var out GenerationTask
	_ = json.Unmarshal(raw, &out)
	out.OwnerUsername = task.OwnerUsername
	if out.Settings == nil {
		out.Settings = map[string]any{}
	}
	if out.ReferenceAssetIDs == nil {
		out.ReferenceAssetIDs = []string{}
	}
	if out.Results == nil {
		out.Results = []GenerationResult{}
	}
	return out
}

func generationIDKey(owner, projectID, id string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID) + "\x00" + strings.TrimSpace(id)
}

func generationIdempotencyScope(owner, projectID, key string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID) + "\x00" + strings.TrimSpace(key)
}

func (s *MemoryStore) CreateGenerationTask(_ context.Context, owner string, input CreateGenerationTaskInput) (GenerationTask, error) {
	input, err := normalizeCreateGenerationTask(input)
	if err != nil || strings.TrimSpace(owner) == "" {
		return GenerationTask{}, ErrGenerationInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if project, ok := s.projects[projectKey(owner, input.ProjectID)]; !ok || project.DeletedAt != nil {
		return GenerationTask{}, ErrProjectNotFound
	}
	idempotencyScope := generationIdempotencyScope(owner, input.ProjectID, input.IdempotencyKey)
	if existingID := s.generationIdempotency[idempotencyScope]; existingID != "" {
		return cloneGenerationTask(s.generationTasks[generationIDKey(owner, input.ProjectID, existingID)]), nil
	}
	key := generationIDKey(owner, input.ProjectID, input.ID)
	if _, exists := s.generationTasks[key]; exists {
		return GenerationTask{}, ErrGenerationInvalid
	}
	task := GenerationTask{
		ID: input.ID, ProjectID: input.ProjectID, OwnerUsername: strings.TrimSpace(owner), NodeID: input.NodeID, NodeVersion: input.NodeVersion,
		Kind: input.Kind, Provider: input.Provider, ModelID: input.ModelID, Mode: input.Mode, State: GenerationQueued,
		Settings: input.Settings, ReferenceAssetIDs: input.ReferenceAssetIDs, Results: []GenerationResult{},
		IdempotencyKey: input.IdempotencyKey, RetryOfTaskID: input.RetryOfTaskID, CreatedAt: input.Now, UpdatedAt: input.Now,
	}
	s.generationTasks[key] = cloneGenerationTask(task)
	s.generationIdempotency[idempotencyScope] = task.ID
	return cloneGenerationTask(task), nil
}

func (s *MemoryStore) GenerationTask(_ context.Context, owner, projectID, id string) (GenerationTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.generationTasks[generationIDKey(owner, projectID, id)]
	if !ok {
		return GenerationTask{}, ErrGenerationNotFound
	}
	return cloneGenerationTask(task), nil
}

func (s *MemoryStore) ListGenerationTasks(_ context.Context, owner, projectID string) ([]GenerationTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if project, ok := s.projects[projectKey(owner, projectID)]; !ok || project.DeletedAt != nil {
		return nil, ErrProjectNotFound
	}
	prefix := strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID) + "\x00"
	out := make([]GenerationTask, 0)
	for key, task := range s.generationTasks {
		if strings.HasPrefix(key, prefix) {
			out = append(out, cloneGenerationTask(task))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

// ListRunnableGenerationTasks returns globally queued work for one mode. A
// running task is included as well so a process restart can resume it.
func (s *MemoryStore) ListRunnableGenerationTasks(_ context.Context, mode string, limit int) ([]GenerationTask, error) {
	return s.ListRunnableGenerationTasksByModes(context.Background(), []string{mode}, limit)
}

// ListRunnableGenerationTasksByModes returns queued and restart-resumable work
// for the requested modes, oldest first. It deliberately excludes terminal work.
func (s *MemoryStore) ListRunnableGenerationTasksByModes(_ context.Context, modes []string, limit int) ([]GenerationTask, error) {
	allowed := make(map[string]struct{}, len(modes))
	for _, mode := range modes {
		mode = strings.TrimSpace(mode)
		if mode != "" {
			allowed[mode] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return nil, ErrGenerationInvalid
	}
	if limit <= 0 {
		limit = 100
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]GenerationTask, 0)
	for _, task := range s.generationTasks {
		if _, ok := allowed[task.Mode]; !ok || (task.State != GenerationQueued && task.State != GenerationRunning) {
			continue
		}
		out = append(out, cloneGenerationTask(task))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID < out[j].ID
		}
		return out[i].UpdatedAt.Before(out[j].UpdatedAt)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemoryStore) TransitionGenerationTask(_ context.Context, owner, projectID, id string, transition GenerationTaskTransition) (GenerationTask, error) {
	if !validGenerationState(transition.State) || transition.Now.IsZero() {
		return GenerationTask{}, ErrGenerationInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := generationIDKey(owner, projectID, id)
	task, ok := s.generationTasks[key]
	if !ok {
		return GenerationTask{}, ErrGenerationNotFound
	}
	if !transition.ExpectedUpdatedAt.IsZero() && !normalizeGenerationTime(transition.ExpectedUpdatedAt).Equal(task.UpdatedAt) {
		return GenerationTask{}, ErrGenerationState
	}
	if !validGenerationTransition(task.State, transition.State) {
		return GenerationTask{}, ErrGenerationState
	}
	if transition.State == GenerationAwaitingConfirmation {
		results, err := normalizeGenerationResults(transition.Results, task.Kind, transition.Now)
		if err != nil || len(results) == 0 {
			return GenerationTask{}, ErrGenerationInvalid
		}
		selected := strings.TrimSpace(transition.SelectedResultID)
		if selected == "" {
			selected = results[0].ID
		}
		if !generationResultExists(results, selected) {
			return GenerationTask{}, ErrGenerationInvalid
		}
		task.Results = results
		task.SelectedResultID = selected
	}
	if transition.State == GenerationCompleted && strings.TrimSpace(transition.SelectedResultID) != "" {
		if !generationResultExists(task.Results, strings.TrimSpace(transition.SelectedResultID)) {
			return GenerationTask{}, ErrGenerationInvalid
		}
		task.SelectedResultID = strings.TrimSpace(transition.SelectedResultID)
	}
	task.State = transition.State
	task.ErrorMessage = strings.TrimSpace(transition.ErrorMessage)
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, transition.Now)
	s.generationTasks[key] = cloneGenerationTask(task)
	return cloneGenerationTask(task), nil
}

func generationResultExists(results []GenerationResult, id string) bool {
	for _, result := range results {
		if result.ID == id {
			return true
		}
	}
	return false
}

func selectedGenerationResult(task GenerationTask, resultID string) (GenerationResult, bool) {
	resultID = strings.TrimSpace(resultID)
	if resultID == "" {
		resultID = task.SelectedResultID
	}
	for _, result := range task.Results {
		if result.ID == resultID {
			return result, true
		}
	}
	return GenerationResult{}, false
}

func (s *MemoryStore) ConfirmGenerationTask(_ context.Context, owner, projectID, id, resultID string, now time.Time) (Canvas, GenerationTask, error) {
	if now.IsZero() {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := generationIDKey(owner, projectID, id)
	task, ok := s.generationTasks[key]
	if !ok {
		return Canvas{}, GenerationTask{}, ErrGenerationNotFound
	}
	if task.State != GenerationAwaitingConfirmation {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	result, ok := selectedGenerationResult(task, resultID)
	if !ok || result.AssetID == "" {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	projectKeyValue := projectKey(owner, projectID)
	project, ok := s.projects[projectKeyValue]
	if !ok || project.DeletedAt != nil {
		return Canvas{}, GenerationTask{}, ErrProjectNotFound
	}
	nodeIndex := -1
	for index := range project.Canvas.Nodes {
		if project.Canvas.Nodes[index].ID == task.NodeID {
			nodeIndex = index
			break
		}
	}
	if nodeIndex < 0 || project.Canvas.Nodes[nodeIndex].Version != task.NodeVersion || project.Canvas.Nodes[nodeIndex].Type != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	node := &project.Canvas.Nodes[nodeIndex]
	node.AssetRef = cloneMap(result.Asset)
	if node.AssetRef == nil {
		node.AssetRef = map[string]any{"id": result.AssetID, "kind": result.Kind}
	}
	node.Status = "confirmed"
	node.Version++
	project.Canvas.Revision++
	project.UpdatedAt = normalizeGenerationTime(now)
	task.State = GenerationCompleted
	task.SelectedResultID = result.ID
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, now)
	s.projects[projectKeyValue] = project
	s.generationTasks[key] = cloneGenerationTask(task)
	return cloneProject(project).Canvas, cloneGenerationTask(task), nil
}

func (s *MemoryStore) ActivateGenerationResult(_ context.Context, owner, projectID, id, resultID string, expectedNodeVersion int64, now time.Time) (Canvas, GenerationTask, error) {
	if expectedNodeVersion < 1 || now.IsZero() {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := generationIDKey(owner, projectID, id)
	task, ok := s.generationTasks[key]
	if !ok {
		return Canvas{}, GenerationTask{}, ErrGenerationNotFound
	}
	if task.State != GenerationCompleted {
		return Canvas{}, GenerationTask{}, ErrGenerationState
	}
	result, ok := selectedGenerationResult(task, resultID)
	if !ok || result.AssetID == "" || result.Kind != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationInvalid
	}
	projectKeyValue := projectKey(owner, projectID)
	project, ok := s.projects[projectKeyValue]
	if !ok || project.DeletedAt != nil {
		return Canvas{}, GenerationTask{}, ErrProjectNotFound
	}
	nodeIndex := -1
	for index := range project.Canvas.Nodes {
		if project.Canvas.Nodes[index].ID == task.NodeID {
			nodeIndex = index
			break
		}
	}
	if nodeIndex < 0 || project.Canvas.Nodes[nodeIndex].Version != expectedNodeVersion || project.Canvas.Nodes[nodeIndex].Type != task.Kind {
		return Canvas{}, GenerationTask{}, ErrGenerationNodeChanged
	}
	node := &project.Canvas.Nodes[nodeIndex]
	node.AssetRef = cloneMap(result.Asset)
	if node.AssetRef == nil {
		node.AssetRef = map[string]any{"id": result.AssetID, "kind": result.Kind}
	}
	node.Status = "confirmed"
	node.Version++
	project.Canvas.Revision++
	project.UpdatedAt = normalizeGenerationTime(now)
	task.SelectedResultID = result.ID
	task.UpdatedAt = nextGenerationUpdatedAt(task.UpdatedAt, now)
	s.projects[projectKeyValue] = project
	s.generationTasks[key] = cloneGenerationTask(task)
	return cloneProject(project).Canvas, cloneGenerationTask(task), nil
}

func (s *MemoryStore) CancelGenerationTask(ctx context.Context, owner, projectID, id string, now time.Time) (GenerationTask, error) {
	return s.TransitionGenerationTask(ctx, owner, projectID, id, GenerationTaskTransition{State: GenerationCancelled, Now: now})
}

func (s *MemoryStore) RetryGenerationTask(_ context.Context, owner, projectID, id string, input RetryGenerationTaskInput) (GenerationTask, error) {
	input.ID = strings.TrimSpace(input.ID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.ID == "" || input.IdempotencyKey == "" || input.Now.IsZero() {
		return GenerationTask{}, ErrGenerationInvalid
	}
	s.mu.Lock()
	original, ok := s.generationTasks[generationIDKey(owner, projectID, id)]
	if !ok {
		s.mu.Unlock()
		return GenerationTask{}, ErrGenerationNotFound
	}
	if original.State != GenerationFailed && original.State != GenerationCancelled {
		s.mu.Unlock()
		return GenerationTask{}, ErrGenerationState
	}
	project := s.projects[projectKey(owner, projectID)]
	for _, node := range project.Canvas.Nodes {
		if node.ID == original.NodeID && node.Version != original.NodeVersion {
			s.mu.Unlock()
			return GenerationTask{}, ErrGenerationNodeChanged
		}
	}
	s.mu.Unlock()
	return s.CreateGenerationTask(context.Background(), owner, CreateGenerationTaskInput{
		ID: input.ID, ProjectID: original.ProjectID, NodeID: original.NodeID, NodeVersion: original.NodeVersion,
		Kind: original.Kind, Provider: original.Provider, ModelID: original.ModelID, Mode: original.Mode,
		Settings: original.Settings, ReferenceAssetIDs: original.ReferenceAssetIDs,
		IdempotencyKey: input.IdempotencyKey, RetryOfTaskID: original.ID, Now: input.Now,
	})
}
