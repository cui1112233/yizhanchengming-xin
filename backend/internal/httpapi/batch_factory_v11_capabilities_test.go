package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestV11CapabilitiesRejectUnsignedRequest(t *testing.T) {
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return time.Unix(1700000000, 0) }})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestV11CapabilitiesAreFalseBeforeSliceOne(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, now, "secret")
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
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v11/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, now.Add(-6*time.Minute), "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestV12CapabilitiesUsesGoContract(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, now, "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"batch.read"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}

func TestSliceThreeUnlocksCompilerPreviewOnly(t *testing.T) {
	caps := CapabilitiesForSlice(3)
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
	caps := CapabilitiesForSlice(3)
	if caps["publish.121"].Reason != "121 视频发布未启用：尚未验证 MP4 提交接口、回执与回读" {
		t.Fatalf("reason=%q", caps["publish.121"].Reason)
	}
}

func TestSliceFiveReports121PublishingWhenTheGoConnectorIsConfigured(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret:             "secret",
		Now:                      func() time.Time { return now },
		Slice:                    5,
		VideoManagementConnector: capabilityVideoManagementConnector{},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/batch-factory/v12/capabilities", nil)
	SignBridgeRequest(req, "alpha-user", false, now, "secret")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"publish.121":{"available":true`) {
		t.Fatalf("publish.121 must reflect the configured Go uploader: %s", rec.Body.String())
	}
}
