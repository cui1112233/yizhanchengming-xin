package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/novelcreation"
	"strings"
	"testing"
	"time"
)

func TestNovelLongImitationTypedRoutesAndRecovery(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, novelcreation.NewMemoryRuntime(), nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	path := "/api/novel-creation/long-imitation"
	input := map[string]any{"idempotencyKey": "imitation", "modelId": "fake-test", "referenceText": strings.Repeat("甲", 100), "targetWordCount": 50000, "targetChapters": 100}
	response := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, input)
	if response.Code != 202 {
		t.Fatalf("create=%d %s", response.Code, response.Body.String())
	}
	var out novelcreation.LongImitationResult
	if e := json.Unmarshal(response.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{path + "/requests/imitation", "/api/novel-creation/projects/" + out.Project.ID + "/long-imitation"} {
		r := signedJSONRequest(t, mux, now, "alice", http.MethodGet, p, nil)
		if r.Code != 200 {
			t.Fatalf("restore=%d %s", r.Code, r.Body.String())
		}
		r = signedJSONRequest(t, mux, now, "bob", http.MethodGet, p, nil)
		if r.Code != 404 {
			t.Fatal("cross-owner read", r.Code)
		}
	}
	input["statistics"] = map[string]any{"fake": true}
	bad := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path, input)
	if bad.Code != 400 {
		t.Fatal("client stats accepted")
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 1 {
		t.Fatal("hidden or duplicate project")
	}
	preview := signedJSONRequest(t, mux, now, "alice", http.MethodPost, path+"/preview", map[string]any{"referenceText": "😀。😀。"})
	if preview.Code != 200 || !strings.Contains(preview.Body.String(), "0.25") {
		t.Fatal("local UTF16 stats absent", preview.Body.String())
	}
}
func TestNovelLongImitationParsePartialFailureAndStrictJSON(t *testing.T) {
	st := novelcreation.NewMemoryStore()
	svc := novelcreation.NewService(st, nil, nil)
	now := time.Now().UTC()
	mux := NewRouter(RouterOptions{BridgeSecret: "secret", Now: func() time.Time { return now }, NovelCreation: svc})
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, _ := writer.CreateFormFile("files", "good.txt")
	part.Write([]byte("第1章 开始\n甲\n第2章 转折\n乙"))
	part, _ = writer.CreateFormFile("files", "bad.docx")
	part.Write([]byte("broken"))
	writer.Close()
	req := httptest.NewRequest("POST", "/api/novel-creation/long-imitation/parse", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(req, "alice", false, now, "secret")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 200 || !strings.Contains(rr.Body.String(), "good.txt") || !strings.Contains(rr.Body.String(), "bad.docx") || !strings.Contains(rr.Body.String(), "failures") {
		t.Fatal("partial parse response", rr.Code, rr.Body.String())
	}
	projects, _ := st.ListProjects(context.Background(), "alice")
	if len(projects) != 0 {
		t.Fatal("parse created hidden project")
	}
	req = httptest.NewRequest("POST", "/api/novel-creation/long-imitation/preview", strings.NewReader(`{"referenceText":"甲"}{"referenceText":"乙"}`))
	req.Header.Set("Content-Type", "application/json")
	SignBridgeRequest(req, "alice", false, now, "secret")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != 400 {
		t.Fatal("trailing JSON accepted", rr.Code)
	}
}
