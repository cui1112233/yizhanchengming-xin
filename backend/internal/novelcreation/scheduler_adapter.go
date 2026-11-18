package novelcreation

import (
	"context"
	"encoding/json"
	prompt "qiantie/backend/internal/prompt/novelcreation"
	"strings"
)

type SchedulerStartInput struct {
	IdempotencyKey         string `json:"idempotencyKey"`
	SessionID              string `json:"sessionId"`
	ModelID                string `json:"modelId"`
	IntervalSeconds        int    `json:"intervalSeconds"`
	DailyChapterCap        int    `json:"dailyChapterCap"`
	MaxConsecutiveFailures int    `json:"maxConsecutiveFailures"`
}

// StartScheduler adapts the observed scheduler controls to durable CM schedules.
// Receipt and plan are committed together; start never dispatches a model call.
func (s *Service) StartScheduler(ctx context.Context, owner, project string, in SchedulerStartInput) (out Schedule, err error) {
	owner, err = requiredOwner(owner)
	if err != nil {
		return
	}
	project = strings.TrimSpace(project)
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	in.SessionID = strings.TrimSpace(in.SessionID)
	in.ModelID = strings.TrimSpace(in.ModelID)
	if err = studioKey(in.IdempotencyKey); err != nil {
		return
	}
	if in.SessionID == "" {
		return out, inputError("请先选择要续写的长篇作品及会话")
	}
	if in.IntervalSeconds == 0 {
		in.IntervalSeconds = 1800
	}
	if in.DailyChapterCap == 0 {
		in.DailyChapterCap = 10
	}
	if in.MaxConsecutiveFailures == 0 {
		in.MaxConsecutiveFailures = 3
	}
	if in.IntervalSeconds < 60 || in.IntervalSeconds > 86400 || in.IntervalSeconds%60 != 0 || in.DailyChapterCap < 1 || in.DailyChapterCap > 100 || in.MaxConsecutiveFailures < 1 || in.MaxConsecutiveFailures > 10 {
		return out, inputError("写作间隔需为1到1440整分钟，每日上限1到100，连续失败阈值1到10")
	}
	digest := fingerprint([]any{project, in})
	err = s.store.Atomic(ctx, owner, project, func(tx Store) error {
		receipt, e := tx.ReserveSubmission(ctx, owner, "scheduler_start", in.IdempotencyKey, digest)
		if e != nil {
			return e
		}
		if receipt.Fingerprint != digest {
			return &DomainError{Code: ErrIdempotencyConflict, Message: "启动编号已用于不同参数，请先核对原调度器"}
		}
		if len(receipt.Result) > 0 {
			var prior Schedule
			if e = json.Unmarshal(receipt.Result, &prior); e != nil {
				return e
			}
			out, e = tx.GetSchedule(ctx, owner, prior.ID)
			return e
		}
		p, e := tx.GetProject(ctx, owner, project)
		if e != nil {
			return e
		}
		if projectReadOnly(p.Status) {
			return archivedError()
		}
		priorPlans, e := tx.ListSchedules(ctx, owner, project)
		if e != nil {
			return e
		}
		for _, prior := range priorPlans {
			if prior.Status != "stopped" && prior.Status != "completed" {
				return inputError("此作品已有未停止的调度计划，请先恢复或停止原计划，避免重复生产")
			}
		}
		session, e := writableSession(ctx, tx, owner, project, in.SessionID)
		if e != nil {
			return e
		}
		if session.Mode == "short_original" || session.Mode == "short_imitation" {
			return inputError("短篇使用独立成品流程，不能通过逐章调度器运行；请选择长篇作品会话")
		}
		child := *s
		child.store = tx
		selected := AutomationInput{}
		// Reuse the last saved session selections rather than asking for another
		// independent set of model/prompt/reference fields on the scheduler page.
		runs, e := tx.ListRuns(ctx, owner, project)
		if e != nil {
			return e
		}
		for _, run := range runs {
			if run.SessionID == session.ID {
				var snap struct {
					ModeInputs     map[string]string `json:"mode_inputs"`
					KnowledgeIDs   []string          `json:"knowledge_ids"`
					NoteIDs        []string          `json:"note_ids"`
					StyleProfileID string            `json:"style_profile_id"`
				}
				if e = json.Unmarshal(run.PromptSnapshot, &snap); e != nil {
					return inputError("原会话设置无法读取，请先在对话中确认创作设置")
				}
				selected = AutomationInput{ModeInputs: snap.ModeInputs, KnowledgeIDs: snap.KnowledgeIDs, NoteIDs: snap.NoteIDs, StyleProfileID: snap.StyleProfileID}
				break
			}
		}
		if session.Mode == "continuation" && strings.TrimSpace(selected.ModeInputs["previousText"]) == "" {
			chapters, e := tx.ListChapters(ctx, owner, project)
			if e != nil {
				return e
			}
			if len(chapters) == 0 {
				return inputError("续写调度前请先导入已完成章节")
			}
			last := chapters[len(chapters)-1]
			selected.ModeInputs = map[string]string{"previousText": last.Content}
		}
		if session.Mode == "fanfiction" && strings.TrimSpace(selected.ModeInputs["existingSetting"]) == "" {
			selected.ModeInputs = map[string]string{"existingSetting": strings.Join([]string{p.World, p.Protagonist, p.CoreConflict}, "\n")}
		}
		settings, e := tx.GetProjectSettings(ctx, owner, project)
		if e != nil {
			return e
		}
		model := in.ModelID
		if model == "" {
			model = p.ModelID
		}
		if model == "" {
			return inputError("请在对话中选择当前可用文本模型")
		}
		out, e = child.createSchedule(ctx, owner, project, ScheduleInput{IdempotencyKey: "scheduler:" + in.IdempotencyKey, SessionID: session.ID, ModelID: model, ThinkingLevel: settings.Agent.ThinkingLevel, IntervalMinutes: in.IntervalSeconds / 60, DailyChapterCap: in.DailyChapterCap, MaxConsecutiveFailures: in.MaxConsecutiveFailures, Enabled: true, Prompt: prompt.SchedulerContinueInstruction, ModeInputs: selected.ModeInputs, KnowledgeIDs: selected.KnowledgeIDs, NoteIDs: selected.NoteIDs, StyleProfileID: selected.StyleProfileID})
		if e != nil {
			return e
		}
		raw, e := json.Marshal(out)
		if e != nil {
			return e
		}
		return tx.CompleteSubmission(ctx, owner, "scheduler_start", in.IdempotencyKey, submissionReceipt{Fingerprint: digest, ProjectID: project, Result: raw})
	})
	return
}
