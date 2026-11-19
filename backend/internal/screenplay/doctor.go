package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"qiantie/backend/internal/prompt/luoshui"
)

// 剧本医生 / 分镜医生（洛水 luoshui_doctors_v1 + luoshui_doctor_shared 取证，2026-06~09 更新日志）：
// 医生以 review 模式审核产出，机器可读 JSON 规范：
//   {doctor, mode:"review", pass, block_next_step, reviewed_at, scope, issues[], suggested_actions[], summary}
//   issue: {id, severity: blocker|major|minor, episode, location, dimension, problem, suggestion, evidence}
// pass 判定：无 blocker → pass=true；任一 blocker → pass=false 且 block_next_step=true（major/minor 不停批次）。
// 剧本医生检查维度（更新日志原文）：剧情逻辑、人物动机、资产引用、节奏推进、镜头缺失、台词违和。
// 分镜词医生检查维度：分镜提示词与剧本一致性——人物错位、场景不连贯、资产引用错误、video_desc 结构缺失。
// V88 差异：医生是独立审核入口（不是 Agent 内置 Skill），rewrite 写回走「选用此版 / 分镜编辑」既有链路，
// 本文件只做 review（洛水 review 同样不改稿，只出报告与建议动作）。

type DoctorSeverity string

const (
	DoctorSeverityBlocker DoctorSeverity = "blocker"
	DoctorSeverityMajor   DoctorSeverity = "major"
	DoctorSeverityMinor   DoctorSeverity = "minor"
)

// DoctorIssue 一条审核问题；对齐洛水 issue 对象。
type DoctorIssue struct {
	ID         string         `json:"id"`
	Severity   DoctorSeverity `json:"severity"`
	Episode    int            `json:"episode,omitempty"`
	Location   string         `json:"location,omitempty"`
	Dimension  string         `json:"dimension,omitempty"`
	Problem    string         `json:"problem"`
	Suggestion string         `json:"suggestion,omitempty"`
	Evidence   string         `json:"evidence,omitempty"`
}

// DoctorReview 医生审核报告；对齐洛水 review JSON 通用字段。
type DoctorReview struct {
	Doctor           string         `json:"doctor"` // screenplay | storyboard
	Mode             string         `json:"mode"`   // 固定 review
	Pass             bool           `json:"pass"`
	BlockNextStep    bool           `json:"block_next_step"`
	ReviewedAt       string         `json:"reviewed_at"`
	Scope            map[string]any `json:"scope"`
	Issues           []DoctorIssue  `json:"issues"`
	SuggestedActions []string       `json:"suggested_actions"`
	Summary          string         `json:"summary"`
}

type doctorReviewDraft struct {
	Issues []struct {
		ID         string `json:"id"`
		Severity   string `json:"severity"`
		Episode    int    `json:"episode"`
		Location   string `json:"location"`
		Dimension  string `json:"dimension"`
		Problem    string `json:"problem"`
		Suggestion string `json:"suggestion"`
		Evidence   string `json:"evidence"`
	} `json:"issues"`
	SuggestedActions []string `json:"suggested_actions"`
	Summary          string   `json:"summary"`
}

// ScreenplayDoctorSystemPrompt 剧本医生系统提示词：只审稿不改稿，输出严格 JSON。
// 内置词注册在 luoshui 语料目录（v88_doctor_screenplay），管理后台发布覆盖版本后
// 运行时自动改用覆盖正文。
func ScreenplayDoctorSystemPrompt() string {
	return luoshui.V88RuntimeBody(luoshui.V88DoctorScreenplayID)
}

// StoryboardDoctorSystemPrompt 分镜词医生系统提示词：只审核一致性，不改分镜。
// 内置词注册在 luoshui 语料目录（v88_doctor_storyboard），后台覆盖后运行时生效。
func StoryboardDoctorSystemPrompt() string {
	return luoshui.V88RuntimeBody(luoshui.V88DoctorStoryboardID)
}

// normalizeDoctorReview 解析模型输出并按洛水规则计算 pass：severity 由服务端裁决，不信任模型自报。
func normalizeDoctorReview(raw, doctor string, scope map[string]any, now time.Time) (DoctorReview, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return DoctorReview{}, fmt.Errorf("%w: 模型输出没有可解析的医生审核 JSON", ErrInvalidState)
	}
	var draft doctorReviewDraft
	if err := json.Unmarshal([]byte(raw[start:end+1]), &draft); err != nil {
		return DoctorReview{}, fmt.Errorf("%w: 医生审核 JSON 解析失败：%v", ErrInvalidState, err)
	}
	issues := make([]DoctorIssue, 0, len(draft.Issues))
	hasBlocker := false
	seen := map[string]bool{}
	for index, item := range draft.Issues {
		problem := strings.TrimSpace(item.Problem)
		if problem == "" {
			continue
		}
		severity := DoctorSeverity(strings.ToLower(strings.TrimSpace(item.Severity)))
		switch severity {
		case DoctorSeverityBlocker:
			hasBlocker = true
		case DoctorSeverityMajor, DoctorSeverityMinor:
		default:
			severity = DoctorSeverityMinor
		}
		id := strings.TrimSpace(item.ID)
		if id == "" {
			id = fmt.Sprintf("issue_%d", index+1)
		}
		for seen[id] {
			id = id + "_dup"
		}
		seen[id] = true
		issues = append(issues, DoctorIssue{
			ID: id, Severity: severity, Episode: item.Episode, Location: strings.TrimSpace(item.Location),
			Dimension: strings.TrimSpace(item.Dimension), Problem: problem,
			Suggestion: strings.TrimSpace(item.Suggestion), Evidence: strings.TrimSpace(item.Evidence),
		})
	}
	actions := make([]string, 0, len(draft.SuggestedActions))
	for _, action := range draft.SuggestedActions {
		if trimmed := strings.TrimSpace(action); trimmed != "" {
			actions = append(actions, trimmed)
		}
	}
	return DoctorReview{
		Doctor: doctor, Mode: "review",
		Pass: !hasBlocker, BlockNextStep: hasBlocker,
		ReviewedAt: now.UTC().Format("2006-01-02T15:04:05Z"),
		Scope:      scope, Issues: issues, SuggestedActions: actions,
		Summary: strings.TrimSpace(draft.Summary),
	}, nil
}

// DoctorExecutor 医生审核回调，由 HTTP 层用已解析的文本模型实现。
type DoctorExecutor func(ctx context.Context, systemPrompt, userPrompt string) (string, error)

// ReviewScreenplayDoctor 剧本医生：审核当前活动剧本全文，产出 review 报告。
func ReviewScreenplayDoctor(ctx context.Context, store Store, owner, projectID string, now time.Time, review DoctorExecutor) (DoctorReview, error) {
	if review == nil {
		return DoctorReview{}, fmt.Errorf("%w: 文本模型不可用", ErrInvalidState)
	}
	document, err := store.ActiveDocument(ctx, owner, projectID)
	if err != nil {
		return DoctorReview{}, err
	}
	if strings.TrimSpace(document.Content) == "" {
		return DoctorReview{}, fmt.Errorf("%w: 还没有剧本内容，先 AI 生成或导入后再来", ErrInvalidState)
	}
	raw, err := review(ctx, ScreenplayDoctorSystemPrompt(), "请审核以下剧本全文：\n\n"+document.Content)
	if err != nil {
		return DoctorReview{}, err
	}
	scope := map[string]any{"episodes": document.EpisodeCount, "full_script": true}
	return normalizeDoctorReview(raw, "screenplay", scope, now)
}

// storyboardDoctorShotLine 一条分镜的紧凑描述，供模型对照剧本。
func storyboardDoctorShotLine(shot StoryboardShot) string {
	parts := []string{fmt.Sprintf("分镜%d", shot.Seq)}
	if strings.TrimSpace(shot.Characters) != "" {
		parts = append(parts, "人物："+shot.Characters)
	}
	if strings.TrimSpace(shot.Scene) != "" {
		parts = append(parts, "场景："+shot.Scene)
	}
	if strings.TrimSpace(shot.Props) != "" {
		parts = append(parts, "道具："+shot.Props)
	}
	if strings.TrimSpace(shot.Line) != "" {
		parts = append(parts, "台词："+shot.Line)
	}
	parts = append(parts, "画面提示词："+strings.TrimSpace(shot.ImagePrompt))
	parts = append(parts, "视频提示词："+strings.TrimSpace(shot.VideoDesc))
	return strings.Join(parts, "｜")
}

// ReviewStoryboardDoctor 分镜医生：对照剧本审核指定集分镜，产出 review 报告。
func ReviewStoryboardDoctor(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, now time.Time, review DoctorExecutor) (DoctorReview, error) {
	if review == nil {
		return DoctorReview{}, fmt.Errorf("%w: 文本模型不可用", ErrInvalidState)
	}
	storyboard, err := store.GetStoryboard(ctx, owner, projectID, episodeNo, engine)
	if err != nil {
		return DoctorReview{}, err
	}
	if len(storyboard.Shots) == 0 {
		return DoctorReview{}, fmt.Errorf("%w: 该集还没有分镜，先生成分镜后再来", ErrInvalidState)
	}
	script := ""
	if document, docErr := store.ActiveDocument(ctx, owner, projectID); docErr == nil {
		script = document.Content
	}
	lines := make([]string, 0, len(storyboard.Shots))
	for _, shot := range storyboard.Shots {
		lines = append(lines, storyboardDoctorShotLine(shot))
	}
	userPrompt := "剧本全文：\n" + script + "\n\n待审核分镜（第 " + fmt.Sprint(episodeNo) + " 集，引擎 " + string(engine) + "）：\n" + strings.Join(lines, "\n")
	raw, err := review(ctx, StoryboardDoctorSystemPrompt(), userPrompt)
	if err != nil {
		return DoctorReview{}, err
	}
	scope := map[string]any{"episode": episodeNo, "engine": string(engine), "shots": len(storyboard.Shots)}
	return normalizeDoctorReview(raw, "storyboard", scope, now)
}
