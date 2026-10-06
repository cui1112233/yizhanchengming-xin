package novelpanel

import (
	"errors"
	"testing"
)

func TestScaffoldShotsDensityKeepsEverySourceLine(t *testing.T) {
	text := "第一句。第二句！\n第三句。"
	compact := ScaffoldShots(text, DensityCompact)
	detailed := ScaffoldShots(text, DensityDetailed)
	if len(compact) != 2 {
		t.Fatalf("compact should keep one editable scaffold per source line, got %d", len(compact))
	}
	if len(detailed) != 3 {
		t.Fatalf("detailed should split sentence boundaries, got %d: %#v", len(detailed), detailed)
	}
	if detailed[0].SourceIndex != 1 || detailed[1].SourceIndex != 1 || detailed[2].SourceIndex != 2 {
		t.Fatalf("source mapping lost: %#v", detailed)
	}
}

func TestValidateShotsRequiresFullCoverageExactBasisAndVisual(t *testing.T) {
	text := "第一段原文。\n第二段原文。"
	valid := []Shot{
		{ID: "s1", SourceIndex: 1, SourceBasis: "第一段原文。", Visual: "近景，人物抬头", DurationSec: 3},
		{ID: "s2", SourceIndex: 2, SourceBasis: "第二段原文。", Visual: "中景，人物转身", DurationSec: 4},
	}
	if err := ValidateShots(text, valid); err != nil {
		t.Fatalf("valid storyboard rejected: %v", err)
	}
	cases := [][]Shot{
		{{ID: "s1", SourceIndex: 1, SourceBasis: "第一段原文。", Visual: "画面"}},
		{{ID: "s1", SourceIndex: 1, SourceBasis: "模型虚构内容", Visual: "画面"}, {ID: "s2", SourceIndex: 2, SourceBasis: "第二段原文。", Visual: "画面"}},
		{{ID: "s1", SourceIndex: 1, SourceBasis: "第一段原文。", Visual: ""}, {ID: "s2", SourceIndex: 2, SourceBasis: "第二段原文。", Visual: "画面"}},
	}
	for _, shots := range cases {
		if err := ValidateShots(text, shots); !errors.Is(err, ErrInvalid) {
			t.Fatalf("expected ErrInvalid for %#v, got %v", shots, err)
		}
	}
}
