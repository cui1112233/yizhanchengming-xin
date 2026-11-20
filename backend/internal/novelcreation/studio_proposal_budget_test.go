package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestStudioShortProposalCannotBypassContextLimit(t *testing.T) {
	for _, field := range []string{"direction", "instruction", "summary", "readerPreferences", "title"} {
		t.Run(field, func(t *testing.T) {
			ctx := context.Background()
			s, st, p, ss := studioFixture(t)
			action := map[string]any{"type": "short_run", "direction": "写一个完整故事", field: strings.Repeat("必须保留的要求", 300)}
			if field == "title" {
				action[field] = strings.Repeat("甲", 200)
			}
			raw, _ := json.Marshal(map[string]any{"message": "请确认短篇方向", "action": action})
			p.reply = string(raw)
			first, e := s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "first", Content: "写短篇", ModelID: "own", ContextBudget: 64000})
			if e != nil || first.PendingAction == nil {
				t.Fatalf("initial proposal failed %+v %v", first, e)
			}
			_, e = s.SendStudioMessage(ctx, "alice", ss.ID, StudioSendInput{IdempotencyKey: "too-small", Content: "改成第一人称", ModelID: "own", ContextBudget: 100})
			if !IsCode(e, ErrInputRequired) {
				t.Fatalf("oversized short proposal was sent: %v", e)
			}
			if p.calls != 1 {
				t.Fatalf("rejected input called provider %d times", p.calls)
			}
			saved, e := s.GetStudioSnapshot(ctx, "alice", ss.ID)
			if e != nil || len(saved.Messages) != len(first.Messages) || saved.PendingAction.ID != first.PendingAction.ID || saved.PendingAction.Status != "pending" {
				t.Fatalf("rejection changed previous proposal %+v %v", saved, e)
			}
			projects, _ := st.ListProjects(ctx, "alice")
			if len(projects) != 0 {
				t.Fatal("rejection created project")
			}
		})
	}
}
