package httpapi

import (
	"bytes"
	"context"
	"errors"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/agentstudio"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type agentHTTPFake struct {
	listed    bool
	actor     agentstudio.Actor
	uploaded  []byte
	uploads   int
	openErr   error
	deleteErr error
	canvas    agentstudio.Canvas
}

func (*agentHTTPFake) ListProjects(context.Context, agentstudio.Actor) ([]agentstudio.Project, error) {
	return nil, nil
}
func (*agentHTTPFake) DeleteProject(context.Context, agentstudio.Actor, int64) error { return nil }
func (*agentHTTPFake) ListMessages(context.Context, agentstudio.Actor, int64) ([]agentstudio.Message, error) {
	return nil, nil
}
func (*agentHTTPFake) ListExecutions(context.Context, agentstudio.Actor, int64) ([]agentstudio.Execution, error) {
	return nil, nil
}
func (*agentHTTPFake) CreateSkill(context.Context, agentstudio.Actor, agentstudio.CreateSkillInput) (agentstudio.Skill, error) {
	return agentstudio.Skill{}, nil
}
func (*agentHTTPFake) ListSkills(context.Context, agentstudio.Actor) ([]agentstudio.Skill, error) {
	return nil, nil
}

func (*agentHTTPFake) CreateProject(context.Context, agentstudio.Actor, agentstudio.CreateProjectInput) (agentstudio.Project, error) {
	return agentstudio.Project{}, nil
}
func (*agentHTTPFake) Continue(context.Context, agentstudio.Actor, int64, agentstudio.ContinueInput) (agentstudio.ContinueResult, error) {
	return agentstudio.ContinueResult{}, nil
}
func (f *agentHTTPFake) ListAttachments(_ context.Context, a agentstudio.Actor, p int64) ([]agentstudio.Attachment, error) {
	f.actor = a
	f.listed = true
	return []agentstudio.Attachment{{ID: 3, ProjectID: p, Filename: "x.txt"}}, nil
}
func (f *agentHTTPFake) UploadAttachment(_ context.Context, a agentstudio.Actor, p int64, _ string, _ string, r io.Reader) (agentstudio.Attachment, error) {
	f.actor = a
	f.uploads++
	f.uploaded, _ = io.ReadAll(r)
	return agentstudio.Attachment{ID: 4, ProjectID: p, Filename: "draft.txt"}, nil
}
func (f *agentHTTPFake) OpenAttachment(_ context.Context, a agentstudio.Actor, p, id int64) (agentstudio.Attachment, io.ReadCloser, error) {
	f.actor = a
	if f.openErr != nil {
		return agentstudio.Attachment{}, nil, f.openErr
	}
	return agentstudio.Attachment{ID: id, ProjectID: p, Filename: "draft.txt", ContentType: "text/plain"}, io.NopCloser(strings.NewReader("body")), nil
}
func (f *agentHTTPFake) DeleteAttachment(_ context.Context, a agentstudio.Actor, _, _ int64) error {
	f.actor = a
	return f.deleteErr
}
func (f *agentHTTPFake) GetCanvas(_ context.Context, a agentstudio.Actor, p int64) (agentstudio.Canvas, error) {
	f.actor = a
	if f.canvas.ProjectID == 0 {
		return agentstudio.Canvas{ProjectID: p, Document: []byte(`{"nodes":[],"edges":[]}`)}, nil
	}
	return f.canvas, nil
}
func (f *agentHTTPFake) SaveCanvas(_ context.Context, a agentstudio.Actor, c agentstudio.Canvas) (agentstudio.Canvas, error) {
	f.actor = a
	c.Revision++
	f.canvas = c
	return c, nil
}
func (f *agentHTTPFake) ListCanvasVersions(_ context.Context, a agentstudio.Actor, p int64) ([]agentstudio.CanvasVersion, error) {
	f.actor = a
	return []agentstudio.CanvasVersion{{ID: 1, ProjectID: p, Revision: 1, Document: []byte(`{"nodes":[]}`)}}, nil
}
func (f *agentHTTPFake) RestoreCanvas(_ context.Context, a agentstudio.Actor, p int64, revision int) (agentstudio.Canvas, error) {
	f.actor = a
	return agentstudio.Canvas{ProjectID: p, Revision: revision + 1, Document: []byte(`{"nodes":["restored"]}`)}, nil
}

func TestAgentAttachmentUploadRequiresCookieSessionCSRFAndCapability(t *testing.T) {
	f := &agentHTTPFake{}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityAgentCreate}}}
	h := NewHandler(Dependencies{AgentStudio: f, Auth: auth})

	newRequest := func() *http.Request {
		var body bytes.Buffer
		form := multipart.NewWriter(&body)
		part, err := form.CreateFormFile("file", "draft.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write([]byte("source")); err != nil {
			t.Fatal(err)
		}
		if err = form.Close(); err != nil {
			t.Fatal(err)
		}
		r := httptest.NewRequest(http.MethodPost, "/api/v1/agent/projects/9/attachments", &body)
		r.Header.Set("Content-Type", form.FormDataContentType())
		r.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
		return r
	}

	csrfRejected := httptest.NewRecorder()
	h.ServeHTTP(csrfRejected, newRequest())
	if csrfRejected.Code != http.StatusForbidden || f.uploads != 0 {
		t.Fatalf("status=%d uploads=%d", csrfRejected.Code, f.uploads)
	}

	req := newRequest()
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated || f.uploads != 1 || string(f.uploaded) != "source" || f.actor.UserID != 7 {
		t.Fatalf("status=%d uploads=%d body=%q actor=%+v", rec.Code, f.uploads, f.uploaded, f.actor)
	}
}

func TestAgentAttachmentReadAndDeleteRespectForbiddenServiceResult(t *testing.T) {
	f := &agentHTTPFake{openErr: agentstudio.ErrForbidden, deleteErr: agentstudio.ErrForbidden}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityAgentView, CapabilityAgentCreate}}}
	h := NewHandler(Dependencies{AgentStudio: f, Auth: auth})

	read := httptest.NewRequest(http.MethodGet, "/api/v1/agent/projects/9/attachments/3/content", nil)
	read.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	readRec := httptest.NewRecorder()
	h.ServeHTTP(readRec, read)
	if readRec.Code != http.StatusForbidden {
		t.Fatalf("read status=%d body=%s", readRec.Code, readRec.Body.String())
	}

	remove := httptest.NewRequest(http.MethodDelete, "/api/v1/agent/projects/9/attachments/3", nil)
	remove.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	sameOrigin(remove)
	removeRec := httptest.NewRecorder()
	h.ServeHTTP(removeRec, remove)
	if removeRec.Code != http.StatusForbidden {
		t.Fatalf("delete status=%d body=%s", removeRec.Code, removeRec.Body.String())
	}

	f.openErr = errors.New("tos down")
	read = httptest.NewRequest(http.MethodGet, "/api/v1/agent/projects/9/attachments/3/content", nil)
	read.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	readRec = httptest.NewRecorder()
	h.ServeHTTP(readRec, read)
	if readRec.Code != http.StatusServiceUnavailable {
		t.Fatalf("storage status=%d", readRec.Code)
	}
}

func TestAgentCanvasRoutesUseCookieSessionCapabilityAndRevision(t *testing.T) {
	f := &agentHTTPFake{}
	auth := &fakeAuthService{user: authn.User{ID: 7, Capabilities: []string{CapabilityAgentView, CapabilityAgentCreate}}}
	h := NewHandler(Dependencies{AgentStudio: f, Auth: auth})

	read := httptest.NewRequest(http.MethodGet, "/api/v1/agent/projects/9/canvas", nil)
	read.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	readRec := httptest.NewRecorder()
	h.ServeHTTP(readRec, read)
	if readRec.Code != http.StatusOK || !strings.Contains(readRec.Body.String(), `"projectId":9`) {
		t.Fatalf("read status=%d body=%s", readRec.Code, readRec.Body.String())
	}

	save := httptest.NewRequest(http.MethodPut, "/api/v1/agent/projects/9/canvas", strings.NewReader(`{"revision":0,"document":{"nodes":["first"],"edges":[]}}`))
	save.Header.Set("Content-Type", "application/json")
	save.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	sameOrigin(save)
	saveRec := httptest.NewRecorder()
	h.ServeHTTP(saveRec, save)
	if saveRec.Code != http.StatusOK || f.canvas.Revision != 1 || string(f.canvas.Document) != `{"nodes":["first"],"edges":[]}` {
		t.Fatalf("save status=%d canvas=%+v", saveRec.Code, f.canvas)
	}

	restore := httptest.NewRequest(http.MethodPost, "/api/v1/agent/projects/9/canvas/versions/1/restore", nil)
	restore.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "session"})
	sameOrigin(restore)
	restoreRec := httptest.NewRecorder()
	h.ServeHTTP(restoreRec, restore)
	if restoreRec.Code != http.StatusOK || !strings.Contains(restoreRec.Body.String(), "restored") {
		t.Fatalf("restore status=%d body=%s", restoreRec.Code, restoreRec.Body.String())
	}
}
func TestAgentAttachmentListUsesAuthenticatedActor(t *testing.T) {
	f := &agentHTTPFake{}
	h := NewHandler(Dependencies{AgentStudio: f})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/agent/projects/9/attachments", nil).WithContext(authn.WithCurrentUser(context.Background(), authn.User{ID: 7, TeamID: 2}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || !f.listed || f.actor.UserID != 7 {
		t.Fatalf("code=%d fake=%+v", w.Code, f)
	}
}
