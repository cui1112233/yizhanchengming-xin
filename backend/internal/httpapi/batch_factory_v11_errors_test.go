package httpapi

import (
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
