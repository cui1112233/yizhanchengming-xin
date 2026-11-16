package httpapi

import (
	"context"
	"fmt"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
)

func syncScriptVideoModelConfig(ctx context.Context, registry batchfactoryv11.VideoProviderRegistry, owner string, config map[string]any, modelKey string) error {
	if registry == nil {
		return batchfactoryv11.ErrUnavailable
	}
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		if !strings.EqualFold(modelString(model, "kind"), "video") || !modelEnabled(model) {
			continue
		}
		if !strings.EqualFold(modelString(model, "modelId"), modelKey) && !strings.EqualFold(modelString(model, "id"), modelKey) {
			continue
		}
		if err := syncShuihuoVideoModel(ctx, registry, owner, model); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("%w: video model is not configured", batchfactoryv11.ErrUnavailable)
}

func scriptVideoModelConfig(_ context.Context, _ batchfactoryv11.Store, _ string, config map[string]any, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
	spec, ok := scriptVideoModelSpec(modelKey)
	if ok && spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("unsupported script video provider")
	}
	models := shuihuoCollection(config, "modelCatalog", nil)
	var exact []map[string]any
	for _, model := range models {
		if modelString(model, "id") == modelKey {
			exact = append(exact, model)
		}
	}
	if len(exact) > 1 {
		return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("视频模型配置ID重复")
	}
	if len(exact) == 1 {
		models = exact
	} else {
		matches := 0
		for _, model := range models {
			if modelString(model, "modelId") == modelKey {
				matches++
			}
		}
		if matches > 1 {
			return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("视频模型重名，请选择具体配置ID")
		}
	}
	for _, model := range models {
		if !strings.EqualFold(modelString(model, "kind"), "video") || !modelEnabled(model) {
			continue
		}
		if modelString(model, "modelId") != modelKey && modelString(model, "id") != modelKey {
			continue
		}
		provider := modelString(model, "provider")
		if !isNativeScriptVideoProvider(provider) {
			if !ok {
				return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("unsupported script video provider")
			}
			provider = batchfactoryv11.VideoProviderForModel(modelKey, provider)
		}
		if isNativeScriptVideoProvider(provider) && (modelString(model, "id") == "" || len(modelString(model, "id")) > nativeScriptConfigIDMaxBytes) {
			return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("原生视频配置缺少有效配置条目ID")
		}
		if isNativeScriptVideoProvider(provider) && modelString(model, "modelId") == "" {
			return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("原生视频配置缺少真实modelId")
		}
		apiKey := shuihuoModelCredential(model)
		createURL := firstText(model["endpoint"], model["baseUrl"])
		tasksURL := firstText(model["tasksEndpoint"])
		resultURL := firstText(model["resultEndpoint"])
		if provider == batchfactoryv11.VideoProviderAutoDLH3 {
			if createURL == "" {
				createURL = batchfactoryv11.DefaultAutoDLH3CreateURL
			}
			if tasksURL == "" {
				tasksURL = batchfactoryv11.DefaultAutoDLH3TasksURL
			}
		}
		if provider == batchfactoryv11.VideoProviderPersonalAPI {
			if createURL == "" {
				createURL = batchfactoryv11.DefaultPersonalVideoCreateURL
			}
			if tasksURL == "" {
				tasksURL = batchfactoryv11.DefaultPersonalVideoTasksURL
			}
			if resultURL == "" {
				resultURL = strings.TrimRight(tasksURL, "/") + "/{id}/result"
			}
		}
		return batchfactoryv11.VideoProviderConfig{ConfigID: modelString(model, "id"), Provider: provider, APIKey: apiKey, Model: firstText(model["modelId"], modelKey), CreateURL: createURL, TasksURL: tasksURL, ResultURL: resultURL, WorkflowID: modelString(model, "workflowId")}, nil
	}
	return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("%w: video model is not configured", batchfactoryv11.ErrUnavailable)
}

func modelEnabled(model map[string]any) bool {
	enabled, exists := model["enabled"].(bool)
	return !exists || enabled
}
