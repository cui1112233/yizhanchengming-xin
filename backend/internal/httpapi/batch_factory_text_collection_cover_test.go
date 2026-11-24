package httpapi

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"testing"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/localartifact"
)

type textCollectionCoverRepo struct {
	txt.Repository
	doc       txt.Document
	pages     []txt.Page
	artifacts []txt.PageArtifact
}

func (r *textCollectionCoverRepo) GetDocument(context.Context, txt.Scope) (txt.Document, error) {
	return r.doc, nil
}
func (r *textCollectionCoverRepo) ListPages(context.Context, txt.Scope) ([]txt.Page, error) {
	return r.pages, nil
}
func (r *textCollectionCoverRepo) ListArtifacts(context.Context, txt.Scope) ([]txt.PageArtifact, error) {
	return r.artifacts, nil
}

func TestBatchFactoryTextCollectionValidatesBodyWithoutDisplayCover(t *testing.T) {
	files := localartifact.NewStore(t.TempDir(), 1<<20)
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	file, err := files.SaveImage("body", "image/png", bytes.NewReader(data.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	repo := &textCollectionCoverRepo{doc: txt.Document{ID: "d", ActiveVersionID: "v", PaginationVersion: 1, Revision: 2}, pages: []txt.Page{{ID: "p", Ordinal: 1, ActiveArtifactID: "body"}}, artifacts: []txt.PageArtifact{{ID: "body", PageID: "p", Kind: "page", ArtifactFile: txt.ArtifactFile{StorageRef: file.StorageRef, MediaType: file.MediaType, ByteSize: file.ByteSize, SHA256: file.SHA256}}}}
	api := textAPI{repo: repo, files: files}
	request := textRequest{Context: context.Background(), scope: txt.Scope{Owner: "owner", BatchID: "batch", BookID: "book"}}
	input := textInput{VersionID: "v", PaginationVersion: 1, ExpectedRevision: 2}
	for _, cover := range []string{"", "missing-display-cover"} {
		repo.doc.CoverArtifactID = cover
		if _, err := api.verifyCollectionFiles(request, input); err != nil {
			t.Fatalf("display cover %q blocked body collection: %v", cover, err)
		}
	}
	repo.artifacts[0].SHA256 = "corrupt"
	if _, err := api.verifyCollectionFiles(request, input); !errors.Is(err, txt.ErrNotReady) {
		t.Fatalf("corrupt body was accepted: %v", err)
	}
}
