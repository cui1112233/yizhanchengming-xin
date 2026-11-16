package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
)

func (s *Service) finishModelCallConfirmation(ctx context.Context, owner string, run Run) (Run, error) {
	e, ok, err := s.store.GetExecution(ctx, owner, run.ID)
	if err != nil {
		return run, err
	}
	if !ok || e.PendingConfirmation == nil {
		return run, &DomainError{Code: ErrPersistenceFailed, Message: "模型调用确认记录缺失"}
	}
	raw, _ := json.Marshal(map[string]any{"confirmationKind": "next_model_call", "confirmationId": e.PendingConfirmation.ID, "callKey": e.PendingConfirmation.Key, "nextModelCall": e.PendingConfirmation.Metadata, "agentSettings": agentFromSnapshot(run.PromptSnapshot), "modelCalls": callMetadata(e)})
	g, ok := s.store.(*guardedStore)
	if !ok {
		return run, leaseLost()
	}
	out, event, err := g.finishRun(ctx, owner, RunUpdate{ID: run.ID, Status: RunAwaitingConfirmation, RuntimeSnapshot: raw}, nil, RunEvent{RunID: run.ID, ProjectID: run.ProjectID, SessionID: run.SessionID, Type: "run.checkpoint", Status: RunAwaitingConfirmation, Message: "请确认后继续下一次模型调用"})
	if err == nil {
		_ = s.runtime.Publish(ctx, event)
	}
	return out, err
}

// ConfirmModelCall authorizes one exact request. Repeating an older token is an
// idempotent read, including after the next checkpoint has already been reached.
func (s *Service) ConfirmModelCall(ctx context.Context, owner, id, token string) (Run, error) {
	owner, id, token = strings.TrimSpace(owner), strings.TrimSpace(id), strings.TrimSpace(token)
	run, err := s.store.GetRun(ctx, owner, id)
	if err != nil {
		return run, err
	}
	err = s.store.Atomic(ctx, owner, run.ProjectID, func(tx Store) error {
		var e error
		run, e = tx.GetRun(ctx, owner, id)
		if e != nil {
			return e
		}
		state, ok, e := tx.GetExecution(ctx, owner, id)
		if e != nil {
			return e
		}
		if !ok || token == "" {
			return &DomainError{Code: ErrInputRequired, Message: "缺少模型调用确认编号"}
		}
		if _, confirmed := state.ConfirmedCalls[token]; confirmed {
			return nil
		}
		if run.Status != RunAwaitingConfirmation || state.CancelRequested || state.PendingConfirmation == nil || state.PendingConfirmation.ID != token {
			return &DomainError{Code: ErrInvalidTransition, Message: "该确认编号已过期，请刷新当前任务"}
		}
		pending := state.PendingConfirmation
		call, ok := state.Calls[pending.Key]
		if !ok || call.State != "prepared" || call.ConfirmationID != token {
			return conflict()
		}
		call.Confirmed = true
		state.Calls[pending.Key] = call
		if state.ConfirmedCalls == nil {
			state.ConfirmedCalls = map[string]string{}
		}
		state.ConfirmedCalls[token] = pending.Key
		state.PendingConfirmation = nil
		state.State = "ready"
		if e = tx.PutExecution(ctx, state); e != nil {
			return e
		}
		var runtimeFields map[string]any
		_ = json.Unmarshal(run.RuntimeSnapshot, &runtimeFields)
		if runtimeFields == nil {
			runtimeFields = map[string]any{}
		}
		delete(runtimeFields, "confirmationId")
		delete(runtimeFields, "callKey")
		delete(runtimeFields, "nextModelCall")
		runtimeFields["confirmedId"] = token
		runtime, _ := json.Marshal(runtimeFields)
		run, e = tx.UpdateRun(ctx, owner, RunUpdate{ID: id, Status: RunQueued, RuntimeSnapshot: runtime})
		if e != nil {
			return e
		}
		_, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: id, ProjectID: run.ProjectID, SessionID: run.SessionID, Status: RunQueued, Type: "run.queued", Message: "已确认，继续当前创作步骤"})
		return e
	})
	if err == nil {
		s.Notify(ctx, owner, id)
	}
	return run, err
}

func callMetadata(e Execution) map[string]*ModelCallMetadata {
	out := map[string]*ModelCallMetadata{}
	for key, c := range e.Calls {
		if c.Metadata != nil {
			out[key] = c.Metadata
		}
	}
	return out
}
func agentFromSnapshot(raw json.RawMessage) any {
	var v struct {
		Settings ProjectSettings `json:"project_settings"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v.Settings.Agent
}
