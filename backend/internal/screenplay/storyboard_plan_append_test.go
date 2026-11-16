package screenplay

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"
)

func TestStoryboardPlanAppendOriginalGolden(t *testing.T) {
	f, e := os.Open("testdata/screenplay_plan_append_original.json.gz")
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
		Cases []struct{ Previous, Chunk, Output string }
	}
	if e = json.NewDecoder(z).Decode(&d); e != nil {
		t.Fatal(e)
	}
	for i, r := range d.Cases {
		if got := appendStoryboardPlanChunk(r.Previous, r.Chunk); got != r.Output {
			t.Errorf("case%d expected%d chars, got%d", i, len([]rune(r.Output)), len([]rune(got)))
		}
	}
}
