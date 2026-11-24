package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"testing"
	"time"
)

type textDownloadRepo struct {
	txt.Repository
	scope      txt.Scope
	collection txt.Collection
	artifacts  []txt.PageArtifact
}

func (r *textDownloadRepo) GetCollection(_ context.Context, s txt.Scope, id string) (txt.Collection, error) {
	if s != r.scope || id != r.collection.ID {
		return txt.Collection{}, txt.ErrNotFound
	}
	return r.collection, nil
}
func (r *textDownloadRepo) ListArtifacts(_ context.Context, s txt.Scope) ([]txt.PageArtifact, error) {
	if s != r.scope {
		return nil, txt.ErrNotFound
	}
	return r.artifacts, nil
}

func TestBatchFactoryTextCollectionDownloadUsesFrozenOrderedImages(t *testing.T) {
	now := time.Unix(1700000000, 0)
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "批次", InitialBatchSettings: batchfactoryv11.SettingsPatch{"productionMode": []byte(`"text"`)}, Books: []batchfactoryv11.CreateBookInput{{Title: "测试书", SourceText: "完整原文。"}}})
	if err != nil {
		t.Fatal(err)
	}
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	repo := &textDownloadRepo{scope: txt.Scope{Owner: "alice", BatchID: batch.ID, BookID: batch.Books[0].ID}, collection: txt.Collection{ID: "collection", Status: "ready", CoverArtifactID: "cover"}}
	originals := [][]byte{}
	for i := 0; i < 2; i++ {
		var data bytes.Buffer
		img := image.NewRGBA(image.Rect(0, 0, 20, 30))
		img.Set(1, 1, color.RGBA{R: uint8(100 + i*100), A: 255})
		_ = png.Encode(&data, img)
		originals = append(originals, append([]byte(nil), data.Bytes()...))
		file, err := files.SaveImage([]string{"page_one", "page_two"}[i], "image/png", bytes.NewReader(data.Bytes()))
		if err != nil {
			t.Fatal(err)
		}
		id := []string{"one", "two"}[i]
		repo.artifacts = append(repo.artifacts, txt.PageArtifact{ID: id, PageID: id, Kind: "page", ArtifactFile: txt.ArtifactFile{StorageRef: file.StorageRef, MediaType: file.MediaType, ByteSize: file.ByteSize, SHA256: file.SHA256}})
		repo.collection.Items = append(repo.collection.Items, txt.CollectionItem{Ordinal: i + 1, PageID: id, ArtifactID: id, StorageRef: file.StorageRef, MediaType: file.MediaType, SHA256: file.SHA256})
	}
	// List order and newer artifacts must not replace the frozen collection.
	repo.collection.Items[0], repo.collection.Items[1] = repo.collection.Items[1], repo.collection.Items[0]
	repo.artifacts = append(repo.artifacts, txt.PageArtifact{ID: "newer", PageID: "one", Kind: "page"}, txt.PageArtifact{ID: "cover", Kind: "cover"})
	router := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, Slice: 1, Store: store, TextRepository: repo, LocalArtifacts: files})
	route := "/api/batch-factory/v12/batches/" + batch.ID + "/books/" + batch.Books[0].ID + "/text-collections/collection/download"
	result := signedJSONRequest(t, router, now, "alice", http.MethodGet, route, nil)
	if result.Code != 200 {
		t.Fatalf("%d %s", result.Code, result.Body.String())
	}
	_, params, err := mime.ParseMediaType(result.Header().Get("Content-Disposition"))
	if err != nil || params["filename"] != "测试书.zip" {
		t.Fatal(params, err)
	}
	archive, err := zip.NewReader(bytes.NewReader(result.Body.Bytes()), int64(result.Body.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(archive.File) != 2 {
		t.Fatal("cover included or page missing", len(archive.File))
	}
	for i, file := range archive.File {
		if file.Name != []string{"测试书-001.png", "测试书-002.png"}[i] {
			t.Fatal(file.Name)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil || !bytes.Equal(data, originals[i]) {
			t.Fatal("download bytes differ from frozen page", err)
		}
	}
	denied := signedJSONRequest(t, router, now, "bob", http.MethodGet, route, nil)
	if denied.Code == 200 {
		t.Fatal("other account downloaded collection")
	}
	if err = os.Remove(filepath.Join(files.Root(), repo.artifacts[0].StorageRef)); err != nil {
		t.Fatal(err)
	}
	missing := signedJSONRequest(t, router, now, "alice", http.MethodGet, route, nil)
	if missing.Code != 409 || missing.Header().Get("Content-Type") == "application/zip" {
		t.Fatalf("missing file returned successful archive: %d", missing.Code)
	}
}
