package httpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
)

func TestProviderImageRequestFallsBackToBase64WhenDownloadFails(t *testing.T) {
	var (
		mu     sync.Mutex
		bodies []string
	)
	const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(string(body), "response_format") {
			_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + tinyPNG + `"}]}`))
			return
		}
		// 默认返回一个无法下载的 CDN 地址，触发 base64 兜底。
		_, _ = w.Write([]byte(`{"data":[{"url":"http://127.0.0.1:1/unavailable.png"}]}`))
	}))
	defer server.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"endpoint":   server.URL + "/v1/images/generations",
			"credential": "test-secret",
		},
	}
	ctx := withShuihuoImageConcurrency(context.Background(), 1)
	if _, _, _, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "test prompt", nil, nil); err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 {
		t.Fatalf("requests=%d, want URL request then base64 fallback", len(bodies))
	}
	if strings.Contains(bodies[0], "response_format") {
		t.Fatalf("first request must use URL mode: %s", bodies[0])
	}
	if !strings.Contains(bodies[1], `"response_format":"b64_json"`) {
		t.Fatalf("fallback request must ask for base64: %s", bodies[1])
	}
}

func TestProviderImageResponseFormatRejectionDetection(t *testing.T) {
	if !shuihuoImageResponseRejectsResponseFormat([]byte(`{"error":{"message":"unknown parameter: response_format"}}`)) {
		t.Fatal("response_format rejection was not detected")
	}
	if shuihuoImageResponseRejectsResponseFormat([]byte(`{"error":{"message":"提示词包含违规内容"}}`)) {
		t.Fatal("content policy rejection must not be treated as response_format rejection")
	}
}

func TestDownloadShuihuoGeneratedImageRetriesTemporaryProviderFailure(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 2 {
			http.Error(w, "temporary image unavailable", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer server.Close()

	mediaType, payload, status, err := downloadShuihuoGeneratedImageWithRetry(context.Background(), server.Client(), server.URL, "", false)
	if err != nil || mediaType != "image/png" || len(payload) == 0 || status != http.StatusOK {
		t.Fatalf("mediaType=%q payload=%d status=%d err=%v", mediaType, len(payload), status, err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts=%d, want 2", got)
	}
}

func TestDecodeShuihuoGeneratedImageRetriesSameOriginWithProviderCredential(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if r.Header.Get("Authorization") != "Bearer test-secret" {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer server.Close()

	mediaType, payload, err := decodeShuihuoGeneratedImage(context.Background(), server.Client(), "", server.URL+"/generated.png", "test-secret", server.URL+"/v1/images/generations")
	if err != nil || mediaType != "image/png" || len(payload) == 0 {
		t.Fatalf("mediaType=%q payload=%d err=%v", mediaType, len(payload), err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("download attempts=%d, want unauthenticated then authorized", got)
	}
}

func TestDecodeShuihuoGeneratedImageDoesNotSendCredentialToOtherOrigin(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer provider.Close()

	var attempts atomic.Int32
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("authorization header leaked to image origin: %q", got)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer imageServer.Close()

	mediaType, payload, err := decodeShuihuoGeneratedImage(context.Background(), imageServer.Client(), "", imageServer.URL+"/generated.png", "test-secret", provider.URL+"/v1/images/generations")
	if err != nil || mediaType != "image/png" || len(payload) == 0 {
		t.Fatalf("mediaType=%q payload=%d err=%v", mediaType, len(payload), err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("download attempts=%d, want one unauthenticated request", got)
	}
}

func TestShuihuoImageDownloadHTTPClientNegotiatesHTTP2(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Errorf("request protocol=%q, want HTTP/2", r.Proto)
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := shuihuoImageDownloadHTTPClient()
	client.Transport.(*http.Transport).TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig
	_, payload, status, err := downloadShuihuoGeneratedImage(context.Background(), client, server.URL, "", false)
	if err != nil || status != http.StatusOK || len(payload) == 0 {
		t.Fatalf("payload=%d status=%d err=%v", len(payload), status, err)
	}
}

func TestShuihuoLongRunningImageHTTPClientNegotiatesHTTP2(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Errorf("request protocol=%q, want HTTP/2", r.Proto)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := shuihuoLongRunningImageHTTPClient()
	client.Transport.(*http.Transport).TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status=%d, want %d", response.StatusCode, http.StatusNoContent)
	}
}

func TestShuihuoImageGenerationHTTPClientSelectsProtocolByEndpoint(t *testing.T) {
	for _, endpoint := range []string{
		"https://api.vveai.com/v1/images/generations",
		"https://www.ggwk1.online/v1/images/generations",
		"https://nvewvip.preview.tencent-zeabur.cn/v1/images/generations",
		"https://abc.preview.tencent-zeabur.cn/v1/images/generations",
		"https://images.example.test/v1/images/generations",
	} {
		client := shuihuoImageGenerationHTTPClient(endpoint)
		transport := client.Transport.(*http.Transport)
		// 共享 Transport 首次被使用后，net/http 会自动填充 TLSNextProto，
		// 因此这里不能再以 TLSNextProto 判断协议能力，改用 ForceAttemptHTTP2
		// 并排除显式 HTTP/1.1 变体。
		if !transport.ForceAttemptHTTP2 || client == shuihuoImageSharedHTTP1Client() {
			t.Fatalf("endpoint %s must use the HTTP/2-capable transport", endpoint)
		}
	}
}

func TestShuihuoImageGenerationHTTPClientKeepsLegacyGatewayOnHTTP1(t *testing.T) {
	shuihuoImageGenerationLegacyHTTP1Hosts["legacy-images.example.test"] = struct{}{}
	defer delete(shuihuoImageGenerationLegacyHTTP1Hosts, "legacy-images.example.test")
	client := shuihuoImageGenerationHTTPClient("https://legacy-images.example.test/v1/images/generations")
	transport := client.Transport.(*http.Transport)
	if transport.ForceAttemptHTTP2 || transport.TLSNextProto == nil {
		t.Fatal("legacy exception gateway must retain the HTTP/1.1 transport")
	}
}

func TestDownloadShuihuoGeneratedImageRetriesEmptySuccessfulResponse(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	}))
	defer server.Close()

	mediaType, payload, status, err := downloadShuihuoGeneratedImageWithRetry(context.Background(), server.Client(), server.URL, "", false)
	if err != nil || mediaType != "image/png" || len(payload) == 0 || status != http.StatusOK {
		t.Fatalf("mediaType=%q payload=%d status=%d err=%v", mediaType, len(payload), status, err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("download attempts=%d, want retry after empty body", got)
	}
}

func TestGPTImageGenerationRequestsAreSerialized(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		for observed := maximum.Load(); current > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		active.Add(-1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer server.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"endpoint":   server.URL + "/v1/images/generations",
			"credential": "test-secret",
		},
	}
	const requests = 4
	// 统一配置设成 1 路时必须保持串行；同一任务共用同一个信号量实例。
	taskCtx := withShuihuoImageConcurrency(context.Background(), 1)
	var workers sync.WaitGroup
	errs := make(chan error, requests)
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, _, _, err := generateShuihuoImageArtifact(taskCtx, nil, files, nil, "alice", target, "test prompt", nil, nil)
			errs <- err
		}()
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := maximum.Load(); got != 1 {
		t.Fatalf("maximum simultaneous gpt-image requests=%d, want 1", got)
	}
}

func TestGPTImageGenerationRequestsHonorConfiguredConcurrency(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		for observed := maximum.Load(); current > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		active.Add(-1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer server.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"endpoint":   server.URL + "/v1/images/generations",
			"credential": "test-secret",
		},
	}
	const requests = 4
	taskCtx := withShuihuoImageConcurrency(context.Background(), 2)
	var workers sync.WaitGroup
	errs := make(chan error, requests)
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, _, _, err := generateShuihuoImageArtifact(taskCtx, nil, files, nil, "alice", target, "test prompt", nil, nil)
			errs <- err
		}()
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := maximum.Load(); got != 2 {
		t.Fatalf("maximum simultaneous gpt-image requests=%d, want 2", got)
	}
}

func TestGPTImageConcurrencyIsPerTaskNotProcessWide(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		for observed := maximum.Load(); current > observed; observed = maximum.Load() {
			if maximum.CompareAndSwap(observed, current) {
				break
			}
		}
		time.Sleep(30 * time.Millisecond)
		active.Add(-1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer server.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"endpoint":   server.URL + "/v1/images/generations",
			"credential": "test-secret",
		},
	}
	// 两本书并行、各自配置 2 路时，全进程应有 4 个请求在跑，而不是共用 2 路。
	taskCtxs := []context.Context{
		withShuihuoImageConcurrency(context.Background(), 2),
		withShuihuoImageConcurrency(context.Background(), 2),
	}
	var workers sync.WaitGroup
	errs := make(chan error, 4)
	for _, ctx := range taskCtxs {
		for range 2 {
			workers.Add(1)
			go func(taskCtx context.Context) {
				defer workers.Done()
				_, _, _, err := generateShuihuoImageArtifact(taskCtx, nil, files, nil, "alice", target, "test prompt", nil, nil)
				errs <- err
			}(ctx)
		}
	}
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := maximum.Load(); got != 4 {
		t.Fatalf("maximum simultaneous gpt-image requests=%d, want 4", got)
	}
}

func TestGPTImageGenerationSlotCanBeCancelledWhileWaiting(t *testing.T) {
	// 出图并发是新语义：一个任务（一本书）内共用同一个信号量实例。
	taskCtx := withShuihuoImageConcurrency(context.Background(), 1)
	release, err := acquireShuihuoGPTImageGenerationSlot(taskCtx, "gpt-image-2")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	ctx, cancel := context.WithCancel(taskCtx)
	cancel()
	if _, err := acquireShuihuoGPTImageGenerationSlot(ctx, "gpt-image-2"); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("cancelled slot wait error=%v, want context canceled", err)
	}
}

func TestGPTImageGenerationSlotIsReleasedBeforeImageDownload(t *testing.T) {
	imageBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	var server *httptest.Server
	posts := make(chan struct{}, 2)
	downloadStarted := make(chan struct{}, 2)
	allowDownload := make(chan struct{})
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts <- struct{}{}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"url":"` + server.URL + `/generated.png"}]}`))
			return
		}
		downloadStarted <- struct{}{}
		<-allowDownload
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(imageBytes)
	}))
	defer server.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider: "personal_api", ModelID: "gpt-image-2", AspectRatio: "1:1",
		CatalogModel: map[string]any{"endpoint": server.URL + "/v1/images/generations", "credential": "test-secret"},
	}
	var workers sync.WaitGroup
	errs := make(chan error, 2)
	var allowDownloadOnce sync.Once
	releaseDownloads := func() { allowDownloadOnce.Do(func() { close(allowDownload) }) }
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, _, _, err := generateShuihuoImageArtifact(context.Background(), nil, files, nil, "alice", target, "test prompt", nil, nil)
			errs <- err
		}()
	}
	defer releaseDownloads()

	<-posts
	<-downloadStarted
	select {
	case <-posts:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second image request stayed blocked while first image was downloading")
	}

	releaseDownloads()
	workers.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRetryableImageTransportErrorCoversBodyReadFailure(t *testing.T) {
	// 上游已返回 200，但响应体在传输中断：必须重试，不能把已生成的图丢掉。
	for _, message := range []string{
		"image provider response body read failed status=200 refs=3: unexpected EOF",
		"image provider response body read failed status=200 refs=0: context deadline exceeded",
	} {
		if !retryableShuihuoAssetImageTransportError(errors.New(message)) {
			t.Fatalf("expected retryable transport error for %q", message)
		}
	}
	if retryableShuihuoAssetImageTransportError(errors.New("image provider returned HTTP 400: 参考图无法读取")) {
		t.Fatal("non-transport provider error must not be retried")
	}
}

func TestProviderRejectedReferencesDetection(t *testing.T) {
	raw := []byte(`{"error":{"message":"目前无法读取到可用于生成的参考图片。请重新上传人物参考图"}}`)
	if !shuihuoImageProviderRejectedReferences(raw) {
		t.Fatal("reference-image failure was not detected")
	}
	if shuihuoImageProviderRejectedReferences([]byte(`{"error":{"message":"提示词包含违规内容"}}`)) {
		t.Fatal("prompt moderation failure must not be treated as reference-image failure")
	}
}
