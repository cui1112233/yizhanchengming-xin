package app

import (
	"bytes"
	"context"
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

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO intakes (name, status) VALUES (?, ?)")).
		WithArgs("知乎测试", intake.StatusPending).
		WillReturnResult(sqlmock.NewResult(11, 1))

	bookQuery := "INSERT INTO books (intake_id, source, platform_id, external_book_id, title, body_ref, original_text, category, genre, gender, gender_source, style, status, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), platform_id = VALUES(platform_id), title = VALUES(title), body_ref = VALUES(body_ref), original_text = VALUES(original_text), category = VALUES(category), genre = VALUES(genre), gender = VALUES(gender), gender_source = VALUES(gender_source), style = VALUES(style), status = VALUES(status), error_message = VALUES(error_message)"
	mock.ExpectExec(regexp.QuoteMeta(bookQuery)).
		WithArgs(int64(11), "知乎付费", "15", "1001", "测试书", "", "", "", "", "", "", "", intake.BookStatusPending, "").
		WillReturnResult(sqlmock.NewResult(21, 1))

	handler := NewHandler(db, fakeFetcher{}, nil, func() time.Time {
		return time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	})
	body := []byte(`{"name":"知乎测试","groups":[{"source":"知乎付费","platformId":"15","books":[{"bookId":"1001","title":"测试书"}]}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/intakes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, status, created_at, updated_at FROM intakes WHERE id = ?")).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "status", "created_at", "updated_at"}).
			AddRow(11, "已完成批次", intake.StatusCompleted, now, now))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO batch_projects (intake_id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id), name = VALUES(name)")).
		WithArgs(int64(11), "已完成批次").
		WillReturnResult(sqlmock.NewResult(51, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO runs (batch_project_id, run_at, status) VALUES (?, ?, ?)")).
		WithArgs(int64(51), now, intake.RunStatusPending).
		WillReturnResult(sqlmock.NewResult(71, 1))

	handler := NewHandler(db, fakeFetcher{}, nil, func() time.Time { return now })
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
