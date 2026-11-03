package batchfactoryv11

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestComicPageImagesKeepCandidatesAndExactlyOnePrimary(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	video := book.Videos[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "四格竖版漫画页",
		ImagePrompt: "可直接生成的四格漫画页",
		ModelID:     "seedream-4.5",
		AspectRatio: "3:4",
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, CreateComicPageImageInput{
		StorageRef: "lea_first.png", MediaType: "image/png", Source: "doubao_local_executor",
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, CreateComicPageImageInput{
		StorageRef: "lea_second.png", MediaType: "image/png", Source: "doubao_local_executor",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !first.IsPrimary || second.IsPrimary {
		t.Fatalf("initial primary state first=%t second=%t", first.IsPrimary, second.IsPrimary)
	}
	if _, err := store.SetPrimaryComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	pages, err := store.ListComicPages(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || len(pages[0].Images) != 2 {
		t.Fatalf("pages=%+v", pages)
	}
	if pages[0].Status != "image_ready" {
		t.Fatalf("page status=%q want image_ready", pages[0].Status)
	}
	primaryCount := 0
	for _, image := range pages[0].Images {
		if image.IsPrimary {
			primaryCount++
			if image.ID != second.ID {
				t.Fatalf("primary image=%q want=%q", image.ID, second.ID)
			}
		}
	}
	if primaryCount != 1 {
		t.Fatalf("primary count=%d", primaryCount)
	}
}

func TestComicPageUpsertHonorsExpectedRevision(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	video := book.Videos[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "出图",
	})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "generating", ExpectedRevision: page.Revision,
	})
	if err != nil || updated.Revision != page.Revision+1 {
		t.Fatalf("conditional upsert revision=%d err=%v, want %d", updated.Revision, err, page.Revision+1)
	}
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "generation_failed", ExpectedRevision: page.Revision,
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale revision write err=%v, want ErrConflict", err)
	}
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "generation_failed", ExpectedRevision: 999,
	}); !errors.Is(err, ErrConflict) {
		t.Fatalf("future revision write err=%v, want ErrConflict", err)
	}
	// 不带 ExpectedRevision 的调用保持原来的覆盖语义。
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "image_ready",
	}); err != nil {
		t.Fatalf("unconditional upsert err=%v", err)
	}
}

func TestComicPageReferenceImagesPersistAndApplyToEveryExistingPage(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{Title: "B", Books: []CreateBookInput{{Title: "K", SourceText: "正文", Videos: []CreateVideoInput{{Label: "一"}, {Label: "二"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	firstVideo, secondVideo := book.Videos[0], book.Videos[1]
	for _, video := range []Video{firstVideo, secondVideo} {
		if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
			ComicPrompt: "漫画页", ImagePrompt: "可直接生成的漫画页",
		}); err != nil {
			t.Fatal(err)
		}
	}
	references := []string{"/api/local-executor/v1/batch-comic-references/" + batch.ID + "/reference-a"}
	page, err := store.SetComicPageReferenceImages(context.Background(), "alice", batch.ID, book.ID, firstVideo.ID, references)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.ReferenceImageURLs) != 1 || page.ReferenceImageURLs[0] != references[0] {
		t.Fatalf("saved references=%v", page.ReferenceImageURLs)
	}
	updated, err := store.ApplyComicPageReferenceImages(context.Background(), "alice", batch.ID, book.ID, firstVideo.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 2 {
		t.Fatalf("updated=%d want 2", updated)
	}
	pages, err := store.ListComicPages(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || len(pages[1].ReferenceImageURLs) != 1 || pages[1].ReferenceImageURLs[0] != references[0] {
		t.Fatalf("pages=%+v", pages)
	}
}

func TestComicCollectionRejectsEveryStoryboardMissingPrimary(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	missingVideoIDs := make([]string, 0, len(book.Videos))
	for _, video := range book.Videos {
		missingVideoIDs = append(missingVideoIDs, video.ID)
	}
	_, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("error=%v", err)
	}
	if len(missing) != len(missingVideoIDs) {
		t.Fatalf("missing=%v want=%v", missing, missingVideoIDs)
	}
	for index := range missingVideoIDs {
		if missing[index] != missingVideoIDs[index] {
			t.Fatalf("missing=%v want=%v", missing, missingVideoIDs)
		}
	}
}

func TestComicCollectionReusesMatchingOrderedPrimaryImages(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	video := book.Videos[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{
		ComicPrompt: "漫画页", ImagePrompt: "漫画页图片",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, CreateComicPageImageInput{
		StorageRef: "primary.png", MediaType: "image/png", Source: "doubao_local_executor",
	}); err != nil {
		t.Fatal(err)
	}

	first, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID, []string{video.ID})
	if err != nil || len(missing) != 0 {
		t.Fatalf("first collection=%#v missing=%v err=%v", first, missing, err)
	}
	second, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID, []string{video.ID})
	if err != nil || len(missing) != 0 {
		t.Fatalf("second collection=%#v missing=%v err=%v", second, missing, err)
	}
	if second.ID != first.ID {
		t.Fatalf("duplicate range created a second collection: first=%q second=%q", first.ID, second.ID)
	}
	collections, err := store.ListComicCollections(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(collections) != 1 {
		t.Fatalf("collections=%#v err=%v, want one collection", collections, err)
	}
}

func TestComicCollectionListKeepsOnlyLatestMatchingPrimaryImages(t *testing.T) {
	store, batch, book := seedDirectorBook(t, "original", false)
	video := book.Videos[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "漫画页图片"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, CreateComicPageImageInput{StorageRef: "primary.png", MediaType: "image/png", Source: "doubao_local_executor"}); err != nil {
		t.Fatal(err)
	}
	first, _, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID, []string{video.ID})
	if err != nil {
		t.Fatal(err)
	}
	duplicate := first
	duplicate.ID = "comic-collection-latest"
	duplicate.CreatedAt = first.CreatedAt.Add(time.Second)
	store.comicCollections[duplicate.ID] = memoryOwned[ComicCollection]{Owner: "alice", Value: duplicate}

	collections, err := store.ListComicCollections(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(collections) != 1 || collections[0].ID != duplicate.ID {
		t.Fatalf("collections=%#v err=%v, want only latest duplicate", collections, err)
	}
}
