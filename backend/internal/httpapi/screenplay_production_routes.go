package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
	"time"
)

func RegisterScreenplayProductionRoutes(mux *http.ServeMux, auth BridgeAuth, runtime *screenplay.ProductionRuntime, authoring ...*screenplay.AuthoringService) {
	if mux == nil || runtime == nil {
		return
	}
	register := func(pattern string, fn func(*http.Request, string) (screenplay.ProductionJob, error)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			owner := identity.Username
			if len(authoring) > 0 && authoring[0] != nil {
				perm := "generate"
				if r.Method == http.MethodGet {
					perm = "read"
				}
				resolved, err := authoring[0].Resolve(r.Context(), owner, r.PathValue("id"), perm)
				if err != nil {
					writeScreenplayError(w, err)
					return
				}
				owner = resolved
			}
			job, err := fn(r, owner)
			if err != nil {
				writeScreenplayError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, job)
		})))
	}
	register("POST /api/comic-screenplays/projects/{id}/production/start", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		var input struct {
			Rebuild bool `json:"rebuild"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil && err != io.EOF {
			return screenplay.ProductionJob{}, screenplay.ErrInvalidArgument
		}
		return runtime.StartWithOptions(r.Context(), o, r.PathValue("id"), input.Rebuild)
	})
	register("POST /api/comic-screenplays/projects/{id}/production/schedule", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		var input struct {
			RunAt   time.Time `json:"run_at"`
			Rebuild bool      `json:"rebuild"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			return screenplay.ProductionJob{}, screenplay.ErrInvalidArgument
		}
		return runtime.Schedule(r.Context(), o, r.PathValue("id"), input.RunAt, input.Rebuild)
	})
	register("POST /api/comic-screenplays/projects/{id}/production/cancel-schedule", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		return runtime.CancelSchedule(r.Context(), o, r.PathValue("id"))
	})
	register("GET /api/comic-screenplays/projects/{id}/production/schedule", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		return runtime.Status(r.Context(), o, r.PathValue("id"))
	})

	register("GET /api/comic-screenplays/projects/{id}/production/status", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		return runtime.Status(r.Context(), o, r.PathValue("id"))
	})
	register("POST /api/comic-screenplays/projects/{id}/production/pause", func(r *http.Request, o string) (screenplay.ProductionJob, error) {
		return runtime.Pause(r.Context(), o, r.PathValue("id"))
	})
}
