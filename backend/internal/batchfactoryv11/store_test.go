package batchfactoryv11

import (
	"context"
	"encoding/json"
	"testing"
)

func TestUpdateBookMetadataCanRenameBook(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{
		Title: "小说 793433",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	updated, err := s.UpdateBookMetadata(ctx, "alice", batch.ID, book.ID, UpdateBookMetadataInput{
		Title:            "她为发小有求必应，我转身离开",
		Metadata:         map[string]any{"sourceBookTitle": "她为发小有求必应，我转身离开"},
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "她为发小有求必应，我转身离开" {
		t.Fatalf("title=%q", updated.Title)
	}
	stale, err := s.UpdateBookMetadata(ctx, "alice", batch.ID, book.ID, UpdateBookMetadataInput{
		Title:            "另一个书名",
		ExpectedRevision: book.Revision,
	})
	if err != ErrConflict {
		t.Fatalf("stale revision should conflict: book=%+v err=%v", stale, err)
	}
	kept, err := s.UpdateBookMetadata(ctx, "alice", batch.ID, book.ID, UpdateBookMetadataInput{
		Metadata:         map[string]any{"style": "悬疑"},
		ExpectedRevision: updated.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if kept.Title != "她为发小有求必应，我转身离开" {
		t.Fatalf("empty title should keep current title: %q", kept.Title)
	}
}

func TestCaptureBookSourceReplacesOnlyGeneratedTitle(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{
		BookID: "7673480334440139800",
		Title:  "小说 7673480334440139800",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]

	updated, err := s.CaptureBookSource(ctx, "alice", batch.ID, book.ID, CaptureBookSourceInput{
		SourceText:       "第一章 正文",
		SourceTitle:      "港岛雨停，再无爱意",
		SourceMetadata:   map[string]any{"sourceBookTitle": "港岛雨停，再无爱意"},
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "港岛雨停，再无爱意" || updated.BookID != "7673480334440139800" {
		t.Fatalf("updated book=%+v", updated)
	}
}

// 书城导入的占位书名「小说 <来源ID>」与内部 bookID 不同，抓正文拿到真实书名后
// 也必须能回填，否则发布阶段永远匹配不到漫画作品。
func TestCaptureBookSourceReplacesPlaceholderTitleWithSourceBookID(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{
		Title: "小说 793433",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	updated, err := s.CaptureBookSource(ctx, "alice", batch.ID, book.ID, CaptureBookSourceInput{
		SourceText:       "第一章 正文",
		SourceBookID:     "793433",
		SourceTitle:      "她为发小有求必应，我转身离开",
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "她为发小有求必应，我转身离开" {
		t.Fatalf("title=%q", updated.Title)
	}
}

func TestCaptureBookSourcePreservesUserTitle(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{
		BookID: "7673480334440139800",
		Title:  "用户手工书名",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]

	updated, err := s.CaptureBookSource(ctx, "alice", batch.ID, book.ID, CaptureBookSourceInput{
		SourceText:       "第一章 正文",
		SourceTitle:      "港岛雨停，再无爱意",
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "用户手工书名" {
		t.Fatalf("user title was overwritten: %+v", updated)
	}
}

func TestCaptureBookSourceKeepsRawTextSeparateFromProcessedText(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	updated, err := s.CaptureBookSource(ctx, "alice", batch.ID, book.ID, CaptureBookSourceInput{
		SourceText:       "处理后的正文",
		RawSourceText:    "上游返回的原始正文\n保留原始换行",
		ExpectedRevision: book.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceText != "处理后的正文" || updated.TxtText != "上游返回的原始正文\n保留原始换行" {
		t.Fatalf("source text separation = source=%q raw=%q", updated.SourceText, updated.TxtText)
	}
}

func TestCaptureBookSourceCanRetainRawOnlyAfterProcessingFailure(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "批量", Books: []CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	updated, err := s.CaptureBookSource(ctx, "alice", batch.ID, book.ID, CaptureBookSourceInput{
		RawSourceText: "完整抓取原文", ReplaceSource: true, ExpectedRevision: book.Revision,
		SourceMetadata: map[string]any{"processingStatus": "failed", "processRetryable": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.SourceText != "" || updated.TxtText != "完整抓取原文" || updated.SourceMetadata["processingStatus"] != "failed" {
		t.Fatalf("failed processing should keep only retryable raw text: %+v", updated)
	}
}

func TestSavingBatchSettingsDoesNotDeleteBookOrVideoOverride(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{Title: "book", Videos: []CreateVideoInput{{Label: "v"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	if _, err = s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID}, SettingsUpdate{Patch: SettingsPatch{"bookOnly": raw(false)}, ExpectedRevision: book.Revision}); err != nil {
		t.Fatal(err)
	}
	updatedBook, _ := s.GetBatch(ctx, "alice", batch.ID)
	book = updatedBook.Books[0]
	video = book.Videos[0]
	if _, err = s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{Patch: SettingsPatch{"videoOnly": raw(0)}, ExpectedRevision: video.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, _ = s.GetBatch(ctx, "alice", batch.ID)
	if _, err = s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"batchOnly": raw("")}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	bookPatch := s.DebugPatch(ScopeRef{Kind: ScopeBook, BatchID: batch.ID, BookID: book.ID})
	videoPatch := s.DebugPatch(ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID})
	if string(bookPatch["bookOnly"]) != "false" || string(videoPatch["videoOnly"]) != "0" {
		t.Fatalf("book=%v video=%v", bookPatch, videoPatch)
	}
}

func TestOptimisticRevisionRejectsStaleWrite(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, _ := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b"})
	if _, err := s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"x": raw(1)}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveSettings(ctx, "alice", ScopeRef{Kind: ScopeBatch, BatchID: batch.ID}, SettingsUpdate{Patch: SettingsPatch{"x": raw(2)}, ExpectedRevision: batch.Revision}); err != ErrConflict {
		t.Fatalf("err=%v", err)
	}
}

func TestAnotherOwnerCannotReadBatch(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, _ := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b"})
	if _, err := s.GetBatch(ctx, "bob", batch.ID); err != ErrNotFound {
		t.Fatalf("err=%v", err)
	}
}

func TestNovelFetchIntakeIsConsumedExactlyOnce(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{{ID: "207", Title: "A"}}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateBatchFromIntake(ctx, "alice", intake.ID, CreateBatchInput{}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateBatchFromIntake(ctx, "alice", intake.ID, CreateBatchInput{}); err != ErrConflict {
		t.Fatalf("err=%v", err)
	}
}

func TestCreateBatchFromIntakePersistsFrozenAutomationSettings(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{{
		ID:             "giant-1",
		Title:          "巨量素材",
		SourceMetadata: map[string]any{"sourceMode": "giant_material"},
	}}})
	if err != nil {
		t.Fatal(err)
	}

	batch, err := s.CreateBatchFromIntake(ctx, "alice", intake.ID, CreateBatchInput{
		InitialBatchSettings: SettingsPatch{
			"textModelId": raw("text-model-a"),
			"automationPresetSnapshot": raw(map[string]any{
				"id":      "preset-1",
				"name":    "全自动预设",
				"version": 3,
			}),
		},
		GiantAutomationPlan: map[string]any{
			"presetId": "preset-1",
			"runMode":  "immediate",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(batch.SettingsState.Patch["textModelId"]); got != `"text-model-a"` {
		t.Fatalf("batch settings=%v", batch.SettingsState.Patch)
	}
	if batch.SettingsState.Revision == 0 || batch.Revision != batch.SettingsState.Revision {
		t.Fatalf("batch revision=%d settings=%+v", batch.Revision, batch.SettingsState)
	}
	plan, ok := batch.Books[0].SourceMetadata["giantAutomationPlan"].(map[string]any)
	if !ok || plan["presetId"] != "preset-1" {
		t.Fatalf("giant plan=%#v", batch.Books[0].SourceMetadata["giantAutomationPlan"])
	}

	loaded, err := s.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(loaded.SettingsState.Patch["automationPresetSnapshot"]); got == "" {
		t.Fatalf("frozen preset missing after reload: %v", loaded.SettingsState.Patch)
	}
}

func TestNovelFetchIntakeAppendsToCurrentBatchAndRetainsSourceBookID(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "当前批量", Books: []CreateBookInput{{BookID: "origin-1", Title: "原书"}}})
	if err != nil {
		t.Fatal(err)
	}
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{{
		BookID: "origin-2", SourceTaskID: "task-2", Title: "AI1·新书", SourceText: "已确认正文", TxtText: "完整原始正文",
		SourceMetadata: map[string]any{"sourceContentVersion": "ai1", "gender": "女频", "genderSource": "bookinfo", "genderReason": "平台分类明确"},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := s.AppendBooksFromIntake(ctx, "alice", batch.ID, intake.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Books) != 2 {
		t.Fatalf("books=%+v", updated.Books)
	}
	got := updated.Books[1]
	if got.BookID != "origin-2" || got.SourceTaskID != "task-2" || got.Title != "AI1·新书" || got.SourceText != "已确认正文" || got.TxtText != "完整原始正文" || got.SourceMetadata["sourceContentVersion"] != "ai1" || got.SourceMetadata["genderSource"] != "bookinfo" {
		t.Fatalf("appended book=%+v", got)
	}
	if _, err := s.AppendBooksFromIntake(ctx, "alice", batch.ID, intake.ID, false); err != ErrConflict {
		t.Fatalf("second append err=%v", err)
	}
}

func TestNovelFetchIntakeSkipsExistingSamePlatformSourceBookID(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, err := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "当前批量", Books: []CreateBookInput{{BookID: "origin-1", Platform: "3", Title: "原书"}}})
	if err != nil {
		t.Fatal(err)
	}
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{BookID: "origin-1", Platform: "3", Title: "不应重复导入", SourceText: "AI 正文"},
		{BookID: "origin-2", Platform: "3", Title: "应当导入的新书", SourceText: "新正文"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := s.AppendBooksFromIntake(ctx, "alice", batch.ID, intake.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Books) != 2 || updated.Books[0].BookID != "origin-1" || updated.Books[1].BookID != "origin-2" || updated.Books[1].Title != "应当导入的新书" {
		t.Fatalf("same-platform duplicate must be skipped while new books append, got %+v", updated.Books)
	}
}

func TestNovelFetchIntakeDeduplicatesSourceBookIDs(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{{ID: "207", Title: "A"}, {ID: "207", Title: "duplicate"}, {ID: "208", Title: "B"}}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 2 || got.Books[0].ID != "207" || got.Books[1].ID != "208" {
		t.Fatalf("books=%+v", got.Books)
	}
}

func TestNovelFetchIntakeKeepsDistinctContentVersionsForTheSameSourceBook(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{BookID: "207", Title: "原书", SourceText: "原文", SourceMetadata: map[string]any{"sourceContentVersion": "original"}},
		{BookID: "207", Title: "AI1·原书", SourceText: "AI 正文", SourceMetadata: map[string]any{"sourceContentVersion": "ai1"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 2 || got.Books[0].Title != "原书" || got.Books[1].Title != "AI1·原书" {
		t.Fatalf("books=%+v", got.Books)
	}
}

func TestNovelFetchIntakeKeepsSameBookIDAcrossPlatforms(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{ID: "737092", Title: "甲", Platform: "3"},
		{ID: "737092", Title: "乙", Platform: "15"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 2 {
		t.Fatalf("cross-platform same ID must keep both books, got %d: %+v", len(got.Books), got.Books)
	}
	if got.Books[0].Platform != "3" || got.Books[1].Platform != "15" || got.Books[0].Title != "甲" || got.Books[1].Title != "乙" {
		t.Fatalf("books=%+v", got.Books)
	}
}

func TestNovelFetchIntakeDeduplicatesSameBookIDWithinPlatform(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{Books: []CreateBookInput{
		{ID: "737092", Title: "甲", Platform: "3"},
		{ID: "737092", Title: "甲重复行", Platform: "3"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	var got NovelFetchIntakeInput
	if err := json.Unmarshal(intake.Payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Books) != 1 || got.Books[0].Title != "甲" {
		t.Fatalf("same-platform duplicate must collapse to the first row, got %+v", got.Books)
	}
}

func TestPromptAndDraftAreOwnerScopedAndDraftRecovers(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	if _, err := s.CreatePrompt(ctx, "alice", Prompt{Name: "p", Kind: "constraint", Content: "alpha"}); err != nil {
		t.Fatal(err)
	}
	if prompts, _ := s.ListPrompts(ctx, "bob", ""); len(prompts) != 0 {
		t.Fatalf("bob prompts=%+v", prompts)
	}
	if _, err := s.SaveDraft(ctx, "alice", Draft{Key: "quality", Kind: "constraint", Scope: "book:1", Content: "draft"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetDraft(ctx, "alice", "quality", "constraint", "book:1")
	if err != nil || got.Content != "draft" || got.Revision != 1 {
		t.Fatalf("draft=%+v err=%v", got, err)
	}
	if _, err := s.GetDraft(ctx, "bob", "quality", "constraint", "book:1"); err != ErrNotFound {
		t.Fatalf("cross-owner err=%v", err)
	}
}

func TestSliceOneChangeImpactDoesNotClaimDirectorInvalidation(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	batch, _ := s.CreateBatch(ctx, "alice", CreateBatchInput{Title: "b", Books: []CreateBookInput{{Title: "book", Videos: []CreateVideoInput{{Label: "v"}}}}})
	impact, err := s.ChangeImpact(ctx, "alice", batch.ID, SettingsUpdate{Patch: SettingsPatch{"modelId": raw("new-model")}})
	if err != nil {
		t.Fatal(err)
	}
	if impact.InvalidatesDirector {
		t.Fatalf("Slice 1 must not claim Director invalidation: %+v", impact)
	}
	if impact.AffectedBooks != 1 || impact.AffectedVideos != 1 || !impact.PreservesOverrides {
		t.Fatalf("impact=%+v", impact)
	}
}

func TestNovelFetchBatchKeepsExplicitSourceBookID(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	intake, err := s.CreateIntake(ctx, "alice", NovelFetchIntakeInput{
		Books: []CreateBookInput{{ID: "source-book-207", Title: "A", SourceText: "原文"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err := s.CreateBatchFromIntake(ctx, "alice", intake.ID, CreateBatchInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Books) != 1 || batch.Books[0].BookID != "source-book-207" {
		t.Fatalf("book=%+v", batch.Books)
	}
}

func TestNovelFetchBatchKeepsSourceLineageFields(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	var input NovelFetchIntakeInput
	if err := json.Unmarshal([]byte(`{"books":[{"id":"book-207","bookId":"book-207","sourceTaskId":"task-207","title":"A","platform":"番茄","sourceText":"原文","txtText":"TXT","txtFileName":"book-207.txt","sourceMetadata":{"platformId":"fanqie"}}]}`), &input); err != nil {
		t.Fatal(err)
	}
	intake, err := s.CreateIntake(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(intake.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	books, ok := payload["books"].([]any)
	if !ok || len(books) != 1 {
		t.Fatalf("payload=%v", payload)
	}
	bookPayload := books[0].(map[string]any)
	if bookPayload["sourceTaskId"] != "task-207" || bookPayload["platform"] != "番茄" || bookPayload["txtText"] != "TXT" {
		t.Fatalf("book payload=%v", bookPayload)
	}
	batch, err := s.CreateBatchFromIntake(ctx, "alice", intake.ID, CreateBatchInput{})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if book.BookID != "book-207" || book.SourceTaskID != "task-207" || book.Platform != "番茄" || book.TxtText != "TXT" || book.TxtFileName != "book-207.txt" {
		t.Fatalf("book=%+v", book)
	}
	if book.SourceMetadata["platformId"] != "fanqie" {
		t.Fatalf("source metadata=%v", book.SourceMetadata)
	}
}
