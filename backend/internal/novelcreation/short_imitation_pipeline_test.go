package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/batchfactoryv11"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

var imitationDeconstructionFixture = "# 拆书报告\n## 一、核心情绪曲线\n压抑到释然\n## 二、人物关系与人设\n追求真相的记者\n## 三、核心价值观\n守住诚实\n## 四、大众底层情绪渴望\n被理解\n## 五、底层核心逻辑\n冲突到解决\n## 六、金句摘录\n只分析手法"
var imitationDirectionsFixture = "=== DIRECTION 1 ===\n题材：现实\n设定：渡口的旧信\n概要：追查线索\n价值观沿用：诚实\n主爽点：真相\n主虐点：误解\n=== DIRECTION 2 ===\n题材：历史\n设定：失踪的账本\n概要：寻找证据\n价值观沿用：诚实\n主爽点：揭晓\n主虐点：孤立\n=== DIRECTION 3 ===\n题材：幻想\n设定：遗忘之城\n概要：找回记忆\n价值观沿用：诚实\n主爽点：团聚\n主虐点：分别"
var imitationBlueprintFixture = "### 1. 放大后的情绪曲线\n怀疑到释然\n### 2. 新人物关系\n陌生人共同寻找真相\n### 3. 新主线结构\n情绪功能→新事件映射：恢复信任，原作遗失物，新作渡口查信\n### 4. 新金句清单\n重新创作句子\n### 5. 角色与命名\n于灯\n### 6. 分段规划\n场景1 200字，场景2 200字，场景3 200字，场景4 200字，场景5 200字\n### 7. 防抄袭对照\n改换人名、地名和事件"

func imitationBlueprintReply(target int) string {
	scenes := []map[string]any{}
	for i := 1; i <= 5; i++ {
		scenes = append(scenes, map[string]any{"name": fmt.Sprintf("场景%d", i), "expectedWords": target / 5})
	}
	raw, _ := json.Marshal(map[string]any{"content": strings.ReplaceAll(imitationBlueprintFixture, "200字", fmt.Sprintf("%d字", target/5)), "scenes": scenes, "eventMappings": []map[string]string{{"emotionFunction": "恢复信任", "originalEvent": "遗失物", "newEvent": "渡口查信"}}})
	return string(raw)
}
func imitationStageReplies() []string {
	return []string{shortControlText(imitationDeconstructionFixture), shortControlText(imitationDirectionsFixture), imitationBlueprintReply(1000), `{"title":"新故事","summary":"揭开渡口真相","chapters":[{"number":1,"title":"旧信","summary":"找到证据并消除误解"}]}`, `{"score":95,"issues":[]}`, shortControlText(strings.Repeat("完整正文。", 200)), `{"score":95,"issues":[]}`, shortControlText("推荐"), `{"content":"=== REVIEW START ===\n## 总评级\n可过审\n## 逐维度审核\n完成\n## 标题建议\n渡口旧信\n## 一句话简介\n旧信揭晓真相\n## 修改清单\n无\n=== REVIEW END ===","rating":"可过审","titleSuggestions":["渡口旧信","旧信与真相"],"synopsis":"旧信揭晓真相","issues":[]}`}
}
func TestNewShortImitationFreezesDedicatedProtocol(t *testing.T) {
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), nil)
	out, e := s.StartShortStory(context.Background(), "alice", shortInput("dedicated"))
	if e != nil {
		t.Fatal(e)
	}
	var snap map[string]any
	_ = json.Unmarshal(out.Run.PromptSnapshot, &snap)
	if snap["short_imitation_version"] != float64(2) {
		t.Fatal("新仿写任务未冻结专用阶段协议")
	}
}
func TestShortImitationDedicatedStagesReachBodyAndArtifacts(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, imitationStageReplies(), a)
	in.Session.Mode = "short_imitation"
	_ = json.Unmarshal([]byte(`{"ShortImitationVersion":1}`), &in)
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
	if e != nil || out.Status != "completed" {
		t.Fatal("专用仿写阶段未接通", e, out.Status)
	}
	if len(provider.requests) != 9 {
		t.Fatal("调用步骤缺失或重复", len(provider.requests))
	}
	for _, i := range []int{3, 5, 8} {
		raw, _ := json.Marshal(provider.requests[i])
		if !strings.Contains(string(raw), "渡口的旧信") || !strings.Contains(string(raw), "守住诚实") {
			t.Fatalf("第%d步未使用前置材料", i)
		}
	}
	raw, _ := json.Marshal(out.Artifacts)
	for _, key := range []string{"imitation/deconstruction.md", "imitation/directions.md", "imitation/blueprint.md", "reviews/editorial.md"} {
		if !strings.Contains(string(raw), key) {
			t.Fatal("缺少可查看的阶段文件", key)
		}
	}
}

func TestShortImitationCancelBetweenStagesStopsRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, imitationStageReplies(), a)
	in.ShortImitationVersion = 1
	in.Session.Mode = "short_imitation"
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	saved := 0
	in.SaveShortStage = func(context.Context, ShortStory) error { saved++; cancel(); return nil }
	_, e := p.buildShortStory(ctx, in, &recordingEventSink{}, story, false)
	if e == nil || len(provider.requests) != 1 || saved != 1 {
		t.Fatal("取消后继续请求", e, len(provider.requests), saved)
	}
}
func TestShortImitationEditorialFailureKeepsManuscriptAndReport(t *testing.T) {
	replies := imitationStageReplies()
	replies[8] = strings.ReplaceAll(replies[8], "可过审", "需修改")
	replies[8] = strings.Replace(replies[8], `"issues":[]`, `"issues":["结尾逻辑需要修改"]`, 1)
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, _, in, story := controlledShortFixture(t, replies, a)
	in.ShortImitationVersion = 1
	in.Session.Mode = "short_imitation"
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	saved := ShortStory{}
	in.SaveShortStage = func(_ context.Context, v ShortStory) error { saved = v; return nil }
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
	if e == nil || out.Status == "completed" || len(saved.Chapters) != 1 || saved.Imitation == nil || saved.Imitation.Editorial.Rating != "需修改" {
		t.Fatal("终审失败丢正文或假通过", e, out.Status)
	}
}
func TestShortImitationInvalidReportsCannotProceed(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, []string{shortControlText("只有一段笼统分析"), shortControlText("仍然没有六维")}, a)
	in.ShortImitationVersion = 1
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	if _, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false); e == nil || len(provider.requests) != 2 {
		t.Fatal("无效拆书进入后续或无限修复", e, len(provider.requests))
	}
}

func TestShortImitationConfirmedMissingMaterialsDoesNotCallModel(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, imitationStageReplies(), a)
	in.ShortImitationVersion = 1
	story.Input.Mode = "short_imitation"
	if _, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); e == nil || len(provider.requests) != 0 {
		t.Fatal("确认后的缺失资料没有提前阻止", e)
	}
}

func TestShortImitationUsesFrozenAdminPromptsAndOwnModelRoutes(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	a.ModelRoutes = map[string]string{"foundation-plan": "own-plan", "chapter-draft": "own-writer", "chapter-audit": "own-review", "memory-index": "own-memory"}
	p, provider, in, story := controlledShortFixture(t, imitationStageReplies(), a)
	in.ShortImitationVersion = 1
	in.Session.Mode = "short_imitation"
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	in.SystemPrompts = &SystemPromptSnapshot{Version: 1, Entries: map[string]FrozenSystemPrompt{}}
	for _, d := range prompt.AdminPrompts() {
		in.SystemPrompts.Entries[d.ID] = FrozenSystemPrompt{ID: d.ID, Version: 1, Revision: 7, Body: "发布版本测试-" + d.ID + "\n" + d.Definition.SystemPrompt}
	}
	routes := []string{}
	p.resolveProvider = func(_ context.Context, _ string, id string) (batchfactoryv11.DirectorProvider, error) {
		routes = append(routes, id)
		return provider, nil
	}
	if _, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false); e != nil {
		t.Fatal(e)
	}
	for index, id := range map[int]string{0: prompt.ShortImitationDeconstruction.ID, 1: prompt.ShortImitationDirections.ID, 2: prompt.ShortImitationBlueprint.ID, 5: prompt.ShortImitationWritingID, 8: prompt.ShortImitationEditorial.ID} {
		if !strings.Contains(provider.requests[index].SystemPrompt, "发布版本测试-"+id) {
			t.Fatal("未使用后台已冻结正文", index, id)
		}
	}
	if routes[0] != "own-plan" || routes[5] != "own-writer" || routes[8] != "own-review" {
		t.Fatal("没有用自家分步骤模型", routes)
	}
}

func TestShortImitationLegacySnapshotKeepsOriginalCalls(t *testing.T) {
	replies := imitationStageReplies()
	legacy := append([]string{shortControlText("旧版参考分析")}, replies[3:8]...)
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, legacy, a)
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	in.Session.Mode = "short_imitation"
	out, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
	if e != nil || out.Status != "completed" || len(provider.requests) != 6 || out.Imitation != nil {
		t.Fatal("旧任务被插入新步骤", e, len(provider.requests))
	}
}

func TestShortImitationRepairCannotInheritMissingReport(t *testing.T) {
	first := strings.TrimSuffix(shortControlText(imitationDeconstructionFixture), "}") + `,"unexpected":true}`
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	p, provider, in, story := controlledShortFixture(t, []string{first, `{}`}, a)
	in.ShortImitationVersion = 1
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	if _, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false); e == nil || len(provider.requests) != 2 {
		t.Fatal("格式修复继承了上一份报告", e, len(provider.requests))
	}
}

func TestShortImitationRetryAfterEditorialRateLimitReusesSuccessfulCalls(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	responses := imitationStageReplies()
	responses = append(responses, responses[8])
	provider := &queuedPipelineProvider{responses: responses, errors: make([]error, 9)}
	provider.errors[8] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("late-retry")
	input.TargetWordCount = 1000
	out, e := s.StartShortStory(ctx, "alice", input)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if !IsCode(e, ErrModelRateLimited) || run.Status != RunFailed {
		t.Fatal("未到终审限流", run.Status, e)
	}
	next, e := s.RetryRun(ctx, "alice", run.ID)
	if e != nil {
		t.Fatal(e)
	}
	run, e = s.Run(ctx, "alice", next.ID)
	if e != nil || run.Status != RunSucceeded || len(provider.requests) != 10 {
		t.Fatalf("末段重试重复生成已完成正文: status=%s calls=%d err=%v", run.Status, len(provider.requests), e)
	}
}

func TestShortImitationRejectsBlueprintWordBudgetAndReusedEvents(t *testing.T) {
	for _, kind := range []string{"words", "events"} {
		t.Run(kind, func(t *testing.T) {
			replies := imitationStageReplies()
			if kind == "words" {
				replies[2] = strings.Replace(replies[2], `"expectedWords":200`, `"expectedWords":999`, 1)
			} else {
				replies[2] = strings.Replace(replies[2], `"newEvent":"渡口查信"`, `"newEvent":"遗失物"`, 1)
			}
			replies = append(replies[:3], replies[2])
			a := NormalizeAgentSettings(AgentSettings{}, 2000)
			p, provider, in, story := controlledShortFixture(t, replies, a)
			in.ShortImitationVersion = 1
			story.Input.Mode = "short_imitation"
			story.Input.ReferenceText = strings.Repeat("参考", 100)
			if _, e := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false); e == nil || len(provider.requests) != 4 {
				t.Fatal("无效蓝图进入正文", e, len(provider.requests))
			}
		})
	}
}

func TestShortImitationInvalidSecondChapterRetryKeepsFirstChapterCalls(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	parts := imitationStageReplies()
	responses := append([]string{}, parts[:3]...)
	responses[2] = imitationBlueprintReply(2000)
	responses = append(responses, `{"title":"两章故事","summary":"完整闭环","chapters":[{"number":1,"title":"第一章","summary":"寻找线索"},{"number":2,"title":"第二章","summary":"真相揭晓"}]}`, parts[4], parts[5], parts[6], shortControlText("第一章已发生事实"), `{}`, `{}`, parts[5], parts[6], parts[7], parts[8])
	provider := &queuedPipelineProvider{responses: responses}
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("second-chapter-retry")
	input.TargetWordCount = 2000
	input.TargetChapters = 2
	out, e := s.StartShortStory(ctx, "alice", input)
	if e != nil {
		t.Fatal(e)
	}
	first, e := s.Run(ctx, "alice", out.Run.ID)
	if !IsCode(e, ErrInvalidOutput) || first.Status != RunFailed || len(provider.requests) != 10 {
		t.Fatal("未在第二章格式失败", first.Status, e, len(provider.requests))
	}
	next, e := s.RetryRun(ctx, "alice", first.ID)
	if e != nil {
		t.Fatal(e)
	}
	last, e := s.Run(ctx, "alice", next.ID)
	if e != nil || last.Status != RunSucceeded || len(provider.requests) != 14 {
		t.Fatal("重试重新调用了已成功的第一章", last.Status, len(provider.requests), e)
	}
}
