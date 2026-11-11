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
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/safehttp"
)

// 模型下拉里的渠道候选项使用固定前缀：channel:<预置渠道> / custom:<自定义渠道>。
// 执行器看到前缀时回读 apiIntegrations，而不是去 modelCatalog 里找。
const (
	integrationModelPrefixChannel = "channel:"
	integrationModelPrefixCustom  = "custom:"
)

type integrationChannelTarget struct {
	Kind        string
	Label       string
	BaseURL     string
	ModelID     string
	Credential  string
	Concurrency int
	// FullURL 为 true 时 BaseURL 就是完整请求地址，不再补齐 /chat/completions。
	FullURL bool
}

func splitIntegrationModelID(modelID string) (prefix, id string, ok bool) {
	modelID = strings.TrimSpace(modelID)
	switch {
	case strings.HasPrefix(modelID, integrationModelPrefixChannel):
		id = strings.TrimSpace(strings.TrimPrefix(modelID, integrationModelPrefixChannel))
		return "channel", id, id != ""
	case strings.HasPrefix(modelID, integrationModelPrefixCustom):
		id = strings.TrimSpace(strings.TrimPrefix(modelID, integrationModelPrefixCustom))
		return "custom", id, id != ""
	}
	return "", "", false
}

// integrationChannelTargetFor 按前缀解析出一条可执行渠道。
// matched=false 表示这个 modelID 不是渠道候选，调用方应继续原有的目录查找。
func integrationChannelTargetFor(ctx context.Context, userConfig *UserConfigStore, owner, modelID string) (integrationChannelTarget, bool, error) {
	prefix, id, matched := splitIntegrationModelID(modelID)
	if !matched {
		return integrationChannelTarget{}, false, nil
	}
	if userConfig == nil {
		return integrationChannelTarget{}, true, errors.New("API 接口配置存储不可用")
	}
	config, err := userConfig.GetConfig(ctx, owner)
	if err != nil {
		return integrationChannelTarget{}, true, fmt.Errorf("读取 API 接口配置失败: %w", err)
	}
	stored := apiIntegrationStoredMap(config)
	values := apiIntegrationValues(stored)
	selection := apiIntegrationChannelSelection(stored)

	if prefix == "channel" {
		for _, preset := range integrationChannelPresets {
			if preset.ID != id {
				continue
			}
			// 与下拉、页面同一口径：填了 Key 就可用，不再检查旧的勾选列表。
			credential := strings.TrimSpace(fmt.Sprint(values[preset.KeyField]))
			if credential == "" || credential == "<nil>" {
				return integrationChannelTarget{}, true, fmt.Errorf("渠道「%s」还没有填写 API Key", preset.Label)
			}
			baseURL := integrationChannelBaseURL(preset, values)
			if baseURL == "" {
				return integrationChannelTarget{}, true, fmt.Errorf("渠道「%s」缺少接口地址", preset.Label)
			}
			return integrationChannelTarget{
				Kind: preset.Group, Label: preset.Label, BaseURL: baseURL,
				ModelID: preset.ModelID, Credential: credential,
			}, true, nil
		}
		return integrationChannelTarget{}, true, fmt.Errorf("渠道 %s 不存在", id)
	}

	for _, raw := range apiIntegrationChannels(stored) {
		channel, ok := raw.(map[string]any)
		if !ok || fmt.Sprint(channel["id"]) != id {
			continue
		}
		group := integrationChannelGroup(channel, selection)
		credential := strings.TrimSpace(fmt.Sprint(channel["apiKey"]))
		if credential == "" || credential == "<nil>" {
			return integrationChannelTarget{}, true, fmt.Errorf("自定义渠道「%s」还没有填写 API Key", fmt.Sprint(channel["name"]))
		}
		baseURL := strings.TrimSpace(fmt.Sprint(channel["baseUrl"]))
		modelID := strings.TrimSpace(fmt.Sprint(channel["model"]))
		if baseURL == "" || baseURL == "<nil>" {
			return integrationChannelTarget{}, true, fmt.Errorf("自定义渠道「%s」缺少接口地址", fmt.Sprint(channel["name"]))
		}
		if modelID == "" || modelID == "<nil>" {
			return integrationChannelTarget{}, true, fmt.Errorf("自定义渠道「%s」缺少模型名", fmt.Sprint(channel["name"]))
		}
		target := integrationChannelTarget{
			Kind: group, Label: fmt.Sprint(channel["name"]), BaseURL: baseURL,
			ModelID: modelID, Credential: credential,
			FullURL: channel["fullUrl"] == true,
		}
		if concurrency, ok := channel["concurrency"].(float64); ok {
			target.Concurrency = int(concurrency)
		}
		return target, true, nil
	}
	return integrationChannelTarget{}, true, fmt.Errorf("自定义渠道 %s 不存在", id)
}

// resolveIntegrationImageCatalogEntry 把 custom:/channel: 候选解析成一条图片
// 目录条目，让图片生成与文本模型共用同一份 API 接口配置。
// matched=false 表示不是渠道候选，调用方应继续使用原有 modelCatalog。
func resolveIntegrationImageCatalogEntry(ctx context.Context, userConfig *UserConfigStore, owner, modelID string) (map[string]any, bool, error) {
	target, matched, err := integrationChannelTargetFor(ctx, userConfig, owner, modelID)
	if !matched {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	if target.Kind != "" && target.Kind != "image" {
		return nil, true, fmt.Errorf("渠道「%s」不是图片渠道，请选择图片分组的渠道", target.Label)
	}
	entry := map[string]any{
		"id":          strings.TrimSpace(modelID),
		"kind":        "image",
		"displayName": target.Label,
		"baseUrl":     target.BaseURL,
		"modelId":     target.ModelID,
		"credential":  target.Credential,
		"provider":    batchfactoryv11.VideoProviderPersonalAPI,
		"enabled":     true,
	}
	if target.FullURL {
		// 完整请求地址直接作为 endpoint，避免被再补一次 /images/generations。
		entry["endpoint"] = target.BaseURL
	}
	return entry, true, nil
}

// resolveLegacyIntegrationImageCatalogEntry keeps old Batch Factory pages
// executable after API channels switched from custom-<model> to
// custom:<channel-id>. Only an exact image-model match is accepted, so a stale
// text/video channel can never be selected by accident.
func resolveLegacyIntegrationImageCatalogEntry(ctx context.Context, userConfig *UserConfigStore, owner, modelID string) (map[string]any, bool, error) {
	modelID = strings.TrimSpace(modelID)
	if userConfig == nil || !strings.HasPrefix(modelID, "custom-") || strings.HasPrefix(modelID, integrationModelPrefixCustom) {
		return nil, false, nil
	}
	wantedModel := strings.TrimSpace(strings.TrimPrefix(modelID, "custom-"))
	if wantedModel == "" {
		return nil, false, nil
	}
	config, err := userConfig.GetConfig(ctx, owner)
	if err != nil {
		return nil, true, fmt.Errorf("读取 API 接口配置失败: %w", err)
	}
	stored := apiIntegrationStoredMap(config)
	selection := apiIntegrationChannelSelection(stored)
	for _, raw := range apiIntegrationChannels(stored) {
		channel, ok := raw.(map[string]any)
		if !ok || !strings.EqualFold(strings.TrimSpace(fmt.Sprint(channel["model"])), wantedModel) {
			continue
		}
		if integrationChannelGroup(channel, selection) != "image" {
			continue
		}
		channelID := strings.TrimSpace(fmt.Sprint(channel["id"]))
		if channelID == "" || channelID == "<nil>" {
			continue
		}
		entry, _, resolveErr := resolveIntegrationImageCatalogEntry(ctx, userConfig, owner, integrationModelPrefixCustom+channelID)
		if entry != nil {
			entry["id"] = modelID
		}
		return entry, true, resolveErr
	}
	return nil, false, nil
}

// integrationTextProvider 是渠道候选项专用的 OpenAI 兼容 provider。
// 与目录模型不同，渠道允许 http 地址（本地 Grok / 大香蕉等），因此不复用
// batchfactoryv11.OpenAICompatibleProvider 的 https 强校验。
type integrationTextProvider struct {
	Label    string
	Endpoint string
	APIKey   string
	Model    string
	Client   *http.Client
}

func (p *integrationTextProvider) Complete(ctx context.Context, input batchfactoryv11.TextCompletionRequest) (string, error) {
	payload := map[string]any{
		"model": p.Model,
		"messages": []map[string]string{
			{"role": "system", "content": input.SystemPrompt},
			{"role": "user", "content": input.UserPrompt},
		},
		"temperature": input.Temperature,
		"max_tokens":  input.MaxTokens,
	}
	if !input.DisableJSONResponse {
		payload["response_format"] = map[string]string{"type": "json_object"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("渠道「%s」请求创建失败", p.Label)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+p.APIKey)
	client := p.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(300 * time.Second)
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("渠道「%s」连接失败: %v", p.Label, err)
	}
	defer response.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(raw))
		if len(snippet) > 300 {
			snippet = snippet[:300]
		}
		return "", fmt.Errorf("渠道「%s」请求失败（HTTP %d）：%s", p.Label, response.StatusCode, snippet)
	}
	if readErr != nil {
		return "", fmt.Errorf("渠道「%s」响应读取失败", p.Label)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded.Choices) == 0 {
		return "", fmt.Errorf("渠道「%s」返回了无法识别的响应", p.Label)
	}
	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		content = strings.TrimSpace(decoded.Choices[0].Text)
	}
	if content == "" {
		return "", fmt.Errorf("渠道「%s」没有返回文本内容", p.Label)
	}
	return content, nil
}

// ResolveIntegrationTextProvider 把 channel:/custom: 候选解析成可用的文本 provider。
// matched=false 表示不是渠道候选；true 时 error 非空代表渠道配置不完整。
func ResolveIntegrationTextProvider(ctx context.Context, userConfig *UserConfigStore, owner, modelID string) (batchfactoryv11.DirectorProvider, bool, error) {
	target, matched, err := integrationChannelTargetFor(ctx, userConfig, owner, modelID)
	if !matched {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	if target.Kind != "" && target.Kind != "text" {
		return nil, true, fmt.Errorf("渠道「%s」不是文本模型渠道", target.Label)
	}
	endpoint := strings.TrimRight(target.BaseURL, "/")
	if !target.FullURL && !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint += "/chat/completions"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, true, fmt.Errorf("渠道「%s」的接口地址格式不正确", target.Label)
	}
	return &integrationTextProvider{
		Label: target.Label, Endpoint: endpoint, APIKey: target.Credential, Model: target.ModelID,
	}, true, nil
}

// integrationImageEndpoint 只处理 OpenAI 兼容的图片接口；
// 地址已带 /images/generations 时原样使用，v1/v3 结尾时直接追加。
func integrationImageEndpoint(base string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("图片渠道地址格式不正确")
	}
	path := strings.TrimRight(parsed.Path, "/")
	switch {
	case strings.HasSuffix(strings.ToLower(path), "/images/generations"):
	case path == "", strings.HasSuffix(path, "/v1"), strings.HasSuffix(path, "/v3"):
		path += "/images/generations"
	default:
		path += "/v1/images/generations"
	}
	parsed.Path = path
	return parsed.String(), nil
}

func decodeIntegrationImageResponse(raw []byte, count int) ([]string, error) {
	var decoded struct {
		Data []struct {
			URL     string `json:"url"`
			Image   string `json:"image_url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded.Data) == 0 {
		return nil, fmt.Errorf("图片模型返回了无法识别的响应")
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
				return nil, fmt.Errorf("图片模型返回了无效图片数据")
			}
			imageURL = "data:" + detectedShuihuoImageMediaType(decodedBytes) + ";base64," + encoded
		}
		if imageURL != "" {
			imageURLs = append(imageURLs, imageURL)
		}
	}
	if len(imageURLs) == 0 {
		return nil, fmt.Errorf("图片模型未返回图片结果")
	}
	if count > 0 && len(imageURLs) > count {
		imageURLs = imageURLs[:count]
	}
	return imageURLs, nil
}

// generateIntegrationChannelImage 用渠道配置构造真实的 /images/generations 请求。
func generateIntegrationChannelImage(ctx context.Context, target integrationChannelTarget, input AgentImageGenerationInput) ([]string, int, error) {
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("图片生成提示词不能为空")
	}
	if target.Kind != "" && target.Kind != "image" {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("渠道「%s」不是图片模型渠道", target.Label)
	}
	endpoint := strings.TrimSpace(target.BaseURL)
	if !target.FullURL {
		fullEndpoint, err := integrationImageEndpoint(target.BaseURL)
		if err != nil {
			return nil, http.StatusUnprocessableEntity, err
		}
		endpoint = fullEndpoint
	}
	payload := map[string]any{"model": target.ModelID, "prompt": prompt}
	if input.Count > 0 {
		payload["n"] = input.Count
	}
	if strings.TrimSpace(input.Quality) != "" {
		payload["quality"] = strings.TrimSpace(input.Quality)
	}
	switch {
	case strings.TrimSpace(input.Size) != "":
		payload["size"] = strings.TrimSpace(input.Size)
	case strings.TrimSpace(input.AspectRatio) != "":
		payload["size"] = shuihuoImageSize(input.AspectRatio)
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
	request.Header.Set("Authorization", "Bearer "+target.Credential)
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
			Message:    shuihuoImageStatusRejectionMessage(response.StatusCode),
			Raw:        string(raw),
			StatusCode: response.StatusCode,
		}
	}
	if readErr != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("%s", appendShuihuoUpstreamRaw("图片通道响应读取中断，请稍后重试", readErr.Error()))
	}
	images, err := decodeIntegrationImageResponse(raw, input.Count)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if input.Count > 0 && len(images) < input.Count {
		return nil, http.StatusBadGateway, fmt.Errorf("图片模型只返回了 %d 张图片，少于请求的 %d 张", len(images), input.Count)
	}
	return images, http.StatusOK, nil
}
