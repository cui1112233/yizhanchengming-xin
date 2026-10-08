package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
)

type intakeAccessStub struct {
	allowed        bool
	err            error
	accessCalls    int
	listCalls      int
	lastUserID     int64
	lastTeamID     int64
	lastElevated   bool
	visibleIntakes []intake.Intake
}

func (s *intakeAccessStub) CanAccessIntake(_ context.Context, _ int64, userID, teamID int64, elevated bool) (bool, error) {
	s.accessCalls++
	s.lastUserID, s.lastTeamID, s.lastElevated = userID, teamID, elevated
	return s.allowed, s.err
}

func (s *intakeAccessStub) ListVisibleIntakes(_ context.Context, userID, teamID int64, elevated bool) ([]intake.Intake, error) {
	s.listCalls++
	s.lastUserID, s.lastTeamID, s.lastElevated = userID, teamID, elevated
	return s.visibleIntakes, s.err
}

type intakeBusinessSpy struct {
	calls int
}

func (s *intakeBusinessSpy) CreateIntake(_ context.Context, _ intake.CreateIntakeInput) (intake.Intake, []intake.Book, error) {
	s.calls++
	return intake.Intake{ID: 11}, nil, nil
}
func (s *intakeBusinessSpy) ExecuteIntake(_ context.Context, id int64, _ int) (intake.ExecuteResult, error) {
	s.calls++
	return intake.ExecuteResult{IntakeID: id, Status: intake.StatusCompleted}, nil
}
func (s *intakeBusinessSpy) RestoreBook(_ context.Context, intakeID, bookID int64, _ int) (intake.Book, error) {
	s.calls++
	return intake.Book{ID: bookID, IntakeID: intakeID}, nil
}
func (s *intakeBusinessSpy) ListIntakes(context.Context) ([]intake.Intake, error) {
	s.calls++
	return []intake.Intake{{ID: 999, Name: "不应读取全局列表"}}, nil
}
func (s *intakeBusinessSpy) ListBooks(_ context.Context, id int64) ([]intake.Book, error) {
	s.calls++
	return []intake.Book{{ID: 21, IntakeID: id, OriginalText: "foreign-secret-original"}}, nil
}
func (s *intakeBusinessSpy) Snapshot(_ context.Context, id int64) (workshop.Snapshot, error) {
	s.calls++
	return workshop.Snapshot{Intake: intake.Intake{ID: id}, Books: []intake.Book{{ID: 21, OriginalText: "foreign-secret-original"}}}, nil
}
func (s *intakeBusinessSpy) Save(_ context.Context, _ int64, settings json.RawMessage) (json.RawMessage, error) {
	s.calls++
	return settings, nil
}
func (s *intakeBusinessSpy) Create(_ context.Context, request pipeline.CreateRequest) (pipeline.CreateResult, error) {
	s.calls++
	return pipeline.CreateResult{Project: intake.BatchProject{ID: 51, IntakeID: request.IntakeID}, Run: intake.Run{ID: 71}}, nil
}

func authenticatedIntakeRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "access"})
	if method != http.MethodGet {
		sameOrigin(req)
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestIntakeObjectRoutesFailClosedBeforeBusinessServices(t *testing.T) {
	routes := []struct {
		name   string
		method string
		path   string
		body   string
		cap    string
	}{
		{name: "execute", method: http.MethodPost, path: "/api/v1/intakes/11/execute", body: `{"maxText":4000}`, cap: CapabilityBatchExecute},
		{name: "books", method: http.MethodGet, path: "/api/v1/intakes/11/books", cap: CapabilityBatchView},
		{name: "workshop read", method: http.MethodGet, path: "/api/v1/intakes/11/workshop", cap: CapabilityBatchView},
		{name: "workshop save", method: http.MethodPut, path: "/api/v1/intakes/11/workshop", body: `{"settings":{}}`, cap: CapabilityBatchConfigure},
		{name: "restore", method: http.MethodPost, path: "/api/v1/intakes/11/books/21/restore", body: `{"maxText":4000}`, cap: CapabilityBatchExecute},
		{name: "create project", method: http.MethodPost, path: "/api/v1/intakes/11/batch-projects", body: `{}`, cap: CapabilityBatchExecute},
	}

	for _, route := range routes {
		t.Run(route.name+" foreign", func(t *testing.T) {
			business := &intakeBusinessSpy{}
			access := &intakeAccessStub{allowed: false}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.cap}}}
			h := NewHandler(Dependencies{Auth: auth, Intakes: business, Reader: business, Workshop: business, Pipeline: business, IntakeAccess: access})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, authenticatedIntakeRequest(route.method, route.path, route.body))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if business.calls != 0 {
				t.Fatalf("business calls = %d, want zero", business.calls)
			}
			if bytes.Contains(rec.Body.Bytes(), []byte("foreign-secret-original")) {
				t.Fatalf("foreign workshop original leaked: %s", rec.Body.String())
			}
		})

		t.Run(route.name+" missing checker", func(t *testing.T) {
			business := &intakeBusinessSpy{}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{route.cap}}}
			h := NewHandler(Dependencies{Auth: auth, Intakes: business, Reader: business, Workshop: business, Pipeline: business})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, authenticatedIntakeRequest(route.method, route.path, route.body))
			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if business.calls != 0 {
				t.Fatalf("business calls = %d, want zero", business.calls)
			}
		})
	}
}

func TestIntakeAccessElevatesOnlyAdminAndOwnerRoles(t *testing.T) {
	for _, tc := range []struct {
		role     string
		elevated bool
	}{
		{role: "owner", elevated: true},
		{role: "admin", elevated: true},
		{role: "dev", elevated: false},
		{role: "manager", elevated: false},
		{role: "member", elevated: false},
	} {
		t.Run(tc.role, func(t *testing.T) {
			business := &intakeBusinessSpy{}
			access := &intakeAccessStub{allowed: true}
			auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: tc.role, Capabilities: []string{CapabilityBatchView}}}
			h := NewHandler(Dependencies{Auth: auth, Reader: business, IntakeAccess: access})
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, authenticatedIntakeRequest(http.MethodGet, "/api/v1/intakes/11/books", ""))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if access.lastElevated != tc.elevated {
				t.Fatalf("elevated = %v, want %v", access.lastElevated, tc.elevated)
			}
		})
	}
}

func TestListIntakesUsesScopedSQLReaderInsteadOfGlobalReader(t *testing.T) {
	business := &intakeBusinessSpy{}
	access := &intakeAccessStub{visibleIntakes: []intake.Intake{{ID: 41, Name: "我的批次", Status: intake.StatusCompleted}}}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	h := NewHandler(Dependencies{Auth: auth, Reader: business, IntakeAccess: access})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authenticatedIntakeRequest(http.MethodGet, "/api/v1/intakes", ""))
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte(`"id":41`)) {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if business.calls != 0 || access.listCalls != 1 || access.lastUserID != 7 || access.lastTeamID != 3 || access.lastElevated {
		t.Fatalf("business=%d access=%+v", business.calls, access)
	}
}

func TestListIntakesFailsClosedWhenVisibilityQueryFails(t *testing.T) {
	business := &intakeBusinessSpy{}
	access := &intakeAccessStub{err: errors.New("ownership database unavailable")}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchView}}}
	h := NewHandler(Dependencies{Auth: auth, Reader: business, IntakeAccess: access})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authenticatedIntakeRequest(http.MethodGet, "/api/v1/intakes", ""))
	if rec.Code != http.StatusServiceUnavailable || !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"AUTH_POLICY_UNAVAILABLE"`)) {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if business.calls != 0 {
		t.Fatalf("global reader calls = %d, want zero", business.calls)
	}
}

func TestCreateIntakePassesAuthenticatedActorScope(t *testing.T) {
	api := &fakeIntakeAPI{createResult: intake.Intake{ID: 11, Status: intake.StatusPending}}
	auth := &fakeAuthService{user: authn.User{ID: 7, TeamID: 3, Role: "member", Capabilities: []string{CapabilityBatchConfigure}}}
	h := NewHandler(Dependencies{Auth: auth, Intakes: api})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authenticatedIntakeRequest(http.MethodPost, "/api/v1/intakes", `{"groups":[{"source":"知乎","platformId":"15","books":[{"bookId":"1"}]}]}`))
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if api.createInput.Actor != (intake.ActorScope{UserID: 7, TeamID: 3}) {
		t.Fatalf("actor = %+v", api.createInput.Actor)
	}
}
