package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"testing"
)

// All responses are local queued fakes. No paid/network model is used.
func TestLongImitationLargeReferenceReallyCallsEachChunkOnce(t *testing.T) {
	ctx := context.Background()
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON(), imitationAnalysisJSON(), imitationAnalysisJSON(), foundationJSON}}
	st := NewMemoryStore()
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("chunks")
	first := strings.Repeat("甲", 50000)
	second := "乙"
	in.ReferenceText = first + second
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation {
		t.Fatalf("chunked pipeline status=%s err=%v", run.Status, e)
	}
	if len(provider.requests) != 4 {
		t.Fatal("expected two chunk analyses, merge, blueprint", len(provider.requests))
	}
	for i, want := range []string{first, second} {
		var payload struct {
			Chunk struct {
				Text  string `json:"text"`
				Index int    `json:"index"`
				Total int    `json:"total"`
			} `json:"reference_chunk"`
		}
		if e = json.Unmarshal([]byte(provider.requests[i].UserPrompt), &payload); e != nil {
			t.Fatal(e)
		}
		if payload.Chunk.Text != want || payload.Chunk.Index != i || payload.Chunk.Total != 2 {
			t.Fatal("actual chunk content/order incomplete", i)
		}
		if strings.Contains(provider.requests[i].UserPrompt, in.ReferenceText) {
			t.Fatal("whole reference duplicated into chunk call")
		}
	}
	for _, request := range provider.requests[2:] {
		if strings.Contains(request.UserPrompt, first) || strings.Contains(request.UserPrompt, `"referenceChunks"`) || strings.Contains(request.UserPrompt, `"effectiveReferenceText"`) {
			t.Fatal("summary/foundation reinjected huge source")
		}
	}
	var foundationPayload map[string]any
	_ = json.Unmarshal([]byte(provider.requests[3].UserPrompt), &foundationPayload)
	if foundationPayload["mode_inputs"] != nil || foundationPayload["notes"] != nil || foundationPayload["knowledge_materials"] != nil || foundationPayload["style_profile"] != "" {
		t.Fatal("creation source reappeared through generic input fields")
	}
	ex, _, _ := st.GetExecution(ctx, "alice", out.Run.ID)
	keys := []string{}
	for key := range ex.Calls {
		keys = append(keys, key)
	}
	if len(keys) != 4 {
		t.Fatal("chunk journal keys collided", keys)
	}
}
func TestLongImitationAnalysisHierarchyNeverMergesMoreThanFour(t *testing.T) {
	provider := &queuedPipelineProvider{}
	for i := 0; i < 7; i++ {
		provider.responses = append(provider.responses, imitationAnalysisJSON())
	}
	provider.responses = append(provider.responses, foundationJSON)
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("hierarchy")
	in.ReferenceText = strings.Repeat("甲", 200000) + "乙"
	out, e := s.StartLongImitation(context.Background(), "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(context.Background(), "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation || len(provider.requests) != 8 {
		t.Fatalf("hierarchy status=%s calls=%d err=%v", run.Status, len(provider.requests), e)
	}
	for index, want := range []int{4, 2} {
		var body struct {
			Sources []LongImitationAnalysis `json:"source_analyses"`
		}
		if e = json.Unmarshal([]byte(provider.requests[5+index].UserPrompt), &body); e != nil {
			t.Fatal(e)
		}
		if len(body.Sources) != want {
			t.Fatal("merge fanout/order changed", len(body.Sources))
		}
		if jsLength(provider.requests[5+index].UserPrompt) > 45000 {
			t.Fatal("hierarchical merge context unbounded")
		}
	}
}
func TestLongImitationRequirementsAndChunkAnalysisInheritFoundationRoute(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{`{"summary":"完整保留大补充的约束"}`, imitationAnalysisJSON(), foundationJSON}}
	models := []string{}
	pipeline := NewPipeline(func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
		if owner != "alice" {
			t.Fatal("owner changed")
		}
		models = append(models, model)
		return provider, nil
	})
	in := imitationInput("route")
	in.SupplementaryInfo = strings.Repeat("甲", 13000)
	creation := &LongImitationCreation{Input: in, EffectiveReferenceText: in.ReferenceText, ReferenceChunks: imitationChunks(in.ReferenceText), StyleDimensions: ImitationStyleDimensions(), Statistics: ImitationStyleStatistics(in.ReferenceText, "")}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.LongImitation = creation
	input.Checkpoint = true
	input.Project.TargetChapterWords = 50000
	input.Agent = &AgentSettings{Version: 1, ModelRoutes: map[string]string{"foundation-plan": "foundation-route"}}
	_, e := pipeline.Run(context.Background(), input, &recordingEventSink{})
	if e != nil || len(models) != 3 {
		t.Fatal("dedicated stages did not run", e, models)
	}
	for _, model := range models {
		if model != "foundation-route" {
			t.Fatal("dedicated stage silently selected a different model", models)
		}
	}
}
func TestLongImitationSelectedLargeStyleAndMemoNeverLeakIntoGenericPromptFields(t *testing.T) {
	ctx := context.Background()
	st := NewMemoryStore()
	largeStyle := strings.Repeat("风格要求", 5000)
	style, e := st.SaveStyleProfile(ctx, "alice", StyleProfile{Title: "大文风", Content: largeStyle})
	if e != nil {
		t.Fatal(e)
	}
	notesProject, e := st.CreateProject(ctx, "alice", CreateProjectInput{Title: "备忘资料", Genre: "other", ModelID: "fake-only"})
	if e != nil {
		t.Fatal(e)
	}
	memoText := strings.Repeat("备忘要求", 4000)
	note, e := st.CreateNote(ctx, "alice", notesProject.ID, ProjectNoteInput{Title: "备忘", Content: memoText})
	if e != nil {
		t.Fatal(e)
	}
	provider := &queuedPipelineProvider{responses: []string{`{"summary":"文风要求已完整读取并提炼"}`, imitationAnalysisJSON(), foundationJSON}}
	s := NewService(st, NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("large-style-memo")
	in.KnowledgeBookIDs = []string{style.ID}
	in.MemoReferences = []ImitationMemoReference{{notesProject.ID, note.ID}}
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation || len(provider.requests) != 3 {
		t.Fatal("large style/memo flow failed", e)
	}
	if strings.Count(provider.requests[0].UserPrompt, largeStyle) != 1 || strings.Contains(provider.requests[0].UserPrompt, memoText) {
		t.Fatal("requirements sources duplicated or wrongly included reference memo")
	}
	if strings.Count(provider.requests[1].UserPrompt, memoText) != 1 || strings.Contains(provider.requests[1].UserPrompt, largeStyle) {
		t.Fatal("memo or style reentered reference analysis more than once")
	}
	if strings.Contains(provider.requests[2].UserPrompt, memoText) || strings.Contains(provider.requests[2].UserPrompt, largeStyle) {
		t.Fatal("source body leaked into blueprint")
	}
	var body map[string]any
	_ = json.Unmarshal([]byte(provider.requests[2].UserPrompt), &body)
	if body["style_profile"] != "" || body["notes"] != nil || body["knowledge_materials"] != nil {
		t.Fatal("generic fields bypassed requirements")
	}
	config, e := s.GetLongImitation(ctx, "alice", out.Project.ID)
	if e != nil || config.Styles[0].Profile.Content != largeStyle || config.Memos[0].Content != memoText {
		t.Fatal("bounded model inputs mutated immutable full sources", e)
	}
}
func TestLongImitationOversizedAnalysisFailsLocallyWithoutRepairOrRebilling(t *testing.T) {
	var oversized LongImitationAnalysis
	_ = json.Unmarshal([]byte(imitationAnalysisJSON()), &oversized)
	for key := range oversized.Dimensions {
		oversized.Dimensions[key] = strings.Repeat("长", 2000)
	}
	raw, _ := json.Marshal(oversized)
	provider := &queuedPipelineProvider{responses: []string{string(raw)}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("oversize-analysis"))
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(context.Background(), "alice", out.Run.ID)
	if !IsCode(e, ErrInvalidOutput) || run.Status != RunFailed || len(provider.requests) != 1 {
		t.Fatal("oversize analysis triggered an extra repair call", e, len(provider.requests))
	}
	retry, e := s.RetryRun(context.Background(), "alice", out.Run.ID)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Run(context.Background(), "alice", retry.ID)
	if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 1 {
		t.Fatal("received oversized result billed again", e)
	}
}
func TestLongImitationOversizedMalformedReplyCannotBypassLimitViaRepair(t *testing.T) {
	for name, reply := range map[string]string{"unknown-field": strings.TrimSuffix(imitationAnalysisJSON(), "}") + `,"padding":"` + strings.Repeat("长", 9000) + `"}`, "truncated": `{"dimensions":"` + strings.Repeat("长", 9000)} {
		t.Run(name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: []string{reply}}
			s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
			out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("malformed-"+name))
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", out.Run.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 1 {
				t.Fatal("oversize raw reply invoked repair", e, len(provider.requests))
			}
			retry, e := s.RetryRun(context.Background(), "alice", out.Run.ID)
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", retry.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 1 {
				t.Fatal("malformed oversize reply rebilled", e)
			}
		})
	}
}
func TestLongImitationRequirementsRawEnvelopeHasIndependentLimit(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{`{"summary":"摘要","padding":"` + strings.Repeat("长", 7000) + `"}`}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("requirements-raw-limit")
	in.SupplementaryInfo = strings.Repeat("补", 13000)
	out, e := s.StartLongImitation(context.Background(), "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Run(context.Background(), "alice", out.Run.ID)
	if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 1 {
		t.Fatal("oversize requirements envelope invoked repair", e, len(provider.requests))
	}
	retry, e := s.RetryRun(context.Background(), "alice", out.Run.ID)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Run(context.Background(), "alice", retry.ID)
	if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 1 {
		t.Fatal("oversize requirements envelope rebilled", e)
	}
}
func TestLongImitationOversizedRepairReplyAlsoCachesLimitFailure(t *testing.T) {
	for name, reply := range map[string]string{"unknown-field": strings.TrimSuffix(imitationAnalysisJSON(), "}") + `,"padding":"` + strings.Repeat("长", 9000) + `"}`, "truncated": `{"dimensions":"` + strings.Repeat("长", 9000)} {
		t.Run(name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: []string{"short bad JSON", reply}}
			s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
			out, e := s.StartLongImitation(context.Background(), "alice", imitationInput("repair-limit-"+name))
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", out.Run.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 2 {
				t.Fatal("normal repair stage count changed", e)
			}
			retry, e := s.RetryRun(context.Background(), "alice", out.Run.ID)
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", retry.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != 2 {
				t.Fatal("oversized repair result rebilled on retry", e, len(provider.requests))
			}
		})
	}
}
func TestLongImitationRequirementsInitialAndRepairLimitMatrix(t *testing.T) {
	unknown := `{"summary":"摘要","padding":"` + strings.Repeat("长", 7000) + `"}`
	truncated := `{"summary":"` + strings.Repeat("长", 7000)
	semantic := `{"summary":"` + strings.Repeat("长", 4500) + `"}`
	for _, tt := range []struct {
		name    string
		replies []string
	}{{"initial-truncated", []string{truncated}}, {"initial-semantic", []string{semantic}}, {"repair-unknown", []string{"bad", unknown}}, {"repair-truncated", []string{"bad", truncated}}, {"repair-semantic", []string{"bad", semantic}}} {
		t.Run(tt.name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: tt.replies}
			s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
			in := imitationInput(tt.name)
			in.SupplementaryInfo = strings.Repeat("补", 13000)
			out, e := s.StartLongImitation(context.Background(), "alice", in)
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", out.Run.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != len(tt.replies) {
				t.Fatal("requirements limit stage count", e, len(provider.requests))
			}
			retry, e := s.RetryRun(context.Background(), "alice", out.Run.ID)
			if e != nil {
				t.Fatal(e)
			}
			_, e = s.Run(context.Background(), "alice", retry.ID)
			if !IsCode(e, ErrInvalidOutput) || len(provider.requests) != len(tt.replies) {
				t.Fatal("received requirements failure rebilled", e, len(provider.requests))
			}
			if !strings.Contains(e.Error(), "重新创建") {
				t.Fatal("limit failure does not explain recovery", e)
			}
		})
	}
}

type imitationArtifactFault struct {
	Store
	fail *bool
}

func (s imitationArtifactFault) Atomic(ctx context.Context, owner, project string, fn func(Store) error) error {
	return s.Store.Atomic(ctx, owner, project, func(tx Store) error { return fn(imitationArtifactFault{tx, s.fail}) })
}
func (s imitationArtifactFault) AtomicCreate(ctx context.Context, owner string, fn func(Store) error) error {
	return s.Store.AtomicCreate(ctx, owner, func(tx Store) error { return fn(imitationArtifactFault{tx, s.fail}) })
}
func (s imitationArtifactFault) PutContentRecord(ctx context.Context, r ContentRecord) error {
	if r.Kind == longImitationKind && r.ID == "analysis-v2-chunk-1" && *s.fail {
		*s.fail = false
		return errors.New("fake second chunk local-save failure")
	}
	return s.Store.PutContentRecord(ctx, r)
}
func TestLongImitationChunkLocalFailureReplaysSavedRawWithoutDoubleCall(t *testing.T) {
	ctx := context.Background()
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON(), imitationAnalysisJSON(), imitationAnalysisJSON(), foundationJSON}}
	st := NewMemoryStore()
	fail := true
	s := NewService(imitationArtifactFault{st, &fail}, NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("partial-local")
	in.ReferenceText = strings.Repeat("甲", 50000) + "乙"
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Run(ctx, "alice", out.Run.ID); !IsCode(e, ErrPersistenceFailed) {
		t.Fatal("local save failure not recoverable", e)
	}
	if len(provider.requests) != 2 {
		t.Fatal("wrong partial attempts", len(provider.requests))
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation || len(provider.requests) != 4 {
		t.Fatalf("partial replay repeated calls or failed: %s calls=%d err=%v", run.Status, len(provider.requests), e)
	}
}
func TestLongImitationSecondChunkUnknownNeverResends(t *testing.T) {
	ctx := context.Background()
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON()}, errors: []error{nil, errors.New("fake connection lost after send")}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("partial-unknown")
	in.ReferenceText = strings.Repeat("甲", 50000) + "乙"
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if !IsCode(e, ErrResultUnknown) || run.Status != RunPaused || len(provider.requests) != 2 {
		t.Fatal("second chunk unknown not preserved", e)
	}
	if _, e = s.RetryRun(ctx, "alice", out.Run.ID); !IsCode(e, ErrResultUnknown) {
		t.Fatal("unknown chunk explicitly retried", e)
	}
	s.Run(ctx, "alice", out.Run.ID)
	if len(provider.requests) != 2 {
		t.Fatal("unknown chunk billed again")
	}
}
func TestLongImitationKnownFailedChunkRetryReusesEarlierCompletedArtifacts(t *testing.T) {
	ctx := context.Background()
	provider := &queuedPipelineProvider{responses: []string{imitationAnalysisJSON(), "bad output", "bad repair", imitationAnalysisJSON(), imitationAnalysisJSON(), foundationJSON}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("partial-known")
	in.ReferenceText = strings.Repeat("甲", 50000) + "乙"
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal(e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e == nil || run.Status != RunFailed || len(provider.requests) != 3 {
		t.Fatal("expected known invalid second chunk", e)
	}
	retry, e := s.RetryRun(ctx, "alice", out.Run.ID)
	if e != nil {
		t.Fatal(e)
	}
	run, e = s.Run(ctx, "alice", retry.ID)
	if e != nil || run.Status != RunAwaitingConfirmation || len(provider.requests) != 6 {
		t.Fatalf("known retry did not reuse completed chunk: status=%s calls=%d err=%v", run.Status, len(provider.requests), e)
	}
	if strings.Contains(provider.requests[3].UserPrompt, strings.Repeat("甲", 50000)) {
		t.Fatal("completed first chunk resent on a new retry run")
	}
}
func TestLongImitationLargeSupplementIsPreparedWithoutShrinkingAcceptedFields(t *testing.T) {
	ctx := context.Background()
	provider := &queuedPipelineProvider{responses: []string{`{"summary":"保留完整补充中的关键要求"}`, imitationAnalysisJSON(), foundationJSON}}
	s := NewService(NewMemoryStore(), NewMemoryRuntime(), pipelineWithProvider(provider))
	in := imitationInput("large-supplement")
	in.SupplementaryInfo = strings.Repeat("补充要求", 5000)
	out, e := s.StartLongImitation(ctx, "alice", in)
	if e != nil {
		t.Fatal("accepted field range was shrunk", e)
	}
	run, e := s.Run(ctx, "alice", out.Run.ID)
	if e != nil || run.Status != RunAwaitingConfirmation {
		t.Fatalf("requirements pipeline status=%s err=%v", run.Status, e)
	}
	if len(provider.requests) != 3 || !strings.Contains(provider.requests[0].UserPrompt, in.SupplementaryInfo) {
		t.Fatal("full supplement never reached a dedicated step")
	}
	if strings.Contains(provider.requests[1].UserPrompt, in.SupplementaryInfo) || strings.Contains(provider.requests[2].UserPrompt, in.SupplementaryInfo) {
		t.Fatal("large supplement was repeatedly reinjected")
	}
	config, e := s.GetLongImitation(ctx, "alice", out.Project.ID)
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(LongImitationSummary(config))
	if !strings.Contains(string(raw), "保留完整补充中的关键要求") {
		t.Fatal("requirements summary not readable")
	}
}
