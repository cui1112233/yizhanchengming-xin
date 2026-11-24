package batchfactorytext

import (
	"context"
	"strings"
	"testing"
)

func TestSmartPaginationFollowsFontGeometryNot150Characters(t *testing.T) {
	text := strings.Repeat("春风吹过山川，今天我们一起读一本好书。\n", 40)
	renderer := Renderer{}
	run := func(size float64, mode string, target int) []PageDraft {
		t.Helper()
		c := DefaultRenderLayout()
		c.FontSize = size
		c.PaginationMode = mode
		c.TargetCharacters = target
		pages, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: c}, renderer.MeasureFunc(context.Background()))
		if err != nil {
			t.Fatal(err)
		}
		var joined strings.Builder
		for _, page := range pages {
			joined.WriteString(page.Text)
			m, err := renderer.Measure(page.Text, c)
			if err != nil || !m.Fits {
				t.Fatalf("page overflow: %+v %v", m, err)
			}
		}
		if joined.String() != text {
			t.Fatal("original text lost or duplicated")
		}
		return pages
	}
	small := run(28, "auto", 150)
	large := run(56, "auto", 150)
	if len(small) >= len(large) {
		t.Fatalf("smaller text should use fewer pages: %d vs %d", len(small), len(large))
	}
	if len([]rune(small[0].Text)) <= 150 {
		t.Fatal("auto mode still capped at 150")
	}
	otherTarget := run(28, "auto", 20)
	if len(otherTarget) != len(small) {
		t.Fatal("auto mode used inactive word target")
	}
	target := run(28, "target", 150)
	t.Logf("same source: auto 28px=%d pages, auto 56px=%d pages, target 150=%d pages; auto first page=%d chars", len(small), len(large), len(target), len([]rune(small[0].Text)))
	if len(target) <= len(small) {
		t.Fatal("target mode no longer honors target")
	}
}

func TestSmartPaginationRejectsUnknownMode(t *testing.T) {
	_, err := Paginate(context.Background(), PaginationInput{Text: "正文。", Layout: LayoutConfig{PaginationMode: "invalid"}}, runeCapacity(100))
	if err == nil {
		t.Fatal("unknown pagination mode accepted")
	}
}
