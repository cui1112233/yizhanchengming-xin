package screenplay

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"qiantie/backend/internal/mergeworker"
)

const assetCheckDueKey = "qiantie:screenplay:asset-check:due"

type AssetCheckStartRequest struct {
	RequestKey     string `json:"requestKey"`
	SourceMode     string `json:"sourceMode"`
	EpisodeNumbers []int  `json:"episodeNumbers"`
	Model          string `json:"model"`
	Confirmed      bool   `json:"confirmed"`
}
type AssetCheckRuntime struct {
	Store       Store
	Repository  AssetCheckRepository
	Authoring   *AuthoringService
	RedisDo     mergeworker.RedisDoFunc
	Generator   func(string) TextGenerator
	BuildSteps  func(Project, []AssetCheckEpisode, []Asset, string) ([]AssetCheckStep, error)
	Interval    time.Duration
	CallTimeout time.Duration
}

func (r *AssetCheckRuntime) Start(ctx context.Context, owner, projectID, actor string, in AssetCheckStartRequest) (AssetCheckJob, error) {
	if r == nil || r.Store == nil || r.Repository == nil || r.RedisDo == nil || r.Generator == nil || r.BuildSteps == nil {
		return AssetCheckJob{}, fmt.Errorf("%w: 资产检查后台尚未配置", ErrInvalidState)
	}
	if !storyboardGenerationKeyRe.MatchString(in.RequestKey) || len(in.EpisodeNumbers) == 0 || len(in.EpisodeNumbers) > 1000 {
		return AssetCheckJob{}, fmt.Errorf("%w: 请提供请求标识并选择有效分集", ErrInvalidArgument)
	}
	if in.SourceMode != "" && in.SourceMode != "script" && in.SourceMode != "original" {
		return AssetCheckJob{}, ErrInvalidArgument
	}
	mode := in.SourceMode
	if mode == "" {
		mode = "script"
	}
	selected := append([]int{}, in.EpisodeNumbers...)
	sort.Ints(selected)
	for i, n := range selected {
		if n <= 0 || i > 0 && n == selected[i-1] {
			return AssetCheckJob{}, ErrInvalidArgument
		}
	}
	project, err := r.Store.GetProject(ctx, owner, projectID)
	if err != nil {
		return AssetCheckJob{}, err
	}
	if project.PermissionMode == PermissionAuto && !in.Confirmed {
		return AssetCheckJob{}, fmt.Errorf("%w: 请明确确认使用文字模型检查资产", ErrInvalidArgument)
	}
	// A lost acknowledgement must recover the original frozen request even if
	// project source or assets changed afterwards. A key cannot name new input.
	previous, findErr := r.Repository.FindAssetCheckJobByRequestKey(ctx, owner, projectID, in.RequestKey)
	if findErr == nil {
		oldSelection := previous.SelectedEpisodeNumbers
		if len(oldSelection) == 0 {
			for _, ep := range previous.Episodes {
				oldSelection = append(oldSelection, ep.EpisodeNo)
			}
			sort.Ints(oldSelection)
		}
		if previous.SourceMode != mode || in.Model != "" && in.Model != previous.Model || fmt.Sprint(oldSelection) != fmt.Sprint(selected) {
			return AssetCheckJob{}, ErrRevisionConflict
		}
		return previous, nil
	}
	if !errors.Is(findErr, ErrNotFound) {
		return AssetCheckJob{}, findErr
	}
	model := firstNonEmpty(in.Model, project.ModelID)
	if model == "" {
		return AssetCheckJob{}, fmt.Errorf("%w: 请选择自己的文字模型", ErrInvalidArgument)
	}
	episodes, err := r.sourceEpisodes(ctx, owner, project, in)
	if err != nil {
		return AssetCheckJob{}, err
	}
	existing, err := r.Store.ListAssets(ctx, owner, projectID, "")
	if err != nil {
		return AssetCheckJob{}, err
	}
	steps, err := r.BuildSteps(project, episodes, existing, model)
	if err != nil {
		return AssetCheckJob{}, err
	}
	if len(steps) == 0 {
		return AssetCheckJob{}, ErrInvalidState
	}
	if _, err = r.RedisDo(ctx, "PING"); err != nil {
		return AssetCheckJob{}, fmt.Errorf("%w: 检查队列暂不可用，尚未发起模型请求", ErrInvalidState)
	}
	job := AssetCheckJob{ID: newScreenplayID("asset_check"), Owner: owner, ProjectID: projectID, Actor: actor, RequestKey: in.RequestKey, SourceMode: mode, SelectedEpisodeNumbers: selected, Model: model, Degree: normalizeBatchDegree(project.AssetDegree), Status: "queued", Episodes: episodes, Steps: steps, Existing: existing, Candidates: []AssetCheckCandidate{}}
	job.InputDigest = generationFingerprint(map[string]any{"source_mode": mode, "model": model, "episodes": episodes, "steps": steps, "degree": job.Degree, "asset_prompt_type": EffectiveCreativeConfig(project).AssetPromptType}, "", "", "")
	job, err = r.Repository.CreateAssetCheckJob(ctx, job)
	if err != nil {
		return job, err
	}
	if job.Status == "queued" || job.Status == "running" {
		notify, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		_, _ = r.RedisDo(notify, "ZADD", assetCheckDueKey, strconv.FormatInt(job.CreatedAt.UnixMilli(), 10), job.ID)
		cancel()
	}
	return job, nil
}

func (r *AssetCheckRuntime) sourceEpisodes(ctx context.Context, owner string, p Project, in AssetCheckStartRequest) ([]AssetCheckEpisode, error) {
	if in.SourceMode != "" && in.SourceMode != "original" && in.SourceMode != "script" {
		return nil, ErrInvalidArgument
	}
	available := map[int]AssetCheckEpisode{}
	if r.Authoring != nil {
		state, err := r.Authoring.Load(ctx, owner, p.ID)
		if err != nil {
			return nil, err
		}
		if state.Archived {
			return nil, ErrInvalidState
		}
		for i, c := range state.Chapters {
			content := firstNonEmpty(c.ScriptText, c.SourceText)
			if in.SourceMode == "original" {
				content = c.SourceText
			}
			available[i+1] = AssetCheckEpisode{ChapterIndex: i, EpisodeNo: i + 1, ChapterID: c.ID, Title: firstNonEmpty(c.Title, fmt.Sprintf("第%d集", i+1)), Content: content}
		}
	}
	if len(available) == 0 {
		if in.SourceMode == "original" {
			available[1] = AssetCheckEpisode{EpisodeNo: 1, Title: "第1集", Content: p.SourceText}
		} else {
			doc, e := r.Store.ActiveDocument(ctx, owner, p.ID)
			if e != nil {
				return nil, e
			}
			episodes := splitEpisodes(doc.Content)
			if len(episodes) == 0 {
				episodes = []episodeInput{{No: 1, Content: doc.Content}}
			}
			for _, ep := range episodes {
				available[ep.No] = AssetCheckEpisode{ChapterIndex: ep.No - 1, EpisodeNo: ep.No, Title: fmt.Sprintf("第%d集", ep.No), Content: ep.Content}
			}
		}
	}
	seen := map[int]bool{}
	numbers := append([]int{}, in.EpisodeNumbers...)
	sort.Ints(numbers)
	out := []AssetCheckEpisode{}
	for _, n := range numbers {
		if n <= 0 || seen[n] {
			return nil, ErrInvalidArgument
		}
		seen[n] = true
		ep, ok := available[n]
		if !ok {
			return nil, fmt.Errorf("%w: 第%d集不存在", ErrInvalidArgument, n)
		}
		if strings.TrimSpace(ep.Content) != "" {
			out = append(out, ep)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: 请至少选择一集有剧本内容的分集", ErrEmptyContent)
	}
	return out, nil
}

func (r *AssetCheckRuntime) Cancel(ctx context.Context, o, p, id string) (AssetCheckJob, error) {
	return r.Repository.MutateAssetCheckJob(ctx, o, p, id, func(j *AssetCheckJob) error {
		if j.Status != "queued" && j.Status != "running" && j.Status != "unknown" {
			return nil
		}
		j.CancelRequested = true
		if j.Status == "queued" || j.Status == "unknown" {
			j.Status = "cancelled"
			j.LeaseToken = ""
			for i := range j.Steps {
				if j.Steps[i].Status == "received" {
					candidates, err := parseAssetCheckCandidates(*j, j.Steps[i], j.Steps[i].RawResponse)
					if err != nil {
						j.Steps[i].Status = "failed"
						j.Steps[i].Error = err.Error()
					} else {
						j.Steps[i].Status = "completed"
						j.Steps[i].Candidates = candidates
					}
				}
				if j.Steps[i].Status == "pending" {
					j.Steps[i].Status = "cancelled"
				}
			}
			for _, ep := range j.Episodes {
				recordAssetCheckEpisodeCompletion(j, ep.ChapterIndex)
			}
			j.Candidates = collectAssetCheckCandidates(*j)
		}
		return nil
	})
}
func (r *AssetCheckRuntime) Discard(ctx context.Context, o, p, id string, revision int64) (AssetCheckJob, error) {
	return r.Repository.MutateAssetCheckJob(ctx, o, p, id, func(j *AssetCheckJob) error {
		if j.Discarded {
			return nil
		}
		if j.Status == "queued" || j.Status == "running" || j.Applied {
			return ErrInvalidState
		}
		if j.Revision != revision {
			return ErrRevisionConflict
		}
		j.Discarded = true
		return nil
	})
}

func (r *AssetCheckRuntime) Run(ctx context.Context) error {
	if r.Repository == nil || r.RedisDo == nil || r.Generator == nil {
		return ErrInvalidState
	}
	interval := r.Interval
	if interval <= 0 {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	slots := make(chan struct{}, 4)
	var wg sync.WaitGroup
	defer wg.Wait()
	for ctx.Err() == nil {
		jobs, err := r.Repository.ListRunnableAssetCheckJobs(ctx)
		if err == nil {
			known := map[string]AssetCheckJob{}
			for _, job := range jobs {
				known[job.ID] = job
				_, _ = r.RedisDo(ctx, "ZADD", assetCheckDueKey, strconv.FormatInt(job.CreatedAt.UnixMilli(), 10), job.ID)
			}
			v, e := r.RedisDo(ctx, "ZRANGEBYSCORE", assetCheckDueKey, "-inf", strconv.FormatInt(time.Now().UnixMilli(), 10), "LIMIT", "0", "100")
			if e == nil {
				if items, ok := v.([]any); ok {
					for _, item := range items {
						id := fmt.Sprint(item)
						job, ok := known[id]
						if !ok {
							_, _ = r.RedisDo(ctx, "ZREM", assetCheckDueKey, id)
							continue
						}
						select {
						case slots <- struct{}{}:
							wg.Add(1)
							go func(job AssetCheckJob) { defer wg.Done(); defer func() { <-slots }(); _ = r.RunOne(ctx, job) }(job)
						default:
						}
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
	return ctx.Err()
}

func (r *AssetCheckRuntime) RunOne(ctx context.Context, input AssetCheckJob) error {
	key := "qiantie:screenplay:asset-check:lease:" + contentDigest(input.Owner+"\x00"+input.ProjectID)
	token := newScreenplayID("check_lease")
	v, err := r.RedisDo(ctx, "SET", key, token, "NX", "PX", "30000")
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = r.RedisDo(cleanup, "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) else return 0 end`, "1", key, token)
	}()
	renew := func(c context.Context) error {
		v, e := r.RedisDo(c, "EVAL", `if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('PEXPIRE',KEYS[1],ARGV[2]) else return 0 end`, "1", key, token, "30000")
		if e != nil {
			return e
		}
		if fmt.Sprint(v) != "1" {
			return ErrLockBusy
		}
		return nil
	}
	job, err := r.Repository.MutateAssetCheckJob(ctx, input.Owner, input.ProjectID, input.ID, func(j *AssetCheckJob) error {
		if j.Status != "queued" && j.Status != "running" {
			return ErrInvalidState
		}
		for i := range j.Steps {
			if j.Steps[i].Status == "calling" {
				j.Steps[i].Status = "unknown"
				j.Steps[i].Error = "上次请求已发出，结果尚未核对，未自动重试"
				j.Status = "unknown"
			}
		}
		if j.Status == "unknown" {
			return nil
		}
		j.Status = "running"
		j.LeaseToken = token
		return nil
	})
	if err != nil {
		return err
	}
	if job.Status != "running" {
		return nil
	}
	if err = renew(ctx); err != nil {
		return err
	}
	_, _ = r.RedisDo(ctx, "ZREM", assetCheckDueKey, job.ID)
	work, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	defer close(done)
	var lost atomic.Bool
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-work.Done():
				return
			case <-tick.C:
				current, e := r.Repository.GetAssetCheckJob(work, job.Owner, job.ProjectID, job.ID)
				if e != nil || current.LeaseToken != token {
					cancel()
					return
				}
				if e = renew(work); e != nil {
					lost.Store(true)
					cancel()
					return
				}
			}
		}
	}()
	var wg sync.WaitGroup
	slots := make(chan struct{}, 3)
episodeLoop:
	for _, episode := range job.Episodes {
		if work.Err() != nil {
			break
		}
		select {
		case slots <- struct{}{}:
			current, readErr := r.Repository.GetAssetCheckJob(work, job.Owner, job.ProjectID, job.ID)
			if readErr != nil || current.CancelRequested || current.LeaseToken != token {
				<-slots
				break episodeLoop
			}
			wg.Add(1)
			go func(ep AssetCheckEpisode) {
				defer wg.Done()
				defer func() { <-slots }()
				r.runEpisode(work, job, token, ep, renew)
			}(episode)
		case <-work.Done():
		}
	}
	wg.Wait()
	saveCtx, saveCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer saveCancel()
	_, err = r.Repository.MutateAssetCheckJob(saveCtx, job.Owner, job.ProjectID, job.ID, func(j *AssetCheckJob) error {
		if j.LeaseToken != token || j.Status != "running" {
			return ErrLockBusy
		}
		unknown, pending, failed, completed := false, false, 0, 0
		for i := range j.Steps {
			step := &j.Steps[i]
			switch step.Status {
			case "calling":
				step.Status = "unknown"
				step.Error = "模型请求结果待核对"
				unknown = true
			case "unknown":
				unknown = true
			case "pending", "received":
				pending = true
			case "failed":
				failed++
			case "completed":
				completed++
			}
			if j.CancelRequested && step.Status == "pending" {
				step.Status = "cancelled"
			}
		}
		switch {
		case j.CancelRequested:
			j.Status = "cancelled"
		case unknown:
			j.Status = "unknown"
		case ctx.Err() != nil || lost.Load() || pending:
			j.Status = "queued"
		case failed > 0 && completed == 0:
			j.Status = "failed"
		default:
			j.Status = "completed"
		}
		j.LeaseToken = ""
		if failed > 0 {
			j.Error = fmt.Sprintf("%d项资产检查失败，请查看对应分集；成功结果已保留", failed)
		}
		j.Candidates = collectAssetCheckCandidates(*j)
		return nil
	})
	return err
}

func (r *AssetCheckRuntime) runEpisode(ctx context.Context, job AssetCheckJob, token string, episode AssetCheckEpisode, renew func(context.Context) error) {
	// The original schedules three episodes, with all three asset kinds inside
	// each episode in parallel. Cancellation stops new episodes, not their results.
	var workers sync.WaitGroup
	for i, initial := range job.Steps {
		if initial.ChapterIndex != episode.ChapterIndex {
			continue
		}
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			r.runStep(ctx, job, token, index, renew)
		}(i)
	}
	workers.Wait()
}

func (r *AssetCheckRuntime) runStep(ctx context.Context, job AssetCheckJob, token string, i int, renew func(context.Context) error) {
	if ctx.Err() != nil {
		return
	}
	current, err := r.Repository.GetAssetCheckJob(ctx, job.Owner, job.ProjectID, job.ID)
	if err != nil || current.LeaseToken != token {
		return
	}
	step := current.Steps[i]
	if step.Status == "completed" || step.Status == "failed" || step.Status == "cancelled" || step.Status == "unknown" {
		return
	}
	if r.Authoring != nil {
		owner, e := r.Authoring.Resolve(ctx, job.Actor, job.ProjectID, "generate")
		if e != nil || owner != job.Owner {
			_ = r.stepUpdate(ctx, job, token, i, func(s *AssetCheckStep) { s.Status = "failed"; s.Error = "当前用户已无检查权限" })
			return
		}
	}
	if step.Status == "pending" {
		if err = renew(ctx); err != nil {
			return
		}
		err = r.stepUpdate(ctx, job, token, i, func(s *AssetCheckStep) { s.Status = "calling"; now := time.Now().UTC(); s.StartedAt = &now })
		if err != nil {
			return
		}
		timeout := r.CallTimeout
		if timeout <= 0 {
			timeout = 300 * time.Second
		}
		callCtx, callCancel := context.WithTimeout(ctx, timeout)
		// Verify the Redis lease again after the journal write, immediately
		// before crossing the billable provider boundary.
		if err = renew(callCtx); err != nil || callCtx.Err() != nil {
			callCancel()
			_ = r.stepUpdate(context.WithoutCancel(ctx), job, token, i, func(s *AssetCheckStep) { s.Status = "pending"; s.StartedAt = nil })
			return
		}
		response, callErr := r.Generator(job.Owner).Generate(callCtx, step.Request)
		callCancel()
		err = r.stepUpdate(context.WithoutCancel(ctx), job, token, i, func(s *AssetCheckStep) {
			now := time.Now().UTC()
			s.FinishedAt = &now
			if callErr != nil {
				s.Status = "unknown"
				s.Error = callErr.Error()
				if errors.Is(callErr, ErrAPIKeyMissing) || errors.Is(callErr, ErrInvalidArgument) {
					s.Status = "failed"
				}
				return
			}
			if len(response.Content) > 8<<20 {
				s.Status = "unknown"
				s.Error = "模型结果过长，需核对"
				return
			}
			s.RawResponse = response.Content
			s.Status = "received"
		})
		if err != nil || callErr != nil {
			return
		}
		step.RawResponse = response.Content
	}
	candidates, parseErr := parseAssetCheckCandidates(job, step, step.RawResponse)
	if err = r.stepUpdate(context.WithoutCancel(ctx), job, token, i, func(s *AssetCheckStep) {
		if parseErr != nil {
			s.Status = "failed"
			s.Error = parseErr.Error()
		} else {
			s.Status = "completed"
			s.Candidates = candidates
		}
	}); err != nil {
		return
	}
}

func (r *AssetCheckRuntime) stepUpdate(ctx context.Context, job AssetCheckJob, token string, index int, update func(*AssetCheckStep)) error {
	bounded, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := r.Repository.MutateAssetCheckJob(bounded, job.Owner, job.ProjectID, job.ID, func(j *AssetCheckJob) error {
		if j.LeaseToken != token || j.Status != "running" {
			return ErrLockBusy
		}
		update(&j.Steps[index])
		switch j.Steps[index].Status {
		case "completed", "failed", "unknown":
			now := time.Now().UTC()
			j.Steps[index].FinishedAt = &now
		}
		recordAssetCheckEpisodeCompletion(j, j.Steps[index].ChapterIndex)
		j.Candidates = collectAssetCheckCandidates(*j)
		return nil
	})
	return err
}

func recordAssetCheckEpisodeCompletion(j *AssetCheckJob, chapter int) {
	finished := true
	for _, step := range j.Steps {
		if step.ChapterIndex == chapter && (step.Status == "pending" || step.Status == "calling" || step.Status == "received") {
			finished = false
		}
	}
	if finished {
		found := false
		for _, previous := range j.EpisodeCompletionOrder {
			if previous == chapter {
				found = true
			}
		}
		if !found {
			j.EpisodeCompletionOrder = append(j.EpisodeCompletionOrder, chapter)
		}
	}
}

func collectAssetCheckCandidates(job AssetCheckJob) []AssetCheckCandidate {
	out := []AssetCheckCandidate{}
	seen := map[string]bool{}
	for _, asset := range job.Existing {
		seen[assetCheckNameKey(asset)] = true
	}
	for _, chapter := range job.EpisodeCompletionOrder {
		for _, step := range job.Steps {
			if step.ChapterIndex != chapter {
				continue
			}
			for _, candidate := range step.Candidates {
				key := assetCheckNameKey(candidate.Asset)
				if !seen[key] {
					seen[key] = true
					out = append(out, candidate)
				}
			}
		}
	}
	return out
}
