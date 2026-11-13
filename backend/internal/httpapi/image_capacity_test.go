package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/imagecapacity"
	"strings"
	"testing"
	"time"
)

func TestCapacityOnlyConfirmsRealImageBytes(t *testing.T) {
	if err := validateCapacityImageBytes([]byte("hello")); err == nil {
		t.Fatal("ordinary text counted as generated image")
	}
	var encoded bytes.Buffer
	png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if err := validateCapacityImageBytes(encoded.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := validateCapacityImageBytes(encoded.Bytes()[:20]); err == nil {
		t.Fatal("truncated PNG accepted")
	}
}

type imageRouteRepo struct{ saved int }

func (s *imageRouteRepo) Resolve(ctx context.Context, r imagecapacity.Request) (imagecapacity.Policy, error) {
	return imagecapacity.Policy{AccountKey: imagecapacity.AccountKey(r), ProfileKey: imagecapacity.ProfileKey(r), Model: r.Model, Size: r.Size, Quality: r.Quality, Resolution: r.Resolution, AccountLimit: 4, ModelLimit: 4, InitialLimit: 4}, nil
}
func (s *imageRouteRepo) Save(context.Context, imagecapacity.Policy) error { s.saved++; return nil }

func TestImageCapacityRoutesAreOwnerScopedAndAdminOnly(t *testing.T) {
	config := NewUserConfigStore(t.TempDir())
	if err := config.write("alice", map[string]any{apiIntegrationsSettingKey: map[string]any{"customChannels": []any{map[string]any{"id": "tested", "group": "image", "model": "gpt-image-2.6-flare", "baseUrl": "https://provider.test/v1", "apiKey": "private-capacity-credential", "params": map[string]any{"quality": "high", "resolution": "2K"}}}}}); err != nil {
		t.Fatal(err)
	}
	repo := &imageRouteRepo{}
	controller := imagecapacity.NewController(repo, func(context.Context, ...string) (any, error) {
		return `{"running":0,"model":0,"waiting":0,"effective":4,"cooling":0}`, nil
	}, "test-route")
	mux := http.NewServeMux()
	now := time.Now()
	auth := BridgeAuth{Secret: "bridge", Now: func() time.Time { return now }}
	registerImageCapacityRoutes(mux, auth, RouterOptions{ImageCapacity: controller, UserConfigStore: config, Store: batchfactoryv11.NewMemoryStore()})
	request := func(method, path, body string, owner bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		SignBridgeRequest(r, "alice", owner, now, "bridge")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	got := request("GET", "/api/config/image-capacity?modelId=custom:tested&aspectRatio=9:16", "", false)
	if got.Code != 200 || strings.Contains(got.Body.String(), "private-capacity-credential") {
		t.Fatalf("GET %d %s", got.Code, got.Body)
	}
	var view map[string]any
	json.Unmarshal(got.Body.Bytes(), &view)
	policy := view["policy"].(map[string]any)
	if view["canManage"] != false || policy["quality"] != "high" || policy["resolution"] != "2K" || policy["size"] != "1024x1536" {
		t.Fatalf("view=%v", view)
	}
	// 并发完全自动判定，保存接口已移除，任何人都不再能手工改并发。
	removed := request("PUT", "/api/config/image-capacity", `{"modelId":"custom:tested","accountLimit":20,"modelLimit":20,"initialLimit":12,"verified":true}`, true)
	if removed.Code != 405 || repo.saved != 0 {
		t.Fatalf("manual concurrency update should be gone: %d saved=%d", removed.Code, repo.saved)
	}
}

func TestImageChannelCapacityPreservesActualRequestParameters(t *testing.T) {
	config := NewUserConfigStore(t.TempDir())
	err := config.write("alice", map[string]any{apiIntegrationsSettingKey: map[string]any{"customChannels": []any{map[string]any{"id": "tested", "group": "image", "name": "tested 2K", "model": "gpt-image-2.6-flare", "baseUrl": "https://provider.test/v1", "apiKey": "secret", "params": map[string]any{"quality": "high", "resolution": "2K"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	entry, matched, err := resolveIntegrationImageCatalogEntry(context.Background(), config, "alice", "custom:tested")
	if err != nil || !matched {
		t.Fatalf("resolve %v", err)
	}
	req, err := newShuihuoProviderImageRequest(context.Background(), "https://provider.test/v1/images/generations", "gpt-image-2.6-flare", "test", "9:16", nil, entry, "")
	if err != nil {
		t.Fatal(err)
	}
	defer req.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["quality"] != "high" || payload["resolution"] != "2K" || payload["size"] != "1024x1536" {
		t.Fatalf("payload=%v", payload)
	}
}

func TestImageCapacityValidationRejectsUserEscalationAndInvalidBounds(t *testing.T) {
	for _, values := range [][3]int{{21, 20, 12}, {20, 21, 12}, {20, 20, 21}, {4, 4, 5}, {0, 4, 4}} {
		if validateImageCapacityLimits(values[0], values[1], values[2]) == nil {
			t.Fatalf("accepted %v", values)
		}
	}
	if err := validateImageCapacityLimits(20, 20, 12); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(imageCapacityQueueHint, "全站") {
		t.Fatal("missing shared queue explanation")
	}
}
