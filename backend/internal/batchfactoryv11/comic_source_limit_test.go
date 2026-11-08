package batchfactoryv11

import (
	"strings"
	"testing"
)

func TestComicSourceTextUsesNearestCompleteBoundaryAfterTargetCharacters(t *testing.T) {
	source := strings.Repeat("甲", 99) + "。" + strings.Repeat("乙", 20) + "。"
	got := comicSourceTextForBook(Book{SourceText: source, SourceMetadata: map[string]any{"comicAnalysisCharacters": 100}})
	if got != strings.Repeat("甲", 99)+"。" {
		t.Fatalf("comic source = %q", got)
	}
}

func TestComicSourceTextKeepsCompleteSourceWhenTargetExceedsSource(t *testing.T) {
	source := "第一段。第二段。"
	if got := comicSourceTextForBook(Book{SourceText: source, SourceMetadata: map[string]any{"comicAnalysisCharacters": 2000}}); got != source {
		t.Fatalf("comic source = %q", got)
	}
}

func TestComicSourceTextUsesAnalysisTargetSeparateFromCaptureLimit(t *testing.T) {
	source := strings.Repeat("甲", 99) + "。" + strings.Repeat("乙", 3900) + "。"
	got := comicSourceTextForBook(Book{SourceText: source, SourceMetadata: map[string]any{
		"comicAnalysisCharacters":  100,
		"contentCaptureCharacters": 4000,
	}})
	if got != strings.Repeat("甲", 99)+"。" {
		t.Fatalf("comic source should use the separate analysis setting, got length=%d", len([]rune(got)))
	}
}

func TestLimitComicSourceSegmentsCapsAtThirtyFiveAndKeepsCompleteSource(t *testing.T) {
	source := strings.Repeat(strings.Repeat("甲", 50)+"。", 72)
	boundaries := EnumerateComicSourceBoundaries(source)
	segments, err := FallbackComicSourceSegments(source, boundaries)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) <= ComicMaxPageCount {
		t.Fatalf("fixture must exceed the cap, got %d segments", len(segments))
	}
	limited, err := LimitComicSourceSegments(source, segments, ComicMaxPageCount)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited) != ComicMaxPageCount {
		t.Fatalf("limited pages=%d, want %d", len(limited), ComicMaxPageCount)
	}
	if err := ValidateComicSourceSegments(source, limited); err != nil {
		t.Fatalf("limited segments must cover complete source: %v", err)
	}
	for _, segment := range limited {
		if !strings.HasSuffix(segment.SourceText, "。") {
			t.Fatalf("segment cut a sentence: %q", segment.SourceText)
		}
	}
}
