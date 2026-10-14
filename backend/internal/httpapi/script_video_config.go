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
	if !ok || spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("unsupported script video provider")
	}
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		if !strings.EqualFold(modelString(model, "kind"), "video") || !modelEnabled(model) {
			continue
		}
		if !strings.EqualFold(modelString(model, "modelId"), modelKey) && !strings.EqualFold(modelString(model, "id"), modelKey) {
			continue
		}
		provider := batchfactoryv11.VideoProviderForModel(modelKey, modelString(model, "provider"))
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
		return batchfactoryv11.VideoProviderConfig{Provider: provider, APIKey: apiKey, Model: firstText(model["modelId"], modelKey), CreateURL: createURL, TasksURL: tasksURL, ResultURL: resultURL, WorkflowID: modelString(model, "workflowId")}, nil
	}
	return batchfactoryv11.VideoProviderConfig{}, fmt.Errorf("%w: video model is not configured", batchfactoryv11.ErrUnavailable)
}

func modelEnabled(model map[string]any) bool {
	enabled, exists := model["enabled"].(bool)
	return !exists || enabled
}
