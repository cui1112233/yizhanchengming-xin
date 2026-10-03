package batchfactory

import (
	"context"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
)

type fakeIntakeStore struct {
	record IntakeRecord
}

func (f *fakeIntakeStore) CreateIntake(_ context.Context, record IntakeRecord) error {
	f.record = record
	return nil
}

func TestIntakeServiceCreatesGroupedBooksInOneIntake(t *testing.T) {
	store := &fakeIntakeStore{}
	service := IntakeService{Store: store, NewID: func() string { return "intake-1" }}

	result, err := service.Create(context.Background(), CreateIntakeInput{
		Owner: "user-1",
		Title: "10月3日批量",
		Groups: []IntakeGroupInput{
			{PlatformID: "zhihu", PlatformName: "知乎", MaxTxt: 2000, Books: []IntakeBookInput{{BookID: "z1"}, {BookID: "z2", ManualGender: novel.GenderFemale}}},
			{PlatformID: "dianzhong", PlatformName: "点众", Books: []IntakeBookInput{{BookID: "d1", Style: "现代通用"}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IntakeID != "intake-1" || result.GroupCount != 2 || result.BookCount != 3 {
		t.Fatalf("result=%+v", result)
	}
	if store.record.Owner != "user-1" || len(store.record.Books) != 3 {
		t.Fatalf("record=%+v", store.record)
	}
	if store.record.Books[0].PlatformName != "知乎" || store.record.Books[0].MaxTxt != 2000 {
		t.Fatalf("first book=%+v", store.record.Books[0])
	}
	if store.record.Books[1].ManualGender != novel.GenderFemale {
		t.Fatalf("manual gender lost: %+v", store.record.Books[1])
	}
	if store.record.Books[2].PlatformName != "点众" || store.record.Books[2].MaxTxt != 4000 || store.record.Books[2].Style != "现代通用" {
		t.Fatalf("third book=%+v", store.record.Books[2])
	}
}

func TestIntakeServiceRejectsDuplicateBookInSamePlatform(t *testing.T) {
	service := IntakeService{Store: &fakeIntakeStore{}, NewID: func() string { return "intake-1" }}
	_, err := service.Create(context.Background(), CreateIntakeInput{
		Owner: "user-1",
		Groups: []IntakeGroupInput{{PlatformID: "2", PlatformName: "番茄", Books: []IntakeBookInput{{BookID: "b1"}, {BookID: "b1"}}}},
	})
	if err == nil {
		t.Fatal("expected duplicate book error")
	}
}

func TestIntakeServiceRejectsInvalidManualGender(t *testing.T) {
	service := IntakeService{Store: &fakeIntakeStore{}, NewID: func() string { return "intake-1" }}
	_, err := service.Create(context.Background(), CreateIntakeInput{
		Owner: "user-1",
		Groups: []IntakeGroupInput{{PlatformID: "2", PlatformName: "番茄", Books: []IntakeBookInput{{BookID: "b1", ManualGender: novel.Gender("other")}}}},
	})
	if err == nil {
		t.Fatal("expected invalid gender error")
	}
}
