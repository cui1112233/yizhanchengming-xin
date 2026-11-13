package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
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
	service      screenplayVideoService
	configs      *UserConfigStore
	poll         time.Duration
	repo         screenplay.ProductionRepository
	References   func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error)
	ReferenceIDs func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error)
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

		if r.repo == nil {
			return artifacts, fmt.Errorf("生产镜头存储未配置，已停止提交")
		}
		if err := ctx.Err(); err != nil {
			return artifacts, err
		}
		input := scriptVideoLocalInput{SourceTaskID: fmt.Sprintf("screenplay:%s:%s:%d", project.ID, storyboard.ID, shot.Seq), ModelKey: modelKey, Prompt: prompt, Duration: shot.Duration, AspectRatio: project.AspectRatio, Resolution: resolution, GenerateAudio: true}
		if r.References != nil {
			input.ImageURLs, err = r.References(ctx, project, shot)
			if err != nil {
				return artifacts, err
			}
		}
		hashInput := input
		if r.ReferenceIDs != nil {
			hashInput.ImageURLs, err = r.ReferenceIDs(ctx, project, shot)
			if err != nil {
				return artifacts, err
			}
		}
		raw, _ := json.Marshal(hashInput)
		hash := fmt.Sprintf("%x", sha256.Sum256(raw))
		input.SourceTaskID += ":" + hash[:16]
		task, getErr := r.repo.GetShotTask(ctx, project.OwnerUsername, project.ID, storyboard.ID, shot.Seq)
		if versions, ok := r.repo.(screenplay.VersionedProductionRepository); ok {
			previous, previousErr := task, getErr
			task, getErr = versions.GetShotTaskVersion(ctx, project.OwnerUsername, project.ID, storyboard.ID, shot.Seq, hash)
			if errors.Is(getErr, screenplay.ErrNotFound) && previousErr == nil && previous.RequestHash != hash {
				job, jobErr := r.repo.GetProductionJob(ctx, project.OwnerUsername, project.ID)
				if jobErr != nil || !job.Rebuild {
					return artifacts, fmt.Errorf("第 %d 集第 %d 镜头请求已改变；请明确选择重建变更镜头，旧任务将保留", episode, shot.Seq)
				}
			}
		}
		if getErr != nil && !errors.Is(getErr, screenplay.ErrNotFound) {
			return artifacts, getErr
		}
		if getErr == nil && task.RequestHash != hash {
			return artifacts, fmt.Errorf("第 %d 集第 %d 镜头内容或模型已改变，请明确创建新版本后再生成", episode, shot.Seq)
		}
		if task.Status == "succeeded" && task.Artifact != "" {
			task.ShotID = shot.ID
			if saveErr := r.repo.SaveShotTask(ctx, task); saveErr != nil {
				return artifacts, saveErr
			}
			artifacts = append(artifacts, task.Artifact)
			continue
		}
		if task.Status == "failed" {
			return artifacts, fmt.Errorf("第 %d 集第 %d 镜头供应商任务已失败：%s", episode, shot.Seq, task.Error)
		}
		if task.TaskID == "" {
			if job, checkErr := r.repo.GetProductionJob(ctx, project.OwnerUsername, project.ID); checkErr == nil && job.Status != "running" {
				return artifacts, screenplay.ErrUserCancelled
			} else if checkErr != nil && !errors.Is(checkErr, screenplay.ErrNotFound) {
				return artifacts, checkErr
			}
			if getErr == nil && task.Status != "not_submitted" {
				return artifacts, fmt.Errorf("第 %d 集第 %d 镜头提交结果未知，已停止重复提交；请核对供应商任务", episode, shot.Seq)
			}
			task = screenplay.ShotTask{Owner: project.OwnerUsername, ProjectID: project.ID, StoryboardID: storyboard.ID, ShotID: shot.ID, Episode: episode, Seq: shot.Seq, Status: "submitting", RequestHash: hash}
			if err = r.repo.SaveShotTask(ctx, task); err != nil {
				return artifacts, err
			}
			if err = ctx.Err(); err != nil {
				task.Status = "not_submitted"
				_ = r.repo.SaveShotTask(context.Background(), task)
				return artifacts, err
			}
			if job, checkErr := r.repo.GetProductionJob(ctx, project.OwnerUsername, project.ID); checkErr == nil && job.Status != "running" {
				task.Status = "not_submitted"
				_ = r.repo.SaveShotTask(context.Background(), task)
				return artifacts, screenplay.ErrUserCancelled
			} else if checkErr != nil && !errors.Is(checkErr, screenplay.ErrNotFound) {
				return artifacts, checkErr
			}
			submission, submitErr := r.service.Submit(ctx, project.OwnerUsername, input)
			if submitErr != nil {
				task.Status = "submission_unknown"
				task.Error = submitErr.Error()
				_ = r.repo.SaveShotTask(context.Background(), task)
				return artifacts, fmt.Errorf("第 %d 集第 %d 镜头提交失败（不会自动重复提交）：%w", episode, shot.Seq, submitErr)
			}
			if strings.TrimSpace(submission.TaskID) == "" {
				return artifacts, fmt.Errorf("供应商未返回任务编号，已停止重复提交")
			}
			task.TaskID = submission.TaskID
			task.Status = "processing"
			if err = r.repo.SaveShotTask(context.Background(), task); err != nil {
				return artifacts, fmt.Errorf("供应商任务 %s 已提交但保存失败，禁止重复提交：%w", submission.TaskID, err)
			}
		}

		for {
			if err = ctx.Err(); err != nil {
				return artifacts, err
			}
			status, pollErr := r.service.Poll(ctx, project.OwnerUsername, task.TaskID)
			if pollErr != nil {
				task.Error = pollErr.Error()
				if saveErr := r.repo.SaveShotTask(context.Background(), task); saveErr != nil {
					return artifacts, saveErr
				}
				return artifacts, fmt.Errorf("第 %d 集第 %d 镜头查询失败（继续时查询原任务）：%w", episode, shot.Seq, pollErr)
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
				task.Status = "succeeded"
				task.Artifact = artifact
				task.Error = ""
				if saveErr := r.repo.SaveShotTask(context.Background(), task); saveErr != nil {
					return artifacts, saveErr
				}
				artifacts = append(artifacts, artifact)
				goto nextShot
			case "failed":
				task.Status = "failed"
				task.Error = status.Error
				if saveErr := r.repo.SaveShotTask(context.Background(), task); saveErr != nil {
					return artifacts, saveErr
				}
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
