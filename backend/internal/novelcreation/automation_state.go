package novelcreation

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// AutomationState is stored with the plan's input JSON. Changes to accounting,
// the active round and new runs commit in the same project transaction.
type AutomationState struct {
	Status                 string            `json:"status"`
	DailyChapterCap        int               `json:"dailyChapterCap"`
	MaxConsecutiveFailures int               `json:"maxConsecutiveFailures"`
	ConsecutiveFailures    int               `json:"consecutiveFailures"`
	TotalSavedChapters     int               `json:"totalSavedChapters"`
	DailySavedChapters     int               `json:"dailySavedChapters"`
	SuccessDay             string            `json:"successDay"`
	StoppedReason          string            `json:"stoppedReason,omitempty"`
	LastFailure            string            `json:"lastFailure,omitempty"`
	WaitingReason          string            `json:"waitingReason,omitempty"`
	EligibleAt             *time.Time        `json:"eligibleAt,omitempty"`
	ActiveRound            string            `json:"activeRound,omitempty"`
	Rounds                 []AutomationRound `json:"rounds"`
}
type AutomationRound struct {
	Key           string     `json:"key"`
	PlannedAt     time.Time  `json:"plannedAt"`
	Immediate     bool       `json:"immediate"`
	RunID         string     `json:"runId"`
	SourceRunID   string     `json:"sourceRunId,omitempty"`
	PreviousRunID string     `json:"previousRunId,omitempty"`
	State         string     `json:"state"`
	ErrorCode     string     `json:"errorCode,omitempty"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
}

var shanghai = time.FixedZone("Asia/Shanghai", 8*60*60)

func (s *Service) automationNow() time.Time {
	if s.automationClock != nil {
		return s.automationClock().UTC()
	}
	return time.Now().UTC()
}
func (a *AutomationState) defaults() {
	if a.DailyChapterCap == 0 {
		a.DailyChapterCap = 10
	}
	if a.MaxConsecutiveFailures == 0 {
		a.MaxConsecutiveFailures = 3
	}
}
func (a *AutomationState) day(now time.Time) {
	day := now.In(shanghai).Format("2006-01-02")
	if a.SuccessDay != day {
		a.SuccessDay = day
		a.DailySavedChapters = 0
	}
	if a.EligibleAt != nil && !now.Before(*a.EligibleAt) {
		a.EligibleAt = nil
	}
}
func (a *AutomationState) active() *AutomationRound {
	for i := range a.Rounds {
		if a.Rounds[i].Key == a.ActiveRound {
			return &a.Rounds[i]
		}
	}
	return nil
}
func temporary(code string) bool {
	return code == ErrModelRateLimited || code == ErrModelOverloaded || code == ErrModelConnectionLost || code == ErrGenerationTimeout
}
func automationWaitReason(state string) string {
	switch state {
	case "awaiting_confirmation":
		return "请确认创作蓝图后继续"
	case "save_pending":
		return "结果保存失败，等待重试保存"
	case "ready":
		return "等待执行或恢复当前运行"
	default:
		return "当前章节正在创作"
	}
}
func busy(err error) bool { return IsCode(err, ErrRunInProgress) || IsCode(err, ErrProjectLocked) }
func errorCode(err error) string {
	var d *DomainError
	if errors.As(err, &d) {
		return d.Code
	}
	return ErrPersistenceFailed
}

// inspectRound follows a confirmed blueprint into its successor. A checkpoint,
// ready run or save_pending result is never consumed as a saved chapter.
func (s *Service) inspectRound(ctx context.Context, owner string, r *AutomationRound) (string, error) {
	run, err := s.store.GetRun(ctx, owner, r.RunID)
	if err != nil {
		return "", err
	}
	if run.Status == RunAwaitingConfirmation || run.Status == RunSucceeded {
		next, ok, e := s.store.FindRunByIdempotencyKey(ctx, owner, "confirm:"+run.ID)
		if e != nil {
			return "", e
		}
		if !ok {
			if run.Status == RunAwaitingConfirmation {
				r.State = "awaiting_confirmation"
				return "waiting", nil
			}
		} else {
			if r.SourceRunID == "" {
				r.SourceRunID = run.ID
			}
			r.RunID = next.ID
			return s.inspectRound(ctx, owner, r)
		}
	}
	execution, exists, executionErr := s.store.GetExecution(ctx, owner, run.ID)
	if executionErr != nil {
		return "", executionErr
	}
	if exists && execution.State == "paused" {
		r.ErrorCode = ErrResultUnknown
		r.ErrorMessage = "上游结果待确认，已暂停自动重试"
		return "halt", nil
	}
	if run.Status == RunPaused || run.ErrorCode == ErrResultUnknown {
		r.ErrorCode = ErrResultUnknown
		r.ErrorMessage = run.ErrorMessage
		return "halt", nil
	}
	if run.Status == RunSucceeded || run.Status == RunSucceededWithWarnings {
		saved, e := s.store.HasSavedRun(ctx, owner, run.ID)
		if e != nil {
			return "", e
		}
		if saved {
			r.FinishedAt = run.FinishedAt
			return "success", nil
		}
		r.ErrorCode = ErrPersistenceFailed
		r.ErrorMessage = "运行未保存有效章节，请检查结果"
		return "halt", nil
	}
	if run.Status == RunFailed || run.Status == RunCancelled {
		// Local journal/save failures remain on the same durable run.
		if exists && (execution.State == "ready" || execution.State == "save_pending") && !execution.CancelRequested {
			r.State = execution.State
			return "waiting", nil
		}
		r.ErrorCode = run.ErrorCode
		r.ErrorMessage = run.ErrorMessage
		if run.Status == RunCancelled {
			return "cancelled", nil
		}
		if temporary(run.ErrorCode) {
			return "failure", nil
		}
		return "halt", nil
	}
	r.State = string(run.Status)
	return "waiting", nil
}
func (s *MemoryStore) HasSavedRun(_ context.Context, owner, id string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for key, v := range s.chapterVersions {
		if v.RunID == id && strings.HasPrefix(key, owner+"\x00") {
			return true, nil
		}
	}
	return false, nil
}
func (s *MySQLStore) HasSavedRun(ctx context.Context, owner, id string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM novel_chapter_versions WHERE owner_username = ? AND run_id = ? LIMIT 1`, owner, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, persistenceError(err)
}
func (s *MemoryStore) InsertAutomationRound(_ context.Context, owner, kind, plan, key, run string, planned time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := memoryKey(owner, kind, plan, key)
	if _, ok := s.automationRounds[k]; ok {
		return conflict()
	}
	s.automationRounds[k] = run
	return nil
}
func (s *MySQLStore) InsertAutomationRound(ctx context.Context, owner, kind, plan, key, run string, planned time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO novel_automation_rounds (owner_username, kind, plan_id, round_key, run_id, planned_at) VALUES (?, ?, ?, ?, ?, ?)`, owner, kind, plan, key, run, planned.UTC())
	return persistenceError(err)
}
func (s *Service) automationChild(tx Store) *Service {
	return &Service{store: tx, runtime: quietRuntime{s.runtime}, pipeline: s.pipeline, control: s.control, automationClock: s.automationClock}
}

// CancelRun's durable changes participate in the outer automation transaction.
// Redis/context signals must wait for that transaction's commit.
type automationQuietCancel struct{ Runtime }

func (automationQuietCancel) Cancel(context.Context, string) error { return nil }
func (s *Service) cancelAutomationInTx(ctx context.Context, tx Store, owner, id string) error {
	child := s.automationChild(tx)
	child.control = newWorkerControl()
	child.runtime = automationQuietCancel{child.runtime}
	_, err := child.CancelRun(ctx, owner, id)
	return err
}
func (s *Service) signalAutomationCancel(ctx context.Context, owner, id string) {
	if id == "" {
		return
	}
	s.control.mu.Lock()
	cancel := s.control.cancel[memoryKey(owner, id)].cancel
	s.control.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	_ = s.runtime.Cancel(ctx, id)
}

// Automation waits before appending a message. The project transaction protects
// this check against concurrent automation dispatches in other processes.
func (s *Service) queueAutomationMessage(ctx context.Context, owner string, input SendMessageInput) (Run, error) {
	normalized, err := NormalizeSendMessageInput(input)
	if err != nil {
		return Run{}, err
	}
	input = normalized
	if _, err = s.prepareGenerationInput(ctx, owner, input.ProjectID, input.SessionID, AutomationInput{ModeInputs: input.ModeInputs, KnowledgeIDs: input.KnowledgeIDs, NoteIDs: input.NoteIDs, StyleProfileID: input.StyleProfileID}); err != nil {
		return Run{}, err
	}
	runs, err := s.store.ListRuns(ctx, owner, input.ProjectID)
	if err != nil {
		return Run{}, err
	}
	for _, r := range runs {
		occupied := !isTerminalRunStatus(r.Status)
		if r.Status == RunFailed {
			e, ok, err := s.store.GetExecution(ctx, owner, r.ID)
			if err != nil {
				return Run{}, err
			}
			occupied = ok && (e.State == "ready" || e.State == "save_pending" || e.State == "paused") && !e.CancelRequested
		}
		if occupied {
			return Run{}, &DomainError{Code: ErrRunInProgress, Message: "该作品正在创作，等待空闲"}
		}
	}
	return s.SendMessage(ctx, owner, input)
}

// ScheduleCursor uses ordered keys rather than an offset into a changing
// candidate set. The scanner captures its pages before settling any rounds.
// Business due dates are not changed to make scans fair.
type ScheduleCursor struct {
	NextRunAt         time.Time
	OwnerUsername, ID string
}

func scheduleAfter(p Schedule, c ScheduleCursor) bool {
	if c.ID == "" {
		return true
	}
	if !p.NextRunAt.Equal(c.NextRunAt) {
		return p.NextRunAt.After(c.NextRunAt)
	}
	if p.OwnerUsername != c.OwnerUsername {
		return p.OwnerUsername > c.OwnerUsername
	}
	return p.ID > c.ID
}
func restoreLegacyScheduleRound(p *Schedule) {
	if p.ActiveRound != "" || len(p.Rounds) != 0 || p.LastRunID == "" {
		return
	}
	planned := p.NextRunAt
	if p.LastRunAt != nil {
		planned = *p.LastRunAt
	}
	key := "legacy:" + p.LastRunID
	p.Rounds = append(p.Rounds, AutomationRound{Key: key, RunID: p.LastRunID, PlannedAt: planned})
	p.ActiveRound = key
}
func restoreLegacyBatchRound(p *Batch, now time.Time) {
	if p.ActiveRound != "" || p.CurrentRunID == "" {
		return
	}
	key := "legacy:" + p.CurrentRunID
	p.Rounds = append(p.Rounds, AutomationRound{Key: key, RunID: p.CurrentRunID, PlannedAt: now})
	p.ActiveRound = key
}
