package httpapi

import (
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

// registerV12IntakeRoutes keeps the public novel-fetch handoff on the same
// durable Go batch store as the V11 workbench routes. The browser uses V12 for
// the handoff, while the production workbench continues reading the intake
// through its V11-compatible contract.
func registerV12IntakeRoutes(mux *http.ServeMux, store batchfactoryv11.Store) {
	mux.HandleFunc("POST /api/batch-factory/v12/intakes/novel-fetch", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input batchfactoryv11.NovelFetchIntakeInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if len(input.Books) == 0 {
			writeStoreError(w, batchfactoryv11.ErrInvalid)
			return
		}
		intake, err := store.CreateIntake(r.Context(), owner, input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"intake":     intake,
			"redirectTo": "/shuihuo-production?intake=" + intake.ID,
		})
	})
	mux.HandleFunc("GET /api/batch-factory/v12/intakes/{intakeId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		intake, err := store.GetIntake(r.Context(), owner, r.PathValue("intakeId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"intake": intake})
	})
}
