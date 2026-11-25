package novelcreation

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"strings"
)

// buildEPUB 生成 EPUB 3 包。只依赖标准库，不引入新的运行依赖。
// 章节正文是纯文本，按空行分段后转义写入 XHTML；不猜测 Markdown 结构。
func buildEPUB(title string, chapters []Chapter) ([]byte, error) {
	bookTitle := strings.TrimSpace(title)
	if bookTitle == "" {
		bookTitle = "未命名作品"
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	// mimetype 必须是第一个条目且不压缩。
	header := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return nil, err
	}
	if _, err = writer.Write([]byte("application/epub+zip")); err != nil {
		return nil, err
	}
	writeFile := func(name string, content string) error {
		w, e := zw.Create(name)
		if e != nil {
			return e
		}
		_, e = w.Write([]byte(content))
		return e
	}
	if err = writeFile("META-INF/container.xml", `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/package.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`); err != nil {
		return nil, err
	}
	if err = writeFile("OEBPS/style.css", `body { font-family: serif; line-height: 1.8; margin: 1em; }
h1 { font-size: 1.4em; margin: 1.2em 0 0.8em; }
p { text-indent: 2em; margin: 0.6em 0; }`); err != nil {
		return nil, err
	}
	manifest := strings.Builder{}
	spine := strings.Builder{}
	navItems := strings.Builder{}
	for index, chapter := range chapters {
		name := fmt.Sprintf("chapter-%04d.xhtml", chapter.Number)
		chapterTitle := fmt.Sprintf("第%d章 %s", chapter.Number, strings.TrimSpace(chapter.Title))
		body := strings.Builder{}
		for _, paragraph := range splitParagraphs(chapter.Content) {
			body.WriteString("<p>" + html.EscapeString(paragraph) + "</p>\n")
		}
		page := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="zh" lang="zh">
<head><meta charset="UTF-8"/><title>%s</title><link rel="stylesheet" href="style.css" type="text/css"/></head>
<body><h1>%s</h1>
%s</body>
</html>`, html.EscapeString(chapterTitle), html.EscapeString(chapterTitle), body.String())
		if err = writeFile("OEBPS/"+name, page); err != nil {
			return nil, err
		}
		manifest.WriteString(fmt.Sprintf(`    <item id="chapter-%d" href="%s" media-type="application/xhtml+xml"/>`+"\n", index+1, name))
		spine.WriteString(fmt.Sprintf(`    <itemref idref="chapter-%d"/>`+"\n", index+1))
		navItems.WriteString(fmt.Sprintf(`      <li><a href="%s">%s</a></li>`+"\n", name, html.EscapeString(chapterTitle)))
	}
	nav := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" xml:lang="zh" lang="zh">
<head><meta charset="UTF-8"/><title>%s</title></head>
<body><nav epub:type="toc" id="toc"><h1>目录</h1><ol>
%s    </ol></nav></body>
</html>`, html.EscapeString(bookTitle), navItems.String())
	if err = writeFile("OEBPS/nav.xhtml", nav); err != nil {
		return nil, err
	}
	packageFile := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id" xml:lang="zh">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:language>zh</dc:language>
    <dc:identifier id="book-id">%s</dc:identifier>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="style" href="style.css" media-type="text/css"/>
%s  </manifest>
  <spine>
%s  </spine>
</package>`, html.EscapeString(bookTitle), html.EscapeString("urn:novel:"+fingerprint([]string{bookTitle})[:16]), manifest.String(), spine.String())
	if err = writeFile("OEBPS/package.opf", packageFile); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// splitParagraphs 按空行分段，保留段落内换行合并。
func splitParagraphs(content string) []string {
	body := stripLocalTitle(content)
	out := []string{}
	var paragraph []string
	flush := func() {
		if len(paragraph) > 0 {
			out = append(out, strings.TrimSpace(strings.Join(paragraph, "")))
			paragraph = paragraph[:0]
		}
	}
	for _, line := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		paragraph = append(paragraph, line)
	}
	flush()
	return out
}
