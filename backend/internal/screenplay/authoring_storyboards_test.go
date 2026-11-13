package screenplay

import (
	"context"
	"errors"
	"testing"
)

func TestAuthoringStoryboardBatchConflictAndMetadata(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice"})
	b, _ := s.SaveStoryboard(ctx, Storyboard{ProjectID: "p", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7, Shots: []StoryboardShot{{ID: "one", Duration: 4, VideoDesc: "一"}, {ID: "two", Duration: 4, VideoDesc: "二"}}})
	_, e := EditStoryboardBatch(ctx, s, "alice", "p", 1, EngineXiaoluoV7, StoryboardBatchEdit{ExpectedRevision: 0, Operation: "delete", ShotIDs: []string{"one"}})
	if !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	_, e = EditStoryboardBatch(ctx, s, "alice", "p", 1, EngineXiaoluoV7, StoryboardBatchEdit{ExpectedRevision: b.Revision, Operation: "duration", ShotIDs: []string{"one"}, Duration: -1})
	if !errors.Is(e, ErrInvalidArgument) {
		t.Fatal(e)
	}
	out, e := EditStoryboardBatch(ctx, s, "alice", "p", 1, EngineXiaoluoV7, StoryboardBatchEdit{ExpectedRevision: b.Revision, Operation: "duplicate", ShotIDs: []string{"one"}})
	if e != nil || len(out.Shots) != 3 || out.Shots[1].ID == "one" || out.Revision != b.Revision+1 {
		t.Fatal(out, e)
	}
	_, e = s.SaveStoryboardRevision(ctx, b, b.Revision)
	if !errors.Is(e, ErrRevisionConflict) {
		t.Fatal("CAS", e)
	}
	_, e = EditStoryboardBatch(ctx, s, "mallory", "p", 1, EngineXiaoluoV7, StoryboardBatchEdit{ExpectedRevision: out.Revision, Operation: "delete", ShotIDs: []string{"one"}})
	if !errors.Is(e, ErrNotFound) {
		t.Fatal("foreign", e)
	}
}
