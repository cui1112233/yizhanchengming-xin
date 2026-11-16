package documenttext

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func smallPDF(text string) []byte {
	return pdfWithStream("BT /F1 16 Tf 40 700 Td (" + text + ") Tj ET")
}
func pdfWithStream(stream string) []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>", fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream)}
	return pdfObjects(objects)
}
func pdfObjects(objects []string) []byte {
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, o := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	x := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, o := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), x)
	return b.Bytes()
}

func TestPDFPageBudget(t *testing.T) {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", ""}
	var kids strings.Builder
	for i := 0; i < 2001; i++ {
		fmt.Fprintf(&kids, "%d 0 R ", i+3)
		objects = append(objects, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>")
	}
	objects[1] = fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count 2001 >>", kids.String())
	p := NewParser()
	if _, e := p.Parse(context.Background(), "pages.pdf", pdfObjects(objects)); e == nil || !strings.Contains(e.Error(), "2000页") {
		t.Fatalf("page budget: %v", e)
	}
}

func TestUnicodeEncryptedScannedAndActualTimeout(t *testing.T) {
	p := NewParser()
	data, e := os.ReadFile("testdata/unicode.pdf")
	if e != nil {
		t.Fatal(e)
	}
	r, e := p.Parse(context.Background(), "novel.pdf", data)
	if e != nil || !strings.Contains(r.Text, "第一章 初见") || !strings.Contains(r.Text, "我们再次相遇") {
		t.Fatalf("unicode %q %v", r.Text, e)
	}
	for _, name := range []string{"encrypted.pdf", "encrypted-empty-user.pdf", "scanned.pdf"} {
		b, e := os.ReadFile("testdata/" + name)
		if e != nil {
			t.Fatal(e)
		}
		_, e = p.Parse(context.Background(), name, b)
		if e == nil {
			t.Fatalf("accepted %s", name)
		}
		if strings.HasPrefix(name, "encrypted") && !strings.Contains(e.Error(), "加密") {
			t.Fatalf("%s: %v", name, e)
		}
	}
	p.Timeout = 10 * time.Millisecond
	start := time.Now()
	_, e = p.Parse(context.Background(), "timeout.pdf", data)
	if e == nil || !strings.Contains(e.Error(), "超时") || time.Since(start) > 2*time.Second {
		t.Fatalf("actual cancellation: elapsed=%s err=%v", time.Since(start), e)
	}
	// The timed-out runtime was actually closed and the process-wide slot freed.
	p.Timeout = 30 * time.Second
	if _, e = p.Parse(context.Background(), "after.pdf", data); e != nil {
		t.Fatal(e)
	}
}
func TestRealBoundedDocuments(t *testing.T) {
	p := NewParser()
	defer p.Close()
	var doc bytes.Buffer
	z := zip.NewWriter(&doc)
	f, _ := z.Create("word/document.xml")
	f.Write([]byte(`<w:document xmlns:w="w"><w:body><w:p><w:r><w:t>第一章 故事</w:t></w:r></w:p><w:p><w:r><w:t>你好世界。</w:t></w:r></w:p></w:body></w:document>`))
	z.Close()
	for _, c := range []struct {
		name string
		data []byte
		want string
	}{{"a.txt", []byte("\xef\xbb\xbf你好世界"), "你好世界"}, {"a.docx", doc.Bytes(), "你好世界"}, {"a.pdf", smallPDF("Hello novel"), "Hello novel"}} {
		t.Run(c.name, func(t *testing.T) {
			r, e := p.Parse(context.Background(), c.name, c.data)
			if e != nil || !strings.Contains(r.Text, c.want) {
				t.Fatalf("%+v %v", r, e)
			}
		})
	}
	for _, c := range []struct {
		name string
		data []byte
	}{{"bad.pdf", []byte("%PDF-broken")}, {"blank.pdf", smallPDF("")}, {"bad.txt", []byte{0xff}}, {"big.txt", make([]byte, (20<<20)+1)}} {
		if _, e := p.Parse(context.Background(), c.name, c.data); e == nil {
			t.Fatalf("accepted %s", c.name)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if _, e := p.Parse(ctx, "a.pdf", smallPDF("Hi")); e == nil || time.Since(start) > time.Second {
		t.Fatalf("cancellation %v", e)
	}
}

func TestDocumentOutputLimitRejectsRatherThanTruncates(t *testing.T) {
	p := NewParser()
	if _, e := p.Parse(context.Background(), "big.txt", []byte(strings.Repeat("字", MaxCharacters+1))); e == nil {
		t.Fatal("oversize text silently truncated")
	}
	stream := "BT /F1 1 Tf 40 700 Td " + strings.Repeat("("+strings.Repeat("A", 20000)+") Tj ", 6) + "ET"
	if _, e := p.Parse(context.Background(), "big.pdf", pdfWithStream(stream)); e == nil || !strings.Contains(e.Error(), "限制") {
		t.Fatalf("PDF page character budget not enforced: %v", e)
	}
}
