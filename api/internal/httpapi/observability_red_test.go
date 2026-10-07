package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

const testRequestIDHeader = "X-Request-ID"

func TestObservabilityRequestIDGeneratedAndReturned(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	got := rec.Header().Get(testRequestIDHeader)
	if got == "" {
		t.Fatal("X-Request-ID is empty; want server-generated request id")
	}
}

func TestObservabilityRequestIDAcceptsSafeSingleClientValue(t *testing.T) {
	const want = "e2e-Run_12:stage.3"
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(testRequestIDHeader, want)
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	if got := rec.Header().Get(testRequestIDHeader); got != want {
		t.Fatalf("X-Request-ID = %q, want trusted safe value %q", got, want)
	}
}

func TestObservabilityRequestIDRejectsUnsafeValues(t *testing.T) {
	cases := []struct {
		name   string
		values []string
	}{
		{name: "newline", values: []string{"good\nforged"}},
		{name: "control", values: []string{"good\x01bad"}},
		{name: "invalid-character", values: []string{"bad/request?id=1"}},
		{name: "too-long", values: []string{strings.Repeat("a", 257)}},
		{name: "multiple-values", values: []string{"one", "two"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.Header[testRequestIDHeader] = tc.values
			rec := httptest.NewRecorder()
			NewHandler().ServeHTTP(rec, req)

			got := rec.Header().Get(testRequestIDHeader)
			if got == "" {
				t.Fatal("server did not replace rejected request id")
			}
			for _, bad := range tc.values {
				if got == bad {
					t.Fatalf("unsafe/multi-valued request id %q was trusted", bad)
				}
			}
			if strings.ContainsAny(got, "\r\n\x00\x01") {
				t.Fatalf("generated request id contains control characters: %q", got)
			}
		})
	}
}

func TestObservabilityNotFoundStillReturnsRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/definitely-not-a-route?token=must-not-matter", nil)
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Header().Get(testRequestIDHeader); got == "" {
		t.Fatal("404 response missing X-Request-ID")
	}
}

func TestObservabilityReadyzRouteExists(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatal("/readyz is missing")
	}
	if got := rec.Header().Get(testRequestIDHeader); got == "" {
		t.Fatal("/readyz response missing X-Request-ID")
	}
}

func TestObservabilityDiagnosticsAreNotPublic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics", nil)
	rec := httptest.NewRecorder()
	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("diagnostics status = %d, want 401 for anonymous request", rec.Code)
	}
	if got := rec.Header().Get(testRequestIDHeader); got == "" {
		t.Fatal("diagnostics error missing X-Request-ID")
	}
}

func TestObservabilityIntakeInternalCauseIsNotReturned(t *testing.T) {
	const marker = "INTERNAL-DSN-password=supersecret"
	h := NewHandler(Dependencies{Intakes: failingIntakeService{err: errors.New(marker)}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", strings.NewReader(`{"name":"safe","groups":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), marker) || strings.Contains(rec.Body.String(), "supersecret") {
		t.Fatalf("response leaked internal cause: %s", rec.Body.String())
	}
	if got := rec.Header().Get(testRequestIDHeader); got == "" {
		t.Fatal("error response missing X-Request-ID")
	}
}

type failingIntakeService struct{ err error }

func (f failingIntakeService) CreateIntake(context.Context, intake.CreateIntakeInput) (intake.Intake, []intake.Book, error) {
	return intake.Intake{}, nil, f.err
}

func (f failingIntakeService) ExecuteIntake(context.Context, int64, int) (intake.ExecuteResult, error) {
	return intake.ExecuteResult{}, f.err
}

func (f failingIntakeService) RestoreBook(context.Context, int64, int64, int) (intake.Book, error) {
	return intake.Book{}, f.err
}
