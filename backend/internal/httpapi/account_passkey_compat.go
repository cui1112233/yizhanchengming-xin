package httpapi

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
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
	root.Handle("POST /api/account-recovery/passkeys", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var input struct {
			Challenge  string `json:"challenge"`
			Name       string `json:"name"`
			Credential struct {
				ID             string   `json:"id"`
				CredentialID   string   `json:"credentialId"`
				ClientDataJSON string   `json:"clientDataJSON"`
				PublicKey      string   `json:"publicKey"`
				Algorithm      int      `json:"algorithm"`
				Transports     []string `json:"transports"`
			} `json:"credential"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 128<<10)).Decode(&input) != nil {
			managedAccountError(w, http.StatusBadRequest, "请求内容无效")
			return
		}
		credentialID := strings.TrimSpace(input.Credential.CredentialID)
		if credentialID == "" {
			credentialID = strings.TrimSpace(input.Credential.ID)
		}
		publicKeyDER, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(input.Credential.PublicKey))
		clientDataRaw, clientErr := base64.RawURLEncoding.DecodeString(strings.TrimSpace(input.Credential.ClientDataJSON))
		if credentialID == "" || len(publicKeyDER) == 0 || len(publicKeyDER) > 4096 || clientErr != nil {
			managedAccountError(w, http.StatusBadRequest, "Passkey 注册数据不完整")
			return
		}
		var clientData struct {
			Type      string `json:"type"`
			Challenge string `json:"challenge"`
			Origin    string `json:"origin"`
		}
		if json.Unmarshal(clientDataRaw, &clientData) != nil || clientData.Type != "webauthn.create" || clientData.Challenge != input.Challenge {
			managedAccountError(w, http.StatusBadRequest, "Passkey challenge 无效")
			return
		}
		_, expectedOrigin := passkeyRequestContext(req)
		if clientData.Origin != expectedOrigin {
			managedAccountError(w, http.StatusForbidden, "Passkey Origin 不匹配")
			return
		}
		publicKey, err := x509.ParsePKIXPublicKey(publicKeyDER)
		if err != nil || (input.Credential.Algorithm == -7 && !isECDSAPublicKey(publicKey)) || (input.Credential.Algorithm == -257 && !isRSAPublicKey(publicKey)) || (input.Credential.Algorithm != -7 && input.Credential.Algorithm != -257) {
			managedAccountError(w, http.StatusBadRequest, "Passkey 公钥或算法不受支持")
			return
		}
		path := filepath.Join(profiles.SystemDir, "passkeys.json")
		state := map[string]any{"credentials": []any{}, "challenges": []any{}}
		if raw, readErr := os.ReadFile(path); readErr == nil {
			_ = json.Unmarshal(raw, &state)
		}
		challenges, _ := state["challenges"].([]any)
		matched := false
		now := time.Now()
		for _, raw := range challenges {
			row, ok := raw.(map[string]any)
			if !ok || toString(row["username"]) != identity.Username || toString(row["kind"]) != "register" || toString(row["challenge"]) != input.Challenge || strings.TrimSpace(toString(row["usedAt"])) != "" {
				continue
			}
			expires, parseErr := time.Parse(time.RFC3339, toString(row["expiresAt"]))
			if parseErr != nil || !expires.After(now) || toString(row["origin"]) != expectedOrigin {
				managedAccountError(w, http.StatusConflict, "Passkey challenge 已失效")
				return
			}
			row["usedAt"] = now.UTC().Format(time.RFC3339)
			matched = true
			break
		}
		if !matched {
			managedAccountError(w, http.StatusConflict, "Passkey challenge 不存在")
			return
		}
		credentials, _ := state["credentials"].([]any)
		for _, raw := range credentials {
			row, ok := raw.(map[string]any)
			if ok && toString(row["credentialId"]) == credentialID {
				managedAccountError(w, http.StatusConflict, "该 Passkey 已绑定")
				return
			}
		}
		name := strings.TrimSpace(input.Name)
		if name == "" {
			name = "Passkey"
		}
		if len(name) > 60 {
			name = name[:60]
		}
		credential := map[string]any{"id": fmt.Sprintf("go-%d", now.UnixNano()), "username": identity.Username, "credentialId": credentialID, "publicKey": input.Credential.PublicKey, "algorithm": input.Credential.Algorithm, "transports": input.Credential.Transports, "name": name, "signCount": 0, "createdAt": now.UTC().Format(time.RFC3339), "lastUsedAt": nil}
		state["credentials"] = append(credentials, credential)
		raw, _ := json.MarshalIndent(state, "", "  ")
		if err := os.WriteFile(path, raw, 0o600); err != nil {
			managedAccountError(w, http.StatusInternalServerError, "Passkey 保存失败")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"credential": credential})
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

func isECDSAPublicKey(value any) bool { _, ok := value.(*ecdsa.PublicKey); return ok }
func isRSAPublicKey(value any) bool   { _, ok := value.(*rsa.PublicKey); return ok }

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

func passkeyAuthenticationOptions(path, username string, req *http.Request) (map[string]any, error) {
	state := map[string]any{"credentials": []any{}, "challenges": []any{}}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &state)
	}
	credentials, _ := state["credentials"].([]any)
	allow := make([]map[string]any, 0)
	for _, raw := range credentials {
		row, ok := raw.(map[string]any)
		if !ok || toString(row["username"]) != username {
			continue
		}
		allow = append(allow, map[string]any{"type": "public-key", "id": row["credentialId"], "transports": row["transports"]})
	}
	if len(allow) == 0 {
		return nil, fmt.Errorf("当前账号没有可用 Passkey")
	}
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		return nil, err
	}
	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes)
	rpID, origin := passkeyRequestContext(req)
	now := time.Now().UTC()
	challenges, _ := state["challenges"].([]any)
	challenges = append(challenges, map[string]any{"id": fmt.Sprintf("go-%d", now.UnixNano()), "username": username, "kind": "authenticate", "challenge": challenge, "rpId": rpID, "origin": origin, "createdAt": now.Format(time.RFC3339), "expiresAt": now.Add(5 * time.Minute).Format(time.RFC3339), "usedAt": nil})
	state["challenges"] = challenges
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return nil, err
	}
	return map[string]any{"challenge": challenge, "rpId": rpID, "timeout": 300000, "userVerification": "required", "allowCredentials": allow}, nil
}

// finishPasskeyAuthentication verifies the WebAuthn assertion produced by the
// browser.  Keeping this check next to the challenge store makes the Go login
// path independent from the legacy Node authentication bridge.
func finishPasskeyAuthentication(path, username string, input map[string]any, req *http.Request) error {
	credentialValue, _ := input["credential"].(map[string]any)
	credentialID := strings.TrimSpace(toString(credentialValue["credentialId"]))
	if credentialID == "" {
		credentialID = strings.TrimSpace(toString(credentialValue["id"]))
	}
	challengeValue := strings.TrimSpace(toString(input["challenge"]))
	clientData, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(toString(credentialValue["clientDataJSON"])))
	if err != nil || len(clientData) == 0 {
		return fmt.Errorf("Passkey 登录数据不完整")
	}
	authenticatorData, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(toString(credentialValue["authenticatorData"])))
	if err != nil || len(authenticatorData) < 37 {
		return fmt.Errorf("Passkey 登录数据不完整")
	}
	signature, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(toString(credentialValue["signature"])))
	if err != nil || len(signature) == 0 || credentialID == "" || challengeValue == "" {
		return fmt.Errorf("Passkey 登录数据不完整")
	}
	var client struct {
		Type      string `json:"type"`
		Challenge string `json:"challenge"`
		Origin    string `json:"origin"`
	}
	if json.Unmarshal(clientData, &client) != nil || client.Type != "webauthn.get" || client.Challenge != challengeValue {
		return fmt.Errorf("Passkey challenge 无效")
	}
	_, expectedOrigin := passkeyRequestContext(req)
	if client.Origin != expectedOrigin {
		return fmt.Errorf("Passkey Origin 不匹配")
	}
	rpID, _ := passkeyRequestContext(req)
	expectedRPHash := sha256.Sum256([]byte(rpID))
	if !equalBytes(authenticatorData[:32], expectedRPHash[:]) {
		return fmt.Errorf("Passkey RP 校验失败")
	}
	flags := authenticatorData[32]
	if flags&0x01 == 0 || flags&0x04 == 0 {
		return fmt.Errorf("Passkey 未完成用户验证")
	}
	signCount := binaryBigEndian(authenticatorData[33:37])
	state := map[string]any{"credentials": []any{}, "challenges": []any{}}
	if raw, readErr := os.ReadFile(path); readErr == nil {
		_ = json.Unmarshal(raw, &state)
	}
	credentials, _ := state["credentials"].([]any)
	var credential map[string]any
	for _, raw := range credentials {
		row, ok := raw.(map[string]any)
		if ok && toString(row["username"]) == username && toString(row["credentialId"]) == credentialID {
			credential = row
			break
		}
	}
	if credential == nil {
		return fmt.Errorf("Passkey 不存在")
	}
	challenges, _ := state["challenges"].([]any)
	now := time.Now().UTC()
	matched := false
	for _, raw := range challenges {
		row, ok := raw.(map[string]any)
		if !ok || toString(row["username"]) != username || toString(row["kind"]) != "authenticate" || toString(row["challenge"]) != challengeValue || strings.TrimSpace(toString(row["usedAt"])) != "" {
			continue
		}
		expires, parseErr := time.Parse(time.RFC3339, toString(row["expiresAt"]))
		if parseErr != nil || !expires.After(now) || toString(row["origin"]) != expectedOrigin || toString(row["rpId"]) != rpID {
			return fmt.Errorf("Passkey challenge 已失效")
		}
		row["usedAt"] = now.Format(time.RFC3339)
		matched = true
		break
	}
	if !matched {
		return fmt.Errorf("Passkey challenge 不存在")
	}
	publicKeyDER, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(toString(credential["publicKey"])))
	if err != nil {
		return fmt.Errorf("Passkey 公钥损坏")
	}
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		return fmt.Errorf("Passkey 公钥损坏")
	}
	clientHash := sha256.Sum256(clientData)
	signed := append(append([]byte{}, authenticatorData...), clientHash[:]...)
	digest := sha256.Sum256(signed)
	algorithm := int(toNumber(credential["algorithm"]))
	switch key := publicKey.(type) {
	case *ecdsa.PublicKey:
		if algorithm != -7 || !ecdsa.VerifyASN1(key, digest[:], signature) {
			return fmt.Errorf("Passkey 签名验证失败")
		}
	case *rsa.PublicKey:
		if algorithm != -257 || rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature) != nil {
			return fmt.Errorf("Passkey 签名验证失败")
		}
	default:
		return fmt.Errorf("Passkey 公钥算法不受支持")
	}
	previous := uint64(toNumber(credential["signCount"]))
	if previous > 0 && signCount > 0 && uint64(signCount) <= previous {
		return fmt.Errorf("Passkey 计数器异常")
	}
	if uint64(signCount) > previous {
		credential["signCount"] = signCount
	}
	credential["lastUsedAt"] = now.Format(time.RFC3339)
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil || os.WriteFile(path, raw, 0o600) != nil {
		return fmt.Errorf("Passkey 登录状态保存失败")
	}
	return nil
}

func toNumber(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		return 0
	}
}

func binaryBigEndian(value []byte) uint32 {
	if len(value) < 4 {
		return 0
	}
	return uint32(value[0])<<24 | uint32(value[1])<<16 | uint32(value[2])<<8 | uint32(value[3])
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
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
