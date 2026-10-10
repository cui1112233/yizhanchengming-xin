package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAccountRecoveryEmailVerifyConsumesToken(t *testing.T) {
	dir := t.TempDir()
	token := "email-verification-token"
	digest := sha256.Sum256([]byte(token))
	state := map[string]any{
		"verifiedEmails": []any{},
		"emailTokens": []any{map[string]any{
			"username": "alice", "email": "alice@example.com", "tokenDigest": hex.EncodeToString(digest[:]),
			"expiresAt": time.Now().UTC().Add(time.Minute).Format(time.RFC3339), "usedAt": nil,
		}},
	}
	raw, _ := json.Marshal(state)
	if err := os.WriteFile(filepath.Join(dir, "account-recovery.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAccountRecoveryStatusCompat(mux, BridgeAuth{Secret: "secret", Now: time.Now, AuthStore: &authTestStore{}}, dir)
	req := httptest.NewRequest(http.MethodPost, "/api/account-recovery/email/verify", strings.NewReader(`{"token":"email-verification-token"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"verified":true`) {
		t.Fatalf("verify response=%d %s", res.Code, res.Body.String())
	}
	used := httptest.NewRecorder()
	mux.ServeHTTP(used, httptest.NewRequest(http.MethodPost, "/api/account-recovery/email/verify", strings.NewReader(`{"token":"email-verification-token"}`)))
	if used.Code != http.StatusConflict {
		t.Fatalf("second verify status=%d body=%s", used.Code, used.Body.String())
	}
}
