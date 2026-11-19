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
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/prompt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
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

var presetSlotLabels = map[string]string{
	"batch.assets-h3-skill":                "批量资产智能提取",
	"batch.character-meta":                 "人物资产补充信息",
	"batch.comic-assets":                   "漫画资产统一提取",
	"batch.comic-image-derivation":         "漫画图片描述词推导",
	"batch.comic-page":                     "漫画分镜与页面生成",
	"batch.comic-reference-scene":          "漫画场景参考图库",
	"batch.comic-style":                    "漫画画风库",
	"batch.comic.assets.character":         "漫画人物资产推导",
	"batch.comic.assets.prop":              "漫画道具资产推导",
	"batch.comic.assets.scene":             "漫画场景资产推导",
	"batch.comic.character-image-template": "漫画角色设定图",
	"batch.hook-adaptation":                "开头剧情改编",
	"batch.opening-meta":                   "开篇信息提取",
	"batch.original-director":              "原文导演拆分",
	"batch.prefix":                         "视频画面风格前缀",
	"batch.prop-meta":                      "道具补充信息",
	"batch.scene-meta":                     "场景补充信息",
	"batch.video-meta":                     "视频分镜提示词",
	"batch.viral-director":                 "爆款导演拆分",
	"batch.visual-meta":                    "画面视觉提示词",
	"novel-fetch.hook":                     "小说开头优化",
	"novel-fetch.induce":                   "小说导读提取",
	"novel.analysis":                       "小说分析",
	"novel.character":                      "小说人物提取",
	"novel.outline":                        "小说大纲提取",
	"script.asset-extraction":              "剧本资产提取",
	"script.audio-match":                   "剧本配音匹配",
	"script.card.protocol":                 "剧本卡片输出规范",
	"script.character-focus":               "角色重点设定",
	"script.constraint.negative":           "通用负面提示词",
	"script.constraint.prefix":             "通用画面前缀词",
	"script.constraint.quality":            "通用画质约束",
	"script.constraint.restriction":        "通用画面限制",
	"script.constraint.wrapper":            "通用约束组合规则",
	"script.continuous":                    "连续剧本生成",
	"script.director.master":               "导演总控规则",
	"script.extract":                       "剧本内容提取",
	"script.format.q版":                     "Q版剧本格式",
	"script.format.screenplay":             "影视剧本格式",
	"script.format.shortdrama":             "短剧剧本格式",
	"script.format.shotlist":               "分镜清单格式",
	"script.format.storyboard":             "故事板格式",
	"script.general":                       "通用剧本生成",
	"script.hook":                          "剧本开头设计",
	"script.novel-panel-extract":           "小说面板提取",
	"script.quick-director":                "快速导演拆分",
	"script.segmented":                     "分段剧本生成",
	"shuihuo.asset.binding":                "水货资产绑定",
	"shuihuo.asset.character-extraction":   "水货人物提取",
	"shuihuo.asset.character-sheet":        "水货角色设定图",
	"shuihuo.asset.extraction":             "水货资产提取",
	"shuihuo.asset.prop-extraction":        "水货道具提取",
	"shuihuo.asset.scene-extraction":       "水货场景提取",
	"shuihuo.prompt.image":                 "水货图片提示词",
	"shuihuo.prompt.negative":              "水货负面提示词",
	"shuihuo.prompt.video":                 "水货视频提示词",
	"shuihuo.segmentation.smart":           "水货智能分段",
}

func presetSlotLabel(slot string) string {
	if label := presetSlotLabels[strings.TrimSpace(slot)]; label != "" {
		return label
	}
	return "未命名调用位置"
}

func RegisterAdminReadCompat(root *http.ServeMux, middleware BridgeAuth, accounts auth.Store, profiles *MemberProfileStore, catalog *prompt.Catalog, files *localartifact.Store) {
	if root == nil || accounts == nil || profiles == nil {
		return
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
		if req.URL.Query().Get("module") == prompt.LuoshuiAdminModule {
			items = prompt.MergeLuoshuiAdminPresets(items)
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
			if base, ok := prompt.LuoshuiAdminPreset(req.PathValue("id"), version); ok {
				writeJSON(w, http.StatusOK, map[string]any{"preset": base})
				return
			}
			writeAdminStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"preset": item})
	})))
	root.Handle("DELETE /api/admin/presets/{id}/{version}", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if adminStore == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "提示词目录尚未接入 Go 存储"})
			return
		}
		identity, _ := BridgeIdentityFromContext(req.Context())
		version, _ := strconv.Atoi(req.PathValue("version"))
		item, err := adminStore.Delete(identity.Username, req.PathValue("id"), version)
		if err != nil {
			if _, ok := prompt.LuoshuiAdminPreset(req.PathValue("id"), version); ok {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "这是内嵌 CM 原文，不能删除；保存覆盖版本后删除覆盖即可回退到原文"})
				return
			}
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
	for _, action := range []string{"publish", "rollback", "archive"} {
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
			} else if action == "rollback" {
				item, err = adminStore.Rollback(identity.Username, req.PathValue("id"), int(version))
			} else {
				item, err = adminStore.Archive(identity.Username, req.PathValue("id"), int(version))
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
		if req.URL.Query().Get("module") == novelprompt.AdminModule {
			writeJSON(w, http.StatusOK, map[string]any{"slots": novelprompt.AdminSlots()})
			return
		}
		if req.URL.Query().Get("module") == prompt.LuoshuiAdminModule {
			writeJSON(w, http.StatusOK, map[string]any{"slots": prompt.LuoshuiPluginSlots()})
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
			slots = append(slots, map[string]string{"id": *item.Slot, "label": presetSlotLabel(*item.Slot), "mode": "primary"})
		}
		writeJSON(w, http.StatusOK, map[string]any{"slots": slots})
	})))
	root.Handle("POST /api/admin/comic-materials/upload", adminPresetAuth(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if files == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "漫画素材存储尚未配置"})
			return
		}
		input, ok := decodeAdminJSON(w, req, 24<<20)
		if !ok {
			return
		}
		assetID, err := saveComicMaterialImage(files, toString(input["dataUrl"]))
		if err != nil {
			writeComicMaterialUploadError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"assetId": assetID})
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
