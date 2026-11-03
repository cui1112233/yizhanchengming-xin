package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

// registerConfiguredImageGenerationRoute keeps the legacy Batch Factory
// image button working while the newer workbench uses book-scoped asset
// routes. Both paths read the same account model catalog.
func registerConfiguredImageGenerationRoute(mux *http.ServeMux, store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store) {
	if mux == nil || store == nil {
		return
	}
	mux.HandleFunc("POST /api/batch-factory/v12/image-generation", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Prompt       string `json:"prompt"`
			ImageModelID string `json:"imageModelId"`
			Size         string `json:"size"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		result, status, err := configuredBatchImage(r.Context(), store, catalogStore, owner, input.Prompt, input.ImageModelID, input.Size)
		if err != nil {
			writeJSON(w, status, map[string]string{"error": err.Error(), "code": errorCodeOr(err, "BFV11_IMAGE_GENERATION_FAILED")})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
}

func configuredBatchImage(ctx context.Context, store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, owner, prompt, imageModelID, size string) (map[string]any, int, error) {
	results, status, err := configuredBatchImages(ctx, store, catalogStore, owner, AgentImageGenerationInput{
		ModelID: imageModelID, Prompt: prompt, Size: size,
	})
	if err != nil {
		return nil, status, err
	}
	if len(results) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型未返回图片结果")
	}
	return map[string]any{"ok": true, "imageUrl": results[0]}, http.StatusOK, nil
}

func configuredBatchImages(ctx context.Context, store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, owner string, input AgentImageGenerationInput) ([]string, int, error) {
	prompt := strings.TrimSpace(input.Prompt)
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("图片生成提示词不能为空")
	}
	config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStore)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	var model map[string]any
	for _, candidate := range shuihuoCollection(config, "modelCatalog", nil) {
		if !strings.EqualFold(modelString(candidate, "kind"), "image") {
			continue
		}
		if input.ModelID != "" && modelString(candidate, "id") != input.ModelID && modelString(candidate, "modelId") != input.ModelID {
			continue
		}
		if enabled, exists := candidate["enabled"].(bool); exists && !enabled {
			continue
		}
		model = candidate
		break
	}
	if model == nil {
		return nil, http.StatusServiceUnavailable, withErrorCode(fmt.Errorf("图片模型未配置；%s", setupHintImage), ErrorCodeImageModelNotConfigured)
	}
	endpointRaw := modelString(model, "endpoint")
	if endpointRaw == "" {
		endpointRaw = modelString(model, "baseUrl")
	}
	endpoint, err := shuihuoImageGenerationEndpoint(endpointRaw)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("图片模型生成接口格式不正确")
	}
	credential := shuihuoModelCredential(model)
	if credential == "" {
		return nil, http.StatusServiceUnavailable, withErrorCode(fmt.Errorf("图片模型凭据未配置；%s", setupHintImage), ErrorCodeImageModelNotConfigured)
	}
	payload := map[string]any{"model": modelString(model, "modelId"), "prompt": prompt}
	if input.Count > 0 {
		payload["n"] = input.Count
	}
	if strings.TrimSpace(input.Quality) != "" {
		payload["quality"] = input.Quality
	}
	if strings.TrimSpace(input.Size) != "" {
		payload["size"] = strings.TrimSpace(input.Size)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型请求创建失败")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+credential)
	response, err := (&http.Client{Timeout: 120 * time.Second}).Do(request)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型请求失败: %v", err)
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if readErr != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型请求失败（HTTP %d）", response.StatusCode)
	}
	var decoded struct {
		Data []struct {
			URL     string `json:"url"`
			Image   string `json:"image_url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded.Data) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型返回了无法识别的响应")
	}
	imageURLs := make([]string, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		imageURL := strings.TrimSpace(item.URL)
		if imageURL == "" {
			imageURL = strings.TrimSpace(item.Image)
		}
		if imageURL == "" && strings.TrimSpace(item.B64JSON) != "" {
			encoded := strings.TrimSpace(item.B64JSON)
			decodedBytes, decodeErr := base64.StdEncoding.DecodeString(encoded)
			if decodeErr != nil || len(decodedBytes) == 0 {
				return nil, http.StatusBadGateway, fmt.Errorf("图片模型返回了无效图片数据")
			}
			imageURL = "data:" + detectedShuihuoImageMediaType(decodedBytes) + ";base64," + encoded
		}
		if imageURL != "" {
			imageURLs = append(imageURLs, imageURL)
		}
	}
	if len(imageURLs) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型未返回图片结果")
	}
	if input.Count > 0 && len(imageURLs) < input.Count {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型只返回了 %d 张图片，少于请求的 %d 张", len(imageURLs), input.Count)
	}
	if input.Count > 0 && len(imageURLs) > input.Count {
		imageURLs = imageURLs[:input.Count]
	}
	return imageURLs, http.StatusOK, nil
}
