package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	cp "qiantie/backend/internal/collectionpublish"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/n8catalog"
	"qiantie/backend/internal/n8comic"
)

type textPublishTestRepo struct {
	txt.Repository
	scope      txt.Scope
	collection txt.Collection
	cover      txt.PageArtifact
}

func (f textPublishTestRepo) GetCollection(_ context.Context, s txt.Scope, id string) (txt.Collection, error) {
	if s != f.scope || id != f.collection.ID {
		return txt.Collection{}, txt.ErrNotFound
	}
	return f.collection, nil
}
func (f textPublishTestRepo) ListArtifacts(_ context.Context, s txt.Scope) ([]txt.PageArtifact, error) {
	if s != f.scope {
		return nil, txt.ErrNotFound
	}
	return []txt.PageArtifact{f.cover}, nil
}

type textPublishTestClient struct {
	confirmedReadback bool
	readbackCalls     int
	calls             int
	images            []n8comic.ImageInput
}

func (*textPublishTestClient) Configured() bool { return true }
func (f *textPublishTestClient) Publish(_ context.Context, r n8comic.PublishRequest) (n8comic.PublishResult, error) {
	f.calls++
	f.images = r.Images
	if f.calls > 1 {
		return n8comic.PublishResult{ImageIDs: []int64{21, 22}, ImageSignatures: []string{"signature1", "signature2"}, ReceiptPath: "/receipt?private-proof=7"}, n8comic.ErrCarouselReceiptPending
	}
	return n8comic.PublishResult{CarouselID: 19, ImageIDs: []int64{11, 12}}, nil
}

func (f *textPublishTestClient) Readback(_ context.Context, in n8comic.ReceiptLookup) (n8comic.PublishResult, error) {
	f.readbackCalls++
	if f.confirmedReadback && in.ReceiptPath == "/receipt?private-proof=7" && len(in.ImageSignatures) == 2 && in.PreviousCarouselID == 19 {
		return n8comic.PublishResult{CarouselID: 29, ImageIDs: []int64{21, 22}}, nil
	}
	return n8comic.PublishResult{}, nil
}

func TestTextPublishSignedReadbackAndForcedPending(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	settings := batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`), "comicPublishSettings": json.RawMessage(`{"producerAdminName":"maker","organizationId":"1","audioCategoryId":3,"audioId":77,"workIds":[45]}`), "publishSettings": json.RawMessage(`{"websiteProfileId":"121-1","websiteProfiles":[{"id":"121-1"}],"bookCategory":"小说"}`)}
	batch, err := store.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{Title: "批次", InitialBatchSettings: settings, Books: []batchfactoryv11.CreateBookInput{{Title: "作品"}}})
	if err != nil {
		t.Fatal(err)
	}
	s := txt.Scope{Owner: "alice", BatchID: batch.ID, BookID: batch.Books[0].ID}
	f := localartifact.NewStore(t.TempDir(), 1<<20)
	png, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9Y9ZgS8AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	pageData := append(append([]byte(nil), png...), 10)
	cover, err := f.SaveImage("collection-cover", "image/png", bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	page, err := f.SaveImage("collection-page", "image/png", bytes.NewReader(pageData))
	if err != nil {
		t.Fatal(err)
	}
	repo := textPublishTestRepo{scope: s, collection: txt.Collection{ID: "text-c", Status: "ready", CoverArtifactID: "c", Items: []txt.CollectionItem{{Ordinal: 1, StorageRef: page.StorageRef, MediaType: "image/png"}}}, cover: txt.PageArtifact{ID: "c", Kind: "cover", ArtifactFile: txt.ArtifactFile{StorageRef: cover.StorageRef, MediaType: "image/png"}}}
	client := &textPublishTestClient{}
	directory := fakeN8ComicAudioDirectory{categories: []n8catalog.AudioCategory{{ID: 3, Name: "分类", Status: "ENABLE"}}, audios: []n8catalog.Audio{{ID: 77, OrganizationID: 1, CategoryIDs: []int64{3}, Status: "ENABLE"}}}
	p := &BatchFactoryN8ComicPublisher{Store: store, Client: client, LocalFiles: f, ProducerDirectory: directory, ProducerAccounts: []string{"maker"}}
	api := NewRouter(RouterOptions{BridgeSecret: "secret", Slice: 1, Store: store, TextRepository: repo, LocalArtifacts: f, N8ComicPublisher: p, Now: func() time.Time { return now }})
	base := "/api/batch-factory/v12/batches/" + s.BatchID + "/books/" + s.BookID + "/"
	path := base + "text-collections/text-c/publish"
	mixed := map[string]any{"collections": []cp.Reference{{Kind: cp.CollectionKindText, ID: "text-c"}, {Kind: cp.CollectionKindComic, ID: "comic-c"}}}
	for _, route := range []string{path, base + "comic-collections/comic-c/publish"} {
		got := signedJSONRequest(t, api, now, "alice", "POST", route, mixed)
		if got.Code != 400 || decodeBody[map[string]string](t, got)["error"] != cp.ErrMixedKinds.Error() || client.calls != 0 {
			t.Fatalf("mixed: %d %s", got.Code, got.Body.String())
		}
	}
	type response struct {
		Published              bool
		Publish, LastPublished *cp.Record
	}
	call := func(method string, input any) response {
		r := signedJSONRequest(t, api, now, "alice", method, path, input)
		if r.Code != 200 {
			t.Fatalf("%s %d %s", method, r.Code, r.Body.String())
		}
		return decodeBody[response](t, r)
	}
	if g := call("GET", nil); g.Published || g.Publish != nil {
		t.Fatal(g)
	}
	automatic := textCollectionPublisher{n8: p, repo: repo}
	confirmed, err := automatic.automation(ctx, txt.EnsureRequest{Scope: s, FrozenSettings: settings, PublishRunID: "run-1", PublishAttempt: 1, PublishRequestKey: "request-1"}, repo.collection)
	if err != nil || !confirmed {
		t.Fatalf("automatic confirmed=%v err=%v", confirmed, err)
	}
	first := call("POST", map[string]any{})
	if first.Publish.RunID != "run-1" || first.Publish.Attempt != 1 || first.Publish.RequestKey != "request-1" {
		t.Fatal(first.Publish)
	}
	if !first.Published || first.Publish.CarouselID != 19 || first.Publish.CategoryID != "小说" || first.Publish.OrganizationID != "1" || len(client.images) != 2 || !bytes.Equal(client.images[0].Content, png) || !bytes.Equal(client.images[1].Content, pageData) {
		t.Fatal(first, client.images)
	}
	if g := call("POST", map[string]any{}); g.Publish.IntentID != first.Publish.IntentID || client.calls != 1 {
		t.Fatal(g)
	}
	if g := call("GET", nil); !g.Published || g.Publish.IntentID != first.Publish.IntentID {
		t.Fatal(g)
	}
	stored, err := store.GetBatch(ctx, s.Owner, s.BatchID)
	if err != nil || stored.Books[0].SourceMetadata["textCollectionId"] != repo.collection.ID {
		t.Fatalf("manual selection=%v err=%v", stored.Books[0].SourceMetadata, err)
	}
	runs, err := store.ListBookStageRuns(ctx, s.Owner, s.BatchID, s.BookID)
	if err != nil || len(runs) != 0 {
		t.Fatalf("manual upload created runs: %v %v", runs, err)
	}
	second := call("POST", map[string]any{"force": true, "requestKey": "another-upload"})
	if second.Published || second.Publish.Status != "pending" || second.LastPublished == nil || second.LastPublished.IntentID != first.Publish.IntentID {
		t.Fatal(second)
	}
	if g := call("GET", nil); g.Published || g.LastPublished.CarouselID != 19 || g.Publish.IntentID != second.Publish.IntentID {
		t.Fatal(g)
	}
	call("POST", map[string]any{"force": true, "requestKey": "another-upload"})
	old := call("POST", map[string]any{"force": true, "requestKey": "request-1"})
	if !old.Published || old.Publish.IntentID != first.Publish.IntentID {
		t.Fatal(old)
	}
	if current := call("GET", nil); current.Publish.IntentID != second.Publish.IntentID || current.Published {
		t.Fatal(current)
	}
	if client.calls != 2 {
		t.Fatalf("duplicate calls=%d", client.calls)
	}
	client.confirmedReadback = true
	current := signedJSONRequest(t, api, now, "alice", "GET", path, nil)
	if bytes.Contains(current.Body.Bytes(), []byte("private-proof")) || bytes.Contains(current.Body.Bytes(), []byte("signature1")) {
		t.Fatal("private readback data leaked")
	}
	got := decodeBody[response](t, current)
	if !got.Published || got.Publish.CarouselID != 29 || got.Publish.IntentID != second.Publish.IntentID || client.calls != 2 || client.readbackCalls == 0 {
		t.Fatal(got)
	}
	bound, err := automatic.automation(ctx, txt.EnsureRequest{Scope: s, FrozenSettings: settings, PublishRunID: "next-run", PublishAttempt: 2, PublishRequestKey: "next-key"}, repo.collection)
	if err != nil || !bound {
		t.Fatal(bound, err)
	}
	reused := call("GET", nil).Publish
	if reused.RunID != "next-run" || reused.Attempt != 2 || reused.RequestKey != "next-key" || reused.IntentID != second.Publish.IntentID || reused.LastPublished.RequestKey != "another-upload" || client.calls != 2 {
		t.Fatal(reused)
	}
	foreign := signedJSONRequest(t, api, now, "bob", http.MethodGet, path, nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign=%d", foreign.Code)
	}
}
