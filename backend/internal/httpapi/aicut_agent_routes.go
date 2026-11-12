package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/aicut"
)

type AICutAgentStarter func(context.Context, string) error

func RegisterAICutAgentRoutes(mux *http.ServeMux, auth BridgeAuth, timelines aicut.Store, agents aicut.AgentStore, start AICutAgentStarter, now func() time.Time) {
	if mux == nil || timelines == nil || agents == nil {
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

	handle("POST /api/ai-cut/projects/{id}/agent-runs", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Instruction      string                `json:"instruction"`
			Mode             aicut.AgentMode       `json:"mode"`
			ModelID          string                `json:"modelId"`
			ActiveSkillID    string                `json:"activeSkillId"`
			ExpectedRevision int64                 `json:"expectedRevision"`
			ClientRequestID  string                `json:"clientRequestId"`
			UIContext        *aicut.AgentUIContext `json:"uiContext,omitempty"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if strings.TrimSpace(input.Instruction) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请输入剪辑要求"})
			return
		}
		if input.Mode == "" {
			input.Mode = aicut.AgentModeAgent
		}
		if input.Mode != aicut.AgentModeAgent && input.Mode != aicut.AgentModeAsk {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Agent 模式无效"})
			return
		}
		projectID := req.PathValue("id")
		if _, err := timelines.Project(req.Context(), owner, projectID); err != nil {
			writeAICutError(w, err)
			return
		}
		current, err := timelines.Timeline(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if current.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		if input.UIContext != nil {
			if err := aicut.ValidateAgentUIContext(*input.UIContext, current); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "当前选择或播放位置无效，请重新选择"})
				return
			}
		}
		createdAt := now().UTC()
		generatedID := newAgentID("aicut-agent")
		clientRequestID := strings.TrimSpace(input.ClientRequestID)
		if clientRequestID == "" {
			clientRequestID = generatedID
		}
		run := aicut.AgentRun{ID: generatedID, ClientRequestID: clientRequestID, Owner: owner, ProjectID: projectID, Instruction: input.Instruction, Mode: input.Mode, ModelID: input.ModelID, ActiveSkillID: input.ActiveSkillID, ExpectedRevision: input.ExpectedRevision, State: aicut.AgentQueued, CreatedAt: createdAt, UpdatedAt: createdAt}
		stored, err := agents.CreateAgentRun(req.Context(), run)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if stored.ID == generatedID {
			if input.UIContext != nil {
				encoded, _ := json.Marshal(input.UIContext)
				_, err := agents.AppendAgentEvent(req.Context(), aicut.AgentEvent{RunID: stored.ID, Owner: owner, ProjectID: projectID, Type: aicut.AgentEventUIContext, Message: string(encoded), Revision: input.ExpectedRevision, CreatedAt: createdAt})
				if err != nil {
					_, _ = agents.UpdateAgentRunState(context.WithoutCancel(req.Context()), stored.ID, aicut.AgentFailed, "保存当前选择失败，请重试", 0, now().UTC())
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "保存当前选择失败，请重试"})
					return
				}
			}
			if start != nil {
				if err := start(req.Context(), stored.ID); err != nil {
					_, _ = agents.UpdateAgentRunState(req.Context(), stored.ID, aicut.AgentFailed, "Agent 队列不可用，请稍后重试", 0, now().UTC())
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Agent 队列不可用，请稍后重试"})
					return
				}
			}
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"run": stored})
	})

	handle("GET /api/ai-cut/projects/{id}/agent-runs/{runID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		run, err := agents.AgentRun(req.Context(), owner, req.PathValue("id"), req.PathValue("runID"))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run": run})
	})

	handle("GET /api/ai-cut/projects/{id}/agent-runs/{runID}/events", func(w http.ResponseWriter, req *http.Request, owner string) {
		after, err := strconv.ParseInt(req.URL.Query().Get("after"), 10, 64)
		if err != nil || after < 0 {
			after = 0
		}
		events, err := agents.AgentEvents(req.Context(), owner, req.PathValue("id"), req.PathValue("runID"), after)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events})
	})

	handle("POST /api/ai-cut/projects/{id}/agent-runs/{runID}/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		run, err := agents.RequestAgentCancel(req.Context(), owner, req.PathValue("id"), req.PathValue("runID"), now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run": run})
	})

	handle("GET /api/ai-cut/projects/{id}/agent-messages", func(w http.ResponseWriter, req *http.Request, owner string) {
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		messages, err := agents.AgentMessages(req.Context(), owner, req.PathValue("id"), limit)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
	})
}
