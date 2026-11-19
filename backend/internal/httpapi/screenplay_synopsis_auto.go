package httpapi

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

// synopsisAttachResult 把 AssetResult 摊平成 map 并附加自动生成的概要字段：
// synopsis_result（本次新生成时非 nil）。概要生成失败时把提示并入 warnings 数组
// （前端 collectWarnings 通道）并带 synopsis_warning 明细。AssetResult 字段原样
// 保留在顶层，老调用方按结构体解码不受影响。
func synopsisAttachResult(result screenplay.AssetResult, synopsis *screenplay.SynopsisResult, warning string) map[string]any {
	payload := map[string]any{}
	if data, err := json.Marshal(result); err == nil {
		_ = json.Unmarshal(data, &payload)
	}
	if synopsis != nil {
		payload["synopsis_result"] = synopsis
	}
	if warning != "" {
		payload["synopsis_warning"] = warning
		warnings, _ := payload["warnings"].([]any)
		payload["warnings"] = append(warnings, warning)
	}
	return payload
}

// 洛水对齐：AI 推理分析流水线在人物/场景/道具解析之前自动附带跑「研读剧本 · 故事概要」
// （synopsis Agent）。V88 在资产解析入口做同样的事：
//   - 项目尚无 synopsis_result 且原文、模型齐备时，自动生成并写回创作配置；
//   - 已有结果直接跳过，不重复消耗模型额度（对应洛水流水线只跑一次）；
//   - 前提不满足或生成失败时静默跳过/告警，绝不阻断资产解析本身（洛水流水线
//     某步失败同样继续跑后续 Agent）。
// 返回值：result 非 nil 表示本次新生成并已落库；nil 表示跳过或失败（失败时带 err）。
func autoGenerateSynopsisIfNeeded(ctx context.Context, store screenplay.Store, owner, projectID, model string, resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error), nowFn func() time.Time) (*screenplay.SynopsisResult, error) {
	project, err := store.GetProject(ctx, owner, projectID)
	if err != nil {
		return nil, err
	}
	// 无原文或未指定模型：无法生成，静默跳过（资产解析照常进行）。
	if strings.TrimSpace(project.SourceText) == "" || strings.TrimSpace(model) == "" {
		return nil, nil
	}
	cs, ok := store.(screenplay.CreativeStore)
	if !ok {
		return nil, nil
	}
	// 已有概要结果：跳过，避免重复生成烧额度。
	if screenplay.EffectiveCreativeConfig(project).SynopsisResult != nil {
		return nil, nil
	}
	generator := screenplayTextGenerator{owner: owner, resolver: resolver}
	result, err := screenplay.GenerateSynopsis(ctx, project.SourceText, func(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
		response, err := generator.Generate(ctx, screenplay.TextRequest{
			Model: model, SystemPrompt: systemPrompt, UserPrompt: userPrompt,
			MaxOutputTokens: 4000, SingleAttempt: true,
		})
		return response.Content, err
	}, nowFn())
	if err != nil {
		return nil, err
	}
	config := screenplay.EffectiveCreativeConfig(project)
	config.SynopsisResult = &result
	if _, err := cs.SaveCreativeConfig(ctx, owner, project.ID, config, project.Revision); err != nil {
		return nil, err
	}
	return &result, nil
}
