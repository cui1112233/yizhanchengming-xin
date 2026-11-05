package httpapi

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
)

func testShuihuoImageTarget(endpoint string) shuihuoAssetImageTarget {
	return shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"endpoint":   endpoint,
			"credential": "test-secret",
		},
	}
}

// 第 1 项：结构化日志必须包含全部字段，且顺序固定、单行输出。
func TestLogImageAttemptFieldOrderIsStable(t *testing.T) {
	var captured bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&captured)
	t.Cleanup(func() { log.SetOutput(previous) })

	logImageAttempt(shuihuoImageAttemptLog{
		elapsed:   1500 * time.Millisecond,
		attempt:   2,
		inner:     3,
		status:    200,
		queued:    250 * time.Millisecond,
		errorKind: "body_read_failed",
		model:     "gpt-image-2",
		detail:    "boom",
	})

	var line string
	for _, candidate := range strings.Split(strings.TrimSpace(captured.String()), "\n") {
		if strings.Contains(candidate, "shuihuo image attempt ") {
			line = candidate
			break
		}
	}
	if line == "" {
		t.Fatalf("structured attempt log not found: %q", captured.String())
	}
	fields := []string{
		"shuihuo image attempt",
		"elapsed_ms=1500",
		"attempt=2",
		"inner_attempt=3",
		"http_status=200",
		"queued_ms=250",
		"error_kind=body_read_failed",
		"model=gpt-image-2",
		`detail="boom"`,
	}
	lastIndex := -1
	for _, field := range fields {
		index := strings.Index(line, field)
		if index < 0 {
			t.Fatalf("field %q missing from line %q", field, line)
		}
		if index < lastIndex {
			t.Fatalf("field %q out of order in line %q", field, line)
		}
		lastIndex = index
	}
}

func TestShuihuoImageAttemptErrorKindClassification(t *testing.T) {
	cases := map[string]string{
		"context deadline exceeded (Client.Timeout exceeded while awaiting headers)": "client_timeout",
		"context deadline exceeded":                     "ctx_deadline",
		"context canceled":                             "ctx_canceled",
		"read tcp 1.2.3.4:1->5.6.7.8:443: connection reset by peer": "reset",
		"wsarecv: An existing connection was forcibly closed":      "reset",
		"unexpected EOF":                               "unexpected_eof",
		"net/http: TLS handshake timeout":              "tls_timeout",
		"image provider response body read failed":     "body_read_failed",
		"something else":                               "other",
	}
	for message, want := range cases {
		if got := shuihuoImageAttemptKindOfError(errors.New(message), 0); got != want {
			t.Fatalf("kind(%q)=%q, want %q", message, got, want)
		}
	}
	if got := shuihuoImageAttemptKindOfStatus(404); got != "http_4xx" {
		t.Fatalf("status 404 kind=%q", got)
	}
	if got := shuihuoImageAttemptKindOfStatus(502); got != "http_5xx" {
		t.Fatalf("status 502 kind=%q", got)
	}
	if got := shuihuoImageAttemptKindOfStatus(200); got != "ok" {
		t.Fatalf("status 200 kind=%q", got)
	}
}

// 第 2 项：发了头却不回 body 时，body 分层计时必须在阈值内失败，而不是等到 900s。
func TestRequestBodyReadTimesOutWithoutWaitingClientTimeout(t *testing.T) {
	previous := shuihuoImageBodyReadTimeoutURL
	shuihuoImageBodyReadTimeoutURL = 200 * time.Millisecond
	t.Cleanup(func() { shuihuoImageBodyReadTimeoutURL = previous })

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-release
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	endpoint := server.URL + "/v1/images/generations"
	target := testShuihuoImageTarget(endpoint)
	started := time.Now()
	_, _, err := requestShuihuoProviderImage(context.Background(), nil, target, target.CatalogModel, endpoint, "test-secret", "test prompt", nil)
	elapsed := time.Since(started)
	if err == nil {
		t.Fatal("expected a body read timeout error")
	}
	if !strings.Contains(err.Error(), "response body read failed") {
		t.Fatalf("error must keep the existing body-read string: %v", err)
	}
	if !strings.Contains(err.Error(), "body read timeout after=200ms mode=url") {
		t.Fatalf("error must report the layered timeout: %v", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("body timeout took %s, want within the layered limit", elapsed)
	}
	var billed *shuihuoImageBilledReadFailure
	if !errors.As(err, &billed) {
		t.Fatalf("status=200 read failure must be billed (no resend): %v", err)
	}
}

// 第 2 项：连响应头都不回时，ResponseHeaderTimeout 必须早于 900s 生效。
func TestResponseHeaderTimeoutFailsFast(t *testing.T) {
	previous := shuihuoImageResponseHeaderTimeout
	shuihuoImageResponseHeaderTimeout = 200 * time.Millisecond
	t.Cleanup(func() { shuihuoImageResponseHeaderTimeout = previous })

	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-release
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	client := &http.Client{Timeout: 15 * time.Minute, Transport: newShuihuoImageTransport(true)}
	started := time.Now()
	response, err := client.Post(server.URL, "application/json", strings.NewReader("{}"))
	if response != nil {
		_ = response.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "timeout awaiting response headers") {
		t.Fatalf("want response header timeout error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 10*time.Second {
		t.Fatalf("header timeout took %s", elapsed)
	}
	if timeout := newShuihuoImageTransport(true).ResponseHeaderTimeout; timeout != 200*time.Millisecond {
		t.Fatalf("transport response header timeout=%s", timeout)
	}
}

// 第 2 项：立刻回 200 空 body 属于快速失败，同样不能等到 900s。
func TestRequestBodyEmpty200FailsFast(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpoint := server.URL + "/v1/images/generations"
	target := testShuihuoImageTarget(endpoint)
	started := time.Now()
	_, _, err := requestShuihuoProviderImage(context.Background(), nil, target, target.CatalogModel, endpoint, "test-secret", "test prompt", nil)
	if err == nil || !strings.Contains(err.Error(), "response is invalid") {
		t.Fatalf("want invalid response error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 5*time.Second {
		t.Fatalf("empty 200 took %s", elapsed)
	}
}

// 第 3 项：同一协议变体必须复用同一个 client 与 Transport。
func TestImageGenerationClientIsSharedPerProtocol(t *testing.T) {
	first := shuihuoImageGenerationHTTPClient("https://images.example.test/v1/images/generations")
	second := shuihuoImageGenerationHTTPClient("https://another.example.test/v1/images/generations")
	if first != second {
		t.Fatal("HTTP/2 image client must be shared, not created per request")
	}
	if first.Transport != second.Transport {
		t.Fatal("shared client must reuse the same Transport")
	}
	transport := first.Transport.(*http.Transport)
	if transport.ResponseHeaderTimeout != 60*time.Second {
		t.Fatalf("response header timeout=%s, want 60s", transport.ResponseHeaderTimeout)
	}
	if transport.MaxIdleConnsPerHost <= 0 || transport.IdleConnTimeout <= 0 {
		t.Fatalf("idle pool not configured: perHost=%d idle=%s", transport.MaxIdleConnsPerHost, transport.IdleConnTimeout)
	}
}

// 第 4 项：200 已返回但 body 中途断开时不再重发（对面已计费）。
func TestBilledReadFailureIsNotRetried(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Error("hijacker unavailable")
			return
		}
		conn, buffer, err := hijacker.Hijack()
		if err != nil {
			t.Errorf("hijack failed: %v", err)
			return
		}
		_, _ = buffer.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 128\r\n\r\n{\"data\":[]}")
		_ = buffer.Flush()
		_ = conn.Close()
	}))
	defer server.Close()

	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "3")

	endpoint := server.URL + "/v1/images/generations"
	target := testShuihuoImageTarget(endpoint)
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	_, _, _, err := generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, files, nil, "alice", target, "test prompt", nil, nil)
	if err == nil {
		t.Fatal("expected a read failure")
	}
	var billed *shuihuoImageBilledReadFailure
	if !errors.As(err, &billed) {
		t.Fatalf("want billed read failure, got %v", err)
	}
	if !strings.Contains(err.Error(), "response body read failed") {
		t.Fatalf("error text must stay compatible: %v", err)
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("provider requests=%d, want exactly 1 (no billed resend)", got)
	}

	// 关闭去重开关后应恢复旧行为：按 maxAttempts 重发。
	requests.Store(0)
	t.Setenv("QIANTIE_IMAGE_DEDUPE_BILLED_RETRY", "0")
	_, _, _, err = generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, files, nil, "alice", target, "test prompt", nil, nil)
	if err == nil {
		t.Fatal("expected a read failure with dedupe disabled")
	}
	if errors.As(err, &billed) {
		t.Fatalf("dedupe disabled must keep the plain retryable error, got %v", err)
	}
	if got := requests.Load(); got != 3 {
		t.Fatalf("provider requests=%d, want 3 retries when dedupe disabled", got)
	}
}
