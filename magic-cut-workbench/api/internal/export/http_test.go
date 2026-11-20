package export

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type creatorFixture struct{ input CreateInput }

func (f *creatorFixture) Create(_ context.Context, owner string, input CreateInput) (Export, worktask.Task, error) {
	f.input = input
	return Export{ID: "mce_1", ProjectID: input.ProjectID, TaskID: "mct_1", OutputName: input.OutputName}, worktask.Task{ID: "mct_1", ProjectID: input.ProjectID, Status: worktask.StatusPending}, nil
}

type fileCreatorFixture struct {
	creatorFixture
	file Export
}

func (f *fileCreatorFixture) File(context.Context, string, string) (Export, error) {
	return f.file, nil
}

func TestHTTPHandlerCreatesOwnedExport(t *testing.T) {
	creator := &creatorFixture{}
	handler := NewHandler(creator, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	body := []byte(`{"projectId":"mcp_1","outputName":"成片","width":640,"height":360,"frameRate":25,"videoCodec":"h264","audioCodec":"aac","container":"mp4"}`)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/magic-cut/exports", bytes.NewReader(body)))
	if recorder.Code != http.StatusAccepted || !strings.Contains(recorder.Body.String(), `"exportId":"mce_1"`) || !strings.Contains(recorder.Body.String(), `"taskId":"mct_1"`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if creator.input.ProjectID != "mcp_1" || creator.input.OutputName != "成片" {
		t.Fatalf("input=%+v", creator.input)
	}
}

func TestHTTPHandlerRejectsUnauthenticatedExport(t *testing.T) {
	handler := NewHandler(&creatorFixture{}, func(context.Context, *http.Request) (string, error) { return "", context.Canceled })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/magic-cut/exports", strings.NewReader(`{}`)))
	if recorder.Code != http.StatusUnauthorized || !strings.Contains(recorder.Body.String(), "AUTH_REQUIRED") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHTTPHandlerDownloadsOwnedCompletedExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "成片.mp4")
	if err := os.WriteFile(path, []byte("playable-video"), 0o644); err != nil {
		t.Fatal(err)
	}
	creator := &fileCreatorFixture{file: Export{ID: "mce_1", Owner: "alice", OutputName: "成片", StorageKey: path, SHA256: "abc"}}
	handler := NewHandler(creator, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/magic-cut/exports/mce_1/file", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != "playable-video" || !strings.Contains(recorder.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
