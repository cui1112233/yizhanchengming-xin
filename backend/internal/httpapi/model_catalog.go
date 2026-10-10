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

var modelKinds = map[string]bool{"text": true, "video": true, "image": true}

func registerModelCatalogRoutes(root *http.ServeMux, auth NovelFetchBridgeAuth, store novelfetchworkshop.Store) {
	mux := http.NewServeMux()
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
				if key != "id" && key != "credential" || key == "credential" {
					current[key] = value
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
	model := map[string]any{"id": id, "kind": kind, "displayName": strings.TrimSpace(fmt.Sprint(input["displayName"])), "modelId": strings.TrimSpace(fmt.Sprint(input["modelId"])), "baseUrl": strings.TrimSpace(fmt.Sprint(input["baseUrl"])), "credential": strings.TrimSpace(fmt.Sprint(input["credential"])), "enabled": input["enabled"] == true}
	if model["displayName"] == "<nil>" || model["displayName"] == "" {
		model["displayName"] = id
	}
	if model["modelId"] == "<nil>" {
		model["modelId"] = ""
	}
	if model["baseUrl"] == "<nil>" {
		model["baseUrl"] = ""
	}
	if model["credential"] == "<nil>" {
		model["credential"] = ""
	}
	return model, nil
}

func publicManagedModel(model map[string]any) map[string]any {
	result := map[string]any{}
	for key, value := range model {
		if key != "credential" {
			result[key] = value
		}
	}
	result["hasCredential"] = strings.TrimSpace(fmt.Sprint(model["credential"])) != "" && fmt.Sprint(model["credential"]) != "<nil>"
	return result
}
