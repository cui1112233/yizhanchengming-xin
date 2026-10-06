package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
)

type fakeShuihuoService struct {
	create func(context.Context, shuihuo.Actor, shuihuo.CreateProjectInput) (shuihuo.ReadModel, error)
	get    func(context.Context, shuihuo.Actor, int64) (shuihuo.ReadModel, error)
	fixed  func(context.Context, shuihuo.Actor, int64, shuihuo.FixedSegmentationInput) ([]shuihuo.Candidate, error)
}

func (f fakeShuihuoService) CreateProject(ctx context.Context, actor shuihuo.Actor, input shuihuo.CreateProjectInput) (shuihuo.ReadModel, error) {
	if f.create != nil {
		return f.create(ctx, actor, input)
	}
	return shuihuo.ReadModel{}, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) ListProjects(context.Context, shuihuo.Actor) ([]shuihuo.Project, error) {
	return []shuihuo.Project{}, nil
}
func (f fakeShuihuoService) GetProject(ctx context.Context, actor shuihuo.Actor, id int64) (shuihuo.ReadModel, error) {
	if f.get != nil {
		return f.get(ctx, actor, id)
	}
	return shuihuo.ReadModel{}, shuihuo.ErrNotFound
}
func (f fakeShuihuoService) ReplaceSource(context.Context, shuihuo.Actor, int64, string) (shuihuo.ReadModel, error) {
	return shuihuo.ReadModel{}, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) DeleteProject(context.Context, shuihuo.Actor, int64) error {
	return nil
}
func (f fakeShuihuoService) ParagraphSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error) {
	return nil, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) FixedSegmentation(ctx context.Context, actor shuihuo.Actor, id int64, input shuihuo.FixedSegmentationInput) ([]shuihuo.Candidate, error) {
	if f.fixed != nil {
		return f.fixed(ctx, actor, id, input)
	}
	return nil, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) ImportSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error) {
	return nil, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) SmartSegmentation(context.Context, shuihuo.Actor, int64, shuihuo.SegmentationInput) ([]shuihuo.Candidate, error) {
	return nil, shuihuo.ErrSmartUnavailable
}
func (f fakeShuihuoService) ConfirmSegmentation(context.Context, shuihuo.Actor, int64, []shuihuo.Candidate) (shuihuo.ReadModel, error) {
	return shuihuo.ReadModel{}, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) CreateSegment(context.Context, shuihuo.Actor, int64, shuihuo.SegmentInput) (shuihuo.Segment, error) {
	return shuihuo.Segment{}, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) UpdateSegment(context.Context, shuihuo.Actor, int64, shuihuo.SegmentInput) (shuihuo.Segment, error) {
	return shuihuo.Segment{}, shuihuo.ErrInvalid
}
func (f fakeShuihuoService) DeleteSegment(context.Context, shuihuo.Actor, int64) error {
	return nil
}
func (f fakeShuihuoService) ReorderSegments(context.Context, shuihuo.Actor, int64, []int64) (shuihuo.ReadModel, error) {
	return shuihuo.ReadModel{}, shuihuo.ErrInvalid
}

func TestShuihuoCreateProjectUsesExistingAuthBoundaryActorContract(t *testing.T) {
	service := fakeShuihuoService{create: func(_ context.Context, actor shuihuo.Actor, input shuihuo.CreateProjectInput) (shuihuo.ReadModel, error) {
		if !actor.BypassOwnership {
			t.Fatalf("auth-disabled test handler should use explicit bypass actor")
		}
		if input.Name != "作品A" {
			t.Fatalf("name=%q", input.Name)
		}
		return shuihuo.ReadModel{Project: shuihuo.Project{ID: 7, Name: input.Name, ProductionMode: "commentary", SegmentationStatus: shuihuo.SegmentationDraft}}, nil
	}}
	handler := NewHandler(Dependencies{Shuihuo: service})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shuihuo-production/projects", strings.NewReader(`{"name":"作品A"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var project shuihuo.Project
	if err := json.Unmarshal(rec.Body.Bytes(), &project); err != nil {
		t.Fatal(err)
	}
	if project.ID != 7 || project.Name != "作品A" {
		t.Fatalf("project=%#v", project)
	}
}

func TestShuihuoProjectOwnershipForbiddenRemains403(t *testing.T) {
	service := fakeShuihuoService{get: func(context.Context, shuihuo.Actor, int64) (shuihuo.ReadModel, error) {
		return shuihuo.ReadModel{}, shuihuo.ErrForbidden
	}}
	handler := NewHandler(Dependencies{Shuihuo: service})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/shuihuo-production/projects/9", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &payload)
	if payload["code"] != "AUTH_FORBIDDEN" {
		t.Fatalf("payload=%v", payload)
	}
}

func TestShuihuoFixedSegmentationRouteReturnsCandidates(t *testing.T) {
	service := fakeShuihuoService{fixed: func(_ context.Context, _ shuihuo.Actor, id int64, input shuihuo.FixedSegmentationInput) ([]shuihuo.Candidate, error) {
		if id != 5 || input.LinesPerSegment != 2 {
			t.Fatalf("id=%d input=%#v", id, input)
		}
		return []shuihuo.Candidate{{Text: "一\n二", Speaker: "旁白"}}, nil
	}}
	handler := NewHandler(Dependencies{Shuihuo: service})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shuihuo-production/projects/5/segmentation/fixed", strings.NewReader(`{"linesPerSegment":2,"text":"一\n二"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Candidates []shuihuo.Candidate `json:"candidates"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Candidates) != 1 || payload.Candidates[0].Text != "一\n二" {
		t.Fatalf("payload=%#v", payload)
	}
}
