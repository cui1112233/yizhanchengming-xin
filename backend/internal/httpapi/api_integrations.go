package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"qiantie/backend/internal/apisecrets"
	"qiantie/backend/internal/auth"
)

// 「API 接口」配置的独立存储键。它只写进用户自己的 api-config.json / MySQL
// settings 里的 apiIntegrations 字段，不碰 modelCatalog，也不碰其他设置。
const apiIntegrationsSettingKey = "apiIntegrations"

// apiIntegrationFields 是服务端白名单：key -> 是否密钥（密钥永不下发明文）。
// 与前端 apiIntegrationSchema.js 的字段 key 保持一致；新增字段必须两边同时加。
var apiIntegrationFields = map[string]bool{
	"openrouterApiKey":       true,
	"kimiOfficialApiKey":     true,
	"deepseekOfficialApiKey": true,
	"doubaoOfficialApiKey":   true,
	"glmOfficialApiKey":      true,
	"qwenOfficialApiKey":     true,
	"mimoOfficialApiKey":     true,

	"seedance2Api": true, "viduApi": true, "runninghubApi": true, "runninghubConcurrency": false,
	"wanxApiKey": true, "localGrokVideoBaseUrl": false, "localGrokVideoApiKey": true,
	"xiaoyunqueAccessKey": true, "xiaoyunqueConcurrency": false, "xiaoyunqueIntlAccessKey": true,
	"xiaoyunqueIntlConcurrency": false, "happyHorseKey": true, "minimaxH3Key": true,

	"localGrokImageBaseUrl": false, "localGrokImageApiKey": true, "localBananaBaseUrl": false,
	"localBananaApiKey": true, "qwenImageApi": true, "zImageApi": true, "volcanoSeedreamApi": true,
	"nano2BaseUrl": false, "nano2ApiKey": true, "grsaiImage2Key": true, "geeknowImageApi": true,

	"jimengBrowserEnabled": false, "jimengConcurrency": false, "jimengTimeoutHours": false,

	"realpersonChannel": false, "deyunaiRegion": false, "ydyAccessKey": true, "ydySecretKey": true,
	"ydyPoolId": false, "gumenApiToken": true, "gumenGroupId": false, "xiaoxiaoModel": false,
	"volcanoProjectName": false, "volcanoAccessKeyId": true, "volcanoSecretAccessKey": true,

	"viduBrowserEnabled": false,

	"subtitleProvider": false, "volcanoEraseVersion": false, "subtitleConcurrency": false,
	"kaipaiAccessKey": true, "kaipaiSecretKey": true, "w550UserNo": false, "w550ApiKey": true,
	"volcanoMediaKitApiKey": true,

	"conchApiUrl": true, "niji7KongyangKey": true, "niji7Concurrency": false, "niji7Style": false,
	"imageHostUrl": true,

	"ttsServiceUrl": false, "ttsApiKey": true, "qwenTtsKey": true, "cosyVoiceKey": true,
	"minimaxApiKey": true, "ttsSentenceLength": false,
}

var apiIntegrationNumberFields = map[string]bool{
	"runninghubConcurrency": true, "xiaoyunqueConcurrency": true, "xiaoyunqueIntlConcurrency": true,
	"jimengConcurrency": true, "jimengTimeoutHours": true, "subtitleConcurrency": true,
	"niji7Concurrency": true, "ttsSentenceLength": true,
}

var apiIntegrationBoolFields = map[string]bool{
	"jimengBrowserEnabled": true, "viduBrowserEnabled": true,
}

// apiIntegrationChannelGroups 是允许保存渠道勾选结果的分组白名单。
// 渠道之间没有“默认使用谁”的概念，这里只保存每组勾选了哪些渠道 id。
var apiIntegrationChannelGroups = map[string]bool{
	"text": true, "video": true, "image": true, "jimeng": true, "realperson": true,
	"vidu": true, "subtitle": true, "storage": true, "tts": true,
}

// apiIntegrationChannelSelection 统一成 { enabled: [...] }：
// 兼容早期单值写法（字符串），并丢弃未知分组与非法 id。
func apiIntegrationChannelSelection(stored map[string]any) map[string]any {
	raw, ok := stored["channelSelection"].(map[string]any)
	if !ok || raw == nil {
		return map[string]any{}
	}
	result := map[string]any{}
	for group, value := range raw {
		if !apiIntegrationChannelGroups[group] {
			continue
		}
		enabled := normalizeChannelSelectionValue(value)
		// 对象形式的空 enabled 表示“用户显式取消勾选”，必须保留，
		// 否则刷新后前端会回落到默认渠道，空状态无法持久。
		if len(enabled) == 0 && !isExplicitChannelSelection(value) {
			continue
		}
		if enabled == nil {
			enabled = []string{}
		}
		result[group] = map[string]any{"enabled": enabled}
	}
	return result
}

// isExplicitChannelSelection 判断值是否为带 enabled 键的对象：
// 只有这种写法才代表用户主动提交了勾选结果（可能为空）。
func isExplicitChannelSelection(value any) bool {
	typed, ok := value.(map[string]any)
	if !ok {
		return false
	}
	_, ok = typed["enabled"]
	return ok
}

func normalizeChannelSelectionValue(value any) []string {
	switch typed := value.(type) {
	case string:
		id := strings.TrimSpace(typed)
		if id == "" || len(id) > 64 {
			return nil
		}
		return []string{id}
	case map[string]any:
		enabled := []string{}
		seen := map[string]bool{}
		switch list := typed["enabled"].(type) {
		case []any:
			for _, item := range list {
				enabled = appendChannelID(enabled, seen, item)
			}
		case []string:
			for _, item := range list {
				enabled = appendChannelID(enabled, seen, item)
			}
		}
		if len(enabled) == 0 {
			return nil
		}
		return enabled
	}
	return nil
}

func appendChannelID(list []string, seen map[string]bool, item any) []string {
	id, ok := item.(string)
	if !ok {
		return list
	}
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 64 || seen[id] || len(list) >= 20 {
		return list
	}
	seen[id] = true
	return append(list, id)
}

func apiIntegrationStoredMap(config map[string]any) map[string]any {
	raw, ok := config[apiIntegrationsSettingKey].(map[string]any)
	if !ok || raw == nil {
		return map[string]any{}
	}
	return raw
}

func apiIntegrationValues(stored map[string]any) map[string]any {
	values, ok := stored["values"].(map[string]any)
	if !ok || values == nil {
		return map[string]any{}
	}
	return values
}

func apiIntegrationPinned(stored map[string]any) []string {
	raw, ok := stored["pinned"].([]any)
	if !ok {
		if list, ok := stored["pinned"].([]string); ok {
			return list
		}
		return []string{}
	}
	result := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, item := range raw {
		key, _ := item.(string)
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		if _, allowed := apiIntegrationFields[key]; !allowed {
			continue
		}
		seen[key] = true
		result = append(result, key)
	}
	return result
}

func apiIntegrationChannels(stored map[string]any) []any {
	raw, ok := stored["customChannels"].([]any)
	if !ok {
		return []any{}
	}
	result := make([]any, 0, len(raw))
	for _, item := range raw {
		channel, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, channel)
	}
	return result
}

// apiIntegrationsView 是下发给浏览器的唯一形态：
// 非密钥字段给值，密钥字段只给 configured 布尔；自定义渠道同理只给 hasApiKey。
func apiIntegrationsView(stored map[string]any) map[string]any {
	values := apiIntegrationValues(stored)
	publicValues := map[string]any{}
	configured := map[string]bool{}
	for key, secret := range apiIntegrationFields {
		value, ok := values[key]
		if !ok || isEmptyIntegrationValue(value) {
			continue
		}
		configured[key] = true
		if !secret {
			publicValues[key] = value
		}
	}
	channels := apiIntegrationChannels(stored)
	publicChannels := make([]map[string]any, 0, len(channels))
	for _, item := range channels {
		channel, _ := item.(map[string]any)
		public := map[string]any{}
		for _, field := range []string{"id", "name", "baseUrl", "model", "concurrency", "apiFormat", "fullUrl", "enabled", "group"} {
			if value, ok := channel[field]; ok {
				public[field] = value
			}
		}
		key, _ := channel["apiKey"].(string)
		public["hasApiKey"] = strings.TrimSpace(key) != ""
		publicChannels = append(publicChannels, public)
	}
	return map[string]any{
		"values":           publicValues,
		"configured":       configured,
		"pinned":           apiIntegrationPinned(stored),
		"initialized":      len(stored) > 0,
		"customChannels":   publicChannels,
		"channelSelection": apiIntegrationChannelSelection(stored),
	}
}

func isEmptyIntegrationValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	}
	return false
}

func coerceIntegrationValue(key string, value any) (any, bool) {
	if apiIntegrationBoolFields[key] {
		if _, ok := value.(bool); ok {
			return value, true
		}
		return nil, false
	}
	if apiIntegrationNumberFields[key] {
		switch typed := value.(type) {
		case float64:
			return typed, true
		case json.Number:
			if number, err := typed.Float64(); err == nil {
				return number, true
			}
		case int:
			return float64(typed), true
		case string:
			trimmed := strings.TrimSpace(typed)
			if trimmed == "" {
				return nil, true
			}
			var number json.Number = json.Number(trimmed)
			if parsed, err := number.Float64(); err == nil {
				return parsed, true
			}
		}
		return nil, false
	}
	text, ok := value.(string)
	if !ok {
		return nil, false
	}
	if len(text) > 8192 {
		return nil, false
	}
	return strings.TrimSpace(text), true
}

func sanitizeIntegrationChannel(input map[string]any, previous map[string]any) (map[string]any, bool) {
	channel := map[string]any{}
	id, _ := input["id"].(string)
	id = strings.TrimSpace(id)
	if id == "" {
		buffer := make([]byte, 6)
		if _, err := rand.Read(buffer); err != nil {
			return nil, false
		}
		id = "ch_" + hex.EncodeToString(buffer)
	}
	channel["id"] = id

	name, _ := input["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 64 {
		return nil, false
	}
	channel["name"] = name

	baseURL, _ := input["baseUrl"].(string)
	baseURL = strings.TrimSpace(baseURL)
	if len(baseURL) > 1024 {
		return nil, false
	}
	channel["baseUrl"] = baseURL

	model, _ := input["model"].(string)
	model = strings.TrimSpace(model)
	if len(model) > 256 {
		return nil, false
	}
	channel["model"] = model

	switch typed := input["concurrency"].(type) {
	case float64:
		if typed < 1 || typed > 100 {
			return nil, false
		}
		channel["concurrency"] = typed
	case nil:
		channel["concurrency"] = nil
	default:
		return nil, false
	}

	enabled := true
	if value, ok := input["enabled"].(bool); ok {
		enabled = value
	}
	channel["enabled"] = enabled

	// group 为空表示「自定义」独立分组里的通用渠道；有值时表示挂在某个
	// 具体分组（例如 text）下、由该分组的「自定义」选项使用。
	group, _ := input["group"].(string)
	group = strings.TrimSpace(group)
	if len(group) > 32 {
		return nil, false
	}
	channel["group"] = group

	apiFormat, _ := input["apiFormat"].(string)
	apiFormat = strings.TrimSpace(apiFormat)
	if apiFormat == "" && previous != nil {
		apiFormat, _ = previous["apiFormat"].(string)
		apiFormat = strings.TrimSpace(apiFormat)
	}
	switch apiFormat {
	case "", "openai_compatible":
		channel["apiFormat"] = "openai_compatible"
	case "custom":
		channel["apiFormat"] = "custom"
	default:
		return nil, false
	}
	if fullURL, ok := input["fullUrl"].(bool); ok {
		channel["fullUrl"] = fullURL
	} else if previous != nil {
		if previousFullURL, ok := previous["fullUrl"].(bool); ok {
			channel["fullUrl"] = previousFullURL
		}
	}

	if clear, ok := input["clearApiKey"].(bool); ok && clear {
		// 显式清空 Key：渠道还在，但因为没凭证不会进入模型下拉。
		channel["apiKey"] = ""
	} else if apiKey, ok := input["apiKey"].(string); ok && strings.TrimSpace(apiKey) != "" {
		if len(apiKey) > 8192 {
			return nil, false
		}
		channel["apiKey"] = strings.TrimSpace(apiKey)
	} else if previous != nil {
		if apiKey, ok := previous["apiKey"]; ok {
			channel["apiKey"] = apiKey
		}
	}
	return channel, true
}

func registerAPIIntegrationRoutes(root *http.ServeMux, middleware BridgeAuth, store *UserConfigStore, accounts auth.Store, profiles ...*MemberProfileStore) {
	if root == nil || middleware.AuthStore == nil || store == nil || accounts == nil {
		return
	}
	var profileStore *MemberProfileStore
	if len(profiles) > 0 {
		profileStore = profiles[0]
	}
	canManage := func(req *http.Request) (BridgeIdentity, bool) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			return BridgeIdentity{}, false
		}
		allowed, _ := configManagementState(identity, profileStore, accounts)
		return identity, allowed
	}

	root.Handle("GET /api/settings/api-integrations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		config := store.read(identity.Username)
		store.mu.Unlock()
		writeJSON(w, http.StatusOK, apiIntegrationsView(apiIntegrationStoredMap(config)))
	})))

	root.Handle("POST /api/settings/api-integrations", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, allowed := canManage(req)
		if identity.Username == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !allowed {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "没有权限修改 API 接口配置"})
			return
		}
		// 用指针区分「没提交这个键」和「提交了空值」：输入类字段自动保存时
		// 只带 values 一个键，未提交的 pinned / customChannels 必须保持原值，
		// 不能被当成整组覆盖。
		var payload struct {
			Values           map[string]any    `json:"values"`
			Pinned           *[]string         `json:"pinned"`
			CustomChannels   *[]map[string]any `json:"customChannels"`
			ChannelSelection map[string]any    `json:"channelSelection"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 512<<10)).Decode(&payload) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		config := store.read(identity.Username)
		stored := apiIntegrationStoredMap(config)
		values := apiIntegrationValues(stored)
		nextValues := map[string]any{}
		for key, value := range values {
			nextValues[key] = value
		}
		for key, value := range payload.Values {
			secret, allowedKey := apiIntegrationFields[key]
			if !allowedKey {
				continue
			}
			if value == nil {
				delete(nextValues, key)
				continue
			}
			coerced, ok := coerceIntegrationValue(key, value)
			if !ok {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字段内容无效：" + key})
				return
			}
			if secret {
				if text, _ := coerced.(string); text == "" {
					delete(nextValues, key)
					continue
				}
			}
			nextValues[key] = coerced
		}
		previousChannels := apiIntegrationChannels(stored)
		previousByID := map[string]map[string]any{}
		for _, item := range previousChannels {
			if channel, ok := item.(map[string]any); ok {
				if id, _ := channel["id"].(string); id != "" {
					previousByID[id] = channel
				}
			}
		}
		channels := make([]any, 0, len(previousChannels))
		if payload.CustomChannels != nil {
			channels = make([]any, 0, len(*payload.CustomChannels))
			for _, input := range *payload.CustomChannels {
				id, _ := input["id"].(string)
				channel, ok := sanitizeIntegrationChannel(input, previousByID[id])
				if !ok {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "自定义渠道内容无效"})
					return
				}
				channels = append(channels, channel)
			}
		} else {
			channels = previousChannels
		}
		next := map[string]any{"values": nextValues, "pinned": apiIntegrationPinned(stored), "customChannels": channels}
		if payload.Pinned != nil {
			next["pinned"] = apiIntegrationPinned(map[string]any{"pinned": toAnySlice(*payload.Pinned)})
		}
		// 渠道选择与字段值分开合并：只覆盖本次提交的分组，其余分组保持不变。
		selection := apiIntegrationChannelSelection(stored)
		for group, value := range payload.ChannelSelection {
			if !apiIntegrationChannelGroups[group] {
				continue
			}
			enabled := normalizeChannelSelectionValue(value)
			if len(enabled) == 0 && !isExplicitChannelSelection(value) {
				delete(selection, group)
				continue
			}
			if enabled == nil {
				enabled = []string{}
			}
			selection[group] = map[string]any{"enabled": enabled}
		}
		next["channelSelection"] = selection
		config[apiIntegrationsSettingKey] = next
		if err := store.write(identity.Username, config); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "配置保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, apiIntegrationsView(next))
	})))

	root.Handle("POST /api/settings/api-integrations/reveal", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, allowed := canManage(req)
		if identity.Username == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !allowed {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "没有权限查看密钥"})
			return
		}
		var payload struct {
			Key string `json:"key"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<10)).Decode(&payload) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		key := strings.TrimSpace(payload.Key)
		store.mu.Lock()
		stored := apiIntegrationStoredMap(store.read(identity.Username))
		store.mu.Unlock()
		if strings.HasPrefix(key, "custom:") {
			id := strings.TrimPrefix(key, "custom:")
			for _, item := range apiIntegrationChannels(stored) {
				channel, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if channelID, _ := channel["id"].(string); channelID != id {
					continue
				}
				apiKey, _ := channel["apiKey"].(string)
				if apisecrets.IsEncrypted(apiKey) {
					// 解密失败时绝不把密文整段回显给前端。
					writeJSON(w, http.StatusConflict, map[string]string{"error": "该渠道密钥无法解密，请重新填写后保存"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"key": key, "value": apiKey})
				return
			}
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "渠道不存在"})
			return
		}
		if secret, ok := apiIntegrationFields[key]; !ok || !secret {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "字段不支持查看"})
			return
		}
		value, _ := apiIntegrationValues(stored)[key].(string)
		if apisecrets.IsEncrypted(value) {
			// 解密失败时绝不把密文整段回显给前端。
			writeJSON(w, http.StatusConflict, map[string]string{"error": "该密钥无法解密，请重新填写后保存"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"key": key, "value": value})
	})))
}

func toAnySlice(list []string) []any {
	result := make([]any, 0, len(list))
	for _, item := range list {
		result = append(result, item)
	}
	return result
}
