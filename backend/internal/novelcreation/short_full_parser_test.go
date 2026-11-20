package novelcreation

import (
	"strings"
	"testing"
)

func TestShortFullParserUsesLastNonemptyBlocksAndCompletesOnlyMissing(t *testing.T) {
	raw := "=== SHORT_FICTION_TITLE ===\n《渡口旧信》\n=== SHORT_FICTION_OPENING_HOOK ===\n一封旧信。\n=== CHAPTER 1 TITLE ===\n第1章 归来\n=== CHAPTER 1 CONTENT ===\n旧正文\n=== CHAPTER 2 TITLE ===\n第2章 信封\n=== CHAPTER 2 CONTENT ===\n\n"
	draft, err := parseShortFullDraft(raw, 2)
	if err != nil || draft.Title != "渡口旧信" || draft.Chapters[0].Title != "归来" || len(missingShortFullChapters(draft)) != 1 {
		t.Fatal(draft, err)
	}
	raw += "\n=== CHAPTER 2 CONTENT ===\n第二章补齐。\n=== CHAPTER 1 CONTENT ===\n新的第一章。\n=== CHAPTER 1 CONTENT ===\n\n"
	draft, err = parseShortFullDraft(raw, 2)
	if err != nil || validateShortFullDraft(draft, 2) != nil || draft.Chapters[0].Content != "新的第一章。" || draft.Chapters[1].Content != "第二章补齐。" || draft.RawContent != raw {
		t.Fatal(draft, err)
	}
	if _, err = parseShortFullDraft(raw, 0); err == nil {
		t.Fatal("invalid count accepted")
	}
}

func TestShortFullParserDuplicateTitleFallbackAndSanitize(t *testing.T) {
	raw := "=== SHORT_FICTION_TITLE ===\n# 《名字》\n=== CHAPTER 1 TITLE ===\n第1章 旧信\n=== CHAPTER 1 TITLE ===\n```markdown\n正文。\n=== AUX_MARKER ===\n尾声。\n```\n=== SHORT_FICTION_OTHER ===\n不能串进正文"
	draft, err := parseShortFullDraft(raw, 1)
	if err != nil || draft.Title != "名字" || draft.Chapters[0].Title != "旧信" || draft.Chapters[0].Content != "正文。\n\n尾声。" {
		t.Fatal(draft, err)
	}
	if shortFullCharacterCount("中 A\t\n😀\u00a0\x1c") != 3 {
		t.Fatal("native count is all non-whitespace codepoints")
	}
}

func TestShortFullParserTitlesUseFirstBlockWhileBodyUsesLastNonempty(t *testing.T) {
	raw := "=== short_fiction_title ===\n第一书名\n=== SHORT_FICTION_TITLE ===\n错误书名\n=== chapter 1 title ===\n第1章 正确章名\n=== CHAPTER 1 TITLE ===\n错误章名\n=== CHAPTER 1 CONTENT ===\n旧正文\n=== chapter 1 content ===\n新正文\n=== CHAPTER 1 CONTENT ===\n\n"
	draft, err := parseShortFullDraft(raw, 1)
	if err != nil || draft.Title != "第一书名" || draft.Chapters[0].Title != "正确章名" || draft.Chapters[0].Content != "新正文" {
		t.Fatal(draft, err)
	}
	if shortFullFirstBlock("=== INTRO ===\n\n=== INTRO ===\n第二块", "INTRO") != "" {
		t.Fatal("first empty block must not be skipped")
	}
}

func TestShortFullParserPreservesOriginalLooseMarkdownFallback(t *testing.T) {
	draft, err := parseShortFullDraft("# 书名\n## 第1章 开始\n第一章内容\n## 第2章 结束\n第二章内容", 2)
	if err != nil || draft.Chapters[0].Content != "第一章内容" || draft.Chapters[1].Content != "第一章内容" {
		t.Fatal("native optional numbered group was silently changed", draft, err)
	}
	// This intentionally records the recovered fallback defect; production prompts
	// request explicit CHAPTER blocks. Do not call it exact numbered Markdown parsing.
}

func TestShortFullParserValidationRenderAndTokenEstimate(t *testing.T) {
	draft, err := parseShortFullDraft("", 2)
	if err != nil || draft.Title != "未命名短篇" || len(missingShortFullChapters(draft)) != 2 || validateShortFullDraft(draft, 2) == nil {
		t.Fatal(draft, err)
	}
	draft = shortFullDraft{Title: "渡口", OpeningHook: "旧信", Chapters: []ShortChapter{{Number: 1, Title: "归来", Content: "正文"}}}
	if validateShortFullDraft(draft, 1) != nil || !strings.Contains(renderShortFullDraft(draft), "## 第1章 归来\n\n正文") {
		t.Fatal(renderShortFullDraft(draft))
	}
	if estimateShortFullTokens(1, 1000) != 12288 || estimateShortFullTokens(2, 5000) != 26096 {
		t.Fatal("token formula differs")
	}
}

func TestShortFullPackageAliasesAndExactBulletCleanup(t *testing.T) {
	raw := "=== SHORT_FICTION_TITLE ===\n# 《包名》\n=== INTRO ===\n第一简介\n=== INTRO ===\n不能覆盖\n=== SELLING_POINTS ===\n- 卖点一\n* 卖点二\n--保留一个横线\n1. 保留编号\n• 保留圆点\n-\n=== COVER_PROMPT ===\n封面"
	got := parseShortFullPackage(raw, "后备标题")
	if got.Title != "包名" || got.Intro != "第一简介" || got.CoverPrompt != "封面" || strings.Join(got.SellingPoints, "|") != "卖点一|卖点二|-保留一个横线|1. 保留编号|• 保留圆点" {
		t.Fatal(got)
	}
	if got := parseShortFullPackage("", "《后备标题》"); got.Title != "后备标题" || got.Intro != "" || len(got.SellingPoints) != 0 {
		t.Fatal(got)
	}
}
