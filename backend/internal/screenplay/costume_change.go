package screenplay

import (
	"context"
	"fmt"
	"strings"
)

// 换装按洛水 2026-08-18 版规则实现，并在 2026-09-05 版补上分镜范围与「不限制场景」：
// 先圈定分镜范围，再按场景筛选，把范围内绑定人物 A 的分镜批量换成 A-睡衣或其他人物资产。
// 只改命中条件的分镜：不在范围内、场景不符、没绑 A、已经是目标人物的分镜都不会动。
// 替换时保留人物在分镜中的顺序，@图片N 的对应关系不变。
// 洛水改的是 character_ids（人物资产 id），V88 分镜存的是「、」分隔的人物名，因此按词元整体替换。

// CostumeChangeOptions 是换装的条件：分镜范围按 seq 闭区间，Scene 为空表示不限制场景。
type CostumeChangeOptions struct {
	ShotStart int    `json:"shot_start"`
	ShotEnd   int    `json:"shot_end"`
	Scene     string `json:"scene"`
	From      string `json:"from"`
	To        string `json:"to"`
}

// CostumeChangeTarget 是一条命中的分镜，Before/After 便于弹窗预览与核对。
type CostumeChangeTarget struct {
	ShotID string `json:"shot_id"`
	Seq    int    `json:"seq"`
	Scene  string `json:"scene"`
	Before string `json:"before"`
	After  string `json:"after"`
}

// CostumeChangePlan 是换装预览：命中条数与逐条明细，不写库。
type CostumeChangePlan struct {
	ShotStart int                  `json:"shot_start"`
	ShotEnd   int                  `json:"shot_end"`
	Scene     string               `json:"scene"`
	From      string               `json:"from"`
	To        string               `json:"to"`
	Total     int                  `json:"total"`
	Matched   int                  `json:"matched"`
	Targets   []CostumeChangeTarget `json:"targets"`
}

// CostumeChangeOptionsView 是弹窗下拉的数据源：人物名、场景名取自资产库，再补上本集分镜里实际出现的名字。
type CostumeChangeOptionsView struct {
	Characters []string `json:"characters"`
	Scenes     []string `json:"scenes"`
	ShotCount  int      `json:"shot_count"`
}

// PlanCostumeChange 只计算命中分镜，不改动入参；预览与执行共用同一份算法。
func PlanCostumeChange(shots []StoryboardShot, options CostumeChangeOptions) (CostumeChangePlan, error) {
	normalized, err := normalizeCostumeChangeOptions(len(shots), options)
	if err != nil {
		return CostumeChangePlan{}, err
	}
	plan := CostumeChangePlan{
		ShotStart: normalized.ShotStart,
		ShotEnd:   normalized.ShotEnd,
		Scene:     normalized.Scene,
		From:      normalized.From,
		To:        normalized.To,
		Total:     len(shots),
		Targets:   []CostumeChangeTarget{},
	}
	for _, shot := range shots {
		if shot.Seq < normalized.ShotStart || shot.Seq > normalized.ShotEnd {
			continue
		}
		if normalized.Scene != "" && strings.TrimSpace(shot.Scene) != normalized.Scene {
			continue
		}
		after, ok := ReplaceCostumeName(shot.Characters, normalized.From, normalized.To)
		if !ok {
			continue
		}
		plan.Targets = append(plan.Targets, CostumeChangeTarget{
			ShotID: shot.ID,
			Seq:    shot.Seq,
			Scene:  shot.Scene,
			Before: shot.Characters,
			After:  after,
		})
	}
	plan.Matched = len(plan.Targets)
	return plan, nil
}

// ApplyCostumeChange 按规则批量替换人物名，整集一次性保存，返回替换后的分镜。
func ApplyCostumeChange(ctx context.Context, store Store, owner, projectID string, episodeNo int, engine StoryboardEngine, options CostumeChangeOptions) (Storyboard, CostumeChangePlan, error) {
	storyboard, err := loadEditableStoryboard(ctx, store, owner, projectID, episodeNo, engine)
	if err != nil {
		return Storyboard{}, CostumeChangePlan{}, err
	}
	plan, err := PlanCostumeChange(storyboard.Shots, options)
	if err != nil {
		return storyboard, plan, err
	}
	if plan.Matched == 0 {
		return storyboard, plan, fmt.Errorf("%w: %s", ErrInvalidState, CostumeChangeEmptyMessage(plan))
	}
	replacements := make(map[string]string, plan.Matched)
	for _, target := range plan.Targets {
		replacements[target.ShotID] = target.After
	}
	for index := range storyboard.Shots {
		after, hit := replacements[storyboard.Shots[index].ID]
		if !hit {
			continue
		}
		storyboard.Shots[index].Characters = after
	}
	saved, err := saveEditedStoryboard(ctx, store, storyboard)
	if err != nil {
		return saved, plan, err
	}
	return saved, plan, nil
}

// ReplaceCostumeName 把人物名按词元整体替换，保留原有顺序与其余人物。
// 已经是目标人物、或没有原人物的分镜返回 false，与洛水「不会动」的判定一致。
func ReplaceCostumeName(characters, from, to string) (string, bool) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" || from == to {
		return characters, false
	}
	tokens := splitCostumeNames(characters)
	if len(tokens) == 0 {
		return characters, false
	}
	hit := false
	for _, token := range tokens {
		if token == to {
			// 已经是目标人物，整条分镜不动，避免把 A、A-睡衣 混在一起。
			return characters, false
		}
	}
	replaced := make([]string, 0, len(tokens))
	seen := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if token == from {
			token = to
			hit = true
		}
		if seen[token] {
			continue
		}
		seen[token] = true
		replaced = append(replaced, token)
	}
	if !hit {
		return characters, false
	}
	return strings.Join(replaced, "、"), true
}

// BuildCostumeChangeOptionsView 汇总人物与场景下拉项：资产库在前，分镜里出现的名字补在后。
func BuildCostumeChangeOptionsView(shots []StoryboardShot, assets []Asset) CostumeChangeOptionsView {
	view := CostumeChangeOptionsView{Characters: []string{}, Scenes: []string{}, ShotCount: len(shots)}
	characters := make(map[string]bool, len(assets))
	scenes := make(map[string]bool, len(assets))
	for _, asset := range assets {
		name := strings.TrimSpace(asset.Name)
		if name == "" {
			continue
		}
		switch asset.Type {
		case AssetCharacter:
			if !characters[name] {
				characters[name] = true
				view.Characters = append(view.Characters, name)
			}
		case AssetScene:
			if !scenes[name] {
				scenes[name] = true
				view.Scenes = append(view.Scenes, name)
			}
		}
	}
	for _, shot := range shots {
		for _, name := range splitCostumeNames(shot.Characters) {
			if characters[name] {
				continue
			}
			characters[name] = true
			view.Characters = append(view.Characters, name)
		}
		scene := strings.TrimSpace(shot.Scene)
		if scene == "" || scenes[scene] {
			continue
		}
		scenes[scene] = true
		view.Scenes = append(view.Scenes, scene)
	}
	return view
}

// CostumeChangeEmptyMessage 复述洛水弹窗里的「没有可换装的分镜」提示，区分限定场景与不限制场景。
func CostumeChangeEmptyMessage(plan CostumeChangePlan) string {
	scope := fmt.Sprintf("分镜 %d-%d", plan.ShotStart, plan.ShotEnd)
	if plan.Scene == "" {
		return fmt.Sprintf("没有可换装的分镜：%s 中需绑定「%s」，并且还不是「%s」", scope, plan.From, plan.To)
	}
	return fmt.Sprintf("没有可换装的分镜：%s 中需绑定「%s」，且已绑定「%s」，并且还不是「%s」", scope, plan.Scene, plan.From, plan.To)
}

// CostumeChangeSuccessMessage 复述洛水完成提示，便于前端原样展示。
func CostumeChangeSuccessMessage(plan CostumeChangePlan) string {
	scope := fmt.Sprintf("分镜 %d-%d", plan.ShotStart, plan.ShotEnd)
	if plan.Scene != "" {
		scope = fmt.Sprintf("%s（%s）", scope, plan.Scene)
	}
	return fmt.Sprintf("已在%s将「%s」换成「%s」，共 %d 个分镜", scope, plan.From, plan.To, plan.Matched)
}

func normalizeCostumeChangeOptions(shotCount int, options CostumeChangeOptions) (CostumeChangeOptions, error) {
	options.From = strings.TrimSpace(options.From)
	options.To = strings.TrimSpace(options.To)
	options.Scene = strings.TrimSpace(options.Scene)
	if options.From == "" || options.To == "" {
		return options, fmt.Errorf("%w: 请选择原人物和替换人物", ErrInvalidArgument)
	}
	if options.From == options.To {
		return options, fmt.Errorf("%w: 原人物和替换人物不能相同", ErrInvalidArgument)
	}
	if options.ShotStart <= 0 {
		options.ShotStart = 1
	}
	if shotCount == 0 {
		return options, fmt.Errorf("%w: 本集还没有分镜，暂无可换装内容", ErrInvalidState)
	}
	if options.ShotStart > shotCount {
		return options, fmt.Errorf("%w: 起始分镜超出本集 %d 个分镜", ErrInvalidArgument, shotCount)
	}
	if options.ShotEnd <= 0 || options.ShotEnd > shotCount {
		options.ShotEnd = shotCount
	}
	if options.ShotEnd < options.ShotStart {
		return options, fmt.Errorf("%w: 结束分镜不能小于起始分镜", ErrInvalidArgument)
	}
	return options, nil
}

func splitCostumeNames(value string) []string {
	names := make([]string, 0, 4)
	for _, part := range strings.FieldsFunc(value, storyboardNameSeparator) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		names = append(names, part)
	}
	return names
}
