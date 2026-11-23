package httpapi

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// A valid PNG header alone must never become a locked/ready page.
func TestBatchFactoryTextUploadReadsCompleteImage(t *testing.T) {
	var full bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 30, 50))
	img.Set(1, 1, color.White)
	if err := png.Encode(&full, img); err != nil {
		t.Fatal(err)
	}
	good := full.Bytes()
	if mime, err := validateTextImage(good); err != nil || mime != "image/png" {
		t.Fatalf("normal PNG: %s, %v", mime, err)
	}
	partial := good[:len(good)/2]
	if _, _, err := image.DecodeConfig(bytes.NewReader(partial)); err != nil {
		t.Fatal("test needs readable dimensions", err)
	}
	if _, err := validateTextImage(partial); err == nil {
		t.Fatal("partial image accepted")
	}
}
