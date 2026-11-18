package screenplay

import (
	"context"
	"fmt"
	"qiantie/backend/internal/prompt/luoshui"
	"strings"
	"time"
)

// Reparse is the original standalone parser entrance, not the degree-based
// extraction/image pipeline. It uses the saved project model and complete input.
func (s *AssetService) Reparse(ctx context.Context, req AssetRequest) (AssetResult, error) {
	if ctx.Err() != nil {
		return AssetResult{}, fmt.Errorf("%w: %w", ErrUserCancelled, ctx.Err())
	}
	if !validAssetType(req.Type) {
		return AssetResult{}, fmt.Errorf("%w: 未知资产类型", ErrInvalidArgument)
	}
	if strings.TrimSpace(req.Content) == "" {
		return AssetResult{}, ErrEmptyContent
	}
	project, err := s.Store.GetProject(ctx, req.Owner, req.ProjectID)
	if err != nil {
		return AssetResult{}, err
	}
	if plan, blocked := pendingPlan(project, req.Confirmed, "3_assets", "重新解析"+assetTypeLabel(req.Type), "使用作品模型重新解析并更新文字，保留图片引用"); blocked {
		return AssetResult{RequiresConfirmation: true, PendingActions: plan.PendingActions}, nil
	}
	if strings.TrimSpace(project.ModelID) == "" {
		return AssetResult{}, fmt.Errorf("%w: 请先在作品模型配置中选择文字模型", ErrInvalidArgument)
	}
	if s.Generator == nil {
		return AssetResult{}, ErrAPIKeyMissing
	}
	system, user, ok := luoshui.NativeAssetMessages(string(req.Type), req.Content)
	if !ok {
		return AssetResult{}, fmt.Errorf("%w: 缺少资产解析提示词", ErrInvalidArgument)
	}
	token := newScreenplayID("reparse")
	lock := "asset-reparse:" + req.Owner + ":" + req.ProjectID
	if err = s.Store.AcquireLock(ctx, lock, token, 15*time.Minute); err != nil {
		return AssetResult{}, err
	}
	defer func() { _ = s.Store.ReleaseLock(context.Background(), lock, token) }()
	response, err := s.Generator.Generate(ctx, TextRequest{Model: project.ModelID, SystemPrompt: system, UserPrompt: user, Temperature: 1, MaxOutputTokens: 20000, JSONResponse: true, SingleAttempt: true})
	if ctx.Err() != nil {
		return AssetResult{}, fmt.Errorf("%w: %w", ErrUserCancelled, ctx.Err())
	}
	if err != nil {
		return AssetResult{}, wrapAIError(err)
	}
	assets, err := parseAssetJSON(req.Type, strings.TrimSpace(response.Content))
	if err != nil {
		return AssetResult{}, err
	}
	if len(assets) == 0 {
		return AssetResult{}, fmt.Errorf("%w: AI未返回有效的%s数据", ErrArtifactInvalid, assetTypeLabel(req.Type))
	}
	old, err := s.Store.ListAssets(ctx, req.Owner, req.ProjectID, req.Type)
	if err != nil {
		return AssetResult{}, err
	}
	byName := map[string]Asset{}
	for _, a := range old {
		byName[a.Name] = a
	}
	merged := []Asset{}
	positions := map[string]int{}
	for _, a := range assets {
		if pos, found := positions[a.Name]; found {
			mergeAsset(&merged[pos], a)
			continue
		}
		a.ID = newScreenplayID("asset")
		a.ProjectID = req.ProjectID
		a.OwnerUsername = req.Owner
		a.Degree = project.AssetDegree
		if !validDegree(a.Degree) {
			a.Degree = DegreeStandard
		}
		a.Revision = 1
		a.CreatedAt = time.Now().UTC()
		a.UpdatedAt = a.CreatedAt
		if previous, exists := byName[a.Name]; exists {
			a.ID = previous.ID
			a.Revision = previous.Revision + 1
			a.CreatedAt = previous.CreatedAt
			if a.Extra == nil {
				a.Extra = map[string]string{}
			}
			for k, v := range previous.Extra {
				_, exists := a.Extra[k]
				// A text parser never selects or replaces persisted media.
				mediaKey := strings.Contains(k, "image") || strings.Contains(k, "video") || strings.Contains(k, "audio") || strings.Contains(k, "reference")
				if !exists || mediaKey {
					a.Extra[k] = v
				}
			}
		}
		positions[a.Name] = len(merged)
		merged = append(merged, a)
	}
	if ctx.Err() != nil {
		return AssetResult{}, fmt.Errorf("%w: %w", ErrUserCancelled, ctx.Err())
	}
	saved, err := s.Store.MergeAssets(ctx, req.Owner, req.ProjectID, merged)
	if err != nil {
		return AssetResult{}, err
	}
	return AssetResult{Assets: saved, ChunkCount: 1, Batch: AssetBatch{ProjectID: req.ProjectID, OwnerUsername: req.Owner, Type: req.Type, Status: "done", ItemCount: len(saved)}}, nil
}
