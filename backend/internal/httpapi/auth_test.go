package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/auth"
)

type authTestStore struct {
	account  auth.Account
	password string
	session  auth.Session
}

func (s *authTestStore) FindAccount(context.Context, string) (auth.Account, error) {
	if s.account.Username == "" {
		return auth.Account{}, auth.ErrNotFound
	}
	return s.account, nil
}
func (s *authTestStore) CreateAccount(context.Context, auth.Account) error            { return nil }
func (s *authTestStore) UpdatePassword(context.Context, string, string, string) error { return nil }
func (s *authTestStore) SetActive(context.Context, string, bool) error                { return nil }
func (s *authTestStore) CreateSession(_ context.Context, session auth.Session) error {
	s.session = session
	return nil
}
func (s *authTestStore) FindSession(_ context.Context, token string, now time.Time) (auth.Session, error) {
	if token == "" || token != s.session.Token || !s.session.ExpiresAt.After(now) {
		return auth.Session{}, auth.ErrNotFound
	}
	return s.session, nil
}
func (s *authTestStore) RevokeSession(_ context.Context, token string) error {
	if token == s.session.Token {
		s.session = auth.Session{}
	}
	return nil
}

func TestAuthRoutesRejectUnknownAccountWithoutLeakingDetails(t *testing.T) {
	store := &authTestStore{}
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, AuthOptions{Store: store, Now: func() time.Time { return time.Unix(100, 0) }})
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"nobody","password":"wrong-password"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized || strings.Contains(res.Body.String(), "password") || strings.Contains(res.Body.String(), "hash") {
		t.Fatalf("login response = %d %s", res.Code, res.Body.String())
	}
}

func TestAuthRoutesLoginSessionAndLogout(t *testing.T) {
	hash, algorithm, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatal(err)
	}
	store := &authTestStore{account: auth.Account{Username: "alice", PasswordHash: hash, PasswordAlgorithm: algorithm, Active: true, IsOwner: true}}
	mux := http.NewServeMux()
	RegisterAuthRoutes(mux, AuthOptions{Store: store, Now: func() time.Time { return time.Unix(100, 0) }})
	login := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"username":"alice","password":"correct-password","remember":true}`))
	login.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	mux.ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK || store.session.Token == "" {
		t.Fatalf("login response = %d %s", loginResponse.Code, loginResponse.Body.String())
	}

	session := httptest.NewRequest(http.MethodGet, "/api/login/session", nil)
	session.Header.Set("Authorization", "Bearer "+store.session.Token)
	sessionResponse := httptest.NewRecorder()
	mux.ServeHTTP(sessionResponse, session)
	if sessionResponse.Code != http.StatusOK || !strings.Contains(sessionResponse.Body.String(), "alice") {
		t.Fatalf("session response = %d %s", sessionResponse.Code, sessionResponse.Body.String())
	}

	logout := httptest.NewRequest(http.MethodPost, "/api/login/logout", nil)
	logout.Header.Set("Authorization", "Bearer "+store.session.Token)
	logoutResponse := httptest.NewRecorder()
	mux.ServeHTTP(logoutResponse, logout)
	if logoutResponse.Code != http.StatusNoContent || store.session.Token != "" {
		t.Fatalf("logout response = %d", logoutResponse.Code)
	}
}
