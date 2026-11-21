package batchfactoryv11

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	promptcatalog "qiantie/backend/internal/prompt"
)

// videoPromptRepairMaxRounds 与漫画页的自动去敏保持一致：每个视频、每个换
// 开头变体在内容审核拒绝后最多自动改写两轮提示词；仍未通过就转人工，避免
// 用同一份提示词反复提交、重复扣费。生产实测 1 轮不足以过审，故为 2 轮。
const videoPromptRepairMaxRounds = 2

const (
	videoPromptRepairCountsKey  = "videoPromptRepairCounts"
	videoPromptRepairMarkersKey = "videoPromptRepairMarkers"
	openingVariantPromptsKey    = "openingVariantPrompts"
)

// ErrVideoPromptRepairExhausted 表示自动去敏轮次已用完。调用方不得再用原
// 提示词重新提交，必须等待人工调整提示词或参考图。
var ErrVideoPromptRepairExhausted = errors.New("视频提示词被内容审核拒绝，自动去敏改写 2 轮后仍未通过，请人工调整提示词或参考图后重试")

const videoPromptRepairSystemPrompt = `你是短视频分镜提示词的内容安全改写器。输入是一条会被视频平台审核拒绝的提示词，请在不改变剧情事实的前提下改写为可过审版本。

必须保留：
- 人物身份、姓名、服装、场景、镜头顺序、时长和时间轴
- 台词、动作目的和画面叙事结果

必须删除或弱化：
- 暴力、血腥、流血、伤口、尸体、武器直接攻击、掐颈、殴打、自残、昏迷、死亡特写
- 露骨、性暗示、未成年人不当描写
- 恐怖、惊悚、引起不适的细节
- 真实公众人物、明星、影视角色、品牌或商标的直接姓名/标识；用“年轻女演员”“某品牌汽车”等身份描述替代
- 真实案件、真实机构或可能引发侵权联想的专有名称

处理方式：
- 用安全等价画面替代：例如“掐住脖子”改为“逼近对峙”，“鲜血直流”改为“踉跄后退”，“举刀”改为“握紧手中物件”
- 删除武器细节与攻击动作，只保留人物关系、情绪和剧情结果
- 不新增人物、不改变镜头数量、不写解释
- 只输出改写后的提示词正文，不要标题、序号、引号或 Markdown 代码块`

func videoPromptRepairUserPrompt(prompt, rejection string) string {
	var builder strings.Builder
	builder.WriteString("被审核拒绝的原始提示词：\n")
	builder.WriteString(strings.TrimSpace(prompt))
	if reason := strings.TrimSpace(rejection); reason != "" {
		builder.WriteString("\n\n平台拒绝原因：\n")
		builder.WriteString(reason)
	}
	builder.WriteString("\n\n请输出改写后的提示词正文。")
	return builder.String()
}

func productionTaskKey(videoID string, variantIndex int) string {
	return fmt.Sprintf("%s#v%d", videoID, variantIndex)
}

// openingVariantPromptOverride 读取某个换开头变体的独立提示词覆盖。
// 视频级 videoPrompt 不能用于变体，否则所有变体会被同一段改写文本覆盖。
func openingVariantPromptOverride(patch SettingsPatch, index int) (string, bool) {
	raw, ok := patch[openingVariantPromptsKey]
	if !ok || len(raw) == 0 {
		return "", false
	}
	var stored map[string]string
	if json.Unmarshal(raw, &stored) != nil {
		return "", false
	}
	value := strings.TrimSpace(stored[strconv.Itoa(index)])
	if value == "" {
		return "", false
	}
	return value, true
}

// videoPromptOverrideFor returns the prompt override that applies to one
// opening-variant slot. Variant slots only accept their own override; the
// VIDEO01-wide override must not leak into 换开头 variants.
func videoPromptOverrideFor(video Video, variantIndex int) (string, bool) {
	if variantIndex > 0 {
		return openingVariantPromptOverride(video.SettingsState.Patch, variantIndex)
	}
	return optionalPromptValue(video.SettingsState.Patch, "videoPrompt")
}

func videoPromptRepairCounts(video Video) map[string]int {
	counts := map[string]int{}
	raw, ok := video.SettingsState.Patch[videoPromptRepairCountsKey]
	if !ok || len(raw) == 0 {
		return counts
	}
	var stored map[string]int
	if json.Unmarshal(raw, &stored) != nil {
		return counts
	}
	for key, value := range stored {
		if value > 0 {
			counts[key] = value
		}
	}
	return counts
}

// videoPromptRepairMarkers 记录自动改写后的提示词指纹。用于区分“自动改写后
// 仍被拒”和“用户已手动修改提示词”：后者不应再被轮次上限拦住。
func videoPromptRepairMarkers(video Video) map[string]string {
	markers := map[string]string{}
	raw, ok := video.SettingsState.Patch[videoPromptRepairMarkersKey]
	if !ok || len(raw) == 0 {
		return markers
	}
	var stored map[string]string
	if json.Unmarshal(raw, &stored) != nil {
		return markers
	}
	for key, value := range stored {
		if strings.TrimSpace(value) != "" {
			markers[key] = value
		}
	}
	return markers
}

func promptDigest(text string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(text))))
}

func updatedOpeningVariantsWithPrompt(video Video, index int, prompt string) ([]OpeningVariant, error) {
	var variants []OpeningVariant
	raw, ok := video.SettingsState.Patch["openingVariants"]
	if ok {
		if err := json.Unmarshal(raw, &variants); err != nil {
			return nil, fmt.Errorf("%w: 换开头变体格式不正确", ErrInvalid)
		}
	}
	updated := false
	for position := range variants {
		if variants[position].Index == index {
			variants[position].Prompt = prompt
			variants[position].Status = "success"
			updated = true
			break
		}
	}
	if !updated {
		variants = append(variants, OpeningVariant{Index: index, Label: openingVariantLabel(index), Prompt: prompt, Status: "success"})
	}
	return variants, nil
}

// RepairRejectedVideoPrompts rewrites video prompts that were rejected by the
// provider's content moderation, mirroring the comic page repair flow. It is a
// no-op when no moderated task exists. When the repair budget is exhausted it
// returns ErrVideoPromptRepairExhausted so the caller stops resubmitting the
// same prompt.
func (s *BookStageService) RepairRejectedVideoPrompts(ctx context.Context, owner, batchID, bookID, videoID string) (bool, error) {
	if s == nil || s.Store == nil || s.Production == nil || s.Production.Store == nil {
		return false, nil
	}
	batch, err := s.Store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return false, err
	}
	book, err := bookFromBatch(batch, bookID)
	if err != nil {
		return false, err
	}
	repository, err := s.Production.repository()
	if err != nil {
		return false, err
	}
	jobs, err := repository.ListProductionJobs(ctx, owner, batchID)
	if err != nil {
		return false, err
	}
	// 同一视频/变体可能有多轮任务。只处理“最新一条任务仍是被审核拒绝”的
	// 组合：修复后重新提交的任务即使还在跑，也不能再判一次修复轮次用尽。
	latestTasks := map[string]ProductionTask{}
	for _, job := range jobs {
		if job.BookID != bookID {
			continue
		}
		for _, task := range job.Tasks {
			key := productionTaskKey(task.VideoID, task.OpeningVariantIndex)
			if current, ok := latestTasks[key]; !ok || task.CreatedAt.After(current.CreatedAt) {
				latestTasks[key] = task
			}
		}
	}
	repaired := false
	exhausted := false
	// 每修好一条就重新读取书籍，拿到最新 revision，避免同一视频的多个变体
	// 连续保存时版本冲突。
	for _, task := range latestTasks {
		if task.Status != ProductionFailed || !isModerationRejectionMessage(task.ErrorMessage) {
			continue
		}
		if strings.TrimSpace(videoID) != "" && task.VideoID != videoID {
			continue
		}
		video, found := videoByID(book, task.VideoID)
		if !found {
			continue
		}
		indexKey := strconv.Itoa(task.OpeningVariantIndex)
		counts := videoPromptRepairCounts(video)
		if counts[indexKey] >= videoPromptRepairMaxRounds {
			// 自动改写之后用户又手动改了提示词：这属于新的提交内容，不能被
			// 轮次上限永久拦住。只有提示词仍是上次自动改写结果时才转人工。
			current, promptErr := s.videoPromptForRepair(ctx, owner, batchID, book, video, task)
			if promptErr != nil {
				return repaired, promptErr
			}
			if marker := videoPromptRepairMarkers(video)[indexKey]; marker != "" && marker != promptDigest(current) {
				log.Printf("batch-factory: video prompt manually edited, allowing resubmit owner=%s book=%s video=%s variant=%d", owner, bookID, video.ID, task.OpeningVariantIndex)
				continue
			}
			exhausted = true
			continue
		}
		prompt, err := s.videoPromptForRepair(ctx, owner, batchID, book, video, task)
		if err != nil {
			return repaired, err
		}
		rewritten, err := s.rewriteRejectedVideoPrompt(ctx, owner, batchID, bookID, prompt, task.ErrorMessage)
		if err != nil {
			return repaired, err
		}
		patch := SettingsPatch{}
		if task.OpeningVariantIndex > 0 {
			variants, variantErr := updatedOpeningVariantsWithPrompt(video, task.OpeningVariantIndex, rewritten)
			if variantErr != nil {
				return repaired, variantErr
			}
			encodedVariants, marshalErr := json.Marshal(variants)
			if marshalErr != nil {
				return repaired, fmt.Errorf("%w: 保存换开头变体失败", ErrInvalid)
			}
			patch["openingVariants"] = encodedVariants
			variantOverrides := map[string]string{}
			if raw, ok := video.SettingsState.Patch[openingVariantPromptsKey]; ok {
				_ = json.Unmarshal(raw, &variantOverrides)
			}
			variantOverrides[indexKey] = rewritten
			encodedOverrides, marshalErr := json.Marshal(variantOverrides)
			if marshalErr != nil {
				return repaired, fmt.Errorf("%w: 保存换开头变体提示词失败", ErrInvalid)
			}
			patch[openingVariantPromptsKey] = encodedOverrides
		} else {
			encoded, marshalErr := json.Marshal(rewritten)
			if marshalErr != nil {
				return repaired, fmt.Errorf("%w: 保存视频提示词失败", ErrInvalid)
			}
			patch["videoPrompt"] = encoded
		}
		counts[indexKey]++
		encodedCounts, marshalErr := json.Marshal(counts)
		if marshalErr != nil {
			return repaired, fmt.Errorf("%w: 保存视频提示词改写次数失败", ErrInvalid)
		}
		patch[videoPromptRepairCountsKey] = encodedCounts
		markers := videoPromptRepairMarkers(video)
		markers[indexKey] = promptDigest(rewritten)
		encodedMarkers, marshalErr := json.Marshal(markers)
		if marshalErr != nil {
			return repaired, fmt.Errorf("%w: 保存视频提示词改写指纹失败", ErrInvalid)
		}
		patch[videoPromptRepairMarkersKey] = encodedMarkers
		if _, err := s.Store.SaveSettings(ctx, owner, ScopeRef{Kind: ScopeVideo, BatchID: batchID, BookID: bookID, VideoID: video.ID}, SettingsUpdate{Patch: patch, ExpectedRevision: video.Revision}); err != nil {
			return repaired, err
		}
		repaired = true
		log.Printf("batch-factory: content moderation rewrite owner=%s book=%s video=%s variant=%d round=%d", owner, bookID, video.ID, task.OpeningVariantIndex, counts[indexKey])
		if latestBatch, refreshErr := s.Store.GetBatch(ctx, owner, batchID); refreshErr == nil {
			if latestBook, latestErr := bookFromBatch(latestBatch, bookID); latestErr == nil {
				book = latestBook
			}
		}
	}
	if !repaired && exhausted {
		return false, ErrVideoPromptRepairExhausted
	}
	return repaired, nil
}

func (s *BookStageService) rewriteRejectedVideoPrompt(ctx context.Context, owner, batchID, bookID, sourcePrompt, rejection string) (string, error) {
	director, err := s.directorForBookTextModel(ctx, owner, batchID, bookID)
	if err != nil {
		return "", err
	}
	if director == nil || director.Provider == nil {
		return "", fmt.Errorf("%w: 文本模型未配置，无法自动改写被审核拒绝的视频提示词", ErrConfigRequired)
	}
	completion, err := director.Provider.Complete(ctx, TextCompletionRequest{
		SystemPrompt: videoPromptRepairSystemPrompt,
		UserPrompt:   videoPromptRepairUserPrompt(sourcePrompt, rejection),
		Temperature:  0.2,
		MaxTokens:    4096,
	})
	if err != nil {
		// A provider/network error means there was no model output to parse. Do
		// not misreport it as an output-format error. For retryable channel
		// outages, use the backend-owned deterministic safety rewrite so video
		// production can continue without repeatedly hitting the same 5xx.
		if errors.Is(err, ErrUpstream) {
			if fallback := promptcatalog.VideoVisualSafetyFallback(sourcePrompt); fallback != "" {
				log.Printf("batch-factory: video prompt rewrite channel unavailable, used deterministic safety fallback owner=%s book=%s", owner, bookID)
				return fallback, nil
			}
		}
		return "", fmt.Errorf("视频提示词自动去敏失败：%w", err)
	}
	rewritten := sanitizeRepairedVideoPrompt(completion)
	if rewritten == "" {
		return "", fmt.Errorf("%w: 自动去敏改写返回了空提示词", ErrUnavailable)
	}
	return rewritten, nil
}

// videoPromptForRepair returns the exact prompt text that should be rewritten.
// Variant slots use their own prompt; VIDEO01 uses the current override or the
// compiled "video" component (which already includes the H3 skeleton).
func (s *BookStageService) videoPromptForRepair(ctx context.Context, owner, batchID string, book Book, video Video, task ProductionTask) (string, error) {
	if override, ok := videoPromptOverrideFor(video, task.OpeningVariantIndex); ok {
		return override, nil
	}
	if task.OpeningVariantIndex > 0 {
		if variant, ok := openingVariantFor(video, task.OpeningVariantIndex); ok {
			return variant.Prompt, nil
		}
	}
	compiler := s.Production.Compiler
	var prompt FinalPrompt
	var err error
	switch {
	case task.OpeningVariantIndex > 0:
		resolver, ok := compiler.(openingVariantPromptResolver)
		if !ok {
			return "", fmt.Errorf("%w: 换开头变体提示词编译器不可用", ErrUnavailable)
		}
		prompt, err = resolver.CompileForOpeningVariant(ctx, owner, batchID, book.ID, video.ID, task.OpeningVariantIndex, 3)
	case task.CompilationID != "":
		resolver, ok := compiler.(h3CompilationPromptResolver)
		if !ok {
			return "", fmt.Errorf("%w: H3 提示词编译器不可用", ErrUnavailable)
		}
		prompt, err = resolver.CompileH3ForProduction(ctx, owner, batchID, book.ID, video.ID, task.CompilationID, 3)
	default:
		if resolver, ok := compiler.(referenceImageLimitPromptResolver); ok {
			prompt, err = resolver.CompileWithReferenceImageLimit(ctx, owner, batchID, book.ID, video.ID, 3)
		} else if resolver, ok := compiler.(strictPromptResolver); ok {
			prompt, err = resolver.CompileForProduction(ctx, owner, batchID, book.ID, video.ID)
		} else {
			prompt, err = compiler.Compile(ctx, owner, batchID, book.ID, video.ID)
		}
	}
	if err != nil {
		return "", err
	}
	for _, component := range prompt.Components {
		if component.Key == "video" && strings.TrimSpace(component.Content) != "" {
			return component.Content, nil
		}
	}
	if strings.TrimSpace(prompt.CompiledPrompt) != "" {
		return prompt.CompiledPrompt, nil
	}
	return "", fmt.Errorf("%w: 找不到可改写的视频提示词", ErrInvalid)
}

func sanitizeRepairedVideoPrompt(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```") {
		if index := strings.Index(text, "\n"); index >= 0 {
			text = text[index+1:]
		}
		text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	}
	text = strings.TrimSpace(text)
	for _, prefix := range []string{"改写后的提示词：", "改写后：", "改写结果：", "视频提示词：", "提示词："} {
		if strings.HasPrefix(text, prefix) {
			text = strings.TrimSpace(strings.TrimPrefix(text, prefix))
		}
	}
	return strings.TrimSpace(text)
}

func videoByID(book Book, videoID string) (Video, bool) {
	for _, video := range book.Videos {
		if video.ID == videoID {
			return video, true
		}
	}
	return Video{}, false
}
