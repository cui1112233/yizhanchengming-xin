package batchfactorytext

import (
	"context"
	"strings"
	"testing"
)

func TestReadingParagraphsPreserveSource(t *testing.T) {
	text := strings.Repeat("春风吹过山川，今天我们一起读一本好书。", 12)
	c := DefaultRenderLayout()
	breaks := readingBreaks(text, c, 888)
	if len(breaks) < 4 {
		t.Fatal("long paragraph has no reading breaks")
	}
	pages, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: c}, (Renderer{}).MeasureFunc(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	var joined strings.Builder
	for _, p := range pages {
		joined.WriteString(p.Text)
	}
	if joined.String() != text {
		t.Fatal("source text was modified")
	}
}

func TestReadingBlockVerticallyCentered(t *testing.T) {
	img, _ := renderCard(t, Renderer{}, RenderInput{Text: "一段简短的正文。"})
	top, bottom := img.Rect.Dy(), 0
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			if img.RGBAAt(x, y) != img.RGBAAt(0, 0) {
				top = min(top, y)
				bottom = max(bottom, y)
			}
		}
	}
	if delta := (top+bottom)/2 - img.Rect.Dy()/2; delta < -15 || delta > 15 {
		t.Fatalf("text center delta %d", delta)
	}
}
