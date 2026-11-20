package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"qiantie/backend/internal/batchfactoryv11"
	catalog "qiantie/backend/internal/prompt"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

func shortFullRaw(title string, contents ...string) string {
	raw := "=== SHORT_FICTION_TITLE ===\n" + title + "\n=== SHORT_FICTION_OPENING_HOOK ===\n渡口来了一封信。\n"
	for i, content := range contents {
		raw += fmtShortFullChapter(i+1, content)
	}
	return raw
}
func fmtShortFullChapter(number int, content string) string {
	raw, _ := json.Marshal(number)
	return "=== CHAPTER " + string(raw) + " TITLE ===\n章节\n=== CHAPTER " + string(raw) + " CONTENT ===\n" + content + "\n"
}
func shortFullPackageReply() string {
	return "=== SHORT_FICTION_PACKAGE_TITLE ===\n不应覆盖正文的包装标题\n=== SHORT_FICTION_INTRO ===\n简介内容。\n=== SHORT_FICTION_SELLING_POINTS ===\n- 卖点一\n- 卖点二\n=== SHORT_FICTION_COVER_PROMPT ===\n竖版书封提示词。"
}
func shortFullFixture(t *testing.T, replies []string, chapters int) (*Pipeline, *queuedPipelineProvider, RunContext, ShortStory) {
	t.Helper()
	p, provider, in, story := controlledShortFixture(t, replies, NormalizeAgentSettings(AgentSettings{Version: 1, MaxRevisionRounds: 0, PassScore: 100}, 2000))
	_ = json.Unmarshal([]byte(`{"ShortFullDraftVersion":1}`), &in)
	in.ShortImitationVersion = 2
	in.Session.Mode = "short_imitation"
	story.Input.Mode = "short_imitation"
	story.Input.TargetChapters = chapters
	var bp ShortImitationBlueprintOutput
	_ = json.Unmarshal([]byte(imitationBlueprintReply(story.Input.TargetWordCount)), &bp)
	directions, _ := parseImitationDirections(imitationDirectionsFixture)
	story.Imitation = &ShortImitationData{Deconstruction: imitationDeconstructionFixture, Directions: imitationDirectionsFixture, SelectedDirection: directions[0], Blueprint: bp.Content, Scenes: bp.Scenes, EventMappings: bp.EventMappings}
	story.Outline.Chapters = nil
	for i := 1; i <= chapters; i++ {
		story.Outline.Chapters = append(story.Outline.Chapters, ShortOutlineChapter{Number: i, Title: "章节", Summary: "完成情节"})
	}
	return p, provider, in, story
}
func shortFullArtifact(story ShortStory, key string) string {
	for _, a := range story.Artifacts {
		if a.Key == key {
			return a.Content
		}
	}
	return ""
}

func TestShortFullPipelineAdoptsValidSecondDraftWithoutPieceLengthGate(t *testing.T) {
	first := shortFullRaw("首稿标题", "首稿完整正文。")
	second := shortFullRaw("二稿标题", "经过修订的正文。")
	p, provider, in, story := shortFullFixture(t, []string{first, "逐维度审核意见", second, shortFullPackageReply(), imitationStageReplies()[8]}, 1)
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err != nil || out.Status != "completed" || len(provider.requests) != 5 || len(out.Chapters) != 1 || out.Chapters[0].Content != "经过修订的正文。" {
		t.Fatalf("full raw draft was treated as piece JSON or length-gated: status=%s calls=%d err=%v", out.Status, len(provider.requests), err)
	}
	if shortFullArtifact(out, "drafts/v001.md") != first || shortFullArtifact(out, "drafts/v002.md") != second {
		t.Fatal("complete draft versions not preserved")
	}
	if out.Title != "二稿标题" || out.Recommendation != "简介内容。\n\n- 卖点一\n- 卖点二" || shortFullArtifact(out, "package/cover-prompt.txt") != "竖版书封提示词。" {
		t.Fatal("package fields not mapped or packaging title replaced accepted draft title", out.Title, out.Recommendation)
	}
	messages := provider.requests[2].CompletionMessages()
	if len(messages) != 4 || messages[0].Content != provider.requests[0].SystemPrompt || messages[1].Content != provider.requests[0].UserPrompt || messages[2].Content != strings.TrimSpace(first) {
		t.Fatal("full revision did not retain original full writer conversation")
	}
	if !provider.requests[0].DisableJSONResponse || !provider.requests[2].DisableJSONResponse {
		t.Fatal("full Markdown request incorrectly requires JSON")
	}
}

func TestShortFullPipelineCompletesMissingChaptersBeforeSavingAndRevision(t *testing.T) {
	first := shortFullRaw("首稿", "第一章正文。", "")
	completion := fmtShortFullChapter(2, "补齐第二章。")
	mergedCanonical := strings.TrimSpace(first) + "\n\n" + strings.TrimSpace(completion)
	first = "\x1c" + first + "\x1f"
	completion = "\x1d" + completion + "\x1e"
	second := shortFullRaw("二稿", "修订第一章。", "修订第二章。")
	p, provider, in, story := shortFullFixture(t, []string{first, completion, "审核意见", second, shortFullPackageReply(), imitationStageReplies()[8]}, 2)
	saved := []ShortStory{}
	in.SaveShortStage = func(_ context.Context, partial ShortStory) error { saved = append(saved, partial); return nil }
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err != nil || out.Status != "completed" || len(provider.requests) != 6 {
		t.Fatal("missing chapter completion failed", out.Status, len(provider.requests), err)
	}
	firstVersion := shortFullArtifact(out, "drafts/v001.md")
	if firstVersion != mergedCanonical || provider.requests[3].Messages[2].Content != mergedCanonical {
		t.Fatal("continuation merge did not match Python strip behavior")
	}
	if !strings.Contains(firstVersion, "补齐第二章") || !strings.Contains(provider.requests[3].Messages[2].Content, "补齐第二章") {
		t.Fatal("revision used incomplete original draft")
	}
	if !strings.Contains(provider.requests[3].UserPrompt, "CHAPTER 1 CONTENT") || !strings.Contains(provider.requests[3].UserPrompt, "CHAPTER 2 CONTENT") {
		t.Fatal("revision reused only missing-chapter output template")
	}
	completeSaved := false
	for _, v := range saved {
		if shortFullArtifact(v, "drafts/v001.md") != "" && len(v.Chapters) == 2 && v.Chapters[1].Content == "补齐第二章。" {
			completeSaved = true
		}
	}
	if !completeSaved {
		t.Fatal("complete v001 was never persisted before later stages")
	}
}

func TestShortFullPipelineRejectsInvalidSecondDraftAndKeepsFirst(t *testing.T) {
	first := shortFullRaw("首稿", "完整首稿。")
	invalid := shortFullRaw("坏二稿", "")
	p, provider, in, story := shortFullFixture(t, []string{first, "审核意见", invalid, shortFullPackageReply(), imitationStageReplies()[8]}, 1)
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err != nil || out.Status != "completed" || len(out.Chapters) != 1 || out.Chapters[0].Content != "完整首稿。" || len(provider.requests) != 5 {
		t.Fatal("invalid revision destroyed usable first draft", err, out.Status)
	}
	if shortFullArtifact(out, "drafts/v002.md") != "" || shortFullArtifact(out, "drafts/v002-rejected.md") != invalid {
		t.Fatal("invalid candidate falsely adopted")
	}
	raw, _ := json.Marshal(out)
	if !strings.Contains(string(raw), `"currentVersion":"v001"`) || !strings.Contains(string(raw), `"revisionStatus":"rejected"`) || !strings.Contains(string(raw), `"revisionWarning":`) {
		t.Fatal("fallback is not explicit")
	}
}

func TestShortFullPipelineStopsAfterThreeMissingChapterAttempts(t *testing.T) {
	first := shortFullRaw("首稿", "第一章。", "")
	continuation := fmtShortFullChapter(1, "仍只有第一章。")
	p, provider, in, story := shortFullFixture(t, []string{first, continuation, continuation, continuation}, 2)
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err == nil || out.Status == "completed" || len(provider.requests) != 4 || shortFullArtifact(out, "drafts/v001.md") != "" {
		t.Fatal("incomplete draft marked complete or unbounded completion", err, len(provider.requests), out.Status)
	}
}

func TestShortFullPipelinePreservesFirstDraftOnProtectedFailures(t *testing.T) {
	first := shortFullRaw("首稿", "完整首稿。")
	for _, kind := range []string{"rate_limit", "unknown_result", "auth", "cancel", "save"} {
		t.Run(kind, func(t *testing.T) {
			p, provider, in, story := shortFullFixture(t, []string{first, "审核意见", shortFullRaw("二稿", "新正文。")}, 1)
			if kind == "save" {
				in.SaveShortStage = func(context.Context, ShortStory) error { return persistenceError(errors.New("save rejected")) }
			} else if kind == "cancel" {
				runtime := NewMemoryRuntime()
				in.Runtime = runtime
				provider.afterCall = func(n int) {
					if n == 1 {
						_ = runtime.Cancel(context.Background(), in.Run.ID)
					}
				}
			} else {
				provider.errors = make([]error, 3)
				switch kind {
				case "rate_limit":
					provider.errors[2] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limit"}
				case "auth":
					provider.errors[2] = &batchfactoryv11.DirectorHTTPError{StatusCode: 401, Body: "unauthorized"}
				case "unknown_result":
					provider.errors[2] = &DomainError{Code: ErrResultUnknown, Message: "unknown"}
				}
			}
			out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
			if err == nil || out.Status == "completed" || shortFullArtifact(out, "drafts/v001.md") != first || len(out.Chapters) != 1 || out.Chapters[0].Content != "完整首稿。" || out.FullDraft.RevisionStatus == "rejected" {
				t.Fatal("protected failure was swallowed as invalid output fallback", kind, out.Status, err)
			}
			if kind == "save" && len(provider.requests) != 1 {
				t.Fatal("continued after durable save failed")
			}
			if kind == "cancel" && len(provider.requests) != 2 {
				t.Fatal("continued after cancellation")
			}
		})
	}
}

func TestShortFullPipelineAcceptsEmptyNilErrorRevisionOnlyAsRejectedOutput(t *testing.T) {
	first := shortFullRaw("首稿", "完整首稿。")
	p, provider, in, story := shortFullFixture(t, []string{first, "审核意见", "", shortFullPackageReply(), imitationStageReplies()[8]}, 1)
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true)
	if err != nil || out.Status != "completed" || out.FullDraft.RevisionStatus != "rejected" || len(provider.requests) != 5 || out.Chapters[0].Content != "完整首稿。" {
		t.Fatal("empty actual revision output did not retain v001", err, out.Status)
	}
}

func TestShortFullPipelineRetryReopensDurableFirstDraftWithoutRewriting(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	base := imitationStageReplies()
	first := shortFullRaw("首稿", "完整首稿。")
	second := shortFullRaw("二稿", "修订正文。")
	replies := append([]string{}, base[:5]...)
	replies = append(replies, first, "审核意见", "", second, shortFullPackageReply(), base[8])
	provider := &queuedPipelineProvider{responses: replies, errors: make([]error, 8)}
	provider.errors[7] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
	makeService := func() *Service { return NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider)) }
	service := makeService()
	input := shortInput("full-retry")
	input.TargetWordCount = 1000
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	_ = json.Unmarshal(started.Run.PromptSnapshot, &snapshot)
	if snapshot["short_full_draft_version"] != float64(1) {
		t.Fatal("new short imitation did not freeze full draft protocol")
	}
	failed, err := service.Run(ctx, "alice", started.Run.ID)
	if !IsCode(err, ErrModelRateLimited) || failed.Status != RunFailed {
		t.Fatal(failed.Status, err)
	}
	service = makeService()
	saved, err := service.GetShortStory(ctx, "alice", started.Project.ID)
	if err != nil || saved.FullDraft.CurrentVersion != "v001" || saved.FullDraft.RevisionStatus != "pending" || saved.Chapters[0].Content != "完整首稿。" || shortFullArtifact(saved, "drafts/v001.md") != first {
		t.Fatal("durable first draft unavailable after failed revision", err)
	}
	next, err := service.RetryRun(ctx, "alice", failed.ID)
	if err != nil {
		t.Fatal(err)
	}
	service = makeService()
	run, err := service.Run(ctx, "alice", next.ID)
	if err != nil || run.Status != RunSucceeded || len(provider.requests) != 11 {
		t.Fatal("whole draft retry repeated completed calls", run.Status, err, len(provider.requests))
	}
	if len(provider.requests[8].Messages) != 4 || provider.requests[8].Messages[2].Content != strings.TrimSpace(first) {
		t.Fatal("retry revision lost original full draft")
	}
	execution, ok, err := store.GetExecution(ctx, "alice", next.ID)
	if err != nil || !ok {
		t.Fatal(err)
	}
	record := execution.Calls[prompt.ShortFullRevision.ID+":v002"]
	if record.Metadata == nil || len(record.Metadata.Messages) != 4 {
		t.Fatal("actual whole revision history not journaled")
	}
}

func TestShortFullPipelineOwnedMySQLCompletionFallbackAndRetry(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("full-sql")
	base := imitationStageReplies()
	replies := append([]string{}, base[:5]...)
	replies[3] = `{"title":"两章故事","summary":"完整闭环","chapters":[{"number":1,"title":"第一章","summary":"找到信件"},{"number":2,"title":"第二章","summary":"真相揭晓"}]}`
	first := shortFullRaw("首稿", "第一章正文。", "")
	completion := fmtShortFullChapter(2, "补齐第二章。")
	invalid := shortFullRaw("坏二稿", "", "仍然不完整。")
	replies = append(replies, first, completion, "全篇审稿意见", invalid, "", shortFullPackageReply(), base[8])
	provider := &queuedPipelineProvider{responses: replies, errors: make([]error, 10)}
	provider.errors[9] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "rate limited"}
	makeService := func() *Service {
		return NewService(NewMySQLStore(db), NewMemoryRuntime(), pipelineWithProvider(provider))
	}
	service := makeService()
	input := shortInput(novelID("full-sql"))
	input.TargetWordCount = 1000
	input.TargetChapters = 2
	started, err := service.StartShortStory(ctx, owner, input)
	if err != nil {
		t.Fatal(err)
	}
	failed, err := service.Run(ctx, owner, started.Run.ID)
	if !IsCode(err, ErrModelRateLimited) || failed.Status != RunFailed || len(provider.requests) != 10 {
		t.Fatal(failed.Status, err, len(provider.requests))
	}
	service = makeService()
	saved, err := service.GetShortStory(ctx, owner, started.Project.ID)
	if err != nil || saved.FullDraft == nil || saved.FullDraft.CurrentVersion != "v001" || saved.FullDraft.RevisionStatus != "rejected" || saved.FullDraft.CompletionAttempts != 1 || len(saved.Chapters) != 2 || saved.Chapters[1].Content != "补齐第二章。" {
		t.Fatal("SQL lost completed first draft or rejection warning", err)
	}
	next, err := service.RetryRun(ctx, owner, failed.ID)
	if err != nil {
		t.Fatal(err)
	}
	service = makeService()
	last, err := service.Run(ctx, owner, next.ID)
	if err != nil || last.Status != RunSucceeded || len(provider.requests) != 12 {
		t.Fatal("SQL retry regenerated completed draft/completion/revision", last.Status, err, len(provider.requests))
	}
	saved, err = makeService().GetShortStory(ctx, owner, started.Project.ID)
	if err != nil || saved.Status != "completed" || saved.FullDraft.CurrentVersion != "v001" || saved.FullDraft.RevisionWarning == "" || shortFullArtifact(saved, "drafts/v002.md") != "" || shortFullArtifact(saved, "drafts/v002-rejected.md") != invalid {
		t.Fatal("SQL fallback result is not explicit/durable", err)
	}
	execution, ok, err := NewMySQLStore(db).GetExecution(ctx, owner, next.ID)
	if err != nil || !ok {
		t.Fatal(err)
	}
	record := execution.Calls[prompt.ShortFullRevision.ID+":v002"]
	if record.Metadata == nil || len(record.Metadata.Messages) != 4 || !strings.Contains(record.Metadata.Messages[2].Content, "补齐第二章") {
		t.Fatal("SQL revision did not use merged first draft")
	}
	if shortArtifactEditable(saved, "drafts/v001.md") || shortArtifactEditable(saved, "drafts/v002-rejected.md") {
		t.Fatal("version evidence is editable")
	}
	t.Log("real owned MySQL: missing chapter completed once; invalid v002 retained v001; package429 resumed after service reopen; 12 fake provider calls, zero paid requests")
}

func TestShortFullPipelineKeepsFrozenPromptsCustomRequirementsAndModelRoutes(t *testing.T) {
	first := shortFullRaw("首稿", "正文。")
	p, provider, in, story := shortFullFixture(t, []string{first, "审核意见", first, shortFullPackageReply(), imitationStageReplies()[8]}, 1)
	in.SystemPrompts = &SystemPromptSnapshot{Version: 1, Entries: map[string]FrozenSystemPrompt{}}
	for _, definition := range prompt.AdminPrompts() {
		in.SystemPrompts.Entries[definition.ID] = FrozenSystemPrompt{ID: definition.ID, Version: 1, Revision: 7, Body: "冻结-" + definition.ID + "\n" + definition.Definition.SystemPrompt}
	}
	in.Agent.CustomWritePrompts = []AgentPrompt{{ID: "short", Category: "short_fiction", Enabled: true, Content: "短篇写作自定义要求"}, {ID: "imitation", Category: "imitation", Enabled: true, Content: "仿写自定义要求"}, {ID: "long", Category: "long_fiction", Enabled: true, Content: "不应进入的长篇要求"}}
	in.Agent.ReviewPrompts = []AgentPrompt{{ID: "review", Enabled: true, Kind: "custom", Content: "全篇自定义审稿"}}
	ids := []int{1, 37}
	in.Agent.ReviewDimensionIDs = &ids
	in.Agent.ModelRoutes = map[string]string{"chapter-draft": "own-writer", "chapter-audit": "own-reviewer", "chapter-revision": "own-reviser", "memory-index": "own-packager"}
	routes := []string{}
	p.resolveProvider = func(_ context.Context, _ string, model string) (batchfactoryv11.DirectorProvider, error) {
		routes = append(routes, model)
		return provider, nil
	}
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
		t.Fatal(err)
	}
	system := provider.requests[0].SystemPrompt
	if !strings.Contains(system, "冻结-"+prompt.ShortFullWriter.ID) || !strings.Contains(system, "冻结-"+prompt.ShortImitationWritingID) || strings.Count(system, "短篇写作自定义要求") != 1 || strings.Count(system, "仿写自定义要求") != 1 || strings.Contains(system, "不应进入的长篇要求") {
		t.Fatal("full writer lost frozen/custom requirements")
	}
	review := provider.requests[1].SystemPrompt
	if !strings.Contains(review, "冻结-"+prompt.ShortFullReview.ID) || strings.Count(review, "全篇自定义审稿") != 1 || strings.Count(review, "内置审稿维度") != 1 || !strings.Contains(review, "37.") || !strings.Contains(review, "1.") {
		t.Fatal("full review lost frozen dimensions/custom review")
	}
	for i, want := range []string{"own-writer", "own-reviewer", "own-reviser", "own-packager", "own-reviewer"} {
		if routes[i] != want {
			t.Fatal("full pipeline bypassed configured routes", routes)
		}
	}
	if provider.requests[0].Temperature != .58 || provider.requests[2].Temperature != .45 || provider.requests[1].Temperature != .3 || provider.requests[1].MaxTokens != 8192 || provider.requests[3].Temperature != .45 || provider.requests[3].MaxTokens != 4096 {
		t.Fatal("native writer/revision temperature changed")
	}
}

func TestShortFullContinuationBudgetUsesOnlyMissingChapters(t *testing.T) {
	first := shortFullRaw("首稿", "第一章。", "第二章。", "")
	complete := fmtShortFullChapter(3, "第三章。")
	revised := shortFullRaw("二稿", "第一章。", "第二章。", "第三章。")
	p, provider, in, story := shortFullFixture(t, []string{first, complete, "审稿意见", revised, shortFullPackageReply(), imitationStageReplies()[8]}, 3)
	story.Input.TargetWordCount = 15000
	var bp ShortImitationBlueprintOutput
	_ = json.Unmarshal([]byte(imitationBlueprintReply(15000)), &bp)
	story.Imitation.Blueprint = bp.Content
	story.Imitation.Scenes = bp.Scenes
	if _, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, true); err != nil {
		t.Fatal(err)
	}
	if provider.requests[0].MaxTokens != 37096 || provider.requests[1].MaxTokens != 15096 || provider.requests[3].MaxTokens != 37096 || provider.requests[1].Temperature != .68 {
		t.Fatal("continuation requested full-book budget", provider.requests[0].MaxTokens, provider.requests[1].MaxTokens, provider.requests[3].MaxTokens)
	}
}

func TestShortFullPipelineRejectsCrossModeSnapshotBeforeAnyProvider(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: []string{imitationStageReplies()[3], imitationStageReplies()[4]}}
	service := NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("full-cross-mode")
	input.Mode = "short_original"
	input.Requirements = "完整故事"
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	run := store.runs[memoryKey("alice", started.Run.ID)]
	var snapshot map[string]any
	_ = json.Unmarshal(run.PromptSnapshot, &snapshot)
	snapshot["short_full_draft_version"] = 1
	run.PromptSnapshot, _ = json.Marshal(snapshot)
	store.runs[memoryKey("alice", run.ID)] = run
	failed, err := service.Run(ctx, "alice", run.ID)
	if !IsCode(err, ErrPersistenceFailed) || failed.Status != RunFailed || len(provider.requests) != 0 {
		t.Fatal("corrupt cross-mode full protocol dispatched provider calls", failed.Status, err, len(provider.requests))
	}
}

func TestShortFullCompletedEditsRefreshCurrentDownloadWithoutChangingVersionEvidence(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	base := imitationStageReplies()
	first := shortFullRaw("首稿", "首稿正文。")
	second := strings.Replace(shortFullRaw("二稿", "二稿正文。"), "渡口来了一封信。", "二稿开篇钩子。", 1)
	replies := append([]string{}, base[:5]...)
	replies = append(replies, first, "审稿意见", second, shortFullPackageReply(), base[8])
	provider := &queuedPipelineProvider{responses: replies}
	service := NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("full-edit-download")
	input.TargetWordCount = 1000
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Run(ctx, "alice", started.Run.ID); err != nil {
		t.Fatal(err)
	}
	story, err := service.GetShortStory(ctx, "alice", started.Project.ID)
	if err != nil {
		t.Fatal(err)
	}
	edited, err := service.EditShortStoryArtifact(ctx, "alice", story.ProjectID, story.Version, "chapters/001.txt", "人工修改后的正文。")
	if err != nil {
		t.Fatal(err)
	}
	current := shortFullArtifact(edited, "drafts/final.md")
	if !strings.Contains(current, "人工修改后的正文。") || strings.Contains(current, "二稿正文。") || !strings.Contains(current, "二稿开篇钩子。") {
		t.Fatal("current whole download stale or hook lost", current)
	}
	outline := *edited.Outline
	outline.Title = "人工标题"
	raw, _ := json.Marshal(outline)
	edited, err = service.EditShortStoryArtifact(ctx, "alice", story.ProjectID, edited.Version, "outline/story.json", string(raw))
	if err != nil {
		t.Fatal(err)
	}
	current = shortFullArtifact(edited, "drafts/final.md")
	if !strings.HasPrefix(current, "# 人工标题\n") || !strings.Contains(current, "人工修改后的正文。") || !strings.Contains(current, "二稿开篇钩子。") || !edited.QualityStale {
		t.Fatal("manual title/body/hook diverged", current)
	}
	if shortFullArtifact(edited, "drafts/v001.md") != first || shortFullArtifact(edited, "drafts/v002.md") != second {
		t.Fatal("immutable generation version was changed by manual edit")
	}
	project, err := store.GetProject(ctx, "alice", story.ProjectID)
	if err != nil || project.Title != "人工标题" || edited.Title != "人工标题" {
		t.Fatal("current title and project title diverged", err)
	}
}

func TestShortFullReviewFailureReopensFrozenFirstDraft(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	base := imitationStageReplies()
	first := shortFullRaw("首稿", "完整首稿。")
	second := shortFullRaw("二稿", "采用正文。")
	replies := append([]string{}, base[:5]...)
	replies = append(replies, first, "", "审稿意见", second, shortFullPackageReply(), base[8])
	provider := &queuedPipelineProvider{responses: replies, errors: make([]error, 7)}
	provider.errors[6] = &batchfactoryv11.DirectorHTTPError{StatusCode: 429, Body: "review rate limit"}
	presetPath := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(presetPath, []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := prompt.EnsureAdminPresets(presetPath); err != nil {
		t.Fatal(err)
	}
	makeService := func() *Service {
		service := NewService(store, NewMemoryRuntime(), pipelineWithProvider(provider))
		service.ConfigureSystemPrompts(catalog.NewCatalog(presetPath))
		return service
	}
	service := makeService()
	input := shortInput("full-review-reopen")
	input.TargetWordCount = 1000
	started, err := service.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Prompts *SystemPromptSnapshot `json:"system_prompts"`
	}
	if err = json.Unmarshal(started.Run.PromptSnapshot, &snapshot); err != nil || snapshot.Prompts == nil || len(snapshot.Prompts.Entries) != 41 {
		t.Fatal("new task did not persist all 41 published presets", err)
	}
	failed, err := service.Run(ctx, "alice", started.Run.ID)
	if !IsCode(err, ErrModelRateLimited) || failed.Status != RunFailed || len(provider.requests) != 7 {
		t.Fatal(failed.Status, err, len(provider.requests))
	}
	service = makeService()
	saved, err := service.GetShortStory(ctx, "alice", started.Project.ID)
	if err != nil || saved.Status != "failed" || saved.FullDraft.CurrentVersion != "v001" || saved.FullDraft.RevisionStatus != "pending" || saved.FullDraft.Review != "" || shortFullArtifact(saved, "drafts/v001.md") != first {
		t.Fatal("GET after review failure lost first draft/pending state", err)
	}
	next, err := service.RetryRun(ctx, "alice", failed.ID)
	if err != nil {
		t.Fatal(err)
	}
	service = makeService()
	if run, err := service.Run(ctx, "alice", next.ID); err != nil || run.Status != RunSucceeded || len(provider.requests) != 11 {
		t.Fatal("review retry repeated completed first draft", run.Status, err, len(provider.requests))
	}
	saved, err = service.GetShortStory(ctx, "alice", started.Project.ID)
	if err != nil || shortFullArtifact(saved, "drafts/v001.md") != first || !strings.Contains(shortFullArtifact(saved, "drafts/final.md"), "渡口来了一封信。") || !strings.Contains(shortFullArtifact(saved, "drafts/final.md"), "采用正文。") || shortFullArtifact(saved, "package/report.md") != shortFullPackageReply() {
		t.Fatal("stage artifacts overwritten or final hook missing", err)
	}
}
