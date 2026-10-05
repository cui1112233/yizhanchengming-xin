package httpapi

import "net/http"

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
	rows := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		rows = append(rows, projectResponse{ID: project.ID, IntakeID: project.IntakeID, Name: project.Name})
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": rows})
}
