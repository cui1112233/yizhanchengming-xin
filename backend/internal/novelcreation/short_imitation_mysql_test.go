package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestShortImitationOwnedMySQLDefaultsCheckpointAndReopen(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	st := NewMySQLStore(db)
	provider := &queuedPipelineProvider{responses: imitationStageReplies()}
	makeService := func() *Service {
		return NewService(NewMySQLStore(db), NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	}
	s := makeService()
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	a.PassScore = 90
	ids := []int{1, 32}
	a.ReviewDimensionIDs = &ids
	if _, e := s.UpdateAgentDefaults(ctx, "stage-owner", a); e != nil {
		t.Fatal(e)
	}
	s = makeService()
	got, e := s.GetAgentDefaults(ctx, "stage-owner")
	if e != nil || got.PassScore != 90 || got.ReviewDimensionIDs == nil || len(*got.ReviewDimensionIDs) != 2 {
		t.Fatal("重开丢账号设置", e)
	}
	input := shortInput("owned-stage-reopen")
	input.TargetWordCount = 1000
	input.PauseAfterOutline = true
	out, e := s.StartShortStory(ctx, "stage-owner", input)
	if e != nil {
		t.Fatal(e)
	}
	freezeLegacyShortFullDraft(t, s, "stage-owner", out.Run.ID)
	first, e := s.Run(ctx, "stage-owner", out.Run.ID)
	if e != nil || first.Status != RunAwaitingConfirmation {
		t.Fatal(first.Status, e)
	}
	if len(provider.requests) != 3 {
		t.Fatal("暂停前步骤不对", len(provider.requests))
	}
	s = makeService()
	story, e := s.GetShortStory(ctx, "stage-owner", out.Project.ID)
	if e != nil || story.Checkpoint != "blueprint" || story.Outline != nil || story.Imitation == nil || story.Imitation.Blueprint == "" || len(story.Artifacts) < 4 {
		t.Fatal("中间资料没有落库", e)
	}
	editedBlueprint := story.Imitation.Blueprint + "\n持久化人工补充：灯塔藏着最后一封信。"
	story, e = s.EditShortStoryArtifact(ctx, "stage-owner", out.Project.ID, story.Version, "imitation/blueprint.md", editedBlueprint)
	if e != nil {
		t.Fatal(e)
	}
	s = makeService()
	story, e = s.GetShortStory(ctx, "stage-owner", out.Project.ID)
	if e != nil || story.Imitation.Blueprint != editedBlueprint {
		t.Fatal("重开丢失蓝图编辑", e)
	}
	next, e := s.ConfirmShortStory(ctx, "stage-owner", out.Run.ID, story.Version)
	if e != nil {
		t.Fatal(e)
	}
	s = makeService()
	last, e := s.Run(ctx, "stage-owner", next.ID)
	if e != nil || last.Status != RunSucceeded {
		t.Fatal(last.Status, e)
	}
	s = makeService()
	story, e = s.GetShortStory(ctx, "stage-owner", out.Project.ID)
	if e != nil || story.Status != "completed" || story.Imitation.Editorial == nil || story.Imitation.Editorial.Rating != "可过审" || len(story.Chapters) != 1 {
		t.Fatal("成品或终审没有持久化", e)
	}
	if len(provider.requests) != 9 {
		t.Fatal("恢复重复调用已完成步骤", len(provider.requests))
	}
	for _, index := range []int{3, 5, 8} {
		raw, _ := json.Marshal(provider.requests[index])
		if !strings.Contains(string(raw), "灯塔藏着最后一封信") {
			t.Fatal("后续步骤未使用人工蓝图", index)
		}
	}
	before := len(provider.requests)
	if _, e = s.Run(ctx, "stage-owner", next.ID); e != nil || len(provider.requests) != before {
		t.Fatal("重复执行重新计费", e)
	}
	if _, e = st.GetProject(ctx, "other-owner", out.Project.ID); e == nil {
		t.Fatal("账号隔离失败")
	}
	var raw string
	if e = db.QueryRow(`SELECT JSON_EXTRACT(prompt_snapshot_json,'$.project_settings.agent.passScore') FROM novel_runs WHERE owner_username=? AND run_id=?`, "stage-owner", out.Run.ID).Scan(&raw); e != nil || !strings.Contains(raw, "90") {
		t.Fatal("MySQL任务没有冻结实际设置", raw, e)
	}
	responses := imitationStageReplies()
	responses = append(responses, responses[8])
	provider = &queuedPipelineProvider{responses: responses, errors: make([]error, 9)}
	provider.errors[8] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
	input.IdempotencyKey = "owned-editorial-retry"
	input.PauseAfterOutline = false
	s = makeService()
	out, e = s.StartShortStory(ctx, "stage-owner", input)
	if e != nil {
		t.Fatal(e)
	}
	freezeLegacyShortFullDraft(t, s, "stage-owner", out.Run.ID)
	first, e = s.Run(ctx, "stage-owner", out.Run.ID)
	if !IsCode(e, ErrModelRateLimited) || first.Status != RunFailed {
		t.Fatal("未触发终审限流", first.Status, e)
	}
	next, e = s.RetryRun(ctx, "stage-owner", first.ID)
	if e != nil {
		t.Fatal(e)
	}
	s = makeService()
	last, e = s.Run(ctx, "stage-owner", next.ID)
	if e != nil || last.Status != RunSucceeded || len(provider.requests) != 10 {
		t.Fatal("MySQL重开重试重复生成", last.Status, len(provider.requests), e)
	}
	t.Log("account defaults, checkpoint/reopen, completed artifacts, terminal429 retry and completed-call reuse verified in owned MySQL; 19 fake provider calls, zero paid requests")
}
