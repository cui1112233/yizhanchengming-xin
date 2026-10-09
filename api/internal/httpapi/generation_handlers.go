package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

type GenerationService interface {
	ProjectSummary(context.Context, int64) (generation.ProjectSummary, error)
	BookSummary(context.Context, int64, int64) (generation.BookGenerationResult, error)
	RunBook(context.Context, generation.RunBookRequest) (generation.BookGenerationResult, error)
	RunBatch(context.Context, generation.RunBatchRequest) (generation.BatchGenerationResult, error)
	RetryStage(context.Context, generation.RetryStageRequest) (generation.BookGenerationResult, error)
	StageResult(context.Context, int64, int64, generation.Stage) (generation.StageRun, error)
	ListPrompts(context.Context) ([]generation.Prompt, error)
	AudioMeasurement(context.Context, int64, int64) (generation.AudioMeasurement, error)
	MeasureAudio(context.Context, generation.AudioMeasurementRequest) (generation.AudioMeasurement, error)
}

type AdminPromptService interface {
	ListAdminPrompts(context.Context, string) ([]generation.AdminPrompt, error)
	GetAdminPrompt(context.Context, string, int) (generation.AdminPrompt, error)
	CreateAdminPromptDraft(context.Context, int64, string, string) error
	UpdateAdminPromptDraft(context.Context, int64, string, int, string) error
	PublishAdminPrompt(context.Context, int64, string, int) error
	RestoreAdminPrompt(context.Context, int64, string, int) error
}

func generationHTTPStatus(err error) int {
	switch {
	case errors.Is(err, generation.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, generation.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, generation.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, generation.ErrProjectArchived):
		return http.StatusConflict
	case errors.Is(err, generation.ErrAudioMeasurementRequired):
		return http.StatusUnprocessableEntity
	case errors.Is(err, generation.ErrAudioProbeUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, generation.ErrUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func (h handler) projectGeneration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && h.deps.Generation == nil {
		writeError(w, http.StatusServiceUnavailable, "generation service unavailable")
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if r.Method == http.MethodGet {
		out, e := h.deps.Generation.ProjectSummary(r.Context(), projectID)
		if e != nil {
			outcome := generation.OutcomeForError(e)
			h.writeServiceError(w, r, generationHTTPStatus(e), outcome.Code, outcome.Message, "generation", "project_summary", e)
			return
		}
		writeJSON(w, http.StatusOK, projectGenerationSummary(out))
		return
	}
	var body generationAdmissionBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.admitGeneration(w, r, task9runtime.GenerationRequest{BatchProjectID: projectID, BookIDs: body.BookIDs, HookEnabled: body.HookEnabled, PlotMode: body.PlotMode, DirectorMode: body.DirectorMode, MatchAudio: body.MatchAudio, ShotDurationLimitSec: body.ShotDurationLimitSec, RequestID: idempotencyKey(r, body.RequestID)})
}

func (h handler) bookGeneration(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && h.deps.Generation == nil {
		writeError(w, http.StatusServiceUnavailable, "generation service unavailable")
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookID, err := parsePositiveID(r.PathValue("bookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if r.Method == http.MethodGet {
		out, e := h.deps.Generation.BookSummary(r.Context(), projectID, bookID)
		if e != nil {
			outcome := generation.OutcomeForError(e)
			h.writeServiceError(w, r, generationHTTPStatus(e), outcome.Code, outcome.Message, "generation", "book_summary", e)
			return
		}
		writeJSON(w, http.StatusOK, projectBookGeneration(out))
		return
	}
	var body generationAdmissionBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.admitGeneration(w, r, task9runtime.GenerationRequest{BatchProjectID: projectID, BookIDs: []int64{bookID}, HookEnabled: body.HookEnabled, PlotMode: body.PlotMode, DirectorMode: body.DirectorMode, MatchAudio: body.MatchAudio, ShotDurationLimitSec: body.ShotDurationLimitSec, RequestID: idempotencyKey(r, body.RequestID)})
}

func (h handler) audioMeasurement(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil {
		writeError(w, http.StatusServiceUnavailable, "generation service unavailable")
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookID, err := parsePositiveID(r.PathValue("bookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if r.Method == http.MethodGet {
		out, e := h.deps.Generation.AudioMeasurement(r.Context(), projectID, bookID)
		if e != nil {
			writeJSON(w, generationHTTPStatus(e), map[string]string{"error": stableGenerationError(e)})
			return
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	var body struct {
		AudioAsset       string   `json:"audioAsset"`
		AudioDurationSec *float64 `json:"audioDurationSec,omitempty"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// AudioDurationSec is compatibility/display input only. It is deliberately
	// ignored here; ffprobe remains the sole duration fact source.
	out, e := h.deps.Generation.MeasureAudio(r.Context(), generation.AudioMeasurementRequest{BatchProjectID: projectID, BookID: bookID, AudioAsset: body.AudioAsset})
	if e != nil {
		writeJSON(w, generationHTTPStatus(e), map[string]string{"error": stableGenerationError(e)})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func stableGenerationError(err error) string {
	switch {
	case errors.Is(err, generation.ErrAudioProbeUnavailable):
		return "audio_probe_unavailable"
	case errors.Is(err, generation.ErrAudioMeasurementRequired):
		return "audio_measurement_required"
	case errors.Is(err, generation.ErrNotFound):
		return "not_found"
	case errors.Is(err, generation.ErrInvalid):
		return "invalid_request"
	default:
		return "generation_unavailable"
	}
}

func (h handler) retryGenerationStage(w http.ResponseWriter, r *http.Request) {
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookID, err := parsePositiveID(r.PathValue("bookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	stage := generation.Stage(strings.ToUpper(strings.TrimSpace(r.PathValue("stage"))))
	var body struct {
		RequestID       string `json:"requestId"`
		SourceBookRunID int64  `json:"sourceBookRunId"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.admitGeneration(w, r, task9runtime.GenerationRequest{BatchProjectID: projectID, BookIDs: []int64{bookID}, RequestID: idempotencyKey(r, body.RequestID), Action: task9runtime.GenerationActionStageRetry, RetryStage: string(stage), SourceBookRunID: body.SourceBookRunID})
}

type generationAdmissionBody struct {
	BookIDs              []int64 `json:"bookIds,omitempty"`
	HookEnabled          bool    `json:"hookEnabled"`
	PlotMode             bool    `json:"plotMode"`
	DirectorMode         string  `json:"directorMode"`
	MatchAudio           bool    `json:"matchAudio"`
	ShotDurationLimitSec int64   `json:"shotDurationLimitSec"`
	RequestID            string  `json:"requestId"`
}

func idempotencyKey(r *http.Request, compatibility string) string {
	if key := strings.TrimSpace(r.Header.Get("Idempotency-Key")); key != "" {
		return key
	}
	return compatibility
}

func (h handler) admitGeneration(w http.ResponseWriter, r *http.Request, req task9runtime.GenerationRequest) {
	if h.deps.GenerationRuntime == nil || !h.deps.GenerationRuntime.Ready() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "GENERATION_RUNTIME_UNAVAILABLE", "message": "生成执行器尚未可用"})
		return
	}
	actor, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	req.RequestedByUserID = actor.ID
	result, err := h.deps.GenerationRuntime.AdmitGeneration(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, task9runtime.ErrInvalidGenerationRequest):
			writeJSON(w, http.StatusBadRequest, map[string]any{"code": "GENERATION_INVALID", "message": "生成请求无效"})
		case errors.Is(err, task9runtime.ErrIdempotencyConflict):
			writeJSON(w, http.StatusConflict, map[string]any{"code": "GENERATION_IDEMPOTENCY_CONFLICT", "message": "幂等键已用于不同请求"})
		case errors.Is(err, task9runtime.ErrProjectArchived):
			writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ARCHIVED", "message": "项目已归档，请先恢复"})
		case errors.Is(err, task9runtime.ErrExecutorUnavailable):
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "GENERATION_RUNTIME_UNAVAILABLE", "message": "生成执行器尚未可用"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"code": "GENERATION_ADMISSION_FAILED", "message": "生成任务保存失败"})
		}
		return
	}
	pollURL := "/api/v1/batch-projects/" + r.PathValue("projectId") + "/generation/runs/" + strconv.FormatInt(result.Run.ID, 10)
	w.Header().Set("Location", pollURL)
	w.Header().Set("Retry-After", "1")
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusAccepted, map[string]any{"created": result.Created, "status": publicAdmissionStatus(result.Run.Status), "runId": result.Run.ID, "taskIds": result.TaskIDs, "pollUrl": pollURL, "dispatch": result.Dispatch})
}

func publicAdmissionStatus(status task9runtime.RunState) string {
	switch status {
	case task9runtime.RunSucceeded:
		return "completed"
	case task9runtime.RunPartialFailed:
		return "partial_failed"
	case task9runtime.RunFailed:
		return "failed"
	case task9runtime.RunPending:
		return "queued"
	default:
		return "running"
	}
}

func (h handler) generationRun(w http.ResponseWriter, r *http.Request) {
	if h.deps.GenerationRuntime == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "RUNTIME_UNAVAILABLE", "message": "生成状态暂不可用"})
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	runID, err := parsePositiveID(r.PathValue("runId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.deps.GenerationRuntime.GenerationRun(r.Context(), projectID, runID)
	if errors.Is(err, task9runtime.ErrRunNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "GENERATION_RUN_NOT_FOUND", "message": "生成任务不存在"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "RUNTIME_UNAVAILABLE", "message": "生成状态暂不可用"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, out)
}

func (h handler) generationStage(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil {
		writeError(w, http.StatusServiceUnavailable, "generation service unavailable")
		return
	}
	projectID, err := parsePositiveID(r.PathValue("projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	bookID, err := parsePositiveID(r.PathValue("bookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	stage := generation.Stage(strings.ToUpper(strings.TrimSpace(r.PathValue("stage"))))
	out, e := h.deps.Generation.StageResult(r.Context(), projectID, bookID, stage)
	if e != nil {
		outcome := generation.OutcomeForError(e)
		h.writeServiceError(w, r, generationHTTPStatus(e), outcome.Code, outcome.Message, "generation", "stage_result", e)
		return
	}
	writeJSON(w, http.StatusOK, projectStageRun(out))
}

func (h handler) generationPrompts(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil {
		writeError(w, http.StatusServiceUnavailable, "generation service unavailable")
		return
	}
	out, e := h.deps.Generation.ListPrompts(r.Context())
	if e != nil {
		writeError(w, generationHTTPStatus(e), "读取 Prompt 列表失败")
		return
	}
	safe := append([]generation.Prompt(nil), out...)
	for i := range safe {
		safe[i].Content = ""
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompts": safe})
}
