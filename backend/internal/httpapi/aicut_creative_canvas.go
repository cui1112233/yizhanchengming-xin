package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"strings"
	"sync"
	"time"
)

// Requests share immutable source objects and deterministic destination identities.
// Serialize copying in this Go process so a retry cannot delete another request's object.
var aicutCreativeTransferMu sync.Mutex

func aicutCreativeID(prefix string, parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return prefix + hex.EncodeToString(sum[:24])
}
func aicutCreativeTaskID(owner, project string) string {
	return aicutCreativeID("aicut-creative-", owner, project)
}

func RegisterAICutCreativeCanvasRoutes(mux *http.ServeMux, auth BridgeAuth, cuts aicut.Store, projects agentworkspace.Store, assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore, now func() time.Time) {
	if mux == nil || cuts == nil || projects == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			owner := strings.TrimSpace(identity.Username)
			if !ok || owner == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			project := strings.TrimSpace(r.PathValue("id"))
			if _, err := cuts.Project(r.Context(), owner, project); err != nil {
				writeAICutError(w, err)
				return
			}
			fn(w, r, owner, project)
		})))
	}
	handle("POST /api/ai-cut/projects/{id}/creative-canvas", func(w http.ResponseWriter, r *http.Request, owner, project string) {
		task := aicutCreativeTaskID(owner, project)
		item, err := projects.Project(r.Context(), owner, task)
		if errors.Is(err, agentworkspace.ErrProjectNotFound) {
			cut, loadErr := cuts.Project(r.Context(), owner, project)
			if loadErr != nil {
				writeAICutError(w, loadErr)
				return
			}
			stamp := now().UTC()
			item = agentworkspace.Project{ID: task, Title: cut.Title + " · 创作画布", TitleMode: "manual", ProjectType: "video", CreatedAt: stamp, UpdatedAt: stamp, Canvas: aicutCreativeSeed()}
			err = projects.CreateProject(r.Context(), owner, item)
			if err == nil || errors.Is(err, agentworkspace.ErrProjectExists) {
				item, err = projects.Project(r.Context(), owner, task)
			}
		}
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"taskId": task, "canvas": agentCanvasForResponse(task, item.Canvas)})
	})
	handle("POST /api/ai-cut/projects/{id}/creative-canvas/assets", func(w http.ResponseWriter, r *http.Request, owner, project string) {
		if assets == nil || objects == nil {
			writeJSON(w, 503, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		var input struct {
			AssetIDs []string `json:"assetIds"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input) != nil || len(input.AssetIDs) < 1 || len(input.AssetIDs) > 30 {
			writeJSON(w, 400, map[string]string{"error": "请选择 1–30 个素材"})
			return
		}
		task := aicutCreativeTaskID(owner, project)
		if _, err := projects.Project(r.Context(), owner, task); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		// Validate every source before writing destination objects or rows.
		sources := []agentworkspace.Asset{}
		seen := map[string]bool{}
		for _, raw := range input.AssetIDs {
			id := strings.TrimSpace(raw)
			if id == "" {
				writeJSON(w, 400, map[string]string{"error": "素材编号不能为空"})
				return
			}
			if seen[id] {
				continue
			}
			seen[id] = true
			source, err := assets.Asset(r.Context(), owner, task, id)
			if err != nil {
				writeJSON(w, 400, map[string]string{"error": "所选素材不存在或不属于当前创作画布：" + id})
				return
			}
			if source.ProjectID != task || agentworkspace.ValidateAsset(source) != nil || source.ByteSize <= 0 || source.ByteSize > aicutAssetUploadMaxBytes {
				writeJSON(w, 400, map[string]string{"error": "素材信息无效：" + id})
				return
			}
			sources = append(sources, source)
		}
		aicutCreativeTransferMu.Lock()
		defer aicutCreativeTransferMu.Unlock()
		dir, err := os.MkdirTemp("", "aicut-creative-*")
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "无法准备素材复制"})
			return
		}
		defer os.RemoveAll(dir)
		prepared := make([]aicutCreativePrepared, 0, len(sources))
		for _, source := range sources {
			targetID := aicutCreativeID("aicut-ca-", owner, project, task, source.ID)
			existing, findErr := cuts.Asset(r.Context(), owner, project, targetID)
			if findErr == nil {
				body, openErr := objects.Open(r.Context(), aicutObjectAsset(existing))
				if openErr == nil {
					_ = body.Close()
					prepared = append(prepared, aicutCreativePrepared{asset: existing, reused: true})
					continue
				}
			} else if !errors.Is(findErr, aicut.ErrAssetNotFound) {
				writeAICutError(w, findErr)
				return
			}
			filename := path.Join(dir, fmt.Sprintf("%d", len(prepared)))
			body, openErr := objects.Open(r.Context(), source)
			if openErr != nil {
				writeJSON(w, 502, map[string]string{"error": "读取创作素材失败：" + source.Name})
				return
			}
			file, createErr := os.Create(filename)
			if createErr != nil {
				body.Close()
				writeJSON(w, 500, map[string]string{"error": "无法准备素材复制"})
				return
			}
			size, copyErr := io.Copy(file, io.LimitReader(body, source.ByteSize+1))
			closeErr := file.Close()
			body.Close()
			if copyErr != nil || closeErr != nil || size != source.ByteSize {
				writeJSON(w, 502, map[string]string{"error": "创作素材内容缺失或长度不符：" + source.Name})
				return
			}
			ext := path.Ext(source.Name)
			if ext == "" {
				switch source.Kind {
				case "image":
					ext = ".image"
				case "video":
					ext = ".video"
				case "audio":
					ext = ".audio"
				}
			}
			key, keyErr := agentworkspace.AgentAssetObjectKey(owner, project, targetID, ext)
			if keyErr != nil {
				writeJSON(w, 400, map[string]string{"error": "素材文件名无效：" + source.Name})
				return
			}
			target := aicut.Asset{ID: targetID, ProjectID: project, Kind: source.Kind, Name: source.Name, ContentType: source.ContentType, ObjectKey: strings.Replace(key, "agent-assets/", "aicut-assets/", 1), ByteSize: size, CreatedAt: now().UTC()}
			if target.Kind == "video" || target.Kind == "audio" {
				probeSource := target
				probeSource.ObjectKey = source.ObjectKey
				duration, probeErr := probeAICutAssetDuration(r.Context(), probeSource, objects)
				if probeErr != nil || duration <= 0 {
					writeJSON(w, 422, map[string]string{"error": "无法读取素材时长：" + target.Name})
					return
				}
				target.DurationMS = duration
			}
			prepared = append(prepared, aicutCreativePrepared{asset: target, filename: filename, repair: findErr == nil})
		}
		result := make([]aicut.Asset, 0, len(prepared))
		for _, item := range prepared {
			target := item.asset
			if !item.reused {
				if err := aicutCreativeCopy(r.Context(), objects, target, item.filename); err != nil {
					writeJSON(w, 502, map[string]string{"error": "复制素材失败：" + target.Name})
					return
				}
				if !item.repair {
					if err := cuts.CreateAsset(r.Context(), owner, target); err != nil {
						if !errors.Is(err, aicut.ErrAssetExists) {
							writeAICutError(w, err)
							return
						}
						target, err = cuts.Asset(r.Context(), owner, project, target.ID)
						if err != nil {
							writeAICutError(w, err)
							return
						}
					}
				}
			}
			target.URL = aicutSignedAssetContentURL(auth, owner, project, target.ID, "content", now())
			result = append(result, target)
		}
		writeJSON(w, 200, map[string]any{"assets": result})
	})
	registerAICutCreativeScript(handle, cuts, projects, now)
}

type aicutCreativePrepared struct {
	asset          aicut.Asset
	filename       string
	reused, repair bool
}

func aicutCreativeCopy(ctx context.Context, objects agentworkspace.AssetObjectStore, target aicut.Asset, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return objects.Put(ctx, aicutObjectAsset(target), file)
}
func aicutCreativeSeed() agentworkspace.Canvas {
	canvas := agentworkspace.Canvas{Version: 1, Appearance: agentworkspace.DefaultAppearance(), Nodes: []agentworkspace.Node{}, Edges: []agentworkspace.Edge{}}
	for index, card := range []struct{ id, title, kind string }{{"story", "故事", "text"}, {"role", "角色", "text"}, {"scene", "场景", "text"}, {"prop", "道具", "text"}, {"storyboard", "分镜", "script"}} {
		id := "creative-" + card.id
		config := map[string]any{"prompt": "", "creativeCardKind": card.id}
		if card.kind == "script" {
			config["sourceText"] = ""
		}
		node := agentworkspace.Node{ID: id, Type: card.kind, Title: card.title, Version: 1, Status: "draft", Position: agentworkspace.Position{X: float64(80 + index%3*360), Y: float64(80 + index/3*340)}, Config: config, SourceNodeIDs: []string{}}
		if index > 0 {
			node.SourceNodeIDs = []string{"creative-story"}
			canvas.Edges = append(canvas.Edges, agentworkspace.Edge{ID: "creative-story-" + card.id, Source: "creative-story", Target: id, Relation: "reference"})
		}
		canvas.Nodes = append(canvas.Nodes, node)
	}
	return canvas
}

func registerAICutCreativeScript(handle func(string, func(http.ResponseWriter, *http.Request, string, string)), cuts aicut.Store, projects agentworkspace.Store, now func() time.Time) {
	handle("POST /api/ai-cut/projects/{id}/creative-canvas/script", func(w http.ResponseWriter, r *http.Request, owner, project string) {
		var input struct {
			NodeIDs          []string `json:"nodeIds"`
			ExpectedRevision *int64   `json:"expectedRevision"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input) != nil || len(input.NodeIDs) < 1 || len(input.NodeIDs) > 10 || input.ExpectedRevision == nil || *input.ExpectedRevision < 0 {
			writeJSON(w, 400, map[string]string{"error": "请选择 1–10 张文字卡，并提供当前时间线版本"})
			return
		}
		canvas, err := projects.Canvas(r.Context(), owner, aicutCreativeTaskID(owner, project))
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		nodes := map[string]agentworkspace.Node{}
		for _, node := range canvas.Nodes {
			nodes[node.ID] = node
		}
		seen := map[string]bool{}
		parts := []string{}
		for _, id := range input.NodeIDs {
			if seen[id] {
				continue
			}
			seen[id] = true
			node, ok := nodes[id]
			if !ok || (node.Type != "text" && node.Type != "script") {
				writeJSON(w, 400, map[string]string{"error": "文字卡不存在或不属于当前创作画布"})
				return
			}
			content := ""
			if node.Type == "script" {
				content, _ = node.Config["sourceText"].(string)
			}
			if strings.TrimSpace(content) == "" {
				content = node.Summary
			}
			if strings.TrimSpace(content) == "" {
				content, _ = node.Config["prompt"].(string)
			}
			content = strings.TrimSpace(content)
			if content == "" {
				writeJSON(w, 400, map[string]string{"error": "所选文字卡为空：" + node.Title})
				return
			}
			parts = append(parts, content)
		}
		timeline, err := cuts.Timeline(r.Context(), owner, project)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if timeline.Revision != *input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		addition := strings.Join(parts, "\n\n")
		if strings.TrimSpace(timeline.Script) != "" {
			timeline.Script += "\n\n" + addition
		} else {
			timeline.Script = addition
		}
		if len([]rune(timeline.Script)) > 100000 {
			writeJSON(w, 400, map[string]string{"error": "送入后的剧本不能超过 10 万字"})
			return
		}
		saved, err := cuts.SaveTimeline(r.Context(), owner, project, timeline, *input.ExpectedRevision, now().UTC())
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"timeline": saved})
	})
}
