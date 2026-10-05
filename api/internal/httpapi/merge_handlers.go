package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

func (h handler) startVideoMerge(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoMerge == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error":"merge_unavailable","message":"merge service unavailable"})
		return
	}
	projectID, err := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if err != nil || projectID <= 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid_request","message":"invalid project id"}); return }
	bookID, err := strconv.ParseInt(r.PathValue("bookId"), 10, 64)
	if err != nil || bookID <= 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid_request","message":"invalid book id"}); return }
	defer r.Body.Close()
	var input struct { ProductionTaskIDs []int64 `json:"productionTaskIds"`; AspectRatio string `json:"aspectRatio"`; Speed float64 `json:"speed"` }
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody)).Decode(&input); err != nil { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid_request","message":"invalid JSON body"}); return }
	result, err := h.deps.VideoMerge.StartFromProductionTasks(r.Context(), video.MergeProductionStartRequest{BatchProjectID:projectID, BookID:bookID, ProductionTaskIDs:input.ProductionTaskIDs, AspectRatio:input.AspectRatio, Speed:input.Speed})
	if err != nil { writeMergeError(w, err); return }
	writeJSON(w, http.StatusAccepted, result)
}

func (h handler) getVideoMerge(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoMerge == nil { writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error":"merge_unavailable","message":"merge service unavailable"}); return }
	jobID, err := strconv.ParseInt(r.PathValue("jobId"), 10, 64)
	if err != nil || jobID <= 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid_request","message":"invalid merge job id"}); return }
	job, attempts, err := h.deps.VideoMerge.Get(r.Context(), jobID)
	if err != nil { writeMergeError(w, err); return }
	writeJSON(w, http.StatusOK, map[string]any{"job":job,"attempts":attempts})
}

func (h handler) retryVideoMerge(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoMerge == nil { writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error":"merge_unavailable","message":"merge service unavailable"}); return }
	attemptID, err := strconv.ParseInt(r.PathValue("attemptId"), 10, 64)
	if err != nil || attemptID <= 0 { writeJSON(w, http.StatusBadRequest, map[string]any{"error":"invalid_request","message":"invalid merge attempt id"}); return }
	result, err := h.deps.VideoMerge.RetryAttempt(r.Context(), attemptID)
	if err != nil { writeMergeError(w, err); return }
	writeJSON(w, http.StatusAccepted, result)
}

func writeMergeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, video.ErrMergeInputNotReady):
		writeJSON(w, http.StatusConflict, map[string]any{"error":"merge_input_not_ready","message":"all merge inputs must be succeeded VIDEO tasks from this project/book"})
	case errors.Is(err, video.ErrMergeRetryNotAllowed):
		writeJSON(w, http.StatusConflict, map[string]any{"error":"merge_retry_not_allowed","message":"only the latest failed merge attempt can be retried"})
	case errors.Is(err, video.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error":"merge_not_found","message":"merge job or attempt not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error":"merge_internal","message":"merge operation failed"})
	}
}
