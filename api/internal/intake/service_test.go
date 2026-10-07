package intake

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/metadata"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
)

type fakeServiceStore struct {
	nextIntakeID int64
	nextBookID   int64
	intakes      map[int64]Intake
	books        map[int64][]Book
	statuses     []Status
}

func newFakeServiceStore() *fakeServiceStore {
	return &fakeServiceStore{nextIntakeID: 1, nextBookID: 1, intakes: map[int64]Intake{}, books: map[int64][]Book{}}
}

func (s *fakeServiceStore) CreateIntake(_ context.Context, name string) (Intake, error) {
	id := s.nextIntakeID
	s.nextIntakeID++
	value := Intake{ID: id, Name: name, Status: StatusPending}
	s.intakes[id] = value
	return value, nil
}

func (s *fakeServiceStore) GetIntake(_ context.Context, id int64) (Intake, error) {
	value, ok := s.intakes[id]
	if !ok {
		return Intake{}, errors.New("not found")
	}
	return value, nil
}

func (s *fakeServiceStore) UpsertBook(_ context.Context, book Book) (Book, error) {
	rows := s.books[book.IntakeID]
	for i := range rows {
		if rows[i].Source == book.Source && rows[i].ExternalBookID == book.ExternalBookID {
			book.ID = rows[i].ID
			rows[i] = book
			s.books[book.IntakeID] = rows
			return book, nil
		}
	}
	book.ID = s.nextBookID
	s.nextBookID++
	s.books[book.IntakeID] = append(rows, book)
	return book, nil
}

func (s *fakeServiceStore) ListBooks(_ context.Context, intakeID int64) ([]Book, error) {
	rows := append([]Book(nil), s.books[intakeID]...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func (s *fakeServiceStore) UpdateIntakeStatus(_ context.Context, intakeID int64, status Status) error {
	value := s.intakes[intakeID]
	value.Status = status
	s.intakes[intakeID] = value
	s.statuses = append(s.statuses, status)
	return nil
}

type fake121Fetcher struct {
	results  map[string]provider121.Result
	errors   map[string]error
	requests []provider121.Request
}

func (f *fake121Fetcher) Fetch(_ context.Context, req provider121.Request) (provider121.Result, error) {
	f.requests = append(f.requests, req)
	key := req.PlatformID + ":" + req.BookID
	if err := f.errors[key]; err != nil {
		return provider121.Result{}, err
	}
	return f.results[key], nil
}

type fakeClassifier struct {
	result ClassificationResult
	err    error
	seen   []ClassificationInput
}

func (f *fakeClassifier) Classify(_ context.Context, input ClassificationInput) (ClassificationResult, error) {
	f.seen = append(f.seen, input)
	return f.result, f.err
}

func TestServiceCreateIntakePreservesBookstoreGroupsAndDedupesWithinGroup(t *testing.T) {
	store := newFakeServiceStore()
	service := NewService(store, nil, nil)

	created, books, err := service.CreateIntake(context.Background(), CreateIntakeInput{
		Name: "知乎 + 黑岩",
		Groups: []BookGroup{
			{
				Source: "知乎付费", PlatformID: "15",
				Books: []BookInput{
					{BookID: "1001", Title: "知乎书一", Gender: "女频", Style: "现代甜文"},
					{BookID: "1001", Title: "知乎书一重复"},
					{BookID: "1002", Title: "知乎书二"},
				},
			},
			{Source: "黑岩付费", PlatformID: "1", Books: []BookInput{{BookID: "1001", Title: "黑岩同 ID"}}},
		},
	})
	if err != nil {
		t.Fatalf("CreateIntake: %v", err)
	}
	if created.ID == 0 || created.Status != StatusPending {
		t.Fatalf("created = %+v", created)
	}
	if len(books) != 3 {
		t.Fatalf("len(books) = %d, want 3", len(books))
	}
	if books[0].Source != "知乎付费" || books[0].PlatformID != "15" || books[0].ExternalBookID != "1001" {
		t.Fatalf("first book = %+v", books[0])
	}
	if books[0].Gender != "女频" || books[0].GenderSource != metadata.SourceManual || books[0].Style != "现代甜文" {
		t.Fatalf("manual metadata lost: %+v", books[0])
	}
	if books[2].Source != "黑岩付费" || books[2].PlatformID != "1" {
		t.Fatalf("cross-source identity lost: %+v", books[2])
	}
}

func TestServiceExecuteIntakeFetches121ResolvesMetadataAndKeepsPartialFailureRetryable(t *testing.T) {
	store := newFakeServiceStore()
	intake, _ := store.CreateIntake(context.Background(), "混合书城")
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intake.ID, Source: "番茄付费", PlatformID: "2", ExternalBookID: "2001", Title: "小说 2001", Status: BookStatusPending})
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intake.ID, Source: "知乎付费", PlatformID: "15", ExternalBookID: "2002", Title: "手工标题", Gender: "女频", GenderSource: metadata.SourceManual, Style: "古风虐文", Status: BookStatusPending})
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intake.ID, Source: "黑岩付费", PlatformID: "1", ExternalBookID: "2003", Status: BookStatusPending})

	fetcher := &fake121Fetcher{
		results: map[string]provider121.Result{
			"2:2001": {Text: "第一章\n正文 A", BookInfo: provider121.BookInfo{BookName: "港岛雨停，再无爱意", Category: "男生生活", Genre: float64(8)}},
			"15:2002": {Text: "正文 B", BookInfo: provider121.BookInfo{BookName: "121 标题不应覆盖手工标题", Category: "男生生活", Genre: "现代言情"}},
		},
		errors: map[string]error{"1:2003": errors.New("upstream timeout")},
	}
	classifier := &fakeClassifier{result: ClassificationResult{Gender: "女频", Style: "现代通用"}}
	service := NewService(store, fetcher, classifier)

	result, err := service.ExecuteIntake(context.Background(), intake.ID, 4000)
	if err != nil {
		t.Fatalf("ExecuteIntake: %v", err)
	}
	if result.Status != StatusPartial || result.Fetched != 2 || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
	if len(fetcher.requests) != 3 {
		t.Fatalf("fetch requests = %d, want 3", len(fetcher.requests))
	}

	books, _ := store.ListBooks(context.Background(), intake.ID)
	first := books[0]
	if first.Status != BookStatusFetched || first.OriginalText != "第一章\n正文 A" {
		t.Fatalf("first fetch state = %+v", first)
	}
	if first.Title != "港岛雨停，再无爱意" || first.Category != "男生生活" || first.Genre != "8" {
		t.Fatalf("first 121 metadata = %+v", first)
	}
	if first.Gender != "男频" || first.GenderSource != metadata.SourceCategory {
		t.Fatalf("first gender = %q/%q", first.Gender, first.GenderSource)
	}
	if first.Style != "现代通用" {
		t.Fatalf("first style = %q", first.Style)
	}

	second := books[1]
	if second.Title != "手工标题" {
		t.Fatalf("manual title overwritten: %q", second.Title)
	}
	if second.Gender != "女频" || second.GenderSource != metadata.SourceManual || second.Style != "古风虐文" {
		t.Fatalf("manual metadata overwritten: %+v", second)
	}

	third := books[2]
	if third.Status != BookStatusRetryableFailed || third.ErrorMessage == "" {
		t.Fatalf("third = %+v", third)
	}
	if got := store.intakes[intake.ID].Status; got != StatusPartial {
		t.Fatalf("intake status = %q, want %q", got, StatusPartial)
	}
}

func TestServiceExecuteIntakeAllProviderFailuresMarksIntakeFailed(t *testing.T) {
	store := newFakeServiceStore()
	intake, _ := store.CreateIntake(context.Background(), "失败批次")
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intake.ID, Source: "番茄付费", PlatformID: "2", ExternalBookID: "3001", Status: BookStatusPending})
	fetcher := &fake121Fetcher{results: map[string]provider121.Result{}, errors: map[string]error{"2:3001": fmt.Errorf("121 unavailable")}}
	service := NewService(store, fetcher, nil)

	result, err := service.ExecuteIntake(context.Background(), intake.ID, 4000)
	if err != nil {
		t.Fatalf("ExecuteIntake: %v", err)
	}
	if result.Status != StatusFailed || result.Fetched != 0 || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
}

func TestServiceExecuteIntakeClassifierFailureIsRetryableAndDoesNotClaimSuccess(t *testing.T) {
	store := newFakeServiceStore()
	intake, _ := store.CreateIntake(context.Background(), "分类失败")
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intake.ID, Source: "番茄付费", PlatformID: "2", ExternalBookID: "4001", Status: BookStatusPending})
	fetcher := &fake121Fetcher{results: map[string]provider121.Result{
		"2:4001": {Text: "正文", BookInfo: provider121.BookInfo{Category: "都市生活", Genre: float64(8)}},
	}, errors: map[string]error{}}
	classifier := &fakeClassifier{err: errors.New("ai unavailable")}
	service := NewService(store, fetcher, classifier)

	result, err := service.ExecuteIntake(context.Background(), intake.ID, 4000)
	if err != nil {
		t.Fatalf("ExecuteIntake: %v", err)
	}
	if result.Status != StatusFailed || result.Failed != 1 {
		t.Fatalf("result = %+v", result)
	}
	books, _ := store.ListBooks(context.Background(), intake.ID)
	if books[0].Status != BookStatusRetryableFailed || books[0].OriginalText != "正文" {
		t.Fatalf("book = %+v", books[0])
	}
}


func TestServiceExecuteIntakeRetryOnlyFailedBooksAndCompletes(t *testing.T) {
	store := newFakeServiceStore()
	intakeValue, _ := store.CreateIntake(context.Background(), "可重试批次")
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intakeValue.ID, Source: "番茄付费", PlatformID: "2", ExternalBookID: "5001", Status: BookStatusPending})
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intakeValue.ID, Source: "知乎付费", PlatformID: "15", ExternalBookID: "5002", Status: BookStatusPending})

	fetcher := &fake121Fetcher{
		results: map[string]provider121.Result{
			"2:5001": {Text: "正文 A", BookInfo: provider121.BookInfo{BookName: "成功书", Category: "男生生活"}},
		},
		errors: map[string]error{"15:5002": errors.New("temporary upstream timeout")},
	}
	service := NewService(store, fetcher, nil)

	first, err := service.ExecuteIntake(context.Background(), intakeValue.ID, 4000)
	if err != nil {
		t.Fatalf("first ExecuteIntake: %v", err)
	}
	if first.Status != StatusPartial || first.Fetched != 1 || first.Failed != 1 {
		t.Fatalf("first = %+v", first)
	}

	delete(fetcher.errors, "15:5002")
	fetcher.results["15:5002"] = provider121.Result{Text: "正文 B", BookInfo: provider121.BookInfo{BookName: "重试成功书", Category: "现代言情"}}

	second, err := service.ExecuteIntake(context.Background(), intakeValue.ID, 4000)
	if err != nil {
		t.Fatalf("second ExecuteIntake: %v", err)
	}
	if second.Status != StatusCompleted || second.Fetched != 2 || second.Failed != 0 {
		t.Fatalf("second = %+v", second)
	}

	firstBookRequests := 0
	secondBookRequests := 0
	for _, request := range fetcher.requests {
		switch request.BookID {
		case "5001":
			firstBookRequests++
		case "5002":
			secondBookRequests++
		}
	}
	if firstBookRequests != 1 || secondBookRequests != 2 {
		t.Fatalf("request counts = first:%d second:%d, want 1/2", firstBookRequests, secondBookRequests)
	}
	books, _ := store.ListBooks(context.Background(), intakeValue.ID)
	if books[0].Status != BookStatusFetched || books[1].Status != BookStatusFetched {
		t.Fatalf("books after retry = %+v", books)
	}
}

func TestServiceExecuteIntakeSanitizesSensitiveFailureBeforePersistence(t *testing.T) {
	store := newFakeServiceStore()
	intakeValue, _ := store.CreateIntake(context.Background(), "敏感错误批次")
	_, _ = store.UpsertBook(context.Background(), Book{IntakeID: intakeValue.ID, Source: "番茄付费", PlatformID: "2", ExternalBookID: "6001", Status: BookStatusPending})
	fetcher := &fake121Fetcher{
		results: map[string]provider121.Result{},
		errors: map[string]error{
			"2:6001": errors.New("provider failed token=secret-123 dsn=user:pass@tcp(127.0.0.1:3306)/novels"),
		},
	}
	service := NewService(store, fetcher, nil)

	result, err := service.ExecuteIntake(context.Background(), intakeValue.ID, 4000)
	if err != nil {
		t.Fatalf("ExecuteIntake: %v", err)
	}
	if result.Status != StatusFailed {
		t.Fatalf("result = %+v", result)
	}
	books, _ := store.ListBooks(context.Background(), intakeValue.ID)
	message := books[0].ErrorMessage
	if strings.Contains(message, "secret-123") || strings.Contains(message, "user:pass") {
		t.Fatalf("sensitive error leaked into persistence: %q", message)
	}
	if !strings.Contains(message, "[REDACTED]") {
		t.Fatalf("sanitized marker missing: %q", message)
	}
}
