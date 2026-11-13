package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"regexp"
	"strconv"
	"strings"
)

var creativeImagePath = regexp.MustCompile(`^/api/(reference-assets/file/(character|scene|prop)/[A-Za-z0-9_-]+/[A-Za-z0-9_-]+|agent/tasks/[A-Za-z0-9_-]+/assets/[A-Za-z0-9_-]+/content|ai-cut/projects/[A-Za-z0-9_-]+/assets/[A-Za-z0-9_-]+/content)$`)

func decodeCreativeRequest(w http.ResponseWriter, r *http.Request, dst any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	if e := d.Decode(dst); e != nil {
		writeJSON(w, 400, map[string]string{"error": "请求内容无效"})
		return false
	}
	if e := d.Decode(new(any)); e != io.EOF {
		writeJSON(w, 400, map[string]string{"error": "请求内容无效"})
		return false
	}
	return true
}
func validateCreativeStyle(v screenplay.UserStyle) error {
	if strings.TrimSpace(v.Name) == "" || len([]rune(v.Name)) > 255 {
		return fmt.Errorf("画风名称不能为空，最多255字")
	}
	if strings.TrimSpace(v.Prompt) == "" || len([]rune(v.Prompt)) > 100000 {
		return fmt.Errorf("画风文字不能为空，最多10万字")
	}
	if len(v.Images) > 16 {
		return fmt.Errorf("每个画风最多16张预览图")
	}
	for _, ref := range v.Images {
		if !creativeImagePath.MatchString(ref) {
			return fmt.Errorf("预览图必须引用本系统已上传的图片")
		}
	}
	return nil
}

// RegisterScreenplayCreativeRoutes keeps user libraries private; shared project readers see only frozen selections.
func RegisterScreenplayCreativeRoutes(mux *http.ServeMux, auth BridgeAuth, store screenplay.Store, authors *screenplay.AuthoringService) {
	cs, ok := store.(screenplay.CreativeStore)
	if mux == nil || !ok {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			actor := strings.TrimSpace(identity.Username)
			if authors != nil && pattern == "PUT /api/screenplays/{id}/{creativeAction}" && r.PathValue("creativeAction") == "creative-config" {
				serveCreativeConfigMutation(w, r, actor, authors, fn)
				return
			}
			fn(w, r, actor)
		})))
	}
	handle("GET /api/screenplays/creative-library", func(w http.ResponseWriter, r *http.Request, actor string) {
		styles, e := cs.ListUserStyles(r.Context(), actor)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		options := []map[string]string{}
		for _, p := range prompt.CreativePromptOptions() {
			options = append(options, map[string]string{"id": p.ID, "name": p.Name, "purpose": strings.ReplaceAll(p.Type, "-", "_"), "version": p.Version, "sha256": p.Digest})
		}
		writeJSON(w, 200, map[string]any{"styles": styles, "prompts": options})
	})
	type styleInput struct {
		Name             string   `json:"name"`
		Prompt           string   `json:"prompt"`
		Images           []string `json:"images"`
		ExpectedRevision *int64   `json:"expected_revision"`
	}
	handle("POST /api/screenplays/styles", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in styleInput
		if !decodeCreativeRequest(w, r, &in) {
			return
		}
		v := screenplay.UserStyle{ID: newAgentID("style"), OwnerUsername: actor, Name: strings.TrimSpace(in.Name), Prompt: in.Prompt, Images: in.Images}
		if e := validateCreativeStyle(v); e != nil {
			writeJSON(w, 400, map[string]string{"error": e.Error()})
			return
		}
		saved, e := cs.SaveUserStyle(r.Context(), v, nil)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 201, map[string]any{"style": saved})
	})
	handle("PUT /api/screenplays/styles/{styleId}", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in styleInput
		if !decodeCreativeRequest(w, r, &in) {
			return
		}
		if in.ExpectedRevision == nil {
			writeJSON(w, 400, map[string]string{"error": "缺少画风版本，请刷新重试"})
			return
		}
		if _, e := cs.GetUserStyle(r.Context(), actor, r.PathValue("styleId")); e != nil {
			writeScreenplayError(w, e)
			return
		}
		v := screenplay.UserStyle{ID: r.PathValue("styleId"), OwnerUsername: actor, Name: strings.TrimSpace(in.Name), Prompt: in.Prompt, Images: in.Images}
		if e := validateCreativeStyle(v); e != nil {
			writeJSON(w, 400, map[string]string{"error": e.Error()})
			return
		}
		saved, e := cs.SaveUserStyle(r.Context(), v, in.ExpectedRevision)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"style": saved})
	})
	handle("DELETE /api/screenplays/styles/{styleId}", func(w http.ResponseWriter, r *http.Request, actor string) {
		revision, e := strconv.ParseInt(r.URL.Query().Get("expected_revision"), 10, 64)
		if e != nil {
			writeJSON(w, 400, map[string]string{"error": "缺少画风版本，请刷新重试"})
			return
		}
		if _, e = cs.GetUserStyle(r.Context(), actor, r.PathValue("styleId")); e == nil {
			e = cs.DeleteUserStyle(r.Context(), actor, r.PathValue("styleId"), revision)
		}
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]bool{"deleted": true})
	})
	handle("GET /api/screenplays/{id}/creative-config", func(w http.ResponseWriter, r *http.Request, actor string) {
		owner := actor
		if authors != nil {
			var e error
			owner, e = authors.Resolve(r.Context(), actor, r.PathValue("id"), "read")
			if e != nil {
				authoringError(w, e)
				return
			}
		}
		p, e := store.GetProject(r.Context(), owner, r.PathValue("id"))
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		c := p.CreativeConfig
		if c == nil {
			c = &screenplay.CreativeConfig{ScriptPromptID: p.PromptID}
			if p.SelectedStyle != "" || p.StyleDesc != "" {
				c.Style = &screenplay.CreativeStyleSnapshot{ID: p.SelectedStyle, Name: p.ArtStyle, Prompt: p.StyleDesc, Images: []string{}}
			}
		}
		writeJSON(w, 200, map[string]any{"creative_config": c, "revision": p.Revision})
	})
	// A final wildcard avoids ServeMux's overlapping /styles/{styleId} pattern.
	// Only the documented creative-config action is accepted here.
	handle("PUT /api/screenplays/{id}/{creativeAction}", func(w http.ResponseWriter, r *http.Request, actor string) {
		if r.PathValue("creativeAction") != "creative-config" {
			http.NotFound(w, r)
			return
		}
		if authors != nil {
			owner, e := authors.Resolve(r.Context(), actor, r.PathValue("id"), "read")
			if e != nil {
				authoringError(w, e)
				return
			}
			if owner != actor {
				authoringError(w, screenplay.ErrAccessDenied)
				return
			}
		}
		var in creativeConfigPatch
		if !decodeCreativeRequest(w, r, &in) {
			return
		}
		if in.ExpectedRevision == nil {
			writeJSON(w, 400, map[string]string{"error": "缺少作品版本，请刷新重试"})
			return
		}
		p, e := store.GetProject(r.Context(), actor, r.PathValue("id"))
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		if p.Revision != *in.ExpectedRevision {
			writeScreenplayError(w, screenplay.ErrRevisionConflict)
			return
		}
		c := screenplay.CreativeConfig{}
		if p.CreativeConfig != nil {
			c = *p.CreativeConfig
		} else {
			c.ScriptPromptID = p.PromptID
			if p.SelectedStyle != "" || p.StyleDesc != "" {
				c.Style = &screenplay.CreativeStyleSnapshot{ID: p.SelectedStyle, Name: p.ArtStyle, Prompt: p.StyleDesc, Images: []string{}}
			}
		}
		if c.PromptSnapshots == nil {
			c.PromptSnapshots = map[string]screenplay.CreativePromptSnapshot{}
		}
		if e = applyCreativePatch(r, cs, actor, &c, in); e != nil {
			writeJSON(w, 400, map[string]string{"error": e.Error()})
			return
		}
		saved, e := cs.SaveCreativeConfig(r.Context(), actor, p.ID, c, *in.ExpectedRevision)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"creative_config": saved.CreativeConfig, "revision": saved.Revision})
	})
}

func serveCreativeConfigMutation(w http.ResponseWriter, r *http.Request, actor string, authors *screenplay.AuthoringService, fn func(http.ResponseWriter, *http.Request, string)) {
	// Use the same project lock as chapter editing/generation, and recheck access after locking.
	owner, e := authors.Resolve(r.Context(), actor, r.PathValue("id"), "read")
	if e != nil {
		authoringError(w, e)
		return
	}
	if owner != actor {
		authoringError(w, screenplay.ErrAccessDenied)
		return
	}
	response := &screenplayWorkflowResponse{headers: make(http.Header)}
	e = authors.WithLock(r.Context(), owner, r.PathValue("id"), func() error {
		resolved, err := authors.Resolve(r.Context(), actor, r.PathValue("id"), "read")
		if err != nil {
			return err
		}
		if resolved != actor {
			return screenplay.ErrAccessDenied
		}
		fn(response, r, actor)
		if response.status >= 200 && response.status < 300 {
			if err = screenplay.InvalidateAuthoringContent(r.Context(), authors.Store, owner, r.PathValue("id"), "config"); err != nil {
				return fmt.Errorf("创意配置已保存，但后续任务状态更新失败：%w", err)
			}
			st, err := authors.Load(r.Context(), owner, r.PathValue("id"))
			if err != nil {
				return fmt.Errorf("创意配置已保存，但工作区读取失败：%w", err)
			}
			if _, err = authors.Save(r.Context(), owner, r.PathValue("id"), st, st.Revision); err != nil {
				return fmt.Errorf("创意配置已保存，但工作区状态保存失败：%w", err)
			}
		}
		return nil
	})
	if e != nil {
		authoringError(w, e)
		return
	}
	for k, vs := range response.headers {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	status := response.status
	if status == 0 {
		status = 200
	}
	w.WriteHeader(status)
	_, _ = w.Write(response.body.Bytes())
}

type creativeConfigPatch struct {
	ExpectedRevision        *int64          `json:"expected_revision"`
	ScriptPromptID          *string         `json:"script_prompt_id"`
	StoryboardPromptID      *string         `json:"storyboard_prompt_id"`
	CharacterDesignPromptID *string         `json:"character_design_prompt_id"`
	SceneDesignPromptID     *string         `json:"scene_design_prompt_id"`
	StyleID                 json.RawMessage `json:"style_id"`
	ImageCoreInstruction    *string         `json:"image_core_instruction"`
	ImageModelID            *string         `json:"image_model_id"`
	ImageEditModelID        *string         `json:"image_edit_model_id"`
}

func applyCreativePatch(r *http.Request, cs screenplay.CreativeStore, actor string, c *screenplay.CreativeConfig, in creativeConfigPatch) error {
	selections := []struct {
		src     *string
		dst     *string
		purpose string
	}{{in.ScriptPromptID, &c.ScriptPromptID, "script"}, {in.StoryboardPromptID, &c.StoryboardPromptID, "storyboard"}, {in.CharacterDesignPromptID, &c.CharacterDesignPromptID, "character-design"}, {in.SceneDesignPromptID, &c.SceneDesignPromptID, "scene-design"}}
	for _, s := range selections {
		id := strings.TrimSpace(*s.dst)
		if s.src != nil {
			id = strings.TrimSpace(*s.src)
		}
		if id == "" && s.src == nil {
			continue
		}
		if frozen, ok := c.PromptSnapshots[s.purpose]; ok && frozen.ID == id && frozen.Type == s.purpose && frozen.Body != "" {
			*s.dst = id
			continue
		}
		p, e := prompt.ResolveCreativePrompt(id, s.purpose)
		if e != nil {
			return e
		}
		*s.dst = id
		c.PromptSnapshots[s.purpose] = p
	}
	if len(in.StyleID) > 0 {
		var id *string
		if e := json.Unmarshal(in.StyleID, &id); e != nil {
			return fmt.Errorf("画风ID无效")
		}
		if id == nil || strings.TrimSpace(*id) == "" {
			c.Style = nil
		} else {
			v, e := cs.GetUserStyle(r.Context(), actor, strings.TrimSpace(*id))
			if e != nil {
				return fmt.Errorf("所选画风不存在或不属于当前账号")
			}
			c.Style = &screenplay.CreativeStyleSnapshot{ID: v.ID, Name: v.Name, Prompt: v.Prompt, Images: v.Images, Revision: v.Revision}
		}
	}
	if in.ImageCoreInstruction != nil {
		if len([]rune(*in.ImageCoreInstruction)) > 100000 {
			return fmt.Errorf("图片核心要求最多10万字")
		}
		c.ImageCoreInstruction = *in.ImageCoreInstruction
	}
	if in.ImageModelID != nil {
		c.ImageModelID = strings.TrimSpace(*in.ImageModelID)
	}
	if in.ImageEditModelID != nil {
		c.ImageEditModelID = strings.TrimSpace(*in.ImageEditModelID)
	}
	return nil
}
