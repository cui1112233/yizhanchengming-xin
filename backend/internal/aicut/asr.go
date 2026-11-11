package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

type ASRProviderConfig struct{ Endpoint, APIKey, Model string }
type ASRInput struct {
	Filename, ContentType, Language string
	Body                            io.Reader
}
type ASRProvider interface {
	Transcribe(context.Context, ASRInput) ([]SRTCue, error)
}

type OpenAICompatibleASRProvider struct {
	config ASRProviderConfig
	client *http.Client
}

func ASRProviderFromCatalog(settings map[string]any, selection string, client *http.Client) (*OpenAICompatibleASRProvider, error) {
	selection = strings.TrimSpace(selection)
	items := make([]map[string]any, 0)
	switch raw := settings["modelCatalog"].(type) {
	case []map[string]any:
		items = raw
	case []any:
		for _, item := range raw {
			if model, ok := item.(map[string]any); ok {
				items = append(items, model)
			}
		}
	}
	for _, model := range items {
		if strings.TrimSpace(fmt.Sprint(model["id"])) != selection || strings.TrimSpace(fmt.Sprint(model["kind"])) != "asr" {
			continue
		}
		enabled, ok := model["enabled"].(bool)
		if !ok {
			enabled = strings.EqualFold(strings.TrimSpace(fmt.Sprint(model["enabled"])), "true")
		}
		if !enabled {
			return nil, errors.New("选择的 ASR 模型未启用")
		}
		endpoint := strings.TrimSpace(fmt.Sprint(model["endpoint"]))
		if endpoint == "<nil>" {
			endpoint = ""
		}
		if endpoint == "" {
			base := strings.TrimRight(strings.TrimSpace(fmt.Sprint(model["baseUrl"])), "/")
			if base != "" && base != "<nil>" {
				endpoint = base + "/audio/transcriptions"
			}
		}
		return NewOpenAICompatibleASRProvider(ASRProviderConfig{Endpoint: endpoint, APIKey: strings.TrimSpace(fmt.Sprint(model["credential"])), Model: strings.TrimSpace(fmt.Sprint(model["modelId"]))}, client)
	}
	return nil, errors.New("选择的 ASR 模型不存在")
}

func NewOpenAICompatibleASRProvider(config ASRProviderConfig, client *http.Client) (*OpenAICompatibleASRProvider, error) {
	config.Endpoint = strings.TrimSpace(config.Endpoint)
	config.APIKey = strings.TrimSpace(config.APIKey)
	config.Model = strings.TrimSpace(config.Model)
	parsed, err := url.Parse(config.Endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || config.APIKey == "" || config.Model == "" {
		return nil, errors.New("ASR 模型配置不完整")
	}
	if client == nil {
		client = safehttp.NewClientWithTimeout(30 * time.Minute)
	}
	return &OpenAICompatibleASRProvider{config: config, client: client}, nil
}

func (p *OpenAICompatibleASRProvider) Transcribe(ctx context.Context, input ASRInput) ([]SRTCue, error) {
	if input.Body == nil {
		return nil, errors.New("语音素材为空")
	}
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	writeDone := make(chan error, 1)
	go func() {
		var writeErr error
		defer func() { _ = multipartWriter.Close(); _ = writer.CloseWithError(writeErr); writeDone <- writeErr }()
		for key, value := range map[string]string{"model": p.config.Model, "response_format": "verbose_json"} {
			if writeErr = multipartWriter.WriteField(key, value); writeErr != nil {
				return
			}
		}
		if language := strings.TrimSpace(input.Language); language != "" {
			if writeErr = multipartWriter.WriteField("language", language); writeErr != nil {
				return
			}
		}
		if writeErr = multipartWriter.WriteField("timestamp_granularities[]", "segment"); writeErr != nil {
			return
		}
		filename := strings.TrimSpace(input.Filename)
		if filename == "" {
			filename = "audio.bin"
		}
		part, err := multipartWriter.CreateFormFile("file", filename)
		if err != nil {
			writeErr = err
			return
		}
		_, writeErr = io.Copy(part, input.Body)
	}()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.Endpoint, reader)
	if err != nil {
		_ = reader.Close()
		return nil, err
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	response, err := p.client.Do(request)
	if err != nil {
		_ = reader.CloseWithError(err)
		<-writeDone
		return nil, fmt.Errorf("ASR 请求失败：%w", err)
	}
	defer response.Body.Close()
	if writeErr := <-writeDone; writeErr != nil {
		return nil, fmt.Errorf("读取语音素材失败：%w", writeErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("ASR 服务返回 %d：%s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Segments []struct {
			Start json.Number `json:"start"`
			End   json.Number `json:"end"`
			Text  string      `json:"text"`
		} `json:"segments"`
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 32<<20))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("ASR 返回内容无法解析：%w", err)
	}
	cues := make([]SRTCue, 0, len(payload.Segments))
	for _, segment := range payload.Segments {
		start, errStart := strconv.ParseFloat(segment.Start.String(), 64)
		end, errEnd := strconv.ParseFloat(segment.End.String(), 64)
		text := strings.TrimSpace(segment.Text)
		if errStart != nil || errEnd != nil || start < 0 || end <= start || text == "" {
			continue
		}
		cues = append(cues, SRTCue{StartMS: int64(start*1000 + .5), EndMS: int64(end*1000 + .5), Text: text})
	}
	if len(cues) == 0 {
		return nil, errors.New("ASR 没有返回带时间码的有效字幕")
	}
	return cues, nil
}
