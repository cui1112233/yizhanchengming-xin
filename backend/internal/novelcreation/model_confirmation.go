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
	dispatch := false
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
			if session, sessionErr := writableSession(ctx, tx, owner, run.ProjectID, run.SessionID); sessionErr != nil {
				// Approved tokens are historical reads, even while the session is archived.
				if session.ProjectID == run.ProjectID && session.ArchivedAt != nil {
					return nil
				}
				return sessionErr
			}
			// Legacy deterministic IDs can already be both approved history and
			// the current unapproved pending. Refresh that pending, never reuse
			// the old approval to authorize it.
			if pending := state.PendingConfirmation; pending != nil && pending.ID == token {
				call, exists := state.Calls[pending.Key]
				if exists && call.State == "prepared" && !call.Confirmed && run.Status == RunAwaitingConfirmation && !state.CancelRequested {
					binding := confirmationBinding(id, pending.Key, call.Fingerprint, pending.Metadata)
					issueCallConfirmation(&state, pending.Key, &call, binding, pending.Metadata)
					if e = tx.PutExecution(ctx, state); e != nil {
						return e
					}
					var snapshot map[string]any
					_ = json.Unmarshal(run.RuntimeSnapshot, &snapshot)
					if snapshot == nil {
						snapshot = map[string]any{}
					}
					snapshot["confirmationKind"] = "next_model_call"
					snapshot["confirmationId"] = state.PendingConfirmation.ID
					snapshot["callKey"] = state.PendingConfirmation.Key
					snapshot["nextModelCall"] = state.PendingConfirmation.Metadata
					snapshot["modelCalls"] = callMetadata(state)
					raw, _ := json.Marshal(snapshot)
					run, e = tx.UpdateRun(ctx, owner, RunUpdate{ID: id, Status: RunAwaitingConfirmation, RuntimeSnapshot: raw})
					if e != nil {
						return e
					}
					_, e = tx.AppendRunEvent(ctx, owner, RunEvent{RunID: id, ProjectID: run.ProjectID, SessionID: run.SessionID, Status: RunAwaitingConfirmation, Type: "run.checkpoint", Message: "确认编号已更新，请确认当前模型调用"})
					return e
				}
			}
			return nil
		}
		if run.Status != RunAwaitingConfirmation || state.CancelRequested || state.PendingConfirmation == nil || state.PendingConfirmation.ID != token {
			return &DomainError{Code: ErrInvalidTransition, Message: "该确认编号已过期，请刷新当前任务"}
		}
		if _, e = writableSession(ctx, tx, owner, run.ProjectID, run.SessionID); e != nil {
			return e
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
		dispatch = e == nil
		return e
	})
	if err == nil && dispatch {
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

func confirmationBinding(run, key, request string, metadata *ModelCallMetadata) string {
	return fingerprint(struct {
		Run, Key, Request string
		Metadata          *ModelCallMetadata
	}{run, key, request, metadata})
}

// A durable generation distinguishes A1/B1/A2 even when configuration A recurs.
// Re-observing the same pending never calls this function.
func issueCallConfirmation(e *Execution, key string, record *CallRecord, binding string, metadata *ModelCallMetadata) {
	for {
		record.ConfirmationGeneration++
		candidate := fingerprint(struct {
			Binding    string
			Generation uint64
		}{binding, record.ConfirmationGeneration})
		if candidate == record.ConfirmationID {
			continue
		}
		if _, approved := e.ConfirmedCalls[candidate]; approved {
			continue
		}
		record.ConfirmationID = candidate
		break
	}
	record.ConfirmationFingerprint = binding
	record.Confirmed = false
	record.Metadata = metadata
	e.Calls[key] = *record
	e.PendingConfirmation = &CallConfirmation{ID: record.ConfirmationID, Key: key, Metadata: metadata}
}
