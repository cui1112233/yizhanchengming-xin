package httpapi

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

const screenplayJSONLimit = 8 << 20

type screenplayTextGenerator struct {
	owner    string
	resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)
}

func (g screenplayTextGenerator) Generate(ctx context.Context, req screenplay.TextRequest) (screenplay.TextResponse, error) {
	if g.resolver == nil {
		return screenplay.TextResponse{}, screenplay.ErrAPIKeyMissing
	}
	provider, err := g.resolver(ctx, g.owner, strings.TrimSpace(req.Model))
	if err != nil {
		return screenplay.TextResponse{}, err
	}
	if provider == nil {
		return screenplay.TextResponse{}, screenplay.ErrAPIKeyMissing
	}
	content, err := provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt: req.SystemPrompt, UserPrompt: req.UserPrompt,
		Temperature: req.Temperature, MaxTokens: req.MaxOutputTokens,
		DisableJSONResponse: !req.JSONResponse,
	})
	return screenplay.TextResponse{Content: content, Model: req.Model}, err
}

func RegisterScreenplayRoutes(mux *http.ServeMux, auth BridgeAuth, store screenplay.Store, resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error), now func() time.Time, mediaRunners ...screenplay.BatchMediaRunner) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	var mediaRunner screenplay.BatchMediaRunner
	if len(mediaRunners) > 0 {
		mediaRunner = mediaRunners[0]
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	decode := func(w http.ResponseWriter, req *http.Request, target any) bool {
		decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, screenplayJSONLimit))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return false
		}
		return true
	}

	handle("GET /api/comic-screenplays/prompts", func(w http.ResponseWriter, req *http.Request, owner string) {
		writeJSON(w, http.StatusOK, map[string]any{"prompts": []map[string]any{
			{"id": "b_director_screenplay", "name": "专业剧本", "mode": "b", "description": "导演级剧本结构，适合完整小说改编"},
			{"id": "a1_novel_to_screenplay", "name": "小说转剧本", "mode": "a1", "description": "按原文章节快速生成分集剧本"},
		}})
	})

	handle("GET /api/comic-screenplays/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		projects, err := store.ListProjects(req.Context(), owner)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
	})
	handle("POST /api/comic-screenplays/projects", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ID             string `json:"id"`
			Title          string `json:"title"`
			ScriptMode     string `json:"scriptMode"`
			AspectRatio    string `json:"aspectRatio"`
			ArtStyle       string `json:"artStyle"`
			StyleDesc      string `json:"styleDesc"`
			SelectedStyle  string `json:"selectedStyle"`
			Synopsis       string `json:"synopsis"`
			ModelID        string `json:"modelId"`
			PromptID       string `json:"promptId"`
			AssetDegree    string `json:"assetDegree"`
			PermissionMode string `json:"permissionMode"`
		}
		if !decode(w, req, &input) {
			return
		}
		if strings.TrimSpace(input.Title) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "作品名称不能为空"})
			return
		}
		id := strings.TrimSpace(input.ID)
		if id == "" {
			id = newAgentID("comic")
		}
		project, err := store.CreateProject(req.Context(), screenplay.Project{ID: id, OwnerUsername: owner, Title: input.Title, ScriptMode: input.ScriptMode, AspectRatio: input.AspectRatio, ArtStyle: input.ArtStyle, StyleDesc: input.StyleDesc, SelectedStyle: input.SelectedStyle, Synopsis: input.Synopsis, ModelID: input.ModelID, PromptID: input.PromptID, AssetDegree: screenplay.AssetDegree(input.AssetDegree), PermissionMode: screenplay.ParsePermissionMode(input.PermissionMode), CreatedAt: now().UTC()})
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"project": project})
	})
	handle("GET /api/comic-screenplays/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		project, err := store.GetProject(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": project})
	})
	handle("PATCH /api/comic-screenplays/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Title            string `json:"title"`
			ScriptMode       string `json:"scriptMode"`
			AspectRatio      string `json:"aspectRatio"`
			ArtStyle         string `json:"artStyle"`
			StyleDesc        string `json:"styleDesc"`
			SelectedStyle    string `json:"selectedStyle"`
			Synopsis         string `json:"synopsis"`
			ModelID          string `json:"modelId"`
			PromptID         string `json:"promptId"`
			AssetDegree      string `json:"assetDegree"`
			PermissionMode   string `json:"permissionMode"`
			ExpectedRevision int64  `json:"expectedRevision"`
		}
		if !decode(w, req, &input) {
			return
		}
		project, err := store.UpdateProject(req.Context(), owner, req.PathValue("id"), screenplay.Project{Title: input.Title, ScriptMode: input.ScriptMode, AspectRatio: input.AspectRatio, ArtStyle: input.ArtStyle, StyleDesc: input.StyleDesc, SelectedStyle: input.SelectedStyle, Synopsis: input.Synopsis, ModelID: input.ModelID, PromptID: input.PromptID, AssetDegree: screenplay.AssetDegree(input.AssetDegree), PermissionMode: screenplay.ParsePermissionMode(input.PermissionMode)}, input.ExpectedRevision)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": project})
	})
	handle("PUT /api/comic-screenplays/projects/{id}/source", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Content string `json:"content"`
		}
		if !decode(w, req, &input) {
			return
		}
		if len([]rune(input.Content)) > screenplay.MaxSingleCallRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "原文超过 60 万字，请按章拆分后分批提交"})
			return
		}
		project, err := store.SaveProjectSource(req.Context(), owner, req.PathValue("id"), input.Content)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"project": project})
	})
	handle("PUT /api/comic-screenplays/projects/{id}/document", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Content string `json:"content"`
		}
		if !decode(w, req, &input) {
			return
		}
		if len([]rune(input.Content)) > screenplay.MaxSingleCallRunes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "剧本超过 60 万字，请拆分后保存"})
			return
		}
		document, err := screenplay.SaveManualDocument(req.Context(), store, owner, req.PathValue("id"), input.Content)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"document": document})
	})
	handle("POST /api/comic-screenplays/projects/{id}/documents/{documentID}/restore", func(w http.ResponseWriter, req *http.Request, owner string) {
		document, err := screenplay.RestoreDocument(req.Context(), store, owner, req.PathValue("id"), req.PathValue("documentID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"document": document})
	})
	handle("DELETE /api/comic-screenplays/projects/{id}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteProject(req.Context(), owner, req.PathValue("id")); err != nil {
			writeScreenplayError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("GET /api/comic-screenplays/projects/{id}/workspace", func(w http.ResponseWriter, req *http.Request, owner string) {
		project, err := store.GetProject(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		payload := map[string]any{"project": project, "documents": []screenplay.Document{}, "shots": []screenplay.Shot{}, "characters": []screenplay.Character{}, "scenes": []screenplay.Scene{}, "assets": []screenplay.Asset{}, "storyboards": []screenplay.Storyboard{}}
		documents, _ := store.ListDocuments(req.Context(), owner, project.ID)
		payload["documents"] = documents
		if active, activeErr := store.ActiveDocument(req.Context(), owner, project.ID); activeErr == nil {
			payload["document"] = active
			shots, _ := store.ListShots(req.Context(), owner, project.ID, active.ID)
			payload["shots"] = shots
			characters, _ := store.ListCharacters(req.Context(), owner, project.ID, active.ID)
			payload["characters"] = characters
			scenes, _ := store.ListScenes(req.Context(), owner, project.ID, active.ID)
			payload["scenes"] = scenes
		}
		assets, _ := store.ListAssets(req.Context(), owner, project.ID, "")
		payload["assets"] = assets
		storyboardsV7, _ := store.ListStoryboards(req.Context(), owner, project.ID, screenplay.EngineXiaoluoV7)
		storyboardsV3, _ := store.ListStoryboards(req.Context(), owner, project.ID, screenplay.EngineXiaoyunqueV3)
		payload["storyboards"] = append(storyboardsV7, storyboardsV3...)
		writeJSON(w, http.StatusOK, payload)
	})
	type shotInput struct {
		EpisodeNo         int      `json:"episodeNo"`
		Heading           string   `json:"heading"`
		ScriptText        string   `json:"scriptText"`
		Characters        []string `json:"characters"`
		Scene             string   `json:"scene"`
		Props             []string `json:"props"`
		LocationReference string   `json:"locationReference"`
	}
	handle("POST /api/comic-screenplays/projects/{id}/shots", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := req.PathValue("id")
		document, err := store.ActiveDocument(req.Context(), owner, projectID)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		var input shotInput
		if !decode(w, req, &input) {
			return
		}
		shot, err := store.CreateShot(req.Context(), screenplay.Shot{DocumentID: document.ID, ProjectID: projectID, OwnerUsername: owner, EpisodeNo: input.EpisodeNo, Heading: input.Heading, ScriptText: input.ScriptText, CharacterNames: input.Characters, SceneName: input.Scene, PropNames: input.Props, LocationReference: input.LocationReference})
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"shot": shot})
	})
	handle("PATCH /api/comic-screenplays/projects/{id}/shots/{shotID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input shotInput
		if !decode(w, req, &input) {
			return
		}
		shot, err := store.UpdateShot(req.Context(), owner, req.PathValue("id"), req.PathValue("shotID"), screenplay.Shot{EpisodeNo: input.EpisodeNo, Heading: input.Heading, ScriptText: input.ScriptText, CharacterNames: input.Characters, SceneName: input.Scene, PropNames: input.Props, LocationReference: input.LocationReference})
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"shot": shot})
	})
	handle("POST /api/comic-screenplays/projects/{id}/shots/{shotID}/move", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Direction string `json:"direction"`
		}
		if !decode(w, req, &input) {
			return
		}
		direction := 0
		switch strings.ToLower(strings.TrimSpace(input.Direction)) {
		case "up":
			direction = -1
		case "down":
			direction = 1
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "移动方向必须是 up 或 down"})
			return
		}
		shots, err := store.MoveShot(req.Context(), owner, req.PathValue("id"), req.PathValue("shotID"), direction)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"shots": shots})
	})
	handle("DELETE /api/comic-screenplays/projects/{id}/shots/{shotID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteShot(req.Context(), owner, req.PathValue("id"), req.PathValue("shotID")); err != nil {
			writeScreenplayError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("POST /api/comic-screenplays/projects/{id}/generate", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Mode          string            `json:"mode"`
			Content       string            `json:"content"`
			Novel         string            `json:"novel"`
			Model         string            `json:"model"`
			PromptID      string            `json:"promptId"`
			ChapterRange  string            `json:"chapterRange"`
			EpisodeTarget string            `json:"episodeTarget"`
			ExtraNote     string            `json:"extraNote"`
			Segments      []string          `json:"segments"`
			ExtraVars     map[string]string `json:"extraVars"`
			Confirmed     bool              `json:"confirmed"`
		}
		if !decode(w, req, &input) {
			return
		}
		service := screenplay.ScreenplayService{Store: store, Generator: screenplayTextGenerator{owner: owner, resolver: resolver}}
		var result screenplay.ScreenplayResult
		var err error
		if strings.EqualFold(input.Mode, "a1") {
			result, err = service.GenerateA1(req.Context(), screenplay.A1Request{Owner: owner, ProjectID: req.PathValue("id"), Content: input.Content, Segments: input.Segments, Model: input.Model, PromptID: input.PromptID, ExtraVars: input.ExtraVars, Confirmed: input.Confirmed})
		} else {
			novel := input.Novel
			if strings.TrimSpace(novel) == "" {
				novel = input.Content
			}
			result, err = service.GenerateB(req.Context(), screenplay.BRequest{Owner: owner, ProjectID: req.PathValue("id"), Novel: novel, Model: input.Model, PromptID: input.PromptID, ChapterRange: input.ChapterRange, EpisodeTarget: input.EpisodeTarget, ExtraNote: input.ExtraNote, ExtraVars: input.ExtraVars, Confirmed: input.Confirmed})
		}
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	handle("POST /api/comic-screenplays/projects/{id}/source/parse", func(w http.ResponseWriter, req *http.Request, owner string) {
		if _, err := store.GetProject(req.Context(), owner, req.PathValue("id")); err != nil {
			writeScreenplayError(w, err)
			return
		}
		if !limitRequestBody(w, req, maxScreenplayUploadBytes) {
			return
		}
		if err := req.ParseMultipartForm(24 << 20); err != nil {
			if isBodyTooLarge(err) {
				writePayloadTooLarge(w, maxScreenplayUploadBytes)
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文件过大或上传格式无效"})
			return
		}
		file, header, err := req.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择 TXT 或 DOCX 文件"})
			return
		}
		defer file.Close()
		data, err := io.ReadAll(io.LimitReader(file, 20<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文件读取失败"})
			return
		}
		ext := strings.ToLower(filepath.Ext(header.Filename))
		var content string
		switch ext {
		case ".txt":
			content = strings.TrimPrefix(string(data), "\ufeff")
		case ".docx":
			content, err = readDOCXText(data)
		default:
			err = errors.New("仅支持 TXT 或 DOCX 文件")
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		content = strings.TrimSpace(content)
		if content == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "文件中没有可读取的正文"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"filename": header.Filename, "content": content, "runes": len([]rune(content))})
	})
	handle("POST /api/comic-screenplays/projects/{id}/assets/extract", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Type        screenplay.AssetType   `json:"type"`
			Degree      screenplay.AssetDegree `json:"degree"`
			Content     string                 `json:"content"`
			Chapters    []string               `json:"chapters"`
			Model       string                 `json:"model"`
			Concurrency int                    `json:"concurrency"`
			Confirmed   bool                   `json:"confirmed"`
		}
		if !decode(w, req, &input) {
			return
		}
		service := screenplay.AssetService{Store: store, Generator: screenplayTextGenerator{owner: owner, resolver: resolver}}
		result, err := service.Extract(req.Context(), screenplay.AssetRequest{Owner: owner, ProjectID: req.PathValue("id"), Type: input.Type, Degree: input.Degree, Content: input.Content, Chapters: input.Chapters, Model: input.Model, Concurrency: input.Concurrency, Confirmed: input.Confirmed})
		if err != nil {
			// 部分批次失败：已成功提取的资产已经落库，把 partial_failure、
			// failed_batches 和已保存资产一起返回，前端据此明确提示失败批次，
			// 不能再整体当成成功，也不能只回一条 500 丢掉结果。
			if errors.Is(err, screenplay.ErrPartialFailure) {
				writeJSON(w, http.StatusOK, result)
				return
			}
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	handle("GET /api/comic-screenplays/projects/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		assets, err := store.ListAssets(req.Context(), owner, req.PathValue("id"), screenplay.AssetType(req.URL.Query().Get("type")))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"assets": assets})
	})
	type assetInput struct {
		Type             screenplay.AssetType   `json:"type"`
		Degree           screenplay.AssetDegree `json:"degree"`
		Name             string                 `json:"name"`
		Identity         string                 `json:"identity"`
		Description      string                 `json:"description"`
		Prompt           string                 `json:"prompt"`
		Voice            string                 `json:"voice"`
		Episodes         string                 `json:"episodes"`
		Tag              string                 `json:"tag"`
		Extra            map[string]string      `json:"extra"`
		ExpectedRevision int64                  `json:"expectedRevision"`
	}
	toAsset := func(projectID, owner string, input assetInput) screenplay.Asset {
		return screenplay.Asset{ProjectID: projectID, OwnerUsername: owner, Type: input.Type, Degree: input.Degree, Name: input.Name, Identity: input.Identity, Description: input.Description, Prompt: input.Prompt, Voice: input.Voice, Episodes: input.Episodes, Tag: input.Tag, Extra: input.Extra}
	}
	handle("POST /api/comic-screenplays/projects/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input assetInput
		if !decode(w, req, &input) {
			return
		}
		asset, err := store.CreateAsset(req.Context(), toAsset(req.PathValue("id"), owner, input))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
	})
	handle("PATCH /api/comic-screenplays/projects/{id}/assets/{assetID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input assetInput
		if !decode(w, req, &input) {
			return
		}
		asset, err := store.UpdateAsset(req.Context(), owner, req.PathValue("id"), req.PathValue("assetID"), toAsset(req.PathValue("id"), owner, input), input.ExpectedRevision)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"asset": asset})
	})
	handle("DELETE /api/comic-screenplays/projects/{id}/assets/{assetID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		if err := store.DeleteAsset(req.Context(), owner, req.PathValue("id"), req.PathValue("assetID")); err != nil {
			writeScreenplayError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handle("POST /api/comic-screenplays/projects/{id}/storyboards/generate", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			Engine    screenplay.StoryboardEngine `json:"engine"`
			EpisodeNo int                         `json:"episodeNo"`
			Content   string                      `json:"content"`
			Episodes  map[int]string              `json:"episodes"`
			Style     string                      `json:"style"`
			Synopsis  string                      `json:"synopsis"`
			Model     string                      `json:"model"`
			Confirmed bool                        `json:"confirmed"`
		}
		if !decode(w, req, &input) {
			return
		}
		service := screenplay.StoryboardService{Store: store, Generator: screenplayTextGenerator{owner: owner, resolver: resolver}}
		result, err := service.Generate(req.Context(), screenplay.StoryboardRequest{Owner: owner, ProjectID: req.PathValue("id"), Engine: input.Engine, EpisodeNo: input.EpisodeNo, Content: input.Content, Episodes: input.Episodes, Style: input.Style, Synopsis: input.Synopsis, Model: input.Model, Confirmed: input.Confirmed})
		if err != nil {
			// 部分剧集失败：已生成的剧集已经落库，返回 partial_failure 和
			// failed_episodes，让前端能明确提示「第 N 集失败」。
			if errors.Is(err, screenplay.ErrPartialFailure) {
				writeJSON(w, http.StatusOK, result)
				return
			}
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	handle("GET /api/comic-screenplays/projects/{id}/storyboards", func(w http.ResponseWriter, req *http.Request, owner string) {
		engine := screenplay.StoryboardEngine(req.URL.Query().Get("engine"))
		if engine != "" {
			rows, err := store.ListStoryboards(req.Context(), owner, req.PathValue("id"), engine)
			if err != nil {
				writeScreenplayError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"storyboards": rows})
			return
		}
		first, _ := store.ListStoryboards(req.Context(), owner, req.PathValue("id"), screenplay.EngineXiaoluoV7)
		second, _ := store.ListStoryboards(req.Context(), owner, req.PathValue("id"), screenplay.EngineXiaoyunqueV3)
		writeJSON(w, http.StatusOK, map[string]any{"storyboards": append(first, second...)})
	})
	type storyboardShotInput struct {
		AfterShotID string  `json:"afterShotId"`
		SceneID     int     `json:"sceneId"`
		Duration    float64 `json:"durationSec"`
		Characters  string  `json:"characters"`
		Scene       string  `json:"scene"`
		Props       string  `json:"props"`
		VideoDesc   string  `json:"videoDesc"`
	}
	storyboardPath := func(w http.ResponseWriter, req *http.Request) (string, int, screenplay.StoryboardEngine, bool) {
		episode, err := strconv.Atoi(req.PathValue("episode"))
		if err != nil || episode <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "集数必须是大于 0 的整数"})
			return "", 0, "", false
		}
		engine := screenplay.StoryboardEngine(strings.TrimSpace(req.PathValue("engine")))
		if engine == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "分镜引擎不能为空"})
			return "", 0, "", false
		}
		return req.PathValue("id"), episode, engine, true
	}
	toStoryboardShot := func(input storyboardShotInput) screenplay.StoryboardShot {
		return screenplay.StoryboardShot{SceneID: input.SceneID, Duration: input.Duration, Characters: input.Characters, Scene: input.Scene, Props: input.Props, VideoDesc: input.VideoDesc}
	}
	handle("POST /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, episode, engine, ok := storyboardPath(w, req)
		if !ok {
			return
		}
		var input storyboardShotInput
		if !decode(w, req, &input) {
			return
		}
		storyboard, err := screenplay.CreateStoryboardShot(req.Context(), store, owner, projectID, episode, engine, input.AfterShotID, toStoryboardShot(input))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"storyboard": storyboard})
	})
	handle("PATCH /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots/{shotID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, episode, engine, ok := storyboardPath(w, req)
		if !ok {
			return
		}
		var input storyboardShotInput
		if !decode(w, req, &input) {
			return
		}
		storyboard, err := screenplay.UpdateStoryboardShot(req.Context(), store, owner, projectID, episode, engine, req.PathValue("shotID"), toStoryboardShot(input))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"storyboard": storyboard})
	})
	handle("DELETE /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots/{shotID}", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, episode, engine, ok := storyboardPath(w, req)
		if !ok {
			return
		}
		storyboard, err := screenplay.DeleteStoryboardShot(req.Context(), store, owner, projectID, episode, engine, req.PathValue("shotID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"storyboard": storyboard})
	})
	handle("POST /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots/{shotID}/split", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, episode, engine, ok := storyboardPath(w, req)
		if !ok {
			return
		}
		storyboard, err := screenplay.SplitStoryboardShot(req.Context(), store, owner, projectID, episode, engine, req.PathValue("shotID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"storyboard": storyboard})
	})
	handle("POST /api/comic-screenplays/projects/{id}/storyboards/{engine}/{episode}/shots/{shotID}/merge-next", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, episode, engine, ok := storyboardPath(w, req)
		if !ok {
			return
		}
		storyboard, err := screenplay.MergeStoryboardShotWithNext(req.Context(), store, owner, projectID, episode, engine, req.PathValue("shotID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"storyboard": storyboard})
	})
	handle("POST /api/comic-screenplays/projects/{id}/batch", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := req.PathValue("id")
		project, err := store.GetProject(req.Context(), owner, projectID)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		if existing, getErr := store.GetBatch(req.Context(), owner, projectID); getErr == nil {
			if screenplay.NormalizeBatchStatus(existing.Status) == screenplay.BatchCompleted {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前批次已经完成，请新建项目后再次执行"})
				return
			}
			updated, mutateErr := store.MutateBatch(req.Context(), owner, projectID, func(batch *screenplay.Batch) error {
				if screenplay.NormalizeBatchStatus(batch.Status) == screenplay.BatchPaused || screenplay.NormalizeBatchStatus(batch.Status) == screenplay.BatchFailed {
					return screenplay.SetBatchStatus(batch, screenplay.BatchRunning)
				}
				return nil
			})
			if mutateErr != nil {
				writeScreenplayError(w, mutateErr)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"batch": updated, "steps": screenplay.StepDefinitions})
			return
		}
		batch := screenplay.NewBatch(newAgentID("comic-batch"), projectID, owner, newAgentID("session"), project.PermissionMode)
		if err := store.CreateBatch(req.Context(), batch); err != nil {
			writeScreenplayError(w, err)
			return
		}
		_ = store.AppendBatchLog(req.Context(), batch.ID, owner, "info", "漫剧生产任务已创建")
		writeJSON(w, http.StatusCreated, map[string]any{"batch": batch, "steps": screenplay.StepDefinitions})
	})
	handle("GET /api/comic-screenplays/projects/{id}/batch", func(w http.ResponseWriter, req *http.Request, owner string) {
		batch, err := store.GetBatch(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		logs, _ := store.ListBatchLogs(req.Context(), batch.ID, owner, 100)
		writeJSON(w, http.StatusOK, map[string]any{"batch": batch, "steps": screenplay.StepDefinitions, "logs": logs})
	})
	handle("POST /api/comic-screenplays/projects/{id}/batch/run-next", func(w http.ResponseWriter, req *http.Request, owner string) {
		executor := screenplay.BatchExecutor{
			Store:     store,
			Generator: screenplayTextGenerator{owner: owner, resolver: resolver},
			Media:     mediaRunner,
		}
		result, err := executor.RunNext(req.Context(), owner, req.PathValue("id"))
		if err != nil {
			writeJSON(w, screenplayErrorStatus(err), map[string]any{"error": err.Error(), "batch": result.Batch, "unit": result.Unit})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})
	handle("POST /api/comic-screenplays/projects/{id}/batch/cancel", func(w http.ResponseWriter, req *http.Request, owner string) {
		batch, err := store.MutateBatch(req.Context(), owner, req.PathValue("id"), func(batch *screenplay.Batch) error {
			if screenplay.NormalizeBatchStatus(batch.Status) == screenplay.BatchCompleted {
				return screenplay.ErrInvalidState
			}
			return screenplay.SetBatchStatus(batch, screenplay.BatchPaused)
		})
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		_ = store.AppendBatchLog(req.Context(), batch.ID, owner, "info", "用户暂停了漫剧生产任务")
		writeJSON(w, http.StatusOK, map[string]any{"batch": batch})
	})
	handle("GET /api/comic-screenplays/projects/{id}/export", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := req.PathValue("id")
		project, err := store.GetProject(req.Context(), owner, projectID)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		payload := map[string]any{"format": "yizhanchengming-comic-screenplay-v1", "exported_at": now().UTC(), "project": project}
		if documents, listErr := store.ListDocuments(req.Context(), owner, projectID); listErr == nil {
			payload["documents"] = documents
		}
		if document, activeErr := store.ActiveDocument(req.Context(), owner, projectID); activeErr == nil {
			payload["active_document"] = document
			shots, _ := store.ListShots(req.Context(), owner, projectID, document.ID)
			characters, _ := store.ListCharacters(req.Context(), owner, projectID, document.ID)
			scenes, _ := store.ListScenes(req.Context(), owner, projectID, document.ID)
			payload["shots"], payload["characters"], payload["scenes"] = shots, characters, scenes
		}
		assets, _ := store.ListAssets(req.Context(), owner, projectID, "")
		storyboardsA, _ := store.ListStoryboards(req.Context(), owner, projectID, screenplay.EngineXiaoluoV7)
		storyboardsB, _ := store.ListStoryboards(req.Context(), owner, projectID, screenplay.EngineXiaoyunqueV3)
		payload["assets"], payload["storyboards"] = assets, append(storyboardsA, storyboardsB...)
		if batch, batchErr := store.GetBatch(req.Context(), owner, projectID); batchErr == nil {
			payload["batch"] = batch
		}
		w.Header().Set("Content-Disposition", `attachment; filename="comic-screenplay-`+projectID+`.json"`)
		writeJSON(w, http.StatusOK, payload)
	})
}

func readDOCXText(data []byte) (string, error) {
	stream, err := openDocxDocumentXML(data)
	if err != nil {
		if errors.Is(err, errDocxMissingDoc) {
			return "", errors.New("DOCX 中没有找到正文")
		}
		if isDocxLimitError(err) {
			return "", err
		}
		return "", errors.New("DOCX 文件已损坏")
	}
	defer stream.Close()
	decoder := xml.NewDecoder(stream)
	var out strings.Builder
	for {
		token, tokenErr := decoder.Token()
		if tokenErr == io.EOF {
			break
		}
		if tokenErr != nil {
			return "", errors.New("DOCX 正文解析失败")
		}
		switch value := token.(type) {
		case xml.CharData:
			out.Write([]byte(value))
		case xml.EndElement:
			if value.Name.Local == "p" {
				out.WriteString("\n")
			}
		}
	}
	return out.String(), nil
}

func screenplayErrorStatus(err error) int {
	status := http.StatusInternalServerError
	switch {
	case strings.Contains(err.Error(), "已存在"):
		status = http.StatusConflict
	case errors.Is(err, screenplay.ErrNotFound), errors.Is(err, screenplay.ErrNoActiveScript):
		status = http.StatusNotFound
	case errors.Is(err, screenplay.ErrRevisionConflict), errors.Is(err, screenplay.ErrLockBusy), errors.Is(err, screenplay.ErrInvalidState):
		status = http.StatusConflict
	case errors.Is(err, screenplay.ErrInvalidArgument), errors.Is(err, screenplay.ErrEmptyContent), errors.Is(err, screenplay.ErrEmptyNovel), errors.Is(err, screenplay.ErrArtifactInvalid), errors.Is(err, screenplay.ErrAIInvalidJSON), errors.Is(err, screenplay.ErrAIInvalidStoryboard):
		status = http.StatusBadRequest
	case errors.Is(err, screenplay.ErrUserCancelled):
		status = 499
	}
	return status
}

func writeScreenplayError(w http.ResponseWriter, err error) {
	status := screenplayErrorStatus(err)
	writeJSON(w, status, map[string]string{"error": strings.TrimSpace(err.Error())})
}
