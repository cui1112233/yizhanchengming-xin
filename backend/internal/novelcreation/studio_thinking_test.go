package novelcreation

import (
	"context"
	"testing"
)

func TestStudioConfirmationThinkingRequirementFrozenAndReplayStable(t *testing.T) {
	for _, selected := range []string{"", "deep"} {
		t.Run("selected-"+selected, func(t *testing.T) {
			s, _, _, ss := studioFixture(t)
			ctx := context.Background()
			snap, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "chat", Content: "想法", ModelID: "own", ThinkingLevel: "fast"})
			if e != nil {
				t.Fatal(e)
			}
			a := StudioActionInput{ActionID: snap.PendingAction.ID, ModelID: "own", ThinkingLevel: selected}
			out, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, a)
			if e != nil {
				t.Fatal(e)
			}
			want := selected
			if want == "" {
				want = "fast"
			}
			if out.Run.ThinkingLevel != want {
				t.Fatalf("queued thinking=%s want=%s", out.Run.ThinkingLevel, want)
			}
			a.ThinkingLevel = "standard"
			again, e := s.ConfirmStudioAction(ctx, "alice", ss.ID, a)
			if e != nil || again.Run.ID != out.Run.ID || again.Run.ThinkingLevel != want {
				t.Fatal("replay changed confirmed parameters", e)
			}
		})
	}
}
