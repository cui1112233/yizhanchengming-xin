package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/prompt"
)

func adminSystemFile(systemDir, name string) string {
	paths := []string{filepath.Join(systemDir, name), filepath.Join(systemDir, "..", "..", "..", "data", "system", name)}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return paths[0]
}

func RegisterAdminReadCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore, catalogs ...*prompt.Catalog) {
	if root == nil || accounts == nil || profiles == nil {
		return
	}
	var catalog *prompt.Catalog
	if len(catalogs) > 0 {
		catalog = catalogs[0]
	}
	var adminStore *prompt.AdminStore
	if catalog != nil {
		adminStore = prompt.NewAdminStore(catalog.Path())
	}
	adminPresetAuth := func(next http.Handler) http.Handler {
		return middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			self := profiles.member(identity.Username, accounts)
			if self == nil || toString(self["role"]) != "dev" {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
				return
			}
			next.ServeHTTP(w, req)
		}))
	}
	root.Handle("GET /api/admin/presets", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if adminStore == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		items, err := adminStore.List(req.URL.Query().Get("module"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"presets": items})
	})))
	root.Handle("GET /api/admin/presets/{id}/{version}", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if adminStore == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		version, _ := strconv.Atoi(req.PathValue("version"))
		item, err := adminStore.Get(req.PathValue("id"), version)
		if err != nil {
			writeAdminStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"preset": item})
	})))
	root.Handle("POST /api/admin/presets/draft", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if adminStore == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		input, ok := decodeAdminJSON(w, req, 256<<10)
		if !ok {
			return
		}
		item, err := adminStore.CreateDraft(identity.Username, input)
		if err != nil {
			writeAdminStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"preset": item})
	})))
	root.Handle("PUT /api/admin/presets/{id}/{version}/draft", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if adminStore == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		input, ok := decodeAdminJSON(w, req, 256<<10)
		if !ok {
			return
		}
		version, _ := strconv.Atoi(req.PathValue("version"))
		item, err := adminStore.UpdateDraft(identity.Username, req.PathValue("id"), version, input)
		if err != nil {
			writeAdminStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"preset": item})
	})))
	for _, action := range []string{"publish", "rollback"} {
		action := action
		root.Handle("POST /api/admin/presets/{id}/"+action, adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if adminStore == nil {
				writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
				return
			}
			identity, _ := BridgeIdentityFromContext(req.Context())
			input, ok := decodeAdminJSON(w, req, 16<<10)
			if !ok {
				return
			}
			version, ok := input["version"].(float64)
			if !ok {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少版本号"})
				return
			}
			var item map[string]any
			var err error
			if action == "publish" {
				item, err = adminStore.Publish(identity.Username, req.PathValue("id"), int(version))
			} else {
				item, err = adminStore.Rollback(identity.Username, req.PathValue("id"), int(version))
			}
			if err != nil {
				writeAdminStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"preset": item})
		})))
	}
	root.Handle("GET /api/admin/preset-slots", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if catalog == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		items, err := catalog.All(req.URL.Query().Get("module"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		seen := map[string]bool{}
		slots := make([]map[string]string, 0)
		for _, item := range items {
			if item.Slot == nil || strings.TrimSpace(*item.Slot) == "" || seen[*item.Slot] {
				continue
			}
			seen[*item.Slot] = true
			slots = append(slots, map[string]string{"id": *item.Slot, "label": *item.Slot, "mode": "primary"})
		}
		writeJSON(w, http.StatusOK, map[string]any{"slots": slots})
	})))
	root.Handle("GET /api/admin/error-logs", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
			return
		}
		limit := 100
		if value, err := strconv.Atoi(req.URL.Query().Get("limit")); err == nil && value > 0 {
			limit = value
		}
		if limit > 500 {
			limit = 500
		}
		entries := []map[string]any{}
		if raw, err := os.ReadFile(adminSystemFile(profiles.SystemDir, "error-logs.json")); err == nil {
			_ = json.Unmarshal(raw, &entries)
		}
		if len(entries) > limit {
			entries = entries[len(entries)-limit:]
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
	})))
	root.Handle("GET /api/admin/accounts", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
			return
		}
		lister, ok := accounts.(accountLister)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "账号目录尚未接入 Go 存储"})
			return
		}
		rows, err := lister.ListAccounts(req.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "账号目录读取失败"})
			return
		}
		result := make([]map[string]any, 0, len(rows))
		for _, account := range rows {
			result = append(result, managedAccountRow(account, managedProfile(profiles, account.Username)))
		}
		writeJSON(w, http.StatusOK, map[string]any{"accounts": result})
	})))
	root.Handle("GET /api/admin/agent-skills", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"skills": readAgentSkills(adminSystemFile(profiles.SystemDir, "agent-skills.json"))})
	})))
	root.Handle("GET /api/admin/agent-skills/{id}/{version}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		self := profiles.member(identity.Username, accounts)
		if self == nil || toString(self["role"]) != "dev" {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
			return
		}
		version, _ := strconv.Atoi(req.PathValue("version"))
		for _, skill := range readAgentSkills(adminSystemFile(profiles.SystemDir, "agent-skills.json")) {
			if toString(skill["id"]) == req.PathValue("id") && intValue(skill["version"]) == version {
				writeJSON(w, http.StatusOK, map[string]any{"skill": skill})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
	})))
	root.Handle("POST /api/admin/agent-skills/draft", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := adminSkillOwner(w, req, accounts, profiles)
		if !ok {
			return
		}
		var input map[string]any
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 80<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "技能内容不合法"})
			return
		}
		id := strings.TrimSpace(toString(input["id"]))
		if !agentSkillIDPattern.MatchString(id) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "技能内容不合法"})
			return
		}
		path := adminSystemFile(profiles.SystemDir, "agent-skills.json")
		skills := readAgentSkills(path)
		version := 1
		for _, skill := range skills {
			if toString(skill["id"]) == id && intValue(skill["version"]) >= version {
				version = intValue(skill["version"]) + 1
			}
		}
		skill, err := normalizeAgentSkill(input, identity.Username, id)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		skill["source"] = "system"
		skill["owner"] = nil
		skill["status"] = "draft"
		skill["version"] = version
		skill["createdBy"] = identity.Username
		skills = append(skills, skill)
		if err := writeAgentSkills(path, skills); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存技能失败"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"skill": publicAgentSkill(skill)})
	})))
	for _, status := range []string{"published", "archived"} {
		status := status
		root.Handle("POST /api/admin/agent-skills/{id}/"+status, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := adminSkillOwner(w, req, accounts, profiles)
			if !ok {
				return
			}
			var input map[string]any
			_ = json.NewDecoder(http.MaxBytesReader(w, req.Body, 16<<10)).Decode(&input)
			version := intValue(input["version"])
			path := adminSystemFile(profiles.SystemDir, "agent-skills.json")
			skills := readAgentSkills(path)
			found := false
			for index, skill := range skills {
				if toString(skill["id"]) == req.PathValue("id") && intValue(skill["version"]) == version {
					found = true
					if status == "published" {
						for otherIndex, other := range skills {
							if otherIndex != index && toString(other["id"]) == req.PathValue("id") && toString(other["status"]) == "published" {
								other["status"] = "archived"
								skills[otherIndex] = other
							}
						}
					}
					skill["status"] = status
					skill["publishedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
					skill["publishedBy"] = identity.Username
					skills[index] = skill
					break
				}
			}
			if !found {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
				return
			}
			if err := writeAgentSkills(path, skills); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存技能失败"})
				return
			}
			for _, skill := range skills {
				if toString(skill["id"]) == req.PathValue("id") && intValue(skill["version"]) == version {
					writeJSON(w, http.StatusOK, map[string]any{"skill": publicAgentSkill(skill)})
					return
				}
			}
		})))
	}
}

func decodeAdminJSON(w http.ResponseWriter, req *http.Request, limit int64) (map[string]any, bool) {
	var input map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, limit)).Decode(&input); err != nil || input == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容不合法"})
		return nil, false
	}
	return input, true
}

func writeAdminStoreError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if typed, ok := err.(*prompt.StoreError); ok {
		switch typed.Code {
		case "invalid":
			status = http.StatusBadRequest
		case "not_found":
			status = http.StatusNotFound
		case "conflict":
			status = http.StatusConflict
		}
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func adminSkillOwner(w http.ResponseWriter, req *http.Request, accounts auth.Store, profiles *MemberProfileStore) (BridgeIdentity, bool) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return BridgeIdentity{}, false
	}
	self := profiles.member(identity.Username, accounts)
	if self == nil || toString(self["role"]) != "dev" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Forbidden"})
		return BridgeIdentity{}, false
	}
	return identity, true
}
