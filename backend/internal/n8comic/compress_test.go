package n8comic

import (
	"bytes"
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
