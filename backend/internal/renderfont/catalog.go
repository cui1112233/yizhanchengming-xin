// Package renderfont owns the font bytes shipped in the Go executable.
package renderfont

import (
	_ "embed"
	"fmt"
	"image"
	"math"
	"sync"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const Noto = "noto-sans-sc"
const XiangHai = "xiang-hai"

//go:embed fonts/NotoSansSC.ttf
var notoBytes []byte

//go:embed fonts/XiangHai.ttf
var xiangHaiBytes []byte

var noto = sync.OnceValues(func() (*truetype.Font, error) { return truetype.Parse(notoBytes) })
var xiangHai = sync.OnceValues(func() (*truetype.Font, error) { return truetype.Parse(xiangHaiBytes) })

type FontInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Coverage  string `json:"coverage"`
	Available bool   `json:"available"`
}

func Catalog() []FontInfo {
	_, nerr := noto()
	_, xerr := xiangHai()
	return []FontInfo{{Noto, "思源黑体", "常用中文、拉丁字母和标点，未收录字符显示方框", nerr == nil}, {XiangHai, "湘海手写", "12 个标题用字，其他字使用思源黑体", nerr == nil && xerr == nil}}
}

func Open(id string, size float64) (font.Face, error) {
	if size <= 0 || math.IsNaN(size) || size > 256 {
		return nil, fmt.Errorf("字号须在 0 到 256 像素之间")
	}
	if id != Noto && id != XiangHai {
		return nil, fmt.Errorf("找不到字体 %q，请重新选择", id)
	}
	n, err := noto()
	if err != nil {
		return nil, fmt.Errorf("加载思源黑体失败: %w", err)
	}
	options := &truetype.Options{Size: size, DPI: 72, GlyphCacheEntries: 16}
	fallback := truetype.NewFace(n, options)
	if id == Noto {
		return fallback, nil
	}
	x, err := xiangHai()
	if err != nil {
		fallback.Close()
		return nil, fmt.Errorf("加载湘海字体失败: %w", err)
	}
	return &mixedFace{Face: fallback, special: truetype.NewFace(x, options), subset: x}, nil
}

// Each face owns its glyph cache. Only immutable parsed fonts are shared.
type mixedFace struct {
	font.Face
	special font.Face
	subset  *truetype.Font
}

func (f *mixedFace) pick(r rune) font.Face {
	if f.subset.Index(r) != 0 {
		return f.special
	}
	return f.Face
}
func (f *mixedFace) Glyph(p fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	return f.pick(r).Glyph(p, r)
}
func (f *mixedFace) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	return f.pick(r).GlyphBounds(r)
}
func (f *mixedFace) GlyphAdvance(r rune) (fixed.Int26_6, bool) { return f.pick(r).GlyphAdvance(r) }
func (f *mixedFace) Kern(a, b rune) fixed.Int26_6 {
	fa, fb := f.pick(a), f.pick(b)
	if fa != fb {
		return 0
	}
	return fa.Kern(a, b)
}
func (f *mixedFace) Close() error { f.special.Close(); return f.Face.Close() }
