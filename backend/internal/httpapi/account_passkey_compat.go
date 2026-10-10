package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

// Passkey registration still stays on the WebAuthn implementation. These
// read/delete handlers keep the security page on Go without exposing keys.
func RegisterAccountPasskeyCompat(root *http.ServeMux, middleware BridgeAuth, profiles *MemberProfileStore, accounts auth.Store) {
	if root == nil || profiles == nil || accounts == nil {
		return
	}
	root.Handle("GET /api/account-recovery/passkeys", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		credentials := readPasskeyCredentials(profiles, identity.Username)
		writeJSON(w, http.StatusOK, map[string]any{"passkeys": credentials})
	})))
	root.Handle("POST /api/account-recovery/passkeys/options", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			CurrentPassword string `json:"currentPassword"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		account, err := accounts.FindAccount(req.Context(), identity.Username)
		if err != nil || !auth.VerifyPassword(input.CurrentPassword, account.PasswordHash, account.PasswordAlgorithm) {
			managedAccountError(w, http.StatusBadRequest, "当前密码不正确")
			return
		}
		challengeBytes := make([]byte, 32)
		if _, err := rand.Read(challengeBytes); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey challenge 生成失败")
			return
		}
		challenge := base64.RawURLEncoding.EncodeToString(challengeBytes)
		rpID, origin := passkeyRequestContext(req)
		path := filepath.Join(profiles.SystemDir, "passkeys.json")
		state := map[string]any{"credentials": []any{}, "challenges": []any{}}
		if raw, readErr := os.ReadFile(path); readErr == nil {
			_ = json.Unmarshal(raw, &state)
		}
		credentials, _ := state["credentials"].([]any)
		exclude := make([]map[string]any, 0)
		for _, raw := range credentials {
			row, ok := raw.(map[string]any)
			if !ok || toString(row["username"]) != identity.Username {
				continue
			}
			exclude = append(exclude, map[string]any{"type": "public-key", "id": row["credentialId"], "transports": row["transports"]})
		}
		challenges, _ := state["challenges"].([]any)
		now := time.Now().UTC()
		challenges = append(challenges, map[string]any{"id": fmt.Sprintf("go-%d", now.UnixNano()), "username": identity.Username, "kind": "register", "challenge": challenge, "rpId": rpID, "origin": origin, "createdAt": now.Format(time.RFC3339), "expiresAt": now.Add(5 * time.Minute).Format(time.RFC3339), "usedAt": nil})
		state["challenges"] = challenges
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey 配置保存失败")
			return
		}
		raw, _ := json.MarshalIndent(state, "", "  ")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey 配置保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"challenge": challenge, "rp": map[string]string{"id": rpID, "name": "一战晟铭"}, "user": map[string]string{"id": base64.RawURLEncoding.EncodeToString([]byte(identity.Username)), "name": identity.Username, "displayName": identity.Username}, "pubKeyCredParams": []map[string]any{{"type": "public-key", "alg": -7}, {"type": "public-key", "alg": -257}}, "timeout": 300000, "attestation": "none", "authenticatorSelection": map[string]string{"residentKey": "preferred", "userVerification": "required"}, "excludeCredentials": exclude})
	})))
	root.Handle("DELETE /api/account-recovery/passkeys/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		path := adminSystemFile(profiles.SystemDir, "passkeys.json")
		state := map[string]any{"credentials": []any{}, "challenges": []any{}}
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		rows, _ := state["credentials"].([]any)
		id := req.PathValue("id")
		removed := map[string]any(nil)
		next := make([]any, 0, len(rows))
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if ok && removed == nil && toString(row["username"]) == identity.Username && (toString(row["id"]) == id || toString(row["credentialId"]) == id) {
				removed = map[string]any{"id": row["id"], "username": row["username"], "name": row["name"]}
				continue
			}
			next = append(next, raw)
		}
		if removed == nil {
			managedAccountError(w, http.StatusNotFound, "Passkey 不存在")
			return
		}
		state["credentials"] = next
		raw, err := json.MarshalIndent(state, "", "  ")
		if err != nil || os.WriteFile(path, raw, 0o600) != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey 删除失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"removed": removed})
	})))
}

func passkeyRequestContext(req *http.Request) (string, string) {
	host := strings.TrimSpace(req.Host)
	if host == "" {
		host = "localhost"
	}
	if strings.HasPrefix(host, "[") {
		if end := strings.Index(host, "]"); end >= 0 {
			host = host[1:end]
		}
	} else if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	rpID := strings.ToLower(host)
	scheme := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	origin := scheme + "://" + strings.TrimSpace(req.Host)
	if parsed, err := url.Parse(origin); err == nil && parsed.Host != "" {
		origin = parsed.Scheme + "://" + parsed.Host
	}
	return rpID, origin
}

func readPasskeyCredentials(profiles *MemberProfileStore, username string) []map[string]any {
	path := adminSystemFile(profiles.SystemDir, "passkeys.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	state := map[string]any{}
	if json.Unmarshal(data, &state) != nil {
		return []map[string]any{}
	}
	rows, _ := state["credentials"].([]any)
	result := make([]map[string]any, 0)
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok || strings.TrimSpace(toString(row["username"])) != username {
			continue
		}
		result = append(result, map[string]any{"id": row["id"], "username": row["username"], "credentialId": row["credentialId"], "name": row["name"], "algorithm": row["algorithm"], "transports": row["transports"], "signCount": row["signCount"], "createdAt": row["createdAt"], "lastUsedAt": row["lastUsedAt"]})
	}
	return result
}
