package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

func libraryReplacementPNG(t *testing.T) []byte {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 3, 2))
	im.Set(1, 1, color.RGBA{R: 230, G: 90, A: 255})
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func libraryReplaceMultipartRequest(t *testing.T, mux http.Handler, now time.Time, actor, id string, expected int64, in screenplayLibraryUploadInput) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range map[string]string{"requestKey": in.RequestKey, "type": string(in.Type), "name": in.Name, "description": in.Description, "category": in.Category, "prompt": in.Prompt, "voice": in.Voice} {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	part, err := w.CreateFormFile("file", in.FileName)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(in.Bytes)
	w.Close()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/comic-screenplays/asset-library/%s/image?expectedRevision=%d", id, expected), &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	SignBridgeRequest(req, actor, false, now, "secret")
	result := httptest.NewRecorder()
	mux.ServeHTTP(result, req)
	return result
}

func TestScreenplayAssetLibraryReplaceImageHTTP(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	ctx := context.Background()
	in := libraryUploadFixture(t)
	entry, err := s.UploadEntry(ctx, "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	in.RequestKey = "replace-http-001"
	in.Bytes = libraryReplacementPNG(t)
	res := libraryReplaceMultipartRequest(t, mux, now, "alice", entry.ID, entry.Revision, in)
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	var payload struct {
		Entry struct {
			ID       string               `json:"id"`
			Revision int64                `json:"revision"`
			Image    screenplay.MediaItem `json:"image"`
		} `json:"entry"`
	}
	if err = json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Entry.ID != entry.ID || payload.Entry.Revision != entry.Revision+1 || payload.Entry.Image.URL == screenplayLibraryPublic(entry)["image"].(screenplay.MediaItem).URL {
		t.Fatal("replacement must refresh visible URL", res.Body.String())
	}
	image := signedJSONRequest(t, mux, now, "alice", "GET", payload.Entry.Image.URL, nil)
	if image.Code != 200 || !bytes.Equal(image.Body.Bytes(), in.Bytes) {
		t.Fatal("image bytes did not update", image.Code)
	}
	if res = libraryReplaceMultipartRequest(t, mux, now, "alice", entry.ID, entry.Revision, in); res.Code != 200 {
		t.Fatal("replay", res.Code)
	}
	if res = libraryReplaceMultipartRequest(t, mux, now, "bob", entry.ID, entry.Revision, in); res.Code != 404 {
		t.Fatal("owner isolation", res.Code)
	}
	if err = s.DeleteEntry(ctx, "alice", entry.ID, payload.Entry.Revision); err != nil {
		t.Fatal(err)
	}
	if res = libraryReplaceMultipartRequest(t, mux, now, "alice", entry.ID, entry.Revision, in); res.Code != 404 {
		t.Fatal("deleted image revived", res.Code)
	}
}

func TestScreenplayAssetLibraryReplaceImageLostCommitAndCancellation(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			s, _, objects, _, _ := assetLibraryServiceFixture(t)
			ctx := context.Background()
			in := libraryUploadFixture(t)
			entry, err := s.UploadEntry(ctx, "alice", in)
			if err != nil {
				t.Fatal(err)
			}
			repo := &commitLostAssetLibraryRepository{AssetLibraryRepository: s.Repository, lookupUnavailable: unknown}
			s.Repository = repo
			in.RequestKey = "replace-lost-commit"
			in.Bytes = libraryReplacementPNG(t)
			got, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
			if unknown && err == nil || !unknown && err != nil {
				t.Fatal(err)
			}
			repo.lookupUnavailable = false
			got, err = s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
			if err != nil {
				t.Fatal(err)
			}
			reader, err := objects.Open(ctx, libraryImageObject(got))
			if err != nil {
				t.Fatal("committed replacement was deleted", err)
			}
			reader.Close()
		})
	}
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	in := libraryUploadFixture(t)
	entry, err := s.UploadEntry(ctx, "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	in.RequestKey = "replace-cancel-after-put"
	in.Bytes = libraryReplacementPNG(t)
	cancelCtx, cancel := context.WithCancel(ctx)
	s.Media.Objects = &cancellingStageObjectStore{AssetObjectStore: objects, cancel: cancel}
	if _, err = s.ReplaceImage(cancelCtx, "alice", entry.ID, entry.Revision, in); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if current.Revision != entry.Revision || current.ImageObject != entry.ImageObject {
		t.Fatal("cancel changed original")
	}
}

func TestScreenplayAssetLibraryReplaceImageConcurrentSameKey(t *testing.T) {
	s, _, _, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	in := libraryUploadFixture(t)
	entry, err := s.UploadEntry(ctx, "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	in.RequestKey = "replace-concurrent"
	in.Bytes = libraryReplacementPNG(t)
	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if current.Revision != entry.Revision+1 {
		t.Fatal("replay incremented revision", current.Revision)
	}
}

func TestScreenplayAssetLibraryReplaceImageKeepsExistingProjectCopyAndStagesNewImage(t *testing.T) {
	s, entry, target, _, _ := stageFixture(t)
	ctx := context.Background()
	first, err := s.StageImage(ctx, "alice", target.ProjectID, entry.ID, stageInput(entry, target))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := s.Media.Cut.Asset(ctx, "alice", target.ProjectID, first.AssetID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := s.Media.Objects.Open(ctx, aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	before, _ := io.ReadAll(reader)
	reader.Close()
	in := libraryUploadFixture(t)
	in.RequestKey = "replace-after-project-copy"
	in.Bytes = libraryReplacementPNG(t)
	replaced, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil {
		t.Fatal(err)
	}
	reader, err = s.Media.Objects.Open(ctx, aicutObjectAsset(stored))
	if err != nil {
		t.Fatal("replacement deleted project copy", err)
	}
	after, _ := io.ReadAll(reader)
	reader.Close()
	if !bytes.Equal(before, after) {
		t.Fatal("existing project image changed")
	}
	input := stageInput(replaced, target)
	input.RequestKey = "stage-replacement-new-key"
	second, err := s.StageImage(ctx, "alice", target.ProjectID, entry.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	stored, err = s.Media.Cut.Asset(ctx, "alice", target.ProjectID, second.AssetID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err = s.Media.Objects.Open(ctx, aicutObjectAsset(stored))
	if err != nil {
		t.Fatal(err)
	}
	after, _ = io.ReadAll(reader)
	reader.Close()
	if !bytes.Equal(after, in.Bytes) {
		t.Fatal("new binding did not use replacement image")
	}
}

func TestScreenplayAssetLibraryReplaceImagePreservesIdentityReplayAndMetadata(t *testing.T) {
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	ctx := context.Background()
	entry, err := s.UploadEntry(ctx, "alice", libraryUploadFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	entry.Snapshot.Extra["private_extension"] = "preserve"
	entry, err = s.Repository.SaveAssetLibraryEntry(ctx, entry, entry.Revision)
	if err != nil {
		t.Fatal(err)
	}
	in := libraryUploadFixture(t)
	in.RequestKey = "replace-image-001"
	in.Bytes = libraryReplacementPNG(t)
	in.Name = "换图后名称"
	in.Description = "换图后描述"
	got, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != entry.ID || got.Revision != entry.Revision+1 || got.ImageObject.ObjectKey == entry.ImageObject.ObjectKey || got.Snapshot.Extra["private_extension"] != "preserve" || got.Snapshot.Prompt != entry.Snapshot.Prompt || got.Name != in.Name {
		t.Fatal(got)
	}
	again, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil || again.Revision != got.Revision || again.ImageObject != got.ImageObject {
		t.Fatal(again, err)
	}
	read, err := objects.Open(ctx, libraryImageObject(got))
	if err != nil {
		t.Fatal(err)
	}
	read.Close()
	changed := in
	changed.Description = "same key different input"
	if _, err = s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, changed); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	in.RequestKey = "replace-image-002"
	if _, err = s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	if _, err = s.ReplaceImage(ctx, "bob", entry.ID, got.Revision, in); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal(err)
	}
	in.Bytes = []byte("invalid")
	if _, err = s.ReplaceImage(ctx, "alice", entry.ID, got.Revision, in); !errors.Is(err, screenplay.ErrInvalidArgument) {
		t.Fatal(err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.ReplaceImage(cancelCtx, "alice", entry.ID, got.Revision, in); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if current.Revision != got.Revision || current.ImageObject != got.ImageObject {
		t.Fatal("failed/cancelled replacement changed saved image")
	}
}
