package material

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	worktask "yizhanchengming/magic-cut-workbench/internal/task"
)

type httpRepository struct{ items []Material }

func (r *httpRepository) SaveImport(context.Context, Material, worktask.Task) error { return nil }
func (r *httpRepository) List(context.Context, string, string) ([]Material, error) {
	return r.items, nil
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
