package novelcreation

import (
	"reflect"
	"testing"
)

// 预期来自 docs/reverse-engineering/2026-11-17-luoshui-novel-style-statistics.md
// 的独立构造案例，不依赖原包 helper 作为测试依赖。
func TestStyleFingerprintDocumentedCases(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		avg     float64
		stddev  float64
		avgPara float64
		minPara int
		maxPara int
		div     float64
		top     []string
		rhet    []string
	}{
		{"空串", "", 0, 0, 0, 0, 0, 0, []string{}, []string{}},
		{"重复短句", "甲乙。甲乙。甲乙。", 2, 0, 9, 9, 9, 0.333, []string{"甲乙...(3次)"}, []string{}},
		{"两句", "甲。乙乙乙！", 2, 1, 6, 6, 6, 0.5, []string{}, []string{}},
		{"表情", "😀。😀。", 2, 0, 6, 6, 6, 0.25, []string{}, []string{}},
		{"分段", "甲。\n\n乙乙。", 1.5, 0.5, 3, 2, 3, 0.667, []string{}, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildStyleFingerprint(tc.text, "")
			if got.AvgSentenceLength != tc.avg || got.SentenceLengthStdDev != tc.stddev || got.AvgParagraphLength != tc.avgPara {
				t.Fatalf("averages=%+v want avg=%v std=%v para=%v", got, tc.avg, tc.stddev, tc.avgPara)
			}
			if got.ParagraphLengthRange.Min != tc.minPara || got.ParagraphLengthRange.Max != tc.maxPara {
				t.Fatalf("range=%+v want [%d,%d]", got.ParagraphLengthRange, tc.minPara, tc.maxPara)
			}
			if got.VocabularyDiversity != tc.div {
				t.Fatalf("diversity=%v want=%v", got.VocabularyDiversity, tc.div)
			}
			if !reflect.DeepEqual(got.TopPatterns, tc.top) {
				t.Fatalf("top=%v want=%v", got.TopPatterns, tc.top)
			}
			if !reflect.DeepEqual(got.RhetoricalFeatures, tc.rhet) {
				t.Fatalf("rhetoric=%v want=%v", got.RhetoricalFeatures, tc.rhet)
			}
		})
	}
}

func TestStyleFingerprintRhetoricalRules(t *testing.T) {
	text := "他像是风，又如是一道光。天崩地裂，翻天覆地。难道你不知道。风雨仿佛在笑。"
	got := buildStyleFingerprint(text, "")
	seen := map[string]bool{}
	for _, feature := range got.RhetoricalFeatures {
		seen[feature] = true
	}
	for _, want := range []string{"比喻(像/如/仿佛)(2处)", "夸张(2处)"} {
		if !seen[want] {
			t.Fatalf("missing %q in %v", want, got.RhetoricalFeatures)
		}
	}
	// 短句节奏会消耗两侧句末符，连续短句只算一次，因此不到阈值。
	// 出现次数不足 2 的规则不得输出。
	for _, unwanted := range []string{"反问", "拟人", "排比", "短句节奏"} {
		for feature := range seen {
			if len(feature) >= len(unwanted) && feature[:len(unwanted)] == unwanted {
				t.Fatalf("under-threshold rule %q must not appear: %v", unwanted, got.RhetoricalFeatures)
			}
		}
	}
}

func TestStyleFingerprintParallelismNonOverlapping(t *testing.T) {
	if got := countParallelism("，甲乙，甲乙，甲乙"); got != 1 {
		t.Fatalf("parallelism=%d want 1 (global non-overlapping)", got)
	}
	if got := countParallelism("，甲乙，丙丁"); got != 0 {
		t.Fatalf("parallelism=%d want 0", got)
	}
}

func TestStyleStatsKeepsLegacyFields(t *testing.T) {
	stats := styleStats("甲。乙。")
	if stats.Sentences != 2 || stats.Characters == 0 || stats.AnalysisVersion != 2 {
		t.Fatalf("legacy fields lost: %+v", stats)
	}
	if stats.AvgSentenceLength != 1 {
		t.Fatalf("fingerprint not attached: %+v", stats.StyleFingerprint)
	}
}
