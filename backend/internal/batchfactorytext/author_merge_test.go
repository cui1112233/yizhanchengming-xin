package batchfactorytext

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestMergeAuthorPageEdit(t *testing.T) {
	for _, tc := range []struct {
		name, author, shown string
		start, end          int
		edit, want          string
		conflict            bool
	}{
		{"only other page", "甲词。\n她走。", "乙词。\n她走。", 4, 7, "她笑。", "甲词。\n她笑。", false},
		{"preserve imported line endings", "甲词。\r\n她走。", "乙词。\n她走。", 4, 7, "她笑。", "甲词。\r\n她笑。", false},
		{"intentional transformed word", "序甲词终。", "序蓝色描写终。", 0, 7, "序亲笔终。", "序亲笔终。", false},
		{"hunks around transformed token", "蓝甲词红。", "蓝乙词红。", 0, 5, "青乙词黄。", "青甲词黄。", false},
		{"emoji and changed length", "🌿甲。\n🙂好。", "🌿长短语。\n🙂好。", 6, 9, "🙂对。", "🌿甲。\n🙂对。", false},
		{"repeated anchor ambiguity", "开猫尾。", "开猫猫尾。", 0, 5, "开狗猫尾。", "", true},
		{"deleting one indistinguishable copy", "甲乙。", "乙乙。", 0, 3, "乙。", "", true},
		{"inserting beside removed author text", "甲隐乙", "甲乙", 0, 2, "甲新乙", "", true},
		{"preserve hidden author text on neighboring replacement", "甲隐乙", "甲乙", 0, 2, "丙乙", "丙隐乙", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MergeAuthorPageEdit(tc.author, Version{ProductionText: tc.shown}, tc.start, tc.end, tc.edit)
			if tc.conflict {
				if !errors.Is(err, ErrConflict) {
					t.Fatalf("want conflict, got %q %v", got, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("want %q, got %q %v", tc.want, got, err)
			}
		})
	}
}

// Enumerate every common subsequence alignment for short strings and compare
// its mandatory pairs against the production algorithm's unchanged anchors.
func TestMergeAuthorAnchorsAgainstExhaustiveAlignments(t *testing.T) {
	words := []string{""}
	for length := 1; length <= 4; length++ {
		for mask := 0; mask < 1<<length; mask++ {
			w := make([]rune, length)
			for i := range w {
				w[i] = '甲'
				if mask&(1<<i) != 0 {
					w[i] = '乙'
				}
			}
			words = append(words, string(w))
		}
	}
	for _, left := range words {
		for _, right := range words {
			a, b := []rune(left), []rune(right)
			current, expected := make([]int, len(a)), make([]int, len(a))
			best := -1
			var find func(int, int, int)
			find = func(i, j, count int) {
				if i == len(a) {
					if count > best {
						copy(expected, current)
						best = count
					} else if count == best {
						for k := range current {
							if current[k] != expected[k] {
								expected[k] = -1
							}
						}
					}
					return
				}
				current[i] = -1
				find(i+1, j, count)
				for k := j; k < len(b); k++ {
					if a[i] == b[k] {
						current[i] = k
						find(i+1, k+1, count+1)
					}
				}
			}
			find(0, 0, 0)
			budget := 10000
			changes, err := editRanges(a, b, &budget)
			if err != nil {
				t.Fatal(err)
			}
			got := make([]int, len(a))
			i, j := 0, 0
			for _, h := range changes {
				for i < h.start {
					got[i] = j
					i++
					j++
				}
				for i < h.end {
					got[i] = -1
					i++
				}
				j = h.replacementEnd
			}
			for i < len(a) {
				got[i] = j
				i++
				j++
			}
			for k := range expected {
				if got[k] != expected[k] {
					t.Fatalf("%q→%q: anchors %v want %v", left, right, got, expected)
				}
			}
		}
	}
}

func TestMergeAuthorVerifiedLengthChangingTrace(t *testing.T) {
	before, after := strings.Repeat("甲", 5000), strings.Repeat("乙", 12000)
	trace, _ := json.Marshal([]map[string]string{{"Before": before, "After": after}})
	v := Version{ProductionText: after + "🙂末。", ChangeTrace: trace}
	got, err := MergeAuthorPageEdit(before+"🙂末。", v, 12000, 12003, "🙂新。")
	if err != nil || got != before+"🙂新。" {
		t.Fatalf("verified trace failed: %v", err)
	}
	v.ChangeTrace = json.RawMessage(`[{"Before":"甲","After":"错"}]`)
	if _, err = MergeAuthorPageEdit(before+"🙂末。", v, 12000, 12003, "🙂新。"); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale trace accepted: %v", err)
	}
}

func TestMergeAuthorMaximumLength(t *testing.T) {
	author := strings.Repeat("正文。", 33330) + "甲词。末。"
	shown := strings.Replace(author, "甲词", "乙词", 1)
	n := len([]rune(shown))
	got, err := MergeAuthorPageEdit(author, Version{ProductionText: shown}, n-2, n, "终。")
	if err != nil || got != strings.Replace(author, "末。", "终。", 1) {
		t.Fatalf("large sparse edit: %v", err)
	}
	if _, err = MergeAuthorPageEdit(strings.Repeat("甲", 100000), Version{ProductionText: strings.Repeat("乙", 100000)}, 1, 2, "丁"); !errors.Is(err, ErrConflict) {
		t.Fatalf("unbounded dense edit: %v", err)
	}
}
