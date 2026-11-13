package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/imagecapacity"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/safehttp"
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

func configuredBatchImages(ctx context.Context, store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, owner string, input AgentImageGenerationInput) (imageResultsOut []string, imageStatusOut int, imageReturnErr error) {
	ctx = imageRequestIdentity(ctx, owner, input.ProjectID)
	ctx, finishImage := beginImageRequestOperation(ctx)
	defer func() { finishImage(imageReturnErr) }()
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
	// 旧批量出图与 Agent 出图此前绕过出图信号量直接打上游，一个用户连续点击
	// 就能把整机图片并发打满。这里复用漫画出图的同一套进程/任务名额；拿不到
	// 名额按“队列繁忙”返回，由调用方稍后重试，绝不静默直连。
	queueCtx, cancelQueue := context.WithTimeout(ctx, shuihuoImageQueueWaitTimeout)
	releaseImageSlot, slotErr := acquireShuihuoGPTImageGenerationSlot(queueCtx, modelString(model, "modelId"))
	cancelQueue()
	if slotErr != nil {
		return nil, http.StatusServiceUnavailable, withErrorCode(shuihuoImageQueueError(ctx, slotErr), automationCodeImageQueueBusy)
	}
	defer releaseImageSlot()
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
	response, err := safehttp.NewClientWithTimeout(120 * time.Second).Do(request)
	if err != nil {
		// 传输层错误不能把 “Post ...: dial tcp ...” 直接抛给界面：主文案只给
		// 中文结论，原始错误放到「上游原文」后面给展开区查看。
		message := "图片通道网络不可达，请检查网络或代理设置"
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netErr) && netErr.Timeout() {
			message = "图片通道连接超时，请稍后重试"
		}
		return nil, http.StatusBadGateway, fmt.Errorf("%s", appendShuihuoUpstreamRaw(message, err.Error()))
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, http.StatusBadGateway, shuihuoImageUpstreamError{
			Message:    shuihuoImageStatusRejectionMessage(response.StatusCode, string(raw)),
			Raw:        string(raw),
			StatusCode: response.StatusCode,
		}
	}
	if readErr != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("%s", appendShuihuoUpstreamRaw("图片通道响应读取中断，请稍后重试", readErr.Error()))
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
			if decodeErr != nil || len(decodedBytes) == 0 || (imagecapacity.Enabled() && validateCapacityImageBytes(decodedBytes) != nil) {
				return nil, http.StatusBadGateway, fmt.Errorf("图片模型返回了无效图片数据")
			}
			imageURL = "data:" + detectedShuihuoImageMediaType(decodedBytes) + ";base64," + encoded
		}
		if imageURL != "" {
			if err := validateCapacityImageURL(imageURL); err != nil {
				return nil, http.StatusBadGateway, err
			}
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
	confirmImageRequestResult(ctx)
	return imageURLs, http.StatusOK, nil
}
