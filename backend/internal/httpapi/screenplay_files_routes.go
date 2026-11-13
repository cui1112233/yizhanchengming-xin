package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"qiantie/backend/internal/screenplay"
	"strings"
)

type screenplaySourceLine struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

func numberedScreenplaySource(text string) (string, []screenplaySourceLine) {
	normalized := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	hash := sha256.Sum256([]byte(normalized))
	lines := []screenplaySourceLine{}
	for i, line := range strings.Split(normalized, "\n") {
		lines = append(lines, screenplaySourceLine{i + 1, line})
	}
	return hex.EncodeToString(hash[:]), lines
}
func registerScreenplayFileRoutes(mux *http.ServeMux, auth BridgeAuth, authors *screenplay.AuthoringService) {
	if authors == nil || authors.Store == nil {
		return
	}
	base := "/api/comic-screenplays/projects/{id}"
	mux.Handle("GET "+base+"/source-numbered", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, _ := BridgeIdentityFromContext(r.Context())
		owner, e := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "read")
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		p, e := authors.Store.GetProject(r.Context(), owner, r.PathValue("id"))
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		hash, lines := numberedScreenplaySource(p.SourceText)
		writeJSON(w, 200, map[string]any{"format": "yzcm-numbered-source-v1", "source_sha256": hash, "lines": lines})
	})))
	mux.Handle("POST "+base+"/storyboards/import", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, _ := BridgeIdentityFromContext(r.Context())
		owner, e := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "edit")
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		var in struct {
			Expected   int64                 `json:"expected_revision"`
			SourceHash string                `json:"source_sha256"`
			Board      screenplay.Storyboard `json:"storyboard"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, screenplayJSONLimit))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&in) != nil {
			writeScreenplayError(w, screenplay.ErrInvalidArgument)
			return
		}
		var result screenplay.Storyboard
		e = authors.WithLock(r.Context(), owner, r.PathValue("id"), func() error {
			resolved, err := authors.Resolve(r.Context(), actor.Username, r.PathValue("id"), "edit")
			if err != nil {
				return err
			}
			if resolved != owner {
				return screenplay.ErrAccessDenied
			}
			p, err := authors.Store.GetProject(r.Context(), owner, r.PathValue("id"))
			if err != nil {
				return err
			}
			hash, _ := numberedScreenplaySource(p.SourceText)
			if in.SourceHash != hash {
				return fmt.Errorf("%w: 原文已变化，请重新导出编号原文后核对", screenplay.ErrRevisionConflict)
			}
			b := in.Board
			if b.Engine != screenplay.EngineXiaoluoV7 && b.Engine != screenplay.EngineXiaoyunqueV3 && b.Engine != screenplay.EngineXiaoluoV9 {
				return screenplay.ErrInvalidArgument
			}
			if b.EpisodeNo < 1 || len(b.Shots) == 0 || len(b.Shots) > 2000 {
				return screenplay.ErrInvalidArgument
			}
			b.ID = ""
			b.OwnerUsername = owner
			b.ProjectID = p.ID
			b.SourceSHA256 = hash
			b.Status = "done"
			for i := range b.Shots {
				sh := &b.Shots[i]
				if sh.Duration <= 0 || sh.Duration > 600 || math.IsNaN(sh.Duration) || math.IsInf(sh.Duration, 0) || strings.TrimSpace(sh.VideoDesc) == "" {
					return screenplay.ErrInvalidArgument
				}
				sh.ID = ""
				sh.StoryboardID = ""
				sh.ProjectID = p.ID
				sh.OwnerUsername = owner
				sh.Seq = i + 1
			}
			store, ok := authors.Store.(interface {
				SaveStoryboardRevision(context.Context, screenplay.Storyboard, int64) (screenplay.Storyboard, error)
			})
			if !ok {
				return screenplay.ErrInvalidState
			}
			result, err = store.SaveStoryboardRevision(r.Context(), b, in.Expected)
			if err != nil {
				return err
			}
			if err = screenplay.InvalidateAuthoringContent(r.Context(), authors.Store, owner, p.ID, "storyboard"); err != nil {
				return err
			}
			state, err := authors.Load(r.Context(), owner, p.ID)
			if err != nil {
				return err
			}
			_, err = authors.Save(r.Context(), owner, p.ID, state, state.Revision)
			return err
		})
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"storyboard": result})
	})))
}
