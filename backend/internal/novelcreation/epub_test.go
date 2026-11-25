package novelcreation

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func epubFixtureChapters() []Chapter {
	return []Chapter{
		{Number: 1, Title: "春雨", Content: "# 春雨\n\n雨水敲着窗，林澜没有开灯。\n\n信纸上只有一枚褪色邮戳。"},
		{Number: 2, Title: "旧信 <特刊>", Content: "邮戳来自城南旧邮局 & 十年前就停业了。"},
	}
}

func TestBuildEPUBStructure(t *testing.T) {
	binary, e := buildEPUB("测试书", epubFixtureChapters())
	if e != nil {
		t.Fatal(e)
	}
	reader, e := zip.NewReader(bytes.NewReader(binary), int64(len(binary)))
	if e != nil {
		t.Fatalf("not a zip: %v", e)
	}
	if len(reader.File) == 0 || reader.File[0].Name != "mimetype" {
		t.Fatalf("mimetype must be first: %+v", reader.File)
	}
	if reader.File[0].Method != zip.Store {
		t.Fatalf("mimetype must be stored uncompressed: %v", reader.File[0].Method)
	}
	names := map[string]bool{}
	for _, file := range reader.File {
		names[file.Name] = true
	}
	for _, want := range []string{"META-INF/container.xml", "OEBPS/package.opf", "OEBPS/nav.xhtml", "OEBPS/style.css", "OEBPS/chapter-0001.xhtml", "OEBPS/chapter-0002.xhtml"} {
		if !names[want] {
			t.Fatalf("missing %s in %v", want, names)
		}
	}
	read := func(name string) string {
		file, e := reader.Open(name)
		if e != nil {
			t.Fatal(e)
		}
		defer file.Close()
		b, e := io.ReadAll(file)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	if got := read("mimetype"); got != "application/epub+zip" {
		t.Fatalf("mimetype=%q", got)
	}
	chapter := read("OEBPS/chapter-0002.xhtml")
	if !strings.Contains(chapter, "旧信 &lt;特刊&gt;") || !strings.Contains(chapter, "&amp;") {
		t.Fatalf("xhtml not escaped: %s", chapter)
	}
	nav := read("OEBPS/nav.xhtml")
	if !strings.Contains(nav, "chapter-0001.xhtml") || !strings.Contains(nav, "春雨") {
		t.Fatalf("nav missing chapter: %s", nav)
	}
	if !strings.Contains(read("OEBPS/package.opf"), `<itemref idref="chapter-2"/>`) {
		t.Fatal("spine missing second chapter")
	}
}

func TestBuildEPUBParagraphSplit(t *testing.T) {
	paragraphs := splitParagraphs("# 标题\n\n第一段\n继续第一段\n\n第二段")
	if len(paragraphs) != 2 || paragraphs[0] != "第一段继续第一段" || paragraphs[1] != "第二段" {
		t.Fatalf("paragraphs=%v", paragraphs)
	}
}

func TestExportProjectSupportsEPUBAndRejectsUnknown(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, NewMemoryRuntime(), nil)
	ctx := context.Background()
	project, e := service.CreateProject(ctx, "alice", CreateProjectInput{Title: "导出书", Genre: "悬疑", ModelID: "own-model"})
	if e != nil {
		t.Fatal(e)
	}
	if e = store.SaveChapterBundle(ctx, "alice", ChapterBundle{ProjectID: project.ID, Chapter: Chapter{ID: "chapter-1", ProjectID: project.ID, Number: 1, Title: "春雨", Content: "雨水敲着窗。", Version: 1}}); e != nil {
		t.Fatal(e)
	}
	doc, e := service.ExportProject(ctx, "alice", project.ID, "epub")
	if e != nil {
		t.Fatal(e)
	}
	if doc.ContentType != "application/epub+zip" || !strings.HasSuffix(doc.Filename, ".epub") || len(doc.Binary) == 0 {
		t.Fatalf("epub doc=%+v", doc)
	}
	if _, e = service.ExportProject(ctx, "alice", project.ID, "pdf"); e == nil || !IsCode(e, ErrInputRequired) {
		t.Fatalf("unknown format must be rejected: %v", e)
	}
}
