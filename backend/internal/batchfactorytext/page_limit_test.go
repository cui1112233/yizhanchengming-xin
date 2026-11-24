package batchfactorytext

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRepeatedBlankLinesDoNotConsumeReadingSpace(t *testing.T) {
	renderer := Renderer{}
	layout := DefaultRenderLayout()
	compact := "第一段完整正文。\n第二段完整正文。\n第三段完整正文。"
	spaced := strings.ReplaceAll(compact, "\n", "\n\n\n\n")
	_, a, err := renderer.MeasureFunc(context.Background())(compact, layout)
	if err != nil {
		t.Fatal(err)
	}
	_, b, err := renderer.MeasureFunc(context.Background())(spaced, layout)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("blank separators inflated height: %v != %v", a, b)
	}
	pages, err := Paginate(context.Background(), PaginationInput{Text: spaced, Layout: layout}, renderer.MeasureFunc(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	var joined strings.Builder
	for _, p := range pages {
		joined.WriteString(p.Text)
	}
	if joined.String() != spaced {
		t.Fatal("original source was altered")
	}
}

func TestTextPaginationEnforcesConfiguredPageLimitWithoutTruncation(t *testing.T) {
	text := strings.Repeat("完整。\n", 36)
	_, err := Paginate(context.Background(), PaginationInput{Text: text}, runeCapacity(4))
	var limit *PageLimitError
	if !errors.As(err, &limit) || limit.Limit != 35 {
		t.Fatalf("default limit not enforced: %v", err)
	}
	pages, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: LayoutConfig{MaxPages: 40}}, runeCapacity(4))
	if err != nil || len(pages) != 36 {
		t.Fatalf("configured limit ignored: %d %v", len(pages), err)
	}
	var joined strings.Builder
	for _, p := range pages {
		joined.WriteString(p.Text)
	}
	if joined.String() != text {
		t.Fatal("content lost after raising limit")
	}
}

func TestReadingPaginationBalancesTinyFinalPageAtParagraphBoundary(t *testing.T) {
	for _, mode := range []string{"auto", "target"} {
		text := strings.Repeat(strings.Repeat("字", 30)+"。\n", 3) + "SUV车窗降下"
		pages, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: LayoutConfig{PaginationMode: mode}}, runeCapacity(96))
		if err != nil || len(pages) != 2 {
			t.Fatalf("pages=%d error=%v", len(pages), err)
		}
		if len([]rune(strings.TrimSpace(pages[1].Text))) < 25 {
			t.Fatalf("tiny final page was not balanced: %q", pages[1].Text)
		}
		if pages[0].Text+pages[1].Text != text || !strings.HasSuffix(pages[0].Text, "\n") {
			t.Fatal("balancing lost text or split a paragraph")
		}
		for _, page := range pages {
			if len([]rune(page.Text)) > 96 {
				t.Fatal("balanced page overflowed")
			}
		}
	}
}
