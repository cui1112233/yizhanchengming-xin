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
