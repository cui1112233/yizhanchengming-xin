package task

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type httpStore struct {
	canceled string
	retried  string
}

func (s *httpStore) List(context.Context, string, int) ([]Task, error) {
	return []Task{{ID: "mct_1", Status: StatusRunning}}, nil
}
func (s *httpStore) Cancel(_ context.Context, _ string, id string) error { s.canceled = id; return nil }
func (s *httpStore) Retry(_ context.Context, _ string, id string) error  { s.retried = id; return nil }

type httpActions struct {
	canceled string
	retried  string
}

func (a *httpActions) CancelTask(_ context.Context, _ string, id string) (bool, error) {
	a.canceled = id
	return true, nil
}
func (a *httpActions) RetryTask(_ context.Context, _ string, id string) (bool, error) {
	a.retried = id
	return true, nil
}

func TestTaskHandlerListsAndCancelsOwnedTasks(t *testing.T) {
	store := &httpStore{}
	handler := NewHandler(store, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/magic-cut/tasks", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"id":"mct_1"`) {
		t.Fatalf("status=%d body=%s", list.Code, list.Body.String())
	}
	cancel := httptest.NewRecorder()
	handler.ServeHTTP(cancel, httptest.NewRequest(http.MethodPost, "/api/magic-cut/tasks/mct_1/cancel", nil))
	if cancel.Code != http.StatusOK || store.canceled != "mct_1" {
		t.Fatalf("status=%d canceled=%q", cancel.Code, store.canceled)
	}
}

func TestTaskHandlerDelegatesRunningExportActionsBeforeGenericStore(t *testing.T) {
	store := &httpStore{}
	actions := &httpActions{}
	handler := NewHandlerWithActions(store, func(context.Context, *http.Request) (string, error) { return "alice", nil }, actions)
	cancel := httptest.NewRecorder()
	handler.ServeHTTP(cancel, httptest.NewRequest(http.MethodPost, "/api/magic-cut/tasks/mct_export/cancel", nil))
	if cancel.Code != http.StatusOK || actions.canceled != "mct_export" || store.canceled != "" {
		t.Fatalf("status=%d delegated=%q generic=%q", cancel.Code, actions.canceled, store.canceled)
	}
}
