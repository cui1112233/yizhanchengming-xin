package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

var modelKinds = map[string]bool{"text": true, "video": true, "image": true, "audio": true}

func registerModelCatalogRoutes(root *http.ServeMux, auth NovelFetchBridgeAuth, store novelfetchworkshop.Store) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/test/text", func(w http.ResponseWriter, req *http.Request) {
		testConfiguredModel(w, req, store, "text")
	})
	mux.HandleFunc("POST /api/test/image", func(w http.ResponseWriter, req *http.Request) {
		testConfiguredModel(w, req, store, "image")
	})
	mux.HandleFunc("GET /api/models", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		models, err := readModelCatalog(req, store, identity.Username, req.URL.Query().Get("kind"), identity.IsOwner)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": models})
	})
	mux.HandleFunc("GET /api/config/models", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		models, err := readModelCatalog(req, store, identity.Username, req.URL.Query().Get("kind"), true)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": models})
	})
	mux.HandleFunc("POST /api/config/models", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		model, err := normalizeManagedModel(input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		settings, _ := store.GetConfig(req.Context(), identity.Username)
		catalog := modelCatalog(settings)
		for _, current := range catalog {
			if current["id"] == model["id"] {
				writeJSON(w, http.StatusConflict, map[string]any{"error": "模型 ID 已存在"})
				return
			}
		}
		catalog = append(catalog, model)
		settings["modelCatalog"] = catalog
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存模型失败"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"model": publicManagedModel(model)})
	})
	mux.HandleFunc("POST /api/config/models/test", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		kind, _ := input["kind"].(string)
		baseURL, _ := input["baseUrl"].(string)
		modelID, _ := input["modelId"].(string)
		credential, _ := input["credential"].(string)
		existingModelID, _ := input["existingModelId"].(string)
		// The settings page intentionally omits an already-saved credential when
		// editing a model. Reuse that secret only for this server-side probe; it
		// is never returned to the browser.
		if strings.TrimSpace(credential) == "" && strings.TrimSpace(existingModelID) != "" {
			if settings, settingsErr := store.GetConfig(req.Context(), identity.Username); settingsErr == nil {
				for _, existing := range modelCatalog(settings) {
					if existing["id"] != strings.TrimSpace(existingModelID) {
						continue
					}
					if savedCredential, ok := existing["credential"].(string); ok {
						credential = strings.TrimSpace(savedCredential)
					}
					if baseURL == "" {
						baseURL = firstText(existing["baseUrl"])
					}
					if modelID == "" {
						modelID = firstText(existing["modelId"])
					}
					break
				}
			}
		}
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		modelID = strings.TrimSpace(modelID)
		credential = strings.TrimSpace(credential)
		if kind != "text" || baseURL == "" || modelID == "" || credential == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "测试文本模型前请填写 Base URL、模型 ID 和 API Key"})
			return
		}
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "Base URL 格式不正确"})
			return
		}
		endpoint := baseURL
		if !strings.HasSuffix(parsed.Path, "/chat/completions") {
			endpoint += "/chat/completions"
		}
		payload, _ := json.Marshal(map[string]any{"model": modelID, "messages": []map[string]string{{"role": "user", "content": "Hi"}}, "max_tokens": 5})
		ctx, cancel := context.WithTimeout(req.Context(), 20*time.Second)
		defer cancel()
		probe, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "文本模型地址无效"})
			return
		}
		probe.Header.Set("Content-Type", "application/json")
		probe.Header.Set("Authorization", "Bearer "+credential)
		response, err := http.DefaultClient.Do(probe)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "文本模型连接失败：" + err.Error()})
			return
		}
		defer response.Body.Close()
		var decoded struct {
			Choices []struct {
				Message map[string]any `json:"message"`
			} `json:"choices"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&decoded)
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": fmt.Sprintf("文本模型连接失败（状态 %d），请检查接口、模型名或密钥", response.StatusCode)})
			return
		}
		if decodeErr != nil || len(decoded.Choices) == 0 || len(decoded.Choices[0].Message) == 0 {
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "文本模型未返回有效答复，请检查模型兼容性"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "文本模型连接成功，现在可以保存并启用。"})
	})
	mux.HandleFunc("POST /api/config/models/test-audio", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		model := map[string]any{
			"kind":            "audio",
			"endpoint":        strings.TrimSpace(fmt.Sprint(input["endpoint"])),
			"requestTemplate": strings.TrimSpace(fmt.Sprint(input["requestTemplate"])),
			"responseMapping": strings.TrimSpace(fmt.Sprint(input["responseMapping"])),
			"credential":      strings.TrimSpace(fmt.Sprint(input["credential"])),
		}
		existingModelID := strings.TrimSpace(fmt.Sprint(input["existingModelId"]))
		if existingModelID != "" {
			if settings, settingsErr := store.GetConfig(req.Context(), identity.Username); settingsErr == nil {
				for _, existing := range modelCatalog(settings) {
					if existing["id"] != existingModelID {
						continue
					}
					for _, key := range []string{"endpoint", "requestTemplate", "responseMapping", "credential"} {
						if strings.TrimSpace(fmt.Sprint(model[key])) == "" {
							model[key] = existing[key]
						}
					}
					break
				}
			}
		}
		if modelString(model, "endpoint") == "" || modelString(model, "requestTemplate") == "" || modelString(model, "responseMapping") == "" || shuihuoModelCredential(model) == "" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "测试音频模型前请填写服务地址、请求模板、返回映射和 API Key"})
			return
		}
		payload, contentType, err := requestConfiguredTTSBytes(req.Context(), ttsRequest{Input: "测试配音", Voice: "zh-CN-XiaoxiaoNeural", Speed: 1, Pitch: "0", Style: "general"}, model)
		if err != nil || len(payload) == 0 || strings.HasPrefix(strings.ToLower(contentType), "application/json") {
			if err == nil {
				err = errors.New("服务未返回音频文件")
			}
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "音频模型连接失败：" + err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "contentType": contentType, "message": "音频模型连接成功，现在可以保存并启用。"})
	})
	mux.HandleFunc("PATCH /api/config/models/{modelId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		var patch map[string]any
		if err := decodeNovelFetchJSON(w, req, &patch); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求内容无效"})
			return
		}
		settings, _ := store.GetConfig(req.Context(), identity.Username)
		catalog := modelCatalog(settings)
		found := false
		var updated map[string]any
		for index, current := range catalog {
			if current["id"] != req.PathValue("modelId") {
				continue
			}
			for key, value := range patch {
				if key == "id" {
					continue
				}
				// Older settings pages called the secret apiKey. Keep accepting
				// that spelling, but normalize it into the single credential field
				// before validation and persistence.
				if key == "apiKey" {
					if strings.TrimSpace(fmt.Sprint(value)) != "" {
						current["credential"] = value
					}
					continue
				}
				if key == "credential" && strings.TrimSpace(fmt.Sprint(value)) == "" {
					continue
				}
				current[key] = value
			}
			// Migrate a legacy stored apiKey before normalizeManagedModel checks
			// the H3 key requirement. The secret never leaves public responses.
			if strings.TrimSpace(fmt.Sprint(current["credential"])) == "" {
				if legacyKey := strings.TrimSpace(fmt.Sprint(current["apiKey"])); legacyKey != "" && legacyKey != "<nil>" {
					current["credential"] = legacyKey
				}
			}
			model, err := normalizeManagedModel(current)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
				return
			}
			catalog[index] = model
			updated = model
			found = true
			break
		}
		if !found {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "模型不存在"})
			return
		}
		settings["modelCatalog"] = catalog
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "保存模型失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"model": publicManagedModel(updated)})
	})
	mux.HandleFunc("DELETE /api/config/models/{modelId}", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		settings, _ := store.GetConfig(req.Context(), identity.Username)
		catalog := modelCatalog(settings)
		filtered := make([]map[string]any, 0, len(catalog))
		removed := false
		for _, model := range catalog {
			if model["id"] == req.PathValue("modelId") {
				removed = true
				continue
			}
			filtered = append(filtered, model)
		}
		if !removed {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "模型不存在"})
			return
		}
		settings["modelCatalog"] = filtered
		if err := store.PutConfig(req.Context(), identity.Username, settings); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "删除模型失败"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/config/models/quotas", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"quotas": []any{}})
	})
	mux.HandleFunc("GET /api/config/models/local-doubao-executor-video/pairing-status", func(w http.ResponseWriter, req *http.Request) {
		identity, _ := BridgeIdentityFromContext(req.Context())
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "无权管理模型"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"executorPaired": false})
	})
	secured := auth.Middleware(mux)
	root.Handle("/api/models", secured)
	root.Handle("/api/models/", secured)
	root.Handle("/api/config/models", secured)
	root.Handle("/api/config/models/", secured)
	root.Handle("/api/test/", secured)
}

func testConfiguredModel(w http.ResponseWriter, req *http.Request, store novelfetchworkshop.Store, kind string) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "kind": kind, "error": "unauthorized"})
		return
	}
	var input map[string]any
	if err := decodeNovelFetchJSON(w, req, &input); err != nil {
		return
	}
	if kind == "image" {
		if nested, ok := input["image"].(map[string]any); ok {
			input = nested
		}
	}
	settings, err := store.GetConfig(req.Context(), identity.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "kind": kind, "error": "无法读取当前模型配置，请稍后重试。"})
		return
	}
	baseURL := firstText(input["baseUrl"], input["base_url"])
	modelID := firstText(input["model"], input["modelId"], input["model_id"])
	credential := firstText(input["apiKey"], input["api_key"], input["credential"])
	if baseURL == "" || modelID == "" || credential == "" {
		for _, model := range modelCatalog(settings) {
			if model["kind"] != kind || model["enabled"] != true {
				continue
			}
			if baseURL == "" {
				baseURL = firstText(model["baseUrl"])
			}
			if modelID == "" {
				modelID = firstText(model["modelId"])
			}
			if credential == "" {
				credential = firstText(model["credential"])
			}
			break
		}
	}
	if baseURL == "" || modelID == "" || credential == "" {
		message := "请填写完整的文本服务地址、模型和 API Key。"
		if kind == "image" {
			message = "请填写完整的生图服务地址、模型和 API Key。"
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": kind, "error": message})
		return
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"ok": false, "kind": kind, "error": "Base URL 格式不正确"})
		return
	}
	client := &http.Client{Timeout: 20 * time.Second}
	if kind == "image" {
		testModelCatalogEndpoint(w, client, baseURL, modelID, credential)
		return
	}
	endpoint := baseURL
	if !strings.HasSuffix(parsed.Path, "/chat/completions") {
		endpoint += "/chat/completions"
	}
	payload, _ := json.Marshal(map[string]any{"model": modelID, "messages": []map[string]string{{"role": "user", "content": "Hi"}}, "max_tokens": 5})
	probe, err := http.NewRequestWithContext(req.Context(), http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": kind, "error": "文本模型地址无效"})
		return
	}
	probe.Header.Set("Content-Type", "application/json")
	probe.Header.Set("Authorization", "Bearer "+credential)
	response, err := client.Do(probe)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "模型服务连接失败，请检查服务地址、网络和代理设置。", "type": "upstream_connection_error"})
		return
	}
	defer response.Body.Close()
	var decoded map[string]any
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "上游文本模型未返回有效 JSON。"})
		return
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": fmt.Sprintf("上游模型服务请求失败（状态 %d）。", response.StatusCode), "type": "upstream_error"})
		return
	}
	choices, _ := decoded["choices"].([]any)
	if len(choices) == 0 {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": kind, "error": "上游文本模型未返回有效答复。"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": kind, "message": mapOrEmpty(mapOrEmpty(choices[0])["message"])})
}

func testModelCatalogEndpoint(w http.ResponseWriter, client *http.Client, baseURL, modelID, credential string) {
	endpoint := baseURL
	if !strings.HasSuffix(endpoint, "/models") {
		endpoint += "/models"
	}
	probe, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "kind": "image", "error": "生图服务地址无效"})
		return
	}
	probe.Header.Set("Authorization", "Bearer "+credential)
	response, err := client.Do(probe)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": "image", "error": "模型服务连接失败，请检查服务地址、网络和代理设置。", "type": "upstream_connection_error"})
		return
	}
	defer response.Body.Close()
	var decoded map[string]any
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "kind": "image", "error": fmt.Sprintf("上游模型服务请求失败（状态 %d）。", response.StatusCode), "type": "upstream_error"})
		return
	}
	listed := false
	if items, ok := decoded["data"].([]any); ok {
		for _, item := range items {
			if modelOrID := firstText(mapOrEmpty(item)["id"]); modelOrID == modelID {
				listed = true
				break
			}
		}
	}
	message := "生图服务连接成功，但当前模型未出现在模型目录中，请确认模型名称或供应商支持。"
	if listed {
		message = "生图模型连接成功，当前模型已在模型目录中找到。"
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "kind": "image", "modelListed": listed, "message": message})
}

func readModelCatalog(req *http.Request, store novelfetchworkshop.Store, owner, kind string, includeDisabled bool) ([]map[string]any, error) {
	if kind != "" && !modelKinds[kind] {
		return nil, errors.New("模型类型无效")
	}
	settings, err := store.GetConfig(req.Context(), owner)
	if err != nil {
		return nil, err
	}
	result := make([]map[string]any, 0)
	for _, model := range modelCatalog(settings) {
		if kind != "" && model["kind"] != kind {
			continue
		}
		if !includeDisabled && model["enabled"] != true {
			continue
		}
		result = append(result, publicManagedModel(model))
	}
	return result, nil
}

func modelCatalog(settings map[string]any) []map[string]any {
	raw, _ := settings["modelCatalog"].([]any)
	if typed, ok := settings["modelCatalog"].([]map[string]any); ok {
		return typed
	}
	catalog := make([]map[string]any, 0, len(raw))
	for _, value := range raw {
		if model, ok := value.(map[string]any); ok {
			catalog = append(catalog, model)
		}
	}
	return catalog
}

func normalizeManagedModel(input map[string]any) (map[string]any, error) {
	id := strings.TrimSpace(fmt.Sprint(input["id"]))
	kind := strings.TrimSpace(fmt.Sprint(input["kind"]))
	if id == "" || id == "<nil>" || kind == "" || !modelKinds[kind] {
		return nil, errors.New("模型 ID 和类型不能为空")
	}
	workflowID := strings.TrimSpace(fmt.Sprint(input["workflowId"]))
	if workflowID == "<nil>" {
		workflowID = strings.TrimSpace(fmt.Sprint(input["workflow_id"]))
	}
	if workflowID == "<nil>" {
		workflowID = ""
	}
	if id == "minimax-h3-video" && workflowID == "" {
		return nil, errors.New("H3 工作流 ID 不能为空")
	}
	credential := strings.TrimSpace(fmt.Sprint(input["credential"]))
	if credential == "<nil>" || credential == "" {
		credential = strings.TrimSpace(fmt.Sprint(input["apiKey"]))
	}
	if credential == "<nil>" {
		credential = ""
	}
	if id == "minimax-h3-video" && credential == "" {
		return nil, errors.New("保存 H3 配置前必须填写 API Key")
	}
	model := map[string]any{"id": id, "kind": kind, "displayName": strings.TrimSpace(fmt.Sprint(input["displayName"])), "modelId": strings.TrimSpace(fmt.Sprint(input["modelId"])), "baseUrl": strings.TrimSpace(fmt.Sprint(input["baseUrl"])), "endpoint": strings.TrimSpace(fmt.Sprint(input["endpoint"])), "requestTemplate": strings.TrimSpace(fmt.Sprint(input["requestTemplate"])), "responseMapping": strings.TrimSpace(fmt.Sprint(input["responseMapping"])), "credential": credential, "workflowId": workflowID, "enabled": input["enabled"] == true}
	if model["displayName"] == "<nil>" || model["displayName"] == "" {
		model["displayName"] = id
	}
	if model["modelId"] == "<nil>" {
		model["modelId"] = ""
	}
	if model["baseUrl"] == "<nil>" {
		model["baseUrl"] = ""
	}
	for _, key := range []string{"endpoint", "requestTemplate", "responseMapping"} {
		if model[key] == "<nil>" {
			model[key] = ""
		}
	}
	return model, nil
}

func publicManagedModel(model map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range model {
		if key != "credential" && key != "apiKey" {
			result[key] = value
		}
	}
	result["hasCredential"] = strings.TrimSpace(fmt.Sprint(model["credential"])) != "" && fmt.Sprint(model["credential"]) != "<nil>"
	return result
}
