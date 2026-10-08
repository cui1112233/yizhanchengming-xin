package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestAccountProfileWritesOnlyCurrentUsersSafeDisplayName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7, Username: "alice", DisplayName: "Alice", Role: "member", TeamID: 3, Capabilities: []string{"batch.view"}}}, Database: db})
	mock.ExpectExec(regexp.QuoteMeta("UPDATE auth_users SET display_name=? WHERE id=? AND active=TRUE")).WithArgs("Alice New", int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	req := httptest.NewRequest(http.MethodPut, "/api/v1/account/profile", bytes.NewBufferString(`{"displayName":"Alice New"}`))
	sameOrigin(req)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, forbidden := range []string{"password", "cookie", "session", "secret", "recovery"} {
		if strings.Contains(strings.ToLower(rec.Body.String()), forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, rec.Body.String())
		}
	}
	if !strings.Contains(rec.Body.String(), "Alice New") || !strings.Contains(rec.Body.String(), "batch.view") {
		t.Fatalf("safe profile body=%s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAccountProfileWriteRejectsCrossOrigin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewHandler(Dependencies{Auth: &fakeAuthService{user: authn.User{ID: 7}}, Database: db})
	req := httptest.NewRequest(http.MethodPut, "http://example.com/api/v1/account/profile", bytes.NewBufferString(`{"displayName":"Alice"}`))
	req.Header.Set("Origin", "https://attacker.example")
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
