package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
)

func TestDeriveComicPageStatusUsesPrimaryImageAsReady(t *testing.T) {
	cases := []struct {
		name             string
		page             batchfactoryv11.ComicPage
		wantPrompt       string
		wantImage        string
		wantGeneration   string
		wantFallbackOnly bool
	}{
		{
			name:           "missing prompts",
			page:           batchfactoryv11.ComicPage{Status: "prompt_ready"},
			wantPrompt:     comicPromptStatusMissing,
			wantImage:      comicImageStatusNone,
			wantGeneration: comicGenerationStateIdle,
		},
		{
			name:           "comic prompt only",
			page:           batchfactoryv11.ComicPage{ComicPrompt: "漫画页", Status: "prompt_ready"},
			wantPrompt:     comicPromptStatusComicOnly,
			wantImage:      comicImageStatusNone,
			wantGeneration: comicGenerationStateIdle,
		},
		{
			name:             "candidate image is not ready",
			page:             batchfactoryv11.ComicPage{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "succeeded", Images: []batchfactoryv11.ComicPageImage{{ID: "img-1"}}},
			wantPrompt:       comicPromptStatusReady,
			wantImage:        comicImageStatusCandidates,
			wantGeneration:   comicGenerationStateSucceeded,
			wantFallbackOnly: true,
		},
		{
			name:           "primary image is ready",
			page:           batchfactoryv11.ComicPage{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "succeeded", Images: []batchfactoryv11.ComicPageImage{{ID: "img-1", IsPrimary: true}}},
			wantPrompt:     comicPromptStatusReady,
			wantImage:      comicImageStatusPrimary,
			wantGeneration: comicGenerationStateSucceeded,
		},
		{
			name:           "failed but has image keeps failed state",
			page:           batchfactoryv11.ComicPage{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "generation_failed", Images: []batchfactoryv11.ComicPageImage{{ID: "img-1", IsPrimary: true}}},
			wantPrompt:     comicPromptStatusReady,
			wantImage:      comicImageStatusPrimary,
			wantGeneration: comicGenerationStateFailed,
		},
		{
			name:           "generating",
			page:           batchfactoryv11.ComicPage{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "generating"},
			wantPrompt:     comicPromptStatusReady,
			wantImage:      comicImageStatusNone,
			wantGeneration: comicGenerationStateGenerating,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hasPrimary, hasAny := comicPageImageFlags(tc.page)
			promptStatus, imageStatus, generationState, primaryFallback := deriveComicPageStatus(tc.page, hasPrimary, hasAny)
			if promptStatus != tc.wantPrompt || imageStatus != tc.wantImage || generationState != tc.wantGeneration {
				t.Fatalf("got prompt=%q image=%q generation=%q, want %q/%q/%q", promptStatus, imageStatus, generationState, tc.wantPrompt, tc.wantImage, tc.wantGeneration)
			}
			if primaryFallback != tc.wantFallbackOnly {
				t.Fatalf("primaryFallback=%v want=%v", primaryFallback, tc.wantFallbackOnly)
			}
		})
	}
}

func TestBuildComicReadinessCountsPrimaryOnlyAndFlagsFallback(t *testing.T) {
	videos := []batchfactoryv11.Video{{ID: "v1"}, {ID: "v2"}, {ID: "v3"}}
	pages := []batchfactoryv11.ComicPage{
		{VideoID: "v1", ComicPrompt: "漫画页1", ImagePrompt: "出图1", Status: "succeeded", Images: []batchfactoryv11.ComicPageImage{{ID: "i1", IsPrimary: true}}},
		{VideoID: "v2", ComicPrompt: "漫画页2", ImagePrompt: "出图2", Status: "succeeded", Images: []batchfactoryv11.ComicPageImage{{ID: "i2"}}},
	}
	readiness := buildComicReadiness(videos, pages, 35)
	if readiness.PagesTotal != 3 || readiness.PromptReadyCount != 2 || readiness.ImageReadyCount != 1 || readiness.PrimaryReadyCount != 1 {
		t.Fatalf("readiness=%+v", readiness)
	}
	if readiness.CanGenerateImages || readiness.CanBuildCollection || readiness.Skippable {
		t.Fatalf("must not be ready with candidate-only and missing row: %+v", readiness)
	}
	if readiness.Limits.MaxPages != 35 {
		t.Fatalf("maxPages=%d", readiness.Limits.MaxPages)
	}
	joined := strings.Join(readiness.Blockers, "|")
	for _, expected := range []string{"缺少漫画页记录：1 页", "尚未出主图：2 页", "历史图片未设主图：1 页"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("blockers=%q missing %q", joined, expected)
		}
	}
}

func TestBuildComicReadinessIsReadyWhenEveryPageHasPrimary(t *testing.T) {
	videos := []batchfactoryv11.Video{{ID: "v1"}}
	pages := []batchfactoryv11.ComicPage{{
		VideoID: "v1", ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "succeeded",
		Images: []batchfactoryv11.ComicPageImage{{ID: "i1", IsPrimary: true}},
	}}
	readiness := buildComicReadiness(videos, pages, 35)
	if !readiness.CanGenerateImages || !readiness.CanBuildCollection || !readiness.Skippable {
		t.Fatalf("readiness=%+v", readiness)
	}
	if len(readiness.Blockers) != 0 {
		t.Fatalf("blockers=%v", readiness.Blockers)
	}
}

// TestComicPagesReadResponseCarriesAuthoritativeStatus locks B's HTTP contract:
// page status fields and book readiness come from the backend, and reading the
// book does not mutate page images (candidate-only rows stay candidate-only).
func TestComicPagesReadResponseCarriesAuthoritativeStatus(t *testing.T) {
	now := time.Unix(1700000000, 0)
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
		Status:      "succeeded",
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
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-pages"
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Pages     []batchfactoryv11.ComicPage `json:"pages"`
		Readiness comicReadiness              `json:"readiness"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v body=%s", err, response.Body.String())
	}
	if len(payload.Pages) != 1 {
		t.Fatalf("pages=%d", len(payload.Pages))
	}
	page := payload.Pages[0]
	if page.PromptStatus != comicPromptStatusReady || page.ImageStatus != comicImageStatusNone || page.GenerationState != comicGenerationStateSucceeded {
		t.Fatalf("page status=%s/%s/%s", page.PromptStatus, page.ImageStatus, page.GenerationState)
	}
	if payload.Readiness.PagesTotal != 1 || payload.Readiness.PromptReadyCount != 1 || payload.Readiness.ImageReadyCount != 0 {
		t.Fatalf("readiness=%+v", payload.Readiness)
	}
	if payload.Readiness.CanGenerateImages != true || payload.Readiness.CanBuildCollection != false || payload.Readiness.Skippable != false {
		t.Fatalf("flags=%+v", payload.Readiness)
	}
	// Read-only guarantee: the GET must not add or promote images.
	stored, err := store.ListComicPages(context.Background(), "alice", batch.ID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored[0].Images) != 0 {
		t.Fatalf("read created images: %+v", stored[0].Images)
	}
}
