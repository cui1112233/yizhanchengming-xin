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
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
	"strings"
	"sync"
	"testing"
	"time"
)

func libraryUploadFixture(t *testing.T) screenplayLibraryUploadInput {
	return screenplayLibraryUploadInput{RequestKey: "manager-upload-001", Type: screenplay.AssetCharacter, Name: "新角色", Description: "描述", Category: "古言", Prompt: "形象提示词", Voice: "声音描述", FileName: "角色.png", Bytes: testImageToolPNG(t)}
}
func libraryMultipartRequest(t *testing.T, mux http.Handler, now time.Time, actor string, input screenplayLibraryUploadInput, extra map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{"requestKey": input.RequestKey, "type": string(input.Type), "name": input.Name, "description": input.Description, "category": input.Category, "prompt": input.Prompt, "voice": input.Voice}
	for k, v := range extra {
		fields[k] = v
	}
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	part, err := writer.CreateFormFile("file", input.FileName)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(input.Bytes)
	writer.Close()
	request := httptest.NewRequest("POST", "/api/comic-screenplays/asset-library", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(request, actor, false, now, "secret")
	result := httptest.NewRecorder()
	mux.ServeHTTP(result, request)
	return result
}
func TestScreenplayAssetLibraryManagerUploadEditDeleteHTTP(t *testing.T) {
	s, _, objects, mux, now := assetLibraryServiceFixture(t)
	input := libraryUploadFixture(t)
	response := libraryMultipartRequest(t, mux, now, "alice", input, nil)
	if response.Code != 201 {
		t.Fatal(response.Code, response.Body.String())
	}
	var result struct {
		Entry struct {
			ID       string               `json:"id"`
			Revision int64                `json:"revision"`
			Prompt   string               `json:"prompt"`
			Voice    string               `json:"voice"`
			Image    screenplay.MediaItem `json:"image"`
		} `json:"entry"`
	}
	json.Unmarshal(response.Body.Bytes(), &result)
	if result.Entry.ID == "" || result.Entry.Prompt != input.Prompt || result.Entry.Voice != input.Voice || result.Entry.Image.URL == "" || strings.Contains(response.Body.String(), "image_object") || strings.Contains(response.Body.String(), "library_upload_digest") {
		t.Fatal(response.Body.String())
	}
	get := signedJSONRequest(t, mux, now, "alice", "GET", result.Entry.Image.URL, nil)
	if get.Code != 200 || !bytes.Equal(get.Body.Bytes(), input.Bytes) {
		t.Fatal(get.Code)
	}
	patch := map[string]any{"expectedRevision": result.Entry.Revision, "name": "改名", "description": "新描述", "category": "新分类", "prompt": "新提示词", "voice": "新声音"}
	path := "/api/comic-screenplays/asset-library/" + result.Entry.ID
	edited := signedJSONRequest(t, mux, now, "alice", "PATCH", path, patch)
	if edited.Code != 200 {
		t.Fatal(edited.Code, edited.Body.String())
	}
	repeated := signedJSONRequest(t, mux, now, "alice", "PATCH", path, patch)
	if repeated.Code != 200 {
		t.Fatal(repeated.Code, repeated.Body.String())
	}
	entry, _ := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", result.Entry.ID)
	if entry.Name != "改名" || entry.Revision != 2 || entry.Snapshot.Extra[screenplayLibraryPromptKey] != "新提示词" || entry.Snapshot.Extra[screenplayLibraryVoiceKey] != "新声音" || entry.Snapshot.Prompt != input.Prompt || entry.Snapshot.Voice != input.Voice {
		t.Fatal(entry)
	}
	again := libraryMultipartRequest(t, mux, now, "alice", input, nil)
	if again.Code != 201 || len(objects.files) != 2 {
		t.Fatal("upload replay after edit did not recover original identity", again.Code, len(objects.files))
	}
	foreign := signedJSONRequest(t, mux, now, "bob", "PATCH", path, patch)
	if foreign.Code != 404 {
		t.Fatal(foreign.Code)
	}
	foreignDelete := signedJSONRequest(t, mux, now, "bob", "DELETE", path+"?expectedRevision=2", nil)
	if foreignDelete.Code != 200 {
		t.Fatal(foreignDelete.Code)
	}
	if _, err := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", result.Entry.ID); err != nil || len(objects.files) != 2 {
		t.Fatal("foreign deletion touched owner's data", err, len(objects.files))
	}
	stale := signedJSONRequest(t, mux, now, "alice", "DELETE", path+"?expectedRevision=1", nil)
	if stale.Code != 409 {
		t.Fatal(stale.Code)
	}
	deleted := signedJSONRequest(t, mux, now, "alice", "DELETE", path+"?expectedRevision=2", nil)
	if deleted.Code != 200 || len(objects.files) != 1 {
		t.Fatal(deleted.Code, deleted.Body.String(), len(objects.files))
	}
	deleted = signedJSONRequest(t, mux, now, "alice", "DELETE", path+"?expectedRevision=2", nil)
	if deleted.Code != 200 {
		t.Fatal("delete replay", deleted.Code)
	}
	if _, err := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", result.Entry.ID); !errors.Is(err, screenplay.ErrNotFound) {
		t.Fatal(err)
	}
}
func TestScreenplayAssetLibraryManagerUploadValidationAndFileNames(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	for _, kind := range []screenplay.AssetType{screenplay.AssetScene, screenplay.AssetProp} {
		input := libraryUploadFixture(t)
		input.Type = kind
		input.Name = ""
		input.FileName = "房间-原名.png"
		input.RequestKey += "-" + string(kind)
		res := libraryMultipartRequest(t, mux, now, "alice", input, nil)
		if res.Code != 201 || !strings.Contains(res.Body.String(), "房间-原名.png") {
			t.Fatal(res.Code, res.Body.String())
		}
	}
	for _, mutate := range []func(*screenplayLibraryUploadInput){func(in *screenplayLibraryUploadInput) { in.Name = " " }, func(in *screenplayLibraryUploadInput) { in.Type = "avatar" }, func(in *screenplayLibraryUploadInput) { in.Bytes = []byte("not an image") }, func(in *screenplayLibraryUploadInput) { in.Bytes = nil }, func(in *screenplayLibraryUploadInput) { in.Bytes = make([]byte, screenplayLibraryImageLimit+1) }} {
		in := libraryUploadFixture(t)
		mutate(&in)
		if _, err := s.UploadEntry(context.Background(), "alice", in); !errors.Is(err, screenplay.ErrInvalidArgument) {
			t.Fatal(err)
		}
	}
	in := libraryUploadFixture(t)
	res := libraryMultipartRequest(t, mux, now, "alice", in, map[string]string{"url": "https://arbitrary.invalid/image.png"})
	if res.Code != 400 {
		t.Fatal(res.Code)
	}
}
func TestScreenplayAssetLibraryManagerUploadConcurrentAndUnknownCommit(t *testing.T) {
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	in := libraryUploadFixture(t)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.UploadEntry(context.Background(), "alice", in); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	rows, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
	if len(rows) != 1 || len(objects.files) != 2 {
		t.Fatal(len(rows), len(objects.files))
	}
	changed := in
	changed.Name = "changed same key"
	if _, err := s.UploadEntry(context.Background(), "alice", changed); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal(err)
	}
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			svc, _, objects, _, _ := assetLibraryServiceFixture(t)
			repo := &commitLostAssetLibraryRepository{AssetLibraryRepository: svc.Repository, lookupUnavailable: unknown}
			svc.Repository = repo
			_, err := svc.UploadEntry(context.Background(), "alice", libraryUploadFixture(t))
			if unknown && err == nil || !unknown && err != nil {
				t.Fatal(err)
			}
			if len(objects.files) != 2 {
				t.Fatal("committed upload object removed", len(objects.files))
			}
			repo.lookupUnavailable = false
			if _, err = svc.UploadEntry(context.Background(), "alice", libraryUploadFixture(t)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type deleteAckLostLibraryRepo struct {
	screenplay.AssetLibraryRepository
}

func (r deleteAckLostLibraryRepo) DeleteAssetLibraryEntry(ctx context.Context, owner, id string, revision int64) error {
	if err := r.AssetLibraryRepository.DeleteAssetLibraryEntry(ctx, owner, id, revision); err != nil {
		return err
	}
	return fmt.Errorf("delete commit acknowledgement lost")
}
func TestScreenplayAssetLibraryManagerDeleteKeepsStagedProjectCopy(t *testing.T) {
	s, entry, target, _, _ := stageFixture(t)
	item, err := s.StageImage(context.Background(), "alice", target.ProjectID, entry.ID, stageInput(entry, target))
	if err != nil {
		t.Fatal(err)
	}
	s.Repository = deleteAckLostLibraryRepo{s.Repository}
	if err = s.DeleteEntry(context.Background(), "alice", entry.ID, entry.Revision); err != nil {
		t.Fatal(err)
	}
	cut, err := s.Media.Cut.Asset(context.Background(), "alice", target.ProjectID, item.AssetID)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := s.Media.Objects.Open(context.Background(), aicutObjectAsset(cut))
	if err != nil {
		t.Fatal("deleting library removed staged image", err)
	}
	reader.Close()
}
func TestScreenplayAssetLibraryManagerCancelledUploadDoesNotWriteOrDeleteCommittedObjects(t *testing.T) {
	s, _, objects, _, _ := assetLibraryServiceFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.UploadEntry(ctx, "alice", libraryUploadFixture(t)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if len(objects.files) != 1 {
		t.Fatal(len(objects.files))
	}
	ctx, cancel = context.WithCancel(context.Background())
	s.Media.Objects = &cancellingStageObjectStore{AssetObjectStore: s.Media.Objects, cancel: cancel}
	if _, err := s.UploadEntry(ctx, "alice", libraryUploadFixture(t)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if len(objects.files) != 1 {
		t.Fatal("cancel left an uncommitted upload", len(objects.files))
	}
	rows, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", "", "", "")
	if len(rows) != 0 {
		t.Fatal(rows)
	}
}

var _ agentworkspace.AssetObjectStore = (*cancellingStageObjectStore)(nil)

func TestScreenplayAssetLibraryManagerDeletedUploadKeyCannotResurrectOrStage(t *testing.T) {
	s, _, objects, mux, now := assetLibraryServiceFixture(t)
	input := libraryUploadFixture(t)
	entry, err := s.UploadEntry(context.Background(), "alice", input)
	if err != nil {
		t.Fatal(err)
	}
	targetProject, _ := s.Media.Projects.CreateProject(context.Background(), screenplay.Project{ID: "tombstone-target", OwnerUsername: "alice"})
	target, _ := s.Media.Projects.CreateAsset(context.Background(), screenplay.Asset{OwnerUsername: "alice", ProjectID: targetProject.ID, Type: screenplay.AssetCharacter, Name: "target"})
	if err = s.DeleteEntry(context.Background(), "alice", entry.ID, entry.Revision); err != nil {
		t.Fatal(err)
	}
	if _, err = s.UploadEntry(context.Background(), "alice", input); !errors.Is(err, screenplay.ErrInvalidState) {
		t.Fatal("late upload revived deleted entry", err)
	}
	rows, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", "", "", "")
	if len(rows) != 0 || len(objects.files) != 1 {
		t.Fatal(rows, len(objects.files))
	}
	result := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+target.ProjectID+"/asset-library/"+entry.ID+"/stage", screenplayLibraryStageInput{RequestKey: "old-picker-key-001", TargetID: target.ID, ExpectedEntryRevision: &entry.Revision})
	if result.Code != 404 {
		t.Fatal("old picker was allowed to stage a deleted revision", result.Code, result.Body.String())
	}
	patch := signedJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+entry.ID, map[string]any{"expectedRevision": entry.Revision, "name": "replacement"})
	if patch.Code != 404 {
		t.Fatal(patch.Code)
	}
}
