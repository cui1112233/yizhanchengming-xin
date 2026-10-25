package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

func comicTestRaw(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestComicPageRoutesKeepCandidatesSeparateFromBookAssets(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{
			Title: "第一本",
			Videos: []batchfactoryv11.CreateVideoInput{
				{Label: "分镜一", VideoPrompt: "视频一"},
				{Label: "分镜二", VideoPrompt: "视频二"},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	api := NewRouter(RouterOptions{
		BridgeSecret:   "secret",
		Now:            func() time.Time { return now },
		Slice:          1,
		Store:          store,
		LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024),
	})
	pagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"

	created := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+book.Videos[0].ID, map[string]any{
		"comicPrompt": "第一张完整漫画页",
		"imagePrompt": "三格竖版漫画，第一张完整漫画页",
		"modelId":     "seedream-5-flash-local",
		"aspectRatio": "3:4",
	})
	if created.Code != http.StatusOK {
		t.Fatalf("create page status=%d body=%s", created.Code, created.Body.String())
	}
	page := decodeBody[struct {
		Page batchfactoryv11.ComicPage `json:"page"`
	}](t, created).Page
	if page.VideoID != book.Videos[0].ID || page.ImagePrompt != "三格竖版漫画，第一张完整漫画页" {
		t.Fatalf("page=%+v", page)
	}

	dataURL := "data:image/png;base64,iVBORw0KGgoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+page.ID+"/images/upload", map[string]any{"dataUrl": dataURL})
	if first.Code != http.StatusCreated {
		t.Fatalf("first upload status=%d body=%s", first.Code, first.Body.String())
	}
	firstImage := decodeBody[struct {
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, first).Image
	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+page.ID+"/images", map[string]any{
		"url":       "https://images.example/candidate-2.png",
		"mediaType": "image/png",
		"source":    "provider",
	})
	if second.Code != http.StatusCreated {
		t.Fatalf("second candidate status=%d body=%s", second.Code, second.Body.String())
	}
	secondImage := decodeBody[struct {
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, second).Image
	if !firstImage.IsPrimary || secondImage.IsPrimary {
		t.Fatalf("primary state first=%t second=%t", firstImage.IsPrimary, secondImage.IsPrimary)
	}

	swapped := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+page.ID+"/images/"+secondImage.ID+"/primary", map[string]any{})
	if swapped.Code != http.StatusOK {
		t.Fatalf("set primary status=%d body=%s", swapped.Code, swapped.Body.String())
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, pagesPath, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, listed).Pages
	if len(pages) != 1 || len(pages[0].Images) != 2 {
		t.Fatalf("pages=%+v", pages)
	}
	if pages[0].Images[0].URL == "" {
		t.Fatalf("stored candidate must expose owned content URL: %+v", pages[0].Images)
	}
	assets, err := store.ListBookAssets(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 0 {
		t.Fatalf("comic candidates leaked into book assets: %+v", assets)
	}
}

func TestComicCollectionRouteRejectsMissingPrimaryAndSnapshotsAllPrimaryImages(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "一"}, {Label: "二"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	files := localartifact.NewStore(t.TempDir(), 1024)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: files})
	pagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"
	collectionsPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-collections"
	dataURL := "data:image/png;base64,iVBORw0KGgoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	for index, video := range book.Videos {
		created := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+video.ID, map[string]any{
			"comicPrompt": "漫画页",
			"imagePrompt": "漫画页直接出图提示词",
		})
		if created.Code != http.StatusOK {
			t.Fatalf("page %d status=%d body=%s", index, created.Code, created.Body.String())
		}
		page := decodeBody[struct {
			Page batchfactoryv11.ComicPage `json:"page"`
		}](t, created).Page
		if index == 0 {
			uploaded := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+page.ID+"/images/upload", map[string]any{"dataUrl": dataURL})
			if uploaded.Code != http.StatusCreated {
				t.Fatalf("upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
			}
		}
	}

	blocked := signedJSONRequest(t, api, now, "alice", http.MethodPost, collectionsPath, map[string]any{})
	if blocked.Code != http.StatusConflict {
		t.Fatalf("missing primary status=%d body=%s", blocked.Code, blocked.Body.String())
	}
	missing := decodeBody[struct {
		MissingVideoIDs []string `json:"missingVideoIds"`
	}](t, blocked).MissingVideoIDs
	if len(missing) != 1 || missing[0] != book.Videos[1].ID {
		t.Fatalf("missing=%v", missing)
	}

	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, signedJSONRequest(t, api, now, "alice", http.MethodGet, pagesPath, nil)).Pages
	uploaded := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+pages[1].ID+"/images/upload", map[string]any{"dataUrl": dataURL})
	if uploaded.Code != http.StatusCreated {
		t.Fatalf("second upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
	}
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, collectionsPath, map[string]any{})
	if created.Code != http.StatusCreated {
		t.Fatalf("collection status=%d body=%s", created.Code, created.Body.String())
	}
	collection := decodeBody[struct {
		Collection batchfactoryv11.ComicCollection `json:"collection"`
	}](t, created).Collection
	if collection.Status != "ready" || len(collection.Items) != 2 || collection.Items[0].Ordinal != 1 || collection.Items[1].Ordinal != 2 {
		t.Fatalf("collection=%+v", collection)
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, collectionsPath, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list collections status=%d body=%s", listed.Code, listed.Body.String())
	}
}

func TestComicPageGenerateUsesPagePromptAndStoresCandidateOutsideAssets(t *testing.T) {
	now := time.Unix(1700000000, 0)
	var providerRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Size   string `json:"size"`
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&providerRequest); err != nil {
			t.Errorf("decode provider request: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
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
	_, err = store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{
		"productionMode": comicTestRaw(t, "comic"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
			"enabled":           true,
			"style":             map[string]any{"presetId": "batch-comic-style-14", "body": "水墨漫画画风"},
			"imagePromptPrefix": "统一前缀",
			"imagePromptSuffix": "统一后缀",
		}}),
	}, ExpectedRevision: batch.Revision})
	if err != nil {
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
		"comicPrompt": "第一页剧情说明",
		"imagePrompt": "三格竖版条漫，人物一致，第一页完整画面",
		"modelId":     "comic-image",
		"aspectRatio": "3:4",
	})
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	result := decodeBody[struct {
		Page  batchfactoryv11.ComicPage      `json:"page"`
		Image batchfactoryv11.ComicPageImage `json:"image"`
	}](t, generated)
	for _, part := range []string{"水墨漫画画风", "统一前缀", "阅读顺序", "三格竖版条漫，人物一致，第一页完整画面", "统一后缀"} {
		if !strings.Contains(providerRequest.Prompt, part) {
			t.Fatalf("provider prompt misses %q: %s", part, providerRequest.Prompt)
		}
	}
	if providerRequest.Model != "comic-provider-v1" || providerRequest.Size != "1536x2048" {
		t.Fatalf("provider request=%+v", providerRequest)
	}
	if result.Page.VideoID != video.ID || result.Image.ComicPageID != result.Page.ID || !result.Image.IsPrimary || result.Image.URL == "" {
		t.Fatalf("result=%+v", result)
	}
	assets, err := store.ListBookAssets(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 0 {
		t.Fatalf("generated comic page leaked into assets: %+v", assets)
	}
	current, err := store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{
		"productionMode": comicTestRaw(t, "comic"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
			"enabled": true,
			"style":   map[string]any{"presetId": "batch-comic-style-15", "body": "新的厚涂漫画画风"},
		}}),
	}, ExpectedRevision: current.SettingsState.Revision}); err != nil {
		t.Fatal(err)
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages", nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list after config change status=%d body=%s", listed.Code, listed.Body.String())
	}
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, listed).Pages
	if len(pages) != 1 || !pages[0].ConfigStale || len(pages[0].Images) != 1 {
		t.Fatalf("stale page should preserve its candidate: %+v", pages)
	}
}

func TestComicPageAssetReferenceURLsUseOnlyMentionedPrimaryImages(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	hero, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "阿月", Prompt: "青衣女子"})
	if err != nil {
		t.Fatal(err)
	}
	unused, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Name: "后院", Prompt: "古宅后院"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, hero.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: "hero-first.png", MediaType: "image/png", Source: batchfactoryv11.VideoProviderDoubaoLocal})
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, hero.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: "hero-primary.png", MediaType: "image/png", Source: batchfactoryv11.VideoProviderDoubaoLocal})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetPrimaryBookAssetImage(ctx, "alice", batch.ID, book.ID, hero.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, unused.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://images.example/unused.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}

	got, err := comicPageAssetReferenceURLs(ctx, store, "alice", batch.ID, book.ID, "阿月在正堂抬头")
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/local-executor/v1/batch-book-asset-images/" + batch.ID + "/" + book.ID + "/" + hero.ID + "/" + second.ID
	if len(got) != 1 || got[0] != want || strings.Contains(got[0], first.ID) {
		t.Fatalf("references=%v want=%q", got, want)
	}
}
