package workshop

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type memoryStore struct{ value json.RawMessage }

func (s *memoryStore) Load(context.Context, int64) (json.RawMessage, error) { return s.value, nil }
func (s *memoryStore) Save(_ context.Context, _ int64, value json.RawMessage) (json.RawMessage, error) {
	s.value = value
	return value, nil
}

type memoryIntakes struct{}

func (memoryIntakes) GetIntake(context.Context, int64) (intake.Intake, error) {
	return intake.Intake{ID: 7, Name: "恢复任务", Status: intake.StatusPartial}, nil
}
func (memoryIntakes) ListBooks(context.Context, int64) ([]intake.Book, error) {
	return []intake.Book{{ID: 11, IntakeID: 7, Title: "已恢复", OriginalText: "服务端原文", Status: intake.BookStatusFetched}}, nil
}

type memoryPrompts struct{}

func (memoryPrompts) ListPrompts(context.Context) ([]generation.Prompt, error) {
	return []generation.Prompt{{Key: generation.PromptScript, Version: 2, Enabled: true}}, nil
}

func TestSnapshotAndSaveUseExistingIntakeFacts(t *testing.T) {
	store := &memoryStore{value: json.RawMessage(`{"processingRules":"old"}`)}
	service := NewService(store, memoryIntakes{}, memoryPrompts{})
	snapshot, err := service.Snapshot(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Intake.ID != 7 || len(snapshot.Books) != 1 || snapshot.Books[0].OriginalText != "服务端原文" || len(snapshot.Prompts) != 1 {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	saved, err := service.Save(context.Background(), 7, json.RawMessage(`{"knowledgeBase":"kb"}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(saved) != "{\"knowledgeBase\":\"kb\"}" {
		t.Fatalf("saved=%s", saved)
	}
	if _, err := service.Save(context.Background(), 7, json.RawMessage(`[]`)); err == nil {
		t.Fatal("array settings must be rejected")
	}
}
