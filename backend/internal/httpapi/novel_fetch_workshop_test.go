package httpapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/novelfetchworkshop"
)

func signNovelFetchRequest(req *http.Request, username string, isOwner bool, now time.Time, secret string) {
	issuedAt := strconv.FormatInt(now.Unix(), 10)
	owner := strconv.FormatBool(isOwner)
	payload := username + "\n" + issuedAt + "\n" + owner + "\n" + req.Method + "\n" + req.URL.Path
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	req.Header.Set(HeaderUsername, username)
	req.Header.Set(HeaderIsOwner, owner)
	req.Header.Set(HeaderIssuedAt, issuedAt)
	req.Header.Set(HeaderSignature, hex.EncodeToString(mac.Sum(nil)))
}

func doNovelFetchRequest(t *testing.T, handler http.Handler, method, path string, body any, now time.Time, secret string) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	signNovelFetchRequest(req, "tester", true, now, secret)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestNovelFetchWorkshopBridgeRoundTrip(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{
		BridgeSecret:    secret,
		Now:             func() time.Time { return now },
		NovelFetchStore: store,
	})

	put := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/tasks/book-1", map[string]any{
		"meta":        map[string]any{"bookName": "测试小说", "targetVersions": []any{"original", "ai1", "ai3"}},
		"original":    "正文",
		"originalRaw": "原始正文",
		"versions":    map[string]any{"ai1": "AI1", "ai3": "AI3"},
		"logs":        []any{map[string]any{"event": "saved"}},
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var document map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	if document["bookId"] != "book-1" || document["original"] != "正文" {
		t.Fatalf("unexpected document: %#v", document)
	}
	versions := document["versions"].(map[string]any)
	if versions["ai1"] != "AI1" || versions["ai3"] != "AI3" {
		t.Fatalf("sparse versions lost: %#v", versions)
	}

	list := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks", nil, now, secret)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var listed struct {
		Tasks []novelfetchworkshop.Document `json:"tasks"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].BookID != "book-1" {
		t.Fatalf("unexpected tasks: %#v", listed.Tasks)
	}

	configPut := doNovelFetchRequest(t, handler, http.MethodPut, "/api/novel-fetch-workshop/config", map[string]any{
		"settings": map[string]any{"fetch": map[string]any{"timeout_seconds": 9}},
	}, now, secret)
	if configPut.Code != http.StatusOK {
		t.Fatalf("config put status=%d body=%s", configPut.Code, configPut.Body.String())
	}
	configGet := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/config", nil, now, secret)
	if configGet.Code != http.StatusOK {
		t.Fatalf("config get status=%d body=%s", configGet.Code, configGet.Body.String())
	}
	var config struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(configGet.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if config.Settings["fetch"].(map[string]any)["timeout_seconds"] != float64(9) {
		t.Fatalf("config did not round trip: %#v", config.Settings)
	}

	deleted := doNovelFetchRequest(t, handler, http.MethodDelete, "/api/novel-fetch-workshop/tasks", map[string]any{"ids": []string{"book-1"}}, now, secret)
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	missing := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/tasks/book-1", nil, now, secret)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestRunAuditPersistsSelectedAndResolvedModelWithoutCredential(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	secret := "test-secret"
	handler := NewRouter(RouterOptions{BridgeSecret: secret, Now: func() time.Time { return now }, NovelFetchStore: novelfetchworkshop.NewMemoryStore()})

	put := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/runs", map[string]any{
		"runId": "run-1", "bookId": "book-1", "stage": "rewrite", "status": "succeeded", "attempts": 1,
		"textModelId": "gemini-3", "modelId": "gemini-3", "modelDisplayName": "Gemini 3", "api_key": "must-not-persist",
	}, now, secret)
	if put.Code != http.StatusOK {
		t.Fatalf("put status=%d body=%s", put.Code, put.Body.String())
	}

	get := doNovelFetchRequest(t, handler, http.MethodGet, "/api/novel-fetch-workshop/runs?bookId=book-1", nil, now, secret)
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	if bytes.Contains(get.Body.Bytes(), []byte("must-not-persist")) || bytes.Contains(bytes.ToLower(get.Body.Bytes()), []byte("api_key")) {
		t.Fatalf("audit leaked secret: %s", get.Body.String())
	}
	var response struct {
		Runs []novelfetchworkshop.RunRecord `json:"runs"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Runs) != 1 || response.Runs[0].TextModelID != "gemini-3" || response.Runs[0].ModelID != "gemini-3" {
		t.Fatalf("unexpected audit records: %#v", response.Runs)
	}
}

func TestNovelFetchWorkshopRejectsUnsignedRequest(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	handler := NewRouter(RouterOptions{BridgeSecret: "test-secret", NovelFetchStore: store})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestNovelFetchWorkshopAcceptsBrowserSessionToken(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	store := novelfetchworkshop.NewMemoryStore()
	authStore := &authTestStore{
		account: auth.Account{Username: "alice", Active: true, IsOwner: true},
		session: auth.Session{Token: "browser-token", Username: "alice", ExpiresAt: now.Add(time.Hour)},
	}
	handler := NewRouter(RouterOptions{
		BridgeSecret:    "unused-bridge-secret",
		Now:             func() time.Time { return now },
		AuthStore:       authStore,
		NovelFetchStore: store,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/novel-fetch-workshop/tasks", nil)
	req.Header.Set("Authorization", "Bearer browser-token")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("browser session status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestNovelFetchMemoryStoreIsOwnerScoped(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", novelfetchworkshop.Document{BookID: "same", Original: "alice"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutDocument(context.Background(), "bob", novelfetchworkshop.Document{BookID: "same", Original: "bob"}); err != nil {
		t.Fatal(err)
	}
	a, _ := store.GetDocument(context.Background(), "alice", "same")
	b, _ := store.GetDocument(context.Background(), "bob", "same")
	if a.Original != "alice" || b.Original != "bob" {
		t.Fatalf("owner isolation broken: alice=%q bob=%q", a.Original, b.Original)
	}
}
