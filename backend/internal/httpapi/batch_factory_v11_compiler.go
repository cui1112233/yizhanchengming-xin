package httpapi

import (
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
)

func registerCompilerRoutes(mux *http.ServeMux, service *batchfactoryv11.PromptCompilerService) {
	registerCompilerRoutesForPrefix(mux, service, "v11")
}

// registerV12CompilerRoutes exposes the same read-only compiler under the
// public workbench namespace. V12 owns the browser contract, while the data
// and compilation rules remain shared with V11.
func registerV12CompilerRoutes(mux *http.ServeMux, service *batchfactoryv11.PromptCompilerService) {
	registerCompilerRoutesForPrefix(mux, service, "v12")
}

func registerCompilerRoutesForPrefix(mux *http.ServeMux, service *batchfactoryv11.PromptCompilerService, version string) {
	mux.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/books/{bookId}/videos/{videoId}/effective-settings", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		value, _, _, _, err := service.ResolveEffective(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"effectiveSettings": value})
	})
	mux.HandleFunc("GET /api/batch-factory/"+version+"/batches/{batchId}/books/{bookId}/videos/{videoId}/final-prompt", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		value, err := service.Compile(r.Context(), owner, r.PathValue("batchId"), r.PathValue("bookId"), r.PathValue("videoId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"finalPrompt": value})
	})
}
