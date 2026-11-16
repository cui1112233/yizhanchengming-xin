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

type scriptVideoPreflightError struct{ Err error }

func (e *scriptVideoPreflightError) Error() string { return e.Err.Error() }
func (e *scriptVideoPreflightError) Unwrap() error { return e.Err }

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
	spec, cfg, err := s.resolveInput(ctx, owner, input)
	if err != nil {
		return scriptVideoSubmission{}, &scriptVideoPreflightError{Err: err}
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
		imageURLs := append([]string(nil), input.ImageURLs...)
		if len(input.ReferenceAssets) > 0 {
			imageURLs = make([]string, 0, len(input.ReferenceAssets))
			for _, reference := range input.ReferenceAssets {
				imageURLs = append(imageURLs, agentExecutorReferenceAssetURL(reference.ProjectID, reference.AssetID))
			}
		}
		job, err := s.localExecutor.CreateJob(ctx, owner, localexecutor.CreateJobInput{SourceTaskID: sourceTaskID, Platform: localexecutor.PlatformDoubao, Payload: map[string]any{"model": model, "prompt": input.Prompt, "duration": input.Duration, "aspectRatio": input.AspectRatio, "resolution": input.Resolution, "imageUrls": imageURLs, "generateAudio": input.GenerateAudio}})
		if err != nil {
			return scriptVideoSubmission{}, err
		}
		return scriptVideoSubmission{TaskID: localScriptVideoTaskPrefix + job.ID, Status: "processing", Provider: spec.Provider}, nil
	}
	adapter, err := newScriptVideoAdapter(cfg, spec.Provider)
	if err != nil {
		return scriptVideoSubmission{}, err
	}
	setScriptVideoClient(adapter, s.client)
	settings := scriptVideoEffectiveSettings(spec.Provider, input.AspectRatio, input.Resolution)
	ref, err := adapter.Submit(ctx, batchfactoryv11.FrozenVideoModel{ID: cfg.Model, MaxDuration: 15, MaxReferenceImages: 9}, batchfactoryv11.FinalPrompt{CompiledPrompt: input.Prompt, DurationSeconds: durationValue(input.Duration), ReferenceImageURLs: input.ImageURLs, SubjectReferenceURIs: input.SubjectReferenceURIs, EffectiveSettings: batchfactoryv11.EffectiveSettings{Values: settings}})
	if err != nil {
		return scriptVideoSubmission{}, err
	}
	prefix := spec.PublicTaskPrefix
	if isNativeScriptVideoProvider(spec.Provider) {
		prefix = nativeScriptTaskPrefix(spec.Provider, cfg.ConfigID)
	}
	return scriptVideoSubmission{TaskID: prefix + ref.ProviderTaskID, Status: "processing", Provider: spec.Provider}, nil
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
		return scriptVideoStatus{}, &scriptVideoPreflightError{Err: fmt.Errorf("unsupported video task")}
	}
	modelKey := scriptVideoTaskModelKey(taskID, provider)
	cfg, err := s.resolver(ctx, owner, modelKey)
	if err != nil {
		return scriptVideoStatus{}, &scriptVideoPreflightError{Err: err}
	}
	if isNativeScriptVideoProvider(provider) && cfg.ConfigID != modelKey {
		return scriptVideoStatus{}, &scriptVideoPreflightError{Err: fmt.Errorf("原视频任务的模型配置已不存在，请恢复原配置条目")}
	}
	if isNativeScriptVideoProvider(provider) && cfg.Provider != provider {
		return scriptVideoStatus{}, &scriptVideoPreflightError{Err: fmt.Errorf("视频任务模型配置协议已更改")}
	}
	adapter, err := newScriptVideoAdapter(cfg, provider)
	if err != nil {
		return scriptVideoStatus{}, &scriptVideoPreflightError{Err: err}
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
