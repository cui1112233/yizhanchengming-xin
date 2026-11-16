package httpapi

import (
	"context"
	"encoding/json"
	"qiantie/backend/internal/characterarchive"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
	"testing"
)

func TestLibrarySubjectUploadReferenceApplyAndClear(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	ctx := context.Background()
	row, err := s.UploadEntry(ctx, "alice", libraryUploadFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	repo := screenplay.NewMemorySubjectUploadRepository()
	s.SubjectUploads = &ScreenplaySubjectUploadService{Repository: repo}
	result := subjectupload.Result{AssetID: "own-asset", AssetURI: "asset://own-asset", GroupID: "own-group", ProjectName: "own-project", Provider: "volcengine", Status: "Active"}
	_, err = repo.SaveSubjectUpload(ctx, screenplay.SubjectUploadOperation{Owner: "alice", RequestKey: "own-upload-ref", Digest: "own-digest", Status: "succeeded", Result: &result}, 0)
	if err != nil {
		t.Fatal(err)
	}
	w := signedJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+row.ID, map[string]any{"expectedRevision": row.Revision, "sora2Info": "manual-opaque", "subjectUploadKey": "own-upload-ref"})
	if w.Code != 200 {
		t.Fatalf("apply ref: %d %s", w.Code, w.Body.String())
	}
	var response struct {
		Entry map[string]any `json:"entry"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Entry["sora2Info"] != "manual-opaque" || response.Entry["subjectUploadKey"] != "own-upload-ref" {
		t.Fatal(response.Entry)
	}
	meta, ok := response.Entry["subjectUpload"].(map[string]any)
	if !ok || meta["groupId"] != "own-group" {
		t.Fatal(response.Entry)
	}
	saved, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", row.ID)
	archive, e := s.ExportCharacters(ctx, "alice")
	if e != nil {
		t.Fatal(e)
	}
	raw, e := characterarchive.DecodeJSON(archive)
	if e != nil {
		t.Fatal(e)
	}
	var rows []map[string]any
	json.Unmarshal(raw, &rows)
	if rows[0]["volcengineAssetId"] != "own-asset" || rows[0]["volcengineAssetGroupId"] != "own-group" {
		t.Fatal(rows)
	}
	// A request from an old client must retain the association.
	w = signedJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+row.ID, map[string]any{"expectedRevision": saved.Revision, "description": "old-client"})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	saved, _ = s.Repository.GetAssetLibraryEntry(ctx, "alice", row.ID)
	w = signedJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+row.ID, map[string]any{"expectedRevision": saved.Revision, "subjectUploadKey": ""})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Entry["subjectUploadKey"] != "" || response.Entry["subjectUpload"] != nil {
		t.Fatal(response.Entry)
	}
	other, _ := s.Repository.GetAssetLibraryEntry(ctx, "alice", row.ID)
	w = signedJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+row.ID, map[string]any{"expectedRevision": other.Revision, "subjectUploadKey": "foreign-missing"})
	if w.Code == 200 {
		t.Fatal("foreign reference accepted")
	}
}
