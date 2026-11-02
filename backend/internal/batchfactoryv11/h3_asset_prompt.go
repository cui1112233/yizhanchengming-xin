package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	h3CharacterRendererKey = "h3-character-normal"
	h3SceneRendererKey     = "h3-scene-normal"
	h3AssetsRendererKey    = "h3-assets-full"
	sdAssetsRendererKey    = "sd-character-scene"
	cmAssetsRendererID     = "batch-character-meta"
)

func usesH3CharacterRenderer(selection PresetSnapshot) bool {
	return strings.EqualFold(strings.TrimSpace(selection.Key), h3CharacterRendererKey)
}

func usesH3SceneRenderer(selection PresetSnapshot) bool {
	return strings.EqualFold(strings.TrimSpace(selection.Key), h3SceneRendererKey)
}

func usesH3AssetRenderer(selection PresetSnapshot) bool {
	return strings.EqualFold(strings.TrimSpace(selection.Key), h3AssetsRendererKey)
}

func usesCMAssetRenderer(selection PresetSnapshot) bool {
	return strings.EqualFold(strings.TrimSpace(selection.ID), cmAssetsRendererID)
}

// usesCombinedAssetRenderer marks presets whose single response already
// contains the complete asset result.  These presets must never fall through
// to the legacy per-character/per-scene compiler, otherwise one user choice
// becomes several hidden AI calls.
func usesCombinedAssetRenderer(selection PresetSnapshot) bool {
	key := strings.TrimSpace(selection.Key)
	return strings.EqualFold(key, h3AssetsRendererKey) || strings.EqualFold(key, sdAssetsRendererKey) || usesCMAssetRenderer(selection)
}

func h3AssetSelection(snapshot DirectorSnapshot, book Book, kind string) (PresetSnapshot, bool) {
	assets := aiReasoningPromptConfig(snapshot.Effective).Assets
	switch kind {
	case "character":
		if assets.appliesTo(book, assets.Extraction) && usesH3AssetRenderer(assets.Extraction) {
			return assets.Extraction, true
		}
		return assets.Character, assets.appliesTo(book, assets.Character) && usesH3CharacterRenderer(assets.Character)
	case "scene":
		return assets.Scene, assets.appliesTo(book, assets.Scene) && usesH3SceneRenderer(assets.Scene)
	default:
		return PresetSnapshot{}, false
	}
}

// BuildH3AssetPromptContract turns one extracted asset into its stable H3
// production prompt. The asset library is passed as Scene Memory so independent
// calls retain names, relationships and the spatial vocabulary established by
// the first extraction pass.
func BuildH3AssetPromptContract(book Book, assets DirectorAssets, kind string, target NamedPrompt, preset PresetSnapshot) (PromptContract, error) {
	if strings.TrimSpace(book.SourceText) == "" || strings.TrimSpace(target.Name) == "" {
		return PromptContract{}, fmt.Errorf("%w: source text and asset name are required", ErrInvalid)
	}
	label := "人物"
	maxTokens := 1800
	if kind == "scene" {
		label, maxTokens = "场景", 1600
	} else if kind != "character" {
		return PromptContract{}, fmt.Errorf("%w: unsupported H3 asset kind", ErrInvalid)
	}
	system := "你是 H3 " + label + "提示词编译器。你只为当前一个" + label + "输出稳定、可直接用于图片和视频模型的中文提示词。"
	system += "\n必须保持原文事实、名称、身份、年龄阶段、关系和已有资产记忆一致；不得编造原文没有的关键设定。"
	if kind == "character" {
		system += "\n人物提示词必须是可直接用于生图和视频一致性的完整视觉定义，依次写清：身份与气质、体型与年龄感、脸型与五官、发型与发色、肤色、服装分层、材质与颜色、鞋履配饰、姿态表情、视觉识别点与必须稳定保持的连续性特征。"
		system += "\n人物 prompt 不少于 120 个中文字；禁止只写“年轻女性、黑长发、穿西装”之类简化标签。"
	}
	system += "\n只输出合法 JSON：{\"prompt\":\"完整中文提示词\"}。不得输出 Markdown、解释或其它字段。"
	if body := strings.TrimSpace(preset.Body); body != "" {
		system += "\n\n当前 H3 " + label + "输出规则：\n" + body
	}
	payload, err := json.MarshalIndent(map[string]any{
		"book_id":       book.ID,
		"title":         book.Title,
		"source_text":   book.SourceText,
		"asset_kind":    kind,
		"current_asset": target,
		"scene_memory": map[string]any{
			"characters": assets.Characters,
			"scenes":     assets.Scenes,
			"props":      assets.Props,
		},
	}, "", "  ")
	if err != nil {
		return PromptContract{}, err
	}
	return PromptContract{SystemPrompt: system, UserPrompt: string(payload), Temperature: 0.2, MaxTokens: maxTokens}, nil
}

func normalizeH3AssetPrompt(raw json.RawMessage, kind string) (string, error) {
	root, err := decodeDirectorObject(raw)
	if err != nil {
		return "", err
	}
	prompt := firstDirectorText(root, "prompt", "asset_prompt")
	if prompt == "" {
		return "", fmt.Errorf("H3 资产提示词为空")
	}
	if kind == "character" && utf8.RuneCountInString(prompt) < 120 {
		return "", fmt.Errorf("H3 人物提示词过于简化：必须至少 120 个字符")
	}
	return prompt, nil
}

func (s *DirectorService) compileH3AssetPrompts(ctx context.Context, book Book, snapshot DirectorSnapshot, assets DirectorAssets) (DirectorAssets, error) {
	compile := func(kind string, values []NamedPrompt) error {
		preset, enabled := h3AssetSelection(snapshot, book, kind)
		if !enabled {
			return nil
		}
		for index := range values {
			contract, err := BuildH3AssetPromptContract(book, assets, kind, values[index], preset)
			if err != nil {
				return err
			}
			completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
			if err != nil {
				return err
			}
			raw, err := ParseDirectorJSON(completion)
			if err != nil {
				return retryableModelOutputError(err)
			}
			prompt, err := normalizeH3AssetPrompt(raw, kind)
			if err != nil {
				return retryableModelOutputError(err)
			}
			values[index].Prompt = prompt
		}
		return nil
	}
	if preset, enabled := h3AssetSelection(snapshot, book, "character"); enabled {
		characters, err := s.compileH3CharacterBatch(ctx, book, preset, assets)
		if err != nil {
			return DirectorAssets{}, err
		}
		assets.Characters = characters
	}
	if err := compile("scene", assets.Scenes); err != nil {
		return DirectorAssets{}, err
	}
	return assets, nil
}
