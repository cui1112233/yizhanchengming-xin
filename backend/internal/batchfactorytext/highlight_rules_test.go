package batchfactorytext

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSelectedWrappedDialogueHighlightsWholeQuote(t *testing.T) {
	in := RenderInput{Text: "旁白。\n“" + strings.Repeat("完整的对话", 6) + "。”", Layout: LayoutConfig{FontSize: 64, HighlightRules: "lines", HighlightLines: "3"}}
	_, selected := renderCard(t, Renderer{}, in)
	in.Layout.HighlightRules = "dialogue"
	_, dialogue := renderCard(t, Renderer{}, in)
	if !bytes.Equal(selected, dialogue) {
		t.Fatal("selecting the dialogue continuation line did not highlight the complete quote")
	}
	text := "第一句。第二句。"
	r := sentenceHighlightAt(text, 1, nil)
	if string([]rune(text)[r.Start:r.End]) != "第一句。" {
		t.Fatal("sentence selection leaked into the next sentence", r)
	}
}

func TestDialogueHighlightContinuesAcrossPagesAndDraftEdits(t *testing.T) {
	source := "旁白：“第一句。\n第二句。\n最后一句。”尾声。"
	page := "第二句。\n"
	start := utf8.RuneCountInString(source[:strings.Index(source, page)])
	end := start + utf8.RuneCountInString(page)
	for _, draft := range []string{page, "改后的第二句。\n"} {
		ranges, ok := DialoguePageRanges(source, draft, start, end)
		if !ok || len(ranges) != 1 || ranges[0].Start != 0 || ranges[0].End != utf8.RuneCountInString(draft) {
			t.Fatal(ranges, ok)
		}
	}
}

func TestDialogueHighlightRangesAndLineRules(t *testing.T) {
	text := "旁白：“你好。”\n他说：\"再见\"。她说：「可以」。"
	ranges := DialogueHighlightRanges(text)
	if len(ranges) != 3 {
		t.Fatal(ranges)
	}
	for _, r := range ranges {
		if !strings.ContainsAny(string([]rune(text)[r.Start:r.End]), "你好再见可以") {
			t.Fatal(r)
		}
	}
	for _, bad := range []string{"0", "-1", "a", "1,,3", "1001"} {
		if _, err := parseHighlightLines(bad); err == nil {
			t.Fatal("invalid line accepted", bad)
		}
	}
	lines, err := parseHighlightLines("1，3, 5")
	if err != nil || len(lines) != 3 || !lines[3] {
		t.Fatal(lines, err)
	}
}

func TestRenderedLineHighlightUsesVisualLinesAndResetsPerPage(t *testing.T) {
	in := RenderInput{Text: "第一行。\n  \n第二行。\n第三行。", Layout: LayoutConfig{HighlightRules: "lines", HighlightLines: "1,3"}}
	img, _ := renderCard(t, Renderer{}, in)
	bands := redBands(img)
	if len(bands) != 2 {
		t.Fatalf("wanted first and third visual line red, got %v", bands)
	}
	in.Layout.FontSize = 64
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) != 2 {
		t.Fatal("line selection did not follow actual layout")
	}
	in.Text = strings.Repeat("完整文字", 16)
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) <= 2 {
		t.Fatal("wrapped sentence was only partially highlighted")
	}
	count := len(redBands(img))
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) != count {
		t.Fatal("visual line numbers did not reset for a new page")
	}
	in.Layout.HighlightRules = "off"
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) != 0 {
		t.Fatal("off still red")
	}
	in.Layout.HighlightRules = "dialogue"
	in.Text = "旁白。\n“完整对话。”"
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) != 1 {
		t.Fatal("dialogue not highlighted")
	}
	in.Layout.HighlightRules = "dialogue,lines"
	in.Layout.HighlightLines = "1"
	img, _ = renderCard(t, Renderer{}, in)
	if len(redBands(img)) != 2 {
		t.Fatal("combined dialogue and line rules failed")
	}
	if _, _, err := (Renderer{}).Render(context.Background(), RenderInput{Text: "字", Layout: LayoutConfig{HighlightRules: "lines", HighlightLines: "0"}}); err == nil {
		t.Fatal("invalid selection accepted")
	}
}
func redBands(img *image.RGBA) []int {
	var bands []int
	active := false
	for y := 0; y < img.Bounds().Dy(); y++ {
		red := false
		for x := 0; x < img.Bounds().Dx(); x++ {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if c.R > 140 && int(c.R) > int(c.G)*2 && int(c.R) > int(c.B)*2 {
				red = true
				break
			}
		}
		if red && !active {
			bands = append(bands, y)
		}
		active = red
	}
	return bands
}
