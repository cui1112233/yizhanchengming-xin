package batchfactorytext

import (
	_ "embed"
	"strings"
)

const DefaultCoverVersion = "builtin-text-cover-v2"

func FormatCoverTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.TrimPrefix(title, "《")
	title = strings.TrimSuffix(title, "》")
	title = strings.TrimSpace(title)
	if title == "" {
		title = "未命名作品"
	}
	return "《" + title + "》"
}

//go:embed assets/default-cover.png
var defaultCoverImage []byte

func DefaultCoverImage() []byte { return defaultCoverImage }
func DefaultCoverLayout(aspect string) LayoutConfig {
	c := DefaultRenderLayout()
	if aspect != "" {
		c.AspectRatio = aspect
	}
	c.FontSize = 64
	c.FontWeight = 600
	c.LineHeight = 1.4
	c.Align = "center"
	c.TextColor = "#FFFFFF"
	c.HighlightRules = "off"
	return c
}
