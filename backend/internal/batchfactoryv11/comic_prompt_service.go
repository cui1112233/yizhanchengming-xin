package batchfactoryv11

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	comicQualityConstraint     = "comic_quality"
	comicRestrictionConstraint = "comic_restriction"
	comicNegativeConstraint    = "comic_negative"
)

const comicTextDensityRule = "【漫画文字密度】每页画面文字总量控制在45-70字，普通上限90字；一段过场旁白不要塞进一个框，旁白拆成多个8-24字的短框，整页最多3个旁白框。对白优先保留关键句，过长内容顺延下一页；不得为了塞下全文缩小字体或让文字遮挡人物和关键画面。"

const comicAssetIdentityRule = "【漫画资产身份表】comic_assets 是本书人物、场景和道具的唯一身份表。每条资产都用 asset_id 区分，name 是原文简称，formal_name 是正式名称，facts、appearance、relationships 是该资产的事实、外形和关系。生成 roles、role_stage、screen_prompt、image_prompt 时，必须按 asset_id 对应到同一条资产；人物在画面提示词中优先写“name（formal_name）”，formal_name 存在时不得只写“我”“他”“她”“夫人”等简称，不得把不同 asset_id 合并成一个人，也不得给资产临时编造新名字。"

type comicPromptProgressFunc func(completed, total int, pageLabel string)

type comicPromptProgressContextKey struct{}

type comicPromptConcurrencyContextKey struct{}

// WithComicPromptProgress lets the automation log expose the page currently
// being processed while prompt requests run in parallel.
func WithComicPromptProgress(ctx context.Context, report func(completed, total int, pageLabel string)) context.Context {
	if report == nil {
		return ctx
	}
	return context.WithValue(ctx, comicPromptProgressContextKey{}, comicPromptProgressFunc(report))
}

func comicPromptProgressFromContext(ctx context.Context) comicPromptProgressFunc {
	if ctx == nil {
		return nil
	}
	progress, _ := ctx.Value(comicPromptProgressContextKey{}).(comicPromptProgressFunc)
	return progress
}

// WithComicPromptConcurrency opts an automation run into bounded page-level
// concurrency. Direct service calls keep the historical serial behavior,
// which is useful for deterministic callers and tests.
func WithComicPromptConcurrency(ctx context.Context, concurrency int) context.Context {
	if concurrency < 1 {
		concurrency = 1
	}
	// 逐页提示词是纯文本请求，不占图片网关名额；实测 35 页在 5 路并发下要
	// 10~11 分钟，提到 8 路可以把排队轮次压下来。
	if concurrency > 8 {
		concurrency = 8
	}
	return context.WithValue(ctx, comicPromptConcurrencyContextKey{}, concurrency)
}

func comicPromptConcurrencyFromContext(ctx context.Context) int {
	if ctx == nil {
		return 1
	}
	if concurrency, ok := ctx.Value(comicPromptConcurrencyContextKey{}).(int); ok && concurrency > 0 {
		return concurrency
	}
	return 1
}

// comicPageStructureLock owns layout and text safety only. Art direction is
// supplied by the selected comic style and must never be fixed here.
const comicPageStructureLock = "【漫画页面结构约束】成图为{{image_aspect_ratio}}{{image_orientation}}的中文漫画页：保留清楚外留边和等宽格间距，格线清晰；根据剧情选择合适格数，阅读顺序固定为从上到下、同一行从左到右。旁白使用矩形框，对白使用圆形或椭圆气泡，简体中文短句不得遮挡脸部、手部或关键道具。具体线稿、上色、材质、人物比例和渲染方式只听所选画风预设，不得固定为其他画风。禁止水印、乱码、倾斜分格、无序拼贴、单张海报或后期拼图。"

func comicPromptTextForAspect(value, aspectRatio string) string {
	orientation := ImageAspectRatioOrientation(aspectRatio)
	value = strings.ReplaceAll(value, "1440×1920 或等比例 3:4", "与 "+aspectRatio+" 画幅匹配的高清分辨率")
	value = strings.ReplaceAll(value, "3:4 竖版", aspectRatio+orientation)
	value = strings.ReplaceAll(value, "3:4竖版", aspectRatio+orientation)
	value = strings.ReplaceAll(value, "{{image_orientation}}", orientation)
	value = strings.ReplaceAll(value, "{{image_aspect_ratio}}", aspectRatio)
	value = strings.ReplaceAll(value, "3:4", aspectRatio)
	// Preset bodies are authored for portrait pages. Landscape and square
	// canvases must not keep portrait-only wording.
	switch {
	case IsPortraitImageAspectRatio(aspectRatio):
		return value
	case aspectRatio == "1:1":
		return strings.NewReplacer(
			"竖版", "方形",
			"竖幅", "方形",
			"竖屏", "方形",
			"条漫", "漫画",
		).Replace(value)
	default:
		return strings.NewReplacer(
			"竖版长条", "横幅",
			"竖版", "横幅",
			"竖幅", "横幅",
			"竖屏", "横屏",
			"条漫", "漫画",
		).Replace(value)
	}
}

// BuildComicPromptContract remains available for previews of one selected page.
func BuildComicPromptContract(book Book, videos []Video, pageIndex int, snapshot DirectorSnapshot) (PromptContract, error) {
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return PromptContract{}, NewConfigRequiredError("请先打开漫画生成")
	}
	if pageIndex < 0 || pageIndex >= len(videos) {
		return PromptContract{}, fmt.Errorf("%w: 漫画页不存在", ErrInvalid)
	}
	video := videos[pageIndex]
	config := aiReasoningPromptConfig(snapshot.Effective)
	if !config.Comic.appliesTo(book) {
		return PromptContract{}, NewConfigRequiredError("请先在 AI 推理中选择漫画提示词")
	}
	// 开（assetImagesEnabled=true）：参考图承担资产视觉信息，提示词只带身份；
	// 关：沿用已存资产描述并合并进提示词。关闭时没有已存描述也不阻断。
	assetImages := AssetImagesEnabled(snapshot.Effective)
	assets := comicAssets(book.AssetRecords)
	if assetImages && len(assets) == 0 {
		return PromptContract{}, NewConfigRequiredError("请先提取漫画资产")
	}
	aspectRatio := EffectiveImageAspectRatio(snapshot.Effective)
	parts := []string{
		"你是 Batch Factory 的漫画页面导演。只输出合法 JSON，不要输出 Markdown、标题或解释。",
		`JSON 必须包含 screen_prompt、image_prompt、roles、scene_presets。screen_prompt 是本页导演说明；image_prompt 是可直接交给图片模型的最终中文提示词；roles 只列本页画面真实出现的人物；scene_presets 只列本页实际使用的场景。`,
		comicAssetIdentityRule,
		"本页必须生成一张完整" + aspectRatio + ImageAspectRatioOrientation(aspectRatio) + "漫画页，直接包含分格、角色、场景、动作、表情、旁白框、对话气泡及必须出现的中文文案。禁止后期拼图或另行叠字。",
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
	previousPage, nextPage := "", ""
	if pageIndex > 0 {
		previousPage = comicPageSourceText(videos[pageIndex-1])
	}
	if pageIndex+1 < len(videos) {
		nextPage = comicPageSourceText(videos[pageIndex+1])
	}
	localContext := comicLocalContext(videos, pageIndex)
	pagePreset = expandRecoveredComicPreset(comicPromptTextForAspect(pagePreset, aspectRatio), book, localContext, !assetImages)
	parts = append(parts,
		comicPromptTextForAspect(comicPageStructureLock, aspectRatio),
		"漫画提示词预设：\n"+pagePreset,
		comicTextDensityRule,
		"当前页内容边界：image_prompt 只能描述 current_page 当前页实际发生的画面、人物、动作、场景和当前页必须出现的文字。previous_page、next_page 和 source_text 仅用于理解人物与剧情连续性，禁止把其中的对白、旁白、事件、场景或文字复制到当前页。不得自行补写当前页没有的台词、旁白或剧情。",
	)
	if derivation := strings.TrimSpace(config.Comic.ImageDerivation.Body); derivation != "" {
		parts = append(parts, "图片描述词推导规则：\n"+comicPromptTextForAspect(derivation, aspectRatio))
	}
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
	payloadAssets := assets
	if assetImages {
		payloadAssets = comicAssetIdentities(book.AssetRecords)
	}
	payload, err := json.Marshal(map[string]any{
		"book_title":    book.Title,
		"source_text":   localContext,
		"page_label":    video.Label,
		"previous_page": previousPage,
		"current_page":  comicPageSourceText(video),
		"next_page":     nextPage,
		"comic_assets":  payloadAssets,
		"aspect_ratio":  aspectRatio,
	})
	if err != nil {
		return PromptContract{}, err
	}
	return PromptContract{SystemPrompt: strings.Join(parts, "\n\n"), UserPrompt: string(payload), Temperature: 0.25, MaxTokens: 4800}, nil
}

func expandRecoveredComicPreset(body string, book Book, localContext string, withDescriptions bool) string {
	assetLines := make([]string, 0, len(book.AssetRecords))
	for _, asset := range book.AssetRecords {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		name := strings.TrimSpace(asset.FormalName)
		if name == "" {
			name = strings.TrimSpace(asset.Name)
		}
		if name == "" {
			continue
		}
		if withDescriptions {
			// 没有已存描述时不拼，也不阻断流程。
			if description := strings.TrimSpace(asset.Prompt); description != "" {
				assetLines = append(assetLines, name+"："+description)
			}
			continue
		}
		// 开路径：参考图承担视觉信息，只保留身份名。
		assetLines = append(assetLines, name)
	}
	replacer := strings.NewReplacer(
		"{summary}", strings.TrimSpace(localContext),
		"{roles_str}", strings.Join(assetLines, "\n"),
		"{context_window}", "当前页及前后各一页",
		"{default_img_example}", "按本页剧情写明人物全名、动作、站位、环境光影、机位景别与必要文字 UI",
		"{{", "{",
		"}}", "}",
	)
	return replacer.Replace(body)
}

func comicLocalContext(videos []Video, pageIndex int) string {
	if pageIndex < 0 || pageIndex >= len(videos) {
		return ""
	}
	start, end := pageIndex-1, pageIndex+1
	if start < 0 {
		start = 0
	}
	if end >= len(videos) {
		end = len(videos) - 1
	}
	parts := make([]string, 0, end-start+1)
	for index := start; index <= end; index++ {
		text := strings.TrimSpace(comicPageSourceText(videos[index]))
		if text == "" {
			continue
		}
		label := "当前页"
		if index < pageIndex {
			label = "上一页"
		} else if index > pageIndex {
			label = "下一页"
		}
		parts = append(parts, label+"："+text)
	}
	return strings.Join(parts, "\n")
}

// BuildComicBookPromptContract asks the text model for every page in one call.
// This mirrors the snow-manga chain and keeps the full book context consistent.
func BuildComicBookPromptContract(book Book, videos []Video, snapshot DirectorSnapshot) (PromptContract, error) {
	if len(videos) == 0 {
		return PromptContract{}, fmt.Errorf("%w: 漫画页不存在", ErrInvalid)
	}
	contract, err := BuildComicPromptContract(book, videos, 0, snapshot)
	if err != nil {
		return PromptContract{}, err
	}
	contract.SystemPrompt = strings.Replace(contract.SystemPrompt,
		"JSON 必须包含 screen_prompt、image_prompt、roles、scene_presets。screen_prompt 是本页导演说明；image_prompt 是可直接交给图片模型的最终中文提示词；roles 只列本页画面真实出现的人物；scene_presets 只列本页实际使用的场景。",
		`只输出合法 JSON：{"pages":[...] }。pages 必须与输入页面一一对应且顺序不变。每项必须包含 video_id、role_stage、scene_presets、screen_prompt、image_prompt、roles；video_id 原样返回；role_stage 为简称到资产正式名称的映射；roles 只列画面真实出场人物；scene_presets 只列实际使用场景。轻量分镜版允许 image_prompt 为空，其他版本不得为空。`, 1)
	pages := make([]map[string]string, 0, len(videos))
	for _, video := range videos {
		pages = append(pages, map[string]string{
			"video_id": video.ID,
			"label":    video.Label,
			"content":  comicPageSourceText(video),
		})
	}
	payloadAssets := comicAssets(book.AssetRecords)
	if AssetImagesEnabled(snapshot.Effective) {
		payloadAssets = comicAssetIdentities(book.AssetRecords)
	}
	payload, err := json.Marshal(map[string]any{
		"book_title":    book.Title,
		"source_text":   book.SourceText,
		"ordered_pages": pages,
		"comic_assets":  payloadAssets,
		"aspect_ratio":  EffectiveImageAspectRatio(snapshot.Effective),
	})
	if err != nil {
		return PromptContract{}, err
	}
	contract.UserPrompt = string(payload)
	contract.MaxTokens = len(videos) * 1200
	if contract.MaxTokens < 3200 {
		contract.MaxTokens = 3200
	}
	return contract, nil
}

func comicPageSourceText(video Video) string {
	if video.ComicSourceSegment != nil && video.ComicSourceSegment.SourceText != "" {
		return video.ComicSourceSegment.SourceText
	}
	return strings.TrimSpace(video.VideoPrompt)
}

// comicAssetIdentities keeps only the identity fields used for role mapping.
// It is the image-first path: visual details come from the reference images,
// so stored descriptions must not be repeated in the prompt.
func comicAssetIdentities(records []BookAsset) []map[string]string {
	assets := make([]map[string]string, 0, len(records))
	for _, asset := range records {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		name := strings.TrimSpace(asset.Name)
		if name == "" {
			continue
		}
		assetID := strings.TrimSpace(asset.ID)
		if assetID == "" {
			assetID = name
		}
		assets = append(assets, map[string]string{
			"asset_id":    assetID,
			"kind":        strings.TrimSpace(asset.Kind),
			"name":        name,
			"formal_name": strings.TrimSpace(asset.FormalName),
		})
	}
	return assets
}

func comicAssets(records []BookAsset) []map[string]string {
	assets := make([]map[string]string, 0, len(records))
	for _, asset := range records {
		if strings.TrimSpace(asset.Source) != "comic" {
			continue
		}
		name, prompt := strings.TrimSpace(asset.Name), strings.TrimSpace(asset.Prompt)
		if name == "" || prompt == "" {
			continue
		}
		assetID := strings.TrimSpace(asset.ID)
		if assetID == "" {
			assetID = name
		}
		assets = append(assets, map[string]string{
			"asset_id":      assetID,
			"kind":          strings.TrimSpace(asset.Kind),
			"name":          name,
			"formal_name":   strings.TrimSpace(asset.FormalName),
			"facts":         strings.TrimSpace(asset.Facts),
			"appearance":    strings.TrimSpace(asset.Appearance),
			"relationships": strings.TrimSpace(asset.Relationships),
			"prompt":        prompt,
		})
	}
	return assets
}

// RunComicPromptExtraction regenerates complete-page comic prompts one page at
// a time so each accepted result can be persisted and displayed immediately.
// The existing director revision and video prompts are deliberately untouched.
func (s *DirectorService) RunComicPromptExtraction(ctx context.Context, owner, batchID, bookID string) ([]Video, error) {
	return s.runComicPromptExtraction(ctx, owner, batchID, bookID, false)
}

// RunMissingComicPromptExtraction fills only pages that do not already have a
// complete prompt pair. It is used by automatic recovery so accepted edits on
// completed pages are never replaced.
func (s *DirectorService) RunMissingComicPromptExtraction(ctx context.Context, owner, batchID, bookID string) ([]Video, error) {
	return s.runComicPromptExtraction(ctx, owner, batchID, bookID, true)
}

func (s *DirectorService) runComicPromptExtraction(ctx context.Context, owner, batchID, bookID string, onlyMissing bool) ([]Video, error) {
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
	if segmentStore, ok := s.Store.(ComicSourceSegmentStore); ok {
		if segmentation, segmentErr := segmentStore.GetComicSourceSegmentation(ctx, owner, batchID, bookID); segmentErr == nil {
			segments := make(map[string]ComicSourceSegment, len(segmentation.Segments))
			for _, segment := range segmentation.Segments {
				segments[segment.VideoID] = segment
			}
			for index := range book.Videos {
				if segment, exists := segments[book.Videos[index].ID]; exists {
					copy := segment
					book.Videos[index].ComicSourceSegment = &copy
				}
			}
		}
	}
	if len(book.Videos) == 0 {
		return nil, fmt.Errorf("%w: 请先生成漫画分页", ErrConflict)
	}
	// 漫画分页和漫画画面提示词不依赖配音时长。音频时长只在视频
	// 分镜/视频生产阶段用于时间轴规划；这里如果复用 snapshotForBook，
	// 开启音频规划但尚未生成配音时会把整条漫画批量流程错误拦住。
	snapshot, err := snapshotForAssetExtraction(batch, book)
	if err != nil {
		return nil, err
	}
	if normalizeProductionMode(snapshot.Effective) != "comic" {
		return nil, fmt.Errorf("%w: 当前书为视频生成模式", ErrConflict)
	}
	targets := book.Videos
	if onlyMissing {
		targets = missingComicPromptVideos(book.Videos, snapshot)
		if len(targets) == 0 {
			return []Video{}, nil
		}
	}
	// Book-level concurrency does not help when a batch contains one book.
	// Keep one request per page, but process at most five pages in parallel so
	// immediate execution does not spend the whole run on a serial chain.
	promptConcurrency := comicPromptConcurrencyFromContext(ctx)
	progress := comicPromptProgressFromContext(ctx)
	if progress != nil {
		progress(0, len(targets), "准备漫画页提示词")
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan Video)
	var workers sync.WaitGroup
	var completed atomic.Int32
	var resultMu sync.Mutex
	updated := make([]Video, 0, len(targets))
	var firstErr error
	process := func(target Video) {
		pageIndex := -1
		for index, page := range book.Videos {
			if page.ID == target.ID {
				pageIndex = index
				break
			}
		}
		if pageIndex < 0 {
			resultMu.Lock()
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: 漫画页不存在", ErrNotFound)
				cancel()
			}
			resultMu.Unlock()
			return
		}
		contract, err := BuildComicPromptContract(book, book.Videos, pageIndex, snapshot)
		if err == nil {
			completion, completionErr := s.Provider.Complete(workCtx, TextCompletionRequest{SystemPrompt: contract.SystemPrompt, UserPrompt: contract.UserPrompt, Temperature: contract.Temperature, MaxTokens: contract.MaxTokens})
			if isTruncatedDirectorOutput(completionErr) {
				completion, completionErr = s.Provider.Complete(workCtx, TextCompletionRequest{
					SystemPrompt:        contract.SystemPrompt + "\n上一轮输出因长度截断。保留当前页必要的分格、人物、动作、场景和关键文字；使用更紧凑的短语，screen_prompt 与 image_prompt 都不要重复描述；现在只输出一个完整、合法的 JSON 对象，不要解释，不要 Markdown。",
					UserPrompt:          contract.UserPrompt,
					Temperature:         0.2,
					MaxTokens:           contract.MaxTokens,
					DisableJSONResponse: true,
				})
			}
			err = completionErr
			if err == nil {
				var parsed map[string]comicPromptResult
				parsed, err = parseComicBookPromptResults(completion, []Video{target}, snapshot)
				if err == nil {
					var page []Video
					page, err = s.persistComicPromptResults(workCtx, owner, batchID, bookID, []Video{target}, parsed)
					if err == nil {
						resultMu.Lock()
						updated = append(updated, page...)
						resultMu.Unlock()
					}
				}
			}
		}
		if err != nil {
			resultMu.Lock()
			if firstErr == nil {
				firstErr = err
				cancel()
			}
			resultMu.Unlock()
			return
		}
		finished := int(completed.Add(1))
		if progress != nil {
			progress(finished, len(targets), target.Label)
		}
	}
	for index := 0; index < promptConcurrency; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-workCtx.Done():
					return
				case target, ok := <-jobs:
					if !ok {
						return
					}
					process(target)
				}
			}
		}()
	}
	for _, target := range targets {
		select {
		case jobs <- target:
		case <-workCtx.Done():
			break
		}
		if workCtx.Err() != nil {
			break
		}
	}
	close(jobs)
	workers.Wait()
	resultMu.Lock()
	err = firstErr
	result := append([]Video(nil), updated...)
	resultMu.Unlock()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, right := len(book.Videos), len(book.Videos)
		for index, page := range book.Videos {
			if page.ID == result[i].ID {
				left = index
			}
			if page.ID == result[j].ID {
				right = index
			}
		}
		return left < right
	})
	return result, nil
}

func missingComicPromptVideos(videos []Video, snapshot DirectorSnapshot) []Video {
	light := aiReasoningPromptConfig(snapshot.Effective).Comic.ID == "batch-comic-page-light"
	missing := make([]Video, 0, len(videos))
	for _, video := range videos {
		if strings.TrimSpace(video.ComicPrompt) == "" || (!light && strings.TrimSpace(video.ComicImagePrompt) == "") {
			missing = append(missing, video)
		}
	}
	return missing
}

func (s *DirectorService) persistComicPromptResults(ctx context.Context, owner, batchID, bookID string, videos []Video, results map[string]comicPromptResult) ([]Video, error) {
	updated := make([]Video, 0, len(videos))
	for _, video := range videos {
		parsed := results[video.ID]
		encodedScreen, _ := json.Marshal(parsed.ScreenPrompt)
		encodedImage, _ := json.Marshal(parsed.ImagePrompt)
		encodedRoles, _ := json.Marshal(parsed.Roles)
		encodedRoleStage, _ := json.Marshal(parsed.RoleStage)
		encodedScenes, _ := json.Marshal(parsed.ScenePresets)
		result, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: video.ID}, SettingsUpdate{
			Patch: SettingsPatch{
				"comicPrompt": encodedScreen, "comicImagePrompt": encodedImage,
				"comicRoles": encodedRoles, "comicRoleStage": encodedRoleStage, "comicScenePresets": encodedScenes,
			}, ExpectedRevision: video.Revision,
		})
		if err != nil {
			return nil, err
		}
		video.ComicPrompt = parsed.ScreenPrompt
		video.ComicImagePrompt = parsed.ImagePrompt
		video.ComicRoles = parsed.Roles
		video.ComicRoleStage = parsed.RoleStage
		video.ComicScenePresets = parsed.ScenePresets
		video.Revision = result.Revision
		video.SettingsState = SettingsState{Patch: result.Patch, Revision: result.Revision}
		updated = append(updated, video)
	}
	return updated, nil
}

type comicPromptResult struct {
	VideoID      string            `json:"video_id"`
	RoleStage    map[string]string `json:"role_stage"`
	ScreenPrompt string            `json:"screen_prompt"`
	ImagePrompt  string            `json:"image_prompt"`
	Roles        []string          `json:"roles"`
	ScenePresets []string          `json:"scene_presets"`
}

func parseComicBookPromptResults(completion string, videos []Video, snapshot DirectorSnapshot) (map[string]comicPromptResult, error) {
	raw, err := ParseDirectorJSON(strings.TrimSpace(completion))
	if err != nil {
		if len(videos) != 1 {
			return nil, fmt.Errorf("%w: 整书漫画提示词不是合法 JSON", ErrInvalid)
		}
		legacy := parseComicPromptResult(completion)
		if legacy.ScreenPrompt == "" {
			return nil, fmt.Errorf("%w: 漫画提示词为空", ErrInvalid)
		}
		legacy.VideoID = videos[0].ID
		return map[string]comicPromptResult{legacy.VideoID: legacy}, nil
	}
	var envelope struct {
		Pages []comicPromptResult `json:"pages"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("%w: 漫画页结果无法解析", ErrInvalid)
	}
	if len(envelope.Pages) == 0 && len(videos) == 1 {
		var legacy comicPromptResult
		if json.Unmarshal(raw, &legacy) == nil && strings.TrimSpace(legacy.ScreenPrompt) != "" {
			legacy.VideoID = videos[0].ID
			envelope.Pages = []comicPromptResult{legacy}
		}
	}
	if len(envelope.Pages) != len(videos) {
		return nil, fmt.Errorf("%w: 漫画页数量与分页不一致", ErrInvalid)
	}
	light := aiReasoningPromptConfig(snapshot.Effective).Comic.ID == "batch-comic-page-light"
	wanted := make(map[string]struct{}, len(videos))
	for _, video := range videos {
		wanted[video.ID] = struct{}{}
	}
	results := make(map[string]comicPromptResult, len(videos))
	for _, page := range envelope.Pages {
		page.VideoID = strings.TrimSpace(page.VideoID)
		page.ScreenPrompt = strings.TrimSpace(page.ScreenPrompt)
		page.ImagePrompt = strings.TrimSpace(page.ImagePrompt)
		page.Roles = normalizedStringList(page.Roles)
		page.ScenePresets = normalizedStringList(page.ScenePresets)
		page.RoleStage = normalizedStringMap(page.RoleStage)
		if _, ok := wanted[page.VideoID]; !ok || results[page.VideoID].VideoID != "" {
			return nil, fmt.Errorf("%w: 漫画页 video_id 无效或重复", ErrInvalid)
		}
		if page.ScreenPrompt == "" || (!light && page.ImagePrompt == "") {
			return nil, fmt.Errorf("%w: 漫画提示词为空", ErrInvalid)
		}
		results[page.VideoID] = page
	}
	return results, nil
}

func parseComicPromptResult(completion string) comicPromptResult {
	completion = strings.TrimSpace(completion)
	if completion == "" {
		return comicPromptResult{}
	}
	var result comicPromptResult
	if raw, err := ParseDirectorJSON(completion); err == nil {
		if err := json.Unmarshal(raw, &result); err == nil {
			result.ScreenPrompt = strings.TrimSpace(result.ScreenPrompt)
			result.ImagePrompt = strings.TrimSpace(result.ImagePrompt)
			result.Roles = normalizedStringList(result.Roles)
			result.ScenePresets = normalizedStringList(result.ScenePresets)
			if result.ScreenPrompt != "" && result.ImagePrompt != "" {
				return result
			}
		}
	}
	// Older providers returned one final prompt. Keep those books usable while
	// new presets return the structured pair above.
	return comicPromptResult{ScreenPrompt: completion, ImagePrompt: completion}
}

func normalizedStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizedStringMap(values map[string]string) map[string]string {
	normalized := make(map[string]string, len(values))
	for key, value := range values {
		if key, value = strings.TrimSpace(key), strings.TrimSpace(value); key != "" && value != "" {
			normalized[key] = value
		}
	}
	return normalized
}
