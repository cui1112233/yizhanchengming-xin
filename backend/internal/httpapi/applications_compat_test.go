package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

type applicationTestStore struct{ accounts map[string]auth.Account }

func (s *applicationTestStore) FindAccount(_ context.Context, username string) (auth.Account, error) {
	account, ok := s.accounts[username]
	if !ok {
		return auth.Account{}, auth.ErrNotFound
	}
	return account, nil
}
func (s *applicationTestStore) CreateAccount(_ context.Context, account auth.Account) error {
	s.accounts[account.Username] = account
	return nil
}
func (s *applicationTestStore) UpdatePassword(context.Context, string, string, string) error {
	return nil
}
func (s *applicationTestStore) SetActive(context.Context, string, bool) error     { return nil }
func (s *applicationTestStore) CreateSession(context.Context, auth.Session) error { return nil }
func (s *applicationTestStore) FindSession(context.Context, string, time.Time) (auth.Session, error) {
	return auth.Session{}, auth.ErrNotFound
}
func (s *applicationTestStore) RevokeSession(context.Context, string) error { return nil }
func (s *applicationTestStore) PurgeAccount(_ context.Context, username string) (int64, error) {
	if _, ok := s.accounts[username]; !ok {
		return 0, auth.ErrNotFound
	}
	delete(s.accounts, username)
	return 1, nil
}

func TestApplicationsCompatSubmitStatusAndWithdraw(t *testing.T) {
	accounts := &applicationTestStore{accounts: map[string]auth.Account{}}
	profiles := NewMemberProfileStore(t.TempDir())
	now := func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) }
	mux := http.NewServeMux()
	RegisterApplicationsCompat(mux, accounts, profiles, now)

	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}
	created := request(http.MethodPost, "/api/applications", `{"username":"new-user","password":"long-password","reason":"测试申请"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	var application map[string]any
	if err := json.Unmarshal(created.Body.Bytes(), &application); err != nil {
		t.Fatal(err)
	}
	id, _ := application["id"].(string)
	if id == "" || application["status"] != "pending" {
		t.Fatalf("created=%v", application)
	}
	status := request(http.MethodPost, "/api/applications/"+id+"/status", `{"username":"new-user","password":"long-password"}`)
	if status.Code != http.StatusOK || strings.Contains(status.Body.String(), "passwordHash") {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
	withdraw := request(http.MethodPost, "/api/applications/"+id+"/withdraw", `{"username":"new-user","password":"long-password"}`)
	if withdraw.Code != http.StatusOK || !strings.Contains(withdraw.Body.String(), `"withdrawn"`) {
		t.Fatalf("withdraw=%d body=%s", withdraw.Code, withdraw.Body.String())
	}
}
