package httpapi

import (
	"context"
	json "encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestNovelFetchWorkshopUsesConfiguredGenderJudgeResolver(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","data":"原文","bookinfo":{"category":"男频","genre":"女频"}}`))
	}))
	defer upstream.Close()
	judgeCalls := 0
	handler := NewRouter(RouterOptions{
		BridgeSecret:    secret,
		Now:             func() time.Time { return now },
		NovelFetchStore: novelfetchworkshop.NewMemoryStore(),
		NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{
			Client:   upstream.Client(),
			Endpoint: upstream.URL,
		},
		NovelFetchAIGenderJudgeResolver: func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error) {
			return func(context.Context, novelfetchworkshop.BookInfo) (string, error) {
				judgeCalls++
				return "女频", nil
			}, nil
		},
	})
	process := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/process", map[string]any{
		"inputText": "12345678906\t冲突分类测试\t待确认",
	}, now, secret)
	if process.Code != http.StatusOK {
		t.Fatalf("process status=%d body=%s", process.Code, process.Body.String())
	}
	fetch := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/tasks/12345678906/fetch", map[string]any{"maxTxt": 4000}, now, secret)
	if fetch.Code != http.StatusOK {
		t.Fatalf("fetch status=%d body=%s", fetch.Code, fetch.Body.String())
	}
	var response struct {
		Task novelfetchworkshop.Document `json:"task"`
	}
	if err := json.Unmarshal(fetch.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if judgeCalls != 1 || response.Task.Gender != "女频" || response.Task.GenderSource != novelfetchworkshop.GenderAI {
		t.Fatalf("gender judge was not used: calls=%d task=%#v", judgeCalls, response.Task)
	}
}

func TestNovelFetchWorkshopGenderFieldsRoundTrip(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})
	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-gender", map[string]any{
		"gender": "女频", "genderSource": "manual", "genderReason": "用户手工填写",
		"bookInfo": map[string]string{"category": "男频", "genre": "都市"},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}
	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-gender", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document novelfetchworkshop.Document
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document.Gender != "女频" || document.GenderSource != "manual" || document.BookInfo.Category != "男频" || document.BookInfo.Genre != "都市" {
		t.Fatalf("unexpected document=%#v", document)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(get.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	var bookInfo map[string]json.RawMessage
	if err := json.Unmarshal(raw["bookInfo"], &bookInfo); err != nil {
		t.Fatal(err)
	}
	if _, ok := bookInfo["category"]; !ok {
		t.Fatalf("bookInfo.category missing from raw JSON: %s", get.Body.String())
	}
	if _, ok := bookInfo["genre"]; !ok {
		t.Fatalf("bookInfo.genre missing from raw JSON: %s", get.Body.String())
	}
	if _, ok := bookInfo["Category"]; ok {
		t.Fatalf("legacy BookInfo.Category key present: %s", get.Body.String())
	}
}
