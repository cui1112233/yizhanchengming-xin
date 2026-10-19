package httpapi

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type agentAssetObjectTestStore struct{ files map[string][]byte }

func (s *agentAssetObjectTestStore) Put(_ context.Context, asset agentworkspace.Asset, body io.Reader) error {
	contents, err := io.ReadAll(body)
	if err == nil {
		s.files[asset.ObjectKey] = contents
	}
	return err
}

func (s *agentAssetObjectTestStore) Open(_ context.Context, asset agentworkspace.Asset) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.files[asset.ObjectKey])), nil
}

func (s *agentAssetObjectTestStore) Delete(_ context.Context, asset agentworkspace.Asset) error {
	delete(s.files, asset.ObjectKey)
	return nil
}

func TestAgentAssetUploadCreatesProjectScopedRecord(t *testing.T) {
	now := time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC)
	projects := agentworkspace.NewMemoryStore()
	if err := projects.CreateProject(context.Background(), "alice", agentworkspace.Project{ID: "agent-1", Title: "项目", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	objects := &agentAssetObjectTestStore{files: map[string][]byte{}}
	RegisterAgentAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, agentworkspace.NewMemoryAssetStore(), objects, func() time.Time { return now })

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "cover.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("\x89PNG\r\n\x1a\ncanvas")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/agent/tasks/agent-1/assets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	SignBridgeRequest(req, "alice", true, now, "secret")
	created := httptest.NewRecorder()
	mux.ServeHTTP(created, req)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"kind":"image"`) || strings.Contains(created.Body.String(), "agent-assets/") {
		t.Fatalf("upload status=%d body=%s", created.Code, created.Body.String())
	}
	if len(objects.files) != 1 {
		t.Fatalf("objects=%d, want 1", len(objects.files))
	}
}
