package material

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type httpRepository struct {
	items []Material
	item  Material
}

func (r *httpRepository) SaveImport(context.Context, Material, worktask.Task) error { return nil }
func (r *httpRepository) List(context.Context, string, string) ([]Material, error) {
	return r.items, nil
}
func (r *httpRepository) Get(_ context.Context, owner, projectID, materialID string) (Material, error) {
	if r.item.Owner != owner || r.item.ProjectID != projectID || r.item.ID != materialID {
		return Material{}, errors.New("material not found")
	}
	return r.item, nil
}

func TestMaterialHandlerRejectsUnsupportedUploadWithExactCode(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "setup.exe")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("not media"))
	_ = writer.Close()
	repository := &httpRepository{}
	service := NewService(ServiceOptions{StorageRoot: t.TempDir(), Repository: repository, NewMaterialID: func() string { return "mcm_1" }, NewTaskID: func() string { return "mct_1" }})
	handler := NewHandler(service, repository, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	request := httptest.NewRequest(http.MethodPost, "/api/magic-cut/projects/mcp_1/materials", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnsupportedMediaType || !strings.Contains(response.Body.String(), `"code":"FORMAT_NOT_SUPPORTED"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestMaterialHandlerListsProjectMaterials(t *testing.T) {
	repository := &httpRepository{items: []Material{{ID: "mcm_1", ProjectID: "mcp_1", OriginalName: "clip.mp4", MediaType: "video", Status: "ready"}}}
	handler := NewHandler(nil, repository, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects/mcp_1/materials", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Materials []Material `json:"materials"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Materials) != 1 || payload.Materials[0].ID != "mcm_1" {
		t.Fatalf("payload=%+v", payload)
	}
}

func TestMaterialHandlerStreamsOwnedFileWithRangeSupport(t *testing.T) {
	path := t.TempDir() + "/clip.mp4"
	if err := os.WriteFile(path, []byte("0123456789"), 0o644); err != nil {
		t.Fatal(err)
	}
	repository := &httpRepository{item: Material{ID: "mcm_1", ProjectID: "mcp_1", Owner: "alice", OriginalName: "clip.mp4", MediaType: "video", StorageKey: path, Status: "ready"}}
	handler := NewHandler(nil, repository, func(context.Context, *http.Request) (string, error) { return "alice", nil })
	request := httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects/mcp_1/materials/mcm_1/file", nil)
	request.Header.Set("Range", "bytes=2-5")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusPartialContent || response.Body.String() != "2345" {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
	if response.Header().Get("Accept-Ranges") != "bytes" || response.Header().Get("Content-Type") != "video/mp4" {
		t.Fatalf("headers=%v", response.Header())
	}
}

func TestMaterialHandlerDoesNotExposeAnotherOwnersFile(t *testing.T) {
	repository := &httpRepository{item: Material{ID: "mcm_1", ProjectID: "mcp_1", Owner: "alice", StorageKey: t.TempDir() + "/secret.mp4", Status: "ready"}}
	handler := NewHandler(nil, repository, func(context.Context, *http.Request) (string, error) { return "bob", nil })
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects/mcp_1/materials/mcm_1/file", nil))

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"MATERIAL_NOT_FOUND"`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
