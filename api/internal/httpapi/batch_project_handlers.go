package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

func (h handler) listBatchProjects(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjects == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_READER_UNAVAILABLE", "message": "批量项目列表暂不可用"})
		return
	}
	if h.deps.Auth != nil && h.deps.BatchProjectAccess == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "AUTH_POLICY_UNAVAILABLE", "message": "项目权限校验暂不可用"})
		return
	}
	filter, err := parseBatchProjectListQuery(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "BATCH_PROJECT_INVALID_REQUEST", "message": err.Error()})
		return
	}
	if h.deps.Auth != nil {
		user, ok := authn.CurrentUser(r.Context())
		if !ok || user.ID <= 0 {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
			return
		}
		filter.UserID, filter.TeamID = user.ID, user.TeamID
		filter.Elevated = strings.EqualFold(user.Role, "admin") || strings.EqualFold(user.Role, "owner")
	} else {
		filter.Elevated = true
	}
	page, err := h.deps.BatchProjects.ListBatchProjects(r.Context(), filter)
	if err != nil {
		h.writeServiceError(w, r, http.StatusInternalServerError, "BATCH_PROJECT_LIST_FAILED", "读取批量项目列表失败", "batch_project", "list", err)
		return
	}
	rows := make([]projectResponse, 0, len(page.Projects))
	for _, project := range page.Projects {
		rows = append(rows, projectResponse{
			ID: project.ID, IntakeID: project.IntakeID, Name: project.Name,
			Sources: project.Sources, BookCount: project.BookCount, Genders: project.Genders, Styles: project.Styles,
			RunStatus: project.RunStatus, FailureCount: project.FailureCount,
			CreatedAt: &project.CreatedAt, UpdatedAt: &project.UpdatedAt, ArchivedAt: project.ArchivedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": rows, "page": page.Page, "limit": page.Limit, "total": page.Total})
}

func parseBatchProjectListQuery(r *http.Request) (intake.BatchProjectListQuery, error) {
	values := r.URL.Query()
	filter := intake.BatchProjectListQuery{
		Query: strings.TrimSpace(values.Get("q")), Source: strings.TrimSpace(values.Get("source")),
		Archived: intake.BatchProjectArchivedActive, Page: 1, Limit: 20, Sort: intake.BatchProjectSortUpdatedDesc,
	}
	if len([]rune(filter.Query)) > 191 || len([]rune(filter.Source)) > 64 {
		return filter, errors.New("搜索条件过长")
	}
	if value := strings.TrimSpace(values.Get("status")); value != "" {
		filter.Status = intake.RunStatus(value)
		switch filter.Status {
		case intake.RunStatusPending, intake.RunStatusQueued, intake.RunStatusScheduled, intake.RunStatusRunning, intake.RunStatusCompleted, intake.RunStatusSucceeded, intake.RunStatusPartialFailed, intake.RunStatusFailed:
		default:
			return filter, errors.New("status 参数无效")
		}
	}
	if value := strings.TrimSpace(values.Get("archived")); value != "" {
		filter.Archived = intake.BatchProjectArchivedFilter(value)
		if filter.Archived != intake.BatchProjectArchivedActive && filter.Archived != intake.BatchProjectArchivedArchived && filter.Archived != intake.BatchProjectArchivedAll {
			return filter, errors.New("archived 参数无效")
		}
	}
	if value := strings.TrimSpace(values.Get("page")); value != "" {
		page, parseErr := strconv.Atoi(value)
		if parseErr != nil || page <= 0 {
			return filter, errors.New("page 参数无效")
		}
		filter.Page = page
	}
	limitValue := strings.TrimSpace(values.Get("limit"))
	if limitValue == "" {
		limitValue = strings.TrimSpace(values.Get("pageSize"))
	}
	if limitValue != "" {
		limit, parseErr := strconv.Atoi(limitValue)
		if parseErr != nil || limit <= 0 || limit > 100 {
			return filter, errors.New("limit 参数必须在 1 到 100 之间")
		}
		filter.Limit = limit
	}
	if value := strings.TrimSpace(values.Get("sort")); value != "" {
		filter.Sort = intake.BatchProjectSort(value)
		if filter.Sort != intake.BatchProjectSortUpdatedDesc && filter.Sort != intake.BatchProjectSortNameAsc {
			return filter, errors.New("sort 参数无效")
		}
	}
	return filter, nil
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
		"project": projectResponse{ID: project.ID, IntakeID: project.IntakeID, Name: project.Name, CreatedAt: &project.CreatedAt, UpdatedAt: &project.UpdatedAt, ArchivedAt: project.ArchivedAt},
		"books":   toBookDetailResponses(books),
	})
}

func (h handler) archiveBatchProject(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjectLifecycle == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目归档暂不可用"})
		return
	}
	projectID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "BATCH_PROJECT_INVALID_REQUEST", "message": "批量项目 ID 无效"})
		return
	}
	actor, ok := authn.CurrentUser(r.Context())
	if !ok || actor.ID <= 0 {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return
	}
	err = h.deps.BatchProjectLifecycle.ArchiveBatchProject(r.Context(), projectID, actor.ID)
	switch {
	case errors.Is(err, intake.ErrBatchProjectActive):
		writeJSON(w, http.StatusConflict, map[string]any{"code": "BATCH_PROJECT_ACTIVE", "message": "项目仍有运行中的任务，暂不能归档"})
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "BATCH_PROJECT_NOT_FOUND", "message": "批量项目不存在"})
	case err != nil:
		h.writeServiceError(w, r, http.StatusInternalServerError, "BATCH_PROJECT_ARCHIVE_FAILED", "归档批量项目失败", "batch_project", "archive", err)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"projectId": projectID, "status": "archived"})
	}
}

func (h handler) restoreBatchProject(w http.ResponseWriter, r *http.Request) {
	if h.deps.BatchProjectLifecycle == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"code": "BATCH_PROJECT_POLICY_UNAVAILABLE", "message": "项目恢复暂不可用"})
		return
	}
	projectID, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": "BATCH_PROJECT_INVALID_REQUEST", "message": "批量项目 ID 无效"})
		return
	}
	err = h.deps.BatchProjectLifecycle.RestoreBatchProject(r.Context(), projectID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "BATCH_PROJECT_NOT_FOUND", "message": "批量项目不存在"})
	case err != nil:
		h.writeServiceError(w, r, http.StatusInternalServerError, "BATCH_PROJECT_RESTORE_FAILED", "恢复批量项目失败", "batch_project", "restore", err)
	default:
		writeJSON(w, http.StatusOK, map[string]any{"projectId": projectID, "status": "active"})
	}
}
