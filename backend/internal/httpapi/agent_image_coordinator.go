package httpapi

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

const agentImageLease = 3 * time.Minute

type agentImageCoordinator struct {
	services AgentNodeGenerationServices
	mu       sync.Mutex
	inFlight map[string]struct{}
	now      func() time.Time
}

func newAgentImageCoordinator(services AgentNodeGenerationServices) *agentImageCoordinator {
	now := services.Now
	if now == nil {
		now = time.Now
	}
	return &agentImageCoordinator{services: services, inFlight: map[string]struct{}{}, now: now}
}

func (s *agentImageCoordinator) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		_ = s.pollPendingOnce(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *agentImageCoordinator) pollPendingOnce(ctx context.Context) error {
	if s == nil || s.services.Projects == nil {
		return nil
	}
	tasks, err := s.services.Projects.ListRunnableGenerationTasksByModes(ctx, []string{"text_to_image", "image_to_image", "image_generation", "image_crop", "image_rotate", "image_mirror", "image_grid_split"}, 100)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Kind != "image" || (task.State == agentworkspace.GenerationRunning && s.now().UTC().Sub(task.UpdatedAt) < agentImageLease) || !s.claim(task.ID) {
			continue
		}
		go func(task agentworkspace.GenerationTask) {
			defer s.release(task.ID)
			_ = s.execute(context.Background(), task)
		}(task)
	}
	return nil
}

func (s *agentImageCoordinator) claim(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inFlight[id]; ok {
		return false
	}
	s.inFlight[id] = struct{}{}
	return true
}
func (s *agentImageCoordinator) release(id string) {
	s.mu.Lock()
	delete(s.inFlight, id)
	s.mu.Unlock()
}

func (s *agentImageCoordinator) execute(ctx context.Context, task agentworkspace.GenerationTask) error {
	services := s.services
	if services.Assets == nil || services.Objects == nil || (imageToolTaskMode(task.Mode) && services.ImageProcess == nil) || (!imageToolTaskMode(task.Mode) && services.ImageGenerate == nil) {
		log.Printf("agent image task unavailable id=%s mode=%s assets=%t objects=%t image_generate=%t image_process=%t", task.ID, task.Mode, services.Assets != nil, services.Objects != nil, services.ImageGenerate != nil, services.ImageProcess != nil)
	}
	if services.Assets == nil {
		return s.fail(ctx, task, "图片生成服务未配置：缺少结果素材记录存储")
	}
	if services.Objects == nil {
		return s.fail(ctx, task, "图片生成服务未配置：缺少结果素材存储")
	}
	if imageToolTaskMode(task.Mode) && services.ImageProcess == nil {
		return s.fail(ctx, task, "图片编辑服务未配置")
	}
	if !imageToolTaskMode(task.Mode) && services.ImageGenerate == nil {
		return s.fail(ctx, task, "图片生成服务未配置：缺少图片模型执行器")
	}
	if task.State == agentworkspace.GenerationQueued {
		var err error
		task, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationRunning, ExpectedUpdatedAt: task.UpdatedAt, Now: s.now().UTC()})
		if err != nil {
			return err
		}
	}
	if imageToolTaskMode(task.Mode) {
		outputs, err := runAgentImageTool(ctx, task.OwnerUsername, task.ProjectID, task, services)
		if err != nil {
			return s.fail(ctx, task, err.Error())
		}
		results, err := persistAgentImageToolOutputs(ctx, task.OwnerUsername, task.ProjectID, outputs, services)
		if err != nil {
			return s.fail(ctx, task, "图片编辑结果保存失败")
		}
		current, err := services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
		if err != nil || current.State == agentworkspace.GenerationCancelled {
			return err
		}
		_, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: results, SelectedResultID: results[0].ID, ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
		return err
	}
	parameters, _ := task.Settings["parameters"].(map[string]any)
	advanced, _ := task.Settings["advanced"].(map[string]any)
	request := agentImageGenerationRequest{NodeID: task.NodeID, NodeVersion: task.NodeVersion, Provider: task.Provider, ModelID: task.ModelID, Mode: task.Mode, Prompt: strings.TrimSpace(fmt.Sprint(task.Settings["prompt"])), Parameters: parameters, Advanced: advanced, Size: strings.TrimSpace(fmt.Sprint(task.Settings["size"])), AspectRatio: strings.TrimSpace(fmt.Sprint(task.Settings["aspectRatio"])), Quality: strings.TrimSpace(fmt.Sprint(task.Settings["quality"])), Resolution: strings.TrimSpace(fmt.Sprint(task.Settings["resolution"])), ReferenceAssetIDs: task.ReferenceAssetIDs}
	if count, ok := task.Settings["count"].(float64); ok {
		request.Count = int(count)
	}
	if count, ok := task.Settings["count"].(int); ok {
		request.Count = count
	}
	input, err := normalizeAgentImageGenerationInput(request)
	if err != nil {
		return s.fail(ctx, task, err.Error())
	}
	input.ProjectID, input.ReferenceAssetIDs = task.ProjectID, task.ReferenceAssetIDs
	urls, _, err := services.ImageGenerate(ctx, task.OwnerUsername, input)
	if err != nil {
		return s.fail(ctx, task, err.Error())
	}
	if len(urls) == 0 {
		return s.fail(ctx, task, "图片模型未返回图片结果")
	}
	ingestor := newAgentMediaAssetIngestor(services.Assets, services.Objects, nil, s.now)
	results := make([]agentworkspace.GenerationResult, 0, len(urls))
	for _, url := range urls {
		asset, saveErr := ingestor.SaveImageResult(ctx, task.OwnerUsername, task.ProjectID, newAgentID("asset"), url)
		if saveErr != nil {
			return s.fail(ctx, task, "图片结果保存失败")
		}
		view := map[string]any{"id": asset.ID, "assetId": asset.ID, "kind": asset.Kind, "name": asset.Name, "url": asset.URL, "contentType": asset.ContentType, "byteSize": asset.ByteSize}
		results = append(results, agentworkspace.GenerationResult{ID: asset.ID, AssetID: asset.ID, Kind: asset.Kind, Asset: view, CreatedAt: asset.CreatedAt})
	}
	current, err := services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil {
		return err
	}
	if current.State == agentworkspace.GenerationCancelled {
		return nil
	}
	_, err = services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationAwaitingConfirmation, Results: results, SelectedResultID: results[0].ID, ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}

func (s *agentImageCoordinator) fail(ctx context.Context, task agentworkspace.GenerationTask, message string) error {
	current, err := s.services.Projects.GenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID)
	if err != nil || current.State == agentworkspace.GenerationCancelled || current.State == agentworkspace.GenerationFailed {
		return err
	}
	_, err = s.services.Projects.TransitionGenerationTask(ctx, task.OwnerUsername, task.ProjectID, task.ID, agentworkspace.GenerationTaskTransition{State: agentworkspace.GenerationFailed, ErrorMessage: strings.TrimSpace(message), ExpectedUpdatedAt: current.UpdatedAt, Now: s.now().UTC()})
	return err
}
