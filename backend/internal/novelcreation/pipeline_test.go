package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
)

const (
	foundationJSON          = `{"title":"雨幕记者","genre":"都市悬疑","world":"旧城连日暴雨","characters":[{"name":"林澜","role":"记者","goal":"找到失踪者","conflict":"报社要求停查"}],"coreConflict":"真相与自保的冲突","volumes":[{"number":1,"title":"雨夜","goal":"找到第一条证据","chapterRange":[1,20]}]}`
	chapterPlanJSON         = `{"chapterNumber":1,"title":"没有寄件人的信","goal":"让林澜拿到线索","conflict":"线索可能是陷阱","turn":"信纸上有失踪者指纹","hook":"楼道出现监视盲区","targetWords":2000}`
	draftJSON               = `{"content":"雨水敲着窗。林澜拆开了那封没有寄件人的信。"}`
	failedAuditJSON         = `{"score":70,"issues":[{"id":"issue-1","severity":"high","description":"指纹线索没有落地","category":"plot","evidence":"正文没有识别指纹"}],"passed":false}`
	revisedJSON             = `{"strategy":"补足线索动作","content":"雨水敲着窗。林澜拆开了那封没有寄件人的信，在封口找到了失踪者的指纹。","resolvedIssueIDs":["issue-1"]}`
	passedAuditJSON         = `{"score":92,"issues":[],"passed":true}`
	settlementJSON          = `{"facts":[{"id":"fact-1","statement":"林澜拿到带有失踪者指纹的信"}],"characterChanges":[{"character":"林澜","change":"决定继续调查"}],"timeline":[{"event":"收到匿名信","position":"第一章雨夜"}],"foreshadowing":[{"thread":"监视盲区","status":"open"}]}`
	memoryJSON              = `{"summary":"林澜收到匿名信并找到指纹","entities":[{"name":"林澜","type":"character","state":"继续调查"}],"openThreads":[{"id":"thread-1","thread":"监视盲区","nextExpectation":"查明谁删除了监控"}]}`
	twoVolumeFoundationJSON = `{"title":"雨幕记者","genre":"都市悬疑","world":"旧城连日暴雨","characters":[{"name":"林澜","role":"记者","goal":"找到失踪者","conflict":"报社要求停查"}],"coreConflict":"真相与自保的冲突","volumes":[{"number":1,"title":"雨夜","goal":"找到第一条证据","chapterRange":[1,10]},{"number":2,"title":"暗河","goal":"追查真正幕后人","chapterRange":[11,20]}]}`
	chapter12PlanJSON       = `{"chapterNumber":12,"title":"暗河入口","goal":"找到幕后人路线","conflict":"线索指向老朋友","turn":"老朋友留下密码","hook":"暗河下游出现灯光","targetWords":2000}`
)

type queuedPipelineProvider struct {
	responses []string
	errors    []error
	requests  []batchfactoryv11.TextCompletionRequest
	afterCall func(int)
}

func (p *queuedPipelineProvider) Complete(_ context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	p.requests = append(p.requests, request)
	index := len(p.requests) - 1
	var response string
	var err error
	if index < len(p.responses) {
		response = p.responses[index]
	}
	if index < len(p.errors) {
		err = p.errors[index]
	}
	if p.afterCall != nil {
		p.afterCall(index)
	}
	return response, err
}

type recordingEventSink struct {
	nextID    int64
	actions   []string
	persisted []RunEvent
	published []RunEvent
}

func (s *recordingEventSink) Persist(_ context.Context, event RunEvent) (RunEvent, error) {
	s.nextID++
	event.EventID = s.nextID
	s.actions = append(s.actions, "persist:"+event.Type+":"+event.Step)
	s.persisted = append(s.persisted, event)
	return event, nil
}

func (s *recordingEventSink) Publish(_ context.Context, event RunEvent) error {
	s.actions = append(s.actions, "publish:"+event.Type+":"+event.Step)
	s.published = append(s.published, event)
	return nil
}

func fixturePipelineRunContext(runtime Runtime) RunContext {
	return RunContext{
		OwnerUsername:  "alice",
		Run:            Run{ID: "run-1", ProjectID: "project-1", SessionID: "session-1", ModelID: "text-1", ThinkingLevel: "medium"},
		Project:        Project{ID: "project-1", Title: "雨幕记者", Genre: "都市悬疑", World: "旧城连日暴雨", Protagonist: "林澜", CoreConflict: "调查失踪案", ModelID: "text-1", TargetChapterWords: 2000},
		Session:        Session{ID: "session-1", ProjectID: "project-1", Mode: "book-create"},
		UserMessage:    "写第一章，从一封匿名信开始。",
		ChapterNumber:  1,
		PreviousTruth:  json.RawMessage(`{"facts":[]}`),
		PreviousMemory: json.RawMessage(`{"summary":"","entities":[],"openThreads":[]}`),
		Runtime:        runtime,
	}
}

func pipelineWithProvider(provider batchfactoryv11.DirectorProvider) *Pipeline {
	return NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		return provider, nil
	})
}

func TestPipelineRevisesFailedAuditAndReturnsSettledBundle(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON,
		revisedJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	sink := &recordingEventSink{}
	var resolvedOwner, resolvedModelID string
	pipeline := NewPipeline(func(_ context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		resolvedOwner, resolvedModelID = owner, modelID
		return provider, nil
	})
	bundle, err := pipeline.Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), sink)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedOwner != "alice" || resolvedModelID != "text-1" {
		t.Fatalf("resolver owner=%q modelID=%q, want alice/text-1", resolvedOwner, resolvedModelID)
	}
	if bundle.FinalRunStatus != RunSucceeded {
		t.Fatalf("status=%s, want %s", bundle.FinalRunStatus, RunSucceeded)
	}
	if bundle.Chapter.Content != "雨水敲着窗。林澜拆开了那封没有寄件人的信，在封口找到了失踪者的指纹。" {
		t.Fatalf("chapter content=%q", bundle.Chapter.Content)
	}
	if len(bundle.Audits) != 2 || bundle.Audits[len(bundle.Audits)-1].Score != 92 {
		t.Fatalf("audits=%+v", bundle.Audits)
	}
	if len(bundle.Revisions) != 1 || bundle.Revisions[0].Strategy != "补足线索动作" {
		t.Fatalf("revisions=%+v", bundle.Revisions)
	}
	if len(bundle.Memories) != 1 || !json.Valid(bundle.Memories[0].Snapshot) || !json.Valid(bundle.Foundation.Snapshot) {
		t.Fatalf("bundle snapshots are not valid JSON: %+v", bundle)
	}

	wantStages := []string{"planning", "writing", "normalizing", "auditing", "revising", "settling", "indexing"}
	var gotStages []string
	for _, event := range sink.persisted {
		if event.Type == "run.step" {
			gotStages = append(gotStages, event.Step)
		}
	}
	if !reflect.DeepEqual(gotStages, wantStages) {
		t.Fatalf("stage order=%v, want %v", gotStages, wantStages)
	}
	if len(sink.actions)%2 != 0 {
		t.Fatalf("unpaired event actions: %v", sink.actions)
	}
	for index := 0; index < len(sink.actions); index += 2 {
		if !strings.HasPrefix(sink.actions[index], "persist:") || !strings.HasPrefix(sink.actions[index+1], "publish:") {
			t.Fatalf("event was not persisted before publish: %v", sink.actions[index:index+2])
		}
		if sink.published[index/2].EventID == 0 {
			t.Fatalf("published event did not use persisted identity: %+v", sink.published[index/2])
		}
	}

	wantDefinitions := []struct {
		definitionContract string
		temperature        float64
		maxTokens          int
	}{
		{novelprompt.FoundationPlan.OutputContract, 0.4, 6000},
		{novelprompt.ChapterPlan.OutputContract, 0.35, 3500},
		{novelprompt.ChapterDraft.OutputContract, 0.75, 12000},
		{novelprompt.ChapterAudit.OutputContract, 0.2, 4500},
		{novelprompt.ChapterRevision.OutputContract, 0.55, 12000},
		{novelprompt.ChapterAudit.OutputContract, 0.2, 4500},
		{novelprompt.TruthSettlement.OutputContract, 0.2, 4500},
		{novelprompt.MemoryIndex.OutputContract, 0.2, 3500},
	}
	if len(provider.requests) != len(wantDefinitions) {
		t.Fatalf("requests=%d, want %d", len(provider.requests), len(wantDefinitions))
	}
	for index, want := range wantDefinitions {
		request := provider.requests[index]
		if !strings.Contains(request.SystemPrompt, want.definitionContract) {
			t.Fatalf("request %d omitted output contract", index)
		}
		if strings.TrimSpace(request.UserPrompt) == "" || request.MaxTokens != want.maxTokens || request.Temperature != want.temperature {
			t.Fatalf("request %d was not explicitly mapped: %+v", index, request)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal([]byte(request.UserPrompt), &payload); err != nil {
			t.Fatalf("request %d user prompt is not JSON: %v", index, err)
		}
		var thinkingLevel string
		if err := json.Unmarshal(payload["thinking_level"], &thinkingLevel); err != nil || thinkingLevel != "medium" {
			t.Fatalf("request %d thinking_level=%q err=%v", index, thinkingLevel, err)
		}
		switch index {
		case 0:
			var projectBrief map[string]any
			var userMessage string
			if json.Unmarshal(payload["project_brief"], &projectBrief) != nil || projectBrief["title"] != "雨幕记者" || json.Unmarshal(payload["user_message"], &userMessage) != nil || userMessage != "写第一章，从一封匿名信开始。" {
				t.Fatalf("foundation request lost key inputs: %s", request.UserPrompt)
			}
		case 1:
			var chapterNumber int
			if json.Unmarshal(payload["chapter_number"], &chapterNumber) != nil || chapterNumber != 1 || len(payload["foundation_plan"]) == 0 {
				t.Fatalf("chapter plan request lost key inputs: %s", request.UserPrompt)
			}
		case 2:
			if len(payload["chapter_plan"]) == 0 || len(payload["foundation_plan"]) == 0 || len(payload["memory_context"]) == 0 {
				t.Fatalf("draft request lost key inputs: %s", request.UserPrompt)
			}
		case 3:
			var content string
			if json.Unmarshal(payload["chapter_content"], &content) != nil || content != "雨水敲着窗。林澜拆开了那封没有寄件人的信。" {
				t.Fatalf("audit request content=%q", content)
			}
		case 4:
			var issues []map[string]any
			if json.Unmarshal(payload["audit_issues"], &issues) != nil || len(issues) != 1 || issues[0]["id"] != "issue-1" {
				t.Fatalf("revision request lost current audit issues: %s", request.UserPrompt)
			}
		case 5:
			var content string
			if json.Unmarshal(payload["chapter_content"], &content) != nil || !strings.Contains(content, "失踪者的指纹") {
				t.Fatalf("reaudit request did not use revised content: %q", content)
			}
		case 6:
			if len(payload["previous_truth"]) == 0 || len(payload["chapter_content"]) == 0 {
				t.Fatalf("settlement request lost truth/content: %s", request.UserPrompt)
			}
		case 7:
			if len(payload["chapter_bundle"]) == 0 || len(payload["previous_memory"]) == 0 || len(payload["truth_settlement"]) == 0 {
				t.Fatalf("memory request lost bundle/memory/settlement: %s", request.UserPrompt)
			}
		}
	}
	var deltas []RunEvent
	for _, event := range sink.persisted {
		if event.Type == "assistant.delta" {
			deltas = append(deltas, event)
		}
	}
	if len(deltas) != len(provider.responses) {
		t.Fatalf("assistant deltas=%d, want %d", len(deltas), len(provider.responses))
	}
	for index, event := range deltas {
		if event.Message != provider.responses[index] {
			t.Fatalf("delta %d=%q, want complete reply %q", index, event.Message, provider.responses[index])
		}
	}
}

func TestPipelineRepairsInvalidJSONOnceUsingTheSameOutputContract(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		`{"title":`, foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunSucceeded {
		t.Fatalf("status=%s", bundle.FinalRunStatus)
	}
	if len(provider.requests) != 7 {
		t.Fatalf("requests=%d, want original plus one repair and five later calls", len(provider.requests))
	}
	repair := provider.requests[1]
	var repairPayload struct {
		InvalidOutput string `json:"invalid_output"`
		ThinkingLevel string `json:"thinking_level"`
	}
	if err := json.Unmarshal([]byte(repair.UserPrompt), &repairPayload); err != nil {
		t.Fatalf("decode repair prompt: %v", err)
	}
	if !strings.Contains(repair.SystemPrompt, novelprompt.FoundationPlan.OutputContract) || repairPayload.InvalidOutput != `{"title":` || repairPayload.ThinkingLevel != "medium" {
		t.Fatalf("repair request did not carry contract and invalid reply: %+v", repair)
	}
}

func TestPipelineRejectsStructurallyInvalidOutputAfterOneRepair(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		`{"title":"少字段"}`,
		`{"title":"仍然少字段","genre":"悬疑"}`,
	}}
	_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if !IsCode(err, ErrInvalidOutput) {
		t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("requests=%d, invalid output may be repaired only once", len(provider.requests))
	}
}

func TestPipelineRejectsUnknownMarkdownNullAndWrongTypeOutputs(t *testing.T) {
	unknownField := strings.TrimSuffix(foundationJSON, "}") + `,"unexpected":"value"}`
	tests := []struct {
		name   string
		output string
	}{
		{name: "unknown field", output: unknownField},
		{name: "markdown fence", output: "```json\n" + foundationJSON + "\n```"},
		{name: "null required field", output: strings.Replace(foundationJSON, `"title":"雨幕记者"`, `"title":null`, 1)},
		{name: "wrong field type", output: strings.Replace(foundationJSON, `"title":"雨幕记者"`, `"title":7`, 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: []string{test.output, test.output}}
			_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
			if !IsCode(err, ErrInvalidOutput) {
				t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
			}
			if len(provider.requests) != 2 {
				t.Fatalf("requests=%d, malformed output should get exactly one repair", len(provider.requests))
			}
		})
	}
}

func TestPipelineDoesNotMergeFieldsAcrossInvalidOriginalAndRepairReplies(t *testing.T) {
	originalMissingConflict := `{"title":"雨幕记者","genre":"都市悬疑","world":"旧城连日暴雨","characters":[{"name":"林澜","role":"记者","goal":"找到失踪者","conflict":"报社要求停查"}],"volumes":[{"number":1,"title":"雨夜","goal":"找到第一条证据","chapterRange":[1,20]}]}`
	repairWithOnlyConflict := `{"coreConflict":"真相与自保的冲突"}`
	provider := &queuedPipelineProvider{responses: []string{originalMissingConflict, repairWithOnlyConflict}}
	_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if !IsCode(err, ErrInvalidOutput) {
		t.Fatalf("err=%v, each reply must independently satisfy the output contract", err)
	}
	if len(provider.requests) != 2 {
		t.Fatalf("requests=%d, pipeline incorrectly accepted merged partial replies", len(provider.requests))
	}
}

func TestPipelineMapsRequestedChapterToItsOutlineVolume(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		twoVolumeFoundationJSON, chapter12PlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ChapterNumber = 12
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Chapter.Number != 12 || bundle.Volume.Number != 2 || bundle.Volume.Title != "暗河" {
		t.Fatalf("chapter=%d volume=%+v, want chapter 12 in volume 2", bundle.Chapter.Number, bundle.Volume)
	}
}

func TestPipelineRejectsChapterPlanNumberDifferentFromRequestedChapter(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		twoVolumeFoundationJSON, chapterPlanJSON, chapterPlanJSON,
	}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ChapterNumber = 12
	_, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if !IsCode(err, ErrInvalidOutput) {
		t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
	}
	if len(provider.requests) != 3 {
		t.Fatalf("requests=%d, mismatched plan should fail after one repair", len(provider.requests))
	}
}

func TestPipelineRejectsRequestedChapterOutsideEveryOutlineVolume(t *testing.T) {
	chapter25Plan := `{"chapterNumber":25,"title":"越界章","goal":"不应被接受","conflict":"卷区间缺失","turn":"无","hook":"无","targetWords":2000}`
	provider := &queuedPipelineProvider{responses: []string{
		twoVolumeFoundationJSON, chapter25Plan, chapter25Plan,
	}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ChapterNumber = 25
	_, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if !IsCode(err, ErrInvalidOutput) {
		t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
	}
	if len(provider.requests) != 3 {
		t.Fatalf("requests=%d, out-of-outline chapter should fail after one repair", len(provider.requests))
	}
}

func TestPipelineRejectsNonStringOptionalAuditFields(t *testing.T) {
	tests := []struct {
		name  string
		audit string
	}{
		{name: "null category", audit: `{"score":70,"issues":[{"id":"issue-1","severity":"high","description":"问题","category":null}],"passed":false}`},
		{name: "numeric category", audit: `{"score":70,"issues":[{"id":"issue-1","severity":"high","description":"问题","category":7}],"passed":false}`},
		{name: "null evidence", audit: `{"score":70,"issues":[{"id":"issue-1","severity":"high","description":"问题","evidence":null}],"passed":false}`},
		{name: "object evidence", audit: `{"score":70,"issues":[{"id":"issue-1","severity":"high","description":"问题","evidence":{"line":1}}],"passed":false}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: []string{
				foundationJSON, chapterPlanJSON, draftJSON, test.audit, test.audit,
			}}
			_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
			if !IsCode(err, ErrInvalidOutput) {
				t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
			}
			if len(provider.requests) != 5 {
				t.Fatalf("requests=%d, audit should fail after exactly one repair", len(provider.requests))
			}
		})
	}
}

func TestPipelineRejectsAuditPassedFlagInconsistentWithThresholdAndBlockingIssues(t *testing.T) {
	tests := []struct {
		name  string
		audit string
	}{
		{name: "passing score marked failed without blocker", audit: `{"score":92,"issues":[],"passed":false}`},
		{name: "failing score marked passed", audit: `{"score":70,"issues":[],"passed":true}`},
		{name: "blocking issue marked passed", audit: `{"score":92,"issues":[{"id":"block-1","severity":"high","description":"关键事实冲突"}],"passed":true}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &queuedPipelineProvider{responses: []string{
				foundationJSON, chapterPlanJSON, draftJSON, test.audit, test.audit,
			}}
			_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
			if !IsCode(err, ErrInvalidOutput) {
				t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
			}
			if len(provider.requests) != 5 {
				t.Fatalf("requests=%d, inconsistent audit should fail after one repair", len(provider.requests))
			}
		})
	}
}

func TestPipelineRevisesPassingScoreWhenAuditHasBlockingIssue(t *testing.T) {
	blockedAudit := `{"score":92,"issues":[{"id":"block-1","severity":"high","description":"关键事实冲突"}],"passed":false}`
	blockingRevision := `{"strategy":"修复关键事实冲突","content":"雨水敲着窗。林澜核对证据后，确认指纹来自失踪者。","resolvedIssueIDs":["block-1"]}`
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, blockedAudit,
		blockingRevision, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunSucceeded || bundle.Chapter.Content != "雨水敲着窗。林澜核对证据后，确认指纹来自失踪者。" {
		t.Fatalf("bundle status=%s content=%q", bundle.FinalRunStatus, bundle.Chapter.Content)
	}
	if len(bundle.Revisions) != 1 || len(bundle.Audits) != 2 {
		t.Fatalf("revisions=%d audits=%d", len(bundle.Revisions), len(bundle.Audits))
	}
}

func TestPipelineRejectsRevisionResolvedIssueIDsOutsideCurrentAudit(t *testing.T) {
	badRevision := `{"strategy":"虚构修复","content":"新正文","resolvedIssueIDs":["issue-from-another-audit"]}`
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON, badRevision, badRevision,
	}}
	_, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if !IsCode(err, ErrInvalidOutput) {
		t.Fatalf("err=%v, want %s", err, ErrInvalidOutput)
	}
	if len(provider.requests) != 6 {
		t.Fatalf("requests=%d, revision must get exactly one contract repair", len(provider.requests))
	}
}

func TestPipelineKeepsBestContentWhenTwoRevisionsDoNotImproveAudit(t *testing.T) {
	firstRevision := `{"strategy":"第一轮","content":"更差的第一稿","resolvedIssueIDs":["issue-1"]}`
	secondRevision := `{"strategy":"第二轮","content":"仍然更差的第二稿","resolvedIssueIDs":["issue-1"]}`
	audit60 := `{"score":60,"issues":[{"id":"issue-1","severity":"high","description":"问题仍在"}],"passed":false}`
	audit65 := `{"score":65,"issues":[{"id":"issue-1","severity":"high","description":"问题仍在"}],"passed":false}`
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, failedAuditJSON,
		firstRevision, audit60, secondRevision, audit65, settlementJSON, memoryJSON,
	}}
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), fixturePipelineRunContext(NewMemoryRuntime()), &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunSucceededWithWarnings {
		t.Fatalf("status=%s, want %s", bundle.FinalRunStatus, RunSucceededWithWarnings)
	}
	if bundle.Chapter.Content != "雨水敲着窗。林澜拆开了那封没有寄件人的信。" {
		t.Fatalf("pipeline did not retain best content: %q", bundle.Chapter.Content)
	}
	if len(bundle.Revisions) != 2 || len(bundle.Audits) != 3 {
		t.Fatalf("revisions=%d audits=%d", len(bundle.Revisions), len(bundle.Audits))
	}
}

func TestPipelineMapsCancellationDeadlineAndUnavailableSeparately(t *testing.T) {
	tests := []struct {
		name      string
		factory   ProviderResolver
		provider  *queuedPipelineProvider
		wantCode  string
		cancelRun bool
	}{
		{name: "context canceled", provider: &queuedPipelineProvider{errors: []error{context.Canceled}}, wantCode: ErrRunCancelled},
		{name: "deadline", provider: &queuedPipelineProvider{errors: []error{context.DeadlineExceeded}}, wantCode: ErrGenerationTimeout},
		{name: "ordinary provider error", provider: &queuedPipelineProvider{errors: []error{errors.New("provider transport failed")}}, wantCode: ErrModelUnavailable},
		{name: "model unavailable", factory: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return nil, errors.New("configured model is offline")
		}, wantCode: ErrModelUnavailable},
		{name: "nil provider", factory: func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
			return nil, nil
		}, wantCode: ErrModelUnavailable},
		{name: "runtime cancellation", provider: &queuedPipelineProvider{responses: []string{foundationJSON}}, wantCode: ErrRunCancelled, cancelRun: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime := NewMemoryRuntime()
			factory := test.factory
			if factory == nil {
				if test.cancelRun {
					test.provider.afterCall = func(call int) {
						if call == 0 {
							if err := runtime.Cancel(context.Background(), "run-1"); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				factory = func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
					return test.provider, nil
				}
			}
			_, err := NewPipeline(factory).Run(context.Background(), fixturePipelineRunContext(runtime), &recordingEventSink{})
			if !IsCode(err, test.wantCode) {
				t.Fatalf("err=%v, want %s", err, test.wantCode)
			}
		})
	}
}

func TestPipelineStopsBeforeResolverAndProviderWhenAlreadyCancelled(t *testing.T) {
	runtime := NewMemoryRuntime()
	if err := runtime.Cancel(context.Background(), "run-1"); err != nil {
		t.Fatal(err)
	}
	resolverCalled := false
	provider := &queuedPipelineProvider{responses: []string{foundationJSON}}
	pipeline := NewPipeline(func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
		resolverCalled = true
		return provider, nil
	})
	_, err := pipeline.Run(context.Background(), fixturePipelineRunContext(runtime), &recordingEventSink{})
	if !IsCode(err, ErrRunCancelled) {
		t.Fatalf("err=%v, want %s", err, ErrRunCancelled)
	}
	if resolverCalled || len(provider.requests) != 0 {
		t.Fatalf("cancelled run reached resolver=%v providerCalls=%d", resolverCalled, len(provider.requests))
	}
}

func TestPipelineCheckpointStopsAfterBlueprint(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{foundationJSON}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.Checkpoint = true
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunAwaitingConfirmation {
		t.Fatalf("status=%s, want %s", bundle.FinalRunStatus, RunAwaitingConfirmation)
	}
	if len(provider.requests) != 1 {
		t.Fatalf("provider calls=%d, checkpoint must stop after the blueprint", len(provider.requests))
	}
	if strings.TrimSpace(bundle.Chapter.Content) != "" {
		t.Fatalf("checkpoint must not produce chapter body: %q", bundle.Chapter.Content)
	}
	if !strings.Contains(string(bundle.Foundation.Snapshot), "雨幕记者") || !strings.Contains(string(bundle.RuntimeSnapshot), "foundationPlan") {
		t.Fatalf("binding snapshot=%s runtime=%s", bundle.Foundation.Snapshot, bundle.RuntimeSnapshot)
	}
}

func TestPipelineReusesConfirmedFoundationWithoutBlueprintCall(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ConfirmedFoundation = json.RawMessage(foundationJSON)
	bundle, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{})
	if err != nil {
		t.Fatal(err)
	}
	if bundle.FinalRunStatus != RunSucceeded {
		t.Fatalf("status=%s", bundle.FinalRunStatus)
	}
	if len(provider.requests) != 5 {
		t.Fatalf("provider calls=%d, confirmed foundation must skip blueprint generation", len(provider.requests))
	}
}

func TestPipelineInjectsModeInputsStyleAndRules(t *testing.T) {
	provider := &queuedPipelineProvider{responses: []string{
		foundationJSON, chapterPlanJSON, draftJSON, passedAuditJSON, settlementJSON, memoryJSON,
	}}
	input := fixturePipelineRunContext(NewMemoryRuntime())
	input.ModeInputs = map[string]string{"referenceText": "参考段落"}
	input.StyleProfile = "冷峻短句"
	input.Rules = []string{"不得出现血腥描写"}
	if _, err := pipelineWithProvider(provider).Run(context.Background(), input, &recordingEventSink{}); err != nil {
		t.Fatal(err)
	}
	for index, request := range provider.requests[:3] {
		for _, expected := range []string{"参考段落", "冷峻短句", "不得出现血腥描写"} {
			if !strings.Contains(request.UserPrompt, expected) {
				t.Fatalf("request[%d] missing %q: %s", index, expected, request.UserPrompt)
			}
		}
	}
}
