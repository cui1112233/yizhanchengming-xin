package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

type agentVideoCoordinator struct {
	projects   agentworkspace.Store
	assets     agentworkspace.AssetStore
	references agentworkspace.ReferenceURLProvider
	videos     *scriptVideoService
	ingestor   *agentVideoAssetIngestor
	executor   *localexecutor.Service
	artifacts  *localartifact.Store
	now        func() time.Time
}

type agentVideoJobRequest struct {
	NodeID            string         `json:"nodeId"`
	NodeVersion       int64          `json:"nodeVersion"`
	IdempotencyKey    string         `json:"idempotencyKey"`
	Prompt            string         `json:"prompt"`
	Provider          string         `json:"provider"`
	ModelID           string         `json:"modelId"`
	Mode              string         `json:"mode"`
	Parameters        map[string]any `json:"parameters"`
	Advanced          map[string]any `json:"advanced"`
	ModelKey          string         `json:"modelKey"`
	Duration          any            `json:"duration"`
	AspectRatio       string         `json:"aspectRatio"`
	Resolution        string         `json:"resolution"`
	ReferenceAssetIDs []string       `json:"referenceAssetIds"`
	IgnoreReferences  bool           `json:"ignoreReferences"`
}

var errUnsupportedVideoReference = errors.New("unsupported video reference")
var errAgentVideoSettings = errors.New("invalid agent video settings")

const agentVideoIngestLease = 5 * time.Minute

func RegisterAgentVideoJobRoutes(root *http.ServeMux, middleware BridgeAuth, coordinator *agentVideoCoordinator) {
	if root == nil || coordinator == nil || coordinator.projects == nil || coordinator.assets == nil || coordinator.videos == nil || coordinator.ingestor == nil {
		return
	}
	if coordinator.now == nil {
		coordinator.now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	handle("POST /api/agent/tasks/{id}/video-jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input agentVideoJobRequest
		if err := decodeAgentVideoJobRequest(w, req, &input); err != nil {
			return
		}
		job, err := coordinator.submit(req.Context(), owner, req.PathValue("id"), input)
		if err != nil {
			writeAgentVideoJobSubmissionError(w, req.Context(), coordinator, owner, job, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": coordinator.viewJob(req.Context(), owner, job)})
	})
	handle("GET /api/agent/tasks/{id}/video-jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		jobs, err := coordinator.projects.ListVideoJobs(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		result := make([]map[string]any, 0, len(jobs))
		for _, job := range jobs {
			result = append(result, coordinator.viewJob(req.Context(), owner, job))
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": result})
	})
	handle("GET /api/agent/tasks/{id}/video-jobs/{jobID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := coordinator.projects.VideoJob(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"))
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		job, err = coordinator.poll(req.Context(), owner, job)
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": coordinator.viewJob(req.Context(), owner, job)})
	})
	handle("POST /api/agent/tasks/{id}/video-jobs/{jobID}/confirm", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := coordinator.projects.ConfirmVideoJob(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"), coordinator.now().UTC())
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": coordinator.viewJob(req.Context(), owner, job)})
	})
	handle("POST /api/agent/tasks/{id}/video-jobs/{jobID}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := coordinator.projects.VideoJob(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"))
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		job, err = coordinator.cancel(req.Context(), owner, job)
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": coordinator.viewJob(req.Context(), owner, job)})
	})
	handle("POST /api/agent/tasks/{id}/video-jobs/{jobID}/retry", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			IdempotencyKey string `json:"idempotencyKey"`
		}
		if err := decodeAgentVideoJobRequest(w, req, &input); err != nil {
			return
		}
		job, err := coordinator.retry(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"), input.IdempotencyKey)
		if err != nil {
			writeAgentVideoJobSubmissionError(w, req.Context(), coordinator, owner, job, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": coordinator.viewJob(req.Context(), owner, job)})
	})
}

func decodeAgentVideoJobRequest(w http.ResponseWriter, req *http.Request, target any) error {
	req.Body = http.MaxBytesReader(w, req.Body, 256<<10)
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "视频任务参数无效"})
		return err
	}
	return nil
}

func (s *agentVideoCoordinator) submit(ctx context.Context, owner, projectID string, request agentVideoJobRequest) (agentworkspace.VideoJob, error) {
	canvas, err := s.projects.Canvas(ctx, owner, strings.TrimSpace(projectID))
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	node, ok := agentVideoNode(canvas, request.NodeID)
	if !ok {
		return agentworkspace.VideoJob{}, fmt.Errorf("视频节点不存在")
	}
	if node.Type != "video" {
		return agentworkspace.VideoJob{}, fmt.Errorf("只能为视频节点生成视频")
	}
	if request.NodeVersion > 0 && request.NodeVersion != node.Version {
		return agentworkspace.VideoJob{}, agentworkspace.ErrVideoJobNodeChanged
	}
	normalized, settings, err := normalizeAgentVideoJobRequest(request)
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	references, ids, err := s.referenceURLs(ctx, owner, projectID, request.ReferenceAssetIDs, request.IgnoreReferences)
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	job, err := s.projects.CreateVideoJob(ctx, owner, agentworkspace.CreateVideoJobInput{ID: newAgentID("video-job"), ProjectID: projectID, NodeID: node.ID, NodeVersion: node.Version, ModelKey: normalized.ModelKey, Settings: settings, ReferenceAssetIDs: ids, IdempotencyKey: strings.TrimSpace(request.IdempotencyKey), Now: s.now().UTC()})
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	return s.submitCreated(ctx, owner, job, references)
}

func (s *agentVideoCoordinator) retry(ctx context.Context, owner, projectID, jobID, idempotencyKey string) (agentworkspace.VideoJob, error) {
	job, err := s.projects.RetryVideoJob(ctx, owner, strings.TrimSpace(projectID), strings.TrimSpace(jobID), agentworkspace.RetryVideoJobInput{ID: newAgentID("video-job"), IdempotencyKey: strings.TrimSpace(idempotencyKey), Now: s.now().UTC()})
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	ignoreReferences, _ := job.Settings["ignoreReferences"].(bool)
	references, _, err := s.referenceURLs(ctx, owner, job.ProjectID, job.ReferenceAssetIDs, ignoreReferences)
	if err != nil {
		job.State, job.ErrorMessage = agentworkspace.VideoJobFailed, err.Error()
		failed, updateErr := s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
		if updateErr != nil {
			return agentworkspace.VideoJob{}, updateErr
		}
		return failed, err
	}
	return s.submitCreated(ctx, owner, job, references)
}

func (s *agentVideoCoordinator) submitCreated(ctx context.Context, owner string, job agentworkspace.VideoJob, references []string) (agentworkspace.VideoJob, error) {
	settings := job.Settings
	submission, err := s.videos.Submit(ctx, owner, scriptVideoLocalInput{
		SourceTaskID: "agent-video:" + job.ID, ModelKey: job.ModelKey, Model: valueString(settings["modelId"]),
		Prompt: valueString(settings["prompt"]), Duration: settings["durationSeconds"], AspectRatio: valueString(settings["aspectRatio"]),
		Resolution: valueString(settings["resolution"]), ImageURLs: references, GenerateAudio: valueBool(settings["generateAudio"]),
	})
	if err != nil {
		job.State, job.ErrorMessage = agentworkspace.VideoJobFailed, err.Error()
		failedJob, updateErr := s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
		if updateErr != nil {
			return agentworkspace.VideoJob{}, updateErr
		}
		return failedJob, err
	}
	job.State, job.ProviderTaskID, job.ErrorMessage = agentworkspace.VideoJobQueued, submission.TaskID, ""
	return s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
}

func normalizeAgentVideoJobRequest(request agentVideoJobRequest) (agentVideoJobRequest, map[string]any, error) {
	parameters := map[string]any{}
	for key, value := range request.Parameters {
		parameters[key] = value
	}
	if _, ok := parameters["durationSeconds"]; !ok && request.Duration != nil {
		parameters["durationSeconds"] = request.Duration
	}
	if _, ok := parameters["ratio"]; !ok && strings.TrimSpace(request.AspectRatio) != "" {
		parameters["ratio"] = strings.TrimSpace(request.AspectRatio)
	}
	if _, ok := parameters["resolution"]; !ok && strings.TrimSpace(request.Resolution) != "" {
		parameters["resolution"] = strings.ToUpper(strings.TrimSpace(request.Resolution))
	}
	if strings.TrimSpace(request.Mode) != "" {
		parameters["mode"] = strings.TrimSpace(request.Mode)
	}
	if err := ValidateModelSettings(map[string]any{"kind": "video"}, "video", parameters); err != nil {
		return agentVideoJobRequest{}, nil, fmt.Errorf("%w: %v", errAgentVideoSettings, err)
	}

	provider := strings.ToLower(strings.TrimSpace(request.Provider))
	modelID := strings.TrimSpace(request.ModelID)
	modelKey := strings.TrimSpace(request.ModelKey)
	if provider == "doubao_local_executor" {
		if !supportedAgentLocalVideoModel(modelID) {
			return agentVideoJobRequest{}, nil, fmt.Errorf("%w: modelId: 所选豆包视频模型不可用", errAgentVideoSettings)
		}
		modelKey = "local-doubao-executor-video"
	} else if modelKey == "local-doubao-executor-video" && provider == "" {
		provider, modelID = "doubao_local_executor", "Seedance 2.0 Fast"
	} else if modelKey == "" {
		modelKey = modelID
	}
	if modelKey == "" {
		return agentVideoJobRequest{}, nil, fmt.Errorf("%w: modelId: 请选择视频模型", errAgentVideoSettings)
	}
	if provider == "" {
		if spec, ok := scriptVideoModelSpec(modelKey); ok {
			provider = spec.Provider
		}
	}
	if modelID == "" {
		modelID = modelKey
	}

	duration := parameters["durationSeconds"]
	ratio := settingString(parameters, "ratio")
	resolution := strings.ToUpper(settingString(parameters, "resolution"))
	count := settingInt(parameters, "count")
	if count == 0 {
		count = 1
	}
	settings := map[string]any{
		"prompt": strings.TrimSpace(request.Prompt), "provider": provider, "modelId": modelID,
		"mode": strings.TrimSpace(request.Mode), "parameters": parameters, "advanced": request.Advanced,
		"duration": duration, "durationSeconds": duration, "aspectRatio": ratio, "resolution": resolution,
		"count": count, "generateAudio": valueBool(parameters["generateAudio"]), "ignoreReferences": request.IgnoreReferences,
	}
	request.Provider, request.ModelID, request.ModelKey = provider, modelID, modelKey
	return request, settings, nil
}

func supportedAgentLocalVideoModel(modelID string) bool {
	for _, supported := range []string{"Seedance 2.5", "Seedance 2.0", "Seedance 2.0 Fast", "Seedance 2.0 Mini"} {
		if strings.TrimSpace(modelID) == supported {
			return true
		}
	}
	return false
}

func (s *agentVideoCoordinator) referenceURLs(ctx context.Context, owner, projectID string, requestedIDs []string, ignore bool) ([]string, []string, error) {
	ids := make([]string, 0, len(requestedIDs))
	seen := map[string]bool{}
	for _, requestedID := range requestedIDs {
		id := strings.TrimSpace(requestedID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		if len(ids) == 10 {
			break
		}
	}
	if len(ids) == 0 {
		return []string{}, []string{}, nil
	}
	assets := make([]agentworkspace.Asset, 0, len(ids))
	for _, id := range ids {
		asset, err := s.assets.Asset(ctx, owner, projectID, id)
		if err != nil {
			return nil, nil, err
		}
		if asset.Kind != "image" {
			return nil, nil, fmt.Errorf("%w: 当前视频生成暂不支持视频参考素材，请移除后重试", errUnsupportedVideoReference)
		}
		assets = append(assets, asset)
	}
	if ignore {
		return []string{}, []string{}, nil
	}
	if s.references == nil {
		return nil, nil, agentworkspace.ErrReferenceURLUnavailable
	}
	urls := make([]string, 0, len(ids))
	for _, asset := range assets {
		ref, err := s.references.ReferenceURL(ctx, asset, 15*time.Minute)
		if err != nil {
			return nil, nil, err
		}
		urls = append(urls, ref)
	}
	return urls, ids, nil
}

func (s *agentVideoCoordinator) poll(ctx context.Context, owner string, job agentworkspace.VideoJob) (agentworkspace.VideoJob, error) {
	if job.State != agentworkspace.VideoJobPreparing && job.State != agentworkspace.VideoJobQueued && job.State != agentworkspace.VideoJobRunning && job.State != agentworkspace.VideoJobResultReady && job.State != agentworkspace.VideoJobIngesting {
		return job, nil
	}
	if strings.TrimSpace(job.ProviderTaskID) == "" {
		return job, nil
	}
	if job.State == agentworkspace.VideoJobIngesting && s.now().UTC().Sub(job.UpdatedAt) < agentVideoIngestLease {
		return job, nil
	}
	status, err := s.videos.Poll(ctx, owner, job.ProviderTaskID)
	if err != nil {
		return s.fail(ctx, owner, job, err)
	}
	if status.Status == "failed" {
		return s.fail(ctx, owner, job, errors.New(status.Error))
	}
	if status.Status != "succeeded" {
		job.State = agentworkspace.VideoJobRunning
		return s.updateOrLatest(ctx, owner, job)
	}
	if job.State != agentworkspace.VideoJobResultReady && job.State != agentworkspace.VideoJobIngesting {
		job.State = agentworkspace.VideoJobResultReady
		transition := job
		job, err = s.projects.UpdateVideoJob(ctx, owner, transition, s.now().UTC())
		if err != nil {
			if errors.Is(err, agentworkspace.ErrVideoJobState) {
				return s.projects.VideoJob(ctx, owner, transition.ProjectID, transition.ID)
			}
			return agentworkspace.VideoJob{}, err
		}
	}
	job.State = agentworkspace.VideoJobIngesting
	job, err = s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
	if err != nil {
		if errors.Is(err, agentworkspace.ErrVideoJobState) {
			return s.projects.VideoJob(ctx, owner, job.ProjectID, job.ID)
		}
		return agentworkspace.VideoJob{}, err
	}
	var asset agentworkspace.Asset
	resultAssetID := "video-result-" + job.ID
	if status.ArtifactID != "" {
		asset, err = s.ingestor.SaveLocal(ctx, owner, job.ProjectID, resultAssetID, status.ArtifactID, s.executor, s.artifacts)
	} else {
		asset, err = s.ingestor.SaveRemote(ctx, owner, job.ProjectID, resultAssetID, status.VideoURL)
	}
	if err != nil {
		return s.fail(ctx, owner, job, err)
	}
	job.State, job.ResultAssetID, job.ErrorMessage = agentworkspace.VideoJobAwaitingConfirmation, asset.ID, ""
	return s.updateOrLatest(ctx, owner, job)
}

func (s *agentVideoCoordinator) fail(ctx context.Context, owner string, job agentworkspace.VideoJob, err error) (agentworkspace.VideoJob, error) {
	job.State, job.ErrorMessage = agentworkspace.VideoJobFailed, strings.TrimSpace(err.Error())
	return s.updateOrLatest(ctx, owner, job)
}

func (s *agentVideoCoordinator) cancel(ctx context.Context, owner string, job agentworkspace.VideoJob) (agentworkspace.VideoJob, error) {
	job.State = agentworkspace.VideoJobCanceling
	claimed, err := s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	job = claimed
	err = s.videos.Cancel(ctx, owner, job.ProviderTaskID)
	if err != nil {
		job.State, job.ErrorMessage = agentworkspace.VideoJobAbandoned, err.Error()
	} else {
		job.State, job.ErrorMessage = agentworkspace.VideoJobCanceled, ""
	}
	return s.updateOrLatest(ctx, owner, job)
}

func (s *agentVideoCoordinator) updateOrLatest(ctx context.Context, owner string, job agentworkspace.VideoJob) (agentworkspace.VideoJob, error) {
	updated, err := s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
	if !errors.Is(err, agentworkspace.ErrVideoJobState) {
		return updated, err
	}
	return s.projects.VideoJob(ctx, owner, job.ProjectID, job.ID)
}

func (s *agentVideoCoordinator) viewJob(ctx context.Context, owner string, job agentworkspace.VideoJob) map[string]any {
	result := map[string]any{"id": job.ID, "projectId": job.ProjectID, "nodeId": job.NodeID, "nodeVersion": job.NodeVersion, "state": job.State, "modelKey": job.ModelKey, "providerTaskId": job.ProviderTaskID, "settings": job.Settings, "referenceAssetIds": job.ReferenceAssetIDs, "resultAssetId": job.ResultAssetID, "errorMessage": job.ErrorMessage, "createdAt": job.CreatedAt, "updatedAt": job.UpdatedAt}
	if job.ResultAssetID != "" {
		if asset, err := s.assets.Asset(ctx, owner, job.ProjectID, job.ResultAssetID); err == nil {
			asset.URL = agentAssetContentURL(asset.ProjectID, asset.ID)
			result["resultAsset"] = asset
		}
	}
	return result
}

func agentVideoNode(canvas agentworkspace.Canvas, id string) (agentworkspace.Node, bool) {
	for _, node := range canvas.Nodes {
		if node.ID == strings.TrimSpace(id) {
			return node, true
		}
	}
	return agentworkspace.Node{}, false
}
func valueString(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}
func valueBool(value any) bool {
	value, _ = value.(bool)
	return value == true
}
func writeAgentVideoJobError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errAgentVideoSettings):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": strings.TrimPrefix(err.Error(), errAgentVideoSettings.Error()+": ")})
	case errors.Is(err, agentworkspace.ErrVideoJobNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "视频任务不存在"})
	case errors.Is(err, agentworkspace.ErrVideoJobExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "重复的视频任务"})
	case errors.Is(err, agentworkspace.ErrVideoJobNodeChanged):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "视频节点已被修改，请重新生成"})
	case errors.Is(err, agentworkspace.ErrVideoJobState):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "当前视频任务状态不允许该操作"})
	case errors.Is(err, agentworkspace.ErrReferenceURLUnavailable):
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "当前素材存储无法提供模型参考图"})
	case errors.Is(err, errUnsupportedReferenceImages):
		writeJSON(w, http.StatusConflict, map[string]any{"code": "UNSUPPORTED_REFERENCE_IMAGES", "error": "当前视频模型不支持参考图，是否允许无参考图生成"})
	case errors.Is(err, errUnsupportedVideoReference):
		writeJSON(w, http.StatusConflict, map[string]any{"code": "UNSUPPORTED_VIDEO_REFERENCE", "error": "当前视频生成暂不支持视频参考素材，请移除后重试"})
	case errors.Is(err, agentworkspace.ErrAssetNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "参考素材不存在"})
	case errors.Is(err, agentworkspace.ErrProjectNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "项目不存在"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
}

func writeAgentVideoJobSubmissionError(w http.ResponseWriter, ctx context.Context, coordinator *agentVideoCoordinator, owner string, job agentworkspace.VideoJob, err error) {
	if strings.TrimSpace(job.ID) == "" {
		writeAgentVideoJobError(w, err)
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]any{
		"error": err.Error(),
		"job":   coordinator.viewJob(ctx, owner, job),
	})
}
