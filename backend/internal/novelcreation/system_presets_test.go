package novelcreation

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	catalog "qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"testing"
)

func TestSystemPresetPublicationAffectsNewTasksOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	_ = os.WriteFile(path, []byte(`[]`), 0600)
	if e := novelprompt.EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	_, e := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PauseAfterBlueprint: boolPointer(false), PassScore: 80, MaxRevisionRounds: 0}})
	if e != nil {
		t.Fatal(e)
	}
	f.service.ConfigureSystemPrompts(catalog.NewCatalog(path))
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON, strings.Replace(chapterPlanJSON, `"chapterNumber":1`, `"chapterNumber":2`, 1), draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = pipelineWithProvider(provider)
	first, e := f.service.SendMessage(ctx, "alice", f.sendInput("preset-first"))
	if e != nil {
		t.Fatal(e)
	}
	admin := catalog.NewAdminStore(path)
	draft, e := admin.Get(novelprompt.ChapterDraft.ID, 1)
	if e != nil {
		t.Fatal(e)
	}
	draft["body"] = "ADMIN_PUBLISHED_BODY_V2"
	created, e := admin.CreateDraft("admin", draft)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = admin.Publish("admin", novelprompt.ChapterDraft.ID, created["version"].(int)); e != nil {
		t.Fatal(e)
	}
	first, e = f.service.Run(ctx, "alice", first.ID)
	if e != nil || first.Status != RunSucceeded {
		t.Fatal(first.Status, e)
	}
	if strings.Contains(provider.requests[2].SystemPrompt, "ADMIN_PUBLISHED_BODY_V2") || !strings.Contains(provider.requests[2].SystemPrompt, novelprompt.ChapterDraft.SystemPrompt) {
		t.Fatal("queued task changed after publication")
	}
	second, e := f.service.SendMessage(ctx, "alice", f.sendInput("preset-second"))
	if e != nil {
		t.Fatal(e)
	}
	second, e = f.service.Run(ctx, "alice", second.ID)
	if e != nil || second.Status != RunSucceeded {
		t.Fatal(second.Status, e)
	}
	if len(provider.requests) != 11 || !strings.Contains(provider.requests[7].SystemPrompt, "ADMIN_PUBLISHED_BODY_V2") {
		t.Fatal("new generation ignored published body", len(provider.requests))
	}
}

func TestSystemPresetDraftArchiveRollbackAndLegacy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if e := os.WriteFile(path, []byte(`[]`), 0600); e != nil {
		t.Fatal(e)
	}
	if e := novelprompt.EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	s := newServiceFixture(t, nil).service
	s.ConfigureSystemPrompts(catalog.NewCatalog(path))
	admin := catalog.NewAdminStore(path)
	draft, e := admin.Get(novelprompt.ChapterDraft.ID, 1)
	if e != nil {
		t.Fatal(e)
	}
	draft["body"] = "UNPUBLISHED_BODY"
	v2, e := admin.CreateDraft("admin", draft)
	if e != nil {
		t.Fatal(e)
	}
	read := func() *SystemPromptSnapshot {
		t.Helper()
		snap, e := s.snapshotSystemPrompts()
		if e != nil {
			t.Fatal(e)
		}
		return snap
	}
	before := read()
	if before.Entries[novelprompt.ChapterDraft.ID].Body != novelprompt.ChapterDraft.SystemPrompt {
		t.Fatal("unpublished draft used")
	}
	if _, e = admin.Publish("admin", novelprompt.ChapterDraft.ID, v2["version"].(int)); e != nil {
		t.Fatal(e)
	}
	if read().Entries[novelprompt.ChapterDraft.ID].Body != "UNPUBLISHED_BODY" {
		t.Fatal("publication ignored")
	}
	if _, e = admin.Rollback("admin", novelprompt.ChapterDraft.ID, 1); e != nil {
		t.Fatal(e)
	}
	if read().Entries[novelprompt.ChapterDraft.ID].Body != novelprompt.ChapterDraft.SystemPrompt {
		t.Fatal("rollback ignored")
	}
	if _, e = admin.Archive("admin", novelprompt.ChapterDraft.ID, 1); e != nil {
		t.Fatal(e)
	}
	provider := &queuedPipelineProvider{responses: []string{draftJSON}}
	var output chapterDraftOutput
	_, e = pipelineWithProvider(provider).callJSON(context.Background(), RunContext{SystemPrompts: read()}, &recordingEventSink{}, provider, modelCallSpec{definition: novelprompt.ChapterDraft, userPayload: map[string]any{}, maxTokens: 100}, &output)
	if e == nil || !strings.Contains(e.Error(), "未发布") || len(provider.requests) != 0 {
		t.Fatal("archived prompt did not stop before model", e)
	}
	if body, e := frozenSystemText(RunContext{}, novelprompt.ChapterDraft.ID, "HISTORICAL_BODY"); e != nil || body != "HISTORICAL_BODY" {
		t.Fatal("historical snapshot changed", e)
	}
	if before.Entries[novelprompt.ChapterDraft.ID].Body != novelprompt.ChapterDraft.SystemPrompt {
		t.Fatal("captured body mutated")
	}
}

func TestSystemPresetNativeRevisionAndIncrementalSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "presets.json")
	if e := os.WriteFile(path, []byte(`[]`), 0600); e != nil {
		t.Fatal(e)
	}
	if e := novelprompt.EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	s := newServiceFixture(t, nil).service
	s.ConfigureSystemPrompts(catalog.NewCatalog(path))
	snap, e := s.snapshotSystemPrompts()
	if e != nil {
		t.Fatal(e)
	}
	for id, item := range snap.Entries {
		item.Body = "PUBLISHED_" + id
		snap.Entries[id] = item
	}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ReviewCycleVersion = 3
	input.SystemPrompts = snap
	localAudit := strings.Replace(failedAuditJSON, `"category":"plot"`, `"category":"plot","repairScope":"local"`, 1)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, localAudit, nativeFixturePatch, passedAuditJSON, settlementJSON, memoryJSON}}
	if _, e = pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(provider.requests[4].SystemPrompt, "PUBLISHED_"+novelprompt.NativeRevisionPresetID) || !strings.Contains(provider.requests[4].SystemPrompt, "PUBLISHED_"+novelprompt.PatchOnlyPresetID) {
		t.Fatal("revision bypassed published base/route")
	}
	for _, p := range novelprompt.AdminPrompts() {
		got, e := frozenSystemText(input, p.ID, "fallback")
		if e != nil || got != "PUBLISHED_"+p.ID {
			t.Fatal(p.ID, e)
		}
	}
	if systemPresetKey(novelprompt.IncrementalTruthSettlement) != novelprompt.IncrementalSettlementPresetID {
		t.Fatal("incremental settlement used rebuild prompt")
	}
}

func TestSystemPresetRealMySQLSnapshotSurvivesRestart(t *testing.T) {
	db := studioOwnedDB(t)
	store := NewMySQLStore(db)
	ctx := context.Background()
	owner := novelID("preset-mysql")
	runner := &stubPipelineRunner{}
	s := NewService(store, NewMemoryRuntime(), runner)
	path := filepath.Join(t.TempDir(), "presets.json")
	if e := os.WriteFile(path, []byte(`[]`), 0600); e != nil {
		t.Fatal(e)
	}
	if e := novelprompt.EnsureAdminPresets(path); e != nil {
		t.Fatal(e)
	}
	s.ConfigureSystemPrompts(catalog.NewCatalog(path))
	p, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "预设快照验收", Genre: "悬疑", World: "旧城", Protagonist: "记者", CoreConflict: "寻找真相", ModelID: "test-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: p.ID, Mode: "long_form_original", Title: "第一章"})
	if e != nil {
		t.Fatal(e)
	}
	job, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: p.ID, SessionID: session.ID, Content: "从匿名信开始", ModelID: "test-only", IdempotencyKey: "freeze"})
	if e != nil {
		t.Fatal(e)
	}
	var frozen struct {
		SystemPrompts *SystemPromptSnapshot `json:"system_prompts"`
	}
	if e = json.Unmarshal(job.PromptSnapshot, &frozen); e != nil || frozen.SystemPrompts == nil {
		t.Fatal("snapshot absent", e)
	}
	admin := catalog.NewAdminStore(path)
	draft, e := admin.Get(novelprompt.ChapterDraft.ID, 1)
	if e != nil {
		t.Fatal(e)
	}
	draft["body"] = "AFTER_RESTART_V2"
	v2, e := admin.CreateDraft("admin", draft)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = admin.Publish("admin", novelprompt.ChapterDraft.ID, v2["version"].(int)); e != nil {
		t.Fatal(e)
	}
	reopened := NewService(NewMySQLStore(db), NewMemoryRuntime(), runner)
	reopened.ConfigureSystemPrompts(catalog.NewCatalog(path))
	got, e := reopened.Run(ctx, owner, job.ID)
	if e != nil || got.Status != RunSucceeded {
		t.Fatal(got.Status, e)
	}
	actual := runner.lastCall().SystemPrompts
	if actual == nil || actual.Entries[novelprompt.ChapterDraft.ID].Version != 1 || actual.Entries[novelprompt.ChapterDraft.ID].Body != novelprompt.ChapterDraft.SystemPrompt {
		t.Fatal("restart replaced frozen prompt")
	}
}
