package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/n8catalog"
	"qiantie/backend/internal/n8comic"
)

type failingProducerDirectory struct{ err error }

func (f failingProducerDirectory) Configured() bool { return true }
func (f failingProducerDirectory) ListProducerAccounts(context.Context) ([]string, error) {
	return nil, f.err
}

func TestProducerDirectoryFailureLogsOnlySafeCategory(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })
	publisher := BatchFactoryN8ComicPublisher{ProducerAccounts: []string{"fallback"}, ProducerDirectory: failingProducerDirectory{err: &n8catalog.DirectoryError{Kind: "http", Status: 401}}}
	accounts, source, warning := publisher.producerAccounts(context.Background())
	if !reflect.DeepEqual(accounts, []string{"fallback"}) || source != "cache" || warning == "" || !strings.Contains(output.String(), "kind=http status=401") {
		t.Fatalf("fallback or diagnostic incorrect: %v %s %s %s", accounts, source, warning, output.String())
	}
	output.Reset()
	publisher.ProducerDirectory = failingProducerDirectory{err: fmt.Errorf("sensitive-token")}
	publisher.producerAccounts(context.Background())
	if strings.Contains(output.String(), "sensitive-token") || !strings.Contains(output.String(), "kind=unknown") {
		t.Fatalf("unsafe diagnostic: %s", output.String())
	}
}

type fakeN8ComicClient struct {
	configured bool
	request    n8comic.PublishRequest
	calls      int
}

func (f *fakeN8ComicClient) Configured() bool { return f.configured }

func (f *fakeN8ComicClient) Publish(_ context.Context, request n8comic.PublishRequest) (n8comic.PublishResult, error) {
	f.calls++
	f.request = request
	return n8comic.PublishResult{CarouselID: 456, ImageIDs: []int64{101, 102}}, nil
}

func TestN8ComicPublishExplainsMissingRequiredSettings(t *testing.T) {
	publisher := &BatchFactoryN8ComicPublisher{
		Store:      batchfactoryv11.NewMemoryStore(),
		LocalFiles: localartifact.NewStore(t.TempDir(), 1<<20),
		Client:     &fakeN8ComicClient{configured: true},
	}
	_, err := publisher.Publish(context.Background(), "alice", "batch-1", "book-1", "collection-1", n8ComicPublishInput{})
	if err == nil || !strings.Contains(err.Error(), "制作者") || !strings.Contains(err.Error(), "音频分类") || !strings.Contains(err.Error(), "音频") || !strings.Contains(err.Error(), "作品") {
		t.Fatalf("publish error=%v, want the missing setting names", err)
	}
}

func TestN8ComicPublishRouteUsesCollectionMainImagesAndServerNamingRule(t *testing.T) {
	now := time.Date(2026, 10, 26, 8, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本", Videos: []batchfactoryv11.CreateVideoInput{{Label: "一"}, {Label: "二"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	for index, video := range book.Videos {
		page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, video.ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
		if err != nil {
			t.Fatal(err)
		}
		content := append([]byte(nil), pngBytes...)
		content = append(content, bytes.Repeat([]byte{'\n'}, index)...)
		artifact, err := files.SaveImage("n8-publish-image-"+page.ID, "image/png", bytes.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"}); err != nil {
			t.Fatal(err)
		}
	}
	collection, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(missing) != 0 {
		t.Fatalf("collection=%+v missing=%v err=%v", collection, missing, err)
	}
	client := &fakeN8ComicClient{configured: true}
	api := NewRouter(RouterOptions{
		BridgeSecret:     "secret",
		Now:              func() time.Time { return now },
		Slice:            1,
		Store:            store,
		LocalArtifacts:   files,
		N8ComicPublisher: &BatchFactoryN8ComicPublisher{Store: store, LocalFiles: files, Client: client, Now: func() time.Time { return now }},
	})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + book.ID + "/comic-collections/" + collection.ID + "/publish"
	response := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"producerAdminName": "崔家铭",
		"audioCategoryId":   3,
		"audioId":           77,
		"workIds":           []int64{2001, 2002},
	})
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if client.request.Name != "崔家铭1026-第一本" || client.request.ProducerAdminName != "崔家铭" || client.request.AudioCategoryID != 3 {
		t.Fatalf("request=%+v", client.request)
	}
	if !reflect.DeepEqual(client.request.WorkIDs, []int64{2001, 2002}) || len(client.request.Images) != 2 {
		t.Fatalf("request=%+v", client.request)
	}
	if !strings.HasSuffix(client.request.Images[0].FileName, ".png") {
		t.Fatalf("image file name=%q", client.request.Images[0].FileName)
	}
	result := decodeBody[struct {
		Publish N8ComicPublishRecord `json:"publish"`
	}](t, response).Publish
	if result.CarouselID != 456 || result.Name != "崔家铭1026-第一本" || result.Status != "published" {
		t.Fatalf("publish=%+v", result)
	}

	repeat := signedJSONRequest(t, api, now, "alice", http.MethodPost, path, map[string]any{
		"producerAdminName": "崔家铭", "audioCategoryId": 3, "workIds": []int64{2001, 2002},
	})
	if repeat.Code != http.StatusOK {
		t.Fatalf("repeat status=%d body=%s", repeat.Code, repeat.Body.String())
	}
	if client.request.Name != "崔家铭1026-第一本" {
		t.Fatal("repeat publish changed client request")
	}
	if client.calls != 1 {
		t.Fatalf("repeat publish called N8 %d times", client.calls)
	}

	readback := signedJSONRequest(t, api, now, "alice", http.MethodGet, path, nil)
	if readback.Code != http.StatusOK {
		t.Fatalf("readback status=%d body=%s", readback.Code, readback.Body.String())
	}
	readbackPayload := decodeBody[struct {
		Published bool                 `json:"published"`
		Publish   N8ComicPublishRecord `json:"publish"`
	}](t, readback)
	if !readbackPayload.Published || readbackPayload.Publish.CarouselID != 456 {
		t.Fatalf("readback=%+v", readbackPayload)
	}
}

type fakeN8ComicAudioDirectory struct {
	audios     []n8catalog.Audio
	categories []n8catalog.AudioCategory
}

func (f fakeN8ComicAudioDirectory) Configured() bool { return true }
func (f fakeN8ComicAudioDirectory) ListProducerAccounts(context.Context) ([]string, error) {
	return nil, nil
}
func (f fakeN8ComicAudioDirectory) ListAudio(context.Context, string) ([]n8catalog.Audio, error) {
	return f.audios, nil
}
func (f fakeN8ComicAudioDirectory) ListAudioCategories(context.Context) ([]n8catalog.AudioCategory, error) {
	return f.categories, nil
}

func TestN8ComicAutomaticPublishSelectsDefaultAudioSettings(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "自动音频", Videos: []batchfactoryv11.CreateVideoInput{{Label: "第一幕"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	page, err := store.UpsertComicPage(context.Background(), "alice", batch.ID, book.ID, book.Videos[0].ID, batchfactoryv11.UpsertComicPageInput{ComicPrompt: "漫画页", ImagePrompt: "出图", Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	pngBytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := files.SaveImage("n8-default-audio", "image/png", bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateComicPageImage(context.Background(), "alice", batch.ID, book.ID, page.ID, batchfactoryv11.CreateComicPageImageInput{StorageRef: artifact.StorageRef, MediaType: "image/png", Source: "upload"}); err != nil {
		t.Fatal(err)
	}
	collection, missing, err := store.CreateComicCollection(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(missing) != 0 {
		t.Fatalf("collection=%+v missing=%v err=%v", collection, missing, err)
	}
	client := &fakeN8ComicClient{configured: true}
	publisher := &BatchFactoryN8ComicPublisher{
		Store:      store,
		LocalFiles: files,
		Client:     client,
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 7, Name: "停用", Status: "DISABLE"}, {ID: 9, Name: "默认", Status: "ENABLE"}},
			audios: []n8catalog.Audio{
				{ID: 201, Name: "旧音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE", UpdatedAt: "2026-11-01 12:00:00"},
				{ID: 202, Name: "最新音频", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE", UpdatedAt: "2026-11-02 12:00:00"},
			},
		},
	}
	settings := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"producerAdminName":"制作者","organizationId":"1","workIds":[9001]}`)}
	if _, err := publisher.PublishFromSettings(context.Background(), "alice", batch.ID, book.ID, collection, settings); err != nil {
		t.Fatalf("automatic publish error=%v", err)
	}
	if client.request.AudioCategoryID != 9 || client.request.AudioID != 202 {
		t.Fatalf("publish request used category=%d audio=%d, want category=9 audio=202", client.request.AudioCategoryID, client.request.AudioID)
	}
}

func TestN8ComicAutomaticPublishExplainsWhenNoDefaultAudioIsAvailable(t *testing.T) {
	publisher := &BatchFactoryN8ComicPublisher{
		Store:      batchfactoryv11.NewMemoryStore(),
		LocalFiles: localartifact.NewStore(t.TempDir(), 1<<20),
		Client:     &fakeN8ComicClient{configured: true},
		ProducerDirectory: fakeN8ComicAudioDirectory{
			categories: []n8catalog.AudioCategory{{ID: 9, Name: "默认", Status: "ENABLE"}},
		},
	}
	settings := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"producerAdminName":"制作者","workIds":[9001]}`)}
	_, err := publisher.PublishFromSettings(context.Background(), "alice", "batch-1", "book-1", batchfactoryv11.ComicCollection{ID: "collection-1"}, settings)
	if err == nil || !strings.Contains(err.Error(), "没有可用指定音频") {
		t.Fatalf("publish error=%v, want no usable audio explanation", err)
	}
}

func TestN8ComicAudioRouteRequiresFiltersAndSortsCandidates(t *testing.T) {
	now := time.Unix(1700000000, 0)
	publisher := &BatchFactoryN8ComicPublisher{Client: &fakeN8ComicClient{configured: true}, ProducerDirectory: fakeN8ComicAudioDirectory{audios: []n8catalog.Audio{
		{ID: 3, Name: "新", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE", UpdatedAt: "2026-11-01 12:00:00"},
		{ID: 4, Name: "同时间更大", OrganizationID: 1, CategoryIDs: []int64{9}, Status: "ENABLE", UpdatedAt: "2026-11-01 12:00:00"},
		{ID: 5, Name: "其他组织", OrganizationID: 2, CategoryIDs: []int64{9}, Status: "ENABLE", UpdatedAt: "2026-11-02 12:00:00"},
	}}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, N8ComicPublisher: publisher})
	missing := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/comic-publish/audios", nil)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("missing status=%d", missing.Code)
	}
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/comic-publish/audios?audioCategoryId=9&organizationId=1", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	payload := decodeBody[struct {
		Audios []n8catalog.Audio `json:"audios"`
	}](t, response)
	if !reflect.DeepEqual([]int64{payload.Audios[0].ID, payload.Audios[1].ID}, []int64{4, 3}) || payload.Audios[0].UpdatedAt != "2026-11-01 12:00:00" {
		t.Fatalf("audios=%+v", payload.Audios)
	}
}

func TestN8ComicPublishStatusReportsServerConfiguration(t *testing.T) {
	now := time.Unix(1700000000, 0)
	client := &fakeN8ComicClient{configured: true}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, N8ComicPublisher: &BatchFactoryN8ComicPublisher{Client: client, ProducerAccounts: []string{"崔家铭", "张三"}}})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/comic-publish/status", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	status := decodeBody[struct {
		Configured       bool     `json:"configured"`
		ProducerAccounts []string `json:"producerAccounts"`
	}](t, response)
	if !status.Configured {
		t.Fatalf("status=%+v", status)
	}
	if !reflect.DeepEqual(status.ProducerAccounts, []string{"崔家铭", "张三"}) {
		t.Fatalf("producer accounts=%v", status.ProducerAccounts)
	}
}

type fakeN8ComicWorkDirectory struct {
	configured bool
	works      []n8catalog.Work
}

func (f fakeN8ComicWorkDirectory) Configured() bool { return f.configured }
func (f fakeN8ComicWorkDirectory) SearchWorks(context.Context, string) ([]n8catalog.Work, error) {
	return f.works, nil
}

func TestN8ComicPublisherResolvesTitleMatchForAutomaticPublish(t *testing.T) {
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "漫画批次",
		Books: []batchfactoryv11.CreateBookInput{{Title: "鼓励式老公，把行动都给了闺蜜"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	publisher := &BatchFactoryN8ComicPublisher{
		Store: store,
		WorksDirectory: fakeN8ComicWorkDirectory{configured: true, works: []n8catalog.Work{
			{ID: 9001, Name: "鼓励式老公,却把行动都给了闺蜜"},
		}},
	}
	ids, err := publisher.resolveComicWorkIDs(context.Background(), "alice", batch.ID, batch.Books[0].ID)
	if err != nil || !reflect.DeepEqual(ids, []int64{9001}) {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestN8ComicWorkSearchRouteReturnsMatchingWorks(t *testing.T) {
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, N8ComicPublisher: &BatchFactoryN8ComicPublisher{
		Client:         &fakeN8ComicClient{configured: true},
		WorksDirectory: fakeN8ComicWorkDirectory{configured: true, works: []n8catalog.Work{{ID: 9001, Name: "小妾造反后"}}},
	}})
	response := signedJSONRequest(t, api, now, "alice", http.MethodGet, "/api/batch-factory/v12/comic-publish/works?query=%E5%B0%8F%E5%A6%BE%E9%80%A0%E5%8F%8D%E5%90%8E", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Works []n8catalog.Work `json:"works"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(payload.Works, []n8catalog.Work{{ID: 9001, Name: "小妾造反后"}}) {
		t.Fatalf("works=%#v", payload.Works)
	}
}
