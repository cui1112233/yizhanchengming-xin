package httpapi

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestParseSmartShuihuoCandidatesAcceptsCandidatesAndSegments(t *testing.T) {
	for _, raw := range []string{
		`{"candidates":[{"text":"第一段","speaker":"旁白"}]}`,
		"```json\n{\"segments\":[{\"sourceText\":\"第二段\",\"subtitleText\":\"第二段\"}]}\n```",
	} {
		got, err := parseSmartShuihuoCandidates(raw, "book-1")
		if err != nil {
			t.Fatalf("parse candidates: %v", err)
		}
		if len(got) != 1 || got[0]["text"] == "" || got[0]["id"] == "" {
			t.Fatalf("unexpected candidates: %#v", got)
		}
	}
}

func TestParseSmartShuihuoCandidatesRejectsInvalidOutput(t *testing.T) {
	if _, err := parseSmartShuihuoCandidates("not-json", "book-1"); err == nil {
		t.Fatal("expected invalid model output to fail")
	}
}

func TestFixedShuihuoCandidatesGroupsLines(t *testing.T) {
	got := fixedShuihuoCandidates("book-1", "一\n二\n三", 2)
	if len(got) != 2 || got[0]["text"] != "一\n二" || got[1]["text"] != "三" {
		t.Fatalf("unexpected fixed candidates: %#v", got)
	}
}

func TestParseShuihuoPromptCandidates(t *testing.T) {
	got, err := parseShuihuoPromptCandidates("```json\n{\"candidates\":[{\"segmentId\":\"s1\",\"prompt\":\"夜景\"}]}\n```")
	if err != nil || len(got) != 1 || got[0]["segmentId"] != "s1" {
		t.Fatalf("unexpected prompt candidates: %#v, err=%v", got, err)
	}
}

func TestShuihuoImageGenerationEndpointAndSize(t *testing.T) {
	got, err := shuihuoImageGenerationEndpoint("https://images.example/v1")
	if err != nil || got != "https://images.example/v1/images/generations" {
		t.Fatalf("endpoint=%q err=%v", got, err)
	}
	if shuihuoImageSize("9:16") != "1024x1792" || shuihuoImageSize("1:1") != "1024x1024" {
		t.Fatalf("unexpected image sizes")
	}
}

func TestExtractShuihuoDocxText(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entry, err := archive.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte(`<w:document xmlns:w="x"><w:body><w:p><w:r><w:t>第一段</w:t></w:r></w:p><w:p><w:r><w:t>第二段</w:t></w:r></w:p></w:body></w:document>`))
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := extractShuihuoDocxText(buffer.Bytes())
	if err != nil || got != "第一段\n第二段" {
		t.Fatalf("text=%q err=%v", got, err)
	}
}
