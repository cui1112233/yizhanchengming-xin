// Package renderfont owns the openly licensed fonts used for text pages.
package renderfont

import (
	"embed"
	"fmt"
	"image"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const Noto = "noto-sans-sc"
const MaShanZheng = "ma-shan-zheng"

// Assets are unmodified upstream release files. The catalog exposes the
// pinned source, embedded copyright notice and license with each binary.
//
//go:embed fonts/NotoSansSC-Regular.otf
var notoBytes []byte

//go:embed fonts/Noto-OFL.txt
var notoLicense string

//go:embed fonts/MaShanZheng-Regular.ttf
var maBytes []byte

//go:embed fonts/MaShanZheng-OFL.txt
var maLicense string

type FontInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Coverage    string `json:"coverage"`
	Available   bool   `json:"available"`
	License     string `json:"license"`
	Copyright   string `json:"copyright"`
	LicenseText string `json:"licenseText"`
	Source      string `json:"source"`
}

type entry struct {
	info   FontInfo
	parsed *sfnt.Font
	err    error
}

func loadEntry(id, name, coverage, source, license string, data []byte) entry {
	e := entry{info: FontInfo{ID: id, Name: name, Coverage: coverage, License: "OFL-1.1", Source: source, LicenseText: license}}
	e.parsed, e.err = opentype.Parse(data)
	if e.err == nil {
		var family string
		family, e.err = e.parsed.Name(nil, sfnt.NameIDFamily)
		if e.err == nil && family != name {
			e.err = fmt.Errorf("字体文件名称 %q 与目录 %q 不一致", family, name)
		}
		if e.err == nil {
			e.info.Copyright, e.err = e.parsed.Name(nil, sfnt.NameIDCopyright)
		}
	}
	e.info.Available = e.err == nil
	return e
}

var noto = sync.OnceValue(func() entry {
	return loadEntry(Noto, "Noto Sans SC", "30,890 个字符，未收录字符显示方框", "https://github.com/notofonts/noto-cjk/blob/f8d157532fbfaeda587e826d4cd5b21a49186f7c/Sans/SubsetOTF/SC/NotoSansSC-Regular.otf", notoLicense, notoBytes)
})
var ma = sync.OnceValue(func() entry {
	return loadEntry(MaShanZheng, "Ma Shan Zheng", "7,015 个字符，缺字时使用 Noto Sans SC，仍缺字时显示方框", "https://github.com/google/fonts/blob/06bfe0d62a98375cb09c942fa18e68bd32cff057/ofl/mashanzheng/MaShanZheng-Regular.ttf", maLicense, maBytes)
})

//go:embed fonts/*.ttf fonts/*OFL.txt
var extraFiles embed.FS

var extraFonts = sync.OnceValue(func() map[string]entry {
	result := map[string]entry{}
	for _, item := range []struct{ id, family, file, license, source string }{
		{"lxgw-wenkai", "霞鶩文楷", "LXGWWenKai-Regular.ttf", "LXGWWenKai-OFL.txt", "https://github.com/lxgw/LxgwWenKai"},
		{"long-cang", "Long Cang", "LongCang-Regular.ttf", "longcang-OFL.txt", "https://github.com/google/fonts/tree/main/ofl/longcang"},
		{"zhi-mang-xing", "Zhi Mang Xing", "ZhiMangXing-Regular.ttf", "zhimangxing-OFL.txt", "https://github.com/google/fonts/tree/main/ofl/zhimangxing"},
		{"zcool-kuaile", "ZCOOL KuaiLe", "ZCOOLKuaiLe-Regular.ttf", "zcoolkuaile-OFL.txt", "https://github.com/google/fonts/tree/main/ofl/zcoolkuaile"},
	} {
		data, _ := extraFiles.ReadFile("fonts/" + item.file)
		license, _ := extraFiles.ReadFile("fonts/" + item.license)
		result[item.id] = loadEntry(item.id, item.family, "缺字时使用 Noto Sans SC", item.source, string(license), data)
	}
	return result
})

func Catalog() []FontInfo {
	result := []FontInfo{noto().info, ma().info}
	for _, id := range []string{"lxgw-wenkai", "long-cang", "zhi-mang-xing", "zcool-kuaile"} {
		result = append(result, extraFonts()[id].info)
	}
	return result
}

type familyFace struct {
	font.Face
	family *sfnt.Font
}

func open(e entry, size float64) (familyFace, error) {
	if e.err != nil {
		return familyFace{}, e.err
	}
	f, err := opentype.NewFace(e.parsed, &opentype.FaceOptions{Size: size, DPI: 72})
	return familyFace{f, e.parsed}, err
}
func Open(id string, size float64) (font.Face, error) {
	if !(size > 0 && size <= 256) {
		return nil, fmt.Errorf("字号须在 0 到 256 像素之间")
	}
	extra, hasExtra := extraFonts()[id]
	if id != Noto && id != MaShanZheng && !hasExtra {
		return nil, fmt.Errorf("找不到字体 %q，请重新选择", id)
	}
	n, err := open(noto(), size)
	if err != nil {
		return nil, err
	}
	f := &mixedFace{familyFace: n, base: n}
	if id == MaShanZheng {
		f.familyFace, err = open(ma(), size)
		if err != nil {
			n.Close()
			return nil, err
		}
	}
	if hasExtra {
		f.familyFace, err = open(extra, size)
		if err != nil {
			n.Close()
			return nil, err
		}
	}
	return f, nil
}

// Parsed fonts can be shared. Lookup buffers and rasterizers belong to each face.
type mixedFace struct {
	familyFace
	base familyFace
	buf  sfnt.Buffer
}

func (f *mixedFace) pick(r rune) (font.Face, rune) {
	if idx, err := f.family.GlyphIndex(&f.buf, r); err == nil && idx != 0 {
		return f.Face, r
	}
	if f.base.family != f.family {
		if idx, err := f.base.family.GlyphIndex(&f.buf, r); err == nil && idx != 0 {
			return f.base.Face, r
		}
	}
	return f.base.Face, '□'
}
func (f *mixedFace) Glyph(p fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	face, r := f.pick(r)
	return face.Glyph(p, r)
}
func (f *mixedFace) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	face, r := f.pick(r)
	return face.GlyphBounds(r)
}
func (f *mixedFace) GlyphAdvance(r rune) (fixed.Int26_6, bool) {
	face, r := f.pick(r)
	return face.GlyphAdvance(r)
}
func (f *mixedFace) Kern(a, b rune) fixed.Int26_6 {
	fa, a := f.pick(a)
	fb, b := f.pick(b)
	if fa != fb {
		return 0
	}
	return fa.Kern(a, b)
}
func (f *mixedFace) Close() error {
	if f.family != f.base.family {
		f.base.Close()
	}
	return f.Face.Close()
}
