package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
	"strconv"
	"strings"
)

// The plaintext draft schema is independently mapped from the public protocol
// represented by pyJianYingDraft's segment/local_materials/track serializers:
// https://github.com/GuanYixuan/pyJianYingDraft . Native app import is not verified.
func screenplayJianyingDraft(project aicut.Project, tl aicut.Timeline, assets []aicut.Asset, paths map[string]string, dimensions ...map[string][2]int) (map[string]any, map[string]any, error) {
	width, height := 1080, 1920
	if project.AspectRatio == "16:9" {
		width, height = 1920, 1080
	} else if project.AspectRatio == "1:1" {
		width, height = 1080, 1080
	}
	materials := map[string][]map[string]any{}
	for _, key := range strings.Fields("videos audios texts speeds stickers effects transitions audio_fades audio_effects canvases chromas common_mask masks material_animations video_effects beats images flowers handwrites hsl keyframes sound_channel_mappings vocal_separations") {
		materials[key] = []map[string]any{}
	}
	byID := map[string]aicut.Asset{}
	for _, a := range assets {
		byID[a.ID] = a
	}
	added := map[string]bool{}
	tracks := []map[string]any{}
	var duration int64
	for ti, track := range tl.Tracks {
		if track.Kind != aicut.TrackVideo && track.Kind != aicut.TrackAudio && track.Kind != aicut.TrackSubtitle {
			return nil, nil, fmt.Errorf("剪映基础草稿暂不支持特效或转场轨道，请导出成片")
		}
		kind := string(track.Kind)
		if kind == "subtitle" {
			kind = "text"
		}
		segments := []map[string]any{}
		for _, c := range track.Clips {
			mid := c.AssetID
			if c.Kind == aicut.ClipSubtitle {
				mid = "text-" + c.ID
				content, _ := json.Marshal(map[string]any{"text": c.Text, "styles": []map[string]any{{"range": []int{0, len([]rune(c.Text))}, "size": 8, "fill": map[string]any{"alpha": 1, "content": map[string]any{"render_type": "solid", "solid": map[string]any{"alpha": 1, "color": []float64{1, 1, 1}}}}, "bold": false, "italic": false, "underline": false, "strokes": []any{}}}})
				materials["texts"] = append(materials["texts"], map[string]any{"id": mid, "content": string(content), "type": "text", "check_flag": 7, "alignment": 1, "typesetting": 0, "global_alpha": 1, "line_feed": 1, "line_spacing": .02, "letter_spacing": 0})
			} else {
				a, ok := byID[mid]
				if !ok || paths[mid] == "" {
					return nil, nil, fmt.Errorf("镜头引用的素材未包含在资料包中：%s", mid)
				}
				if !added[mid] {
					added[mid] = true
					if a.Kind == "audio" {
						materials["audios"] = append(materials["audios"], map[string]any{"id": mid, "local_material_id": mid, "music_id": mid, "name": a.Name, "path": paths[mid], "type": "extract_music", "duration": a.DurationMS * 1000, "category_name": "local", "source_platform": 0, "wave_points": []any{}})
					} else {
						w, h := width, height
						if len(dimensions) > 0 {
							size, ok := dimensions[0][a.ID]
							if !ok || size[0] <= 0 || size[1] <= 0 {
								return nil, nil, fmt.Errorf("素材画面尺寸读取失败：%s", a.Name)
							}
							w, h = size[0], size[1]
						}
						typ := "video"
						d := a.DurationMS * 1000
						if a.Kind == "image" {
							typ = "photo"
							d = 10800000000
						}
						materials["videos"] = append(materials["videos"], map[string]any{"id": mid, "material_id": mid, "local_material_id": "", "material_name": a.Name, "path": paths[mid], "media_path": "", "type": typ, "duration": d, "width": w, "height": h, "check_flag": 63487, "category_name": "local", "crop_ratio": "free", "crop_scale": 1, "crop": map[string]any{"upper_left_x": 0, "upper_left_y": 0, "upper_right_x": 1, "upper_right_y": 0, "lower_left_x": 0, "lower_left_y": 1, "lower_right_x": 1, "lower_right_y": 1}})
					}
				}
			}
			speedID := "speed-" + c.ID
			materials["speeds"] = append(materials["speeds"], map[string]any{"id": speedID, "type": "speed", "mode": 0, "speed": 1, "curve_speed": nil})
			seg := map[string]any{"id": c.ID, "material_id": mid, "target_timerange": map[string]int64{"start": c.StartMS * 1000, "duration": c.DurationMS * 1000}, "source_timerange": map[string]int64{"start": c.SourceStartMS * 1000, "duration": c.DurationMS * 1000}, "speed": 1, "volume": 1, "visible": true, "reverse": false, "extra_material_refs": []string{speedID}, "common_keyframes": []any{}, "keyframe_refs": []any{}, "track_attribute": 0, "track_render_index": 0, "render_index": ti, "enable_adjust": true, "enable_lut": true, "last_nonzero_volume": 1, "is_tone_modify": false}
			if track.Muted {
				seg["volume"] = 0
			}
			if c.Kind == aicut.ClipSubtitle {
				seg["source_timerange"] = nil
			}
			if c.Kind != aicut.ClipAudio {
				transform := aicut.CanvasTransform{ScaleX: 1, ScaleY: 1, Opacity: 1}
				if c.CanvasTransform != nil {
					transform = *c.CanvasTransform
				}
				seg["clip"] = map[string]any{"alpha": transform.Opacity, "rotation": transform.Rotation, "scale": map[string]float64{"x": transform.ScaleX, "y": transform.ScaleY}, "transform": map[string]float64{"x": transform.X, "y": transform.Y}, "flip": map[string]bool{"horizontal": false, "vertical": false}}
				seg["uniform_scale"] = map[string]any{"on": true, "value": 1}
			}
			segments = append(segments, seg)
			if end := (c.StartMS + c.DurationMS) * 1000; end > duration {
				duration = end
			}
		}
		tracks = append(tracks, map[string]any{"id": track.ID, "name": track.Name, "type": kind, "attribute": 0, "flag": 0, "is_default_name": false, "segments": segments})
	}
	draft := map[string]any{"id": project.ID, "name": project.Title, "version": 360000, "new_version": "110.0.0", "fps": 30, "duration": duration, "canvas_config": map[string]any{"width": width, "height": height, "ratio": "original"}, "config": map[string]any{"maintrack_adsorb": true, "material_save_mode": 0, "video_mute": false}, "materials": materials, "tracks": tracks, "keyframes": map[string]any{"videos": []any{}, "audios": []any{}, "texts": []any{}}, "relationships": []any{}, "source": "default", "color_space": 0, "platform": map[string]any{"app_id": 3704, "app_source": "lv", "app_version": "5.9.0", "os": "windows"}}
	meta := map[string]any{"draft_id": project.ID, "draft_name": project.Title, "draft_fold_path": ".", "draft_root_path": ".", "draft_cover": "", "tm_duration": duration, "tm_draft_modified": 0, "draft_materials": []any{}, "draft_materials_copied_info": []any{}}
	return draft, meta, nil
}

const screenplayJianyingRelinkScript = `$ErrorActionPreference = 'Stop'
$draftRoot = $PSScriptRoot
$draftPath = Join-Path $draftRoot 'draft_content.json'
$draft = Get-Content -LiteralPath $draftPath -Raw -Encoding UTF8 | ConvertFrom-Json
foreach ($group in @($draft.materials.videos, $draft.materials.audios)) {
 foreach ($material in $group) { if ($material.path) { $material.path = [IO.Path]::GetFullPath((Join-Path $draftRoot $material.path)) } }
}
$draft | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $draftPath -Encoding UTF8
$metaPath = Join-Path $draftRoot 'draft_meta_info.json'
$meta = Get-Content -LiteralPath $metaPath -Raw -Encoding UTF8 | ConvertFrom-Json
$meta.draft_fold_path = $draftRoot
$meta.draft_root_path = [IO.Path]::GetDirectoryName($draftRoot)
$meta | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $metaPath -Encoding UTF8
Write-Output '素材地址已连接。请在剪映中打开此草稿目录。'
`

func probeScreenplayDimensions(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore) ([2]int, error) {
	var size [2]int
	dir, e := os.MkdirTemp("", "screenplay-draft-probe-")
	if e != nil {
		return size, e
	}
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "source"+filepath.Ext(asset.Name))
	if e = copyAICutObjectToFile(ctx, objects, asset, file); e != nil {
		return size, e
	}
	cmd, _, e := ffmpegbin.Command(ctx, "ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", file)
	if e != nil {
		return size, e
	}
	raw, e := cmd.Output()
	if e != nil {
		return size, e
	}
	parts := strings.Split(strings.TrimSpace(string(raw)), "x")
	if len(parts) != 2 {
		return size, fmt.Errorf("素材画面尺寸读取失败")
	}
	size[0], e = strconv.Atoi(parts[0])
	if e != nil {
		return size, e
	}
	size[1], e = strconv.Atoi(parts[1])
	if e != nil {
		return size, e
	}
	return size, nil
}
