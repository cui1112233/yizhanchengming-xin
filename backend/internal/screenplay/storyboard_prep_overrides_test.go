package screenplay

import (
	"context"
	"strings"
	"testing"
)

func TestStoryboardRequestPrefixAndExtraContentReachPrompt(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	if _, err := store.MutateStoryboardPreferences(ctx, "alice", func(v *StoryboardPreferences) error {
		if _, _, err := ApplyStoryboardPreferencesAction(v, StoryboardPreferencesAction{ExpectedRevision: int64Ptr(0), Action: "prefix_save", Preset: StoryboardPrefixPreset{ID: "saved", Name: "账号前缀", Content: "账号已选前缀不得泄漏到本次请求"}}); err != nil {
			return err
		}
		_, _, err := ApplyStoryboardPreferencesAction(v, StoryboardPreferencesAction{ExpectedRevision: &v.Revision, Action: "prefix_select", PresetID: "saved"})
		return err
	}); err != nil {
		t.Fatal(err)
	}

	prefix := "本次弹窗前缀"
	if _, err := service.Generate(ctx, StoryboardRequest{
		IdempotencyKey: "test-prefix-extra-1",
		Owner:          "alice",
		ProjectID:      "p1",
		Engine:         EngineXiaoluoV7,
		Content:        "# 第1集\n开场",
		Prefix:         &prefix,
		ExtraContent:   "本次补充要求",
	}); err != nil {
		t.Fatal(err)
	}
	if len(generator.requests) != 1 {
		t.Fatalf("model calls = %d, want 1", len(generator.requests))
	}
	sent := generator.requests[0]
	if !strings.HasPrefix(sent.SystemPrompt, prefix+"\n\n") || strings.Contains(sent.SystemPrompt, "账号已选前缀不得泄漏到本次请求") {
		t.Fatalf("request prefix did not override saved preference: %q", sent.SystemPrompt)
	}
	if !strings.Contains(sent.UserPrompt, "【额外内容】\n本次补充要求") {
		t.Fatalf("extra content missing from storyboard prompt: %q", sent.UserPrompt)
	}
}
