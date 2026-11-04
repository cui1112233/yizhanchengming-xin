package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestWriteStoreErrorKeepsSpecificConflictReason(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, fmt.Errorf("%w: current book has no failed stage", batchfactoryv11.ErrConflict))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "current book has no failed stage") {
		t.Fatalf("body must retain the actionable conflict reason, got %s", rec.Body.String())
	}
}

func TestWriteStoreErrorKeepsTransientUpstreamReasonAndUsesServiceUnavailable(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, fmt.Errorf("%w: media duration probe failed: stream error", batchfactoryv11.ErrUpstream))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "media duration probe failed") {
		t.Fatalf("body must retain the actionable upstream reason, got %s", rec.Body.String())
	}
}

func TestWriteStoreErrorKeepsSpecificInvalidReason(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, fmt.Errorf("%w: AutoDL H3 API key is required", batchfactoryv11.ErrInvalid))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "AutoDL H3 API key is required") {
		t.Fatalf("body must retain the actionable validation reason, got %s", rec.Body.String())
	}
}

func TestWriteStoreErrorUnknownErrorReturnsInternalCode(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, errors.New("boom"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"INTERNAL"`) {
		t.Fatalf("unknown errors must expose a stable code, got %s", rec.Body.String())
	}
}

func TestWriteStoreErrorN8DirectoryRejectionIsUpstreamWithCode(t *testing.T) {
	rec := httptest.NewRecorder()
	writeStoreError(rec, n8DirectoryError(errors.New("漫画后台作品目录未配置，无法按书名自动匹配作品 ID")))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, ErrorCodeN8DirectoryRejected) {
		t.Fatalf("body must carry %s, got %s", ErrorCodeN8DirectoryRejected, body)
	}
	if !strings.Contains(body, "漫画后台作品目录未配置") {
		t.Fatalf("body must retain the directory reason, got %s", body)
	}
}
