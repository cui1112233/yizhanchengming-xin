package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestScreenplayInputStylesAndAtomicSource(t *testing.T) {
	now := time.Now()
	path := filepath.Join(t.TempDir(), "presets.json")
	if err := os.WriteFile(path, []byte(`[{"id":"own-style","module":"batch-factory","name":"自有画风","status":"published","body":"后端私有风格提示词","protocolLock":{"slot":"batch.comic-style"}},{"id":"wrong","module":"batch-factory","name":"其他预设","status":"published","body":"不能选","protocolLock":{"slot":"batch.comic-quality"}}]`), 0600); err != nil {
		t.Fatal(err)
	}
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	registerScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now }, prompt.NewCatalog(path))
	styles := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/comic-screenplays/styles", nil)
	if styles.Code != 200 || !containsText(styles.Body.String(), "own-style") || containsText(styles.Body.String(), "后端私有") || containsText(styles.Body.String(), "wrong") {
		t.Fatalf("styles=%d %s", styles.Code, styles.Body.String())
	}
	created := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"id": "p", "title": "漫剧", "sourceText": "  完整原文\n第二行  ", "selectedStyle": "own-style", "styleDesc": "前端伪造"})
	if created.Code != 201 {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	p, err := store.GetProject(context.Background(), "alice", "p")
	if err != nil || p.SourceText != "  完整原文\n第二行  " || p.StyleDesc != "后端私有风格提示词" || p.ArtStyle != "自有画风" {
		t.Fatalf("project=%+v err=%v", p, err)
	}
	if containsText(created.Body.String(), "后端私有风格提示词") {
		t.Fatal("prompt leaked")
	}
	for _, id := range []string{"wrong", "missing"} {
		bad := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"title": "无效画风", "selectedStyle": id})
		if bad.Code != 400 {
			t.Fatalf("id %s: %d %s", id, bad.Code, bad.Body.String())
		}
	}
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/comic-screenplays/projects", map[string]any{"title": "空原文", "sourceText": "  "})
	if empty.Code != 400 {
		t.Fatalf("empty=%d", empty.Code)
	}
	list, _ := store.ListProjects(context.Background(), "alice")
	if len(list) != 1 {
		t.Fatalf("unexpected projects %+v", list)
	}
	updated := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/p", map[string]any{"title": "修改名称", "selectedStyle": "own-style", "styleDesc": "客户端覆盖", "expectedRevision": p.Revision})
	if updated.Code != 200 {
		t.Fatalf("update=%d %s", updated.Code, updated.Body.String())
	}
	p, _ = store.GetProject(context.Background(), "alice", "p")
	if p.StyleDesc != "后端私有风格提示词" || p.SourceText != "  完整原文\n第二行  " {
		t.Fatalf("updated=%+v", p)
	}
	conflict := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/p", map[string]any{"selectedStyle": "own-style", "expectedRevision": p.Revision - 1})
	if conflict.Code != 409 {
		t.Fatalf("conflict=%d %s", conflict.Code, conflict.Body.String())
	}
	if err := os.WriteFile(path, []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	removed := signedJSONRequest(t, mux, now, "alice", http.MethodPatch, "/api/comic-screenplays/projects/p", map[string]any{"selectedStyle": "own-style", "expectedRevision": p.Revision})
	if removed.Code != 400 || !containsText(removed.Body.String(), "重新选择") {
		t.Fatalf("removed=%d %s", removed.Code, removed.Body.String())
	}
}

func TestScreenplayDraftFileParsingDoesNotCreateProjects(t *testing.T) {
	now := time.Now()
	store := screenplay.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterScreenplayRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	var docx bytes.Buffer
	archive := zip.NewWriter(&docx)
	entry, _ := archive.Create("word/document.xml")
	_, _ = entry.Write([]byte(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>DOCX 正文</w:t></w:r></w:p></w:body></w:document>`))
	_ = archive.Close()
	for _, tc := range []struct {
		name     string
		data     []byte
		expected string
		code     int
	}{
		{"story.txt", []byte("\ufeffTXT 正文"), "TXT 正文", 200},
		{"story.docx", docx.Bytes(), "DOCX 正文", 200},
		{"broken.docx", []byte("bad"), "", 400},
		{"empty.txt", []byte("  "), "", 400},
		{"bad.pdf", []byte("正文"), "", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			file, _ := form.CreateFormFile("file", tc.name)
			_, _ = file.Write(tc.data)
			_ = form.Close()
			req := httptest.NewRequest(http.MethodPost, "/api/comic-screenplays/source/parse", &body)
			req.Header.Set("Content-Type", form.FormDataContentType())
			SignBridgeRequest(req, "alice", false, now, "secret")
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.code || (tc.expected != "" && !containsText(rec.Body.String(), tc.expected)) {
				t.Fatalf("parse=%d %s", rec.Code, rec.Body.String())
			}
		})
	}
	projects, _ := store.ListProjects(context.Background(), "alice")
	if len(projects) != 0 {
		t.Fatalf("parse created %d projects", len(projects))
	}
	for _, path := range []string{"/api/comic-screenplays/source/parse", "/api/comic-screenplays/styles"} {
		method := http.MethodGet
		if strings.Contains(path, "parse") {
			method = http.MethodPost
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		if rec.Code != 401 {
			t.Fatalf("unauth %s = %d", path, rec.Code)
		}
	}
}

func TestScreenplayStyleLegacyNamesAndInvalidRemovedPreset(t *testing.T) {
	current := screenplay.Project{ArtStyle: "旧画风", SelectedStyle: "旧画风", StyleDesc: "历史描述"}
	got, err := resolveScreenplayStyle(nil, "旧画风", "旧画风", &current)
	if err != nil || got.StyleDesc != "历史描述" {
		t.Fatalf("legacy=%+v %v", got, err)
	}
	current.SelectedStyle = ""
	got, err = resolveScreenplayStyle(nil, "旧画风", "旧画风", &current)
	if err != nil || got.ArtStyle != "旧画风" || got.StyleDesc != "历史描述" {
		t.Fatalf("name-only legacy=%+v %v", got, err)
	}
	current.SelectedStyle = "batch-comic-style-removed"
	if _, err := resolveScreenplayStyle(nil, current.SelectedStyle, current.ArtStyle, &current); err == nil {
		t.Fatal("removed ID accepted")
	}
}
