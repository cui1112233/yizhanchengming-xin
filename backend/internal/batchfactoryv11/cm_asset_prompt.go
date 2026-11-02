package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// extractCMSceneAndPropPrompts follows the published CM preset contract. CM
// deliberately keeps scene and prop extraction as separate model phases, so
// a character-only first response cannot silently produce an incomplete
// asset library.
func (s *DirectorService) extractCMSceneAndPropPrompts(ctx context.Context, book Book, snapshot DirectorSnapshot, preset PresetSnapshot, assets DirectorAssets) (DirectorAssets, error) {
	scenes, err := s.runCMAssetPhase(ctx, book, snapshot, preset, assets, "scene")
	if err != nil {
		return DirectorAssets{}, err
	}
	props, err := s.runCMAssetPhase(ctx, book, snapshot, preset, DirectorAssets{
		Characters: assets.Characters,
		Scenes:     scenes,
	}, "prop")
	if err != nil {
		return DirectorAssets{}, err
	}
	return DirectorAssets{
		Characters: assets.Characters,
		Scenes:     mergeCMNamedPrompts(scenes, assets.Scenes),
		Props:      mergeCMNamedPrompts(props, assets.Props),
	}, nil
}

func (s *DirectorService) runCMAssetPhase(ctx context.Context, book Book, snapshot DirectorSnapshot, preset PresetSnapshot, assets DirectorAssets, kind string) ([]NamedPrompt, error) {
	contract, err := buildCMAssetPhaseContract(book, snapshot, preset, assets, kind)
	if err != nil {
		return nil, err
	}
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{
		SystemPrompt: contract.SystemPrompt,
		UserPrompt:   contract.UserPrompt,
		Temperature:  contract.Temperature,
		MaxTokens:    contract.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		return nil, retryableModelOutputError(err)
	}
	return normalizeCMAssetPhaseOutput(raw, kind)
}

func buildCMAssetPhaseContract(book Book, snapshot DirectorSnapshot, preset PresetSnapshot, assets DirectorAssets, kind string) (PromptContract, error) {
	phaseBody, err := cmAssetPhaseBody(preset.Body, kind)
	if err != nil {
		return PromptContract{}, err
	}
	style := rawString(snapshot.Effective, "h3StyleAnalysis", "")
	styleJSON, _ := json.Marshal(map[string]string{"style_system_analysis": style})
	characterJSON, _ := json.Marshal(map[string]any{"characters": assets.Characters})
	sceneJSON, _ := json.Marshal(map[string]any{"scenes": assets.Scenes})
	body := strings.NewReplacer(
		"${STYLE_JSON}", string(styleJSON),
		"${CHARACTER_GRAPH_JSON}", string(characterJSON),
		"${SCENE_JSON}", string(sceneJSON),
		"${SOURCE}", book.SourceText,
	).Replace(phaseBody)

	arrayKey, promptKey, label := "scenes", "scene_prompt", "场景"
	if kind == "prop" {
		arrayKey, promptKey, label = "props", "prop_prompt", "道具"
	}
	system := fmt.Sprintf("你是 CM 资产提取的%s阶段执行器。只执行当前%s阶段，不要返回人物，不要返回其它阶段内容。只输出一个合法 JSON 对象：{\"%s\":[{\"canonical_name\":\"标准名称\",\"%s\":\"完整稳定提示词\"}]}。没有可确认的%s时返回空数组。不得输出 Markdown 或解释。", label, label, arrayKey, promptKey, label)
	if strings.TrimSpace(body) != "" {
		system += "\n\n当前 CM 阶段规则：\n" + body
	}
	payload, err := json.MarshalIndent(map[string]any{
		"book_id":         book.ID,
		"title":           book.Title,
		"source_text":     book.SourceText,
		"phase":           kind,
		"characters":      assets.Characters,
		"frozen_scenes":   assets.Scenes,
		"output_contract": map[string]string{"array": arrayKey, "prompt": promptKey},
	}, "", "  ")
	if err != nil {
		return PromptContract{}, err
	}
	return PromptContract{
		SystemPrompt: system,
		UserPrompt:   string(payload),
		Temperature:  0.2,
		MaxTokens:    2600,
	}, nil
}

func cmAssetPhaseBody(body, kind string) (string, error) {
	if kind != "scene" && kind != "prop" {
		return "", fmt.Errorf("%w: unsupported CM asset phase", ErrInvalid)
	}
	text := strings.TrimSpace(body)
	if text == "" {
		return "", nil
	}
	startMarker := "AI 调用六：独立场景提取"
	if kind == "prop" {
		startMarker = "AI 调用七：独立道具提取"
	}
	start := strings.Index(text, startMarker)
	if start < 0 {
		return text, nil
	}
	text = text[start:]
	if kind == "scene" {
		if end := strings.Index(text, "AI 调用七：独立道具提取"); end > 0 {
			text = text[:end]
		}
	}
	for _, marker := range []string{"9. 最终资产组装", "9.1 最终数据结构", "10. 本次精简与补全说明"} {
		if end := strings.Index(text, marker); end > 0 {
			text = text[:end]
		}
	}
	return strings.TrimSpace(text), nil
}

func normalizeCMAssetPhaseOutput(raw json.RawMessage, kind string) ([]NamedPrompt, error) {
	root, err := decodeDirectorObject(raw)
	if err != nil {
		return nil, err
	}
	arrayKey, label := "scenes", "scenes"
	nameFields := []string{"canonical_name", "scene_name", "name", "场景名称"}
	promptFields := []string{"scene_prompt", "prompt", "visual_anchor", "stable_environment", "description", "描述"}
	if kind == "prop" {
		arrayKey, label = "props", "props"
		nameFields = []string{"canonical_name", "prop_name", "name", "道具名称"}
		promptFields = []string{"prop_prompt", "prompt", "visual_anchor", "stable_appearance", "description", "描述"}
	}
	items := directorArray(root[arrayKey])
	seen := map[string]struct{}{}
	out := make([]NamedPrompt, 0, len(items))
	for index, item := range items {
		entry, ok := directorMap(item)
		if !ok {
			return nil, fmt.Errorf("%s[%d] 格式无效", label, index)
		}
		name := firstDirectorText(entry, nameFields...)
		prompt := firstDirectorText(entry, promptFields...)
		if name == "" {
			return nil, fmt.Errorf("%s[%d] 缺少名称", label, index)
		}
		if prompt == "" {
			return nil, fmt.Errorf("%s[%d] 缺少提示词", label, index)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("%s存在重复名称：%s", label, name)
		}
		seen[name] = struct{}{}
		out = append(out, NamedPrompt{Name: name, Prompt: prompt})
	}
	return out, nil
}

func mergeCMNamedPrompts(primary, fallback []NamedPrompt) []NamedPrompt {
	seen := map[string]struct{}{}
	out := make([]NamedPrompt, 0, len(primary)+len(fallback))
	for _, values := range [][]NamedPrompt{primary, fallback} {
		for _, item := range values {
			name := strings.TrimSpace(item.Name)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			item.Name = name
			out = append(out, item)
		}
	}
	return out
}
