package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

func TestRepairComicAssetPromptUsesRewriteSanitizesAndCaches(t *testing.T) {
	key := comicAssetRepairKey("batch-1", "book-1", "asset-1")
	shuihuoRepairedPrompts.Delete(key)
	defer shuihuoRepairedPrompts.Delete(key)
	calls := 0
	rewrite := func(_ context.Context, _ string, request novelfetchworkshop.RewriteRequest) (string, error) {
		calls++
		if !strings.Contains(request.UserPrompt, "三视图") || !strings.Contains(request.UserPrompt, "参考图") {
			t.Fatalf("asset constraints missing from rewrite request: %s", request.UserPrompt)
		}
		return strings.Repeat("保留角色身份与漫画画风，正面、侧面、背面三视图，参考图约束。", 4) + " 刀叉", nil
	}
	original := strings.Repeat("角色资产，画风统一，正面、侧面、背面三视图，参考图约束。", 4) + " 刀叉"
	repaired := repairComicImagePrompt(context.Background(), "alice", original, "内容审核拒绝", rewrite, true)
	if calls != 1 || repaired == original || strings.Contains(repaired, "持刀") {
		t.Fatalf("repaired=%q calls=%d", repaired, calls)
	}
	baseRevision := comicPromptRevision("资产原始提示词")
	storeShuihuoRepairedPrompt(key, shuihuoRepairedPrompt{Prompt: repaired, PromptRevision: baseRevision, PageRevision: 0})
	cached, ok := loadShuihuoRepairedPrompt(key, baseRevision)
	if !ok || cached.Prompt != repaired {
		t.Fatalf("cached repaired prompt=%+v ok=%t", cached, ok)
	}
	_ = repairComicImagePrompt
}

func TestComicBookGenerationContextPreservesParentDeadlineWhileIgnoringCancellation(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	child, childCancel := comicBookGenerationContext(parent)
	defer childCancel()

	cancel()
	select {
	case <-child.Done():
		t.Fatal("detached comic generation should ignore an early browser cancellation")
	case <-time.After(2 * time.Millisecond):
	}

	select {
	case <-child.Done():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("detached comic generation did not preserve the parent deadline")
	}
}

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
	// 列表行用 lite 视图：保留状态与图片地址，剥离提示词正文。
	liteListed := signedJSONRequest(t, api, now, "alice", http.MethodGet, pagesPath+"?view=lite", nil)
	if liteListed.Code != http.StatusOK {
		t.Fatalf("lite list status=%d body=%s", liteListed.Code, liteListed.Body.String())
	}
	litePages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, liteListed).Pages
	if len(litePages) != 1 || len(litePages[0].Images) != 2 {
		t.Fatalf("lite pages=%+v", litePages)
	}
	if litePages[0].ComicPrompt != "" || litePages[0].ImagePrompt != "" || len(litePages[0].ReferenceImageURLs) != 0 {
		t.Fatalf("lite view must strip prompt text: %+v", litePages[0])
	}
	if litePages[0].GenerationState == "" || litePages[0].PromptStatus == "" || litePages[0].ImageStatus == "" {
		t.Fatalf("lite view must keep page state: %+v", litePages[0])
	}
	assets, err := store.ListBookAssets(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 0 {
		t.Fatalf("comic candidates leaked into book assets: %+v", assets)
	}
}

func TestComicPageAssetLookupUsesStructuredVisibleBindings(t *testing.T) {
	lookup := comicPageAssetLookupText(batchfactoryv11.Video{
		ComicRoles: []string{"林晚", "陆沉"}, ComicScenePresets: []string{"办公室"},
	}, "完整中文漫画页")
	for _, expected := range []string{"完整中文漫画页", "林晚", "陆沉", "办公室"} {
		if !strings.Contains(lookup, expected) {
			t.Fatalf("asset lookup text missing %q: %q", expected, lookup)
		}
	}
}

func TestComicPageGenerationRejectsARecentDurableGenerationTask(t *testing.T) {
	now := time.Now().UTC()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	if _, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{
		ComicPrompt: "完整漫画页",
		ImagePrompt: "完整出图描述",
		Status:      "generating",
	}); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret:    "secret",
		Now:             func() time.Time { return now },
		Slice:           1,
		Store:           store,
		LocalArtifacts:  localartifact.NewStore(t.TempDir(), 1024),
		NovelFetchStore: novelfetchworkshop.NewMemoryStore(),
	})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + video.ID + "/generate"
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"comicPrompt": "完整漫画页",
		"imagePrompt": "完整出图描述",
	})
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "当前漫画页已有生成任务") {
		t.Fatalf("recent durable task must stay protected: status=%d body=%s", response.Code, response.Body.String())
	}
	pages, err := store.ListComicPages(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0].Status != "generating" {
		t.Fatalf("protected page=%+v, want generating", pages)
	}
}

func TestComicPageAssetReferencesIgnoreUnboundNamesInImagePrompt(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	visible, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "comic", Name: "林晚", Prompt: "黑发女性"})
	if err != nil {
		t.Fatal(err)
	}
	unbound, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "comic", Name: "我", Prompt: "第一人称角色"})
	if err != nil {
		t.Fatal(err)
	}
	visibleImage, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, visible.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://images.example/linwan.png", MediaType: "image/png", Source: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	unboundImage, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, unbound.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://images.example/first-person.png", MediaType: "image/png", Source: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetPrimaryBookAssetImage(ctx, "alice", batch.ID, book.ID, visible.ID, visibleImage.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetPrimaryBookAssetImage(ctx, "alice", batch.ID, book.ID, unbound.ID, unboundImage.ID); err != nil {
		t.Fatal(err)
	}
	refs, err := comicPageAssetReferenceURLsForPage(ctx, store, "alice", batch.ID, book.ID, batchfactoryv11.Video{
		ComicRoles: []string{"林晚"},
	}, "我走进房间，林晚回头看我")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0] != "https://images.example/linwan.png" {
		t.Fatalf("references=%v, want only the visibly bound character", refs)
	}
}

func TestComicPageReferenceRoutesSaveAndApplyCurrentReferenceToAllPages(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "一"}, {Label: "二"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	for index := range book.Videos {
		video := book.Videos[index]
		if _, err := store.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeVideo, BatchID: batch.ID, BookID: book.ID, VideoID: video.ID}, batchfactoryv11.SettingsUpdate{
			Patch: batchfactoryv11.SettingsPatch{"comicPrompt": comicTestRaw(t, "漫画页"), "comicImagePrompt": comicTestRaw(t, "可直接生成的漫画页")}, ExpectedRevision: video.Revision,
		}); err != nil {
			t.Fatal(err)
		}
		batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
		if err != nil {
			t.Fatal(err)
		}
		book = batch.Books[0]
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024)})
	pagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"
	response := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+book.Videos[0].ID, map[string]any{"comicPrompt": "漫画页", "imagePrompt": "可直接生成的漫画页"})
	if response.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", response.Code, response.Body.String())
	}
	reference := "/api/local-executor/v1/batch-comic-references/" + batch.ID + "/reference-a"
	saved := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+book.Videos[0].ID+"/reference-images", map[string]any{"referenceImageUrls": []string{reference}})
	if saved.Code != http.StatusOK {
		t.Fatalf("save status=%d body=%s", saved.Code, saved.Body.String())
	}
	applied := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+book.Videos[0].ID+"/reference-images/apply-to-all", map[string]any{})
	if applied.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", applied.Code, applied.Body.String())
	}
	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, pagesPath, nil)
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, listed).Pages
	if len(pages) != 2 || len(pages[1].ReferenceImageURLs) != 1 || pages[1].ReferenceImageURLs[0] != reference {
		t.Fatalf("pages=%+v", pages)
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

func TestComicCollectionRouteAcceptsSelectedPageRange(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文", Videos: []batchfactoryv11.CreateVideoInput{{Label: "一"}, {Label: "二"}, {Label: "三"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024)})
	pagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"
	collectionsPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-collections"
	dataURL := "data:image/png;base64,iVBORw0KGgoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	for _, video := range book.Videos {
		created := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+video.ID, map[string]any{"comicPrompt": "漫画页", "imagePrompt": "漫画页直接出图提示词"})
		if created.Code != http.StatusOK {
			t.Fatalf("page status=%d body=%s", created.Code, created.Body.String())
		}
		page := decodeBody[struct {
			Page batchfactoryv11.ComicPage `json:"page"`
		}](t, created).Page
		uploaded := signedJSONRequest(t, api, now, "alice", http.MethodPost, pagesPath+"/"+page.ID+"/images/upload", map[string]any{"dataUrl": dataURL})
		if uploaded.Code != http.StatusCreated {
			t.Fatalf("upload status=%d body=%s", uploaded.Code, uploaded.Body.String())
		}
	}
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, collectionsPath, map[string]any{"videoIds": []string{book.Videos[0].ID, book.Videos[1].ID}})
	if created.Code != http.StatusCreated {
		t.Fatalf("selected collection status=%d body=%s", created.Code, created.Body.String())
	}
	collection := decodeBody[struct {
		Collection batchfactoryv11.ComicCollection `json:"collection"`
	}](t, created).Collection
	if len(collection.Items) != 2 || collection.Items[0].VideoID != book.Videos[0].ID || collection.Items[1].VideoID != book.Videos[1].ID {
		t.Fatalf("selected collection=%+v", collection)
	}
}

func TestComicCollectionPreflightRejectsPageWithoutExtractedPrompts(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024)})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-collections"
	blocked := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{})
	if blocked.Code != http.StatusConflict || !strings.Contains(blocked.Body.String(), "分页和分镜提示词") {
		t.Fatalf("missing prompts should block collection: status=%d body=%s", blocked.Code, blocked.Body.String())
	}
}

func TestComicCollectionPreflightRejectsPageWithStaleImageSnapshot(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "原文内容", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	digest := sha256.Sum256([]byte(book.SourceText))
	if _, err := store.ReplaceComicSourceSegments(ctx, "alice", batch.ID, book.ID, fmt.Sprintf("%x", digest), []batchfactoryv11.ComicSourceSegmentDraft{{StartOffset: 0, EndOffset: len(book.SourceText), SourceText: book.SourceText, CharacterCount: len([]rune(book.SourceText)), BoundaryType: "test"}}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book = batch.Books[0]
	page, err := store.UpsertComicPage(ctx, "alice", batch.ID, book.ID, book.Videos[0].ID, batchfactoryv11.UpsertComicPageInput{
		ComicPrompt: "已经完成分页的内容", ImagePrompt: "已经完成的漫画图片提示词", SnapshotHash: "old-snapshot",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(ctx, "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{URL: "https://images.example/page.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024)})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-collections"
	blocked := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{})
	if blocked.Code != http.StatusCreated {
		t.Fatalf("style/model changes must not block collection: status=%d body=%s", blocked.Code, blocked.Body.String())
	}
}

func TestComicPageConfigStaleOnlyWhenSourceDigestChanges(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	page := batchfactoryv11.ComicPage{VideoID: "page-1", SnapshotHash: "old", Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}}}
	batch := batchfactoryv11.Batch{ID: "batch-1", Books: []batchfactoryv11.Book{{ID: "book-1", SourceText: "新正文", Videos: []batchfactoryv11.Video{{ID: "page-1", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceDigest: "old-source-digest"}}}}}}
	if !comicPageConfigStale(ctx, store, nil, "alice", batch, "book-1", page) {
		t.Fatal("changed source digest must mark the page stale")
	}
	digest := sha256.Sum256([]byte(batch.Books[0].SourceText))
	batch.Books[0].Videos[0].ComicSourceSegment.SourceDigest = fmt.Sprintf("%x", digest)
	if comicPageConfigStale(ctx, store, nil, "alice", batch, "book-1", page) {
		t.Fatal("current source digest must not be marked stale")
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
		"productionMode":   comicTestRaw(t, "comic"),
		"imageAspectRatio": comicTestRaw(t, "1:1"),
		"imageModelId":     comicTestRaw(t, "comic-image"),
		"imageProvider":    comicTestRaw(t, "personal_api"),
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
		"comicPrompt": "第一页分格导演说明：上格人物特写，中格手机界面，下格浴室全景",
		"imagePrompt": "三格竖版条漫，人物一致，第一页完整画面",
		"aspectRatio": "16:9",
		// Simulate a stale page/editor payload. The saved book settings above
		// must still win and route this request to the personal API model.
		"provider": "doubao_local_executor",
		"modelId":  "Seedream 4.5",
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
	if !strings.Contains(providerRequest.Prompt, "第一页分格导演说明：上格人物特写，中格手机界面，下格浴室全景") {
		t.Fatalf("provider prompt lost director content: %s", providerRequest.Prompt)
	}
	if providerRequest.Model != "comic-provider-v1" || providerRequest.Size != "1024x1024" {
		t.Fatalf("provider request=%+v", providerRequest)
	}
	if result.Page.AspectRatio != "1:1" {
		t.Fatalf("comic page ratio=%q, want unified image ratio 1:1", result.Page.AspectRatio)
	}
	if result.Page.ModelID != "comic-image" {
		t.Fatalf("comic page model=%q, want unified image model", result.Page.ModelID)
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
	if len(pages) != 1 || pages[0].ConfigStale || len(pages[0].Images) != 1 {
		t.Fatalf("configuration changes must preserve the current page without stale warning: %+v", pages)
	}
}

func TestComicPagePromptPreviewReturnsTheExactCompletePromptWithoutGenerating(t *testing.T) {
	now := time.Unix(1700000000, 0)
	presetPath := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(presetPath, []byte(`[
		{"id":"batch-comic-style-16","module":"batch-factory","name":"精致火柴2DQ版漫风","kind":"base","version":3,"revision":1,"status":"published","body":"精致头部，极简火柴四肢","previewAssetId":"test-style-asset","protocolLock":{"format":"style","slot":"batch.comic-style"}}
	]`), 0o600); err != nil {
		t.Fatal(err)
	}
	presetCatalog := prompt.NewCatalog(presetPath)
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
			"enabled": true,
			// Older saved batches may retain the selected style ID without its
			// private body. Prompt preview must hydrate it from the published catalog.
			"style":             map[string]any{"presetId": "batch-comic-style-16", "presetName": "精致火柴2DQ版漫风"},
			"imagePromptPrefix": "统一前缀",
			"imagePromptSuffix": "统一后缀",
			"globalReferenceImage": map[string]any{
				"assetId":    "scene:batch-comic-reference-scene-urban",
				"url":        "/api/local-executor/v1/batch-comic-reference-scenes/batch-comic-reference-scene-urban",
				"previewUrl": "/api/presets/comic-reference-scene/batch-comic-reference-scene-urban/preview",
				"source":     "builtin", "enabled": true, "applyToPages": true,
				"scene": map[string]any{"presetId": "batch-comic-reference-scene-urban", "presetName": "都市", "presetSlot": "batch.comic-reference-scene", "presetVersion": 1},
			},
		}}),
	}, ExpectedRevision: batch.Revision})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(context.Background(), "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book, video = batch.Books[0], batch.Books[0].Videos[0]
	accounts := &governanceAuthStore{accounts: map[string]auth.Account{"alice": {Username: "alice", Active: true, IsOwner: true}}}
	profiles := NewMemberProfileStore(t.TempDir())
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024), PresetCatalog: presetCatalog, AuthStore: accounts, MemberProfileStore: profiles})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + video.ID + "/prompt-preview"
	preview := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"comicPrompt":        "分格导演说明：上格特写，中格插入格，下格大主格",
		"imagePrompt":        "三格竖版漫画，都市浴室",
		"referenceImageUrls": []string{"/api/local-executor/v1/batch-comic-references/" + batch.ID + "/manual-reference-a"},
	})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	result := decodeBody[struct {
		ComicPrompt     string   `json:"comicPrompt"`
		ImagePrompt     string   `json:"imagePrompt"`
		CompiledPrompt  string   `json:"compiledPrompt"`
		ReferenceURLs   []string `json:"referenceImageUrls"`
		ReferenceImages []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			URL        string `json:"url"`
			PreviewURL string `json:"previewUrl"`
			Source     string `json:"source"`
			Removable  bool   `json:"removable"`
		} `json:"referenceImages"`
	}](t, preview)
	if result.ComicPrompt != "分格导演说明：上格特写，中格插入格，下格大主格" || result.ImagePrompt != "三格竖版漫画，都市浴室" {
		t.Fatalf("raw prompts changed: %+v", result)
	}
	for _, part := range []string{"画风名称：精致火柴2DQ版漫风", "精致头部，极简火柴四肢", "第1张参考图是画风库参考图", "统一前缀", "阅读顺序", "三格竖版漫画，都市浴室", "统一后缀"} {
		if !strings.Contains(result.CompiledPrompt, part) {
			t.Fatalf("compiled prompt misses %q: %s", part, result.CompiledPrompt)
		}
	}
	if !strings.Contains(result.CompiledPrompt, "分格导演说明：上格特写，中格插入格，下格大主格") {
		t.Fatalf("preview prompt lost director content: %s", result.CompiledPrompt)
	}
	if len(result.ReferenceURLs) != 2 || len(result.ReferenceImages) != 2 || result.ReferenceURLs[0] != "/api/local-executor/v1/batch-comic-materials/batch-comic-style-16" {
		t.Fatalf("references=%v images=%+v", result.ReferenceURLs, result.ReferenceImages)
	}
	style := result.ReferenceImages[0]
	if style.ID != "batch-comic-style-16" || style.PreviewURL != "/api/presets/comic-material/batch-comic-style-16/preview" || style.Removable {
		t.Fatalf("style reference metadata=%+v", style)
	}
	manual := result.ReferenceImages[1]
	if manual.Name != "手动参考图 1" || manual.Source != "manual-upload" || !manual.Removable || !strings.Contains(manual.PreviewURL, "/comic-references/manual-reference-a/content") {
		t.Fatalf("manual reference metadata=%+v", manual)
	}
	pages := decodeBody[struct {
		Pages []batchfactoryv11.ComicPage `json:"pages"`
	}](t, signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/"+batch.ID+"/books/"+book.ID+"/comic-pages", nil)).Pages
	if len(pages) != 0 {
		t.Fatalf("read-only preview persisted pages: %+v", pages)
	}
}

func TestComicPagePromptPreviewRejectsNonDeveloperAccounts(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "member", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.SaveSettings(context.Background(), "member", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{
		Patch: batchfactoryv11.SettingsPatch{
			"productionMode": comicTestRaw(t, "comic"),
			"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
				"enabled": true,
				"style":   map[string]any{"presetId": "batch-comic-style-14", "body": "水墨漫画画风"},
			}}),
		}, ExpectedRevision: batch.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(context.Background(), "member", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	accounts := &governanceAuthStore{accounts: map[string]auth.Account{"member": {Username: "member", Active: true}}}
	profiles := NewMemberProfileStore(t.TempDir())
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, LocalArtifacts: localartifact.NewStore(t.TempDir(), 1024), AuthStore: accounts, MemberProfileStore: profiles})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + video.ID + "/prompt-preview"

	response := signedJSONRequest(t, api, now, "member", http.MethodPost, path, map[string]any{
		"comicPrompt": "分格导演描述",
		"imagePrompt": "本页基础出图描述",
	})

	if response.Code != http.StatusForbidden {
		t.Fatalf("non-developer preview status=%d body=%s", response.Code, response.Body.String())
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

func TestComicPageHasPrimarySkipsOnlyPagesWithASelectedMainImage(t *testing.T) {
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "ready", Images: []batchfactoryv11.ComicPageImage{{ID: "candidate"}, {ID: "main", IsPrimary: true}}},
		{VideoID: "candidate-only", Images: []batchfactoryv11.ComicPageImage{{ID: "candidate"}}},
	}
	if !comicPageHasPrimary(pages, "ready") {
		t.Fatal("page with a primary image must be reusable")
	}
	if comicPageHasPrimary(pages, "candidate-only") || comicPageHasPrimary(pages, "missing") {
		t.Fatal("page without a primary image must still be generated")
	}
}

func TestMissingComicAssetImageIDsReturnsOnlyAssetsWithoutPrimaryImages(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	ready, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "comic", Name: "已完成", Prompt: "人物"})
	if err != nil {
		t.Fatal(err)
	}
	missing, err := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Source: "comic", Name: "待生成", Prompt: "场景"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, ready.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://example.com/ready.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	current, _ := store.GetBatch(ctx, "alice", batch.ID)

	ids, err := missingComicAssetImageIDs(ctx, store, "alice", current.ID, current.Books[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != missing.ID {
		t.Fatalf("missing ids=%v, want only %s", ids, missing.ID)
	}
}

func TestComicPageAssetReferencesUseCharacterImagesOnly(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "漫画批次", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	character, _ := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Source: "comic", Name: "月锦", Prompt: "人物"})
	scene, _ := store.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "scene", Source: "comic", Name: "正房明厅", Prompt: "场景"})
	for _, asset := range []batchfactoryv11.BookAsset{character, scene} {
		if _, err := store.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://example.com/" + asset.ID + ".png", MediaType: "image/png", Source: "provider"}); err != nil {
			t.Fatal(err)
		}
	}

	references, err := comicPageAssetReferenceURLs(ctx, store, "alice", batch.ID, book.ID, "月锦站在正房明厅")
	if err != nil {
		t.Fatal(err)
	}
	if len(references) != 1 || !strings.Contains(references[0], character.ID) {
		t.Fatalf("references=%v, want character image only", references)
	}
}

func TestComicPageConfigStaleUsesSavedPageReferences(t *testing.T) {
	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	settings := batchfactoryv11.SettingsPatch{
		"productionMode": comicTestRaw(t, "comic"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
			"enabled": true,
			"style":   map[string]any{"presetId": "batch-comic-style-14", "body": "水墨漫画画风"},
		}}),
	}
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: settings, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	batch, err = store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	pageReference := "/api/local-executor/v1/batch-comic-references/" + batch.ID + "/manual-reference"
	effective := batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
	contract, err := batchfactoryv11.BuildComicImageContract(effective, batchfactoryv11.ComicImagePurposePage, "", "完整漫画页", nil, []string{pageReference})
	if err != nil {
		t.Fatal(err)
	}
	page, err := store.UpsertComicPage(ctx, "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{
		ComicPrompt: "完整漫画页", ImagePrompt: contract.Prompt, ReferenceImageURLs: []string{pageReference}, SnapshotHash: contract.SnapshotHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(ctx, "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{URL: "https://example.com/page.png", MediaType: "image/png", Source: "provider"}); err != nil {
		t.Fatal(err)
	}
	current, err := store.GetBatch(ctx, "alice", batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := store.ListComicPages(ctx, "alice", batch.ID, book.ID)
	if err != nil || len(pages) != 1 {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	if comicPageConfigStale(ctx, store, nil, "alice", current, book.ID, pages[0]) {
		t.Fatalf("page with unchanged saved reference was marked stale: %+v", pages[0])
	}
}

// TestComicPagePutIsPatchNotFullRow is the contract for editing prompts: PUT
// only replaces fields the client actually sent. Server-owned snapshotHash,
// status, reference images and page images must survive.
func TestComicPagePutIsPatchNotFullRow(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{
			Title:  "第一本",
			Videos: []batchfactoryv11.CreateVideoInput{{Label: "分镜一", VideoPrompt: "视频一"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	video := book.Videos[0]
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	pagesPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"

	const snapshotHash = "contract-snapshot-hash"
	created, err := store.UpsertComicPage(ctx, "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{
		ComicPrompt:        "导演页描述",
		ImagePrompt:        "出图描述",
		ReferenceImageURLs: []string{"https://example.com/ref.png"},
		ModelID:            "seedream-5-flash-local",
		AspectRatio:        "3:4",
		Status:             "image_ready",
		SnapshotHash:       snapshotHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(ctx, "alice", batch.ID, book.ID, created.ID, batchfactoryv11.CreateComicPageImageInput{
		URL: "https://example.com/page.png", MediaType: "image/png", Source: "provider",
	}); err != nil {
		t.Fatal(err)
	}

	response := signedJSONRequest(t, api, now, "alice", http.MethodPut, pagesPath+"/"+video.ID, map[string]any{
		"comicPrompt": "修改后的导演页描述",
		"imagePrompt": "修改后的出图描述",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", response.Code, response.Body.String())
	}
	page := decodeBody[struct {
		Page batchfactoryv11.ComicPage `json:"page"`
	}](t, response).Page
	if page.ComicPrompt != "修改后的导演页描述" || page.ImagePrompt != "修改后的出图描述" {
		t.Fatalf("prompts not patched: %+v", page)
	}
	// SnapshotHash is server-owned (json:"-"), so verify it on the stored row.
	pages, err := store.ListComicPages(ctx, "alice", batch.ID, book.ID)
	if err != nil || len(pages) != 1 {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	stored := pages[0]
	if stored.SnapshotHash != snapshotHash {
		t.Fatalf("snapshotHash cleared by PUT: %q", stored.SnapshotHash)
	}
	if stored.Status != "image_ready" || stored.ModelID != "seedream-5-flash-local" || stored.AspectRatio != "3:4" {
		t.Fatalf("status/model/aspect replaced by PUT: %+v", stored)
	}
	if len(stored.ReferenceImageURLs) != 1 || stored.ReferenceImageURLs[0] != "https://example.com/ref.png" {
		t.Fatalf("referenceImageUrls replaced by PUT: %+v", stored.ReferenceImageURLs)
	}
	if len(stored.Images) != 1 {
		t.Fatalf("page images lost after PUT: %+v", stored.Images)
	}

	// Creating a page through the same PUT still applies defaults.
	batch, err = store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "第二批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第二本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "分镜一"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	newBook := batch.Books[0]
	newVideo := newBook.Videos[0]
	newPath := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + newBook.ID + "/comic-pages/" + newVideo.ID
	newPageResponse := signedJSONRequest(t, api, now, "alice", http.MethodPut, newPath, map[string]any{
		"comicPrompt": "新页导演描述",
		"imagePrompt": "新页出图描述",
	})
	if newPageResponse.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", newPageResponse.Code, newPageResponse.Body.String())
	}
	newPage := decodeBody[struct {
		Page batchfactoryv11.ComicPage `json:"page"`
	}](t, newPageResponse).Page
	if newPage.Status != "prompt_ready" || newPage.AspectRatio != "9:16" {
		t.Fatalf("new page defaults=%+v", newPage)
	}
}

func TestClassifyComicPageShortageSeparatesInFlightStaleAndFailed(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	videos := []batchfactoryv11.Video{{ID: "v1"}, {ID: "v2"}, {ID: "v3"}, {ID: "v4"}}
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "v1", Status: "generating", UpdatedAt: now.Add(-2 * time.Minute)},
		{VideoID: "v2", Status: "generating", UpdatedAt: now.Add(-40 * time.Minute)},
		{VideoID: "v3", Status: "generation_failed", UpdatedAt: now},
	}
	shortage := classifyComicPageShortage(pages, videos, []string{"v1", "v2", "v3", "v4"}, now)
	if shortage.Total != 4 || shortage.InFlight != 1 || shortage.Stale != 1 || shortage.Failed != 2 {
		t.Fatalf("shortage=%+v", shortage)
	}
	if !errors.Is(shortage, batchfactoryv11.ErrConflict) {
		t.Fatalf("shortage should unwrap ErrConflict, got %v", shortage)
	}
	if message := shortage.Error(); !strings.Contains(message, "1 页仍在生成") || !strings.Contains(message, "1 页生成租约已过期") || !strings.Contains(message, "2 页生成失败") {
		t.Fatalf("shortage message=%q", message)
	}
}

func comicPageLiteralTestSettings(t *testing.T, presetID string) batchfactoryv11.SettingsPatch {
	t.Helper()
	return batchfactoryv11.SettingsPatch{
		"productionMode": comicTestRaw(t, "comic"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
			"presetId": presetID,
			"style":    map[string]any{"presetId": "batch-comic-style-08", "body": "韩漫条漫画风"},
		}}),
	}
}

func comicPageLiteralTestPage(source string) batchfactoryv11.Video {
	return batchfactoryv11.Video{ID: "video-1", ComicSourceSegment: &batchfactoryv11.ComicSourceSegment{SourceText: source}}
}

func TestComicPageFinalImagePromptInjectsLiteralSourceOnlyForIllustratedText(t *testing.T) {
	source := "夜色压低了整条街。\n「你迟到了。」陆淮安把外套递过来，声音很轻。"
	page := comicPageLiteralTestPage(source)
	for _, presetID := range []string{
		"batch-comic-page",
		"batch-comic-page-fusion-bw",
		"batch-comic-page-light",
		"batch-comic-page-premium-strip",
	} {
		settings := comicPageLiteralTestSettings(t, presetID)
		contract, err := batchfactoryv11.BuildComicImageContract(settings, batchfactoryv11.ComicImagePurposePage, "", "画面上方文字区，下方插画区。", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		got, appended := comicPageFinalImagePrompt(contract.Prompt, settings, page)
		if appended || got != contract.Prompt {
			t.Fatalf("%s 不得注入原文块", presetID)
		}
		if strings.Contains(got, source) || strings.Contains(got, "【文字区必须逐字排版") {
			t.Fatalf("%s 的出图 prompt 不得包含原文块:\n%s", presetID, got)
		}
	}
	settings := comicPageLiteralTestSettings(t, batchfactoryv11.ComicPageIllustratedTextPresetID)
	contract, err := batchfactoryv11.BuildComicImageContract(settings, batchfactoryv11.ComicImagePurposePage, "", "画面上方文字区，下方插画区。", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, appended := comicPageFinalImagePrompt(contract.Prompt, settings, page)
	if !appended {
		t.Fatal("CM·文字插画版必须注入原文块")
	}
	if !strings.Contains(got, source) {
		t.Fatalf("最终出图 prompt 必须逐字包含 current_page 原文:\n%s", got)
	}
	if strings.Count(got, source) != 1 {
		t.Fatalf("原文只应出现一次:\n%s", got)
	}
	again, appendedAgain := comicPageFinalImagePrompt(got, settings, page)
	if appendedAgain || again != got {
		t.Fatal("重复注入必须被去重")
	}
}

func TestComicPageFinalImagePromptSkipsEmptySourceWithoutError(t *testing.T) {
	settings := comicPageLiteralTestSettings(t, batchfactoryv11.ComicPageIllustratedTextPresetID)
	contract, err := batchfactoryv11.BuildComicImageContract(settings, batchfactoryv11.ComicImagePurposePage, "", "画面上方文字区，下方插画区。", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, page := range []batchfactoryv11.Video{
		{ID: "video-1"},
		comicPageLiteralTestPage(" \n "),
	} {
		got, appended := comicPageFinalImagePrompt(contract.Prompt, settings, page)
		if appended || got != contract.Prompt {
			t.Fatal("空 current_page 时不得追加约束块")
		}
	}
}

// TestComicPageManualPromptEditInvalidatesRepairedPromptAndGenerationUsesNewPrompt
// 复现 21 版必修 bug：人工改提示词后，内存里的旧修复版必须立即失效，
// 生成必须使用数据库里的新提示词。
func TestComicPageManualPromptEditInvalidatesRepairedPromptAndGenerationUsesNewPrompt(t *testing.T) {
	now := time.Unix(1700000000, 0)
	var providerPrompt string
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode provider request: %v", err)
			return
		}
		providerPrompt = request.Prompt
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII="}]}`))
	}))
	defer provider.Close()

	ctx := context.Background()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第九页"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book, video := batch.Books[0], batch.Books[0].Videos[0]
	if _, err := store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: batch.ID}, batchfactoryv11.SettingsUpdate{Patch: batchfactoryv11.SettingsPatch{
		"productionMode":   comicTestRaw(t, "comic"),
		"imageAspectRatio": comicTestRaw(t, "1:1"),
		"imageModelId":     comicTestRaw(t, "comic-image"),
		"imageProvider":    comicTestRaw(t, "personal_api"),
		"aiPromptConfig": comicTestRaw(t, map[string]any{"comic": map[string]any{
			"enabled": true,
			"style":   map[string]any{"presetId": "batch-comic-style-14", "body": "水墨漫画画风"},
		}}),
	}, ExpectedRevision: batch.Revision}); err != nil {
		t.Fatal(err)
	}
	catalog := novelfetchworkshop.NewMemoryStore()
	if err := catalog.PutConfig(ctx, "alice", map[string]any{"modelCatalog": []any{
		map[string]any{"id": "comic-image", "kind": "image", "enabled": true, "provider": "personal_api", "modelId": "comic-provider-v1", "baseUrl": provider.URL, "credential": "test-secret"},
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.UpsertComicPage(ctx, "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{
		ComicPrompt: "旧分镜：复视室对峙",
		ImagePrompt: "旧画面：复视室里两人对坐，气氛紧绷",
	}); err != nil {
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
	key := comicPageRepairKey(batch.ID, book.ID, video.ID)
	storeShuihuoRepairedPrompt(key, shuihuoRepairedPrompt{Prompt: "旧修复版：复视室对峙，旧气泡「最后没钱治病，抱憾绝望离世」", PromptRevision: comicPromptRevision("旧画面"), PageRevision: 1})
	defer shuihuoRepairedPrompts.Delete(key)

	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages/" + video.ID
	edited := signedJSONRequest(t, api, now, "alice", http.MethodPut, path, map[string]any{
		"comicPrompt": "第九页：空走廊，散落假金条与药袋，远处模糊背影",
		"imagePrompt": "空走廊，地面散落假金条与药袋，走廊尽头一个模糊背影，冷色调",
	})
	if edited.Code != http.StatusOK {
		t.Fatalf("manual edit status=%d body=%s", edited.Code, edited.Body.String())
	}
	if _, ok := shuihuoRepairedPrompts.Load(key); ok {
		t.Fatal("人工修改提示词后修复版缓存必须立即失效")
	}

	generated := signedJSONRequest(t, api, now, "alice", http.MethodPost, path+"/generate", map[string]any{})
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	if !strings.Contains(providerPrompt, "空走廊") || !strings.Contains(providerPrompt, "假金条") {
		t.Fatalf("生成必须使用数据库里的新提示词: %s", providerPrompt)
	}
	if strings.Contains(providerPrompt, "复视室对峙") {
		t.Fatalf("生成不得再复用旧修复版提示词: %s", providerPrompt)
	}
}

// TestRepairedPromptDroppedWhenDatabasePromptRevisionChanged 证明缓存版本
// 落后于数据库时会被丢弃，而不是继续压住数据库提示词。
func TestRepairedPromptDroppedWhenDatabasePromptRevisionChanged(t *testing.T) {
	key := comicPageRepairKey("batch-rev", "book-rev", "video-rev")
	defer shuihuoRepairedPrompts.Delete(key)
	storeShuihuoRepairedPrompt(key, shuihuoRepairedPrompt{Prompt: "旧修复版提示词正文", PromptRevision: comicPromptRevision("旧提示词"), PageRevision: 7})

	if _, ok := loadShuihuoRepairedPrompt(key, comicPromptRevision("旧提示词")); !ok {
		t.Fatal("版本一致时应命中缓存")
	}
	if _, ok := loadShuihuoRepairedPrompt(key, comicPromptRevision("人工改写后的提示词")); ok {
		t.Fatal("数据库提示词版本变化后必须丢弃缓存")
	}
	if _, ok := shuihuoRepairedPrompts.Load(key); ok {
		t.Fatal("版本不符时应删除缓存条目")
	}
}

// TestRepairedPromptExpiresAfterInjectedTTL 用注入时钟验证 TTL 兜底，
// 不 sleep。
func TestRepairedPromptExpiresAfterInjectedTTL(t *testing.T) {
	originalNow := shuihuoRepairedPromptNow
	defer func() { shuihuoRepairedPromptNow = originalNow }()
	current := time.Date(2026, 11, 6, 3, 42, 0, 0, time.UTC)
	shuihuoRepairedPromptNow = func() time.Time { return current }

	key := comicPageRepairKey("batch-ttl", "book-ttl", "video-ttl")
	defer shuihuoRepairedPrompts.Delete(key)
	revision := comicPromptRevision("同一份数据库提示词")
	storeShuihuoRepairedPrompt(key, shuihuoRepairedPrompt{Prompt: "修复版提示词正文", PromptRevision: revision, PageRevision: 3})

	current = current.Add(shuihuoRepairedPromptTTL - time.Second)
	if _, ok := loadShuihuoRepairedPrompt(key, revision); !ok {
		t.Fatal("TTL 内应继续复用修复版")
	}
	current = current.Add(2 * time.Second)
	if _, ok := loadShuihuoRepairedPrompt(key, revision); ok {
		t.Fatal("TTL 到期后必须丢弃缓存并重新改写")
	}
}

// TestManualPromptOverrideNeverFallsBackToRepairedPrompt 证明人工覆盖后
// 该页不再回落到自动改写版本，且失效入口能按页/按书精确清理。
func TestManualPromptOverrideNeverFallsBackToRepairedPrompt(t *testing.T) {
	const batchID, bookID = "batch-scope", "book-scope"
	pageA, pageB := "video-a", "video-b"
	keyA := comicPageRepairKey(batchID, bookID, pageA)
	keyB := comicPageRepairKey(batchID, bookID, pageB)
	assetKey := comicAssetRepairKey(batchID, bookID, "asset-a")
	for _, key := range []string{keyA, keyB, assetKey} {
		storeShuihuoRepairedPrompt(key, shuihuoRepairedPrompt{Prompt: "自动修复版正文", PromptRevision: comicPromptRevision("旧提示词"), PageRevision: 1})
	}
	defer func() {
		for _, key := range []string{keyA, keyB, assetKey} {
			shuihuoRepairedPrompts.Delete(key)
		}
	}()

	invalidateShuihuoRepairedPrompt(batchID, bookID, pageA)
	if _, ok := loadShuihuoRepairedPrompt(keyA, comicPromptRevision("人工覆盖后的提示词")); ok {
		t.Fatal("人工覆盖后该页不得再回落自动改写版本")
	}
	if _, ok := shuihuoRepairedPrompts.Load(keyA); ok {
		t.Fatal("人工覆盖后该页缓存必须删除")
	}
	if _, ok := shuihuoRepairedPrompts.Load(keyB); !ok {
		t.Fatal("页级失效不得误删兄弟页缓存")
	}
	invalidateShuihuoRepairedPrompt(batchID, bookID, "")
	if _, ok := shuihuoRepairedPrompts.Load(keyB); ok {
		t.Fatal("书级失效必须清理本书所有页缓存")
	}
	if _, ok := shuihuoRepairedPrompts.Load(assetKey); ok {
		t.Fatal("书级失效必须清理本书资产修复缓存")
	}
}

// TestComicPromptRepairCoversVisualFieldsAndRecordsLedger 证明自动修复能
// 改写画面描述，并把修复次数、字段与前后摘要写进页级账本。
func TestComicPromptRepairCoversVisualFieldsAndRecordsLedger(t *testing.T) {
	before := "复视室里两人对峙，「最后没钱治病，抱憾绝望离世」，气氛紧绷"
	after := "空走廊，散落假金条与药袋，远处一个模糊背影，「最后没钱治病，抱憾绝望离世」"
	fields := comicPromptRepairFields(before, after)
	if len(fields) != 1 || fields[0] != "画面描述" {
		t.Fatalf("只改画面时必须标记画面描述: %v", fields)
	}
	both := comicPromptRepairFields(before, "空走廊，远处模糊背影，「这一次我先走」")
	if len(both) != 2 || both[0] != "台词" || both[1] != "画面描述" {
		t.Fatalf("台词与画面同时改写时都必须标记: %v", both)
	}

	entry := map[string]any{}
	now := time.Date(2026, 11, 6, 3, 42, 0, 0, time.UTC)
	recordComicPromptRepair(entry, comicPromptRepair{
		VideoID: "video-09", Label: "第 9 页", Fields: fields, Before: before, After: after, At: now,
		Attempts: 1, LastRefusalRaw: "上游原文：content policy violation",
	}, now)
	recordComicPromptRepair(entry, comicPromptRepair{
		VideoID: "video-09", Label: "第 9 页", Fields: fields, Before: after, After: after + "，冷色调", At: now.Add(time.Minute),
		Attempts: 2, LastRefusalRaw: "上游原文：content policy violation (round 2)",
	}, now.Add(time.Minute))
	ledger, _ := entry["nodes"].(map[string]any)
	comicNode, _ := ledger["comic_images"].(map[string]any)
	pageLedger, _ := comicNode["pageLedger"].(map[string]any)
	record, _ := pageLedger["video-09"].(map[string]any)
	if record == nil {
		t.Fatalf("页级账本缺失: %+v", entry)
	}
	if repaired, _ := record["autoRepaired"].(bool); !repaired {
		t.Fatalf("必须标记该页被自动改写过: %+v", record)
	}
	if count, _ := record["repairCount"].(int); count != 2 {
		t.Fatalf("修复次数=%v, want 2", record["repairCount"])
	}
	if attempts, _ := record["repairAttempts"].(int); attempts != 2 {
		t.Fatalf("自动改写轮次=%v, want 2", record["repairAttempts"])
	}
	if refusal, _ := record["lastRefusalRaw"].(string); !strings.Contains(refusal, "round 2") {
		t.Fatalf("最后拒绝原文=%q", refusal)
	}
	if repairedFields, _ := record["repairedFields"].([]string); len(repairedFields) != 1 || repairedFields[0] != "画面描述" {
		t.Fatalf("修复字段=%v", record["repairedFields"])
	}
	for _, key := range []string{"repairBefore", "repairAfter", "repairedAt"} {
		if text, _ := record[key].(string); strings.TrimSpace(text) == "" {
			t.Fatalf("账本缺少 %s: %+v", key, record)
		}
	}
	// 整本成功后的清零只重置当前账期，不能抹掉“被自动改写过”的历史。
	if !resetAutomationPageLedgerNode(entry, "comic_images", now.Add(2*time.Minute)) {
		t.Fatal("账本应有可重置的当前账期")
	}
	if repaired, _ := record["autoRepaired"].(bool); !repaired {
		t.Fatalf("清零后仍须保留自动改写历史: %+v", record)
	}
	if count, _ := record["repairCount"].(int); count != 2 {
		t.Fatalf("清零后修复次数不得丢失: %v", record["repairCount"])
	}
}
