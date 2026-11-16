package screenplay

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func originalCompletePlanningFixture(t *testing.T) string {
	t.Helper()
	f, e := os.Open("testdata/storyboard_plan_coverage_original.json.gz")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var d struct {
		Cases []struct {
			Input, Density string
			Output         StoryboardPlanCoverage
		}
	}
	if e = json.NewDecoder(z).Decode(&d); e != nil {
		t.Fatal(e)
	}
	for _, r := range d.Cases {
		if r.Density == "standard" && r.Output.OK {
			return r.Input
		}
	}
	t.Fatal("missing original complete fixture")
	return ""
}

func TestStoryboardPlanningPersistsPartialAndCompleteWithoutRebilling(t *testing.T) {
	for _, complete := range []bool{false, true} {
		t.Run(map[bool]string{false: "partial", true: "complete"}[complete], func(t *testing.T) {
			ctx := context.Background()
			_, store, _ := newStoryboardService(t, PermissionBypass)
			p, _ := store.GetProject(ctx, "alice", "p1")
			cfg := EffectiveCreativeConfig(p)
			cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
			if _, e := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision); e != nil {
				t.Fatal(e)
			}
			authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
			state, e := authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "本集", ScriptText: "自造剧本"}}}, 0)
			if e != nil {
				t.Fatal(e)
			}
			body := "## 一、总览\n部分稿"
			want := "partial"
			if complete {
				body = originalCompletePlanningFixture(t)
				want = "succeeded"
			}
			gen := &fakeGenerator{responses: []string{body}}
			auto := false
			in := StoryboardPlanRequest{AutoContinue: &auto, ChapterID: "c", Model: "own-model", ExpectedRevision: &state.Revision, IdempotencyKey: "one"}
			result, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen)
			if e != nil {
				t.Fatal(e)
			}
			if result.Chapter.PlanningStatus != want || result.Chapter.PlanningCoverage == nil || result.Chapter.PlanningCoverage.OK != complete {
				t.Fatalf("wrong completion: %+v", result)
			}
			if _, e = authors.GenerateStoryboardPlan(ctx, "alice", "p1", in, gen); e != nil || gen.callCount() != 1 {
				t.Fatal("repeated request billed", e, gen.callCount())
			}
			saved, e := authors.Load(ctx, "alice", "p1")
			if e != nil || !reflect.DeepEqual(saved.Chapters[0].PlanningCoverage, result.Chapter.PlanningCoverage) {
				t.Fatal("coverage not persisted", e)
			}
			retry := in
			retry.IdempotencyKey = "empty"
			retry.ExpectedRevision = &saved.Revision
			empty := &fakeGenerator{responses: []string{""}}
			failed, e := authors.GenerateStoryboardPlan(ctx, "alice", "p1", retry, empty)
			if e == nil || failed.Chapter.PlanningText != body || !reflect.DeepEqual(failed.Chapter.PlanningCoverage, result.Chapter.PlanningCoverage) {
				t.Fatal("empty output destroyed prior plan", failed, e)
			}
		})
	}
}
