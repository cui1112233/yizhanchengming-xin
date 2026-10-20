package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

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
		return "Seedream 4.5", true
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

func newAgentImageGenerator(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, executors *localexecutor.Service, files *localartifact.Store) AgentImageGenerator {
	return func(ctx context.Context, owner string, input AgentImageGenerationInput) ([]string, int, error) {
		localModel, local := agentLocalDoubaoImageModel(input)
		if strings.EqualFold(strings.TrimSpace(input.Provider), "doubao_local_executor") && !local {
			return nil, http.StatusUnprocessableEntity, fmt.Errorf("所选豆包图片模型不可用")
		}
		if local {
			results := make([]string, 0, input.Count)
			for index := 0; index < input.Count; index++ {
				result, status, err := generateAgentLocalDoubaoImage(ctx, executors, files, owner, localModel, input)
				if err != nil {
					return nil, status, err
				}
				results = append(results, result)
			}
			return results, http.StatusOK, nil
		}
		if store == nil || catalogStore == nil {
			return nil, http.StatusServiceUnavailable, fmt.Errorf("图片模型执行器未配置")
		}
		result, status, err := configuredBatchImages(ctx, store, catalogStore, owner, input)
		if err != nil {
			return nil, status, err
		}
		return result, status, nil
	}
}

func generateAgentLocalDoubaoImage(ctx context.Context, executors *localexecutor.Service, files *localartifact.Store, owner, model string, input AgentImageGenerationInput) (string, int, error) {
	if executors == nil || files == nil {
		return "", http.StatusServiceUnavailable, fmt.Errorf("豆包本地执行器未配置")
	}
	items, err := executors.List(ctx, owner)
	if err != nil {
		return "", http.StatusServiceUnavailable, fmt.Errorf("读取豆包本地执行器状态失败: %w", err)
	}
	online := false
	for _, item := range items {
		if item.Online {
			online = true
			break
		}
	}
	if !online {
		return "", http.StatusServiceUnavailable, fmt.Errorf("没有在线的豆包本地执行器")
	}
	mediaType, storageRef, err := waitForLocalExecutorImageWithSettings(ctx, executors, files, owner, model, input)
	if err != nil {
		return "", http.StatusBadGateway, err
	}
	file, err := files.Open(storageRef)
	if err != nil {
		return "", http.StatusBadGateway, fmt.Errorf("读取豆包图片结果失败: %w", err)
	}
	defer file.Close()
	const maxAgentImageBytes = 32 << 20
	raw, err := io.ReadAll(io.LimitReader(file, maxAgentImageBytes+1))
	if err != nil || len(raw) == 0 || len(raw) > maxAgentImageBytes {
		return "", http.StatusBadGateway, fmt.Errorf("豆包图片结果无效或过大")
	}
	return "data:" + mediaType + ";base64," + base64.StdEncoding.EncodeToString(raw), http.StatusOK, nil
}

func waitForLocalExecutorImageWithSettings(ctx context.Context, service *localexecutor.Service, files *localartifact.Store, owner, model string, input AgentImageGenerationInput) (string, string, error) {
	return waitForLocalExecutorImage(ctx, service, files, owner, model, input.Prompt, input.Size, map[string]any{
		"aspectRatio": input.AspectRatio, "quality": input.Quality, "resolution": input.Resolution, "count": 1,
	})
}
