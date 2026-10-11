package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

var adminGrantCapabilities = map[string]bool{
	"account:review": true,
	"preset:draft":   true,
	"preset:publish": true,
}

var adminGrantScopes = map[string]bool{
	"*":           true,
	"script":      true,
	"novel-panel": true,
}

// RegisterAdminGovernanceCompat keeps the account-application and backend-grant
// contracts available while the rest of the admin surface is being moved to Go.
// Secrets in application records are never returned to the browser.
func RegisterAdminGovernanceCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/admin/applications", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := requireAdminGovernance(w, req, accounts, profiles, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"applications": readAdminGovernanceArray(profiles, "applications.json")})
	})))
	root.Handle("POST /api/admin/accounts", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := requireAdminGovernance(w, req, accounts, profiles, true)
		if !ok {
			return
		}
		input, ok := decodeAdminJSON(w, req, 32<<10)
		if !ok {
			return
		}
		account, err := createAdminAccount(req.Context(), profiles, accounts, identity.Username, input)
		if err != nil {
			writeAdminGovernanceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"account": account})
	})))
	root.Handle("POST /api/admin/accounts/{username}/status", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := requireAdminGovernance(w, req, accounts, profiles, false)
		if !ok {
			return
		}
		input, ok := decodeAdminJSON(w, req, 8<<10)
		if !ok {
			return
		}
		active, exists := input["active"].(bool)
		if !exists {
			writeAdminGovernanceError(w, auth.ErrInvalid)
			return
		}
		account, err := setAdminAccountStatus(req.Context(), profiles, accounts, identity.Username, req.PathValue("username"), active)
		if err != nil {
			writeAdminGovernanceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"account": account})
	})))
	root.Handle("POST /api/admin/accounts/{username}/reset-password", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := requireAdminGovernance(w, req, accounts, profiles, false)
		if !ok {
			return
		}
		input, ok := decodeAdminJSON(w, req, 8<<10)
		if !ok {
			return
		}
		password := strings.TrimSpace(toString(input["password"]))
		if len(password) < 8 {
			writeAdminGovernanceError(w, auth.ErrInvalid)
			return
		}
		account, err := resetAdminAccountPassword(req.Context(), profiles, accounts, identity.Username, req.PathValue("username"), password)
		if err != nil {
			writeAdminGovernanceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"account": account})
	})))
	for _, action := range []string{"approve", "reject"} {
		action := action
		root.Handle("POST /api/admin/applications/{id}/"+action, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := requireAdminGovernance(w, req, accounts, profiles, false)
			if !ok {
				return
			}
			application, err := reviewAdminApplication(req.Context(), profiles, accounts, identity.Username, req.PathValue("id"), action == "approve")
			if err != nil {
				writeAdminGovernanceError(w, err)
				return
			}
			if action == "approve" {
				writeJSON(w, http.StatusOK, application)
			} else {
				writeJSON(w, http.StatusOK, map[string]any{"application": application})
			}
		})))
	}
	root.Handle("GET /api/admin/grants", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := requireAdminGovernance(w, req, accounts, profiles, true); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"grants": readAdminGovernanceArray(profiles, "grants.json")})
	})))
	root.Handle("POST /api/admin/grants", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := requireAdminGovernance(w, req, accounts, profiles, true)
		if !ok {
			return
		}
		input, ok := decodeAdminJSON(w, req, 16<<10)
		if !ok {
			return
		}
		grant, err := createAdminGrant(req.Context(), profiles, accounts, identity.Username, input)
		if err != nil {
			writeAdminGovernanceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"grant": grant})
	})))
	root.Handle("DELETE /api/admin/grants/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := requireAdminGovernance(w, req, accounts, profiles, true)
		if !ok {
			return
		}
		grant, err := revokeAdminGrant(profiles, identity.Username, req.PathValue("id"))
		if err != nil {
			writeAdminGovernanceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"grant": grant})
	})))
}

func requireAdminGovernance(w http.ResponseWriter, req *http.Request, accounts auth.Store, profiles *MemberProfileStore, ownerOnly bool) (BridgeIdentity, bool) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return BridgeIdentity{}, false
	}
	account, err := accounts.FindAccount(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "账号不存在"})
		return BridgeIdentity{}, false
	}
	if ownerOnly {
		if !account.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "仅主账号可以调整后台授权"})
			return BridgeIdentity{}, false
		}
		return identity, true
	}
	profile := profiles.member(identity.Username, accounts)
	if !account.IsOwner && (profile == nil || strings.ToLower(strings.TrimSpace(toString(profile["role"]))) != "dev") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "无账号治理权限"})
		return BridgeIdentity{}, false
	}
	return identity, true
}

func readAdminGovernanceArray(profiles *MemberProfileStore, name string) []map[string]any {
	path := adminSystemFile(profiles.SystemDir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil || rows == nil {
		return []map[string]any{}
	}
	if name == "applications.json" {
		for _, row := range rows {
			delete(row, "passwordHash")
			delete(row, "passwordAlgorithm")
		}
	}
	return rows
}

func writeAdminGovernanceArray(profiles *MemberProfileStore, name string, rows []map[string]any) error {
	path := adminSystemFile(profiles.SystemDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func reviewAdminApplication(ctx context.Context, profiles *MemberProfileStore, accounts auth.Store, actor, id string, approve bool) (map[string]any, error) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	path := adminSystemFile(profiles.SystemDir, "applications.json")
	// readAdminGovernanceArray intentionally removes the password hash, so read
	// the private source again for approval and only sanitize the response.
	privateRows := readAdminGovernancePrivateArray(path)
	index := -1
	for i, row := range privateRows {
		if toString(row["id"]) == strings.TrimSpace(id) {
			index = i
			break
		}
	}
	if index < 0 {
		return nil, auth.ErrNotFound
	}
	application := privateRows[index]
	if strings.ToLower(strings.TrimSpace(toString(application["status"]))) != "pending" {
		return nil, auth.ErrInvalid
	}
	if approve {
		username := strings.TrimSpace(toString(application["username"]))
		passwordHash := strings.TrimSpace(toString(application["passwordHash"]))
		if username == "" || passwordHash == "" {
			return nil, auth.ErrInvalid
		}
		if _, err := accounts.FindAccount(ctx, username); err == nil {
			return nil, auth.ErrInvalid
		} else if err != auth.ErrNotFound {
			return nil, err
		}
		algorithm := strings.TrimSpace(toString(application["passwordAlgorithm"]))
		if algorithm == "" {
			algorithm = "scrypt"
		}
		created := time.Now().UTC()
		if err := accounts.CreateAccount(ctx, auth.Account{Username: username, PasswordHash: passwordHash, PasswordAlgorithm: algorithm, Active: true, CreatedAt: created, UpdatedAt: created}); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	application["status"] = map[bool]string{true: "approved", false: "rejected"}[approve]
	application["updatedAt"], application["reviewedAt"], application["reviewedBy"] = now, now, actor
	if err := writeAdminGovernanceArray(profiles, "applications.json", privateRows); err != nil {
		return nil, err
	}
	public := cloneAdminGovernanceMap(application)
	delete(public, "passwordHash")
	delete(public, "passwordAlgorithm")
	if approve {
		account, err := accounts.FindAccount(ctx, strings.TrimSpace(toString(application["username"])))
		if err != nil {
			return nil, err
		}
		return map[string]any{"application": public, "account": map[string]any{"username": account.Username, "active": account.Active, "isOwner": account.IsOwner}}, nil
	}
	return public, nil
}

func createAdminGrant(ctx context.Context, profiles *MemberProfileStore, accounts auth.Store, actor string, input map[string]any) (map[string]any, error) {
	subject := strings.TrimSpace(toString(input["subject"]))
	capability := strings.TrimSpace(toString(input["capability"]))
	scope := strings.TrimSpace(toString(input["scope"]))
	if !managedUsernamePattern.MatchString(subject) || !adminGrantCapabilities[capability] || !adminGrantScopes[scope] {
		return nil, auth.ErrInvalid
	}
	account, err := accounts.FindAccount(ctx, subject)
	if err != nil {
		return nil, err
	}
	if !account.Active || account.IsOwner {
		return nil, auth.ErrInvalid
	}
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	rows := readAdminGovernancePrivateArray(adminSystemFile(profiles.SystemDir, "grants.json"))
	for _, row := range rows {
		if toString(row["subject"]) == subject && toString(row["capability"]) == capability && toString(row["scope"]) == scope {
			return nil, auth.ErrInvalid
		}
	}
	grant := map[string]any{"id": newAdminGrantID(), "at": time.Now().UTC().Format(time.RFC3339Nano), "actor": actor, "subject": subject, "capability": capability, "scope": scope}
	rows = append(rows, grant)
	if err := writeAdminGovernanceArray(profiles, "grants.json", rows); err != nil {
		return nil, err
	}
	return grant, nil
}

func createAdminAccount(ctx context.Context, profiles *MemberProfileStore, accounts auth.Store, _ string, input map[string]any) (map[string]any, error) {
	username := strings.TrimSpace(toString(input["username"]))
	password := strings.TrimSpace(toString(input["password"]))
	if !managedUsernamePattern.MatchString(username) || len(password) < 8 {
		return nil, auth.ErrInvalid
	}
	if _, err := accounts.FindAccount(ctx, username); err == nil {
		return nil, auth.ErrInvalid
	} else if err != auth.ErrNotFound {
		return nil, err
	}
	hash, algorithm, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := accounts.CreateAccount(ctx, auth.Account{Username: username, PasswordHash: hash, PasswordAlgorithm: algorithm, Active: true, CreatedAt: now, UpdatedAt: now}); err != nil {
		return nil, err
	}
	if err := saveManagedProfile(profiles, username, username, "member", nil); err != nil {
		return nil, err
	}
	account, err := accounts.FindAccount(ctx, username)
	if err != nil {
		return nil, err
	}
	return managedAccountRow(account, managedProfile(profiles, username)), nil
}

func setAdminAccountStatus(ctx context.Context, profiles *MemberProfileStore, accounts auth.Store, _ string, username string, active bool) (map[string]any, error) {
	account, err := accounts.FindAccount(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	if account.IsOwner && !active {
		return nil, auth.ErrInvalid
	}
	if err := accounts.SetActive(ctx, account.Username, active); err != nil {
		return nil, err
	}
	account.Active = active
	return managedAccountRow(account, managedProfile(profiles, account.Username)), nil
}

func resetAdminAccountPassword(ctx context.Context, profiles *MemberProfileStore, accounts auth.Store, _ string, username, password string) (map[string]any, error) {
	account, err := accounts.FindAccount(ctx, strings.TrimSpace(username))
	if err != nil {
		return nil, err
	}
	hash, algorithm, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	if err := accounts.UpdatePassword(ctx, account.Username, hash, algorithm); err != nil {
		return nil, err
	}
	account.UpdatedAt = time.Now().UTC()
	return managedAccountRow(account, managedProfile(profiles, account.Username)), nil
}

func revokeAdminGrant(profiles *MemberProfileStore, _ string, id string) (map[string]any, error) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	rows := readAdminGovernancePrivateArray(adminSystemFile(profiles.SystemDir, "grants.json"))
	for index, row := range rows {
		if toString(row["id"]) != strings.TrimSpace(id) {
			continue
		}
		rows = append(rows[:index], rows[index+1:]...)
		if err := writeAdminGovernanceArray(profiles, "grants.json", rows); err != nil {
			return nil, err
		}
		return row, nil
	}
	return nil, auth.ErrNotFound
}

func readAdminGovernancePrivateArray(path string) []map[string]any {
	raw, err := os.ReadFile(path)
	if err != nil {
		return []map[string]any{}
	}
	var rows []map[string]any
	if json.Unmarshal(raw, &rows) != nil || rows == nil {
		return []map[string]any{}
	}
	return rows
}

func cloneAdminGovernanceMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func writeAdminGovernanceError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if err == auth.ErrNotFound {
		status = http.StatusNotFound
	}
	if err == auth.ErrInvalid {
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func newAdminGrantID() string {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "grant-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return "grant-" + hex.EncodeToString(raw[:])
}
