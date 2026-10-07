package httpapi

import (
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novelpanel"
	"net/http"
)

func (h handler) getNovelPanel(w http.ResponseWriter, r *http.Request) {
	if h.deps.NovelPanel == nil {
		writeError(w, 503, "小说面板服务不可用")
		return
	}
	id, e := parsePositiveID(r.PathValue("id"))
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	v, e := h.deps.NovelPanel.GetWorkspace(r.Context(), id)
	if e != nil {
		novelPanelError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"workspace": v})
}
func (h handler) saveNovelPanel(w http.ResponseWriter, r *http.Request) {
	if h.deps.NovelPanel == nil {
		writeError(w, 503, "小说面板服务不可用")
		return
	}
	id, e := parsePositiveID(r.PathValue("id"))
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	var body novelpanel.SaveRequest
	if e = decodeJSON(w, r, &body); e != nil {
		writeError(w, 400, e.Error())
		return
	}
	body.Workspace.ProjectID = id
	v, e := h.deps.NovelPanel.Save(r.Context(), body)
	if e != nil {
		novelPanelError(w, e)
		return
	}
	writeJSON(w, 200, v)
}
func (h handler) listNovelPanelHistory(w http.ResponseWriter, r *http.Request) {
	if h.deps.NovelPanel == nil {
		writeError(w, 503, "小说面板服务不可用")
		return
	}
	id, e := parsePositiveID(r.PathValue("id"))
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	v, e := h.deps.NovelPanel.ListHistory(r.Context(), id, 50)
	if e != nil {
		novelPanelError(w, e)
		return
	}
	writeJSON(w, 200, map[string]any{"items": v})
}
func (h handler) restoreNovelPanelHistory(w http.ResponseWriter, r *http.Request) {
	if h.deps.NovelPanel == nil {
		writeError(w, 503, "小说面板服务不可用")
		return
	}
	id, e := parsePositiveID(r.PathValue("id"))
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	var body struct {
		ExpectedRevision int64 `json:"expectedRevision"`
	}
	if e = decodeJSON(w, r, &body); e != nil {
		writeError(w, 400, e.Error())
		return
	}
	v, e := h.deps.NovelPanel.Restore(r.Context(), novelpanel.RestoreRequest{ProjectID: id, HistoryID: r.PathValue("historyId"), ExpectedRevision: body.ExpectedRevision})
	if e != nil {
		novelPanelError(w, e)
		return
	}
	writeJSON(w, 200, v)
}
func novelPanelError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, novelpanel.ErrNotFound):
		writeError(w, 404, "小说面板记录不存在")
	case errors.Is(e, novelpanel.ErrConflict):
		writeJSON(w, 409, map[string]any{"code": "NOVEL_PANEL_CONFLICT", "message": "小说面板已被其他会话更新，请重新载入"})
	case errors.Is(e, novelpanel.ErrInvalid):
		writeJSON(w, 422, map[string]any{"code": "NOVEL_PANEL_INVALID", "message": e.Error()})
	default:
		writeError(w, 422, "小说面板操作失败")
	}
}
