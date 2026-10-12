package giantmaterialplugin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qiantie/backend/internal/giantmaterialresolver"
)

type fakeResolver struct {
	material giantmaterialresolver.Material
	err      error
}

func (f fakeResolver) Resolve(context.Context, string) (giantmaterialresolver.Material, error) {
	return f.material, f.err
}

func TestHandlerResolvesMaterialWithPluginToken(t *testing.T) {
	handler := NewHandler("plugin-token", fakeResolver{material: giantmaterialresolver.Material{
		MaterialID: "material-1", VideoURL: "https://material.hnqingyuwen.top/video.mp4", DurationSeconds: 3,
	}})
	req := httptest.NewRequest(http.MethodPost, "/v1/giant-material/resolve", strings.NewReader(`{"giantMaterialId":"7689285397448523826"}`))
	req.Header.Set("Authorization", "Bearer plugin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Material giantmaterialresolver.Material `json:"material"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Material.GiantMaterialID != "7689285397448523826" || body.Material.MaterialID != "material-1" {
		t.Fatalf("material=%#v", body.Material)
	}
}

func TestHandlerRejectsMissingPluginToken(t *testing.T) {
	handler := NewHandler("plugin-token", fakeResolver{})
	req := httptest.NewRequest(http.MethodPost, "/v1/giant-material/resolve", strings.NewReader(`{"giantMaterialId":"7689285397448523826"}`))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestNewQingyuHandlerValidatesRequestBeforeCallingUpstream(t *testing.T) {
	handler := NewQingyuHandler("plugin-token", "https://n8.hnqingyuwen.top/center-api/material/video/select", "upstream-token")
	req := httptest.NewRequest(http.MethodPost, "/v1/giant-material/resolve", strings.NewReader(`{"giantMaterialId":"invalid"}`))
	req.Header.Set("Authorization", "Bearer plugin-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}
