package batchfactory

import (
	"context"
	"errors"
	"strconv"
	"strings"

	one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type BookFetcher interface {
	FetchBook(context.Context, string, string, int) (one21.BookResponse, error)
}

type PrepareBookInput struct {
	BookID           string
	PlatformID       string
	MaxTxt           int
	ManualGender     novel.Gender
	Style            string
	VerifiedGenreMap map[string]novel.Gender
}

type PreparedBook struct {
	BookID     string
	BookName   string
	Author     string
	SourceText string
	Category   string
	Genre      int
	Gender     novel.GenderResult
	Style      string
	NeedsAI    bool
	BookInfo   one21.BookInfo
}

type BookPreparationService struct {
	Fetcher BookFetcher
}

func (s BookPreparationService) Prepare(ctx context.Context, input PrepareBookInput) (PreparedBook, error) {
	if s.Fetcher == nil {
		return PreparedBook{}, errors.New("book fetcher is required")
	}
	response, err := s.Fetcher.FetchBook(ctx, strings.TrimSpace(input.BookID), strings.TrimSpace(input.PlatformID), input.MaxTxt)
	if err != nil {
		return PreparedBook{}, err
	}
	gender := novel.ResolveGender(novel.GenderInput{
		Manual:           input.ManualGender,
		Category:         response.BookInfo.Category,
		Genre:            strconv.Itoa(response.BookInfo.Genre),
		VerifiedGenreMap: input.VerifiedGenreMap,
	})
	style := strings.TrimSpace(input.Style)
	return PreparedBook{
		BookID:     response.BookInfo.BookID,
		BookName:   response.BookInfo.BookName,
		Author:     response.BookInfo.Author,
		SourceText: response.Data,
		Category:   response.BookInfo.Category,
		Genre:      response.BookInfo.Genre,
		Gender:     gender,
		Style:      style,
		NeedsAI:    gender.Gender == novel.GenderUnknown || style == "",
		BookInfo:   response.BookInfo,
	}, nil
}
