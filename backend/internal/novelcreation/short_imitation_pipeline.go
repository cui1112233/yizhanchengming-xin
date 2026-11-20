package novelcreation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"regexp"
	"strings"
)

type ShortImitationScene struct {
	Name          string `json:"name"`
	ExpectedWords int    `json:"expectedWords"`
}
type ShortImitationEvent struct {
	EmotionFunction string `json:"emotionFunction"`
	OriginalEvent   string `json:"originalEvent"`
	NewEvent        string `json:"newEvent"`
}
type ShortImitationBlueprintOutput struct {
	Content       string                `json:"content"`
	Scenes        []ShortImitationScene `json:"scenes"`
	EventMappings []ShortImitationEvent `json:"eventMappings"`
}
type ShortImitationData struct {
	Scenes            []ShortImitationScene    `json:"scenes"`
	EventMappings     []ShortImitationEvent    `json:"eventMappings"`
	Deconstruction    string                   `json:"deconstruction"`
	Directions        string                   `json:"directions"`
	SelectedDirection string                   `json:"selectedDirection"`
	Blueprint         string                   `json:"blueprint"`
	Editorial         *ShortImitationEditorial `json:"editorial,omitempty"`
}
type ShortImitationEditorial struct {
	Content          string   `json:"content"`
	Rating           string   `json:"rating"`
	TitleSuggestions []string `json:"titleSuggestions"`
	Synopsis         string   `json:"synopsis"`
	Issues           []string `json:"issues"`
}

func (v *ShortImitationEditorial) UnmarshalJSON(raw []byte) error {
	type plain ShortImitationEditorial
	var p plain
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&p); e != nil {
		return e
	}
	*v = ShortImitationEditorial(p)
	return nil
}
func validateImitationReport(v shortText, sections ...string) error {
	if e := validShortText(v, 24000); e != nil {
		return e
	}
	for _, s := range sections {
		if !strings.Contains(v.Content, s) {
			return inputError("仿写阶段报告缺少必要区块：" + s)
		}
		if s == "REVIEW START" || s == "REVIEW END" {
			continue
		}
		lines := strings.Split(v.Content, "\n")
		filled := false
		for i, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "#") || !strings.Contains(line, s) {
				continue
			}
			for _, next := range lines[i+1:] {
				if strings.HasPrefix(strings.TrimSpace(next), "#") {
					break
				}
				if strings.TrimSpace(next) != "" && !strings.Contains(next, "REVIEW END") {
					filled = true
					break
				}
			}
			if filled {
				break
			}
		}
		if !filled {
			return inputError("仿写阶段区块为空或缺少标题：" + s)
		}
	}
	return nil
}

var directionMarkers = regexp.MustCompile(`(?m)^[ \t]*===\s*DIRECTION\s+([123])\s*===[ \t]*\r?$`)

func parseImitationDirections(text string) ([]string, error) {
	marks := directionMarkers.FindAllStringSubmatchIndex(text, -1)
	if len(marks) != 3 {
		return nil, inputError("仿写方向必须包含DIRECTION 1、2、3三个独立区块")
	}
	out := []string{}
	for i, m := range marks {
		if text[m[2]:m[3]] != fmt.Sprint(i+1) {
			return nil, inputError("仿写方向编号重复或顺序不正确")
		}
		end := len(text)
		if i < 2 {
			end = marks[i+1][0]
		}
		part := strings.TrimSpace(text[m[0]:end])
		for _, label := range []string{"题材", "设定", "概要", "价值观沿用", "主爽点", "主虐点"} {
			if !strings.Contains(part, label) {
				return nil, inputError("仿写方向缺少：" + label)
			}
		}
		out = append(out, part)
	}
	return out, nil
}
func shortImitationArtifacts(story *ShortStory) {
	if story.Imitation == nil {
		return
	}
	d := story.Imitation
	upsert := func(key, title, content string) {
		if content == "" {
			return
		}
		for i := range story.Artifacts {
			if story.Artifacts[i].Key == key {
				story.Artifacts[i] = ShortArtifact{Key: key, Title: title, Content: content}
				return
			}
		}
		story.Artifacts = append(story.Artifacts, ShortArtifact{Key: key, Title: title, Content: content})
	}
	upsert("imitation/deconstruction.md", "六维拆书报告", d.Deconstruction)
	upsert("imitation/directions.md", "三个仿写方向", d.Directions)
	upsert("imitation/selected-direction.md", "采用的仿写方向", d.SelectedDirection)
	upsert("imitation/blueprint.md", "仿写创作蓝图", d.Blueprint)
	if d.Editorial != nil {
		upsert("reviews/editorial.md", "编辑审核意见单", d.Editorial.Content)
	}
}
func saveShortStage(ctx context.Context, in RunContext, story *ShortStory) error {
	shortImitationArtifacts(story)
	if in.SaveShortStage != nil {
		return in.SaveShortStage(ctx, *story)
	}
	return nil
}

func validateShortImitationBlueprint(v ShortImitationBlueprintOutput, target int) error {
	if e := validateImitationReport(shortText{v.Content}, "情绪曲线", "人物关系", "主线结构", "金句", "角色", "分段规划", "防抄袭"); e != nil {
		return e
	}
	if len(v.Scenes) < 5 || len(v.Scenes) > 8 {
		return inputError("仿写蓝图必须包含5至8个分段计划")
	}
	total := 0
	seen := map[string]bool{}
	for _, scene := range v.Scenes {
		name := strings.TrimSpace(scene.Name)
		if name == "" || jsLength(name) > 128 || seen[name] || !strings.Contains(v.Content, name) || scene.ExpectedWords <= 0 || scene.ExpectedWords > target {
			return inputError("仿写蓝图分段名称或预计字数无效")
		}
		seen[name] = true
		total += scene.ExpectedWords
	}
	if total != target {
		return inputError(fmt.Sprintf("仿写蓝图预计字数合计%d，与目标%d不一致", total, target))
	}
	if len(v.EventMappings) == 0 || len(v.EventMappings) > 50 {
		return inputError("仿写蓝图缺少原作事件与新事件映射，或数量超限")
	}
	for _, event := range v.EventMappings {
		if strings.TrimSpace(event.OriginalEvent) == strings.TrimSpace(event.NewEvent) {
			return inputError("仿写蓝图必须更换事件，不能照搬原事件")
		}
		for _, field := range []string{event.EmotionFunction, event.OriginalEvent, event.NewEvent} {
			if strings.TrimSpace(field) == "" || jsLength(field) > 4000 || !strings.Contains(v.Content, field) {
				return inputError("仿写蓝图情绪功能与事件映射缺失或与报告不一致")
			}
		}
	}
	return nil
}
func validateSavedShortImitation(d *ShortImitationData, target int) error {
	if d == nil {
		return inputError("已确认仿写缺少拆书和蓝图，已停止生成")
	}
	if e := validateImitationReport(shortText{d.Deconstruction}, "核心情绪曲线", "人物关系", "核心价值观", "情绪渴望", "核心逻辑", "金句"); e != nil {
		return e
	}
	if e := validateShortImitationBlueprint(ShortImitationBlueprintOutput{Content: d.Blueprint, Scenes: d.Scenes, EventMappings: d.EventMappings}, target); e != nil {
		return e
	}
	directions, e := parseImitationDirections(d.Directions)
	if e != nil {
		return e
	}
	for _, v := range directions {
		if v == d.SelectedDirection {
			return nil
		}
	}
	return inputError("已保存的仿写方向与本次方向列表不一致")
}
func (p *Pipeline) prepareShortImitation(ctx context.Context, in RunContext, sink EventSink, story *ShortStory) error {
	if story.Imitation == nil {
		story.Imitation = &ShortImitationData{}
	}
	d := story.Imitation
	if d.Deconstruction == "" {
		reports := []string{}
		for i, chunk := range imitationChunks(story.Input.ReferenceText) {
			if e := p.emitStage(ctx, in, sink, RunPlanning, "六维拆书", 5); e != nil {
				return e
			}
			var reference string
			if e := json.Unmarshal(chunk, &reference); e != nil {
				return e
			}
			v := shortText{}
			if e := p.shortCall(ctx, in, sink, prompt.ShortImitationDeconstruction, fmt.Sprintf(":chunk:%d", i), RunPlanning, map[string]any{"reference_text": reference, "creation": shortBrief(*story)}, &v, func() error {
				return validateImitationReport(v, "核心情绪曲线", "人物关系", "核心价值观", "情绪渴望", "核心逻辑", "金句")
			}, 16000); e != nil {
				return e
			}
			reports = append(reports, v.Content)
		}
		if len(reports) == 0 {
			return inputError("仿写短篇需要参考原文")
		}
		for level := 0; len(reports) > 1; level++ {
			next := []string{}
			for start := 0; start < len(reports); start += 4 {
				end := start + 4
				if end > len(reports) {
					end = len(reports)
				}
				v := shortText{}
				if e := p.shortCall(ctx, in, sink, prompt.ShortImitationDeconstruction, fmt.Sprintf(":merge:%d:%d", level, start), RunPlanning, map[string]any{"source_reports": reports[start:end], "instruction": "合并分块拆书，保留六维与整篇情绪链"}, &v, func() error {
					return validateImitationReport(v, "核心情绪曲线", "人物关系", "核心价值观", "情绪渴望", "核心逻辑", "金句")
				}, 16000); e != nil {
					return e
				}
				next = append(next, v.Content)
			}
			reports = next
		}
		d.Deconstruction = reports[0]
		if e := saveShortStage(ctx, in, story); e != nil {
			return e
		}
	}
	if d.Directions == "" {
		if e := p.emitStage(ctx, in, sink, RunPlanning, "生成三个仿写方向", 8); e != nil {
			return e
		}
		v := shortText{}
		if e := p.shortCall(ctx, in, sink, prompt.ShortImitationDirections, "", RunPlanning, map[string]any{"deconstruction": d.Deconstruction, "creation": shortBrief(*story)}, &v, func() error {
			if e := validShortText(v, 24000); e != nil {
				return e
			}
			_, e := parseImitationDirections(v.Content)
			return e
		}, 16000); e != nil {
			return e
		}
		d.Directions = v.Content
		if e := saveShortStage(ctx, in, story); e != nil {
			return e
		}
	}
	directions, e := parseImitationDirections(d.Directions)
	if e != nil {
		return e
	}
	if d.SelectedDirection == "" {
		selected := strings.TrimSpace(story.Input.SelectedDirection)
		if selected == "" {
			selected = directions[0]
		} else {
			found := false
			for _, v := range directions {
				if selected == v {
					found = true
				}
			}
			if !found {
				return inputError("所选仿写方向不在本次生成的三个方向中")
			}
		}
		d.SelectedDirection = selected
	}
	if d.Blueprint == "" {
		if e := p.emitStage(ctx, in, sink, RunPlanning, "生成仿写创作蓝图", 10); e != nil {
			return e
		}
		v := ShortImitationBlueprintOutput{}
		if e := p.shortCall(ctx, in, sink, prompt.ShortImitationBlueprint, "", RunPlanning, map[string]any{"deconstruction": d.Deconstruction, "selected_direction": d.SelectedDirection, "creation": shortBrief(*story)}, &v, func() error {
			return validateShortImitationBlueprint(v, story.Input.TargetWordCount)
		}, 16000); e != nil {
			return e
		}
		d.Blueprint = v.Content
		d.Scenes = v.Scenes
		d.EventMappings = v.EventMappings
		if e := saveShortStage(ctx, in, story); e != nil {
			return e
		}
	}
	return nil
}
func (p *Pipeline) finishShortImitation(ctx context.Context, in RunContext, sink EventSink, story *ShortStory) error {
	if story.Imitation == nil {
		return inputError("仿写前置材料缺失，不能进行编辑审核")
	}
	if e := p.emitStage(ctx, in, sink, RunWriting, "仿写编辑终审", 98); e != nil {
		return e
	}
	v := ShortImitationEditorial{}
	if e := p.shortCall(ctx, in, sink, prompt.ShortImitationEditorial, "", RunWriting, map[string]any{"creation": shortBrief(*story), "chapters": story.Chapters}, &v, func() error {
		if e := validateImitationReport(shortText{v.Content}, "REVIEW START", "REVIEW END", "总评级", "逐维度审核", "标题建议", "一句话简介", "修改清单"); e != nil {
			return e
		}
		if v.Rating != "可过审" && v.Rating != "需修改" && v.Rating != "打回" {
			return inputError("编辑审核评级无效")
		}
		if !strings.Contains(v.Content, v.Rating) || len(v.Issues) > 50 || jsLength(v.Synopsis) > 2000 {
			return inputError("编辑审核报告与评级不符或内容超限")
		}
		for _, issue := range v.Issues {
			if strings.TrimSpace(issue) == "" || jsLength(issue) > 4000 {
				return inputError("编辑审核修改项为空或过长")
			}
		}
		if len(v.TitleSuggestions) < 2 || len(v.TitleSuggestions) > 3 || strings.TrimSpace(v.Synopsis) == "" || v.Issues == nil {
			return inputError("编辑审核须包含2至3个标题、简介及修改清单")
		}
		for _, title := range v.TitleSuggestions {
			if strings.TrimSpace(title) == "" || jsLength(title) > 255 {
				return inputError("编辑审核标题为空或过长")
			}
		}
		if v.Rating == "可过审" && len(v.Issues) > 0 || v.Rating != "可过审" && len(v.Issues) == 0 {
			return inputError("编辑审核评级与修改清单矛盾")
		}
		return nil
	}, 16000); e != nil {
		return e
	}
	story.Imitation.Editorial = &v
	if e := saveShortStage(ctx, in, story); e != nil {
		return e
	}
	if in.ShortImitationVersion < 2 && v.Rating != "可过审" {
		return inputError("仿写编辑终审" + v.Rating + "，正文与意见已保存，请按修改清单处理：" + truncateRunes(strings.Join(v.Issues, "；"), 300))
	}
	return nil
}
