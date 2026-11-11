package novelcreation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestMemoryStoreCreateProjectAndSessionAreOwnerScoped(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	project, err := store.CreateProject(ctx, "alice", CreateProjectInput{
		Title: "迷雾记者", Genre: "都市悬疑", ModelID: "text-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetProject(ctx, "bob", project.ID); !IsCode(err, ErrProjectNotFound) {
		t.Fatalf("cross-owner project read err=%v", err)
	}
	session, err := store.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: project.ID, Mode: "book-create"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetSession(ctx, "bob", session.ID); !IsCode(err, ErrSessionNotFound) {
		t.Fatalf("cross-owner session read err=%v", err)
	}
}

func TestMemoryStoreKeepsRunIdempotencyEventsAndChapterBundleOwnerScoped(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	project, err := store.CreateProject(ctx, "alice", CreateProjectInput{Title: "迷雾记者", Genre: "都市悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: project.ID, Mode: "book-create"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := store.CreateRun(ctx, "alice", Run{ID: "run-1", ProjectID: project.ID, SessionID: session.ID, IdempotencyKey: "send-1", Status: RunQueued})
	if err != nil {
		t.Fatal(err)
	}
	found, ok, err := store.FindRunByIdempotencyKey(ctx, "alice", "send-1")
	if err != nil || !ok || found.ID != run.ID {
		t.Fatalf("found=%+v ok=%v err=%v", found, ok, err)
	}
	if _, ok, err := store.FindRunByIdempotencyKey(ctx, "bob", "send-1"); err != nil || ok {
		t.Fatalf("cross-owner run found=%v err=%v", ok, err)
	}
	if _, err := store.CreateRun(ctx, "alice", Run{ID: "run-2", ProjectID: project.ID, SessionID: session.ID, IdempotencyKey: "send-1", Status: RunQueued}); !IsCode(err, ErrRunInProgress) {
		t.Fatalf("duplicate idempotency err=%v", err)
	}
	firstMessage, err := store.AppendMessage(ctx, "alice", Message{SessionID: session.ID, ProjectID: project.ID, Role: "user", Content: "先写提纲"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendMessage(ctx, "alice", Message{SessionID: session.ID, ProjectID: project.ID, Role: "assistant", Content: "已开始规划"}); err != nil {
		t.Fatal(err)
	}
	messages, err := store.ListMessages(ctx, "alice", session.ID, firstMessage.Sequence)
	if err != nil || len(messages) != 1 || messages[0].Sequence != firstMessage.Sequence+1 {
		t.Fatalf("messages after cursor=%+v err=%v", messages, err)
	}
	event, err := store.AppendRunEvent(ctx, "alice", RunEvent{RunID: run.ID, SessionID: session.ID, ProjectID: project.ID, Type: "status", Status: RunPlanning, Message: "正在规划"})
	if err != nil {
		t.Fatal(err)
	}
	events, err := store.ListRunEvents(ctx, "alice", run.ID, 0)
	if err != nil || len(events) != 1 || events[0].EventID != event.EventID {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	if err := store.SaveChapterBundle(ctx, "alice", ChapterBundle{ProjectID: project.ID, RunID: run.ID, Chapter: Chapter{ID: "chapter-1", Number: 1, Title: "雨夜", Content: "雨落在旧城。"}}); err != nil {
		t.Fatal(err)
	}
	if saved, ok := store.chapterBundles[memoryKey("alice", project.ID, "chapter-1")]; !ok || saved.Chapter.Content != "雨落在旧城。" {
		t.Fatalf("chapter bundle not persisted: %+v ok=%v", saved, ok)
	}
}

func TestMemoryStoreRejectsCrossProjectBindingsForSameOwner(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	first, err := store.CreateProject(ctx, "alice", CreateProjectInput{Title: "甲", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateProject(ctx, "alice", CreateProjectInput{Title: "乙", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	firstSession, err := store.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: first.ID, Mode: "book-create"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendMessage(ctx, "alice", Message{SessionID: firstSession.ID, ProjectID: second.ID, Role: "user", Content: "错误绑定"}); !IsCode(err, ErrSessionNotFound) {
		t.Fatalf("cross-project message err=%v", err)
	}
	if _, err := store.CreateRun(ctx, "alice", Run{ProjectID: second.ID, SessionID: firstSession.ID, IdempotencyKey: "cross-project", Status: RunQueued}); !IsCode(err, ErrSessionNotFound) {
		t.Fatalf("cross-project run err=%v", err)
	}
}

func TestMemoryStoreRejectsInvalidJSONPayloads(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	project, err := store.CreateProject(ctx, "alice", CreateProjectInput{Title: "迷雾记者", Genre: "都市悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: project.ID, Mode: "book-create"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateRun(ctx, "alice", Run{ProjectID: project.ID, SessionID: session.ID, IdempotencyKey: "invalid-json", Status: RunQueued, PromptSnapshot: json.RawMessage(`{`)}); !IsCode(err, ErrPersistenceFailed) {
		t.Fatalf("invalid run JSON err=%v", err)
	}
	run, err := store.CreateRun(ctx, "alice", Run{ProjectID: project.ID, SessionID: session.ID, IdempotencyKey: "valid-json", Status: RunQueued})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		call func() error
	}{
		{name: "message", call: func() error {
			_, err := store.AppendMessage(ctx, "alice", Message{SessionID: session.ID, ProjectID: project.ID, Role: "user", Content: "继续", References: json.RawMessage(`{`)})
			return err
		}},
		{name: "run update", call: func() error {
			_, err := store.UpdateRun(ctx, "alice", RunUpdate{ID: run.ID, RuntimeSnapshot: json.RawMessage(`{`)})
			return err
		}},
		{name: "event", call: func() error {
			_, err := store.AppendRunEvent(ctx, "alice", RunEvent{RunID: run.ID, SessionID: session.ID, ProjectID: project.ID, Type: "status", Payload: json.RawMessage(`{`)})
			return err
		}},
		{name: "chapter bundle", call: func() error {
			return store.SaveChapterBundle(ctx, "alice", ChapterBundle{ProjectID: project.ID, RunID: run.ID, Chapter: Chapter{ID: "chapter-invalid-json", Number: 1}, Foundation: Foundation{Snapshot: json.RawMessage(`{`)}})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); !IsCode(err, ErrPersistenceFailed) {
				t.Fatalf("invalid JSON err=%v", err)
			}
		})
	}
}
