package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
)

type fakeStoryboardService struct{ recompileCalls int }

func (*fakeStoryboardService) Storyboard(context.Context, int64, int64) (generation.StoryboardDocument, error) {
	return generation.StoryboardDocument{}, nil
}
func (*fakeStoryboardService) SaveStoryboardCard(context.Context, int64, int64, generation.SaveStoryboardCardRequest) (generation.StoryboardDocument, error) {
	return generation.StoryboardDocument{}, nil
}
func (*fakeStoryboardService) DeleteStoryboardCard(context.Context, int64, int64, int64, int) (generation.StoryboardDocument, error) {
	return generation.StoryboardDocument{}, nil
}
func (*fakeStoryboardService) ReorderStoryboard(context.Context, int64, int64, []int64, int) (generation.StoryboardDocument, error) {
	return generation.StoryboardDocument{}, nil
}
func (f *fakeStoryboardService) RecompileStoryboard(context.Context, int64, int64, string) (generation.BookGenerationResult, error) {
	f.recompileCalls++
	return generation.BookGenerationResult{}, nil
}

func TestRuntimeStoryboardRecompileFailsClosedWithoutCallingLegacyService(t *testing.T) {
	legacy := &fakeStoryboardService{}
	h := NewHandler(Dependencies{
		Auth:                  &fakeAuthService{user: authn.User{ID: 5, Capabilities: []string{CapabilityBatchExecute}}},
		BatchProjectAccess:    &fakeRuntimeAccess{allowed: true},
		BatchProjectLifecycle: &fakeBatchProjectLifecycle{},
		ScriptStoryboards:     legacy,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/batch-projects/3/books/11/storyboard/recompile", bytes.NewBufferString(`{"requestId":"compile-1"}`))
	req.AddCookie(&http.Cookie{Name: AccessCookieName, Value: "valid"})
	sameOrigin(req)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "STORYBOARD_RECOMPILE_ASYNC_REQUIRED") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if legacy.recompileCalls != 0 {
		t.Fatalf("legacy recompile calls=%d", legacy.recompileCalls)
	}
}
