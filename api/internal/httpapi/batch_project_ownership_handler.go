package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

// createOwnedBatchProject preserves the existing pipeline creation behavior and,
// when authentication is enabled, pins the newly created project to the current
// user/team before returning it to the browser. Existing pre-Task15 projects
// without an ownership row remain admin/owner-only for publishing.
func (h handler) createOwnedBatchProject(w http.ResponseWriter, r *http.Request) {
	if h.deps.Pipeline == nil {
		writeError(w, http.StatusServiceUnavailable, "pipeline service unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var request batchProjectRequest
	if r.ContentLength != 0 {
		if err := decodeJSON(w, r, &request); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	create := pipeline.CreateRequest{IntakeID: id, Name: request.Name}
	if raw := strings.TrimSpace(request.RunAt); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "runAt 必须是 RFC3339 时间")
			return
		}
		create.RunAt = parsed
	}
	result, err := h.deps.Pipeline.Create(r.Context(), create)
	if err != nil {
		h.writeServiceError(w, r, http.StatusUnprocessableEntity, "BATCH_PROJECT_CREATE_FAILED", "创建批量项目失败", "batch_project", "create", err)
		return
	}

	if h.deps.Publishing != nil {
		if actor, ok := authn.CurrentUser(r.Context()); ok {
			if err := h.deps.Publishing.ClaimBatchProject(r.Context(), actor, result.Project.ID); err != nil {
				writePublishingError(w, err)
				return
			}
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"project": projectResponse{ID: result.Project.ID, IntakeID: result.Project.IntakeID, Name: result.Project.Name},
		"run": runResponse{ID: result.Run.ID, BatchProjectID: result.Run.BatchProjectID, RunAt: result.Run.RunAt, Status: result.Run.Status},
	})
}
