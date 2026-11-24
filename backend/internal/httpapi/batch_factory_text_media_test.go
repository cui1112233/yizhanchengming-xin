package httpapi

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	txt "qiantie/backend/internal/batchfactorytext"
	"strings"
	"testing"
)

func TestBatchFactoryTextPresetPreviewReflowsWithoutChangingDraft(t *testing.T) {
	original := strings.Repeat("完整的句子需要保留，不可以删除任何正文。\n", 12)
	input := txt.RenderInput{Text: original, Layout: txt.LayoutConfig{PaginationMode: "auto", FontSize: 46, FontWeight: 500, LineHeight: 1.65, ParagraphSpacing: 32, Padding: 96, MaxHeightRatio: 0.75}}
	renderer := txt.Renderer{}
	if _, _, err := renderer.Render(context.Background(), input); err == nil {
		t.Fatal("fixture must overflow a single page")
	}
	prepared, count, err := textPreviewInput(context.Background(), input, renderer)
	if err != nil {
		t.Fatal(err)
	}
	if count < 2 || prepared.Text == original {
		t.Fatal("overflowing draft did not reflow")
	}
	if input.Text != original {
		t.Fatal("preview modified the original draft")
	}
	if !strings.HasPrefix(original, prepared.Text) {
		t.Fatal("preview is not a source slice")
	}
	if _, _, err := renderer.Render(context.Background(), prepared); err != nil {
		t.Fatalf("reflow preview still overflows: %v", err)
	}
	small := input
	small.Text = "一段完整的短句。"
	prepared, count, err = textPreviewInput(context.Background(), small, renderer)
	if err != nil || count != 1 || prepared.Text != small.Text {
		t.Fatal("fitting page was changed", err)
	}
}

// A valid PNG header alone must never become a locked/ready page.
func TestBatchFactoryTextUploadReadsCompleteImage(t *testing.T) {
	var full bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 30, 50))
	img.Set(1, 1, color.White)
	if err := png.Encode(&full, img); err != nil {
		t.Fatal(err)
	}
	good := full.Bytes()
	if mime, err := validateTextImage(good); err != nil || mime != "image/png" {
		t.Fatalf("normal PNG: %s, %v", mime, err)
	}
	partial := good[:len(good)/2]
	if _, _, err := image.DecodeConfig(bytes.NewReader(partial)); err != nil {
		t.Fatal("test needs readable dimensions", err)
	}
	if _, err := validateTextImage(partial); err == nil {
		t.Fatal("partial image accepted")
	}
}
