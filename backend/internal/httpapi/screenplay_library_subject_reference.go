package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
)

const librarySubjectResultKey = "_cm_library_subject_result"
const librarySubjectRequestKey = "_cm_library_subject_request"

func (s *ScreenplayAssetLibrary) resolveLibrarySubjectReference(ctx context.Context, owner string, kind screenplay.AssetType, key *string) (*subjectupload.Result, error) {
	if key == nil {
		return nil, nil
	}
	if kind != screenplay.AssetCharacter {
		return nil, fmt.Errorf("%w: 仅角色支持主体上传关联", screenplay.ErrInvalidArgument)
	}
	if *key == "" {
		return nil, nil
	}
	if !validAssetLibraryRequestKey(*key) {
		return nil, screenplay.ErrInvalidArgument
	}
	if s.SubjectUploads == nil || s.SubjectUploads.Repository == nil {
		return nil, screenplay.ErrInvalidState
	}
	op, err := s.SubjectUploads.Repository.GetSubjectUpload(ctx, owner, *key)
	if err != nil {
		return nil, err
	}
	if op.Status != "succeeded" || op.Result == nil || op.Result.AssetID == "" {
		return nil, fmt.Errorf("%w: 请先核对上传成功后再应用", screenplay.ErrInvalidArgument)
	}
	value := *op.Result
	return &value, nil
}
func setLibrarySubjectReference(entry *screenplay.AssetLibraryEntry, key *string, value *subjectupload.Result) {
	if key == nil {
		return
	}
	extra := make(map[string]string, len(entry.Snapshot.Extra)+2)
	for k, v := range entry.Snapshot.Extra {
		extra[k] = v
	}
	extra[librarySubjectRequestKey] = *key
	extra[librarySubjectResultKey] = ""
	if value != nil {
		raw, _ := json.Marshal(value)
		extra[librarySubjectResultKey] = string(raw)
	}
	entry.Snapshot.Extra = extra
}
func librarySubjectResult(entry screenplay.AssetLibraryEntry) *subjectupload.Result {
	if entry.Type != screenplay.AssetCharacter {
		return nil
	}
	if raw, ok := entry.Snapshot.Extra[librarySubjectResultKey]; ok {
		var result subjectupload.Result
		if json.Unmarshal([]byte(raw), &result) == nil && result.AssetID != "" {
			return &result
		}
		return nil
	}
	var record map[string]json.RawMessage
	if json.Unmarshal([]byte(entry.Snapshot.Extra[screenplay.LibraryArchiveJSON]), &record) != nil {
		return nil
	}
	text := func(key string) string { var v string; _ = json.Unmarshal(record[key], &v); return v }
	r := subjectupload.Result{AssetID: text("volcengineAssetId"), AssetURI: text("volcengineAssetUri"), GroupID: text("volcengineAssetGroupId"), ProjectName: text("volcengineProjectName"), Provider: text("realPersonAssetProvider")}
	if r.AssetID == "" {
		return nil
	}
	return &r
}
func librarySubjectReferenceEqual(a, b screenplay.AssetLibraryEntry) bool {
	return a.Snapshot.Extra[librarySubjectRequestKey] == b.Snapshot.Extra[librarySubjectRequestKey] && a.Snapshot.Extra[librarySubjectResultKey] == b.Snapshot.Extra[librarySubjectResultKey]
}
func librarySubjectReferenceSource(entry *screenplay.AssetLibraryEntry, source screenplay.Asset) {
	if source.Type != screenplay.AssetCharacter {
		return
	}
	var r subjectupload.Result
	if json.Unmarshal([]byte(source.Extra["subject_upload_result"]), &r) == nil && r.AssetID != "" {
		empty := ""
		setLibrarySubjectReference(entry, &empty, &r)
	}
}
func exportLibrarySubjectReference(record map[string]json.RawMessage, entry screenplay.AssetLibraryEntry) {
	r := librarySubjectResult(entry)
	_, explicit := entry.Snapshot.Extra[librarySubjectResultKey]
	if r == nil && !explicit {
		return
	}
	if r == nil {
		r = &subjectupload.Result{}
	}
	for k, v := range map[string]string{"volcengineAssetId": r.AssetID, "volcengineAssetUri": r.AssetURI, "volcengineAssetGroupId": r.GroupID, "volcengineProjectName": r.ProjectName, "realPersonAssetProvider": r.Provider} {
		raw, _ := json.Marshal(v)
		record[k] = raw
	}
	// A current typed result supersedes stale provider payloads from imported files.
	record["volcengineAssetData"] = json.RawMessage("null")
	record["volcengineAssetImageUrl"] = json.RawMessage(`""`)
	record["volcengineAssetUrl"] = json.RawMessage(`""`)
}
