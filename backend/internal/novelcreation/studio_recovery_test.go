package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"
)

type studioFaultStore struct {
	Store
	state string
	fail  *bool
}

func (s studioFaultStore) AtomicCreate(ctx context.Context, owner string, fn func(Store) error) error {
	return s.Store.AtomicCreate(ctx, owner, func(tx Store) error { return fn(studioFaultStore{Store: tx, state: s.state, fail: s.fail}) })
}
func (s studioFaultStore) SaveStudio(ctx context.Context, owner, id string, raw json.RawMessage, create bool) error {
	var r studioRecord
	_ = json.Unmarshal(raw, &r)
	state := ""
	if len(r.Turns) > 0 {
		state = r.Turns[len(r.Turns)-1].State
	}
	if s.fail != nil && *s.fail && ((s.state == "confirmed" && r.Transition != nil) || state == s.state) {
		*s.fail = false
		return errors.New("injected studio persistence failure")
	}
	return s.Store.SaveStudio(ctx, owner, id, raw, create)
}
func TestStudioSavedRawRecoveryWithoutProviderAndAtomicTransitionRollback(t *testing.T) {
	s, st, p, ss := studioFixture(t)
	ctx := context.Background()
	fail := true
	s.store = studioFaultStore{Store: st, state: "completed", fail: &fail}
	_, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "recover", Content: "幻想故事", ModelID: "own"})
	if e == nil {
		t.Fatal("expected local interpretation save fault")
	}
	reopened := NewService(st, NewMemoryRuntime(), studioTestPipeline{provider: p})
	snap, e := reopened.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || snap.PendingAction == nil || p.calls != 1 {
		t.Fatalf("raw recovery: %+v %v calls=%d", snap, e, p.calls)
	}
	snap2, e := reopened.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || snap2.PendingAction.ID != snap.PendingAction.ID || len(snap2.Messages) != 2 {
		t.Fatal("GET recovery not idempotent")
	}
	fail = true
	reopened.store = studioFaultStore{Store: st, state: "confirmed", fail: &fail}
	a := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"}
	if _, e = reopened.ConfirmStudioAction(ctx, "alice", ss.ID, a); e == nil {
		t.Fatal("expected atomic confirmation rollback")
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 0 {
		t.Fatal("partial project survived rollback")
	}
	dispatch, _ := st.ListDispatch(ctx, 20)
	if len(dispatch) != 0 {
		t.Fatal("partial run survived rollback")
	}
	snap, e = reopened.GetStudioSnapshot(ctx, "alice", ss.ID)
	if e != nil || snap.PendingAction.Status != "pending" {
		t.Fatal("action mutated on rollback")
	}
	if _, e = reopened.ConfirmStudioAction(ctx, "alice", ss.ID, a); e != nil {
		t.Fatal(e)
	}
	if p.calls != 1 {
		t.Fatal("confirm called planner")
	}
}
func TestStudioExpiredUnsentAndUncertainRecovery(t *testing.T) {
	for _, state := range []string{"prepared", "may_have_sent"} {
		t.Run(state, func(t *testing.T) {
			s, st, p, ss := studioFixture(t)
			ctx := context.Background()
			e := st.AtomicCreate(ctx, "alice", func(tx Store) error {
				r, e := readStudio(ctx, tx, "alice", ss.ID, true)
				if e != nil {
					return e
				}
				r.Turns = []studioTurnRecord{{StudioTurn: StudioTurn{ID: "old", IdempotencyKey: "old", State: state, ModelID: "own"}, SentAt: time.Now().Add(-3 * time.Minute)}}
				return writeStudio(ctx, tx, "alice", r, false)
			})
			if e != nil {
				t.Fatal(e)
			}
			snap, e := s.GetStudioSnapshot(ctx, "alice", ss.ID)
			if e != nil {
				t.Fatal(e)
			}
			want := "rejected"
			if state == "may_have_sent" {
				want = "unknown"
			}
			if snap.Turn.State != want || p.calls != 0 {
				t.Fatal("incorrect recovery", snap.Turn)
			}
			_, e = s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "new", Content: "继续", ModelID: "own"})
			if state == "may_have_sent" && (!IsCode(e, ErrResultUnknown) || p.calls != 0) {
				t.Fatal("uncertain result resubmitted")
			}
			if state == "prepared" && (e != nil || p.calls != 1) {
				t.Fatal("unsent request remains stuck", e)
			}
		})
	}
}
func TestStudioConcurrentConfirmationAndCreation(t *testing.T) {
	s, st, _, ss := studioFixture(t)
	ctx := context.Background()
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "想法", ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	ids := make(chan string, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			same, e := s.CreateStudioSession(ctx, "alice", StudioCreateSessionInput{IdempotencyKey: "new", Mode: "long_original"})
			if e != nil || same.ID != ss.ID {
				errs <- errors.New("creation receipt mismatch")
				return
			}
			out, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own"})
			if e != nil {
				errs <- e
				return
			}
			ids <- out.Project.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	first := ""
	for id := range ids {
		if first == "" {
			first = id
		}
		if first != id {
			t.Fatal("parallel duplicate")
		}
	}
	projects, _ := st.ListProjects(ctx, "alice")
	if len(projects) != 1 {
		t.Fatal("parallel duplicate projects")
	}
}
func TestStudioPublicSnapshotOmitsPrivateJournal(t *testing.T) {
	s, _, _, ss := studioFixture(t)
	snap, e := s.SendStudioMessage(context.Background(), "alice", ss.ID, StudioSendInput{IdempotencyKey: "public-key", Content: "想法", ModelID: "own"})
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(snap)
	var view map[string]any
	_ = json.Unmarshal(raw, &view)
	if _, ok := view["Turns"]; ok {
		t.Fatal("journal leaked")
	}
	if snap.Turn.IdempotencyKey != "public-key" {
		t.Fatal("lost recovery key")
	}
}

func TestStudioSameSendKeyCannotSwitchSession(t *testing.T) {
	s, _, p, ss := studioFixture(t)
	ctx := context.Background()
	in := StudioSendInput{IdempotencyKey: "shared-key", Content: "故事", ModelID: "own"}
	if _, e := s.SendStudioMessage(ctx, "alice", ss.ID, in); e != nil {
		t.Fatal(e)
	}
	other, e := s.CreateStudioSession(ctx, "alice", StudioCreateSessionInput{IdempotencyKey: "other", Mode: "long_original"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SendStudioMessage(ctx, "alice", other.ID, in); !IsCode(e, ErrIdempotencyConflict) {
		t.Fatal("same key switched session", e)
	}
	if p.calls != 1 {
		t.Fatal("switch session rebilled")
	}
}

func TestStudioConcurrentSendUsesSinglePaidAttempt(t *testing.T) {
	s, _, p, ss := studioFixture(t)
	ctx := context.Background()
	p.started = make(chan struct{})
	p.release = make(chan struct{})
	in := StudioSendInput{IdempotencyKey: "one", Content: "想法", ModelID: "own"}
	done := make(chan error, 1)
	go func() { _, e := s.SendStudioMessage(ctx, "alice", ss.ID, in); done <- e }()
	select {
	case <-p.started:
	case <-time.After(time.Second):
		t.Fatal("provider did not start")
	}
	same, e := s.SendStudioMessage(ctx, "alice", ss.ID, in)
	if e != nil || same.Turn.State != "may_have_sent" {
		t.Fatal("duplicate didn't recover current turn", e)
	}
	different := in
	different.IdempotencyKey = "two"
	if _, e = s.SendStudioMessage(ctx, "alice", ss.ID, different); !IsCode(e, ErrResultUnknown) {
		t.Fatal("concurrent second turn accepted", e)
	}
	close(p.release)
	if e = <-done; e != nil {
		t.Fatal(e)
	}
	if p.calls != 1 {
		t.Fatal("parallel paid requests")
	}
}
