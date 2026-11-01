package httpapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/localartifact"
)

func TestDownloadShuihuoGeneratedImageRetriesTemporaryProviderFailure(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
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
	if got := attempts.Load(); got != 3 {
		t.Fatalf("attempts=%d, want 3", got)
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
	vveai := shuihuoImageGenerationHTTPClient("https://api.vveai.com/v1/images/generations")
	vveaiTransport := vveai.Transport.(*http.Transport)
	if !vveaiTransport.ForceAttemptHTTP2 || vveaiTransport.TLSNextProto != nil {
		t.Fatal("VVEAI endpoint must use the HTTP/2-capable transport")
	}

	other := shuihuoImageGenerationHTTPClient("https://images.example.test/v1/images/generations")
	otherTransport := other.Transport.(*http.Transport)
	if otherTransport.ForceAttemptHTTP2 || otherTransport.TLSNextProto == nil {
		t.Fatal("other image gateways must retain the HTTP/1.1 transport")
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
	var workers sync.WaitGroup
	errs := make(chan error, requests)
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			// 统一配置设成 1 路时必须保持串行。
			ctx := withShuihuoImageConcurrency(context.Background(), 1)
			_, _, _, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "test prompt", nil, nil)
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
	var workers sync.WaitGroup
	errs := make(chan error, requests)
	for range requests {
		workers.Add(1)
		go func() {
			defer workers.Done()
			ctx := withShuihuoImageConcurrency(context.Background(), 2)
			_, _, _, err := generateShuihuoImageArtifact(ctx, nil, files, nil, "alice", target, "test prompt", nil, nil)
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

func TestGPTImageGenerationSlotCanBeCancelledWhileWaiting(t *testing.T) {
	release, err := acquireShuihuoGPTImageGenerationSlot(withShuihuoImageConcurrency(context.Background(), 1), "gpt-image-2")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	ctx, cancel := context.WithCancel(withShuihuoImageConcurrency(context.Background(), 1))
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
