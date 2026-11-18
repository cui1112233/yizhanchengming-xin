package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/novelcreation"
)

func registerNovelShortStoryRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), s *novelcreation.Service, starter NovelCreationRunStarter) {
	handle("POST /api/novel-creation/short-stories", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in novelcreation.ShortStoryInput
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.StartShortStory(r.Context(), owner, in)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, out.Run.ID)
		}
		writeJSON(w, 202, out)
	})
	handle("GET /api/novel-creation/short-stories/requests/{key}", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, e := s.GetShortStoryRequest(r.Context(), owner, r.PathValue("key"))
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})
	handle("GET /api/novel-creation/projects/{projectID}/short-story", func(w http.ResponseWriter, r *http.Request, owner string) {
		out, e := s.GetShortStory(r.Context(), owner, r.PathValue("projectID"))
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"story": out})
	})
	handle("PUT /api/novel-creation/projects/{projectID}/short-story/outline", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in struct {
			Version int                        `json:"version"`
			Outline novelcreation.ShortOutline `json:"outline"`
		}
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.EditShortStoryOutline(r.Context(), owner, r.PathValue("projectID"), in.Version, in.Outline)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"story": out})
	})
	handle("POST /api/novel-creation/runs/{runID}/confirm-short-story", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in struct {
			Version int `json:"version"`
		}
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.ConfirmShortStory(r.Context(), owner, r.PathValue("runID"), in.Version)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		if starter != nil {
			starter(context.Background(), owner, out.ID)
		}
		writeJSON(w, 202, map[string]any{"run": out})
	})
}
