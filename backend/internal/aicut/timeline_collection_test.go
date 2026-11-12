package aicut

import (
	"context"
	"errors"
	"testing"
	"time"
)

func collectionFixture(t *testing.T) (context.Context, *MemoryStore, *TimelineCollectionStore) {
	t.Helper()
	ctx := context.Background()
	b := NewMemoryStore()
	if e := b.CreateProject(ctx, "alice", Project{ID: "p", Title: "P", CreatedAt: time.Now(), UpdatedAt: time.Now()}); e != nil {
		t.Fatal(e)
	}
	return ctx, b, NewMemoryTimelineCollectionStore(b)
}
func TestTimelineCollectionIsolationHistory(t *testing.T) {
	ctx, b, s := collectionFixture(t)
	a, e := s.Create(ctx, "alice", "p", "A")
	if e != nil {
		t.Fatal(e)
	}
	other, _ := s.Create(ctx, "alice", "p", "B")
	v, e := s.Save(ctx, "alice", "p", a.ID, Timeline{Script: "first"}, 0, time.Now())
	if e != nil || v.Revision != 1 {
		t.Fatalf("save=%+v %v", v, e)
	}
	if _, e = s.Save(ctx, "alice", "p", a.ID, Timeline{}, 0, time.Now()); !errors.Is(e, ErrRevisionConflict) {
		t.Fatalf("conflict=%v", e)
	}
	for _, id := range []string{DefaultTimelineID, other.ID} {
		v, e = s.Timeline(ctx, "alice", "p", id)
		if e != nil || v.Script != "" || v.Revision != 0 {
			t.Fatalf("isolation=%+v %v", v, e)
		}
	}
	v, h, e := s.Undo(ctx, "alice", "p", a.ID, 1, time.Now())
	if e != nil || v.Script != "" || v.Revision != 2 || !h.CanRedo {
		t.Fatalf("undo=%+v %+v %v", v, h, e)
	}
	v, _, e = s.Redo(ctx, "alice", "p", a.ID, 2, time.Now())
	if e != nil || v.Script != "first" || v.Revision != 3 {
		t.Fatalf("redo=%+v %v", v, e)
	}
	legacy, _ := b.Timeline(ctx, "alice", "p")
	if legacy.Revision != 0 {
		t.Fatal("legacy changed")
	}
	if _, e = s.Timeline(ctx, "bob", "p", a.ID); !errors.Is(e, ErrProjectNotFound) {
		t.Fatalf("owner=%v", e)
	}
}
func TestTimelineCollectionCRUDLimit(t *testing.T) {
	ctx, _, s := collectionFixture(t)
	a, _ := s.Create(ctx, "alice", "p", "A")
	if _, e := s.Rename(ctx, "alice", "p", a.ID, "Renamed"); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Save(ctx, "alice", "p", a.ID, Timeline{Script: "original"}, 0, time.Now()); e != nil {
		t.Fatal(e)
	}
	b, e := s.Duplicate(ctx, "alice", "p", a.ID, "Copy")
	if e != nil {
		t.Fatal(e)
	}
	v, _ := s.Timeline(ctx, "alice", "p", b.ID)
	if v.Script != "original" || v.Revision != 0 {
		t.Fatalf("copy=%+v", v)
	}
	if e = s.Reorder(ctx, "alice", "p", []string{b.ID, DefaultTimelineID, a.ID}); e != nil {
		t.Fatal(e)
	}
	items, _ := s.List(ctx, "alice", "p")
	if items[0].ID != b.ID || items[2].Title != "Renamed" {
		t.Fatalf("items=%+v", items)
	}
	if e = s.Reorder(ctx, "alice", "p", []string{b.ID, b.ID, a.ID}); !errors.Is(e, ErrInvalidTimelineCollection) {
		t.Fatalf("order=%v", e)
	}
	if e = s.Delete(ctx, "alice", "p", DefaultTimelineID); !errors.Is(e, ErrDefaultTimelineProtected) {
		t.Fatalf("default=%v", e)
	}
	if e = s.Delete(ctx, "alice", "p", a.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Timeline(ctx, "alice", "p", a.ID); !errors.Is(e, ErrTimelineNotFound) {
		t.Fatalf("deleted=%v", e)
	}
	for i := 2; i < 50; i++ {
		if _, e = s.Create(ctx, "alice", "p", "More"); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = s.Create(ctx, "alice", "p", "Overflow"); !errors.Is(e, ErrTimelineLimit) {
		t.Fatalf("limit=%v", e)
	}
}
func TestTimelineCollectionScopeAndAssets(t *testing.T) {
	ctx, _, s := collectionFixture(t)
	a, _ := s.Create(ctx, "alice", "p", "A")
	_, e := s.Save(ctx, "alice", "p", a.ID, Timeline{Markers: []TimelineMarker{{ID: "m", AssetID: "foreign"}}}, 0, time.Now())
	if !errors.Is(e, ErrAssetNotFound) {
		t.Fatalf("asset=%v", e)
	}
	sc, e := s.ScopedStore(ctx, "alice", "p", a.ID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = sc.SaveTimeline(ctx, "alice", "p", Timeline{Script: "scoped"}, 0, time.Now()); e != nil {
		t.Fatal(e)
	}
	v, _ := s.Timeline(ctx, "alice", "p", a.ID)
	if v.Script != "scoped" {
		t.Fatal(v)
	}
	if e = sc.CreateASRJob(ctx, "alice", ASRJob{ProjectID: "p"}); !errors.Is(e, ErrTimelineJobUnsupported) {
		t.Fatalf("job=%v", e)
	}
	if _, e = sc.Timeline(ctx, "alice", "other"); !errors.Is(e, ErrProjectNotFound) {
		t.Fatalf("escape=%v", e)
	}
}

func TestTimelineCollectionDuplicateRegeneratesIdentities(t *testing.T) {
	ctx, base, s := collectionFixture(t)
	_, err := base.SaveTimeline(ctx, "alice", "p", Timeline{Markers: []TimelineMarker{{ID: "mark", Label: "Cue"}}, Tracks: []Track{{ID: "text", Kind: TrackSubtitle, Clips: []Clip{{ID: "caption", Kind: ClipSubtitle, Text: "Hello", DurationMS: 1000, CompoundID: "group", CompoundName: "Group", LinkGroupID: "linked"}}}}}, 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.Duplicate(ctx, "alice", "p", DefaultTimelineID, "Copy")
	if err != nil {
		t.Fatal(err)
	}
	v, _ := s.Timeline(ctx, "alice", "p", a.ID)
	if v.Tracks[0].ID == "text" || v.Tracks[0].Clips[0].ID == "caption" || v.Markers[0].ID == "mark" || v.Tracks[0].Clips[0].CompoundID == "group" || v.Tracks[0].Clips[0].LinkGroupID == "linked" {
		t.Fatalf("copied identities: %+v", v)
	}
}

func TestTimelineCollectionConcurrentWritersAndUndoBranch(t *testing.T) {
	ctx, _, s := collectionFixture(t)
	a, _ := s.Create(ctx, "alice", "p", "A")
	results := make(chan error, 16)
	for i := 0; i < 16; i++ {
		go func() {
			_, e := s.Save(ctx, "alice", "p", a.ID, Timeline{Script: "winner"}, 0, time.Now())
			results <- e
		}()
	}
	successes := 0
	for i := 0; i < 16; i++ {
		e := <-results
		if e == nil {
			successes++
		} else if !errors.Is(e, ErrRevisionConflict) {
			t.Fatal(e)
		}
	}
	if successes != 1 {
		t.Fatalf("successes=%d", successes)
	}
	if _, _, e := s.Undo(ctx, "alice", "p", a.ID, 1, time.Now()); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Save(ctx, "alice", "p", a.ID, Timeline{Script: "branch"}, 2, time.Now()); e != nil {
		t.Fatal(e)
	}
	h, e := s.History(ctx, "alice", "p", a.ID)
	if e != nil || h.CanRedo || !h.CanUndo {
		t.Fatalf("history=%+v %v", h, e)
	}
	if _, _, e = s.Redo(ctx, "alice", "p", a.ID, 3, time.Now()); !errors.Is(e, ErrTimelineHistoryBoundary) {
		t.Fatalf("redo=%v", e)
	}
}

func TestTimelineCollectionScopedJobsCannotReadOrCancelDefaultJobs(t *testing.T) {
	ctx, _, s := collectionFixture(t)
	a, _ := s.Create(ctx, "alice", "p", "A")
	sc, e := s.ScopedStore(ctx, "alice", "p", a.ID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = sc.ListASRJobs(ctx, "alice", "p"); !errors.Is(e, ErrTimelineJobUnsupported) {
		t.Fatalf("default job read=%v", e)
	}
	if _, e = sc.RequestExportCancel(ctx, "alice", "p", "default-job", time.Now()); !errors.Is(e, ErrTimelineJobUnsupported) {
		t.Fatalf("default cancel=%v", e)
	}
}

func TestTimelineCollectionRejectsForeignSegmentationMatte(t *testing.T) {
	ctx, base, s := collectionFixture(t)
	a, _ := s.Create(ctx, "alice", "p", "A")
	if e := base.CreateAsset(ctx, "alice", Asset{ID: "source", ProjectID: "p", Kind: "video", Name: "Source", ContentType: "video/mp4", ObjectKey: "source.mp4", ByteSize: 1, CreatedAt: time.Now()}); e != nil {
		t.Fatal(e)
	}
	v := Timeline{Tracks: []Track{{ID: "video", Kind: TrackVideo, Clips: []Clip{{ID: "clip", Kind: ClipVideo, AssetID: "source", DurationMS: 1000, SourceDurationMS: 1000, VisualProcessing: &VisualProcessing{PersonSegmentation: PersonSegmentation{MatteAssetID: "foreign"}}}}}}}
	if _, e := s.Save(ctx, "alice", "p", a.ID, v, 0, time.Now()); !errors.Is(e, ErrAssetNotFound) {
		t.Fatalf("foreign matte=%v", e)
	}
}
