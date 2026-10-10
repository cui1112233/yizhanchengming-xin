package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestV12FetchOriginalsUsesGoNovelFetcherAndKeepsRawText(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("bookid") != "book-1" || r.URL.Query().Get("platform") != "15" || r.URL.Query().Get("max_txt") != "100" {
			t.Fatalf("query=%s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"成功","data":"完整正文超过限制","bookinfo":{"category":"女频","genre":"现代"}}`))
	}))
	defer upstream.Close()
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now },
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
	})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{
		"platform": 15, "bookIds": []string{"book-1"}, "maxTxt": 100,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"\"status\":\"ok\"", "完整正文超过限制", "\"gender\":\"女频\"", "\"rawData\""} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
}

func TestV12FetchOriginalsRejectsInvalidInput(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{}})
	rec := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/fetch-originals", map[string]any{"platform": 15, "bookIds": []string{}, "maxTxt": 100})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
