package screenplay

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

type planningRoundGenerator struct {
	mode, complete string
	requests       []TextRequest
}

func (g *planningRoundGenerator) Generate(_ context.Context, r TextRequest) (TextResponse, error) {
	g.requests = append(g.requests, r)
	n := len(g.requests)
	if g.mode == "error" || g.mode == "late-error" && n > 1 {
		return TextResponse{}, errors.New("OWN_CONTROLLED_LLM_FAILURE")
	}
	text := "## 一、总览\nOWN_CONTROLLED_PARTIAL_PLAN"
	if g.mode == "complete" || g.mode == "complete-second" && n > 1 {
		text = g.complete
	}
	if g.mode == "empty" || g.mode == "late-empty" && n > 1 {
		text = ""
	}
	if g.mode == "plain-append" {
		text = strings.Repeat("a", 100)
		if n > 1 {
			text = strings.Repeat("b", 100)
		}
	}
	return TextResponse{Content: text}, nil
}
func TestStoryboardPlanningOriginalRoundGolden(t *testing.T) {
	f, e := os.Open("testdata/screenplay_plan_rounds_original.json.gz")
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
			Label string
			Input struct {
				AutoContinue *bool `json:"auto_continue"`
				MaxRounds    *int  `json:"max_rounds"`
			}
			Output struct {
				Success bool
				Data    struct {
					Rounds    int
					PlanChars int `json:"plan_chars"`
					Complete  bool
				}
			}
			Calls           []struct{ Messages []TextMessage }
			ChapterReadback struct {
				Data []struct{ StoryboardResult json.RawMessage }
			} `json:"chapter_readback"`
		}
	}
	if e = json.NewDecoder(z).Decode(&d); e != nil {
		t.Fatal(e)
	}
	complete := ""
	for _, r := range d.Cases {
		if r.Label == "complete" {
			var raw struct{ Plan string }
			if e = json.Unmarshal(r.ChapterReadback.Data[0].StoryboardResult, &raw); e != nil {
				t.Fatal(e)
			}
			complete = raw.Plan
		}
	}
	if complete == "" {
		t.Fatal("missing original complete output")
	}
	for _, row := range d.Cases {
		t.Run(row.Label, func(t *testing.T) {
			gen := &planningRoundGenerator{mode: row.Label, complete: complete}
			system, user := row.Calls[0].Messages[0].Content, row.Calls[0].Messages[1].Content
			limit := storyboardPlanningRoundLimit(StoryboardPlanRequest{AutoContinue: row.Input.AutoContinue, MaxRounds: row.Input.MaxRounds})
			result, e := generateStoryboardPlanRounds(context.Background(), gen, "OWN_FAKE_MODEL", system, user, "standard", limit)
			if row.Output.Success != (e == nil) {
				t.Fatalf("original success%t; CM error%v", row.Output.Success, e)
			}
			if !row.Output.Success {
				return
			}
			var saved struct{ Plan string }
			if e = json.Unmarshal(row.ChapterReadback.Data[0].StoryboardResult, &saved); e != nil {
				t.Fatal(e)
			}
			if result.Text != saved.Plan || result.Rounds != row.Output.Data.Rounds || len([]rune(result.Text)) != row.Output.Data.PlanChars || EvaluateStoryboardPlan(result.Text, "standard").OK != row.Output.Data.Complete {
				t.Fatalf("round result differs: rounds%d/%d chars%d/%d", result.Rounds, row.Output.Data.Rounds, len([]rune(result.Text)), row.Output.Data.PlanChars)
			}
			// Provider retries are below the business-round boundary. Native late-error
			// captures repeated identical requests there; compare our attempted round.
			for i, req := range gen.requests {
				if !reflect.DeepEqual(req.Messages, row.Calls[i].Messages) {
					t.Fatalf("round%d message history differs", i+1)
				}
			}
		})
	}
}
