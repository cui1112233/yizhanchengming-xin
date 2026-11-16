package prompt

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestScreenplayNativePlanningOriginalMessages(t *testing.T) {
	f, e := os.Open("testdata/screenplay_planner_original.json.gz")
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
		Guides         []struct{ Density, Continuation, Text string }
		FirstMessages  []struct{ Role, Content string } `json:"first_messages"`
		SecondMessages []struct{ Role, Content string } `json:"second_messages"`
		ContinueCases  []struct{ Input, Output string } `json:"continue_cases"`
		Previous       struct {
			Rows []struct {
				Index        int
				Plan, Output string
			}
			Calls []struct {
				Messages []struct{ Role, Content string }
			}
		}
	}
	if e = json.NewDecoder(z).Decode(&d); e != nil {
		t.Fatal(e)
	}
	for _, r := range d.Guides {
		if got := ScreenplayNativePlanGuide("OWN_WORK", "OWN_EPISODE", "OWN_PROMPT", r.Density, r.Continuation); got != r.Text {
			t.Fatalf("guide differs %s/%s", r.Density, r.Continuation)
		}
	}
	script := "第1集\n1-1 操场 日 外\nAlice：下一幕由我创造。\n△Alice走进Park。"
	system, user := ScreenplayNativePlanningRequest("本作品", "第1集", "OWN_PROMPT", "OWN_STORYBOARD_RULE", "standard", "", script)
	if system != d.FirstMessages[0].Content || user != d.FirstMessages[1].Content {
		a, b := []rune(user), []rune(d.FirstMessages[1].Content)
		i := 0
		for i < len(a) && i < len(b) && a[i] == b[i] {
			i++
		}
		endA, endB := i+80, i+80
		if endA > len(a) {
			endA = len(a)
		}
		if endB > len(b) {
			endB = len(b)
		}
		t.Fatalf("first request differs at%d (systemmatch=%t) CM=%q original=%q lengths=%d/%d", i, system == d.FirstMessages[0].Content, string(a[i:endA]), string(b[i:endB]), len(a), len(b))
	}
	for _, r := range d.ContinueCases {
		if got := ScreenplayNativePlanContinue(r.Input); got != r.Output {
			t.Fatalf("continuation differs for %d chars\nexpected %s\nactual %s", len([]rune(r.Input)), r.Output, got)
		}
	}
	var previous string
	for _, r := range d.Previous.Rows {
		if r.Index == 1 && len([]rune(r.Plan)) > 3000 {
			previous = r.Plan
		}
	}
	_, user = ScreenplayNativePlanningRequest("本作品", "second", "OWN_PROMPT", "OWN_STORYBOARD_RULE", "standard", previous, script)
	if !reflect.DeepEqual(user, d.Previous.Calls[0].Messages[1].Content) {
		t.Fatal("previous episode excerpt request differs")
	}
}
