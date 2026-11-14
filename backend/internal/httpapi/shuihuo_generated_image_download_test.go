package httpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
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

func TestProviderImageGenerationRequestUsesURLAndDownloadsResult(t *testing.T) {
	const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="
	pngBytes, err := base64.StdEncoding.DecodeString(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	var (
		mu        sync.Mutex
		bodies    []string
		serverURL string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/result.png" {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngBytes)
			return
		}
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"url":"` + serverURL + `/result.png"}]}`))
	}))
	defer server.Close()
	serverURL = server.URL

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
	if len(bodies) != 1 {
		t.Fatalf("requests=%d, want one image generation request", len(bodies))
	}
	if strings.Contains(bodies[0], `"response_format"`) {
		t.Fatalf("initial generation request must preserve the working gateway default: %s", bodies[0])
	}
}

func TestProviderImageEditRequestUsesBase64Directly(t *testing.T) {
	reference := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4e, 0x47})
	req, err := newShuihuoImageEditRequest(context.Background(), "https://images.example/v1/images/generations", "gpt-image-2", "test prompt", "1024x1536", []string{reference})
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `name="response_format"`) || !strings.Contains(string(body), "b64_json") {
		t.Fatalf("edit request must directly ask for base64 so a generated image does not depend on a second CDN download: %s", body)
	}

	base64Req, err := newShuihuoImageEditRequestWithResponseFormat(context.Background(), "https://images.example/v1/images/generations", "gpt-image-2", "test prompt", "1024x1536", []string{reference}, "b64_json")
	if err != nil {
		t.Fatal(err)
	}
	base64Body, err := io.ReadAll(base64Req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(base64Body), `name="response_format"`) || !strings.Contains(string(base64Body), "b64_json") {
		t.Fatalf("fallback edit request must ask for base64: %s", base64Body)
	}
}

func TestProviderImageEditRequestFallsBackWhenGatewayRejectsResponseFormat(t *testing.T) {
	const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="
	var (
		mu      sync.Mutex
		formats []string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		format := r.FormValue("response_format")
		mu.Lock()
		formats = append(formats, format)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if format != "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"unknown parameter: response_format"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + tinyPNG + `"}]}`))
	}))
	defer server.Close()

	target := shuihuoAssetImageTarget{
		Provider:    "personal_api",
		ModelID:     "gpt-image-2",
		AspectRatio: "1:1",
		CatalogModel: map[string]any{
			"supportsImageEdits": true,
		},
	}
	reference := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4e, 0x47})
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	if _, _, err := requestShuihuoProviderImage(context.Background(), files, target, target.CatalogModel, server.URL+"/v1/images/generations", "test-secret", "test prompt", []string{reference}); err != nil {
		t.Fatalf("compatibility fallback failed: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(formats) != 2 || formats[0] != "b64_json" || formats[1] != "" {
		t.Fatalf("response formats=%v, want [b64_json <omitted>]", formats)
	}
}

func TestProviderImageUsesCompleteJSONBeforeResetWithoutResubmitting(t *testing.T) {
	const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="
	png, err := base64.StdEncoding.DecodeString(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/images/generations":
			conn, rw, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijack: %v", err)
				return
			}
			body := []byte(`{"data":[{"url":"` + serverURL + `/result.png"}]}`)
			_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nTransfer-Encoding: chunked\r\n\r\n")
			_, _ = rw.WriteString(fmt.Sprintf("%x\r\n", len(body)))
			_, _ = rw.Write(body)
			_, _ = rw.WriteString("\r\n")
			_ = rw.Flush()
			_ = conn.Close()
		case "/result.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(png)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

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
	mediaType, storageRef, err := requestShuihuoProviderImage(context.Background(), files, target, target.CatalogModel, server.URL+"/v1/images/generations", "test-secret", "test prompt", nil)
	if err != nil {
		t.Fatalf("complete JSON received before reset must be used: %v", err)
	}
	if mediaType != "image/png" || storageRef == "" {
		t.Fatalf("mediaType=%q storageRef=%q", mediaType, storageRef)
	}
}

// 上游已返回图片（HTTP 200，按次计费）但本地磁盘写满时，必须返回「已计费
// 未落盘」错误，且资产级重试不得再次请求上游。
func TestProviderImageDiskPressureDoesNotResubmitAfterBilling(t *testing.T) {
	const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="
	png, err := base64.StdEncoding.DecodeString(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	var requests int32
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/images/generations":
			atomic.AddInt32(&requests, 1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[{"url":"` + serverURL + `/result.png"}]}`))
		case "/result.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(png)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	files.SetWriteGuard(func() error {
		return fmt.Errorf("%w: 94.0%% used", localartifact.ErrDiskPressure)
	})
	target := shuihuoAssetImageTarget{Provider: "personal_api", ModelID: "gpt-image-2", AspectRatio: "1:1", CatalogModel: map[string]any{}}
	_, _, err = requestShuihuoProviderImage(context.Background(), files, target, target.CatalogModel, serverURL+"/v1/images/generations", "test-secret", "test prompt", nil)
	if err == nil {
		t.Fatal("磁盘保护拒绝落盘时必须返回失败")
	}
	var billed *shuihuoImageBilledSaveFailure
	if !errors.As(err, &billed) {
		t.Fatalf("err=%v, want shuihuoImageBilledSaveFailure", err)
	}
	if !errors.Is(err, localartifact.ErrDiskPressure) {
		t.Fatalf("错误链丢失磁盘压力语义: %v", err)
	}
	if retryableShuihuoAssetImageError(err) {
		t.Fatal("已计费未落盘不得自动重发")
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("upstream requests=%d, want 1", got)
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

func TestShuihuoImageGenerationHTTPClientIsolatesGGWKHTTP2Connections(t *testing.T) {
	first := shuihuoImageGenerationHTTPClient("https://www.ggwk1.online/v1/images/generations")
	second := shuihuoImageGenerationHTTPClient("https://www.ggwk1.online/v1/images/generations")
	if first == second || first.Transport == second.Transport {
		t.Fatal("ggwk image requests must not share an HTTP/2 connection")
	}
	for _, client := range []*http.Client{first, second} {
		transport := client.Transport.(*http.Transport)
		if !transport.ForceAttemptHTTP2 || !transport.DisableKeepAlives {
			t.Fatal("ggwk image requests must use one isolated HTTP/2 connection")
		}
		// 该网关常用 200 响应头 + 数分钟慢正文。客户端一旦带 120s 总超时，
		// 正文会在传输中被掐断，表现为「HTTP 200 但图片取不回」。
		if client.Timeout != 0 {
			t.Fatalf("ggwk image client must not carry an absolute timeout, got %s", client.Timeout)
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

func TestGPTImageConcurrencyIsProcessWideAcrossTasks(t *testing.T) {
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
	// 设计意图：出图并发是「进程级 + 按图片域名共享」的，不是每本书各算一份。
	// 同一图片域名冷启动只放 shuihuoAdaptiveInitialLimit（当前 2）路，连续成功
	// 后才按自适应阶梯升到调用方配置值；所以两本书各自配置 2 路时，同一域名
	// 同时最多仍是 2 个请求，而不是 4 个，避免多个批次把慢速 Base64 响应压垮。
	// 以后要调整并发策略，改 backend/internal/httpapi/shuihuo_image_adaptive_concurrency.go
	// 的冷启动初值，以及 backend/internal/httpapi/shuihuo_projects.go 里的
	// shuihuoImageConcurrencyDefault，并同步更新本用例断言；只调大调用方的
	// withShuihuoImageConcurrency 不会突破这里的进程级上限。
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
	if got := maximum.Load(); got != shuihuoAdaptiveInitialLimit {
		t.Fatalf("maximum simultaneous gpt-image requests=%d, want process-wide cold-start %d", got, shuihuoAdaptiveInitialLimit)
	}
	if maximum.Load() > int32(shuihuoImageConcurrencyDefault) {
		t.Fatalf("cold-start concurrency=%d exceeds configured default %d", maximum.Load(), shuihuoImageConcurrencyDefault)
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

func TestRetryableImageTransportErrorUsesTypesNotMessageText(t *testing.T) {
	// 上游已返回 200，但响应体在传输中断：底层类型必须被判为可重试，
	// 不能再依赖错误文案（文案已中文化，关键词会漏判）。
	if !retryableShuihuoAssetImageTransportError(fmt.Errorf("图片通道响应读取中断（HTTP 200）：%w", io.ErrUnexpectedEOF)) {
		t.Fatal("unexpected EOF 类型必须可重试")
	}
	if !retryableShuihuoAssetImageTransportError(fmt.Errorf("图片通道请求失败：%w", context.DeadlineExceeded)) {
		t.Fatal("deadline exceeded 类型必须可重试")
	}
	// 只有文案、没有传输层类型时不再重试（旧实现靠英文关键词，会误判）。
	if retryableShuihuoAssetImageTransportError(errors.New("image provider response body read failed status=200: unexpected EOF")) {
		t.Fatal("纯文案错误不应再被判定为可重试")
	}
	// 业务拒绝带明确 HTTP 状态但没有传输层类型：不重试。
	if retryableShuihuoAssetImageTransportError(shuihuoImageUpstreamError{Message: "图片通道拒绝了本次请求（HTTP 400）", StatusCode: http.StatusBadRequest}) {
		t.Fatal("4xx 业务拒绝必须不可重试")
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
