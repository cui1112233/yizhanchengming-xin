package novelcreation

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

func TestNativeSettlementBoundaries(t *testing.T) {
	var delta truthSettlementOutput
	_ = json.Unmarshal([]byte(`{"facts":[],"characterChanges":[],"timeline":[],"foreshadowing":[{"thread":"已解开的暗号","status":"open"}]}`), &delta)
	merged, raw, e := mergeNativeSettlement(json.RawMessage(previousSettlementFixture), delta, json.RawMessage(`{}`), 2)
	if e != nil || len(merged.Facts) != 1 || len(merged.Foreshadowing) != 2 || merged.Foreshadowing[1].Status != "resolved" {
		t.Fatalf("empty delta lost state or reopened resolved hook: %+v %v", merged, e)
	}
	again, _, e := mergeNativeSettlement(raw, delta, json.RawMessage(`{}`), 2)
	if e != nil || len(again.Timeline) != len(merged.Timeline) || len(again.Foreshadowing) != 2 {
		t.Fatal("repeat duplicated state", e)
	}
	delta.Facts = []settledFact{{ID: "old-fact", Statement: "钥匙已经交还"}}
	updated, _, e := mergeNativeSettlement(raw, delta, json.RawMessage(`{}`), 3)
	if e != nil || len(updated.Facts) != 1 || updated.Facts[0].Statement != "钥匙已经交还" {
		t.Fatal("targeted update failed", e)
	}
	delta.Facts = append(delta.Facts, delta.Facts[0])
	if _, _, e = mergeNativeSettlement(raw, delta, json.RawMessage(`{}`), 3); e == nil {
		t.Fatal("conflicting IDs accepted")
	}
	for _, bad := range []string{`not json`, `null`, `{"unknown_format":true}`, `{"facts":["opaque string"]}`} {
		if _, _, e = mergeNativeSettlement(json.RawMessage(bad), truthSettlementOutput{}, json.RawMessage(`{}`), 1); e == nil {
			t.Fatal("invalid prior truth accepted", bad)
		}
	}
}

func TestNativeVersion3KeepsFormerSettlementProtocol(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ReviewCycleVersion = 3
	input.PreviousTruth = json.RawMessage(previousSettlementFixture)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(bundle.RuntimeSnapshot), "settlement_delta") || strings.Contains(provider.requests[4].SystemPrompt, "CM数据适配约定") {
		t.Fatal("old frozen settlement protocol migrated")
	}
}

func TestNativeSettlementContinuityRealMySQL(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("state-carry")
	secondPlan := strings.Replace(chapterPlanJSON, `"chapterNumber":1`, `"chapterNumber":2`, 1)
	delta := `{"facts":[{"id":"fact-2","statement":"记者第二天找到了证人"}],"characterChanges":[],"timeline":[],"foreshadowing":[]}`
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON, secondPlan, draftJSON, passedAuditJSON, delta, memoryJSON}}
	store := NewMySQLStore(db)
	s := NewService(store, NewMemoryRuntime(), NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) { return provider, nil }))
	project, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "未提及旧资料跨章保留", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: "long_original", Title: "跨章记忆"})
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.UpdateProjectSettings(ctx, owner, project.ID, ProjectSettings{Agent: AgentSettings{Version: 1, PauseAfterBlueprint: boolPointer(false), PassScore: 80, MaxRevisionRounds: 0}})
	if e != nil {
		t.Fatal(e)
	}
	for _, key := range []string{"first", "second"} {
		r, err := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "继续写下一章", ModelID: "fake-only", IdempotencyKey: key})
		if err != nil {
			t.Fatal(err)
		}
		r, err = s.Run(ctx, owner, r.ID)
		if err != nil || r.Status != RunSucceeded {
			t.Fatalf("%s: %s %v", key, r.Status, err)
		}
	}
	chapters, e := store.ListChapters(ctx, owner, project.ID)
	if e != nil || len(chapters) != 2 {
		t.Fatal("chapters missing", e, len(chapters))
	}
	third, e := s.SendMessage(ctx, owner, SendMessageInput{ProjectID: project.ID, SessionID: session.ID, Content: "继续第三章", ModelID: "fake-only", IdempotencyKey: "third-snapshot-only"})
	if e != nil {
		t.Fatal(e)
	}
	var snapshot struct {
		Truth truthSettlementOutput `json:"previous_truth"`
	}
	if e = json.Unmarshal(third.PromptSnapshot, &snapshot); e != nil || len(snapshot.Truth.Facts) != 2 || len(snapshot.Truth.Foreshadowing) != 1 || snapshot.Truth.Foreshadowing[0].Thread != "监视盲区" {
		t.Fatal("durable predecessor forgot unrelated fact/hook", e, string(third.PromptSnapshot))
	}
	if len(provider.requests) != 11 || !strings.Contains(provider.requests[10].UserPrompt, "监视盲区") {
		t.Fatal("merged truth not supplied to memory index", len(provider.requests))
	}
	t.Log("2 chapters saved; chapter2 delta omitted prior hook/fact; chapter3 frozen input reread both facts and old hook from MySQL; 11 fake calls, 0 paid calls")
}

const previousSettlementFixture = `{"facts":[{"id":"old-fact","statement":"钥匙留在门外"}],"characterChanges":[{"character":"林澜","change":"记住门牌号"}],"timeline":[{"event":"收到钥匙","position":"前夜"}],"foreshadowing":[{"thread":"未找到的钥匙主人","status":"open"},{"thread":"已解开的暗号","status":"resolved"}]}`

func TestNativeSettlementCarriesUnmentionedTruth(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ReviewCycleVersion = 4
	input.PreviousTruth = json.RawMessage(previousSettlementFixture)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if e != nil {
		t.Fatal(e)
	}
	var snapshot struct {
		Truth truthSettlementOutput `json:"truthSettlement"`
	}
	if e = json.Unmarshal(bundle.RuntimeSnapshot, &snapshot); e != nil {
		t.Fatal(e)
	}
	if len(snapshot.Truth.Facts) != 2 || len(snapshot.Truth.Foreshadowing) != 3 || !strings.Contains(provider.requests[5].UserPrompt, "未找到的钥匙主人") {
		t.Fatalf("old truth disappeared: %s", bundle.RuntimeSnapshot)
	}
	if !strings.Contains(string(bundle.RuntimeSnapshot), "settlement_delta") {
		t.Fatal("original delta evidence missing")
	}
}

func TestNativeSettlementCancellationStopsBeforeSave(t *testing.T) {
	runtime := NewMemoryRuntime()
	input := fixturePipelineRunContext(runtime)
	input.ReviewCycleVersion = 4
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	provider.afterCall = func(n int) {
		if n == 4 {
			_ = runtime.Cancel(context.Background(), input.Run.ID)
		}
	}
	bundle, e := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if !IsCode(e, ErrRunCancelled) || bundle.Chapter.Content != "" || len(provider.requests) != 5 {
		t.Fatalf("cancelled settlement continued: %v calls=%d", e, len(provider.requests))
	}
}
