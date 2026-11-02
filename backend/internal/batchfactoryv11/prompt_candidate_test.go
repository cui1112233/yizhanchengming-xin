package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestVideoReadbackKeepsAIOriginalPromptSeparateFromCurrentPrompt(t *testing.T) {
	video := Video{
		VideoPrompt: "用户正在使用的提示词",
		SettingsState: SettingsState{Patch: SettingsPatch{
			"videoPrompt":           raw("用户保存后的提示词"),
			"aiOriginalVideoPrompt": raw("AI 最近确认的原始提示词"),
		}},
	}

	applyVideoPromptOverrides(&video)
	encoded, err := json.Marshal(video)
	if err != nil {
		t.Fatal(err)
	}
	var readback struct {
		VideoPrompt           string `json:"videoPrompt"`
		AIOriginalVideoPrompt string `json:"aiOriginalVideoPrompt"`
	}
	if err := json.Unmarshal(encoded, &readback); err != nil {
		t.Fatal(err)
	}
	if readback.VideoPrompt != "用户保存后的提示词" {
		t.Fatalf("current prompt=%q", readback.VideoPrompt)
	}
	if readback.AIOriginalVideoPrompt != "AI 最近确认的原始提示词" {
		t.Fatalf("AI original prompt=%q", readback.AIOriginalVideoPrompt)
	}
}

func TestVideoPromptEditDurablyInitializesMissingAIOriginalFromPreEditPrompt(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{
		Title: "batch",
		Books: []CreateBookInput{{
			Title:  "book",
			Videos: []CreateVideoInput{{Label: "VIDEO 01", VideoPrompt: "初始 AI 提示词"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	if _, err := store.SaveSettings(context.Background(), "alice", ScopeRef{Kind: ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, SettingsUpdate{
		Patch: SettingsPatch{"videoPrompt": raw("用户编辑后的提示词")}, ExpectedRevision: video.Revision,
	}); err != nil {
		t.Fatal(err)
	}

	reloaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Books[0].Videos[0]
	if got.VideoPrompt != "用户编辑后的提示词" || got.AIOriginalVideoPrompt != "初始 AI 提示词" {
		t.Fatalf("reloaded video=%+v", got)
	}
	if original, ok := optionalPromptValue(got.SettingsState.Patch, "aiOriginalVideoPrompt"); !ok || original != "初始 AI 提示词" {
		t.Fatalf("original patch=%s", got.SettingsState.Patch["aiOriginalVideoPrompt"])
	}

	reloadedAgain, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloadedAgain.Books[0].Videos[0].AIOriginalVideoPrompt; got != "初始 AI 提示词" {
		t.Fatalf("original did not survive reload: %q", got)
	}
}

func TestCandidateAcceptConflictDoesNotPartiallyPersistSelectedVideo(t *testing.T) {
	base := NewMemoryStore()
	store := &candidateAcceptConflictStore{MemoryStore: base}
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{
		Title: "batch",
		Books: []CreateBookInput{{
			Title: "book", SourceText: "林晚进入客厅。",
			Videos: []CreateVideoInput{{Label: "VIDEO 01", VideoPrompt: "原来的当前提示词"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	service := &DirectorService{Store: store, Provider: &queuedDirectorProvider{values: []string{validDirectorJSON()}}}
	candidate, err := service.CreateDirectorPromptCandidate(context.Background(), "alice", batch.ID, book.ID, book.Revision)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	video := loaded.Books[0].Videos[0]
	store.rejectCandidateBookUpdate = true

	_, err = service.AcceptDirectorPromptCandidate(context.Background(), "alice", batch.ID, book.ID, video.ID, candidate.ID, video.Revision)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("accept error=%v, want revision conflict", err)
	}
	reloaded, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Books[0].Videos[0]
	if got.VideoPrompt != "原来的当前提示词" || got.AIOriginalVideoPrompt != "" || got.Revision != video.Revision {
		t.Fatalf("failed accept partially persisted video: %+v", got)
	}
	if _, err := service.DirectorPromptCandidate(context.Background(), "alice", batch.ID, book.ID); err != nil {
		t.Fatalf("candidate was changed by failed accept: %v", err)
	}
}

func TestCreateDirectorPromptCandidateRetriesToExistingVideoCount(t *testing.T) {
	store := NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", CreateBatchInput{
		Title: "batch",
		Books: []CreateBookInput{{
			Title: "book", SourceText: "林晚进入客厅并放下玻璃杯。",
			Videos: []CreateVideoInput{
				{Label: "VIDEO 01", VideoPrompt: "原来的第一段提示词"},
				{Label: "VIDEO 02", VideoPrompt: "原来的第二段提示词"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	provider := &queuedDirectorProvider{values: []string{validDirectorJSON(), twoDirectorVideosJSON(t)}}
	service := &DirectorService{Store: store, Provider: provider}

	candidate, err := service.CreateDirectorPromptCandidate(context.Background(), "alice", batch.ID, book.ID, book.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("provider calls=%d, want 2 after storyboard-count retry", len(provider.calls))
	}
	if !strings.Contains(provider.calls[0].SystemPrompt, "恰好 2 个 VIDEO") || !strings.Contains(provider.calls[0].UserPrompt, `"existing_video_count": 2`) {
		t.Fatalf("candidate count constraint missing from first request:\n%s\n%s", provider.calls[0].SystemPrompt, provider.calls[0].UserPrompt)
	}
	if !strings.Contains(provider.calls[1].SystemPrompt, "上一轮返回了 1 个 storyboard") {
		t.Fatalf("count-retry instruction missing:\n%s", provider.calls[1].SystemPrompt)
	}
	if len(candidate.VideoPrompts) != 2 {
		t.Fatalf("candidate prompts=%d, want 2", len(candidate.VideoPrompts))
	}
}

type candidateAcceptConflictStore struct {
	*MemoryStore
	rejectCandidateBookUpdate bool
}

func (s *candidateAcceptConflictStore) SaveSettings(ctx context.Context, owner string, ref ScopeRef, update SettingsUpdate) (SettingsResult, error) {
	if s.rejectCandidateBookUpdate && ref.Kind == ScopeBook && ref.BookID != "" {
		if _, updating := update.Patch["directorPromptCandidate"]; updating || len(update.RestoreKeys) > 0 {
			return SettingsResult{}, ErrConflict
		}
	}
	return s.MemoryStore.SaveSettings(ctx, owner, ref, update)
}

// The future atomic repository operation reports its conflict before changing
// either scope. The current two-write service ignores this method and exposes
// the partial-write bug through SaveSettings above.
func (s *candidateAcceptConflictStore) AcceptDirectorPromptCandidate(context.Context, string, string, string, string, string, int64) error {
	return ErrConflict
}
