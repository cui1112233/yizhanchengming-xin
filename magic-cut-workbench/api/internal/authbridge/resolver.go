package authbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrUnauthenticated = errors.New("account is not authenticated")

type Resolver struct {
	baseURL *url.URL
	client  *http.Client
}

func NewResolver(baseURL string, client *http.Client) (*Resolver, error) {
	parsed, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid account service URL")
	}
	if client == nil {
		client = &http.Client{}
	}
	clientCopy := *client
	if clientCopy.Timeout == 0 {
		clientCopy.Timeout = 15 * time.Second
	}
	return &Resolver{baseURL: parsed, client: &clientCopy}, nil
}

func (r *Resolver) Resolve(ctx context.Context, request *http.Request) (string, error) {
	cookie, err := request.Cookie(accountCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", ErrUnauthenticated
	}
	target := *r.baseURL
	target.Path = "/api/login/session"
	upstream, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return "", err
	}
	upstream.Header.Set("Authorization", "Bearer "+strings.TrimSpace(cookie.Value))
	upstream.Header.Set("X-Qiantie-Auth-Bridge", "1")
	response, err := r.client.Do(upstream)
	if err != nil {
		return "", fmt.Errorf("resolve account: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthenticated
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("account service returned %d", response.StatusCode)
	}
	var account struct {
		Username string `json:"username"`
		Active   bool   `json:"active"`
	}
	if err := json.NewDecoder(response.Body).Decode(&account); err != nil {
		return "", fmt.Errorf("decode account: %w", err)
	}
	if !account.Active || strings.TrimSpace(account.Username) == "" {
		return "", ErrUnauthenticated
	}
	return strings.TrimSpace(account.Username), nil
}
