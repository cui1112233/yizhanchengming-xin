package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
)

func (h handler) createPublishingAccount(w http.ResponseWriter, r *http.Request) {
	if h.deps.Publishing == nil { writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"PUBLISH_UNAVAILABLE","message":"发布服务暂不可用"}); return }
	var input publishing.CreateAccountInput
	decoder := json.NewDecoder(r.Body); decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil { writeJSON(w,http.StatusBadRequest,map[string]any{"code":"PUBLISH_INVALID_REQUEST","message":"发布账号参数无效"}); return }
	user, _ := authn.CurrentUser(r.Context())
	account, err := h.deps.Publishing.CreateAccount(r.Context(), user, input); if err != nil { writePublishingError(w,err); return }
	writeJSON(w,http.StatusCreated,map[string]any{"account":account.Public()})
}

func (h handler) listPublishingAccounts(w http.ResponseWriter, r *http.Request) {
	user,_ := authn.CurrentUser(r.Context()); accounts,err := h.deps.Publishing.ListAccounts(r.Context(),user); if err != nil { writePublishingError(w,err); return }
	result := make([]publishing.PublicAccount,0,len(accounts)); for _,account := range accounts { result=append(result,account.Public()) }
	writeJSON(w,http.StatusOK,map[string]any{"accounts":result})
}

func (h handler) createPublishIntent(w http.ResponseWriter, r *http.Request) {
	var input publishing.CreateIntentInput
	decoder := json.NewDecoder(r.Body); decoder.DisallowUnknownFields(); if err := decoder.Decode(&input); err != nil { writeJSON(w,http.StatusBadRequest,map[string]any{"code":"PUBLISH_INVALID_REQUEST","message":"发布意图参数无效"}); return }
	user,_ := authn.CurrentUser(r.Context()); intent,err := h.deps.Publishing.CreateIntent(r.Context(),user,input); if err != nil { writePublishingError(w,err); return }
	writeJSON(w,http.StatusCreated,map[string]any{"intent":intent})
}

func (h handler) getPublishIntent(w http.ResponseWriter, r *http.Request) {
	id,err := strconv.ParseInt(r.PathValue("id"),10,64); if err != nil || id<=0 { writeJSON(w,http.StatusBadRequest,map[string]any{"code":"PUBLISH_INVALID_REQUEST","message":"发布意图 ID 无效"}); return }
	user,_ := authn.CurrentUser(r.Context()); intent,err := h.deps.Publishing.GetIntent(r.Context(),user,id); if err != nil { writePublishingError(w,err); return }; writeJSON(w,http.StatusOK,map[string]any{"intent":intent})
}

func (h handler) listPublishAudits(w http.ResponseWriter, r *http.Request) {
	var projectID int64
	if raw := strings.TrimSpace(r.URL.Query().Get("batchProjectId")); raw != "" { value,err := strconv.ParseInt(raw,10,64); if err != nil || value<=0 { writeJSON(w,http.StatusBadRequest,map[string]any{"code":"PUBLISH_INVALID_REQUEST","message":"batchProjectId 无效"}); return }; projectID=value }
	user,_ := authn.CurrentUser(r.Context()); audits,err := h.deps.Publishing.ListAudits(r.Context(),user,projectID); if err != nil { writePublishingError(w,err); return }; writeJSON(w,http.StatusOK,map[string]any{"audits":audits})
}

func writePublishingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err,publishing.ErrInvalid): writeJSON(w,http.StatusBadRequest,map[string]any{"code":"PUBLISH_INVALID_REQUEST","message":"发布参数无效"})
	case errors.Is(err,publishing.ErrForbidden): writeJSON(w,http.StatusForbidden,map[string]any{"code":"AUTH_FORBIDDEN","message":"你没有执行此发布操作的权限"})
	case errors.Is(err,publishing.ErrNotFound): writeJSON(w,http.StatusNotFound,map[string]any{"code":"PUBLISH_NOT_FOUND","message":"发布资源不存在"})
	case errors.Is(err,publishing.ErrUnavailable): writeJSON(w,http.StatusServiceUnavailable,map[string]any{"code":"PUBLISH_UNAVAILABLE","message":"发布服务暂不可用"})
	default: writeJSON(w,http.StatusInternalServerError,map[string]any{"code":"PUBLISH_FAILED","message":"发布操作失败"})
	}
}
