package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	txt "qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

type settingLibrary struct {
	txt.Repository
	owners map[string]txt.LibraryBackground
}

func TestBatchFactoryTextHighlightSettingsSnapshot(t *testing.T) {
	x := textRequest{book: batchfactoryv11.Book{Title: "测试书", SourceMetadata: map[string]any{"illustrationHighlightKeywords": []string{"关键词"}}}, settings: batchfactoryv11.SettingsPatch{"textHighlightRules": json.RawMessage(`"dialogue,lines,keywords"`), "textHighlightLines": json.RawMessage(`"1,3"`)}}
	got, err := bookTextSnapshot(x)
	if err != nil {
		t.Fatal(err)
	}
	if got.Layout.HighlightRules != "dialogue,lines,keywords" || got.Layout.HighlightLines != "1,3" || !got.HighlightReuseEnabled || len(got.HighlightWords) != 1 {
		t.Fatal(got)
	}
	x.settings["textHighlightRules"] = json.RawMessage(`"off"`)
	off, err := bookTextSnapshot(x)
	if err != nil || off.HighlightReuseEnabled || off.CoverFingerprint != got.CoverFingerprint {
		t.Fatal("body rules changed cover or failed to turn off", err)
	}
}

func (lib settingLibrary) ListBackgrounds(_ context.Context, owner string) ([]txt.LibraryBackground, error) {
	if bg, ok := lib.owners[owner]; ok {
		return []txt.LibraryBackground{bg}, nil
	}
	return nil, nil
}
func (lib settingLibrary) GetBackground(_ context.Context, owner, id string) (txt.LibraryBackground, error) {
	if bg, ok := lib.owners[owner]; ok && bg.ID == id {
		return bg, nil
	}
	return txt.LibraryBackground{}, txt.ErrNotFound
}
func (lib settingLibrary) SaveBackground(_ context.Context, owner, name string, file txt.ArtifactFile) (txt.LibraryBackground, error) {
	bg := txt.LibraryBackground{ID: owner + "-bg", Name: name, ArtifactFile: file}
	lib.owners[owner] = bg
	return bg, nil
}

func TestBatchFactoryTextPersonalSettingsLibrary(t *testing.T) {
	now := time.Unix(1700000000, 0)
	lib := settingLibrary{owners: make(map[string]txt.LibraryBackground)}
	api := NewRouter(RouterOptions{Slice: 1, Store: batchfactoryv11.NewMemoryStore(), TextRepository: lib, LocalArtifacts: localartifact.NewStore(t.TempDir(), 32<<20), BridgeSecret: "secret", Now: func() time.Time { return now }})
	path := "/api/batch-factory/v12/text-backgrounds"
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", path+"?name=paper", bytes.NewReader(data.Bytes()))
	SignBridgeRequest(req, "alice", false, now, "secret")
	uploaded := httptest.NewRecorder()
	api.ServeHTTP(uploaded, req)
	if uploaded.Code != 201 {
		t.Fatalf("upload %d %s", uploaded.Code, uploaded.Body.String())
	}
	for owner, expected := range map[string]int{"alice": 1, "bob": 0} {
		rr := signedJSONRequest(t, api, now, owner, "GET", path, nil)
		result := decodeBody[struct{ Items []textMediaView }](t, rr)
		if rr.Code != 200 || len(result.Items) != expected {
			t.Fatalf("owner %s response %d %s", owner, rr.Code, rr.Body.String())
		}
	}
	owned := signedJSONRequest(t, api, now, "alice", "GET", path+"/alice-bg/content", nil)
	if owned.Code != 200 || !bytes.Equal(data.Bytes(), owned.Body.Bytes()) {
		t.Fatalf("owner could not read uploaded image: status %d", owned.Code)
	}
	rr := signedJSONRequest(t, api, now, "bob", "GET", path+"/alice-bg/content", nil)
	if rr.Code != 404 {
		t.Fatalf("foreign library response %d", rr.Code)
	}
}

func TestBatchFactoryTextPaginationModeSettings(t *testing.T) {
	for _, mode := range []string{"auto", "target"} {
		settings := batchfactoryv11.SettingsPatch{"textPaginationMode": json.RawMessage(`"` + mode + `"`), "textPageTarget": json.RawMessage(`150`)}
		snap, err := bookTextSnapshot(textRequest{settings: settings})
		if err != nil || snap.Layout.PaginationMode != mode || snap.Layout.TargetCharacters != 150 {
			t.Fatalf("pagination mode not carried into renderer: %+v %v", snap.Layout, err)
		}
	}
	snap, err := bookTextSnapshot(textRequest{settings: batchfactoryv11.SettingsPatch{}})
	if err != nil || snap.Layout.PaginationMode != "auto" {
		t.Fatalf("expected smart default: %+v %v", snap.Layout, err)
	}
}

func TestBatchFactoryTextCoverSparseSettings(t *testing.T) {
	batch := batchfactoryv11.SettingsPatch{"textFontSize": json.RawMessage("56"), "textPadding": json.RawMessage("0"), "textCover": json.RawMessage(`{"layout":{"fontId":"ma-shan-zheng"},"background":{"kind":"color","color":"#123456"}}`)}
	book := batchfactoryv11.SettingsPatch{"textCover": json.RawMessage(`{"layout":{"align":"right"}}`)}
	snap, err := bookTextSnapshot(textRequest{settings: effectiveTextSettings(batch, book)})
	if err != nil || snap.Cover.Layout.FontID != "ma-shan-zheng" || snap.Cover.Layout.FontSize != 56 || snap.Cover.Layout.Align != "right" || snap.Cover.Layout.Padding != 0 || snap.Cover.Background != nil {
		t.Fatalf("cover resolved to %+v, error %v", snap, err)
	}
}

// The front-end save/reopen test produces this exact settings payload.
func TestBatchFactoryTextSavedProfileResolvesForPublish(t *testing.T) {
	data, err := os.ReadFile("testdata/text_profile_snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	batch, book := publishFieldsFixture(t, `{"websiteProfileId":"121-4","organization":"3","bookCategory":"已有分类"}`)
	if err := json.Unmarshal(data, &book.SettingsState.Patch); err != nil {
		t.Fatal(err)
	}
	fields, err := buildPublishFields(batch, book, nil)
	if err != nil {
		t.Fatal(err)
	}
	for key, expected := range map[string]string{"organization": "7", "jieya_num": "8", "gunping_num": "0", "font_color_styles": "[1,2]"} {
		if fields[key] != expected {
			t.Errorf("publish field %s = %q, want %q", key, fields[key], expected)
		}
	}
	var publish map[string]any
	if err := json.Unmarshal(effectiveTextSettings(batch.SettingsState.Patch, book.SettingsState.Patch)["publishSettings"], &publish); err != nil || publish["bookCategory"] != "已有分类" {
		t.Fatalf("book did not inherit category: %v, %v", publish, err)
	}
}
