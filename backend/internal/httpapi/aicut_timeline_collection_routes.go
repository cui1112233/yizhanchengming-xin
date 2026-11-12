package httpapi

import (
	"errors"
	"net/http"
	"qiantie/backend/internal/aicut"
	"strings"
	"time"
)

func RegisterAICutTimelineCollectionRoutes(mux *http.ServeMux, auth BridgeAuth, store *aicut.TimelineCollectionStore, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, r, strings.TrimSpace(identity.Username))
		})))
	}
	root := "/api/ai-cut/projects/{id}/timelines"
	item := root + "/{timelineId}"
	handle("GET "+root, func(w http.ResponseWriter, r *http.Request, o string) {
		items, e := store.List(r.Context(), o, r.PathValue("id"))
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"timelines": items})
	})
	metadata := func(w http.ResponseWriter, r *http.Request, o, action string) {
		var input struct {
			Title string `json:"title"`
		}
		if !decodeAICutJSON(w, r, &input) {
			return
		}
		var out aicut.TimelineMetadata
		var e error
		status := http.StatusCreated
		switch action {
		case "create":
			out, e = store.Create(r.Context(), o, r.PathValue("id"), input.Title)
		case "duplicate":
			out, e = store.Duplicate(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"), input.Title)
		default:
			status = http.StatusOK
			if r.PathValue("timelineId") == aicut.DefaultTimelineID {
				if _, e = store.Timeline(r.Context(), o, r.PathValue("id"), aicut.DefaultTimelineID); e == nil {
					e = aicut.ErrDefaultTimelineProtected
				}
			} else {
				out, e = store.Rename(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"), input.Title)
			}
		}
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, status, map[string]any{"timeline": out})
	}
	handle("POST "+root, func(w http.ResponseWriter, r *http.Request, o string) { metadata(w, r, o, "create") })
	handle("PATCH "+item, func(w http.ResponseWriter, r *http.Request, o string) { metadata(w, r, o, "rename") })
	handle("POST "+item+"/duplicate", func(w http.ResponseWriter, r *http.Request, o string) { metadata(w, r, o, "duplicate") })
	handle("DELETE "+item, func(w http.ResponseWriter, r *http.Request, o string) {
		if e := store.Delete(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId")); e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("POST "+root+"/reorder", func(w http.ResponseWriter, r *http.Request, o string) {
		var input struct {
			TimelineIDs []string `json:"timelineIds"`
		}
		if !decodeAICutJSON(w, r, &input) {
			return
		}
		if e := store.Reorder(r.Context(), o, r.PathValue("id"), input.TimelineIDs); e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		items, e := store.List(r.Context(), o, r.PathValue("id"))
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"timelines": items})
	})
	handle("GET "+item+"/snapshot", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := store.Timeline(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"))
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"timeline": v})
	})
	handle("PUT "+item+"/snapshot", func(w http.ResponseWriter, r *http.Request, o string) {
		var input struct {
			ExpectedRevision int64          `json:"expectedRevision"`
			Timeline         aicut.Timeline `json:"timeline"`
		}
		if !decodeAICutJSON(w, r, &input) {
			return
		}
		v, e := store.Save(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"), input.Timeline, input.ExpectedRevision, now().UTC())
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"timeline": v})
	})
	handle("GET "+item+"/history", func(w http.ResponseWriter, r *http.Request, o string) {
		v, e := store.History(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"))
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"history": v})
	})
	move := func(w http.ResponseWriter, r *http.Request, o string, undo bool) {
		var input struct {
			ExpectedRevision int64 `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, r, &input) {
			return
		}
		var v aicut.Timeline
		var h aicut.TimelineHistoryState
		var e error
		if undo {
			v, h, e = store.Undo(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"), input.ExpectedRevision, now().UTC())
		} else {
			v, h, e = store.Redo(r.Context(), o, r.PathValue("id"), r.PathValue("timelineId"), input.ExpectedRevision, now().UTC())
		}
		if e != nil {
			writeAICutTimelineCollectionError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"timeline": v, "history": h})
	}
	handle("POST "+item+"/history/undo", func(w http.ResponseWriter, r *http.Request, o string) { move(w, r, o, true) })
	handle("POST "+item+"/history/redo", func(w http.ResponseWriter, r *http.Request, o string) { move(w, r, o, false) })
}
func writeAICutTimelineCollectionError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, aicut.ErrTimelineNotFound):
		writeJSON(w, 404, map[string]string{"error": "时间线不存在"})
	case errors.Is(e, aicut.ErrTimelineLimit):
		writeJSON(w, 409, map[string]string{"error": "每个工程最多保留50条时间线"})
	case errors.Is(e, aicut.ErrDefaultTimelineProtected):
		writeJSON(w, 400, map[string]string{"error": "默认时间线名称不可修改，也不能删除"})
	case errors.Is(e, aicut.ErrInvalidTimelineCollection):
		writeJSON(w, 400, map[string]string{"error": "时间线名称或排序无效"})
	case errors.Is(e, aicut.ErrTimelineJobUnsupported):
		writeJSON(w, 409, map[string]string{"error": "此任务暂不支持所选时间线"})
	default:
		writeAICutError(w, e)
	}
}
