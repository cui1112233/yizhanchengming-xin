package screenplay

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPlanningExistingOriginalGolden(t *testing.T) {
	f, err := os.Open("testdata/screenplay_plan_existing_original.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var data struct {
		Cases []struct {
			Label, Seed, Mode, Text string
			Success                 bool
			Rounds                  int
			Messages                []TextMessage
		}
	}
	if err = json.NewDecoder(z).Decode(&data); err != nil {
		t.Fatal(err)
	}
	if len(data.Cases) != 12 {
		t.Fatal("native case coverage changed", len(data.Cases))
	}
	for _, row := range data.Cases {
		t.Run(row.Label, func(t *testing.T) {
			gen := &planningRoundGenerator{mode: row.Mode, complete: originalCompletePlanningFixture(t)}
			out, e := generateStoryboardPlanContinuation(context.Background(), gen, "OWN_FAKE_MODEL", row.Messages[0].Content, row.Messages[1].Content, "standard", row.Seed)
			if (e == nil) != row.Success {
				t.Fatal("native success mismatch", e)
			}
			if len(gen.requests) != 1 || !reflect.DeepEqual(gen.requests[0].Messages, row.Messages) {
				t.Fatal("native first messages differ", len(gen.requests))
			}
			if row.Success && (out.Text != row.Text || out.Rounds != row.Rounds) {
				t.Fatalf("native output differs: chars %d/%d rounds%d/%d", len([]rune(out.Text)), len([]rune(row.Text)), out.Rounds, row.Rounds)
			}
		})
	}
}
