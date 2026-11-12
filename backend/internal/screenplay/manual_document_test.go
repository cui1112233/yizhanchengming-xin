package screenplay

import (
	"context"
	"testing"
)

func TestSaveManualDocumentCreatesVersionAndKeepsStructuredRows(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	project, err := store.CreateProject(ctx, Project{ID: "manual-project", OwnerUsername: "alice", Title: "漫剧"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.SaveScreenplay(ctx, Document{ID: "doc-1", ProjectID: project.ID, OwnerUsername: "alice", Source: "b", Content: "第一版"},
		[]Shot{{ID: "shot-1", ProjectID: project.ID, OwnerUsername: "alice", EpisodeNo: 1, SortOrder: 1, Heading: "1-1", ScriptText: "镜头"}},
		[]Character{{ID: "character-1", ProjectID: project.ID, OwnerUsername: "alice", Name: "小Y"}},
		[]Scene{{ID: "scene-1", ProjectID: project.ID, OwnerUsername: "alice", Name: "客厅"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	saved, err := SaveManualDocument(ctx, store, "alice", project.ID, "第二版")
	if err != nil {
		t.Fatal(err)
	}
	if saved.Version != 2 || saved.Content != "第二版" || saved.Source != "manual" {
		t.Fatalf("unexpected saved document: %#v", saved)
	}
	if saved.ID == first.ID {
		t.Fatal("manual save must create an immutable document version")
	}
	shots, _ := store.ListShots(ctx, "alice", project.ID, saved.ID)
	characters, _ := store.ListCharacters(ctx, "alice", project.ID, saved.ID)
	scenes, _ := store.ListScenes(ctx, "alice", project.ID, saved.ID)
	if len(shots) != 1 || len(characters) != 1 || len(scenes) != 1 {
		t.Fatalf("structured rows were not retained: shots=%d characters=%d scenes=%d", len(shots), len(characters), len(scenes))
	}
	if shots[0].ID == "shot-1" || shots[0].DocumentID != saved.ID {
		t.Fatalf("shot was not cloned into the new version: %#v", shots[0])
	}
}

func TestRestoreDocumentCreatesTraceableVersion(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	project, err := store.CreateProject(ctx, Project{ID: "restore-project", OwnerUsername: "alice", Title: "漫剧"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.SaveScreenplay(ctx, Document{ID: "doc-old", ProjectID: project.ID, OwnerUsername: "alice", Source: "b", Content: "旧版本"}, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SaveManualDocument(ctx, store, "alice", project.ID, "新版本"); err != nil {
		t.Fatal(err)
	}

	restored, err := RestoreDocument(ctx, store, "alice", project.ID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Version != 3 || restored.Content != "旧版本" || restored.Source != "restore" {
		t.Fatalf("unexpected restored document: %#v", restored)
	}
	active, err := store.ActiveDocument(ctx, "alice", project.ID)
	if err != nil || active.ID != restored.ID {
		t.Fatalf("restored document is not active: %#v, %v", active, err)
	}
}

func TestSaveManualDocumentRejectsEmptyContent(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_, _ = store.CreateProject(ctx, Project{ID: "empty-project", OwnerUsername: "alice", Title: "漫剧"})
	if _, err := SaveManualDocument(ctx, store, "alice", "empty-project", "  "); err == nil {
		t.Fatal("expected empty content error")
	}
}
