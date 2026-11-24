package batchfactorytext

import (
	"context"
	"strings"
	"testing"
)

func TestHigherTargetRepaginatesWholeTextWithoutLosingParagraphs(t *testing.T) {
	ctx := context.Background()
	text := strings.Repeat("她整理好信件，推开门走向车站。\n第二天清晨，她终于见到了等待多年的亲人。\n", 24)
	layout := DefaultRenderLayout()
	layout.PaginationMode, layout.TargetCharacters, layout.FontSize = "target", 150, 36
	measure := (Renderer{}).MeasureFunc(ctx)
	before, err := Paginate(ctx, PaginationInput{Text: text, Layout: layout}, measure)
	if err != nil {
		t.Fatal(err)
	}
	layout.TargetCharacters = 200
	after, err := Paginate(ctx, PaginationInput{Text: text, Layout: layout}, measure)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) >= len(before) {
		t.Fatalf("150: %d pages, 200: %d pages", len(before), len(after))
	}
	var joined strings.Builder
	for _, page := range after {
		joined.WriteString(page.Text)
	}
	if joined.String() != text {
		t.Fatal("repagination changed or dropped original text")
	}
}
