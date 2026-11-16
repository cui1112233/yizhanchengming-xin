package httpapi

import (
	"context"
	"errors"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
	"time"
)

func registerScreenplayPlanningProgressRoute(mux *http.ServeMux, auth BridgeAuth, authors *screenplay.AuthoringService) {
	mux.Handle("GET /api/comic-screenplays/projects/{id}/storyboard-plan-progress", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		owner, e := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "read")
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		key, chapterID := r.URL.Query().Get("request_key"), r.URL.Query().Get("chapter_id")
		if strings.TrimSpace(key) == "" || len(key) > 256 || chapterID == "" {
			writeJSON(w, 400, map[string]string{"error": "请选择有效的规划请求与章节"})
			return
		}
		state, e := authors.Load(r.Context(), owner, r.PathValue("id"))
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		var chapter *screenplay.Chapter
		for i := range state.Chapters {
			if state.Chapters[i].ID == chapterID {
				chapter = &state.Chapters[i]
				break
			}
		}
		if chapter == nil {
			writeScreenplayError(w, screenplay.ErrNotFound)
			return
		}
		if chapter.PlanningRequestKey != key {
			writeJSON(w, 409, map[string]string{"error": "规划请求已变化，请刷新后核对当前请求"})
			return
		}
		out := map[string]any{"chapter_id": chapterID, "request_key": key, "status": chapter.PlanningStatus, "progress": nil, "available": authors.PlanningProgress != nil, "revision": state.Revision}
		if chapter.PlanningStatus != "running" {
			out["chapter"] = chapter
		}
		if authors.PlanningProgress == nil {
			writeJSON(w, 200, out)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		progress, e := authors.PlanningProgress.Get(ctx, owner, r.PathValue("id"), key)
		if errors.Is(e, screenplay.ErrNotFound) {
			writeJSON(w, 200, out)
			return
		}
		if e != nil {
			if chapter.PlanningStatus != "running" {
				out["available"] = false
				writeJSON(w, 200, out)
			} else {
				writeJSON(w, 503, map[string]any{"error": "暂时无法读取规划进度，生成请求不会重复提交", "available": false, "status": chapter.PlanningStatus})
			}
			return
		}
		if progress.ChapterID != chapterID {
			writeJSON(w, 409, map[string]string{"error": "规划进度与当前章节不一致"})
			return
		}
		out["progress"] = progress
		writeJSON(w, 200, out)
	})))
}
