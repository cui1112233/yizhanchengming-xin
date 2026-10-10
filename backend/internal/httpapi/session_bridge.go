package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// NewLegacySessionResolver lets the Go service recognize a token created by the
// still-running Node service during the migration window. It only asks Node
// whether the token is valid; it never stores or logs the token.
func NewLegacySessionResolver(baseURL string) func(context.Context, string) (BridgeIdentity, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	client := &http.Client{Timeout: 3 * time.Second}
	return func(ctx context.Context, token string) (BridgeIdentity, error) {
		if baseURL == "" || strings.TrimSpace(token) == "" {
			return BridgeIdentity{}, errors.New("legacy auth bridge is not configured")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/login/session", nil)
		if err != nil {
			return BridgeIdentity{}, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		response, err := client.Do(req)
		if err != nil {
			return BridgeIdentity{}, err
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return BridgeIdentity{}, errors.New("legacy login session rejected")
		}
		var payload struct {
			Username string `json:"username"`
			Active   bool   `json:"active"`
			IsOwner  bool   `json:"isOwner"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil || strings.TrimSpace(payload.Username) == "" || !payload.Active {
			return BridgeIdentity{}, errors.New("legacy login session is invalid")
		}
		return BridgeIdentity{Username: strings.TrimSpace(payload.Username), IsOwner: payload.IsOwner}, nil
	}
}
