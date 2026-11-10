package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
)

type createEditPlanInput struct {
	ClientRequestID    string   `json:"clientRequestId"`
	ModelID            string   `json:"modelId"`
	ExpectedRevision   int64    `json:"expectedRevision"`
	AssetIDs           []string `json:"assetIds"`
	Goal               string   `json:"goal"`
	TargetDurationMS   int64    `json:"targetDurationMs"`
	AspectRatio        string   `json:"aspectRatio"`
	IncludeCaptions    bool     `json:"includeCaptions"`
	IncludeTransitions bool     `json:"includeTransitions"`
	IncludeFilters     bool     `json:"includeFilters"`
	IncludeBGM         bool     `json:"includeBgm"`
	ExtraRequirement   string   `json:"extraRequirement"`
}

var (
	editPlanTransitions = []string{"fade", "flash-black", "flash-white", "slide-left", "slide-right", "zoom", "blur"}
	editPlanFilters     = []string{"filter-original", "filter-bright", "filter-warm-sun", "filter-cool-blue", "filter-cinema", "filter-monochrome", "filter-soft"}
	editPlanTextStyles  = []string{"text-basic-caption", "text-bold-title", "text-outline-caption", "text-shadow-caption", "text-panel-caption", "text-gradient-title"}
)

func writeEditPlanError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, aicut.ErrEditPlanNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "剪辑草案不存在"})
	case errors.Is(err, aicut.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "时间线已更新，草案已过期，请重新生成"})
	case errors.Is(err, aicut.ErrInvalidEditPlan), errors.Is(err, aicut.ErrEditPlanApply):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "剪辑草案内容无效"})
	default:
		writeAICutError(w, err)
	}
}

func RegisterAICutEditPlanRoutes(mux *http.ServeMux, auth BridgeAuth, timelines aicut.Store, plans aicut.EditPlanStore, resolver AICutTextProviderResolver, now func() time.Time) {
	if mux == nil || timelines == nil || plans == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("POST /api/ai-cut/projects/{id}/edit-plans", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input createEditPlanInput
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		projectID := req.PathValue("id")
		if existing, err := plans.EditPlanByRequest(req.Context(), owner, projectID, strings.TrimSpace(input.ClientRequestID)); err == nil {
			writeJSON(w, http.StatusOK, map[string]any{"plan": existing})
			return
		} else if !errors.Is(err, aicut.ErrEditPlanNotFound) {
			writeEditPlanError(w, err)
			return
		}
		if len(input.AssetIDs) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请至少选择一个素材"})
			return
		}
		if strings.TrimSpace(input.ModelID) == "" || resolver == nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		timeline, err := timelines.Timeline(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if timeline.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		request := aicut.EditPlanRequest{AssetIDs: input.AssetIDs, Goal: input.Goal, TargetDurationMS: input.TargetDurationMS, AspectRatio: input.AspectRatio, IncludeCaptions: input.IncludeCaptions, IncludeTransitions: input.IncludeTransitions, IncludeFilters: input.IncludeFilters, IncludeBGM: input.IncludeBGM, ExtraRequirement: input.ExtraRequirement}
		contract, err := prompt.AICutEditPlanContract(timeline, request, editPlanTransitions, editPlanFilters, editPlanTextStyles)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		createdAt := now().UTC()
		plan := aicut.EditPlan{ID: newAgentID("edit-plan"), ClientRequestID: strings.TrimSpace(input.ClientRequestID), Owner: owner, ProjectID: projectID, BaseRevision: timeline.Revision, ModelID: strings.TrimSpace(input.ModelID), State: aicut.EditPlanPlanning, Request: request, Operations: []aicut.EditOperation{}, CreatedAt: createdAt, UpdatedAt: createdAt}
		plan, err = plans.CreateEditPlan(req.Context(), plan)
		if err != nil {
			writeEditPlanError(w, err)
			return
		}
		provider, err := resolver(req.Context(), owner, plan.ModelID)
		if err != nil || provider == nil {
			_, _ = plans.UpdateEditPlanState(req.Context(), owner, projectID, plan.ID, aicut.EditPlanFailed, "文本模型未配置", now().UTC())
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "文本模型未配置，请先到 API 配置启用文本模型"})
			return
		}
		result, err := provider.Complete(req.Context(), batchfactoryv11.TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			_, _ = plans.UpdateEditPlanState(context.Background(), owner, projectID, plan.ID, aicut.EditPlanCanceled, "剪辑草案生成已取消", now().UTC())
			writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "剪辑草案生成已取消"})
			return
		}
		if err != nil {
			_, _ = plans.UpdateEditPlanState(req.Context(), owner, projectID, plan.ID, aicut.EditPlanFailed, err.Error(), now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "剪辑草案生成失败：" + err.Error()})
			return
		}
		ready, err := aicut.ParseEditPlanModelOutput(result, aicut.EditPlanEnvelope{ID: plan.ID, ClientRequestID: plan.ClientRequestID, Owner: owner, ProjectID: projectID, BaseRevision: plan.BaseRevision, ModelID: plan.ModelID, Request: plan.Request, CreatedAt: plan.CreatedAt, UpdatedAt: now().UTC()})
		if err != nil {
			_, _ = plans.UpdateEditPlanState(req.Context(), owner, projectID, plan.ID, aicut.EditPlanFailed, "模型返回的剪辑草案格式无效", now().UTC())
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "模型返回的剪辑草案格式无效"})
			return
		}
		ready, err = plans.UpdateEditPlan(req.Context(), ready)
		if err != nil {
			writeEditPlanError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"plan": ready})
	})

	handle("GET /api/ai-cut/projects/{id}/edit-plans/{planId}", func(w http.ResponseWriter, req *http.Request, owner string) {
		plan, err := plans.EditPlan(req.Context(), owner, req.PathValue("id"), req.PathValue("planId"))
		if err != nil {
			writeEditPlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
	})

	handle("POST /api/ai-cut/projects/{id}/edit-plans/{planId}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		plan, err := plans.UpdateEditPlanState(req.Context(), owner, req.PathValue("id"), req.PathValue("planId"), aicut.EditPlanCanceled, "", now().UTC())
		if err != nil {
			writeEditPlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"plan": plan})
	})

	handle("POST /api/ai-cut/projects/{id}/edit-plans/{planId}/apply", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			EnabledOperationIDs []string `json:"enabledOperationIds"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		plan, timeline, summary, err := plans.ApplyStoredEditPlan(req.Context(), owner, req.PathValue("id"), req.PathValue("planId"), input.EnabledOperationIDs, now().UTC())
		if err != nil {
			writeEditPlanError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"plan": plan, "timeline": timeline, "summary": summary})
	})
}
