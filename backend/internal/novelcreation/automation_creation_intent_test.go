package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestAutomationCreationIntentReplayBeforeMutableValidation(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			note, _ := f.store.CreateNote(ctx, "alice", f.project.ID, ProjectNoteInput{Content: "原引用"})
			data, _ := json.Marshal(map[string]any{"idempotencyKey": "stable-original", "sessionId": f.session.ID, "prompt": "原始要求", "noteIds": []string{note.ID}, "enabled": true})
			var scheduleIn ScheduleInput
			var batchIn BatchInput
			json.Unmarshal(data, &scheduleIn)
			json.Unmarshal(data, &batchIn)
			create := func(s *Service) (string, error) {
				if kind == "schedule" {
					p, e := s.CreateSchedule(ctx, "alice", f.project.ID, scheduleIn)
					return p.ID, e
				}
				p, e := s.CreateBatch(ctx, "alice", f.project.ID, batchIn)
				return p.ID, e
			}
			id, e := create(f.service)
			if e != nil {
				t.Fatal(e)
			}
			// References and mutable plan configuration can change after successful creation.
			f.store.DeleteNote(ctx, "alice", f.project.ID, note.ID)
			if kind == "schedule" {
				p, _ := f.store.GetSchedule(ctx, "alice", id)
				p.Prompt = "后来修改"
				p.Status = "stopped"
				p.Enabled = false
				p.TotalSavedChapters = 7
				f.store.UpdateSchedule(ctx, "alice", p)
			} else {
				p, _ := f.store.GetBatch(ctx, "alice", id)
				p.Status = BatchStopped
				p.CompletedChapters = 4
				p.Prompt = "后来修改"
				f.store.UpdateBatch(ctx, "alice", p)
			}
			other := NewService(f.store, NewMemoryRuntime(), nil)
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					got, err := create(other)
					if err != nil || got != id {
						t.Errorf("recovery got=%s want=%s err=%v", got, id, err)
					}
				}()
			}
			wg.Wait()
			if kind == "schedule" {
				plans, _ := f.store.ListSchedules(ctx, "alice", f.project.ID)
				if len(plans) != 1 || plans[0].Prompt != "后来修改" || plans[0].TotalSavedChapters != 7 || plans[0].Enabled {
					t.Fatalf("reset/recreated: %+v", plans)
				}
				scheduleIn.Prompt = "不同请求"
				_, e = f.service.CreateSchedule(ctx, "alice", f.project.ID, scheduleIn)
			} else {
				plans, _ := f.store.ListBatches(ctx, "alice", f.project.ID)
				if len(plans) != 1 || plans[0].Status != BatchStopped || plans[0].CompletedChapters != 4 {
					t.Fatalf("reset/recreated: %+v", plans)
				}
				batchIn.Prompt = "不同请求"
				_, e = f.service.CreateBatch(ctx, "alice", f.project.ID, batchIn)
			}
			if !IsCode(e, ErrIdempotencyConflict) {
				t.Fatalf("changed payload accepted: %v", e)
			}
		})
	}
}

func TestAutomationCreationIntentConcurrentInitialAndArchivedReadOnly(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			other := NewService(f.store, NewMemoryRuntime(), nil)
			si := ScheduleInput{IdempotencyKey: "simultaneous", SessionID: f.session.ID, Prompt: "原请求", Enabled: true}
			bi := BatchInput{IdempotencyKey: "simultaneous", SessionID: f.session.ID, Prompt: "原请求"}
			create := func(s *Service) (string, error) {
				if kind == "schedule" {
					p, e := s.CreateSchedule(ctx, "alice", f.project.ID, si)
					return p.ID, e
				}
				p, e := s.CreateBatch(ctx, "alice", f.project.ID, bi)
				return p.ID, e
			}
			start := make(chan struct{})
			ids := make(chan string, 8)
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					<-start
					s := f.service
					if i%2 == 1 {
						s = other
					}
					id, e := create(s)
					if e != nil {
						t.Error(e)
					}
					ids <- id
				}(i)
			}
			close(start)
			wg.Wait()
			close(ids)
			id := ""
			for got := range ids {
				if id == "" {
					id = got
				}
				if got != id || got == "" {
					t.Fatalf("different initial IDs: %s/%s", id, got)
				}
			}
			if kind == "schedule" {
				p, _ := f.store.GetSchedule(ctx, "alice", id)
				p.Enabled = false
				p.Status = "stopped"
				f.store.UpdateSchedule(ctx, "alice", p)
			} else {
				p, _ := f.store.GetBatch(ctx, "alice", id)
				p.Status = BatchStopped
				f.store.UpdateBatch(ctx, "alice", p)
			}
			if e := f.service.ArchiveProject(ctx, "alice", f.project.ID); e != nil {
				t.Fatal(e)
			}
			before, _ := f.store.GetProject(ctx, "alice", f.project.ID)
			got, e := create(other)
			after, _ := f.store.GetProject(ctx, "alice", f.project.ID)
			if e != nil || got != id || !reflect.DeepEqual(before, after) {
				t.Fatalf("archived replay mutated project: %s %v", got, e)
			}
			if kind == "schedule" {
				plans, _ := f.store.ListSchedules(ctx, "alice", f.project.ID)
				if len(plans) != 1 {
					t.Fatal(plans)
				}
				si.IdempotencyKey = "new"
				_, e = f.service.CreateSchedule(ctx, "alice", f.project.ID, si)
				if _, foreign := f.service.CreateSchedule(ctx, "bob", f.project.ID, si); foreign == nil {
					t.Fatal("foreign creation")
				}
			} else {
				plans, _ := f.store.ListBatches(ctx, "alice", f.project.ID)
				if len(plans) != 1 {
					t.Fatal(plans)
				}
				bi.IdempotencyKey = "new"
				_, e = f.service.CreateBatch(ctx, "alice", f.project.ID, bi)
				if _, foreign := f.service.CreateBatch(ctx, "bob", f.project.ID, bi); foreign == nil {
					t.Fatal("foreign creation")
				}
			}
			if e == nil {
				t.Fatal("archived new intent dispatched")
			}
		})
	}
}

type failCreationStore struct{ Store }

func (s failCreationStore) Atomic(ctx context.Context, o, p string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, o, p, func(tx Store) error { return fn(failCreationStore{tx}) })
}
func (s failCreationStore) CreateSchedule(ctx context.Context, o string, p Schedule) (Schedule, error) {
	p, e := s.Store.CreateSchedule(ctx, o, p)
	if e == nil {
		e = errors.New("fail after insert")
	}
	return p, e
}
func (s failCreationStore) CreateBatch(ctx context.Context, o string, p Batch) (Batch, error) {
	p, e := s.Store.CreateBatch(ctx, o, p)
	if e == nil {
		e = errors.New("fail after insert")
	}
	return p, e
}
func TestAutomationCreationIntentMemoryRollbackAndLegacy(t *testing.T) {
	for _, kind := range []string{"schedule", "batch"} {
		t.Run(kind, func(t *testing.T) {
			f := newServiceFixture(t, nil)
			ctx := context.Background()
			bad := NewService(failCreationStore{f.store}, NewMemoryRuntime(), nil)
			si := ScheduleInput{IdempotencyKey: "rollback", SessionID: f.session.ID, Prompt: "原请求"}
			bi := BatchInput{IdempotencyKey: "rollback", SessionID: f.session.ID}
			if kind == "schedule" {
				if _, e := bad.CreateSchedule(ctx, "alice", f.project.ID, si); e == nil {
					t.Fatal("missing failure")
				}
				plans, _ := f.store.ListSchedules(ctx, "alice", f.project.ID)
				if len(plans) != 0 {
					t.Fatal("partial persisted")
				}
				if _, e := f.service.CreateSchedule(ctx, "alice", f.project.ID, si); e != nil {
					t.Fatal(e)
				}
				si.IdempotencyKey = ""
				a, _ := f.service.CreateSchedule(ctx, "alice", f.project.ID, si)
				b, _ := f.service.CreateSchedule(ctx, "alice", f.project.ID, si)
				if a.ID == b.ID {
					t.Fatal("legacy changed")
				}
			} else {
				if _, e := bad.CreateBatch(ctx, "alice", f.project.ID, bi); e == nil {
					t.Fatal("missing failure")
				}
				plans, _ := f.store.ListBatches(ctx, "alice", f.project.ID)
				if len(plans) != 0 {
					t.Fatal("partial persisted")
				}
				if _, e := f.service.CreateBatch(ctx, "alice", f.project.ID, bi); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}

func TestAutomationCreationIntentScheduleEditKeepsOriginalIdentityAndUnknownInputs(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	session, _ := f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "short_imitation"})
	original := ScheduleInput{IdempotencyKey: "before-edit", SessionID: session.ID, Prompt: "原指令", ModeInputs: map[string]string{"referenceText": "原参考"}}
	p, e := f.service.CreateSchedule(ctx, "alice", f.project.ID, original)
	if e != nil {
		t.Fatal(e)
	}
	meta := *p.Creation
	p.ModeInputs["futureField"] = "保留原字段"
	f.store.UpdateSchedule(ctx, "alice", p)
	edited := original
	edited.Prompt = "修改后的指令"
	edited.ModeInputs = map[string]string{"referenceText": "明确修改参考"}
	saved, e := f.service.UpdateSchedule(ctx, "alice", f.project.ID, p.ID, edited)
	if e != nil {
		t.Fatal(e)
	}
	if saved.SessionID != session.ID || saved.ModeInputs["referenceText"] != "明确修改参考" || saved.ModeInputs["futureField"] != "保留原字段" || !reflect.DeepEqual(*saved.Creation, meta) {
		t.Fatalf("edit lost fields: %+v", saved)
	}
	recovered, e := f.service.CreateSchedule(ctx, "alice", f.project.ID, original)
	if e != nil || recovered.ID != p.ID || recovered.Prompt != "修改后的指令" {
		t.Fatalf("%+v %v", recovered, e)
	}
}
