package batchfactoryv11

import (
	"context"
	"testing"
)

func TestGetBatchReturnsSettingsStateForEveryScope(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{
		Title: "batch",
		Books: []CreateBookInput{{Title: "book", Videos: []CreateVideoInput{{Label: "video"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := s.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.SettingsState.Revision != 1 || len(got.SettingsState.Patch) != 0 {
		t.Fatalf("batch settingsState=%+v", got.SettingsState)
	}
	if got.Books[0].SettingsState.Revision != 1 || len(got.Books[0].SettingsState.Patch) != 0 {
		t.Fatalf("book settingsState=%+v", got.Books[0].SettingsState)
	}
	if got.Books[0].Videos[0].SettingsState.Revision != 1 || len(got.Books[0].Videos[0].SettingsState.Patch) != 0 {
		t.Fatalf("video settingsState=%+v", got.Books[0].Videos[0].SettingsState)
	}
}

func TestGetBatchReadsBackSparsePatchesAndCurrentRevisions(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{
		Title: "batch",
		Books: []CreateBookInput{{Title: "book", Videos: []CreateVideoInput{{Label: "video"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]

	if _, err := s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{
		Patch: SettingsPatch{"enabled": raw(false)}, ExpectedRevision: batch.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{
		Patch: SettingsPatch{"empty": raw("")}, ExpectedRevision: book.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"zero": raw(0)}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.SettingsState.Revision != 2 || string(got.SettingsState.Patch["enabled"]) != "false" {
		t.Fatalf("batch settingsState=%+v", got.SettingsState)
	}
	if got.Books[0].SettingsState.Revision != 2 || string(got.Books[0].SettingsState.Patch["empty"]) != `""` {
		t.Fatalf("book settingsState=%+v", got.Books[0].SettingsState)
	}
	if got.Books[0].Videos[0].SettingsState.Revision != 2 || string(got.Books[0].Videos[0].SettingsState.Patch["zero"]) != "0" {
		t.Fatalf("video settingsState=%+v", got.Books[0].Videos[0].SettingsState)
	}
}
