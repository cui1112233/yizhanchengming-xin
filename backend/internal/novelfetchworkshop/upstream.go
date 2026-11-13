package novelfetchworkshop

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

const DefaultUpstreamEndpoint = "https://txt.121w.com/api.php"

type UpstreamFetcher struct {
	Client   *http.Client
	Endpoint string
}

type FetchResult struct {
	Code        int
	Message     string
	OriginalRaw string
	BookInfo    BookInfo
}

type FetchAndProcessResult struct {
	Code             int
	Message          string
	OriginalRaw      string
	ProcessedText    string
	BookInfo         BookInfo
	OriginalRawChars int
	ProcessedChars   int
	MaxTxt           int
}

type upstreamResponse struct {
	Code     int      `json:"code"`
	Message  string   `json:"msg"`
	Data     string   `json:"data"`
	BookInfo BookInfo `json:"bookinfo"`
}

func (f UpstreamFetcher) FetchAndProcess(ctx context.Context, bookID, platformID string, maxTxt int) (FetchAndProcessResult, error) {
	// maxTxt is a local processing limit. Sending it upstream can irreversibly
	// truncate the source before OriginalRaw is persisted.
	fetched, err := f.Fetch(ctx, bookID, platformID, 0)
	if err != nil {
		return FetchAndProcessResult{}, err
	}
	processed := ProcessRawText(fetched.OriginalRaw, maxTxt)
	return FetchAndProcessResult{
		Code:             fetched.Code,
		Message:          fetched.Message,
		OriginalRaw:      processed.OriginalRaw,
		ProcessedText:    processed.ProcessedText,
		BookInfo:         fetched.BookInfo,
		OriginalRawChars: processed.OriginalRawChars,
		ProcessedChars:   processed.ProcessedChars,
		MaxTxt:           processed.MaxTxt,
	}, nil
}

func (f UpstreamFetcher) Fetch(ctx context.Context, bookID, platformID string, maxTxt int) (FetchResult, error) {
	bookID = strings.TrimSpace(bookID)
	platformID = strings.TrimSpace(platformID)
	if bookID == "" || platformID == "" {
		return FetchResult{}, fmt.Errorf("121 upstream 参数不完整：缺少书号或书城 ID（bookid=%q platform=%q）", bookID, platformID)
	}
	endpoint := strings.TrimSpace(f.Endpoint)
	if endpoint == "" {
		endpoint = DefaultUpstreamEndpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream endpoint invalid: %w", err)
	}
	query := parsed.Query()
	query.Set("bookid", bookID)
	query.Set("platform", platformID)
	if maxTxt > 0 {
		query.Set("max_txt", strconv.Itoa(maxTxt))
	} else {
		query.Del("max_txt")
	}
	parsed.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream request creation failed: %w", err)
	}
	client := f.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(20 * time.Second)
	}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return FetchResult{}, fmt.Errorf("121 upstream request timed out: %w", context.DeadlineExceeded)
		}
		return FetchResult{}, fmt.Errorf("121 upstream request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return FetchResult{}, fmt.Errorf("121 upstream HTTP status %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream response read failed: %w", err)
	}
	payload, err := decodeUpstreamPayload(body)
	if err != nil {
		return FetchResult{}, fmt.Errorf("121 upstream returned invalid JSON: %w", err)
	}
	if payload.Code != http.StatusOK {
		return FetchResult{}, fmt.Errorf("121 upstream business error %d: %s（bookid=%s platform=%s）", payload.Code, strings.TrimSpace(payload.Message), bookID, platformID)
	}
	return FetchResult{Code: payload.Code, Message: payload.Message, OriginalRaw: payload.Data, BookInfo: payload.BookInfo}, nil
}

// decodeUpstreamPayload 容忍 121 在 JSON 前输出的 PHP 告警/HTML 片段。
// 对方接口报错时会把 Warning 直接 print 在 json_encode 之前，严格按 JSON
// 解码只会看到首个字符 '<'，真正的业务错误（如“获取书籍信息失败”）被丢掉。
func decodeUpstreamPayload(body []byte) (upstreamResponse, error) {
	var payload upstreamResponse
	if err := json.Unmarshal(bytes.TrimSpace(body), &payload); err == nil {
		return payload, nil
	}
	start := bytes.IndexByte(body, '{')
	if start < 0 {
		return payload, fmt.Errorf("上游返回的不是 JSON：%s", upstreamBodySnippet(body))
	}
	if err := json.Unmarshal(bytes.TrimSpace(body[start:]), &payload); err == nil {
		return payload, nil
	}
	return payload, fmt.Errorf("上游返回的不是 JSON：%s", upstreamBodySnippet(body))
}

// upstreamBodySnippet 只保留一小段可读文本，避免把整页 HTML 写进错误信息。
func upstreamBodySnippet(body []byte) string {
	const limit = 160
	text := strings.Join(strings.Fields(string(body)), " ")
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit]) + "…"
	}
	if text == "" {
		return "（空响应）"
	}
	return text
}
