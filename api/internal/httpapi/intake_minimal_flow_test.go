package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type intakeAccessStub struct{ allowed bool }

func (s intakeAccessStub) ClaimIntake(context.Context, int64, int64, int64) error { return nil }
func (s intakeAccessStub) CanAccessIntake(context.Context, int64, int64, int64, bool) (bool, error) {
	return s.allowed, nil
}

func TestIntakeReadIsBlockedByOwnershipAfterCapabilityCheck(t *testing.T) {
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	h := NewHandler(Dependencies{Auth: auth, IntakeAccess: intakeAccessStub{allowed: false}})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/intakes/51/books", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), "AUTH_FORBIDDEN") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

type intakeBookReaderStub struct{ book intake.Book }
func (s intakeBookReaderStub) GetBook(context.Context, int64, int64) (intake.Book, error) { return s.book, nil }

func TestIntakeBookDetailReturnsPersistedOriginalText(t *testing.T) {
	h := NewHandler(Dependencies{IntakeBooks: intakeBookReaderStub{book: intake.Book{
		ID: 9, IntakeID: 51, ExternalBookID: "1001", OriginalText: "完整正文", Status: intake.BookStatusFetched,
	}}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/intakes/51/books/9", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"originalText":"完整正文"`) {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
