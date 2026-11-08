package httpapi

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"qiantie/backend/internal/imagegenerationcache"
	"qiantie/backend/internal/localartifact"
)

type readyImageCache struct {
	entry  imagegenerationcache.Entry
	claims int
}

func (cache *readyImageCache) Claim(context.Context, string, string, string, time.Duration) (imagegenerationcache.Entry, string, error) {
	cache.claims++
	return cache.entry, imagegenerationcache.ClaimReady, nil
}
func (cache *readyImageCache) MarkReady(context.Context, string, string, string, string, imagegenerationcache.Timing) error {
	return nil
}
func (cache *readyImageCache) MarkFailed(context.Context, string, string, error) error { return nil }

func cacheTestPNG(t *testing.T) []byte {
	t.Helper()
	var body bytes.Buffer
	if err := png.Encode(&body, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

func TestImageGenerationFingerprintChangesWithEveryMaterialInput(t *testing.T) {
	base := shuihuoImageFingerprintInput{Owner: "alice", Endpoint: "https://one.example/v1/images/generations", ModelID: "gpt-image-2", Prompt: "prompt", AspectRatio: "9:16", References: []string{"ref-a"}, WatermarkEnabled: true, WatermarkText: "mark"}
	first := shuihuoImageGenerationFingerprint(base)
	changes := []shuihuoImageFingerprintInput{
		func() shuihuoImageFingerprintInput {
			value := base
			value.Endpoint = "https://two.example/v1/images/generations"
			return value
		}(),
		func() shuihuoImageFingerprintInput { value := base; value.ModelID = "other"; return value }(),
		func() shuihuoImageFingerprintInput { value := base; value.Prompt = "other"; return value }(),
		func() shuihuoImageFingerprintInput { value := base; value.AspectRatio = "1:1"; return value }(),
		func() shuihuoImageFingerprintInput { value := base; value.References = []string{"ref-b"}; return value }(),
		func() shuihuoImageFingerprintInput { value := base; value.WatermarkText = "other"; return value }(),
	}
	for index, changed := range changes {
		if got := shuihuoImageGenerationFingerprint(changed); got == first {
			t.Fatalf("change %d did not change fingerprint %q", index, got)
		}
	}
}

func TestGenerateShuihuoImageArtifactReusesReadyCacheWithoutProviderCall(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	saved, err := files.SaveImage("cached-image", "image/png", bytes.NewReader(cacheTestPNG(t)))
	if err != nil {
		t.Fatal(err)
	}
	cache := &readyImageCache{entry: imagegenerationcache.Entry{State: imagegenerationcache.StateReady, StorageRef: saved.StorageRef, MediaType: "image/png"}}
	ctx := withShuihuoImageGenerationCache(context.Background(), cache)
	target := shuihuoAssetImageTarget{Provider: "personal_api", ModelID: "gpt-image-2", AspectRatio: "1:1", CatalogModel: map[string]any{"endpoint": "://invalid", "credential": "secret"}}

	mediaType, storageRef, source, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "prompt", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "image/png" || storageRef != saved.StorageRef || source != "provider" || cache.claims != 1 {
		t.Fatalf("result=(%q,%q,%q) claims=%d", mediaType, storageRef, source, cache.claims)
	}
}

func TestExplicitRegenerationBypassesReadyCache(t *testing.T) {
	cache := &readyImageCache{entry: imagegenerationcache.Entry{State: imagegenerationcache.StateReady, StorageRef: "cached.png", MediaType: "image/png"}}
	ctx := withShuihuoImageGenerationCacheBypass(withShuihuoImageGenerationCache(context.Background(), cache))
	target := shuihuoAssetImageTarget{Provider: "personal_api", ModelID: "gpt-image-2", AspectRatio: "1:1", CatalogModel: map[string]any{"endpoint": "://invalid", "credential": "secret"}}
	_, _, _, err := generateShuihuoImageArtifact(ctx, nil, localartifact.NewStore(t.TempDir(), 1<<20), nil, "alice", target, "prompt", nil, nil)
	if err == nil {
		t.Fatal("explicit regeneration unexpectedly reused cache")
	}
	if cache.claims != 0 {
		t.Fatalf("cache claims=%d, want 0", cache.claims)
	}
}

var _ = color.RGBA{}
