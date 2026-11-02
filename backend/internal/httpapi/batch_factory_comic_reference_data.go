package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/prompt"
)

const maxComicReferenceImageBytes = 16 << 20
const compactComicReferenceImageBytes = 1 << 20

func resolveComicReferenceImageData(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, catalog *prompt.Catalog, owner string, references []string) ([]string, error) {
	resolved := make([]string, 0, len(references))
	for _, reference := range references {
		value := strings.TrimSpace(reference)
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "data:image/") {
			resolved = append(resolved, value)
			continue
		}
		parsed, err := url.Parse(value)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && !isLocalComicReferenceURL(value) {
			resolved = append(resolved, value)
			continue
		}

		payload, mediaType, err := openComicReferenceImage(ctx, store, files, catalog, owner, value)
		if err != nil {
			return nil, err
		}
		payload, mediaType, err = compactComicReferenceImage(payload, mediaType)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, "data:"+mediaType+";base64,"+base64.StdEncoding.EncodeToString(payload))
	}
	return resolved, nil
}

func isLocalComicReferenceURL(value string) bool {
	parts := comicReferenceURLParts(value)
	if len(parts) == 0 || parts[0] != "api" {
		return false
	}
	if len(parts) == 5 && parts[1] == "local-executor" && parts[2] == "v1" {
		switch parts[3] {
		case "batch-comic-style-previews", "batch-comic-reference-scenes", "batch-comic-materials":
			return true
		}
	}
	if len(parts) == 6 && parts[1] == "local-executor" && parts[2] == "v1" && parts[3] == "batch-comic-references" {
		return true
	}
	if len(parts) == 8 && parts[1] == "local-executor" && parts[2] == "v1" && parts[3] == "batch-book-asset-images" {
		return true
	}
	return len(parts) == 5 && parts[1] == "presets" && (parts[2] == "comic-style" || parts[2] == "comic-reference-scene" || parts[2] == "comic-material") && parts[4] == "preview"
}

func compactComicReferenceImage(payload []byte, mediaType string) ([]byte, string, error) {
	if len(payload) <= compactComicReferenceImageBytes || (mediaType != "image/png" && mediaType != "image/jpeg") {
		return payload, mediaType, nil
	}

	source, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return nil, "", fmt.Errorf("%w: 无法压缩漫画参考图: %v", batchfactoryv11.ErrInvalid, err)
	}
	bounds := source.Bounds()
	flattened := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(flattened, flattened.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(flattened, flattened.Bounds(), source, bounds.Min, draw.Over)

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, flattened, &jpeg.Options{Quality: 82}); err != nil {
		return nil, "", fmt.Errorf("%w: 无法压缩漫画参考图: %v", batchfactoryv11.ErrInvalid, err)
	}
	return encoded.Bytes(), "image/jpeg", nil
}

func openComicReferenceImage(ctx context.Context, store batchfactoryv11.Store, files *localartifact.Store, catalog *prompt.Catalog, owner, reference string) ([]byte, string, error) {
	parts := comicReferenceURLParts(reference)
	switch {
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "presets" && parts[4] == "preview" && parts[2] == "comic-style":
		payload, mediaType, ok := prompt.OpenComicStylePreview(parts[3])
		if !ok {
			return nil, "", fmt.Errorf("%w: 漫画画风参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return payload, mediaType, nil
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "presets" && parts[4] == "preview" && parts[2] == "comic-reference-scene":
		payload, mediaType, ok := prompt.OpenComicReferenceScene(parts[3])
		if !ok {
			return nil, "", fmt.Errorf("%w: 漫画场景参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return payload, mediaType, nil
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "presets" && parts[4] == "preview" && parts[2] == "comic-material":
		if catalog == nil {
			return nil, "", fmt.Errorf("%w: 漫画画风素材库不可用", batchfactoryv11.ErrUnavailable)
		}
		assetID, found, err := catalog.PublishedComicMaterialAsset(parts[3])
		if err != nil {
			return nil, "", err
		}
		if !found {
			return nil, "", fmt.Errorf("%w: 漫画画风参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return readComicReferenceArtifact(files, assetID, "")
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-style-previews":
		payload, mediaType, ok := prompt.OpenComicStylePreview(parts[4])
		if !ok {
			return nil, "", fmt.Errorf("%w: 漫画画风参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return payload, mediaType, nil
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-reference-scenes":
		payload, mediaType, ok := prompt.OpenComicReferenceScene(parts[4])
		if !ok {
			return nil, "", fmt.Errorf("%w: 漫画场景参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return payload, mediaType, nil
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-materials":
		if catalog == nil {
			return nil, "", fmt.Errorf("%w: 漫画画风素材库不可用", batchfactoryv11.ErrUnavailable)
		}
		assetID, found, err := catalog.PublishedComicMaterialAsset(parts[4])
		if err != nil {
			return nil, "", err
		}
		if !found {
			return nil, "", fmt.Errorf("%w: 漫画画风参考图不存在", batchfactoryv11.ErrNotFound)
		}
		return readComicReferenceArtifact(files, assetID, "")
	case len(parts) == 6 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-comic-references":
		if store == nil {
			return nil, "", fmt.Errorf("%w: 漫画参考图存储不可用", batchfactoryv11.ErrUnavailable)
		}
		if _, err := store.GetBatch(ctx, owner, parts[4]); err != nil {
			return nil, "", err
		}
		return readComicReferenceArtifact(files, parts[5], "")
	case len(parts) == 8 && parts[0] == "api" && parts[1] == "local-executor" && parts[3] == "batch-book-asset-images":
		if store == nil {
			return nil, "", fmt.Errorf("%w: 漫画资产图存储不可用", batchfactoryv11.ErrUnavailable)
		}
		images, err := store.ListBookAssetImages(ctx, owner, parts[4], parts[5], parts[6])
		if err != nil {
			return nil, "", err
		}
		for _, image := range images {
			if image.ID == parts[7] && strings.TrimSpace(image.StorageRef) != "" {
				return readComicReferenceArtifact(files, image.StorageRef, image.MediaType)
			}
		}
		return nil, "", fmt.Errorf("%w: 漫画人物资产图不存在", batchfactoryv11.ErrNotFound)
	default:
		return nil, "", fmt.Errorf("%w: 无法读取漫画参考图 %q", batchfactoryv11.ErrInvalid, reference)
	}
}

func readComicReferenceArtifact(files *localartifact.Store, idOrStorageRef, mediaType string) ([]byte, string, error) {
	if files == nil {
		return nil, "", fmt.Errorf("%w: 漫画参考图存储不可用", batchfactoryv11.ErrUnavailable)
	}
	candidates := []string{strings.TrimSpace(idOrStorageRef)}
	if !strings.Contains(candidates[0], ".") {
		candidates = []string{candidates[0] + ".png", candidates[0] + ".jpg", candidates[0] + ".webp"}
	}
	for _, candidate := range candidates {
		file, err := files.Open(candidate)
		if err != nil {
			continue
		}
		payload, readErr := io.ReadAll(io.LimitReader(file, maxComicReferenceImageBytes+1))
		_ = file.Close()
		if readErr != nil {
			return nil, "", readErr
		}
		if len(payload) > maxComicReferenceImageBytes {
			return nil, "", fmt.Errorf("%w: 漫画参考图超过 16MB", batchfactoryv11.ErrInvalid)
		}
		if strings.TrimSpace(mediaType) == "" {
			mediaType = http.DetectContentType(payload)
		}
		if mediaType != "image/png" && mediaType != "image/jpeg" && mediaType != "image/webp" {
			return nil, "", fmt.Errorf("%w: 漫画参考图格式不支持", batchfactoryv11.ErrInvalid)
		}
		return payload, mediaType, nil
	}
	return nil, "", fmt.Errorf("%w: 漫画参考图文件不存在", batchfactoryv11.ErrNotFound)
}
