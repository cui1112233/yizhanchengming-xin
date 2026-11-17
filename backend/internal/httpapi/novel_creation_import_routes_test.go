package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"net/http/httptest"
	"os"
	"qiantie/backend/internal/novelcreation"
	"regexp"
	"strings"
	"testing"
	"time"
)

type chapterImportRejectPipeline struct{}

func TestNovelImportReadHTTPAccountScope(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	style, err := svc.SaveStyleProfile(context.Background(), "alice", novelcreation.StyleProfile{Title: "文风甲", Content: "不应在列表公开的完整文风"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.SaveStyleProfile(context.Background(), "bob", novelcreation.StyleProfile{Title: "其他账号文风", Content: "其他文风"}); err != nil {
		t.Fatal(err)
	}
	styles := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/novel-creation/style-profiles", nil)
	if styles.Code != 200 || !strings.Contains(styles.Body.String(), style.ID) || strings.Contains(styles.Body.String(), "不应在列表") || strings.Contains(styles.Body.String(), "其他账号") {
		t.Fatalf("owner style metadata=%d %s", styles.Code, styles.Body.String())
	}
	input := map[string]any{"title": "无会话导入书", "chapters": []map[string]any{{"title": "开场", "content": "正式正文"}}, "knowledgeBookIds": []string{style.ID}, "idempotencyKey": "read-contract"}
	res := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	var out novelcreation.ImportChaptersResult
	if res.Code != 201 || json.Unmarshal(res.Body.Bytes(), &out) != nil {
		t.Fatalf("import=%d %s", res.Code, res.Body.String())
	}
	url := "/api/novel-creation/projects/" + out.Book.ID + "/chapters"
	got := signedJSONRequest(t, mux, now, "alice", http.MethodGet, url, nil)
	if got.Code != 200 || !strings.Contains(got.Body.String(), "正式正文") {
		t.Fatalf("read=%d %s", got.Code, got.Body.String())
	}
	for _, id := range []string{out.Book.ID, "missing"} {
		foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/novel-creation/projects/"+id+"/chapters", nil)
		if foreign.Code != 404 {
			t.Fatalf("foreign read=%d %s", foreign.Code, foreign.Body.String())
		}
	}
	sessions, err := svc.ListSessions(context.Background(), "alice", out.Book.ID)
	if err != nil || len(sessions) != 0 {
		t.Fatalf("read created sessions: %+v %v", sessions, err)
	}
}

// This opt-in case only reads an already imported fixture; it performs no DDL or DML.
func TestNovelImportReadHTTPRealMySQL(t *testing.T) {
	raw := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_DSN")
	if raw == "" {
		t.Skip("opt-in owned MySQL read acceptance")
	}
	c, err := mysql.ParseDSN(raw)
	schema, addr := os.Getenv("QIANTIE_TEST_NOVEL_SCHEMA"), os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_ADDR")
	bookID := os.Getenv("QIANTIE_TEST_NOVEL_READ_BOOK")
	if err != nil || !regexp.MustCompile(`^novelparity20261117[a-f0-9]{12}$`).MatchString(schema) || !regexp.MustCompile(`^127\.0\.0\.1:[0-9]{4,5}$`).MatchString(addr) || strings.HasSuffix(addr, ":3306") || c.DBName != schema || c.Addr != addr || c.Net != "tcp" || c.User != "np"+strings.TrimPrefix(schema, "novelparity20261117") || !regexp.MustCompile(`^project-[a-f0-9]+$`).MatchString(bookID) {
		t.Fatal("refusing non-owned MySQL read target")
	}
	db, err := sql.Open("mysql", raw)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var actual, version, nonce string
	if err = db.QueryRow(`SELECT DATABASE(),VERSION()`).Scan(&actual, &version); err != nil || actual != schema || !strings.HasPrefix(version, "8.4.") {
		t.Fatal("owned MySQL identity changed")
	}
	if err = db.QueryRow(`SELECT nonce FROM novel_parity_fixture_owner WHERE task_id='novel-parity-task8'`).Scan(&nonce); err != nil || len(nonce) != 32 || nonce != os.Getenv("QIANTIE_TEST_NOVEL_NONCE") {
		t.Fatal("owner nonce mismatch")
	}
	var owner string
	if err = db.QueryRow(`SELECT owner_username FROM novel_projects WHERE project_id=? AND deleted_at IS NULL`, bookID).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	var expected int
	if err = db.QueryRow(`SELECT COUNT(*) FROM novel_chapters WHERE owner_username=? AND project_id=?`, owner, bookID).Scan(&expected); err != nil || expected < 1 {
		t.Fatalf("missing persisted fixture chapters: %d %v", expected, err)
	}
	svc := novelcreation.NewService(novelcreation.NewMySQLStore(db), novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	url := "/api/novel-creation/projects/" + bookID + "/chapters"
	read := signedJSONRequest(t, mux, now, owner, http.MethodGet, url, nil)
	var got struct {
		Chapters []novelcreation.Chapter `json:"chapters"`
	}
	if read.Code != 200 || json.Unmarshal(read.Body.Bytes(), &got) != nil || len(got.Chapters) != expected || got.Chapters[0].Content == "" {
		t.Fatalf("persisted chapters=%d %s", read.Code, read.Body.String())
	}
	if denied := signedJSONRequest(t, mux, now, "foreign-read-owner", http.MethodGet, url, nil); denied.Code != 404 {
		t.Fatalf("foreign project read=%d", denied.Code)
	}
	var styleOwner, styleID string
	if err = db.QueryRow(`SELECT owner_username, style_profile_id FROM novel_style_profiles ORDER BY owner_username, style_profile_id LIMIT 1`).Scan(&styleOwner, &styleID); err != nil {
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
		styleOwner = owner
	}
	var styleCount int
	if err = db.QueryRow(`SELECT COUNT(*) FROM novel_style_profiles WHERE owner_username=?`, styleOwner).Scan(&styleCount); err != nil {
		t.Fatal(err)
	}
	styles := signedJSONRequest(t, mux, now, styleOwner, http.MethodGet, "/api/novel-creation/style-profiles", nil)
	var metadata struct {
		Styles []novelcreation.StyleProfileSummary `json:"styles"`
	}
	if styles.Code != 200 || json.Unmarshal(styles.Body.Bytes(), &metadata) != nil || len(metadata.Styles) != styleCount || (styleID != "" && !strings.Contains(styles.Body.String(), styleID)) || strings.Contains(styles.Body.String(), `"content"`) {
		t.Fatalf("style metadata=%d %s", styles.Code, styles.Body.String())
	}
	other := signedJSONRequest(t, mux, now, "foreign-read-owner", http.MethodGet, "/api/novel-creation/style-profiles", nil)
	if other.Code != 200 || (styleID != "" && strings.Contains(other.Body.String(), styleID)) {
		t.Fatalf("foreign style read=%d %s", other.Code, other.Body.String())
	}
	t.Logf("GET-only real MySQL: book=%s persisted chapters=%d, style metadata=%d, foreign-owner rejection verified; no model/session/write", bookID, expected, styleCount)
}

func (chapterImportRejectPipeline) Run(context.Context, novelcreation.RunContext, novelcreation.EventSink) (novelcreation.ChapterBundle, error) {
	panic("chapter import must never invoke generation")
}

func TestNovelChapterImportHTTPRejectsTrailingPayload(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	req := httptest.NewRequest(http.MethodPost, "/api/novel-creation/chapter-imports", strings.NewReader(`{"title":"书","chapters":[{"content":"正文"}],"idempotencyKey":"trailing"} {"owner":"bob"}`))
	SignBridgeRequest(req, "alice", false, now, "secret")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != 400 {
		t.Fatalf("trailing unsupported payload accepted: %d %s", res.Code, res.Body.String())
	}
	books, e := svc.ListProjects(context.Background(), "alice")
	if e != nil || len(books) != 0 {
		t.Fatalf("rejected payload wrote books: %+v %v", books, e)
	}
}

func TestNovelChapterImportHTTPPersistenceAndReplay(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	input := map[string]any{"title": "导入作品", "chapters": []map[string]any{{"title": "开场", "content": "雨停了。", "wordCount": 999}, {"content": "船靠岸。"}}, "knowledgeBookIds": []string{}, "idempotencyKey": "import-1"}
	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if first.Code != http.StatusCreated {
		t.Fatalf("import=%d %s", first.Code, first.Body.String())
	}
	var got struct {
		Book     novelcreation.Project
		Chapters []novelcreation.Chapter
		Message  string
	}
	if err := json.Unmarshal(first.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Message != "章节已导入" || len(got.Chapters) != 2 || got.Book.ModelID != "" || got.Chapters[0].WordCount == 999 {
		t.Fatalf("%+v", got)
	}
	replay := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if replay.Code != http.StatusCreated || replay.Body.String() != first.Body.String() {
		t.Fatalf("replay=%d %s", replay.Code, replay.Body.String())
	}
	input["title"] = "冲突"
	changed := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if changed.Code != http.StatusConflict {
		t.Fatalf("changed=%d %s", changed.Code, changed.Body.String())
	}
	input["bookId"] = got.Book.ID
	input["idempotencyKey"] = "other"
	delete(input, "title")
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
	input["unsupported"] = true
	bad := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("unknown=%d", bad.Code)
	}
}

func TestNovelChapterImportHTTPRealMySQL(t *testing.T) {
	raw := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_DSN")
	if raw == "" {
		t.Skip("opt-in owned MySQL HTTP acceptance")
	}
	c, e := mysql.ParseDSN(raw)
	schema := os.Getenv("QIANTIE_TEST_NOVEL_SCHEMA")
	addr := os.Getenv("QIANTIE_TEST_NOVEL_MYSQL_ADDR")
	if e != nil || !regexp.MustCompile(`^novelparity20261117[a-f0-9]{12}$`).MatchString(schema) || !regexp.MustCompile(`^127\.0\.0\.1:[0-9]{4,5}$`).MatchString(addr) || strings.HasSuffix(addr, ":3306") || c.DBName != schema || c.Addr != addr || c.Net != "tcp" || c.User != "np"+strings.TrimPrefix(schema, "novelparity20261117") {
		t.Fatal("refusing non-owned MySQL target")
	}
	db, e := sql.Open("mysql", raw)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	var actual, version, nonce string
	if e = db.QueryRow(`SELECT DATABASE(),VERSION()`).Scan(&actual, &version); e != nil || actual != schema || !strings.HasPrefix(version, "8.4.") {
		t.Fatal("owned MySQL identity changed")
	}
	if e = db.QueryRow(`SELECT nonce FROM novel_parity_fixture_owner WHERE task_id='novel-parity-task8'`).Scan(&nonce); e != nil || len(nonce) != 32 || nonce != os.Getenv("QIANTIE_TEST_NOVEL_NONCE") {
		t.Fatal("owner nonce mismatch")
	}
	owner := "http-import-" + time.Now().UTC().Format("150405.000000000")
	st := novelcreation.NewMySQLStore(db)
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	input := map[string]any{"title": "HTTP实库导入", "idempotencyKey": "http-real", "knowledgeBookIds": []string{}, "chapters": []map[string]any{{"title": "第十章 渡口", "content": "雨停了。"}, {"content": "故人归来。"}}}
	res := signedJSONRequest(t, mux, now, owner, http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if res.Code != http.StatusCreated {
		t.Fatalf("status=%d %s", res.Code, res.Body.String())
	}
	var out novelcreation.ImportChaptersResult
	if e = json.Unmarshal(res.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	svc = novelcreation.NewService(novelcreation.NewMySQLStore(db), novelcreation.NewMemoryRuntime(), chapterImportRejectPipeline{})
	mux = NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	replay := signedJSONRequest(t, mux, now, owner, http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if replay.Code != 201 || replay.Body.String() != res.Body.String() {
		t.Fatalf("replay=%d %s", replay.Code, replay.Body.String())
	}
	for _, format := range []string{"txt", "md"} {
		doc, e := svc.ExportProject(context.Background(), owner, out.Book.ID, format)
		if e != nil || !strings.Contains(doc.Content, "故人归来。") {
			t.Fatalf("export %s %v", format, e)
		}
	}
	var count int
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_chapter_versions WHERE owner_username=? AND project_id=? AND version_no=1 AND run_id IS NULL`, owner, out.Book.ID).Scan(&count); e != nil || count != 2 {
		t.Fatalf("versions=%d %v", count, e)
	}
	if e = db.QueryRow(`SELECT COUNT(*) FROM novel_runs WHERE owner_username=?`, owner).Scan(&count); e != nil || count != 0 {
		t.Fatalf("runs=%d %v", count, e)
	}
	input["bookId"] = out.Book.ID
	input["idempotencyKey"] = "append"
	delete(input, "title")
	appended := signedJSONRequest(t, mux, now, owner, http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if appended.Code != 201 {
		t.Fatalf("append=%d %s", appended.Code, appended.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "foreign-http", http.MethodPost, "/api/novel-creation/chapter-imports", input)
	if foreign.Code != 404 {
		t.Fatalf("foreign=%d", foreign.Code)
	}
	t.Logf("HTTP->MySQL->new-router replay book=%s version1=2 zero-runs TXT/MD verified schema=%s", out.Book.ID, schema)
}
