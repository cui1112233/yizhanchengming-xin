package httpapi

import (
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"net/http"
	"strings"
)

func (h handler) scriptStoryboard(w http.ResponseWriter, r *http.Request) {
	if h.deps.ScriptStoryboards == nil {
		writeError(w, http.StatusServiceUnavailable, "script storyboard service unavailable")
		return
	}
	p, e := parsePositiveID(r.PathValue("projectId"))
	if e != nil {
		writeError(w, http.StatusBadRequest, e.Error())
		return
	}
	b, e := parsePositiveID(r.PathValue("bookId"))
	if e != nil {
		writeError(w, http.StatusBadRequest, e.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		out, e := h.deps.ScriptStoryboards.Storyboard(r.Context(), p, b)
		if e != nil {
			writeError(w, generationHTTPStatus(e), "读取分镜失败")
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPost:
		if strings.HasSuffix(r.URL.Path, "/cards") {
			var in generation.SaveStoryboardCardRequest
			if e = decodeJSON(w, r, &in); e != nil {
				writeError(w, http.StatusBadRequest, e.Error())
				return
			}
			out, e := h.deps.ScriptStoryboards.SaveStoryboardCard(r.Context(), p, b, in)
			if e != nil {
				writeError(w, generationHTTPStatus(e), "保存分镜失败")
				return
			}
			writeJSON(w, http.StatusOK, out)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/recompile") {
			writeJSON(w, http.StatusConflict, map[string]any{
				"code":    "STORYBOARD_RECOMPILE_ASYNC_REQUIRED",
				"message": "运行时生成记录必须通过异步任务重新编译",
			})
			return
		}
		var in struct {
			CardIDs         []int64 `json:"cardIds"`
			ExpectedVersion int     `json:"expectedVersion"`
		}
		if e = decodeJSON(w, r, &in); e != nil {
			writeError(w, http.StatusBadRequest, e.Error())
			return
		}
		out, e := h.deps.ScriptStoryboards.ReorderStoryboard(r.Context(), p, b, in.CardIDs, in.ExpectedVersion)
		if e != nil {
			writeError(w, generationHTTPStatus(e), "排序分镜失败")
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPut:
		id, e := parsePositiveID(r.PathValue("cardId"))
		if e != nil {
			writeError(w, http.StatusBadRequest, e.Error())
			return
		}
		var in generation.SaveStoryboardCardRequest
		if e = decodeJSON(w, r, &in); e != nil {
			writeError(w, http.StatusBadRequest, e.Error())
			return
		}
		in.Card.ID = id
		out, e := h.deps.ScriptStoryboards.SaveStoryboardCard(r.Context(), p, b, in)
		if e != nil {
			writeError(w, generationHTTPStatus(e), "保存分镜失败")
			return
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodDelete:
		id, e := parsePositiveID(r.PathValue("cardId"))
		if e != nil {
			writeError(w, http.StatusBadRequest, e.Error())
			return
		}
		var in struct {
			ExpectedVersion int `json:"expectedVersion"`
		}
		if e = decodeJSON(w, r, &in); e != nil {
			writeError(w, http.StatusBadRequest, e.Error())
			return
		}
		out, e := h.deps.ScriptStoryboards.DeleteStoryboardCard(r.Context(), p, b, id, in.ExpectedVersion)
		if e != nil {
			writeError(w, generationHTTPStatus(e), "删除分镜失败")
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}
