package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localexecutor"
)

var errUnsupportedReferenceImages = errors.New("unsupported reference images")

type scriptVideoSubmission struct {
	TaskID   string
	Status   string
	Provider string
}

type scriptVideoStatus struct {
	TaskID     string
	Status     string
	Provider   string
	VideoURL   string
	ArtifactID string
	Error      string
}

type scriptVideoService struct {
	resolver      scriptVideoConfigResolver
	client        *http.Client
	localExecutor *localexecutor.Service
}

func (s *scriptVideoService) Submit(ctx context.Context, owner string, input scriptVideoLocalInput) (scriptVideoSubmission, error) {
	input.ModelKey = strings.TrimSpace(input.ModelKey)
	input.Prompt = strings.TrimSpace(input.Prompt)
	if input.Prompt == "" || len([]rune(input.Prompt)) > 12000 {
		return scriptVideoSubmission{}, fmt.Errorf("分镜视频提示词不能为空或超过 12000 个字符")
	}
	spec, known := scriptVideoModelSpec(input.ModelKey)
	if !known {
		return scriptVideoSubmission{}, fmt.Errorf("unsupported video provider")
	}
	if len(input.ImageURLs) > 0 && !spec.AcceptsReferenceImages {
		return scriptVideoSubmission{}, errUnsupportedReferenceImages
	}
	if spec.Provider == batchfactoryv11.VideoProviderDoubaoLocal {
		if s.localExecutor == nil {
			return scriptVideoSubmission{}, fmt.Errorf("本地执行器服务不可用")
		}
		sourceTaskID := strings.TrimSpace(input.SourceTaskID)
		if sourceTaskID == "" {
			sourceTaskID = newAgentID("agent-video-source")
		}
		model := strings.TrimSpace(input.Model)
		if model == "" {
			model = "Seedance 2.0 Fast"
		}
		job, err := s.localExecutor.CreateJob(ctx, owner, localexecutor.CreateJobInput{SourceTaskID: sourceTaskID, Platform: localexecutor.PlatformDoubao, Payload: map[string]any{"model": model, "prompt": input.Prompt, "duration": input.Duration, "aspectRatio": input.AspectRatio, "resolution": input.Resolution, "imageUrls": input.ImageURLs, "generateAudio": input.GenerateAudio}})
		if err != nil {
			return scriptVideoSubmission{}, err
		}
		return scriptVideoSubmission{TaskID: localScriptVideoTaskPrefix + job.ID, Status: "processing", Provider: spec.Provider}, nil
	}
	if s.resolver == nil {
		return scriptVideoSubmission{}, fmt.Errorf("视频模型配置不可用")
	}
	cfg, err := s.resolver(ctx, owner, input.ModelKey)
	if err != nil {
		return scriptVideoSubmission{}, err
	}
	adapter, err := newScriptVideoAdapter(cfg, spec.Provider)
	if err != nil {
		return scriptVideoSubmission{}, err
	}
	setScriptVideoClient(adapter, s.client)
	settings := scriptVideoEffectiveSettings(spec.Provider, input.AspectRatio, input.Resolution)
	ref, err := adapter.Submit(ctx, batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15, MaxReferenceImages: 9}, batchfactoryv11.FinalPrompt{CompiledPrompt: input.Prompt, DurationSeconds: durationValue(input.Duration), ReferenceImageURLs: input.ImageURLs, EffectiveSettings: batchfactoryv11.EffectiveSettings{Values: settings}})
	if err != nil {
		return scriptVideoSubmission{}, err
	}
	return scriptVideoSubmission{TaskID: spec.PublicTaskPrefix + ref.ProviderTaskID, Status: "processing", Provider: spec.Provider}, nil
}

func scriptVideoEffectiveSettings(provider, aspectRatio, resolution string) batchfactoryv11.SettingsPatch {
	aspectRatio = strings.TrimSpace(aspectRatio)
	switch {
	case strings.HasPrefix(aspectRatio, "16:9"):
		aspectRatio = "16:9"
	case strings.HasPrefix(aspectRatio, "9:16"):
		aspectRatio = "9:16"
	default:
		aspectRatio = ""
	}
	resolution = strings.TrimSpace(resolution)
	if strings.EqualFold(resolution, "auto") {
		resolution = ""
	}
	if provider == batchfactoryv11.VideoProviderAutoDLH3 && resolution != "" && !strings.HasSuffix(resolution, "横") && !strings.HasSuffix(resolution, "竖") {
		if aspectRatio == "16:9" {
			resolution += "横"
		} else if aspectRatio == "9:16" {
			resolution += "竖"
		}
	}
	settings := batchfactoryv11.SettingsPatch{}
	if aspectRatio != "" {
		encoded, _ := json.Marshal(aspectRatio)
		settings["videoAspectRatio"] = encoded
	}
	if resolution != "" {
		encoded, _ := json.Marshal(resolution)
		settings["videoResolution"] = encoded
	}
	return settings
}

func (s *scriptVideoService) Poll(ctx context.Context, owner, taskID string) (scriptVideoStatus, error) {
	taskID = strings.TrimSpace(taskID)
	if strings.HasPrefix(taskID, localScriptVideoTaskPrefix) {
		if s.localExecutor == nil {
			return scriptVideoStatus{}, fmt.Errorf("本地执行器服务不可用")
		}
		job, err := s.localExecutor.GetJob(ctx, owner, strings.TrimPrefix(taskID, localScriptVideoTaskPrefix))
		if err != nil {
			return scriptVideoStatus{}, err
		}
		out := scriptVideoStatus{TaskID: taskID, Provider: batchfactoryv11.VideoProviderDoubaoLocal, Status: "processing", ArtifactID: job.ArtifactID}
		switch job.State {
		case localexecutor.JobFailed, localexecutor.JobCancelled:
			out.Status, out.Error = "failed", job.ErrorMessage
		case localexecutor.JobSucceeded:
			out.Status = "succeeded"
			if job.ArtifactID != "" {
				out.VideoURL = "/api/shuihuo-production/local-executor-artifacts/" + job.ArtifactID
			}
		}
		return out, nil
	}
	prefix, provider := scriptVideoProvider(taskID)
	if prefix == "" || s.resolver == nil {
		return scriptVideoStatus{}, fmt.Errorf("unsupported video task")
	}
	modelKey := map[string]string{batchfactoryv11.VideoProviderPersonalAPI: "yd2-mini-video", batchfactoryv11.VideoProviderAutoDLH3: "minimax-h3-video", batchfactoryv11.VideoProviderYFAISeedance: "seedance-2-0-official"}[provider]
	cfg, err := s.resolver(ctx, owner, modelKey)
	if err != nil {
		return scriptVideoStatus{}, err
	}
	adapter, err := newScriptVideoAdapter(cfg, provider)
	if err != nil {
		return scriptVideoStatus{}, err
	}
	setScriptVideoClient(adapter, s.client)
	ref, err := adapter.Poll(ctx, batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15}, batchfactoryv11.ProviderTaskRef{ProviderTaskID: strings.TrimPrefix(taskID, prefix)})
	if err != nil {
		return scriptVideoStatus{}, err
	}
	out := scriptVideoStatus{TaskID: taskID, Provider: provider, Status: "processing"}
	if ref.State == batchfactoryv11.ProductionFailed {
		out.Status, out.Error = "failed", ref.ErrorMessage
	} else if ref.State == batchfactoryv11.ProductionSucceeded {
		out.Status, out.VideoURL = "succeeded", ref.MediaURL
	}
	return out, nil
}

func (s *scriptVideoService) Cancel(ctx context.Context, owner, taskID string) error {
	if !strings.HasPrefix(strings.TrimSpace(taskID), localScriptVideoTaskPrefix) || s.localExecutor == nil {
		return fmt.Errorf("当前视频提供方不支持取消")
	}
	_, err := s.localExecutor.CancelJob(ctx, owner, strings.TrimPrefix(taskID, localScriptVideoTaskPrefix))
	return err
}
