package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/videomanagementconnector"
)

type capabilityVideoManagementConnector struct{}

func (capabilityVideoManagementConnector) Login(context.Context, videomanagementconnector.LoginInput) (videomanagementconnector.Session, error) {
	return videomanagementconnector.Session{}, nil
}

func (capabilityVideoManagementConnector) CheckSession(context.Context, videomanagementconnector.Session) error {
	return nil
}

func (capabilityVideoManagementConnector) RefreshCatalogs(context.Context, videomanagementconnector.Session) (videomanagementconnector.Catalogs, error) {
	return videomanagementconnector.Catalogs{}, nil
}

func (capabilityVideoManagementConnector) Submit(context.Context, videomanagementconnector.Session, videomanagementconnector.SubmissionInput) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}

func (capabilityVideoManagementConnector) GetSubmission(context.Context, videomanagementconnector.Session, string) (videomanagementconnector.Submission, error) {
	return videomanagementconnector.Submission{}, nil
}

func capabilityNow() time.Time { return time.Unix(1700000000, 0) }

// capabilityRouterOptions registers every route the staged runtime can own so
// the golden table can compare slice semantics with the real route surface.
func capabilityRouterOptions(slice int, productionEnabled, mergeEnabled bool) RouterOptions {
	store := batchfactoryv11.NewMemoryStore()
	return RouterOptions{
		BridgeSecret: "secret",
		Now:          capabilityNow,
		Slice:        slice,
		Store:        store,
		Director:     &batchfactoryv11.DirectorService{Store: store},
		Compiler:     &batchfactoryv11.PromptCompilerService{Store: store},
		Production:   &batchfactoryv11.ProductionService{Store: store, Enabled: productionEnabled},
		Merge:        &batchfactoryv11.MergeService{Store: store, Enabled: mergeEnabled},
	}
}

func loadCapabilities(t *testing.T, api http.Handler, path string) map[string]Capability {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("capabilities status=%d body=%s", rec.Code, rec.Body.String())
	}
	var caps map[string]Capability
	if err := json.Unmarshal(rec.Body.Bytes(), &caps); err != nil {
		t.Fatalf("decode capabilities: %v body=%s", err, rec.Body.String())
	}
	return caps
}

func TestV11CapabilitiesRejectUnsignedRequest(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestV11CapabilitiesAreFalseBeforeSliceOne(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"settings.edit":{"available":false`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestV11CapabilitiesRejectExpiredSignature(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow().Add(-6*time.Minute), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestV12CapabilitiesUsesGoContract(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow, Slice: 1})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"batch.read"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

// TestCapabilityTruthTableAcrossSlices is the golden lock: the advertised
// capability surface for slice 0..6 must exactly match the staged release
// semantics, and every available capability must have a really registered
// route (deriveCapabilities only reports available when ServeMux matched it).
func TestCapabilityTruthTableAcrossSlices(t *testing.T) {
	baseCapabilities := []string{"batch.read", "batch.create", "settings.edit", "snapshot.read", "override.edit"}
	directorCapabilities := []string{"director.run", "hook.review", "working-front.viral"}
	for slice := 0; slice <= 6; slice++ {
		api := NewRouter(capabilityRouterOptions(slice, true, true))
		caps := loadCapabilities(t, api, "/api/batch-factory/v12/capabilities")
		expectAvailable := func(key string) bool {
			switch {
			case slice >= 1 && capabilityListContains(baseCapabilities, key):
				return true
			case slice >= 2 && capabilityListContains(directorCapabilities, key):
				return true
			case slice >= 3 && key == "compiler.preview":
				return true
			case slice >= 4 && (key == "production.submit" || key == "production.cancel"):
				return true
			case slice >= 5 && key == "merge.run":
				return true
			default:
				return false
			}
		}
		for _, spec := range capabilitySpecs {
			got := caps[spec.Key]
			want := expectAvailable(spec.Key)
			if got.Available != want {
				t.Fatalf("slice=%d %s available=%v want=%v reason=%q", slice, spec.Key, got.Available, want, got.Reason)
			}
			if got.Available && got.Reason != "" {
				t.Fatalf("slice=%d %s available but carries reason %q", slice, spec.Key, got.Reason)
			}
		}
		// publish capabilities need an explicit external/connector configuration.
		for _, key := range []string{"publish.121", "publish.yadi"} {
			if caps[key].Available {
				t.Fatalf("slice=%d %s must stay disabled without an external publisher", slice, key)
			}
		}
	}
}

func TestSliceThreeUnlocksCompilerPreviewOnly(t *testing.T) {
	api := NewRouter(capabilityRouterOptions(3, true, true))
	caps := loadCapabilities(t, api, "/api/batch-factory/v12/capabilities")
	if !caps["compiler.preview"].Available {
		t.Fatalf("compiler.preview=%+v", caps["compiler.preview"])
	}
	for _, key := range []string{"production.submit", "merge.run", "publish.121", "publish.yadi"} {
		if caps[key].Available {
			t.Fatalf("%s must remain unavailable in Slice 3", key)
		}
	}
}

func TestUnavailable121PublishingExplainsMissingVideoSubmissionContract(t *testing.T) {
	api := NewRouter(capabilityRouterOptions(3, true, true))
	caps := loadCapabilities(t, api, "/api/batch-factory/v12/capabilities")
	if caps["publish.121"].Reason != "121 视频发布未启用：尚未验证 MP4 提交接口、回执与回读" {
		t.Fatalf("reason=%q", caps["publish.121"].Reason)
	}
}

func TestSliceFiveReports121PublishingWhenTheGoConnectorIsConfigured(t *testing.T) {
	api := NewRouter(RouterOptions{
		BridgeSecret:             "secret",
		Now:                      capabilityNow,
		Slice:                    5,
		Store:                    batchfactoryv11.NewMemoryStore(),
		VideoManagementConnector: capabilityVideoManagementConnector{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"publish.121":{"available":true`) {
		t.Fatalf("publish.121 must reflect the configured Go uploader: %s", rec.Body.String())
	}
}

// TestUnregisteredV12RouteReturnsJSON404 locks the F contract: unknown routes
// still answer 404, but the body is a structured "capability not enabled"
// payload instead of bare text.
func TestUnregisteredV12RouteReturnsJSON404(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow, Slice: 0})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/batches/summary", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"CAPABILITY_NOT_ENABLED"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"requiredSlice":1`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("404 body is not JSON: %v body=%s", err, rec.Body.String())
	}
	if payload["error"] != "该能力未开启" {
		t.Fatalf("error=%v", payload["error"])
	}
}

// TestMethodMismatchKeeps405 guards against a wildcard 404 rewrite swallowing
// the real 405 semantics.
func TestMethodMismatchKeeps405(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: capabilityNow, Slice: 3})
	req := httptest.NewRequest(http.MethodPost, "/api/batch-factory/v12/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, capabilityNow(), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "CAPABILITY_NOT_ENABLED") {
		t.Fatalf("405 must not be rewritten: %s", rec.Body.String())
	}
}

func capabilityListContains(values []string, key string) bool {
	for _, value := range values {
		if value == key {
			return true
		}
	}
	return false
}
