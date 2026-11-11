package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"

	_ "golang.org/x/image/webp"
)

const aicutBackgroundRemovalMaxBytes = 20 << 20

type aicutBackgroundRemover func(context.Context, string, aicut.Asset, []byte) ([]byte, string, error)

type aicutBackgroundRemovalRequest struct {
	Name string `json:"name,omitempty"`
}

// RegisterAICutBackgroundRemovalRoutes creates a new transparent project image;
// it never mutates or replaces the user's original asset.
func RegisterAICutBackgroundRemovalRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, catalog novelfetchworkshop.Store, remover aicutBackgroundRemover, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/assets/{assetId}/remove-background", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		owner := strings.TrimSpace(identity.Username)
		projectID, assetID := strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("assetId"))
		if _, err := store.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		asset, err := store.Asset(req.Context(), owner, projectID, assetID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "image" || !strings.HasPrefix(strings.ToLower(asset.ContentType), "image/") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "背景移除目前仅支持图片素材"})
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		var input aicutBackgroundRemovalRequest
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		body, err := objects.Open(req.Context(), aicutObjectAsset(asset))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "原始图片文件不存在"})
			return
		}
		source, readErr := io.ReadAll(io.LimitReader(body, aicutBackgroundRemovalMaxBytes+1))
		_ = body.Close()
		if readErr != nil || len(source) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "读取原始图片失败"})
			return
		}
		if len(source) > aicutBackgroundRemovalMaxBytes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "背景移除图片不能超过 20 MB"})
			return
		}
		activeRemover := remover
		if activeRemover == nil {
			activeRemover = func(ctx context.Context, owner string, asset aicut.Asset, source []byte) ([]byte, string, error) {
				return requestConfiguredAICutBackgroundRemoval(ctx, owner, asset, source, catalog)
			}
		}
		result, _, err := activeRemover(req.Context(), owner, asset, source)
		if err != nil {
			if req.Context().Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "背景移除已取消"})
				return
			}
			status := http.StatusBadGateway
			if errors.Is(err, batchfactoryv11.ErrUnavailable) || errors.Is(err, batchfactoryv11.ErrInvalid) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": "背景移除失败：" + err.Error()})
			return
		}
		transparentPNG, err := normalizeAICutTransparentPNG(result)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "背景移除失败：" + err.Error()})
			return
		}
		createdAt := now().UTC()
		createdID := newAgentID("aicut-asset")
		objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, createdID, ".png")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "透明素材地址创建失败"})
			return
		}
		created := aicut.Asset{ID: createdID, ProjectID: projectID, Kind: "image", Name: aicutBackgroundRemovalAssetName(input.Name, asset.Name, createdAt), ContentType: "image/png", ObjectKey: strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1), ByteSize: int64(len(transparentPNG)), CreatedAt: createdAt}
		if err := objects.Put(req.Context(), aicutObjectAsset(created), bytes.NewReader(transparentPNG)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "透明素材保存失败"})
			return
		}
		if err := store.CreateAsset(req.Context(), owner, created); err != nil {
			_ = objects.Delete(req.Context(), aicutObjectAsset(created))
			writeAICutError(w, err)
			return
		}
		created.URL = aicutSignedAssetContentURL(auth, owner, projectID, created.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": created})
	})))
}

func requestConfiguredAICutBackgroundRemoval(ctx context.Context, owner string, asset aicut.Asset, source []byte, catalog novelfetchworkshop.Store) ([]byte, string, error) {
	if catalog == nil {
		return nil, "", fmt.Errorf("%w: 未配置支持图片编辑的模型", batchfactoryv11.ErrUnavailable)
	}
	config, err := catalog.GetConfig(ctx, owner)
	if err != nil {
		return nil, "", fmt.Errorf("%w: 无法读取图片模型配置", batchfactoryv11.ErrUnavailable)
	}
	var selected map[string]any
	for _, model := range modelCatalog(config) {
		if !strings.EqualFold(modelString(model, "kind"), "image") || model["enabled"] != true || !shuihuoImageModelSupportsEdits(model, modelString(model, "modelId")) {
			continue
		}
		if firstText(model["endpoint"], model["baseUrl"]) == "" || shuihuoModelCredential(model) == "" || modelString(model, "modelId") == "" {
			continue
		}
		selected = model
		break
	}
	if selected == nil {
		return nil, "", fmt.Errorf("%w: 未配置支持图片编辑的模型", batchfactoryv11.ErrUnavailable)
	}
	endpoint, err := shuihuoImageGenerationEndpoint(firstText(selected["endpoint"], selected["baseUrl"]))
	if err != nil {
		return nil, "", fmt.Errorf("%w: 图片模型地址无效", batchfactoryv11.ErrInvalid)
	}
	mediaType := strings.TrimSpace(strings.Split(asset.ContentType, ";")[0])
	if mediaType == "" {
		mediaType = detectedShuihuoImageMediaType(source)
	}
	reference := "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(source)
	ratio := aicutImageAspectRatio(source)
	req, err := newShuihuoImageEditRequestWithResponseFormat(ctx, endpoint, modelString(selected, "modelId"), prompt.AICutBackgroundRemoval(), shuihuoImageSizeForModel(modelString(selected, "modelId"), ratio), []string{reference}, "b64_json")
	if err != nil {
		return nil, "", fmt.Errorf("%w: 无法创建图片编辑请求", batchfactoryv11.ErrUpstream)
	}
	credential := shuihuoModelCredential(selected)
	req.Header.Set("Authorization", "Bearer "+credential)
	response, err := shuihuoImageGenerationHTTPClient(endpoint).Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: 图片编辑服务不可用", batchfactoryv11.ErrUpstream)
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 24<<20))
	if readErr != nil {
		return nil, "", fmt.Errorf("%w: 读取图片编辑结果失败", batchfactoryv11.ErrUpstream)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", fmt.Errorf("%w: 图片编辑服务返回 HTTP %d%s", batchfactoryv11.ErrUpstream, response.StatusCode, shuihuoImageProviderErrorDetail(raw))
	}
	var result struct {
		Data []struct {
			B64JSON string `json:"b64_json"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &result) != nil || len(result.Data) == 0 {
		return nil, "", fmt.Errorf("%w: 图片编辑服务没有返回图片", batchfactoryv11.ErrUpstream)
	}
	resultType, payload, err := decodeShuihuoGeneratedImage(ctx, shuihuoImageGenerationHTTPClient(endpoint), result.Data[0].B64JSON, result.Data[0].URL, credential, endpoint)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", batchfactoryv11.ErrUpstream, err)
	}
	return payload, resultType, nil
}

func normalizeAICutTransparentPNG(payload []byte) ([]byte, error) {
	decoded, _, err := image.Decode(bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("服务没有返回有效图片")
	}
	if !imageHasTransparency(decoded) {
		return nil, fmt.Errorf("服务没有返回透明背景")
	}
	var output bytes.Buffer
	if err := png.Encode(&output, decoded); err != nil {
		return nil, fmt.Errorf("透明图片编码失败")
	}
	return output.Bytes(), nil
}

func aicutImageHasTransparency(payload []byte) bool {
	decoded, _, err := image.Decode(bytes.NewReader(payload))
	return err == nil && imageHasTransparency(decoded)
}

func imageHasTransparency(decoded image.Image) bool {
	bounds := decoded.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := decoded.At(x, y).RGBA()
			if alpha < 0xffff {
				return true
			}
		}
	}
	return false
}

func aicutImageAspectRatio(payload []byte) string {
	config, _, err := image.DecodeConfig(bytes.NewReader(payload))
	if err != nil || config.Width <= 0 || config.Height <= 0 {
		return "1:1"
	}
	ratio := float64(config.Width) / float64(config.Height)
	if ratio >= 1.25 {
		return "16:9"
	}
	if ratio <= .8 {
		return "9:16"
	}
	return "1:1"
}

func aicutBackgroundRemovalAssetName(requested, source string, now time.Time) string {
	name := strings.TrimSpace(requested)
	if name == "" {
		name = strings.TrimSpace(strings.TrimSuffix(source, path.Ext(source))) + "-透明背景"
	}
	name = strings.TrimSuffix(name, path.Ext(name))
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return '-'
		}
		return r
	}, name)
	runes := []rune(strings.Trim(name, " .-"))
	if len(runes) > 64 {
		runes = runes[:64]
	}
	if len(runes) == 0 {
		return "透明背景-" + now.Format("20060102-150405") + ".png"
	}
	return string(runes) + ".png"
}
