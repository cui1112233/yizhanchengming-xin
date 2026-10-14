package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localexecutor"
)

type scriptVideoConfigResolver func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error)

type scriptVideoAdapter interface {
	batchfactoryv11.ProductionAdapter
	batchfactoryv11.ProductionPoller
}

type scriptVideoRoutes struct {
	resolver      scriptVideoConfigResolver
	client        *http.Client
	localExecutor *localexecutor.Service
}

func RegisterScriptVideoRoutes(root *http.ServeMux, auth BridgeAuth, client any, resolver scriptVideoConfigResolver) {
	var httpClient *http.Client
	var executor *localexecutor.Service
	switch typed := client.(type) {
	case *http.Client:
		httpClient = typed
	case *localexecutor.Service:
		executor = typed
	}
	if root == nil {
		return
	}
	routes := &scriptVideoRoutes{resolver: resolver, client: httpClient, localExecutor: executor}
	handler := auth.Middleware(http.HandlerFunc(routes.handle))
	root.Handle("POST /api/script-video", handler)
	root.Handle("GET /api/script-video/{taskId}", handler)
	root.Handle("GET /api/script-video/{taskId}/download", handler)
}

func (r *scriptVideoRoutes) handle(w http.ResponseWriter, req *http.Request) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if req.Method == http.MethodPost {
		r.submit(w, req, identity.Username)
		return
	}
	if strings.HasSuffix(req.URL.Path, "/download") {
		r.download(w, req, identity.Username)
		return
	}
	r.poll(w, req, identity.Username)
}

// download is the Go-owned replacement for the old Node bridge endpoint. It
// deliberately resolves the task once more so a stale browser link cannot
// download a task that is still processing or has failed.
func (r *scriptVideoRoutes) download(w http.ResponseWriter, req *http.Request, owner string) {
	taskID := strings.TrimSuffix(strings.TrimSpace(req.PathValue("taskId")), "/download")
	if strings.HasPrefix(taskID, localScriptVideoTaskPrefix) {
		if r.localExecutor == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "本地执行器服务不可用"})
			return
		}
		job, err := r.localExecutor.GetJob(req.Context(), owner, strings.TrimPrefix(taskID, localScriptVideoTaskPrefix))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		if job.State != localexecutor.JobSucceeded || strings.TrimSpace(job.ArtifactID) == "" {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "status": string(job.State), "error": "视频尚未生成完成"})
			return
		}
		http.Redirect(w, req, "/api/shuihuo-production/local-executor-artifacts/"+url.PathEscape(job.ArtifactID), http.StatusFound)
		return
	}
	prefix, provider := scriptVideoProvider(taskID)
	if prefix == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported video task"})
		return
	}
	if r.resolver == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "视频模型配置不可用"})
		return
	}
	modelKey := map[string]string{batchfactoryv11.VideoProviderAutoDLH3: "minimax-h3-video", batchfactoryv11.VideoProviderYFAISeedance: "seedance-2-0-official"}[provider]
	cfg, err := r.resolver(req.Context(), owner, modelKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	adapter, err := newScriptVideoAdapter(cfg, provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	setScriptVideoClient(adapter, r.client)
	ref, err := adapter.Poll(req.Context(), batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: strings.TrimPrefix(taskID, prefix)})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	if ref.State == batchfactoryv11.ProductionFailed {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "status": "failed", "error": ref.ErrorMessage})
		return
	}
	if ref.State != batchfactoryv11.ProductionSucceeded || strings.TrimSpace(ref.MediaURL) == "" {
		writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "status": "processing", "error": "视频尚未生成完成"})
		return
	}
	http.Redirect(w, req, ref.MediaURL, http.StatusFound)
}

func scriptVideoProvider(taskID string) (string, string) {
	for _, item := range []struct{ prefix, provider string }{{"h3:", batchfactoryv11.VideoProviderAutoDLH3}, {"yfai:", batchfactoryv11.VideoProviderYFAISeedance}} {
		if strings.HasPrefix(taskID, item.prefix) {
			return item.prefix, item.provider
		}
	}
	return "", ""
}

func (r *scriptVideoRoutes) submit(w http.ResponseWriter, req *http.Request, owner string) {
	var input scriptVideoLocalInput
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	input.ModelKey = strings.TrimSpace(input.ModelKey)
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "分镜视频提示词不能为空"})
		return
	}
	if len([]rune(input.Prompt)) > 12000 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "分镜视频提示词不能超过 12000 个字符"})
		return
	}
	spec, known := scriptVideoModelSpec(input.ModelKey)
	if !known {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported video provider"})
		return
	}
	if len(input.ImageURLs) > 0 && !spec.AcceptsReferenceImages {
		writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "status": "unsupported_reference_images", "code": "UNSUPPORTED_REFERENCE_IMAGES", "error": "当前视频模型不支持参考图，是否允许无参考图生成"})
		return
	}
	if spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		if r.localExecutor == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "本地执行器服务不可用"})
			return
		}
		job, err := r.localExecutor.CreateJob(req.Context(), owner, localexecutor.CreateJobInput{SourceTaskID: "script-video", Platform: localexecutor.PlatformDoubao, Payload: map[string]any{"prompt": input.Prompt, "duration": input.Duration, "aspectRatio": input.AspectRatio, "resolution": input.Resolution, "imageUrls": input.ImageURLs}})
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "taskId": localScriptVideoTaskPrefix + job.ID, "status": "processing"})
		return
	}
	if r.resolver == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "视频模型配置不可用"})
		return
	}
	cfg, err := r.resolver(req.Context(), owner, input.ModelKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	model := batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15, MaxReferenceImages: 9}
	prompt := batchfactoryv11.FinalPrompt{CompiledPrompt: input.Prompt, DurationSeconds: durationValue(input.Duration), ReferenceImageURLs: input.ImageURLs, EffectiveSettings: batchfactoryv11.EffectiveSettings{Values: batchfactoryv11.SettingsPatch{}}}
	adapter, err := newScriptVideoAdapter(cfg, spec.Provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	setScriptVideoClient(adapter, r.client)
	ref, err := adapter.Submit(req.Context(), model, prompt)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "taskId": spec.PublicTaskPrefix + ref.ProviderTaskID, "status": "processing", "provider": spec.Provider})
}

func (r *scriptVideoRoutes) poll(w http.ResponseWriter, req *http.Request, owner string) {
	taskID := strings.TrimSpace(req.PathValue("taskId"))
	prefix := ""
	provider := ""
	if strings.HasPrefix(taskID, localScriptVideoTaskPrefix) {
		if r.localExecutor == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "本地执行器服务不可用"})
			return
		}
		job, err := r.localExecutor.GetJob(req.Context(), owner, strings.TrimPrefix(taskID, localScriptVideoTaskPrefix))
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		result := map[string]any{"ok": true, "taskId": taskID, "status": "processing"}
		switch job.State {
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			result["status"], result["error"] = "failed", job.ErrorMessage
		case localexecutor.JobSucceeded:
			result["status"] = "succeeded"
			result["videoUrl"] = "/api/shuihuo-production/local-executor-artifacts/" + job.ArtifactID
		}
		writeJSON(w, http.StatusOK, result)
		return
	}
	for _, item := range []struct{ prefix, provider string }{{"h3:", batchfactoryv11.VideoProviderAutoDLH3}, {"yfai:", batchfactoryv11.VideoProviderYFAISeedance}} {
		if strings.HasPrefix(taskID, item.prefix) {
			prefix, provider = item.prefix, item.provider
			break
		}
	}
	if prefix == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported video task"})
		return
	}
	modelKey := map[string]string{batchfactoryv11.VideoProviderAutoDLH3: "minimax-h3-video", batchfactoryv11.VideoProviderYFAISeedance: "seedance-2-0-official"}[provider]
	if r.resolver == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "视频模型配置不可用"})
		return
	}
	cfg, err := r.resolver(req.Context(), owner, modelKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	adapter, err := newScriptVideoAdapter(cfg, provider)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	setScriptVideoClient(adapter, r.client)
	ref, err := adapter.Poll(req.Context(), batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: strings.TrimPrefix(taskID, prefix)})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	result := map[string]any{"ok": true, "taskId": taskID, "status": "processing", "provider": provider}
	if ref.State == batchfactoryv11.ProductionFailed {
		result["status"] = "failed"
		result["error"] = ref.ErrorMessage
	} else if ref.State == batchfactoryv11.ProductionSucceeded {
		result["status"] = "succeeded"
		result["videoUrl"] = ref.MediaURL
	}
	writeJSON(w, http.StatusOK, result)
}

func setScriptVideoClient(adapter scriptVideoAdapter, client *http.Client) {
	if client == nil {
		return
	}
	switch typed := adapter.(type) {
	case *batchfactoryv11.AutoDLH3VideoAdapter:
		typed.Client = client
	case *batchfactoryv11.YFAISeedanceAdapter:
		typed.Client = client
	case *batchfactoryv11.YadiVideoAdapter:
		typed.Client = client
	}
}

func newScriptVideoAdapter(cfg batchfactoryv11.VideoProviderConfig, provider string) (scriptVideoAdapter, error) {
	switch provider {
	case batchfactoryv11.VideoProviderAutoDLH3:
		return &batchfactoryv11.AutoDLH3VideoAdapter{CreateURL: cfg.CreateURL, TasksURL: cfg.TasksURL, APIKey: cfg.APIKey, Model: cfg.Model, WorkflowID: cfg.WorkflowID}, nil
	case batchfactoryv11.VideoProviderYFAISeedance:
		return &batchfactoryv11.YFAISeedanceAdapter{BaseURL: cfg.CreateURL, APIKey: cfg.APIKey, Model: cfg.Model}, nil
	case batchfactoryv11.VideoProviderPersonalAPI:
		return &batchfactoryv11.YadiVideoAdapter{CreateURL: cfg.CreateURL, TasksURL: cfg.TasksURL, ResultURL: cfg.ResultURL, APIKey: cfg.APIKey, Model: cfg.Model}, nil
	default:
		return nil, fmt.Errorf("unsupported video provider")
	}
}

func durationValue(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case string:
		var result int
		_, _ = fmt.Sscanf(typed, "%d", &result)
		return result
	default:
		return 10
	}
}
