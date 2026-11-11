package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// automationPreflightIssue 是预检面板可以直接渲染的一条问题。
// Code 保持稳定，前端据此决定跳转入口；Message 是给用户看的中文文案。
type automationPreflightIssue struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	FixTarget string `json:"fixTarget,omitempty"`
	FixLabel  string `json:"fixLabel,omitempty"`
}

const (
	automationFixBatchSettings   = "batch_settings"
	automationFixBookSettings    = "book_settings"
	automationFixPreset          = "preset"
	automationFixAdminModels     = "admin_models"
	automationFixPublishSettings = "publish_settings"
)

type automationPreflightOverride struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type automationPreflightBook struct {
	BookID    string                        `json:"bookId"`
	Title     string                        `json:"title"`
	Passed    bool                          `json:"passed"`
	Blocking  []automationPreflightIssue    `json:"blocking"`
	Warnings  []automationPreflightIssue    `json:"warnings"`
	Overrides []automationPreflightOverride `json:"overrides"`
}

type automationPreflightSummary struct {
	Total        int `json:"total"`
	Passed       int `json:"passed"`
	Blocked      int `json:"blocked"`
	WarningBooks int `json:"warningBooks"`
}

// automationTaskPeek 是当前批次活动任务的只读快照。
type automationTaskPeek struct {
	State string
	JobID string
	Since string
}

// automationPreflight 只读取批次、预设、系统预设库与任务状态，不写 job、
// 不写书状态、不写 counts、不落盘任何设置，也不发起任何供应商生成调用。
type automationPreflight struct {
	BatchID         string                     `json:"batchId"`
	RunMode         string                     `json:"runMode"`
	ProductionMode  string                     `json:"productionMode"`
	PresetID        string                     `json:"presetId"`
	ReadOnly        bool                       `json:"readOnly"`
	CheckedAt       string                     `json:"checkedAt"`
	OK              bool                       `json:"ok"`
	TaskState       string                     `json:"taskState,omitempty"`
	RunningJobID    string                     `json:"runningJobId,omitempty"`
	RunningSince    string                     `json:"runningSince,omitempty"`
	TaskPaused      bool                       `json:"taskPaused"`
	Summary         automationPreflightSummary `json:"summary"`
	RunnableBookIDs []string                   `json:"runnableBookIds"`
	BatchIssues     []automationPreflightIssue `json:"batchIssues"`
	Books           []automationPreflightBook  `json:"books"`
}

// Preflight 复用 Start 的入参形状（presetId/bookIds/runMode/productionMode/
// settingsPatch），但把所有“预设合并 + 提示词冻结”都只在内存里模拟一次，
// 用于点击「立即执行」之前展示逐书阻断原因。
func (c *BatchAutomationCoordinator) Preflight(ctx context.Context, owner, batchID string, input map[string]any) (*automationPreflight, error) {
	if c == nil || c.store == nil {
		return nil, fmt.Errorf("%w: Go 自动化调度器未配置", batchfactoryv11.ErrUnavailable)
	}
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(batchID) == "" {
		return nil, batchfactoryv11.ErrInvalid
	}
	batch, err := c.store.GetBatch(ctx, owner, batchID)
	if err != nil {
		return nil, err
	}
	books, err := selectedAutomationBooks(batch, input["bookIds"])
	if err != nil {
		return nil, err
	}
	runMode := automationStringValue(input["runMode"], "full_submit")
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")

	result := &automationPreflight{
		BatchID:         batchID,
		RunMode:         runMode,
		ProductionMode:  productionMode,
		ReadOnly:        true,
		CheckedAt:       c.now().Format(time.RFC3339),
		BatchIssues:     []automationPreflightIssue{},
		Books:           make([]automationPreflightBook, 0, len(books)),
		RunnableBookIDs: []string{},
	}

	// 活动任务在面板打开时就要可见，而不是等提交后才吃 409。peek 只做
	// AutomationStateStore.Load（纯 SELECT），不归一化、不 Save、不落盘。
	if peek := c.peekActiveAutomationJob(ctx, owner, batchID); peek != nil {
		result.TaskState = peek.State
		result.RunningJobID = peek.JobID
		result.RunningSince = peek.Since
		result.TaskPaused = peek.State == "paused"
		if !automationBoolValue(input["queueIfActive"]) {
			result.BatchIssues = append(result.BatchIssues, automationPreflightIssue{
				Code: "automation_job_active",
				Message: fmt.Sprintf(
					"当前批次已有自动生产任务在进行（%s%s）；请先停止当前任务后再重试",
					automationTaskStateLabel(peek.State), automationTaskSinceText(peek.Since),
				),
			})
		}
	}

	// 预设：和 Start 一样“预设只是默认值，批次统一配置优先”。只在内存里
	// 合并并冻结提示词正文，不调用 SaveSettings。
	presetID := automationStringValue(input["presetId"], "")
	result.PresetID = presetID
	settings, presetIssue := c.preflightSettings(ctx, owner, batch, input, presetID)
	if presetIssue != nil {
		result.BatchIssues = append(result.BatchIssues, *presetIssue)
	}
	batch.SettingsState.Patch = settings
	result.ProductionMode = comicSettingString(settings, "productionMode", productionMode)

	// 全局执行器缺失属于整批不可执行，单独列出，避免每本书重复同一条原因。
	result.BatchIssues = append(result.BatchIssues, c.automationGlobalPreflightIssues(ctx, owner, batch, runMode)...)

	for _, book := range books {
		issues := c.automationBookPreflightIssues(ctx, owner, batch, book, runMode, true)
		blocking := make([]automationPreflightIssue, 0, len(issues))
		blocking = append(blocking, issues...)
		warnings := c.automationBookPreflightWarnings(batch, book, c.automationEffectiveSettings(batch, book))
		overrides := automationBookOverrideKeys(book.SettingsState.Patch)
		passed := len(blocking) == 0
		result.Books = append(result.Books, automationPreflightBook{
			BookID:    book.ID,
			Title:     strings.TrimSpace(book.Title),
			Passed:    passed,
			Blocking:  blocking,
			Warnings:  warnings,
			Overrides: overrides,
		})
		if passed {
			result.RunnableBookIDs = append(result.RunnableBookIDs, book.ID)
		}
	}

	result.Summary.Total = len(result.Books)
	for _, entry := range result.Books {
		if entry.Passed {
			result.Summary.Passed++
		} else {
			result.Summary.Blocked++
		}
		if len(entry.Warnings) > 0 {
			result.Summary.WarningBooks++
		}
	}
	result.OK = len(result.BatchIssues) == 0 && result.Summary.Passed > 0
	return result, nil
}

// peekActiveAutomationJob 只读地找出当前批次的活动任务。它刻意绕开
// loadLocked：后者在状态归一化或旧单文件迁移时会 Save/写盘，预检不能复用。
// 没有持久化 store 时宁可返回 nil，也不去读文件布局——文件布局的兼容读取
// 可能触发迁移写盘；生产环境始终走 MySQL 的 AutomationStateStore。
func (c *BatchAutomationCoordinator) peekActiveAutomationJob(ctx context.Context, owner, batchID string) *automationTaskPeek {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	store := c.automationStore
	c.mu.Unlock()
	if store == nil {
		return nil
	}
	state, found, err := store.Load(ctx)
	if err != nil || !found {
		return nil
	}
	var picked *automationTaskPeek
	var pickedUpdated time.Time
	for jobID, job := range state.Jobs {
		if job == nil || automationStringValue(job["owner"], "") != owner || automationStringValue(job["batchId"], "") != batchID {
			continue
		}
		switch automationStringValue(job["state"], "") {
		case "running", "paused", "scheduled", "needs_attention":
		default:
			continue
		}
		updated := automationStoredJobTime(job["updatedAt"])
		// 同一批次理论上只有一个活动任务；历史脏数据按 updatedAt 取最新的一个。
		if picked != nil && !updated.After(pickedUpdated) {
			continue
		}
		since := automationStoredJobTime(job["startedAt"])
		if since.IsZero() {
			since = automationStoredJobTime(job["scheduledAt"])
		}
		if since.IsZero() {
			since = updated
		}
		id := automationStringValue(job["id"], "")
		if id == "" {
			id = jobID
		}
		sinceText := ""
		if !since.IsZero() {
			sinceText = since.UTC().Format(time.RFC3339)
		}
		picked = &automationTaskPeek{State: automationStringValue(job["state"], ""), JobID: id, Since: sinceText}
		pickedUpdated = updated
	}
	return picked
}

func automationTaskStateLabel(state string) string {
	switch state {
	case "running":
		return "运行中"
	case "paused":
		return "已暂停"
	case "scheduled":
		return "已定时"
	case "needs_attention":
		return "需要处理"
	default:
		return state
	}
}

func automationTaskSinceText(since string) string {
	if strings.TrimSpace(since) == "" {
		return ""
	}
	return "，开始于 " + since
}

// automationEffectiveSettings 返回批次 + 单书合并后的设置（只读快照）。
func (c *BatchAutomationCoordinator) automationEffectiveSettings(batch batchfactoryv11.Batch, book batchfactoryv11.Book) batchfactoryv11.SettingsPatch {
	return batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)
}

// preflightSettings 在内存中模拟 Start 的预设合并与提示词冻结，返回最终生效
// 的批次级设置。任何 hydrate 失败都转成一条批次级阻断项，而不是接口 500。
func (c *BatchAutomationCoordinator) preflightSettings(ctx context.Context, owner string, batch batchfactoryv11.Batch, input map[string]any, presetID string) (batchfactoryv11.SettingsPatch, *automationPreflightIssue) {
	current := map[string]any{}
	for key, raw := range batch.SettingsState.Patch {
		var value any
		if json.Unmarshal(raw, &value) == nil {
			current[key] = value
		}
	}
	if presetID != "" {
		if c.presets == nil {
			return batch.SettingsState.Patch, &automationPreflightIssue{
				Code:      "preset_service_missing",
				Message:   "自动化预设服务未配置；请先在自动生产中选择预设",
				FixTarget: automationFixPreset,
				FixLabel:  "自动化预设",
			}
		}
		preset, err := c.presets.Get(owner, presetID)
		if err != nil || preset == nil {
			return batch.SettingsState.Patch, &automationPreflightIssue{
				Code:      "preset_not_found",
				Message:   "自动化预设不存在或不属于当前账号；请重新选择预设",
				FixTarget: automationFixPreset,
				FixLabel:  "自动化预设",
			}
		}
		if len(preset.Config) > 0 {
			current = mergeAutomationPresetConfig(current, preset.Config)
		}
	}

	if overlay, ok := input["settingsPatch"].(map[string]any); ok {
		for key, value := range overlay {
			current[key] = value
		}
	}
	if mode := strings.TrimSpace(automationStringValue(input["productionMode"], "")); mode == "comic" || mode == "video" {
		current["productionMode"] = mode
	}
	current, hydrateIssue := c.preflightHydratePromptConfig(current)
	patch, err := automationSettingsPatchFromMap(current)
	if err != nil {
		return batch.SettingsState.Patch, &automationPreflightIssue{
			Code:      "settings_patch_invalid",
			Message:   "统一配置读取失败；请在统一配置里重新保存一次",
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置",
		}
	}
	return patch, hydrateIssue
}

// preflightHydratePromptConfig 与 applyPreset 使用同一个冻结函数，保证预检
// 结论和真正启动时一致；只是结果留在内存里，不写回存档。
func (c *BatchAutomationCoordinator) preflightHydratePromptConfig(config map[string]any) (map[string]any, *automationPreflightIssue) {
	value, ok := config["aiPromptConfig"]
	if !ok || value == nil || c.promptCatalog == nil {
		return config, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return config, &automationPreflightIssue{
			Code:      "prompt_config_invalid",
			Message:   "提示词配置格式不正确；请在统一配置里重新选择提示词预设",
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置",
		}
	}
	hydrated, hydrateErr := hydrateBatchFactoryPromptConfig(c.promptCatalog, raw)
	if hydrateErr != nil {
		return config, &automationPreflightIssue{
			Code:      "prompt_preset_unavailable",
			Message:   "所选提示词预设不存在或未发布：" + hydrateErr.Error(),
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置",
		}
	}
	var decoded map[string]any
	if json.Unmarshal(hydrated, &decoded) == nil {
		config["aiPromptConfig"] = decoded
	}
	return config, nil
}

func automationSettingsPatchFromMap(values map[string]any) (batchfactoryv11.SettingsPatch, error) {
	patch := batchfactoryv11.SettingsPatch{}
	for key, value := range values {
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		patch[key] = raw
	}
	return patch, nil
}

// automationGlobalPreflightIssues 是 validateAutomationGlobalPreflight 的结构化
// 版本：只覆盖“整批不可执行”的条件。
func (c *BatchAutomationCoordinator) automationGlobalPreflightIssues(ctx context.Context, owner string, batch batchfactoryv11.Batch, runMode string) []automationPreflightIssue {
	issues := make([]automationPreflightIssue, 0, 3)
	if c.stages == nil || c.stages.Director == nil || c.stages.Director.Provider == nil {
		issues = append(issues, automationPreflightIssue{
			Code:      "text_executor_missing",
			Message:   "文本模型执行器未配置",
			FixTarget: automationFixAdminModels,
			FixLabel:  "管理员 → 水货生产模型",
		})
	}
	productionMode := comicSettingString(batch.SettingsState.Patch, "productionMode", "video")
	if productionMode == "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.ImageReady == nil {
			issues = append(issues, automationPreflightIssue{
				Code:      "image_executor_missing",
				Message:   "图片生成执行器未配置",
				FixTarget: automationFixAdminModels,
				FixLabel:  "管理员 → 水货生产模型",
			})
		} else if ready, reason := c.stages.ImageReady(ctx, owner); !ready {
			issues = append(issues, automationPreflightIssue{
				Code:      "image_executor_unavailable",
				Message:   reasonOrDefault(reason, "图片生成执行器不可用"),
				FixTarget: automationFixAdminModels,
				FixLabel:  "管理员 → 水货生产模型",
			})
		}
	}
	if productionMode != "comic" && runMode != "storyboard_only" {
		if c.stages == nil || c.stages.VideoReady == nil {
			issues = append(issues, automationPreflightIssue{
				Code:      "video_executor_missing",
				Message:   "视频生成执行器未配置",
				FixTarget: automationFixAdminModels,
				FixLabel:  "管理员 → 水货生产模型",
			})
		} else if ready, reason := c.stages.VideoReady(ctx, owner); !ready {
			issues = append(issues, automationPreflightIssue{
				Code:      "video_executor_unavailable",
				Message:   reasonOrDefault(reason, "视频生成执行器不可用"),
				FixTarget: automationFixAdminModels,
				FixLabel:  "管理员 → 水货生产模型",
			})
		}
	}
	if runMode == "full_submit" {
		if productionMode == "comic" && c.comicPublish == nil {
			issues = append(issues, automationPreflightIssue{
				Code:      "comic_publish_executor_missing",
				Message:   "漫画发布执行器未配置",
				FixTarget: automationFixPublishSettings,
				FixLabel:  "发布设置",
			})
		}
		if productionMode != "comic" && c.publish121 == nil && c.external == nil {
			issues = append(issues, automationPreflightIssue{
				Code:      "video_publish_executor_missing",
				Message:   "视频发布执行器未配置",
				FixTarget: automationFixPublishSettings,
				FixLabel:  "发布设置",
			})
		}
	}
	return issues
}

// automationBookPreflightIssues 逐书预检的结构化版本。extended=false 时与运行
// 时的 automationBookPreflightReason 完全一致（保证线上阻断行为不变）；
// extended=true 时额外检查提示词预设、正文冻结与模型可用性，供执行前展示。
func (c *BatchAutomationCoordinator) automationBookPreflightIssues(ctx context.Context, owner string, batch batchfactoryv11.Batch, book batchfactoryv11.Book, runMode string, extended bool) []automationPreflightIssue {
	issues := []automationPreflightIssue{}
	if isDirectFirstSourceBlocked(book) && len(productionLines(book)) == 0 && firstMetadataString(book.SourceMetadata, "giantMaterialId", "qingyuMaterialId") == "" {
		title := strings.TrimSpace(book.Title)
		if title == "" {
			title = "未命名作品"
		}
		issues = append(issues, automationPreflightIssue{
			Code:      "source_fetch_blocked",
			Message:   fmt.Sprintf("《%s》书城原文获取失败，等待人工处理", title),
			FixTarget: automationFixBookSettings,
			FixLabel:  "单书设置",
		})
		return issues
	}
	effective := c.automationEffectiveSettings(batch, book)
	textModelID := comicSettingString(effective, "textModelId", "")
	if textModelID == "" {
		issues = append(issues, automationPreflightIssue{
			Code:      "text_model_missing",
			Message:   "文本模型未配置",
			FixTarget: automationFixAdminModels,
			FixLabel:  "管理员 → 水货生产模型",
		})
	} else if extended && !c.automationTextModelAvailable(ctx, owner, textModelID) {
		issues = append(issues, automationPreflightIssue{
			Code:      "text_model_unavailable",
			Message:   "选择的文本模型不可用：" + textModelID + "；请重新选择文本模型或启用该模型",
			FixTarget: automationFixAdminModels,
			FixLabel:  "管理员 → 水货生产模型",
		})
	}
	if runMode == "storyboard_only" {
		return issues
	}
	productionMode := batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch)
	if productionMode == "comic" {
		if comicSettingString(effective, "imageModelId", "") == "" {
			issues = append(issues, automationPreflightIssue{
				Code:      "image_model_missing",
				Message:   "图片模型未配置；" + setupHintImage,
				FixTarget: automationFixAdminModels,
				FixLabel:  "管理员 → 水货生产模型",
			})
		}
	} else if comicSettingString(effective, "videoModelId", "") == "" {
		issues = append(issues, automationPreflightIssue{
			Code:      "video_model_missing",
			Message:   "视频模型未配置；" + setupHintVideoModel,
			FixTarget: automationFixBookSettings,
			FixLabel:  "书设置 → 视频生成",
		})
	}
	if runMode == "full_submit" {
		publishIssues := automationBookPublishIssues(productionMode, effective)
		issues = append(issues, publishIssues...)
		if productionMode == "comic" && len(publishIssues) == 0 && c.comicPublishPreflight != nil {
			if err := c.comicPublishPreflight(ctx, owner, batch.ID, book.ID, effective); err != nil {
				issues = append(issues, automationPreflightIssue{
					Code:      "comic_publish_directory_invalid",
					Message:   err.Error(),
					FixTarget: automationFixPublishSettings,
					FixLabel:  "发布设置",
				})
			}
		}
	}
	if extended && productionMode == "comic" {
		issues = append(issues, automationComicPromptIssues(effective, book)...)
	}
	return issues
}

func automationBookPublishIssues(productionMode string, effective batchfactoryv11.SettingsPatch) []automationPreflightIssue {
	publishKey := "publishSettings"
	if productionMode == "comic" {
		publishKey = "comicPublishSettings"
	}
	var publish map[string]any
	if raw := effective[publishKey]; len(raw) > 0 {
		_ = json.Unmarshal(raw, &publish)
	}
	if publish != nil && hasAutomationPublishProfile(publish) {
		return nil
	}
	if productionMode == "comic" {
		return []automationPreflightIssue{{
			Code:      "comic_publish_profile_missing",
			Message:   "漫画发布设置未绑定网站或组织；请到「统一配置 → 发布设置」绑定组织/制作者",
			FixTarget: automationFixPublishSettings,
			FixLabel:  "发布设置",
		}}
	}
	return []automationPreflightIssue{{
		Code:      "publish_profile_missing",
		Message:   "发布配置未绑定网站或组织",
		FixTarget: automationFixPublishSettings,
		FixLabel:  "发布设置",
	}}
}

// automationComicPromptIssues 检查漫画页 / 资产 / 画风预设是否已选且正文已冻结。
// 正文未冻结是 Start 时 applyPreset hydrate 能补的，但单书覆盖无法被补，
// 所以这里一律提示清楚是哪一层配置需要修。
func automationComicPromptIssues(effective batchfactoryv11.SettingsPatch, book batchfactoryv11.Book) []automationPreflightIssue {
	issues := []automationPreflightIssue{}
	config := comicPromptConfig(effective)
	if strings.TrimSpace(config.ID) == "" && strings.TrimSpace(config.Body) == "" {
		issues = append(issues, automationPreflightIssue{
			Code:      "comic_page_preset_missing",
			Message:   "未选择漫画页提示词预设；请在「统一配置 → 漫画提示词」选择漫画分页预设",
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置",
		})
	} else if strings.TrimSpace(config.Body) == "" {
		issues = append(issues, automationPreflightIssue{
			Code:      "comic_prompt_body_not_frozen",
			Message:   "漫画页提示词正文未冻结（只保存了预设 ID）；请重新保存一次统一配置",
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置",
		})
	}
	if batchfactoryv11.AssetImagesEnabled(effective) {
		assets := config.Assets
		if strings.TrimSpace(assets.ID) == "" && strings.TrimSpace(assets.Body) == "" {
			issues = append(issues, automationPreflightIssue{
				Code:      "comic_assets_preset_missing",
				Message:   "请在资产设置中选择漫画资产提示词",
				FixTarget: automationFixBatchSettings,
				FixLabel:  "统一配置",
			})
		} else if strings.TrimSpace(assets.Body) == "" {
			issues = append(issues, automationPreflightIssue{
				Code:      "comic_assets_prompt_body_not_frozen",
				Message:   "漫画资产提示词正文未冻结（只保存了预设 ID）；请重新保存一次统一配置",
				FixTarget: automationFixBatchSettings,
				FixLabel:  "统一配置",
			})
		}
	}
	if config.TargetPageCount < 0 {
		issues = append(issues, automationPreflightIssue{
			Code:      "comic_page_target_invalid",
			Message:   "漫画目标页数为负数；请重新选择漫画提示词预设",
			FixTarget: automationFixBatchSettings,
			FixLabel:  "统一配置 → 漫画提示词",
		})
	}
	return issues
}

func comicPromptConfig(effective batchfactoryv11.SettingsPatch) batchfactoryv11.AIReasoningComicModule {
	raw, ok := effective["aiPromptConfig"]
	if !ok {
		return batchfactoryv11.AIReasoningComicModule{}
	}
	var config struct {
		Comic batchfactoryv11.AIReasoningComicModule `json:"comic"`
	}
	if json.Unmarshal(raw, &config) != nil {
		return batchfactoryv11.AIReasoningComicModule{}
	}
	return config.Comic
}

// automationBookPreflightWarnings 只产出“不阻断但执行时很可能失败”的风险项。
func (c *BatchAutomationCoordinator) automationBookPreflightWarnings(batch batchfactoryv11.Batch, book batchfactoryv11.Book, effective batchfactoryv11.SettingsPatch) []automationPreflightIssue {
	warnings := []automationPreflightIssue{}
	productionMode := batchfactoryv11.EffectiveProductionMode(batch.SettingsState.Patch, book.SettingsState.Patch)
	if productionMode != "comic" {
		return warnings
	}
	if target := comicPromptConfig(effective).TargetPageCount; target > 0 {
		existing := len(book.Videos)
		if existing > 0 && existing != target {
			warnings = append(warnings, automationPreflightIssue{
				Code: "comic_page_count_mismatch",
				Message: fmt.Sprintf(
					"已有 %d 个漫画分页与目标 %d 页不一致，执行后可能报「漫画页数量与分页不一致」；请确认漫画提示词预设或先清理旧分页",
					existing, target,
				),
				FixTarget: automationFixBatchSettings,
				FixLabel:  "统一配置 → 漫画提示词",
			})
		}
	}
	return warnings
}

// automationTextModelAvailable 通过只读的账号模型解析确认所选文本模型仍可用。
// 没有解析器时退回“执行器是否已配置”的判断，绝不发起生成请求。
func (c *BatchAutomationCoordinator) automationTextModelAvailable(ctx context.Context, owner, modelID string) bool {
	if strings.TrimSpace(modelID) == "" {
		return false
	}
	if c.stages == nil {
		return false
	}
	if c.stages.TextProviderResolver == nil {
		return c.stages.Director != nil && c.stages.Director.Provider != nil
	}
	provider, err := c.stages.TextProviderResolver(ctx, owner, modelID)
	return err == nil && provider != nil
}

// automationBookOverrideKeys 列出本书相对批次统一配置生效的单书覆盖项，
// 用于提示“改统一配置对这本来无效”。
func automationBookOverrideKeys(patch batchfactoryv11.SettingsPatch) []automationPreflightOverride {
	if len(patch) == 0 {
		return []automationPreflightOverride{}
	}
	labels := map[string]string{
		"aiPromptConfig":       "漫画/视频提示词（含漫画页与资产预设）",
		"productionMode":       "生产模式",
		"textModelId":          "文本模型",
		"imageModelId":         "图片模型",
		"videoModelId":         "视频模型",
		"comicPublishSettings": "漫画发布设置",
		"publishSettings":      "发布设置",
		"assetImagesEnabled":   "资产图片开关",
		"comic":                "漫画分页/页数",
	}
	keys := make([]string, 0, len(patch))
	for key := range patch {
		keys = append(keys, key)
	}
	sortStrings(keys)
	overrides := make([]automationPreflightOverride, 0, len(keys))
	for _, key := range keys {
		label, ok := labels[key]
		if !ok {
			continue
		}
		overrides = append(overrides, automationPreflightOverride{Key: key, Label: label})
	}
	return overrides
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func registerBatchAutomationPreflightRoute(mux *http.ServeMux, auth BridgeAuth, coordinator *BatchAutomationCoordinator, version string) {
	if mux == nil || coordinator == nil {
		return
	}
	mux.Handle("POST /api/batch-factory/"+version+"/batches/{batchId}/automation/preflight", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		result, err := coordinator.Preflight(r.Context(), owner, r.PathValue("batchId"), input)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})))
}
