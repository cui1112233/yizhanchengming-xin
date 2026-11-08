package batchfactoryv11

import (
	"strings"
	"testing"
)

func TestEnumerateComicSourceBoundariesKeepsQuotedDialogueWhole(t *testing.T) {
	source := "他问：\"你真的要走吗？\"她没有回答。\n\"那我等你。\"\n\"不用等了。\"\n天亮以后，他们各自离开。"
	boundaries := EnumerateComicSourceBoundaries(source)
	if len(boundaries) < 3 {
		t.Fatalf("expected several safe boundaries, got %#v", boundaries)
	}
	firstDialogueEnd := strings.Index(source, "她没有回答。") + len("她没有回答。")
	continuousDialogueMiddle := strings.Index(source, "\"不用等了。\"")
	if !hasComicBoundary(boundaries, firstDialogueEnd) {
		t.Fatalf("expected sentence boundary at %d: %#v", firstDialogueEnd, boundaries)
	}
	if hasComicBoundary(boundaries, continuousDialogueMiddle) {
		t.Fatalf("must not split between consecutive dialogue lines: %#v", boundaries)
	}
}

func TestFallbackComicSourceSegmentsTargetsPageRangeAndReconstructsSource(t *testing.T) {
	sentence := "这是一个完整的剧情句子，人物正在推进事件并且留下关键证据。"
	source := strings.Repeat(sentence, 18)
	segments, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) < 5 {
		t.Fatalf("expected page-sized segments, got %d", len(segments))
	}
	var rebuilt strings.Builder
	for index, segment := range segments {
		rebuilt.WriteString(segment.SourceText)
		if index < len(segments)-1 && (segment.CharacterCount < 90 || segment.CharacterCount > 150) {
			t.Fatalf("segment %d has unexpected count %d", index, segment.CharacterCount)
		}
	}
	if rebuilt.String() != source {
		t.Fatal("segments do not reconstruct source exactly")
	}
	if err := ValidateComicSourceSegments(source, segments); err != nil {
		t.Fatalf("expected valid segments: %v", err)
	}
}

func TestFallbackComicSourceSegmentsUsesAboutTwentyPagesForTwoThousandCharacters(t *testing.T) {
	// Each complete sentence is exactly 50 non-space characters. A 2,000
	// character source therefore has 40 safe sentence endings and should group
	// two sentences per page under the 90-120 character comic page contract.
	sentence := strings.Repeat("甲", 49) + "。"
	source := strings.Repeat(sentence, 40)
	segments, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 20 {
		t.Fatalf("2,000 characters produced %d pages, want 20", len(segments))
	}
	for index, segment := range segments {
		if segment.CharacterCount < 90 || segment.CharacterCount > 120 {
			t.Fatalf("segment %d has %d characters, want 90-120", index+1, segment.CharacterCount)
		}
	}
	if err := ValidateComicSourceSegments(source, segments); err != nil {
		t.Fatalf("expected complete source coverage: %v", err)
	}
}

func TestFallbackComicSourceSegmentsKeepsOverlongIndivisibleSentence(t *testing.T) {
	source := strings.Repeat("很长但仍属于同一句话", 20) + "。下一句结束。"
	segments, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected two complete sentences, got %#v", segments)
	}
	if segments[0].CharacterCount <= 150 || segments[0].Reason != ComicSegmentReasonOverlongUnit {
		t.Fatalf("expected documented overlong unit, got %#v", segments[0])
	}
	if segments[0].SourceText != strings.Repeat("很长但仍属于同一句话", 20)+"。" {
		t.Fatal("overlong sentence was cut")
	}
}

func TestValidateComicSourceSegmentsRejectsGapAndMutation(t *testing.T) {
	source := "第一句完整。第二句也完整。"
	segments := []ComicSourceSegmentDraft{
		{StartOffset: 0, EndOffset: len("第一句完整。"), SourceText: "第一句完整。"},
		{StartOffset: len("第一句完整。") + 1, EndOffset: len(source), SourceText: "第二句被改。"},
	}
	if err := ValidateComicSourceSegments(source, segments); err == nil {
		t.Fatal("expected continuity validation failure")
	}
}

func hasComicBoundary(boundaries []ComicSourceBoundary, offset int) bool {
	for _, boundary := range boundaries {
		if boundary.Offset == offset {
			return true
		}
	}
	return false
}
