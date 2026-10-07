package agentstudio

import (
	"context"
	"errors"
	"testing"
)

func TestCanvasServicePersistsRevisionsAndRestoresWithoutOverwritingHistory(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, UnavailableExecutor{})
	actor := Actor{UserID: 7}
	project, err := service.CreateProject(context.Background(), actor, CreateProjectInput{Title: "Canvas"})
	if err != nil {
		t.Fatal(err)
	}

	initial, err := service.GetCanvas(context.Background(), actor, project.ID)
	if err != nil || initial.Revision != 0 {
		t.Fatalf("initial=%+v err=%v", initial, err)
	}
	first, err := service.SaveCanvas(context.Background(), actor, Canvas{ProjectID: project.ID, Revision: initial.Revision, Document: []byte(`{"nodes":["one"],"edges":[]}`)})
	if err != nil || first.Revision != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if _, err = service.SaveCanvas(context.Background(), actor, Canvas{ProjectID: project.ID, Revision: 0, Document: []byte(`{"nodes":["stale"]}`)}); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale save error=%v, want conflict", err)
	}
	second, err := service.SaveCanvas(context.Background(), actor, Canvas{ProjectID: project.ID, Revision: first.Revision, Document: []byte(`{"nodes":["two"],"edges":[]}`)})
	if err != nil || second.Revision != 2 {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	versions, err := service.ListCanvasVersions(context.Background(), actor, project.ID)
	if err != nil || len(versions) != 2 {
		t.Fatalf("versions=%+v err=%v", versions, err)
	}
	restored, err := service.RestoreCanvas(context.Background(), actor, project.ID, 1)
	if err != nil || restored.Revision != 3 || string(restored.Document) != `{"nodes":["one"],"edges":[]}` {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
}
