package n8comic

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"testing"
)

func noisyPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rng := rand.New(rand.NewSource(7))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256)), 255})
		}
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatalf("encode sample: %v", err)
	}
	return buffer.Bytes()
}

func TestCompressImageForUploadShrinksOversizedPage(t *testing.T) {
	content := noisyPNG(t, 1400, 2100)
	if len(content) <= n8ImageMaxBytes() {
		t.Fatalf("sample size=%d, want above %d", len(content), n8ImageMaxBytes())
	}
	output, mediaType, err := compressImageForUpload(content, "image/png")
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if mediaType != "image/jpeg" {
		t.Fatalf("mediaType=%q, want image/jpeg", mediaType)
	}
	if len(output) > n8ImageMaxBytes() {
		t.Fatalf("compressed size=%d, want <=%d", len(output), n8ImageMaxBytes())
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(output))
	if err != nil || format != "jpeg" {
		t.Fatalf("decode compressed format=%q err=%v", format, err)
	}
	if config.Width > 1400 || config.Height > 2100 {
		t.Fatalf("compressed %dx%d larger than original", config.Width, config.Height)
	}
	ratio := float64(config.Width) / float64(config.Height)
	if math.Abs(ratio-float64(1400)/float64(2100)) > 0.01 {
		t.Fatalf("aspect ratio changed: %dx%d", config.Width, config.Height)
	}
}

func TestCompressImageForUploadKeepsSmallImage(t *testing.T) {
	content := noisyPNG(t, 64, 96)
	output, mediaType, err := compressImageForUpload(content, "image/png")
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if mediaType != "image/png" || !bytes.Equal(output, content) {
		t.Fatalf("small image should stay untouched, mediaType=%q size=%d", mediaType, len(output))
	}
}

func TestCompressImageForUploadKeepsGIF(t *testing.T) {
	content := append([]byte(nil), make([]byte, n8ImageMaxBytes()+1024)...)
	output, mediaType, err := compressImageForUpload(content, "image/gif")
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if mediaType != "image/gif" || len(output) != len(content) {
		t.Fatalf("gif should stay untouched, mediaType=%q size=%d", mediaType, len(output))
	}
}

// sampleWebP 是一张 64x64 的真实 WebP。手动换图允许上传 WebP，但 N8 不接受，
// 发布前必须转码，否则会出现“xxx.png 无效：N8 只接受 JPG…”的报错。
const sampleWebP = "UklGRlQAAABXRUJQVlA4IEgAAADwAwCdASpAAEAAPpFIoEwlpCMiIggAsBIJaQDTUoAAEmwFeIXGAXEAAP7paR//jePAryXZX//G83mtXXZs2wBgVPTL84yAAAA="

func TestCompressImageForUploadTranscodesWebP(t *testing.T) {
	content, err := base64.StdEncoding.DecodeString(sampleWebP)
	if err != nil {
		t.Fatalf("decode sample: %v", err)
	}
	output, mediaType, err := compressImageForUpload(content, "image/webp")
	if err != nil {
		t.Fatalf("compress webp: %v", err)
	}
	if mediaType != "image/jpeg" {
		t.Fatalf("mediaType=%q, want image/jpeg", mediaType)
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(output))
	if err != nil || format != "jpeg" {
		t.Fatalf("transcoded format=%q err=%v", format, err)
	}
	if _, _, err := imageDimensions(output, mediaType); err != nil {
		t.Fatalf("transcoded image must pass N8 validation: %v", err)
	}
}

func TestCompressImageForUploadCorrectsMissingMediaType(t *testing.T) {
	content := noisyPNG(t, 64, 96)
	output, mediaType, err := compressImageForUpload(content, "")
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if mediaType != "image/png" || !bytes.Equal(output, content) {
		t.Fatalf("real png should keep bytes and use image/png, mediaType=%q", mediaType)
	}
}

func TestCompressImageForUploadRejectsUnknownPayload(t *testing.T) {
	if _, _, err := compressImageForUpload([]byte("not an image at all"), "image/webp"); err == nil {
		t.Fatal("unknown payload with unsupported media type must fail instead of being named .png")
	}
}
