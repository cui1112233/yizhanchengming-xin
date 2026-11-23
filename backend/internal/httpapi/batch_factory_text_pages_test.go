package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/storage"
	"strings"
	"testing"
	"time"
)

func TestBatchFactoryTextSignedModeGuard(t *testing.T) {
	s := batchfactoryv11.NewMemoryStore()
	b, err := s.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "批次", Books: []batchfactoryv11.CreateBookInput{{Title: "书"}}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	api := NewRouter(RouterOptions{Slice: 1, Store: s, BridgeSecret: "secret", Now: func() time.Time { return now }})
	path := "/api/batch-factory/v12/batches/" + b.ID + "/books/" + b.Books[0].ID + "/text-pages"
	if got := signedJSONRequest(t, api, now, "alice", "GET", path, nil); got.Code != 409 {
		t.Fatalf("other mode: %d %s", got.Code, got.Body.String())
	}
	if got := signedJSONRequest(t, api, now, "bob", "GET", path, nil); got.Code != 404 {
		t.Fatalf("foreign: %d", got.Code)
	}
}

// The DSN must name a disposable schema. Book and page IDs are unique per run.
func TestBatchFactoryTextSignedMySQLLifecycle(t *testing.T) {
	dsn := os.Getenv("TEXT_REPO_TEST_DSN")
	if dsn == "" {
		t.Skip("set TEXT_REPO_TEST_DSN for the signed MySQL lifecycle")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, ddl := range append([]string{"CREATE TABLE IF NOT EXISTS batch_factory_v12_books(id VARCHAR(64) PRIMARY KEY,owner_username VARCHAR(191) NOT NULL,batch_id VARCHAR(64) NOT NULL) ENGINE=InnoDB"}, storage.V12TextProductionStatements()...) {
		if _, err = db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	s := batchfactoryv11.NewMemoryStore()
	const original = "这是第一句。这是第二句。"
	inputBook := batchfactoryv11.CreateBatchInput{Title: "批次", InitialBatchSettings: batchfactoryv11.SettingsPatch{"productionMode": mustJSON("text"), "sensitiveMode": mustJSON("off")}, Books: []batchfactoryv11.CreateBookInput{{Title: "书名", TxtText: original, SourceText: original, SourceMetadata: map[string]any{"rawSourceStatus": "verified"}}}}
	var b batchfactoryv11.Batch
	for {
		b, err = s.CreateBatch(context.Background(), "alice", inputBook)
		if err != nil {
			t.Fatal(err)
		}
		var used bool
		if err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM batch_factory_v12_books WHERE id=?)", b.Books[0].ID).Scan(&used); err != nil {
			t.Fatal(err)
		}
		if !used {
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	book := b.Books[0]
	if _, err = db.Exec("INSERT INTO batch_factory_v12_books(id,owner_username,batch_id) VALUES(?,?,?)", book.ID, "alice", b.ID); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1700000000, 0)
	files := localartifact.NewStore(t.TempDir(), 32<<20)
	api := NewRouter(RouterOptions{Slice: 1, Store: s, TextRepository: batchfactorytext.NewMySQLRepository(db), LocalArtifacts: files, BridgeSecret: "secret", Now: func() time.Time { return now }})
	base := "/api/batch-factory/v12/batches/" + b.ID + "/books/" + book.ID + "/text-"
	call := func(method, path string, body any, code int) *httptest.ResponseRecorder {
		t.Helper()
		rr := signedJSONRequest(t, api, now, "alice", method, base+path, body)
		if rr.Code != code {
			t.Fatalf("%s %s got %d want %d: %s", method, path, rr.Code, code, rr.Body.String())
		}
		return rr
	}
	type state struct {
		Document batchfactorytext.Document
		Pages    []batchfactorytext.Page
	}
	get := func() state { return decodeBody[state](t, call("GET", "pages", nil, 200)) }
	call("POST", "prepare", map[string]any{}, 200)
	st := get()
	if len(st.Pages) != 1 || st.Pages[0].Text != original || st.Document.OriginalText != original {
		t.Fatalf("state %+v", st)
	}
	pg := st.Pages[0]
	call("POST", "pages/"+pg.ID+"/split", map[string]any{"offset": 2, "expectedRevision": st.Document.Revision}, 400)
	call("POST", "pages/"+pg.ID+"/split", map[string]any{"offset": 6, "expectedRevision": st.Document.Revision}, 200)
	st = get()
	if len(st.Pages) != 2 {
		t.Fatalf("split %+v", st)
	}
	call("POST", "pages/"+st.Pages[0].ID+"/merge", map[string]any{"expectedRevision": st.Document.Revision}, 200)
	st = get()
	pg = st.Pages[0]
	call("PATCH", "pages/"+pg.ID, map[string]any{"expectedRevision": pg.Revision + 1, "background": map[string]string{"kind": "color", "color": "#AA0000"}}, 409)
	img := call("POST", "pages/"+pg.ID+"/preview", map[string]any{"background": map[string]string{"kind": "color", "color": "#AA0000"}}, 200).Body.Bytes()
	decoded, err := png.Decode(bytes.NewReader(img))
	if err != nil || decoded.Bounds().Dx() != 1080 {
		t.Fatalf("preview %v", err)
	}
	r, g, _, _ := decoded.At(2, 2).RGBA()
	if r != 0xAAAA || g != 0 {
		t.Fatal("preview background mismatch")
	}
	rev := func() map[string]any { return map[string]any{"expectedRevision": get().Document.Revision} }
	input := func() map[string]any {
		d := get().Document
		return map[string]any{"versionId": d.ActiveVersionID, "paginationVersion": d.PaginationVersion}
	}
	call("POST", "collections", input(), 409)
	call("POST", "cover/render", rev(), 201)
	call("POST", "collections", input(), 409)
	call("POST", "pages/"+pg.ID+"/render", map[string]any{"expectedRevision": pg.Revision}, 201)
	result := call("POST", "collections", input(), 201)
	if !strings.Contains(result.Body.String(), "contentUrl") || strings.Contains(result.Body.String(), "storageRef") {
		t.Fatalf("collection media: %s", result.Body.String())
	}
	var collection struct {
		Items           []struct{ ContentURL string }
		CoverContentURL string
	}
	collection = decodeBody[struct {
		Items           []struct{ ContentURL string }
		CoverContentURL string
	}](t, result)
	mediaPath := collection.Items[0].ContentURL
	rows, err := batchfactorytext.NewMySQLRepository(db).ListArtifacts(context.Background(), batchfactorytext.Scope{Owner: "alice", BatchID: b.ID, BookID: book.ID})
	if err != nil {
		t.Fatal(err)
	}
	for _, art := range rows {
		if strings.HasSuffix(mediaPath, "/"+art.ID+"/content") {
			file, err := files.LocalPath(art.StorageRef)
			if err != nil {
				t.Fatal(err)
			}
			original, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(file, []byte("损坏"), 0600); err != nil {
				t.Fatal(err)
			}
			call("POST", "collections", input(), 409)
			if err = os.WriteFile(file, original, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	media := signedJSONRequest(t, api, now, "alice", "GET", mediaPath, nil)
	if media.Code != 200 {
		t.Fatalf("owner read %s got %d", mediaPath, media.Code)
	}
	if _, err := png.Decode(bytes.NewReader(media.Body.Bytes())); err != nil {
		t.Fatal(err)
	}
	if foreign := signedJSONRequest(t, api, now, "bob", "GET", mediaPath, nil); foreign.Code != 404 {
		t.Fatalf("foreign read: %d", foreign.Code)
	}
	upload := func(path string, payload []byte, code int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("POST", base+path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "image/png")
		SignBridgeRequest(req, "alice", false, now, "secret")
		rr := httptest.NewRecorder()
		api.ServeHTTP(rr, req)
		if rr.Code != code {
			t.Fatalf("upload %s: %d %s", path, rr.Code, rr.Body.String())
		}
		return rr
	}
	bg := decodeBody[struct{ ID string }](t, upload("backgrounds?name=test", img, 201))
	if bg.ID == "" {
		t.Fatal("missing background")
	}
	call("POST", "pages/"+pg.ID+"/preview", map[string]any{"background": map[string]string{"kind": "image", "id": bg.ID}}, 200)
	call("GET", "backgrounds", nil, http.StatusOK)
	upload("backgrounds?name=bad", []byte("bad"), 400)
	fmtRev := func(r int64) string { return fmt.Sprint(r) }
	upload("pages/"+pg.ID+"/upload?expectedRevision="+fmtRev(get().Pages[0].Revision), img, 201)
	st = get()
	if !st.Pages[0].Locked {
		t.Fatal("unlocked after upload")
	}
	call("POST", "pages/"+pg.ID+"/render", map[string]any{"expectedRevision": st.Pages[0].Revision}, 409)
	before := st.Document
	_, err = s.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: b.ID, Content: "修改后的整段文字。"})
	if err != nil {
		t.Fatal(err)
	}
	call("POST", "prepare", map[string]any{"expectedRevision": before.Revision}, 409)
	after := get()
	if after.Document.ActiveVersionID != before.ActiveVersionID || after.Document.Revision != before.Revision || after.Pages[0].ActiveArtifactID != st.Pages[0].ActiveArtifactID {
		t.Fatalf("failed reflow changed visible version/locked artifact")
	}

	call("POST", "pages/"+pg.ID+"/restore", map[string]any{"expectedRevision": st.Pages[0].Revision}, 200)
	upload("cover/upload?expectedRevision="+fmtRev(get().Document.Revision), img, 201)
	call("POST", "cover/render", rev(), 409)
	call("POST", "cover/restore", rev(), 200)
	call("POST", "cover/render", rev(), 201)
	fresh, err := s.GetBatch(context.Background(), "alice", b.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: b.ID}, batchfactoryv11.SettingsUpdate{ExpectedRevision: fresh.Revision, Patch: batchfactoryv11.SettingsPatch{"textBackground": mustJSON(map[string]string{"type": "color", "color": "#CC0000"})}})
	if err != nil {
		t.Fatal(err)
	}
	call("POST", "prepare", map[string]any{}, 200)
	if get().Document.CoverArtifactID != "" {
		t.Fatal("changed background retained stale system cover")
	}

	// Frozen collections still resolve the original image after a replacement.
	frozen := signedJSONRequest(t, api, now, "alice", "GET", mediaPath, nil)
	if !bytes.Equal(media.Body.Bytes(), frozen.Body.Bytes()) {
		t.Fatal("frozen collection changed")
	}

}
