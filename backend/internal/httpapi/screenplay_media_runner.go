package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/screenplay"
)

type screenplayVideoService interface {
	Submit(context.Context, string, scriptVideoLocalInput) (scriptVideoSubmission, error)
	Poll(context.Context, string, string) (scriptVideoStatus, error)
}

// screenplayVideoRunner 把漫剧分镜接到已经存在的在线视频生成通道。
// 它不保存密钥；只读取当前账户在“设置”中保存的视频模型。
type screenplayVideoRunner struct {
	service screenplayVideoService
	configs *UserConfigStore
	poll    time.Duration
}

func (r *screenplayVideoRunner) ProduceEpisode(ctx context.Context, project screenplay.Project, episode int, storyboard screenplay.Storyboard) ([]string, error) {
	if r == nil || r.service == nil || r.configs == nil {
		return nil, fmt.Errorf("漫剧视频生产服务不可用")
	}
	config, err := r.configs.GetConfig(ctx, project.OwnerUsername)
	if err != nil {
		return nil, err
	}
	modelKey, resolution := screenplayVideoDefaults(config)
	if modelKey == "" {
		return nil, fmt.Errorf("请先在设置中选择并保存一个可用的视频模型")
	}
	if len(storyboard.Shots) == 0 {
		return nil, fmt.Errorf("第 %d 集没有可生产的分镜镜头", episode)
	}
	interval := r.poll
	if interval <= 0 {
		interval = 2 * time.Second
	}
	artifacts := make([]string, 0, len(storyboard.Shots))
	for _, shot := range storyboard.Shots {
		prompt := strings.TrimSpace(shot.VideoDesc)
		if prompt == "" {
			return artifacts, fmt.Errorf("第 %d 集第 %d 镜头缺少视频提示词", episode, shot.Seq)
		}
		submission, submitErr := r.service.Submit(ctx, project.OwnerUsername, scriptVideoLocalInput{
			SourceTaskID: fmt.Sprintf("screenplay:%s:%d:%d", project.ID, episode, shot.Seq),
			ModelKey:     modelKey, Prompt: prompt, Duration: shot.Duration,
			AspectRatio: project.AspectRatio, Resolution: resolution, GenerateAudio: true,
		})
		if submitErr != nil {
			return artifacts, fmt.Errorf("第 %d 集第 %d 镜头提交失败：%w", episode, shot.Seq, submitErr)
		}
		for {
			status, pollErr := r.service.Poll(ctx, project.OwnerUsername, submission.TaskID)
			if pollErr != nil {
				return artifacts, fmt.Errorf("第 %d 集第 %d 镜头查询失败：%w", episode, shot.Seq, pollErr)
			}
			switch status.Status {
			case "succeeded":
				artifact := strings.TrimSpace(status.VideoURL)
				if artifact == "" && strings.TrimSpace(status.ArtifactID) != "" {
					artifact = "/api/shuihuo-production/local-executor-artifacts/" + strings.TrimSpace(status.ArtifactID)
				}
				if artifact == "" {
					return artifacts, fmt.Errorf("第 %d 集第 %d 镜头完成但没有视频地址", episode, shot.Seq)
				}
				artifacts = append(artifacts, artifact)
				goto nextShot
			case "failed":
				return artifacts, fmt.Errorf("第 %d 集第 %d 镜头生成失败：%s", episode, shot.Seq, firstText(status.Error, "视频通道返回失败"))
			}
			select {
			case <-ctx.Done():
				return artifacts, ctx.Err()
			case <-time.After(interval):
			}
		}
	nextShot:
	}
	return artifacts, nil
}

func screenplayVideoDefaults(config map[string]any) (string, string) {
	defaults, _ := config["scriptDefaults"].(map[string]any)
	modelKey := strings.TrimSpace(fmt.Sprint(defaults["videoModelKey"]))
	resolution := strings.TrimSpace(fmt.Sprint(defaults["videoResolution"]))
	if resolution == "" || resolution == "<nil>" {
		resolution = "720p"
	}
	if modelKey == "<nil>" {
		modelKey = ""
	}
	return modelKey, resolution
}
