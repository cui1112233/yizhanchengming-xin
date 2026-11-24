package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	artifacts  []txt.PageArtifact
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
	return f.artifacts, nil
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
		return n8comic.PublishResult{ImageIDs: []int64{21, 22, 23}, ImageSignatures: []string{"signature1", "signature2", "signature3"}, ReceiptPath: "/receipt?private-proof=7"}, n8comic.ErrCarouselReceiptPending
	}
	return n8comic.PublishResult{CarouselID: 19, ImageIDs: []int64{11, 12, 13}}, nil
}

func (f *textPublishTestClient) Readback(_ context.Context, in n8comic.ReceiptLookup) (n8comic.PublishResult, error) {
	f.readbackCalls++
	if f.confirmedReadback && in.ReceiptPath == "/receipt?private-proof=7" && len(in.ImageSignatures) == 3 && in.PreviousCarouselID == 19 {
		return n8comic.PublishResult{ReceiptIntentID: in.IntentID, CarouselID: 29, ImageIDs: []int64{21, 22, 23}}, nil
	}
	return n8comic.PublishResult{}, nil
}

func TestTextPublishSettingsIgnoreLegacyWebsiteMapping(t *testing.T) {
	settings := batchfactoryv11.SettingsPatch{"comicPublishSettings": json.RawMessage(`{"producerAdminName":"maker","organizationId":"1","audioCategoryId":3,"audioId":77,"workIds":[45]}`)}
	before, err := workflowPublishSettingsKey(settings)
	if err != nil {
		t.Fatal(err)
	}
	settings["publishSettings"] = json.RawMessage(`{"websiteProfileId":"old-site","bookCategory":"旧分类"}`)
	after, err := workflowPublishSettingsKey(settings)
	if err != nil || before != after {
		t.Fatalf("legacy fields changed upload routing: %v", err)
	}
	if err := cp.ValidateMapping(cp.PublishRequest{Kind: cp.CollectionKindText, Producer: "maker", OrganizationID: "1", AudioCategoryID: 3, AudioID: 77, WorkIDs: []int64{45}}); err != nil {
		t.Fatalf("valid shared mapping rejected: %v", err)
	}
}

func TestTextPublishSignedReadbackAndForcedPending(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	settings := batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`), "comicPublishSettings": json.RawMessage(`{"producerAdminName":"maker","organizationId":"1","audioCategoryId":3,"audioId":77,"workIds":[45]}`)}
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
	pageData2 := append(append([]byte(nil), png...), 20)
	pageData3 := append(append([]byte(nil), png...), 30)
	cover, err := f.SaveImage("collection-cover", "image/png", bytes.NewReader(png))
	if err != nil {
		t.Fatal(err)
	}
	page, err := f.SaveImage("collection-page-1", "image/png", bytes.NewReader(pageData))
	if err != nil {
		t.Fatal(err)
	}
	page2, err := f.SaveImage("collection-page-2", "image/png", bytes.NewReader(pageData2))
	if err != nil {
		t.Fatal(err)
	}
	page3, err := f.SaveImage("collection-page-3", "image/png", bytes.NewReader(pageData3))
	if err != nil {
		t.Fatal(err)
	}
	repo := textPublishTestRepo{
		scope: s,
		collection: txt.Collection{ID: "text-c", Status: "ready", CoverArtifactID: "c", Items: []txt.CollectionItem{
			{Ordinal: 1, StorageRef: page.StorageRef, MediaType: "image/png"},
			{Ordinal: 2, StorageRef: page2.StorageRef, MediaType: "image/png"},
			{Ordinal: 3, StorageRef: page3.StorageRef, MediaType: "image/png"},
		}},
		artifacts: []txt.PageArtifact{{ID: "c", Kind: "cover", ArtifactFile: txt.ArtifactFile{StorageRef: cover.StorageRef, MediaType: "image/png"}}},
	}
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
	foreignScopes := [][]string{
		{"bob", s.BatchID, s.BookID, "text-c"},
		{s.Owner, "other-batch", s.BookID, "text-c"},
		{s.Owner, s.BatchID, "other-book", "text-c"},
		{s.Owner, s.BatchID, s.BookID, "other-collection"},
	}
	for i, scope := range foreignScopes {
		key := "foreign-" + strconv.Itoa(i)
		encoded, _ := json.Marshal(scope)
		if err := store.BindTextPublishKeys(ctx, map[string]string{key: string(encoded)}); err != nil {
			t.Fatal(err)
		}
		for _, force := range []bool{true, false} {
			resp := signedJSONRequest(t, api, now, s.Owner, "POST", path, map[string]any{"force": force, "requestKey": key})
			if resp.Code != http.StatusConflict || client.calls != 0 || client.readbackCalls != 0 {
				t.Fatalf("foreign scope %d force=%t: HTTP %d calls=%d reads=%d", i, force, resp.Code, client.calls, client.readbackCalls)
			}
		}
	}
	for _, invalid := range []string{"", "bad key!"} {
		resp := signedJSONRequest(t, api, now, s.Owner, "POST", path, map[string]any{"force": true, "requestKey": invalid})
		if resp.Code != http.StatusBadRequest || client.calls != 0 {
			t.Fatalf("invalid confirmation: HTTP %d calls=%d", resp.Code, client.calls)
		}
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
	if !first.Published || first.Publish.CarouselID != 19 || first.Publish.CategoryID != "" || first.Publish.OrganizationID != "1" || len(client.images) != 3 || !bytes.Equal(client.images[0].Content, pageData) || !bytes.Equal(client.images[1].Content, pageData2) || !bytes.Equal(client.images[2].Content, pageData3) {
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
	coord := &BatchAutomationCoordinator{store: store, textPublishCollection: func(_ context.Context, _ txt.Scope, id string) (txt.Collection, error) {
		return txt.Collection{ID: id, Status: "ready"}, nil
	}}
	if _, err := coord.textPublishOverrides(ctx, s.Owner, s.BatchID, []string{s.BookID}, map[string]textPublishOverride{s.BookID: {CollectionID: "other-collection", Force: true, RequestKey: "another-upload"}}); !errors.Is(err, batchfactoryv11.ErrConflict) {
		t.Fatalf("manual confirmation reused by automation on another collection: %v", err)
	}
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
	reply := call("GET", nil)
	if reply.LastPublished == nil || reply.LastPublished.RequestKey != "another-upload" || reply.LastPublished.RunID != "" {
		t.Fatalf("origin=%+v", reply.LastPublished)
	}
	reused := reply.Publish
	if reused.RunID != "next-run" || reused.Attempt != 2 || reused.RequestKey != "next-key" || reused.IntentID != second.Publish.IntentID || reused.LastPublished.RequestKey != "another-upload" || client.calls != 2 {
		t.Fatal(reused)
	}
	// Simulate an interrupted manual upload written before key ownership was
	// enforced. Reading it remains possible, but continuing it must not submit.
	legacy := *first.Publish
	legacy.IntentID, legacy.RunID, legacy.RequestKey = "old-dispatch", "", "foreign-0"
	legacy.Status, legacy.Phase = "uploading", "preparing"
	saveLegacy := func() {
		t.Helper()
		if _, err := store.SaveDraft(ctx, s.Owner, batchfactoryv11.Draft{Scope: s.BatchID, Key: "text-c", Kind: textPublishDraftKind, Content: string(mustJSON(legacy))}); err != nil {
			t.Fatal(err)
		}
	}
	saveLegacy()
	if got := call("GET", nil); got.Publish.IntentID != legacy.IntentID {
		t.Fatal(got)
	}
	if got := signedJSONRequest(t, api, now, s.Owner, "POST", path, map[string]any{}); got.Code != http.StatusConflict || client.calls != 2 {
		t.Fatalf("resumed another scope's stored confirmation: HTTP %d calls=%d", got.Code, client.calls)
	}
	legacy.Status = "pending"
	saveLegacy()
	beforeReads := client.readbackCalls
	if got := signedJSONRequest(t, api, now, s.Owner, "POST", path, map[string]any{}); got.Code != http.StatusConflict || client.readbackCalls != beforeReads || client.calls != 2 {
		t.Fatalf("continued pending foreign confirmation: HTTP %d reads=%d calls=%d", got.Code, client.readbackCalls, client.calls)
	}
	legacy.Status = "failed"
	saveLegacy()
	if got := call("POST", map[string]any{"force": true, "requestKey": "fresh-after-failure"}); got.Publish.RequestKey != "fresh-after-failure" || client.calls != 3 {
		t.Fatalf("fresh confirmed recovery: %+v calls=%d", got, client.calls)
	}
	foreign := signedJSONRequest(t, api, now, "bob", http.MethodGet, path, nil)
	if foreign.Code != 404 {
		t.Fatalf("foreign=%d", foreign.Code)
	}
}

func TestPublicDraftRoutesProtectCollectionReceipts(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Slice: 1, Store: store, Now: func() time.Time { return now }})
	ordinary := map[string]any{"key": "plain", "kind": "scratch", "scope": "batch", "content": "hello"}
	good := signedJSONRequest(t, router, now, "alice", "PUT", "/api/batch-factory/v12/drafts", ordinary)
	if good.Code != 200 {
		t.Fatalf("ordinary draft should work: %d %s", good.Code, good.Body.String())
	}
	good = signedJSONRequest(t, router, now, "alice", "GET", "/api/batch-factory/v12/drafts?key=plain&kind=scratch&scope=batch", nil)
	if good.Code != 200 || !bytes.Contains(good.Body.Bytes(), []byte("hello")) {
		t.Fatal(good.Code, good.Body.String())
	}
	for _, kind := range []string{n8ComicPublishDraftKind, n8ComicPublishDraftKind + "-history", textPublishDraftKind, textPublishDraftKind + "-history", "collection-publish-text-history"} {
		d := batchfactoryv11.Draft{Key: "book-c", Scope: "batch", Kind: kind, Content: `{"status":"pending","_receiptPath":"private-tag"}`}
		if _, err := store.SaveDraft(ctx, "alice", d); err != nil {
			t.Fatal(err)
		}
		for _, version := range []string{"v11", "v12"} {
			path := "/api/batch-factory/" + version + "/drafts"
			for _, k := range []string{kind, strings.ToUpper(kind)} {
				query := url.Values{"kind": {k}, "key": {d.Key}, "scope": {d.Scope}}
				response := signedJSONRequest(t, router, now, "alice", "GET", path+"?"+query.Encode(), nil)
				if response.Code < 400 || bytes.Contains(response.Body.Bytes(), []byte("private-tag")) {
					t.Fatalf("%s %s GET: %d", version, k, response.Code)
				}
				response = signedJSONRequest(t, router, now, "alice", "PUT", path, map[string]any{"kind": k, "scope": d.Scope, "key": d.Key, "content": `{"status":"published","carouselId":99}`})
				if response.Code < 400 {
					t.Fatalf("%s %s PUT: %d", version, k, response.Code)
				}
			}
			response := signedJSONRequest(t, router, now, "alice", "GET", path, nil)
			if bytes.Contains(response.Body.Bytes(), []byte("private-tag")) {
				t.Fatal("private field exposed")
			}
			response = signedJSONRequest(t, router, now, "alice", "PUT", path, map[string]any{"kind": "scratch", "scope": "s", "key": "k", "content": `{"_imageSignatures":["private"]}`})
			if response.Code < 400 {
				t.Fatal("accepted private receipt fields")
			}
		}
		got, err := store.GetDraft(ctx, "alice", d.Key, d.Kind, d.Scope)
		if err != nil || got.Content != d.Content {
			t.Fatal(got, err)
		}
	}
}
