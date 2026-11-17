package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func readZIPSceneMessage(t *testing.T, user string) (string, map[string]any) {
	t.Helper()
	if !strings.HasPrefix(user, "【章节原文】\n") {
		t.Fatal("missing ZIP chapter reference envelope", user)
	}
	original, rest, ok := strings.Cut(strings.TrimPrefix(user, "【章节原文】\n"), "\n\n【当前剧本场景JSON（仅1个场景）】\n")
	if !ok {
		t.Fatal("missing single scene envelope", user)
	}
	raw, _, ok := strings.Cut(rest, "\n\n【处理要求】\n")
	if !ok {
		t.Fatal("missing requirement envelope", user)
	}
	var scene map[string]any
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&scene); err != nil {
		t.Fatal(err)
	}
	return original, scene
}

func TestZIPSceneMessageUsesBoundChapterOriginalAndOneStructuredScene(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	s.Authoring = &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	raw := `{"title":"测试","scenes":[{"scene_number":"1-1","location":"屋内","script_content":[{"element_id":"ID-A","element_type":"action","description":"甲开窗"}],"_source_scene_index":0,"_internal":"hidden"},{"scene_number":"1-2","location":"车站","script_content":[{"element_id":"ID-B","element_type":"action","description":"乙上车"}]}]}`
	script, err := professionalScriptMarkdown(raw)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Authoring.Save(context.Background(), "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "bound", Title: "本章", SourceText: "真正小说原文：甲开窗，乙上车。", ScriptText: script, RawResult: raw}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	g.responses = []string{zip164JSON, zip164JSON}
	_, err = s.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "zip-scene-message", Owner: "alice", ProjectID: "p1", PromptID: "164", ChapterID: "bound"})
	if err != nil {
		t.Fatal(err)
	}
	if len(g.requests) != 2 {
		t.Fatal("scene request count", len(g.requests))
	}
	for i, r := range g.requests {
		original, scene := readZIPSceneMessage(t, r.UserPrompt)
		if original != "真正小说原文：甲开窗，乙上车。" {
			t.Fatal("script substituted for chapter original", original)
		}
		if scene["scene_number"] != []string{"1-1", "1-2"}[i] {
			t.Fatal("wrong current scene", scene)
		}
		if _, ok := scene["_internal"]; ok {
			t.Fatal("internal metadata leaked")
		}
		if _, ok := scene["_source_scene_index"]; ok {
			t.Fatal("source metadata leaked")
		}
		serialized, _ := json.Marshal(scene)
		if strings.Contains(string(serialized), []string{"ID-B", "ID-A"}[i]) || !strings.Contains(string(serialized), []string{"ID-A", "ID-B"}[i]) {
			t.Fatal("source elements mixed", scene)
		}
		if !strings.Contains(r.SystemPrompt, script) {
			t.Fatal("164 script variable lost")
		}
	}
}

func TestZIPSceneMessagePlainInputStaysWithinItsEpisode(t *testing.T) {
	s, _, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{zip164JSON}
	in := StoryboardRequest{IdempotencyKey: "zip-multi-original", Owner: "alice", ProjectID: "p1", PromptID: "164", Episodes: map[int]string{1: "## 1-1 屋内\n甲开窗。\n## 1-2 门外\n甲关门。", 2: "## 2-1 车站\n乙上车。"}}
	if _, err := s.Generate(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if len(g.requests) != 3 {
		t.Fatal(len(g.requests))
	}
	for i, r := range g.requests {
		original, scene := readZIPSceneMessage(t, r.UserPrompt)
		ep := 1
		if i == 2 {
			ep = 2
		}
		if original != in.Episodes[ep] {
			t.Fatal("chapter reference crossed episodes", original)
		}
		if _, ok := scene["content"].(string); !ok {
			t.Fatal("plain source not wrapped", scene)
		}
	}
}

func TestZIPSceneMessageRetryOnlyFailedSceneUsesIdenticalMessage(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{zip164JSON, "not-json", zip164JSON}
	in := StoryboardRequest{IdempotencyKey: "zip-msg-partial", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "# 第1集\n## 1-1 房间\n甲开窗。\n## 1-2 门外\n甲离开。"}
	if _, err := s.Generate(context.Background(), in); err == nil || len(g.requests) != 2 {
		t.Fatal(err, len(g.requests))
	}
	state, err := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := state.Attempts[in.IdempotencyKey].Episodes[1].Calls["storyboard_scene_1"]
	if checkpoint.Status != "received" || checkpoint.RawResponse != zip164JSON {
		t.Fatal("successful scene not checkpointed", checkpoint)
	}
	in.RetryOf = in.IdempotencyKey
	in.IdempotencyKey = "zip-msg-retry"
	in.ConfirmedRetry = true
	if _, err = s.Generate(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if len(g.requests) != 3 || g.requests[1].UserPrompt != g.requests[2].UserPrompt || g.requests[1].SystemPrompt != g.requests[2].SystemPrompt {
		t.Fatal("retry repeated success or changed failed scene")
	}
	saved, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, "prompt_164")
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Shots) != 2 || saved.Shots[0].SourceSceneIndex != 0 || saved.Shots[1].SourceSceneIndex != 1 {
		t.Fatal("scene ownership lost", saved.Shots)
	}
}

func TestZIPSceneMessageChangedOriginalCannotReplayOldResult(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	s.Authoring = &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	state, err := s.Authoring.Save(context.Background(), "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "source-change", Title: "本章", SourceText: "旧原文", ScriptText: "当前剧本"}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	g.responses = []string{zip164JSON}
	req := StoryboardRequest{IdempotencyKey: "zip-msg-source-key", Owner: "alice", ProjectID: "p1", ChapterID: "source-change", PromptID: "164"}
	if _, err = s.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	state.Chapters[0].SourceText = "用户修改后的原文"
	if _, err = s.Authoring.Save(context.Background(), "alice", "p1", state, state.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Generate(context.Background(), req); !errors.Is(err, ErrRevisionConflict) || g.callCount() != 1 {
		t.Fatal("stale request replayed or rebilled", err, g.callCount())
	}
}

func TestZIPSceneMessageDirectOverrideNeverBorrowsUnrelatedOriginal(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	s.Authoring = &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
	if _, err := s.Authoring.Save(context.Background(), "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "other", Title: "其他", SourceText: "不应注入的旧小说", ScriptText: "不相关剧本"}}}, 0); err != nil {
		t.Fatal(err)
	}
	g.responses = []string{zip164JSON}
	if _, err := s.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "zip-direct-override", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "用户新输入的内容"}); err != nil {
		t.Fatal(err)
	}
	original, _ := readZIPSceneMessage(t, g.requests[0].UserPrompt)
	if original != "用户新输入的内容" || strings.Contains(g.requests[0].UserPrompt, "不应注入") {
		t.Fatal("borrowed unrelated original", original)
	}
}

func TestZIPSceneMessageCancellationDoesNotSaveLateScene(t *testing.T) {
	s, store, _ := newStoryboardService(t, PermissionBypass)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	s.Generator = storyboardGenerationTestFunc(func(_ context.Context, in TextRequest) (TextResponse, error) {
		calls++
		readZIPSceneMessage(t, in.UserPrompt)
		if calls == 2 {
			cancel()
		}
		return TextResponse{Content: zip164JSON}, nil
	})
	_, err := s.Generate(ctx, StoryboardRequest{IdempotencyKey: "zip-msg-cancel", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "# 第1集\n## 1-1 屋内\n甲开窗。\n## 1-2 门外\n甲离开。"})
	if !errors.Is(err, context.Canceled) || calls != 2 {
		t.Fatal("cancel not reported", err, calls)
	}
	rows, _ := store.ListStoryboards(context.Background(), "alice", "p1", "prompt_164")
	if len(rows) != 0 {
		t.Fatal("late result saved")
	}
	state, _ := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
	if state.Attempts["zip-msg-cancel"].Episodes[1].Calls["storyboard_scene_1"].RawResponse != zip164JSON {
		t.Fatal("successful checkpoint lost")
	}
}
