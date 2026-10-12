package giantmaterialresolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPPluginClientResolveUsesVersionedContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/giant-material/resolve" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer plugin-token" {
			t.Fatalf("Authorization = %q", got)
		}
		var input struct {
			GiantMaterialID string `json:"giantMaterialId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.GiantMaterialID != "7689285397448523826" {
			t.Fatalf("giantMaterialId = %q", input.GiantMaterialID)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"material": Material{
			MaterialID: "material-1", GiantMaterialID: input.GiantMaterialID,
			VideoURL: "https://material.hnqingyuwen.top/video.mp4", DurationSeconds: 3,
		}})
	}))
	defer server.Close()

	client := NewHTTPPluginClient(server.Client(), server.URL, "plugin-token")
	material, err := client.Resolve(context.Background(), "7689285397448523826")
	if err != nil {
		t.Fatal(err)
	}
	if material.MaterialID != "material-1" || material.GiantMaterialID != "7689285397448523826" {
		t.Fatalf("material = %#v", material)
	}
}

func TestHTTPPluginClientResolveMapsPluginUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": "PLUGIN_UNAVAILABLE"})
	}))
	defer server.Close()

	client := NewHTTPPluginClient(server.Client(), server.URL, "plugin-token")
	_, err := client.Resolve(context.Background(), "7689285397448523826")
	if got := CodeOf(err); got != "GIANT_PLUGIN_UNAVAILABLE" {
		t.Fatalf("code = %q", got)
	}
}
