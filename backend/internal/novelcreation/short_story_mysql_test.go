package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestShortStoryRealMySQLCheckpointAndReopen(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("short-sql")
	provider := &queuedPipelineProvider{responses: []string{`{"content":"参考写法分析"}`, `{"title":"雨夜","summary":"完整闭环","chapters":[{"number":1,"title":"追寻","summary":"真相揭晓"}]}`, `{"score":95,"issues":[]}`}}
	body, _ := json.Marshal(shortText{Content: strings.Repeat("完整正文。", 200)})
	provider.responses = append(provider.responses, string(body), `{"score":96,"issues":[]}`, `{"content":"真实推荐"}`)
	s := NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("sql-short")
	input.TargetWordCount = 1000
	input.PauseAfterOutline = true
	out, e := s.StartShortStory(ctx, owner, input)
	if e != nil {
		t.Fatal(e)
	}
	replay, e := s.StartShortStory(ctx, owner, input)
	if e != nil || replay.Run.ID != out.Run.ID {
		t.Fatal("SQL duplicate", e)
	}
	run, e := s.Run(ctx, owner, out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation {
		t.Fatal("SQL checkpoint", run.Status, e)
	}
	reopened := NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
	story, e := reopened.GetShortStory(ctx, owner, out.Project.ID)
	if e != nil || story.Outline == nil {
		t.Fatal("SQL reopen", e)
	}
	story.Outline.Summary = "SQL人工编辑"
	edited, e := reopened.EditShortStoryOutline(ctx, owner, out.Project.ID, story.Version, *story.Outline)
	if e != nil {
		t.Fatal("SQL edit", e)
	}
	next, e := reopened.ConfirmShortStory(ctx, owner, out.Run.ID, edited.Version)
	if e != nil {
		t.Fatal("SQL confirm", e)
	}
	run, e = reopened.Run(ctx, owner, next.ID)
	if e != nil || run.Status != RunSucceeded {
		t.Fatal("SQL completion", run.Status, e)
	}
	final := NewService(NewMySQLStore(db), NewMemoryRuntime(), nil)
	saved, e := final.GetShortStory(ctx, owner, out.Project.ID)
	if e != nil || saved.Status != "completed" || len(saved.Artifacts) != 3 || len(saved.Chapters) != 1 || saved.Outline.Summary != "SQL人工编辑" {
		t.Fatal("SQL persisted independent output", e)
	}
	chapters, e := final.store.ListChapters(ctx, owner, out.Project.ID)
	if e != nil || len(chapters) != 0 {
		t.Fatal("ordinary chapter pollution", e)
	}
	if _, e = final.GetShortStory(ctx, owner+"-other", out.Project.ID); e == nil {
		t.Fatal("SQL cross owner leak")
	}
	t.Log("real MySQL: create/replay/checkpoint/edit/confirm/complete/new-store-reopen passed; all model outputs are local test doubles")
}
