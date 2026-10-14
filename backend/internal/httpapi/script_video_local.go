package httpapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/localexecutor"
)

const localScriptVideoTaskPrefix = "go-local:"

type scriptVideoLocalInput struct {
	ModelKey    string   `json:"modelKey"`
	Prompt      string   `json:"prompt"`
	BookID      string   `json:"bookId"`
	VideoID     string   `json:"videoId"`
	Model       string   `json:"model"`
	Duration    any      `json:"duration"`
	AspectRatio string   `json:"aspectRatio"`
	Resolution  string   `json:"resolution"`
	ImageURLs   []string `json:"imageUrls"`
}

func validateLocalScriptVideoInput(input scriptVideoLocalInput) error {
	spec, ok := scriptVideoModelSpec(strings.TrimSpace(input.ModelKey))
	if !ok || spec.Provider != "doubao_local_executor" {
		return fmt.Errorf("unsupported local video model")
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return fmt.Errorf("prompt is required")
	}
	return nil
}

func RegisterLocalScriptVideoRoutes(root *http.ServeMux, auth BridgeAuth, service *localexecutor.Service) {
	if root == nil || service == nil {
		return
	}
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
			return
		}
		if req.Method == http.MethodPost {
			var input scriptVideoLocalInput
			if err := decodeExecutorJSON(w, req, &input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request"})
				return
			}
			if err := validateLocalScriptVideoInput(input); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
				return
			}
			model := strings.TrimSpace(input.Model)
			if model == "" {
				model = "doubao"
			}
			source := "script-video:" + time.Now().UTC().Format("20060102150405.000000000")
			job, err := service.CreateJob(req.Context(), identity.Username, localexecutor.CreateJobInput{
				SourceTaskID: source,
				Platform:     localexecutor.PlatformDoubao,
				Payload: map[string]any{
					"bookId": input.BookID, "videoId": input.VideoID, "model": model,
					"prompt": input.Prompt, "duration": input.Duration, "aspectRatio": input.AspectRatio,
					"resolution": input.Resolution, "imageUrls": input.ImageURLs,
				},
			})
			if err != nil {
				writeLocalJobError(w, err)
				return
			}
			writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "taskId": localScriptVideoTaskPrefix + job.ID, "status": "processing"})
			return
		}
		id := strings.TrimPrefix(req.PathValue("taskId"), localScriptVideoTaskPrefix)
		job, err := service.GetJob(req.Context(), identity.Username, id)
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		status := "processing"
		result := map[string]any{"ok": true, "taskId": localScriptVideoTaskPrefix + job.ID, "status": status}
		switch job.State {
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			result["status"] = "failed"
			result["error"] = job.ErrorMessage
		case localexecutor.JobSucceeded:
			result["status"] = "succeeded"
			if job.ArtifactID != "" {
				result["videoUrl"] = "/api/shuihuo-production/local-executor-artifacts/" + job.ArtifactID
			}
		}
		writeJSON(w, http.StatusOK, result)
	}))
	root.Handle("POST /api/script-video", handler)
	root.Handle("GET /api/script-video/{taskId}", handler)
}
