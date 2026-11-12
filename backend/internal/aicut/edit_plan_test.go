package aicut

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validEditPlanForTest() EditPlan {
	now := time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC)
	return EditPlan{
		ID: "plan-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: "cut-1",
		BaseRevision: 3, ModelID: "model-1", State: EditPlanReady,
		Request:    EditPlanRequest{AssetIDs: []string{"asset-1"}, Goal: "剪成一分钟竖屏视频", TargetDurationMS: 60000, AspectRatio: "9:16", IncludeCaptions: true},
		Summary:    EditPlanSummary{SourceDurationMS: 120000, EstimatedDurationMS: 60000},
		Operations: []EditOperation{{ID: "op-1", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 0, EndMS: 1000, Reason: "去掉片尾空白", Enabled: true}},
		CreatedAt:  now, UpdatedAt: now,
	}
}

func TestNormalizeEditPlanAcceptsValidPlan(t *testing.T) {
	got, err := NormalizeEditPlan(validEditPlanForTest())
	if err != nil {
		t.Fatal(err)
	}
	if got.Operations[0].Type != EditTrimClip || got.State != EditPlanReady {
		t.Fatalf("got=%#v", got)
	}
}

func TestNormalizeEditPlanRejectsUnknownOperation(t *testing.T) {
	plan := validEditPlanForTest()
	plan.Operations[0].Type = "run_shell"
	if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("err=%v, want ErrInvalidEditPlan", err)
	}
}

func TestNormalizeEditPlanRejectsInvalidTimeRanges(t *testing.T) {
	for _, operation := range []EditOperation{
		{ID: "op-1", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: -1, EndMS: 1000, Reason: "非法开始", Enabled: true},
		{ID: "op-2", Type: EditTrimClip, TrackID: "video-1", ClipID: "clip-1", StartMS: 1000, EndMS: 1000, Reason: "空区间", Enabled: true},
	} {
		plan := validEditPlanForTest()
		plan.Operations = []EditOperation{operation}
		if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
			t.Fatalf("operation=%#v err=%v", operation, err)
		}
	}
}

func TestNormalizeEditPlanRejectsDuplicateOperationIDs(t *testing.T) {
	plan := validEditPlanForTest()
	plan.Operations = append(plan.Operations, plan.Operations[0])
	if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("err=%v, want duplicate rejection", err)
	}
}

func TestNormalizeEditPlanRejectsUnknownResources(t *testing.T) {
	for _, operation := range []EditOperation{
		{ID: "op-filter", Type: EditSetFilter, TrackID: "video-1", ClipID: "clip-1", ResourceID: "filter-does-not-exist", Reason: "未知滤镜", Enabled: true},
		{ID: "op-transition", Type: EditSetTransition, TrackID: "video-1", ClipID: "clip-2", ResourceID: "transition-does-not-exist", DurationMS: 300, Reason: "未知转场", Enabled: true},
		{ID: "op-effect", Type: EditAddEffectLayer, TrackID: "effect-1", ResourceID: "effect-does-not-exist", StartMS: 0, DurationMS: 500, Reason: "未知特效", Enabled: true},
	} {
		plan := validEditPlanForTest()
		plan.Operations = []EditOperation{operation}
		if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
			t.Fatalf("operation=%#v err=%v", operation, err)
		}
	}
}

func TestNormalizeEditPlanRejectsMissingReasonAndTooManyOperations(t *testing.T) {
	plan := validEditPlanForTest()
	plan.Operations[0].Reason = ""
	if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("missing reason err=%v", err)
	}

	plan = validEditPlanForTest()
	plan.Operations = make([]EditOperation, 201)
	for i := range plan.Operations {
		plan.Operations[i] = EditOperation{ID: "op-" + strings.Repeat("x", i%5) + string(rune('A'+i%26)), Type: EditAddCaption, TrackID: "subtitle-1", StartMS: int64(i * 1000), DurationMS: 500, Text: "字幕", Reason: "补充字幕", Enabled: true}
	}
	if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("too many operations err=%v", err)
	}
}

func TestNormalizeEditPlanRejectsInvalidState(t *testing.T) {
	plan := validEditPlanForTest()
	plan.State = "mystery"
	if _, err := NormalizeEditPlan(plan); !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("err=%v", err)
	}
}

func TestParseEditPlanModelOutputAcceptsMarkdownFenceAndKeepsServerEnvelope(t *testing.T) {
	envelope := EditPlanEnvelope{ID: "plan-server", ClientRequestID: "request-server", Owner: "alice", ProjectID: "cut-1", BaseRevision: 9, ModelID: "model-7", Request: EditPlanRequest{AssetIDs: []string{"asset-1"}, Goal: "粗剪", IncludeCaptions:true, AspectRatio: "9:16"}, CreatedAt: time.Unix(100, 0).UTC(), UpdatedAt: time.Unix(100, 0).UTC()}
	raw := "```json\n{\"summary\":{\"sourceDurationMs\":5000,\"estimatedDurationMs\":3000},\"operations\":[{\"id\":\"op-1\",\"type\":\"add_caption\",\"trackId\":\"subtitle-1\",\"startMs\":0,\"durationMs\":1000,\"text\":\"开场\",\"reason\":\"补充字幕\",\"enabled\":true}],\"warnings\":[\"素材较短\"]}\n```"
	got, err := ParseEditPlanModelOutput(raw, envelope)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != envelope.ID || got.BaseRevision != 9 || got.State != EditPlanReady || got.Owner != "alice" {
		t.Fatalf("server envelope lost: %#v", got)
	}
	if len(got.Operations) != 1 || got.Operations[0].Type != EditAddCaption {
		t.Fatalf("operations=%#v", got.Operations)
	}
}

func TestParseEditPlanModelOutputRejectsTrailingContent(t *testing.T) {
	_, err := ParseEditPlanModelOutput(`{"summary":{},"operations":[]} trailing`, EditPlanEnvelope{ID: "plan-1", ClientRequestID: "request-1", Owner: "alice", ProjectID: "cut-1", ModelID: "model-1", Request: EditPlanRequest{AssetIDs: []string{"asset-1"}, Goal: "粗剪", IncludeCaptions:true, AspectRatio: "9:16"}})
	if !errors.Is(err, ErrInvalidEditPlan) {
		t.Fatalf("err=%v", err)
	}
}

func TestEditPlanRejectsDisabledOptions(t *testing.T) {
	for _, op := range []EditOperation{
		{ID: "caption", Type: EditAddCaption, TrackID: "sub", Text: "x", DurationMS: 1000, Reason: "r", Enabled: true},
		{ID: "transition", Type: EditSetTransition, TrackID: "video", ClipID: "clip", ResourceID: "transition-fade", DurationMS: 300, Reason: "r", Enabled: true},
		{ID: "filter", Type: EditSetFilter, TrackID: "video", ClipID: "clip", ResourceID: "filter-cinema", Reason: "r", Enabled: true},
		{ID: "bgm", Type: EditAddBGM, TrackID: "audio", AssetID: "asset-1", DurationMS: 1000, Reason: "r", Enabled: true},
	} {
		t.Run(string(op.Type), func(t *testing.T) {
			p := validEditPlanForTest()
			p.Request.IncludeCaptions = false
			p.Operations = []EditOperation{op}
			if _, err := NormalizeEditPlan(p); err == nil {
				t.Fatal("disabled option accepted")
			}
		})
	}
}
func TestEditPlanDurationUsesAppliedTimeline(t *testing.T) {
	p := editPlanForApply(EditOperation{ID: "move", Type: EditMoveClip, TrackID: "video-1", ClipID: "clip-2", StartMS: 5000, Reason: "r", Enabled: true})
	p.Summary.EstimatedDurationMS = 999999
	_, sum, err := ApplyEditPlan(editPlanBaseTimeline(), p)
	if err != nil {
		t.Fatal(err)
	}
	if sum.EstimatedDurationMS != 7000 {
		t.Fatalf("duration=%d", sum.EstimatedDurationMS)
	}
}
