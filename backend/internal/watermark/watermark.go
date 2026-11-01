// Package watermark 在生成的漫画图片角落叠加白底方框文字（如“热门小说”）。
package watermark

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

//go:embed NotoSansSC.ttf
var fontBytes []byte

var parsedFont *truetype.Font

func init() {
	var err error
	parsedFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(fmt.Sprintf("watermark: parse embedded font: %v", err))
	}
}

// Options 控制水印开关与文字内容。
type Options struct {
	Enabled bool
	Text    string
}

// Apply 在图片右下角叠加一个白底黑字的水印方框。原图尺寸过小或解析失败时
// 返回原图，不阻塞生成。
func Apply(data []byte, mediaType string, opts Options) ([]byte, string) {
	if !opts.Enabled || len(data) == 0 {
		return data, mediaType
	}
	text := opts.Text
	if text == "" {
		text = "热门小说"
	}

	var src image.Image
	var err error
	switch mediaType {
	case "image/png":
		src, err = png.Decode(bytes.NewReader(data))
	case "image/jpeg":
		src, err = jpeg.Decode(bytes.NewReader(data))
	default:
		src, err = png.Decode(bytes.NewReader(data))
	}
	if err != nil {
		return data, mediaType
	}

	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < 100 || height < 100 {
		return data, mediaType
	}

	// 方框尺寸：按图片宽度 1/4 且不小于 80px；高度按文字字号计算。
	fontSize := float64(width) / 22
	if fontSize < 14 {
		fontSize = 14
	}
	if fontSize > 48 {
		fontSize = 48
	}
	paddingX := int(fontSize * 0.55)
	paddingY := int(fontSize * 0.35)
	margin := int(float64(width) * 0.025)
	if margin < 8 {
		margin = 8
	}

	face := truetype.NewFace(parsedFont, &truetype.Options{Size: fontSize, DPI: 72})
	defer face.Close()

	// 测量文字宽度
	drawer := &font.Drawer{Face: face}
	textWidth := drawer.MeasureString(text).Ceil()
	boxWidth := textWidth + paddingX*2
	boxHeight := int(fontSize) + paddingY*2
	if boxWidth > width/2 {
		// 文字太长，缩小字号重试一次
		fontSize = fontSize * 0.8
		face = truetype.NewFace(parsedFont, &truetype.Options{Size: fontSize, DPI: 72})
		defer face.Close()
		drawer = &font.Drawer{Face: face}
		textWidth = drawer.MeasureString(text).Ceil()
		boxWidth = textWidth + paddingX*2
		boxHeight = int(fontSize) + paddingY*2
	}

	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	// 方框位置：右下角
	x0 := width - boxWidth - margin
	y0 := height - boxHeight - margin
	x1 := width - margin
	y1 := height - margin
	if x0 < 0 {
		x0 = margin
	}
	if y0 < 0 {
		y0 = margin
	}

	// 白底方框
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(dst, image.Rect(x0, y0, x1, y1), &image.Uniform{white}, image.Point{}, draw.Src)

	// 黑色文字
	black := color.RGBA{0, 0, 0, 255}
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(black),
		Face: face,
	}
	// 垂直居中
	baseline := y0 + paddingY + int(fontSize*0.85)
	d.Dot = fixed.P(x0+paddingX, baseline)
	d.DrawString(text)

	var out bytes.Buffer
	switch mediaType {
	case "image/jpeg":
		if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 92}); err != nil {
			return data, mediaType
		}
	case "image/png":
		if err := png.Encode(&out, dst); err != nil {
			return data, mediaType
		}
	default:
		if err := png.Encode(&out, dst); err != nil {
			return data, mediaType
		}
		mediaType = "image/png"
	}
	return out.Bytes(), mediaType
}
