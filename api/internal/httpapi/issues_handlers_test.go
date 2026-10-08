package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestIssuesProjectionScopesAndRedactsFacts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewHandler(Dependencies{Database: db, Auth: &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Capabilities: []string{CapabilityBatchView}}}})
	count := regexp.QuoteMeta("SELECT COUNT(*) FROM (") + ".*"
	mock.ExpectQuery(count).WithArgs(int64(7), int64(3)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{"id", "project_id", "book_id", "source", "status", "error_code", "request_id", "error_message", "at"}).
		AddRow(9, 12, 23, "stage_run", "failed", "", "request-9", "provider token=secret-value failed", time.Date(2026, 10, 7, 1, 2, 3, 0, time.UTC))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM (")+".*").WithArgs(int64(7), int64(3), 20, 0).WillReturnRows(rows)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?page=1", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret-value") {
		t.Fatalf("unsafe error leaked: %s", rec.Body.String())
	}
	var body struct {
		Total   int `json:"total"`
		Entries []struct {
			ID        string `json:"id"`
			RequestID string `json:"requestId"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 1 || len(body.Entries) != 1 || body.Entries[0].ID != "stage_run-9" || body.Entries[0].RequestID != "request-9" {
		t.Fatalf("unexpected projection: %s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestIssuesProjectionNormalizesUnionTextCollation(t *testing.T) {
	for _, want := range []string{
		"CONVERT(COALESCE(br.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci",
		"CONVERT(COALESCE(mt.error_message,'') USING utf8mb4) COLLATE utf8mb4_unicode_ci",
		"_utf8mb4'media_task' COLLATE utf8mb4_unicode_ci",
	} {
		if !strings.Contains(issueProjectionSQL, want) {
			t.Fatalf("issues UNION must normalize text collation; missing %q", want)
		}
	}
}
