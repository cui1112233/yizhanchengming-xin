package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
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

func generationHTTPStatus(err error) int {
	switch {
	case errors.Is(err, generation.ErrInvalid): return http.StatusBadRequest
	case errors.Is(err, generation.ErrNotFound): return http.StatusNotFound
	case errors.Is(err, generation.ErrConflict): return http.StatusConflict
	case errors.Is(err, generation.ErrAudioMeasurementRequired): return http.StatusUnprocessableEntity
	case errors.Is(err, generation.ErrAudioProbeUnavailable): return http.StatusServiceUnavailable
	case errors.Is(err, generation.ErrUnavailable): return http.StatusServiceUnavailable
	default: return http.StatusInternalServerError
	}
}

func (h handler) projectGeneration(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil { writeError(w,http.StatusServiceUnavailable,"generation service unavailable"); return }
	projectID,err:=parsePositiveID(r.PathValue("projectId")); if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	if r.Method==http.MethodGet { out,e:=h.deps.Generation.ProjectSummary(r.Context(),projectID); if e!=nil{writeError(w,generationHTTPStatus(e),"读取生成状态失败");return}; writeJSON(w,http.StatusOK,out); return }
	var req generation.RunBatchRequest; if err:=decodeJSON(w,r,&req);err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}; req.BatchProjectID=projectID
	out,e:=h.deps.Generation.RunBatch(r.Context(),req); if e!=nil{writeJSON(w,http.StatusMultiStatus,out);return}; writeJSON(w,http.StatusOK,out)
}

func (h handler) bookGeneration(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil { writeError(w,http.StatusServiceUnavailable,"generation service unavailable"); return }
	projectID,err:=parsePositiveID(r.PathValue("projectId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	bookID,err:=parsePositiveID(r.PathValue("bookId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	if r.Method==http.MethodGet {out,e:=h.deps.Generation.BookSummary(r.Context(),projectID,bookID);if e!=nil{writeError(w,generationHTTPStatus(e),"读取小说生成状态失败");return};writeJSON(w,http.StatusOK,out);return}
	var req generation.RunBookRequest;if err:=decodeJSON(w,r,&req);err!=nil{writeError(w,http.StatusBadRequest,err.Error());return};req.BatchProjectID,req.BookID=projectID,bookID
	out,e:=h.deps.Generation.RunBook(r.Context(),req);if e!=nil{writeJSON(w,generationHTTPStatus(e),out);return};writeJSON(w,http.StatusOK,out)
}

func (h handler) audioMeasurement(w http.ResponseWriter, r *http.Request) {
	if h.deps.Generation == nil { writeError(w,http.StatusServiceUnavailable,"generation service unavailable"); return }
	projectID,err:=parsePositiveID(r.PathValue("projectId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	bookID,err:=parsePositiveID(r.PathValue("bookId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	if r.Method==http.MethodGet {
		out,e:=h.deps.Generation.AudioMeasurement(r.Context(),projectID,bookID)
		if e!=nil { writeJSON(w,generationHTTPStatus(e),map[string]string{"error":stableGenerationError(e)}); return }
		writeJSON(w,http.StatusOK,out); return
	}
	var body struct{ AudioAsset string `json:"audioAsset"` }
	if err:=decodeJSON(w,r,&body);err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	out,e:=h.deps.Generation.MeasureAudio(r.Context(),generation.AudioMeasurementRequest{BatchProjectID:projectID,BookID:bookID,AudioAsset:body.AudioAsset})
	if e!=nil { writeJSON(w,generationHTTPStatus(e),map[string]string{"error":stableGenerationError(e)}); return }
	writeJSON(w,http.StatusOK,out)
}

func stableGenerationError(err error) string {
	switch {
	case errors.Is(err,generation.ErrAudioProbeUnavailable): return "audio_probe_unavailable"
	case errors.Is(err,generation.ErrAudioMeasurementRequired): return "audio_measurement_required"
	case errors.Is(err,generation.ErrNotFound): return "not_found"
	case errors.Is(err,generation.ErrInvalid): return "invalid_request"
	default: return "generation_unavailable"
	}
}

func (h handler) retryGenerationStage(w http.ResponseWriter,r *http.Request){
	if h.deps.Generation==nil{writeError(w,http.StatusServiceUnavailable,"generation service unavailable");return}
	projectID,err:=parsePositiveID(r.PathValue("projectId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return};bookID,err:=parsePositiveID(r.PathValue("bookId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	stage:=generation.Stage(strings.ToUpper(strings.TrimSpace(r.PathValue("stage"))))
	var body struct{RequestID string `json:"requestId"`};if err:=decodeJSON(w,r,&body);err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	out,e:=h.deps.Generation.RetryStage(r.Context(),generation.RetryStageRequest{BatchProjectID:projectID,BookID:bookID,Stage:stage,RequestID:body.RequestID});if e!=nil{writeJSON(w,generationHTTPStatus(e),out);return};writeJSON(w,http.StatusOK,out)
}

func (h handler) generationStage(w http.ResponseWriter,r *http.Request){
	if h.deps.Generation==nil{writeError(w,http.StatusServiceUnavailable,"generation service unavailable");return}
	projectID,err:=parsePositiveID(r.PathValue("projectId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return};bookID,err:=parsePositiveID(r.PathValue("bookId"));if err!=nil{writeError(w,http.StatusBadRequest,err.Error());return}
	stage:=generation.Stage(strings.ToUpper(strings.TrimSpace(r.PathValue("stage"))));out,e:=h.deps.Generation.StageResult(r.Context(),projectID,bookID,stage);if e!=nil{writeError(w,generationHTTPStatus(e),"读取 Stage 结果失败");return};writeJSON(w,http.StatusOK,out)
}

func (h handler) generationPrompts(w http.ResponseWriter,r *http.Request){
	if h.deps.Generation==nil{writeError(w,http.StatusServiceUnavailable,"generation service unavailable");return};out,e:=h.deps.Generation.ListPrompts(r.Context());if e!=nil{writeError(w,generationHTTPStatus(e),"读取 Prompt 列表失败");return};writeJSON(w,http.StatusOK,map[string]any{"prompts":out})
}
