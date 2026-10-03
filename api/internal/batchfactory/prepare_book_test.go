package batchfactory

import (
	"context"
	"testing"

	one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type fakeBookFetcher struct {
	response one21.BookResponse
	calls    int
}

func (f *fakeBookFetcher) FetchBook(context.Context, string, string, int) (one21.BookResponse, error) {
	f.calls++
	return f.response, nil
}

func TestPrepareBookManualGenderIsNeverOverwritten(t *testing.T) {
	fetcher := &fakeBookFetcher{response: one21.BookResponse{
		Data: "正文",
		BookInfo: one21.BookInfo{BookID: "123", BookName: "测试书", Category: "男生生活", Genre: 8},
	}}
	service := BookPreparationService{Fetcher: fetcher}

	result, err := service.Prepare(context.Background(), PrepareBookInput{
		BookID: "123", PlatformID: "2", MaxTxt: 2000,
		ManualGender: novel.GenderFemale,
		Style: "现代女主",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Gender.Gender != novel.GenderFemale || result.Gender.Source != novel.GenderSourceManual {
		t.Fatalf("manual gender was overwritten: %#v", result.Gender)
	}
	if result.NeedsAI {
		t.Fatal("complete deterministic metadata should skip AI")
	}
}

func TestPrepareBookUses121CategoryBeforeAI(t *testing.T) {
	fetcher := &fakeBookFetcher{response: one21.BookResponse{
		Data: "正文",
		BookInfo: one21.BookInfo{BookID: "7673480334440139800", BookName: "港岛雨停，再无爱意", Category: "男生生活", Genre: 8},
	}}
	service := BookPreparationService{Fetcher: fetcher}

	result, err := service.Prepare(context.Background(), PrepareBookInput{
		BookID: "7673480334440139800", PlatformID: "2", MaxTxt: 2000, Style: "现代通用",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Gender.Gender != novel.GenderMale || result.Gender.Source != novel.GenderSource121Category {
		t.Fatalf("expected category-derived male gender, got %#v", result.Gender)
	}
	if result.NeedsAI {
		t.Fatal("category + existing style should skip AI")
	}
	if result.SourceText != "正文" || result.BookName != "港岛雨停，再无爱意" {
		t.Fatalf("121 data was not preserved: %#v", result)
	}
}

func TestPrepareBookMarksAIOnlyWhenMetadataStillMissing(t *testing.T) {
	fetcher := &fakeBookFetcher{response: one21.BookResponse{
		Data: "正文",
		BookInfo: one21.BookInfo{BookID: "123", Category: "都市故事", Genre: 999},
	}}
	service := BookPreparationService{Fetcher: fetcher}

	result, err := service.Prepare(context.Background(), PrepareBookInput{BookID: "123", PlatformID: "2", MaxTxt: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if result.Gender.Gender != novel.GenderUnknown || !result.NeedsAI {
		t.Fatalf("expected unresolved metadata to request AI fallback, got %#v", result)
	}
}
