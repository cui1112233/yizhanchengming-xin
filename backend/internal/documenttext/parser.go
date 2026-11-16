// Package documenttext extracts bounded plain text without host filesystem access.
package documenttext

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxFileBytes = 20 << 20
const MaxCharacters = 2_000_000

type Result struct {
	Text       string `json:"text"`
	Encoding   string `json:"encoding"`
	EmptyPages []int  `json:"emptyPages,omitempty"`
}
type Parser struct{ Timeout time.Duration }

func NewParser() *Parser     { return &Parser{Timeout: 30 * time.Second} }
func (*Parser) Close() error { return nil }
func (p *Parser) Parse(ctx context.Context, name string, data []byte) (out Result, err error) {
	if len(data) > MaxFileBytes {
		return out, errors.New("文件超过20MiB，请拆分后重试")
	}
	if err = ctx.Err(); err != nil {
		return
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer func() {
		if x := recover(); x != nil {
			out = Result{}
			err = fmt.Errorf("文档损坏或资源超限")
		}
		if ctx.Err() != nil {
			out = Result{}
			err = fmt.Errorf("文档解析已取消或超时: %w", ctx.Err())
		}
	}()
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt":
		data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
		if !utf8.Valid(data) {
			return out, errors.New("文本必须使用UTF-8编码")
		}
		out.Text = string(data)
		out.Encoding = "UTF-8"
	case ".docx":
		var r io.ReadCloser
		r, err = OpenDOCXDocumentXML(data)
		if err != nil {
			return
		}
		defer r.Close()
		d := xml.NewDecoder(r)
		var b strings.Builder
		count := 0
		inText := false
		for {
			if err = ctx.Err(); err != nil {
				return
			}
			var tok xml.Token
			tok, err = d.Token()
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				return
			}
			switch v := tok.(type) {
			case xml.StartElement:
				if v.Name.Local == "t" {
					inText = true
				}
			case xml.EndElement:
				if v.Name.Local == "t" {
					inText = false
				}
				if v.Name.Local == "p" {
					b.WriteByte('\n')
					count++
				}
			case xml.CharData:
				if inText {
					count += utf8.RuneCount(v)
					if count > MaxCharacters {
						return out, errors.New("正文超过200万字元，请拆分")
					}
					b.Write(v)
				}
			}
		}
		out.Text = b.String()
		out.Encoding = "DOCX XML"
	case ".pdf":
		out, err = extractPDF(ctx, data)
	default:
		return out, errors.New("仅支持TXT、DOCX和可提取文字的PDF")
	}
	if err != nil {
		return Result{}, err
	}
	out.Text = strings.TrimSpace(out.Text)
	if utf8.RuneCountInString(out.Text) > MaxCharacters {
		return Result{}, errors.New("正文超过200万字元，请拆分")
	}
	if out.Text == "" {
		return Result{}, errors.New("未提取到正文，可能是扫描件，请上传可复制文字的文档或改用文本")
	}
	return
}
