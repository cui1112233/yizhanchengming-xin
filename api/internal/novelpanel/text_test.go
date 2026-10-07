package novelpanel

import "testing"

func TestProcessTextNormalizesWithoutReordering(t *testing.T) {
	input := "  第一段。\r\n\r\n\u3000第二段！  \r\n\r\n\r\n第三段。\r\n"
	got := ProcessText(input, TextProcessingSettings{
		TrimLineWhitespace:       true,
		CollapseBlankLines:       true,
		NormalizeFullWidthSpaces: true,
	})
	want := "第一段。\n\n第二段！\n\n第三段。"
	if got != want {
		t.Fatalf("processed text mismatch\nwant: %q\n got: %q", want, got)
	}
	lines := SourceLines(got)
	if len(lines) != 3 || lines[0].Index != 1 || lines[1].Text != "第二段！" || lines[2].RawLine != 5 {
		t.Fatalf("unexpected source lines: %#v", lines)
	}
}
