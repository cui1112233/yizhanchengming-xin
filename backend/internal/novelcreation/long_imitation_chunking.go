package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
	"time"
)

const imitationMergeFanout = 4

type ImitationRequirementField struct {
	Field      string `json:"field"`
	Characters int    `json:"characters"`
}
type LongImitationRequirements struct {
	Summary      string                      `json:"summary"`
	Method       string                      `json:"method"`
	SourceFields []ImitationRequirementField `json:"sourceFields"`
	Fingerprint  string                      `json:"fingerprint"`
}
type imitationRequirementSource struct {
	Field string `json:"field"`
	Text  string `json:"text"`
	Part  int    `json:"part"`
	Parts int    `json:"parts"`
}
type imitationCachedArtifact struct {
	Version      int             `json:"version"`
	Binding      string          `json:"binding"`
	Value        json.RawMessage `json:"value"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}
type imitationOutputLimitError struct{ message string }

func (e *imitationOutputLimitError) Error() string { return e.message }
func imitationLimitMessage(message string) string {
	return message + "；已保存超限结果，重试不会再次调用模型。请调整模型或输入后重新创建。"
}
func imitationLimitFailure(cause *imitationOutputLimitError) error {
	return &DomainError{Code: ErrInvalidOutput, Message: imitationLimitMessage(cause.Error()), Cause: cause}
}
func checkImitationRawReply(definitionID, reply string) error {
	limit := 0
	label := ""
	switch definitionID {
	case novelprompt.LongImitationAnalysis.ID:
		limit = 8000
		label = "文风拆解原始回复"
	case novelprompt.LongImitationRequirements.ID:
		limit = 6000
		label = "创作要求原始回复"
	}
	if limit > 0 && jsLength(reply) > limit {
		return imitationLimitFailure(&imitationOutputLimitError{fmt.Sprintf("%s超过%d UTF16字元上限", label, limit)})
	}
	return nil
}

func imitationRequirementSources(c *LongImitationCreation) []imitationRequirementSource {
	in := c.Input
	out := []imitationRequirementSource{}
	for _, s := range []imitationRequirementSource{{Field: "storyIdea", Text: in.StoryIdea}, {Field: "readerPreferences", Text: in.ReaderPreferences}, {Field: "supplementaryInfo", Text: in.SupplementaryInfo}} {
		if s.Text != "" {
			out = append(out, s)
		}
	}
	for _, s := range c.Styles {
		out = append(out, imitationRequirementSource{Field: "style:" + s.Profile.ID, Text: s.Profile.Content})
	}
	return out
}
func needsImitationRequirements(c *LongImitationCreation) bool {
	total := 0
	for _, s := range imitationRequirementSources(c) {
		total += jsLength(s.Text)
	}
	return total > 12000
}
func imitationModelBrief(in RunContext) map[string]any {
	c := in.LongImitation
	i := c.Input
	input := map[string]any{"title": i.Title, "selectedDirection": i.SelectedDirection, "targetWordCount": i.TargetWordCount, "targetChapters": i.TargetChapters, "modelId": i.ModelID, "thinkingLevel": i.ThinkingLevel, "knowledgeBookIds": i.KnowledgeBookIDs, "memoReferences": i.MemoReferences}
	styles := []map[string]string{}
	for _, s := range c.Styles {
		item := map[string]string{"id": s.Profile.ID, "title": s.Profile.Title, "fingerprint": s.Digest}
		if in.ImitationRequirements == nil {
			item["content"] = s.Profile.Content
		}
		styles = append(styles, item)
	}
	if in.ImitationRequirements == nil {
		input["storyIdea"] = i.StoryIdea
		input["readerPreferences"] = i.ReaderPreferences
		input["supplementaryInfo"] = i.SupplementaryInfo
	} else {
		input["largeFieldsPreprocessed"] = true
	}
	memos := []map[string]string{}
	for _, n := range c.Memos {
		memos = append(memos, map[string]string{"projectId": n.ProjectID, "id": n.ID, "title": n.Title, "fingerprint": fingerprint(n)})
	}
	return map[string]any{"input": input, "statistics": c.Statistics, "styleDimensions": c.StyleDimensions, "styles": styles, "memos": memos, "requirements": in.ImitationRequirements, "method": "cm_bounded_long_imitation_v2", "referenceChunkCount": len(c.ReferenceChunks)}
}
func persistImitationValue(ctx context.Context, in RunContext, id string, value any) error {
	j, _ := ctx.Value(journalKey{}).(*callJournal)
	if j == nil {
		return nil
	}
	e := guarded(ctx, j.store, j.lease, func(tx Store, _ *Execution) error {
		return putContent(ctx, tx, in.OwnerUsername, in.Project.ID, longImitationKind, id, "completed", time.Time{}, value)
	})
	if e != nil {
		return persistenceError(e)
	}
	return nil
}
func loadImitationArtifact(ctx context.Context, in RunContext, id string, payload any, target any, validate func() error) (bool, error) {
	j, _ := ctx.Value(journalKey{}).(*callJournal)
	if j == nil {
		return false, nil
	}
	var cached imitationCachedArtifact
	found, e := readContent(ctx, j.store, in.OwnerUsername, in.Project.ID, longImitationKind, id, &cached)
	if e != nil {
		return false, persistenceError(e)
	}
	if !found {
		return false, nil
	}
	if cached.Version != 1 || cached.Binding != fingerprint(payload) {
		return false, conflict()
	}
	if cached.ErrorMessage != "" {
		return false, &DomainError{Code: ErrInvalidOutput, Message: cached.ErrorMessage, Cause: &imitationOutputLimitError{cached.ErrorMessage}}
	}
	if e = json.Unmarshal(cached.Value, target); e != nil {
		return false, persistenceError(e)
	}
	if e = validate(); e != nil {
		return false, persistenceError(e)
	}
	return true, nil
}
func cacheImitationArtifact(ctx context.Context, in RunContext, id string, payload, value any) error {
	raw, e := json.Marshal(value)
	if e != nil {
		return persistenceError(e)
	}
	return persistImitationValue(ctx, in, id, imitationCachedArtifact{Version: 1, Binding: fingerprint(payload), Value: raw})
}
func cacheImitationLimitFailure(ctx context.Context, in RunContext, id string, payload any, cause error) error {
	var limit *imitationOutputLimitError
	if !errors.As(cause, &limit) {
		return nil
	}
	return persistImitationValue(ctx, in, id, imitationCachedArtifact{Version: 1, Binding: fingerprint(payload), ErrorMessage: imitationLimitMessage(limit.Error())})
}
func (p *Pipeline) imitationRequirements(ctx context.Context, in RunContext, sink EventSink) (*LongImitationRequirements, error) {
	if in.LongImitation == nil || !needsImitationRequirements(in.LongImitation) {
		return nil, nil
	}
	if in.ImitationRequirements != nil {
		return in.ImitationRequirements, nil
	}
	if len(in.ConfirmedFoundation) > 0 {
		return nil, &DomainError{Code: ErrPersistenceFailed, Message: "缺少已保存的大文本要求提炼，请先完成创作蓝图"}
	}
	sources := imitationRequirementSources(in.LongImitation)
	fields := []ImitationRequirementField{}
	groups := [][]imitationRequirementSource{}
	group := []imitationRequirementSource{}
	size := 0
	for _, source := range sources {
		fields = append(fields, ImitationRequirementField{source.Field, jsLength(source.Text)})
		chunks := imitationChunks(source.Text)
		for index, raw := range chunks {
			var text string
			_ = json.Unmarshal(raw, &text)
			n := jsLength(text)
			if size+n > 50000 && len(group) > 0 {
				groups = append(groups, group)
				group = nil
				size = 0
			}
			group = append(group, imitationRequirementSource{Field: source.Field, Text: text, Part: index, Parts: len(chunks)})
			size += n
		}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	binding := fingerprint(sources)
	summaries := []string{}
	for i, g := range groups {
		if e := p.emitStage(ctx, in, sink, RunPlanning, fmt.Sprintf("正在提炼创作要求（%d/%d）", i+1, len(groups)), 3); e != nil {
			return nil, e
		}
		value, e := p.imitationRequirementCall(ctx, in, sink, fmt.Sprintf(":source:%d", i), fmt.Sprintf("requirements-v1-source-%d", i), map[string]any{"requirement_sources": g, "source_fingerprint": binding})
		if e != nil {
			return nil, e
		}
		summaries = append(summaries, value)
	}
	for level := 0; len(summaries) > 1; level++ {
		next := []string{}
		for i := 0; i < len(summaries); i += imitationMergeFanout {
			end := i + imitationMergeFanout
			if end > len(summaries) {
				end = len(summaries)
			}
			if end-i == 1 {
				next = append(next, summaries[i])
				continue
			}
			key := fmt.Sprintf(":merge:%d:%d", level, i/imitationMergeFanout)
			value, e := p.imitationRequirementCall(ctx, in, sink, key, fmt.Sprintf("requirements-v1-merge-%d-%d", level, i/imitationMergeFanout), map[string]any{"requirement_summaries": summaries[i:end], "source_fingerprint": binding})
			if e != nil {
				return nil, e
			}
			next = append(next, value)
		}
		summaries = next
	}
	if len(summaries) == 0 {
		return nil, inputError("创作要求为空")
	}
	result := &LongImitationRequirements{Summary: summaries[0], Method: "cm_bounded_requirements_v1", SourceFields: fields, Fingerprint: binding}
	if e := persistImitationValue(ctx, in, "requirements", result); e != nil {
		return nil, e
	}
	return result, nil
}
func (p *Pipeline) imitationRequirementCall(ctx context.Context, in RunContext, sink EventSink, key, id string, payload map[string]any) (string, error) {
	var output struct {
		Summary string `json:"summary"`
	}
	validate := func() error {
		if jsLength(output.Summary) > 4000 {
			return &imitationOutputLimitError{"创作要求摘要超过4000 UTF16字元上限"}
		}
		if strings.TrimSpace(output.Summary) == "" {
			return fmt.Errorf("requirements summary must contain 1..4000 UTF16 units")
		}
		return nil
	}
	found, e := loadImitationArtifact(ctx, in, id, payload, &output, validate)
	if e != nil {
		return "", e
	}
	if !found {
		_, e = p.callJSON(ctx, in, sink, nil, modelCallSpec{definition: novelprompt.LongImitationRequirements, key: key, step: RunPlanning, userPayload: payload, temperature: .2, maxTokens: 6500, validate: validate}, &output)
		if e != nil {
			if cacheError := cacheImitationLimitFailure(ctx, in, id, payload, e); cacheError != nil {
				return "", cacheError
			}
			return "", e
		}
		if e = cacheImitationArtifact(ctx, in, id, payload, output); e != nil {
			return "", e
		}
	}
	return output.Summary, nil
}
func (p *Pipeline) imitationAnalysisCall(ctx context.Context, in RunContext, sink EventSink, key, id string, payload map[string]any) (LongImitationAnalysis, error) {
	var value LongImitationAnalysis
	validate := func() error { return validateImitationAnalysis(value) }
	found, e := loadImitationArtifact(ctx, in, id, payload, &value, validate)
	if e != nil {
		return value, e
	}
	if !found {
		_, e = p.callJSON(ctx, in, sink, nil, modelCallSpec{definition: novelprompt.LongImitationAnalysis, key: key, step: RunPlanning, userPayload: payload, temperature: .35, maxTokens: 10000, validate: validate}, &value)
		if e != nil {
			if cacheError := cacheImitationLimitFailure(ctx, in, id, payload, e); cacheError != nil {
				return value, cacheError
			}
			return value, e
		}
		value.SelectedDirection = ""
		if e = cacheImitationArtifact(ctx, in, id, payload, value); e != nil {
			return value, e
		}
	}
	return value, nil
}
