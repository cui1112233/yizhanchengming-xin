package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	status, err := r.service().Poll(req.Context(), owner, taskID)
	if err != nil {
		r.writeServiceError(w, taskID, err)
		return
	}
	if status.Status == "failed" && !strings.HasPrefix(taskID, localScriptVideoTaskPrefix) {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "status": "failed", "error": status.Error})
		return
	}
	if status.Status != "succeeded" || status.VideoURL == "" {
		writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "status": status.Status, "error": "视频尚未生成完成"})
		return
	}
	http.Redirect(w, req, status.VideoURL, http.StatusFound)
}
func (r *scriptVideoRoutes) service() *scriptVideoService {
	return &scriptVideoService{resolver: r.resolver, client: r.client, localExecutor: r.localExecutor}
}

func scriptVideoProvider(taskID string) (string, string) {
	if prefix, provider, _ := nativeScriptTaskParts(taskID); prefix != "" {
		return prefix, provider
	}
	for _, item := range []struct{ prefix, provider string }{{"yd2:", batchfactoryv11.VideoProviderPersonalAPI}, {"h3:", batchfactoryv11.VideoProviderAutoDLH3}, {"yfai:fanke-video-api:", batchfactoryv11.VideoProviderYFAISeedance}, {"yfai:", batchfactoryv11.VideoProviderYFAISeedance}} {
		if strings.HasPrefix(taskID, item.prefix) {
			return item.prefix, item.provider
		}
	}
	return "", ""
}

func scriptVideoTaskModelKey(taskID, provider string) string {
	if _, _, key := nativeScriptTaskParts(taskID); key != "" {
		return key
	}
	if strings.HasPrefix(taskID, "yfai:fanke-video-api:") {
		return "fanke-video-api"
	}
	return map[string]string{batchfactoryv11.VideoProviderPersonalAPI: "yd2-mini-video", batchfactoryv11.VideoProviderAutoDLH3: "minimax-h3-video", batchfactoryv11.VideoProviderYFAISeedance: "seedance-2-0-official"}[provider]
}

func (r *scriptVideoRoutes) submit(w http.ResponseWriter, req *http.Request, owner string) {
	var input scriptVideoLocalInput
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 256<<10)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	out, err := r.service().Submit(req.Context(), owner, input)
	if err != nil {
		if errors.Is(err, errUnsupportedReferenceImages) {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "status": "unsupported_reference_images", "code": "UNSUPPORTED_REFERENCE_IMAGES", "error": "当前视频模型不支持参考图，是否允许无参考图生成"})
			return
		}

		var preflight *scriptVideoPreflightError
		preflightFailed := errors.As(err, &preflight)
		if input.ModelKey == "local-doubao-executor-video" {
			if r.localExecutor == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "本地执行器服务不可用"})
				return
			}
			if !preflightFailed {
				writeLocalJobError(w, err)
				return
			}
		}
		if spec, known := scriptVideoModelSpec(input.ModelKey); known && spec.Provider != batchfactoryv11.VideoProviderDoubaoLocal && r.resolver == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "视频模型配置不可用"})
			return
		}
		status := http.StatusBadGateway

		if errors.As(err, &preflight) {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "taskId": out.TaskID, "status": out.Status, "provider": out.Provider})
}
func (r *scriptVideoRoutes) poll(w http.ResponseWriter, req *http.Request, owner string) {
	out, err := r.service().Poll(req.Context(), owner, strings.TrimSpace(req.PathValue("taskId")))
	if err != nil {
		r.writeServiceError(w, strings.TrimSpace(req.PathValue("taskId")), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "taskId": out.TaskID, "status": out.Status, "provider": out.Provider, "videoUrl": out.VideoURL, "error": out.Error})
}

func (r *scriptVideoRoutes) writeServiceError(w http.ResponseWriter, taskID string, err error) {
	var preflight *scriptVideoPreflightError
	if errors.As(err, &preflight) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if strings.HasPrefix(taskID, localScriptVideoTaskPrefix) {
		if r.localExecutor == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "本地执行器服务不可用"})
			return
		}
		writeLocalJobError(w, err)
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
}

func setScriptVideoClient(adapter scriptVideoAdapter, client *http.Client) {
	if client == nil {
		return
	}
	switch typed := adapter.(type) {
	case *nativeScriptVideoAdapter:
		typed.Client = client
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
	case scriptVideoProviderArk, scriptVideoProviderSora:
		return &nativeScriptVideoAdapter{cfg: cfg}, nil
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
