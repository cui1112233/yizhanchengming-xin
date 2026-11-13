package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/safehttp"
)

// 上游返回响应头后，生成阶段已经结束，必须把生图槽交给下一页；慢正文由
// 独立的传输槽控制，不能继续堵住整个生成队列。
func TestShuihuoImageGenerationSlotIsReleasedAfterResponseHeaders(t *testing.T) {
	original := shuihuoProcessImageGenerationSemaphore
	shuihuoProcessImageGenerationSemaphore = newShuihuoImageSemaphoreRaw(1)
	t.Cleanup(func() { shuihuoProcessImageGenerationSemaphore = original })

	headersSent := make(chan struct{})
	allowBody := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		close(headersSent)
		<-allowBody
		encoded := base64.StdEncoding.EncodeToString([]byte("test-image"))
		_, _ = io.WriteString(w, `{"data":[{"b64_json":"`+encoded+`"}]}`)
	}))
	defer server.Close()

	target := testShuihuoImageTarget(server.URL + "/v1/images/generations")
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	finished := make(chan error, 1)
	go func() {
		_, _, err := requestShuihuoProviderImage(context.Background(), files, target, target.CatalogModel, target.CatalogModel["endpoint"].(string), "test-secret", "test prompt", nil)
		finished <- err
	}()

	select {
	case <-headersSent:
	case <-time.After(2 * time.Second):
		t.Fatal("provider response headers were not received")
	}

	secondCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	secondRelease, err := acquireShuihuoGPTImageGenerationSlot(secondCtx, "gpt-image-2")
	if err != nil {
		close(allowBody)
		<-finished
		t.Fatalf("next image stayed blocked after response headers: %v", err)
	}
	secondRelease()

	close(allowBody)
	if err := <-finished; err == nil {
		// test-image is intentionally not a valid image; persistence may reject it.
		// The assertion above is only about the slot lifetime.
	}
}

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
		header:    600 * time.Millisecond,
		body:      300 * time.Millisecond,
		download:  200 * time.Millisecond,
		save:      100 * time.Millisecond,
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
		"header_ms=600",
		"body_ms=300",
		"download_ms=200",
		"save_ms=100",
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
		"context deadline exceeded": "ctx_deadline",
		"context canceled":          "ctx_canceled",
		"read tcp 1.2.3.4:1->5.6.7.8:443: connection reset by peer": "reset",
		"wsarecv: An existing connection was forcibly closed":       "reset",
		"unexpected EOF":                           "unexpected_eof",
		"net/http: TLS handshake timeout":          "tls_timeout",
		"image provider response body read failed": "body_read_failed",
		"response body stalled after=30s without new bytes received=12 mode=url": "body_stalled",
		"response body slow total=5m0s received=1024 mode=url":                   "body_slow",
		"something else": "other",
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

func TestShuihuoImageSlowBodyBudgetCanReceiveLargeGeneratedImages(t *testing.T) {
	// HTTP 200 后供应商已经可能计费。只要正文还在持续返回，就不能因为
	// 任意的「总耗时」把已付费图片丢掉；只由无数据进展的 stall 窗口断流。
	if shuihuoImageBodyTotalTimeoutURL != 0 {
		t.Fatalf("URL body total timeout=%s, want disabled", shuihuoImageBodyTotalTimeoutURL)
	}
	if shuihuoImageBodyTotalTimeoutBase64 != 0 {
		t.Fatalf("base64 body total timeout=%s, want disabled", shuihuoImageBodyTotalTimeoutBase64)
	}
	if shuihuoImageRequestTimeout != 0 {
		t.Fatalf("request timeout=%s, want disabled after layered timeouts", shuihuoImageRequestTimeout)
	}
}

// 第 2 项：发了头却不回 body 时，停滞计时必须在阈值内失败，而不是等到整请求超时。
func TestRequestBodyStallsWithoutWaitingClientTimeout(t *testing.T) {
	previous := shuihuoImageBodyStallTimeoutURL
	shuihuoImageBodyStallTimeoutURL = 200 * time.Millisecond
	t.Cleanup(func() { shuihuoImageBodyStallTimeoutURL = previous })

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
		t.Fatal("expected a body stalled error")
	}
	if !strings.Contains(err.Error(), "图片通道响应读取中断") {
		t.Fatalf("error must keep the Chinese body-read headline: %v", err)
	}
	if !strings.Contains(err.Error(), "response body stalled after=200ms without new bytes received=0 mode=url") {
		t.Fatalf("error must report the stalled timeout: %v", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("body stall took %s, want within the layered limit", elapsed)
	}
	var billed *shuihuoImageBilledReadFailure
	if !errors.As(err, &billed) {
		t.Fatalf("status=200 stalled read must be billed: %v", err)
	}
	if billed.kind != shuihuoImageBodyFailureStalled {
		t.Fatalf("billed kind=%q, want BODY_STALLED", billed.kind)
	}
}

// 19 版：正文分块到达、总耗时超过旧闸门但始终有进展时必须判成功。
func TestRequestBodyArrivingLateWithinTimeoutSucceeds(t *testing.T) {
	body := []byte(`{"created":1,"data":[]}`)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-release
		_, _ = w.Write(body)
	}))
	defer server.Close()
	timer := time.AfterFunc(300*time.Millisecond, func() { close(release) })
	defer timer.Stop()

	response, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()
	started := time.Now()
	raw, err := readShuihuoImageResponseBody(response, "url", 900*time.Millisecond, 3*time.Second)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("body arriving inside the window must succeed: %v", err)
	}
	if !bytes.Equal(raw, body) {
		t.Fatalf("unexpected body: %q", raw)
	}
	if elapsed < 250*time.Millisecond {
		t.Fatalf("body read returned in %s, expected to wait for the delayed body", elapsed)
	}
}

// URL 与 base64 两档的停滞阈值、总上限都必须对齐，避免出现两套口径。
func TestImageBodyTimeoutAligned(t *testing.T) {
	if shuihuoImageBodyStallTimeoutURL != shuihuoImageBodyStallTimeoutBase64 {
		t.Fatalf("body stall timeouts must be aligned: url=%s base64=%s",
			shuihuoImageBodyStallTimeoutURL, shuihuoImageBodyStallTimeoutBase64)
	}
	if shuihuoImageBodyTotalTimeoutURL != shuihuoImageBodyTotalTimeoutBase64 {
		t.Fatalf("body total timeouts must be aligned: url=%s base64=%s",
			shuihuoImageBodyTotalTimeoutURL, shuihuoImageBodyTotalTimeoutBase64)
	}
}

// 线上供应商会先返回 200，再持续传输较大的图片正文；默认停滞窗口负责识别
// 真正断流，总上限则必须覆盖慢速但仍在前进的大图响应。
func TestImageRequestDefaultsAllowSlowProviderBody(t *testing.T) {
	if shuihuoImageResponseHeaderTimeout != time.Minute {
		t.Fatalf("response header timeout=%s, want 1m", shuihuoImageResponseHeaderTimeout)
	}
	if shuihuoImageRequestTimeout != 0 {
		t.Fatalf("request timeout=%s, want disabled", shuihuoImageRequestTimeout)
	}
	if shuihuoImageBodyStallTimeoutURL != 120*time.Second || shuihuoImageBodyStallTimeoutBase64 != 120*time.Second {
		t.Fatalf("body stall timeouts=%s/%s, want 120s/120s", shuihuoImageBodyStallTimeoutURL, shuihuoImageBodyStallTimeoutBase64)
	}
	if shuihuoImageBodyTotalTimeoutURL != 0 || shuihuoImageBodyTotalTimeoutBase64 != 0 {
		t.Fatalf("body total timeouts=%s/%s, want disabled", shuihuoImageBodyTotalTimeoutURL, shuihuoImageBodyTotalTimeoutBase64)
	}
}

func TestBodyReaderDoesNotDropContinuouslyArrivingPaidImage(t *testing.T) {
	chunks := make([][]byte, 10)
	for index := range chunks {
		chunks[index] = []byte("paid-image-chunk")
	}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&pacedBodyReader{chunks: chunks, interval: 20 * time.Millisecond}),
	}

	started := time.Now()
	raw, err := readShuihuoImageResponseBody(response, "base64", 80*time.Millisecond, 0)
	if err != nil {
		t.Fatalf("continuously arriving paid image must not be cut off: %v", err)
	}
	if len(raw) != len(chunks)*len(chunks[0]) {
		t.Fatalf("received=%d bytes, want %d", len(raw), len(chunks)*len(chunks[0]))
	}
	if elapsed := time.Since(started); elapsed < 150*time.Millisecond {
		t.Fatalf("elapsed=%s, test did not cross the old total-timeout shape", elapsed)
	}
}

// 连响应头都不回时，注入的 ResponseHeaderTimeout 必须早于整请求超时生效。
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

	client := safehttp.NewClientWithTransportWithoutTimeout(newShuihuoImageTransport(true))
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
	if err == nil || !strings.Contains(err.Error(), "图片通道返回的内容不是有效图片数据") {
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
	if transport.ResponseHeaderTimeout != shuihuoImageResponseHeaderTimeout {
		t.Fatalf("response header timeout=%s, want %s", transport.ResponseHeaderTimeout, shuihuoImageResponseHeaderTimeout)
	}
	if first.Timeout != shuihuoImageRequestTimeout {
		t.Fatalf("client timeout=%s, want %s", first.Timeout, shuihuoImageRequestTimeout)
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
	if !strings.Contains(err.Error(), "图片通道响应读取中断") {
		t.Fatalf("error text must keep the Chinese body-read headline: %v", err)
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

// pacedBodyReader 按固定间隔返回数据块，模拟「慢速但一直有数据」的正文。
// 间隔取毫秒级，测试不依赖真实的 30s/300s 阈值。
type pacedBodyReader struct {
	chunks   [][]byte
	index    int
	interval time.Duration
}

func (r *pacedBodyReader) Read(p []byte) (int, error) {
	if r.index > 0 && r.interval > 0 {
		time.Sleep(r.interval)
	}
	if r.index >= len(r.chunks) {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[r.index])
	r.index++
	return n, nil
}

// blockedBodyReader 先返回一块数据，随后永久阻塞，模拟上游 200 后断流。
type blockedBodyReader struct {
	first   []byte
	sent    bool
	release chan struct{}
}

func (r *blockedBodyReader) Read(p []byte) (int, error) {
	if !r.sent {
		r.sent = true
		return copy(p, r.first), nil
	}
	<-r.release
	return 0, io.EOF
}

// endlessBodyReader 持续返回数据块，用于触发总上限兜底。
type endlessBodyReader struct {
	chunk    []byte
	interval time.Duration
	release  chan struct{}
}

func (r *endlessBodyReader) Read(p []byte) (int, error) {
	if r.interval > 0 {
		time.Sleep(r.interval)
	}
	select {
	case <-r.release:
		return 0, io.EOF
	default:
		return copy(p, r.chunk), nil
	}
}

// ① 慢速但持续有数据：每块都会重置停滞计时，必须成功，即使总耗时远超停滞阈值。
func TestBodyStallReaderSucceedsWithContinuousChunks(t *testing.T) {
	chunks := make([][]byte, 8)
	for index := range chunks {
		chunks[index] = []byte("01234567")
	}
	reader := &pacedBodyReader{chunks: chunks, interval: 20 * time.Millisecond}
	response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(reader)}

	started := time.Now()
	raw, err := readShuihuoImageResponseBody(response, "url", 120*time.Millisecond, 5*time.Second)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatalf("continuous slow chunks must succeed, got %v", err)
	}
	if len(raw) != 8*len(chunks[0]) {
		t.Fatalf("received=%d bytes, want %d", len(raw), 8*len(chunks[0]))
	}
	// 总耗时已超过停滞阈值，但因为没有「连续停滞」所以不应失败。
	if elapsed < 100*time.Millisecond {
		t.Fatalf("elapsed=%s, test did not exercise the slow path", elapsed)
	}
}

// ② 连续无数据超过停滞阈值：失败并归类 BODY_STALLED，且已计费类型可被识别。
func TestBodyStallFailsAndClassifiesBodyStalled(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(&blockedBodyReader{first: []byte("partial"), release: release}),
	}

	started := time.Now()
	raw, err := readShuihuoImageResponseBody(response, "url", 60*time.Millisecond, 5*time.Second)
	elapsed := time.Since(started)
	var stalled *shuihuoImageBodyStallError
	if !errors.As(err, &stalled) {
		t.Fatalf("want shuihuoImageBodyStallError, got %v", err)
	}
	if string(raw) != "partial" {
		t.Fatalf("partial bytes=%q, want %q", raw, "partial")
	}
	if stalled.received != len("partial") || stalled.stall != 60*time.Millisecond || stalled.mode != "url" {
		t.Fatalf("stalled detail=%+v", stalled)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("stall detection took %s", elapsed)
	}
	if kind := shuihuoImageAttemptKindOfError(err, http.StatusOK); kind != "body_stalled" {
		t.Fatalf("attempt kind=%q, want body_stalled", kind)
	}
	billed := &shuihuoImageBilledReadFailure{message: err.Error(), kind: shuihuoImageBodyFailureStalled}
	if billed.Kind() != shuihuoImageBodyFailureStalled {
		t.Fatalf("billed kind=%q", billed.Kind())
	}
	if !errors.Is(billed, batchfactoryv11.ErrUpstream) {
		t.Fatal("billed failure must keep ErrUpstream semantics")
	}
}

// ③ 总上限兜底：一直有数据但到总上限仍未读完，归类 BODY_SLOW，而不是 BODY_STALLED。
func TestBodyTotalCapFailsAsBodySlow(t *testing.T) {
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(&endlessBodyReader{
			chunk:    bytes.Repeat([]byte("x"), 4*1024),
			interval: 10 * time.Millisecond,
			release:  release,
		}),
	}

	_, err := readShuihuoImageResponseBody(response, "url", 2*time.Second, 80*time.Millisecond)
	var slow *shuihuoImageBodySlowError
	if !errors.As(err, &slow) {
		t.Fatalf("want shuihuoImageBodySlowError, got %v", err)
	}
	if slow.received <= 0 {
		t.Fatalf("slow received=%d, want data before the total cap", slow.received)
	}
	if slow.total != 80*time.Millisecond || slow.mode != "url" {
		t.Fatalf("slow detail=%+v", slow)
	}
	if kind := shuihuoImageAttemptKindOfError(err, http.StatusOK); kind != "body_slow" {
		t.Fatalf("attempt kind=%q, want body_slow", kind)
	}
	var stalled *shuihuoImageBodyStallError
	if errors.As(err, &stalled) {
		t.Fatal("total cap must not be classified as BODY_STALLED")
	}
}

// HTTP 200 表示供应商已经接单并可能计费。即使正文最终断流，也不能自动
// 再发一张；应保留失败记录，交给用户明确决定是否付费重试。
func TestBilledStalledResponseIsNotAutomaticallyResent(t *testing.T) {
	previousStall := shuihuoImageBodyStallTimeoutURL
	shuihuoImageBodyStallTimeoutURL = 100 * time.Millisecond
	t.Cleanup(func() { shuihuoImageBodyStallTimeoutURL = previousStall })
	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "3")

	var requests atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-release
	}))
	t.Cleanup(func() {
		close(release)
		server.Close()
	})

	endpoint := server.URL + "/v1/images/generations"
	target := testShuihuoImageTarget(endpoint)
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	_, _, _, err := generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, files, nil, "alice", target, "test prompt", nil, nil)
	var billed *shuihuoImageBilledReadFailure
	if !errors.As(err, &billed) {
		t.Fatalf("want billed stalled failure, got %v", err)
	}
	if billed.kind != shuihuoImageBodyFailureStalled {
		t.Fatalf("billed kind=%q, want BODY_STALLED", billed.kind)
	}
	if billed.retried {
		t.Fatal("billed response must not be automatically resent")
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("provider requests=%d, want exactly 1 billed request", got)
	}
}
