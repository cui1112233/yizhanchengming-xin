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

// upstreamBodyLooksLikeErrorPage 识别 121 在业务码正常时返回的 HTML 错误页。
// 这类响应的 data 字段是 "502 Bad Gateway" 之类的网页源码，直接当正文保存
// 会污染原文并让后续改文把整页 HTML 当成小说内容。
func upstreamBodyLooksLikeErrorPage(text string) bool {
	lower := strings.ToLower(text)
	for _, marker := range []string{
		"<!doctype html", "<html", "<h1>502 bad gateway", "502 bad gateway",
		"504 gateway time-out", "502 gateway", "chrome-error", "err_connection",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func (f UpstreamFetcher) Fetch(ctx context.Context, bookID, platformID string, maxTxt int) (FetchResult, error) {
	bookID = strings.TrimSpace(bookID)
	platformID = strings.TrimSpace(platformID)
	failure := func(class UpstreamFailureClass, message string, cause error) *UpstreamError {
		return &UpstreamError{
			Class:     class,
			Operation: "fetch",
			Message:   fmt.Sprintf("%s（bookid=%s platform=%s）", message, bookID, platformID),
			Cause:     cause,
		}
	}
	if bookID == "" || platformID == "" {
		return FetchResult{}, failure(UpstreamFailurePermanent, "参数不完整：缺少书号或书城 ID", nil)
	}
	endpoint := strings.TrimSpace(f.Endpoint)
	if endpoint == "" {
		endpoint = DefaultUpstreamEndpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return FetchResult{}, failure(UpstreamFailureConfiguration, "endpoint invalid", nil)
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
		return FetchResult{}, failure(UpstreamFailureConfiguration, "request creation failed", nil)
	}
	client := f.Client
	if client == nil {
		client = safehttp.NewClientWithTimeout(20 * time.Second)
	}
	response, err := client.Do(request)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return FetchResult{}, failure(UpstreamFailureTransient, "request timed out", context.DeadlineExceeded)
		}
		// The client wrapper includes the full request URL and its query parameters.
		if urlErr, ok := err.(*url.Error); ok {
			err = urlErr.Err
		}
		return FetchResult{}, failure(UpstreamFailureTransient, "request failed", err)
	}
	defer response.Body.Close()
	status := response.StatusCode
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		class := UpstreamFailurePermanent
		switch {
		case status == http.StatusUnauthorized || status == http.StatusForbidden:
			class = UpstreamFailureConfiguration
		case status == http.StatusRequestTimeout || status == http.StatusTooManyRequests ||
			(status >= http.StatusInternalServerError && status < 600):
			class = UpstreamFailureTransient
		}
		fetchErr := failure(class, "request rejected", nil)
		fetchErr.HTTPStatus = status
		return FetchResult{}, fetchErr
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return FetchResult{}, failure(UpstreamFailureTransient, "response read failed", err)
	}
	payload, err := decodeUpstreamPayload(body)
	if err != nil {
		return FetchResult{}, failure(UpstreamFailureTransient, "returned invalid JSON", nil)
	}
	if payload.Code == 0 {
		return FetchResult{}, failure(UpstreamFailureTransient, "上游响应缺少业务状态", nil)
	}
	if payload.Code != http.StatusOK {
		message := "上游返回业务错误"
		// Preserve known messages used by callers without echoing arbitrary upstream text.
		switch text := strings.TrimSpace(payload.Message); text {
		case "获取书籍信息失败", "参数不完整", "无权访问":
			message = text
		}
		fetchErr := failure(UpstreamFailurePermanent, message, nil)
		fetchErr.BusinessCode = payload.Code
		return FetchResult{}, fetchErr
	}
	if upstreamBodyLooksLikeErrorPage(payload.Data) {
		return FetchResult{}, failure(UpstreamFailureTransient, "上游返回错误页面，请稍后重试", nil)
	}
	if strings.TrimSpace(payload.Data) == "" {
		return FetchResult{}, failure(UpstreamFailureTransient, "上游返回的正文为空，请稍后重试", nil)
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
