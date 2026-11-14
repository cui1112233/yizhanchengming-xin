package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"qiantie/backend/internal/screenplay"
)

func decodeAssetCheckRequest(w http.ResponseWriter, r *http.Request, input any, emptyOK bool) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(input)
	if emptyOK && err == io.EOF {
		return true
	}
	if err != nil {
		writeScreenplayError(w, fmt.Errorf("%w: 检查请求格式不正确", screenplay.ErrInvalidArgument))
		return false
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		writeScreenplayError(w, screenplay.ErrInvalidArgument)
		return false
	}
	return true
}

// These routes only create/read durable drafts. Model calls belong to the
// server lifetime worker; applying a reviewed draft is an explicit edit action.
func RegisterScreenplayAssetCheckRoutes(mux *http.ServeMux, auth BridgeAuth, runtime *screenplay.AssetCheckRuntime, authors *screenplay.AuthoringService) {
	if mux == nil || runtime == nil || runtime.Store == nil || runtime.Repository == nil {
		return
	}
	base := "/api/comic-screenplays/projects/{id}/asset-checks"
	register := func(pattern, permission string, fn func(http.ResponseWriter, *http.Request, string, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(actor.Username) == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			owner := actor.Username
			var err error
			if authors != nil {
				owner, err = authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), permission)
			} else {
				_, err = runtime.Store.GetProject(r.Context(), owner, r.PathValue("id"))
			}
			if err != nil {
				writeScreenplayError(w, err)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			if r.Method == http.MethodGet {
				fn(w, r, owner, actor.Username)
				return
			}
			// Match authoring writes: serialize the project edit and recheck membership
			// after taking that lock, so a revoked grant cannot authorize a late apply.
			if authors != nil {
				err = authors.WithLock(r.Context(), owner, r.PathValue("id"), func() error {
					current, e := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), permission)
					if e != nil {
						return e
					}
					if current != owner {
						return screenplay.ErrAccessDenied
					}
					fn(w, r, owner, actor.Username)
					return nil
				})
				if err != nil {
					writeScreenplayError(w, err)
				}
				return
			}
			fn(w, r, owner, actor.Username)
		})))
	}
	register("GET "+base, "read", func(w http.ResponseWriter, r *http.Request, owner, actor string) {
		limit := 30
		if raw := r.URL.Query().Get("limit"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 200 {
				writeScreenplayError(w, screenplay.ErrInvalidArgument)
				return
			}
			limit = n
		}
		jobs, err := runtime.Repository.ListAssetCheckJobs(r.Context(), owner, r.PathValue("id"), limit)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		summaries := []screenplay.AssetCheckSummary{}
		for _, job := range jobs {
			summaries = append(summaries, job.Summary())
		}
		writeJSON(w, 200, map[string]any{"jobs": summaries})
	})
	register("GET "+base+"/{jobId}", "read", func(w http.ResponseWriter, r *http.Request, owner, actor string) {
		job, err := runtime.Repository.GetAssetCheckJob(r.Context(), owner, r.PathValue("id"), r.PathValue("jobId"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"job": job.Summary()})
	})
	register("POST "+base, "generate", func(w http.ResponseWriter, r *http.Request, owner, actor string) {
		var input screenplay.AssetCheckStartRequest
		if !decodeAssetCheckRequest(w, r, &input, false) {
			return
		}
		job, err := runtime.Start(r.Context(), owner, r.PathValue("id"), actor, input)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": job.Summary()})
	})
	register("POST "+base+"/{jobId}/cancel", "generate", func(w http.ResponseWriter, r *http.Request, owner, actor string) {
		var input struct{}
		if !decodeAssetCheckRequest(w, r, &input, true) {
			return
		}
		job, err := runtime.Cancel(r.Context(), owner, r.PathValue("id"), r.PathValue("jobId"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"job": job.Summary()})
	})
	for _, operation := range []string{"apply", "discard"} {
		operation := operation
		register("POST "+base+"/{jobId}/"+operation, "edit", func(w http.ResponseWriter, r *http.Request, owner, actor string) {
			var input struct {
				Expected *int64 `json:"expectedRevision"`
			}
			if !decodeAssetCheckRequest(w, r, &input, false) {
				return
			}
			if input.Expected == nil || *input.Expected < 0 {
				writeScreenplayError(w, screenplay.ErrInvalidArgument)
				return
			}
			var job screenplay.AssetCheckJob
			var err error
			if operation == "apply" {
				job, err = runtime.Repository.ApplyAssetCheckJob(r.Context(), owner, r.PathValue("id"), r.PathValue("jobId"), *input.Expected)
			} else {
				job, err = runtime.Discard(r.Context(), owner, r.PathValue("id"), r.PathValue("jobId"), *input.Expected)
			}
			if err != nil {
				writeScreenplayError(w, err)
				return
			}
			writeJSON(w, 200, map[string]any{"job": job.Summary()})
		})
	}
}
