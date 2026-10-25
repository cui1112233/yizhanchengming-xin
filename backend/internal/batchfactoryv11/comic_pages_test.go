package batchfactoryv11

import (
	"context"
	"errors"
	"testing"
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
