package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"qiantie/backend/internal/screenplay"
)

func RegisterScreenplayStoryboardPreferenceRoutes(mux *http.ServeMux, auth BridgeAuth, projects screenplay.Store, authors *screenplay.AuthoringService) {
	store, ok := projects.(screenplay.StoryboardPreferencesStore)
	if !ok || mux == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, r, actor.Username)
		})))
	}
	handle("GET /api/screenplays/storyboard-preferences", func(w http.ResponseWriter, r *http.Request, actor string) {
		v, e := store.LoadStoryboardPreferences(r.Context(), actor)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"state": v, "capabilities": map[string]bool{"manage": true}})
	})
	handle("GET /api/screenplays/{id}/storyboard-preferences", func(w http.ResponseWriter, r *http.Request, actor string) {
		owner := actor
		var e error
		if authors != nil {
			owner, e = authors.Resolve(r.Context(), actor, r.PathValue("id"), "read")
		} else {
			_, e = projects.GetProject(r.Context(), actor, r.PathValue("id"))
		}
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		v, e := store.LoadStoryboardPreferences(r.Context(), owner)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		if owner != actor {
			only := []screenplay.StoryboardPrefixPreset{}
			for _, p := range v.PrefixPresets {
				if p.ID == v.SelectedPrefixID {
					only = append(only, p)
				}
			}
			v.PrefixPresets = only
			v.FavoriteIDs = []string{}
			v.RecentIDs = []string{}
		}
		writeJSON(w, 200, map[string]any{"state": v, "capabilities": map[string]bool{"manage": owner == actor}})
	})
	handle("POST /api/screenplays/storyboard-preferences/actions", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in screenplay.StoryboardPreferencesAction
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
		if e := decoder.Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "前缀导入或偏好数据格式无效、或超过8MB"})
			return
		}
		if e := decoder.Decode(new(any)); e != io.EOF {
			writeJSON(w, 400, map[string]string{"error": "请求只能包含一个JSON对象"})
			return
		}
		imported, skipped := 0, 0
		v, e := store.MutateStoryboardPreferences(r.Context(), actor, func(v *screenplay.StoryboardPreferences) error {
			var e error
			imported, skipped, e = screenplay.ApplyStoryboardPreferencesAction(v, in)
			return e
		})
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"state": v, "imported_count": imported, "skipped_count": skipped})
	})
}
