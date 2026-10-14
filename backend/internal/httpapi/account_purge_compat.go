package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

type accountPurger interface {
	PurgeAccount(context.Context, string) (int64, error)
}

// RegisterAccountPurgeCompat owns the last destructive account-recovery
// endpoint that used to live in Node. The operation remains deliberately
// guarded: only DEV may perform it, the target must be archived/inactive, and
// the exact username must be typed as confirmation.
func RegisterAccountPurgeCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore, avatarDir string) {
	if root == nil || middleware.AuthStore == nil || accounts == nil || profiles == nil {
		return
	}
	root.Handle("POST /api/account-recovery/purge/{username}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		actor := profiles.member(identity.Username, accounts)
		targetUsername := strings.TrimSpace(req.PathValue("username"))
		target := profiles.member(targetUsername, accounts)
		if actor == nil || toString(actor["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "仅 DEV 可以永久删除账号"})
			return
		}
		var input struct {
			CurrentPassword string `json:"currentPassword"`
			Confirmation    string `json:"confirmation"`
			Reason          string `json:"reason"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		actorAccount, err := accounts.FindAccount(req.Context(), identity.Username)
		if err != nil || !auth.VerifyPassword(input.CurrentPassword, actorAccount.PasswordHash, actorAccount.PasswordAlgorithm) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前 DEV 密码不正确"})
			return
		}
		if target == nil || target["isOwner"] == true {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "目标账号不可删除"})
			return
		}
		if toString(target["role"]) != "member" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "仅允许永久删除 MEMBER；请先完成团队转移与角色调整"})
			return
		}
		if active, ok := target["active"].(bool); !ok || active {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "请先归档并停用成员，再执行永久删除"})
			return
		}
		if !memberArchiveExists(profiles, targetUsername) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "成员尚未归档"})
			return
		}
		if input.Confirmation != targetUsername {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入完整账号名确认永久删除"})
			return
		}
		purger, ok := accounts.(accountPurger)
		if !ok {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "当前账号存储不支持永久删除"})
			return
		}
		removed, err := purger.PurgeAccount(req.Context(), targetUsername)
		if err != nil {
			if err == auth.ErrNotFound {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "目标账号不可删除"})
				return
			}
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "账号永久删除失败"})
			return
		}
		removedPasskeys := purgePasskeyRows(profiles.SystemDir, targetUsername)
		removedMFA := purgeMFARow(filepath.Join(profiles.SystemDir, "mfa.json"), targetUsername)
		redactDeletedMember(profiles, targetUsername)
		_, _ = revokeRecoveryToken(filepath.Join(profiles.SystemDir, "account-recovery.json"), "email", targetUsername)
		_, _ = revokeRecoveryToken(filepath.Join(profiles.SystemDir, "account-recovery.json"), "password", targetUsername)
		removeMemberAvatars(avatarDir, targetUsername)
		tombstone := appendDeletedAccountTombstone(profiles.SystemDir, targetUsername, identity.Username, input.Reason)
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "removed": removed, "tombstone": tombstone, "revoked": removed, "removedPasskeys": removedPasskeys, "removedMfa": removedMFA, "auditRetained": true})
	})))
}

func memberArchiveExists(profiles *MemberProfileStore, username string) bool {
	for _, row := range profiles.readArray("member-archives.json") {
		if toString(row["username"]) == username && strings.TrimSpace(toString(row["archivedAt"])) != "" && strings.TrimSpace(toString(row["restoredAt"])) == "" {
			return true
		}
	}
	return false
}

func redactDeletedMember(profiles *MemberProfileStore, username string) {
	profiles.mu.Lock()
	defer profiles.mu.Unlock()
	rows := profiles.readArray("member-profiles.json")
	for _, row := range rows {
		if toString(row["username"]) == username {
			row["displayName"], row["avatarUrl"], row["boundTo"], row["monthlyTokenLimit"] = "已删除用户", nil, nil, nil
			row["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
		}
	}
	_ = profiles.writeArray("member-profiles.json", rows)
	details := profiles.readArray("member-profile-details.json")
	filteredDetails := details[:0]
	for _, row := range details {
		if toString(row["username"]) != username {
			filteredDetails = append(filteredDetails, row)
		}
	}
	_ = profiles.writeArray("member-profile-details.json", filteredDetails)
}

func purgePasskeyRows(systemDir, username string) int {
	path := adminSystemFile(systemDir, "passkeys.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var state map[string]any
	if json.Unmarshal(raw, &state) != nil {
		return 0
	}
	rows, _ := state["credentials"].([]any)
	next := make([]any, 0, len(rows))
	removed := 0
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if ok && toString(row["username"]) == username {
			removed++
			continue
		}
		next = append(next, value)
	}
	if removed > 0 {
		state["credentials"] = next
		if data, marshalErr := json.MarshalIndent(state, "", "  "); marshalErr == nil {
			_ = os.WriteFile(path, data, 0o600)
		}
	}
	return removed
}

func purgeMFARow(path, username string) int {
	rows, err := readMFAFile(path)
	if err != nil {
		return 0
	}
	next := rows[:0]
	removed := 0
	for _, row := range rows {
		if row.Username == username {
			removed++
			continue
		}
		next = append(next, row)
	}
	if removed > 0 {
		_ = writeMFAFile(path, next)
	}
	return removed
}

func removeMemberAvatars(dir, username string) {
	if strings.TrimSpace(dir) == "" {
		return
	}
	for _, ext := range []string{"png", "jpg", "webp"} {
		_ = os.Remove(filepath.Join(dir, username+"."+ext))
	}
}

func appendDeletedAccountTombstone(systemDir, username, actor, reason string) map[string]any {
	path := filepath.Join(systemDir, "account-recovery.json")
	state := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &state)
	}
	rows, _ := state["deletedAccounts"].([]any)
	tombstone := map[string]any{"username": username, "deletedBy": actor, "deletedAt": time.Now().UTC().Format(time.RFC3339), "reason": strings.TrimSpace(reason)}
	state["deletedAccounts"] = append(rows, tombstone)
	if data, err := json.MarshalIndent(state, "", "  "); err == nil {
		_ = os.MkdirAll(filepath.Dir(path), 0o700)
		_ = os.WriteFile(path, data, 0o600)
	}
	return tombstone
}
