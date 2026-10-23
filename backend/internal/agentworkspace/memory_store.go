package agentworkspace

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu                      sync.Mutex
	projects                map[string]Project
	imports                 map[string]LegacyImport
	videoJobs               map[string]VideoJob
	videoIdempotency        map[string]string
	generationTasks         map[string]GenerationTask
	generationIdempotency   map[string]string
	scriptShots             map[string]ScriptShot
	shotGenerationLinks     map[string]ShotGenerationLink
	batchFactoryBindings    map[string]BatchFactoryBinding
	batchFactoryProjects    map[string]string
	batchFactoryNodeOrigins map[string]BatchFactoryNodeOrigin
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects: make(map[string]Project), imports: make(map[string]LegacyImport), videoJobs: make(map[string]VideoJob), videoIdempotency: make(map[string]string),
		generationTasks: make(map[string]GenerationTask), generationIdempotency: make(map[string]string),
		scriptShots: make(map[string]ScriptShot), shotGenerationLinks: make(map[string]ShotGenerationLink),
		batchFactoryBindings: make(map[string]BatchFactoryBinding), batchFactoryProjects: make(map[string]string), batchFactoryNodeOrigins: make(map[string]BatchFactoryNodeOrigin),
	}
}

func projectKey(owner, id string) string                   { return owner + "\x00" + id }
func importKey(owner, path string) string                  { return owner + "\x00" + path }
func videoJobKey(owner, id string) string                  { return owner + "\x00" + id }
func videoIdempotencyKey(owner, idempotency string) string { return owner + "\x00" + idempotency }

func cloneProject(project Project) Project {
	raw, _ := json.Marshal(project)
	var out Project
	_ = json.Unmarshal(raw, &out)
	if out.Messages == nil {
		out.Messages = []Message{}
	}
	if out.Canvas.Nodes == nil {
		out.Canvas.Nodes = []Node{}
	}
	if out.Canvas.Edges == nil {
		out.Canvas.Edges = []Edge{}
	}
	if out.Executions == nil {
		out.Executions = []Execution{}
	}
	return out
}

func (s *MemoryStore) ListProjects(_ context.Context, owner string) ([]ProjectSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ProjectSummary, 0)
	for key, project := range s.projects {
		if project.DeletedAt != nil || !strings.HasPrefix(key, owner+"\x00") {
			continue
		}
		out = append(out, ProjectSummary{ID: project.ID, Title: project.Title, TitleMode: project.TitleMode, Preview: project.Preview, MessageCount: len(project.Messages), CreatedAt: project.CreatedAt, UpdatedAt: project.UpdatedAt})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (s *MemoryStore) Project(_ context.Context, owner, id string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, ok := s.projects[projectKey(owner, id)]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	return cloneProject(project), nil
}

func normalizeProject(project Project) Project {
	project.ID = strings.TrimSpace(project.ID)
	project.Title = strings.TrimSpace(project.Title)
	if project.Title == "" {
		project.Title = "新项目"
	}
	if project.TitleMode == "" {
		project.TitleMode = "automatic"
	}
	if project.Messages == nil {
		project.Messages = []Message{}
	}
	if project.Canvas.Version == 0 {
		project.Canvas = EmptyCanvas()
	} else {
		project.Canvas.Appearance = normalizeCanvasAppearance(project.Canvas.Appearance, project.Canvas.appearanceConfigured)
		project.Canvas.appearanceConfigured = true
	}
	if project.Executions == nil {
		project.Executions = []Execution{}
	}
	return project
}

func (s *MemoryStore) CreateProject(_ context.Context, owner string, project Project) error {
	if err := ValidateCanvasForStorage(project.Canvas); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, project.ID)
	if _, exists := s.projects[key]; exists {
		return ErrProjectExists
	}
	s.projects[key] = cloneProject(normalizeProject(project))
	return nil
}

func (s *MemoryStore) RenameProject(_ context.Context, owner, id, title string, now time.Time) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	project.Title = strings.TrimSpace(title)
	project.TitleMode = "manual"
	project.UpdatedAt = now
	s.projects[key] = project
	return cloneProject(project), nil
}

func (s *MemoryStore) ClearMessages(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	project.Messages = []Message{}
	project.Preview = ""
	project.UpdatedAt = now
	s.projects[key] = project
	return nil
}

func (s *MemoryStore) DeleteProject(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return ErrProjectNotFound
	}
	project.DeletedAt = &now
	project.UpdatedAt = now
	s.projects[key] = project
	return nil
}

func (s *MemoryStore) AppendExchange(_ context.Context, owner, id string, user, assistant Message) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Project{}, ErrProjectNotFound
	}
	project.Messages = append(project.Messages, user, assistant)
	project.Preview = assistant.Content
	if len(project.Preview) > 512 {
		project.Preview = project.Preview[:512]
	}
	project.UpdatedAt = assistant.CreatedAt
	s.projects[key] = project
	return cloneProject(project), nil
}

func (s *MemoryStore) Canvas(_ context.Context, owner, id string) (Canvas, error) {
	project, err := s.Project(context.Background(), owner, id)
	if err != nil {
		return Canvas{}, err
	}
	return project.Canvas, nil
}

func (s *MemoryStore) SaveCanvas(_ context.Context, owner, id string, canvas Canvas, expected int64, now time.Time) (Canvas, error) {
	if err := ValidateCanvasForStorage(canvas); err != nil {
		return Canvas{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(owner, id)
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Canvas{}, ErrProjectNotFound
	}
	if project.Canvas.Revision != expected {
		return Canvas{}, ErrRevisionConflict
	}
	canvas.Version = 1
	canvas.Revision = expected + 1
	canvas.Appearance = normalizeCanvasAppearance(canvas.Appearance, canvas.appearanceConfigured)
	canvas.appearanceConfigured = true
	if canvas.Nodes == nil {
		canvas.Nodes = []Node{}
	}
	if canvas.Edges == nil {
		canvas.Edges = []Edge{}
	}
	project.Canvas = canvas
	project.UpdatedAt = now
	s.projects[key] = project
	return cloneProject(project).Canvas, nil
}

func (s *MemoryStore) Executions(_ context.Context, owner, id string) ([]Execution, error) {
	project, err := s.Project(context.Background(), owner, id)
	if err != nil {
		return nil, err
	}
	return project.Executions, nil
}

func (s *MemoryStore) BeginNodeExecution(_ context.Context, owner, projectID string, execution Execution) (Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(strings.TrimSpace(owner), strings.TrimSpace(projectID))
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Execution{}, ErrProjectNotFound
	}
	execution.ID = strings.TrimSpace(execution.ID)
	execution.ToolName = strings.TrimSpace(execution.ToolName)
	execution.IdempotencyKey = strings.TrimSpace(execution.IdempotencyKey)
	if execution.ID == "" || execution.ToolName == "" || execution.IdempotencyKey == "" {
		return Execution{}, ErrExecutionState
	}
	for existingKey, existingProject := range s.projects {
		if !strings.HasPrefix(existingKey, strings.TrimSpace(owner)+"\x00") {
			continue
		}
		for _, existing := range existingProject.Executions {
			if existing.ID == execution.ID || existing.IdempotencyKey == execution.IdempotencyKey {
				return Execution{}, ErrExecutionExists
			}
		}
	}
	execution.Status = "running"
	project.Executions = append(project.Executions, execution)
	s.projects[key] = project
	return execution, nil
}

func (s *MemoryStore) SetNodeExecutionAwaitingConfirmation(_ context.Context, owner, projectID, executionID string, detail map[string]any, now time.Time) (Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(strings.TrimSpace(owner), strings.TrimSpace(projectID))
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Execution{}, ErrProjectNotFound
	}
	for index := range project.Executions {
		execution := &project.Executions[index]
		if execution.ID != strings.TrimSpace(executionID) {
			continue
		}
		if execution.Status != "running" {
			return Execution{}, ErrExecutionState
		}
		execution.Status = "awaiting_confirmation"
		execution.Detail = cloneMap(detail)
		execution.UpdatedAt = now.UTC()
		project.UpdatedAt = now.UTC()
		s.projects[key] = project
		return cloneProject(project).Executions[index], nil
	}
	return Execution{}, ErrExecutionNotFound
}

func (s *MemoryStore) ApplyNodeExecution(_ context.Context, owner, projectID string, input ApplyNodeExecutionInput) (Canvas, Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(strings.TrimSpace(owner), strings.TrimSpace(projectID))
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Canvas{}, Execution{}, ErrProjectNotFound
	}
	executionIndex := -1
	for index := range project.Executions {
		if project.Executions[index].ID == strings.TrimSpace(input.ExecutionID) {
			executionIndex = index
			break
		}
	}
	if executionIndex < 0 {
		return Canvas{}, Execution{}, ErrExecutionNotFound
	}
	executionStatus := project.Executions[executionIndex].Status
	if executionStatus != "running" && !(input.NodeType == "image" && executionStatus == "awaiting_confirmation") {
		return Canvas{}, Execution{}, ErrExecutionState
	}
	nodeIndex := -1
	for index := range project.Canvas.Nodes {
		node := project.Canvas.Nodes[index]
		if node.ID == input.NodeID {
			nodeIndex = index
			if node.Type != input.NodeType || node.Version != input.NodeVersion {
				return Canvas{}, Execution{}, ErrExecutionNodeChanged
			}
			break
		}
	}
	if nodeIndex < 0 {
		return Canvas{}, Execution{}, ErrExecutionNodeChanged
	}
	node := &project.Canvas.Nodes[nodeIndex]
	switch input.NodeType {
	case "text":
		if strings.TrimSpace(input.Summary) == "" {
			return Canvas{}, Execution{}, ErrExecutionState
		}
		node.Summary = strings.TrimSpace(input.Summary)
	case "image":
		if strings.TrimSpace(stringValue(input.AssetRef["id"])) == "" || stringValue(input.AssetRef["kind"]) != "image" {
			return Canvas{}, Execution{}, ErrExecutionState
		}
		node.AssetRef = cloneMap(input.AssetRef)
	default:
		return Canvas{}, Execution{}, ErrExecutionState
	}
	node.Version++
	node.Status = "confirmed"
	project.Canvas.Revision++
	project.UpdatedAt = input.Now.UTC()
	execution := &project.Executions[executionIndex]
	execution.Status = "completed"
	execution.Detail = cloneMap(input.Detail)
	execution.UpdatedAt = input.Now.UTC()
	s.projects[key] = project
	return cloneProject(project).Canvas, cloneProject(project).Executions[executionIndex], nil
}

func (s *MemoryStore) FailNodeExecution(_ context.Context, owner, projectID, executionID string, detail map[string]any, now time.Time) (Execution, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := projectKey(strings.TrimSpace(owner), strings.TrimSpace(projectID))
	project, ok := s.projects[key]
	if !ok || project.DeletedAt != nil {
		return Execution{}, ErrProjectNotFound
	}
	for index := range project.Executions {
		execution := &project.Executions[index]
		if execution.ID != strings.TrimSpace(executionID) {
			continue
		}
		if execution.Status != "running" {
			return Execution{}, ErrExecutionState
		}
		execution.Status = "failed"
		execution.Detail = cloneMap(detail)
		execution.UpdatedAt = now.UTC()
		project.UpdatedAt = now.UTC()
		s.projects[key] = project
		return cloneProject(project).Executions[index], nil
	}
	return Execution{}, ErrExecutionNotFound
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	raw, _ := json.Marshal(input)
	var output map[string]any
	_ = json.Unmarshal(raw, &output)
	return output
}

func (s *MemoryStore) ImportProject(ctx context.Context, owner string, project Project) error {
	err := s.CreateProject(ctx, owner, project)
	if err == ErrProjectExists {
		return nil
	}
	return err
}

func (s *MemoryStore) LegacyImport(_ context.Context, owner, path string) (LegacyImport, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.imports[importKey(owner, path)]
	return record, ok, nil
}

func (s *MemoryStore) RecordLegacyImport(_ context.Context, record LegacyImport) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.imports[importKey(record.OwnerUsername, record.SourcePath)] = record
	return nil
}
