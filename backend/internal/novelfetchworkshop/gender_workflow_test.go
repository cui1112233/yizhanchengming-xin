package novelfetchworkshop

import (
	"context"
	"testing"
)

func TestResolveGenderManualWinsOverBookInfoAndAI(t *testing.T) {
	calls := 0
	got := ResolveGender(context.Background(), "女频", BookInfo{Category: "男频", Genre: "男性向"}, func(context.Context, BookInfo) (string, error) {
		calls++
		return "男频", nil
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
	ai := func(context.Context, BookInfo) (string, error) { calls++; return "男频", nil }
	got := ResolveGender(context.Background(), "", BookInfo{Category: "男频", Genre: "女频"}, ai)
	if got.Gender != "男频" || got.Source != GenderAI || calls != 1 {
		t.Fatalf("conflict got=%#v calls=%d", got, calls)
	}
	got = ResolveGender(context.Background(), "", BookInfo{}, ai)
	if got.Gender != "男频" || got.Source != GenderAI || calls != 2 {
		t.Fatalf("unknown got=%#v calls=%d", got, calls)
	}
}

func TestResolveGenderAIUnableReturnsPendingConfirmation(t *testing.T) {
	got := ResolveGender(context.Background(), "", BookInfo{Category: "男频", Genre: "女频"}, func(context.Context, BookInfo) (string, error) {
		return "", nil
	})
	if got.Gender != "待确认" || got.Source != GenderUnknown {
		t.Fatalf("got=%#v", got)
	}
}
