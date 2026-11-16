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
	_ "golang.org/x/image/webp"
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

// n8UploadAcceptedTypes 是 N8 允许直接上传的图片格式。手动换图允许上传
// WebP，但 N8 不接受 WebP，因此这些格式必须在发布前转码，否则会被判
// 为“无效图片”。
var n8UploadAcceptedTypes = map[string]string{
	"image/jpeg": "image/jpeg",
	"image/jpg":  "image/jpeg",
	"image/gif":  "image/gif",
	"image/bmp":  "image/bmp",
	"image/png":  "image/png",
}

// webpTranscodeQuality 是 WebP 转 JPEG 的质量；漫画页以线条和文字为主，
// 这里用较高画质，避免细小文字发糊。
const webpTranscodeQuality = 92

// compressImageForUpload 保证返回的内容是 N8 接受的格式，并把超过体积上限
// 的图片重编码成 JPEG。手动换图允许上传 WebP，但 N8 不接受 WebP，所以这类
// 图片必须在发布前转码；GIF 动图原样返回，避免改变原有语义。
func compressImageForUpload(content []byte, mediaType string) ([]byte, string, error) {
	lowered := strings.ToLower(strings.TrimSpace(mediaType))
	if lowered == "image/gif" {
		return content, "image/gif", nil
	}
	declared, declaredOK := n8UploadAcceptedTypes[lowered]
	withinLimit := len(content) <= n8ImageMaxBytes()
	if declaredOK && withinLimit {
		return content, declared, nil
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || config.Width < 1 || config.Height < 1 {
		if declaredOK {
			// BMP 等格式由 imageDimensions 单独解析尺寸，这里保持原样。
			return content, declared, nil
		}
		return content, mediaType, fmt.Errorf("图片内容无法识别，无法转码为 N8 支持的格式（声明格式 %s）", fallbackMediaType(mediaType))
	}
	if config.Width*config.Height > n8ImageMaxPixels {
		return content, mediaType, nil
	}
	if real, ok := n8UploadAcceptedTypes["image/"+strings.ToLower(format)]; ok && real != "image/gif" && withinLimit && !declaredOK {
		// 声明格式缺失或写错，但内容本身已经是 N8 支持的格式：按真实格式上传，
		// 不做无谓的重编码，避免画质损失。
		log.Printf("n8 image media type corrected %s -> %s", fallbackMediaType(mediaType), real)
		return content, real, nil
	}
	source, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		if declaredOK {
			return content, declared, nil
		}
		return content, mediaType, fmt.Errorf("图片格式 %s 无法转码为 JPEG", fallbackMediaType(mediaType))
	}
	if !declaredOK && withinLimit {
		// 体积未超限，只需要换格式，不需要缩边降质。
		var buffer bytes.Buffer
		if err := jpeg.Encode(&buffer, source, &jpeg.Options{Quality: webpTranscodeQuality}); err != nil {
			return content, mediaType, fmt.Errorf("图片转码失败：%w", err)
		}
		log.Printf("n8 image transcoded %s %dKB -> jpeg %dKB %dx%d", fallbackMediaType(mediaType), len(content)/1024, buffer.Len()/1024, config.Width, config.Height)
		return buffer.Bytes(), "image/jpeg", nil
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

// fallbackMediaType 用于拼错误文案：声明格式为空时用「未知格式」兜底，
// 不让用户看到一句没有主语的中文。
func fallbackMediaType(mediaType string) string {
	trimmed := strings.TrimSpace(mediaType)
	if trimmed == "" {
		return "未知格式"
	}
	return trimmed
}
