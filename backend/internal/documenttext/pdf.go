package documenttext

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

// The gate bounds compilation, linear memory and output allocations across all
// services in this process. Each pool inherits the parse deadline, so expiry
// closes actual WASM execution (including initialization), not just its caller.
var pdfGate = make(chan struct{}, 1)
var pdfCompilationCache = wazero.NewCompilationCache()

func extractPDF(ctx context.Context, data []byte) (out Result, err error) {
	select {
	case pdfGate <- struct{}{}:
	case <-ctx.Done():
		return out, ctx.Err()
	}
	defer func() { <-pdfGate }()
	pool, err := webassembly.Init(webassembly.Config{Context: ctx, MaxTotal: 1, FSConfig: wazero.NewFSConfig(), Stdout: io.Discard, Stderr: io.Discard, RuntimeConfig: wazero.NewRuntimeConfig().WithCompilationCache(pdfCompilationCache).WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesExceptionHandling).WithMemoryLimitPages(4096).WithCloseOnContextDone(true)})
	if err != nil {
		return out, fmt.Errorf("PDF解析器无法启动或资源超限: %w", err)
	}
	defer pool.Close()
	instance, err := pool.GetInstanceWithContext(ctx)
	if err != nil {
		return out, err
	}
	defer instance.Close()
	doc, err := instance.OpenDocument(&requests.OpenDocument{File: &data})
	if err != nil {
		last, _ := instance.FPDF_GetLastError(&requests.FPDF_GetLastError{})
		if last != nil && (last.Error == responses.FPDF_GetLastErrorErrorPassword || last.Error == responses.FPDF_GetLastErrorErrorSecurity) {
			return out, errors.New("不支持加密PDF，请先导出无密码版本")
		}
		return out, fmt.Errorf("PDF损坏或资源超限: %w", err)
	}
	defer instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	security, err := instance.FPDF_GetSecurityHandlerRevision(&requests.FPDF_GetSecurityHandlerRevision{Document: doc.Document})
	if err != nil {
		return out, err
	}
	if security.SecurityHandlerRevision != -1 {
		return out, errors.New("不支持加密PDF，请先导出无密码版本")
	}
	pages, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return out, err
	}
	if pages.PageCount > 2000 {
		return out, errors.New("PDF超过2000页，请拆分")
	}
	var b strings.Builder
	total := 0
	for i := 0; i < pages.PageCount; i++ {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		page, e := instance.FPDF_LoadPage(&requests.FPDF_LoadPage{Document: doc.Document, Index: i})
		if e != nil {
			return out, e
		}
		text, e := pdfPageText(ctx, instance, requests.Page{ByReference: &page.Page}, &total)
		instance.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: page.Page})
		if e != nil {
			return out, e
		}
		if strings.TrimSpace(text) == "" {
			out.EmptyPages = append(out.EmptyPages, i+1)
		}
		b.WriteString(text)
		b.WriteByte('\n')
	}
	out.Text = b.String()
	out.Encoding = "PDF Unicode"
	return out, nil
}
func pdfPageText(ctx context.Context, p pdfium.Pdfium, page requests.Page, total *int) (string, error) {
	t, err := p.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: page})
	if err != nil {
		return "", err
	}
	defer p.FPDFText_ClosePage(&requests.FPDFText_ClosePage{TextPage: t.TextPage})
	n, err := p.FPDFText_CountChars(&requests.FPDFText_CountChars{TextPage: t.TextPage})
	if err != nil {
		return "", err
	}
	if n.Count < 0 || n.Count > 100000 || *total+n.Count > MaxCharacters {
		return "", errors.New("PDF正文超过资源限制，请拆分（单页10万、总计200万字元）")
	}
	*total += n.Count
	var b strings.Builder
	for start := 0; start < n.Count; start += 4096 {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		chunk, err := p.FPDFText_GetText(&requests.FPDFText_GetText{TextPage: t.TextPage, StartIndex: start, Count: min(4096, n.Count-start)})
		if err != nil {
			return "", err
		}
		b.WriteString(chunk.Text)
	}
	return b.String(), nil
}
