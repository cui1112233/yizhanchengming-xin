package batchfactoryv11

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/safehttp"
)

// DirectorHTTPError keeps the upstream response body so a 400/403 is never
// reduced to a bare status code again. The body is what tells operators
// whether the model rejected the request shape, the model name, or the API key.
type DirectorHTTPError struct {
	StatusCode int
	Model      string
	Body       string
}

// errDirectorOutputTruncated 标识模型因输出长度上限被截断，调用方据此改用
// 更紧凑的提示词重试。用哨兵而不是英文文案匹配，避免界面被迫展示英文。
var errDirectorOutputTruncated = fmt.Errorf("文本模型输出被长度截断，请重试或增大输出上限")

// IsModelOutputTruncated 判断错误是否属于“模型输出被长度上限截断”。
// 这类失败既可能是资产过多，也可能是模型把思考 token 计入输出预算，
// 属于可重试的通道抖动，不能按配置错误直接判成永久失败。
func IsModelOutputTruncated(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errDirectorOutputTruncated) ||
		strings.Contains(err.Error(), "文本模型输出被长度截断") ||
		// 兼容历史记录与测试替身使用的旧英文标记。
		strings.Contains(err.Error(), "director provider output truncated (finish_reason=length)")
}

// truncationRetryMaxTokens 给“截断后重试”一个更大的输出预算。
// 部分 OpenAI 兼容中转会把思考 token 计入 max_tokens，用同一个预算重试
// 仍然会在同一位置截断，所以重试至少给到 12000，最多 32000。
func truncationRetryMaxTokens(current int) int {
	const minimum = 12000
	const maximum = 32000
	next := current * 2
	if next < minimum {
		next = minimum
	}
	if next > maximum {
		next = maximum
	}
	return next
}

// Error 只把中文结论交给界面；状态码、模型名、上游 JSON 与 request id 全部
// 放到第二行「上游原文」里，由前端折叠到「展开原始信息」。
func (e *DirectorHTTPError) Error() string {
	body := directorHTTPBodySnippet([]byte(e.Body), 400)
	if body == "" {
		body = http.StatusText(e.StatusCode)
	}
	raw := fmt.Sprintf("HTTP %d", e.StatusCode)
	if model := strings.TrimSpace(e.Model); model != "" {
		raw += " · model=" + model
	}
	if body != "" {
		raw += " · " + body
	}
	return directorHTTPFailureHeadline(e.StatusCode, e.Body) + "\n上游原文：" + raw
}

// directorHTTPFailureHeadline turns the upstream body into one actionable
// Chinese conclusion. The raw body (including request id and model name) is
// appended by Error() on its own line, never as the headline.
func directorHTTPFailureHeadline(status int, body string) string {
	lower := strings.ToLower(body)
	switch {
	case strings.Contains(lower, "prohibited_content") || strings.Contains(lower, "prohibited content") ||
		strings.Contains(body, "成人内容"):
		return "上游模型以成人内容为由拒绝了本次请求。系统已自动执行两轮安全修复，仍未通过审核；本页已停止继续调用，避免重复扣费。请调整这一页的描述，或更换文本模型。"
	case strings.Contains(lower, "content_policy") || strings.Contains(lower, "content policy") ||
		strings.Contains(lower, "content filter") || strings.Contains(lower, "moderation") ||
		strings.Contains(lower, "safety") || strings.Contains(body, "审核") || strings.Contains(body, "违规"):
		return "上游模型以内容审核为由拒绝了本次请求。系统已自动执行两轮安全修复，仍未通过审核；本页已停止继续调用，避免重复扣费。请调整这一页的描述，或更换文本模型。"
	case strings.Contains(lower, "insufficient quota") || strings.Contains(lower, "not enough quota") ||
		strings.Contains(lower, "insufficient balance") || strings.Contains(lower, "quota exceeded") ||
		strings.Contains(lower, "exceeded your current quota") ||
		strings.Contains(body, "余额不足") || strings.Contains(body, "额度不足") || strings.Contains(body, "欠费"):
		return "文本模型账户余额不足或已欠费，请充值后重试。"
	case status == http.StatusTooManyRequests:
		return "上游文本模型触发限流，请稍后重试。"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		// 中转站欠费时也常以 401 Invalid token 返回，不能只说“密钥失效”。
		if strings.Contains(lower, "invalid token") || strings.Contains(body, "令牌无效") || strings.Contains(body, "无效的令牌") {
			return "文本模型令牌无效或账户已欠费，请检查模型密钥、确认账户余额后重试。"
		}
		return "文本模型密钥无效或账户余额不足，请检查模型密钥、确认账户余额后重试。"
	case status == http.StatusRequestEntityTooLarge ||
		strings.Contains(lower, "context length") || strings.Contains(lower, "context_length") ||
		strings.Contains(lower, "maximum context") || strings.Contains(lower, "max_tokens") ||
		strings.Contains(lower, "token limit") || strings.Contains(lower, "too long") ||
		strings.Contains(body, "超长") || strings.Contains(body, "过长"):
		return "本次请求超出模型上下文长度，请缩短这一页的描述或更换模型。"
	case status == http.StatusNotFound || strings.Contains(lower, "model not found") ||
		strings.Contains(lower, "unknown model") || strings.Contains(lower, "invalid model"):
		return "文本模型不存在或当前账号不支持，请更换模型。"
	case status >= 500:
		return "文本模型通道暂时不可用，请稍后重试。"
	default:
		// 状态码属于参数细节，只放在第二行「上游原文」，主文案不再罗列。
		return "文本模型拒绝了本次请求，请展开原始信息查看上游原因。"
	}
}

// directorTransportError 处理网络层失败：主文案给中文结论，原始 net/http
// 错误放到「上游原文」行，避免把 "Post ...: dial tcp ..." 顶给用户。
type directorTransportError struct{ err error }

func (e *directorTransportError) Error() string {
	headline := "文本模型请求没有完成，请稍后重试。"
	switch {
	case errors.Is(e.err, context.DeadlineExceeded):
		headline = "文本模型请求超时，请稍后重试。"
	case errors.Is(e.err, context.Canceled):
		headline = "文本模型请求已取消。"
	default:
		lower := strings.ToLower(e.err.Error())
		for _, hint := range []string{"connection", "dial ", "eof", "tls", "x509", "no such host", "timeout", "broken pipe"} {
			if strings.Contains(lower, hint) {
				headline = "文本模型通道连接失败，请检查网络或模型地址后重试。"
				break
			}
		}
	}
	return headline + "\n上游原文：" + e.err.Error()
}

func (e *directorTransportError) Unwrap() error { return e.err }

// Unwrap separates retryable channel failures from user-fixable model/key
// configuration problems. 400/401/403/404/422 must not be advertised as a
// transient "capability unavailable" outage.
func (e *DirectorHTTPError) Unwrap() error {
	if directorProviderRetryableStatus(e.StatusCode) {
		return ErrUpstream
	}
	return ErrConfigRequired
}

func directorProviderRetryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return true
	}
	return status >= 500
}

func directorHTTPBodySnippet(raw []byte, limit int) string {
	message := strings.Join(strings.Fields(string(raw)), " ")
	if message == "" {
		return ""
	}
	runes := []rune(message)
	if len(runes) > limit {
		return string(runes[:limit]) + "..."
	}
	return message
}

type DirectorProvider interface {
	Complete(context.Context, TextCompletionRequest) (string, error)
}

type OpenAICompatibleProvider struct {
	Endpoint string
	APIKey   string
	Model    string
	Client   *http.Client
	// retryClient 用于协议不匹配时的一次性 HTTP/2 重试；生产为 nil，按 Endpoint 自动构造。
	retryClient *http.Client
}

// Director planning is intentionally granted more time than a short asset or
// video-provider request. The provider returns one non-streaming JSON result,
// so it may not send headers until the complete H3 director plan is ready.
const directorProviderRequestTimeout = 300 * time.Second

func (p *OpenAICompatibleProvider) Validate() error {
	u, err := url.Parse(strings.TrimSpace(p.Endpoint))
	if err != nil || u.Hostname() == "" {
		return fmt.Errorf("文本模型接口地址无效，请在模型设置中改为完整的 https 地址")
	}
	// 本地模型服务（如自建免费模型网关）通常只监听 http 回环地址；
	// 仅对回环主机放行 http，公网地址仍然强制 https。
	if u.Scheme != "https" && !(u.Scheme == "http" && isLoopbackEndpointHost(u.Hostname())) {
		return fmt.Errorf("文本模型接口地址无效，请在模型设置中改为完整的 https 地址")
	}
	if strings.TrimSpace(p.APIKey) == "" || strings.TrimSpace(p.Model) == "" {
		return fmt.Errorf("文本模型缺少密钥或模型名，请在模型设置中补全")
	}
	return nil
}

func isLoopbackEndpointHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// director400LooksLikeParameterRejection 判断 400 是否更像“参数不被上游接受”，
// 而不是模型名/凭据错误。只有参数类 400 才值得去掉 response_format、压低
// max_tokens 再试一次；模型不存在、密钥错误重试没有意义，还会多打一次上游。
func director400LooksLikeParameterRejection(raw []byte) bool {
	text := strings.ToLower(string(raw))
	for _, deny := range []string{"model not found", "model_not_found", "模型不存在", "does not exist", "unauthorized", "insufficient"} {
		if strings.Contains(text, deny) {
			return false
		}
	}
	for _, hint := range []string{"请求参数有误", "invalid_request_error", "unsupported", "response_format", "max_tokens", "param"} {
		if strings.Contains(text, hint) {
			return true
		}
	}
	return false
}

func (p *OpenAICompatibleProvider) Complete(ctx context.Context, input TextCompletionRequest) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	payload := map[string]any{
		"model":       p.Model,
		"messages":    input.CompletionMessages(),
		"temperature": input.Temperature,
		"max_tokens":  input.MaxTokens,
	}
	// SD director cards and opening variants have a deliberately plain-text
	// wire protocol. Sending OpenAI's JSON-only response_format for those calls
	// contradicts their required ===VIDEO/时长 lines and causes false failures.
	if !input.DisableJSONResponse {
		payload["response_format"] = map[string]string{"type": "json_object"}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	// A 429/5xx is a channel hiccup, not a broken book. Retry a bounded number
	// of times; a 400/401/403/404/422 is a configuration/model problem and is
	// returned immediately with the upstream body so it can be diagnosed.
	const maxAttempts = 3
	var raw []byte
	var statusCode int
	// 部分 OpenAI 兼容中转会对 response_format 或超大的 max_tokens 直接返回
	// 400「请求参数有误」。这种 400 不是凭据/余额问题，允许去掉 response_format
	// 并把输出预算压到安全值后再试一次；仍然 400 才按确定性失败返回。
	fallbackTried := false
	for attempt := 1; attempt <= maxAttempts+1; attempt++ {
		resp, reqErr := p.do(ctx, body)
		if reqErr != nil {
			if SingleAttempt(ctx) || !directorProviderRetryableTransportError(reqErr) || attempt == maxAttempts {
				return "", &directorTransportError{err: reqErr}
			}
			if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
				return "", &directorTransportError{err: reqErr}
			}
			continue
		}
		raw, err = io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		statusCode = resp.StatusCode
		if err != nil {
			if SingleAttempt(ctx) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || attempt == maxAttempts {
				return "", &directorTransportError{err: err}
			}
			if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
				return "", &directorTransportError{err: err}
			}
			continue
		}
		if statusCode >= 200 && statusCode < 300 {
			break
		}
		httpErr := &DirectorHTTPError{StatusCode: statusCode, Model: strings.TrimSpace(p.Model), Body: string(raw)}
		log.Printf("director provider error model=%s status=%d endpoint=%s body=%s", p.Model, statusCode, directorProviderEndpointHost(p.Endpoint), directorHTTPBodySnippet(raw, 400))
		if !SingleAttempt(ctx) && statusCode == http.StatusBadRequest && !fallbackTried && director400LooksLikeParameterRejection(raw) {
			fallbackTried = true
			fallbackMaxTokens := input.MaxTokens
			if fallbackMaxTokens > 8192 {
				fallbackMaxTokens = 8192
			}
			fallbackPayload := map[string]any{
				"model":       p.Model,
				"messages":    input.CompletionMessages(),
				"temperature": input.Temperature,
				"max_tokens":  fallbackMaxTokens,
			}
			if fallbackBody, marshalErr := json.Marshal(fallbackPayload); marshalErr == nil {
				log.Printf("director provider retry without response_format model=%s endpoint=%s max_tokens=%d", p.Model, directorProviderEndpointHost(p.Endpoint), fallbackMaxTokens)
				body = fallbackBody
				continue
			}
		}
		if SingleAttempt(ctx) || !directorProviderRetryableStatus(statusCode) || attempt >= maxAttempts {
			return "", httpErr
		}
		if waitErr := directorProviderWaitRetry(ctx, attempt); waitErr != nil {
			return "", httpErr
		}
	}
	if statusCode < 200 || statusCode >= 300 {
		// Defensive: the loop only exits with a 2xx status or an explicit return.
		return "", &DirectorHTTPError{StatusCode: statusCode, Model: strings.TrimSpace(p.Model), Body: string(raw)}
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text         string `json:"text"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil || len(decoded.Choices) == 0 {
		return "", fmt.Errorf("文本模型返回的响应格式无效，请重试或更换模型")
	}
	choice := decoded.Choices[0]
	// A partial JSON response must never reach the H3 parser as a generic
	// "invalid input".  OpenAI-compatible services expose this condition as
	// finish_reason=length; make it retryable and tell the caller to use a
	// model/output budget that can complete the frozen director contract.
	if strings.EqualFold(strings.TrimSpace(choice.FinishReason), "length") {
		return "", errDirectorOutputTruncated
	}
	content := strings.TrimSpace(choice.Message.Content)
	if content == "" {
		content = strings.TrimSpace(choice.Text)
	}
	if content == "" {
		return "", fmt.Errorf("文本模型返回了空内容，请重试或更换模型")
	}
	return content, nil
}

// do sends one logical request. It keeps the HTTP/2 protocol-mismatch retry
// exactly once, separate from the bounded 429/5xx attempt loop.
func (p *OpenAICompatibleProvider) do(ctx context.Context, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)
	client := p.Client
	if client == nil {
		client = openAICompatibleHTTPClient(p.Endpoint)
	}
	resp, err := SingleAttemptClient(ctx, client, req).Do(req)
	if err == nil {
		return resp, nil
	}
	// 服务端按 HTTP/2 回帧、客户端按 HTTP/1.1 解析时会产生 malformed HTTP
	// response。这里只切换协议重发一次；重试仍失败则保留原始错误。
	if SingleAttempt(ctx) || !directorProtocolMismatchError(err) {
		return nil, err
	}
	retryClient := p.retryClient
	if retryClient == nil {
		retryClient = openAICompatibleHTTP2Client(p.Endpoint)
	}
	retryReq, retryReqErr := http.NewRequestWithContext(ctx, http.MethodPost, p.Endpoint, bytes.NewReader(body))
	if retryReqErr != nil {
		return nil, err
	}
	retryReq.Header.Set("Content-Type", "application/json")
	retryReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	retryResp, retryErr := retryClient.Do(retryReq)
	if retryErr != nil {
		return nil, err
	}
	return retryResp, nil
}

func directorProviderRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// The protocol switch already retried once; do not multiply it.
	return !directorProtocolMismatchError(err)
}

func directorProviderWaitRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(attempt) * 200 * time.Millisecond
	if attempt > 1 {
		delay = 500 * time.Millisecond
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func directorProviderEndpointHost(endpoint string) string {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return ""
	}
	return parsed.Host
}

// openAICompatibleLegacyHTTP1Hosts 是已知会重置长连接的旧网关反向例外表。
// 默认所有 https 端点都走 HTTP/2；只有列在这里的 host 继续强制 HTTP/1.1。
var openAICompatibleLegacyHTTP1Hosts = map[string]struct{}{}

func openAICompatibleHTTPClient(endpoint string) *http.Client {
	return openAICompatibleHTTPClientForHosts(endpoint, openAICompatibleLegacyHTTP1Hosts)
}

func openAICompatibleHTTPClientForHosts(endpoint string, legacyHTTP1 map[string]struct{}) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	host := ""
	if err == nil {
		host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(parsed.Hostname()), "."))
	}
	if openAICompatibleHostPrefersHTTP1(host, legacyHTTP1) {
		transport.ForceAttemptHTTP2 = false
		transport.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	} else {
		transport.ForceAttemptHTTP2 = true
		transport.TLSNextProto = nil
	}
	return safehttp.NewClientWithTransport(directorProviderRequestTimeout, transport)
}

// openAICompatibleHostPrefersHTTP1 只对反向例外表内的旧网关返回 true。
// Tencent Zeabur preview 域名可能只提供 HTTP/2，明确保留在允许 h2 的范围。
func openAICompatibleHostPrefersHTTP1(host string, legacyHTTP1 map[string]struct{}) bool {
	if host == "" {
		return false
	}
	if host == "nvewvip.preview.tencent-zeabur.cn" || strings.HasSuffix(host, ".preview.tencent-zeabur.cn") {
		return false
	}
	_, ok := legacyHTTP1[host]
	return ok
}

// openAICompatibleHTTP2Client 强制 HTTP/2，用于协议不匹配时的一次性重试。
func openAICompatibleHTTP2Client(endpoint string) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = true
	transport.TLSNextProto = nil
	return safehttp.NewClientWithTransport(directorProviderRequestTimeout, transport)
}

// directorProtocolMismatchError 识别“服务端按 HTTP/2 回帧、客户端却按 H1 解析”
// 的错误。这类错误只允许一次协议切换重试。
func directorProtocolMismatchError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "malformed HTTP response") || strings.Contains(message, "\x00\x00\x1e\x04")
}
