package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
)

type AgentAPI interface {
	ListThreads(context.Context, string, int) ([]agent.Thread, error)
	CreateThread(context.Context, string, string) (agent.Thread, error)
	GetThread(context.Context, string, string) (agent.ThreadSnapshot, error)
	SendMessage(context.Context, string, string, agent.SendMessageInput) (agent.SendMessageResult, error)
	ListTasks(context.Context, string, string) ([]agent.Task, error)
	UpdateTask(context.Context, string, string, agent.UpdateTaskInput) (agent.Task, error)
	UpdateToolCall(context.Context, string, string, agent.UpdateToolCallInput) (agent.ToolCall, error)
}

type agentHandler struct {
	api   AgentAPI
	owner OwnerResolver
}

func NewAgentHandler(api AgentAPI, owner OwnerResolver) http.Handler {
	return &agentHandler{api: api, owner: owner}
}

func (h *agentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.api == nil || h.owner == nil {
		http.Error(w, "agent service unavailable", http.StatusServiceUnavailable)
		return
	}
	owner, err := h.owner(r)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	owner = strings.TrimSpace(owner)

	switch {
	case r.URL.Path == "/api/agent/threads":
		h.handleThreads(w, r, owner)
	case r.URL.Path == "/api/agent/tasks":
		h.handleTasks(w, r, owner)
	case strings.HasPrefix(r.URL.Path, "/api/agent/tasks/"):
		h.handleTaskResource(w, r, owner)
	case strings.HasPrefix(r.URL.Path, "/api/agent/tool-calls/"):
		h.handleToolCallResource(w, r, owner)
	case strings.HasPrefix(r.URL.Path, "/api/agent/threads/"):
		h.handleThreadResource(w, r, owner)
	default:
		http.NotFound(w, r)
	}
}

func (h *agentHandler) handleThreads(w http.ResponseWriter, r *http.Request, owner string) {
	switch r.Method {
	case http.MethodGet:
		limit := 50
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > 100 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = parsed
		}
		threads, err := h.api.ListThreads(r.Context(), owner, limit)
		if err != nil {
			http.Error(w, "failed to list agent threads", http.StatusInternalServerError)
			return
		}
		agentWriteJSON(w, http.StatusOK, map[string]any{"threads": threads})
	case http.MethodPost:
		var body struct {
			Title string `json:"title"`
		}
		if err := decodeAgentJSON(w, r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		thread, err := h.api.CreateThread(r.Context(), owner, body.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		agentWriteJSON(w, http.StatusCreated, map[string]any{"thread": thread})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *agentHandler) handleThreadResource(w http.ResponseWriter, r *http.Request, owner string) {
	rest := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/agent/threads/"), "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}
	threadID := strings.TrimSpace(parts[0])
	if len(parts) == 1 {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		snapshot, err := h.api.GetThread(r.Context(), owner, threadID)
		if err != nil {
			handleAgentError(w, err)
			return
		}
		agentWriteJSON(w, http.StatusOK, snapshot)
		return
	}
	if len(parts) == 2 && parts[1] == "messages" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var body agent.SendMessageInput
		if err := decodeAgentJSON(w, r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		result, err := h.api.SendMessage(r.Context(), owner, threadID, body)
		if err != nil {
			handleAgentError(w, err)
			return
		}
		agentWriteJSON(w, http.StatusCreated, result)
		return
	}
	http.NotFound(w, r)
}

func (h *agentHandler) handleTasks(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	tasks, err := h.api.ListTasks(r.Context(), owner, strings.TrimSpace(r.URL.Query().Get("thread_id")))
	if err != nil {
		http.Error(w, "failed to list agent tasks", http.StatusInternalServerError)
		return
	}
	agentWriteJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (h *agentHandler) handleTaskResource(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	taskID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/agent/tasks/"), "/")
	if taskID == "" || strings.Contains(taskID, "/") {
		http.NotFound(w, r)
		return
	}
	var body agent.UpdateTaskInput
	if err := decodeAgentJSON(w, r, &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	updated, err := h.api.UpdateTask(r.Context(), owner, taskID, body)
	if err != nil {
		handleAgentError(w, err)
		return
	}
	agentWriteJSON(w, http.StatusOK, map[string]any{"task": updated})
}

func (h *agentHandler) handleToolCallResource(w http.ResponseWriter, r *http.Request, owner string) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	toolID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/agent/tool-calls/"), "/")
	if toolID == "" || strings.Contains(toolID, "/") {
		http.NotFound(w, r)
		return
	}
	var body agent.UpdateToolCallInput
	if err := decodeAgentJSON(w, r, &body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	updated, err := h.api.UpdateToolCall(r.Context(), owner, toolID, body)
	if err != nil {
		handleAgentError(w, err)
		return
	}
	agentWriteJSON(w, http.StatusOK, map[string]any{"tool_call": updated})
}

func decodeAgentJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("multiple or trailing json values")
	}
	return nil
}

func handleAgentError(w http.ResponseWriter, err error) {
	if errors.Is(err, agent.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "required") || strings.Contains(message, "invalid") || strings.Contains(message, "exceeds") || strings.Contains(message, "too many") || strings.Contains(message, "cannot exceed") {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Error(w, "agent request failed", http.StatusInternalServerError)
}

func agentWriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
