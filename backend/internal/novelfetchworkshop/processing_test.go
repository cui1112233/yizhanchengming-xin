package novelfetchworkshop

import "testing"

func TestProcessRawTextPreservesFullRawAndLimitsProcessedText(t *testing.T) {
	body := ProcessRawText("春风又绿江南岸", 4)
	if body.OriginalRaw != "春风又绿江南岸" {
		t.Fatalf("OriginalRaw = %q", body.OriginalRaw)
	}
	if body.ProcessedText != "春风又绿" {
		t.Fatalf("ProcessedText = %q", body.ProcessedText)
	}
	if body.OriginalRawChars != 7 || body.ProcessedChars != 4 || body.MaxTxt != 4 {
		t.Fatalf("body = %+v", body)
	}
}

func TestProcessRawTextDoesNotTruncateShortText(t *testing.T) {
	body := ProcessRawText("正文内容", 20)
	if body.ProcessedText != "正文内容" || body.OriginalRawChars != 4 || body.ProcessedChars != 4 {
		t.Fatalf("body = %+v", body)
	}
}

func TestProcessRawTextHandlesEmptyAndInvalidMaxTxt(t *testing.T) {
	for _, input := range []struct {
		name string
		raw  string
		max  int
	}{
		{name: "empty", raw: "", max: 10},
		{name: "zero", raw: "正文", max: 0},
		{name: "negative", raw: "正文", max: -1},
	} {
		t.Run(input.name, func(t *testing.T) {
			body := ProcessRawText(input.raw, input.max)
			if body.OriginalRawChars < 0 || body.ProcessedChars < 0 || body.MaxTxt < 0 {
				t.Fatalf("body = %+v", body)
			}
			if body.OriginalRaw != input.raw {
				t.Fatalf("OriginalRaw = %q", body.OriginalRaw)
			}
			if input.max <= 0 && (body.ProcessedText != "" || body.ProcessedChars != 0 || body.MaxTxt != 0) {
				t.Fatalf("invalid max body = %+v", body)
			}
		})
	}
}
