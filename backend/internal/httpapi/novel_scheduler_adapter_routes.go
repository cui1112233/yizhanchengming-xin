package httpapi

import (
	"net/http"
	"qiantie/backend/internal/novelcreation"
)

func registerNovelSchedulerAdapterRoutes(handle func(string, func(http.ResponseWriter, *http.Request, string)), s *novelcreation.Service) {
	handle("POST /api/novel-creation/projects/{projectID}/scheduler/start", func(w http.ResponseWriter, r *http.Request, owner string) {
		var in novelcreation.SchedulerStartInput
		if !decodeLongImitation(w, r, &in) {
			return
		}
		out, e := s.StartScheduler(r.Context(), owner, r.PathValue("projectID"), in)
		if e != nil {
			writeNovelCreationError(w, e)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"schedule": out})
	})
}
