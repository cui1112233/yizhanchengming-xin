package batchfactorytext

import (
	"context"
	"errors"
	"math/rand"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

func runeCapacity(n int) MeasureFunc {
	return func(text string, _ LayoutConfig) (bool, float64, error) {
		used := utf8.RuneCountInString(text)
		return used <= n, float64(used), nil
	}
}

func TestPaginateParagraphsAndSentences(t *testing.T) {
	for _, c := range []struct {
		name, text       string
		target, capacity int
		want             []string
	}{
		{"完整段落优先于剩余空间", "第一。\n甲。乙。", 150, 6, []string{"第一。\n", "甲。乙。"}},
		{"引号复合标点和省略号", "“走吧！？”她沉默……然后点头。", 150, 6, []string{"“走吧！？”", "她沉默……", "然后点头。"}},
		{"句号后新开半角引号", `过了一会儿。"别走。"`, 150, 6, []string{"过了一会儿。", `"别走。"`}},
		{"相邻半角对白", `"行。""别走。"`, 150, 5, []string{`"行。"`, `"别走。"`}},
		{"保留英文缩写撇号", `Don't.'Go!'`, 150, 6, []string{"Don't.", "'Go!'"}},
		{"英文所有格撇号", `James' cat.'Go!'`, 150, 11, []string{"James' cat.", "'Go!'"}},
		{"中英文引号嵌套", `“她说'好！'”"走。"`, 150, 8, []string{"“她说'好！'”", `"走。"`}},
		{"未闭合对白", `说完了。"请等。`, 150, 4, []string{"说完了。", `"请等。`}},
		{"从上一段延续的对白", "\"走。\n再见。\"她回家了。", 150, 5, []string{"\"走。\n", "再见。\"", "她回家了。"}},
		{"英语句号和小数", "It is 3.14. OK! Go?", 150, 12, []string{"It is 3.14. ", "OK! Go?"}},
		{"分号和多个右引号", "「他说『来！』」好；再见;", 150, 8, []string{"「他说『来！』」", "好；再见;"}},
		{"无结尾标点的尾段", "他答应了。后来", 150, 5, []string{"他答应了。", "后来"}},
		{"句末换行移至下一页", "你来了。\n好。", 150, 4, []string{"你来了。", "\n好。"}},
		{"页间保留长空行", "你好。\n\n\n好的。", 150, 5, []string{"你好。\n\n", "\n好的。"}},
		{"刚好占满一页的英文句子", "Good. Yes.", 150, 5, []string{"Good.", " Yes."}},
		{"保留空行缩进和表情", "\r\n 甲🙂。\r\n\r\n　乙。 \r", 150, 7, []string{"\n 甲🙂。\n\n", "　乙。 \n"}},
		{"超过目标仍保留放得下的整段", "甲。乙。丙。丁。", 3, 9, []string{"甲。乙。丙。丁。"}},
		{"多个段落按柔性目标分组", "一。\n二。\n三。\n四。", 6, 200, []string{"一。\n二。\n", "三。\n四。"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			pages, err := Paginate(context.Background(), PaginationInput{Text: c.text, Layout: LayoutConfig{TargetCharacters: c.target}}, runeCapacity(c.capacity))
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, p := range pages {
				got = append(got, p.Text)
			}
			if !slices.Equal(got, c.want) {
				t.Fatalf("pages = %q, want %q", got, c.want)
			}
			if err := validateDrafts(c.text, pages); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPaginateUsesDefaultDensityAndActualMeasurement(t *testing.T) {
	p := strings.Repeat("花", 73) + "。"
	text := p + "\n" + p + "\n" + p + "\n" + p
	layout := LayoutConfig{FontID: "sample", FontSize: 40, MaxHeightRatio: .67}
	measure := func(text string, got LayoutConfig) (bool, float64, error) {
		if got.FontSize != 40 || got.FontID != "sample" || got.MaxHeightRatio != .67 || got.TargetCharacters != 150 {
			t.Fatalf("unexpected layout %#v", got)
		}
		return true, 12, nil
	}
	got, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: layout}, measure)
	if err != nil || len(got) != 2 || got[0].Text != p+"\n"+p+"\n" {
		t.Fatalf("pages = %+v err=%v", got, err)
	}

	measure = func(text string, _ LayoutConfig) (bool, float64, error) {
		width := 0
		for _, r := range text {
			if r == '宽' {
				width += 4
			} else {
				width++
			}
		}
		return width <= 5, float64(width), nil
	}
	got, err = Paginate(context.Background(), PaginationInput{Text: "宽。窄。宽。"}, measure)
	if err != nil || len(got) != 3 {
		t.Fatalf("measured pages = %#v, %v", got, err)
	}
}

func TestPaginateOverflowReportsWholeSentenceAndNoPartialResult(t *testing.T) {
	// A stray trailing quote has no following words and must not become a page on its own.
	pages, err := Paginate(context.Background(), PaginationInput{Text: `她走了。"`}, runeCapacity(4))
	var dangling *SentenceOverflowError
	if pages != nil || !errors.As(err, &dangling) || dangling.EndOffset != 5 {
		t.Fatalf("dangling quote: %+v, %v", pages, err)
	}
	for _, sentence := range []string{"一整句没有中间句号。", "Mr. Black came.", "数值3.14159。", "mail.example.com。", "E.g. Bob came.", "J. Doe ran."} {
		in := PaginationInput{Text: "🙂。\r\n" + sentence}
		pages, err := Paginate(context.Background(), in, runeCapacity(6))
		var overflow *SentenceOverflowError
		if pages != nil || !errors.As(err, &overflow) || overflow.Page != 2 || overflow.StartOffset != 3 || overflow.EndOffset != 3+utf8.RuneCountInString(sentence) {
			t.Fatalf("%q: pages=%v error=%#v", sentence, pages, err)
		}
	}
}

func TestPaginateInvalidInputMeasurementAndCancellation(t *testing.T) {
	for _, text := range []string{"", "\t\r\n  ", string([]byte{0xff})} {
		if _, err := Paginate(context.Background(), PaginationInput{Text: text}, runeCapacity(20)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q: %v", text, err)
		}
	}
	if _, err := Paginate(context.Background(), PaginationInput{Text: "正文"}, nil); !errors.Is(err, ErrInvalid) {
		t.Fatalf("nil measure: %v", err)
	}
	failed := errors.New("字体不可用")
	pages, err := Paginate(context.Background(), PaginationInput{Text: "甲。\n乙。"}, func(string, LayoutConfig) (bool, float64, error) { return true, 0, failed })
	if pages != nil || !errors.Is(err, failed) {
		t.Fatalf("%+v, %v", pages, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	pages, err = Paginate(ctx, PaginationInput{Text: "正文"}, func(string, LayoutConfig) (bool, float64, error) { cancel(); return true, 1, nil })
	if pages != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel returned %v, %v", pages, err)
	}
}

func TestPaginateGeneratedTextPreservesLegalBoundariesAndFit(t *testing.T) {
	rng := rand.New(rand.NewSource(504))
	for sample := 0; sample < 250; sample++ {
		var b strings.Builder
		legal := map[int]bool{}
		offset := 0
		for i := 0; i < 30; i++ {
			s := []string{"甲。", "【好！】", "🙂来了？", "Go!", "等……", "3.14。"}[rng.Intn(6)]
			offset += utf8.RuneCountInString(s)
			legal[offset] = true // The trailing separator may straddle a page break.
			switch rng.Intn(4) {
			case 0:
				s += "\r\n"
				offset++
			case 1:
				s += "\r\n\r\n"
				legal[offset+1] = true
				offset += 2
			}
			b.WriteString(s)
			legal[offset] = true
		}
		text := b.String()
		capacity := 10 + rng.Intn(31)
		pages, err := Paginate(context.Background(), PaginationInput{Text: text, Layout: LayoutConfig{TargetCharacters: 4 + rng.Intn(40)}}, runeCapacity(capacity))
		if err != nil {
			t.Fatalf("%d: %v", sample, err)
		}
		var joined strings.Builder
		for _, p := range pages {
			if !legal[p.EndOffset] || p.StartOffset != utf8.RuneCountInString(joined.String()) || p.Text == "" || utf8.RuneCountInString(p.Text) > capacity {
				t.Fatalf("%d: invalid page %+v", sample, p)
			}
			joined.WriteString(p.Text)
		}
		if joined.String() != NormalizeLineEndings(text) || strings.Contains(joined.String(), "长段落续页") {
			t.Fatalf("%d: wrong content %q", sample, joined.String())
		}
	}
}
