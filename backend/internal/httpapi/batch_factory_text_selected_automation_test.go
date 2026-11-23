package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/n8catalog"
)

func TestSignedSelectedTextAutomationUsesOnlyRequestedBook(t *testing.T) {
	now := func() time.Time { return time.Unix(1700000000, 0) }
	store := batchfactoryv11.NewMemoryStore()
	batch := preflightComicBatch(t, store, 2, 0)
	batch = savePreflightBatchSettings(t, store, batch, batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`)})
	presets, err := NewFileAutomationPresetStore(filepath.Join(t.TempDir(), "presets.json"), now)
	if err != nil {
		t.Fatal(err)
	}
	preset, err := presets.Create("alice", "文本合集", map[string]any{"productionMode": "text", "sensitiveMode": "off"})
	if err != nil {
		t.Fatal(err)
	}
	selected := batch.Books[1].ID
	coord := NewBatchAutomationCoordinator(store, presets, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, NewFileAutomationStatusReader(filepath.Join(t.TempDir(), "state.json")), now)
	requests := make(chan txt.EnsureRequest, 2)
	coord.textEnsure = func(_ context.Context, req txt.EnsureRequest) (txt.EnsureResult, error) {
		requests <- req
		return txt.EnsureResult{Stage: "text_collection", Collection: txt.Collection{ID: "selected-collection"}}, nil
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: now, Slice: 2, Store: store, AutomationCoordinator: coord})
	path := "/api/batch-factory/v12/batches/" + batch.ID + "/automation/"
	payload := map[string]any{"presetId": preset.ID, "bookIds": []string{selected}, "productionMode": "text", "runMode": "video_no_submit", "scheduleMode": "execute", "autoPublish": false}
	pre := signedJSONRequest(t, api, now(), "alice", http.MethodPost, path+"preflight", payload)
	if pre.Code != http.StatusOK {
		t.Fatalf("preflight %d %s", pre.Code, pre.Body.String())
	}
	check := decodeBody[automationPreflight](t, pre)
	if len(check.Books) != 1 || check.Books[0].BookID != selected || !check.Books[0].Passed {
		t.Fatalf("preflight selection %s", pre.Body.String())
	}
	start := signedJSONRequest(t, api, now(), "alice", http.MethodPost, path+"start", payload)
	if start.Code != http.StatusCreated {
		t.Fatalf("start %d %s", start.Code, start.Body.String())
	}
	job := waitForTerminalAutomationJob(t, coord, "alice", batch.ID)
	books := mapOrEmpty(job["books"])
	if len(books) != 1 || mapOrEmpty(books[selected])["textCollectionId"] != "selected-collection" {
		t.Fatalf("selection %+v", job)
	}
	select {
	case req := <-requests:
		if req.Owner != "alice" || req.BookID != selected || req.Target != "collection" {
			t.Fatalf("text request %+v", req)
		}
	default:
		t.Fatal("text workflow did not receive selected book")
	}
	if len(requests) > 0 {
		t.Fatal("unselected book was dispatched")
	}
}

type selectedRetryTextRepo struct {
	txt.Repository
	collections map[txt.Scope]txt.Collection
	artifacts   []txt.PageArtifact
}

func (r selectedRetryTextRepo) GetDocument(context.Context, txt.Scope) (txt.Document, error) {
	return txt.Document{}, txt.ErrNotFound
}
func (r selectedRetryTextRepo) GetVersion(_ context.Context, scope txt.Scope, id string) (txt.Version, error) {
	if r.collections[scope].VersionID != id {
		return txt.Version{}, txt.ErrNotFound
	}
	return txt.Version{ID: id}, nil
}
func (r selectedRetryTextRepo) GetCollection(_ context.Context, scope txt.Scope, id string) (txt.Collection, error) {
	c, ok := r.collections[scope]
	if !ok || c.ID != id {
		return txt.Collection{}, txt.ErrNotFound
	}
	return c, nil
}
func (r selectedRetryTextRepo) ListArtifacts(_ context.Context, scope txt.Scope) ([]txt.PageArtifact, error) {
	if _, ok := r.collections[scope]; !ok {
		return nil, txt.ErrNotFound
	}
	return r.artifacts, nil
}

func TestSignedTextRetryTwoBooksCreatesOnlyConfirmedRepublish(t *testing.T) {
	ctx := context.Background()
	now := func() time.Time { return time.Unix(1700000000, 0) }
	store := batchfactoryv11.NewMemoryStore()
	settings := batchfactoryv11.SettingsPatch{
		"productionMode": json.RawMessage(`"text"`), "sensitiveMode": json.RawMessage(`"off"`),
		"comicPublishSettings": json.RawMessage(`{"producerAdminName":"maker","organizationId":"1","audioCategoryId":3,"audioId":77,"workIds":[45]}`),
		"publishSettings":      json.RawMessage(`{"websiteProfileId":"121-1","websiteProfiles":[{"id":"121-1"}],"bookCategory":"小说"}`),
	}
	batch := savePreflightBatchSettings(t, store, preflightComicBatch(t, store, 2, 0), settings)
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	image, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	cover, err := files.SaveImage("cover", "image/png", bytes.NewReader(image))
	if err != nil {
		t.Fatal(err)
	}
	page, err := files.SaveImage("page", "image/png", bytes.NewReader(append(image, 10)))
	if err != nil {
		t.Fatal(err)
	}
	client := &fakeN8ComicClient{configured: true}
	directory := fakeN8ComicAudioDirectory{categories: []n8catalog.AudioCategory{{ID: 3, Name: "分类", Status: "ENABLE"}}, audios: []n8catalog.Audio{{ID: 77, OrganizationID: 1, CategoryIDs: []int64{3}, Status: "ENABLE"}}}
	publisher := &BatchFactoryN8ComicPublisher{Store: store, Client: client, LocalFiles: files, ProducerDirectory: directory, ProducerAccounts: []string{"maker"}}
	coord := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, NewFileAutomationStatusReader(filepath.Join(t.TempDir(), "state.json")), now)
	repo := selectedRetryTextRepo{collections: map[txt.Scope]txt.Collection{}, artifacts: []txt.PageArtifact{
		{ID: "cover", Kind: "cover", ArtifactFile: txt.ArtifactFile{StorageRef: cover.StorageRef, MediaType: cover.MediaType, ByteSize: cover.ByteSize, SHA256: cover.SHA256}},
		{ID: "page-image", Kind: "page", PageID: "page", ArtifactFile: txt.ArtifactFile{StorageRef: page.StorageRef, MediaType: page.MediaType, ByteSize: page.ByteSize, SHA256: page.SHA256}},
	}}
	current := collectionDraftRecords{store}
	entries := map[string]any{}
	frozen, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	overrides := map[string]any{}
	for i, book := range batch.Books {
		ids = append(ids, book.ID)
		scope := txt.Scope{Owner: "alice", BatchID: batch.ID, BookID: book.ID}
		id := "collection-" + book.ID
		repo.collections[scope] = txt.Collection{ID: id, Status: "ready", VersionID: "v", PaginationVersion: 1, CoverArtifactID: "cover",
			Items: []txt.CollectionItem{{Ordinal: 1, PageID: "page", ArtifactID: "page-image", StorageRef: page.StorageRef, MediaType: page.MediaType, SHA256: page.SHA256}}}
		record := cp.Record{Kind: cp.CollectionKindText, BookID: book.ID, CollectionID: id, IntentID: "original-" + book.ID, Status: "published", CarouselID: 19, ImageIDs: []int64{11, 12}, PageCount: 2}
		if err := current.Save(ctx, textPublishRequest(scope, id), record); err != nil {
			t.Fatal(err)
		}
		entries[book.ID] = map[string]any{"bookId": book.ID, "stage": "text_publish", "status": "failed", "error": "上传失败", "frozenTextSettings": string(frozen)}
		overrides[book.ID] = map[string]any{"collectionId": id, "force": i == 0}
	}

	job := map[string]any{"id": "original", "owner": "alice", "batchId": batch.ID, "state": "failed", "runMode": "full_submit", "concurrency": 1, "books": entries}
	if err := coord.setJob("alice", batch.ID, "original", job); err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Now: now, Slice: 2, Store: store, AutomationCoordinator: coord, TextRepository: repo, LocalArtifacts: files, N8ComicPublisher: publisher})
	resp := signedJSONRequest(t, api, now(), "alice", http.MethodPost, "/api/batch-factory/v12/batches/"+batch.ID+"/automation/retry", map[string]any{"bookIds": ids, "textPublishOverrides": overrides})
	if resp.Code != http.StatusOK {
		t.Fatalf("retry %d %s", resp.Code, resp.Body.String())
	}
	job = waitForTerminalAutomationJob(t, coord, "alice", batch.ID)
	for i, id := range ids {
		scope := txt.Scope{Owner: "alice", BatchID: batch.ID, BookID: id}
		collection := repo.collections[scope]
		entry := mapOrEmpty(mapOrEmpty(job["books"])[id])
		rec, ok, err := current.Load(ctx, textPublishRequest(scope, collection.ID))
		runs, runErr := store.ListBookStageRuns(ctx, "alice", batch.ID, id)
		if runErr != nil || len(runs) != 1 || runs[0].Stage != "text_publish" || rec.RunID != runs[0].ID || rec.Attempt != runs[0].Attempt || rec.RequestKey != runs[0].RequestID {
			t.Fatalf("book %s runs=%+v receipt=%+v err=%v", id, runs, rec, runErr)
		}
		if err != nil || !ok || !rec.Confirmed() || entry["status"] != "completed" || rec.CollectionID != collection.ID {
			t.Fatalf("book %s entry=%v receipt=%+v err=%v", id, entry, rec, err)
		}
		if (rec.IntentID == "original-"+id) != (i == 1) {
			t.Fatalf("incorrect publish intent for %s: %+v", id, rec)
		}
	}
	if client.calls != 1 {
		t.Fatalf("two retries issued %d external requests; want one confirmed republish", client.calls)
	}
}
