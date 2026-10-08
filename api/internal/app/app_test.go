package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
)

type fakeFetcher struct{}

func (fakeFetcher) Fetch(context.Context, provider121.Request) (provider121.Result, error) {
	return provider121.Result{}, nil
}

func TestNewHandlerWiresIntakeServiceToMySQLStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	accessHashBytes := sha256.Sum256([]byte("access"))
	accessHash := hex.EncodeToString(accessHashBytes[:])
	mock.ExpectQuery(regexp.QuoteMeta("SELECT u.id, u.username, u.display_name, u.role, COALESCE(u.team_id, 0) FROM auth_sessions s JOIN auth_users u ON u.id = s.user_id WHERE s.access_token_hash = ? AND s.revoked_at IS NULL AND s.access_expires_at > ? AND u.active = TRUE LIMIT 1")).WithArgs(accessHash, sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "display_name", "role", "team_id"}).AddRow(7, "alice", "Alice", "member", 3))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT capability FROM auth_user_capabilities WHERE user_id = ? ORDER BY capability")).WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"capability"}).AddRow("batch.configure"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).WithArgs("知乎测试", intake.StatusPending).WillReturnResult(sqlmock.NewResult(11, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO auth_intake_ownership (intake_id, owner_user_id, team_id) VALUES (?, ?, ?)")).WithArgs(int64(11), int64(7), int64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	bookQuery := "INSERT INTO books (intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), platform_id = VALUES(platform_id), title = VALUES(title), body_ref = VALUES(body_ref), original_text = VALUES(original_text), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	mock.ExpectExec(regexp.QuoteMeta(bookQuery)).WithArgs(int64(11), "知乎付费", "15", "1001", "测试书", "", "", "", "", "", "", "", intake.BookStatusPending, "").WillReturnResult(sqlmock.NewResult(21, 1))
	handler := NewHandler(db, fakeFetcher{}, nil, func() time.Time { return time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC) })
	body := []byte(`{"name":"知乎测试","groups":[{"source":"知乎付费","platformId":"15","books":[{"bookId":"1001","title":"测试书"}]}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "ycm_access", Value: "access"})
	req.Host = "app.example"
	req.Header.Set("Origin", "http://app.example")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNewHandlerWiresPipelineToSameMySQLStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM batch_projects WHERE intake_id = ? AND archived_at IS NOT NULL)")).WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"archived"}).AddRow(false))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, status, created_at, updated_at FROM intakes WHERE id = ?")).WithArgs(int64(11)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}).AddRow(11, "已完成批次", intake.StatusCompleted, now, now))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = IF(archived_at IS NULL, VALUES(name), name)")).WithArgs(int64(11), "已完成批次").WillReturnResult(sqlmock.NewResult(51, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ?")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT archived_at FROM batch_projects WHERE id = ? FOR UPDATE")).WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"archived_at"}).AddRow(nil))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO runs (batch_project_id,idempotency_key,run_at,status) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)")).WithArgs(int64(51), "pipeline:intake:11:immediate", now, intake.RunStatusPending).WillReturnResult(sqlmock.NewResult(71, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,batch_project_id,run_at,status FROM runs WHERE id=?")).WithArgs(int64(71)).WillReturnRows(sqlmock.NewRows([]string{"id", "batch_project_id", "run_at", "status"}).AddRow(71, 51, now, intake.RunStatusPending))
	mock.ExpectCommit()
	handler := newHandler(db, fakeFetcher{}, nil, func() time.Time { return now }, false)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes/11/batch-projects", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNewHandlerWiresBatchProjectReaderToMySQLStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	mock.ExpectQuery("^SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("^SELECT bp.id").WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "intake_id", "name", "sources", "book_count", "genders", "styles", "run_status", "failure_count", "created_at", "updated_at", "archived_at"}).AddRow(51, 11, "知乎批次", "知乎", 2, "男频", "悬疑", intake.RunStatusPending, 0, now, now, nil))
	handler := newHandler(db, fakeFetcher{}, nil, func() time.Time { return now }, false)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	for _, want := range [][]byte{[]byte(`"id":51`), []byte(`"name":"知乎批次"`), []byte(`"sources":["知乎"]`), []byte(`"bookCount":2`), []byte(`"runStatus":"pending"`)} {
		if !bytes.Contains(rec.Body.Bytes(), want) {
			t.Fatalf("body = %s missing %s", rec.Body.String(), want)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
