package aicut

import (
	"context"
	"sync"
	"time"
)

type AgentStore interface {
	CreateAgentRun(context.Context, AgentRun) (AgentRun, error)
	AgentRun(context.Context, string, string, string) (AgentRun, error)
	ClaimAgentRun(context.Context, string, time.Time) (AgentRun, error)
	UpdateAgentRunState(context.Context, string, AgentRunState, string, int, time.Time) (AgentRun, error)
	RequestAgentCancel(context.Context, string, string, string, time.Time) (AgentRun, error)
	AppendAgentMessage(context.Context, AgentMessage) error
	AgentMessages(context.Context, string, string, int) ([]AgentMessage, error)
	AppendAgentEvent(context.Context, AgentEvent) (AgentEvent, error)
	AgentEvents(context.Context, string, string, string, int64) ([]AgentEvent, error)
	AgentToolResult(context.Context, string, string, string, string) (ToolResult, bool, error)
	SaveAgentToolResult(context.Context, string, string, string, ToolResult) error
}

type MemoryAgentStore struct {
	mu       sync.Mutex
	runs     map[string]AgentRun
	requests map[string]string
	messages []AgentMessage
	events   map[string][]AgentEvent
	tools    map[string]ToolResult
}

func NewMemoryAgentStore() *MemoryAgentStore {
	return &MemoryAgentStore{runs: map[string]AgentRun{}, requests: map[string]string{}, events: map[string][]AgentEvent{}, tools: map[string]ToolResult{}}
}

func agentRequestKey(owner, projectID, clientRequestID string) string {
	return owner + "\x00" + projectID + "\x00" + clientRequestID
}

func (s *MemoryAgentStore) CreateAgentRun(_ context.Context, run AgentRun) (AgentRun, error) {
	normalized, err := normalizeAgentRun(run)
	if err != nil {
		return AgentRun{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := agentRequestKey(normalized.Owner, normalized.ProjectID, normalized.ClientRequestID)
	if id := s.requests[key]; id != "" {
		existing := s.runs[id]
		if canonicalASRTimeline(existing.TimelineID) != canonicalASRTimeline(normalized.TimelineID) {
			return AgentRun{}, ErrInvalidAgentRun
		}
		return existing, nil
	}
	s.runs[normalized.ID] = normalized
	s.requests[key] = normalized.ID
	return normalized, nil
}

func (s *MemoryAgentStore) AgentRun(_ context.Context, owner, projectID, id string) (AgentRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok || run.Owner != owner || run.ProjectID != projectID {
		return AgentRun{}, ErrAgentRunNotFound
	}
	return run, nil
}

func (s *MemoryAgentStore) ClaimAgentRun(_ context.Context, id string, now time.Time) (AgentRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		return AgentRun{}, ErrAgentRunNotFound
	}
	if run.State != AgentQueued {
		return AgentRun{}, ErrInvalidAgentTransition
	}
	run.State, run.UpdatedAt = AgentRunning, now.UTC()
	s.runs[id] = run
	return run, nil
}

func (s *MemoryAgentStore) UpdateAgentRunState(_ context.Context, id string, state AgentRunState, message string, turnCount int, now time.Time) (AgentRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok {
		return AgentRun{}, ErrAgentRunNotFound
	}
	if !validAgentTransition(run.State, state) || turnCount < run.TurnCount || turnCount > 30 {
		return AgentRun{}, ErrInvalidAgentTransition
	}
	run.State, run.Error, run.TurnCount, run.UpdatedAt = state, message, turnCount, now.UTC()
	s.runs[id] = run
	return run, nil
}

func (s *MemoryAgentStore) RequestAgentCancel(_ context.Context, owner, projectID, id string, now time.Time) (AgentRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[id]
	if !ok || run.Owner != owner || run.ProjectID != projectID {
		return AgentRun{}, ErrAgentRunNotFound
	}
	if run.State == AgentSucceeded || run.State == AgentFailed || run.State == AgentCanceled {
		return run, nil
	}
	if run.State != AgentCancelRequested {
		when := now.UTC()
		run.State, run.CancelRequestedAt, run.UpdatedAt = AgentCancelRequested, &when, when
		s.runs[id] = run
	}
	return run, nil
}

func (s *MemoryAgentStore) AppendAgentMessage(_ context.Context, message AgentMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[message.RunID]
	if !ok || run.Owner != message.Owner || run.ProjectID != message.ProjectID {
		return ErrAgentRunNotFound
	}
	s.messages = append(s.messages, message)
	return nil
}

func (s *MemoryAgentStore) AgentMessages(_ context.Context, owner, projectID string, limit int) ([]AgentMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]AgentMessage, 0)
	for _, message := range s.messages {
		if message.Owner == owner && message.ProjectID == projectID {
			items = append(items, message)
		}
	}
	if limit > 0 && len(items) > limit {
		items = items[len(items)-limit:]
	}
	return items, nil
}

func (s *MemoryAgentStore) AppendAgentEvent(_ context.Context, event AgentEvent) (AgentEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[event.RunID]
	if !ok || run.Owner != event.Owner || run.ProjectID != event.ProjectID {
		return AgentEvent{}, ErrAgentRunNotFound
	}
	event.Sequence = int64(len(s.events[event.RunID]) + 1)
	s.events[event.RunID] = append(s.events[event.RunID], event)
	return event, nil
}

func (s *MemoryAgentStore) AgentEvents(_ context.Context, owner, projectID, runID string, after int64) ([]AgentEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[runID]
	if !ok || run.Owner != owner || run.ProjectID != projectID {
		return nil, ErrAgentRunNotFound
	}
	items := make([]AgentEvent, 0)
	for _, event := range s.events[runID] {
		if event.Sequence > after {
			items = append(items, event)
		}
	}
	return items, nil
}

func (s *MemoryAgentStore) AgentToolResult(_ context.Context, owner, projectID, runID, callID string) (ToolResult, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[runID]
	if !ok || run.Owner != owner || run.ProjectID != projectID {
		return ToolResult{}, false, ErrAgentRunNotFound
	}
	result, exists := s.tools[runID+"\x00"+callID]
	return result, exists, nil
}

func (s *MemoryAgentStore) SaveAgentToolResult(_ context.Context, owner, projectID, runID string, result ToolResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[runID]
	if !ok || run.Owner != owner || run.ProjectID != projectID {
		return ErrAgentRunNotFound
	}
	key := runID + "\x00" + result.CallID
	if _, exists := s.tools[key]; !exists {
		s.tools[key] = result
	}
	return nil
}

func (s *MemoryAgentStore) AgentTimelineMessages(_ context.Context, o, p, id string, limit int) ([]AgentMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []AgentMessage{}
	for _, m := range s.messages {
		run, ok := s.runs[m.RunID]
		if ok && m.Owner == o && m.ProjectID == p && canonicalASRTimeline(run.TimelineID) == canonicalASRTimeline(id) {
			out = append(out, m)
		}
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}
