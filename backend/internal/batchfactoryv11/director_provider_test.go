package batchfactoryv11

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDirectorHTTPErrorKeepsRawOnSecondLine(t *testing.T) {
	body := `{"error":{"message":"PROHIBITED_CONTENT/成人内容拒绝，请修改提示词后重试 (request id: 20261111095229058933748268d9d6d705sP3)","type":"invalid_request_error","param":"","code":null}}`
	err := &DirectorHTTPError{StatusCode: http.StatusBadRequest, Model: "[anti-2]gemini-3-8-flash", Body: body}
	headline, raw, found := strings.Cut(err.Error(), "\n上游原文：")
	if !found {
		t.Fatalf("raw upstream must be on its own line: %q", err.Error())
	}
	for _, part := range []string{"成人内容", "自动执行两轮安全修复", "停止继续调用", "更换文本模型"} {
		if !strings.Contains(headline, part) {
			t.Fatalf("headline misses %q: %s", part, headline)
		}
	}
	for _, banned := range []string{"HTTP 400", "model=", "PROHIBITED_CONTENT", "invalid_request_error", "request id"} {
		if strings.Contains(headline, banned) {
			t.Fatalf("headline must not expose %q: %s", banned, headline)
		}
	}
	for _, part := range []string{"HTTP 400", "model=[anti-2]gemini-3-8-flash", "PROHIBITED_CONTENT", "request id: 2026"} {
		if !strings.Contains(raw, part) {
			t.Fatalf("raw misses %q: %s", part, raw)
		}
	}
	if !errors.Is(err, ErrConfigRequired) {
		t.Fatalf("400 must stay a configuration-class error: %v", err)
	}
}

func TestDirectorHTTPFailureHeadlineMapsCommonRejections(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"成人内容", 400, `{"error":{"message":"PROHIBITED_CONTENT"}}`, "成人内容"},
		{"通用审核", 400, `{"error":{"message":"content_policy violation"}}`, "内容审核"},
		{"限流", 429, "", "限流"},
		{"鉴权", 401, "", "余额不足"},
		{"欠费令牌", 401, `{"error":{"message":"Invalid token"}}`, "欠费"},
		{"余额不足", 403, `{"error":{"message":"insufficient quota"}}`, "余额不足"},
		{"模型不存在", 404, "", "不存在"},
		{"上游错误", 502, "", "暂时不可用"},
		{"未知 400", 400, `{"error":{"message":"bad shape"}}`, "文本模型拒绝了本次请求"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := directorHTTPFailureHeadline(tc.status, tc.body)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("headline=%q want contains %q", got, tc.want)
			}
			if tc.name == "未知 400" && strings.Contains(got, "参数或请求格式不合法") {
				t.Fatalf("unknown 400 must not be reported as a malformed request: %q", got)
			}
		})
	}
}

type directorFailingTransport struct{ err error }

func (s directorFailingTransport) RoundTrip(*http.Request) (*http.Response, error) { return nil, s.err }

func TestDirectorTransportErrorKeepsRawOnSecondLine(t *testing.T) {
	provider := OpenAICompatibleProvider{
		Endpoint: "https://text.example.test/v1/chat/completions",
		APIKey:   "test-key",
		Model:    "test-model",
		Client: &http.Client{Transport: directorFailingTransport{
			err: fmt.Errorf(`Post "https://text.example.test/v1/chat/completions": dial tcp 10.0.0.1:443: connect: connection refused`),
		}},
	}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{SystemPrompt: "s", UserPrompt: "u"})
	if err == nil {
		t.Fatal("expected transport failure")
	}
	headline, raw, found := strings.Cut(err.Error(), "\n上游原文：")
	if !found {
		t.Fatalf("raw upstream must be on its own line: %q", err.Error())
	}
	if !strings.Contains(headline, "文本模型通道连接失败") || strings.Contains(headline, "dial tcp") {
		t.Fatalf("headline=%q want Chinese conclusion without net/http details", headline)
	}
	if !strings.Contains(raw, "connection refused") {
		t.Fatalf("raw=%q want original net/http error", raw)
	}
}

func TestOpenAICompatibleHTTPClientUsesHTTP2ForGGWKEndpoint(t *testing.T) {
	client := openAICompatibleHTTPClient("https://www.ggwk1.online/v1/chat/completions")
	transport, ok := client.Transport.(*http.Transport)
	if !ok || !transport.ForceAttemptHTTP2 || transport.TLSNextProto != nil {
		t.Fatalf("transport=%#v, want HTTP/2-capable transport", client.Transport)
	}
}

func TestOpenAICompatibleHTTPClientDefaultsToHTTP2(t *testing.T) {
	for _, endpoint := range []string{
		"https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		"https://abc.preview.tencent-zeabur.cn/v1/chat/completions",
		"https://text.example.test/v1/chat/completions",
	} {
		client := openAICompatibleHTTPClient(endpoint)
		transport, ok := client.Transport.(*http.Transport)
		if !ok || !transport.ForceAttemptHTTP2 || transport.TLSNextProto != nil {
			t.Fatalf("endpoint %s transport=%#v, want HTTP/2-capable transport", endpoint, client.Transport)
		}
	}
}

func TestOpenAICompatibleHTTPClientKeepsLegacyGatewayOnHTTP1(t *testing.T) {
	legacy := map[string]struct{}{"legacy.example.test": {}}
	client := openAICompatibleHTTPClientForHosts("https://legacy.example.test/v1/chat/completions", legacy)
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.ForceAttemptHTTP2 || transport.TLSNextProto == nil {
		t.Fatalf("transport=%#v, want HTTP/1.1-only transport for the legacy exception host", client.Transport)
	}
}

type directorStubTransport struct {
	calls int
	fail  bool
	body  string
}

func (s *directorStubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	s.calls++
	if s.fail {
		return nil, fmt.Errorf("net/http: HTTP/1.x transport connection broken: malformed HTTP response \"\\x00\\x00\\x1e\\x04\"")
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func TestOpenAICompatibleProviderRetriesOnceWithHTTP2OnProtocolMismatch(t *testing.T) {
	first := &directorStubTransport{fail: true}
	second := &directorStubTransport{body: `{"choices":[{"message":{"content":"{}"}}]}`}
	provider := OpenAICompatibleProvider{
		Endpoint:    "https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		APIKey:      "test-key",
		Model:       "test-model",
		Client:      &http.Client{Transport: first},
		retryClient: &http.Client{Transport: second},
	}
	content, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if content != "{}" {
		t.Fatalf("content=%q, want {}", content)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("calls first=%d second=%d, want one attempt plus exactly one HTTP/2 retry", first.calls, second.calls)
	}
}

func TestOpenAICompatibleProviderReportsOriginalErrorWhenHTTP2RetryFails(t *testing.T) {
	first := &directorStubTransport{fail: true}
	second := &directorStubTransport{fail: true}
	provider := OpenAICompatibleProvider{
		Endpoint:    "https://nvewvip.preview.tencent-zeabur.cn/v1/chat/completions",
		APIKey:      "test-key",
		Model:       "test-model",
		Client:      &http.Client{Transport: first},
		retryClient: &http.Client{Transport: second},
	}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "malformed HTTP response") {
		t.Fatalf("error=%v, want the original malformed-response error", err)
	}
	if first.calls != 1 || second.calls != 1 {
		t.Fatalf("calls first=%d second=%d, want exactly one retry", first.calls, second.calls)
	}
}

func TestOpenAICompatibleProviderRequestsJSONObjectResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		format, ok := request["response_format"].(map[string]any)
		if !ok || format["type"] != "json_object" {
			t.Fatalf("response_format=%#v, want json_object", request["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		Endpoint: server.URL,
		APIKey:   "test-key",
		Model:    "test-model",
		Client:   server.Client(),
	}
	if _, err := provider.Complete(context.Background(), TextCompletionRequest{}); err != nil {
		t.Fatalf("complete: %v", err)
	}
}

func TestOpenAICompatibleProviderDoesNotForceJSONObjectForPlainTextProtocol(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if _, exists := request["response_format"]; exists {
			t.Fatalf("plain-text request must not force JSON response: %#v", request["response_format"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"===VIDEO 01===\n时长：10秒\n资产引用：人物=无；场景=无；道具=无\n正文"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	if _, err := provider.Complete(context.Background(), TextCompletionRequest{DisableJSONResponse: true}); err != nil {
		t.Fatalf("complete: %v", err)
	}
}

func TestOpenAICompatibleProviderReportsTruncatedStructuredOutput(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"finish_reason":"length","message":{"content":"{\"director_cards\":["}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{
		Endpoint: server.URL,
		APIKey:   "test-key",
		Model:    "test-model",
		Client:   server.Client(),
	}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if !errors.Is(err, errDirectorOutputTruncated) {
		t.Fatalf("error=%v, want explicit truncation sentinel", err)
	}
	if !strings.Contains(err.Error(), "文本模型输出被长度截断") {
		t.Fatalf("error=%v, want a Chinese truncation diagnostic", err)
	}
}

func TestOpenAICompatibleProviderRetainsUpstreamBodyAndClassifies400(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"model not found: [anti-2]gemini-3.8-flash","type":"invalid_request_error"}}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "[anti-2]gemini-3.8-flash", Client: server.Client()}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil {
		t.Fatal("a 400 must fail the request")
	}
	if !strings.Contains(err.Error(), "model not found") || !strings.Contains(err.Error(), "HTTP 400") {
		t.Fatalf("error must retain the upstream body, got %v", err)
	}
	if !errors.Is(err, ErrConfigRequired) {
		t.Fatalf("400 is a model/config error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("400 calls=%d, want exactly 1 (no retry)", got)
	}
}

func TestOpenAICompatibleProviderRetriesRateLimitOnce(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if attempt == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{}"}}]}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	content, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err != nil {
		t.Fatalf("a 429 must be retried, got %v", err)
	}
	if content != "{}" {
		t.Fatalf("content=%q, want {}", content)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("429 calls=%d, want one retry", got)
	}
}

func TestOpenAICompatibleProviderStopsAfterBoundedServerErrorRetries(t *testing.T) {
	var calls int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"upstream exploded"}`))
	}))
	defer server.Close()

	provider := OpenAICompatibleProvider{Endpoint: server.URL, APIKey: "test-key", Model: "test-model", Client: server.Client()}
	_, err := provider.Complete(context.Background(), TextCompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "upstream exploded") {
		t.Fatalf("error must retain the upstream body, got %v", err)
	}
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("5xx must remain a retryable upstream error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("5xx calls=%d, want exactly 3 attempts", got)
	}
}
