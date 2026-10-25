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

// comicPageStructureLock owns layout and text safety only. Art direction is
// supplied by the selected comic style and must never be fixed here.
const comicPageStructureLock = "【漫画页面结构约束】成图为{{image_aspect_ratio}}画幅中文漫画页：保留清楚外留边和等宽格间距，格线清晰；根据剧情选择合适格数，阅读顺序固定为从上到下、同一行从左到右。旁白使用矩形框，对白使用圆形或椭圆气泡，简体中文短句不得遮挡脸部、手部或关键道具。具体线稿、上色、材质、人物比例和渲染方式只听所选画风预设，不得固定为其他画风。禁止水印、乱码、倾斜分格、无序拼贴、单张海报或后期拼图。"

func comicPromptTextForAspect(value, aspectRatio string) string {
	value = strings.ReplaceAll(value, "1440×1920 或等比例 3:4", "与 "+aspectRatio+" 画幅匹配的高清分辨率")
	value = strings.ReplaceAll(value, "3:4 竖版", aspectRatio+" 画幅")
	value = strings.ReplaceAll(value, "3:4竖版", aspectRatio+"画幅")
	value = strings.ReplaceAll(value, "{{image_aspect_ratio}}", aspectRatio)
	return strings.ReplaceAll(value, "3:4", aspectRatio)
}

// BuildComicPromptContract creates a complete page prompt. It never receives
// video motion text, so a comic page cannot inherit timing or camera movement.
func BuildComicPromptContract(book Book, video Video, snapshot DirectorSnapshot) (PromptContract, error) {
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return PromptContract{}, fmt.Errorf("%w: 请先打开漫画生成", ErrConflict)
	}
	config := aiReasoningPromptConfig(snapshot.Effective)
	if !config.Comic.appliesTo(book) {
		return PromptContract{}, fmt.Errorf("%w: 请先在 AI 推理中选择漫画提示词", ErrConflict)
	}
	assets := comicAssets(book.AssetRecords)
	if len(assets) == 0 {
		return PromptContract{}, fmt.Errorf("%w: 请先提取漫画资产", ErrConflict)
	}
	aspectRatio := EffectiveImageAspectRatio(snapshot.Effective)
	parts := []string{
		"你是 Batch Factory 的漫画提示词生成器。只输出一条可直接交给图片模型生成完整漫画页的中文提示词，不要输出 Markdown、JSON、标题或解释。",
		"该提示词必须生成一张完整" + aspectRatio + "画幅条漫页，直接包含分格、角色、场景、动作、表情、旁白框、对话气泡及必须出现的中文文案。禁止后期拼图或另行叠字。",
		"不得包含视频时长、运镜、镜头推进、音频、配音或视频生成指令。",
	}
	if style := strings.TrimSpace(config.Comic.Style.Body); style != "" {
		parts = append(parts, "本书画风预设：\n"+comicPromptTextForAspect(style, aspectRatio))
	}
	if prefix := strings.TrimSpace(config.Comic.ImagePromptPrefix); prefix != "" {
		parts = append(parts, "用户图片描述词全局前缀：\n"+comicPromptTextForAspect(prefix, aspectRatio))
	}
	pagePreset := config.Comic.body()
	if strings.TrimSpace(config.Comic.Style.Body) != "" {
		pagePreset = stripLegacyComicStyle(pagePreset)
	}
	parts = append(parts,
		comicPromptTextForAspect(comicPageStructureLock, aspectRatio),
		"漫画提示词预设：\n"+comicPromptTextForAspect(pagePreset, aspectRatio),
	)
	for _, category := range []string{comicQualityConstraint, comicRestrictionConstraint, comicNegativeConstraint} {
		for _, selection := range config.Comic.Constraints {
			if constraintCategory(selection) == category && strings.TrimSpace(selection.Body) != "" {
				body := selection.Body
				if strings.TrimSpace(config.Comic.Style.Body) != "" {
					body = stripLegacyComicStyle(body)
				}
				parts = append(parts, comicPromptTextForAspect(body, aspectRatio))
			}
		}
	}
	payload, err := json.Marshal(map[string]any{
		"book_title":   book.Title,
		"source_text":  book.SourceText,
		"page_label":   video.Label,
		"comic_assets": assets,
		"aspect_ratio": aspectRatio,
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
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return nil, fmt.Errorf("%w: 当前书为视频生成模式", ErrConflict)
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
