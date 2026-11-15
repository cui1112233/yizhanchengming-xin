package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/characterarchive"
	"qiantie/backend/internal/screenplay"
)

func bulkHTTPRequest(t *testing.T, mux http.Handler, now time.Time, actor, key string, files map[string][]byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("requestKey", key)
	for name, raw := range files {
		part, err := writer.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write(raw)
	}
	writer.Close()
	r := httptest.NewRequest("POST", "/api/comic-screenplays/asset-library/characters/import", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(r, actor, false, now, "secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}
func bulkJSONRequest(t *testing.T, mux http.Handler, now time.Time, actor, method, url string, input any) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(input)
	r := httptest.NewRequest(method, url, bytes.NewReader(raw))
	SignBridgeRequest(r, actor, false, now, "secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	return w
}
func TestLibraryBulkHTTPNativeAppendReplayMetadataAndExport(t *testing.T) {
	s, asset, _, mux, now := assetLibraryServiceFixture(t)
	ctx := context.Background()
	native, err := os.ReadFile("../characterarchive/testdata/native.lschar")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"folder/characters.lschar": native, "folder/characters.json": []byte(`bad fallback`)}
	for _, key := range []string{"native-import-one", "native-import-one", "native-import-two"} {
		w := bulkHTTPRequest(t, mux, now, "alice", key, files)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	entries, _ := s.Repository.ListAssetLibrary(ctx, "alice", screenplay.AssetCharacter, "", "")
	if len(entries) != 2 {
		t.Fatal(len(entries))
	}
	if entries[0].Snapshot.Extra[screenplay.LibraryArchiveID] == entries[1].Snapshot.Extra[screenplay.LibraryArchiveID] {
		t.Fatal("ID collision")
	}
	bad := bulkHTTPRequest(t, mux, now, "alice", "native-import-one", map[string][]byte{"characters.json": []byte(`[{"name":"changed"}]`)})
	if bad.Code != 409 {
		t.Fatal(bad.Code, bad.Body.String())
	}
	metadata := bulkHTTPRequest(t, mux, now, "alice", "metadata-import-one", map[string][]byte{"characters.json": []byte(`[{"id":"missing","name":"无图角色","custom":{"nested":[1,true]},"voice":"voice"}]`)})
	if metadata.Code != 200 || !strings.Contains(metadata.Body.String(), `"image_available":false`) {
		t.Fatal(metadata.Code, metadata.Body.String())
	}
	entries, _ = s.Repository.ListAssetLibrary(ctx, "alice", screenplay.AssetCharacter, "", "")
	var missing screenplay.AssetLibraryEntry
	for _, e := range entries {
		if e.Name == "无图角色" {
			missing = e
		}
	}
	stage := stageInput(missing, asset)
	_, err = s.StageImage(ctx, "alice", asset.ProjectID, missing.ID, stage)
	if err == nil || !strings.Contains(err.Error(), "先补图") {
		t.Fatal(err)
	}
	title := "已编辑角色"
	_, err = s.UpdateEntry(ctx, "alice", missing.ID, screenplayLibraryMetadataInput{Expected: &missing.Revision, Name: &title})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := s.ExportCharacters(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := characterarchive.DecodeJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(decoded, []byte(`"nested":[1,true]`)) || !bytes.Contains(decoded, []byte(title)) || bytes.Contains(decoded, []byte("object_key")) || bytes.Contains(decoded, []byte("screenplay-library")) {
		t.Fatal(string(decoded))
	}
	exported := bulkJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/asset-library/characters/export?q=no-match", nil)
	if exported.Code != 200 || !strings.Contains(exported.Header().Get("Content-Disposition"), "characters.lschar") {
		t.Fatal(exported.Code, exported.Body.String())
	}
	bob := bulkJSONRequest(t, mux, now, "bob", "GET", "/api/comic-screenplays/asset-library/characters/export", nil)
	if bob.Code == 200 {
		t.Fatal("owner isolation")
	}
	upload := libraryUploadFixture(t)
	upload.RequestKey = "metadata-supply-image"
	upload.Name = title
	current, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", missing.ID)
	supplied, err := s.ReplaceImage(ctx, "alice", missing.ID, current.Revision, upload)
	if err != nil || supplied.Snapshot.Extra[screenplay.LibraryMetadataOnly] != "" {
		t.Fatal(err, supplied)
	}
}
func TestLibraryBulkHTTPRejectsBeforeAnyWriteAndNoFallback(t *testing.T) {
	s, _, objects, mux, now := assetLibraryServiceFixture(t)
	baseline := len(objects.files)
	cases := []map[string][]byte{
		{"characters.lschar": []byte("broken"), "characters.json": []byte(`[{"name":"fallback"}]`)},
		{"characters.json": []byte(`[]`)}, {"characters.json": []byte(`{"characters":[]}`)}, {"characters.json": []byte(`[null]`)},
		{"characters.json": []byte(`[{"name":"ok"},{"name":"bad","imagePath":"missing.png"}]`)},
		{"characters.json": []byte(`[{"name":"bad","imagePath":"../escape.png"}]`)},
		{"characters.json": []byte(`[{"name":"bad","image":"https://example.invalid/x.png"}]`)},
		{"characters.json": []byte(`[{"name":"bad","audio":"C:\\private\\voice.wav"}]`)},
		{"a/characters.json": []byte(`[{"name":"x"}]`), "b/characters.json": []byte(`[{"name":"y"}]`)},
		{"../characters.json": []byte(`[{"name":"x"}]`)},
	}
	for i, files := range cases {
		w := bulkHTTPRequest(t, mux, now, "alice", fmt.Sprintf("bad-case-%08d", i), files)
		if w.Code == 200 {
			t.Fatal(i, w.Body.String())
		}
		entries, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
		if len(entries) != 0 || len(objects.files) != baseline {
			t.Fatal("partial import", i)
		}
	}
}
func TestLibraryBulkFolderMediaPortableAndClearIsolation(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	ctx := context.Background()
	png := testImageToolPNG(t)
	audio := append([]byte("RIFF"), make([]byte, 4)...)
	audio = append(audio, []byte("WAVEfmt ")...)
	audio = append(audio, make([]byte, 40)...)
	files := map[string][]byte{"folder/characters.json": []byte(`[{"id":"a","name":"带资源","imagePath":"image.png","audio":"voice.wav","audioFileName":"voice.wav","unknown":{"v":42}}]`), "folder/image.png": png, "folder/voice.wav": audio}
	w := bulkHTTPRequest(t, mux, now, "alice", "folder-media-one", files)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	raw, err := s.ExportCharacters(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := characterarchive.DecodeJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(decoded, []byte(base64.StdEncoding.EncodeToString(png))) || !bytes.Contains(decoded, []byte(base64.StdEncoding.EncodeToString(audio))) {
		t.Fatal("portable bytes missing")
	}
	entries, _ := s.Repository.ListAssetLibrary(ctx, "alice", screenplay.AssetCharacter, "", "")
	expected := []screenplay.AssetLibraryExpected{{ID: entries[0].ID, Revision: entries[0].Revision}}
	source := libraryUploadFixture(t)
	source.Type = screenplay.AssetScene
	source.RequestKey = "scene-preserved-one"
	scene, err := s.UploadEntry(ctx, "alice", source)
	if err != nil {
		t.Fatal(err)
	}
	w = bulkJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/asset-library/characters/clear", libraryClearInput{RequestKey: "stale-clear-one", Expected: []screenplay.AssetLibraryExpected{}})
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	for range 2 {
		w = bulkJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/asset-library/characters/clear", libraryClearInput{RequestKey: "clear-full-one", Expected: expected})
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"cleared_count":1`) {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if _, err = s.Repository.GetAssetLibraryEntry(ctx, "alice", scene.ID); err != nil {
		t.Fatal("scene changed", err)
	}
	if _, err = s.Repository.GetAssetLibraryEntry(ctx, "alice", entries[0].ID); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal(err)
	}
	w = bulkHTTPRequest(t, mux, now, "alice", "folder-media-new", map[string][]byte{"characters.lschar": raw})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestLibraryBulkConcurrentReplayAndFailedBatch(t *testing.T) {
	s, _, _, _, _ := assetLibraryServiceFixture(t)
	in := libraryBulkFiles{RequestKey: "concurrent-bulk", Archive: "characters.json", Files: map[string][]byte{"characters.json": []byte(`[{"id":"a","name":"并发角色"}]`)}}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.ImportCharacters(context.Background(), "alice", in); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
	if len(entries) != 1 {
		t.Fatal(len(entries))
	}
}

type bulkCommitLostRepository struct {
	*screenplay.MemoryAssetLibraryRepository
	lost bool
}

func (r *bulkCommitLostRepository) ApplyAssetLibraryBulk(ctx context.Context, owner, key, digest, action string, entries []screenplay.AssetLibraryEntry, expected []screenplay.AssetLibraryExpected) (screenplay.AssetLibraryBulkResult, error) {
	result, err := r.MemoryAssetLibraryRepository.ApplyAssetLibraryBulk(ctx, owner, key, digest, action, entries, expected)
	if err == nil && !r.lost {
		r.lost = true
		return screenplay.AssetLibraryBulkResult{}, errors.New("commit acknowledgement lost")
	}
	return result, err
}
func TestLibraryBulkUnknownCommitRetainsMediaAndSameKeyRecovers(t *testing.T) {
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	s.Repository = &bulkCommitLostRepository{MemoryAssetLibraryRepository: screenplay.NewMemoryAssetLibraryRepository()}
	record := `[{"id":"one","name":"recover","image":"data:image/png;base64,` + base64.StdEncoding.EncodeToString(testImageToolPNG(t)) + `"}]`
	in := libraryBulkFiles{RequestKey: "lost-commit-bulk", Archive: "characters.json", Files: map[string][]byte{"characters.json": []byte(record)}}
	baseline := len(objects.files)
	if _, err := s.ImportCharacters(context.Background(), "alice", in); err == nil {
		t.Fatal("expected unknown acknowledgement")
	}
	if len(objects.files) != baseline+1 {
		t.Fatal("committed image missing")
	}
	result, err := s.ImportCharacters(context.Background(), "alice", in)
	if err != nil || len(result.Entries) != 1 || len(objects.files) != baseline+1 {
		t.Fatal(result, err, len(objects.files))
	}
	if _, err = s.ExportCharacters(context.Background(), "alice"); err != nil {
		t.Fatal("recovered image unreadable", err)
	}
}

func TestLibraryBulkExportCountLimitMatchesReimport(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	ctx := context.Background()
	records := make([]map[string]string, 10000)
	for i := range records {
		records[i] = map[string]string{"id": fmt.Sprintf("original-%d", i), "name": fmt.Sprintf("角色%d", i)}
	}
	raw, _ := json.Marshal(records)
	first, err := s.ImportCharacters(ctx, "alice", libraryBulkFiles{RequestKey: "count-limit-first", Archive: "characters.json", Files: map[string][]byte{"characters.json": raw}})
	if err != nil || len(first.Entries) != 10000 {
		t.Fatalf("first legal import: %d %v", len(first.Entries), err)
	}
	exported, err := s.ExportCharacters(ctx, "alice")
	if err != nil {
		t.Fatal("10000 must export", err)
	}
	parsed, err := (libraryBulkFiles{Archive: "characters.lschar", Files: map[string][]byte{"characters.lschar": exported}}).records()
	if err != nil || len(parsed) != 10000 {
		t.Fatalf("boundary export must be reimportable: %d %v", len(parsed), err)
	}
	_, err = s.ImportCharacters(ctx, "alice", libraryBulkFiles{RequestKey: "count-limit-second", Archive: "characters.json", Files: map[string][]byte{"characters.json": []byte(`[{"id":"extra","name":"第10001个角色"}]`)}})
	if err != nil {
		t.Fatal("second legal import", err)
	}
	response := bulkJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/asset-library/characters/export", nil)
	if response.Code != 400 || !strings.Contains(response.Body.String(), "10000") || response.Header().Get("Content-Disposition") != "" {
		t.Fatalf("oversized library must return an explicit limit error instead of an unusable archive: status=%d disposition=%q", response.Code, response.Header().Get("Content-Disposition"))
	}
	rows, err := s.Repository.ListAssetLibrary(ctx, "alice", screenplay.AssetCharacter, "", "")
	if err != nil || len(rows) != 10001 {
		t.Fatalf("export refusal must retain complete library: %d %v", len(rows), err)
	}
}
