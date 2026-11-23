package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"qiantie/backend/internal/batchfactorytext"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/storage"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type textDraftTestStore struct {
	*batchfactoryv11.MemoryStore
	drafts *batchfactoryv11.MySQLStore
}

func (s textDraftTestStore) GetDraft(ctx context.Context, owner, key, kind, scope string) (batchfactoryv11.Draft, error) {
	return s.drafts.GetDraft(ctx, owner, key, kind, scope)
}
func (s textDraftTestStore) SaveDraft(ctx context.Context, owner string, d batchfactoryv11.Draft) (batchfactoryv11.Draft, error) {
	return s.drafts.SaveDraft(ctx, owner, d)
}
func (s textDraftTestStore) SaveDraftTx(ctx context.Context, tx *sql.Tx, owner string, d batchfactoryv11.Draft) (batchfactoryv11.Draft, error) {
	return s.drafts.SaveDraftTx(ctx, tx, owner, d)
}

type pausingTextBatchRead struct {
	textDraftTestStore
	reads           atomic.Int32
	entered, resume chan struct{}
}

func (s *pausingTextBatchRead) GetDraft(ctx context.Context, owner, key, kind, scope string) (batchfactoryv11.Draft, error) {
	b, err := s.textDraftTestStore.GetDraft(ctx, owner, key, kind, scope)
	// Pause after the request reads its production source, before pagination.
	if s.reads.Add(1) == 1 {
		close(s.entered)
		select {
		case <-s.resume:
		case <-ctx.Done():
			return b, ctx.Err()
		}
	}
	return b, err
}

type textSQLSourceReader interface {
	LockProductionSource(context.Context, *sql.Tx, string, string, string) (batchfactoryv11.Book, string, error)
}

// Other batch calls use the signed fixture while the publication fence reads
// the same MySQL record/draft tables and runs the same locking queries as production.
type textSQLSourceStore struct {
	textDraftTestStore
	textSQLSourceReader
}
type pausingTextPublication struct {
	batchfactorytext.Repository
	entered, resume chan struct{}
}

func (p *pausingTextPublication) PreparePages(ctx context.Context, s batchfactorytext.Scope, v batchfactorytext.VersionInput, u batchfactorytext.PaginationUpdate) (int64, error) {
	close(p.entered)
	select {
	case <-p.resume:
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	return p.Repository.PreparePages(ctx, s, v, u)
}

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
	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS batch_factory_v12_drafts(owner_username VARCHAR(191),draft_key VARCHAR(100),kind VARCHAR(64),scope VARCHAR(64),content LONGTEXT,revision BIGINT NOT NULL,updated_at DATETIME(6),PRIMARY KEY(owner_username,draft_key,kind,scope)) ENGINE=InnoDB"); err != nil {
		t.Fatal(err)
	}
	s := textDraftTestStore{batchfactoryv11.NewMemoryStore(), batchfactoryv11.NewMySQLStore(db)}
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
	rules := novelfetchworkshop.NewMemoryStore()
	api := NewRouter(RouterOptions{Slice: 1, Store: s, NovelFetchStore: rules, TextRepository: batchfactorytext.NewMySQLRepository(db), LocalArtifacts: files, BridgeSecret: "secret", Now: func() time.Time { return now }})
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
	runPaused := func(opts RouterOptions, entered chan struct{}, resume chan struct{}, method, path string, body any, change func()) {
		t.Helper()
		opts.Slice = 1
		opts.LocalArtifacts = files
		opts.BridgeSecret = "secret"
		opts.Now = func() time.Time { return now }
		release := sync.OnceFunc(func() { close(resume) })
		defer release()
		api := NewRouter(opts)
		finished := make(chan *httptest.ResponseRecorder, 1)
		go func() { finished <- signedJSONRequest(t, api, now, "alice", method, base+path, body) }()
		select {
		case <-entered:
		case <-time.After(15 * time.Second):
			t.Fatal("slow prepare never reached the synchronization point")
		}
		change()
		release()
		select {
		case got := <-finished:
			if got.Code != 409 {
				t.Fatalf("stale prepare got %d: %s", got.Code, got.Body.String())
			}
		case <-time.After(15 * time.Second):
			t.Fatal("stale prepare never finished")
		}
	}
	runSlow := func(opts RouterOptions, entered, resume chan struct{}, change func()) {
		runPaused(opts, entered, resume, "POST", "prepare", map[string]any{}, change)
	}
	entered, resume := make(chan struct{}), make(chan struct{})
	raceStore := &pausingTextBatchRead{textDraftTestStore: s, entered: entered, resume: resume}
	runSlow(RouterOptions{Store: raceStore, TextRepository: batchfactorytext.NewMySQLRepository(db)}, entered, resume, func() {
		call("POST", "pages/"+pg.ID+"/split", map[string]any{"offset": 6, "expectedRevision": st.Document.Revision}, 200)
	})
	st = get()
	if len(st.Pages) != 2 {
		t.Fatalf("split %+v", st)
	}
	first, second := st.Pages[0], st.Pages[1]
	baseVersion := st.Document.ActiveVersionID
	pagePath := "pages/" + first.ID
	edit := map[string]any{"expectedRevision": first.Revision, "expectedDocumentRevision": st.Document.Revision, "text": "这是手动改写的第一句。"}
	oldPixels := call("POST", pagePath+"/preview", map[string]any{}, 200).Body.Bytes()
	newPixels := call("POST", pagePath+"/preview", edit, 200).Body.Bytes()
	if bytes.Equal(oldPixels, newPixels) || get().Pages[0].Text != first.Text {
		t.Fatal("draft preview must render new text without saving")
	}
	edit["expectedDocumentRevision"] = st.Document.Revision + 1
	call("PATCH", pagePath, edit, 409)
	edit["expectedDocumentRevision"] = st.Document.Revision
	call("PATCH", pagePath, map[string]any{"text": "未说完", "expectedRevision": first.Revision, "expectedDocumentRevision": st.Document.Revision}, 400)
	call("PATCH", pagePath, edit, 200)
	st = get()
	if len(st.Pages) != 2 || st.Pages[0].Text != edit["text"] || st.Pages[1].Text != second.Text || st.Document.OriginalText != original || st.Document.ActiveVersionID == baseVersion || st.Pages[0].EndOffset != st.Pages[1].StartOffset {
		t.Fatalf("edit did not preserve version and boundaries: %+v", st)
	}
	if actual := call("POST", "pages/"+st.Pages[0].ID+"/preview", map[string]any{}, 200).Body.Bytes(); !bytes.Equal(actual, newPixels) {
		t.Fatal("preview and saved pixels differ")
	}
	prior, err := batchfactorytext.NewMySQLRepository(db).GetVersion(context.Background(), batchfactorytext.Scope{Owner: "alice", BatchID: b.ID, BookID: book.ID}, baseVersion)
	if err != nil || prior.ProductionText != original {
		t.Fatalf("original version mutated: %+v %v", prior, err)
	}

	set := func(key string, value any) {
		t.Helper()
		fresh, err := s.GetBatch(context.Background(), "alice", b.ID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.SaveSettings(context.Background(), "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: b.ID}, batchfactoryv11.SettingsUpdate{ExpectedRevision: fresh.Revision, Patch: batchfactoryv11.SettingsPatch{key: mustJSON(value)}})
		if err != nil {
			t.Fatal(err)
		}
	}
	assertText := func(want string) {
		t.Helper()
		curr := get()
		var joined strings.Builder
		for _, p := range curr.Pages {
			joined.WriteString(p.Text)
		}
		if joined.String() != want || curr.Document.OriginalText != original {
			t.Fatalf("production=%q, original=%q, want production=%q original=%q", joined.String(), curr.Document.OriginalText, want, original)
		}
	}
	authored := st.Pages[0].Text + st.Pages[1].Text
	set("sensitiveMode", "system")
	call("POST", "prepare", map[string]any{}, 200)
	assertText(authored)
	if err = rules.PutConfig(context.Background(), "alice", map[string]any{"knowledgeLibraries": map[string]any{"layout_rules": map[string]any{"sensitive_rules": []any{map[string]any{"enabled": true, "find": "手动改写", "replace": "系统调整", "apply_to_original": true}}}}}); err != nil {
		t.Fatal(err)
	}
	call("POST", "prepare", map[string]any{}, 200)
	assertText(strings.ReplaceAll(authored, "手动改写", "系统调整"))
	processed := get()
	shown := processed.Pages[0]
	call("PATCH", "pages/"+shown.ID, map[string]any{"expectedRevision": shown.Revision, "expectedDocumentRevision": processed.Document.Revision, "text": shown.Text, "background": map[string]string{"kind": "color", "color": "#101010"}}, 200)
	set("sensitiveMode", "off")
	call("POST", "prepare", map[string]any{}, 200)
	assertText(authored)
	// A content-changing policy can repaginate. Recreate a manual split before
	// checking that layout-only updates retain it.
	st = get()
	if len(st.Pages) == 1 {
		call("POST", "pages/"+st.Pages[0].ID+"/split", map[string]any{"offset": len([]rune(edit["text"].(string))), "expectedRevision": st.Document.Revision}, 200)
	}
	st = get()
	call("PATCH", "pages/"+st.Pages[1].ID, map[string]any{"expectedRevision": st.Pages[1].Revision, "layout": map[string]string{"textColor": "#225599"}}, 200)
	for _, p := range get().Pages {
		call("POST", "pages/"+p.ID+"/render", map[string]any{"expectedRevision": p.Revision}, 201)
	}
	oldPages := get().Pages
	set("textColor", "#006688")
	call("POST", "prepare", map[string]any{}, 200)
	st = get()
	if len(st.Pages) != 2 || st.Pages[0].ID != oldPages[0].ID || st.Pages[1].ID != oldPages[1].ID || st.Pages[0].Text != edit["text"] {
		t.Fatalf("color lost manual pages: %+v", st)
	}
	if st.Pages[0].ActiveArtifactID != "" || oldPages[1].ActiveArtifactID == "" || st.Pages[1].ActiveArtifactID != oldPages[1].ActiveArtifactID {
		t.Fatal("color change invalidated incorrect pages")
	}
	set("textFontSize", 44)
	call("POST", "prepare", map[string]any{}, 200)
	st = get()
	if len(st.Pages) != 2 || st.Pages[0].ID != oldPages[0].ID {
		t.Fatal("geometry lost fitting manual boundaries")
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
	call("POST", "pages/"+pg.ID+"/preview", map[string]any{"text": "编辑锁定页。"}, 409)
	call("PATCH", "pages/"+pg.ID, map[string]any{"text": "编辑锁定页。", "expectedRevision": st.Pages[0].Revision, "expectedDocumentRevision": st.Document.Revision}, 409)
	lockedArtifact := st.Pages[0].ActiveArtifactID
	set("textBackground", map[string]string{"type": "color", "color": "#111111"})
	call("POST", "prepare", map[string]any{}, 200)
	st = get()
	if !st.Pages[0].Locked || st.Pages[0].ActiveArtifactID != lockedArtifact {
		t.Fatal("color changed a locked upload")
	}
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

	// Pause after pinning the material and rendering layout. Updating only the
	// upstream working draft must still prevent the old publication.
	entered, resume = make(chan struct{}), make(chan struct{})
	paused := &pausingTextPublication{Repository: batchfactorytext.NewMySQLRepository(db), entered: entered, resume: resume}
	head := get().Document
	runSlow(RouterOptions{Store: s, TextRepository: paused}, entered, resume, func() {
		_, err = s.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: b.ID, Content: "这是另一个新的工作稿。"})
		if err != nil {
			t.Fatal(err)
		}
	})
	if get().Document.Revision != head.Revision {
		t.Fatal("stale request changed the document")
	}
	call("POST", "prepare", map[string]any{}, 200)
	if get().Pages[0].Text != "这是另一个新的工作稿。" {
		t.Fatal("new working draft was not activated")
	}
	for _, ddl := range []string{
		"CREATE TABLE IF NOT EXISTS batch_factory_v12_book_records(book_id VARCHAR(64) PRIMARY KEY, source_text LONGTEXT,txt_text LONGTEXT,source_metadata_json JSON) ENGINE=InnoDB",
	} {
		if _, err = db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	fresh, err = s.GetBatch(context.Background(), "alice", b.ID)
	if err != nil {
		t.Fatal(err)
	}
	verified := fresh.Books[0]
	meta, err := json.Marshal(verified.SourceMetadata)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO batch_factory_v12_book_records(book_id,source_text,txt_text,source_metadata_json) VALUES(?,?,?,?)", book.ID, verified.SourceText, verified.TxtText, meta); err != nil {
		t.Fatal(err)
	}

	sqlSource := textSQLSourceStore{s, batchfactoryv11.NewMySQLStore(db)}
	opts := RouterOptions{Slice: 1, Store: sqlSource, TextRepository: batchfactorytext.NewMySQLRepository(db), LocalArtifacts: files, BridgeSecret: "secret", Now: func() time.Time { return now }}
	if rr := signedJSONRequest(t, NewRouter(opts), now, "alice", "POST", base+"prepare", map[string]any{}); rr.Code != 200 {
		t.Fatalf("SQL source fence rejected current source: %d %s", rr.Code, rr.Body.String())
	}
	head = get().Document
	entered, resume = make(chan struct{}), make(chan struct{})
	opts.TextRepository = &pausingTextPublication{Repository: batchfactorytext.NewMySQLRepository(db), entered: entered, resume: resume}
	rawEdit := get().Pages[0]
	runPaused(opts, entered, resume, "PATCH", "pages/"+rawEdit.ID, map[string]any{"text": "过期编辑。", "expectedRevision": rawEdit.Revision, "expectedDocumentRevision": head.Revision}, func() {
		if _, err = db.Exec("UPDATE batch_factory_v12_book_records SET txt_text='并发更新的原文。' WHERE book_id=?", book.ID); err != nil {
			t.Fatal(err)
		}
	})
	unchanged, err := s.GetDraft(context.Background(), "alice", "working-front:"+book.ID, "working-front-content", b.ID)
	if err != nil || unchanged.Content != verified.SourceText || get().Document.Revision != head.Revision {
		t.Fatalf("stale raw edit changed author or document: %v", err)
	}
	if _, err = db.Exec("UPDATE batch_factory_v12_book_records SET txt_text=? WHERE book_id=?", verified.TxtText, book.ID); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"UPDATE batch_factory_v12_drafts SET content='并发改稿。' WHERE scope=?", "UPDATE batch_factory_v12_book_records SET source_text='并发改稿。' WHERE book_id=?"} {
		entered, resume = make(chan struct{}), make(chan struct{})
		opts.TextRepository = &pausingTextPublication{Repository: batchfactorytext.NewMySQLRepository(db), entered: entered, resume: resume}
		runSlow(opts, entered, resume, func() {
			key := book.ID
			if strings.Contains(q, " WHERE scope=") {
				key = b.ID
			}
			if _, err = db.Exec(q, key); err != nil {
				t.Fatal(err)
			}
		})
		if get().Document.Revision != head.Revision {
			t.Fatal("stale SQL source changed the document")
		}
		if _, err = db.Exec("UPDATE batch_factory_v12_drafts SET content=? WHERE scope=?", verified.SourceText, b.ID); err != nil {
			t.Fatal(err)
		}
	}
	longDraft := strings.Repeat("他点头。", 20)
	if _, err = s.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: b.ID, Content: longDraft}); err != nil {
		t.Fatal(err)
	}
	call("POST", "prepare", map[string]any{}, 200)
	beforeResize := get()
	if len(beforeResize.Pages) != 1 {
		t.Fatalf("geometry fixture needs one page, got %d", len(beforeResize.Pages))
	}
	set("textFontSize", 180)
	call("POST", "prepare", map[string]any{}, 409)
	if get().Document.Revision != beforeResize.Document.Revision {
		t.Fatal("overflow modified page boundaries")
	}
	call("POST", "prepare", map[string]any{"reflow": true, "expectedRevision": beforeResize.Document.Revision}, 200)
	resized := get()
	var rebuilt strings.Builder
	for _, page := range resized.Pages {
		rebuilt.WriteString(page.Text)
	}
	if len(resized.Pages) < 2 || rebuilt.String() != longDraft {
		t.Fatal("explicit reflow lost text or did not repaginate")
	}
	set("textFontSize", 44)
	const spaced = "你来了。\n好。\n尾页。"
	if _, err = s.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: b.ID, Content: spaced}); err != nil {
		t.Fatal(err)
	}
	call("POST", "prepare", map[string]any{}, 200)
	d := get().Document
	drafts, err := batchfactorytext.Paginate(context.Background(), batchfactorytext.PaginationInput{Text: spaced}, func(line string, _ batchfactorytext.LayoutConfig) (bool, float64, error) {
		n := len([]rune(line))
		return n <= 4, float64(n), nil
	})
	if err != nil || len(drafts) < 2 || drafts[0].Text != "你来了。" {
		t.Fatalf("separator fixture: %+v %v", drafts, err)
	}
	scope := batchfactorytext.Scope{Owner: "alice", BookID: book.ID, BatchID: b.ID}
	if _, err = batchfactorytext.NewMySQLRepository(db).ReplaceSystemPages(context.Background(), scope, batchfactorytext.PaginationUpdate{VersionID: d.ActiveVersionID, ExpectedRevision: d.Revision, Pages: drafts}); err != nil {
		t.Fatal(err)
	}
	withSeparators := get()
	spill := withSeparators.Pages[1]
	call("POST", "pages/"+spill.ID+"/split", map[string]any{"expectedRevision": withSeparators.Document.Revision, "offset": spill.StartOffset + 1}, 400)
	last := withSeparators.Pages[len(withSeparators.Pages)-1]
	updateLast := strings.Replace(last.Text, "尾页", "末页", 1)
	call("PATCH", "pages/"+last.ID, map[string]any{"text": updateLast, "expectedRevision": last.Revision, "expectedDocumentRevision": withSeparators.Document.Revision}, 200)
	assertText(strings.Replace(spaced, "尾页", "末页", 1))
	// Simulate a manual version created by the older editor, which saved pages
	// without persisting the new author baseline in the working draft.
	manual := get()
	legacyConfig := map[string]any{}
	if err = json.Unmarshal(manual.Document.ConfigSnapshot, &legacyConfig); err != nil {
		t.Fatal(err)
	}
	delete(legacyConfig, "authorDraftKey")
	cfg, err := json.Marshal(legacyConfig)
	if err != nil {
		t.Fatal(err)
	}
	legacyPages := make([]batchfactorytext.PageDraft, 0, len(manual.Pages))
	for _, p := range manual.Pages {
		legacyPages = append(legacyPages, p.PageDraft)
	}
	repo := batchfactorytext.NewMySQLRepository(db)
	if _, err = repo.ReplaceSystemPages(context.Background(), scope, batchfactorytext.PaginationUpdate{VersionID: manual.Document.ActiveVersionID, ExpectedRevision: manual.Document.Revision, Pages: legacyPages, ConfigSnapshot: cfg}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SaveDraft(context.Background(), "alice", batchfactoryv11.Draft{Key: "working-front:" + book.ID, Kind: "working-front-content", Scope: b.ID, Content: spaced}); err != nil {
		t.Fatal(err)
	}
	set("sensitiveMode", "system")
	call("POST", "prepare", map[string]any{}, 200)
	assertText(strings.Replace(spaced, "尾页", "末页", 1))
	// Frozen collections still resolve the original image after a replacement.
	frozen := signedJSONRequest(t, api, now, "alice", "GET", mediaPath, nil)
	if !bytes.Equal(media.Body.Bytes(), frozen.Body.Bytes()) {
		t.Fatal("frozen collection changed")
	}

}
