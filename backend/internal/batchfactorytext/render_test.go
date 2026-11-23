package batchfactorytext

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"testing"

	"qiantie/backend/internal/renderfont"
)

func renderCard(t *testing.T, renderer Renderer, in RenderInput) (*image.RGBA, []byte) {
	t.Helper()
	data, kind, err := renderer.Render(context.Background(), in)
	if err != nil || kind != "image/png" {
		t.Fatalf("render %q: %v", kind, err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	out := image.NewRGBA(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			out.Set(x, y, img.At(x, y))
		}
	}
	return out, data
}

func TestRenderFontsIncludeChineseFallback(t *testing.T) {
	catalog := renderfont.Catalog()
	if len(catalog) != 2 || !catalog[1].Available || catalog[1].Coverage == "" {
		t.Fatal(catalog)
	}
	a, err := renderfont.Open("noto-sans-sc", 44)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := renderfont.Open("xiang-hai", 44)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	for _, r := range "你好🍎" {
		bounds, _, ok := b.GlyphBounds(r)
		if !ok || bounds.Empty() {
			t.Fatalf("missing visible glyph %q", r)
		}
		n, _, _ := a.GlyphBounds(r)
		if (bounds == n) != (r != '你') {
			t.Fatalf("wrong font for %q", r)
		}
	}
	if _, err := renderfont.Open("absent", 44); err == nil {
		t.Fatal("unknown font accepted")
	}
}

func TestRenderCanonicalPixels(t *testing.T) {
	r := Renderer{}
	in := RenderInput{Text: "今天下雨了。\r\n我说：“明天见。”", Background: Background{Kind: "color", Color: "#123456"}}
	img, data := renderCard(t, r, in)
	if img.Bounds() != image.Rect(0, 0, 1080, 1440) || img.RGBAAt(0, 0) != (color.RGBA{18, 52, 86, 255}) {
		t.Fatalf("background %v %v", img.Bounds(), img.At(0, 0))
	}
	n := 0
	for y := 0; y < 1440; y++ {
		for x := 0; x < 1080; x++ {
			if img.RGBAAt(x, y) != img.RGBAAt(0, 0) {
				n++
				if x < 96 || y < 180 || y >= 1260 {
					t.Fatalf("out of text region %d,%d", x, y)
				}
			}
		}
	}
	if n < 100 {
		t.Fatal("text missing")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "fa266605f737b0eb08a7a8d0556b71ab2dbea6d3eb7916a1712553b3ea0efe06" {
		t.Fatalf("Noto card SHA-256 %s", got)
	}
}

func TestRenderSparseStyleKeepsDefaultGeometry(t *testing.T) {
	r := Renderer{}
	text := strings.Repeat("这是测试文章。", 11) + "\n下一段"
	normal, err := r.Measure(text, LayoutConfig{})
	if err != nil {
		t.Fatal(err)
	}
	colorOnly, err := r.Measure(text, LayoutConfig{TextColor: "#777777"})
	if err != nil || normal != colorOnly {
		t.Fatalf("changing color changed geometry: %+v vs %+v", normal, colorOnly)
	}
	complete := DefaultRenderLayout()
	complete.Padding = 0
	complete.ParagraphSpacing = 0
	compact, err := r.Measure(text, complete)
	if err != nil || compact.UsedHeight >= normal.UsedHeight {
		t.Fatalf("explicit zero spacing %+v, %v", compact, err)
	}
}

func TestRenderAlignmentAndSpacing(t *testing.T) {
	r := Renderer{}
	c := DefaultRenderLayout()
	var left [3]int
	for i, align := range []string{"left", "center", "right"} {
		c.Align = align
		img, _ := renderCard(t, r, RenderInput{Text: "字", Layout: c})
		left[i] = img.Rect.Dx()
		for y := 180; y < 260; y++ {
			for x := 0; x < img.Rect.Dx(); x++ {
				if img.RGBAAt(x, y) != img.RGBAAt(0, 0) {
					left[i] = min(left[i], x)
				}
			}
		}
	}
	if left[0] < 96 || left[1]-left[0] < 300 || left[2]-left[1] < 300 || left[2] >= 1080-96 {
		t.Fatal(left)
	}
	m, err := r.Measure("第一段\n第二段", c)
	if err != nil {
		t.Fatal(err)
	}
	c.ParagraphSpacing += 19
	n, err := r.Measure("第一段\n第二段", c)
	if err != nil || n.UsedHeight != m.UsedHeight+19 {
		t.Fatal(n, m, err)
	}
	c.LineHeight += .5
	z, err := r.Measure("第一段\n第二段", c)
	if err != nil || z.UsedHeight <= n.UsedHeight {
		t.Fatal(z, n, err)
	}
}

func TestRenderMeasureAgreement(t *testing.T) {
	r := Renderer{}
	c := DefaultRenderLayout()
	c.MaxHeightRatio = .18
	para := strings.Repeat("小镇黄昏。", 9)
	input := strings.Repeat(para+"\n", 8)
	pages, err := Paginate(context.Background(), PaginationInput{Text: input, Layout: c}, r.MeasureFunc(context.Background()))
	if err != nil || len(pages) < 2 {
		t.Fatalf("pagination %d %v", len(pages), err)
	}
	for _, page := range pages {
		m, err := r.Measure(page.Text, c)
		if err != nil || !m.Fits || m.UsedHeight > m.MaxHeight {
			t.Fatalf("%+v %v", m, err)
		}
		if _, _, err := r.Render(context.Background(), RenderInput{Text: page.Text, Layout: c}); err != nil {
			t.Fatal(err)
		}
	}
	if m, err := r.Measure(input, c); err != nil || m.Fits {
		t.Fatal(m, err)
	}
	if _, _, err := r.Render(context.Background(), RenderInput{Text: input, Layout: c}); err == nil {
		t.Fatal("overflow rendered")
	}
}

func TestRenderHighlightRuneGeometry(t *testing.T) {
	words := []string{"来啦"}
	calls := 0
	r := Renderer{Highlights: KeywordHighlightSource(func(_ context.Context, owner, book, version string) ([]string, error) {
		calls++
		if owner != "u" || book != "b" || version != "v" {
			return nil, errors.New("scope")
		}
		return words, nil
	})}
	in := RenderInput{Owner: "u", BookID: "b", VersionID: "v", Text: "你🍎来啦", Layout: DefaultRenderLayout()}
	_, plain := renderCard(t, r, in)
	if calls != 0 {
		t.Fatal("disabled lookup")
	}
	in.HighlightReuseEnabled = true
	img, red := renderCard(t, r, in)
	if calls != 1 || bytes.Equal(plain, red) {
		t.Fatal("emphasis lost")
	}
	ranges := KeywordHighlightRanges("你🍎来啦", words)
	if len(ranges) != 1 || ranges[0] != (HighlightRange{2, 4}) {
		t.Fatal(ranges)
	}
	count := 0
	for y := 180; y < 300; y++ {
		for x := 0; x < 1080; x++ {
			if img.RGBAAt(x, y) == (color.RGBA{229, 57, 53, 255}) {
				if x < 180 {
					t.Fatal("wrong rune red")
				}
				count++
			}
		}
	}
	if count == 0 {
		t.Fatal("no red pixels")
	}
	words = nil
	_, normal := renderCard(t, r, in)
	if !bytes.Equal(normal, plain) {
		t.Fatal("no-record output differs")
	}
}

func TestRenderBackgroundAndLayout(t *testing.T) {
	c := DefaultRenderLayout()
	c.AspectRatio = "16:9"
	c.Align = "right"
	c.FontWeight = 700
	tile := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			tile.SetRGBA(x, y, color.RGBA{3, 5, 7, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, tile); err != nil {
		t.Fatal(err)
	}
	r := Renderer{}
	in := RenderInput{Text: "字", Layout: c, Background: Background{Kind: "image", ID: "bg"}, BackgroundImage: buf.Bytes()}
	img, bold := renderCard(t, r, in)
	if img.Bounds().Dx()*9 != img.Bounds().Dy()*16 || img.RGBAAt(0, 0) != (color.RGBA{3, 5, 7, 255}) {
		t.Fatal(img.Bounds(), img.At(0, 0))
	}
	for y := 0; y < img.Rect.Dy(); y++ {
		if img.RGBAAt(img.Rect.Dx()/4, y) != img.RGBAAt(0, 0) {
			t.Fatal("text not right-aligned")
		}
	}
	in.Layout.FontWeight = 400
	_, normal := renderCard(t, r, in)
	if bytes.Equal(normal, bold) {
		t.Fatal("weight ignored")
	}
	in.Layout.FontWeight = 300
	_, light := renderCard(t, r, in)
	if bytes.Equal(normal, light) {
		t.Fatal("light stroke ignored")
	}
	in.BackgroundImage = []byte("damaged")
	if _, _, err := r.Render(context.Background(), in); err == nil {
		t.Fatal("invalid background accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := r.Render(ctx, in); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := r.Measure("字", LayoutConfig{FontSize: math.NaN()}); err == nil {
		t.Fatal("NaN size")
	}
}

func TestRenderImageBackgroundCenterCrop(t *testing.T) {
	// A square card takes the central two columns of this 6x2 panorama.
	tile := image.NewRGBA(image.Rect(0, 0, 6, 2))
	green := color.RGBA{4, 126, 8, 255}
	red := color.RGBA{250, 10, 7, 255}
	for x := 0; x < 6; x++ {
		for y := 0; y < 2; y++ {
			v := red
			if x == 2 || x == 3 {
				v = green
			}
			tile.SetRGBA(x, y, v)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, tile); err != nil {
		t.Fatal(err)
	}
	r := Renderer{}
	c := DefaultRenderLayout()
	c.AspectRatio = "1:1"
	img, _ := renderCard(t, r, RenderInput{Text: "背景", Layout: c, Background: Background{Kind: "library", ID: "a"}, BackgroundImage: buf.Bytes()})
	for _, p := range []image.Point{{0, 0}, {1079, 1079}} {
		if img.RGBAAt(p.X, p.Y) != green {
			t.Fatal("wrong crop", img.At(p.X, p.Y))
		}
	}
}
