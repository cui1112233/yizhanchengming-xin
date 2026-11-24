package batchfactorytext

import _ "embed"

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
	c.TextColor = "#123040"
	c.HighlightRules = "off"
	return c
}
