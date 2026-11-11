package httpapi

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
)

func buildDocxForTest(t *testing.T, parts map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range parts {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

func TestDocxPlainTextNormal(t *testing.T) {
	data := buildDocxForTest(t, map[string]string{
		"word/document.xml": `<?xml version="1.0"?><w:document xmlns:w="w"><w:body><w:p><w:r><w:t>第一段</w:t></w:r></w:p><w:p><w:r><w:t>第二段</w:t></w:r></w:p></w:body></w:document>`,
	})
	text, err := docxPlainText(data)
	if err != nil {
		t.Fatalf("docxPlainText: %v", err)
	}
	if !strings.Contains(text, "第一段") || !strings.Contains(text, "第二段") {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestOpenDocxRejectsTooManyEntries(t *testing.T) {
	previous := docxSafeLimits.MaxEntries
	docxSafeLimits.MaxEntries = 2
	defer func() { docxSafeLimits.MaxEntries = previous }()
	data := buildDocxForTest(t, map[string]string{
		"word/document.xml": "<doc/>",
		"word/styles.xml":   "<styles/>",
		"word/settings.xml": "<settings/>",
	})
	if _, err := openDocxDocumentXML(data); !errors.Is(err, errDocxTooManyEntries) {
		t.Fatalf("err = %v, want errDocxTooManyEntries", err)
	}
}

func TestOpenDocxRejectsOversizeEntry(t *testing.T) {
	previous := docxSafeLimits.MaxEntryBytes
	docxSafeLimits.MaxEntryBytes = 16
	defer func() { docxSafeLimits.MaxEntryBytes = previous }()
	data := buildDocxForTest(t, map[string]string{
		"word/document.xml": strings.Repeat("<p>超长正文</p>", 20),
	})
	if _, err := openDocxDocumentXML(data); !errors.Is(err, errDocxTooLarge) {
		t.Fatalf("err = %v, want errDocxTooLarge", err)
	}
}

func TestOpenDocxRejectsOversizeTotal(t *testing.T) {
	previous := docxSafeLimits.MaxTotalBytes
	docxSafeLimits.MaxTotalBytes = 32
	defer func() { docxSafeLimits.MaxTotalBytes = previous }()
	data := buildDocxForTest(t, map[string]string{
		"word/document.xml": strings.Repeat("a", 20),
		"word/theme.xml":    strings.Repeat("b", 20),
	})
	if _, err := openDocxDocumentXML(data); !errors.Is(err, errDocxTooLarge) {
		t.Fatalf("err = %v, want errDocxTooLarge", err)
	}
}

func TestOpenDocxMissingDocument(t *testing.T) {
	data := buildDocxForTest(t, map[string]string{"word/styles.xml": "<styles/>"})
	if _, err := openDocxDocumentXML(data); !errors.Is(err, errDocxMissingDoc) {
		t.Fatalf("err = %v, want errDocxMissingDoc", err)
	}
}
