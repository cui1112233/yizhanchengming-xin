package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStoryboardScenesKeepChapterContextAndRetryOnlyInvalidScene(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{storyboardJSON, "not-json", storyboardJSON}
	in := generationTestRequest("scene-parent")
	in.Content = "# 第1集\n## 1-1：办公室\n甲推门进入。\n## 1-2：车站\n乙离开。"
	first, err := s.Generate(context.Background(), in)
	if err == nil || g.callCount() != 2 || len(first.Storyboards) != 0 {
		t.Fatalf("first=%+v err=%v calls=%d", first, err, g.callCount())
	}
	if strings.Contains(g.requests[0].UserPrompt, "乙离开") || !strings.Contains(g.requests[0].SystemPrompt, "乙离开") {
		t.Fatal("current source and adjacent context must be separated")
	}
	if !strings.Contains(g.requests[1].SystemPrompt, "甲推门进入") {
		t.Fatal("same-chapter previous scene missing")
	}
	state, _ := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
	if state.Attempts[in.IdempotencyKey].Episodes[1].Calls["storyboard_scene_1"].Status != "received" {
		t.Fatal("successful scene must be checkpointed")
	}
	in.RetryOf = in.IdempotencyKey
	in.IdempotencyKey = "scene-retry"
	in.ConfirmedRetry = true
	out, err := s.Generate(context.Background(), in)
	if err != nil || g.callCount() != 3 || len(out.Storyboards) != 1 || len(out.Storyboards[0].Shots) != 4 {
		t.Fatalf("retry=%+v err=%v calls=%d", out, err, g.callCount())
	}
	if !strings.Contains(g.requests[2].UserPrompt, "乙离开") || strings.Contains(g.requests[2].UserPrompt, "甲推门进入") {
		t.Fatal("retry must only regenerate failed scene")
	}
	board, _ := store.GetStoryboard(context.Background(), "alice", "p1", 1, EngineXiaoluoV7)
	for i, shot := range board.Shots {
		if shot.Seq != i+1 {
			t.Fatal("stable total sequence")
		}
	}
	_, err = s.Generate(context.Background(), in)
	if err != nil || g.callCount() != 3 {
		t.Fatal("duplicate retry paid again", err)
	}
}

func TestStoryboardAdjacentContextDoesNotCrossEpisodes(t *testing.T) {
	s, _, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{storyboardJSON, storyboardJSON}
	in := generationTestRequest("separate-chapters")
	in.Content = ""
	in.Episodes = map[int]string{1: "第一章独有正文", 2: "第二章独有正文"}
	_, err := s.Generate(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(g.requests[0].SystemPrompt, "第二章独有正文") || strings.Contains(g.requests[1].SystemPrompt, "第一章独有正文") {
		t.Fatal("chapter is not a scene neighbour")
	}
}

func TestStoryboardSceneHeadingEmptyStopsBeforeProvider(t *testing.T) {
	s, _, g := newStoryboardService(t, PermissionBypass)
	in := generationTestRequest("empty-scene")
	in.Content = "## 1-1：空场景\n## 1-2：有效场景\n乙离开"
	_, err := s.Generate(context.Background(), in)
	if !errors.Is(err, ErrEmptyContent) || g.callCount() != 0 {
		t.Fatal(err, g.callCount())
	}
}

func TestStoryboardStructuredSceneKeepsElementIDsAndEditedTextWins(t *testing.T) {
	s, store, g := newStoryboardService(t, PermissionBypass)
	repo := NewMemoryAuthoringRepository()
	s.Authoring = &AuthoringService{Store: store, Repository: repo}
	raw := `{"episode_index":1,"title":"窗外","scenes":[{"scene_number":"1-1","location":"家","script_content":[{"element_id":"1-1-001","element_type":"action","description":"甲开窗"}]},{"scene_number":"1-2","location":"车站","script_content":[{"element_id":"1-2-001","element_type":"action","description":"乙上车"}]}]}`
	script, err := professionalScriptMarkdown(raw)
	if err != nil {
		t.Fatal(err)
	}
	state, err := s.Authoring.Save(context.Background(), "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "structured", Title: "窗外", ScriptText: script, RawResult: raw}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	g.responses = []string{storyboardJSON, storyboardJSON}
	req := generationTestRequest("structured-scenes")
	req.ChapterID = "structured"
	_, err = s.Generate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.requests) != 2 || !strings.Contains(g.requests[0].UserPrompt, "1-1-001") || strings.Contains(g.requests[0].UserPrompt, "1-2-001") {
		t.Fatal("lost source element identity")
	}
	state.Chapters[0].ScriptText = "## 1-1：改稿\n甲重新关窗"
	_, err = s.Authoring.Save(context.Background(), "alice", "p1", state, state.Revision)
	if err != nil {
		t.Fatal(err)
	}
	req.IdempotencyKey = "edited-scenes"
	_, err = s.Generate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(g.requests[2].UserPrompt, "重新关窗") || strings.Contains(g.requests[2].UserPrompt, "1-1-001") {
		t.Fatal("stale structured result overrode edited script")
	}
}
