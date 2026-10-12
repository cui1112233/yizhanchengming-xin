package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/giantmaterialresolver"
)

type fakeGiantMaterialResolver struct {
	material giantmaterialresolver.Material
	err      error
}

func (f fakeGiantMaterialResolver) Resolve(context.Context, string) (giantmaterialresolver.Material, error) {
	return f.material, f.err
}

func TestGiantMaterialResolverRouteReturnsNormalizedMaterialWithoutToken(t *testing.T) {
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret",
		Now:          func() time.Time { return time.Unix(100, 0) },
		GiantMaterialResolver: fakeGiantMaterialResolver{material: giantmaterialresolver.Material{
			MaterialID: "10122315", VideoURL: "https://material.hnqingyuwen.top/a.mp4", DurationSeconds: 3,
			Books: []giantmaterialresolver.Book{{PlatformBookID: "748725", PlatformName: "七猫", Title: "测试书"}},
		}},
	})
	recorder := signedJSONRequest(t, api, time.Unix(100, 0), "alice", http.MethodPost, "/api/shuihuo-production/giant-material-resolve", map[string]any{"giantMaterialId": "7689285397448523826"})
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if containsAny(recorder.Body.String(), "N8-Admin-Token", "server-token") {
		t.Fatalf("credential leaked: %s", recorder.Body.String())
	}
}

func TestGiantMaterialResolverErrorSeparatesPluginFailureKinds(t *testing.T) {
	for _, tc := range []struct {
		code string
		want int
	}{
		{code: "GIANT_PLUGIN_UNAVAILABLE", want: http.StatusServiceUnavailable},
		{code: "GIANT_PLUGIN_TIMEOUT", want: http.StatusGatewayTimeout},
		{code: "GIANT_PLUGIN_UPSTREAM_FAILED", want: http.StatusBadGateway},
		{code: "GIANT_PLUGIN_ENDPOINT_INVALID", want: http.StatusInternalServerError},
	} {
		t.Run(tc.code, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeGiantMaterialResolverError(recorder, &giantmaterialresolver.CodedError{Code: tc.code})
			if recorder.Code != tc.want {
				t.Fatalf("status = %d, want %d body=%s", recorder.Code, tc.want, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tc.code) {
				t.Fatalf("body = %s", recorder.Body.String())
			}
		})
	}
}
