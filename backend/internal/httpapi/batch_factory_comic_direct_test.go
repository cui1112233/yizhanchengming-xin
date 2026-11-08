package httpapi

import (
	"context"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/n8comic"
)

const comicDirectTestPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

type fakeComicDirectClient struct {
	configured bool
	nextID     int64
	uploads    [][]n8comic.ImageInput
	creates    []n8comic.CreateCarouselRequest
}

func (f *fakeComicDirectClient) Configured() bool { return f.configured }

func (f *fakeComicDirectClient) Publish(context.Context, n8comic.PublishRequest) (n8comic.PublishResult, error) {
	return n8comic.PublishResult{}, nil
}

func (f *fakeComicDirectClient) UploadImages(_ context.Context, request n8comic.UploadImagesRequest) ([]int64, error) {
	f.uploads = append(f.uploads, request.Images)
	ids := make([]int64, len(request.Images))
	for index := range request.Images {
		f.nextID++
		ids[index] = f.nextID
	}
	return ids, nil
}

func (f *fakeComicDirectClient) CreateCarousel(_ context.Context, request n8comic.CreateCarouselRequest) (int64, error) {
	f.creates = append(f.creates, request)
	return 9001, nil
}

func comicDirectTestRouter(t *testing.T, now time.Time) (*fakeComicDirectClient, http.Handler) {
	t.Helper()
	client := &fakeComicDirectClient{configured: true}
	api := NewRouter(RouterOptions{
		BridgeSecret:          "secret",
		Now:                   func() time.Time { return now },
		N8ComicPublisher:      &BatchFactoryN8ComicPublisher{Client: client},
		ComicDirectRecordPath: filepath.Join(t.TempDir(), "comic-direct.json"),
	})
	return client, api
}

func comicDirectPage(ordinal int) map[string]any {
	return map[string]any{"ordinal": ordinal, "dataUrl": "data:image/png;base64," + comicDirectTestPNG}
}

func TestComicDirectUploadsInChunksAndPublishesOnce(t *testing.T) {
	now := time.Unix(1700000000, 0)
	client, api := comicDirectTestRouter(t, now)
	path := "/api/batch-factory/v12/comic-direct/import"

	first := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"materialId": "1355219141083795456", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 3, "finalize": false,
		"pages": []any{comicDirectPage(2), comicDirectPage(1)},
	})
	if first.Code != http.StatusOK {
		t.Fatalf("first status=%d body=%s", first.Code, first.Body.String())
	}
	firstResult := decodeBody[comicDirectResponse](t, first)
	if firstResult.TotalUploaded != 2 || firstResult.UploadedNow != 2 || firstResult.Finalized {
		t.Fatalf("first=%+v", firstResult)
	}
	if len(firstResult.Pages) != 2 || firstResult.Pages[0].Ordinal != 1 || firstResult.Pages[1].Ordinal != 2 {
		t.Fatalf("pages not ordered: %+v", firstResult.Pages)
	}
	if len(client.uploads) != 1 || len(client.uploads[0]) != 2 {
		t.Fatalf("uploads=%d", len(client.uploads))
	}

	second := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"materialId": "1355219141083795456", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 3, "finalize": true,
		"pages":           []any{comicDirectPage(3)},
		"workIds":         []int64{2001},
		"audioCategoryId": 9,
		"audioId":         77,
	})
	if second.Code != http.StatusOK {
		t.Fatalf("second status=%d body=%s", second.Code, second.Body.String())
	}
	secondResult := decodeBody[comicDirectResponse](t, second)
	if secondResult.CarouselID != 9001 || secondResult.Name != "崔家铭1115-第一本" || secondResult.TotalUploaded != 3 {
		t.Fatalf("second=%+v", secondResult)
	}
	if len(client.creates) != 1 {
		t.Fatalf("creates=%d", len(client.creates))
	}
	if !reflect.DeepEqual(client.creates[0].ImageIDs, []int64{1, 2, 3}) {
		t.Fatalf("carousel image ids=%v", client.creates[0].ImageIDs)
	}
	if !reflect.DeepEqual(client.creates[0].WorkIDs, []int64{2001}) || client.creates[0].AudioCategoryID != 9 || client.creates[0].AudioID != 77 {
		t.Fatalf("carousel request=%+v", client.creates[0])
	}

	repeat := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"materialId": "1355219141083795456", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 3, "finalize": true, "pages": []any{comicDirectPage(3)},
		"workIds": []int64{2001}, "audioCategoryId": 9, "audioId": 77,
	})
	replay := decodeBody[comicDirectResponse](t, repeat)
	if !replay.AlreadyPublished || replay.CarouselID != 9001 || len(client.creates) != 1 {
		t.Fatalf("replay=%+v creates=%d", replay, len(client.creates))
	}

	chunkReplay := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"materialId": "1355219141083795456", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 3, "finalize": false, "pages": []any{comicDirectPage(1)},
	})
	if payload := decodeBody[comicDirectResponse](t, chunkReplay); payload.UploadedNow != 0 || len(client.uploads) != 2 {
		t.Fatalf("chunk replay=%+v uploads=%d", payload, len(client.uploads))
	}
}

func TestComicDirectRejectsMissingPageOnFinalize(t *testing.T) {
	now := time.Unix(1700000000, 0)
	_, api := comicDirectTestRouter(t, now)
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/comic-direct/import", map[string]any{
		"materialId": "m1", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 2, "finalize": true, "pages": []any{comicDirectPage(1)},
		"workIds": []int64{2001}, "audioCategoryId": 9, "audioId": 77,
	})
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "缺少第 2 页") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestComicDirectRejectsNonImage(t *testing.T) {
	now := time.Unix(1700000000, 0)
	client, api := comicDirectTestRouter(t, now)
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, "/api/batch-factory/v12/comic-direct/import", map[string]any{
		"materialId": "m2", "bookTitle": "第一本", "producerAdminName": "崔家铭",
		"totalPages": 2, "finalize": false,
		"pages": []any{map[string]any{"ordinal": 1, "dataUrl": "data:text/plain;base64,aGVsbG8="}},
	})
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "不是有效图片") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if len(client.uploads) != 0 {
		t.Fatalf("uploads=%d", len(client.uploads))
	}
}
