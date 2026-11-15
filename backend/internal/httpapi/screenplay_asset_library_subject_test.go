package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
)

func TestLibrarySubjectHTTPPreserveClearValidate(t *testing.T) {
	s, _, _, mux, now := assetLibraryServiceFixture(t)
	value := "  {\"主体\":\"雪😀\"}\n "
	in := libraryUploadFixture(t)
	w := libraryMultipartRequest(t, mux, now, "alice", in, map[string]string{"sora2Info": value})
	if w.Code != 201 {
		t.Fatalf("subject multipart: %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Entry map[string]any `json:"entry"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	id := out.Entry["id"].(string)
	if out.Entry["sora2Info"] != value {
		t.Fatal(out.Entry)
	}
	for _, clear := range []bool{false, true} {
		row, _ := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", id)
		patch := map[string]any{"expectedRevision": row.Revision, "description": "updated"}
		if clear {
			patch["sora2Info"] = ""
		}
		w = bulkJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+id, patch)
		if w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
		json.Unmarshal(w.Body.Bytes(), &out)
		want := value
		if clear {
			want = ""
		}
		if out.Entry["sora2Info"] != want {
			t.Fatal(out.Entry)
		}
	}
	row, _ := s.Repository.GetAssetLibraryEntry(context.Background(), "alice", id)
	w = bulkJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+id, map[string]any{"expectedRevision": row.Revision, "sora2Info": strings.Repeat("雪", 5462)})
	if w.Code != 400 {
		t.Fatal(w.Code, w.Body.String())
	}
	in.Type = screenplay.AssetScene
	in.RequestKey = "subject-scene-001"
	w = libraryMultipartRequest(t, mux, now, "alice", in, map[string]string{"sora2Info": ""})
	if w.Code != 400 {
		t.Fatal("non character accepted", w.Code)
	}
}
func TestLibrarySubjectArchiveFallbackCurrentAndExport(t *testing.T) {
	for i, record := range []string{`{"name":"camel","sora2Info":"  文😀\n","sora2_info":"old","unknown":{"secret":true}}`, `{"name":"snake","sora2_info":"蛇"}`, `{"name":"empty","sora2Info":"","sora2_info":"old"}`} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s, _, _, mux, now := assetLibraryServiceFixture(t)
			w := bulkHTTPRequest(t, mux, now, "alice", "subject-import-001", map[string][]byte{"characters.json": []byte("[" + record + "]")})
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			rows, _ := s.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
			var original map[string]any
			json.Unmarshal([]byte(record), &original)
			want, ok := original["sora2Info"]
			if !ok {
				want = original["sora2_info"]
			}
			if screenplayLibraryPublic(rows[0])["sora2Info"] != want {
				t.Fatal("archive fallback", screenplayLibraryPublic(rows[0]))
			}
			w = bulkJSONRequest(t, mux, now, "alice", "PATCH", "/api/comic-screenplays/asset-library/"+rows[0].ID, map[string]any{"expectedRevision": rows[0].Revision, "sora2Info": ""})
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			exported, e := s.ExportCharacters(context.Background(), "alice")
			if e != nil {
				t.Fatal(e)
			}
			s2, _, _, mux2, now2 := assetLibraryServiceFixture(t)
			w = bulkHTTPRequest(t, mux2, now2, "alice", "subject-reimport-001", map[string][]byte{"characters.lschar": exported})
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			after, _ := s2.Repository.ListAssetLibrary(context.Background(), "alice", screenplay.AssetCharacter, "", "")
			if screenplayLibraryPublic(after[0])["sora2Info"] != "" {
				t.Fatal("cleared subject resurrected")
			}
			var raw map[string]any
			json.Unmarshal([]byte(after[0].Snapshot.Extra[screenplay.LibraryArchiveJSON]), &raw)
			if _, ok := original["sora2_info"]; ok && raw["sora2_info"] != "" {
				t.Fatal("stale snake", raw)
			}
			if original["unknown"] != nil && raw["unknown"] == nil {
				t.Fatal("lost unknown metadata")
			}
		})
	}
	for i, value := range []string{`42`, `null`, `{}`, fmt.Sprintf("%q", strings.Repeat("雪", 5462))} {
		_, _, _, mux, now := assetLibraryServiceFixture(t)
		w := bulkHTTPRequest(t, mux, now, "alice", fmt.Sprintf("subject-invalid-%d", i), map[string][]byte{"characters.json": []byte(`[{"name":"bad","sora2Info":` + value + `}]`)})
		if w.Code != 400 {
			t.Fatal("invalid archive accepted", w.Code, w.Body.String())
		}
	}
}

func TestLibrarySubjectReplaceDigestLostCommitAndSource(t *testing.T) {
	ctx := context.Background()
	s, source, _, _, _ := assetLibraryServiceFixture(t)
	value := "\n {\"id\":\"纯文本😀\"}  "
	in := libraryUploadFixture(t)
	in.Sora2Info = &value
	entry, err := s.UploadEntry(ctx, "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	changed := "different"
	in.Sora2Info = &changed
	if _, err = s.UploadEntry(ctx, "alice", in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal("create replay ignored subject", err)
	}
	in.Sora2Info = nil
	in.RequestKey = "subject-replace-001"
	replaced, err := s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in)
	if err != nil || librarySubject(replaced) != value {
		t.Fatal("image lost subject", err, replaced)
	}
	in.Sora2Info = &changed
	if _, err = s.ReplaceImage(ctx, "alice", entry.ID, entry.Revision, in); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal("replace replay ignored subject", err)
	}
	in.RequestKey = "subject-audio-001"
	in.Bytes = nil
	in.AudioBytes = libraryReferenceWAV()
	in.AudioName = "ref.wav"
	replaced, err = s.ReplaceReferenceAudio(ctx, "alice", entry.ID, replaced.Revision, in)
	if err != nil || librarySubject(replaced) != changed {
		t.Fatal("audio subject", err)
	}
	in.RequestKey = "subject-image-again"
	in.Bytes = testImageToolPNG(t)
	in.AudioBytes = nil
	in.AudioName = ""
	in.Sora2Info = nil
	replaced, err = s.ReplaceImage(ctx, "alice", entry.ID, replaced.Revision, in)
	if err != nil || librarySubject(replaced) != changed {
		t.Fatal("image preserve", err)
	}
	empty := ""
	old := replaced.Revision
	replaced, err = s.UpdateEntry(ctx, "alice", entry.ID, screenplayLibraryMetadataInput{Expected: &old, Sora2Info: &empty})
	if err != nil || librarySubject(replaced) != "" {
		t.Fatal(err)
	}
	if _, err = s.UpdateEntry(ctx, "alice", entry.ID, screenplayLibraryMetadataInput{Expected: &old, Sora2Info: &changed}); !errors.Is(err, screenplay.ErrRevisionConflict) {
		t.Fatal("stale subject accepted", err)
	}
	// A failed acknowledgement may be recovered only for the same subject.
	base := s.Repository
	s.Repository = &commitLostAssetLibraryRepository{AssetLibraryRepository: base}
	rev := replaced.Revision
	replaced, err = s.UpdateEntry(ctx, "alice", entry.ID, screenplayLibraryMetadataInput{Expected: &rev, Sora2Info: &value})
	if err != nil || librarySubject(replaced) != value {
		t.Fatal("lost commit subject", err)
	}
	s.Repository = base
	source.Extra["sora2_info"] = value
	source, err = s.Media.Projects.UpdateAsset(ctx, "alice", source.ProjectID, source.ID, source, source.Revision)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := s.SaveSource(ctx, "alice", source.ProjectID, sourceLibraryInput(s, source))
	if err != nil || librarySubject(saved) != value {
		t.Fatal("source subject", err, saved)
	}
	projectAssets, _ := s.Media.Projects.ListAssets(ctx, "alice", source.ProjectID, "")
	for _, asset := range projectAssets {
		if _, ok := asset.Extra[screenplayLibrarySubjectKey]; ok {
			t.Fatal("SaveSource mutated original project extra")
		}
	}
}

func TestLibrarySubjectPrivateArchivePriorityAndBoundary(t *testing.T) {
	entry := screenplay.AssetLibraryEntry{Type: screenplay.AssetCharacter, Snapshot: screenplay.Asset{Extra: map[string]string{screenplay.LibraryArchiveJSON: `{"sora2Info":"camel","sora2_info":"snake","supplier":{"token":"secret"}}`}}}
	if librarySubject(entry) != "camel" {
		t.Fatal("camel priority")
	}
	empty := ""
	setLibrarySubject(&entry, &empty)
	if librarySubject(entry) != "" {
		t.Fatal("clear resurrects raw")
	}
	public, _ := json.Marshal(screenplayLibraryPublic(entry))
	if strings.Contains(string(public), "secret") || strings.Contains(string(public), "snapshot") || strings.Contains(string(public), "supplier") {
		t.Fatal("private metadata exposed", string(public))
	}
	exact := strings.Repeat("雪", 5461) + "a"
	if len(exact) != 16384 || validateLibrarySubject(screenplay.AssetCharacter, &exact) != nil {
		t.Fatal("valid boundary rejected")
	}
	overflow := exact + "a"
	if validateLibrarySubject(screenplay.AssetCharacter, &overflow) == nil {
		t.Fatal("over boundary accepted")
	}
}
