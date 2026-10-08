package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/webui"
)

func TestRetiredIssuesAndAgentAPIsAlwaysReturnSafeNotFound(t *testing.T) {
	ui := webui.NewHandler(
		NewHandler(),
		fstest.MapFS{"index.html": {Data: []byte("<h1>SPA fallback must not serve retired APIs</h1>")}},
		webui.BuildInfo{},
	)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "issues", method: http.MethodGet, path: "/api/v1/issues?page=1"},
		{name: "list projects", method: http.MethodGet, path: "/api/v1/agent/projects"},
		{name: "create project", method: http.MethodPost, path: "/api/v1/agent/projects"},
		{name: "delete project", method: http.MethodDelete, path: "/api/v1/agent/projects/9"},
		{name: "list messages", method: http.MethodGet, path: "/api/v1/agent/projects/9/messages"},
		{name: "continue project", method: http.MethodPost, path: "/api/v1/agent/projects/9/messages"},
		{name: "list executions", method: http.MethodGet, path: "/api/v1/agent/projects/9/executions"},
		{name: "list skills", method: http.MethodGet, path: "/api/v1/agent/skills"},
		{name: "create skill", method: http.MethodPost, path: "/api/v1/agent/skills"},
		{name: "read canvas", method: http.MethodGet, path: "/api/v1/agent/projects/9/canvas"},
		{name: "save canvas", method: http.MethodPut, path: "/api/v1/agent/projects/9/canvas"},
		{name: "list canvas versions", method: http.MethodGet, path: "/api/v1/agent/projects/9/canvas/versions"},
		{name: "restore canvas", method: http.MethodPost, path: "/api/v1/agent/projects/9/canvas/versions/2/restore"},
		{name: "list attachments", method: http.MethodGet, path: "/api/v1/agent/projects/9/attachments"},
		{name: "upload attachment", method: http.MethodPost, path: "/api/v1/agent/projects/9/attachments"},
		{name: "read attachment", method: http.MethodGet, path: "/api/v1/agent/projects/9/attachments/3/content"},
		{name: "delete attachment", method: http.MethodDelete, path: "/api/v1/agent/projects/9/attachments/3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ui.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			assertErrorEnvelope(t, rec, http.StatusNotFound, "NOT_FOUND")
			if strings.Contains(rec.Body.String(), "SPA fallback") {
				t.Fatalf("retired API fell through to SPA: %s", rec.Body.String())
			}
		})
	}
}
