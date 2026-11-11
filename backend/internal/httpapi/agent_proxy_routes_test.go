package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentproxy"
)

func TestAgentProxyWorkspaceSessionMessageLifecycle(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	store := agentproxy.NewMemoryStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentProxyRoutes(mux, auth, store, func() time.Time { return now })

	unsigned := httptest.NewRecorder()
	mux.ServeHTTP(unsigned, httptest.NewRequest(http.MethodGet, "/api/agent/proxy/workspaces", nil))
	if unsigned.Code != http.StatusUnauthorized {
		t.Fatalf("unsigned request status=%d body=%s", unsigned.Code, unsigned.Body.String())
	}

	created := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/workspaces", `{"name":"内容工作区","description":"骨架验证","permissionMode":"auto"}`)
	if created.Code != http.StatusOK {
		t.Fatalf("create workspace status=%d body=%s", created.Code, created.Body.String())
	}
	var workspaceBody struct {
		Workspace agentproxy.Workspace `json:"workspace"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &workspaceBody); err != nil {
		t.Fatal(err)
	}
	workspace := workspaceBody.Workspace
	if workspace.ID == "" || workspace.PermissionMode != agentproxy.PermissionAuto {
		t.Fatalf("workspace=%+v", workspace)
	}

	// bypass 第一版不开放。
	bypass := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/workspaces", `{"name":"越权","permissionMode":"bypass"}`)
	if bypass.Code != http.StatusBadRequest || !strings.Contains(bypass.Body.String(), "bypass") {
		t.Fatalf("bypass status=%d body=%s", bypass.Code, bypass.Body.String())
	}

	list := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/workspaces", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), workspace.ID) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	sessionCreated := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/workspaces/"+workspace.ID+"/sessions", `{"title":"第一轮对话"}`)
	if sessionCreated.Code != http.StatusOK {
		t.Fatalf("create session status=%d body=%s", sessionCreated.Code, sessionCreated.Body.String())
	}
	var sessionBody struct {
		Session agentproxy.Session `json:"session"`
	}
	if err := json.Unmarshal(sessionCreated.Body.Bytes(), &sessionBody); err != nil {
		t.Fatal(err)
	}
	session := sessionBody.Session
	if session.ID == "" || session.Status != agentproxy.SessionStatusActive || session.PermissionMode != agentproxy.PermissionAuto {
		t.Fatalf("session=%+v", session)
	}

	// auto 档位可以直接追加消息。
	appended := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/sessions/"+session.ID+"/messages", `{"role":"user","content":"列出工作区","metadata":{"source":"test"}}`)
	if appended.Code != http.StatusOK {
		t.Fatalf("append status=%d body=%s", appended.Code, appended.Body.String())
	}
	messages := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/sessions/"+session.ID+"/messages", "")
	if messages.Code != http.StatusOK || !strings.Contains(messages.Body.String(), "列出工作区") {
		t.Fatalf("messages status=%d body=%s", messages.Code, messages.Body.String())
	}

	skills := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/workspaces/"+workspace.ID+"/skills", "")
	if skills.Code != http.StatusOK || !strings.Contains(skills.Body.String(), `"skillsEnabled":false`) {
		t.Fatalf("skills status=%d body=%s", skills.Code, skills.Body.String())
	}
	jobs := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/workspaces/"+workspace.ID+"/scheduled-jobs", "")
	if jobs.Code != http.StatusOK || !strings.Contains(jobs.Body.String(), `"schedulingEnabled":false`) {
		t.Fatalf("jobs status=%d body=%s", jobs.Code, jobs.Body.String())
	}

	updated := agentRequest(t, mux, now, http.MethodPatch, "/api/agent/proxy/workspaces/"+workspace.ID, `{"name":"改名后的工作区","permissionMode":"plan"}`)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), "改名后的工作区") || !strings.Contains(updated.Body.String(), `"permissionMode":"plan"`) {
		t.Fatalf("update workspace status=%d body=%s", updated.Code, updated.Body.String())
	}

	if deleted := agentRequest(t, mux, now, http.MethodDelete, "/api/agent/proxy/sessions/"+session.ID, ""); deleted.Code != http.StatusOK {
		t.Fatalf("delete session status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	if missing := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/sessions/"+session.ID, ""); missing.Code != http.StatusNotFound {
		t.Fatalf("deleted session status=%d body=%s", missing.Code, missing.Body.String())
	}
	if deleted := agentRequest(t, mux, now, http.MethodDelete, "/api/agent/proxy/workspaces/"+workspace.ID, ""); deleted.Code != http.StatusOK {
		t.Fatalf("delete workspace status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	if missing := agentRequest(t, mux, now, http.MethodGet, "/api/agent/proxy/workspaces/"+workspace.ID, ""); missing.Code != http.StatusNotFound {
		t.Fatalf("deleted workspace status=%d body=%s", missing.Code, missing.Body.String())
	}
}

func TestAgentProxyPlanModeRejectsExecution(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 30, 0, 0, time.UTC)
	store := agentproxy.NewMemoryStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentProxyRoutes(mux, auth, store, func() time.Time { return now })

	workspace := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/workspaces", `{"name":"只读工作区","permissionMode":"plan"}`)
	var workspaceBody struct {
		Workspace agentproxy.Workspace `json:"workspace"`
	}
	if err := json.Unmarshal(workspace.Body.Bytes(), &workspaceBody); err != nil {
		t.Fatal(err)
	}
	sessionCreated := agentRequest(t, mux, now, http.MethodPost, "/api/agent/proxy/workspaces/"+workspaceBody.Workspace.ID+"/sessions", `{"title":"只读会话"}`)
	var sessionBody struct {
		Session agentproxy.Session `json:"session"`
	}
	if err := json.Unmarshal(sessionCreated.Body.Bytes(), &sessionBody); err != nil {
		t.Fatal(err)
	}
	path := "/api/agent/proxy/sessions/" + sessionBody.Session.ID + "/messages"

	if denied := agentRequest(t, mux, now, http.MethodPost, path, `{"role":"user","content":"执行一下","execute":true}`); denied.Code != http.StatusForbidden {
		t.Fatalf("plan execute status=%d body=%s", denied.Code, denied.Body.String())
	}
	if denied := agentRequest(t, mux, now, http.MethodPost, path, `{"role":"tool","content":"已执行"}`); denied.Code != http.StatusForbidden {
		t.Fatalf("plan tool status=%d body=%s", denied.Code, denied.Body.String())
	}
	if allowed := agentRequest(t, mux, now, http.MethodPost, path, `{"role":"assistant","content":"建议先确认场景"}`); allowed.Code != http.StatusOK {
		t.Fatalf("plan suggestion status=%d body=%s", allowed.Code, allowed.Body.String())
	}

	// 切到 auto 后允许执行类消息。
	if switched := agentRequest(t, mux, now, http.MethodPatch, "/api/agent/proxy/sessions/"+sessionBody.Session.ID, `{"permissionMode":"auto"}`); switched.Code != http.StatusOK {
		t.Fatalf("switch mode status=%d body=%s", switched.Code, switched.Body.String())
	}
	if allowed := agentRequest(t, mux, now, http.MethodPost, path, `{"role":"tool","content":"已执行","execute":true}`); allowed.Code != http.StatusOK {
		t.Fatalf("auto execute status=%d body=%s", allowed.Code, allowed.Body.String())
	}
	if invalid := agentRequest(t, mux, now, http.MethodPatch, "/api/agent/proxy/sessions/"+sessionBody.Session.ID, `{"status":"running"}`); invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid status code=%d body=%s", invalid.Code, invalid.Body.String())
	}
}
