package novelcreation

import (
	"context"
	"encoding/json"
	"fmt"

	"sort"
	"strings"

	"unicode/utf8"
)

type ImitationDirection struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
type ImitationEmotion struct {
	Stage   string `json:"stage"`
	Emotion string `json:"emotion"`
	Reason  string `json:"reason"`
}
type LongImitationAnalysis struct {
	Dimensions        map[string]string    `json:"dimensions"`
	EmotionCurve      []ImitationEmotion   `json:"emotionCurve"`
	CoreValues        []string             `json:"coreValues"`
	Directions        []ImitationDirection `json:"directions"`
	SelectedDirection string               `json:"selectedDirection"`
}

func validateImitationAnalysis(v LongImitationAnalysis) error {
	raw, _ := json.Marshal(v)
	if jsLength(string(raw)) > 8000 {
		return &imitationOutputLimitError{"文风拆解输出超过8000 UTF16字元上限"}
	}
	if len(v.Dimensions) != 8 || len(v.EmotionCurve) == 0 || len(v.EmotionCurve) > 100 || len(v.CoreValues) == 0 || len(v.CoreValues) > 50 || len(v.Directions) == 0 || len(v.Directions) > 20 {
		return fmt.Errorf("eight dimensions, emotionCurve, coreValues and directions are required")
	}
	for _, d := range ImitationStyleDimensions() {
		if strings.TrimSpace(v.Dimensions[d.Key]) == "" || jsLength(v.Dimensions[d.Key]) > 20000 {
			return fmt.Errorf("dimension %s is required", d.Key)
		}
	}
	for _, v := range v.EmotionCurve {
		if strings.TrimSpace(v.Stage) == "" || strings.TrimSpace(v.Emotion) == "" || strings.TrimSpace(v.Reason) == "" {
			return fmt.Errorf("incomplete emotion curve")
		}
	}
	for _, d := range v.Directions {
		if strings.TrimSpace(d.Title) == "" || strings.TrimSpace(d.Summary) == "" {
			return fmt.Errorf("incomplete direction")
		}
	}
	for _, s := range v.CoreValues {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("empty core value")
		}
	}
	return nil
}
func (p *Pipeline) imitationAnalysis(ctx context.Context, in RunContext, sink EventSink) (*LongImitationAnalysis, error) {
	if in.LongImitation == nil {
		return nil, nil
	}
	if in.ImitationAnalysis != nil {
		return in.ImitationAnalysis, nil
	}
	if len(in.ConfirmedFoundation) > 0 {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "缺少已保存的仿写拆解，请先完成蓝图生成"}
	}
	chunks := in.LongImitation.ReferenceChunks
	if len(chunks) == 0 {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "参考分块缺失"}
	}
	brief := imitationModelBrief(in)
	values := []LongImitationAnalysis{}
	for i, chunk := range chunks {
		var text string
		if e := json.Unmarshal(chunk, &text); e != nil {
			return nil, persistenceError(e)
		}
		if jsLength(text) > 50000 {
			return nil, inputError("参考单块超过50000 UTF16字元")
		}
		if e := p.emitStage(ctx, in, sink, RunPlanning, fmt.Sprintf("正在拆解参考第%d/%d块", i+1, len(chunks)), 5); e != nil {
			return nil, e
		}
		payload := map[string]any{"creation_brief": brief, "reference_chunk": map[string]any{"index": i, "total": len(chunks), "text": text}}
		value, e := p.imitationAnalysisCall(ctx, in, sink, fmt.Sprintf(":chunk:%d", i), fmt.Sprintf("analysis-v2-chunk-%d", i), payload)
		if e != nil {
			return nil, e
		}
		values = append(values, value)
	}
	for level := 0; len(values) > 1; level++ {
		next := []LongImitationAnalysis{}
		for i := 0; i < len(values); i += imitationMergeFanout {
			end := i + imitationMergeFanout
			if end > len(values) {
				end = len(values)
			}
			if end-i == 1 {
				next = append(next, values[i])
				continue
			}
			payload := map[string]any{"creation_brief": brief, "source_analyses": values[i:end], "merge_level": level, "group_index": i / imitationMergeFanout}
			value, e := p.imitationAnalysisCall(ctx, in, sink, fmt.Sprintf(":merge:%d:%d", level, i/imitationMergeFanout), fmt.Sprintf("analysis-v2-merge-%d-%d", level, i/imitationMergeFanout), payload)
			if e != nil {
				return nil, e
			}
			next = append(next, value)
		}
		values = next
	}
	analysis := values[0]
	analysis.SelectedDirection = in.LongImitation.Input.SelectedDirection
	if analysis.SelectedDirection == "" {
		analysis.SelectedDirection = analysis.Directions[0].Title + "：" + analysis.Directions[0].Summary
	}
	if e := persistImitationValue(ctx, in, "analysis", analysis); e != nil {
		return nil, e
	}
	return &analysis, nil
}
func validateImitationFoundation(f foundationPlanOutput, c *LongImitationCreation) error {
	if e := validateFoundation(f); e != nil {
		return e
	}
	if c == nil {
		return nil
	}
	if utf8.RuneCountInString(strings.TrimSpace(f.Title)) > 255 || (c.AutoTitle && strings.TrimSpace(f.Title) == c.ProvisionalTitle) {
		return fmt.Errorf("foundation must supply a real book title within 255 characters")
	}
	volumes := append([]foundationVolume(nil), f.Volumes...)
	sort.Slice(volumes, func(i, j int) bool { return volumes[i].ChapterRange[0] < volumes[j].ChapterRange[0] })
	next := 1
	for _, v := range volumes {
		if v.ChapterRange[0] != next {
			return fmt.Errorf("outline must cover all target chapters without overlap/gaps")
		}
		next = v.ChapterRange[1] + 1
	}
	if next != c.Input.TargetChapters+1 {
		return fmt.Errorf("outline must end at targetChapters=%d", c.Input.TargetChapters)
	}
	return nil
}
func validateImitationConfirmation(snapshot, blueprint json.RawMessage) error {
	var input struct {
		Creation *LongImitationCreation `json:"long_imitation"`
	}
	if e := json.Unmarshal(snapshot, &input); e != nil {
		return persistenceError(e)
	}
	if input.Creation == nil {
		return nil
	}
	var f foundationPlanOutput
	if e := decodeAndValidate(string(blueprint), &f, func() error { return validateImitationFoundation(f, input.Creation) }); e != nil {
		return &DomainError{Code: ErrInputRequired, Message: "蓝图章节范围必须连续覆盖本作品目标章数", Cause: e}
	}
	return nil
}
func validateImitationPlan(v chapterPlanOutput, n int, volumes []foundationVolume, c *LongImitationCreation) error {
	if c == nil {
		return validateChapterPlan(v, n, volumes)
	}
	copy := v
	copy.TargetWords = 2000
	if e := validateChapterPlan(copy, n, volumes); e != nil {
		return e
	}
	if n > c.Input.TargetChapters || v.TargetWords != c.Input.TargetWordCount {
		return fmt.Errorf("plan must respect targetChapters=%d and targetWords=%d", c.Input.TargetChapters, c.Input.TargetWordCount)
	}
	return nil
}
func applyImitationTitle(ctx context.Context, st Store, owner string, run Run, f Foundation, c *LongImitationCreation) error {
	if c == nil || !c.AutoTitle {
		return nil
	}
	var plan foundationPlanOutput
	if e := json.Unmarshal(foundationPlanFromSnapshot(f.Snapshot), &plan); e != nil {
		return persistenceError(e)
	}
	title := strings.TrimSpace(plan.Title)
	if title == "" || title == c.ProvisionalTitle {
		return &DomainError{Code: ErrInvalidOutput, Message: "蓝图未生成真实作品名称"}
	}
	project, e := st.GetProject(ctx, owner, run.ProjectID)
	if e != nil {
		return e
	}
	if project.Title != c.ProvisionalTitle {
		return nil
	}
	if _, e = st.RenameProject(ctx, owner, run.ProjectID, title); e != nil {
		return e
	}
	session, e := st.GetSession(ctx, owner, run.SessionID)
	if e != nil {
		return e
	}
	if session.Title == c.ProvisionalTitle {
		session.Title = title
		_, e = st.UpdateSession(ctx, owner, session)
	}
	return e
}
