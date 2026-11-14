package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var storyboardGenerationKeyRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)

type StoryboardGenerationEpisodeSummary struct {
	Scenes    []StoryboardGenerationSceneSummary `json:"scenes,omitempty"`
	EpisodeNo int                                `json:"episode_no"`
	Status    string                             `json:"status"`
	Error     string                             `json:"error,omitempty"`
}

type StoryboardGenerationSceneSummary struct {
	Index  int    `json:"index"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
type StoryboardGenerationSummary struct {
	RequestKey        string                               `json:"request_key"`
	RetryOf           string                               `json:"retry_of,omitempty"`
	InputDigest       string                               `json:"input_digest"`
	Engine            StoryboardEngine                     `json:"engine"`
	Status            string                               `json:"status"`
	ChapterID         string                               `json:"chapter_id,omitempty"`
	PlanningChapterID string                               `json:"planning_chapter_id,omitempty"`
	UseOriginal       bool                                 `json:"use_original_content_directly"`
	UsePlanning       *bool                                `json:"use_planning"`
	PlanningApplied   bool                                 `json:"planning_applied"`
	EpisodeNumbers    []int                                `json:"episode_numbers"`
	Episodes          []StoryboardGenerationEpisodeSummary `json:"episodes"`
	Error             string                               `json:"error,omitempty"`
	CanRetry          bool                                 `json:"can_retry"`
	CancelRequested   bool                                 `json:"cancel_requested"`
	CreatedAt         time.Time                            `json:"created_at"`
	UpdatedAt         time.Time                            `json:"updated_at"`
}

func (a StoryboardGenerationAttempt) Summary() StoryboardGenerationSummary {
	out := StoryboardGenerationSummary{RequestKey: a.RequestKey, RetryOf: a.RetryOf, InputDigest: a.InputDigest, Engine: a.Engine, Status: a.Status, ChapterID: a.ChapterID, PlanningChapterID: a.PlanningChapterID, UseOriginal: a.UseOriginal, EpisodeNumbers: a.EpisodeNumbers, Error: a.Error, CancelRequested: a.CancelRequested, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, Episodes: []StoryboardGenerationEpisodeSummary{}}
	out.CanRetry = a.Status == "failed" || a.Status == "partial" || a.Status == "cancelled"
	out.UsePlanning = a.UsePlanning
	out.PlanningApplied = a.PlanningApplied
	for _, n := range a.EpisodeNumbers {
		ep := a.Episodes[n]
		if ep == nil {
			continue
		}
		scenes := []StoryboardGenerationSceneSummary{}
		count := ep.SceneCount
		for phase := range ep.Calls {
			if strings.HasPrefix(phase, "storyboard_scene_") {
				if i, e := strconv.Atoi(strings.TrimPrefix(phase, "storyboard_scene_")); e == nil && i > count {
					count = i
				}
			}
		}
		for i := 1; i <= count; i++ {
			phase := fmt.Sprintf("storyboard_scene_%d", i)
			if count == 1 {
				phase = "storyboard"
			}
			call, ok := ep.Calls[phase]
			status := "pending"
			if ok {
				status = call.Status
			}
			scenes = append(scenes, StoryboardGenerationSceneSummary{Index: i, Status: status, Error: call.Error})
		}
		out.Episodes = append(out.Episodes, StoryboardGenerationEpisodeSummary{EpisodeNo: n, Status: ep.Status, Error: ep.Error, Scenes: scenes})
		if ep.Status == "unknown" || ep.Status == "calling" || ep.Status == "saving" {
			out.CanRetry = false
		}
		for _, call := range ep.Calls {
			if call.Status == "calling" || call.Status == "unknown" {
				out.CanRetry = false
			}
		}
	}
	return out
}
func generationAttemptResult(a StoryboardGenerationAttempt) StoryboardResult {
	out := StoryboardResult{EpisodeCount: len(a.EpisodeNumbers), Storyboards: []Storyboard{}, Warnings: []string{}}
	for _, n := range a.EpisodeNumbers {
		ep := a.Episodes[n]
		if ep == nil {
			continue
		}
		out.AddedAssetCount += len(ep.AddedAssetIDs)
		out.Warnings = append(out.Warnings, ep.Warnings...)
		if ep.Status == "succeeded" && ep.Storyboard != nil {
			out.Storyboards = append(out.Storyboards, *ep.Storyboard)
			out.GroupCount += len(ep.Storyboard.Shots)
		} else if a.Status != "running" {
			out.FailedEpisodes = append(out.FailedEpisodes, n)
		}
	}
	out.PartialFailure = len(out.Storyboards) > 0 && len(out.FailedEpisodes) > 0
	return out
}
func (a StoryboardGenerationAttempt) Public() map[string]any {
	raw, _ := json.Marshal(a.Summary())
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	result := generationAttemptResult(a)
	summary := a.Summary()
	result.GenerationAttempt = &summary
	out["result"] = result
	return out
}
func (a StoryboardGenerationAttempt) replay() (StoryboardResult, error) {
	result := generationAttemptResult(a)
	summary := a.Summary()
	result.GenerationAttempt = &summary
	switch a.Status {
	case "succeeded":
		return result, nil
	case "partial":
		return result, fmt.Errorf("%w: %s", ErrPartialFailure, a.Error)
	case "cancelled":
		return result, fmt.Errorf("%w: %s", ErrUserCancelled, a.Error)
	default:
		return result, fmt.Errorf("%w: 生成请求%s当前为%s，请先核对已保存结果；不会自动重复提交", ErrInvalidState, a.RequestKey, a.Status)
	}
}

type storyboardGenerationSession struct {
	store               StoryboardGenerationStore
	owner, project, key string
	attempt             StoryboardGenerationAttempt
}

func generationFingerprint(base map[string]any, roles, scenes, props string) string {
	copy := map[string]any{}
	for k, v := range base {
		copy[k] = v
	}
	copy["roles"] = roles
	copy["scenes"] = scenes
	copy["props"] = props
	raw, _ := json.Marshal(copy)
	return contentDigest(string(raw))
}
func generationLibrary(assets []Asset, prior *StoryboardGenerationAttempt) (string, string, string) {
	excluded := map[string]bool{}
	if prior != nil {
		for _, ep := range prior.Episodes {
			for _, id := range ep.AddedAssetIDs {
				excluded[id] = true
			}
			for _, asset := range ep.PendingAssets {
				excluded[asset.ID] = true
			}
		}
	}
	byType := map[AssetType][]Asset{}
	for _, a := range assets {
		if !excluded[a.ID] {
			byType[a.Type] = append(byType[a.Type], a)
		}
	}
	return formatAssetLibrary(byType[AssetCharacter]), formatAssetLibrary(byType[AssetScene]), formatAssetLibrary(byType[AssetProp])
}
func beginStoryboardGeneration(ctx context.Context, store StoryboardGenerationStore, req StoryboardRequest, engine StoryboardEngine, numbers []int, base map[string]any, assets []Asset, external map[int][]Storyboard, current []Storyboard) (*storyboardGenerationSession, *StoryboardGenerationAttempt, error) {
	var started *StoryboardGenerationAttempt
	fresh := false
	_, e := store.MutateStoryboardGenerations(ctx, req.Owner, req.ProjectID, func(state *StoryboardGenerationState) error {
		old := state.Attempts[req.IdempotencyKey]
		prior := old
		if prior == nil && req.RetryOf != "" {
			prior = state.Attempts[req.RetryOf]
			if prior == nil {
				return ErrNotFound
			}
		}
		roles, scenes, props := generationLibrary(assets, prior)
		digest := generationFingerprint(base, roles, scenes, props)
		if old != nil {
			started = old
			if old.InputDigest != digest {
				return fmt.Errorf("%w: 同一生成请求的已保存输入已变化，请核对原结果后另行明确生成", ErrRevisionConflict)
			}
			return nil
		}
		targets := map[int]bool{}
		for _, n := range numbers {
			targets[n] = true
		}
		for _, a := range state.Attempts {
			if a.Engine != engine {
				continue
			}
			for n, ep := range a.Episodes {
				if !targets[n] {
					continue
				}
				blocked := ep.Status == "calling" || ep.Status == "unknown" || ep.Status == "saving"
				for _, call := range ep.Calls {
					blocked = blocked || call.Status == "calling" || call.Status == "unknown"
				}
				if a.Status == "running" || blocked {
					started = a
					return fmt.Errorf("%w: 当前目标已有%s生成请求，请先核对%s，不能换标识重复提交", ErrInvalidState, a.Status, a.RequestKey)
				}
			}
		}
		if req.RetryOf != "" {
			if !req.ConfirmedRetry {
				return fmt.Errorf("%w: 续跑须用户明确确认", ErrInvalidArgument)
			}
			if prior.InputDigest != digest || prior.Engine != engine || !sameGenerationEpisodes(prior.EpisodeNumbers, numbers) {
				started = prior
				return fmt.Errorf("%w: 续跑的完整目标和输入须与原请求一致", ErrRevisionConflict)
			}
			if !prior.Summary().CanRetry {
				started = prior
				return fmt.Errorf("%w: 原请求状态%s不可自动重付", ErrInvalidState, prior.Status)
			}
			for _, ep := range prior.Episodes {
				if ep.Status != "succeeded" || ep.Storyboard == nil {
					continue
				}
				matches := false
				for _, board := range current {
					if board.EpisodeNo == ep.EpisodeNo && board.ID == ep.Storyboard.ID && board.Revision == ep.Storyboard.Revision {
						matches = true
						break
					}
				}
				if !matches {
					started = prior
					return fmt.Errorf("%w: 原请求的已成功分镜后来被修改，不能混用旧上下文续跑", ErrRevisionConflict)
				}
			}
		} else {
			for _, a := range state.Attempts {
				if a.Engine != engine || !sameGenerationEpisodes(a.EpisodeNumbers, numbers) || (a.Status != "failed" && a.Status != "partial" && a.Status != "cancelled") {
					continue
				}
				r, sc, pr := generationLibrary(assets, a)
				if generationFingerprint(base, r, sc, pr) == a.InputDigest {
					started = a
					return fmt.Errorf("%w: 同输入已有失败或部分结果，请核对后使用retryOf明确续跑，不能直接换标识重付", ErrInvalidState)
				}
			}
		}
		now := time.Now().UTC()
		a := &StoryboardGenerationAttempt{RequestKey: req.IdempotencyKey, RetryOf: req.RetryOf, InputDigest: digest, Engine: engine, Status: "running", ChapterID: req.ChapterID, PlanningChapterID: req.PlanningChapterID, UseOriginal: req.UseOriginalContentDirectly, EpisodeNumbers: numbers, Episodes: map[int]*StoryboardGenerationEpisode{}, CreatedAt: now, UpdatedAt: now, Roles: roles, Scenes: scenes, Props: props, ExternalHistory: external}
		a.UsePlanning = req.UsePlanning
		a.PromptID, _ = base["prompt_id"].(string)
		a.PromptDigest, _ = base["prompt_digest"].(string)
		a.OutputProtocol, _ = base["output_protocol"].(string)
		a.ParserVersion, _ = base["parser_version"].(string)
		if planning, ok := base["planning"].(map[int]string); ok {
			for _, text := range planning {
				a.PlanningApplied = a.PlanningApplied || text != ""
			}
		}
		for _, n := range numbers {
			ep := &StoryboardGenerationEpisode{EpisodeNo: n, Status: "pending", Calls: map[string]StoryboardGenerationCall{}}
			if counts, ok := base["scene_counts"].(map[int]int); ok {
				ep.SceneCount = counts[n]
			}
			if req.RetryOf != "" {
				parent := prior.Episodes[n]
				if parent != nil && parent.Status == "succeeded" {
					raw, _ := json.Marshal(parent)
					_ = json.Unmarshal(raw, ep)
				} else if parent != nil {
					for phase, call := range parent.Calls {
						if (phase == "storyboard" || strings.HasPrefix(phase, "storyboard_scene_")) && call.Status == "received" {
							if _, e := parseStoryboardJSON(call.RawResponse); e == nil {
								ep.Calls[phase] = call
							}
						}
					}
					ep.AddedAssetIDs = append([]string{}, parent.AddedAssetIDs...)
					ep.PendingAssets = append([]Asset{}, parent.PendingAssets...)
					if len(ep.PendingAssets) > 0 {
						if call, ok := parent.Calls["asset_sync"]; ok && call.Status == "received" {
							ep.Calls["asset_sync"] = call
						}
					}
				}
			}
			a.Episodes[n] = ep
		}
		a.Result = generationAttemptResult(*a)
		state.Attempts[req.IdempotencyKey] = a
		started = a
		fresh = true
		return nil
	})
	if e != nil {
		return nil, started, e
	}
	if !fresh {
		return nil, started, nil
	}
	return &storyboardGenerationSession{store: store, owner: req.Owner, project: req.ProjectID, key: req.IdempotencyKey, attempt: *started}, nil, nil
}
func sameGenerationEpisodes(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func (s *storyboardGenerationSession) mutate(ctx context.Context, fn func(*StoryboardGenerationAttempt) error) error {
	state, e := s.store.MutateStoryboardGenerations(ctx, s.owner, s.project, func(v *StoryboardGenerationState) error {
		a := v.Attempts[s.key]
		if a == nil {
			return ErrNotFound
		}
		if e := fn(a); e != nil {
			return e
		}
		a.UpdatedAt = time.Now().UTC()
		return nil
	})
	if e == nil {
		s.attempt = *state.Attempts[s.key]
	}
	return e
}
func (s *storyboardGenerationSession) call(ctx context.Context, episode int, phase string, request TextRequest, generator TextGenerator) (TextResponse, error) {
	raw, _ := json.Marshal(request)
	digest := contentDigest(string(raw))
	if cached, ok := s.attempt.Episodes[episode].Calls[phase]; ok && cached.Status == "received" {
		if cached.InputDigest != digest {
			return TextResponse{}, fmt.Errorf("%w: 已收到模型结果的输入与续跑不一致", ErrRevisionConflict)
		}
		return TextResponse{Content: cached.RawResponse, Model: request.Model}, nil
	}
	if e := s.mutate(ctx, func(a *StoryboardGenerationAttempt) error {
		ep := a.Episodes[episode]
		if a.Status != "running" || ep == nil {
			return ErrInvalidState
		}
		ep.Status = "calling"
		ep.Calls[phase] = StoryboardGenerationCall{Status: "calling", InputDigest: digest}
		return nil
	}); e != nil {
		return TextResponse{}, e
	}
	response, callErr := generator.Generate(ctx, request)
	saveErr := s.mutate(context.WithoutCancel(ctx), func(a *StoryboardGenerationAttempt) error {
		ep := a.Episodes[episode]
		call := ep.Calls[phase]
		if callErr != nil {
			call.Status = "unknown"
			if errors.Is(callErr, ErrAPIKeyMissing) || errors.Is(callErr, ErrInvalidArgument) {
				call.Status = "failed"
			}
			call.Error = callErr.Error()
			ep.Status = call.Status
			ep.Error = call.Error
		} else {
			call.Status = "received"
			if len(response.Content) > 8<<20 {
				call.Status = "unknown"
				call.Error = "模型返回超过安全存储上限，结果需人工核对"
				ep.Status = "unknown"
			} else {
				call.RawResponse = response.Content
				ep.Status = "received"
			}
		}
		ep.Calls[phase] = call
		return nil
	})
	if saveErr != nil {
		return TextResponse{}, saveErr
	}
	if callErr != nil {
		return response, callErr
	}
	if ctx.Err() != nil {
		return response, ctx.Err()
	}
	if s.attempt.Episodes[episode].Calls[phase].Status == "unknown" {
		return response, ErrInvalidState
	}
	return response, nil
}
func (s *StoryboardService) generationCall(ctx context.Context, req StoryboardRequest, episode int, phase string, input TextRequest) (TextResponse, error) {
	if req.generationSession == nil {
		return TextResponse{}, ErrInvalidState
	}
	return req.generationSession.call(ctx, episode, phase, input, s.Generator)
}
func (s *storyboardGenerationSession) recordAssets(ctx context.Context, episode int, assets []Asset) error {
	return s.mutate(context.WithoutCancel(ctx), func(a *StoryboardGenerationAttempt) error {
		for _, asset := range assets {
			a.Episodes[episode].AddedAssetIDs = append(a.Episodes[episode].AddedAssetIDs, asset.ID)
		}
		return nil
	})
}

func (s *storyboardGenerationSession) recordAssetPlan(ctx context.Context, episode int, assets []Asset) error {
	return s.mutate(context.WithoutCancel(ctx), func(a *StoryboardGenerationAttempt) error {
		a.Episodes[episode].PendingAssets = append([]Asset{}, assets...)
		return nil
	})
}

func (s *storyboardGenerationSession) episodeError(episode int, err error) error {
	return s.mutate(context.Background(), func(a *StoryboardGenerationAttempt) error {
		ep := a.Episodes[episode]
		ep.Error = err.Error()
		ep.Warnings = append(ep.Warnings, fmt.Sprintf("第%d集生成失败：%v", episode, err))
		return nil
	})
}
func (s *storyboardGenerationSession) finish(ctx context.Context, result *StoryboardResult, runErr *error) {
	e := s.mutate(context.WithoutCancel(ctx), func(a *StoryboardGenerationAttempt) error {
		unknown := false
		success := 0
		for _, n := range a.EpisodeNumbers {
			ep := a.Episodes[n]
			if ep.Status == "succeeded" {
				success++
				continue
			}
			uncertain := false
			for _, call := range ep.Calls {
				uncertain = uncertain || call.Status == "unknown" || call.Status == "calling"
			}
			if uncertain {
				ep.Status = "unknown"
				unknown = true
			} else if ctx.Err() != nil {
				ep.Status = "cancelled"
			} else {
				ep.Status = "failed"
			}
			if ep.Error == "" && *runErr != nil {
				ep.Error = (*runErr).Error()
			}
		}
		a.CancelRequested = ctx.Err() != nil
		switch {
		case success == len(a.EpisodeNumbers):
			a.Status = "succeeded"
		case unknown:
			a.Status = "unknown"
		case ctx.Err() != nil:
			a.Status = "cancelled"
		case success > 0:
			a.Status = "partial"
		default:
			a.Status = "failed"
		}
		if *runErr != nil {
			a.Error = (*runErr).Error()
		}
		a.Result = generationAttemptResult(*a)
		return nil
	})
	if e != nil {
		if *runErr == nil {
			*runErr = e
		}
		return
	}
	*result = generationAttemptResult(s.attempt)
	summary := s.attempt.Summary()
	result.GenerationAttempt = &summary
	if s.attempt.Status == "succeeded" {
		*runErr = nil
	}
}
func generationExternalHistory(boards []Storyboard, numbers []int) map[int][]Storyboard {
	targets := map[int]bool{}
	for _, n := range numbers {
		targets[n] = true
	}
	sort.SliceStable(boards, func(i, j int) bool { return boards[i].EpisodeNo < boards[j].EpisodeNo })
	out := map[int][]Storyboard{}
	for _, n := range numbers {
		rows := []Storyboard{}
		for _, board := range boards {
			if board.EpisodeNo >= n || targets[board.EpisodeNo] {
				continue
			}
			shots := append([]StoryboardShot{}, board.Shots...)
			sort.SliceStable(shots, func(i, j int) bool { return shots[i].Seq < shots[j].Seq })
			for _, shot := range shots {
				rows = append(rows, Storyboard{EpisodeNo: board.EpisodeNo, Shots: []StoryboardShot{shot}})
			}
		}
		if len(rows) > 5 {
			rows = rows[len(rows)-5:]
		}
		out[n] = rows
	}
	return out
}
func (s *storyboardGenerationSession) history(episode int) string {
	boards := append([]Storyboard{}, s.attempt.ExternalHistory[episode]...)
	for _, n := range s.attempt.EpisodeNumbers {
		if n >= episode {
			break
		}
		ep := s.attempt.Episodes[n]
		if ep.Status == "succeeded" && ep.Storyboard != nil {
			boards = append(boards, *ep.Storyboard)
		}
	}
	sort.SliceStable(boards, func(i, j int) bool { return boards[i].EpisodeNo < boards[j].EpisodeNo })
	rows := []StoryboardShot{}
	for _, board := range boards {
		rows = append(rows, board.Shots...)
	}
	if len(rows) > 5 {
		rows = rows[len(rows)-5:]
	}
	if len(rows) == 0 {
		return ""
	}
	raw, _ := json.Marshal(storyboardSemanticRows(rows))
	return string(raw)
}
func storyboardSemanticRows(shots []StoryboardShot) []map[string]any {
	rows := []map[string]any{}
	for _, shot := range shots {
		rows = append(rows, map[string]any{"characters": shot.Characters, "scene": shot.Scene, "props": shot.Props, "video_desc": shot.VideoDesc, "duration_sec": shot.Duration})
	}
	return rows
}
