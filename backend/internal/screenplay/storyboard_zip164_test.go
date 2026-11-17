package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const zip164JSON = `{"storyboard":[{"id":1,"source_elements":["ID-001"],"characters":"顾苒苒","scene":"办公室","image_prompt":"图1坐下，图2起身，图3开门，图4离开。","voice_actor":"","emotion":"平静","line":"","video_desc":"图1至图4依次推进。角色映射：@顾苒苒.角色"}]}`

func TestZIP164RequestParseSaveAndRepeat(t *testing.T) {
	s, store, gen := newStoryboardService(t, PermissionBypass)
	gen.responses = []string{zip164JSON}
	req := StoryboardRequest{IdempotencyKey: "zip164-normal", Owner: "alice", ProjectID: "p1", PromptID: "164", Model: "our-configured-text-model", Content: "# 第1集：A\n顾苒苒走出办公室。"}
	if _, err := s.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if gen.callCount() != 1 {
		t.Fatalf("calls=%d", gen.callCount())
	}
	r := gen.requests[0]
	if r.Model != req.Model || !r.JSONResponse || !strings.Contains(r.SystemPrompt, "顾苒苒走出办公室。") {
		t.Fatalf("request not wired: %+v", r)
	}
	for _, token := range []string{"{{剧本内容}}", "{{前场景末尾/后场景开头分镜}}", "{ROLES_PLACEHOLDER}", "{SCENES_PLACEHOLDER}", "{PROPS_PLACEHOLDER}"} {
		if strings.Contains(r.SystemPrompt, token) {
			t.Fatalf("unrendered %s", token)
		}
	}
	saved, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, StoryboardEngine("prompt_164"))
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Shots) != 1 || saved.Shots[0].ImagePrompt == "" || saved.Shots[0].VideoDesc == "" || saved.Shots[0].Duration != 0 || string(saved.Shots[0].SourceElements) != `["ID-001"]` {
		t.Fatalf("lost original fields: %+v", saved.Shots)
	}
	if _, err := s.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if gen.callCount() != 1 {
		t.Fatal("repeat made another supplier request")
	}
}

func TestZIP164InvalidOutputDoesNotSave(t *testing.T) {
	for _, raw := range []string{strings.Replace(zip164JSON, `"image_prompt":"图1坐下，图2起身，图3开门，图4离开。",`, "", 1), strings.Replace(zip164JSON, `["ID-001"]`, `123`, 1)} {
		s, store, gen := newStoryboardService(t, PermissionBypass)
		gen.responses = []string{raw}
		_, err := s.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "zip164-invalid", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "# 第1集：A\n正文"})
		if err == nil {
			t.Fatal("invalid output accepted")
		}
		rows, listErr := store.ListStoryboards(context.Background(), "alice", "p1", StoryboardEngine("prompt_164"))
		if listErr != nil || len(rows) != 0 {
			t.Fatal("invalid result saved", listErr)
		}
	}
}

func TestZIP164ContextAndAssetsRespectSavedChoices(t *testing.T) {
	for _, historyOn := range []bool{false, true} {
		s, store, gen := newStoryboardService(t, PermissionBypass)
		ctx := context.Background()
		p, _ := store.GetProject(ctx, "alice", "p1")
		cfg := EffectiveCreativeConfig(p)
		cfg.StoryboardUseContextHistory = historyOn
		cfg.StoryboardAutoSyncAssets = false
		if _, err := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision); err != nil {
			t.Fatal(err)
		}
		if _, err := store.CreateAsset(ctx, Asset{ID: "hero164", ProjectID: "p1", OwnerUsername: "alice", Type: AssetCharacter, Degree: DegreeMaster, Name: "顾苒苒", Description: "资产独有锚点164"}); err != nil {
			t.Fatal(err)
		}
		if _, err := store.SaveStoryboard(ctx, Storyboard{ProjectID: "p1", OwnerUsername: "alice", EpisodeNo: 1, Engine: "prompt_164", Shots: []StoryboardShot{{Seq: 1, VideoDesc: "真实已保存历史164"}}}); err != nil {
			t.Fatal(err)
		}
		gen.responses = []string{zip164JSON}
		_, err := s.Generate(ctx, StoryboardRequest{IdempotencyKey: "zip164-context", Owner: "alice", ProjectID: "p1", PromptID: "164", EpisodeNo: 2, Content: "第2集\n当前集独有内容164"})
		if err != nil {
			t.Fatal(err)
		}
		system := gen.requests[0].SystemPrompt
		if !strings.Contains(system, "资产独有锚点164") || !strings.Contains(system, "当前集独有内容164") {
			t.Fatal("chapter or assets missing")
		}
		if strings.Contains(system, "真实已保存历史164") != historyOn {
			t.Fatal("context switch ignored", system)
		}
		if gen.callCount() != 1 {
			t.Fatal("unexpected extra generation")
		}
	}
}

func TestZIP164EmptyAndCancelledDoNotGenerate(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		s, store, gen := newStoryboardService(t, PermissionBypass)
		ctx := context.Background()
		content := " "
		if cancelled {
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			cancel()
			content = "第1集\n正文"
		}
		_, err := s.Generate(ctx, StoryboardRequest{IdempotencyKey: "zip164-empty-cancel", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: content})
		if err == nil || gen.callCount() != 0 {
			t.Fatal("invalid request generated", err, gen.callCount())
		}
		if cancelled && !errors.Is(err, context.Canceled) {
			t.Fatal("cancel not reported", err)
		}
		rows, _ := store.ListStoryboards(context.Background(), "alice", "p1", "prompt_164")
		if len(rows) != 0 {
			t.Fatal("cancel/empty saved results")
		}
	}
}
