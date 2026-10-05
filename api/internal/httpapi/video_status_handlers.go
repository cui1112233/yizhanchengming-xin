package httpapi

import (
	"net/http"
	"strconv"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

func (h handler) projectVideoStatus(w http.ResponseWriter, r *http.Request) {
	if h.deps.VideoStatus == nil {
		writeVideoError(w, video.ErrorProviderUnavailable, "video status service unavailable")
		return
	}
	projectID, err := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if err != nil || projectID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_request", "message": "invalid project id"})
		return
	}
	status, err := h.deps.VideoStatus.ProjectStatus(r.Context(), projectID)
	if err != nil {
		writeVideoServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}
