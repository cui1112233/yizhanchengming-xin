package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
)

const agentLocalDoubaoImageModelID = "local-doubao-executor-image"

func agentLocalDoubaoImageModel(input AgentImageGenerationInput) (string, bool) {
	provider := strings.ToLower(strings.TrimSpace(input.Provider))
	modelID := strings.TrimSpace(input.ModelID)
	if modelID == agentLocalDoubaoImageModelID && provider == "" {
		return agentLocalDoubaoImageModelForQuality(input.Quality), true
	}
	if modelID == agentLocalDoubaoImageModelID && provider == "doubao_local_executor" {
		return agentLocalDoubaoImageModelForQuality(input.Quality), true
	}
	if provider != "doubao_local_executor" {
		return "", false
	}
	for _, supported := range []string{"Seedream 5.0 Flash", "Seedream 5.0 Pro", "Seedream 4.5", "Seedream 4.0"} {
		if modelID == supported {
			return supported, true
		}
	}
	return "", false
}

func agentLocalDoubaoImageModelForQuality(quality string) string {
	switch strings.ToLower(strings.TrimSpace(quality)) {
	case "low":
		return "Seedream 5.0 Flash"
	case "high":
		return "Seedream 5.0 Pro"
	default:
		return "Seedream 4.5"
	}
}

func newAgentImageGenerator(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, userConfig *UserConfigStore, executors *localexecutor.Service, files *localartifact.Store, assets agentworkspace.AssetStore, references agentworkspace.ReferenceURLProvider) AgentImageGenerator {
	return func(ctx context.Context, owner string, input AgentImageGenerationInput) ([]string, int, error) {
		ctx = imageRequestIdentity(ctx, owner, input.ProjectID)
		// channel:/custom: 候选走 API 接口配置里的地址、Key 和模型名。
		if target, matched, err := integrationChannelTargetFor(ctx, userConfig, owner, input.ModelID); matched {
			if err != nil {
				return nil, http.StatusUnprocessableEntity, err
			}
			return generateIntegrationChannelImage(ctx, target, input)
		}
		localModel, local := agentLocalDoubaoImageModel(input)
		if strings.EqualFold(strings.TrimSpace(input.Provider), "doubao_local_executor") && !local {
			return nil, http.StatusUnprocessableEntity, fmt.Errorf("所选豆包图片模型不可用")
		}
		if local {
			return generateAgentLocalDoubaoImage(ctx, executors, files, assets, references, owner, localModel, input)
		}
		if store == nil || catalogStore == nil {
			return nil, http.StatusServiceUnavailable, withErrorCode(fmt.Errorf("图片模型执行器未配置"), ErrorCodeImageModelNotConfigured)
		}
		result, status, err := configuredBatchImages(ctx, store, catalogStore, owner, input)
		if err != nil {
			return nil, status, err
		}
		return result, status, nil
	}
}

func generateAgentLocalDoubaoImage(ctx context.Context, executors *localexecutor.Service, files *localartifact.Store, assets agentworkspace.AssetStore, references agentworkspace.ReferenceURLProvider, owner, model string, input AgentImageGenerationInput) ([]string, int, error) {
	if executors == nil || files == nil {
		return nil, http.StatusServiceUnavailable, withErrorCode(fmt.Errorf("豆包本地执行器未配置"), ErrorCodeExecutorNotReady)
	}
	items, err := executors.List(ctx, owner)
	if err != nil {
		return nil, http.StatusServiceUnavailable, fmt.Errorf("读取豆包本地执行器状态失败: %w", err)
	}
	online := false
	for _, item := range items {
		if item.Online {
			online = true
			break
		}
	}
	if !online {
		return nil, http.StatusServiceUnavailable, withErrorCode(fmt.Errorf("没有在线的豆包本地执行器"), ErrorCodeExecutorNotReady)
	}
	referenceImages, err := agentLocalImageReferenceURLs(ctx, assets, owner, input.ProjectID, input.ReferenceAssetIDs)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("读取图片参考素材失败: %w", err)
	}
	artifacts, err := waitForLocalExecutorAgentImages(ctx, executors, owner, model, input, referenceImages)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	const maxAgentImageBytes = 32 << 20
	results := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		file, openErr := files.Open(artifact.StorageRef)
		if openErr != nil {
			return nil, http.StatusBadGateway, fmt.Errorf("读取豆包图片结果失败: %w", openErr)
		}
		raw, readErr := io.ReadAll(io.LimitReader(file, maxAgentImageBytes+1))
		file.Close()
		if readErr != nil || len(raw) == 0 || len(raw) > maxAgentImageBytes {
			return nil, http.StatusBadGateway, fmt.Errorf("豆包图片结果无效或过大")
		}
		results = append(results, "data:"+artifact.MediaType+";base64,"+base64.StdEncoding.EncodeToString(raw))
	}
	return results, http.StatusOK, nil
}

func agentLocalDoubaoImagePrompt(prompt string, count int) string {
	if count <= 1 {
		return prompt
	}
	return strings.TrimSpace(prompt) + fmt.Sprintf("\n\n请在同一次生成中输出 %d 张独立的图片，每张都符合以上要求。", count)
}

func waitForLocalExecutorAgentImages(ctx context.Context, service *localexecutor.Service, owner, model string, input AgentImageGenerationInput, referenceImages []string) ([]localexecutor.ArtifactView, error) {
	payload := map[string]any{"type": "image", "model": model, "prompt": agentLocalDoubaoImagePrompt(input.Prompt, input.Count), "size": input.Size, "aspectRatio": input.AspectRatio, "quality": input.Quality, "resolution": input.Resolution, "count": input.Count, "referenceImages": referenceImages}
	job, err := service.CreateJob(ctx, owner, localexecutor.CreateJobInput{SourceTaskID: "agent-canvas:image:" + time.Now().UTC().Format("20060102150405.000000000"), Platform: localexecutor.PlatformDoubao, Payload: payload})
	if err != nil {
		return nil, fmt.Errorf("创建豆包图片任务失败: %w", err)
	}
	deadline := time.NewTimer(180 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		current, getErr := service.GetJob(ctx, owner, job.ID)
		if getErr != nil {
			return nil, fmt.Errorf("读取豆包图片任务失败: %w", getErr)
		}
		switch current.State {
		case localexecutor.JobSucceeded:
			artifacts, artifactsErr := service.GetJobArtifacts(ctx, owner, job.ID)
			if artifactsErr != nil {
				return nil, fmt.Errorf("读取豆包图片结果失败: %w", artifactsErr)
			}
			return artifacts, nil
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			return nil, fmt.Errorf("豆包图片任务失败: %s", artifactErrorMessage(current.ErrorMessage))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			return nil, fmt.Errorf("豆包图片任务超时")
		case <-ticker.C:
		}
	}
}

func agentLocalImageReferenceURLs(ctx context.Context, assets agentworkspace.AssetStore, owner, projectID string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	if assets == nil || strings.TrimSpace(projectID) == "" {
		return nil, agentworkspace.ErrReferenceURLUnavailable
	}
	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		asset, err := assets.Asset(ctx, owner, projectID, id)
		if err != nil {
			return nil, err
		}
		if asset.Kind != "image" {
			return nil, fmt.Errorf("素材 %s 不是图片", id)
		}
		urls = append(urls, agentExecutorReferenceAssetURL(projectID, asset.ID))
	}
	return urls, nil
}

func agentImageReferenceURLs(ctx context.Context, assets agentworkspace.AssetStore, references agentworkspace.ReferenceURLProvider, owner, projectID string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	if assets == nil || references == nil || strings.TrimSpace(projectID) == "" {
		return nil, agentworkspace.ErrReferenceURLUnavailable
	}
	urls := make([]string, 0, len(ids))
	for _, id := range ids {
		asset, err := assets.Asset(ctx, owner, projectID, id)
		if err != nil {
			return nil, err
		}
		if asset.Kind != "image" {
			return nil, fmt.Errorf("素材 %s 不是图片", id)
		}
		url, err := references.ReferenceURL(ctx, asset, 15*time.Minute)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func waitForLocalExecutorImageWithSettings(ctx context.Context, service *localexecutor.Service, files *localartifact.Store, owner, model string, input AgentImageGenerationInput, referenceImages []string) (string, string, error) {
	return waitForLocalExecutorImage(ctx, service, files, owner, model, input.Prompt, input.Size, map[string]any{
		"aspectRatio": input.AspectRatio, "quality": input.Quality, "resolution": input.Resolution, "count": 1,
		"imageUrls": referenceImages,
	})
}
