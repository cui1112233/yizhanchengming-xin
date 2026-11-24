package batchfactorytext

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
	"qiantie/backend/internal/renderfont"
)

type HighlightRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Ranges returns half-open rune offsets relative to the supplied normalized page.
// A book/version without saved illustration emphasis returns no ranges.
type HighlightResolver interface {
	Ranges(ctx context.Context, owner, bookID, versionID, text string) ([]HighlightRange, error)
}

// KeywordHighlightSource adapts saved illustration keywords to page-local
// emphasis. The reader gets the owner and book/version identity, and returns
// an empty list when that version has no reusable data.
type KeywordHighlightSource func(ctx context.Context, owner, bookID, versionID string) ([]string, error)

func (source KeywordHighlightSource) Ranges(ctx context.Context, owner, bookID, versionID, text string) ([]HighlightRange, error) {
	words, err := source(ctx, owner, bookID, versionID)
	if err != nil {
		return nil, err
	}
	return KeywordHighlightRanges(text, words), nil
}

// KeywordHighlightRanges maps existing emphasis phrases to the occurrences on
// this page. It performs no keyword selection or inference.
func KeywordHighlightRanges(text string, words []string) []HighlightRange {
	var out []HighlightRange
	for _, w := range words {
		if strings.TrimSpace(w) == "" {
			continue
		}
		byteOffset, runeOffset := 0, 0
		for byteOffset < len(text) {
			i := strings.Index(text[byteOffset:], w)
			if i < 0 {
				break
			}
			runeOffset += utf8.RuneCountInString(text[byteOffset : byteOffset+i])
			end := runeOffset + utf8.RuneCountInString(w)
			out = append(out, HighlightRange{runeOffset, end})
			byteOffset += i + len(w)
			runeOffset = end
		}
	}
	return mergeRanges(out)
}

func mergeRanges(rs []HighlightRange) []HighlightRange {
	sort.Slice(rs, func(i, j int) bool { return rs[i].Start < rs[j].Start })
	out := rs[:0]
	for _, v := range rs {
		if n := len(out); n > 0 && v.Start <= out[n-1].End {
			out[n-1].End = max(out[n-1].End, v.End)
		} else {
			out = append(out, v)
		}
	}
	return out
}

type Renderer struct{ Highlights HighlightResolver }

type RenderInput struct {
	Owner, BookID, VersionID string
	Text                     string
	Layout                   LayoutConfig
	Background               Background
	// The caller resolves authenticated storage references into bytes.
	BackgroundImage       []byte
	HighlightReuseEnabled bool
	DialogueRanges        []HighlightRange
	DialogueRangesSet     bool
	VerticalCenterRatio   float64
	// Preview samples the actual cropped background behind the glyphs.
	WarnLowContrast func()
}
type MeasureResult struct {
	Fits                  bool
	UsedHeight, MaxHeight float64
	Width, Height, Lines  int
}

func DefaultRenderLayout() LayoutConfig {
	return LayoutConfig{AspectRatio: "3:4", TargetCharacters: DefaultTargetCharacters, MaxPages: DefaultMaxPages, PaginationMode: "auto", FontID: renderfont.Noto, FontSize: 44, FontWeight: 400, LineHeight: 1.65, ParagraphSpacing: 24, Padding: 96, Align: "left", TextColor: "#FFFFFF", HighlightColor: "#E53935", MaxHeightRatio: .75, paddingSet: true, paragraphSpacingSet: true}
}

var canvasSizes = map[string]image.Point{
	"3:4": {1080, 1440}, "4:3": {1440, 1080}, "9:16": {1080, 1920}, "16:9": {1920, 1080}, "21:9": {2520, 1080}, "1:1": {1080, 1080}, "3:2": {1620, 1080}, "2:3": {1080, 1620},
}

func resolveLayout(c LayoutConfig) (LayoutConfig, image.Point, error) {
	if err := validateHighlightRules(c); err != nil {
		return c, image.Point{}, err
	}
	defaults := DefaultRenderLayout()
	if c.PaginationMode != "" && c.PaginationMode != "auto" && c.PaginationMode != "target" {
		return c, image.Point{}, fmt.Errorf("不支持分页方式 %q", c.PaginationMode)
	}
	// Sparse book snapshots inherit missing spacing. The layout editor starts
	// from DefaultRenderLayout and can set either measurement explicitly to zero.
	if c.Padding == 0 && !c.paddingSet {
		c.Padding = defaults.Padding
	}
	if c.ParagraphSpacing == 0 && !c.paragraphSpacingSet {
		c.ParagraphSpacing = defaults.ParagraphSpacing
	}
	if c.AspectRatio == "" {
		c.AspectRatio = defaults.AspectRatio
	}
	if c.FontID == "" {
		c.FontID = defaults.FontID
	}
	if c.FontSize == 0 {
		c.FontSize = defaults.FontSize
	}
	if c.FontWeight == 0 {
		c.FontWeight = defaults.FontWeight
	}
	if c.LineHeight == 0 {
		c.LineHeight = defaults.LineHeight
	}
	if c.MaxHeightRatio == 0 {
		c.MaxHeightRatio = defaults.MaxHeightRatio
	}
	if c.Align == "" {
		c.Align = defaults.Align
	}
	if c.TextColor == "" {
		c.TextColor = defaults.TextColor
	}
	if c.HighlightColor == "" {
		c.HighlightColor = defaults.HighlightColor
	}
	size, ok := canvasSizes[c.AspectRatio]
	if !ok {
		return c, size, fmt.Errorf("不支持画幅 %q", c.AspectRatio)
	}
	if !(c.FontSize >= 1 && c.FontSize <= 256 && c.LineHeight >= 1 && c.LineHeight <= 8 && c.ParagraphSpacing >= 0 && c.ParagraphSpacing <= float64(size.Y) && c.MaxHeightRatio > 0 && c.MaxHeightRatio <= 1) || c.Padding < 0 || c.Padding >= (min(size.X, size.Y)+1)/2 || c.FontWeight < 100 || c.FontWeight > 900 {
		return c, size, fmt.Errorf("字号、字重、间距、边距或文字区域超出可用范围")
	}
	if c.Align != "left" && c.Align != "right" && c.Align != "center" {
		return c, size, fmt.Errorf("不支持文字对齐方式 %q", c.Align)
	}
	return c, size, nil
}

func rgb(s string) (color.RGBA, error) {
	if len(s) == 7 && s[0] == '#' {
		b, err := hex.DecodeString(s[1:])
		if err == nil {
			return color.RGBA{R: b[0], G: b[1], B: b[2], A: 255}, nil
		}
	}
	return color.RGBA{}, fmt.Errorf("无效颜色 %q，请使用 #RRGGBB", s)
}

type glyphPlacement struct {
	r     rune
	index int
	x     fixed.Int26_6
}
type lineGeometry struct {
	glyphs                             []glyphPlacement
	left, right, advance, above, below fixed.Int26_6
}
type linePainter func(line lineGeometry, x, y int) error
type glyphShape struct {
	bounds  fixed.Rectangle26_6
	advance fixed.Int26_6
}

func boldSpread(c LayoutConfig) int {
	// Both bundled fonts have regular outlines; additional strokes share the
	// measured width and ink bounds used by the painter.
	return max(0, int(math.Round(float64(c.FontWeight-400)*c.FontSize/6800)))
}

// The shipped fonts have regular outlines. For light weights, reduce ink
// on the right-hand edge of each stroke while retaining fine strokes.
func lightGlyphMask(mask image.Image, origin image.Point, size image.Point, weight int) *image.Alpha {
	out := image.NewAlpha(image.Rectangle{Max: size})
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			_, _, _, a := mask.At(origin.X+x, origin.Y+y).RGBA()
			var neighbor uint32
			if x+1 < size.X {
				_, _, _, neighbor = mask.At(origin.X+x+1, origin.Y+y).RGBA()
			}
			if a > neighbor {
				a -= (a - neighbor) * uint32(400-weight) / 400
			}
			out.Pix[y*out.Stride+x] = uint8(a >> 8)
		}
	}
	return out
}

func layoutText(ctx context.Context, text string, c LayoutConfig, size image.Point, face font.Face, paint linePainter) (MeasureResult, error) {
	if !utf8.ValidString(text) {
		return MeasureResult{}, fmt.Errorf("正文不是有效的 UTF-8 文本")
	}
	text = NormalizeLineEndings(text)
	width := size.X - 2*c.Padding
	limit := min(int(math.Floor(float64(size.Y)*c.MaxHeightRatio)), size.Y-2*c.Padding)
	m := MeasureResult{Fits: true, MaxHeight: float64(limit), Width: size.X, Height: size.Y}
	metrics := face.Metrics()
	baselineTop := (size.Y - limit) / 2
	if paint != nil {
		measured, err := layoutText(ctx, text, c, size, face, nil)
		if err != nil {
			return measured, err
		}
		baselineTop = (size.Y - int(measured.UsedHeight)) / 2
	}
	reading := readingBreaks(text, c, width)
	spread := boldSpread(c)
	step := int(math.Ceil(c.FontSize * c.LineHeight))
	var line lineGeometry
	var prev rune
	var top, lastHeight int
	var hasGlyph bool
	var hasVisibleGlyph bool
	shapes := make(map[rune]glyphShape)
	flush := func() error {
		ascent := max(metrics.Ascent.Ceil(), line.above.Ceil())
		desc := max(metrics.Descent.Ceil(), line.below.Ceil())
		lastHeight = ascent + desc
		if line.right.Ceil()-line.left.Floor() > width {
			m.Fits = false
		}
		if paint != nil {
			x := c.Padding - line.left.Floor()
			gap := width - (line.right.Ceil() - line.left.Floor())
			switch c.Align {
			case "center":
				x += gap / 2
			case "right":
				x += gap
			}
			if err := paint(line, x, baselineTop+top+ascent); err != nil {
				return err
			}
		}
		m.Lines++
		m.UsedHeight = float64(top + lastHeight)
		line = lineGeometry{glyphs: line.glyphs[:0]}
		prev = 0
		hasGlyph = false
		hasVisibleGlyph = false
		return nil
	}
	runeIndex := 0
	runes := []rune(text)
	for _, r := range text {
		if reading[runeIndex] && hasGlyph && r != '\n' {
			if err := flush(); err != nil {
				return m, err
			}
			top += max(lastHeight, step) + max(int(math.Round(c.ParagraphSpacing)), step/2)
		}
		if runeIndex%256 == 0 {
			if err := ctx.Err(); err != nil {
				return m, err
			}
		}
		if r == '\n' {
			// Blank source separators retain their offsets but do not add another
			// empty line on top of the configured paragraph spacing.
			if !hasVisibleGlyph {
				line = lineGeometry{glyphs: line.glyphs[:0]}
				prev = 0
				hasGlyph = false
				runeIndex++
				continue
			}
			if err := flush(); err != nil {
				return m, err
			}
			top += max(lastHeight, step) + int(math.Round(c.ParagraphSpacing))
			runeIndex++
			continue
		}
		// Tabs expand to four measured spaces. They retain their original index.
		display := r
		n := 1
		if r == '\t' {
			display = ' '
			n = 4
		}
		shape, found := shapes[display]
		if !found {
			bounds, ad, ok := face.GlyphBounds(display)
			if !ok {
				return m, fmt.Errorf("无法测量第 %d 个字符", runeIndex+1)
			}
			shape = glyphShape{bounds, ad}
			shapes[display] = shape
		}
		advance := shape.advance * fixed.Int26_6(n)
		x := line.advance
		if hasGlyph {
			x += face.Kern(prev, display)
		}
		right := max(line.right, x+advance, x+shape.bounds.Max.X+fixed.I(spread))
		left := min(line.left, x+shape.bounds.Min.X)
		// Keep an opening mark with its next character, and reserve space for
		// closing marks following the current character. Measurement and painting
		// use the same lookahead; original rune indexes never change.
		reservedRight, reservedLeft := right, left
		cursor, previous := x+advance, display
		for j := runeIndex + 1; j < len(runes); j++ {
			next := runes[j]
			if next == '\n' || (!strings.ContainsRune("，。！？；：、）》】」』”’", next) && !strings.ContainsRune("（《【「『“‘", previous)) {
				break
			}
			bounds, ad, ok := face.GlyphBounds(next)
			if !ok {
				break
			}
			cursor += face.Kern(previous, next)
			reservedRight = max(reservedRight, cursor+ad, cursor+bounds.Max.X+fixed.I(spread))
			reservedLeft = min(reservedLeft, cursor+bounds.Min.X)
			cursor += ad
			previous = next
		}
		if hasGlyph && reservedRight.Ceil()-reservedLeft.Floor() > width {
			if err := flush(); err != nil {
				return m, err
			}
			top += max(lastHeight, step)
			x = 0
			left = min(0, shape.bounds.Min.X)
			right = max(advance, shape.bounds.Max.X+fixed.I(spread))
		}
		line.advance = x + advance
		line.right = right
		line.left = left
		line.above = max(line.above, -shape.bounds.Min.Y)
		line.below = max(line.below, shape.bounds.Max.Y)
		if paint != nil {
			line.glyphs = append(line.glyphs, glyphPlacement{display, runeIndex, x})
		}
		prev = display
		hasGlyph = true
		hasVisibleGlyph = hasVisibleGlyph || strings.TrimSpace(string(r)) != ""
		runeIndex++
	}
	if hasGlyph {
		if err := flush(); err != nil {
			return m, err
		}
	}
	m.Fits = m.Fits && m.UsedHeight <= m.MaxHeight
	return m, ctx.Err()
}

func (r Renderer) Measure(text string, c LayoutConfig) (MeasureResult, error) {
	return r.measure(context.Background(), text, c)
}

// MeasureFunc is the directly usable real-font callback for Paginate.
func (r Renderer) MeasureFunc(ctx context.Context) MeasureFunc {
	return func(text string, c LayoutConfig) (bool, float64, error) {
		m, err := r.measure(ctx, text, c)
		return m.Fits, m.UsedHeight, err
	}
}
func (r Renderer) measure(ctx context.Context, text string, c LayoutConfig) (MeasureResult, error) {
	if err := ctx.Err(); err != nil {
		return MeasureResult{}, err
	}
	c, size, err := resolveLayout(c)
	if err != nil {
		return MeasureResult{}, err
	}
	face, err := renderfont.Open(c.FontID, c.FontSize)
	if err != nil {
		return MeasureResult{}, err
	}
	defer face.Close()
	return layoutText(ctx, text, c, size, face, nil)
}

func drawBackground(dst *image.RGBA, b Background, encoded []byte) error {
	base := b.Color
	if base == "" {
		base = "#111318"
	}
	col, err := rgb(base)
	if err != nil {
		return err
	}
	draw.Draw(dst, dst.Bounds(), image.NewUniform(col), image.Point{}, draw.Src)
	switch b.Kind {
	case "", "color", "solid":
		return nil
	case "image", "upload", "library":
	default:
		return fmt.Errorf("不支持背景类型 %q", b.Kind)
	}
	if len(encoded) == 0 || len(encoded) > 32<<20 {
		return fmt.Errorf("背景图片为空或过大，请重新选择")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(encoded))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width) > 32_000_000/int64(cfg.Height) {
		return fmt.Errorf("背景图片损坏或分辨率过大，请重新选择")
	}
	img, _, err := image.Decode(bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("背景图片无法解码: %w", err)
	}
	// Center-crop before scaling, keeping extreme panorama dimensions bounded.
	src := img.Bounds()
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	if int64(src.Dx())*int64(h) > int64(src.Dy())*int64(w) {
		target := max(1, int(int64(src.Dy())*int64(w)/int64(h)))
		src.Min.X += (src.Dx() - target) / 2
		src.Max.X = src.Min.X + target
	} else {
		target := max(1, int(int64(src.Dx())*int64(h)/int64(w)))
		src.Min.Y += (src.Dy() - target) / 2
		src.Max.Y = src.Min.Y + target
	}
	xdraw.ApproxBiLinear.Scale(dst, dst.Rect, img, src, draw.Over, nil)
	return nil
}

func (r Renderer) Render(ctx context.Context, in RenderInput) ([]byte, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	c, size, err := resolveLayout(in.Layout)
	if err != nil {
		return nil, "", err
	}
	normal, err := rgb(c.TextColor)
	if err != nil {
		return nil, "", err
	}
	highlight, err := rgb(c.HighlightColor)
	if err != nil {
		return nil, "", err
	}
	face, err := renderfont.Open(c.FontID, c.FontSize)
	if err != nil {
		return nil, "", err
	}
	defer face.Close()
	text := NormalizeLineEndings(in.Text)
	m, err := layoutText(ctx, text, c, size, face, nil)
	if err != nil {
		return nil, "", err
	}
	if !m.Fits {
		return nil, "", fmt.Errorf("本页排版超出文字区域，请调整字号、边距或重新分页")
	}
	var ranges []HighlightRange
	if ((c.HighlightRules == "" && in.HighlightReuseEnabled) || HighlightRuleEnabled(c.HighlightRules, "keywords")) && r.Highlights != nil {
		ranges, err = r.Highlights.Ranges(ctx, in.Owner, in.BookID, in.VersionID, text)
		if err != nil {
			return nil, "", err
		}
		limit := utf8.RuneCountInString(text)
		for _, v := range ranges {
			if v.Start < 0 || v.End <= v.Start || v.End > limit {
				return nil, "", fmt.Errorf("标红位置与当前页正文不匹配")
			}
		}
		ranges = mergeRanges(append([]HighlightRange(nil), ranges...))
	}
	dialogue := in.DialogueRanges
	if !in.DialogueRangesSet {
		dialogue = DialogueHighlightRanges(text)
	}
	if HighlightRuleEnabled(c.HighlightRules, "dialogue") {
		limit := utf8.RuneCountInString(text)
		for _, v := range dialogue {
			v.Start = max(0, v.Start)
			v.End = min(limit, v.End)
			if v.Start < v.End {
				ranges = append(ranges, v)
			}
		}
		ranges = mergeRanges(ranges)
	}
	selectedLines := map[int]bool{}
	if HighlightRuleEnabled(c.HighlightRules, "lines") {
		selectedLines, _ = parseHighlightLines(c.HighlightLines)
		visualLine := 0
		_, err := layoutText(ctx, text, c, size, face, func(line lineGeometry, _, _ int) error {
			for _, g := range line.glyphs {
				if !unicode.IsSpace(g.r) {
					visualLine++
					if selectedLines[visualLine] {
						ranges = append(ranges, sentenceHighlightAt(text, g.index, dialogue))
					}
					break
				}
			}
			return nil
		})
		if err != nil {
			return nil, "", err
		}
		ranges = mergeRanges(ranges)
	}
	dst := image.NewRGBA(image.Rectangle{Max: size})
	if err = drawBackground(dst, in.Background, in.BackgroundImage); err != nil {
		return nil, "", err
	}
	fg, accent := image.NewUniform(normal), image.NewUniform(highlight)
	rangeIndex := 0
	spread := boldSpread(c)
	paint := func(line lineGeometry, x, y int) error {
		if in.VerticalCenterRatio > 0 {
			y += int(math.Round(float64(size.Y) * (in.VerticalCenterRatio - 0.5)))
		}
		for _, g := range line.glyphs {
			for rangeIndex < len(ranges) && ranges[rangeIndex].End <= g.index {
				rangeIndex++
			}
			ink := fg
			if rangeIndex < len(ranges) && ranges[rangeIndex].Start <= g.index {
				ink = accent
			}
			dot := fixed.P(x, y)
			dot.X += g.x
			rect, mask, mp, _, ok := face.Glyph(dot, g.r)
			if !ok {
				return fmt.Errorf("无法绘制第 %d 个字符", g.index+1)
			}
			if in.WarnLowContrast != nil && !rect.Empty() {
				point := rect.Min.Add(rect.Size().Div(2))
				if point.In(dst.Rect) && lowContrast(ink.C.(color.RGBA), dst.RGBAAt(point.X, point.Y)) {
					in.WarnLowContrast()
					in.WarnLowContrast = nil
				}
			}
			if c.FontWeight < 400 {
				mask = lightGlyphMask(mask, mp, rect.Size(), c.FontWeight)
				mp = image.Point{}
			}
			for dx := 0; dx <= spread; dx++ {
				draw.DrawMask(dst, rect.Add(image.Pt(dx, 0)), ink, image.Point{}, mask, mp, draw.Over)
			}
		}
		return nil
	}
	if _, err := layoutText(ctx, text, c, size, face, paint); err != nil {
		return nil, "", err
	}
	var out bytes.Buffer
	if err = png.Encode(&out, dst); err != nil {
		return nil, "", err
	}
	if err = ctx.Err(); err != nil {
		return nil, "", err
	}
	return out.Bytes(), "image/png", nil
}

func lowContrast(fg, bg color.RGBA) bool {
	linear := func(b uint8) float64 {
		c := float64(b) / 255
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	lum := func(c color.RGBA) float64 { return .2126*linear(c.R) + .7152*linear(c.G) + .0722*linear(c.B) }
	a, b := lum(fg), lum(bg)
	return (max(a, b)+.05)/(min(a, b)+.05) < 3
}
