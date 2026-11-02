package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type AgentVideoToolServices struct {
	Projects  agentworkspace.Store
	Assets    agentworkspace.AssetStore
	Objects   agentworkspace.AssetObjectStore
	Processor AgentVideoProcessor
	Now       func() time.Time
}

type agentVideoToolRequest struct {
	NodeID         string         `json:"nodeId"`
	NodeVersion    int64          `json:"nodeVersion"`
	Operation      string         `json:"operation"`
	SourceAssetID  string         `json:"sourceAssetId"`
	Settings       map[string]any `json:"settings"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

func RegisterAgentVideoToolRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentVideoToolServices) {
	if root == nil || services.Projects == nil || services.Assets == nil || services.Objects == nil || services.Processor == nil {
		return
	}
	if services.Now == nil {
		services.Now = time.Now
	}
	root.Handle("POST /api/agent/tasks/{id}/video-tools", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input agentVideoToolRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "视频处理参数无效"})
			return
		}
		operation, err := agentworkspace.ParseVideoToolOperation(input.Operation)
		if err != nil || strings.TrimSpace(input.NodeID) == "" || input.NodeVersion < 1 || strings.TrimSpace(input.SourceAssetID) == "" || strings.TrimSpace(input.IdempotencyKey) == "" || agentworkspace.ValidateVideoToolSettings(operation, input.Settings) != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "视频处理参数不完整"})
			return
		}
		owner, projectID := strings.TrimSpace(identity.Username), strings.TrimSpace(req.PathValue("id"))
		project, err := services.Projects.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		if agentworkspace.NormalizeProjectType(project.ProjectType) == "comic" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "漫画作品不支持视频处理"})
			return
		}
		canvas, err := services.Projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		var target agentworkspace.Node
		for _, node := range canvas.Nodes {
			if node.ID == strings.TrimSpace(input.NodeID) {
				target = node
				break
			}
		}
		if target.ID == "" || target.Version != input.NodeVersion || target.Type != operation.ResultKind() || len(target.AssetRef) != 0 {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "请先新建一个空白结果节点，再执行视频处理"})
			return
		}
		source, err := services.Assets.Asset(req.Context(), owner, projectID, strings.TrimSpace(input.SourceAssetID))
		if err != nil || source.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择当前项目中的视频素材"})
			return
		}
		task, err := services.Projects.CreateGenerationTask(req.Context(), owner, agentworkspace.CreateGenerationTaskInput{
			ID: newAgentID("generation"), ProjectID: projectID, NodeID: target.ID, NodeVersion: target.Version,
			Kind: target.Type, Provider: "local_ffmpeg", ModelID: "ffmpeg", Mode: operation.GenerationMode(),
			Settings:          map[string]any{"operation": string(operation), "settings": cloneAgentSettings(input.Settings), "sourceAssetId": source.ID},
			ReferenceAssetIDs: []string{source.ID}, IdempotencyKey: strings.TrimSpace(input.IdempotencyKey), Now: services.Now().UTC(),
		})
		if err != nil {
			writeAgentGenerationTaskError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"task": task})
	})))
}

const agentVideoToolLease = 3 * time.Minute

type agentVideoToolCoordinator struct {
	services AgentVideoToolServices
	mu       sync.Mutex
	inFlight map[string]struct{}
	now      func() time.Time
}

func newAgentVideoToolCoordinator(services AgentVideoToolServices) *agentVideoToolCoordinator {
	now := services.Now
	if now == nil {
		now = time.Now
	}
	return &agentVideoToolCoordinator{services: services, inFlight: map[string]struct{}{}, now: now}
}

func (s *agentVideoToolCoordinator) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		_ = s.pollPendingOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *agentVideoToolCoordinator) pollPendingOnce(ctx context.Context) error {
	tasks, err := s.services.Projects.ListRunnableGenerationTasksByModes(ctx, []string{"video_trim", "video_crop", "video_extract_frames", "video_separate_audio"}, 100)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if (task.State == agentworkspace.GenerationRunning && s.now().UTC().Sub(task.UpdatedAt) < agentVideoToolLease) || !s.claim(task.ID) {
			continue
		}
		go func(task agentworkspace.GenerationTask) {
			defer s.release(task.ID)
			_ = s.execute(context.Background(), task)
		}(task)
	}
	return nil
}

func (s *agentVideoToolCoordinator) claim(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inFlight[id]; ok {
		return false
	}
	s.inFlight[id] = struct{}{}
	return true
}
func (s *agentVideoToolCoordinator) release(id string) {
	s.mu.Lock()
	delete(s.inFlight, id)
	s.mu.Unlock()
}

func (s *agentVideoToolCoordinator) execute(ctx context.Context, task agentworkspace.GenerationTask) error {
	if task.State == agentworkspace.GenerationQueued {
		var err error
		task, err = s.services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: s.now().UTC()})
		if err != nil {
			return err
		}
	}
	operation, err := agentworkspace.ParseVideoToolOperation(strings.TrimPrefix(task.Mode, "video_"))
	if err != nil {
		return s.fail(ctx, task, "视频处理操作无效")
	}
	settings, _ := task.Settings["settings"].(map[string]any)
	source, contents, err := readAgentVideoSource(ctx, task.OwnerUsername, task.ProjectID, fmt.Sprint(task.Settings["sourceAssetId"]), s.services.Assets, s.services.Objects)
	if err != nil {
		return s.fail(ctx, task, "源视频无法读取")
	}
	outputs, err := s.services.Processor.Process(ctx, AgentVideoProcessInput{Operation: operation, Source: source, SourceBytes: contents, Settings: settings})
	if err != nil {
		return s.fail(ctx, task, err.Error())
	}
	results, err := persistAgentVideoToolOutputs(ctx, task.OwnerUsername, task.ProjectID, task.Kind, outputs, s.services)
	if err != nil {
		return s.fail(ctx, task, "视频处理结果保存失败")
	}
	current, err := s.services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil || current.State == agentworkspace.GenerationCancelled {
		return err
	}
	_, err = s.services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: results, SelectedResultID: results[0].ID, ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}

func (s *agentVideoToolCoordinator) fail(ctx context.Context, task agentworkspace.GenerationTask, message string) error {
	current, err := s.services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil || current.State == agentworkspace.GenerationCancelled || current.State == agentworkspace.GenerationFailed {
		return err
	}
	_, err = s.services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: strings.TrimSpace(message), ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}

func persistAgentVideoToolOutputs(ctx context.Context, owner, projectID, expectedKind string, outputs []AgentVideoProcessOutput, services AgentVideoToolServices) ([]agentworkspace.GenerationResult, error) {
	if len(outputs) == 0 {
		return nil, errors.New("video processor returned no outputs")
	}
	results := make([]agentworkspace.GenerationResult, 0, len(outputs))
	for index, output := range outputs {
		if output.Kind != expectedKind || len(output.Bytes) == 0 || strings.TrimSpace(output.Extension) == "" {
			return nil, errors.New("invalid video processor output")
		}
		assetID := newAgentID("asset")
		key, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, output.Extension)
		if err != nil {
			return nil, err
		}
		asset := agentworkspace.Asset{ID: assetID, ProjectID: projectID, Kind: output.Kind, Name: path.Base(strings.TrimSpace(output.Name)), ContentType: output.ContentType, ObjectKey: key, ByteSize: int64(len(output.Bytes)), CreatedAt: services.Now().UTC()}
		if asset.Name == "." || asset.Name == "" {
			asset.Name = fmt.Sprintf("视频处理结果-%d.%s", index+1, output.Extension)
		}
		if err := services.Objects.Put(ctx, asset, bytes.NewReader(output.Bytes)); err != nil {
			return nil, err
		}
		if err := services.Assets.CreateAsset(ctx, owner, asset); err != nil {
			_ = services.Objects.Delete(ctx, asset)
			return nil, err
		}
		asset.URL = agentAssetContentURL(projectID, asset.ID)
		results = append(results, agentworkspace.GenerationResult{ID: asset.ID, AssetID: asset.ID, Kind: asset.Kind, Asset: map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": asset.Kind, "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType, "byteSize": asset.ByteSize}, CreatedAt: asset.CreatedAt})
	}
	return results, nil
}
