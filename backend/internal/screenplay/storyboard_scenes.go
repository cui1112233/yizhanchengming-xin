package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/prompt/luoshui"
)

type storyboardSceneInput struct{ Content string }

func storyboardScenesForEpisode(req StoryboardRequest, state *AuthoringState, episode episodeInput) ([]storyboardSceneInput, error) {
	if state != nil && !req.UseOriginalContentDirectly {
		for i, chapter := range state.Chapters {
			if (req.ChapterID != "" && chapter.ID != req.ChapterID) || (req.ChapterID == "" && i+1 != episode.No) {
				continue
			}
			// Edited text is authoritative. Reuse structured source only while it
			// exactly corresponds to the saved editable script, preserving element IDs.
			if rendered, e := professionalScriptMarkdown(chapter.RawResult); e == nil && strings.TrimSpace(rendered) == strings.TrimSpace(episode.Content) {
				if object, ok := extractJSONObject(chapter.RawResult); ok {
					var doc struct {
						Scenes []json.RawMessage `json:"scenes"`
					}
					if json.Unmarshal([]byte(object), &doc) == nil && len(doc.Scenes) > 0 {
						units := make([]storyboardSceneInput, 0, len(doc.Scenes))
						for _, scene := range doc.Scenes {
							units = append(units, storyboardSceneInput{Content: string(scene)})
						}
						return units, nil
					}
				}
			}
		}
	}
	return splitStoryboardScenes(episode.Content)
}

// Mirrors the recovered frontend heading families. Unknown prose remains one
// source unit; it is never silently reassigned to the neighbouring chapter.
var storyboardSceneHeading = regexp.MustCompile(`(?i)^(?:[0-9]{1,3}\s*[-－—–]\s*[0-9]{1,3}(?:\s|[：:]|[\p{Han}])|第[一二三四五六七八九十百千万\d]+[场幕]|场景\s*[一二三四五六七八九十百千万\d]+|scene\s*\d+|(?:int|ext|内景|外景|内/外|外/内|日景|夜景)[.．\s、:：-]|[0-9]+[.、]\s*(?:内景|外景|int|ext|场景))`)

func splitStoryboardScenes(content string) ([]storyboardSceneInput, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	starts := []int{}
	for i, line := range lines {
		if storyboardSceneHeading.MatchString(strings.TrimSpace(strings.Trim(line, "*#> \t"))) {
			starts = append(starts, i)
		}
	}
	if len(starts) == 0 {
		if strings.TrimSpace(content) == "" {
			return nil, ErrEmptyContent
		}
		return []storyboardSceneInput{{Content: strings.TrimSpace(content)}}, nil
	}
	scenes := []storyboardSceneInput{}
	for i, start := range starts {
		end := len(lines)
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		if strings.TrimSpace(strings.Join(lines[start+1:end], "\n")) == "" {
			return nil, fmt.Errorf("%w: 第%d场只有标题，没有场景正文", ErrEmptyContent, i+1)
		}
		text := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
		// Preserve any non-heading chapter preamble in the first source unit.
		if i == 0 && start > 0 {
			text = strings.TrimSpace(strings.Join(lines[:start], "\n")) + "\n" + text
		}
		scenes = append(scenes, storyboardSceneInput{Content: text})
	}
	return scenes, nil
}

func (s *StoryboardService) generateSceneShots(ctx context.Context, req StoryboardRequest, episode episodeInput, units []storyboardSceneInput, selected prompt.CreativePromptSnapshot, roles, scenes, props, style, synopsis, prefix, planning, model string, useHistory bool, aspectRatio string) ([]StoryboardShot, error) {
	if req.sceneOnlyIndex == nil && selected.SourceID > 0 && !useHistory && len(units) > 1 && s.SceneConcurrency != 1 {
		return s.generateConcurrentScenes(ctx, req, episode, units, selected, roles, scenes, props, style, synopsis, prefix, planning, model, aspectRatio)
	}
	all := []StoryboardShot{}
	var firstErr error
	for index, unit := range units {
		if req.sceneOnlyIndex != nil && index != *req.sceneOnlyIndex {
			continue
		}
		if err := ctx.Err(); err != nil {
			return all, err
		}
		phase := "storyboard"
		if len(units) > 1 {
			phase = fmt.Sprintf("storyboard_scene_%d", index+1)
		}
		vars := map[string]string{"CURRENT_SCENE_CONTENT": unit.Content, "CURRENT_SCENE_POSITION": strconv.Itoa(index + 1), "TOTAL_SCENE_COUNT": strconv.Itoa(len(units)), "ROLES_PLACEHOLDER": roles, "SCENES_PLACEHOLDER": scenes, "PROPS_PLACEHOLDER": props, "style": style, "synopsis": synopsis, "PREVIOUS_SCENE_POSITION": "", "PREVIOUS_SCENE_CONTENT": "", "NEXT_SCENE_POSITION": "", "NEXT_SCENE_CONTENT": ""}
		if index > 0 {
			vars["PREVIOUS_SCENE_POSITION"] = strconv.Itoa(index)
			vars["PREVIOUS_SCENE_CONTENT"] = units[index-1].Content
		}
		vars["STYLES_PLACEHOLDER"] = style
		vars["aspect_ratio"] = aspectRatio
		if index+1 < len(units) {
			vars["NEXT_SCENE_POSITION"] = strconv.Itoa(index + 2)
			vars["NEXT_SCENE_CONTENT"] = units[index+1].Content
		}
		user := unit.Content
		if selected.SourceID > 0 {
			var err error
			user, err = prompt.StoryboardSceneUserMessage(episode.OriginalContent, unit.Content)
			if err != nil {
				return all, fmt.Errorf("第%d集第%d场消息构建失败：%w", episode.No, index+1, err)
			}
		}
		if planning != "" {
			user += "\n\n【本次分镜规划】\n" + planning
		}
		if useHistory {
			history := req.generationSession.history(episode.No)
			if history != "" {
				user += "\n\n【最近5条已生成分镜历史，仅供衔接参考，不重复输出】\n" + history
			}
			if len(all) > 0 {
				recent := all
				if len(recent) > 5 {
					recent = recent[len(recent)-5:]
				}
				raw, _ := json.Marshal(storyboardSemanticRows(recent))
				user += "\n\n【本章此前已完成场景，仅供衔接参考】\n" + string(raw)
			}
		}
		if selected.SourceID == 164 {
			// Render's braces wrap these keys, matching the ZIP's double-braced Chinese variables.
			vars["{剧本内容}"] = episode.Content
			vars["{前场景末尾/后场景开头分镜}"] = ""
			if useHistory {
				history := req.generationSession.history(episode.No)
				if len(all) > 0 {
					recent := all
					if len(recent) > 5 {
						recent = recent[len(recent)-5:]
					}
					raw, _ := json.Marshal(storyboardSemanticRows(recent))
					history += "\n" + string(raw)
				}
				vars["{前场景末尾/后场景开头分镜}"] = strings.TrimSpace(history)
			}
		}
		system := luoshui.Render(selected.Body, vars)
		if strings.TrimSpace(prefix) != "" {
			system = prefix + "\n\n" + system
		}
		response, err := s.generationCall(ctx, req, episode.No, phase, TextRequest{Model: model, SystemPrompt: system, UserPrompt: user, JSONResponse: selected.ResponseMode != "json_with_report" && selected.ResponseMode != "text"})
		if err == nil {
			err = ctx.Err()
		}
		var shots []StoryboardShot
		if err == nil {
			shots, err = parseStoryboardJSON(response.Content)
		}
		if err == nil {
			err = validateSelectedStoryboardProtocol(selected, response.Content, shots)
		}
		if err != nil {
			err = wrapAIError(err)
			if ctx.Err() == nil && req.generationSession.callStatus(episode.No, phase) == "received" {
				if saveErr := req.generationSession.mutate(context.WithoutCancel(ctx), func(a *StoryboardGenerationAttempt) error {
					call := a.Episodes[episode.No].Calls[phase]
					call.Status = "invalid"
					call.Error = err.Error()
					a.Episodes[episode.No].Calls[phase] = call
					return nil
				}); saveErr != nil {
					return all, saveErr
				}
			}
			wrapped := fmt.Errorf("第%d集第%d场：%w", episode.No, index+1, err)
			if saveErr := req.generationSession.episodeError(episode.No, wrapped); saveErr != nil {
				return all, saveErr
			}
			if firstErr == nil {
				firstErr = wrapped
			}
			if ctx.Err() != nil || req.generationSession.callStatus(episode.No, phase) == "unknown" {
				return all, wrapped
			}
			continue
		}
		for local := range shots {
			shots[local].ID = newScreenplayID("sb_shot")
			shots[local].SourceSceneIndex = index
			shots[local].SourceSceneLocalOrder = local + 1
			shots[local].Seq = len(all) + 1
			if shots[local].Duration == 0 && selected.FixedDurationSec > 0 {
				shots[local].Duration = selected.FixedDurationSec
			}
			var fields map[string]any
			_ = json.Unmarshal(shots[local].ProtocolFields, &fields)
			fields["_cm_contract"] = map[string]any{"prompt_id": selected.ID, "prompt_digest": selected.Digest, "output_protocol": selected.OutputProtocol, "parser_version": selected.ParserVersion}
			shots[local].ProtocolFields, _ = json.Marshal(fields)
			all = append(all, shots[local])
		}
	}
	return all, firstErr
}

func validateSelectedStoryboardProtocol(p prompt.CreativePromptSnapshot, raw string, shots []StoryboardShot) error {
	// Recovered templates have differing output contracts. Legacy engines retain
	// their existing warning-only duration checks until explicitly versioned.
	if p.SourceID == 0 {
		return nil
	}
	seenIDs := map[float64]bool{}
	for i, shot := range shots {
		var fields map[string]json.RawMessage
		if json.Unmarshal(shot.ProtocolFields, &fields) != nil {
			return ErrAIInvalidStoryboard
		}
		for _, key := range p.RequiredFields {
			v, ok := fields[key]
			if !ok || string(v) == "null" {
				return fmt.Errorf("%w: 第%d组缺少%s", ErrAIInvalidStoryboard, i+1, key)
			}
		}
		if value, ok := fields["id"]; ok {
			var id float64
			if json.Unmarshal(value, &id) != nil || id <= 0 || id != float64(int(id)) || seenIDs[id] {
				return fmt.Errorf("%w: 第%d组id须为正整数", ErrAIInvalidStoryboard, i+1)
			}
			seenIDs[id] = true
		}
		if value, ok := fields["scene_id"]; ok {
			if p.OutputProtocol == "storyboard_nine_grid_v1" {
				var scene string
				if json.Unmarshal(value, &scene) != nil || strings.TrimSpace(scene) == "" {
					return fmt.Errorf("%w: 第%d组scene_id须为场景文字编号", ErrAIInvalidStoryboard, i+1)
				}
			} else {
				var scene float64
				if json.Unmarshal(value, &scene) != nil || scene <= 0 || scene != float64(int(scene)) {
					return fmt.Errorf("%w: 第%d组scene_id须为正整数", ErrAIInvalidStoryboard, i+1)
				}
			}
		}
		if duration, ok := fields["duration_sec"]; ok {
			var value float64
			if json.Unmarshal(duration, &value) != nil || value <= 0 {
				return fmt.Errorf("%w: 第%d组duration_sec须为正数", ErrAIInvalidStoryboard, i+1)
			}
		}
		for _, key := range []string{"video_desc", "image_prompt", "image_desc", "scene", "voice_actor", "emotion", "line", "characters", "props", "duration_calc", "timecode_global", "content_density"} {
			if value, ok := fields[key]; ok {
				var text string
				if string(value) == "null" || json.Unmarshal(value, &text) != nil {
					return fmt.Errorf("%w: 第%d组%s应为文字", ErrAIInvalidStoryboard, i+1, key)
				}
				optionalImage := key == "image_prompt" && prompt.StoryboardAllowsMissingImagePrompt(p)
				if (key == "video_desc" || key == "image_prompt" || key == "image_desc") && strings.TrimSpace(text) == "" && !optionalImage {
					return fmt.Errorf("%w: 第%d组%s为空", ErrAIInvalidStoryboard, i+1, key)
				}
			}
		}
		if value, ok := fields["source_elements"]; ok {
			var elements []string
			if json.Unmarshal(value, &elements) != nil {
				return fmt.Errorf("%w: 第%d组source_elements应为文字数组", ErrAIInvalidStoryboard, i+1)
			}
		}
		if shot.Duration > 0 && ((p.MinDurationSec > 0 && shot.Duration < p.MinDurationSec) || (p.MaxDurationSec > 0 && shot.Duration > p.MaxDurationSec) || (p.FixedDurationSec > 0 && shot.Duration != p.FixedDurationSec)) {
			return fmt.Errorf("%w: 第%d组时长不符合所选模板", ErrAIInvalidStoryboard, i+1)
		}
	}
	if p.ResponseMode == "json_with_report" {
		object, ok := extractJSONObject(raw)
		if !ok {
			return ErrAIInvalidJSON
		}
		_, tail, _ := strings.Cut(raw, object)
		for _, marker := range p.ReportMarkers {
			if !strings.Contains(tail, marker) {
				return fmt.Errorf("%w: 所选模板缺少JSON外报告%s", ErrAIInvalidStoryboard, marker)
			}
		}
		if p.SourceID == 137 {
			for _, section := range []string{"场景", "剧本元素总数", "台词总数", "遗漏", "连贯性检查", "映射完整性"} {
				if !strings.Contains(tail, section) {
					return fmt.Errorf("%w: 分镜自检缺少%s", ErrAIInvalidStoryboard, section)
				}
			}
		}
	}
	return nil
}
