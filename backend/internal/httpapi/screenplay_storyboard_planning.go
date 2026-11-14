package httpapi

import (
	"context"
	"errors"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

func RegisterScreenplayStoryboardPlanRoutes(mux *http.ServeMux, auth BridgeAuth, authors *screenplay.AuthoringService, resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)) {
	if mux == nil || authors == nil {
		return
	}
	mux.Handle("POST /api/comic-screenplays/projects/{id}/storyboard-assistant-draft", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		var in screenplay.StoryboardAssistantDraftRequest
		if !decodeCreativeRequest(w, r, &in) {
			return
		}
		text, e := authors.StoryboardAssistantDraft(r.Context(), owner, r.PathValue("id"), in)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"requestText": text, "chapter_id": in.ChapterID, "task": in.Task, "plan_mode": in.PlanMode})
	})))
	mux.Handle("POST /api/comic-screenplays/projects/{id}/storyboard-plan", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		owner, e := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "generate")
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		var in screenplay.StoryboardPlanRequest
		if !decodeCreativeRequest(w, r, &in) {
			return
		}
		out, e := authors.GenerateStoryboardPlan(r.Context(), owner, r.PathValue("id"), in, screenplayTextGenerator{owner: owner, resolver: resolver})
		if e != nil {
			if errors.Is(e, context.Canceled) {
				writeJSON(w, 499, map[string]any{"error": "已取消规划，已有规划保留", "cancelled": true})
				return
			}
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})))
}
