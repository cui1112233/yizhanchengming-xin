package batchfactoryv11

import (
	"context"
	"strings"
	"testing"
)

func TestComicSourceSegmentStoreReplacesAndReadsOrderedPages(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	drafts, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.ReplaceComicSourceSegments(context.Background(), "alice", batch.ID, batch.Books[0].ID, sourceDigest(source), drafts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != len(drafts) || result.Stale || result.Version != 1 {
		t.Fatalf("unexpected segmentation: %#v", result)
	}
	read, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Books[0].Videos) != len(drafts) || read.Books[0].Videos[0].ComicSourceSegment == nil {
		t.Fatalf("segments were not projected to video pages: %#v", read.Books[0].Videos)
	}
	var rebuilt strings.Builder
	for index, segment := range result.Segments {
		if segment.Ordinal != index || segment.VideoID == "" || len(segment.SafeBoundaries) == 0 {
			t.Fatalf("segment %d missing page identity or boundaries: %#v", index, segment)
		}
		rebuilt.WriteString(segment.SourceText)
	}
	if rebuilt.String() != source {
		t.Fatal("persisted segments do not reconstruct the source")
	}
}

func TestComicSourceSegmentStoreBindsDraftsToExistingPages(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	drafts, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	videos := make([]CreateVideoInput, len(drafts))
	for index := range videos {
		videos[index] = CreateVideoInput{Label: "已有漫画页"}
	}
	batch, err = store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "绑定测试", Books: []CreateBookInput{{Title: "正文", SourceText: source, Videos: videos}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	videoIDs := make([]string, len(book.Videos))
	for index := range book.Videos {
		videoIDs[index] = book.Videos[index].ID
	}
	result, err := store.BindComicSourceSegments(context.Background(), "alice", batch.ID, book.ID, sourceDigest(source), videoIDs, drafts)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != len(videoIDs) {
		t.Fatalf("segments=%d, want %d", len(result.Segments), len(videoIDs))
	}
	for index, segment := range result.Segments {
		if segment.VideoID != videoIDs[index] {
			t.Fatalf("segment %d video=%s, want existing %s", index, segment.VideoID, videoIDs[index])
		}
	}
}

func TestComicSourceSegmentStoreLeftSplitKeepsCurrentIDOnRight(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	segmentation := replaceComicSourceFixture(t, store, batch, source)
	current := segmentation.Segments[0]
	boundary := current.SafeBoundaries[0].Offset
	result, err := store.SplitComicSourceSegment(context.Background(), "alice", batch.ID, batch.Books[0].ID, current.VideoID, ComicSourceSplitInput{Direction: "left", BoundaryOffset: boundary, ExpectedRevision: current.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != len(segmentation.Segments)+1 || result.Segments[1].VideoID != current.VideoID {
		t.Fatalf("left split did not preserve current ID on latter text: %#v", result.Segments[:2])
	}
	if result.Segments[0].SourceText+result.Segments[1].SourceText != current.SourceText {
		t.Fatal("left split changed source text")
	}
	if !result.Segments[0].PromptStale || !result.Segments[1].PromptStale {
		t.Fatal("split pages must mark prompts stale")
	}
}

func TestComicSourceSegmentStoreRightSplitKeepsCurrentIDOnLeft(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	segmentation := replaceComicSourceFixture(t, store, batch, source)
	current := segmentation.Segments[0]
	boundary := current.SafeBoundaries[0].Offset
	result, err := store.SplitComicSourceSegment(context.Background(), "alice", batch.ID, batch.Books[0].ID, current.VideoID, ComicSourceSplitInput{Direction: "right", BoundaryOffset: boundary, ExpectedRevision: current.Revision})
	if err != nil {
		t.Fatal(err)
	}
	if result.Segments[0].VideoID != current.VideoID || result.Segments[0].SourceText+result.Segments[1].SourceText != current.SourceText {
		t.Fatalf("right split did not preserve current ID on first text: %#v", result.Segments[:2])
	}
	before := result
	if _, err := store.SplitComicSourceSegment(context.Background(), "alice", batch.ID, batch.Books[0].ID, current.VideoID, ComicSourceSplitInput{Direction: "right", BoundaryOffset: boundary, ExpectedRevision: current.Revision}); err != ErrConflict {
		t.Fatalf("expected revision conflict, got %v", err)
	}
	after, _ := store.GetComicSourceSegmentation(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if len(after.Segments) != len(before.Segments) {
		t.Fatal("conflicted split changed stored pages")
	}
}

func TestComicSourceSegmentStoreMarksEditedBookSourceStale(t *testing.T) {
	store, batch, source := comicSourceStoreFixture(t)
	replaceComicSourceFixture(t, store, batch, source)
	owned := store.batches[batch.ID]
	owned.Value.Books[0].SourceText += "新增正文。"
	store.batches[batch.ID] = owned
	result, err := store.GetComicSourceSegmentation(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Stale || result.SourceDigest == result.CurrentDigest {
		t.Fatalf("edited source was not marked stale: %#v", result)
	}
}

func comicSourceStoreFixture(t *testing.T) (*MemoryStore, Batch, string) {
	t.Helper()
	store := NewMemoryStore()
	source := strings.Repeat("这是一句完整剧情，角色完成动作并推动冲突。", 16)
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "漫画", Books: []CreateBookInput{{Title: "第一本", SourceText: source}}})
	if err != nil {
		t.Fatal(err)
	}
	return store, batch, source
}

func replaceComicSourceFixture(t *testing.T, store *MemoryStore, batch Batch, source string) ComicSourceSegmentation {
	t.Helper()
	drafts, err := FallbackComicSourceSegments(source, EnumerateComicSourceBoundaries(source))
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.ReplaceComicSourceSegments(context.Background(), "alice", batch.ID, batch.Books[0].ID, sourceDigest(source), drafts)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
