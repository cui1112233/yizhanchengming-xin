package metadataai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/novel"
	workerpkg "github.com/cui1112233/yizhanchengming-xin/api/internal/worker"
)

const maxSourceText = 120000

var allowedStyles = map[string]struct{}{
	"古风虐文": {}, "古风甜文": {}, "古风通用": {}, "年代虐文": {}, "年代甜文": {}, "年代通用": {},
	"现代虐文": {}, "现代甜文": {}, "现代悬疑": {}, "现代通用": {}, "男频都市": {}, "现代女主": {},
	"玄幻": {}, "历史": {}, "爆款BGM": {}, "家庭奇葩": {}, "家庭伤感": {}, "职场打脸": {},
}

const systemPrompt = `你负责判断小说的频道和处理风格。只返回一个 JSON 对象，不要解释，不要 Markdown。
JSON 格式必须为：{"gender":"男或女","style":"风格名"}
其中 style 只能从以下列表选择：古风虐文、古风甜文、古风通用、年代虐文、年代甜文、年代通用、现代虐文、现代甜文、现代悬疑、现代通用、男频都市、现代女主、玄幻、历史、爆款BGM、家庭奇葩、家庭伤感、职场打脸。`

type Config struct {
	Endpoint   string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type Classifier struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func New(cfg Config) (*Classifier, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	model := strings.TrimSpace(cfg.Model)
	if endpoint == "" {
		return nil, errors.New("AI chat completions endpoint is required")
	}
	if model == "" {
		return nil, errors.New("AI model is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 120 * time.Second}
	}
	return &Classifier{endpoint: endpoint, apiKey: strings.TrimSpace(cfg.APIKey), model: model, client: client}, nil
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type classificationPayload struct {
	Gender string `json:"gender"`
	Style  string `json:"style"`
}

func (c *Classifier) Classify(ctx context.Context, book workerpkg.IntakeBook) (workerpkg.AIClassification, error) {
	text := strings.TrimSpace(book.SourceText)
	if text == "" {
		return workerpkg.AIClassification{}, errors.New("AI metadata classification requires source text")
	}
	if len(text) > maxSourceText {
		text = text[:maxSourceText]
	}

	userContent := text
	if strings.TrimSpace(book.BookID) != "" {
		userContent = "书籍ID：" + strings.TrimSpace(book.BookID) + "\n\n正文：\n" + text
	}
	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		MaxTokens:   256,
		Temperature: 0.4,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return workerpkg.AIClassification{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return workerpkg.AIClassification{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return workerpkg.AIClassification{}, fmt.Errorf("AI metadata request: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return workerpkg.AIClassification{}, fmt.Errorf("read AI metadata response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return workerpkg.AIClassification{}, fmt.Errorf("AI metadata request failed: HTTP %d: %s", resp.StatusCode, compact(responseBody))
	}

	var decoded chatResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return workerpkg.AIClassification{}, fmt.Errorf("decode AI metadata response: %w", err)
	}
	if decoded.Error != nil && strings.TrimSpace(decoded.Error.Message) != "" {
		return workerpkg.AIClassification{}, errors.New(decoded.Error.Message)
	}
	if len(decoded.Choices) == 0 {
		return workerpkg.AIClassification{}, errors.New("AI metadata response has no choices")
	}
	return parseClassification(decoded.Choices[0].Message.Content)
}

func parseClassification(content string) (workerpkg.AIClassification, error) {
	cleaned := strings.TrimSpace(content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```JSON")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start < 0 || end < start {
		return workerpkg.AIClassification{}, errors.New("AI metadata response does not contain JSON")
	}
	var payload classificationPayload
	if err := json.Unmarshal([]byte(cleaned[start:end+1]), &payload); err != nil {
		return workerpkg.AIClassification{}, fmt.Errorf("decode AI metadata classification: %w", err)
	}

	var gender novel.Gender
	switch strings.TrimSpace(payload.Gender) {
	case "男", "男频", "male", "Male", "MALE":
		gender = novel.GenderMale
	case "女", "女频", "female", "Female", "FEMALE":
		gender = novel.GenderFemale
	default:
		return workerpkg.AIClassification{}, fmt.Errorf("AI classifier returned invalid gender %q", payload.Gender)
	}
	style := strings.TrimSpace(payload.Style)
	if _, ok := allowedStyles[style]; !ok {
		return workerpkg.AIClassification{}, fmt.Errorf("AI classifier returned invalid style %q", style)
	}
	return workerpkg.AIClassification{Gender: gender, Style: style}, nil
}

func compact(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > 500 {
		return text[:500]
	}
	return text
}
