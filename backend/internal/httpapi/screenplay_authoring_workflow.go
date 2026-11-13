package httpapi

import (
	"bytes"

	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
)

type screenplayWorkflowResponse struct {
	headers http.Header
	status  int
	body    bytes.Buffer
}

func (r *screenplayWorkflowResponse) Header() http.Header { return r.headers }
func (r *screenplayWorkflowResponse) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}
func (r *screenplayWorkflowResponse) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = 200
	}
	return r.body.Write(b)
}
func screenplayWorkflowRight(pattern string) string {
	if strings.HasPrefix(pattern, "GET ") || strings.HasSuffix(pattern, "/source/parse") {
		return "read"
	}
	if strings.HasSuffix(pattern, "/generate") || strings.HasSuffix(pattern, "/assets/extract") || strings.HasSuffix(pattern, "/batch") || strings.HasSuffix(pattern, "/batch/run-next") {
		return "generate"
	}
	return "edit"
}
func screenplayWorkflowChange(pattern string) string {
	if strings.HasPrefix(pattern, "GET ") || strings.HasSuffix(pattern, "/source/parse") || strings.Contains(pattern, "/batch") || pattern == "DELETE /api/comic-screenplays/projects/{id}" {
		return ""
	}
	if strings.Contains(pattern, "/storyboards") {
		return "storyboard"
	}
	if strings.Contains(pattern, "/assets") {
		return "asset"
	}
	if strings.HasSuffix(pattern, "/source") {
		return "source"
	}
	if pattern == "PATCH /api/comic-screenplays/projects/{id}" {
		return "config"
	}
	return "document"
}
func serveScreenplayWorkflow(w http.ResponseWriter, r *http.Request, actor, pattern string, authors *screenplay.AuthoringService, fn func(http.ResponseWriter, *http.Request, string)) {
	right := screenplayWorkflowRight(pattern)
	owner, e := authors.Resolve(r.Context(), actor, r.PathValue("id"), right)
	if e != nil {
		authoringError(w, e)
		return
	}
	if pattern == "DELETE /api/comic-screenplays/projects/{id}" && owner != actor {
		authoringError(w, screenplay.ErrAccessDenied)
		return
	}
	if right == "read" || strings.HasSuffix(pattern, "/batch/run-next") {
		fn(w, r, owner)
		return
	}
	e = authors.WithLock(r.Context(), owner, r.PathValue("id"), func() error {
		// Check again after locking, including grants revoked while waiting for the lock.
		resolved, er := authors.Resolve(r.Context(), actor, r.PathValue("id"), right)
		if er != nil {
			return er
		}
		if resolved != owner {
			return screenplay.ErrAccessDenied
		}
		if strings.HasSuffix(pattern, "/batch/cancel") || pattern == "DELETE /api/comic-screenplays/projects/{id}" {
			if repo, ok := authors.Store.(screenplay.ProductionRepository); ok {
				if er = repo.SetProductionState(r.Context(), owner, r.PathValue("id"), "paused", ""); er != nil {
					return er
				}
			}
		}
		kind := screenplayWorkflowChange(pattern)
		var before screenplay.Project
		if kind == "config" {
			before, er = authors.Store.GetProject(r.Context(), owner, r.PathValue("id"))
			if er != nil {
				return er
			}
		}
		response := &screenplayWorkflowResponse{headers: make(http.Header)}
		fn(response, r, owner)
		if response.status >= 200 && response.status < 300 && kind != "" {
			if kind == "config" {
				after, er := authors.Store.GetProject(r.Context(), owner, r.PathValue("id"))
				if er != nil {
					return er
				}
				if before.ScriptMode == after.ScriptMode && before.AspectRatio == after.AspectRatio && before.ArtStyle == after.ArtStyle && before.SelectedStyle == after.SelectedStyle && before.Synopsis == after.Synopsis && before.ModelID == after.ModelID && before.PromptID == after.PromptID && before.AssetDegree == after.AssetDegree {
					kind = ""
				}
			}
			if kind != "" {
				if er = screenplay.InvalidateAuthoringContent(r.Context(), authors.Store, owner, r.PathValue("id"), kind); er != nil {
					return er
				}
				st, er := authors.Load(r.Context(), owner, r.PathValue("id"))
				if er != nil {
					return er
				}
				if _, er = authors.Save(r.Context(), owner, r.PathValue("id"), st, st.Revision); er != nil {
					return er
				}
			}
		}
		for key, values := range response.headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		status := response.status
		if status == 0 {
			status = 200
		}
		w.WriteHeader(status)
		_, er = w.Write(response.body.Bytes())
		return er
	})
	if e != nil {
		authoringError(w, e)
	}
}
