package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
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
	rows := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
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

func (h handler) saveScriptOriginalText(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjectDetails == nil || h.deps.ScriptBooks == nil {
		writeError(w, http.StatusServiceUnavailable, "script text service unavailable")
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
	var body struct {
		OriginalText string `json:"originalText"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	body.OriginalText = strings.TrimSpace(body.OriginalText)
	if body.OriginalText == "" || len([]rune(body.OriginalText)) > 2_000_000 {
		writeError(w, http.StatusBadRequest, "原文长度必须在 1 到 2000000 字符之间")
		return
	}
	project, err := h.deps.BatchProjectDetails.GetBatchProject(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusNotFound, "批量项目不存在")
		return
	}
	book, err := h.deps.ScriptBooks.UpdateBookOriginalText(r.Context(), project.IntakeID, bookID, body.OriginalText)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "小说不属于当前项目")
			return
		}
		writeError(w, http.StatusInternalServerError, "保存原文失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"book": toBookDetailResponses([]intake.Book{book})[0]})
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
		"books":   toBookDetailResponses(books),
	})
}
