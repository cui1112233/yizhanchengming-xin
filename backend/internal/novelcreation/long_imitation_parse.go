package novelcreation

import (
	"context"
	"path/filepath"
	"qiantie/backend/internal/documenttext"
	"strings"
)

func (s *Service) PreviewLongImitation(ctx context.Context, owner, text string) (map[string]any, error) {
	if _, e := requiredOwner(owner); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if jsLength(text) > LongImitationMaxUTF16 {
		return nil, inputError("原文不能超过200万个UTF16字元")
	}
	return map[string]any{"statistics": ImitationStyleStatistics(text, ""), "styleDimensions": ImitationStyleDimensions()}, nil
}
func (s *Service) ParseLongImitationReference(ctx context.Context, owner, filename string, data []byte) ([]MaterialChapter, error) {
	if _, e := requiredOwner(owner); e != nil {
		return nil, e
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".txt" && ext != ".docx" {
		return nil, inputError("仅支持TXT和DOCX文件")
	}
	if len(data) > documenttext.MaxFileBytes {
		return nil, inputError("单个文件不能超过20MiB")
	}
	parser := s.documentParser
	if parser == nil {
		parser = documenttext.NewParser()
	}
	result, e := parser.Parse(ctx, filename, data)
	if e != nil {
		return nil, inputError(e.Error())
	}
	if jsLength(result.Text) > LongImitationMaxUTF16 {
		return nil, inputError("文件正文超过200万个UTF16字元，请拆分")
	}
	if jsTrim(result.Text) == "" {
		return nil, inputError("文件没有可用正文")
	}
	chapters := splitMaterialChapters(result.Text)
	if len(chapters) > 2000 {
		return nil, inputError("文件超过2000章，请拆分")
	}
	for i := range chapters {
		chapters[i].Characters = jsNonspaceLength(chapters[i].Content)
	}
	return chapters, nil
}
