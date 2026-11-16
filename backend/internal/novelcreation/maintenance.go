package novelcreation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	novelprompt "qiantie/backend/internal/prompt/novelcreation"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

const RunAnalyzing RunStatus = "analyzing"

type MaintenanceInput struct {
	Kind           string `json:"kind"`
	SessionID      string `json:"sessionId"`
	StyleID        string `json:"styleId,omitempty"`
	IdempotencyKey string `json:"idempotencyKey"`
}
type maintenanceSnapshot struct {
	Operation            string          `json:"operation"`
	AgentContractVersion int             `json:"agent_contract_version"`
	Settings             ProjectSettings `json:"project_settings"`
	Chapters             []Chapter       `json:"chapters"`
	Sources              []ChapterSource `json:"sources"`
	Revision             int             `json:"revision"`
	PreviousTruth        json.RawMessage `json:"previous_truth"`
	PreviousMemory       json.RawMessage `json:"previous_memory"`
	Style                *StyleSource    `json:"style,omitempty"`
	StyleFingerprint     string          `json:"style_fingerprint,omitempty"`
}
type MaintenanceResult struct {
	Snapshot     maintenanceSnapshot `json:"snapshot"`
	Entries      []ContinuityEntry   `json:"entries,omitempty"`
	StyleContent string              `json:"styleContent,omitempty"`
}

func operationKind(run Run) string {
	var v struct {
		Operation string `json:"operation"`
	}
	_ = json.Unmarshal(run.PromptSnapshot, &v)
	return v.Operation
}
func (s *Service) QueueMaintenance(ctx context.Context, owner, project string, in MaintenanceInput) (Run, error) {
	if in.Kind != "continuity_refresh" && in.Kind != "style_analysis" {
		return Run{}, inputError("请选择事实记忆刷新或文风分析")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" || len(in.IdempotencyKey) > 96 {
		return Run{}, inputError("缺少请求编号")
	}
	return s.queueAtomic(ctx, owner, project, fingerprint(in), func(child *Service) (Run, error) {
		tx := child.store
		if old, ok, e := tx.FindRunByIdempotencyKey(ctx, owner, in.IdempotencyKey); e != nil {
			return old, e
		} else if ok {
			return old, nil
		}
		_, e := writableSession(ctx, tx, owner, project, in.SessionID)
		if e != nil {
			return Run{}, e
		}
		p, e := tx.GetProject(ctx, owner, project)
		if e != nil {
			return Run{}, e
		}
		settings, e := tx.GetProjectSettings(ctx, owner, project)
		if e != nil {
			return Run{}, e
		}
		settings.Agent = NormalizeAgentSettings(settings.Agent, p.TargetChapterWords)
		snap := maintenanceSnapshot{Operation: in.Kind, AgentContractVersion: 1, Settings: settings, PreviousTruth: json.RawMessage(`{}`), PreviousMemory: json.RawMessage(`{}`)}
		if in.Kind == "continuity_refresh" {
			chapters, e := tx.ListChapters(ctx, owner, project)
			if e != nil {
				return Run{}, e
			}
			if len(chapters) == 0 {
				return Run{}, inputError("暂无已保存章节可刷新")
			}
			state, e := child.GetContinuityState(ctx, owner, project)
			if e != nil {
				return Run{}, e
			}
			snap.Revision = state.Revision
			snap.Sources = chapterSources(chapters)
			from := state.DirtyFrom
			if from == 0 {
				from = 1
			}
			before := []Chapter{}
			total := 0
			for _, c := range chapters {
				if c.Number < from {
					before = append(before, c)
				} else {
					snap.Chapters = append(snap.Chapters, c)
					total += utf8.RuneCountInString(c.Content)
				}
			}
			if total > 2_000_000 {
				return Run{}, inputError("待刷新正文超过200万字元，请先整理作品")
			}
			if len(before) > 0 {
				previous, e := child.loadContinuity(ctx, owner, project, before)
				if e != nil {
					return Run{}, e
				}
				snap.PreviousTruth = previous.Truth
				snap.PreviousMemory = previous.Memory
			}
		} else {
			var source StyleSource
			ok, e := readContent(ctx, tx, owner, project, "style", in.StyleID, &source)
			if e != nil {
				return Run{}, e
			}
			if !ok {
				return Run{}, inputError("文风来源不存在")
			}
			profile, e := tx.GetStyleProfile(ctx, owner, in.StyleID)
			if e != nil {
				return Run{}, e
			}
			source.Profile = profile
			snap.Style = &source
			snap.StyleFingerprint = fingerprint(profile)
		}
		raw, _ := json.Marshal(snap)
		run, e := tx.CreateRun(ctx, owner, Run{ProjectID: project, SessionID: in.SessionID, IdempotencyKey: in.IdempotencyKey, Status: RunQueued, ModelID: p.ModelID, ThinkingLevel: settings.Agent.ThinkingLevel, PromptSnapshot: raw})
		if e != nil {
			return run, e
		}
		if snap.Style != nil {
			v := *snap.Style
			v.AnalysisRunID = run.ID
			v.Status = "queued"
			if e = putContent(ctx, tx, owner, project, "style", in.StyleID, v.Status, time.Time{}, v); e != nil {
				return run, e
			}
		}
		_, e = child.appendEvent(ctx, owner, run, "run.queued", RunQueued, "分析任务已加入队列", 0, map[string]any{"operation": in.Kind})
		return run, e
	})
}
func (s *Service) runMaintenance(ctx context.Context, owner string, run Run) (Run, error) {
	var snap maintenanceSnapshot
	if e := json.Unmarshal(run.PromptSnapshot, &snap); e != nil {
		return s.finishWithError(ctx, owner, run, ErrInvalidOutput, "分析输入快照损坏", e)
	}
	if e := s.validateMaintenanceSources(ctx, owner, run.ProjectID, snap); e != nil {
		return s.finishWithError(ctx, owner, run, ErrInvalidTransition, "来源已改版，请重新发起分析", e)
	}
	p, ok := s.pipeline.(*Pipeline)
	if !ok || p.resolveProvider == nil {
		return s.finishWithError(ctx, owner, run, ErrModelUnavailable, "分析模型尚未配置", nil)
	}
	project, e := s.store.GetProject(ctx, owner, run.ProjectID)
	if e != nil {
		return run, e
	}
	session, e := s.store.GetSession(ctx, owner, run.SessionID)
	if e != nil {
		return run, e
	}
	if session.ArchivedAt != nil {
		return s.finishWithError(ctx, owner, run, ErrInvalidTransition, "会话已归档", nil)
	}
	input := RunContext{OwnerUsername: owner, Run: run, Project: project, Session: session, Agent: &snap.Settings.Agent, Runtime: s.runtime}
	sink := serviceEventSink{s, owner}
	if e = p.emitStage(ctx, input, sink, RunAnalyzing, "正在分析已保存资料", 5); e != nil {
		return run, e
	}
	result := MaintenanceResult{Snapshot: snap}
	if snap.Operation == "style_analysis" {
		if snap.Style == nil {
			return s.finishWithError(ctx, owner, run, ErrInputRequired, "文风来源缺失", nil)
		}
		var out struct {
			Content string `json:"content"`
		}
		_, e = p.callJSON(ctx, input, sink, nil, modelCallSpec{definition: novelprompt.StyleAnalysis, step: RunAnalyzing, userPayload: map[string]any{"reference_text": snap.Style.Source.Content, "statistics": snap.Style.Statistics, "user_description": snap.Style.UserDescription}, temperature: .2, maxTokens: 3500, validate: func() error {
			if strings.TrimSpace(out.Content) == "" || utf8.RuneCountInString(out.Content) > 12000 {
				return inputError("文风分析正文为空或过长")
			}
			return nil
		}}, &out)
		result.StyleContent = out.Content
	} else {
		truth, memory := snap.PreviousTruth, snap.PreviousMemory
		for _, chapter := range snap.Chapters {
			var settlement truthSettlementOutput
			var index memoryIndexOutput
			key := fmt.Sprintf(":refresh:%d", chapter.Number)
			var rawTruth, rawMemory json.RawMessage
			rawTruth, e = p.callJSON(ctx, input, sink, nil, modelCallSpec{definition: novelprompt.TruthSettlement, key: key, step: RunAnalyzing, userPayload: map[string]any{"chapter_content": chapter.Content, "previous_truth": rawJSONValue(truth), "chapter_number": chapter.Number}, temperature: .2, maxTokens: 4000, validate: func() error { return validateSettlement(settlement) }}, &settlement)
			if e != nil {
				break
			}
			rawMemory, e = p.callJSON(ctx, input, sink, nil, modelCallSpec{definition: novelprompt.MemoryIndex, key: key, step: RunAnalyzing, userPayload: map[string]any{"chapter_content": chapter.Content, "truth_settlement": rawJSONValue(rawTruth), "previous_memory": rawJSONValue(memory), "chapter_number": chapter.Number}, temperature: .2, maxTokens: 3500, validate: func() error { return validateMemory(index) }}, &index)
			if e != nil {
				break
			}
			result.Entries = append(result.Entries, ContinuityEntry{Source: chapterSources([]Chapter{chapter})[0], Truth: rawTruth, Memory: rawMemory})
			truth, memory = rawTruth, rawMemory
		}
	}
	if e != nil {
		if IsCode(e, ErrSecondaryConfirmation) {
			return s.finishModelCallConfirmation(ctx, owner, run)
		}
		if isDurableWriteFailure(e) || IsCode(e, ErrPersistenceFailed) {
			return s.recoverDurableWriteFailure(ctx, owner, run, e)
		}
		code, message := domainFailure(e)
		status := RunFailed
		if code == ErrResultUnknown {
			status = RunPaused
		}
		if errors.Is(e, context.Canceled) || code == ErrRunCancelled {
			status = RunCancelled
		}
		return s.finish(ctx, owner, run, status, code, message, nil, e)
	}
	j, _ := ctx.Value(journalKey{}).(*callJournal)
	if j == nil {
		return run, leaseLost()
	}
	if e = guarded(ctx, j.store, j.lease, func(tx Store, ex *Execution) error { ex.Maintenance = &result; ex.State = "save_pending"; return nil }); e != nil {
		return run, e
	}
	return s.saveMaintenance(ctx, owner, run, result)
}
func (s *Service) saveMaintenance(ctx context.Context, owner string, run Run, result MaintenanceResult) (Run, error) {
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, leaseLost()
	}
	err := guarded(ctx, g.Store, g.lease, func(tx Store, ex *Execution) error {
		hash := fingerprint(result)
		if ex.CommittedHash != "" {
			if ex.CommittedHash == hash {
				return nil
			}
			return conflict()
		}
		snap := result.Snapshot
		check := &Service{store: tx}
		if e := check.validateMaintenanceSources(ctx, owner, run.ProjectID, snap); e != nil {
			return conflict()
		}
		if snap.Operation == "continuity_refresh" {
			chapters, e := tx.ListChapters(ctx, owner, run.ProjectID)
			if e != nil {
				return e
			}
			if !reflect.DeepEqual(chapterSources(chapters), snap.Sources) {
				return conflict()
			}
			var old ContinuityState
			if _, e = readContent(ctx, tx, owner, run.ProjectID, "continuity", "current", &old); e != nil {
				return e
			}
			if old.Revision != snap.Revision {
				return conflict()
			}
			state := ContinuityState{Revision: old.Revision, AnalysisVersion: old.AnalysisVersion + 1, RunID: run.ID, Sources: snap.Sources}
			if len(result.Entries) == 0 {
				return conflict()
			}
			for _, entry := range old.Entries {
				if entry.Source.Number < result.Entries[0].Source.Number {
					state.Entries = append(state.Entries, entry)
				}
			}
			state.Entries = append(state.Entries, result.Entries...)
			if e = putContent(ctx, tx, owner, run.ProjectID, "continuity", novelID("analysis"), "fresh", time.Time{}, state); e != nil {
				return e
			}
			if e = putContent(ctx, tx, owner, run.ProjectID, "continuity", "current", "fresh", time.Time{}, state); e != nil {
				return e
			}
		} else {
			if snap.Style == nil {
				return conflict()
			}
			p, e := tx.GetStyleProfile(ctx, owner, snap.Style.Profile.ID)
			if e != nil {
				return e
			}
			if fingerprint(p) != snap.StyleFingerprint {
				return conflict()
			}
			source := *snap.Style
			p.Content = result.StyleContent
			source.Profile, e = tx.SaveStyleProfile(ctx, owner, p)
			if e != nil {
				return e
			}
			source.AnalysisVersion++
			source.AnalysisRunID = run.ID
			source.Status = "succeeded"
			if e = putContent(ctx, tx, owner, run.ProjectID, "style", p.ID, "succeeded", time.Time{}, source); e != nil {
				return e
			}
		}
		runtime, _ := json.Marshal(map[string]any{"operation": snap.Operation, "modelCalls": callMetadata(*ex), "agentSettings": snap.Settings.Agent, "saved": true})
		now := time.Now().UTC()
		if _, e := tx.UpdateRun(ctx, owner, RunUpdate{ID: run.ID, Status: RunSucceeded, RuntimeSnapshot: runtime, FinishedAt: &now}); e != nil {
			return e
		}
		if _, e := tx.AppendRunEvent(ctx, owner, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.succeeded", Status: RunSucceeded, Progress: 100, Message: "分析结果已保存"}); e != nil {
			return e
		}
		ex.CommittedHash = hash
		ex.State = "done"
		return nil
	})
	if IsCode(err, ErrIdempotencyConflict) {
		return s.finishWithError(ctx, owner, run, ErrInvalidTransition, "来源已改版，分析结果未覆盖最新内容，请重新发起分析", err)
	}
	if err != nil {
		return run, err
	}
	return s.store.GetRun(ctx, owner, run.ID)
}

func (s *Service) validateMaintenanceSources(ctx context.Context, owner, project string, snap maintenanceSnapshot) error {
	if snap.Operation == "continuity_refresh" {
		ch, e := s.store.ListChapters(ctx, owner, project)
		if e != nil {
			return e
		}
		if !reflect.DeepEqual(chapterSources(ch), snap.Sources) {
			return conflict()
		}
		state, e := s.GetContinuityState(ctx, owner, project)
		if e != nil {
			return e
		}
		if state.Revision != snap.Revision {
			return conflict()
		}
		return nil
	}
	if snap.Style == nil {
		return conflict()
	}
	p, e := s.store.GetStyleProfile(ctx, owner, snap.Style.Profile.ID)
	if e != nil {
		return e
	}
	if fingerprint(p) != snap.StyleFingerprint {
		return conflict()
	}
	source := snap.Style.Source
	if documentID(source.ID) {
		j, e := getDocument(ctx, s.store, owner, project, strings.TrimPrefix(source.ID, "docref-"))
		if e != nil {
			return e
		}
		selected, e := selectedDocument(j, DocumentSelection{source.SourceVersion, source.SelectedChapterIDs})
		if e != nil {
			return e
		}
		if fingerprint(selected) != fingerprint(source) {
			return conflict()
		}
	}
	return nil
}
