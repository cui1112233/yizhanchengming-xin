package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

func RegisterPublicTeamInviteCompat(root *http.ServeMux, systemDir string, accounts auth.Store, profiles *MemberProfileStore) {
	if root == nil || strings.TrimSpace(systemDir) == "" || accounts == nil || profiles == nil {
		return
	}
	root.Handle("GET /api/login/invite/{token}", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := strings.TrimSpace(req.PathValue("token"))
		digestBytes := sha256.Sum256([]byte(token))
		digest := hex.EncodeToString(digestBytes[:])
		state := map[string]any{"teams": []any{}, "invites": []any{}}
		if raw, err := os.ReadFile(adminSystemFile(systemDir, "team-collaboration.json")); err == nil {
			_ = json.Unmarshal(raw, &state)
		}
		var found map[string]any
		for _, item := range arrayOrEmpty(state["invites"]) {
			row := mapOrEmpty(item)
			if toString(row["tokenDigest"]) == digest {
				found = row
				break
			}
		}
		if found == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "邀请不存在"})
			return
		}
		expires, _ := time.Parse(time.RFC3339, toString(found["expiresAt"]))
		expired := !expires.IsZero() && time.Now().UTC().After(expires)
		used := strings.TrimSpace(toString(found["usedAt"])) != ""
		managerName := toString(found["managerUsername"])
		manager := profiles.member(managerName, accounts)
		teamName := ""
		for _, item := range arrayOrEmpty(state["teams"]) {
			team := mapOrEmpty(item)
			if toString(team["id"]) == toString(found["teamId"]) {
				teamName = toString(team["name"])
				break
			}
		}
		invite := map[string]any{"id": found["id"], "teamId": found["teamId"], "teamName": teamName, "managerUsername": managerName, "createdAt": found["createdAt"], "expiresAt": found["expiresAt"], "monthlyTokenLimit": found["monthlyTokenLimit"], "apiScopes": found["apiScopes"], "usedAt": found["usedAt"], "usedBy": found["usedBy"], "available": !expired && !used, "expired": expired}
		var managerPublic map[string]any
		if manager != nil {
			managerPublic = map[string]any{"username": managerName, "displayName": manager["displayName"]}
		}
		writeJSON(w, http.StatusOK, map[string]any{"invite": invite, "manager": managerPublic})
	}))
}
