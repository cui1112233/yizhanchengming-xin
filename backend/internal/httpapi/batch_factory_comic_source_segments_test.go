package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type comicSegmentationHTTPProvider struct {
	calls int
}

func (p *comicSegmentationHTTPProvider) Complete(_ context.Context, _ batchfactoryv11.TextCompletionRequest) (string, error) {
	p.calls++
	return `{"boundaryIds":["b2","b4","b6","b8","b10","b12","b14","b16","b18","b20","b22","b24"]}`, nil
}

func testSourceDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func TestComicSourceSegmentRoutesCreateListAndSplit(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	source := strings.Repeat("这是一句完整剧情，人物行动并推动冲突。", 16)
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: source}}})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	base := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-source-segments"

	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/segment", map[string]any{})
	if created.Code != http.StatusOK {
		t.Fatalf("segment status=%d body=%s", created.Code, created.Body.String())
	}
	result := decodeBody[struct {
		Segmentation batchfactoryv11.ComicSourceSegmentation `json:"segmentation"`
	}](t, created).Segmentation
	if len(result.Segments) < 2 || result.Stale {
		t.Fatalf("unexpected segmentation: %#v", result)
	}

	listed := signedJSONRequest(t, api, now, "alice", http.MethodGet, base, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	current := result.Segments[0]
	if len(current.SafeBoundaries) == 0 {
		t.Fatalf("first page has no manual split boundary: %#v", current)
	}
	split := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/"+current.VideoID+"/split", map[string]any{
		"direction": "right", "boundaryOffset": current.SafeBoundaries[0].Offset, "expectedRevision": current.Revision,
	})
	if split.Code != http.StatusOK {
		t.Fatalf("split status=%d body=%s", split.Code, split.Body.String())
	}
	after := decodeBody[struct {
		Segmentation batchfactoryv11.ComicSourceSegmentation `json:"segmentation"`
	}](t, split).Segmentation
	if len(after.Segments) != len(result.Segments)+1 || after.Segments[0].VideoID != current.VideoID {
		t.Fatalf("unexpected right split result: %#v", after.Segments)
	}
}

func TestComicSourceSegmentRouteUsesAISafeBoundarySelection(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	source := strings.Repeat("第一段剧情完整结束。", 6) + strings.Repeat("第二段剧情完整结束。", 6) + strings.Repeat("第三段剧情完整结束。", 6) + strings.Repeat("第四段剧情完整结束。", 6)
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: source}}})
	if err != nil {
		t.Fatal(err)
	}
	provider := &comicSegmentationHTTPProvider{}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store,
		Director: &batchfactoryv11.DirectorService{Store: store, Provider: provider},
	})
	base := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-source-segments"
	created := signedJSONRequest(t, api, now, "alice", http.MethodPost, base+"/segment", map[string]any{})
	if created.Code != http.StatusOK {
		t.Fatalf("segment status=%d body=%s", created.Code, created.Body.String())
	}
	if provider.calls != 1 {
		t.Fatalf("AI segmentation calls=%d, want 1", provider.calls)
	}
	result := decodeBody[struct {
		Segmentation batchfactoryv11.ComicSourceSegmentation `json:"segmentation"`
	}](t, created).Segmentation
	if len(result.Segments) == 0 {
		t.Fatal("AI-backed segmentation returned no segments")
	}
}

func TestComicSourceSegmentRoutesRejectUnsafeAndConflictedSplit(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	source := strings.Repeat("这是完整句子，人物继续行动。", 12)
	batch, _ := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "漫画", Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", SourceText: source}}})
	drafts, _ := batchfactoryv11.FallbackComicSourceSegments(source, batchfactoryv11.EnumerateComicSourceBoundaries(source))
	segmentation, _ := store.ReplaceComicSourceSegments(context.Background(), "alice", batch.ID, batch.Books[0].ID, testSourceDigest(source), drafts)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store})
	base := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/comic-source-segments/" + segmentation.Segments[0].VideoID + "/split"
	unsafe := signedJSONRequest(t, api, now, "alice", http.MethodPost, base, map[string]any{"direction": "left", "boundaryOffset": segmentation.Segments[0].StartOffset + 1, "expectedRevision": segmentation.Segments[0].Revision})
	if unsafe.Code != http.StatusBadRequest {
		t.Fatalf("unsafe split status=%d body=%s", unsafe.Code, unsafe.Body.String())
	}
	conflict := signedJSONRequest(t, api, now, "alice", http.MethodPost, base, map[string]any{"direction": "left", "boundaryOffset": segmentation.Segments[0].SafeBoundaries[0].Offset, "expectedRevision": 999})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status=%d body=%s", conflict.Code, conflict.Body.String())
	}
}
