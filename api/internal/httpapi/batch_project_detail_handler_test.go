package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeBatchProjectDetailReader struct {
	project intake.BatchProject
	books   []intake.Book
}

func (f *fakeBatchProjectDetailReader) GetBatchProject(context.Context, int64) (intake.BatchProject, error) {
	return f.project, nil
}

func (f *fakeBatchProjectDetailReader) ListBooks(context.Context, int64) ([]intake.Book, error) {
	return f.books, nil
}

func TestBatchProjectDetailRouteIsRegistered(t *testing.T) {
	handler := NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/51", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body=%s, want 503 when detail reader is not wired", rec.Code, rec.Body.String())
	}
}

func TestBatchProjectDetailReturnsRealProjectBooksAndErrors(t *testing.T) {
	reader := &fakeBatchProjectDetailReader{
		project: intake.BatchProject{ID: 51, IntakeID: 11, Name: "真实批次"},
		books: []intake.Book{
			{ID: 31, IntakeID: 11, ExternalBookID: "1001", Title: "成功小说", Source: "知乎", PlatformID: "15", Gender: "女频", Style: "情感", Status: intake.BookStatusFetched},
			{ID: 32, IntakeID: 11, ExternalBookID: "1002", Title: "失败小说", Source: "点众", PlatformID: "7", Gender: "男频", Style: "悬疑", Status: intake.BookStatusRetryableFailed, ErrorMessage: "121 upstream error"},
		},
	}
	handler := NewHandler(Dependencies{BatchProjectDetails: reader})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/batch-projects/51", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`"project"`, `"id":51`, `"name":"真实批次"`,
		`"books"`, `"id":31`, `"bookId":"1001"`, `"title":"成功小说"`, `"source":"知乎"`, `"platformId":"15"`, `"gender":"女频"`, `"style":"情感"`, `"status":"fetched"`,
		`"id":32`, `"bookId":"1002"`, `"title":"失败小说"`, `"source":"点众"`, `"platformId":"7"`, `"gender":"男频"`, `"style":"悬疑"`, `"status":"retryable_failed"`, `"errorMessage":"121 upstream error"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body = %s, missing %s", body, want)
		}
	}
}
