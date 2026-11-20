package project

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeRepository struct {
	createdOwner string
	createdInput CreateInput
	projects     []Project
}

func (f *fakeRepository) Create(_ context.Context, owner string, input CreateInput) (Project, error) {
	f.createdOwner, f.createdInput = owner, input
	return Project{ID: "project-1", Owner: owner, Name: input.Name, Width: input.Width, Height: input.Height, FrameRate: input.FrameRate}, nil
}

func (f *fakeRepository) List(_ context.Context, _ string) ([]Project, error) { return f.projects, nil }

func TestHTTPCreateUsesAuthenticatedOwnerAndReturnsPersistedProject(t *testing.T) {
	repository := &fakeRepository{}
	handler := NewHandler(repository, func(context.Context, *http.Request) (string, error) { return "owner-a", nil })
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/magic-cut/projects", strings.NewReader(`{"name":"竖屏工程","width":1080,"height":1920,"frameRate":30,"idempotencyKey":"create-1"}`))
	request.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.createdOwner != "owner-a" || repository.createdInput.Name != "竖屏工程" {
		t.Fatalf("owner=%q input=%+v", repository.createdOwner, repository.createdInput)
	}
	var response struct {
		Project Project `json:"project"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Project.ID != "project-1" {
		t.Fatalf("project=%+v", response.Project)
	}
}

func TestHTTPListReturnsAnArrayForAnEmptyOwnerLibrary(t *testing.T) {
	handler := NewHandler(&fakeRepository{projects: []Project{}}, func(context.Context, *http.Request) (string, error) { return "owner-a", nil })
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"projects":[]`) {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
