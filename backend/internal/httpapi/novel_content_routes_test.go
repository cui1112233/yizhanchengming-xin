package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/novelcreation"
	"qiantie/backend/internal/projectretention"
	"testing"
	"time"
)

func TestNovelDocumentUploadBatchAndFileLimits(t *testing.T) {
	s := novelcreation.NewService(novelcreation.NewMemoryStore(), novelcreation.NewMemoryRuntime(), nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	for _, tc := range []struct {
		name        string
		count, size int
	}{{"batch", 9, 1}, {"file", 1, (20 << 20) + 1}} {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			m := multipart.NewWriter(&b)
			for n := 0; n < tc.count; n++ {
				f, _ := m.CreateFormFile("files", "sample.txt")
				io.CopyN(f, repeatByte('x'), int64(tc.size))
			}
			m.Close()
			r := httptest.NewRequest(http.MethodPost, "/api/novel-creation/projects/p/documents/upload", &b)
			r.Header.Set("Content-Type", m.FormDataContentType())
			SignBridgeRequest(r, "alice", false, now, "secret")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != 400 {
				t.Fatalf("%s: %d %s", tc.name, w.Code, w.Body.String())
			}
		})
	}
}

type repeatByte byte

func (r repeatByte) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = byte(r)
	}
	return len(b), nil
}

func TestNovelContentHTTPPasteSelectionOwnershipArchive(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	s := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	p, e := s.CreateProject(context.Background(), "alice", novelcreation.CreateProjectInput{Title: "作品", Genre: "幻想", ModelID: "fake"})
	if e != nil {
		t.Fatal(e)
	}
	path := "/api/novel-creation/projects/" + p.ID
	availability := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/novel-creation/documents/availability", nil)
	if availability.Code != 200 || !containsText(availability.Body.String(), `"uploadAvailable":false`) {
		t.Fatal(availability.Body.String())
	}
	empty := signedJSONRequest(t, mux, now, "alice", http.MethodGet, path+"/export?format=txt", nil)
	if empty.Code == 200 || !containsText(empty.Body.String(), "暂无可导出正文") {
		t.Fatal(empty.Body.String())
	}
	paste := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/documents/paste", map[string]any{"title": "资料", "content": "第一章 初见\n你好。\n第二章 重逢\n世界。"})
	if paste.Code != 200 {
		t.Fatal(paste.Body.String())
	}
	var body struct {
		Document novelcreation.DocumentJob `json:"document"`
	}
	if e = json.Unmarshal(paste.Body.Bytes(), &body); e != nil {
		t.Fatal(e)
	}
	j := body.Document
	if j.Status != "ready" || len(j.Chapters) != 2 {
		t.Fatal(paste.Body.String())
	}
	selection := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/documents/"+j.ID+"/selection", map[string]any{"expectedVersion": j.Version, "chapterIds": []string{j.Chapters[1].ID}})
	if selection.Code != 200 || !containsText(selection.Body.String(), "第二章") {
		t.Fatal(selection.Body.String())
	}
	foreign := signedJSONRequest(t, mux, now, "bob", http.MethodGet, path+"/documents/"+j.ID, nil)
	if foreign.Code == 200 {
		t.Fatal("cross-owner document exposed")
	}
	archived := signedJSONRequest(t, mux, now, "alice", http.MethodDelete, path, nil)
	if archived.Code != 200 {
		t.Fatal(archived.Body.String())
	}
	blocked := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/documents/paste", map[string]any{"content": "late"})
	if blocked.Code == 200 {
		t.Fatal("write after archive")
	}
	restored := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/restore", nil)
	if restored.Code != 200 {
		t.Fatal(restored.Body.String())
	}
}

func TestNovelContentPurgingRestoreConflictHTTP(t *testing.T) {
	s := novelcreation.NewService(novelcreation.NewMemoryStore(), novelcreation.NewMemoryRuntime(), nil)
	s.ConfigureRetention(nil, func(context.Context, string, string) error { return projectretention.ErrPurging })
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: s})
	w := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/novel-creation/projects/p/restore", nil)
	if w.Code != http.StatusConflict || !containsText(w.Body.String(), "PROJECT_PURGING") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
