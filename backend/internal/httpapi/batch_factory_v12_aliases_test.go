package httpapi

import (
	"net/http"
	"sort"
	"strings"
	"testing"
)

// frozenV12V11AliasPatterns is the reviewed inventory of V12 routes that
// forward to the V11 implementation. It is intentionally duplicated here so a
// new parallel V12 entry cannot appear without an explicit diff to this list.
// If a route is deliberately added or retired, update both this list and
// v12V11AliasPatterns in batch_factory_v12_aliases.go in the same change.
var frozenV12V11AliasPatterns = []string{
	"DELETE /api/batch-factory/v12/batches/{batchId}",
	"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}",
	"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"DELETE /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/tasks/{taskId}",
	"GET /api/batch-factory/v12/batches/{batchId}",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/{imageId}/content",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages",
	"GET /api/batch-factory/v12/batches/{batchId}/merge-cover/{artifactId}",
	"GET /api/batch-factory/v12/batches/{batchId}/merge-media/{artifactId}",
	"GET /api/batch-factory/v12/config-versions",
	"GET /api/batch-factory/v12/drafts",
	"GET /api/batch-factory/v12/prompts",
	"GET /api/batch-factory/v12/publish/{provider}/credential",
	"GET /api/batch-factory/v12/publish/{provider}/intents/{intentId}/audits",
	"PATCH /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}",
	"POST /api/batch-factory/v12/batches",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/upload",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/director-prompt-candidate",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/hook",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/hooks/{hookId}/approve",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/merge",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/merge/upload",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/opening-variants",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/production",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages/retry",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/stages/{stage}",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/director-prompt-candidate/accept",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/upload",
	"POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/working-front/viral",
	"POST /api/batch-factory/v12/batches/{batchId}/change-impact",
	"POST /api/batch-factory/v12/batches/{batchId}/director",
	"POST /api/batch-factory/v12/batches/{batchId}/intakes/{intakeId}/books",
	"POST /api/batch-factory/v12/batches/{batchId}/merge",
	"POST /api/batch-factory/v12/batches/{batchId}/merge-artifacts/migrate-to-tos",
	"POST /api/batch-factory/v12/batches/{batchId}/production",
	"POST /api/batch-factory/v12/batches/{batchId}/production/cancel",
	"POST /api/batch-factory/v12/config-versions",
	"POST /api/batch-factory/v12/intakes/manual",
	"POST /api/batch-factory/v12/intakes/{intakeId}/batches",
	"POST /api/batch-factory/v12/prompts",
	"POST /api/batch-factory/v12/publish/{provider}/intents",
	"POST /api/batch-factory/v12/publish/{provider}/intents/{intentId}/confirm",
	"POST /api/batch-factory/v12/publish/{provider}/intents/{intentId}/submit",
	"POST /api/batch-factory/v12/private-comic/tasks",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/{assetId}/images/{imageId}/primary",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/metadata",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/override",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/source",
	"PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/videos/{videoId}/override",
	"PUT /api/batch-factory/v12/batches/{batchId}/settings",
	"PUT /api/batch-factory/v12/config-versions/{versionId}",
	"PUT /api/batch-factory/v12/drafts",
	"PUT /api/batch-factory/v12/publish/{provider}/credential",
	"PUT /api/batch-factory/v12/video-provider/config",
}

func TestV12V11AliasesMatchFrozenInventory(t *testing.T) {
	actual := append([]string(nil), v12V11AliasPatterns...)
	sort.Strings(actual)
	frozen := append([]string(nil), frozenV12V11AliasPatterns...)
	sort.Strings(frozen)

	if len(actual) != len(frozen) {
		t.Fatalf("v12 alias count=%d, frozen inventory=%d; adding or removing a V12 alias requires an explicit inventory update", len(actual), len(frozen))
	}
	for i := range frozen {
		if actual[i] != frozen[i] {
			t.Fatalf("v12 alias inventory drift at %d: actual=%q frozen=%q", i, actual[i], frozen[i])
		}
	}
}

func TestV12V11AliasesHaveNoDuplicatesAndStayUnderV12Prefix(t *testing.T) {
	allowedMethods := map[string]struct{}{
		http.MethodGet: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {}, http.MethodDelete: {},
	}
	seen := make(map[string]struct{}, len(v12V11AliasPatterns))
	for _, pattern := range v12V11AliasPatterns {
		parts := strings.SplitN(pattern, " ", 2)
		if len(parts) != 2 {
			t.Fatalf("alias pattern %q is not method-qualified", pattern)
		}
		if _, ok := allowedMethods[parts[0]]; !ok {
			t.Fatalf("alias pattern %q uses unexpected method %q", pattern, parts[0])
		}
		if !strings.HasPrefix(parts[1], "/api/batch-factory/v12/") {
			t.Fatalf("alias pattern %q is not under /api/batch-factory/v12/", pattern)
		}
		if _, exists := seen[pattern]; exists {
			t.Fatalf("duplicate v12 alias pattern %q would panic ServeMux registration", pattern)
		}
		seen[pattern] = struct{}{}
	}
}

func TestV12RuntimeIndexAliasIsNotReintroduced(t *testing.T) {
	for _, pattern := range v12V11AliasPatterns {
		if strings.HasSuffix(pattern, "/runtime-index") {
			t.Fatalf("dead parallel entry %q was reintroduced; the V11 runtime-index route is the only supported one", pattern)
		}
	}
}
