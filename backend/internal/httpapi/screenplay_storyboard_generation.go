package httpapi

import (
	"net/http"
	"qiantie/backend/internal/screenplay"
	"sort"
	"strconv"
)

func RegisterScreenplayStoryboardGenerationRoutes(mux *http.ServeMux, auth BridgeAuth, projects screenplay.Store, authors *screenplay.AuthoringService) {
	store, ok := projects.(screenplay.StoryboardGenerationStore)
	if !ok || mux == nil {
		return
	}
	for _, pattern := range []string{"GET /api/comic-screenplays/projects/{id}/storyboard-generation-attempts", "GET /api/comic-screenplays/projects/{id}/storyboard-generation-attempts/{key}"} {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			owner := actor.Username
			var e error
			if authors != nil {
				owner, e = authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "read")
			} else {
				_, e = projects.GetProject(r.Context(), owner, r.PathValue("id"))
			}
			if e != nil {
				writeScreenplayError(w, e)
				return
			}
			state, e := store.LoadStoryboardGenerations(r.Context(), owner, r.PathValue("id"))
			if e != nil {
				writeScreenplayError(w, e)
				return
			}
			if key := r.PathValue("key"); key != "" {
				a := state.Attempts[key]
				if a == nil {
					writeScreenplayError(w, screenplay.ErrNotFound)
					return
				}
				writeJSON(w, 200, map[string]any{"attempt": a.Public()})
				return
			}
			attempts := []screenplay.StoryboardGenerationAttempt{}
			episode, _ := strconv.Atoi(r.URL.Query().Get("episode"))
			engine := r.URL.Query().Get("engine")
			for _, a := range state.Attempts {
				if engine != "" && string(a.Engine) != engine {
					continue
				}
				if episode > 0 {
					found := false
					for _, n := range a.EpisodeNumbers {
						found = found || n == episode
					}
					if !found {
						continue
					}
				}
				attempts = append(attempts, *a)
			}
			sort.SliceStable(attempts, func(i, j int) bool { return attempts[i].CreatedAt.After(attempts[j].CreatedAt) })
			if len(attempts) > 50 {
				attempts = attempts[:50]
			}
			out := []map[string]any{}
			for _, a := range attempts {
				out = append(out, a.Public())
			}
			writeJSON(w, 200, map[string]any{"attempts": out})
		})))
	}
}
