package intake

import (
	"context"
	"fmt"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/metadata"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
)

type serviceStore interface {
	CreateIntake(ctx context.Context, name string) (Intake, error)
	UpsertBook(ctx context.Context, book Book) (Book, error)
	ListBooks(ctx context.Context, intakeID int64) ([]Book, error)
	UpdateIntakeStatus(ctx context.Context, id int64, status Status) error
}

type Fetcher interface {
	Fetch(ctx context.Context, req provider121.Request) (provider121.Result, error)
}

type Classifier interface {
	Classify(ctx context.Context, input ClassificationInput) (ClassificationResult, error)
}

type ClassificationInput struct {
	BookID   string
	Title    string
	Text     string
	Category string
	Genre    string
	Gender   string
}

type ClassificationResult struct {
	Gender string
	Style  string
}

type Service struct {
	store      serviceStore
	fetcher    Fetcher
	classifier Classifier
}

type BookInput struct {
	BookID string
	Title  string
	Gender string
	Style  string
}

type BookGroup struct {
	Source     string
	PlatformID string
	Books      []BookInput
}

type CreateIntakeInput struct {
	Name   string
	Groups []BookGroup
}

type ExecuteResult struct {
	IntakeID int64  `json:"intakeId"`
	Fetched  int    `json:"fetched"`
	Failed   int    `json:"failed"`
	Status   Status `json:"status"`
}

// RestoreBook refreshes one existing book through the same provider121 and
// persistence path used by ExecuteIntake. It deliberately creates no task,
// queue, run, or worker of its own.
func (s *Service) RestoreBook(ctx context.Context, intakeID, bookID int64, maxText int) (Book, error) {
	if s.store == nil || s.fetcher == nil {
		return Book{}, fmt.Errorf("intake store and 121 fetcher are required")
	}
	books, err := s.store.ListBooks(ctx, intakeID)
	if err != nil {
		return Book{}, fmt.Errorf("读取 intake 书籍: %w", err)
	}
	var book Book
	found := false
	for _, value := range books {
		if value.ID == bookID {
			book = value
			found = true
			break
		}
	}
	if !found {
		return Book{}, fmt.Errorf("书籍不属于当前 intake")
	}
	if err := s.store.UpdateIntakeStatus(ctx, intakeID, StatusRunning); err != nil {
		return Book{}, fmt.Errorf("更新 intake 状态: %w", err)
	}
	fetched, fetchErr := s.fetcher.Fetch(ctx, provider121.Request{BookID: book.ExternalBookID, PlatformID: book.PlatformID, MaxText: maxText})
	if fetchErr != nil {
		book.Status = BookStatusRetryableFailed
		book.ErrorMessage = observability.SafeError(fetchErr)
		if _, err := s.store.UpsertBook(ctx, book); err != nil {
			return Book{}, fmt.Errorf("记录 121 获取失败: %w", err)
		}
		_ = s.refreshStatus(ctx, intakeID)
		return book, nil
	}
	book.OriginalText = fetched.Text
	book.Category = strings.TrimSpace(fetched.BookInfo.Category)
	book.Genre = normalizeGenre(fetched.BookInfo.Genre)
	if incoming := strings.TrimSpace(fetched.BookInfo.BookName); incoming != "" && isGeneratedTitle(book.Title, book.ExternalBookID) {
		book.Title = incoming
	}
	manualGender := ""
	if book.GenderSource == metadata.SourceManual || book.GenderSource == "input" {
		manualGender = book.Gender
	}
	resolvedGender, genderSource := metadata.ResolveGender(manualGender, book.Category, book.Genre, "")
	manualStyle := strings.TrimSpace(book.Style)
	ai := ClassificationResult{}
	if s.classifier != nil && (resolvedGender == "" || manualStyle == "") {
		ai, err = s.classifier.Classify(ctx, ClassificationInput{BookID: book.ExternalBookID, Title: book.Title, Text: fetched.Text, Category: book.Category, Genre: book.Genre, Gender: resolvedGender})
		if err != nil {
			book.Gender = resolvedGender
			book.GenderSource = genderSource
			book.Status = BookStatusRetryableFailed
			book.ErrorMessage = "AI 分类失败: " + observability.SafeError(err)
			_, saveErr := s.store.UpsertBook(ctx, book)
			if saveErr != nil {
				return Book{}, fmt.Errorf("记录 AI 分类失败: %w", saveErr)
			}
			_ = s.refreshStatus(ctx, intakeID)
			return book, nil
		}
	}
	book.Gender, book.GenderSource = metadata.ResolveGender(manualGender, book.Category, book.Genre, ai.Gender)
	book.Style, _ = metadata.ResolveStyle(manualStyle, "", ai.Style)
	book.Status = BookStatusFetched
	book.ErrorMessage = ""
	stored, err := s.store.UpsertBook(ctx, book)
	if err != nil {
		return Book{}, fmt.Errorf("保存已获取书籍: %w", err)
	}
	_ = s.refreshStatus(ctx, intakeID)
	return stored, nil
}

func (s *Service) refreshStatus(ctx context.Context, intakeID int64) error {
	books, err := s.store.ListBooks(ctx, intakeID)
	if err != nil {
		return err
	}
	fetched, failed := 0, 0
	for _, book := range books {
		if book.Status == BookStatusFetched {
			fetched++
		}
		if book.Status == BookStatusRetryableFailed {
			failed++
		}
	}
	status := StatusCompleted
	if failed > 0 && fetched == 0 {
		status = StatusFailed
	} else if failed > 0 {
		status = StatusPartial
	}
	return s.store.UpdateIntakeStatus(ctx, intakeID, status)
}

func NewService(store serviceStore, fetcher Fetcher, classifier Classifier) *Service {
	return &Service{store: store, fetcher: fetcher, classifier: classifier}
}

func (s *Service) CreateIntake(ctx context.Context, input CreateIntakeInput) (Intake, []Book, error) {
	if s.store == nil {
		return Intake{}, nil, fmt.Errorf("intake store is required")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "小说获取批次"
	}
	if len(input.Groups) == 0 {
		return Intake{}, nil, fmt.Errorf("至少添加一个书城")
	}

	prepared := make([]Book, 0)
	seen := make(map[string]struct{})
	for _, group := range input.Groups {
		source := strings.TrimSpace(group.Source)
		platformID := strings.TrimSpace(group.PlatformID)
		if source == "" || platformID == "" {
			return Intake{}, nil, fmt.Errorf("书城名称和 121 platformId 不能为空")
		}
		for _, row := range group.Books {
			bookID := strings.TrimSpace(row.BookID)
			if bookID == "" {
				return Intake{}, nil, fmt.Errorf("Book ID 不能为空")
			}
			key := source + "\x00" + bookID
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			gender, genderSource := metadata.ResolveGender(row.Gender, "", "", "")
			style, _ := metadata.ResolveStyle(row.Style, "", "")
			title := strings.TrimSpace(row.Title)
			if title == "" {
				title = "小说 " + bookID
			}
			prepared = append(prepared, Book{
				Source:         source,
				PlatformID:     platformID,
				ExternalBookID: bookID,
				Title:          title,
				Gender:         gender,
				GenderSource:   genderSource,
				Style:          style,
				Status:         BookStatusPending,
			})
			if len(prepared) > 200 {
				return Intake{}, nil, fmt.Errorf("单次最多添加 200 本小说")
			}
		}
	}
	if len(prepared) == 0 {
		return Intake{}, nil, fmt.Errorf("至少添加一本小说")
	}

	intakeValue, err := s.store.CreateIntake(ctx, name)
	if err != nil {
		return Intake{}, nil, err
	}
	created := make([]Book, 0, len(prepared))
	for _, book := range prepared {
		book.IntakeID = intakeValue.ID
		stored, err := s.store.UpsertBook(ctx, book)
		if err != nil {
			_ = s.store.UpdateIntakeStatus(ctx, intakeValue.ID, StatusFailed)
			return intakeValue, created, fmt.Errorf("保存书籍 %s: %w", book.ExternalBookID, err)
		}
		created = append(created, stored)
	}
	return intakeValue, created, nil
}

func (s *Service) ExecuteIntake(ctx context.Context, intakeID int64, maxText int) (ExecuteResult, error) {
	result := ExecuteResult{IntakeID: intakeID}
	if s.store == nil || s.fetcher == nil {
		return result, fmt.Errorf("intake store and 121 fetcher are required")
	}
	books, err := s.store.ListBooks(ctx, intakeID)
	if err != nil {
		return result, fmt.Errorf("读取 intake 书籍: %w", err)
	}
	if len(books) == 0 {
		return result, fmt.Errorf("intake 没有可执行书籍")
	}
	if err := s.store.UpdateIntakeStatus(ctx, intakeID, StatusRunning); err != nil {
		return result, fmt.Errorf("更新 intake 状态: %w", err)
	}

	for _, book := range books {
		if book.Status == BookStatusFetched {
			result.Fetched++
			continue
		}
		fetched, fetchErr := s.fetcher.Fetch(ctx, provider121.Request{
			BookID: book.ExternalBookID, PlatformID: book.PlatformID, MaxText: maxText,
		})
		if fetchErr != nil {
			book.Status = BookStatusRetryableFailed
			book.ErrorMessage = observability.SafeError(fetchErr)
			if _, err := s.store.UpsertBook(ctx, book); err != nil {
				return result, fmt.Errorf("记录 121 获取失败: %w", err)
			}
			result.Failed++
			continue
		}

		book.OriginalText = fetched.Text
		book.Category = strings.TrimSpace(fetched.BookInfo.Category)
		book.Genre = normalizeGenre(fetched.BookInfo.Genre)
		if incoming := strings.TrimSpace(fetched.BookInfo.BookName); incoming != "" && isGeneratedTitle(book.Title, book.ExternalBookID) {
			book.Title = incoming
		}

		manualGender := ""
		if book.GenderSource == metadata.SourceManual || book.GenderSource == "input" {
			manualGender = book.Gender
		}
		resolvedGender, genderSource := metadata.ResolveGender(manualGender, book.Category, book.Genre, "")
		manualStyle := strings.TrimSpace(book.Style)

		ai := ClassificationResult{}
		if s.classifier != nil && (resolvedGender == "" || manualStyle == "") {
			ai, err = s.classifier.Classify(ctx, ClassificationInput{
				BookID:   book.ExternalBookID,
				Title:    book.Title,
				Text:     fetched.Text,
				Category: book.Category,
				Genre:    book.Genre,
				Gender:   resolvedGender,
			})
			if err != nil {
				book.Gender = resolvedGender
				book.GenderSource = genderSource
				book.Status = BookStatusRetryableFailed
				book.ErrorMessage = "AI 分类失败: " + observability.SafeError(err)
				if _, saveErr := s.store.UpsertBook(ctx, book); saveErr != nil {
					return result, fmt.Errorf("记录 AI 分类失败: %w", saveErr)
				}
				result.Failed++
				continue
			}
		}
		book.Gender, book.GenderSource = metadata.ResolveGender(manualGender, book.Category, book.Genre, ai.Gender)
		book.Style, _ = metadata.ResolveStyle(manualStyle, "", ai.Style)
		book.Status = BookStatusFetched
		book.ErrorMessage = ""
		if _, err := s.store.UpsertBook(ctx, book); err != nil {
			return result, fmt.Errorf("保存已获取书籍: %w", err)
		}
		result.Fetched++
	}

	switch {
	case result.Failed == 0:
		result.Status = StatusCompleted
	case result.Fetched == 0:
		result.Status = StatusFailed
	default:
		result.Status = StatusPartial
	}
	if err := s.store.UpdateIntakeStatus(ctx, intakeID, result.Status); err != nil {
		return result, fmt.Errorf("写入 intake 最终状态: %w", err)
	}
	return result, nil
}

func normalizeGenre(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func isGeneratedTitle(title, bookID string) bool {
	current := strings.TrimSpace(title)
	id := strings.TrimSpace(bookID)
	return current == "" || current == id || current == "小说 "+id
}
