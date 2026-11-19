package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// 分镜检查（洛水 storyboard_check 技能，bundle 反编译取证）：
//   - 洛水描述：「检查原文分镜与AI分镜的缺失/重复/越界，自动补齐缺失分镜并生成检查报告。输出 storyboard_check_report.json」
//   - 洛水运行时 Bt()：从原文按 /^(\d+)[.、．]\s*/gm 提取序号行，与 AI 结果的 scene_id 对照，
//     得到 missing（原文有 AI 没有）、extra（AI 有原文没有，即越界）、duplicates（scene_id 重复），
//     报告结构 {total, originalTotal, errors, missing, extra, duplicates, minNum, maxNum, passed}。
// V88 等价实现：分镜由规划链生成、seq 即镜头序号，没有「原文序号行」这个中间产物，
// 因此检查基准改为「序号 1..最大序号 连续完整」：
//   - missing  = 1..maxSeq 中缺席的序号（跳号）
//   - duplicates = 同一 seq 出现多次
//   - invalid  = seq <= 0 的异常序号（洛水「异常ID」）
//   - extra    = 洛水按原文序号行对照的「多余（越界）」，V88 无对照基准恒为空（保留字段对齐报告结构）
// 修复与洛水一致：缺失分镜逐条调用 AI 补全并插入，完成后可下载 storyboard_check_report.json。

// StoryboardCheckReport 对齐洛水 storyboard_check_report.json 的结构（字段名 snake_case）。
type StoryboardCheckReport struct {
	Total         int   `json:"total"`          // 分镜条数（洛水 total = scene_id 数量）
	ExpectedTotal int   `json:"expected_total"` // 1..maxSeq 应有条数（洛水 originalTotal = 原文序号行数）
	MinSeq        int   `json:"min_seq"`        // 洛水 minNum
	MaxSeq        int   `json:"max_seq"`        // 洛水 maxNum
	Missing       []int `json:"missing"`        // 缺失：序号范围内缺席的编号
	Extra         []int `json:"extra"`          // 洛水「多余（越界）」；V88 无原文序号行对照，恒为空
	Duplicates    []int `json:"duplicates"`     // 重复：同一编号出现多次
	Invalid       []int `json:"invalid"`        // 异常ID：seq <= 0
	Passed        bool  `json:"passed"`         // 洛水 passed = 无缺失/多余/重复
}

// CheckStoryboardShots 对单集分镜做序号完整性检查，只读不写库。
// 空分镜列表返回 ErrInvalidState，与洛水「AI结果为空，请先生成内容」语义一致。
func CheckStoryboardShots(shots []StoryboardShot) (StoryboardCheckReport, error) {
	if len(shots) == 0 {
		return StoryboardCheckReport{}, fmt.Errorf("%w: 本集还没有分镜，请先生成内容", ErrInvalidState)
	}
	report := StoryboardCheckReport{
		Total:     len(shots),
		Missing:   []int{},
		Extra:     []int{},
		Duplicates: []int{},
		Invalid:   []int{},
	}
	counts := make(map[int]int, len(shots))
	for _, shot := range shots {
		counts[shot.Seq]++
		if shot.Seq <= 0 {
			report.Invalid = append(report.Invalid, shot.Seq)
		} else {
			if report.MinSeq == 0 || shot.Seq < report.MinSeq {
				report.MinSeq = shot.Seq
			}
			if shot.Seq > report.MaxSeq {
				report.MaxSeq = shot.Seq
			}
		}
	}
	sort.Ints(report.Invalid)
	for seq := 1; seq <= report.MaxSeq; seq++ {
		if counts[seq] == 0 {
			report.Missing = append(report.Missing, seq)
		}
	}
	for seq, count := range counts {
		if count > 1 {
			report.Duplicates = append(report.Duplicates, seq)
		}
	}
	sort.Ints(report.Duplicates)
	report.ExpectedTotal = report.MaxSeq
	report.Passed = len(report.Missing) == 0 && len(report.Duplicates) == 0 && len(report.Invalid) == 0
	return report, nil
}

// StoryboardCheckFillPrompt 生成补全用户提示词：给出前后分镜上下文与目标编号，要求只输出 JSON。
func StoryboardCheckFillPrompt(seq int, prev, next *StoryboardShot) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("分镜序号 %d 缺失，请补全这条分镜。", seq))
	appendNeighbor := func(role string, shot *StoryboardShot) {
		if shot == nil {
			return
		}
		builder.WriteString(fmt.Sprintf("\n%s（序号 %d）：", role, shot.Seq))
		if strings.TrimSpace(shot.VideoDesc) != "" {
			builder.WriteString("\n视频提示词：" + shot.VideoDesc)
		}
		if strings.TrimSpace(shot.ImagePrompt) != "" {
			builder.WriteString("\n画面提示词：" + shot.ImagePrompt)
		}
		if strings.TrimSpace(shot.Characters) != "" {
			builder.WriteString("\n人物：" + shot.Characters)
		}
		if strings.TrimSpace(shot.Scene) != "" {
			builder.WriteString("\n场景：" + shot.Scene)
		}
		if strings.TrimSpace(shot.Line) != "" {
			builder.WriteString("\n台词：" + shot.Line)
		}
	}
	appendNeighbor("前一分镜", prev)
	appendNeighbor("后一分镜", next)
	builder.WriteString("\n要求与前后分镜衔接自然、风格一致。只输出一个 JSON 对象，字段：" +
		"characters（人物名，多个用 、 分隔）、scene（场景）、props（道具）、video_desc（视频提示词）、" +
		"image_prompt（画面提示词）、line（台词，可为空字符串）、duration_sec（推荐时长秒数，数字）。" +
		"不要输出解释或 markdown 代码块。")
	return builder.String()
}

// StoryboardCheckFillSystemPrompt 补全用的系统提示词，与洛水「缺失分镜自动调用AI补全」对应。
func StoryboardCheckFillSystemPrompt() string {
	return "你是分镜脚本补全器。根据给定的前后分镜上下文，为缺失编号生成一条衔接自然的新分镜。" +
		"只输出一个 JSON 对象，不要任何解释、前后缀或代码块标记。"
}

// storyboardFillDraft 是模型返回的补全草稿，字段名与洛水分镜 JSON 对齐。
type storyboardFillDraft struct {
	Characters  string          `json:"characters"`
	Scene       string          `json:"scene"`
	Props       string          `json:"props"`
	VideoDesc   string          `json:"video_desc"`
	ImagePrompt string          `json:"image_prompt"`
	Line        string          `json:"line"`
	Duration    json.Number     `json:"duration_sec"`
	Extra       json.RawMessage `json:"-"`
}

// parseStoryboardFillJSON 从模型输出里截取第一个 { 到最后一个 } 解析补全草稿，容忍代码块包裹。
func parseStoryboardFillJSON(text string) (storyboardFillDraft, error) {
	var draft storyboardFillDraft
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return draft, fmt.Errorf("%w: 模型输出没有可解析的分镜 JSON", ErrInvalidState)
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &draft); err != nil {
		return draft, fmt.Errorf("%w: 分镜 JSON 解析失败：%v", ErrInvalidState, err)
	}
	return draft, nil
}

// fillStoryboardGenerate 单条补全回调，由 HTTP 层用已解析的文本模型实现。
type fillStoryboardGenerate func(ctx context.Context, prompt string) (string, error)

// FillMissingStoryboardShot 补全一条缺失分镜并整集保存：
// 只允许补当前报告中确实缺失的序号；新分镜按 seq 插入排序位置，时长取前一分镜推荐时长，缺省 5 秒。
func FillMissingStoryboardShot(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, seq int, generate fillStoryboardGenerate) (Storyboard, StoryboardShot, error) {
	if generate == nil {
		return Storyboard{}, StoryboardShot{}, fmt.Errorf("%w: 文本模型不可用", ErrInvalidState)
	}
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, StoryboardShot{}, err
	}
	report, err := CheckStoryboardShots(storyboard.Shots)
	if err != nil {
		return Storyboard{}, StoryboardShot{}, err
	}
	missing := false
	for _, value := range report.Missing {
		if value == seq {
			missing = true
			break
		}
	}
	if !missing {
		return Storyboard{}, StoryboardShot{}, fmt.Errorf("%w: 序号 %d 不是缺失分镜，不能补全", ErrInvalidState, seq)
	}
	var prev, next *StoryboardShot
	for index := range storyboard.Shots {
		if storyboard.Shots[index].Seq < seq {
			if prev == nil || storyboard.Shots[index].Seq > prev.Seq {
				prev = &storyboard.Shots[index]
			}
		}
		if storyboard.Shots[index].Seq > seq {
			if next == nil || storyboard.Shots[index].Seq < next.Seq {
				next = &storyboard.Shots[index]
			}
		}
	}
	prompt := StoryboardCheckFillPrompt(seq, prev, next)
	raw, err := generate(ctx, prompt)
	if err != nil {
		return storyboard, StoryboardShot{}, err
	}
	draft, err := parseStoryboardFillJSON(raw)
	if err != nil {
		return storyboard, StoryboardShot{}, err
	}
	duration := 5.0
	if prev != nil && prev.Duration > 0 {
		duration = prev.Duration
	}
	if draft.Duration != "" {
		if value, err := draft.Duration.Float64(); err == nil && value > 0 {
			duration = value
		}
	}
	shot := StoryboardShot{
		ID:          newScreenplayID("sb_shot"),
		StoryboardID: storyboard.ID,
		ProjectID:   storyboard.ProjectID,
		OwnerUsername: storyboard.OwnerUsername,
		Seq:         seq,
		Duration:    duration,
		Characters:  strings.TrimSpace(draft.Characters),
		Scene:       strings.TrimSpace(draft.Scene),
		Props:       strings.TrimSpace(draft.Props),
		VideoDesc:   strings.TrimSpace(draft.VideoDesc),
		ImagePrompt: strings.TrimSpace(draft.ImagePrompt),
		Line:        strings.TrimSpace(draft.Line),
	}
	insertAt := len(storyboard.Shots)
	for index := range storyboard.Shots {
		if storyboard.Shots[index].Seq > seq {
			insertAt = index
			break
		}
	}
	storyboard.Shots = append(storyboard.Shots, StoryboardShot{})
	copy(storyboard.Shots[insertAt+1:], storyboard.Shots[insertAt:])
	storyboard.Shots[insertAt] = shot
	saved, err := saveEditedStoryboard(ctx, store, storyboard)
	if err != nil {
		return Storyboard{}, StoryboardShot{}, err
	}
	return saved, shot, nil
}
