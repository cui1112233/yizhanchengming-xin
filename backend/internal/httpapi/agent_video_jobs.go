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
	NodeID            string   `json:"nodeId"`
	IdempotencyKey    string   `json:"idempotencyKey"`
	Prompt            string   `json:"prompt"`
	ModelKey          string   `json:"modelKey"`
	Duration          any      `json:"duration"`
	AspectRatio       string   `json:"aspectRatio"`
	Resolution        string   `json:"resolution"`
	ReferenceAssetIDs []string `json:"referenceAssetIds"`
	IgnoreReferences  bool     `json:"ignoreReferences"`
}

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
		current, err := coordinator.projects.VideoJob(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"))
		if err != nil {
			writeAgentVideoJobError(w, err)
			return
		}
		request := requestFromJob(current, input.IdempotencyKey)
		job, err := coordinator.submit(req.Context(), owner, current.ProjectID, request)
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
	references, ids, err := s.referenceURLs(ctx, owner, projectID, request.ReferenceAssetIDs, request.IgnoreReferences)
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	settings := map[string]any{"prompt": strings.TrimSpace(request.Prompt), "duration": request.Duration, "aspectRatio": strings.TrimSpace(request.AspectRatio), "resolution": strings.TrimSpace(request.Resolution), "ignoreReferences": request.IgnoreReferences}
	job, err := s.projects.CreateVideoJob(ctx, owner, agentworkspace.CreateVideoJobInput{ID: newAgentID("video-job"), ProjectID: projectID, NodeID: node.ID, NodeVersion: node.Version, ModelKey: strings.TrimSpace(request.ModelKey), Settings: settings, ReferenceAssetIDs: ids, IdempotencyKey: strings.TrimSpace(request.IdempotencyKey), Now: s.now().UTC()})
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	submission, err := s.videos.Submit(ctx, owner, scriptVideoLocalInput{ModelKey: job.ModelKey, Prompt: valueString(settings["prompt"]), Duration: settings["duration"], AspectRatio: valueString(settings["aspectRatio"]), Resolution: valueString(settings["resolution"]), ImageURLs: references})
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

func (s *agentVideoCoordinator) referenceURLs(ctx context.Context, owner, projectID string, requestedIDs []string, ignore bool) ([]string, []string, error) {
	if ignore {
		return []string{}, []string{}, nil
	}
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
	if s.references == nil {
		return nil, nil, agentworkspace.ErrReferenceURLUnavailable
	}
	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		asset, err := s.assets.Asset(ctx, owner, projectID, id)
		if err != nil {
			return nil, nil, err
		}
		if asset.Kind != "image" {
			return nil, nil, fmt.Errorf("视频参考素材必须是图片")
		}
		ref, err := s.references.ReferenceURL(ctx, asset, 15*time.Minute)
		if err != nil {
			return nil, nil, err
		}
		urls = append(urls, ref)
	}
	return urls, ids, nil
}

func (s *agentVideoCoordinator) poll(ctx context.Context, owner string, job agentworkspace.VideoJob) (agentworkspace.VideoJob, error) {
	if job.State != agentworkspace.VideoJobPreparing && job.State != agentworkspace.VideoJobQueued && job.State != agentworkspace.VideoJobRunning && job.State != agentworkspace.VideoJobResultReady {
		return job, nil
	}
	if strings.TrimSpace(job.ProviderTaskID) == "" {
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
		return s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
	}
	job.State = agentworkspace.VideoJobResultReady
	job, err = s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
	if err != nil {
		return agentworkspace.VideoJob{}, err
	}
	var asset agentworkspace.Asset
	if status.ArtifactID != "" {
		asset, err = s.ingestor.SaveLocal(ctx, owner, job.ProjectID, status.ArtifactID, s.executor, s.artifacts)
	} else {
		asset, err = s.ingestor.SaveRemote(ctx, owner, job.ProjectID, status.VideoURL)
	}
	if err != nil {
		return s.fail(ctx, owner, job, err)
	}
	job.State, job.ResultAssetID, job.ErrorMessage = agentworkspace.VideoJobAwaitingConfirmation, asset.ID, ""
	return s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
}

func (s *agentVideoCoordinator) fail(ctx context.Context, owner string, job agentworkspace.VideoJob, err error) (agentworkspace.VideoJob, error) {
	job.State, job.ErrorMessage = agentworkspace.VideoJobFailed, strings.TrimSpace(err.Error())
	return s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
}

func (s *agentVideoCoordinator) cancel(ctx context.Context, owner string, job agentworkspace.VideoJob) (agentworkspace.VideoJob, error) {
	err := s.videos.Cancel(ctx, owner, job.ProviderTaskID)
	if err != nil {
		job.State, job.ErrorMessage = agentworkspace.VideoJobAbandoned, err.Error()
	} else {
		job.State, job.ErrorMessage = agentworkspace.VideoJobCanceled, ""
	}
	return s.projects.UpdateVideoJob(ctx, owner, job, s.now().UTC())
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
func requestFromJob(job agentworkspace.VideoJob, key string) agentVideoJobRequest {
	return agentVideoJobRequest{NodeID: job.NodeID, IdempotencyKey: key, Prompt: valueString(job.Settings["prompt"]), ModelKey: job.ModelKey, Duration: job.Settings["duration"], AspectRatio: valueString(job.Settings["aspectRatio"]), Resolution: valueString(job.Settings["resolution"]), ReferenceAssetIDs: append([]string(nil), job.ReferenceAssetIDs...)}
}

func writeAgentVideoJobError(w http.ResponseWriter, err error) {
	switch {
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
