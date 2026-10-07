package httpapi

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestWorkspaceHistoryScopesSearchAndPaginates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewHandler(Dependencies{Database: db, Auth: &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Capabilities: []string{CapabilityBatchView}}}})
	args := []driver.Value{int64(7), int64(3), "%故事%", "%故事%", "%故事%", "failed"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM") + ".*").WithArgs(args...).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT bp.id,b.id") + ".*").WithArgs(append(args, 20, 0)...).WillReturnRows(sqlmock.NewRows([]string{"project_id", "book_id", "project_name", "title", "status", "updated_at"}).AddRow(12, 23, "可访问项目", "故事书", "failed", time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/history?q=故事&status=failed", nil)
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Total   int `json:"total"`
		Entries []struct {
			ProjectID int64     `json:"projectId"`
			UpdatedAt time.Time `json:"updatedAt"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 1 || len(body.Entries) != 1 || body.Entries[0].ProjectID != 12 || body.Entries[0].UpdatedAt.IsZero() {
		t.Fatalf("unexpected body=%s", rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
