package novelcreation

import (
	"context"
	"encoding/json"
	"testing"
)

func TestStudioMaterialActionConfirmationDoesNotCreateOrCallModel(t *testing.T) {
	for _, kind := range []string{"style_imitation", "continuation_import"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s, st, p, ss := studioFixture(t)
			p.reply = `{"message":"请补充材料","action":{"type":"` + kind + `"}}`
			snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "需要导入材料", ModelID: "own"})
			if e != nil || snap.PendingAction == nil {
				t.Fatalf("material proposal rejected %+v %v", snap, e)
			}
			// Opening an intake form must remain usable even without a model service.
			reopened := NewService(st, NewMemoryRuntime(), nil)
			in := StudioActionInput{ActionID: snap.PendingAction.ID}
			out, e := reopened.ConfirmStudioAction(ctx, "alice", ss.ID, in)
			if e != nil {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(out)
			var route map[string]any
			_ = json.Unmarshal(raw, &route)
			expected := "imitation"
			if kind == "continuation_import" {
				expected = "chapters"
			}
			if route["assistedAction"] != kind || route["targetRoute"] != expected || out.Project.ID != "" || out.Run.ID != "" {
				t.Fatalf("not a pure intake route %s", raw)
			}
			again, e := reopened.ConfirmStudioAction(ctx, "alice", ss.ID, in)
			if e != nil || fingerprint(again) != fingerprint(out) {
				t.Fatal("receipt changed", e)
			}
			projects, _ := st.ListProjects(ctx, "alice")
			if len(projects) != 0 || p.calls != 1 {
				t.Fatal("material navigation generated/created")
			}
			if _, e = reopened.ConfirmStudioAction(ctx, "bob", ss.ID, in); e == nil {
				t.Fatal("cross-account route")
			}
			follow, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "follow", Content: "继续讨论", ModelID: "own"})
			if e != nil || follow.Turn.State != "completed" {
				t.Fatal("intake route locked chat", e)
			}
		})
	}
}

func TestStudioMaterialCancelledActionCannotOpen(t *testing.T) {
	ctx := context.Background()
	s, _, p, ss := studioFixture(t)
	p.reply = `{"message":"补充材料","action":{"type":"continuation_import"}}`
	snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "a", Content: "导入", ModelID: "own"})
	if e != nil || snap.PendingAction == nil {
		t.Fatalf("missing action %+v %v", snap, e)
	}
	in := StudioActionInput{ActionID: snap.PendingAction.ID}
	if _, e = s.CancelStudioAction(ctx, "alice", ss.ID, in); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ConfirmStudioAction(ctx, "alice", ss.ID, in); e == nil {
		t.Fatal("cancelled intake action opened")
	}
}
