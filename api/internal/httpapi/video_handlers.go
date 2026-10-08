package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type VideoService interface {
	Start(context.Context, video.StartRequest) (video.StartResult, error)
	PollTask(context.Context, int64) (video.ProductionTask, error)
	CancelTask(context.Context, int64) (video.ProductionTask, error)
	RetryTask(context.Context, int64, string) (video.StartResult, error)
}

type VideoConfigService interface {
	Get(context.Context, string, string) (video.ProviderConfigView, error)
	Save(context.Context, video.ProviderConfigInput) (video.ProviderConfigView, error)
	Status(context.Context, string, string) (video.ProviderStatusView, error)
}

func (h handler) getVideoProviderConfig(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoConfig == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video provider configuration service unavailable")
		return
	}
	view, err := h.deps.VideoConfig.Get(r.Context(), r.PathValue("provider"), r.PathValue("model"))
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h handler) getVideoProviderStatus(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoConfig == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video provider configuration service unavailable")
		return
	}
	view, err := h.deps.VideoConfig.Status(r.Context(), r.PathValue("provider"), r.PathValue("model"))
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h handler) putVideoProviderConfig(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoConfig == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video provider configuration service unavailable")
		return
	}
	defer r.Body.Close()
	var input video.ProviderConfigInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	input.ProviderKey = r.PathValue("provider")
	input.Model = r.PathValue("model")
	view, err := h.deps.VideoConfig.Save(r.Context(), input)
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (h handler) startVideo(w http.ResponseWriter, r *http.Request) {
	if h.deps.Video == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video service unavailable")
		return
	}
	projectID, err := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if err != nil || projectID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid project id"})
		return
	}
	bookID, err := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if err != nil || bookID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid book id"})
		return
	}
	defer r.Body.Close()
	var input struct {
		Provider  string `json:"provider"`
		Model     string `json:"model"`
		RequestID string `json:"requestId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
		return
	}
	result, err := h.deps.Video.Start(r.Context(), video.StartRequest{BatchProjectID: projectID, BookID: bookID, Provider: input.Provider, Model: input.Model, RequestID: input.RequestID})
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func (h handler) pollVideoTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.Video == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video service unavailable")
		return
	}
	taskID, ok := parseVideoTaskID(w, r)
	if !ok {
		return
	}
	task, err := h.deps.Video.PollTask(r.Context(), taskID)
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h handler) cancelVideoTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.Video == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video service unavailable")
		return
	}
	taskID, ok := parseVideoTaskID(w, r)
	if !ok {
		return
	}
	task, err := h.deps.Video.CancelTask(r.Context(), taskID)
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h handler) retryVideoTask(w http.ResponseWriter, r *http.Request) {
	if h.deps.Video == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video service unavailable")
		return
	}
	taskID, ok := parseVideoTaskID(w, r)
	if !ok {
		return
	}
	defer r.Body.Close()
	var input struct {
		RequestID string `json:"requestId"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid JSON body"})
			return
		}
	}
	result, err := h.deps.Video.RetryTask(r.Context(), taskID, input.RequestID)
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}

func parseVideoTaskID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	taskID, err := strconv.ParseInt(r.PathValue("taskId"), 10, 64)
	if err != nil || taskID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid task id"})
		return 0, false
	}
	return taskID, true
}

func writeVideoServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, video.ErrProjectArchived) {
		writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ARCHIVED", "message": "项目已归档，请先恢复后再生成视频"})
		return
	}
	if errors.Is(err, video.ErrFinalPromptNotReady) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "final_prompt_not_ready", "message": "completed FINAL_PROMPT is required"})
		return
	}
	var providerErr *video.ProviderError
	if errors.As(err, &providerErr) {
		writeVideoError(w, providerErr.Code, providerErr.Message)
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "video_internal", "message": "video operation failed"})
}

func writeVideoError(w http.ResponseWriter, code video.ErrorCode, message string) {
	status := http.StatusBadGateway
	switch code {
	case video.ErrorProviderUnconfigured, video.ErrorProviderCancelUnsupported, video.ErrorVideoRetryNotAllowed:
		status = http.StatusConflict
	case video.ErrorProviderConfigDecryptFailed:
		status = http.StatusInternalServerError
	case video.ErrorProviderUnavailable:
		status = http.StatusServiceUnavailable
	case video.ErrorProviderAuthFailed, video.ErrorProviderRequestFailed, video.ErrorProviderInvalidResponse:
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]any{"error": code, "message": message})
}
