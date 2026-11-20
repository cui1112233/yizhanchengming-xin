package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestShortV2PausesBeforeGenericOutlineAndResumesWithoutSecondPause(t *testing.T) {
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	pause := true
	a.PauseAfterBlueprint = &pause
	p, provider, in, story := controlledShortFixture(t, imitationStageReplies(), a)
	in.ShortImitationVersion = 2
	story.Outline = nil
	story.Input.Mode = "short_imitation"
	story.Input.ReferenceText = strings.Repeat("参考", 100)
	out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
	if err != nil || out.Status != "awaiting_confirmation" || out.Outline != nil || len(provider.requests) != 3 {
		t.Fatalf("blueprint checkpoint must precede generic outline: status=%s outline=%v calls=%d err=%v", out.Status, out.Outline, len(provider.requests), err)
	}
	raw, _ := json.Marshal(out)
	if !strings.Contains(string(raw), `"checkpoint":"blueprint"`) {
		t.Fatal("missing explicit blueprint checkpoint")
	}
	out, err = p.buildShortStory(context.Background(), in, &recordingEventSink{}, out, true)
	if err != nil || out.Status != "completed" || len(provider.requests) != 9 {
		t.Fatalf("resume failed or paused twice: %s %d %v", out.Status, len(provider.requests), err)
	}
}

func TestShortV2NonpassEditorialIsCompletedButV1RemainsFailed(t *testing.T) {
	for _, version := range []int{1, 2} {
		t.Run(string(rune('0'+version)), func(t *testing.T) {
			replies := imitationStageReplies()
			replies[8] = strings.ReplaceAll(replies[8], "可过审", "需修改")
			replies[8] = strings.Replace(replies[8], `"issues":[]`, `"issues":["结尾逻辑需要修改"]`, 1)
			p, _, in, story := controlledShortFixture(t, replies, NormalizeAgentSettings(AgentSettings{}, 2000))
			in.ShortImitationVersion = version
			story.Input.Mode = "short_imitation"
			story.Input.ReferenceText = strings.Repeat("参考", 100)
			out, err := p.buildShortStory(context.Background(), in, &recordingEventSink{}, story, false)
			if version == 1 {
				if err == nil || out.Status == "completed" {
					t.Fatal("v1 changed", err)
				}
				return
			}
			if err != nil || out.Status != "completed" || out.Imitation.Editorial.Rating != "需修改" || len(out.Chapters) != 1 {
				t.Fatal("valid nonpass review discarded completed manuscript", err, out.Status)
			}
		})
	}
}

func TestShortBlueprintEditPersistsAndEntersConfirmedGeneration(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: imitationStageReplies()}
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	input := shortInput("blueprint-edit")
	input.TargetWordCount = 1000
	input.PauseAfterOutline = true
	started, err := s.StartShortStory(ctx, "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.EditShortStoryArtifact(ctx, "alice", started.Project.ID, started.Story.Version, "imitation/blueprint.md", "premature"); err == nil {
		t.Fatal("queued edit accepted")
	}
	run, err := s.Run(ctx, "alice", started.Run.ID)
	if err != nil || run.Status != RunAwaitingConfirmation {
		t.Fatal(run.Status, err)
	}
	s = NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	story, err := s.GetShortStory(ctx, "alice", started.Project.ID)
	if err != nil || story.Checkpoint != "blueprint" || story.Outline != nil {
		t.Fatal("checkpoint lost", err)
	}
	for _, artifact := range story.Artifacts {
		if artifact.Editable != (artifact.Key == "imitation/blueprint.md") {
			t.Fatal("incorrect waiting editor permission", artifact.Key)
		}
	}
	editedText := story.Imitation.Blueprint + "\n人工补充：灯塔藏着最后一封信。"
	for _, test := range []struct {
		owner, key, content string
		version             int
	}{
		{"bob", "imitation/blueprint.md", editedText, story.Version},
		{"alice", "../../secrets", editedText, story.Version},
		{"alice", "imitation/blueprint.md", editedText, story.Version - 1},
		{"alice", "imitation/blueprint.md", strings.ReplaceAll(editedText, "渡口查信", "更换事件"), story.Version},
		{"alice", "imitation/blueprint.md", strings.ReplaceAll(editedText, "200字", "999字"), story.Version},
	} {
		if _, err = s.EditShortStoryArtifact(ctx, test.owner, story.ProjectID, test.version, test.key, test.content); err == nil {
			t.Fatal("invalid artifact edit accepted", test)
		}
	}
	edited, err := s.EditShortStoryArtifact(ctx, "alice", story.ProjectID, story.Version, "imitation/blueprint.md", editedText)
	if err != nil {
		t.Fatal(err)
	}
	reopened := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	saved, err := reopened.GetShortStory(ctx, "alice", story.ProjectID)
	if err != nil || saved.Imitation.Blueprint != editedText {
		t.Fatal("edit did not persist", err)
	}
	next, err := reopened.ConfirmShortStory(ctx, "alice", run.ID, edited.Version)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := reopened.ConfirmShortStory(ctx, "alice", run.ID, edited.Version)
	if err != nil || replay.ID != next.ID {
		t.Fatal("duplicate confirm not idempotent", err)
	}
	if _, err = reopened.ConfirmShortStory(ctx, "alice", run.ID, edited.Version-1); !IsCode(err, ErrIdempotencyConflict) {
		t.Fatal("stale confirmation accepted", err)
	}
	if _, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, saved.Version, "imitation/blueprint.md", editedText); err == nil {
		t.Fatal("active queued edit accepted")
	}
	last, err := reopened.Run(ctx, "alice", next.ID)
	if err != nil || last.Status != RunSucceeded || len(provider.requests) != 9 {
		t.Fatal(last.Status, err, len(provider.requests))
	}
	for _, index := range []int{3, 5, 8} {
		raw, _ := json.Marshal(provider.requests[index])
		if !strings.Contains(string(raw), "灯塔藏着最后一封信") {
			t.Fatal("edited blueprint not consumed", index)
		}
	}
	finished, err := reopened.GetShortStory(ctx, "alice", story.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	edited, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, finished.Version, "chapters/001.txt", "人工修改正文。")
	if err != nil || edited.Chapters[0].Content != "人工修改正文。" || !edited.QualityStale {
		t.Fatal("manuscript edit not synchronized or rating misleading", err)
	}
	if _, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, edited.Version, "chapters/999.txt", "unknown"); err == nil {
		t.Fatal("unknown file accepted")
	}
	if _, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, edited.Version, "reviews/editorial.md", "伪造评价"); err == nil {
		t.Fatal("report evidence can be overwritten")
	}
	outline := *edited.Outline
	outline.Title = "人工新标题"
	outlineRaw, _ := json.Marshal(outline)
	if _, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, edited.Version, "outline/story.json", string(outlineRaw)+" {}"); err == nil {
		t.Fatal("extra JSON accepted")
	}
	edited, err = reopened.EditShortStoryArtifact(ctx, "alice", story.ProjectID, edited.Version, "outline/story.json", string(outlineRaw))
	if err != nil {
		t.Fatal(err)
	}
	project, err := st.GetProject(ctx, "alice", story.ProjectID)
	if err != nil || project.Title != "人工新标题" || edited.Title != "人工新标题" {
		t.Fatal("story and project titles diverge", err)
	}
}

func TestShortV2ConfirmChangesOnlyContinuationModel(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	provider := &queuedPipelineProvider{responses: imitationStageReplies()}
	models := []string{}
	service := NewService(store, NewMemoryRuntime(), NewPipeline(func(_ context.Context, _ string, model string) (batchfactoryv11.DirectorProvider, error) {
		models = append(models, model)
		return provider, nil
	}))
	input := shortInput("confirm-model")
	input.TargetWordCount = 1000
	input.PauseAfterOutline = true
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
	if _, err = service.ConfirmShortStoryWithModel(ctx, "alice", started.Run.ID, story.Version, "next-model", "invalid"); err == nil {
		t.Fatal("invalid thinking accepted")
	}
	next, err := service.ConfirmShortStoryWithModel(ctx, "alice", started.Run.ID, story.Version, "next-model", "high")
	if err != nil || next.ModelID != "next-model" || next.ThinkingLevel != "deep" {
		t.Fatal("continuation did not use selected model", next, err)
	}
	old, err := store.GetRun(ctx, "alice", started.Run.ID)
	if err != nil || old.ModelID != "fake-only" || old.ThinkingLevel != "standard" {
		t.Fatal("source model changed", old, err)
	}
	same, err := service.ConfirmShortStoryWithModel(ctx, "alice", started.Run.ID, story.Version, "next-model", "high")
	if err != nil || same.ID != next.ID {
		t.Fatal("model confirm duplicate changed result", err)
	}
	if _, err = service.ConfirmShortStoryWithModel(ctx, "alice", started.Run.ID, story.Version, "different", "high"); !IsCode(err, ErrIdempotencyConflict) {
		t.Fatal("different duplicate accepted", err)
	}
	if _, err = service.Run(ctx, "alice", next.ID); err != nil {
		t.Fatal("new model continuation failed", err)
	}
	for index, model := range models {
		want := "fake-only"
		if index >= 3 {
			want = "next-model"
		}
		if model != want {
			t.Fatal("selected continuation model did not reach provider", index, model)
		}
	}
}
