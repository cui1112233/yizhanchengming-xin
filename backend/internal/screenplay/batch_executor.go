package screenplay

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// BatchMediaRunner 是漫剧批次与真实媒体生产能力之间的端口。
// 实现必须在返回前把产物持久化，不能只返回一个展示用状态。
type BatchMediaRunner interface {
	ProduceEpisode(ctx context.Context, project Project, episode int, storyboard Storyboard) ([]string, error)
}

// BatchExecutor 把九步状态机接到已有的剧本、资产、分镜和媒体服务。
// 每次 RunNext 只执行一个最小单元，便于暂停、重试和刷新恢复。
type BatchExecutor struct {
	Store     Store
	Generator TextGenerator
	Media     BatchMediaRunner
	Finalizer ProductionFinalizer
}

type BatchRunResult struct {
	Batch     *Batch   `json:"batch"`
	Unit      Unit     `json:"unit"`
	Artifacts []string `json:"artifacts,omitempty"`
}

type batchUnitOutcome struct {
	artifacts    []string
	review       *Review
	episodeCount int
}

// RunNext 执行并落库一个步骤或一集。失败也会写回 failed 和错误原因。
func (e BatchExecutor) RunNext(ctx context.Context, owner, projectID string) (BatchRunResult, error) {
	if e.Store == nil || strings.TrimSpace(owner) == "" || strings.TrimSpace(projectID) == "" {
		return BatchRunResult{}, fmt.Errorf("%w: 批次执行参数不完整", ErrInvalidArgument)
	}
	token := newScreenplayID("batch-run")
	lockName := "screenplay-batch:" + owner + ":" + projectID
	if err := e.Store.AcquireLock(ctx, lockName, token, 15*time.Minute); err != nil {
		return BatchRunResult{}, err
	}
	defer func() { _ = e.Store.ReleaseLock(context.Background(), lockName, token) }()
	authoringLock := "authoring:" + owner + ":" + projectID
	if err := e.Store.AcquireLock(ctx, authoringLock, token, 15*time.Minute); err != nil {
		return BatchRunResult{}, err
	}
	defer func() { _ = e.Store.ReleaseLock(context.Background(), authoringLock, token) }()
	if renewer, ok := e.Store.(productionLockRenewer); ok {
		runCtx, cancel := context.WithCancel(ctx)
		ctx = runCtx
		defer cancel()
		done := make(chan struct{})
		defer close(done)
		go func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-runCtx.Done():
					return
				case <-ticker.C:
					for _, name := range []string{lockName, authoringLock} {
						if err := renewer.RenewProductionLock(runCtx, name, token, 15*time.Minute); err != nil {
							cancel()
							return
						}
					}
				}
			}
		}()
	}

	var unit Unit
	var regenerate bool
	started, err := e.Store.MutateBatch(ctx, owner, projectID, func(batch *Batch) error {
		if NormalizeBatchStatus(batch.Status) == BatchPaused {
			return fmt.Errorf("%w: 批次已暂停，请先继续任务", ErrInvalidState)
		}
		var nextErr error
		unit, nextErr = NextUnit(batch)
		if nextErr != nil {
			return nextErr
		}
		if step := batch.Step(unit.StepKey); step != nil {
			regenerate = step.Status == StepStale
			if ep := step.Episodes[unit.Episode]; ep != nil && ep.Status == StepStale {
				regenerate = true
			}
		}
		return StartStep(batch, unit)
	})
	if err != nil {
		return BatchRunResult{Batch: started, Unit: unit}, err
	}
	_ = e.Store.AppendBatchLog(ctx, started.ID, owner, "info", "开始执行："+unit.String())

	outcome, runErr := e.executeUnit(ctx, owner, projectID, unit, regenerate)
	if runErr != nil {
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, ErrUserCancelled) {
			paused, persistErr := e.Store.MutateBatch(context.Background(), owner, projectID, func(batch *Batch) error {
				PauseStep(batch, unit)
				return nil
			})
			_ = e.Store.AppendBatchLog(context.Background(), started.ID, owner, "info", "用户暂停："+unit.String())
			if persistErr != nil {
				return BatchRunResult{Batch: paused, Unit: unit}, persistErr
			}
			return BatchRunResult{Batch: paused, Unit: unit}, fmt.Errorf("%w: %v", ErrUserCancelled, runErr)
		}
		failed, persistErr := e.Store.MutateBatch(context.Background(), owner, projectID, func(batch *Batch) error {
			FailStep(batch, unit, runErr.Error())
			if regenerate && (unit.StepKey == "3_assets" || unit.StepKey == "5_storyboard") {
				st := batch.Step(unit.StepKey)
				st.Status = StepStale
				if ep := st.Episodes[unit.Episode]; ep != nil {
					ep.Status = StepStale
				}
			}
			return nil
		})
		_ = e.Store.AppendBatchLog(context.Background(), started.ID, owner, "error", unit.String()+"失败："+runErr.Error())
		if persistErr != nil {
			return BatchRunResult{Batch: failed, Unit: unit}, fmt.Errorf("%v；状态写回失败：%w", runErr, persistErr)
		}
		return BatchRunResult{Batch: failed, Unit: unit}, runErr
	}

	completed, err := e.Store.MutateBatch(ctx, owner, projectID, func(batch *Batch) error {
		wasPaused := batch.Status == BatchPaused
		defer func() {
			if wasPaused {
				batch.Status = BatchPaused
			}
		}()
		if outcome.episodeCount > 0 {
			for _, key := range []string{"5_storyboard", "6_storyboard_doctor", "7_video"} {
				if step := batch.Step(key); step != nil {
					step.TotalEpisodes = outcome.episodeCount
				}
			}
		}
		if outcome.review != nil {
			ApplyDoctorReview(batch, unit, outcome.review)
		} else {
			CompleteStep(batch, unit, outcome.artifacts)
		}
		return nil
	})
	if err != nil {
		return BatchRunResult{Batch: completed, Unit: unit, Artifacts: outcome.artifacts}, err
	}
	_ = e.Store.AppendBatchLog(ctx, completed.ID, owner, "info", "执行完成："+unit.String())
	return BatchRunResult{Batch: completed, Unit: unit, Artifacts: outcome.artifacts}, nil
}

func (e BatchExecutor) executeUnit(ctx context.Context, owner, projectID string, unit Unit, regenerate bool) (batchUnitOutcome, error) {
	project, err := e.Store.GetProject(ctx, owner, projectID)
	if err != nil {
		return batchUnitOutcome{}, err
	}
	switch unit.StepKey {
	case "0_ingest":
		count := 0
		if document, activeErr := e.Store.ActiveDocument(ctx, owner, projectID); activeErr == nil {
			count = document.EpisodeCount
			if count <= 0 {
				count = countBatchEpisodes(document.Content)
			}
		}
		if count <= 0 {
			if strings.TrimSpace(project.SourceText) == "" {
				return batchUnitOutcome{}, ErrEmptyContent
			}
			count = countBatchEpisodes(project.SourceText)
		}
		return batchUnitOutcome{artifacts: []string{"source://" + project.ID}, episodeCount: count}, nil

	case "1_screenplay":
		if document, activeErr := e.Store.ActiveDocument(ctx, owner, projectID); activeErr == nil && strings.TrimSpace(document.Content) != "" {
			count := document.EpisodeCount
			if count <= 0 {
				count = countBatchEpisodes(document.Content)
			}
			return batchUnitOutcome{artifacts: []string{"document://" + document.ID}, episodeCount: count}, nil
		}
		if e.Generator == nil {
			return batchUnitOutcome{}, fmt.Errorf("%w: 未配置文本模型，无法生成剧本", ErrInvalidState)
		}
		service := ScreenplayService{Store: e.Store, Generator: e.Generator}
		var result ScreenplayResult
		if strings.EqualFold(strings.TrimSpace(project.ScriptMode), "a1") {
			result, err = service.GenerateA1(ctx, A1Request{Owner: owner, ProjectID: projectID, Content: project.SourceText, Model: project.ModelID, PromptID: project.PromptID, Confirmed: true})
		} else {
			result, err = service.GenerateB(ctx, BRequest{Owner: owner, ProjectID: projectID, Novel: project.SourceText, Model: project.ModelID, PromptID: project.PromptID, Confirmed: true})
		}
		if err != nil {
			return batchUnitOutcome{}, err
		}
		return batchUnitOutcome{artifacts: []string{"document://" + result.Document.ID}, episodeCount: result.EpisodeCount}, nil

	case "2_screenplay_doctor":
		document, activeErr := e.Store.ActiveDocument(ctx, owner, projectID)
		if activeErr != nil || strings.TrimSpace(document.Content) == "" {
			return batchUnitOutcome{}, ErrNoActiveScript
		}
		return e.doctor(ctx, project, unit, document.Content)

	case "3_assets":
		allArtifacts := make([]string, 0)
		for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
			rows, listErr := e.Store.ListAssets(ctx, owner, projectID, kind)
			if listErr != nil {
				return batchUnitOutcome{}, listErr
			}
			if len(rows) == 0 || regenerate {
				if e.Generator == nil {
					return batchUnitOutcome{}, fmt.Errorf("%w: %s资产为空且未配置文本模型", ErrInvalidState, kind)
				}
				result, extractErr := (&AssetService{Store: e.Store, Generator: e.Generator}).Extract(ctx, AssetRequest{
					Owner: owner, ProjectID: projectID, Type: kind, Degree: normalizeBatchDegree(project.AssetDegree),
					Model: project.ModelID, Confirmed: true,
				})
				if extractErr != nil {
					return batchUnitOutcome{}, extractErr
				}
				rows = result.Assets
			}
			for _, asset := range rows {
				allArtifacts = append(allArtifacts, "asset://"+asset.ID)
			}
		}
		return batchUnitOutcome{artifacts: allArtifacts}, nil

	case "4_preset_doctor":
		assets := []Asset{}
		for _, kind := range []AssetType{AssetCharacter, AssetScene, AssetProp} {
			rows, listErr := e.Store.ListAssets(ctx, owner, projectID, kind)
			if listErr != nil {
				return batchUnitOutcome{}, listErr
			}
			if len(rows) == 0 {
				return batchUnitOutcome{}, fmt.Errorf("%w: 缺少%s资产", ErrArtifactInvalid, kind)
			}
			assets = append(assets, rows...)
		}
		return e.doctor(ctx, project, unit, assets)

	case "5_storyboard":
		episode := unit.Episode
		if episode <= 0 {
			episode = 1
		}
		if storyboard, found := existingBatchStoryboard(ctx, e.Store, owner, projectID, episode); found && !regenerate {
			return batchUnitOutcome{artifacts: []string{"storyboard://" + storyboard.ID}}, nil
		}
		if e.Generator == nil {
			return batchUnitOutcome{}, fmt.Errorf("%w: 未配置文本模型，无法生成分镜", ErrInvalidState)
		}
		result, generateErr := (&StoryboardService{Store: e.Store, Generator: e.Generator}).Generate(ctx, StoryboardRequest{
			Owner: owner, ProjectID: projectID, Engine: EngineXiaoyunqueV3, EpisodeNo: episode,
			Style: project.ArtStyle, Synopsis: project.Synopsis, Model: project.ModelID, Confirmed: true,
		})
		if generateErr != nil {
			return batchUnitOutcome{}, generateErr
		}
		artifacts := make([]string, 0, len(result.Storyboards))
		for _, storyboard := range result.Storyboards {
			artifacts = append(artifacts, "storyboard://"+storyboard.ID)
		}
		return batchUnitOutcome{artifacts: artifacts}, nil

	case "6_storyboard_doctor":
		episode := unit.Episode
		if episode <= 0 {
			episode = 1
		}
		storyboard, found := existingBatchStoryboard(ctx, e.Store, owner, projectID, episode)
		if !found || storyboard.ShotCount <= 0 || len(storyboard.Shots) == 0 {
			return batchUnitOutcome{}, fmt.Errorf("%w: 第 %d 集没有可审核的分镜", ErrArtifactInvalid, episode)
		}
		return e.doctor(ctx, project, unit, storyboard)

	case "7_video":
		if e.Media == nil {
			return batchUnitOutcome{}, fmt.Errorf("%w: 尚未配置漫剧视频生产适配器", ErrInvalidState)
		}
		episode := unit.Episode
		if episode <= 0 {
			episode = 1
		}
		storyboard, found := existingBatchStoryboard(ctx, e.Store, owner, projectID, episode)
		if !found {
			return batchUnitOutcome{}, fmt.Errorf("%w: 第 %d 集分镜不存在", ErrArtifactInvalid, episode)
		}
		artifacts, mediaErr := e.Media.ProduceEpisode(ctx, project, episode, storyboard)
		if mediaErr != nil {
			return batchUnitOutcome{}, mediaErr
		}
		if len(artifacts) == 0 {
			return batchUnitOutcome{}, fmt.Errorf("%w: 视频生产没有返回任何产物", ErrArtifactInvalid)
		}
		return batchUnitOutcome{artifacts: artifacts}, nil

	case "8_finalize":
		if e.Finalizer == nil {
			return batchUnitOutcome{}, ErrClipsReady
		}
		artifacts, err := e.Finalizer.Finalize(ctx, project)
		if err != nil {
			return batchUnitOutcome{}, err
		}
		if len(artifacts) == 0 {
			return batchUnitOutcome{}, fmt.Errorf("%w: 交付结果为空", ErrArtifactInvalid)
		}
		for _, a := range artifacts {
			if !(strings.HasPrefix(a, "https://") || strings.HasPrefix(a, "/api/")) {
				return batchUnitOutcome{}, fmt.Errorf("%w: 交付必须提供可访问媒体或工程", ErrArtifactInvalid)
			}
		}
		return batchUnitOutcome{artifacts: artifacts}, nil
	default:
		return batchUnitOutcome{}, fmt.Errorf("%w: 未知批次步骤 %s", ErrInvalidArgument, unit.StepKey)
	}
}

func normalizeBatchDegree(degree AssetDegree) AssetDegree {
	switch degree {
	case DegreeStandard, DegreeMoreDiff, DegreeMaster, DegreeMasterReal:
		return degree
	default:
		return DegreeMaster
	}
}

func existingBatchStoryboard(ctx context.Context, store Store, owner, projectID string, episode int) (Storyboard, bool) {
	for _, engine := range []StoryboardEngine{EngineXiaoyunqueV3, EngineXiaoluoV7} {
		row, err := store.GetStoryboard(ctx, owner, projectID, episode, engine)
		if err == nil {
			return row, true
		}
	}
	return Storyboard{}, false
}

var batchEpisodeHeading = regexp.MustCompile(`(?m)^#[ \t]*第[ \t]*([0-9]+)[ \t]*集`)

func countBatchEpisodes(content string) int {
	matches := batchEpisodeHeading.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		if strings.TrimSpace(content) == "" {
			return 0
		}
		return 1
	}
	seen := map[string]bool{}
	for _, match := range matches {
		if len(match) > 1 {
			seen[match[1]] = true
		}
	}
	numbers := make([]string, 0, len(seen))
	for number := range seen {
		numbers = append(numbers, number)
	}
	sort.Strings(numbers)
	return len(numbers)
}
