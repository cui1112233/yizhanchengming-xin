package httpapi

import (
	"net/http"
	"strings"
)

// registerV12V11Aliases keeps the V12 browser contract on the Go store for
// operations whose V11 implementation already has the same owner-scoped
// semantics. The remaining V12 adapters stay on the compatibility bridge.
func registerV12V11Aliases(v12, v11 *http.ServeMux) {
	if v12 == nil || v11 == nil {
		return
	}
	for _, pattern := range []string{
		"GET /api/batch-factory/v12/batches/{batchId}",
		"GET /api/batch-factory/v12/batches/{batchId}/runtime-index",
		"GET /api/batch-factory/v12/batches/{batchId}/status",
		"GET /api/batch-factory/v12/batches/{batchId}/merge-status",
		"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages",
		"POST /api/batch-factory/v12/intakes/{intakeId}/batches",
		"PUT /api/batch-factory/v12/batches/{batchId}/settings",
		"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/metadata",
		"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/source",
		"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/override",
		"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/override",
		"POST /api/batch-factory/v12/batches/{batchId}/change-impact",
		"GET /api/batch-factory/v12/config-versions",
		"POST /api/batch-factory/v12/config-versions",
		"PUT /api/batch-factory/v12/config-versions/{versionId}",
		"GET /api/batch-factory/v12/prompts",
		"POST /api/batch-factory/v12/prompts",
		"GET /api/batch-factory/v12/drafts",
		"PUT /api/batch-factory/v12/drafts",
		"DELETE /api/batch-factory/v12/batches/{batchId}",
		"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}",
	} {
		v12.Handle(pattern, v12AliasHandler(v11))
	}
}

func v12AliasHandler(v11 *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		clone := req.Clone(req.Context())
		urlCopy := *req.URL
		urlCopy.Path = strings.Replace(urlCopy.Path, "/api/batch-factory/v12/", "/api/batch-factory/v11/", 1)
		urlCopy.RawPath = ""
		clone.URL = &urlCopy
		v11.ServeHTTP(w, clone)
	})
}
