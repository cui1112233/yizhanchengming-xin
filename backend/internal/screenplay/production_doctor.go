package screenplay

import (
	"context"
	"encoding/json"
	"fmt"
	"qiantie/backend/internal/prompt"
	"strings"
)

func (e BatchExecutor) doctor(ctx context.Context, p Project, u Unit, content any) (batchUnitOutcome, error) {
	if e.Generator == nil {
		return batchUnitOutcome{}, fmt.Errorf("%w: 质量检查需要配置文本模型；基础结构检查不能代替质量审核", ErrInvalidState)
	}
	payload := map[string]any{"target": content}
	if e.Store != nil && u.StepKey != "2_screenplay_doctor" {
		if doc, err := e.Store.ActiveDocument(ctx, p.OwnerUsername, p.ID); err == nil {
			payload["screenplay"] = doc.Content
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return batchUnitOutcome{}, err
	}
	response, err := e.Generator.Generate(ctx, TextRequest{Model: p.ModelID, SystemPrompt: prompt.ScreenplayProductionDoctor, UserPrompt: u.StepKey + "\n" + string(raw), JSONResponse: true, MaxOutputTokens: 6000, Temperature: 0.1})
	if err != nil {
		return batchUnitOutcome{}, wrapAIError(err)
	}
	text := strings.TrimSpace(response.Content)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(strings.TrimSpace(text), "```")
	}
	var wire struct {
		Pass    *bool         `json:"pass"`
		Block   *bool         `json:"block_next_step"`
		Summary string        `json:"summary"`
		Issues  []ReviewIssue `json:"issues"`
		Report  string        `json:"report_markdown"`
	}
	if err = json.Unmarshal([]byte(text), &wire); err != nil || wire.Pass == nil || wire.Block == nil || strings.TrimSpace(wire.Summary) == "" {
		return batchUnitOutcome{}, fmt.Errorf("%w: 质检结果缺少明确结论", ErrAIInvalidJSON)
	}
	review := &Review{ID: newScreenplayID("review"), ProjectID: p.ID, OwnerUsername: p.OwnerUsername, StepKey: u.StepKey, TargetKey: u.String(), Mode: "review", Pass: *wire.Pass, BlockNextStep: *wire.Block, Summary: wire.Summary, Issues: wire.Issues, ReportMarkdown: wire.Report, PromptID: "screenplay_production_doctor", CreatedAt: nowUTC(), UpdatedAt: nowUTC()}
	for _, issue := range review.Issues {
		if issue.Level == "error" {
			review.Pass = false
		}
	}
	if !review.Pass {
		review.BlockNextStep = true
	}
	return batchUnitOutcome{review: review}, nil
}
