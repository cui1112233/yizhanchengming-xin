package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/agentstudio"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (h handler) agent(w http.ResponseWriter, r *http.Request) (AgentStudioService, agentstudio.Actor, bool) {
	if h.deps.AgentStudio == nil {
		writeJSON(w, 503, map[string]any{"code": "AGENT_UNAVAILABLE", "message": "Agent 服务暂不可用"})
		return nil, agentstudio.Actor{}, false
	}
	u, ok := authn.CurrentUser(r.Context())
	if !ok || u.ID <= 0 {
		writeJSON(w, 401, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期"})
		return nil, agentstudio.Actor{}, false
	}
	return h.deps.AgentStudio, agentstudio.Actor{UserID: u.ID, TeamID: u.TeamID}, true
}
func (h handler) createAgentProject(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	var in agentstudio.CreateProjectInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in) != nil || strings.TrimSpace(in.Title) == "" {
		writeError(w, 400, "invalid agent project")
		return
	}
	p, e := s.CreateProject(r.Context(), a, in)
	if errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 403, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有关联项目访问权限"})
		return
	}
	if e != nil {
		writeError(w, 422, "cannot create agent project")
		return
	}
	writeJSON(w, 201, map[string]any{"project": p})
}
func (h handler) listAgentProjects(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	out, e := s.ListProjects(r.Context(), a)
	if e != nil {
		writeError(w, 503, "cannot list agent projects")
		return
	}
	writeJSON(w, 200, map[string]any{"projects": out})
}
func (h handler) deleteAgentProject(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	id, e := agentProjectID(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	e = s.DeleteProject(r.Context(), a, id)
	if errors.Is(e, agentstudio.ErrNotFound) || errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeError(w, 503, "cannot delete agent project")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h handler) listAgentMessages(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	id, e := agentProjectID(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	out, e := s.ListMessages(r.Context(), a, id)
	if errors.Is(e, agentstudio.ErrNotFound) || errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeError(w, 503, "cannot list messages")
		return
	}
	writeJSON(w, 200, map[string]any{"messages": out})
}
func (h handler) listAgentExecutions(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	id, e := agentProjectID(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	out, e := s.ListExecutions(r.Context(), a, id)
	if errors.Is(e, agentstudio.ErrNotFound) || errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeError(w, 503, "cannot list executions")
		return
	}
	writeJSON(w, 200, map[string]any{"executions": out})
}
func (h handler) listAgentSkills(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	out, e := s.ListSkills(r.Context(), a)
	if e != nil {
		writeError(w, 503, "cannot list skills")
		return
	}
	writeJSON(w, 200, map[string]any{"skills": out, "systemSkills": []map[string]string{{"key": "agent.system", "name": "创作系统技能"}}})
}
func (h handler) createAgentSkill(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	var in agentstudio.CreateSkillInput
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
		writeError(w, 400, "invalid skill")
		return
	}
	out, e := s.CreateSkill(r.Context(), a, in)
	if errors.Is(e, agentstudio.ErrInvalid) {
		writeError(w, 400, "invalid skill")
		return
	}
	if e != nil {
		writeError(w, 503, "cannot create skill")
		return
	}
	writeJSON(w, 201, map[string]any{"skill": out})
}
func (h handler) continueAgentProject(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	id, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || id < 1 {
		writeError(w, 400, "invalid projectId")
		return
	}
	var in agentstudio.ContinueInput
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in) != nil {
		writeError(w, 400, "invalid message")
		return
	}
	out, e := s.Continue(r.Context(), a, id, in)
	if errors.Is(e, agentstudio.ErrExecutorUnavailable) {
		writeJSON(w, 503, map[string]any{"code": "executor_unavailable", "message": "当前没有可用执行器", "result": out})
		return
	}
	if errors.Is(e, agentstudio.ErrNotFound) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeError(w, 422, "agent execution failed")
		return
	}
	writeJSON(w, 201, map[string]any{"result": out})
}

func agentProjectID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid projectId")
	}
	return id, nil
}

func writeCanvas(w http.ResponseWriter, status int, canvas agentstudio.Canvas) {
	writeJSON(w, status, map[string]any{"canvas": map[string]any{
		"projectId": canvas.ProjectID,
		"revision":  canvas.Revision,
		"document":  json.RawMessage(canvas.Document),
	}})
}

func writeCanvasError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, agentstudio.ErrForbidden):
		writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有该项目访问权限"})
	case errors.Is(err, agentstudio.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"code": "not_found"})
	case errors.Is(err, agentstudio.ErrConflict):
		writeJSON(w, http.StatusConflict, map[string]any{"code": "canvas_revision_conflict", "message": "画布已更新，请刷新后再保存"})
	case errors.Is(err, agentstudio.ErrInvalid):
		writeError(w, http.StatusBadRequest, "invalid canvas")
	default:
		return false
	}
	return true
}

func (h handler) getAgentCanvas(w http.ResponseWriter, r *http.Request) {
	s, actor, ok := h.agent(w, r)
	if !ok {
		return
	}
	projectID, err := agentProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	canvas, err := s.GetCanvas(r.Context(), actor, projectID)
	if writeCanvasError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "cannot read canvas")
		return
	}
	writeCanvas(w, http.StatusOK, canvas)
}

func (h handler) saveAgentCanvas(w http.ResponseWriter, r *http.Request) {
	s, actor, ok := h.agent(w, r)
	if !ok {
		return
	}
	projectID, err := agentProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var input struct {
		Revision int             `json:"revision"`
		Document json.RawMessage `json:"document"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil || !json.Valid(input.Document) {
		writeError(w, http.StatusBadRequest, "invalid canvas")
		return
	}
	canvas, err := s.SaveCanvas(r.Context(), actor, agentstudio.Canvas{ProjectID: projectID, Revision: input.Revision, Document: append([]byte(nil), input.Document...)})
	if writeCanvasError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "cannot save canvas")
		return
	}
	writeCanvas(w, http.StatusOK, canvas)
}

func (h handler) listAgentCanvasVersions(w http.ResponseWriter, r *http.Request) {
	s, actor, ok := h.agent(w, r)
	if !ok {
		return
	}
	projectID, err := agentProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	versions, err := s.ListCanvasVersions(r.Context(), actor, projectID)
	if writeCanvasError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "cannot read canvas history")
		return
	}
	items := make([]map[string]any, 0, len(versions))
	for _, version := range versions {
		items = append(items, map[string]any{"id": version.ID, "projectId": version.ProjectID, "revision": version.Revision, "document": json.RawMessage(version.Document), "createdAt": version.CreatedAt})
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": items})
}

func (h handler) restoreAgentCanvas(w http.ResponseWriter, r *http.Request) {
	s, actor, ok := h.agent(w, r)
	if !ok {
		return
	}
	projectID, err := agentProjectID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	revision, err := strconv.Atoi(r.PathValue("revision"))
	if err != nil || revision < 1 {
		writeError(w, http.StatusBadRequest, "invalid revision")
		return
	}
	canvas, err := s.RestoreCanvas(r.Context(), actor, projectID, revision)
	if writeCanvasError(w, err) {
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "cannot restore canvas")
		return
	}
	writeCanvas(w, http.StatusOK, canvas)
}

func agentIDs(r *http.Request) (int64, int64, error) {
	p, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || p < 1 {
		return 0, 0, errors.New("invalid projectId")
	}
	id, e := strconv.ParseInt(r.PathValue("attachmentId"), 10, 64)
	if e != nil || id < 1 {
		return 0, 0, errors.New("invalid attachmentId")
	}
	return p, id, nil
}
func (h handler) agentAttachments(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	p, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || p < 1 {
		writeError(w, 400, "invalid projectId")
		return
	}
	out, e := s.ListAttachments(r.Context(), a, p)
	if errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 403, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有该项目访问权限"})
		return
	}
	if errors.Is(e, agentstudio.ErrNotFound) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeError(w, 422, "cannot list attachments")
		return
	}
	writeJSON(w, 200, map[string]any{"attachments": out})
}
func (h handler) uploadAgentAttachment(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	p, e := strconv.ParseInt(r.PathValue("projectId"), 10, 64)
	if e != nil || p < 1 {
		writeError(w, 400, "invalid projectId")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20+1024)
	if e = r.ParseMultipartForm(8 << 20); e != nil {
		writeError(w, 400, "invalid multipart upload")
		return
	}
	f, head, e := r.FormFile("file")
	if e != nil {
		writeError(w, 400, "file is required")
		return
	}
	defer f.Close()
	ct := head.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	v, e := s.UploadAttachment(r.Context(), a, p, head.Filename, ct, f)
	if errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 403, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有该项目访问权限"})
		return
	}
	if errors.Is(e, agentstudio.ErrNotFound) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeJSON(w, 422, map[string]any{"code": "attachment_upload_failed", "message": "附件上传失败"})
		return
	}
	writeJSON(w, 201, map[string]any{"attachment": v})
}
func (h handler) readAgentAttachment(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	p, id, e := agentIDs(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	v, body, e := s.OpenAttachment(r.Context(), a, p, id)
	if errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 403, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有该项目访问权限"})
		return
	}
	if errors.Is(e, agentstudio.ErrNotFound) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeJSON(w, 503, map[string]any{"code": "storage_unavailable", "message": "附件暂不可读取"})
		return
	}
	defer body.Close()
	w.Header().Set("Content-Type", v.ContentType)
	filename := strings.NewReplacer("\"", "", "\r", "", "\n", "").Replace(v.Filename)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	_, _ = io.Copy(w, body)
}
func (h handler) deleteAgentAttachment(w http.ResponseWriter, r *http.Request) {
	s, a, ok := h.agent(w, r)
	if !ok {
		return
	}
	p, id, e := agentIDs(r)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	e = s.DeleteAttachment(r.Context(), a, p, id)
	if errors.Is(e, agentstudio.ErrForbidden) {
		writeJSON(w, 403, map[string]any{"code": "AUTH_FORBIDDEN", "message": "没有该项目访问权限"})
		return
	}
	if errors.Is(e, agentstudio.ErrNotFound) {
		writeJSON(w, 404, map[string]any{"code": "not_found"})
		return
	}
	if e != nil {
		writeJSON(w, 503, map[string]any{"code": "attachment_delete_failed", "message": "附件删除失败"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
