package batchfactorytext

import (
	"image"
	"math"
	"testing"
)

func TestDefaultCoverTitleUsesBookBracketsWhiteAndUpperPosition(t *testing.T) {
	for _, title := range []string{"文本功能验收书", "《文本功能验收书》"} {
		if got := FormatCoverTitle(title); got != "《文本功能验收书》" {
			t.Fatal(got)
		}
	}
	layout := DefaultCoverLayout("3:4")
	if layout.TextColor != "#FFFFFF" {
		t.Fatal("cover title is not white")
	}
	in := RenderInput{Text: FormatCoverTitle("文本功能验收书"), Layout: layout, Background: Background{Kind: "image"}, BackgroundImage: DefaultCoverImage(), VerticalCenterRatio: 0.38}
	actual, _ := renderCard(t, Renderer{}, in)
	base := image.NewRGBA(actual.Bounds())
	if err := drawBackground(base, in.Background, in.BackgroundImage); err != nil {
		t.Fatal(err)
	}
	top, bottom := actual.Bounds().Dy(), 0
	for y := 0; y < actual.Bounds().Dy(); y++ {
		for x := 0; x < actual.Bounds().Dx(); x++ {
			if actual.RGBAAt(x, y) != base.RGBAAt(x, y) {
				top = min(top, y)
				bottom = max(bottom, y)
			}
		}
	}
	if bottom <= top || math.Abs(float64(top+bottom)/2-float64(actual.Bounds().Dy())*0.38) > float64(actual.Bounds().Dy())*0.05 {
		t.Fatalf("title position %d..%d", top, bottom)
	}
}
