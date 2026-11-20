package novelcreation

import (
	"strings"
	"testing"
)

func TestShortFullTemplatesRenderNativeFieldsAndMissingOnly(t *testing.T) {
	input := ShortFullPromptInput{Direction: "方向{{not_a_template}}", OutlineMarkdown: "  方案  ", DraftMarkdown: "  原稿  ", DraftTitle: "题名", Review: "  审稿  ", ChapterCount: 3, CharsPerChapter: 2000}
	for name, template := range map[string]string{"writer": ShortFullWriterUserTemplate, "continuation": ShortFullContinuationUserTemplate, "review": ShortFullReviewUserTemplate, "revision": ShortFullRevisionUserTemplate, "package": ShortFullPackageUserTemplate} {
		t.Run(name, func(t *testing.T) {
			text, err := RenderShortFullPrompt(template, ShortFullChapterBlockTemplate, input)
			if err != nil || text == "" {
				t.Fatal(text, err)
			}
			for _, token := range fullPlaceholder.FindAllString(text, -1) {
				if token != "{{not_a_template}}" {
					t.Fatalf("unrendered field %s", token)
				}
			}
		})
	}
	input.MissingChapters = []int{2, 3}
	text, err := RenderShortFullPrompt(ShortFullContinuationUserTemplate, ShortFullChapterBlockTemplate, input)
	if err != nil || strings.Contains(text, "=== CHAPTER 1 CONTENT ===") || !strings.Contains(text, "=== CHAPTER 2 CONTENT ===") || !strings.Contains(text, "  原稿  ") {
		t.Fatal(text, err)
	}
	text, err = RenderShortFullPrompt("{{outline_markdown}}|{{outline_markdown_trimmed}}|{{draft_markdown_trimmed}}|{{review_trimmed}}", ShortFullChapterBlockTemplate, input)
	if err != nil || text != "  方案  |方案|原稿|审稿" {
		t.Fatal(text, err)
	}
	input.Review = "\x1c审稿\x1f"
	text, err = RenderShortFullPrompt("{{review_trimmed}}", ShortFullChapterBlockTemplate, input)
	if err != nil || text != "审稿" {
		t.Fatal("Python strip control-whitespace mismatch", text, err)
	}
}

func TestShortFullTemplatesRejectUnknownFieldsAndBadChapterNumbers(t *testing.T) {
	input := ShortFullPromptInput{ChapterCount: 3, CharsPerChapter: 1000}
	for _, template := range []string{"{{unknown_field}}", "{{direction}}\n{{forgotten}}"} {
		if _, err := RenderShortFullPrompt(template, ShortFullChapterBlockTemplate, input); err == nil {
			t.Fatal("unknown admin field reached model", template)
		}
	}
	for _, missing := range [][]int{{0}, {4}, {1, 1}} {
		input.MissingChapters = missing
		if _, err := RenderShortFullPrompt(ShortFullContinuationUserTemplate, ShortFullChapterBlockTemplate, input); err == nil {
			t.Fatal("invalid missing chapter accepted", missing)
		}
	}
	input.MissingChapters = nil
	if _, err := RenderShortFullPrompt("{{chapter_blocks}}", "{{unknown_chapter}}", input); err == nil {
		t.Fatal("invalid chapter template accepted")
	}
}
