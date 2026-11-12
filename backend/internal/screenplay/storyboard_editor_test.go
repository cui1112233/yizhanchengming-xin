package screenplay

import (
	"context"
	"errors"
	"testing"
)

func seedStoryboardEditorTest(t *testing.T) (*MemoryStore, Storyboard) {
	t.Helper()
	store := NewMemoryStore()
	row, err := store.SaveStoryboard(context.Background(), Storyboard{
		ProjectID: "comic-1", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{
			{ID: "sb-shot-1", Seq: 1, Duration: 2, Characters: "小Y", Scene: "客厅", Props: "手机", VideoDesc: "小Y拿起手机"},
			{ID: "sb-shot-2", Seq: 2, Duration: 4, Characters: "CM", Scene: "走廊", Props: "", VideoDesc: "CM走进走廊"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, row
}

func TestStoryboardShotEditorCreateAfterAndUpdate(t *testing.T) {
	ctx := context.Background()
	store, _ := seedStoryboardEditorTest(t)

	created, err := CreateStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-1", StoryboardShot{
		Duration: 3, Characters: "小Y、CM", Scene: "门口", Props: "钥匙", VideoDesc: "两人在门口会面",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Shots) != 3 || created.Shots[1].VideoDesc != "两人在门口会面" || created.Shots[1].Seq != 2 || created.Shots[2].Seq != 3 {
		t.Fatalf("created shots = %+v", created.Shots)
	}
	newID := created.Shots[1].ID
	if newID == "" || newID == "sb-shot-1" || newID == "sb-shot-2" {
		t.Fatalf("new shot id = %q", newID)
	}

	updated, err := UpdateStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, newID, StoryboardShot{
		Duration: 5, Characters: "小Y", Scene: "玄关", Props: "雨伞", VideoDesc: "小Y收起雨伞",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Shots[1].ID != newID || updated.Shots[1].Duration != 5 || updated.Shots[1].Scene != "玄关" || updated.Revision <= created.Revision {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestStoryboardShotEditorSplitMergeDeleteAndPersist(t *testing.T) {
	ctx := context.Background()
	store, _ := seedStoryboardEditorTest(t)

	split, err := SplitStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(split.Shots) != 3 || split.Shots[1].Duration != 2 || split.Shots[2].Duration != 2 || split.Shots[2].ID == "sb-shot-2" {
		t.Fatalf("split shots = %+v", split.Shots)
	}

	merged, err := MergeStoryboardShotWithNext(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Shots) != 2 || merged.Shots[0].Duration != 4 || merged.Shots[0].VideoDesc != "小Y拿起手机\nCM走进走廊" || merged.Shots[0].Seq != 1 || merged.Shots[1].Seq != 2 {
		t.Fatalf("merged shots = %+v", merged.Shots)
	}

	deleted, err := DeleteStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, merged.Shots[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Shots) != 1 || deleted.Shots[0].Seq != 1 {
		t.Fatalf("deleted shots = %+v", deleted.Shots)
	}
	persisted, err := store.GetStoryboard(ctx, "alice", "comic-1", 1, EngineXiaoluoV7)
	if err != nil || len(persisted.Shots) != 1 || persisted.Revision != deleted.Revision {
		t.Fatalf("persisted = %+v err=%v", persisted, err)
	}
}

func TestStoryboardShotEditorRejectsInvalidBoundaries(t *testing.T) {
	ctx := context.Background()
	store, _ := seedStoryboardEditorTest(t)

	if _, err := UpdateStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "missing", StoryboardShot{Duration: 1, VideoDesc: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update missing err = %v", err)
	}
	if _, err := MergeStoryboardShotWithNext(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-2"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("merge last err = %v", err)
	}
	short, err := UpdateStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "sb-shot-1", StoryboardShot{Duration: 0.4, VideoDesc: "短镜头"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = SplitStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, short.Shots[0].ID); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("split short err = %v", err)
	}
	if _, err = CreateStoryboardShot(ctx, store, "alice", "comic-1", 1, EngineXiaoluoV7, "missing", StoryboardShot{Duration: 1, VideoDesc: "x"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("create after missing err = %v", err)
	}
}
