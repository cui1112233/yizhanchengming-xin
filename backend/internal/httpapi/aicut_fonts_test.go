package httpapi

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"qiantie/backend/internal/aicut"
)

func testAICutTTF(family string) []byte {
	encoded := utf16.Encode([]rune(family))
	name := make([]byte, 18+len(encoded)*2)
	binary.BigEndian.PutUint16(name[2:4], 1)
	binary.BigEndian.PutUint16(name[4:6], 18)
	binary.BigEndian.PutUint16(name[6:8], 3)
	binary.BigEndian.PutUint16(name[8:10], 1)
	binary.BigEndian.PutUint16(name[10:12], 0x0409)
	binary.BigEndian.PutUint16(name[12:14], 1)
	binary.BigEndian.PutUint16(name[14:16], uint16(len(encoded)*2))
	for index, value := range encoded {
		binary.BigEndian.PutUint16(name[18+index*2:20+index*2], value)
	}
	font := make([]byte, 28+len(name))
	binary.BigEndian.PutUint32(font[0:4], 0x00010000)
	binary.BigEndian.PutUint16(font[4:6], 1)
	copy(font[12:16], []byte("name"))
	binary.BigEndian.PutUint32(font[20:24], 28)
	binary.BigEndian.PutUint32(font[24:28], uint32(len(name)))
	copy(font[28:], name)
	return font
}

func TestParseAICutFontFamilyReadsInternalName(t *testing.T) {
	got, err := parseAICutFontFamily(testAICutTTF("一战测试字体"))
	if err != nil || got != "一战测试字体" {
		t.Fatalf("family=%q err=%v", got, err)
	}
}

func TestDetectedAICutAssetTypeAcceptsOnlyRealTTFOrOTF(t *testing.T) {
	if contentType, kind := detectedAICutAssetType("application/octet-stream", "title.ttf", testAICutTTF("Title Font")); contentType != "font/ttf" || kind != "font" {
		t.Fatalf("ttf detection=%q %q", contentType, kind)
	}
	if _, kind := detectedAICutAssetType("font/ttf", "renamed.ttf", []byte("not-a-font")); kind != "" {
		t.Fatalf("invalid renamed font accepted as %q", kind)
	}
}

func TestAICutFontUploadReturnsFamilyAndListsIt(t *testing.T) {
	now := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-font", Title: "字体", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	objects := &agentAssetObjectTestStore{files: map[string][]byte{}}
	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, objects, func() time.Time { return now })

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "title.ttf")
	_, _ = part.Write(testAICutTTF("一战测试字体"))
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/ai-cut/projects/cut-font/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(req, "alice", true, now, "secret")
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, req)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"kind":"font"`) || !strings.Contains(created.Body.String(), `"fontFamily":"一战测试字体"`) {
		t.Fatalf("upload=%d %s", created.Code, created.Body.String())
	}

	listed := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/ai-cut/projects/cut-font/assets", nil)
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), `"fontFamily":"一战测试字体"`) {
		t.Fatalf("list=%d %s", listed.Code, listed.Body.String())
	}
	var payload struct {
		Assets []aicut.Asset `json:"assets"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &payload); err != nil || len(payload.Assets) != 1 || payload.Assets[0].FontFamily != "一战测试字体" {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
}

func TestPrepareAICutExportFontsCopiesReferencedFontAndUsesInternalFamily(t *testing.T) {
	now := time.Date(2026, 11, 11, 9, 30, 0, 0, time.UTC)
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "cut-font-export", Title: "字体导出", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	asset := aicut.Asset{ID: "font-1", ProjectID: "cut-font-export", Kind: "font", Name: "title.ttf", ContentType: "font/ttf", ObjectKey: "font-1.ttf", ByteSize: int64(len(testAICutTTF("导出字体"))), CreatedAt: now}
	if err := store.CreateAsset(t.Context(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	objects := &agentAssetObjectTestStore{files: map[string][]byte{asset.ObjectKey: testAICutTTF("导出字体")}}
	timeline := aicut.Timeline{Tracks: []aicut.Track{{ID: "text-1", Kind: aicut.TrackSubtitle, Name: "标题", Settings: aicut.TrackSettings{FontAssetID: "font-1", FontFamily: "伪造名称"}}}}
	prepared, fontsDir, err := prepareAICutExportFonts(t.Context(), "alice", "cut-font-export", t.TempDir(), timeline, store, objects)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Tracks[0].Settings.FontFamily != "导出字体" || timeline.Tracks[0].Settings.FontFamily != "伪造名称" {
		t.Fatalf("prepared=%+v original=%+v", prepared.Tracks[0].Settings, timeline.Tracks[0].Settings)
	}
	files, err := os.ReadDir(fontsDir)
	if err != nil || len(files) != 1 || filepath.Ext(files[0].Name()) != ".ttf" {
		t.Fatalf("font files=%v err=%v", files, err)
	}
}
