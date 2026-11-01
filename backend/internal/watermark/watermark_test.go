package watermark

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestApplyAddsWatermark(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			src.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}

	out, mediaType := Apply(buf.Bytes(), "image/png", Options{Enabled: true, Text: "热门小说"})
	if mediaType != "image/png" {
		t.Fatalf("mediaType=%q", mediaType)
	}
	if len(out) == len(buf.Bytes()) {
		t.Fatal("output should differ after watermark")
	}

	decoded, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 800 || decoded.Bounds().Dy() != 600 {
		t.Fatalf("size changed: %v", decoded.Bounds())
	}
}

func TestApplyDisabledReturnsOriginal(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, mt := Apply(buf.Bytes(), "image/png", Options{Enabled: false, Text: "test"})
	if mt != "image/png" || len(out) != len(buf.Bytes()) {
		t.Fatal("disabled should return original")
	}
}

func TestApplySmallImageReturnsOriginal(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 50, 50))
	var buf bytes.Buffer
	if err := png.Encode(&buf, src); err != nil {
		t.Fatal(err)
	}
	out, mt := Apply(buf.Bytes(), "image/png", Options{Enabled: true, Text: "热门小说"})
	if mt != "image/png" || len(out) != len(buf.Bytes()) {
		t.Fatal("small image should return original")
	}
}
