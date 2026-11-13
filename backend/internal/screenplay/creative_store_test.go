package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func TestCreativeStoreIsolationSnapshotsAndClear(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()
	_, e := s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice"})
	if e != nil {
		t.Fatal(e)
	}
	v, e := s.SaveUserStyle(ctx, UserStyle{ID: "style-a", OwnerUsername: "alice", Name: "私人画风", Prompt: "细腻的钢笔线", Images: []string{}}, nil)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.GetUserStyle(ctx, "bob", v.ID); !errors.Is(e, ErrNotFound) {
		t.Fatalf("foreign style: %v", e)
	}
	if list, _ := s.ListUserStyles(ctx, "bob"); len(list) != 0 {
		t.Fatal("library leaked")
	}
	pmt, e := prompt.ResolveCreativePrompt(prompt.CreativeScript144, "script")
	if e != nil {
		t.Fatal(e)
	}
	c := CreativeConfig{ScriptPromptID: pmt.ID, Style: &CreativeStyleSnapshot{ID: v.ID, Name: v.Name, Prompt: v.Prompt, Images: v.Images}, PromptSnapshots: map[string]CreativePromptSnapshot{"script": pmt}, ImageCoreInstruction: "保留轮廓"}
	saved, e := s.SaveCreativeConfig(ctx, "alice", "p", c, 0)
	if e != nil || saved.Revision != 1 || saved.StyleDesc != v.Prompt {
		t.Fatalf("save=%+v %v", saved, e)
	}
	// The caller cannot mutate stored snapshots through shared map/slice pointers.
	saved.CreativeConfig.Style.Prompt = "tampered"
	saved.CreativeConfig.PromptSnapshots["script"] = CreativePromptSnapshot{}
	loaded, e := s.GetProject(ctx, "alice", "p")
	if e != nil || loaded.CreativeConfig.Style.Prompt != v.Prompt || loaded.CreativeConfig.PromptSnapshots["script"].Body != pmt.Body {
		t.Fatalf("snapshot alias/loss: %v", e)
	}
	raw, _ := json.Marshal(loaded)
	if strings.Contains(string(raw), pmt.Body[:min(50, len(pmt.Body))]) {
		t.Fatal("system body leaked")
	}
	c.Style = nil
	cleared, e := s.SaveCreativeConfig(ctx, "alice", "p", c, 1)
	if e != nil || cleared.ArtStyle != "" || cleared.StyleDesc != "" || cleared.SelectedStyle != "" {
		t.Fatalf("clear=%+v %v", cleared, e)
	}
	if _, e = s.SaveCreativeConfig(ctx, "alice", "p", c, 1); !errors.Is(e, ErrRevisionConflict) {
		t.Fatalf("stale write=%v", e)
	}
	if _, e = s.SaveCreativeConfig(ctx, "bob", "p", c, 2); !errors.Is(e, ErrNotFound) {
		t.Fatalf("foreign save=%v", e)
	}
	b, e := encodeCreativeConfig(&c)
	if e != nil {
		t.Fatal(e)
	}
	round, e := decodeCreativeConfig(b)
	if e != nil || round.PromptSnapshots["script"].Body != pmt.Body {
		t.Fatalf("body persistence=%v", e)
	}
}
