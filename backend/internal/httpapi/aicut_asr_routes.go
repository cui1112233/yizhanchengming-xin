package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/aicut"
)

func RegisterAICutASRRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, service *aicut.ASRService) {
	if mux == nil || store == nil || service == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	handle("POST /api/ai-cut/projects/{id}/asr-jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			AssetID          string `json:"assetId"`
			ModelID          string `json:"modelId"`
			Language         string `json:"language"`
			ExpectedRevision int64  `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		job, err := service.Submit(req.Context(), owner, aicut.ASRSubmit{ProjectID: req.PathValue("id"), AssetID: input.AssetID, ModelID: input.ModelID, Language: input.Language, ExpectedRevision: input.ExpectedRevision})
		if err != nil {
			if errors.Is(err, aicut.ErrProjectNotFound) || errors.Is(err, aicut.ErrAssetNotFound) {
				writeAICutError(w, err)
				return
			}
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": job})
	})
	handle("GET /api/ai-cut/projects/{id}/asr-jobs", func(w http.ResponseWriter, req *http.Request, owner string) {
		jobs, err := store.ListASRJobs(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
	})
	handle("GET /api/ai-cut/projects/{id}/asr-jobs/{jobID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := store.ASRJob(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	})
	handle("POST /api/ai-cut/projects/{id}/asr-jobs/{jobID}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		job, err := service.Cancel(req.Context(), owner, req.PathValue("id"), req.PathValue("jobID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	})
}
