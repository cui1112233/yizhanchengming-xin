package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/screenplay"
)

func libraryAudioMultipart(t *testing.T, mux http.Handler, now time.Time, method, url string, in screenplayLibraryUploadInput) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for k, v := range map[string]string{"requestKey": in.RequestKey, "type": string(in.Type), "name": in.Name, "description": in.Description, "category": in.Category, "prompt": in.Prompt, "voice": in.Voice} {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []struct {
		field, name string
		raw         []byte
	}{{"file", in.FileName, in.Bytes}, {"audio", in.AudioName, in.AudioBytes}} {
		if file.raw == nil {
			continue
		}
		part, err := w.CreateFormFile(file.field, file.name)
		if err != nil {
			t.Fatal(err)
		}
		part.Write(file.raw)
	}
	w.Close()
	req := httptest.NewRequest(method, url, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	SignBridgeRequest(req, "alice", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	return res
}

func TestScreenplayAssetLibraryReferenceAudioHTTPAndOwnerIsolation(t *testing.T) {
	s, _, objects, mux, now := assetLibraryServiceFixture(t)
	in := libraryUploadFixture(t)
	in.AudioBytes = libraryReferenceWAV()
	in.AudioName = "初始.wav"
	res := libraryAudioMultipart(t, mux, now, "POST", "/api/comic-screenplays/asset-library", in)
	if res.Code != 201 {
		t.Fatal(res.Code, res.Body.String())
	}
	var result struct {
		Entry struct {
			ID       string               `json:"id"`
			Revision int64                `json:"revision"`
			Audio    string               `json:"audio"`
			Name     string               `json:"audioFileName"`
			Image    screenplay.MediaItem `json:"image"`
		} `json:"entry"`
	}
	json.Unmarshal(res.Body.Bytes(), &result)
	got := signedJSONRequest(t, mux, now, "alice", "GET", result.Entry.Audio, nil)
	if got.Code != 200 || !bytes.Equal(got.Body.Bytes(), in.AudioBytes) || got.Header().Get("Content-Type") != "audio/wav" {
		t.Fatal("saved audio not served", got.Code)
	}
	if got = signedJSONRequest(t, mux, now, "bob", "GET", result.Entry.Audio, nil); got.Code != 404 {
		t.Fatal("cross-user audio", got.Code)
	}
	oldImage := result.Entry.Image.AssetID
	id := result.Entry.ID
	in.Bytes = nil
	in.AudioBytes[48] = 9
	in.AudioName = "替换.wav"
	in.RequestKey = "reference-http-update"
	route := "/api/comic-screenplays/asset-library/" + id + "/audio?expectedRevision=1"
	res = libraryAudioMultipart(t, mux, now, "PUT", route, in)
	if res.Code != 200 {
		t.Fatal(res.Code, res.Body.String())
	}
	json.Unmarshal(res.Body.Bytes(), &result)
	if result.Entry.Revision != 2 || result.Entry.Name != "替换.wav" || result.Entry.Image.AssetID != oldImage {
		t.Fatal(result)
	}
	if res = libraryAudioMultipart(t, mux, now, "PUT", route, in); res.Code != 200 {
		t.Fatal("replay", res.Code)
	}
	in.RequestKey = "reference-http-stale"
	if res = libraryAudioMultipart(t, mux, now, "PUT", route, in); res.Code != 409 {
		t.Fatal("stale revision", res.Code)
	}
	entry, _ := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", id)
	audio, _ := libraryAudioObject(entry)
	if err := s.DeleteEntry(context.Background(), "alice", id, 2); err != nil {
		t.Fatal(err)
	}
	if got = signedJSONRequest(t, mux, now, "alice", "GET", result.Entry.Audio, nil); got.Code != 404 {
		t.Fatal("deleted audio served", got.Code)
	}
	objects.mu.Lock()
	_, exists := objects.files[audio.ObjectKey]
	objects.mu.Unlock()
	if exists {
		t.Fatal("audio object not removed with deleted library entry")
	}
}

func TestScreenplayAssetLibraryReferenceAudioLostCommitConcurrentAndCancellation(t *testing.T) {
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
			in.Bytes = nil
			in.AudioBytes = libraryReferenceWAV()
			in.AudioName = "重试.wav"
			in.RequestKey = "reference-lost-commit"
			_, err = s.ReplaceReferenceAudio(ctx, "alice", entry.ID, entry.Revision, in)
			if unknown && err == nil || !unknown && err != nil {
				t.Fatal(err)
			}
			repo.lookupUnavailable = false
			saved, err := s.ReplaceReferenceAudio(ctx, "alice", entry.ID, entry.Revision, in)
			if err != nil {
				t.Fatal(err)
			}
			audio, err := libraryAudioObject(saved)
			if err != nil {
				t.Fatal(err)
			}
			reader, err := objects.Open(ctx, audio)
			if err != nil {
				t.Fatal("committed audio deleted", err)
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
	in.Bytes = nil
	in.AudioBytes = libraryReferenceWAV()
	in.AudioName = "并发.wav"
	in.RequestKey = "reference-concurrent"
	var wg sync.WaitGroup
	errs := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.ReplaceReferenceAudio(ctx, "alice", entry.ID, entry.Revision, in)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if current.Revision != 2 {
		t.Fatal(current.Revision)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	s.Media.Objects = &cancellingStageObjectStore{AssetObjectStore: objects, cancel: cancel}
	in.RequestKey = "reference-cancel"
	in.AudioBytes[48] = 3
	if _, err = s.ReplaceReferenceAudio(cancelCtx, "alice", entry.ID, current.Revision, in); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	same, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", entry.ID)
	if same.Revision != 2 || same.Snapshot.Extra[libraryAudioKey] != current.Snapshot.Extra[libraryAudioKey] {
		t.Fatal("cancel changed audio")
	}
}

func TestScreenplayAssetLibraryReferenceAudioCannotReferenceAnotherOwnerObject(t *testing.T) {
	in := libraryUploadFixture(t)
	in.AudioBytes = libraryReferenceWAV()
	in.AudioName = "声音.wav"
	s, _, _, _, _ := assetLibraryServiceFixture(t)
	audio, err := s.prepareLibraryAudio("bob", in)
	if err != nil {
		t.Fatal(err)
	}
	entry := screenplay.AssetLibraryEntry{Owner: "alice", Type: screenplay.AssetCharacter, Snapshot: screenplay.Asset{Extra: map[string]string{}}}
	setLibraryAudio(&entry, audio)
	if _, err = libraryAudioObject(entry); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal("foreign object accepted", err)
	}
}
