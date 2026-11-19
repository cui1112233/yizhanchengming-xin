package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"qiantie/backend/internal/batchfactoryv11"
	"reflect"
	"strings"
	"testing"
)

func TestAgentSettingsPresenceAndApprovedDefaults(t *testing.T) {
	for _, tc := range []struct {
		raw    string
		rounds int
	}{{`{}`, 3}, {`{"maxRevisionRounds":0}`, 0}, {`{"maxRevisionRounds":2}`, 2}} {
		var a AgentSettings
		if err := json.Unmarshal([]byte(tc.raw), &a); err != nil {
			t.Fatal(err)
		}
		a = NormalizeAgentSettings(a, 2000)
		raw, _ := json.Marshal(a)
		var fields map[string]any
		_ = json.Unmarshal(raw, &fields)
		if a.MaxRevisionRounds != tc.rounds || fields["passScore"] != float64(80) || fields["baseTemperature"] != .65 || fields["temperatureStep"] != .15 || fields["enableLengthNorm"] != true || fields["enablePostWrite"] != true || fields["enableAiTells"] != true {
			t.Fatalf("input=%s normalized=%s", tc.raw, raw)
		}
	}
}

func TestAgentPipelineRoutesZeroRevisionAndTemperature(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	if err := json.Unmarshal([]byte(`{"Agent":{"maxRevisionRounds":0,"passScore":80,"baseTemperature":0,"temperatureStep":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false,"modelRoutes":{"foundation-plan":"planner","chapter-plan":"planner","chapter-draft":"writer","chapter-audit":"critic","truth-settlement":"truth","memory-index":"memory"}}}`), &input); err != nil {
		t.Fatal(err)
	}
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, settlementJSON, memoryJSON}}
	var models []string
	pipeline := NewPipeline(func(_ context.Context, _, m string) (batchfactoryv11.DirectorProvider, error) {
		models = append(models, m)
		return provider, nil
	})
	bundle, err := pipeline.Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Revisions) != 0 || bundle.FinalRunStatus != RunSucceededWithWarnings {
		t.Fatalf("explicit zero revision ignored: %+v", bundle)
	}
	if !reflect.DeepEqual(models, []string{"planner", "planner", "writer", "critic", "truth", "memory"}) {
		t.Fatalf("actual stage models=%v", models)
	}
	for _, r := range provider.requests {
		if r.Temperature != 0 {
			t.Fatalf("explicit zero temperature lost: %v", r.Temperature)
		}
	}
}

func TestAgentOptionalStepsUseFinalContentAndEachToggle(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		optional  int
	}{{"all", `{}`, 1}, {"length off", `{"enableLengthNorm":false}`, 0}, {"post off", `{"enablePostWrite":false}`, 1}, {"expression off", `{"enableAiTells":false}`, 1}, {"all off", `{"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			input := fixturePipelineRunContext(NewMemoryRuntime())
			var a AgentSettings
			_ = json.Unmarshal([]byte(tc.raw), &a)
			raw, _ := json.Marshal(map[string]any{"Agent": a})
			_ = json.Unmarshal(raw, &input)
			responses := []string{foundationJSON, chapterPlanJSON, draftJSON}
			for i := 0; i < tc.optional; i++ {
				responses = append(responses, `{"content":"整理后的完整正文。"}`)
			}
			responses = append(responses, passedAuditJSON, settlementJSON, memoryJSON)
			provider := &queuedPipelineProvider{responses: responses}
			bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
			if err != nil {
				t.Fatal(err)
			}
			if len(provider.requests) != 6+tc.optional {
				t.Fatalf("actual optional calls=%d", len(provider.requests)-6)
			}
			if tc.optional > 0 {
				if bundle.Chapter.Content != "整理后的完整正文。" {
					t.Fatalf("lost final optional output: %s", bundle.Chapter.Content)
				}
				for _, r := range provider.requests[len(provider.requests)-3:] {
					if !strings.Contains(r.UserPrompt, "整理后的完整正文") {
						t.Fatalf("audit/settlement/memory did not consume final content: %s", r.UserPrompt)
					}
				}
			}
		})
	}
}

func TestAgentRevisionTemperatureBoundAndPassScore(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	_ = json.Unmarshal([]byte(`{"Agent":{"maxRevisionRounds":1,"passScore":95,"baseTemperature":1.9,"temperatureStep":0.5,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}}`), &input)
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, revisedJSON, `{"score":92,"issues":[],"passed":false}`, settlementJSON, memoryJSON}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunSucceededWithWarnings || provider.requests[4].Temperature != 2 {
		t.Fatalf("status=%s revision temp=%v", bundle.FinalRunStatus, provider.requests[4].Temperature)
	}
	if !strings.Contains(provider.requests[3].UserPrompt, `"pass_score":95`) {
		t.Fatal("effective pass score missing from real audit request")
	}
}

func TestAgentSecondaryConfirmationSameRunAndStaleToken(t *testing.T) {
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	if _, err := f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a}); err != nil {
		t.Fatal(err)
	}
	p := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = pipelineWithProvider(p)
	run := queuedWorkerRun(t, f)
	paused, err := f.service.Run(ctx, "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != RunAwaitingConfirmation || len(p.requests) != 1 {
		t.Fatalf("gate did not pause BEFORE second external call: status=%s calls=%d", paused.Status, len(p.requests))
	}
	var state struct {
		ConfirmationID string `json:"confirmationId"`
		Kind           string `json:"confirmationKind"`
	}
	_ = json.Unmarshal(paused.RuntimeSnapshot, &state)
	if state.ConfirmationID == "" || state.Kind != "next_model_call" {
		t.Fatalf("missing typed confirmation: %s", paused.RuntimeSnapshot)
	}
	if _, err = f.service.ConfirmRun(ctx, "alice", run.ID, nil); err == nil {
		t.Fatal("secondary confirmation accepted no expected token")
	}
	// Exercise the public HTTP-independent contract without compile-time coupling.
	confirmed, err := f.service.ConfirmModelCall(ctx, "alice", run.ID, state.ConfirmationID)
	if err != nil || confirmed.ID != run.ID {
		t.Fatalf("same run continuation=%s %v", confirmed.ID, err)
	}
	second, err := f.service.Run(ctx, "alice", run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if second.Status != RunAwaitingConfirmation || len(p.requests) != 2 {
		t.Fatalf("history replay resent or skipped gate: %s calls=%d", second.Status, len(p.requests))
	}
	var next struct {
		ConfirmationID string `json:"confirmationId"`
	}
	_ = json.Unmarshal(second.RuntimeSnapshot, &next)
	if next.ConfirmationID == state.ConfirmationID {
		t.Fatal("confirmation token reused for another call")
	}
	if _, err = f.service.ConfirmModelCall(ctx, "alice", run.ID, state.ConfirmationID); err != nil {
		t.Fatal(err)
	}
	again, _ := f.store.GetRun(ctx, "alice", run.ID)
	if again.Status != RunAwaitingConfirmation {
		t.Fatal("late old confirmation approved a newer paid request")
	}
	// A restart continues with the same journal and each specific token.
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(p))
	for i := 0; i < 4; i++ {
		current, _ := f.store.GetRun(ctx, "alice", run.ID)
		_ = json.Unmarshal(current.RuntimeSnapshot, &next)
		if _, err = f.service.ConfirmModelCall(ctx, "alice", run.ID, next.ConfirmationID); err != nil {
			t.Fatal(err)
		}
		current, err = f.service.Run(ctx, "alice", run.ID)
		if err != nil {
			t.Fatal(err)
		}
	}
	done, _ := f.store.GetRun(ctx, "alice", run.ID)
	if done.Status != RunSucceeded || len(p.requests) != 6 {
		t.Fatalf("same run finish=%s calls=%d", done.Status, len(p.requests))
	}
}

// This is the exact old settings/run envelope, before Agent contract v1.
func legacyAgentRun(t *testing.T, f serviceFixture, r Run) Run {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(r.PromptSnapshot, &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "agent_contract_version")
	delete(fields, "review_cycle_version") // This helper represents a pre-versioned frozen task.
	settings := fields["project_settings"].(map[string]any)
	settings["agent"] = map[string]any{"targetChapterWords": 2000, "maxRevisionRounds": 2, "thinkingLevel": "standard"}
	r.PromptSnapshot, _ = json.Marshal(fields)
	f.store.mu.Lock()
	f.store.runs[memoryKey("alice", r.ID)] = r
	f.store.mu.Unlock()
	return r
}

type metadataQueuedProvider struct {
	*queuedPipelineProvider
	metadata batchfactoryv11.TextModelMetadata
}

func (p *metadataQueuedProvider) TextModelMetadata() batchfactoryv11.TextModelMetadata {
	return p.metadata
}

func TestAgentJournalReplaysOldMetadataAndRecordsActualParameters(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	_, _ = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a})
	unsupported := false
	p := &metadataQueuedProvider{queuedPipelineProvider: &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}, metadata: batchfactoryv11.TextModelMetadata{ActualModel: "actual-model", ChannelID: "channel-safe", ConfigRevision: "v1", SupportsTemperature: &unsupported, SupportsMaxTokens: &unsupported}}
	f.service.pipeline = pipelineWithProvider(p)
	f.service.store = faultJournalStore{f.store, &onceWriteFault{kind: "event"}}
	r := queuedWorkerRun(t, f)
	if _, err := f.service.Run(ctx, "alice", r.ID); err == nil {
		t.Fatal("expected local event failure")
	}
	p.metadata.ConfigRevision = "v2"
	f.service = NewService(f.store, NewMemoryRuntime(), pipelineWithProvider(p))
	done, err := f.service.Run(ctx, "alice", r.ID)
	if err != nil || done.Status != RunSucceeded || len(p.requests) != 6 {
		t.Fatalf("completed paid response was not replayed: status=%s calls=%d err=%v", done.Status, len(p.requests), err)
	}
	var runtime struct {
		Calls map[string]ModelCallMetadata `json:"modelCalls"`
		Agent AgentSettings                `json:"agentSettings"`
	}
	if err = json.Unmarshal(done.RuntimeSnapshot, &runtime); err != nil {
		t.Fatal(err)
	}
	if len(runtime.Calls) != 6 || runtime.Calls["novel-creation.foundation-plan"].Provider.ConfigRevision != "v1" || runtime.Calls["novel-creation.chapter-plan"].Provider.ConfigRevision != "v2" {
		t.Fatalf("actual configuration history lost: %s", done.RuntimeSnapshot)
	}
	for _, request := range p.requests {
		if !request.OmitTemperature || !request.OmitMaxTokens {
			t.Fatal("advertised unsupported parameter still sent")
		}
	}
	if runtime.Calls["novel-creation.chapter-draft"].TemperatureStatus != "不支持，未发送" {
		t.Fatal("unsupported effective parameter status missing")
	}
}

func TestAgentRevalidatesRoutesBeforeNewAndRepairCalls(t *testing.T) {
	for _, first := range []string{foundationJSON, `{"title":`} {
		input := fixturePipelineRunContext(NewMemoryRuntime())
		_ = json.Unmarshal([]byte(`{"Agent":{"modelRoutes":{"foundation-plan":"chosen"},"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}}`), &input)
		p := &queuedPipelineProvider{responses: []string{first}}
		resolves := 0
		pipeline := NewPipeline(func(_ context.Context, _, id string) (batchfactoryv11.DirectorProvider, error) {
			resolves++
			if resolves > 1 {
				return nil, errors.New("selected model was deleted or unauthorized")
			}
			if id != "chosen" {
				t.Fatalf("unexpected fallback=%s", id)
			}
			return p, nil
		})
		_, err := pipeline.Run(context.Background(), input, &recordingEventSink{})
		if (!IsCode(err, ErrModelUnavailable) && !IsCode(err, ErrModelAuthFailed)) || len(p.requests) != 1 || resolves != 2 {
			t.Fatalf("permission change did not stop next call: requests=%d resolves=%d err=%v", len(p.requests), resolves, err)
		}
	}
}

func TestAgentSecondaryRepairAndAutomationWaitBeforePaidCall(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	_, _ = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a})
	p := &queuedPipelineProvider{responses: []string{`{"title":`, foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = pipelineWithProvider(p)
	batch, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1, ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	dispatched, err := f.service.AdvanceBatches(ctx)
	if err != nil || len(dispatched) != 1 {
		t.Fatalf("dispatch=%v %v", dispatched, err)
	}
	id := dispatched[0].Run.ID
	for count := 1; count <= len(p.responses); count++ {
		current, err := f.service.Run(ctx, "alice", id)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.requests) != count {
			t.Fatalf("repair/history duplicated: actual calls=%d want=%d", len(p.requests), count)
		}
		next, err := f.service.AdvanceBatches(ctx)
		if err != nil || len(next) != 0 {
			t.Fatalf("confirmation created another run: %v %v", next, err)
		}
		saved, _ := f.store.GetBatch(ctx, "alice", batch.ID)
		if current.Status == RunSucceededWithWarnings {
			if saved.CompletedChapters != 1 || saved.TotalSavedChapters != 1 || saved.ConsecutiveFailures != 0 {
				t.Fatalf("saved warnings did not count exactly once: %+v", saved)
			}
			return
		}
		if saved.WaitingReason != "请确认后继续下一次模型调用" {
			t.Fatalf("model checkpoint mislabeled as blueprint: %s", saved.WaitingReason)
		}
		if current.Status != RunAwaitingConfirmation || saved.CompletedChapters != 0 || saved.TotalSavedChapters != 0 || saved.ConsecutiveFailures != 0 {
			t.Fatalf("waiting confirmation counted as a chapter/failure: %+v", saved)
		}
		var confirmation struct {
			ID  string `json:"confirmationId"`
			Key string `json:"callKey"`
		}
		_ = json.Unmarshal(current.RuntimeSnapshot, &confirmation)
		if count == 1 && confirmation.Key != "novel-creation.foundation-plan:repair" {
			t.Fatalf("repair did not use a distinct gated key: %s", current.RuntimeSnapshot)
		}
		if _, err = f.service.ConfirmModelCall(ctx, "alice", id, confirmation.ID); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("bounded run did not save warnings")
}

func TestAgentActualBASEJournalReplaysWithoutNewDefaultsOrResolver(t *testing.T) {
	raw, err := os.ReadFile("testdata/task5-base-agent-replay.json")
	if err != nil {
		t.Fatal(err)
	}
	var base struct {
		PromptSnapshot json.RawMessage                         `json:"promptSnapshot"`
		SourceSHA      string                                  `json:"sourceSHA"`
		Run            Run                                     `json:"run"`
		Execution      Execution                               `json:"execution"`
		Requests       []batchfactoryv11.TextCompletionRequest `json:"requests"`
	}
	if err = json.Unmarshal(raw, &base); err != nil {
		t.Fatal(err)
	}
	if base.SourceSHA != "83dab59691d99fcfeb67bf547b3e7d041920dfe9" {
		t.Fatal("wrong BASE evidence")
	}
	keys := []string{"novel-creation.foundation-plan", "novel-creation.foundation-plan:repair", "novel-creation.chapter-plan", "novel-creation.chapter-draft", "novel-creation.chapter-audit:0", "novel-creation.chapter-revision:1", "novel-creation.chapter-audit:1", "novel-creation.truth-settlement", "novel-creation.memory-index"}
	if len(base.Requests) != 9 || len(base.Execution.Calls) != 9 {
		t.Fatal("incomplete actual BASE capture")
	}
	for i, key := range keys {
		if base.Execution.Calls[key].Fingerprint != fingerprint(base.Requests[i]) {
			t.Fatalf("new request contract changed BASE fingerprint for %s", key)
		}
	}
	base.Run.PromptSnapshot = base.PromptSnapshot
	f := newServiceFixture(t, nil)
	ctx := context.Background()
	settings, _ := f.service.GetProjectSettings(ctx, "alice", f.project.ID)
	if !settings.Agent.EnableLengthNorm || !settings.Agent.EnablePostWrite || !settings.Agent.EnableAiTells {
		t.Fatal("new defaults should be enabled outside the legacy snapshot")
	}
	f.store.mu.Lock()
	f.store.runs[memoryKey("alice", base.Run.ID)] = base.Run
	f.store.executions[memoryKey("alice", base.Run.ID)] = base.Execution
	f.store.mu.Unlock()
	resolves := 0
	provider := &queuedPipelineProvider{}
	f.service.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		resolves++
		return provider, nil
	})
	done, err := f.service.Run(ctx, "alice", base.Run.ID)
	if err != nil || done.Status != RunSucceeded || resolves != 0 || len(provider.requests) != 0 {
		t.Fatalf("BASE replay regenerated paid work: status=%s resolves=%d calls=%d err=%v", done.Status, resolves, len(provider.requests), err)
	}
	after, _, _ := f.store.GetExecution(ctx, "alice", base.Run.ID)
	if len(after.Calls) != 9 {
		t.Fatal("new default steps were appended to the old run")
	}
	for key, before := range base.Execution.Calls {
		if after.Calls[key].Fingerprint != before.Fingerprint || after.Calls[key].Response != before.Response {
			t.Fatalf("BASE historical call changed: %s", key)
		}
	}
	t.Log("actual BASE requests=9; fingerprints equal=9; current replay provider=0; resolver=0; optional additions=0")
}

func TestAgentPermissionFailureIsUnsentAndStopsAutomation(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	_, _ = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a})
	provider := &queuedPipelineProvider{responses: []string{foundationJSON}}
	resolves := 0
	f.service.pipeline = NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		resolves++
		if resolves > 1 {
			return nil, errors.New("selected model removed from current authorized catalog")
		}
		return provider, nil
	})
	batch, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 2, ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	dispatched, err := f.service.AdvanceBatches(ctx)
	if err != nil || len(dispatched) != 1 {
		t.Fatalf("dispatch=%v %v", dispatched, err)
	}
	id := dispatched[0].Run.ID
	run, err := f.service.Run(ctx, "alice", id)
	if err == nil || run.Status != RunFailed || run.ErrorCode == ErrResultUnknown {
		t.Fatalf("unsent permission failure treated as uncertain: %+v %v", run, err)
	}
	execution, _, _ := f.store.GetExecution(ctx, "alice", id)
	if execution.Calls["novel-creation.foundation-plan"].State != "completed" || execution.Calls["novel-creation.chapter-plan"].State != "prepared" || hasUncertainCall(execution) {
		t.Fatal("permission validation occurred after marking may_have_sent")
	}
	if _, err = f.service.AdvanceBatches(ctx); err != nil {
		t.Fatal(err)
	}
	stopped, _ := f.store.GetBatch(ctx, "alice", batch.ID)
	if stopped.Status != BatchFailed || stopped.CompletedChapters != 0 || len(provider.requests) != 1 {
		t.Fatalf("deterministic model failure retried automatically: %+v calls=%d", stopped, len(provider.requests))
	}
}

func TestAgentBlueprintChildKeepsSameSecondaryHistoryAndOneBatchCount(t *testing.T) {
	ctx := context.Background()
	f := newServiceFixture(t, nil)
	var err error
	f.session, err = f.service.CreateSession(ctx, "alice", CreateSessionInput{ProjectID: f.project.ID, Mode: "long_original"})
	if err != nil {
		t.Fatal(err)
	}
	var a AgentSettings
	_ = json.Unmarshal([]byte(`{"requireSecondaryConfirm":true,"maxRevisionRounds":0,"enableLengthNorm":false,"enablePostWrite":false,"enableAiTells":false}`), &a)
	_, _ = f.service.UpdateProjectSettings(ctx, "alice", f.project.ID, ProjectSettings{Agent: a})
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	f.service.pipeline = pipelineWithProvider(provider)
	batch, err := f.service.CreateBatch(ctx, "alice", f.project.ID, BatchInput{SessionID: f.session.ID, TotalChapters: 1, ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	dispatched, err := f.service.AdvanceBatches(ctx)
	if err != nil || len(dispatched) != 1 {
		t.Fatal(err)
	}
	source := dispatched[0].Run
	blueprint, err := f.service.Run(ctx, "alice", source.ID)
	if err != nil || blueprint.Status != RunAwaitingConfirmation {
		t.Fatal(err)
	}
	child, err := f.service.ConfirmRun(ctx, "alice", source.ID, nil)
	if err != nil || child.ID == source.ID {
		t.Fatalf("blueprint semantics changed: %v", err)
	}
	duplicate, err := f.service.ConfirmRun(ctx, "alice", source.ID, nil)
	if err != nil || duplicate.ID != child.ID {
		t.Fatal("blueprint confirmation was not idempotent")
	}
	for calls := 2; calls <= 6; calls++ {
		current, err := f.service.Run(ctx, "alice", child.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(provider.requests) != calls {
			t.Fatalf("blueprint or history regenerated: calls=%d", len(provider.requests))
		}
		next, err := f.service.AdvanceBatches(ctx)
		if err != nil || len(next) != 0 {
			t.Fatalf("checkpoint added a second chapter run: %v %v", next, err)
		}
		saved, _ := f.store.GetBatch(ctx, "alice", batch.ID)
		if calls == 6 {
			if current.Status != RunSucceeded || saved.CompletedChapters != 1 || saved.TotalSavedChapters != 1 {
				t.Fatalf("final counter=%+v", saved)
			}
			return
		}
		if current.Status != RunAwaitingConfirmation || saved.CompletedChapters != 0 || saved.Rounds[0].RunID != child.ID || saved.Rounds[0].SourceRunID != source.ID {
			t.Fatalf("secondary checkpoint was confused with blueprint successor: %+v", saved)
		}
		var pending struct {
			ID string `json:"confirmationId"`
		}
		_ = json.Unmarshal(current.RuntimeSnapshot, &pending)
		if _, err = f.service.ConfirmModelCall(ctx, "alice", child.ID, pending.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAgentEveryPhaseRouteReachesItsActualProvider(t *testing.T) {
	input := fixturePipelineRunContext(NewMemoryRuntime())
	a := NormalizeAgentSettings(AgentSettings{}, 2000)
	order := []string{"foundation-plan", "chapter-plan", "chapter-draft", "length-normalize", "chapter-audit", "chapter-revision", "chapter-audit", "truth-settlement", "memory-index"}
	for _, phase := range order {
		a.ModelRoutes[phase] = phase + "-model"
	}
	input.Agent = &a
	provider := &queuedPipelineProvider{responses: []string{foundationJSON, chapterPlanJSON, draftJSON, draftJSON, failedAuditJSON, revisedJSON, passedAuditJSON, settlementJSON, memoryJSON}}
	var actual []string
	pipeline := NewPipeline(func(_ context.Context, _, id string) (batchfactoryv11.DirectorProvider, error) {
		actual = append(actual, id)
		return provider, nil
	})
	if _, err := pipeline.Run(context.Background(), input, &recordingEventSink{}); err != nil {
		t.Fatal(err)
	}
	if len(actual) != len(order) {
		t.Fatalf("actual phase calls=%v", actual)
	}
	for i, phase := range order {
		if actual[i] != phase+"-model" {
			t.Fatalf("phase=%s routed to %s", phase, actual[i])
		}
	}
}
