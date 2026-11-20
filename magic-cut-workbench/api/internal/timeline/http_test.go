package timeline

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type conflictRepository struct{}

func (conflictRepository) Get(context.Context, string, string) (Timeline, error) { return Empty(), nil }
func (conflictRepository) Put(context.Context, string, string, Timeline, uint64) (Timeline, error) {
	return Timeline{}, &RevisionConflictError{CurrentRevision: 4}
}

func TestHandlerReturnsCurrentRevisionOnConflict(t *testing.T) {
	handler := NewHandler(conflictRepository{}, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	body := `{"expectedRevision":2,"timeline":{"revision":2,"durationMs":0,"tracks":[]}}`
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/api/magic-cut/projects/mcp_1/timeline", strings.NewReader(body)))
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"currentRevision":4`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
