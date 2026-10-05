package httpapi

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func (h handler) listBatchProjects(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjects == nil {
		writeError(w, http.StatusServiceUnavailable, "batch project reader unavailable")
		return
	}
	projects, err := h.deps.BatchProjects.ListBatchProjects(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取批量项目列表失败")
		return
	}

	var current authn.User
	if h.deps.Auth != nil {
		var ok bool
		current, ok = authn.CurrentUser(r.Context())
		if !ok || current.ID <= 0 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
	}

	rows := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		if h.deps.Auth != nil {
			allowed, accessErr := h.batchProjectAllowed(r.Context(), current, project.ID)
			if accessErr != nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
				return
			}
			if !allowed {
				continue
			}
		}
		rows = append(rows, projectResponse{
			ID:        project.ID,
			IntakeID:  project.IntakeID,
			Name:      project.Name,
			Sources:   project.Sources,
			BookCount: project.BookCount,
			Genders:   project.Genders,
			Styles:    project.Styles,
			RunStatus: project.RunStatus,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": rows})
}

func (h handler) getBatchProject(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjectDetails == nil {
		writeError(w, http.StatusServiceUnavailable, "batch project detail reader unavailable")
		return
	}
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	project, err := h.deps.BatchProjectDetails.GetBatchProject(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "批量项目不存在")
			return
		}
		writeError(w, http.StatusInternalServerError, "读取批量项目失败")
		return
	}
	books, err := h.deps.BatchProjectDetails.ListBooks(r.Context(), project.IntakeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取批量项目小说失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project": projectResponse{ID: project.ID, IntakeID: project.IntakeID, Name: project.Name},
		"books":   toBookResponses(books),
	})
}
