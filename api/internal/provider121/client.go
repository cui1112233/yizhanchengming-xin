package provider121

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultEndpoint = "https://txt.121w.com/api.php"
	DefaultTimeout  = 20 * time.Second
	maxResponseBody = 2 << 20
)

var numericID = regexp.MustCompile(`^\d{1,20}$`)

type Config struct {
	Endpoint   string
	HTTPClient *http.Client
	Timeout    time.Duration
}

type Client struct {
	endpoint *url.URL
	http     *http.Client
	timeout  time.Duration
}

type Request struct {
	BookID     string
	PlatformID string
	MaxText    int
}

type BookInfo struct {
	BookName string
	Category string
	Genre    any
}

type Result struct {
	Text     string
	BookInfo BookInfo
	Message  string
}

type RemoteError struct {
	HTTPStatus int
}

func (e *RemoteError) Error() string {
	return fmt.Sprintf("121 HTTP 请求失败（HTTP %d）", e.HTTPStatus)
}

type UpstreamError struct {
	Code    int
	Message string
}

func (e *UpstreamError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return fmt.Sprintf("121 获取失败（code=%d）：%s", e.Code, e.Message)
	}
	return fmt.Sprintf("121 获取失败（code=%d）", e.Code)
}

func NewClient(cfg Config) (*Client, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("解析 121 endpoint: %w", err)
	}
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("121 endpoint 必须使用 https")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("121 endpoint 缺少 host")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("121 endpoint 不能包含用户信息、查询参数或 fragment")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return &Client{endpoint: parsed, http: httpClient, timeout: timeout}, nil
}

func (c *Client) Fetch(ctx context.Context, input Request) (Result, error) {
	bookID := strings.TrimSpace(input.BookID)
	platformID := strings.TrimSpace(input.PlatformID)
	if !numericID.MatchString(bookID) {
		return Result{}, fmt.Errorf("书籍 ID 格式不正确")
	}
	if !numericID.MatchString(platformID) {
		return Result{}, fmt.Errorf("121 platform ID 格式不正确")
	}
	if input.MaxText < 100 || input.MaxText > 100000 {
		return Result{}, fmt.Errorf("max_txt 必须在 100–100000 之间")
	}

	target := *c.endpoint
	query := target.Query()
	query.Set("bookid", bookID)
	query.Set("platform", platformID)
	query.Set("max_txt", strconv.Itoa(input.MaxText))
	target.RawQuery = query.Encode()

	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, target.String(), nil)
	if err != nil {
		return Result{}, fmt.Errorf("创建 121 请求: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	response, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("121 请求失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Result{}, &RemoteError{HTTPStatus: response.StatusCode}
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBody+1))
	if err != nil {
		return Result{}, fmt.Errorf("读取 121 响应: %w", err)
	}
	if len(body) > maxResponseBody {
		return Result{}, fmt.Errorf("121 响应过大")
	}

	var payload struct {
		Code     int             `json:"code"`
		Message  string          `json:"msg"`
		Data     string          `json:"data"`
		BookInfo json.RawMessage `json:"bookinfo"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Result{}, fmt.Errorf("121 返回非 JSON 数据: %w", err)
	}
	if payload.Code != http.StatusOK {
		return Result{}, &UpstreamError{Code: payload.Code, Message: strings.TrimSpace(payload.Message)}
	}
	if strings.TrimSpace(payload.Data) == "" {
		return Result{}, fmt.Errorf("121 成功响应缺少正文")
	}

	bookInfo, err := decodeBookInfo(payload.BookInfo)
	if err != nil {
		return Result{}, err
	}
	return Result{Text: payload.Data, BookInfo: bookInfo, Message: strings.TrimSpace(payload.Message)}, nil
}

func decodeBookInfo(raw json.RawMessage) (BookInfo, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return BookInfo{}, nil
	}
	var value struct {
		BookName  string `json:"book_name"`
		WorkTitle string `json:"work_title"`
		Category  string `json:"category"`
		Genre     any    `json:"genre"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return BookInfo{}, fmt.Errorf("121 bookinfo 格式不正确: %w", err)
	}
	bookName := strings.TrimSpace(value.WorkTitle)
	if bookName == "" {
		bookName = strings.TrimSpace(value.BookName)
	}
	return BookInfo{
		BookName: bookName,
		Category: strings.TrimSpace(value.Category),
		Genre:    value.Genre,
	}, nil
}
