package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	h3CharacterFactsHeading      = "## 调用一：H3 人物事实提取"
	h3AssetFactsHeading          = "## 调用一：H3 人物场景道具事实提取"
	h3CharacterAppearanceHeading = "## 调用二：H3 全人物外形编译"
)

// h3CharacterPromptPhases keeps the two model instructions inside the single
// H3 system preset selected by the user. Older saved H3 presets remain readable:
// their body is retained as the appearance phase and the caller keeps its
// legacy extraction contract for the first request.
func h3CharacterPromptPhases(body string) (facts string, appearance string) {
	trimmed := strings.TrimSpace(body)
	factsAt := strings.Index(trimmed, h3AssetFactsHeading)
	if factsAt < 0 {
		factsAt = strings.Index(trimmed, h3CharacterFactsHeading)
	}
	appearanceAt := strings.Index(trimmed, h3CharacterAppearanceHeading)
	if factsAt < 0 || appearanceAt < 0 || appearanceAt <= factsAt {
		return "", trimmed
	}
	facts = strings.TrimSpace(trimmed[factsAt:appearanceAt])
	appearance = strings.TrimSpace(trimmed[appearanceAt:])
	return facts, appearance
}

// Request-local IDs keep reordered model responses attached to the extracted
// identities. Persistent asset identity remains owned by the asset repository.
func (s *DirectorService) compileH3CharacterBatch(ctx context.Context, book Book, preset PresetSnapshot, assets DirectorAssets) ([]NamedPrompt, error) {
	if len(assets.Characters) == 0 {
		return assets.Characters, nil
	}
	if strings.TrimSpace(preset.Body) == "" {
		return nil, fmt.Errorf("%w: H3 人物后台预设正文为空", ErrInvalid)
	}
	characters := make([]map[string]string, len(assets.Characters))
	for i, c := range assets.Characters {
		characters[i] = map[string]string{"character_id": fmt.Sprintf("C%03d", i+1), "name": c.Name, "prompt": c.Prompt}
	}
	payload, err := json.Marshal(map[string]any{"source_text": book.SourceText, "characters": characters, "scenes": assets.Scenes, "props": assets.Props})
	if err != nil {
		return nil, err
	}
	_, appearanceRule := h3CharacterPromptPhases(preset.Body)
	response, err := s.Provider.Complete(ctx, TextCompletionRequest{
		SystemPrompt: appearanceRule + "\n" + staticAssetPromptRule + "\n\n传输协议：一次返回全部人物的 JSON 对象 {\"characters\":[{\"character_id\":\"原样返回输入标识\",\"name\":\"原样返回姓名\",\"prompt\":\"完整外形正文\"}]}。不得遗漏、重复或新增人物。",
		UserPrompt:   string(payload), Temperature: 0.2, MaxTokens: 32000,
	})
	if err != nil {
		return nil, err
	}
	raw, err := ParseDirectorJSON(response)
	if err != nil {
		return nil, fmt.Errorf("%w: H3 人物批量结果不完整: %v", ErrInvalid, err)
	}
	var result struct {
		Characters []struct {
			ID     string `json:"character_id"`
			Name   string `json:"name"`
			Prompt string `json:"prompt"`
		} `json:"characters"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("%w: H3 人物批量结果格式错误", ErrInvalid)
	}
	if len(result.Characters) != len(characters) {
		return nil, fmt.Errorf("%w: H3 人物批量结果人数不完整", ErrInvalid)
	}
	byID := make(map[string]int, len(characters))
	for i, c := range characters {
		byID[c["character_id"]] = i
	}
	seen := map[string]bool{}
	out := append([]NamedPrompt(nil), assets.Characters...)
	for _, c := range result.Characters {
		i, ok := byID[c.ID]
		if !ok || seen[c.ID] || c.Name != out[i].Name {
			return nil, fmt.Errorf("%w: H3 人物标识重复、未知或姓名不匹配", ErrInvalid)
		}
		seen[c.ID] = true
		if strings.TrimSpace(c.Prompt) == "" {
			return nil, fmt.Errorf("%w: H3 人物外形为空", ErrInvalid)
		}
		out[i].Prompt = strings.TrimSpace(c.Prompt)
	}
	return out, nil
}
