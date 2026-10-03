package worker

import (
	"context"
	"errors"
	"strconv"
	"strings"

	one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type IntakeBook struct {
	ID           int64
	BookID       string
	PlatformID   string
	MaxTxt       int
	ManualGender novel.Gender
	Category     string
	Genre        int
	Style        string
	SourceText   string
}

type FetchedBook struct {
	BookID     string
	BookName   string
	Author     string
	SourceText string
	Category   string
	Genre      int
	BookInfo   one21.BookInfo
}

type ResolvedMetadata struct {
	Gender  novel.GenderResult
	Style   string
	NeedsAI bool
}

type AIClassification struct {
	Gender novel.Gender
	Style  string
}

type IntakeBookRepository interface {
	ListBooks(context.Context, string) ([]IntakeBook, error)
	SaveFetched(context.Context, int64, FetchedBook) error
	SaveResolved(context.Context, int64, ResolvedMetadata) error
}

type IntakeBookFetcher interface {
	FetchBook(context.Context, string, string, int) (one21.BookResponse, error)
}

type MetadataClassifier interface {
	Classify(context.Context, IntakeBook) (AIClassification, error)
}

type BatchCreator interface {
	CreateFromIntake(context.Context, string, string) (string, error)
}

type IntakeExecutor struct {
	Books            IntakeBookRepository
	Fetcher          IntakeBookFetcher
	Classifier       MetadataClassifier
	Batches          BatchCreator
	VerifiedGenreMap map[string]novel.Gender
}

func (e IntakeExecutor) Execute(ctx context.Context, job pipeline.Job, stage pipeline.Stage) error {
	if e.Books == nil {
		return errors.New("intake book repository is required")
	}
	if strings.TrimSpace(job.IntakeID) == "" {
		return errors.New("pipeline job intake id is required")
	}

	if stage == pipeline.StageCreateBatch {
		if e.Batches == nil {
			return errors.New("batch creator is required")
		}
		_, err := e.Batches.CreateFromIntake(ctx, job.IntakeID, job.ID)
		return err
	}

	books, err := e.Books.ListBooks(ctx, job.IntakeID)
	if err != nil {
		return err
	}

	switch stage {
	case pipeline.StageFetchBook:
		if e.Fetcher == nil {
			return errors.New("121 book fetcher is required")
		}
		for _, book := range books {
			maxTxt := book.MaxTxt
			if maxTxt <= 0 {
				maxTxt = 4000
			}
			response, err := e.Fetcher.FetchBook(ctx, book.BookID, book.PlatformID, maxTxt)
			if err != nil {
				return err
			}
			if err := e.Books.SaveFetched(ctx, book.ID, FetchedBook{
				BookID: response.BookInfo.BookID, BookName: response.BookInfo.BookName,
				Author: response.BookInfo.Author, SourceText: response.Data,
				Category: response.BookInfo.Category, Genre: response.BookInfo.Genre,
				BookInfo: response.BookInfo,
			}); err != nil {
				return err
			}
		}
		return nil

	case pipeline.StageResolveMetadata:
		for _, book := range books {
			gender := e.resolveGender(book)
			style := strings.TrimSpace(book.Style)
			if err := e.Books.SaveResolved(ctx, book.ID, ResolvedMetadata{
				Gender: gender, Style: style,
				NeedsAI: gender.Gender == novel.GenderUnknown || style == "",
			}); err != nil {
				return err
			}
		}
		return nil

	case pipeline.StageAIClassify:
		for _, book := range books {
			gender := e.resolveGender(book)
			style := strings.TrimSpace(book.Style)
			if gender.Gender != novel.GenderUnknown && style != "" {
				if err := e.Books.SaveResolved(ctx, book.ID, ResolvedMetadata{Gender: gender, Style: style, NeedsAI: false}); err != nil {
					return err
				}
				continue
			}
			if e.Classifier == nil {
				return errors.New("AI metadata classifier is required for unresolved book metadata")
			}
			classification, err := e.Classifier.Classify(ctx, book)
			if err != nil {
				return err
			}
			if gender.Gender == novel.GenderUnknown {
				if classification.Gender != novel.GenderMale && classification.Gender != novel.GenderFemale {
					return errors.New("AI classifier returned unresolved gender")
				}
				gender = novel.GenderResult{Gender: classification.Gender, Source: novel.GenderSourceAI}
			}
			if style == "" {
				style = strings.TrimSpace(classification.Style)
				if style == "" {
					return errors.New("AI classifier returned empty style")
				}
			}
			if err := e.Books.SaveResolved(ctx, book.ID, ResolvedMetadata{Gender: gender, Style: style, NeedsAI: false}); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("unsupported intake pipeline stage")
	}
}

func (e IntakeExecutor) resolveGender(book IntakeBook) novel.GenderResult {
	return novel.ResolveGender(novel.GenderInput{
		Manual: book.ManualGender,
		Category: book.Category,
		Genre: strconv.Itoa(book.Genre),
		VerifiedGenreMap: e.VerifiedGenreMap,
	})
}
