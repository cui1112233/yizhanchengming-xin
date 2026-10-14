package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestScriptVideoRoutesSubmitAndPollH3KeepLegacyContract(t *testing.T) {
	var calls int
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"task_id":"h3-task-1","status":"queued"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"SUCCESS","url":"https://cdn.example/video.mp4"}`))
	}))
	defer provider.Close()
	client := provider.Client()
	transport := client.Transport
	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		copyReq := req.Clone(req.Context())
		parsed, err := url.Parse(provider.URL + req.URL.Path)
		if err != nil {
			return nil, err
		}
		copyReq.URL = parsed
		return transport.RoundTrip(copyReq)
	})
	resolver := func(context.Context, string, string) (batchfactoryv11.VideoProviderConfig, error) {
		return batchfactoryv11.VideoProviderConfig{
			Provider:   batchfactoryv11.VideoProviderAutoDLH3,
			APIKey:     "secret",
			Model:      batchfactoryv11.AutoDLH3Model,
			CreateURL:  "https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}",
			TasksURL:   "https://autodl.art/api/v1/comfyui/comfyui_workflow/result/{id}",
			WorkflowID: batchfactoryv11.AutoDLH3NoImageWorkflow,
		}, nil
	}
	mux := http.NewServeMux()
	RegisterScriptVideoRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return time.Unix(100, 0) }}, client, resolver)
	post := httptest.NewRequest(http.MethodPost, "/api/script-video", strings.NewReader(`{"modelKey":"minimax-h3-video","prompt":"镜头"}`))
	SignBridgeRequest(post, "tester", false, time.Unix(100, 0), "secret")
	postResult := httptest.NewRecorder()
	mux.ServeHTTP(postResult, post)
	if postResult.Code != http.StatusAccepted {
		t.Fatalf("submit status = %d, body=%s", postResult.Code, postResult.Body.String())
	}
	var submitted map[string]any
	_ = json.NewDecoder(postResult.Body).Decode(&submitted)
	if submitted["taskId"] != "h3:h3-task-1" || submitted["provider"] != batchfactoryv11.VideoProviderAutoDLH3 {
		t.Fatalf("submit body = %#v", submitted)
	}
	get := httptest.NewRequest(http.MethodGet, "/api/script-video/h3:h3-task-1", nil)
	SignBridgeRequest(get, "tester", false, time.Unix(100, 0), "secret")
	getResult := httptest.NewRecorder()
	mux.ServeHTTP(getResult, get)
	if getResult.Code != http.StatusOK || !strings.Contains(getResult.Body.String(), `"status":"succeeded"`) {
		t.Fatalf("poll status=%d body=%s calls=%d", getResult.Code, getResult.Body.String(), calls)
	}
	download := httptest.NewRequest(http.MethodGet, "/api/script-video/h3:h3-task-1/download", nil)
	SignBridgeRequest(download, "tester", false, time.Unix(100, 0), "secret")
	downloadResult := httptest.NewRecorder()
	mux.ServeHTTP(downloadResult, download)
	if downloadResult.Code != http.StatusFound || downloadResult.Header().Get("Location") != "https://cdn.example/video.mp4" {
		t.Fatalf("download status=%d location=%q body=%s", downloadResult.Code, downloadResult.Header().Get("Location"), downloadResult.Body.String())
	}
}

func TestScriptVideoRoutesRejectReferenceImagesForLocalModel(t *testing.T) {
	mux := http.NewServeMux()
	RegisterScriptVideoRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return time.Unix(100, 0) }}, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/script-video", strings.NewReader(`{"modelKey":"local-doubao-executor-video","prompt":"镜头","imageUrls":["https://cdn.example/a.png"]}`))
	SignBridgeRequest(req, "tester", false, time.Unix(100, 0), "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusConflict || !strings.Contains(res.Body.String(), "UNSUPPORTED_REFERENCE_IMAGES") {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}
