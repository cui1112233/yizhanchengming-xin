package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	comicQualityConstraint     = "comic_quality"
	comicRestrictionConstraint = "comic_restriction"
	comicNegativeConstraint    = "comic_negative"
)

// BuildComicPromptContract creates a complete page prompt. It never receives
// video motion text, so a comic page cannot inherit timing or camera movement.
func BuildComicPromptContract(book Book, video Video, snapshot DirectorSnapshot) (PromptContract, error) {
	config := aiReasoningPromptConfig(snapshot.Effective)
	if !config.Comic.appliesTo(book) {
		return PromptContract{}, fmt.Errorf("%w: 请先在 AI 推理中选择漫画提示词", ErrConflict)
	}
	assets := comicAssets(book.AssetRecords)
	if len(assets) == 0 {
		return PromptContract{}, fmt.Errorf("%w: 请先提取漫画资产", ErrConflict)
	}
	parts := []string{
		"你是 Batch Factory 的漫画提示词生成器。只输出一条可直接交给图片模型生成完整漫画页的中文提示词，不要输出 Markdown、JSON、标题或解释。",
		"该提示词必须生成一张完整3:4竖版条漫页，直接包含分格、角色、场景、动作、表情、旁白框、对话气泡及必须出现的中文文案。禁止后期拼图或另行叠字。",
		"不得包含视频时长、运镜、镜头推进、音频、配音或视频生成指令。",
		"漫画提示词预设：\n" + config.Comic.body(),
	}
	for _, category := range []string{comicQualityConstraint, comicRestrictionConstraint, comicNegativeConstraint} {
		for _, selection := range config.Comic.Constraints {
			if constraintCategory(selection) == category && strings.TrimSpace(selection.Body) != "" {
				parts = append(parts, selection.Body)
			}
		}
	}
	payload, err := json.Marshal(map[string]any{
		"book_title": book.Title,
		"source_text": book.SourceText,
		"page_label":  video.Label,
		"comic_assets": assets,
		"aspect_ratio": "3:4",
	})
	if err != nil {
		return PromptContract{}, err
	}
	return PromptContract{SystemPrompt: strings.Join(parts, "\n\n"), UserPrompt: string(payload), Temperature: 0.25, MaxTokens: 3200}, nil
}

func comicAssets(records []BookAsset) []map[string]string {
	assets := make([]map[string]string, 0, len(records))
	for _, asset := range records {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		if name, prompt := strings.TrimSpace(asset.Name), strings.TrimSpace(asset.Prompt); name != "" && prompt != "" {
			assets = append(assets, map[string]string{"kind": asset.Kind, "name": name, "prompt": prompt})
		}
	}
	return assets
}

// RunComicPromptExtraction regenerates only the complete-page comic prompts.
// The existing director revision and video prompts are deliberately untouched.
func (s *DirectorService) RunComicPromptExtraction(ctx context.Context, owner, batchID, bookID string) ([]Video, error) {
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
	book, err = s.productionBook(ctx, owner, batchID, book)
	if err != nil {
		return nil, err
	}
	if len(book.Videos) == 0 {
		return nil, fmt.Errorf("%w: 请先提取视频分镜", ErrConflict)
	}
	snapshot, err := snapshotForBook(batch, book)
	if err != nil {
		return nil, err
	}
	updated := make([]Video, 0, len(book.Videos))
	for _, video := range book.Videos {
		contract, err := BuildComicPromptContract(book, video, snapshot)
		if err != nil {
			return nil, err
		}
		prompt, err := s.Provider.Complete(ctx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
		if err != nil {
			return nil, err
		}
		prompt = strings.TrimSpace(prompt)
		if prompt == "" {
			return nil, fmt.Errorf("%w: 漫画提示词为空", ErrInvalid)
		}
		encoded, _ := json.Marshal(prompt)
		result, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: video.ID}, SettingsUpdate{
			Patch: SettingsPatch{"comicPrompt": encoded}, ExpectedRevision: video.Revision,
		})
		if err != nil {
			return nil, err
		}
		video.ComicPrompt = prompt
		video.Revision = result.Revision
		video.SettingsState = SettingsState{Patch: result.Patch, Revision: result.Revision}
		updated = append(updated, video)
	}
	return updated, nil
}
