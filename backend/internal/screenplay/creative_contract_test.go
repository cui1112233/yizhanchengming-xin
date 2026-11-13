package screenplay

import (
	"context"
	"errors"
	"qiantie/backend/internal/prompt"
	"strings"
	"testing"
)

func TestCreativeChapterSelectedPromptAndLosslessJSON(t *testing.T) {
	req, err := BuildChapterTextRequest(Project{PromptID: prompt.CreativeScript144, StyleDesc: "水墨风格"}, Chapter{SourceText: "原文全文"}, "保留结尾", `{"assets":[{"type":"character","name":"甲","description":"主角"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(req.SystemPrompt, "{ROLES_PLACEHOLDER}") || !strings.Contains(req.SystemPrompt, "甲：主角") || !strings.Contains(req.SystemPrompt, "零遗漏") || !strings.Contains(req.UserPrompt, "水墨风格") || !strings.Contains(req.UserPrompt, "原文全文") || !req.JSONResponse {
		t.Fatalf("request=%+v", req)
	}
	raw := ` {"episode_index":1,"title":"测试","scenes":[{"scene_number":"1-1","location":"家","time":"日","indoor_outdoor":"内","characters":["甲"],"script_content":[{"element_type":"action","description":"门开了","extra":"保留"},{"element_type":"dialogue","character":"甲","content":"完整尾句。","emotion":"惊讶","is_voiceover":true}]}]} `
	chapter := Chapter{}
	if err := ApplyChapterGeneration(&chapter, raw); err != nil {
		t.Fatal(err)
	}
	if chapter.RawResult != raw || !strings.Contains(chapter.ScriptText, "完整尾句。") || !strings.Contains(chapter.ScriptText, "门开了") {
		t.Fatalf("chapter=%+v", chapter)
	}
	if len(ParseScreenplayMarkdown(chapter.ScriptText).Shots) != 1 {
		t.Fatal(chapter.ScriptText)
	}
}
func TestCreativePromptWrongPurposeStopsBeforeCall(t *testing.T) {
	_, err := BuildChapterTextRequest(Project{PromptID: prompt.CreativeScene289}, Chapter{SourceText: "原文"}, "", "")
	if err == nil {
		t.Fatal("wrong purpose accepted")
	}
}
func TestCreativeStoryboardThirtySecondsDoesNotChangeLegacy(t *testing.T) {
	shots := []StoryboardShot{{Seq: 1, Duration: 30, Scene: "家"}}
	if len(validateStoryboardGroups(EngineXiaoluoV9, 1, shots, knownNames{})) != 0 {
		t.Fatal("V9 rejects 30s")
	}
	if len(validateStoryboardGroups(EngineXiaoluoV7, 1, shots, knownNames{})) == 0 {
		t.Fatal("V7 accepts 30s")
	}
}
func TestCreativeStoryboardSelected307FinalRequest(t *testing.T) {
	svc, store, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{`{"storyboard":[{"id":1,"scene_id":1,"duration_sec":30,"characters":"甲","scene":"家","props":"无","video_desc":"完整视频描述"}]}`}
	p, err := store.GetProject(context.Background(), "alice", "p1")
	if err != nil {
		t.Fatal(err)
	}
	p.StyleDesc = "独立水墨画风"
	if _, err = store.UpdateProject(context.Background(), "alice", "p1", p, p.Revision); err != nil {
		t.Fatal(err)
	}
	out, err := svc.Generate(context.Background(), StoryboardRequest{Owner: "alice", ProjectID: "p1", Engine: EngineXiaoyunqueV3, PromptID: prompt.CreativeStoryboard307, Content: "# 第1集：家\n角色甲开门。"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Storyboards) != 1 || out.Storyboards[0].Engine != EngineXiaoluoV9 || out.Storyboards[0].Shots[0].Duration != 30 {
		t.Fatalf("out=%+v", out)
	}
	req := g.requests[0]
	if !strings.Contains(req.SystemPrompt, "独立水墨画风") || !strings.Contains(req.SystemPrompt, "角色甲开门。") || !strings.Contains(req.SystemPrompt, "三十秒") {
		t.Fatal("selected template, style or source missing from final request")
	}
}
func TestCreativeWrongStoryboardPurposeNoProviderCall(t *testing.T) {
	svc, _, g := newStoryboardService(t, PermissionBypass)
	_, err := svc.Generate(context.Background(), StoryboardRequest{Owner: "alice", ProjectID: "p1", PromptID: prompt.CreativeCharacter213, Content: "# 第1集：家\n角色甲"})
	if err == nil || g.callCount() != 0 {
		t.Fatalf("err=%v calls=%d", err, g.callCount())
	}
}

type lateCanceledGenerator struct {
	cancel  context.CancelFunc
	content string
}

func (g lateCanceledGenerator) Generate(context.Context, TextRequest) (TextResponse, error) {
	g.cancel()
	return TextResponse{Content: g.content}, nil
}
func TestCreativeCanceledLateResponseDoesNotSave(t *testing.T) {
	for _, mode := range []string{"a1", "b", "storyboard"} {
		t.Run(mode, func(t *testing.T) {
			store := NewMemoryStore()
			_, err := store.CreateProject(context.Background(), Project{ID: "p", OwnerUsername: "o", PermissionMode: PermissionBypass})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			content := sampleScreenplay
			if mode == "storyboard" {
				content = storyboardJSON
			}
			g := lateCanceledGenerator{cancel: cancel, content: content}
			var runErr error
			switch mode {
			case "a1":
				_, runErr = (&ScreenplayService{Store: store, Generator: g}).GenerateA1(ctx, A1Request{Owner: "o", ProjectID: "p", Content: "原文"})
			case "b":
				_, runErr = (&ScreenplayService{Store: store, Generator: g}).GenerateB(ctx, BRequest{Owner: "o", ProjectID: "p", Novel: "原文"})
			case "storyboard":
				_, runErr = (&StoryboardService{Store: store, Generator: g}).Generate(ctx, StoryboardRequest{Owner: "o", ProjectID: "p", Content: "# 第1集：原文\n开门"})
			}
			if !errors.Is(runErr, context.Canceled) {
				t.Fatalf("error=%v", runErr)
			}
			docs, _ := store.ListDocuments(context.Background(), "o", "p")
			boards, _ := store.ListStoryboards(context.Background(), "o", "p", "")
			if len(docs) != 0 || len(boards) != 0 {
				t.Fatal("canceled response was saved")
			}
		})
	}
}
func TestCreativeProfessionalDocumentKeepsRawAndEditable(t *testing.T) {
	raw := ` {"episode_index":1,"title":"测试","scenes":[{"scene_number":"1-1","location":"家","time":"日","indoor_outdoor":"内","characters":["甲"],"script_content":[{"element_type":"action","description":"门开了"},{"element_type":"dialogue","character":"甲","content":"完整尾句。"}]}]} `
	store := NewMemoryStore()
	_, _ = store.CreateProject(context.Background(), Project{ID: "p", OwnerUsername: "o", PromptID: prompt.CreativeScript144, PermissionMode: PermissionBypass})
	g := &fakeGenerator{responses: []string{raw}}
	result, err := (&ScreenplayService{Store: store, Generator: g}).GenerateB(context.Background(), BRequest{Owner: "o", ProjectID: "p", Novel: "原文"})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := store.ActiveDocument(context.Background(), "o", "p")
	if err != nil {
		t.Fatal(err)
	}
	if saved.RawResult != raw || !strings.HasPrefix(saved.Content, "# 第1集") || len(result.Shots) != 1 {
		t.Fatalf("raw/editable not saved separately: %+v", saved)
	}
}
func TestCreativeMalformedProfessionalResultDoesNotOverwrite(t *testing.T) {
	for _, raw := range []string{`{"scenes":[]}`, `{"scenes":[{"location":"家","script_content":[]}]}`, `{"scenes":[{"location":"家","script_content":[{"element_type":"action","description":""}]}]}`, `{"scenes":[{"location":"家","script_content":[{"element_type":"dialogue","character":"甲","content":["错误数组"]}]}]}`} {
		c := Chapter{ScriptText: "原成功剧本", RawResult: "原结果"}
		if ApplyChapterGeneration(&c, raw) == nil || c.ScriptText != "原成功剧本" || c.RawResult != "原结果" {
			t.Fatal("invalid result overwrote previous chapter")
		}
	}
}
func TestCreativeSelectedLegacyStoryboardCannotUseV9Validation(t *testing.T) {
	for _, tc := range []struct {
		id     string
		engine StoryboardEngine
	}{{"storyboard_xiaoluo_v7_15s", EngineXiaoluoV7}, {"storyboard_xiaoyunque_v3", EngineXiaoyunqueV3}} {
		for _, source := range []string{"request", "project"} {
			t.Run(tc.id+"/"+source, func(t *testing.T) {
				svc, store, g := newStoryboardService(t, PermissionBypass)
				g.responses = []string{`{"storyboard":[{"id":1,"scene_id":1,"duration_sec":30,"characters":"甲","scene":"家","props":"无","video_desc":"画风: 水墨\n完整视频描述"}]}`}
				req := StoryboardRequest{Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV9, Content: "# 第1集：家\n角色甲开门。"}
				if source == "request" {
					req.PromptID = tc.id
				} else {
					p, err := store.GetProject(context.Background(), "alice", "p1")
					if err != nil {
						t.Fatal(err)
					}
					if _, err = store.SaveCreativeConfig(context.Background(), "alice", "p1", CreativeConfig{StoryboardPromptID: tc.id}, p.Revision); err != nil {
						t.Fatal(err)
					}
				}
				out, err := svc.Generate(context.Background(), req)
				if err != nil {
					t.Fatal(err)
				}
				if len(out.Storyboards) != 1 || out.Storyboards[0].Engine != tc.engine {
					t.Fatalf("selected prompt stored under wrong protocol: %+v", out)
				}
				if !strings.Contains(strings.Join(out.Warnings, "|"), "15s") {
					t.Fatalf("legacy 30s output missing legacy duration warning: %v", out.Warnings)
				}
				expected, err := prompt.ResolveCreativePrompt(tc.id, "storyboard")
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(g.requests[0].SystemPrompt, "三十秒") && !strings.Contains(expected.Body, "三十秒") {
					t.Fatal("V9 body was selected")
				}
			})
		}
	}
}
