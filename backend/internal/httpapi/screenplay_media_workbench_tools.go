package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/screenplay"
	"strconv"
	"strings"
	"time"
)

func (s *ScreenplayMediaWorkbench) splitGrid(ctx context.Context, p screenplay.Project, in screenplayMediaAction, st screenplay.MediaState) (any, error) {
	if in.Rows < 1 || in.Rows > 8 || in.Columns < 1 || in.Columns > 8 {
		return nil, fmt.Errorf("行数和列数须为1到8")
	}
	count := in.Rows * in.Columns
	if len(in.TargetIDs) > 0 && len(in.TargetIDs) != count {
		return nil, fmt.Errorf("每格绑定对象数量须与拆分格数相同")
	}
	for _, id := range in.TargetIDs {
		if e := s.validTarget(ctx, p, id); e != nil {
			return nil, e
		}
	}
	item, e := mediaItem(st, in.ItemID)
	if e != nil {
		return nil, e
	}
	if item.Kind != "image" || s.Cut == nil || s.Objects == nil {
		return nil, fmt.Errorf("请选择已保存图片")
	}
	asset, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
	if e != nil {
		return nil, e
	}
	body, e := s.Objects.Open(ctx, aicutObjectAsset(asset))
	if e != nil {
		return nil, e
	}
	raw, e := io.ReadAll(io.LimitReader(body, (32<<20)+1))
	body.Close()
	if e != nil || len(raw) > 32<<20 {
		return nil, fmt.Errorf("源图片读取失败或超过32MB")
	}
	config, _, e := image.DecodeConfig(bytes.NewReader(raw))
	if e != nil || config.Width <= 0 || config.Height <= 0 || config.Width > 20000 || config.Height > 20000 || config.Width*config.Height > 20000000 {
		return nil, fmt.Errorf("源图片无法读取或画面超过2000万像素")
	}
	outputs, e := NewAgentImageProcessor().Process(ctx, AgentImageProcessInput{Operation: agentworkspace.ImageGridSplit, Source: aicutObjectAsset(asset), SourceBytes: raw, Settings: map[string]any{"rows": in.Rows, "columns": in.Columns}})
	if e != nil {
		return nil, e
	}
	items := []screenplay.MediaItem{}
	for i, out := range outputs {
		target := ""
		if len(in.TargetIDs) > 0 {
			target = in.TargetIDs[i]
		}
		saved, e := s.saveBytes(ctx, p, "image", out.Name, out.ContentType, bytes.NewReader(out.Bytes), target)
		if e != nil {
			return nil, fmt.Errorf("已保存%d格，后续保存失败：%w", len(items), e)
		}
		items = append(items, saved)
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error {
		for _, item := range items {
			if item.TargetID != "" {
				v.Selected[item.TargetID] = item.ID
			}
		}
		return nil
	})
	return map[string]any{"createdItems": items}, e
}

type screenplayPackageManifest struct {
	Format   string            `json:"format"`
	Timeline aicut.Timeline    `json:"timeline"`
	Media    map[string]string `json:"media"`
	Assets   []aicut.Asset     `json:"assets"`
}
type screenplayImportAsset struct {
	OriginalID  string
	Name        string
	Kind        string
	ContentType string
	Raw         []byte
	DurationMS  int64
}

func parseScreenplayMediaPackage(ctx context.Context, archive *zip.Reader) (screenplayPackageManifest, []screenplayImportAsset, error) {
	var manifest screenplayPackageManifest
	entries := map[string]*zip.File{}
	var total uint64
	if len(archive.File) > 1024 {
		return manifest, nil, fmt.Errorf("资料包文件过多")
	}
	for _, f := range archive.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 || strings.Contains(f.Name, "\\") || path.IsAbs(f.Name) || path.Clean(f.Name) != f.Name || strings.HasPrefix(f.Name, "../") || strings.Contains(f.Name, ":") {
			return manifest, nil, fmt.Errorf("资料包包含不安全的路径")
		}
		if _, exists := entries[f.Name]; exists {
			return manifest, nil, fmt.Errorf("资料包包含重复文件")
		}
		total += f.UncompressedSize64
		if total > 512<<20 || f.UncompressedSize64 > 128<<20 {
			return manifest, nil, fmt.Errorf("资料包超过大小限制")
		}
		entries[f.Name] = f
	}
	mf := entries["timeline.json"]
	if mf == nil || mf.UncompressedSize64 > 8<<20 {
		return manifest, nil, fmt.Errorf("请选择本系统导出的媒体资料包，缺少时间线文件")
	}
	body, e := mf.Open()
	if e != nil {
		return manifest, nil, e
	}
	raw, e := io.ReadAll(io.LimitReader(body, (8<<20)+1))
	body.Close()
	if e != nil {
		return manifest, nil, e
	}
	if e = json.Unmarshal(raw, &manifest); e != nil || manifest.Format != "v88-media-package-v1" || len(manifest.Media) == 0 {
		return manifest, nil, fmt.Errorf("此资料包不是本系统媒体工程格式")
	}
	metadata := map[string]aicut.Asset{}
	for _, a := range manifest.Assets {
		metadata[a.ID] = a
	}
	assets := []screenplayImportAsset{}
	validation := aicut.NewMemoryStore()
	if e = validation.CreateProject(ctx, "validate", aicut.Project{ID: "validate", Title: "验证"}); e != nil {
		return manifest, nil, e
	}
	for id, name := range manifest.Media {
		if !strings.HasPrefix(name, "media/") || entries[name] == nil {
			return manifest, nil, fmt.Errorf("资料包缺少媒体：%s", name)
		}
		if _, e = agentworkspace.AgentAssetObjectKey("validate", "validate", id, "bin"); e != nil {
			return manifest, nil, fmt.Errorf("资料包素材标识无效")
		}
		f := entries[name]
		body, e = f.Open()
		if e != nil {
			return manifest, nil, e
		}
		contents, e := io.ReadAll(io.LimitReader(body, (128<<20)+1))
		body.Close()
		if e != nil || len(contents) > 128<<20 || len(contents) == 0 {
			return manifest, nil, fmt.Errorf("资料包媒体读取失败")
		}
		ct := http.DetectContentType(contents)
		if ct == "application/octet-stream" {
			candidate := mime.TypeByExtension(path.Ext(name))
			if strings.HasPrefix(candidate, "audio/") || strings.HasPrefix(candidate, "video/") {
				ct = candidate
			}
		}
		kind := strings.Split(ct, "/")[0]
		if !agentworkspace.ValidAssetKind(kind) {
			return manifest, nil, fmt.Errorf("资料包包含不支持的媒体：%s", name)
		}
		duration, e := validateScreenplayImportMedia(ctx, contents, kind, path.Ext(name))
		if e != nil {
			return manifest, nil, fmt.Errorf("媒体%s无效：%w", name, e)
		}
		label := metadata[id].Name
		if label == "" {
			label = path.Base(name)
		}
		asset := screenplayImportAsset{OriginalID: id, Name: path.Base(label), Kind: kind, ContentType: ct, Raw: contents, DurationMS: duration}
		assets = append(assets, asset)
		if e = validation.CreateAsset(ctx, "validate", aicut.Asset{ID: id, ProjectID: "validate", Kind: kind, Name: asset.Name, ContentType: ct, ObjectKey: name, ByteSize: int64(len(contents)), DurationMS: duration, CreatedAt: time.Now()}); e != nil {
			return manifest, nil, e
		}
	}
	for _, track := range manifest.Timeline.Tracks {
		for _, clip := range track.Clips {
			if clip.AssetID == "" {
				continue
			}
			a, e := validation.Asset(ctx, "validate", "validate", clip.AssetID)
			if e != nil {
				return manifest, nil, fmt.Errorf("时间线引用了资料包外的素材")
			}
			if string(clip.Kind) != a.Kind {
				return manifest, nil, fmt.Errorf("时间线素材类型不匹配")
			}
			if a.Kind != "image" && clip.SourceStartMS+clip.DurationMS > a.DurationMS+40 {
				return manifest, nil, fmt.Errorf("时间线裁切超出实际媒体时长")
			}
		}
	}
	if _, e = validation.SaveTimeline(ctx, "validate", "validate", manifest.Timeline, 0, time.Now()); e != nil {
		return manifest, nil, fmt.Errorf("资料包时间线无效：%w", e)
	}
	return manifest, assets, nil
}
func validateScreenplayImportMedia(ctx context.Context, raw []byte, kind, extension string) (int64, error) {
	if kind == "image" {
		cfg, _, e := image.DecodeConfig(bytes.NewReader(raw))
		if e != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 20000 || cfg.Height > 20000 || cfg.Width*cfg.Height > 20000000 {
			return 0, fmt.Errorf("图片无效或超过2000万像素")
		}
		return 0, nil
	}
	dir, e := os.MkdirTemp("", "screenplay-import-")
	if e != nil {
		return 0, e
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "source"+extension)
	if e = os.WriteFile(file, raw, 0600); e != nil {
		return 0, e
	}
	cmd, _, e := ffmpegbin.Command(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", file)
	if e != nil {
		return 0, e
	}
	out, e := cmd.Output()
	if e != nil {
		return 0, e
	}
	sec, e := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if e != nil || sec <= 0 {
		return 0, fmt.Errorf("媒体时长无效")
	}
	return int64(sec*1000 + .5), nil
}
func (s *ScreenplayMediaWorkbench) importMedia(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
	r.Body = http.MaxBytesReader(w, r.Body, (512<<20)+(1<<20))
	if e := r.ParseMultipartForm(8 << 20); e != nil {
		writeJSON(w, 400, map[string]string{"error": "资料包无效或超过512MB"})
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, header, e := r.FormFile("file")
	if e != nil {
		writeJSON(w, 400, map[string]string{"error": "请选择资料包"})
		return
	}
	defer file.Close()
	archive, e := zip.NewReader(file, header.Size)
	if e != nil {
		writeJSON(w, 400, map[string]string{"error": "请选择ZIP资料包"})
		return
	}
	expected, e := strconv.ParseInt(r.FormValue("expectedRevision"), 10, 64)
	if e != nil {
		writeJSON(w, 400, map[string]string{"error": "请提供当前时间线版本"})
		return
	}
	if e = s.ensureCut(r.Context(), p); e != nil {
		writeScreenplayError(w, e)
		return
	}
	current, e := s.Cut.Timeline(r.Context(), p.OwnerUsername, p.ID)
	if e != nil {
		writeScreenplayError(w, e)
		return
	}
	if current.Revision != expected {
		writeAICutError(w, aicut.ErrRevisionConflict)
		return
	}
	manifest, assets, e := parseScreenplayMediaPackage(r.Context(), archive)
	if e != nil {
		writeJSON(w, 400, map[string]string{"error": e.Error()})
		return
	}
	token := newAgentID("import-lock")
	lock := "screenplay-media-import:" + p.OwnerUsername + ":" + p.ID
	if e = s.Projects.AcquireLock(r.Context(), lock, token, 10*time.Minute); e != nil {
		writeScreenplayError(w, e)
		return
	}
	defer s.Projects.ReleaseLock(context.WithoutCancel(r.Context()), lock, token)
	mapping := map[string]string{}
	items := []screenplay.MediaItem{}
	for _, a := range assets {
		item, err := s.saveBytes(r.Context(), p, a.Kind, a.Name, a.ContentType, bytes.NewReader(a.Raw), "")
		if err != nil {
			writeJSON(w, 422, map[string]any{"error": fmt.Sprintf("已导入%d个素材，后续失败：%s；原时间线未被覆盖", len(items), err.Error()), "createdItems": items})
			return
		}
		mapping[a.OriginalID] = item.AssetID
		items = append(items, item)
	}
	for ti := range manifest.Timeline.Tracks {
		for ci := range manifest.Timeline.Tracks[ti].Clips {
			c := &manifest.Timeline.Tracks[ti].Clips[ci]
			if c.AssetID != "" {
				c.AssetID = mapping[c.AssetID]
			}
			c.ID = newAgentID("clip-import")
		}
		manifest.Timeline.Tracks[ti].ID = newAgentID("track-import")
	}
	for i := range manifest.Timeline.Markers {
		if manifest.Timeline.Markers[i].AssetID != "" {
			manifest.Timeline.Markers[i].AssetID = mapping[manifest.Timeline.Markers[i].AssetID]
		}
	}
	timeline, e := s.Cut.SaveTimeline(r.Context(), p.OwnerUsername, p.ID, manifest.Timeline, expected, time.Now().UTC())
	if e != nil {
		writeJSON(w, 409, map[string]any{"error": "媒体已保存，但时间线版本已变化，请刷新后重新应用", "createdItems": items})
		return
	}
	if e = screenplay.InvalidateMediaContent(r.Context(), s.Projects, p.OwnerUsername, p.ID, "edit"); e != nil {
		writeJSON(w, 422, map[string]string{"error": "工程已导入，但生产状态暂停失败，请刷新核对：" + e.Error()})
		return
	}
	writeJSON(w, 201, map[string]any{"createdItems": items, "timeline": timeline})
}

// This is the same maintained default-service catalog used by TtsPage. Custom
// providers keep their own voice IDs and are not claimed to support these IDs.
func (s *ScreenplayMediaWorkbench) voiceDirectory(ctx context.Context, owner string) map[string]any {
	if s.Catalog != nil {
		if _, found := configuredTTSModel(ctx, s.Catalog, owner); found {
			return map[string]any{"voices": []any{}, "source": "configured", "reason": "当前自定义配音服务没有公开音色目录，请填写该服务支持的音色ID"}
		}
	}
	return map[string]any{"source": "builtin-default-tts", "voices": []map[string]string{{"id": "zh-CN-XiaoxiaoNeural", "name": "晓晓（女声·温柔）"}, {"id": "zh-CN-XiaochenNeural", "name": "晓辰（女声·知性）"}, {"id": "zh-CN-YunxiNeural", "name": "云希（男声·清朗）"}, {"id": "zh-CN-YunyangNeural", "name": "云扬（男声·阳光）"}, {"id": "zh-CN-XiaoyiNeural", "name": "晓伊（女声·甜美）"}, {"id": "zh-CN-YunjianNeural", "name": "云健（男声·稳重）"}}}
}
