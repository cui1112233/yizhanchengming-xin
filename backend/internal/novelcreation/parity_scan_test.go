package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestEmptyRunFixtureMatchesActualSnapshot(t *testing.T) {
	f := newServiceFixture(t, nil)
	snap, e := f.service.GetSessionSnapshot(context.Background(), "alice", f.session.ID)
	if e != nil {
		t.Fatal(e)
	}
	actual, e := json.Marshal(snap.Run)
	if e != nil {
		t.Fatal(e)
	}
	fixture, e := os.ReadFile("testdata/empty-run.json")
	if e != nil {
		t.Fatal(e)
	}
	var a, b any
	json.Unmarshal(actual, &a)
	json.Unmarshal(fixture, &b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("browser empty DTO differs from real Go snapshot: %s", actual)
	}
}

type paritySavedReadFailureStore struct{ Store }

func (paritySavedReadFailureStore) HasSavedRun(context.Context, string, string) (bool, error) {
	return false, errors.New("saved fact read failed")
}
func TestShortExportSnapshotUsesPersistedVersionFact(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	se, e := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "short_original"})
	if e != nil {
		t.Fatal(e)
	}
	f.session = se
	check := func(s Snapshot, want bool) {
		t.Helper()
		b, e := json.Marshal(s)
		if e != nil {
			t.Fatal(e)
		}
		var wire map[string]any
		json.Unmarshal(b, &wire)
		v, ok := wire["canExportShort"].(bool)
		if !ok || v != want {
			t.Fatalf("canExportShort=%v present=%t want=%t", v, ok, want)
		}
	}
	empty, e := f.service.GetSessionSnapshot(ctx, "alice", se.ID)
	if e != nil {
		t.Fatal(e)
	}
	check(empty, false)
	r, e := f.service.SendMessage(ctx, "alice", f.sendInput("saved-short"))
	if e != nil {
		t.Fatal(e)
	}
	queued, e := f.service.GetSnapshot(ctx, "alice", r.ID, 0)
	if e != nil {
		t.Fatal(e)
	}
	check(queued, false)
	if _, e = f.service.Run(ctx, "alice", r.ID); e != nil {
		t.Fatal(e)
	}
	for _, getter := range []func() (Snapshot, error){func() (Snapshot, error) { return f.service.GetSnapshot(ctx, "alice", r.ID, 0) }, func() (Snapshot, error) { return f.service.GetSessionSnapshot(ctx, "alice", se.ID) }} {
		s, e := getter()
		if e != nil {
			t.Fatal(e)
		}
		check(s, true)
	}
	stored, e := f.store.GetRun(ctx, "alice", r.ID)
	if e != nil {
		t.Fatal(e)
	}
	if can, e := f.service.canExportShort(ctx, "alice", stored, Session{Mode: "long_original"}); e != nil || can {
		t.Fatal("saved long mode offered short export")
	}
	maintenance := stored
	maintenance.PromptSnapshot = json.RawMessage(`{"operation":"style_analysis"}`)
	if can, e := f.service.canExportShort(ctx, "alice", maintenance, se); e != nil || can {
		t.Fatal("maintenance offered short export despite historical saved fact")
	}
	cs, _ := f.store.ListChapters(ctx, "alice", f.project.ID)
	if _, e = f.service.EditChapter(ctx, "alice", f.project.ID, cs[0].ID, ChapterEdit{ExpectedVersion: cs[0].Version, Content: "手动新版本"}); e != nil {
		t.Fatal(e)
	}
	s, e := f.service.GetSnapshot(ctx, "alice", r.ID, 0)
	if e != nil {
		t.Fatal(e)
	}
	check(s, true)
	for _, prompt := range []string{`{"checkpoint":true}`, `{"operation":"style_analysis"}`} {
		source, e := f.store.CreateRun(ctx, "alice", Run{ID: novelID("source"), ProjectID: f.project.ID, SessionID: se.ID, IdempotencyKey: novelID("source-key"), Status: RunSucceeded, PromptSnapshot: json.RawMessage(prompt)})
		if e != nil {
			t.Fatal(e)
		}
		s, e = f.service.GetSnapshot(ctx, "alice", source.ID, 0)
		if e != nil {
			t.Fatal(e)
		}
		check(s, false)
	}
	current, e := f.store.GetSession(ctx, "alice", se.ID)
	if e != nil {
		t.Fatal(e)
	}
	current.CurrentRunID = r.ID
	if _, e = f.store.UpdateSession(ctx, "alice", current); e != nil {
		t.Fatal(e)
	}
	f.service.store = paritySavedReadFailureStore{f.store}
	if _, e = f.service.GetSnapshot(ctx, "alice", r.ID, 0); e == nil {
		t.Fatal("saved-result read failure hidden")
	}
	if _, e = f.service.GetSessionSnapshot(ctx, "alice", se.ID); e == nil {
		t.Fatal("session hid saved-result read failure")
	}
}

func TestBatchScannerSkipsArchivedAndPurgingProjects(t *testing.T) {
	for _, status := range []string{"draft", "archived", "purging"} {
		t.Run(status, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			b, e := f.service.CreateBatch(context.Background(), "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1})
			if e != nil {
				t.Fatal(e)
			}
			if status != "draft" {
				if e = f.service.ArchiveProject(context.Background(), "alice", f.project.ID); e != nil {
					t.Fatal(e)
				}
			}
			if status == "purging" {
				p := f.store.projects[memoryKey("alice", f.project.ID)]
				p.Status = status
				f.store.projects[memoryKey("alice", f.project.ID)] = p
			}
			batches, e := f.store.ListActiveBatches(context.Background(), 20)
			if e != nil {
				t.Fatal(e)
			}
			if status == "draft" {
				if len(batches) != 1 || batches[0].ID != b.ID {
					t.Fatal("normal batch filtered out")
				}
			} else if len(batches) != 0 {
				t.Fatalf("%s project leaked %d batches into bounded scanner", status, len(batches))
			}
		})
	}
}
