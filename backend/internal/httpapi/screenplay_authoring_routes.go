package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strconv"
	"strings"
)

// ResolveScreenplayAccess never returns an owner's identity before checking a specific right.
func ResolveScreenplayAccess(ctx context.Context, store screenplay.Store, repo screenplay.AuthoringRepository, actor, id, right string) (string, error) {
	return (screenplay.AuthoringService{Store: store, Repository: repo}).Resolve(ctx, actor, id, right)
}
func ScreenplayProjectArchived(ctx context.Context, repo screenplay.AuthoringRepository, o, id string) bool {
	st, e := repo.Load(ctx, o, id)
	return e == nil && st.Archived
}
func RegisterScreenplayAuthoringRoutes(mux *http.ServeMux, auth BridgeAuth, store screenplay.Store, repo screenplay.AuthoringRepository, resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)) {
	if mux == nil || store == nil || repo == nil {
		return
	}
	svc := screenplay.AuthoringService{Store: store, Repository: repo}
	handle := func(pattern, right string, fn func(http.ResponseWriter, *http.Request, string, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ident, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			owner := ident.Username
			if r.PathValue("id") != "" {
				var e error
				owner, e = svc.Resolve(r.Context(), ident.Username, r.PathValue("id"), right)
				if e != nil {
					authoringError(w, e)
					return
				}
			}
			if r.Method != http.MethodGet && right == "edit" && !strings.HasSuffix(r.URL.Path, "/assistant/apply") {
				err := svc.WithLock(r.Context(), owner, r.PathValue("id"), func() error { fn(w, r, owner, ident.Username); return nil })
				if err != nil {
					authoringError(w, err)
				}
			} else {
				fn(w, r, owner, ident.Username)
			}
		})))
	}
	decode := func(w http.ResponseWriter, r *http.Request, v any) bool {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, screenplayJSONLimit))
		d.DisallowUnknownFields()
		if e := d.Decode(v); e != nil {
			authoringError(w, screenplay.ErrInvalidArgument)
			return false
		}
		return true
	}
	respond := func(w http.ResponseWriter, st screenplay.AuthoringState, e error, actor string) {
		if e != nil {
			authoringError(w, e)
			return
		}
		access := map[string]any{"owner": st.Owner == actor, "read": true, "edit": st.Owner == actor, "generate": st.Owner == actor}
		for _, g := range st.Members {
			if g.Username == actor {
				access["edit"] = g.Edit
				access["generate"] = g.Generate
			}
		}
		if actor != st.Owner {
			st.Members = nil
		}
		payload := map[string]any{"state": st, "access": access}
		if st.Archived {
			payload["notice"] = "作品已归档，后续生产和定时计划已暂停；恢复作品不会自动重新生成。"
		}
		writeJSON(w, 200, payload)
	}
	base := "/api/comic-screenplays/projects/{id}"
	for _, path := range []string{"chapters", "display", "assistant/chat", "team"} {
		handle("GET "+base+"/"+path, "read", func(w http.ResponseWriter, r *http.Request, o, a string) {
			st, e := svc.Load(r.Context(), o, r.PathValue("id"))
			respond(w, st, e, a)
		})
	}
	for _, path := range []string{"chapters", "display", "team"} {
		path := path
		handle("PUT "+base+"/"+path, "edit", func(w http.ResponseWriter, r *http.Request, o, a string) {
			var in struct {
				Expected int64                    `json:"expected_revision"`
				Chapters []screenplay.Chapter     `json:"chapters"`
				Display  map[string]any           `json:"display"`
				Members  []screenplay.MemberGrant `json:"members"`
			}
			if !decode(w, r, &in) {
				return
			}
			st, e := svc.Load(r.Context(), o, r.PathValue("id"))
			if e != nil {
				authoringError(w, e)
				return
			}
			switch path {
			case "chapters":
				st.Chapters = in.Chapters
			case "display":
				if e = validateAuthoringDisplay(in.Display); e != nil {
					authoringError(w, e)
					return
				}
				st.Display = in.Display
			case "team":
				if o != a {
					authoringError(w, screenplay.ErrAccessDenied)
					return
				}
				seen := map[string]bool{}
				for _, g := range in.Members {
					if g.Username == o || seen[g.Username] || g.Username == "" || (!g.Read && (g.Edit || g.Generate)) {
						authoringError(w, screenplay.ErrInvalidArgument)
						return
					}
					seen[g.Username] = true
					if auth.AuthStore == nil {
						authoringError(w, fmt.Errorf("%w: 账号校验服务未配置", screenplay.ErrInvalidState))
						return
					}
					account, er := auth.AuthStore.FindAccount(r.Context(), g.Username)
					if er != nil || !account.Active {
						authoringError(w, fmt.Errorf("%w: 分享账号不存在或未启用", screenplay.ErrInvalidArgument))
						return
					}
				}
				st.Members = in.Members
			}
			st, e = svc.Save(r.Context(), o, st.ProjectID, st, in.Expected)
			if e == nil && path == "chapters" {
				for _, c := range st.Chapters {
					if strings.TrimSpace(c.ScriptText) != "" {
						_, e = screenplay.PublishChapterCollection(r.Context(), store, o, st.ProjectID, st.Chapters)
						break
					}
				}
				if e == nil {
					e = screenplay.InvalidateAuthoringContent(r.Context(), store, o, st.ProjectID, "chapter")
				}
			}
			respond(w, st, e, a)
		})
	}
	for _, path := range []string{"archive", "restore"} {
		path := path
		handle("POST "+base+"/"+path, "edit", func(w http.ResponseWriter, r *http.Request, o, a string) {
			if o != a {
				authoringError(w, screenplay.ErrAccessDenied)
				return
			}
			var in struct {
				Expected int64 `json:"expected_revision"`
			}
			if !decode(w, r, &in) {
				return
			}
			st, e := svc.Load(r.Context(), o, r.PathValue("id"))
			if e == nil {
				if st.Revision != in.Expected {
					authoringError(w, screenplay.ErrRevisionConflict)
					return
				}
				if path == "archive" {
					if e = screenplay.PauseArchivedProduction(r.Context(), store, o, st.ProjectID); e != nil {
						authoringError(w, e)
						return
					}
				}
				st.Archived = path == "archive"
				st, e = svc.Save(r.Context(), o, st.ProjectID, st, in.Expected)
			}
			respond(w, st, e, a)
		})
	}
	for _, path := range []string{"archived", "shared"} {
		path := path
		handle("GET /api/comic-screenplays/"+path, "read", func(w http.ResponseWriter, r *http.Request, o, a string) {
			rows, e := repo.List(r.Context())
			if e != nil {
				authoringError(w, e)
				return
			}
			projects := []screenplay.Project{}
			for _, st := range rows {
				want := path == "archived" && st.Owner == a && st.Archived
				if path == "shared" && !st.Archived && st.Owner != a {
					resolved, er := svc.Resolve(r.Context(), a, st.ProjectID, "read")
					want = er == nil && resolved == st.Owner
				}
				if want {
					p, er := store.GetProject(r.Context(), st.Owner, st.ProjectID)
					if er == nil {
						projects = append(projects, p)
					}
				}
			}
			writeJSON(w, 200, map[string]any{"projects": projects})
		})
	}
	handle("POST "+base+"/assistant/apply", "edit", func(w http.ResponseWriter, r *http.Request, o, a string) {
		var in struct {
			Expected   int64  `json:"expected_revision"`
			ProposalID string `json:"proposal_id"`
			Confirmed  bool   `json:"confirmed"`
		}
		if !decode(w, r, &in) {
			return
		}
		st, e := svc.Apply(r.Context(), o, r.PathValue("id"), in.ProposalID, in.Expected, in.Confirmed)
		respond(w, st, e, a)
	})
	for _, path := range []string{"chapters/plan", "chapters/{chapter}/script", "assistant/chat"} {
		path := path
		handle("POST "+base+"/"+path, "generate", func(w http.ResponseWriter, r *http.Request, o, a string) {
			var in struct {
				Expected    int64  `json:"expected_revision"`
				Instruction string `json:"instruction"`
				Message     string `json:"message"`
			}
			if !decode(w, r, &in) {
				return
			}
			var out screenplay.AuthoringState
			e := svc.WithLock(r.Context(), o, r.PathValue("id"), func() error {
				st, e := svc.Load(r.Context(), o, r.PathValue("id"))
				if e != nil {
					return e
				}
				if st.Revision != in.Expected {
					return screenplay.ErrRevisionConflict
				}
				project, e := store.GetProject(r.Context(), o, st.ProjectID)
				if e != nil {
					return e
				}
				system := prompt.ScreenplayAuthoringAssistant
				user := in.Message
				jsonResponse := true
				var chapterIndex = -1
				contextData := map[string]any{"project": project, "chapters": st.Chapters, "conversation": st.Messages}
				if doc, er := store.ActiveDocument(r.Context(), o, st.ProjectID); er == nil {
					contextData["document"] = doc
					shots, er := store.ListShots(r.Context(), o, st.ProjectID, doc.ID)
					if er != nil {
						return er
					}
					contextData["shots"] = shots
				}
				assets, er := store.ListAssets(r.Context(), o, st.ProjectID, "")
				if er != nil {
					return er
				}
				contextData["assets"] = assets
				switch path {
				case "chapters/plan":
					system = prompt.ScreenplayAuthoringPlan
					user = in.Instruction
				case "chapters/{chapter}/script":
					system = prompt.ScreenplayAuthoringChapter
					user = in.Instruction
					jsonResponse = false
					for i, c := range st.Chapters {
						if c.ID == r.PathValue("chapter") {
							chapterIndex = i
							contextData["selected_chapter"] = c
						}
					}
					if chapterIndex < 0 {
						return screenplay.ErrNotFound
					}
				default:
					if strings.TrimSpace(user) == "" {
						return screenplay.ErrEmptyContent
					}
				}
				b, _ := json.Marshal(contextData)
				request := screenplay.TextRequest{Model: project.ModelID, SystemPrompt: system, UserPrompt: string(b) + "\n用户要求：" + user, JSONResponse: jsonResponse, MaxOutputTokens: 16000, Temperature: 0.4}
				if chapterIndex >= 0 {
					request, e = screenplay.BuildChapterTextRequest(project, st.Chapters[chapterIndex], user, string(b))
					if e != nil {
						return e
					}
				}
				result, e := (screenplayTextGenerator{owner: o, resolver: resolver}).Generate(r.Context(), request)
				if e != nil {
					return e
				}
				if e = r.Context().Err(); e != nil {
					return e
				}
				raw := strings.TrimSpace(result.Content)
				if raw == "" {
					return screenplay.ErrAIEmptyContent
				}
				raw = strings.TrimPrefix(raw, "```json")
				raw = strings.TrimSuffix(raw, "```")
				switch path {
				case "chapters/plan":
					var parsed struct {
						Chapters []screenplay.Chapter `json:"chapters"`
					}
					if json.Unmarshal([]byte(raw), &parsed) != nil || len(parsed.Chapters) == 0 {
						return screenplay.ErrAIInvalidJSON
					}
					st.Chapters = parsed.Chapters
				case "chapters/{chapter}/script":
					if e = screenplay.ApplyChapterGeneration(&st.Chapters[chapterIndex], result.Content); e != nil {
						return e
					}
					if _, e = screenplay.PublishChapterCollection(r.Context(), store, o, st.ProjectID, st.Chapters); e != nil {
						return e
					}
				default:
					var parsed struct {
						Reply   string                       `json:"reply"`
						Summary string                       `json:"summary"`
						Actions []screenplay.AuthoringAction `json:"actions"`
					}
					if json.Unmarshal([]byte(raw), &parsed) != nil || strings.TrimSpace(parsed.Reply) == "" {
						return screenplay.ErrAIInvalidJSON
					}
					if len(parsed.Actions) > 0 {
						if len(parsed.Actions) > 1 {
							return screenplay.ErrInvalidArgument
						}
						if e = screenplay.ValidateAuthoringActions(parsed.Actions); e != nil {
							return e
						}
						st.Proposals = append(st.Proposals, screenplay.AssistantProposal{ID: newAgentID("proposal"), Summary: parsed.Summary, Actions: parsed.Actions, Status: "pending", Revision: st.Revision + 1})
					}
					st.Messages = append(st.Messages, screenplay.AssistantMessage{Role: "user", Content: user}, screenplay.AssistantMessage{Role: "assistant", Content: parsed.Reply})
				}
				out, e = svc.Save(r.Context(), o, st.ProjectID, st, in.Expected)
				if e == nil && path == "chapters/{chapter}/script" {
					e = screenplay.InvalidateAuthoringContent(r.Context(), store, o, st.ProjectID, "chapter")
				}
				return e
			})
			respond(w, out, e, a)
		})
	}
	handle("POST "+base+"/storyboards/{engine}/{episode}/shots/batch", "edit", func(w http.ResponseWriter, r *http.Request, o, a string) {
		var in screenplay.StoryboardBatchEdit
		if !decode(w, r, &in) {
			return
		}
		episode, e := strconv.Atoi(r.PathValue("episode"))
		if e != nil {
			authoringError(w, screenplay.ErrInvalidArgument)
			return
		}
		out, e := screenplay.EditStoryboardBatch(r.Context(), store, o, r.PathValue("id"), episode, screenplay.StoryboardEngine(r.PathValue("engine")), in)
		if e != nil {
			authoringError(w, e)
			return
		}
		if e = screenplay.InvalidateAuthoringContent(r.Context(), store, o, r.PathValue("id"), "storyboard"); e != nil {
			authoringError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"storyboard": out})
	})
}
func authoringError(w http.ResponseWriter, e error) {
	status := 500
	switch {
	case errors.Is(e, screenplay.ErrAccessDenied):
		status = 403
	case errors.Is(e, screenplay.ErrNotFound):
		status = 404
	case errors.Is(e, screenplay.ErrRevisionConflict), errors.Is(e, screenplay.ErrLockBusy):
		status = 409
	case errors.Is(e, screenplay.ErrInvalidArgument), errors.Is(e, screenplay.ErrEmptyContent):
		status = 400
	case errors.Is(e, screenplay.ErrAPIKeyMissing), errors.Is(e, screenplay.ErrInvalidState):
		status = 422
	}
	writeJSON(w, status, map[string]string{"error": e.Error()})
}
func validateAuthoringDisplay(d map[string]any) error {
	for k, v := range d {
		switch k {
		case "editor_version":
			if v != "classic" && v != "studio" {
				return screenplay.ErrInvalidArgument
			}
		case "active_tab":
			ok := false
			for _, s := range []string{"source", "chapters", "assistant", "assets", "storyboard", "media", "production"} {
				if v == s {
					ok = true
				}
			}
			if !ok {
				return screenplay.ErrInvalidArgument
			}
		case "sidebar_collapsed", "show_descriptions":
			if _, ok := v.(bool); !ok {
				return screenplay.ErrInvalidArgument
			}
		case "font_size":
			n, ok := v.(float64)
			if !ok || n < 12 || n > 22 {
				return screenplay.ErrInvalidArgument
			}
		default:
			return screenplay.ErrInvalidArgument
		}
	}
	return nil
}
