package httpapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

func TestProbeAICutAssetDurationReadsRealMediaDuration(t *testing.T) {
	if _, err := ffmpegbin.ResolveFFmpeg(); err != nil {
		t.Skip(err)
	}
	if _, err := ffmpegbin.ResolveFFprobe(); err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	media := filepath.Join(dir, "sample.mp4")
	if err := ffmpegbin.Run(t.Context(), "ffmpeg", "-hide_banner", "-y", "-f", "lavfi", "-i", "color=c=black:s=64x64:d=1.2", "-c:v", "libx264", "-pix_fmt", "yuv420p", media); err != nil {
		t.Fatal(err)
	}
	objects := agentworkspace.NewLocalAssetObjectStore(filepath.Join(dir, "objects"))
	asset := aicut.Asset{ID: "asset-1", ProjectID: "cut-1", Kind: "video", Name: "sample.mp4", ContentType: "video/mp4", ObjectKey: "aicut-assets/cut-1/sample.mp4"}
	file, err := os.Open(media)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := objects.Put(t.Context(), aicutObjectAsset(asset), file); err != nil {
		t.Fatal(err)
	}
	got, err := probeAICutAssetDuration(context.Background(), asset, objects)
	if err != nil {
		t.Fatal(err)
	}
	if got < 1100 || got > 1300 {
		t.Fatalf("duration=%d", got)
	}
}
