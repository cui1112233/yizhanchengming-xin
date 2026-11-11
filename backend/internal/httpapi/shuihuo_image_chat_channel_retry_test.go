package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

const testChatChannelImageReply = "目前我无法在这个聊天中直接生成图片，请使用支持图像生成的模型或工具。"

const testTinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="

func TestShuihuoImageProviderChatChannelResponseDetection(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"chinese chat reply 400", http.StatusBadRequest, testChatChannelImageReply, true},
		{"english chat reply 400", http.StatusBadRequest, "I can't generate images in this chat.", true},
		{"json chat reply 400", http.StatusBadRequest, `{"error":{"message":"目前无法在此聊天中直接生成图片"}}`, true},
		{"json error 400", http.StatusBadRequest, `{"error":{"message":"bad request"}}`, false},
		{"plain non-chat 400", http.StatusBadRequest, "invalid request", false},
		{"chat reply 200", http.StatusOK, testChatChannelImageReply, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(tc.body)
			if got := shuihuoImageProviderChatChannelResponse(tc.status, body); got != tc.want {
				t.Fatalf("detected=%t want=%t body=%q", got, tc.want, tc.body)
			}
		})
	}
}

func TestGenerateShuihuoAssetImageArtifactRetriesChatChannelReplies(t *testing.T) {
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "3")
	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	attempts := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(testChatChannelImageReply))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + testTinyPNGBase64 + `"}]}`))
	}))
	defer provider.Close()

	files := localartifact.NewStore(t.TempDir(), 1<<20)
	target := shuihuoAssetImageTarget{
		Provider: "personal_api", ModelID: "comic-image", AspectRatio: "1:1",
		CatalogModel: map[string]any{"endpoint": provider.URL + "/v1/images/generations", "credential": "test-secret"},
	}
	mediaType, storageRef, source, err := generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, files, nil, "alice", target, "三格竖版漫画页", nil, nil)
	if err != nil {
		t.Fatalf("chat-channel 400 must be retried until success: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("provider attempts=%d, want 3", attempts)
	}
	// source 记录图片来源；上游直出的图按既有约定为 "provider"，不是空串。
	if mediaType == "" || storageRef == "" || source != "provider" {
		t.Fatalf("mediaType=%q storageRef=%q source=%q", mediaType, storageRef, source)
	}
}

func TestGenerateShuihuoAssetImageArtifactChatChannelFailureKeepsUpstreamText(t *testing.T) {
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "2")
	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	attempts := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(testChatChannelImageReply))
	}))
	defer provider.Close()

	target := shuihuoAssetImageTarget{
		Provider: "personal_api", ModelID: "comic-image", AspectRatio: "1:1",
		CatalogModel: map[string]any{"endpoint": provider.URL + "/v1/images/generations", "credential": "test-secret"},
	}
	_, _, _, err := generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, localartifact.NewStore(t.TempDir(), 1<<20), nil, "alice", target, "三格竖版漫画页", nil, nil)
	if err == nil {
		t.Fatal("chat-channel failure must surface after the retry budget")
	}
	if attempts != 2 {
		t.Fatalf("provider attempts=%d, want 2 (first try + 1 retry)", attempts)
	}
	if !errors.Is(err, errShuihuoImageChatChannel) {
		t.Fatalf("error must stay classifiable: %v", err)
	}
	var chatErr shuihuoImageChatChannelError
	if !errors.As(err, &chatErr) || chatErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("typed chat-channel error missing: %v", err)
	}
	for _, part := range []string{"HTTP 400", "在这个聊天中直接生成图片"} {
		if !strings.Contains(err.Error(), part) {
			t.Fatalf("terminal error misses upstream text %q: %v", part, err)
		}
	}
}

func TestShuihuoImageUpstreamErrorKeepsRawBelowChineseHeadline(t *testing.T) {
	err := shuihuoImageUpstreamError{
		Message: "图片通道拒绝了本次请求（HTTP 400），当前配置的模型可能不支持出图",
		Raw:     testChatChannelImageReply,
	}
	headline, raw, found := strings.Cut(err.Error(), "\n上游原文：")
	if !found {
		t.Fatalf("raw upstream must be on its own line: %q", err.Error())
	}
	if headline != "图片通道拒绝了本次请求（HTTP 400），当前配置的模型可能不支持出图" {
		t.Fatalf("headline=%q", headline)
	}
	for _, english := range []string{"temporary upstream failure", "image provider returned", "invalid input"} {
		if strings.Contains(headline, english) {
			t.Fatalf("headline leaks english sentinel %q: %q", english, headline)
		}
	}
	if !strings.Contains(raw, "在这个聊天中直接生成图片") {
		t.Fatalf("raw upstream text missing: %q", raw)
	}
	if !errors.Is(err, batchfactoryv11.ErrUpstream) {
		t.Fatalf("error must stay classifiable: %v", err)
	}
}

func TestShuihuoImageUpstreamErrorDropsLeadingColonInRaw(t *testing.T) {
	err := shuihuoImageUpstreamError{
		Message: "图片通道鉴权失败（HTTP 401），请检查渠道 Key 是否有效、是否具备出图权限",
		Raw:     ": Incorrect API key provided: wrong-key",
	}
	got := err.Error()
	if strings.Contains(got, "上游原文：:") {
		t.Fatalf("raw upstream must not start with a second colon: %q", got)
	}
	if !strings.Contains(got, "上游原文：Incorrect API key provided: wrong-key") {
		t.Fatalf("raw upstream text missing: %q", got)
	}
}

func TestComicPageGenerateChatChannelRetryWritesSinglePageAndImage(t *testing.T) {
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "3")
	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	now := time.Unix(1700000000, 0)
	attempts := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(testChatChannelImageReply))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"` + testTinyPNGBase64 + `"}]}`))
	}))
	defer provider.Close()

	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	raw := func(value any) json.RawMessage {
		encoded, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		return encoded
	}
	if _, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{
		"productionMode":   raw("comic"),
		"imageAspectRatio": raw("1:1"),
		"imageModelId":     raw("comic-image"),
		"imageProvider":    raw("personal_api"),
		"aiPromptConfig": raw(map[string]any{"comic": map[string]any{
			"enabled":           true,
			"style":             map[string]any{"presetId": "batch-comic-style-chat-retry", "body": "水墨漫画画风"},
			"imagePromptPrefix": "统一前缀",
			"imagePromptSuffix": "统一后缀",
		}}),
	}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(context.Background(), "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "comic-image", "kind": "image", "enabled": true, "provider": "personal_api", "modelId": "comic-provider-v1", "baseUrl": provider.URL, "credential": "test-secret"},
	}}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret:    "secret",
		Now:             func() time.Time { return now },
		Slice:           1,
		Store:           store,
		LocalArtifacts:  localartifact.NewStore(t.TempDir(), 1<<20),
		NovelFetchStore: catalog,
	})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + video.ID + "/generate"
	generated := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"comicPrompt": "三格导演说明：上格人物特写，中格手机界面，下格浴室全景",
		"imagePrompt": "三格竖版条漫，人物一致，第一页完整画面",
		"aspectRatio": "16:9",
	})
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	result := decodeBody[struct {
		Page  batchfactoryv11.ComicPage      `json:"page"`
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, generated)
	if attempts != 3 {
		t.Fatalf("provider attempts=%d, want 3", attempts)
	}
	if result.Page.VideoID != video.ID || result.Image.ComicPageID != result.Page.ID || result.Image.URL == "" {
		t.Fatalf("result=%+v", result)
	}

	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages", nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, listed).Pages
	if len(pages) != 1 {
		t.Fatalf("retries must not duplicate page records: %+v", pages)
	}
	if len(pages[0].Images) != 1 || !pages[0].Images[0].IsPrimary {
		t.Fatalf("retries must not duplicate or replace the successful image: %+v", pages[0].Images)
	}
}
