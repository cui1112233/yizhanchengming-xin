package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestResolveGenderManualWinsOverBookInfoAndAI(t *testing.T) {
	calls := 0
	got := ResolveGender(context.Background(), "女频", BookInfo{Category: "男频", Genre: "男性向"}, func(context.Context, BookInfo) (string, error) {
		calls++
		return `{"gender":"男频","confidence":0.94,"reason":"平台分类冲突时按正文主线判断"}`, nil
	})
	if got.Gender != "女频" || got.Source != GenderManual || calls != 0 {
		t.Fatalf("got=%#v calls=%d", got, calls)
	}
}

func TestResolveGenderUsesBookInfoWithoutAI(t *testing.T) {
	calls := 0
	got := ResolveGender(context.Background(), "", BookInfo{Category: "", Genre: "女性向"}, func(context.Context, BookInfo) (string, error) {
		calls++
		return "男频", nil
	})
	if got.Gender != "女频" || got.Source != GenderBookInfo || calls != 0 {
		t.Fatalf("got=%#v calls=%d", got, calls)
	}
}

func TestResolveGenderUsesStrongThemeMetadataWithoutAI(t *testing.T) {
	calls := 0
	got := ResolveGender(context.Background(), "", BookInfo{Themes: "言情、现实情感、婚恋", Labels: "现代"}, func(context.Context, BookInfo) (string, error) {
		calls++
		return "男频", nil
	})
	if got.Gender != "女频" || got.Source != GenderBookInfo || calls != 0 {
		t.Fatalf("theme metadata decision = %#v, calls=%d", got, calls)
	}
}

func TestResolveGenderPreservesExplicitCategoryWhenGenreIsOrdinaryText(t *testing.T) {
	calls := 0
	got := ResolveGender(context.Background(), "", BookInfo{Category: "男频", Genre: "都市"}, func(context.Context, BookInfo) (string, error) {
		calls++
		return "女频", nil
	})
	if got.Gender != "男频" || got.Source != GenderBookInfo || calls != 0 {
		t.Fatalf("got=%#v calls=%d", got, calls)
	}
}

func TestResolveGenderUsesInjectedAIOnlyForConflictAndUnknown(t *testing.T) {
	calls := 0
	ai := func(context.Context, BookInfo) (string, error) {
		calls++
		return `{"gender":"男频","confidence":0.93,"reason":"男主成长线为主"}`, nil
	}
	got := ResolveGender(context.Background(), "", BookInfo{Category: "男频", Genre: "女频"}, ai)
	if got.Gender != "男频" || got.Source != GenderAI || got.Confidence != 0.93 || got.Reason != "男主成长线为主" || calls != 1 {
		t.Fatalf("conflict got=%#v calls=%d", got, calls)
	}
	got = ResolveGender(context.Background(), "", BookInfo{}, ai)
	if got.Gender != "男频" || got.Source != GenderAI || calls != 2 {
		t.Fatalf("unknown got=%#v calls=%d", got, calls)
	}
}

func TestResolveGenderKeepsValidLowConfidenceAIResult(t *testing.T) {
	got := ResolveGender(context.Background(), "", BookInfo{Category: "男频", Genre: "女频"}, func(context.Context, BookInfo) (string, error) {
		return `{"gender":"女频","confidence":0.61,"reason":"线索不充分"}`, nil
	})
	if got.Gender != "女频" || got.Source != GenderAI || got.Confidence != 0.61 {
		t.Fatalf("got=%#v, want valid AI classification with recorded confidence", got)
	}
}
