package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type observabilityIntakeService struct{}

func (observabilityIntakeService) CreateIntake(context.Context, intake.CreateIntakeInput) (intake.Intake, []intake.Book, error) {
	return intake.Intake{}, nil, nil
}

func (observabilityIntakeService) ExecuteIntake(context.Context, int64, int) (intake.ExecuteResult, error) {
	return intake.ExecuteResult{IntakeID: 77, Fetched: 1, Status: intake.StatusCompleted}, nil
}

type observabilityCountingReader struct {
	listBooksCalls int
}

func (r *observabilityCountingReader) ListIntakes(context.Context) ([]intake.Intake, error) {
	return nil, nil
}

func (r *observabilityCountingReader) ListBooks(context.Context, int64) ([]intake.Book, error) {
	r.listBooksCalls++
	return []intake.Book{{ID: 901, IntakeID: 77, Title: "private-book", Status: intake.BookStatusFetched}}, nil
}

func TestObservabilityExecuteIntakeDoesNotPerformLoggingOnlyResourceRead(t *testing.T) {
	reader := &observabilityCountingReader{}
	h := NewHandler(Dependencies{Intakes: observabilityIntakeService{}, Reader: reader})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes/77/execute", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if reader.listBooksCalls != 0 {
		t.Fatalf("observability added %d ListBooks read(s) after execution; logging must only observe already-authorized operation results", reader.listBooksCalls)
	}
}
