package screenplay

import (
	"errors"
	"testing"
)

func TestNewBatchHasNineStepsInSpecOrder(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionAuto)
	if len(batch.Steps) != 9 {
		t.Fatalf("steps = %d, want 9", len(batch.Steps))
	}
	want := []string{
		"0_ingest", "1_screenplay", "2_screenplay_doctor", "3_assets", "4_preset_doctor",
		"5_storyboard", "6_storyboard_doctor", "7_video", "8_finalize",
	}
	for i, key := range want {
		if StepDefinitions[i].Key != key {
			t.Fatalf("step[%d] = %s, want %s", i, StepDefinitions[i].Key, key)
		}
		if batch.Steps[key] == nil || batch.Steps[key].Status != StepPending {
			t.Fatalf("step %s 初始状态不是 pending", key)
		}
	}
	if batch.Status != BatchRunning {
		t.Fatalf("status = %s, want running", batch.Status)
	}
}

func TestStepLabelsMatchSpec(t *testing.T) {
	want := map[string]string{
		"0_ingest": "整理原文", "1_screenplay": "生成剧本", "2_screenplay_doctor": "剧本审核",
		"3_assets": "资产提取", "4_preset_doctor": "预设质检", "5_storyboard": "按集分镜",
		"6_storyboard_doctor": "分镜审核", "7_video": "批量出视频", "8_finalize": "收尾编号",
	}
	for key, label := range want {
		def, ok := StepByKey(key)
		if !ok {
			t.Fatalf("缺少步键 %s", key)
		}
		if def.Label != label {
			t.Fatalf("%s label = %s, want %s", key, def.Label, label)
		}
	}
	if !IsEpisodicStep("5_storyboard") || !IsEpisodicStep("6_storyboard_doctor") || IsEpisodicStep("3_assets") {
		t.Fatal("按集步骤判定错误")
	}
	for _, key := range []string{"2_screenplay_doctor", "4_preset_doctor", "6_storyboard_doctor"} {
		if !IsDoctorStep(key) {
			t.Fatalf("%s 应为医生步", key)
		}
	}
}

func TestNextUnitWalksNineSteps(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	for i, def := range StepDefinitions {
		unit, err := NextUnit(batch)
		if err != nil {
			t.Fatalf("第 %d 步 NextUnit 失败: %v", i, err)
		}
		if unit.StepKey != def.Key {
			t.Fatalf("unit = %s, want %s", unit.StepKey, def.Key)
		}
		if err := StartStep(batch, unit); err != nil {
			t.Fatalf("StartStep %s: %v", def.Key, err)
		}
		CompleteStep(batch, unit, []string{"artifact-" + def.Key})
	}
	if batch.Status != BatchCompleted {
		t.Fatalf("status = %s, want completed", batch.Status)
	}
	if _, err := NextUnit(batch); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("完成后 NextUnit err = %v, want ErrInvalidState", err)
	}
}

func TestEpisodicStepIteratesEpisodes(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	// 完成 0-4 步。
	for _, key := range []string{"0_ingest", "1_screenplay", "2_screenplay_doctor", "3_assets", "4_preset_doctor"} {
		batch.Steps[key].Status = StepDone
	}
	storyboard := batch.Steps["5_storyboard"]
	storyboard.TotalEpisodes = 3

	for ep := 1; ep <= 3; ep++ {
		unit, err := NextUnit(batch)
		if err != nil {
			t.Fatalf("第 %d 集 NextUnit 失败: %v", ep, err)
		}
		if unit.StepKey != "5_storyboard" || unit.Episode != ep {
			t.Fatalf("unit = %+v, want 5_storyboard ep%d", unit, ep)
		}
		if err := StartStep(batch, unit); err != nil {
			t.Fatalf("StartStep: %v", err)
		}
		CompleteStep(batch, unit, []string{"chapter_" + string(rune('0'+ep)) + "_storyboard.json"})
		if batch.Steps["5_storyboard"].Status == StepDone && ep < 3 {
			t.Fatalf("第 %d 集后整步被提前标记完成", ep)
		}
	}
	if batch.Steps["5_storyboard"].Status != StepDone {
		t.Fatalf("三集完成后步状态 = %s", batch.Steps["5_storyboard"].Status)
	}
	unit, err := NextUnit(batch)
	if err != nil || unit.StepKey != "6_storyboard_doctor" {
		t.Fatalf("next = %+v err=%v, want 6_storyboard_doctor", unit, err)
	}
}

func TestDoctorGateBlocksOnlyWhenBothFlags(t *testing.T) {
	cases := []struct {
		name       string
		pass       bool
		block      bool
		wantStatus StepStatus
		wantBatch  BatchStatus
	}{
		{"pass", true, true, StepDone, BatchRunning},
		{"fail-but-not-block", false, false, StepDone, BatchRunning},
		{"fail-and-block", false, true, StepBlocked, BatchWaitingUser},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
			unit := Unit{StepKey: "2_screenplay_doctor", Label: "剧本审核"}
			batch.Steps["0_ingest"].Status = StepDone
			batch.Steps["1_screenplay"].Status = StepDone
			batch.Steps["2_screenplay_doctor"].Status = StepInProgress
			ApplyDoctorReview(batch, unit, &Review{Pass: tc.pass, BlockNextStep: tc.block, Summary: "review"})
			if got := batch.Steps["2_screenplay_doctor"].Status; got != tc.wantStatus {
				t.Fatalf("step status = %s, want %s", got, tc.wantStatus)
			}
			if batch.Status != tc.wantBatch {
				t.Fatalf("batch status = %s, want %s", batch.Status, tc.wantBatch)
			}
		})
	}
}

func TestBlockedBatchRejectsRunNextUntilGateCleared(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	batch.Steps["0_ingest"].Status = StepDone
	batch.Steps["1_screenplay"].Status = StepDone
	unit := Unit{StepKey: "2_screenplay_doctor", Label: "剧本审核"}
	batch.Steps["2_screenplay_doctor"].Status = StepInProgress
	ApplyDoctorReview(batch, unit, &Review{Pass: false, BlockNextStep: true})

	if _, err := NextUnit(batch); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("闸门未解除时 NextUnit err = %v, want ErrInvalidState", err)
	}
	if err := StartStep(batch, unit); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("闸门未解除时 StartStep err = %v", err)
	}

	// 人工把 review 改成 pass=true 后恢复。
	batch.Steps["2_screenplay_doctor"].Review.Pass = true
	if err := ResumeBatch(batch); err != nil {
		t.Fatalf("ResumeBatch: %v", err)
	}
	if batch.Status != BatchRunning {
		t.Fatalf("status = %s, want running", batch.Status)
	}
	next, err := NextUnit(batch)
	if err != nil || next.StepKey != "2_screenplay_doctor" {
		t.Fatalf("next = %+v err=%v", next, err)
	}
}

func TestEpisodeDoctorGate(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	step := batch.Steps["6_storyboard_doctor"]
	step.Status = StepInProgress
	step.TotalEpisodes = 2
	unit := Unit{StepKey: "6_storyboard_doctor", Episode: 2, Label: "分镜审核"}
	ApplyDoctorReview(batch, unit, &Review{Pass: false, BlockNextStep: true})
	if step.Status != StepBlocked || batch.Status != BatchWaitingUser {
		t.Fatalf("step=%s batch=%s, want blocked/waiting_user", step.Status, batch.Status)
	}
	if len(batch.PendingActions) != 1 || batch.PendingActions[0].RequiresConfirm != true {
		t.Fatalf("pending actions = %+v", batch.PendingActions)
	}
}

func TestRewriteMarksStaleAndClearsGate(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	for _, key := range []string{"0_ingest", "1_screenplay", "2_screenplay_doctor"} {
		batch.Steps[key].Status = StepDone
	}
	unit := Unit{StepKey: "4_preset_doctor", Label: "预设质检"}
	batch.Steps["4_preset_doctor"].Status = StepInProgress
	ApplyDoctorReview(batch, unit, &Review{Pass: false, BlockNextStep: true})
	hit := MarkStale(batch, []string{"3_assets", "4_preset_doctor"})
	if hit != 2 {
		t.Fatalf("stale hit = %d, want 2", hit)
	}
	if batch.Steps["3_assets"].Status != StepStale {
		t.Fatalf("3_assets = %s, want stale", batch.Steps["3_assets"].Status)
	}
	if len(batch.PendingActions) != 0 {
		t.Fatalf("pending actions 未清理: %+v", batch.PendingActions)
	}
	if batch.Status != BatchRunning {
		t.Fatalf("status = %s, want running", batch.Status)
	}
	next, err := NextUnit(batch)
	if err != nil || next.StepKey != "3_assets" {
		t.Fatalf("next = %+v err=%v, want 3_assets", next, err)
	}
}

func TestFailStepDoesNotEnterWaitingUser(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	unit := Unit{StepKey: "3_assets", Label: "资产提取"}
	batch.Steps["3_assets"].Status = StepInProgress
	FailStep(batch, unit, "AI 未返回有效JSON")
	if batch.Steps["3_assets"].Status != StepFailed {
		t.Fatalf("step status = %s, want failed", batch.Steps["3_assets"].Status)
	}
	if batch.Status != BatchFailed {
		t.Fatalf("batch status = %s, want failed", batch.Status)
	}
	if batch.LastError != "AI 未返回有效JSON" {
		t.Fatalf("last_error = %q", batch.LastError)
	}
}

func TestValidateTransition(t *testing.T) {
	cases := []struct {
		from, to BatchStatus
		ok       bool
	}{
		{BatchRunning, BatchWaitingUser, true},
		{BatchWaitingUser, BatchRunning, true},
		{BatchRunning, BatchCompleted, true},
		{BatchPaused, BatchRunning, true},
		{BatchCompleted, BatchRunning, false},
		{BatchFailed, BatchCompleted, false},
	}
	for _, tc := range cases {
		err := ValidateTransition(tc.from, tc.to)
		if tc.ok && err != nil {
			t.Fatalf("%s→%s 应允许: %v", tc.from, tc.to, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s→%s 应拒绝", tc.from, tc.to)
		}
	}
}

func TestPermissionModeDefaults(t *testing.T) {
	if ParsePermissionMode("") != PermissionAuto {
		t.Fatal("交互批次默认应为 auto")
	}
	if ParsePermissionMode("bypass") != PermissionBypass {
		t.Fatal("bypass 解析失败")
	}
	if NormalizeBatchStatus("awaiting_user") != BatchWaitingUser {
		t.Fatal("awaiting_user 别名未归一")
	}
}

func TestSkipDoneStepsIsIdempotent(t *testing.T) {
	batch := NewBatch("b1", "p1", "alice", "s1", PermissionBypass)
	batch.Steps["0_ingest"].Status = StepSkipped
	unit, err := NextUnit(batch)
	if err != nil || unit.StepKey != "1_screenplay" {
		t.Fatalf("unit = %+v err=%v", unit, err)
	}
}

func TestPendingEpisodeInitiallyOne(t *testing.T) {
	step := &StepState{Key: "5_storyboard", Status: StepPending, TotalEpisodes: 4}
	if got := nextEpisode(step); got != 1 {
		t.Fatalf("nextEpisode = %d, want 1", got)
	}
	step.Episodes = map[int]*EpisodeState{
		1: {EpisodeNo: 1, Status: StepDone},
		2: {EpisodeNo: 2, Status: StepSkipped},
		3: {EpisodeNo: 3, Status: StepFailed},
	}
	if got := nextEpisode(step); got != 3 {
		t.Fatalf("nextEpisode = %d, want 3（失败集可重跑）", got)
	}
}

func TestCompleteStepDoesNotDuplicateArtifactsOnRetry(t *testing.T) {
	batch := NewBatch("project-1", "owner", "session-1", "batch-1", PermissionBypass)
	unit := Unit{StepKey: "0_ingest"}
	if err := StartStep(batch, unit); err != nil {
		t.Fatalf("StartStep: %v", err)
	}
	CompleteStep(batch, unit, []string{"source://project-1"})
	CompleteStep(batch, unit, []string{"source://project-1", "document://doc-1"})

	got := batch.Step("0_ingest").Artifacts
	if len(got) != 2 || got[0] != "source://project-1" || got[1] != "document://doc-1" {
		t.Fatalf("unexpected deduplicated artifacts: %#v", got)
	}
}
