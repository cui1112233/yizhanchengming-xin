package n8comic

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const (
	// defaultN8ImageMaxBytes 是 N8 / 巨量素材对单张图片的体积上限。2K 漫画页
	// 原始 PNG 通常在 5–7MB，直接上传会被上游判成“图片过大”。
	defaultN8ImageMaxBytes = 2 << 20
	// n8ImageMaxPixels 防解压炸弹：单张原图解码后不超过 4000 万像素。
	n8ImageMaxPixels = 40_000_000
)

// n8UploadLadder 是压缩阶梯：先保尺寸降质量，再小幅缩边。命中第一个
// 不超过体积上限的结果就返回，避免每页都做完整枚举。
var n8UploadLadder = []struct {
	scale   float64
	quality int
}{
	{1, 85},
	{1, 72},
	{0.85, 82},
	{0.85, 72},
	{0.7, 82},
}

func n8ImageMaxBytes() int {
	raw := strings.TrimSpace(os.Getenv("QIANTIE_N8_IMAGE_MAX_BYTES"))
	if raw == "" {
		return defaultN8ImageMaxBytes
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 256*1024 || value > 32<<20 {
		return defaultN8ImageMaxBytes
	}
	return value
}

// compressImageForUpload 把超过体积上限的图片重编码成 JPEG。未超过上限、
// 无法解码（如 BMP）或 GIF 动图一律原样返回，避免改变原有语义。
func compressImageForUpload(content []byte, mediaType string) ([]byte, string, error) {
	if len(content) <= n8ImageMaxBytes() {
		return content, mediaType, nil
	}
	if strings.EqualFold(mediaType, "image/gif") {
		return content, mediaType, nil
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width < 1 || config.Height < 1 {
		return content, mediaType, nil
	}
	if config.Width*config.Height > n8ImageMaxPixels {
		return content, mediaType, nil
	}
	source, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return content, mediaType, nil
	}
	for _, step := range n8UploadLadder {
		width := int(float64(config.Width)*step.scale + 0.5)
		height := int(float64(config.Height)*step.scale + 0.5)
		if width < 1 {
			width = 1
		}
		if height < 1 {
			height = 1
		}
		frame := source
		if width != config.Width || height != config.Height {
			scaled := image.NewRGBA(image.Rect(0, 0, width, height))
			xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), source, source.Bounds(), xdraw.Src, nil)
			frame = scaled
		}
		var buffer bytes.Buffer
		if err := jpeg.Encode(&buffer, frame, &jpeg.Options{Quality: step.quality}); err != nil {
			return content, mediaType, fmt.Errorf("压缩图片失败：%w", err)
		}
		if buffer.Len() <= n8ImageMaxBytes() {
			log.Printf("n8 image compressed %s %dKB -> %dKB %dx%d q%d", mediaType, len(content)/1024, buffer.Len()/1024, width, height, step.quality)
			return buffer.Bytes(), "image/jpeg", nil
		}
	}
	return content, mediaType, nil
}
