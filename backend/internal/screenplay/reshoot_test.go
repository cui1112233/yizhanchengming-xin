package screenplay

import (
	"errors"
	"testing"
	"time"
)

func TestValidateReshootRange(t *testing.T) {
	if err := ValidateReshootRange(0, 3.2); err != nil {
		t.Fatalf("正常片段应通过：%v", err)
	}
	cases := []struct{ start, end float64 }{{-1, 2}, {2, 2}, {2, 2.05}}
	for _, item := range cases {
		if err := ValidateReshootRange(item.start, item.end); !errors.Is(err, ErrInvalidArgument) {
			t.Fatalf("范围 %.2f-%.2f 应被拒绝，实际 %v", item.start, item.end, err)
		}
	}
}

func TestNormalizeReshootDuration(t *testing.T) {
	cases := []struct{ duration, start, end, want float64 }{
		{0, 1.2, 4.6, 3},
		{0, 1.2, 1.6, 1},
		{45, 0, 4, 30},
		{0.2, 0, 4, 1},
		{12, 0, 4, 12},
	}
	for _, item := range cases {
		if got := NormalizeReshootDuration(item.duration, item.start, item.end); got != item.want {
			t.Fatalf("时长 %.1f（%.1f-%.1f）期望 %.0f，实际 %.0f", item.duration, item.start, item.end, item.want, got)
		}
	}
}

func TestReshootPromptKeepsFrameDescription(t *testing.T) {
	if got := ReshootPrompt("  "); got != ReshootFramePromptPrefix {
		t.Fatalf("空提示词应只保留首尾帧说明，实际 %q", got)
	}
	if got := ReshootPrompt("把外套换成红色"); got != ReshootFramePromptPrefix+"，把外套换成红色" {
		t.Fatalf("提示词应接在首尾帧说明之后，实际 %q", got)
	}
}

func TestPlanSplice(t *testing.T) {
	plan, err := PlanSplice(12, 3, 5.5)
	if err != nil || !plan.HasHead || !plan.HasTail || plan.HeadEnd != 3 || plan.TailStart != 5.5 || plan.TailEnd != 12 {
		t.Fatalf("中段重拍应保留前后段：%+v err=%v", plan, err)
	}
	plan, err = PlanSplice(12, 0, 4)
	if err != nil || plan.HasHead || !plan.HasTail {
		t.Fatalf("从 0 开始只保留后段：%+v err=%v", plan, err)
	}
	plan, err = PlanSplice(12, 8, 12.4)
	if err != nil || !plan.HasHead || plan.HasTail {
		t.Fatalf("到尾结束只保留前段：%+v err=%v", plan, err)
	}
	if _, err = PlanSplice(0, 0, 2); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("未知时长应拒绝，实际 %v", err)
	}
	if _, err = PlanSplice(10, 11, 12); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("起点超出时长应拒绝，实际 %v", err)
	}
}

func TestMediaStateReshootRoundTrip(t *testing.T) {
	state := MediaState{}
	task := ReshootTask{ShotID: "shot-1", Seq: 3, Start: 1, End: 4, Status: ReshootStatusGenerating}
	state.PutReshoot(task)
	saved, ok := state.FindReshoot("shot-1")
	if !ok || saved.Seq != 3 || saved.UpdatedAt.IsZero() {
		t.Fatalf("写入后应能按分镜取回：%+v", saved)
	}
	if !state.MutateReshoot("shot-1", func(item *ReshootTask) { item.Status = ReshootStatusDone; item.Progress = 100 }) {
		t.Fatal("修改应成功")
	}
	saved, _ = state.FindReshoot("shot-1")
	if saved.Status != ReshootStatusDone || saved.Progress != 100 {
		t.Fatalf("修改未写回：%+v", saved)
	}
	if _, ok := state.FindReshoot("missing"); ok {
		t.Fatal("不存在的分镜不应返回记录")
	}
	if saved.UpdatedAt.After(time.Now().Add(time.Second)) {
		t.Fatal("时间戳应为当前时间")
	}
}
