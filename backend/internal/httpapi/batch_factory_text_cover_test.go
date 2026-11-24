package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"strings"
	"testing"
	"time"
)

type textCoverRepo struct {
	txt.Repository
	document  txt.Document
	artifacts []txt.PageArtifact
	reads     int
}

func (r *textCoverRepo) GetDocument(context.Context, txt.Scope) (txt.Document, error) {
	r.reads++
	if r.document.ID == "" {
		return txt.Document{}, txt.ErrNotFound
	}
	return r.document, nil
}
func (r *textCoverRepo) ListArtifacts(context.Context, txt.Scope) ([]txt.PageArtifact, error) {
	return r.artifacts, nil
}

func TestBatchFactoryTextDefaultCoverWorksBeforeProductionAndRetainsUploads(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "批次标题", InitialBatchSettings: batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"text"`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "书名作为封面标题", SourceText: "正文不应出现在封面。"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	repo := &textCoverRepo{}
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, TextRepository: repo, LocalArtifacts: files})
	route := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/text-cover/display"
	cover := signedJSONRequest(t, router, now, "alice", http.MethodGet, route, nil)
	if cover.Code != 200 {
		t.Fatalf("%d %s", cover.Code, cover.Body.String())
	}
	if _, err = png.Decode(bytes.NewReader(cover.Body.Bytes())); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(cover.Body.Bytes(), txt.DefaultCoverImage()) {
		t.Fatal("cover did not render book title")
	}
	if repo.reads != 1 {
		t.Fatal("cover read unexpectedly prepared production")
	}
	summary := signedJSONRequest(t, router, now, "alice", http.MethodGet, "/api/batch-factory/v12/batches/summary", nil)
	if summary.Code != 200 || !strings.Contains(summary.Body.String(), "text-cover/display") || !strings.Contains(summary.Body.String(), `"productionMode":"text"`) {
		t.Fatal("text project summary has no default cover", summary.Body.String())
	}
	denied := signedJSONRequest(t, router, now, "bob", http.MethodGet, route, nil)
	if denied.Code == 200 {
		t.Fatal("other account accessed cover")
	}
	var upload bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 30, 50))
	img.Set(1, 1, color.White)
	_ = png.Encode(&upload, img)
	file, err := files.SaveImage("locked_cover", "image/png", bytes.NewReader(upload.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	repo.document = txt.Document{ID: "document", CoverLocked: true, CoverArtifactID: "manual"}
	repo.artifacts = []txt.PageArtifact{{ID: "manual", Kind: "cover", ArtifactFile: txt.ArtifactFile{StorageRef: file.StorageRef, MediaType: file.MediaType}}}
	locked := signedJSONRequest(t, router, now, "alice", http.MethodGet, route, nil)
	if locked.Code != 200 || !bytes.Equal(locked.Body.Bytes(), upload.Bytes()) {
		t.Fatal("manual cover was overwritten")
	}
}
