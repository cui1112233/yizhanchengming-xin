package httpapi

import (
	"context"
	"qiantie/backend/internal/aicut"
	"time"
)

func storedAICutTimelineID(id string) string {
	if id == "" {
		return aicut.DefaultTimelineID
	}
	return id
}

type aicutRequestAgentStore struct{ aicut.AgentStore }

func (s aicutRequestAgentStore) CreateAgentRun(ctx context.Context, run aicut.AgentRun) (aicut.AgentRun, error) {
	run.TimelineID = aicutTimelineIDFromContext(ctx)
	if run.TimelineID == aicut.DefaultTimelineID {
		run.TimelineID = ""
	}
	out, e := s.AgentStore.CreateAgentRun(ctx, run)
	if e == nil && storedAICutTimelineID(out.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.AgentRun{}, aicut.ErrAgentRunNotFound
	}
	return out, e
}
func (s aicutRequestAgentStore) AgentRun(ctx context.Context, o, p, id string) (aicut.AgentRun, error) {
	run, e := s.AgentStore.AgentRun(ctx, o, p, id)
	if e != nil {
		return aicut.AgentRun{}, e
	}
	if storedAICutTimelineID(run.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.AgentRun{}, aicut.ErrAgentRunNotFound
	}
	return run, nil
}
func (s aicutRequestAgentStore) RequestAgentCancel(ctx context.Context, o, p, id string, n time.Time) (aicut.AgentRun, error) {
	if _, e := s.AgentRun(ctx, o, p, id); e != nil {
		return aicut.AgentRun{}, e
	}
	return s.AgentStore.RequestAgentCancel(ctx, o, p, id, n)
}
func (s aicutRequestAgentStore) AgentEvents(ctx context.Context, o, p, id string, after int64) ([]aicut.AgentEvent, error) {
	if _, e := s.AgentRun(ctx, o, p, id); e != nil {
		return nil, e
	}
	return s.AgentStore.AgentEvents(ctx, o, p, id, after)
}
func (s aicutRequestAgentStore) AgentMessages(ctx context.Context, o, p string, limit int) ([]aicut.AgentMessage, error) {
	if reader, ok := s.AgentStore.(interface {
		AgentTimelineMessages(context.Context, string, string, string, int) ([]aicut.AgentMessage, error)
	}); ok {
		return reader.AgentTimelineMessages(ctx, o, p, aicutTimelineIDFromContext(ctx), limit)
	}
	messages, e := s.AgentStore.AgentMessages(ctx, o, p, 200)
	if e != nil {
		return nil, e
	}
	out := []aicut.AgentMessage{}
	for _, message := range messages {
		if _, e := s.AgentRun(ctx, o, p, message.RunID); e == nil {
			out = append(out, message)
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

type aicutRequestEditPlanStore struct {
	aicut.EditPlanStore
	timelines *aicut.TimelineCollectionStore
}

func (s aicutRequestEditPlanStore) EditPlan(ctx context.Context, o, p, id string) (aicut.EditPlan, error) {
	plan, e := s.EditPlanStore.EditPlan(ctx, o, p, id)
	if e != nil {
		return aicut.EditPlan{}, e
	}
	if storedAICutTimelineID(plan.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.EditPlan{}, aicut.ErrEditPlanNotFound
	}
	return plan, nil
}
func (s aicutRequestEditPlanStore) EditPlanByRequest(ctx context.Context, o, p, id string) (aicut.EditPlan, error) {
	plan, e := s.EditPlanStore.EditPlanByRequest(ctx, o, p, id)
	if e != nil {
		return aicut.EditPlan{}, e
	}
	if storedAICutTimelineID(plan.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.EditPlan{}, aicut.ErrEditPlanNotFound
	}
	return plan, nil
}
func (s aicutRequestEditPlanStore) CreateEditPlan(ctx context.Context, plan aicut.EditPlan) (aicut.EditPlan, error) {
	plan.TimelineID = aicutTimelineIDFromContext(ctx)
	if plan.TimelineID == aicut.DefaultTimelineID {
		plan.TimelineID = ""
	}
	plan.Request.TimelineID = plan.TimelineID
	out, e := s.EditPlanStore.CreateEditPlan(ctx, plan)
	if e == nil && storedAICutTimelineID(out.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.EditPlan{}, aicut.ErrEditPlanNotFound
	}
	return out, e
}
func (s aicutRequestEditPlanStore) UpdateEditPlan(ctx context.Context, plan aicut.EditPlan) (aicut.EditPlan, error) {
	if _, e := s.EditPlan(ctx, plan.Owner, plan.ProjectID, plan.ID); e != nil {
		return aicut.EditPlan{}, e
	}
	if storedAICutTimelineID(plan.TimelineID) != aicutTimelineIDFromContext(ctx) {
		return aicut.EditPlan{}, aicut.ErrEditPlanNotFound
	}
	return s.EditPlanStore.UpdateEditPlan(ctx, plan)
}
func (s aicutRequestEditPlanStore) UpdateEditPlanState(ctx context.Context, o, p, id string, state aicut.EditPlanState, message string, n time.Time) (aicut.EditPlan, error) {
	if _, e := s.EditPlan(ctx, o, p, id); e != nil {
		return aicut.EditPlan{}, e
	}
	return s.EditPlanStore.UpdateEditPlanState(ctx, o, p, id, state, message, n)
}
func (s aicutRequestEditPlanStore) ApplyStoredEditPlan(ctx context.Context, o, p, id string, enabled []string, n time.Time) (aicut.EditPlan, aicut.Timeline, aicut.EditPlanApplySummary, error) {
	if _, e := s.EditPlan(ctx, o, p, id); e != nil {
		return aicut.EditPlan{}, aicut.Timeline{}, aicut.EditPlanApplySummary{}, e
	}
	if s.timelines == nil {
		return s.EditPlanStore.ApplyStoredEditPlan(ctx, o, p, id, enabled, n)
	}
	return s.timelines.ApplyStoredEditPlan(ctx, s.EditPlanStore, o, p, id, aicutTimelineIDFromContext(ctx), enabled, n)
}
