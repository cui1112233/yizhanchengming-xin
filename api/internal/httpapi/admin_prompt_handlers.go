package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
)

func (h handler) requireAnyAdminCapability(next http.Handler) http.Handler {
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authn.CurrentUser(r.Context())
		if !ok || !authn.HasAnyAdminCapability(user.Capabilities) {
			writeAdminError(w, r, http.StatusForbidden, "ADMIN_CAPABILITY_REQUIRED", "缺少后台访问权限")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (h handler) requireAdminCapability(capability string, next http.Handler) http.Handler {
	return h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := authn.CurrentUser(r.Context())
		if !ok || !authn.HasCapability(user.Capabilities, capability) {
			writeAdminError(w, r, http.StatusForbidden, "ADMIN_CAPABILITY_REQUIRED", "缺少所需后台权限")
			return
		}
		next.ServeHTTP(w, r)
	}))
}

type adminPromptMetadata struct {
	ID            int64      `json:"id"`
	Key           string     `json:"key"`
	Version       int        `json:"version"`
	Lifecycle     string     `json:"lifecycle"`
	Enabled       bool       `json:"enabled"`
	SeedSource    string     `json:"seedSource"`
	ContentSHA256 string     `json:"contentSha256"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func toAdminPromptMetadata(prompt generation.AdminPrompt) adminPromptMetadata {
	return adminPromptMetadata{ID: prompt.ID, Key: prompt.Key, Version: prompt.Version, Lifecycle: prompt.Lifecycle, Enabled: prompt.Enabled, SeedSource: prompt.SeedSource, ContentSHA256: prompt.ContentSHA256, PublishedAt: prompt.PublishedAt, CreatedAt: prompt.CreatedAt, UpdatedAt: prompt.UpdatedAt}
}

func (h handler) adminCapabilities(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeAdminError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期")
		return
	}
	capabilities := make([]string, 0)
	for _, capability := range user.Capabilities {
		if strings.HasPrefix(capability, "admin.") {
			capabilities = append(capabilities, capability)
		}
	}
	sort.Strings(capabilities)
	writeJSON(w, http.StatusOK, map[string]any{"capabilities": capabilities})
}

func (h handler) adminPromptList(w http.ResponseWriter, r *http.Request) {
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	prompts, err := h.deps.AdminPrompts.ListAdminPrompts(r.Context(), r.URL.Query().Get("key"))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	metadata := make([]adminPromptMetadata, 0, len(prompts))
	for _, prompt := range prompts {
		metadata = append(metadata, toAdminPromptMetadata(prompt))
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompts": metadata})
}

func (h handler) adminPromptDetail(w http.ResponseWriter, r *http.Request) {
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	version, err := parsePositiveID(r.PathValue("version"))
	if err != nil {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "版本号无效")
		return
	}
	prompt, err := h.deps.AdminPrompts.GetAdminPrompt(r.Context(), r.PathValue("key"), int(version))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompt": map[string]any{
		"id": prompt.ID, "key": prompt.Key, "version": prompt.Version, "content": prompt.Content,
		"enabled": prompt.Enabled, "lifecycle": prompt.Lifecycle, "seedSource": prompt.SeedSource,
		"contentSha256": prompt.ContentSHA256, "publishedAt": prompt.PublishedAt,
		"createdAt": prompt.CreatedAt, "updatedAt": prompt.UpdatedAt,
	}})
}

func (h handler) adminPromptCreateDraft(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeAdminError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期")
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(w, r, &body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "提示词正文不能为空")
		return
	}
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	if err := h.deps.AdminPrompts.CreateAdminPromptDraft(r.Context(), user.ID, r.PathValue("key"), body.Content); err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	prompt, err := h.latestAdminPrompt(r, r.PathValue("key"))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"prompt": toAdminPromptMetadata(prompt)})
}

func (h handler) adminPromptUpdateDraft(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeAdminError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期")
		return
	}
	version, err := parsePositiveID(r.PathValue("version"))
	if err != nil {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "版本号无效")
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := decodeJSON(w, r, &body); err != nil || strings.TrimSpace(body.Content) == "" {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "提示词正文不能为空")
		return
	}
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	if err := h.deps.AdminPrompts.UpdateAdminPromptDraft(r.Context(), user.ID, r.PathValue("key"), int(version), body.Content); err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	prompt, err := h.deps.AdminPrompts.GetAdminPrompt(r.Context(), r.PathValue("key"), int(version))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompt": toAdminPromptMetadata(prompt)})
}

func (h handler) adminPromptPublish(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeAdminError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期")
		return
	}
	version, err := parsePositiveID(r.PathValue("version"))
	if err != nil {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "版本号无效")
		return
	}
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	if err := h.deps.AdminPrompts.PublishAdminPrompt(r.Context(), user.ID, r.PathValue("key"), int(version)); err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	prompt, err := h.deps.AdminPrompts.GetAdminPrompt(r.Context(), r.PathValue("key"), int(version))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompt": toAdminPromptMetadata(prompt)})
}

func (h handler) adminPromptRestore(w http.ResponseWriter, r *http.Request) {
	user, ok := authn.CurrentUser(r.Context())
	if !ok {
		writeAdminError(w, r, http.StatusUnauthorized, "AUTH_UNAUTHENTICATED", "登录状态无效或已过期")
		return
	}
	sourceVersion, err := parsePositiveID(r.PathValue("version"))
	if err != nil {
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "版本号无效")
		return
	}
	if h.deps.AdminPrompts == nil {
		writeAdminError(w, r, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "提示词管理服务暂不可用")
		return
	}
	if err := h.deps.AdminPrompts.RestoreAdminPrompt(r.Context(), user.ID, r.PathValue("key"), int(sourceVersion)); err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	prompt, err := h.latestAdminPrompt(r, r.PathValue("key"))
	if err != nil {
		writeAdminPromptOperationError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"sourceVersion": sourceVersion, "newVersion": prompt.Version, "prompt": toAdminPromptMetadata(prompt)})
}

func (h handler) latestAdminPrompt(r *http.Request, key string) (generation.AdminPrompt, error) {
	prompts, err := h.deps.AdminPrompts.ListAdminPrompts(r.Context(), key)
	if err != nil {
		return generation.AdminPrompt{}, err
	}
	if len(prompts) == 0 {
		return generation.AdminPrompt{}, generation.ErrNotFound
	}
	latest := prompts[0]
	for _, prompt := range prompts[1:] {
		if prompt.Version > latest.Version {
			latest = prompt
		}
	}
	return latest, nil
}

func writeAdminPromptOperationError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, generation.ErrInvalid):
		writeAdminError(w, r, http.StatusBadRequest, "ADMIN_INVALID_REQUEST", "请求无效")
	case errors.Is(err, generation.ErrNotFound):
		writeAdminError(w, r, http.StatusNotFound, "ADMIN_NOT_FOUND", "提示词版本不存在")
	case errors.Is(err, generation.ErrConflict):
		writeAdminError(w, r, http.StatusConflict, "ADMIN_PROMPT_STATE_CONFLICT", "提示词状态不允许此操作")
	default:
		writeAdminError(w, r, http.StatusInternalServerError, "ADMIN_OPERATION_FAILED", "后台操作失败")
	}
}

func writeAdminError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{"code": code, "message": message, "request_id": observability.RequestID(r.Context())})
}
