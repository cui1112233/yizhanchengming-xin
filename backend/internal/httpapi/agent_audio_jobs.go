package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/novelfetchworkshop"
)

type AgentAudioServices struct {
	Projects   agentworkspace.Store
	Catalog    novelfetchworkshop.Store
	Assets     agentworkspace.AssetStore
	Objects    agentworkspace.AssetObjectStore
	Generator  agentworkspace.AudioGenerator
	Processor  agentworkspace.AudioProcessor
	Waveform   agentAudioWaveform
	QueueAudio bool
	Now        func() time.Time
}

type agentAudioRequest struct {
	NodeID         string         `json:"nodeId"`
	NodeVersion    int64          `json:"nodeVersion"`
	Operation      string         `json:"operation"`
	Prompt         string         `json:"prompt"`
	Provider       string         `json:"provider"`
	ModelID        string         `json:"modelId"`
	SourceAssetID  string         `json:"sourceAssetId"`
	Settings       map[string]any `json:"settings"`
	Parameters     map[string]any `json:"parameters"`
	Advanced       map[string]any `json:"advanced"`
	IdempotencyKey string         `json:"idempotencyKey"`
}

func RegisterAgentAudioJobRoutes(root *http.ServeMux, middleware BridgeAuth, services AgentAudioServices) {
	if root == nil || services.Projects == nil || services.Assets == nil || services.Objects == nil {
		return
	}
	if services.Now == nil {
		services.Now = time.Now
	}
	root.Handle("POST /api/agent/tasks/{id}/audio-generations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		owner := strings.TrimSpace(identity.Username)
		projectID := strings.TrimSpace(req.PathValue("id"))
		project, err := services.Projects.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		var request agentAudioRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "音频任务参数无效"})
			return
		}
		node, found := agentAudioNode(project.Canvas, request.NodeID)
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "音频节点不存在"})
			return
		}
		if node.Version != request.NodeVersion {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "音频节点已被修改，请重新执行"})
			return
		}
		settings := request.Settings
		if settings == nil {
			settings = request.Parameters
		}
		input, err := agentworkspace.NormalizeAudioJobRequest(agentworkspace.AudioJobRequest{
			Operation: agentworkspace.AudioOperation(request.Operation), Prompt: request.Prompt,
			Provider: request.Provider, ModelID: request.ModelID, SourceAssetID: request.SourceAssetID,
			Settings: settings, Advanced: request.Advanced,
		})
		if err != nil || strings.TrimSpace(request.IdempotencyKey) == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "音频操作或参数无效"})
			return
		}
		provider, modelID := input.Provider, input.ModelID
		if input.Operation.IsProcessing() {
			provider, modelID = "local_ffmpeg", "ffmpeg"
		} else if provider == "" {
			provider = "personal_api"
		}
		capability, err := resolveAgentCanvasInputCapability(req.Context(), services.Catalog, owner, "audio", provider, modelID)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		resolvedInputs, err := agentResolvedCanvasInputSnapshotWithCapability(project.Canvas, node.ID, input.Prompt, compactAudioReferences(input.SourceAssetID), capability)
		if err != nil {
			writeAgentCanvasInputError(w, err)
			return
		}
		now := services.Now().UTC()
		requestedTaskID := newAgentID("generation")
		task, err := services.Projects.CreateGenerationTask(req.Context(), owner, agentworkspace.CreateGenerationTaskInput{
			ID: requestedTaskID, ProjectID: projectID, NodeID: node.ID, NodeVersion: node.Version,
			Kind: "audio", Provider: provider, ModelID: modelID, Mode: string(input.Operation),
			Settings:          map[string]any{"parameters": input.Settings, "advanced": input.Advanced, "sourceAssetId": input.SourceAssetID, "prompt": input.Prompt, "resolvedInputs": resolvedInputs},
			ReferenceAssetIDs: compactAudioReferences(input.SourceAssetID), IdempotencyKey: request.IdempotencyKey, Now: now,
		})
		if err != nil {
			writeAgentAudioError(w, err)
			return
		}
		preparedBatchTask := task.State == agentworkspace.GenerationQueued && stringValue(task.Settings["submissionIdempotencyKey"]) == strings.TrimSpace(request.IdempotencyKey)
		if task.ID != requestedTaskID && !preparedBatchTask {
			writeJSON(w, http.StatusOK, map[string]any{"task": task})
			return
		}
		if services.QueueAudio {
			writeJSON(w, http.StatusAccepted, map[string]any{"task": task})
			return
		}
		running, err := services.Projects.TransitionGenerationTask(req.Context(), owner, projectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: services.Now().UTC()})
		if err != nil {
			writeAgentAudioError(w, err)
			return
		}
		outputs, err := runAgentAudio(req.Context(), owner, projectID, input, services)
		if err != nil {
			failed, _ := services.Projects.TransitionGenerationTask(req.Context(), owner, projectID, running.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: err.Error(), Now: services.Now().UTC()})
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error(), "task": failed})
			return
		}
		assets, results, err := persistAgentAudioOutputs(req.Context(), owner, projectID, outputs, services)
		if err != nil {
			failed, _ := services.Projects.TransitionGenerationTask(req.Context(), owner, projectID, running.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: err.Error(), Now: services.Now().UTC()})
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "音频素材保存失败", "task": failed})
			return
		}
		awaiting, err := services.Projects.TransitionGenerationTask(req.Context(), owner, projectID, running.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: results, SelectedResultID: results[0].ID, Now: services.Now().UTC()})
		if err != nil {
			writeAgentAudioError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": awaiting, "assets": assets})
	})))
	root.Handle("GET /api/agent/tasks/{id}/audio-assets/{assetId}/waveform", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if services.Waveform == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "音频波形服务尚未配置"})
			return
		}
		owner := strings.TrimSpace(identity.Username)
		projectID := strings.TrimSpace(req.PathValue("id"))
		asset, err := services.Assets.Asset(req.Context(), owner, projectID, strings.TrimSpace(req.PathValue("assetId")))
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		if asset.Kind != "audio" || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(asset.ContentType)), "audio/") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "仅支持音频素材波形预览"})
			return
		}
		body, err := services.Objects.Open(req.Context(), asset)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频素材读取失败"})
			return
		}
		defer body.Close()
		contents, err := io.ReadAll(io.LimitReader(body, agentAssetUploadMaxBytes+1))
		if err != nil || len(contents) == 0 || len(contents) > agentAssetUploadMaxBytes {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频素材读取失败"})
			return
		}
		summary, err := services.Waveform.Summary(req.Context(), asset, contents)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "音频波形解析失败"})
			return
		}
		writeJSON(w, http.StatusOK, summary)
	})))
}

const agentAudioLease = 3 * time.Minute

type agentAudioCoordinator struct {
	services AgentAudioServices
	mu       sync.Mutex
	inFlight map[string]struct{}
	now      func() time.Time
}

func newAgentAudioCoordinator(services AgentAudioServices) *agentAudioCoordinator {
	now := services.Now
	if now == nil {
		now = time.Now
	}
	return &agentAudioCoordinator{services: services, inFlight: map[string]struct{}{}, now: now}
}

func (s *agentAudioCoordinator) Run(ctx context.Context) {
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

func (s *agentAudioCoordinator) pollPendingOnce(ctx context.Context) error {
	if s == nil || s.services.Projects == nil {
		return nil
	}
	tasks, err := s.services.Projects.ListRunnableGenerationTasksByModes(ctx, []string{"speech", "music", "sound_effect", "trim", "speed", "smart_split", "custom_split"}, 100)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Kind != "audio" || (task.State == agentworkspace.GenerationRunning && s.now().UTC().Sub(task.UpdatedAt) < agentAudioLease) || !s.claim(task.ID) {
			continue
		}
		go func(task agentworkspace.GenerationTask) {
			defer s.release(task.ID)
			_ = s.execute(context.Background(), task)
		}(task)
	}
	return nil
}

func (s *agentAudioCoordinator) claim(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inFlight[id]; ok {
		return false
	}
	s.inFlight[id] = struct{}{}
	return true
}
func (s *agentAudioCoordinator) release(id string) {
	s.mu.Lock()
	delete(s.inFlight, id)
	s.mu.Unlock()
}

func (s *agentAudioCoordinator) execute(ctx context.Context, task agentworkspace.GenerationTask) error {
	services := s.services
	if task.State == agentworkspace.GenerationQueued {
		var err error
		task, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: s.now().UTC()})
		if err != nil {
			return err
		}
	}
	settings, _ := task.Settings["parameters"].(map[string]any)
	advanced, _ := task.Settings["advanced"].(map[string]any)
	input, err := agentworkspace.NormalizeAudioJobRequest(agentworkspace.AudioJobRequest{Operation: agentworkspace.AudioOperation(task.Mode), Prompt: stringValue(task.Settings["prompt"]), Provider: task.Provider, ModelID: task.ModelID, SourceAssetID: stringValue(task.Settings["sourceAssetId"]), Settings: settings, Advanced: advanced})
	if err != nil {
		return s.fail(ctx, task, "音频操作或参数无效")
	}
	outputs, err := runAgentAudio(ctx, task.OwnerUsername, task.ProjectID, input, services)
	if err != nil {
		return s.fail(ctx, task, err.Error())
	}
	_, results, err := persistAgentAudioOutputs(ctx, task.OwnerUsername, task.ProjectID, outputs, services)
	if err != nil {
		return s.fail(ctx, task, "音频素材保存失败")
	}
	current, err := services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil || current.State == agentworkspace.GenerationCancelled {
		return err
	}
	_, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: results, SelectedResultID: results[0].ID, ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}

func (s *agentAudioCoordinator) fail(ctx context.Context, task agentworkspace.GenerationTask, message string) error {
	current, err := s.services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil || current.State == agentworkspace.GenerationCancelled || current.State == agentworkspace.GenerationFailed {
		return err
	}
	_, err = s.services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: strings.TrimSpace(message), ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}

func agentAudioNode(canvas agentworkspace.Canvas, id string) (agentworkspace.Node, bool) {
	for _, node := range canvas.Nodes {
		if node.ID == strings.TrimSpace(id) && node.Type == "audio" {
			return node, true
		}
	}
	return agentworkspace.Node{}, false
}

func compactAudioReferences(sourceID string) []string {
	if strings.TrimSpace(sourceID) == "" {
		return []string{}
	}
	return []string{strings.TrimSpace(sourceID)}
}

func runAgentAudio(ctx context.Context, owner, projectID string, input agentworkspace.AudioJobRequest, services AgentAudioServices) ([]agentworkspace.AudioOutput, error) {
	if input.Operation.IsGeneration() {
		if services.Generator == nil {
			return nil, fmt.Errorf("当前音频模型尚未配置")
		}
		return services.Generator.Generate(ctx, owner, agentworkspace.AudioGenerateInput{Operation: input.Operation, Prompt: input.Prompt, Provider: input.Provider, ModelID: input.ModelID, Settings: input.Settings, Advanced: input.Advanced})
	}
	if services.Processor == nil {
		return nil, fmt.Errorf("音频处理器尚未配置")
	}
	source, err := services.Assets.Asset(ctx, owner, projectID, input.SourceAssetID)
	if err != nil {
		return nil, err
	}
	if source.Kind != "audio" {
		return nil, agentworkspace.ErrInvalidAudioJob
	}
	body, err := services.Objects.Open(ctx, source)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	contents, err := io.ReadAll(io.LimitReader(body, agentAssetUploadMaxBytes+1))
	if err != nil || len(contents) == 0 || len(contents) > agentAssetUploadMaxBytes {
		return nil, fmt.Errorf("源音频读取失败")
	}
	return services.Processor.Process(ctx, agentworkspace.AudioProcessInput{Operation: input.Operation, Source: source, SourceBytes: contents, Settings: input.Settings, Advanced: input.Advanced})
}

func persistAgentAudioOutputs(ctx context.Context, owner, projectID string, outputs []agentworkspace.AudioOutput, services AgentAudioServices) ([]agentworkspace.Asset, []agentworkspace.GenerationResult, error) {
	if len(outputs) == 0 {
		return nil, nil, fmt.Errorf("音频服务没有返回结果")
	}
	assets := make([]agentworkspace.Asset, 0, len(outputs))
	results := make([]agentworkspace.GenerationResult, 0, len(outputs))
	for index, output := range outputs {
		if len(output.Bytes) == 0 || !strings.HasPrefix(strings.ToLower(strings.TrimSpace(output.ContentType)), "audio/") {
			return nil, nil, fmt.Errorf("音频服务返回了无效文件")
		}
		extension := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(output.Extension)), ".")
		if extension == "" {
			extension = "mp3"
		}
		assetID := newAgentID("audio")
		objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, assetID, extension)
		if err != nil {
			return nil, nil, err
		}
		name := strings.TrimSpace(output.Name)
		if name == "" {
			name = fmt.Sprintf("音频结果-%d.%s", index+1, extension)
		}
		asset := agentworkspace.Asset{ID: assetID, ProjectID: projectID, Kind: "audio", Name: path.Base(name), ContentType: output.ContentType, ObjectKey: objectKey, ByteSize: int64(len(output.Bytes)), CreatedAt: services.Now().UTC()}
		if err := services.Objects.Put(ctx, asset, bytes.NewReader(output.Bytes)); err != nil {
			return nil, nil, err
		}
		if err := services.Assets.CreateAsset(ctx, owner, asset); err != nil {
			_ = services.Objects.Delete(ctx, asset)
			return nil, nil, err
		}
		asset.URL = agentAssetContentURL(projectID, asset.ID)
		assets = append(assets, asset)
		results = append(results, agentworkspace.GenerationResult{ID: asset.ID, AssetID: asset.ID, Kind: "audio", Asset: map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": "audio", "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType, "byteSize": asset.ByteSize}, CreatedAt: asset.CreatedAt})
	}
	return assets, results, nil
}

func writeAgentAudioError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrProjectNotFound), errors.Is(err, agentworkspace.ErrAssetNotFound), errors.Is(err, agentworkspace.ErrGenerationNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "音频项目、任务或素材不存在"})
	case errors.Is(err, agentworkspace.ErrGenerationNodeChanged), errors.Is(err, agentworkspace.ErrGenerationState):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "音频节点或任务状态已变化"})
	case errors.Is(err, agentworkspace.ErrInvalidAudioJob), errors.Is(err, agentworkspace.ErrGenerationInvalid):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "音频操作或参数无效"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "音频任务执行失败"})
	}
}
