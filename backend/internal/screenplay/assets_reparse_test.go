package screenplay

import (
	"context"
	"errors"
	"qiantie/backend/internal/prompt/luoshui"
	"testing"
)

func TestStandaloneAssetReparseMessages(t *testing.T) {
	for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
		t.Run(string(kind), func(t *testing.T) {
			s, store, g := newAssetService(t, PermissionBypass)
			p, _ := store.GetProject(context.Background(), "alice", "p1")
			p.ModelID = "own-saved-model"
			p.StyleDesc = "must-not-enter-parse"
			if _, e := store.UpdateProject(context.Background(), "alice", "p1", p, p.Revision); e != nil {
				t.Fatal(e)
			}
			key := map[AssetType]string{AssetCharacter: "characters", AssetScene: "scenes", AssetProp: "props"}[kind]
			g.responses = []string{`{"` + key + `":[{"name":"甲","prompt":"新描述"}]}`}
			r, e := s.Reparse(context.Background(), AssetRequest{Owner: "alice", ProjectID: "p1", Type: kind, Content: "完整原文", Model: "ignored-request-model", Confirmed: true})
			if e != nil {
				t.Fatal(e)
			}
			if len(r.Assets) != 1 {
				t.Fatalf("assets=%+v", r.Assets)
			}
			got := g.requests[0]
			system, user, ok := luoshui.NativeAssetMessages(string(kind), "完整原文")
			if !ok || got.SystemPrompt != system || got.UserPrompt != user {
				t.Fatal("messages differ from native capture")
			}
			if got.Model != "own-saved-model" || !got.SingleAttempt || got.Temperature != 1 || got.MaxOutputTokens != 20000 {
				t.Fatalf("request parameters differ: model=%s temperature=%v tokens=%d", got.Model, got.Temperature, got.MaxOutputTokens)
			}
		})
	}
}

func TestStandaloneAssetReparseInvalidPreservesOld(t *testing.T) {
	for _, raw := range []string{"", "invalid", `{"characters":[]}`, `{"wrong":[]}`} {
		t.Run(raw, func(t *testing.T) {
			s, store, g := newAssetService(t, PermissionBypass)
			p, _ := store.GetProject(context.Background(), "alice", "p1")
			p.ModelID = "own-model"
			_, _ = store.UpdateProject(context.Background(), "alice", "p1", p, p.Revision)
			old, _ := store.CreateAsset(context.Background(), Asset{OwnerUsername: "alice", ProjectID: "p1", Type: AssetCharacter, Name: "甲", Prompt: "旧提示词", Extra: map[string]string{"image_path": "old.png"}})
			g.responses = []string{raw}
			_, e := s.Reparse(context.Background(), AssetRequest{Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Content: "正文", Confirmed: true})
			if e == nil {
				t.Fatal("invalid result reported success")
			}
			rows, _ := store.ListAssets(context.Background(), "alice", "p1", AssetCharacter)
			if len(rows) != 1 || rows[0].ID != old.ID || rows[0].Prompt != old.Prompt || rows[0].Extra["image_path"] != "old.png" {
				t.Fatalf("old result changed: %+v", rows)
			}
		})
	}
}

func TestStandaloneAssetReparseRejectsInvalidInput(t *testing.T) {
	s, _, g := newAssetService(t, PermissionBypass)
	req := AssetRequest{Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Content: "正文", Confirmed: true}
	if _, e := s.Reparse(context.Background(), req); !errors.Is(e, ErrInvalidArgument) {
		t.Fatalf("missing model=%v", e)
	}
	req.Content = ""
	if _, e := s.Reparse(context.Background(), req); !errors.Is(e, ErrEmptyContent) {
		t.Fatalf("empty=%v", e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req.Content = "正文"
	if _, e := s.Reparse(ctx, req); !errors.Is(e, ErrUserCancelled) {
		t.Fatalf("cancel=%v", e)
	}
	if g.callCount() != 0 {
		t.Fatal("invalid request called model")
	}
}

func TestStandaloneAssetReparseRepeatKeepsIdentityAndMedia(t *testing.T) {
	s, store, g := newAssetService(t, PermissionBypass)
	p, _ := store.GetProject(context.Background(), "alice", "p1")
	p.ModelID = "own-model"
	_, _ = store.UpdateProject(context.Background(), "alice", "p1", p, p.Revision)
	old, _ := store.CreateAsset(context.Background(), Asset{OwnerUsername: "alice", ProjectID: "p1", Type: AssetCharacter, Name: "甲", Prompt: "旧", Extra: map[string]string{"image_path": "old.png"}})
	g.responses = []string{`{"characters":[{"name":"甲","prompt":"新","image_path":"model-must-not-replace.png"}]}`}
	for i := 0; i < 2; i++ {
		if _, e := s.Reparse(context.Background(), AssetRequest{Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Content: "正文", Confirmed: true}); e != nil {
			t.Fatal(e)
		}
	}
	rows, _ := store.ListAssets(context.Background(), "alice", "p1", AssetCharacter)
	if len(rows) != 1 || rows[0].ID != old.ID || rows[0].Prompt != "新" || rows[0].Extra["image_path"] != "old.png" {
		t.Fatalf("lost identity/media: %+v", rows)
	}
}

func TestStandaloneAssetReparseLateCancellationDoesNotSave(t *testing.T) {
	s, store, _ := newAssetService(t, PermissionBypass)
	p, _ := store.GetProject(context.Background(), "alice", "p1")
	p.ModelID = "own-model"
	_, _ = store.UpdateProject(context.Background(), "alice", "p1", p, p.Revision)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Generator = lateCanceledGenerator{cancel: cancel, content: charactersJSON}
	_, e := s.Reparse(ctx, AssetRequest{Owner: "alice", ProjectID: "p1", Type: AssetCharacter, Content: "正文", Confirmed: true})
	if !errors.Is(e, ErrUserCancelled) {
		t.Fatalf("error=%v", e)
	}
	rows, _ := store.ListAssets(context.Background(), "alice", "p1", AssetCharacter)
	if len(rows) != 0 {
		t.Fatalf("saved late response: %+v", rows)
	}
}
