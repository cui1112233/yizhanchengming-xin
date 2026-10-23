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
	"qiantie/backend/internal/localexecutor"
)

type agentAssetObjectTestStore struct{ files map[string][]byte }

func (s *agentAssetObjectTestStore) Put(_ context.Context, asset agentworkspace.Asset, body io.Reader) error {
	contents, err := io.ReadAll(body)
	if err == nil {
		s.files[asset.ObjectKey] = contents
	}
	return err
}

func TestAgentExecutorReferenceAssetRouteRequiresPairedOwner(t *testing.T) {
	now := time.Date(2026, 10, 22, 4, 0, 0, 0, time.UTC)
	assets := agentworkspace.NewMemoryAssetStore()
	asset := agentworkspace.Asset{ID: "image-1", ProjectID: "agent-1", Kind: "image", Name: "cover.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/agent-1/image-1.png", ByteSize: 4, CreatedAt: now}
	if err := assets.CreateAsset(context.Background(), "alice", asset); err != nil {
		t.Fatal(err)
	}
	objects := &agentAssetObjectTestStore{files: map[string][]byte{asset.ObjectKey: {137, 80, 78, 71}}}
	service := localexecutor.NewService(localexecutor.NewMemoryStore(), func() time.Time { return now })
	pairing, err := service.CreatePairing(context.Background(), "alice", localexecutor.PlatformDoubao)
	if err != nil {
		t.Fatal(err)
	}
	paired, err := service.Pair(context.Background(), localexecutor.PairInput{Code: pairing.Code, Platform: localexecutor.PlatformDoubao, DeviceName: "test", OS: "windows", Version: "1"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentExecutorReferenceAssetRoutes(mux, service, assets, objects)
	req := httptest.NewRequest(http.MethodGet, agentExecutorReferenceAssetURL("agent-1", "image-1"), nil)
	req.Header.Set("Authorization", "Bearer "+paired.Token)
	result := httptest.NewRecorder()
	mux.ServeHTTP(result, req)
	if result.Code != http.StatusOK || result.Header().Get("Content-Type") != "image/png" || result.Body.String() != "\x89PNG" {
		t.Fatalf("response=%d headers=%v body=%q", result.Code, result.Header(), result.Body.String())
	}
	denied := httptest.NewRecorder()
	mux.ServeHTTP(denied, httptest.NewRequest(http.MethodGet, agentExecutorReferenceAssetURL("agent-1", "image-1"), nil))
	if denied.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous response=%d body=%s", denied.Code, denied.Body.String())
	}
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
