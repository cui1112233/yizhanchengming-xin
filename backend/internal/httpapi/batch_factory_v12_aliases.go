package httpapi

import (
	"net/http"
	"strings"
)

// v12V11AliasPatterns is the frozen list of V12 routes served by forwarding to
// the V11 implementation. Adding an entry here grows the parallel V12 surface,
// so every change must be reviewed against the single-source V11 route and the
// frozen inventory test in batch_factory_v12_aliases_test.go.
//
// Removed: GET /api/batch-factory/v12/batches/{batchId}/runtime-index. No
// frontend caller or V12 test depended on it, so it was a dead parallel entry.
var v12V11AliasPatterns = []string{
	"GET /api/batch-factory/v12/batches/{batchId}",
	"POST /api/batch-factory/v12/intakes/{intakeId}/batches",
	"POST /api/batch-factory/v12/intakes/manual",
	"POST /api/batch-factory/v12/batches/{batchId}/intakes/{intakeId}/books",
	"PUT /api/batch-factory/v12/batches/{batchId}/settings",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/metadata",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/source",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/override",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/override",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages/{stage}",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages/retry",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/director-prompt-candidate/accept",
	"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"POST /api/batch-factory/v12/batches/{batchId}/director",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/opening-variants",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/hook",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/working-front/viral",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/hooks/{hookId}/approve",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/production",
	"POST /api/batch-factory/v12/batches/{batchId}/production",
	"POST /api/batch-factory/v12/batches/{batchId}/production/cancel",
	"POST /api/batch-factory/v12/batches/{batchId}/merge",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/merge",
	"POST /api/batch-factory/v12/batches/{batchId}/merge-artifacts/migrate-to-tos",
	"GET /api/batch-factory/v12/batches/{batchId}/merge-media/{artifactId}",
	"GET /api/batch-factory/v12/batches/{batchId}/merge-cover/{artifactId}",
	"GET /api/batch-factory/v12/publish/{provider}/credential",
	"PUT /api/batch-factory/v12/publish/{provider}/credential",
	"POST /api/batch-factory/v12/publish/{provider}/intents",
	"POST /api/batch-factory/v12/publish/{provider}/intents/{intentId}/confirm",
	"POST /api/batch-factory/v12/publish/{provider}/intents/{intentId}/submit",
	"GET /api/batch-factory/v12/publish/{provider}/intents/{intentId}/audits",
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
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets",
	"PATCH /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/upload",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/{imageId}/primary",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/{imageId}/content",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/upload",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/merge/upload",
	"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/tasks/{taskId}",
	"PUT /api/batch-factory/v12/video-provider/config",
}

// registerV12V11Aliases keeps the V12 browser contract on the Go store for
// operations whose V11 implementation already has the same owner-scoped
// semantics. The remaining V12 adapters stay on the compatibility bridge.
func registerV12V11Aliases(v12, v11 *http.ServeMux) {
	if v12 == nil || v11 == nil {
		return
	}
	for _, pattern := range v12V11AliasPatterns {
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
