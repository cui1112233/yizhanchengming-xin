package videomanagementconnector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type fakeService struct{}

func (fakeService) Login(context.Context, LoginInput) (Session, error) {
	return Session{Cookie: "session=secret", TargetUsername: "tester"}, nil
}
func (fakeService) RefreshCatalogs(context.Context, Session) (Catalogs, error) {
	return Catalogs{Organizations: []CatalogItem{{ID: "org-1", Name: "默认组织"}}}, nil
}
func (fakeService) Submit(context.Context, Session, SubmissionInput) (Submission, error) {
	return Submission{ID: "submission-1", Status: "submitted"}, nil
}
func (fakeService) GetSubmission(context.Context, Session, string) (Submission, error) {
	return Submission{ID: "submission-1", Status: "published"}, nil
}

func TestConnectorRejectsMissingInternalToken(t *testing.T) {
	s := NewServer(Config{InternalToken: "internal-token"}, fakeService{}, NewMemorySessionStore())
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/login", strings.NewReader(`{"username":"u","password":"p"}`))
	s.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestConnectorLoginAndCatalogUseOwnerSession(t *testing.T) {
	store := NewMemorySessionStore()
	s := NewServer(Config{InternalToken: "internal-token"}, fakeService{}, store)
	login := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/sessions/login", strings.NewReader(`{"username":"u","password":"p"}`))
	req.Header.Set("Authorization", "Bearer internal-token")
	req.Header.Set("X-Owner", "alice")
	s.Handler().ServeHTTP(login, req)
	if login.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", login.Code, login.Body.String())
	}
	catalogs := httptest.NewRecorder()
	refresh := httptest.NewRequest(http.MethodPost, "/v1/catalogs/refresh", nil)
	refresh.Header.Set("Authorization", "Bearer internal-token")
	refresh.Header.Set("X-Owner", "alice")
	s.Handler().ServeHTTP(catalogs, refresh)
	if catalogs.Code != http.StatusOK || !strings.Contains(catalogs.Body.String(), "默认组织") {
		t.Fatalf("catalog status=%d body=%s", catalogs.Code, catalogs.Body.String())
	}
}

func TestFileSessionStoreEncryptsCookieAtRest(t *testing.T) {
	root := t.TempDir()
	store, err := NewFileSessionStore(root, []byte("connector-test-key"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), "alice", Session{Cookie: "session=secret-cookie"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(store.pathForTest("alice"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "secret-cookie") {
		t.Fatal("plaintext cookie persisted")
	}
	loaded, err := store.Load(context.Background(), "alice")
	if err != nil || loaded.Cookie != "session=secret-cookie" {
		t.Fatalf("loaded=%#v err=%v", loaded, err)
	}
}
