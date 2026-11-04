package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func isComicAssetExtractionPresetID(id string) bool {
	switch strings.TrimSpace(id) {
	case "batch-comic-assets", "vip-smart-character-derive":
		return true
	default:
		return false
	}
}

func normalizeLegacyComicBookAsset(asset BookAsset) BookAsset {
	if asset.Source == "manual" && isComicAssetExtractionPresetID(asset.ExtractionPresetID) {
		asset.Source = "comic"
	}
	return asset
}

// BuildComicAssetExtractionContract turns the selected backend-owned comic
// asset preset into the one text-model call that establishes comic continuity.
func BuildComicAssetExtractionContract(book Book, snapshot DirectorSnapshot) (PromptContract, error) {
	if strings.TrimSpace(book.SourceText) == "" {
		return PromptContract{}, fmt.Errorf("%w: source text is required", ErrInvalid)
	}
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return PromptContract{}, fmt.Errorf("%w: 请先打开漫画生成", ErrConflict)
	}
	preset := aiReasoningPromptConfig(snapshot.Effective).Comic.Assets
	if strings.TrimSpace(preset.Body) == "" {
		// Missing selection is a configuration problem, not a stale revision.
		// Keep the message free of any sentinel prefix so the UI can show the
		// exact action the user must take.
		return PromptContract{}, NewConfigRequiredError("请在资产设置中选择漫画资产提示词")
	}
	payloadValues := map[string]any{"book_id": book.ID, "title": book.Title, "source_text": book.SourceText, "production": "comic"}
	if bookGender := bookGenderForPrompt(book); bookGender != "" {
		payloadValues["book_gender"] = bookGender
	}
	payload, err := json.Marshal(payloadValues)
	if err != nil {
		return PromptContract{}, err
	}
	system := `你是漫画制作资产提取器。只从小说原文提取实际出现的人物、场景与关键道具，并为每项写可直接用于中文漫画图片生成的稳定视觉提示词。
不要生成分镜、镜头、视频、改写正文或解释。不要杜撰原文没有出现的重要资产。
只输出一个合法 JSON 对象，格式严格为：
{"characters":[{"name":"人物名","formal_name":"正式人物名单中的标准名","facts":"人物事实","appearance":"人物外形","relationships":"人物关系","prompt":"只合并人物事实与人物外形的完整漫画视觉提示词"}],"scenes":[{"name":"场景名","prompt":"空间和氛围的可见特征"}],"props":[{"name":"道具名","prompt":"材质、外观和状态"}]}
	每个数组可以为空；每个元素必须同时有非空 name 与 prompt。`
	system += "\n人物关系只写入 relationships 字段；不得把夫妻、亲子、主仆、朋友等关系文字写进 character prompt。character prompt 只能描述身份事实和可见外形，不写人物之间的关系。"
	if genderRule := firstPersonGenderRule(book); genderRule != "" {
		system += "\n\n" + genderRule
	}
	system += "\n\n当前漫画资产提示词：\n" + strings.TrimSpace(preset.Body)
	return PromptContract{SystemPrompt: system, UserPrompt: string(payload), Temperature: 0.2, MaxTokens: 12000}, nil
}

// RunComicAssetExtraction keeps comic assets separate from video assets. The
// workflow invokes it from the existing asset-settings action when comic mode
// is active, so the user has one clear place to configure and execute it.
func (s *DirectorService) RunComicAssetExtraction(ctx context.Context, owner, batchID, bookID string) ([]BookAsset, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return nil, err
	}
	book, err = s.comicProductionBook(ctx, owner, batchID, book)
	if err != nil {
		return nil, err
	}
	snapshot, err := snapshotForAssetExtraction(batch, book)
	if err != nil {
		return nil, err
	}
	contract, err := BuildComicAssetExtractionContract(book, snapshot)
	if err != nil {
		return nil, err
	}
	completion, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
	if err != nil {
		return nil, err
	}
	raw, err := ParseDirectorJSON(completion)
	if err != nil {
		return nil, retryableModelOutputError(err)
	}
	assets, err := NormalizeAssetExtractionOutput(raw)
	if err != nil {
		return nil, retryableModelOutputError(err)
	}
	// Keep comic assets aligned with video assets: if first-person "我" has no
	// explicit gender in the source or model output, preserve the resolved
	// book-frequency classification as a durable visual anchor.
	assets = applyBookGenderFallback(book, assets)
	return s.Store.PersistComicExtractedBookAssets(ctx, owner, book, snapshot, assets)
}

func comicBookAssets(book Book, snapshot DirectorSnapshot, assets DirectorAssets) []BookAsset {
	preset := aiReasoningPromptConfig(snapshot.Effective).Comic.Assets
	rows := make([]BookAsset, 0, len(assets.Characters)+len(assets.Scenes)+len(assets.Props))
	appendRows := func(kind string, values []NamedPrompt) {
		for _, value := range values {
			name, prompt := strings.TrimSpace(value.Name), strings.TrimSpace(value.Prompt)
			if name == "" || prompt == "" {
				continue
			}
			rows = append(rows, BookAsset{BatchID: book.BatchID, BookID: book.ID, Kind: kind, Name: name, FormalName: strings.TrimSpace(value.FormalName), Facts: strings.TrimSpace(value.Facts), Appearance: strings.TrimSpace(value.Appearance), Relationships: strings.TrimSpace(value.Relationships), Prompt: mergeVisualAssetPrompt(value.Facts, value.Appearance, prompt, value.Relationships), Source: "comic", ExtractionPresetID: strings.TrimSpace(preset.ID), ExtractionPresetVersion: int64(preset.Version)})
		}
	}
	appendRows("character", assets.Characters)
	appendRows("scene", assets.Scenes)
	appendRows("prop", assets.Props)
	return rows
}
