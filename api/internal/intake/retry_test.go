package intake

import (
	"context"
	"errors"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
)

type retryFakeStore struct{ *fakeServiceStore }

func (s retryFakeStore) GetBook(_ context.Context, intakeID, bookID int64) (Book, error) {
	for _, book := range s.books[intakeID] {
		if book.ID == bookID {
			return book, nil
		}
	}
	return Book{}, ErrNotFound
}

func TestRetryBookRetriesOnlyFailedBookAndRestoresCompletedIntake(t *testing.T) {
	base := newFakeServiceStore()
	store := retryFakeStore{base}
	value, _ := base.CreateIntake(context.Background(), "重试批次")
	okBook, _ := base.UpsertBook(context.Background(), Book{IntakeID: value.ID, Source: "阳光", PlatformID: "4", ExternalBookID: "1001", OriginalText: "已成功正文", Status: BookStatusFetched})
	failedBook, _ := base.UpsertBook(context.Background(), Book{IntakeID: value.ID, Source: "阳光", PlatformID: "4", ExternalBookID: "1002", Status: BookStatusRetryableFailed, ErrorMessage: "timeout"})
	fetcher := &fake121Fetcher{results: map[string]provider121.Result{
		"4:1002": {Text: "重试正文", BookInfo: provider121.BookInfo{BookName: "重试成功", Category: "男生生活", Genre: "都市"}},
	}, errors: map[string]error{}}
	service := NewService(store, fetcher, nil)

	book, summary, err := service.RetryBook(context.Background(), value.ID, failedBook.ID, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if book.Status != BookStatusFetched || book.OriginalText != "重试正文" {
		t.Fatalf("book=%+v", book)
	}
	if summary.Status != StatusCompleted || summary.Fetched != 2 || summary.Failed != 0 {
		t.Fatalf("summary=%+v", summary)
	}
	if len(fetcher.requests) != 1 || fetcher.requests[0].BookID != "1002" {
		t.Fatalf("requests=%+v", fetcher.requests)
	}
	unchanged, _ := store.GetBook(context.Background(), value.ID, okBook.ID)
	if unchanged.OriginalText != "已成功正文" {
		t.Fatalf("successful book was modified: %+v", unchanged)
	}
}

func TestRetryBookDoesNotRefetchBodyAfterClassifierFailure(t *testing.T) {
	base := newFakeServiceStore()
	store := retryFakeStore{base}
	value, _ := base.CreateIntake(context.Background(), "分类重试")
	failedBook, _ := base.UpsertBook(context.Background(), Book{
		IntakeID: value.ID, Source: "阳光", PlatformID: "4", ExternalBookID: "2001",
		OriginalText: "已保存正文", Status: BookStatusRetryableFailed, ErrorMessage: "AI 分类失败",
	})
	fetcher := &fake121Fetcher{results: map[string]provider121.Result{}, errors: map[string]error{}}
	classifier := &fakeClassifier{result: ClassificationResult{Gender: "女频", Style: "情感"}}
	service := NewService(store, fetcher, classifier)

	book, _, err := service.RetryBook(context.Background(), value.ID, failedBook.ID, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if len(fetcher.requests) != 0 {
		t.Fatalf("body was refetched: %+v", fetcher.requests)
	}
	if book.Status != BookStatusFetched || book.Style != "情感" {
		t.Fatalf("book=%+v", book)
	}
}

func TestRetryBookRejectsPendingBook(t *testing.T) {
	base := newFakeServiceStore()
	store := retryFakeStore{base}
	value, _ := base.CreateIntake(context.Background(), "未执行批次")
	book, _ := base.UpsertBook(context.Background(), Book{IntakeID: value.ID, Source: "阳光", PlatformID: "4", ExternalBookID: "3000", Status: BookStatusPending})
	service := NewService(store, &fake121Fetcher{}, nil)
	_, _, err := service.RetryBook(context.Background(), value.ID, book.ID, 4000)
	if !errors.Is(err, ErrBookNotRetryable) {
		t.Fatalf("err=%v", err)
	}
}

func TestRetryBookRejectsAlreadyFetchedBook(t *testing.T) {
	base := newFakeServiceStore()
	store := retryFakeStore{base}
	value, _ := base.CreateIntake(context.Background(), "无需重试")
	book, _ := base.UpsertBook(context.Background(), Book{IntakeID: value.ID, Source: "阳光", PlatformID: "4", ExternalBookID: "3001", OriginalText: "正文", Status: BookStatusFetched})
	service := NewService(store, &fake121Fetcher{}, nil)
	_, _, err := service.RetryBook(context.Background(), value.ID, book.ID, 4000)
	if !errors.Is(err, ErrBookNotRetryable) {
		t.Fatalf("err=%v", err)
	}
}
