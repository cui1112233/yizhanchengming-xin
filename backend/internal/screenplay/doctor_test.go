package screenplay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func doctorTestClock() time.Time { return time.Date(2026, 11, 19, 2, 10, 0, 0, time.UTC) }

func TestNormalizeDoctorReviewComputesPassFromBlockers(t *testing.T) {
	raw := "```json\n" + `{"issues":[
		{"id":"ep1_opening_weak","severity":"major","episode":1,"location":"第1场","dimension":"节奏推进","problem":"开篇铺垫过长","suggestion":"开篇即冲突","evidence":"前3场无冲突"},
		{"id":"ghost_character","severity":"blocker","location":"分镜3","dimension":"资产引用","problem":"出现剧本没有的人物「白桃」"}
	],"suggested_actions":["rewrite_episode_1"],"summary":"有阻断问题，先修人物引用"}` + "\n```"
	review, err := normalizeDoctorReview(raw, "screenplay", map[string]any{"episodes": 1}, doctorTestClock())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if review.Pass || !review.BlockNextStep {
		t.Fatalf("blocker should fail the gate: %+v", review)
	}
	if review.Doctor != "screenplay" || review.Mode != "review" {
		t.Fatalf("doctor/mode wrong: %+v", review)
	}
	if review.ReviewedAt != "2026-11-19T02:10:00Z" {
		t.Fatalf("reviewed_at = %s", review.ReviewedAt)
	}
	if len(review.Issues) != 2 {
		t.Fatalf("issues = %d", len(review.Issues))
	}
	if review.Issues[1].ID != "ghost_character" || review.Issues[1].Severity != DoctorSeverityBlocker {
		t.Fatalf("issue[1] = %+v", review.Issues[1])
	}
	if review.Issues[0].Suggestion != "开篇即冲突" || review.Issues[0].Evidence != "前3场无冲突" {
		t.Fatalf("issue[0] fields lost: %+v", review.Issues[0])
	}
	if len(review.SuggestedActions) != 1 || review.SuggestedActions[0] != "rewrite_episode_1" {
		t.Fatalf("actions = %v", review.SuggestedActions)
	}
	if review.Summary != "有阻断问题，先修人物引用" {
		t.Fatalf("summary = %q", review.Summary)
	}
}

func TestNormalizeDoctorReviewMajorMinorStillPass(t *testing.T) {
	raw := `{"issues":[{"id":"a","severity":"major","problem":"爽点弱"},{"id":"b","severity":"minor","problem":"台词偏长"}],"suggested_actions":[],"summary":"整体可用"}`
	review, err := normalizeDoctorReview(raw, "storyboard", map[string]any{"episode": 1}, doctorTestClock())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if !review.Pass || review.BlockNextStep {
		t.Fatalf("major/minor must not block: %+v", review)
	}
}

func TestNormalizeDoctorReviewUnknownSeverityDefaultsMinorAndDedupesIDs(t *testing.T) {
	raw := `{"issues":[{"id":"x","severity":"urgent","problem":"p1"},{"id":"x","severity":"","problem":"p2"},{}],"suggested_actions":[" "],"summary":""}`
	review, err := normalizeDoctorReview(raw, "screenplay", nil, doctorTestClock())
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(review.Issues) != 2 {
		t.Fatalf("empty-problem entries must be dropped, got %d", len(review.Issues))
	}
	if review.Issues[0].Severity != DoctorSeverityMinor || review.Issues[1].Severity != DoctorSeverityMinor {
		t.Fatalf("severity defaults wrong: %+v", review.Issues)
	}
	if review.Issues[0].ID == review.Issues[1].ID {
		t.Fatalf("duplicate ids not deduped: %v %v", review.Issues[0].ID, review.Issues[1].ID)
	}
	if len(review.SuggestedActions) != 0 {
		t.Fatalf("blank action kept: %v", review.SuggestedActions)
	}
}

func TestNormalizeDoctorReviewRejectsNonJSON(t *testing.T) {
	if _, err := normalizeDoctorReview("模型说话了但没有 JSON", "screenplay", nil, doctorTestClock()); err == nil || !strings.Contains(err.Error(), "医生审核 JSON") {
		t.Fatalf("expected JSON rejection, got %v", err)
	}
}

func TestReviewScreenplayDoctorUsesActiveDocument(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "doc-project", OwnerUsername: "alice", Title: "剧本医生"}); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveManualDocument(ctx, store, "alice", "doc-project", "# 第1集\n\n第一场\n林川走进巷口。"); err != nil {
		t.Fatal(err)
	}
	var gotSystem, gotUser string
	executor := func(_ context.Context, systemPrompt, userPrompt string) (string, error) {
		gotSystem, gotUser = systemPrompt, userPrompt
		return `{"issues":[],"suggested_actions":["discuss_with_user"],"summary":"通过"}`, nil
	}
	review, err := ReviewScreenplayDoctor(ctx, store, "alice", "doc-project", doctorTestClock(), executor)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if !strings.Contains(gotSystem, "剧本医生") || !strings.Contains(gotUser, "林川走进巷口") {
		t.Fatalf("prompts wrong: system=%q user=%q", gotSystem, gotUser)
	}
	if !review.Pass || review.Summary != "通过" {
		t.Fatalf("review = %+v", review)
	}
	if review.Scope["episodes"] != 0 && review.Scope["episodes"] != documentEpisodesForTest(store, "alice", "doc-project") {
		t.Fatalf("scope episodes unexpected: %v", review.Scope)
	}
}

func documentEpisodesForTest(store Store, owner, projectID string) int {
	document, _ := store.ActiveDocument(context.Background(), owner, projectID)
	return document.EpisodeCount
}

func TestReviewScreenplayDoctorRequiresScript(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "empty-project", OwnerUsername: "alice", Title: "无剧本"}); err != nil {
		t.Fatal(err)
	}
	_, err := ReviewScreenplayDoctor(ctx, store, "alice", "empty-project", doctorTestClock(), func(context.Context, string, string) (string, error) { return "{}", nil })
	if !errors.Is(err, ErrNoActiveScript) {
		t.Fatalf("expected ErrNoActiveScript, got %v", err)
	}
}

func TestReviewStoryboardDoctorBuildsShotLines(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "sb-project", OwnerUsername: "alice", Title: "分镜医生"}); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveManualDocument(ctx, store, "alice", "sb-project", "# 第1集\n\n第一场\n林川走进巷口。"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveStoryboard(ctx, Storyboard{
		ProjectID: "sb-project", OwnerUsername: "alice", EpisodeNo: 1, Engine: EngineXiaoluoV7,
		Shots: []StoryboardShot{
			{ID: "d1", Seq: 1, Characters: "林川", Scene: "巷口", Line: "谁？", ImagePrompt: "巷口夜景", VideoDesc: "镜头推进"},
			{ID: "d2", Seq: 2, Characters: "白桃", Scene: "", Line: "", ImagePrompt: "卧室", VideoDesc: ""},
		},
	}); err != nil {
		t.Fatal(err)
	}
	var gotUser string
	executor := func(_ context.Context, systemPrompt, userPrompt string) (string, error) {
		gotUser = userPrompt
		if !strings.Contains(systemPrompt, "分镜词医生") && !strings.Contains(systemPrompt, "分镜医生") {
			t.Fatalf("system prompt should be storyboard doctor: %q", systemPrompt)
		}
		return `{"issues":[{"id":"v2","severity":"blocker","location":"分镜2","dimension":"video_desc 结构缺失","problem":"视频提示词为空"}],"suggested_actions":["patch_shot_field"],"summary":"1 个阻断"}`, nil
	}
	review, err := ReviewStoryboardDoctor(ctx, store, "alice", "sb-project", 1, EngineXiaoluoV7, doctorTestClock(), executor)
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	for _, want := range []string{"林川走进巷口", "分镜1", "人物：林川", "场景：巷口", "视频提示词：", "分镜2"} {
		if !strings.Contains(gotUser, want) {
			t.Fatalf("user prompt missing %q: %q", want, gotUser)
		}
	}
	if review.Pass || !review.BlockNextStep {
		t.Fatalf("blocker must fail gate: %+v", review)
	}
	if review.Scope["shots"] != 2 || review.Scope["episode"] != 1 {
		t.Fatalf("scope = %v", review.Scope)
	}
}

func TestReviewStoryboardDoctorRequiresShots(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if _, err := store.CreateProject(ctx, Project{ID: "sb-empty", OwnerUsername: "alice", Title: "无分镜"}); err != nil {
		t.Fatal(err)
	}
	_, err := ReviewStoryboardDoctor(ctx, store, "alice", "sb-empty", 1, EngineXiaoluoV7, doctorTestClock(), func(context.Context, string, string) (string, error) { return "{}", nil })
	if err == nil {
		t.Fatalf("expected rejection for missing/empty storyboard")
	}
}
