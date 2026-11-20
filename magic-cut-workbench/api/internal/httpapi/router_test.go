package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthIdentifiesMagicCutService(t *testing.T) {
	handler := New(Dependencies{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		OK      bool   `json:"ok"`
		Service string `json:"service"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.OK || body.Service != "magic-cut-4040" {
		t.Fatalf("body = %+v", body)
	}
}

func TestUnknownAPIRouteReturnsJSONNotTheFrontend(t *testing.T) {
	handler := New(Dependencies{})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/not-real", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", recorder.Code, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
}

func TestLoginRoutesUseTheAccountBridge(t *testing.T) {
	auth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/login/session" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := New(Dependencies{Auth: auth})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/login/session", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAuthBootstrapUsesTheDedicatedCrossPortHandler(t *testing.T) {
	bootstrap := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := New(Dependencies{AuthBootstrap: bootstrap})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/auth/bootstrap", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProjectRoutesUseTheIsolatedProjectHandler(t *testing.T) {
	projects := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := New(Dependencies{Projects: projects})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/magic-cut/projects", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestMaterialRoutesUseTheIsolatedMaterialHandler(t *testing.T) {
	materials := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	handler := New(Dependencies{Materials: materials})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/magic-cut/projects/mcp_1/materials", nil))

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d; body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTimelineRoutesUseTheVersionedTimelineHandler(t *testing.T) {
	timeline := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted) })
	handler := New(Dependencies{Timeline: timeline})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/magic-cut/projects/mcp_1/timeline", nil))
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d", recorder.Code)
	}
}

func TestTaskRoutesUseTheTaskCenterHandler(t *testing.T) {
	tasks := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted) })
	handler := New(Dependencies{Tasks: tasks})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/magic-cut/tasks", nil))
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status=%d", recorder.Code)
	}
}
