package aicut

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrEditPlanNotFound = errors.New("ai-cut edit plan not found")

type EditPlanStore interface {
	CreateEditPlan(context.Context, EditPlan) (EditPlan, error)
	EditPlan(context.Context, string, string, string) (EditPlan, error)
	EditPlanByRequest(context.Context, string, string, string) (EditPlan, error)
	UpdateEditPlan(context.Context, EditPlan) (EditPlan, error)
	UpdateEditPlanState(context.Context, string, string, string, EditPlanState, string, time.Time) (EditPlan, error)
	ApplyStoredEditPlan(context.Context, string, string, string, []string, time.Time) (EditPlan, Timeline, EditPlanApplySummary, error)
}

func (s *MemoryStore) UpdateEditPlan(_ context.Context, plan EditPlan) (EditPlan, error) {
	normalized, err := NormalizeEditPlan(plan)
	if err != nil {
		return EditPlan{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := editPlanKey(normalized.Owner, normalized.ProjectID, normalized.ID)
	current, ok := s.editPlans[key]
	if !ok {
		return EditPlan{}, ErrEditPlanNotFound
	}
	if canonicalASRTimeline(current.TimelineID) != canonicalASRTimeline(normalized.TimelineID) || current.ClientRequestID != normalized.ClientRequestID || !validEditPlanStateTransition(current.State, normalized.State) {
		return EditPlan{}, ErrInvalidEditPlan
	}
	s.editPlans[key] = cloneEditPlan(normalized)
	return cloneEditPlan(normalized), nil
}

func cloneEditPlan(plan EditPlan) EditPlan {
	raw, _ := json.Marshal(plan)
	var cloned EditPlan
	_ = json.Unmarshal(raw, &cloned)
	// Owner deliberately never leaves the backend JSON contract, so preserve it
	// explicitly when making the in-memory defensive copy.
	cloned.Owner = plan.Owner
	return cloned
}

func editPlanKey(owner, projectID, id string) string {
	return owner + "\x00" + projectID + "\x00" + id
}

func editPlanRequestKey(owner, projectID, requestID string) string {
	return owner + "\x00" + projectID + "\x00" + requestID
}

func validEditPlanStateTransition(from, to EditPlanState) bool {
	if from == to {
		return true
	}
	switch from {
	case EditPlanQueued:
		return to == EditPlanPlanning || to == EditPlanCanceled || to == EditPlanFailed
	case EditPlanPlanning:
		return to == EditPlanReady || to == EditPlanCanceled || to == EditPlanFailed
	case EditPlanReady:
		return to == EditPlanApplied || to == EditPlanCanceled || to == EditPlanStale
	default:
		return false
	}
}

func (s *MemoryStore) CreateEditPlan(_ context.Context, plan EditPlan) (EditPlan, error) {
	normalized, err := NormalizeEditPlan(plan)
	if err != nil {
		return EditPlan{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if item, ok := s.projects[aicutProjectKey(normalized.Owner, normalized.ProjectID)]; !ok || item.project.DeletedAt != nil {
		return EditPlan{}, ErrProjectNotFound
	}
	requestKey := editPlanRequestKey(normalized.Owner, normalized.ProjectID, normalized.ClientRequestID)
	if id := s.editPlanRequests[requestKey]; id != "" {
		existing := s.editPlans[editPlanKey(normalized.Owner, normalized.ProjectID, id)]
		if canonicalASRTimeline(existing.TimelineID) != canonicalASRTimeline(normalized.TimelineID) {
			return EditPlan{}, ErrInvalidEditPlan
		}
		return cloneEditPlan(existing), nil
	}
	key := editPlanKey(normalized.Owner, normalized.ProjectID, normalized.ID)
	if _, exists := s.editPlans[key]; exists {
		return EditPlan{}, ErrInvalidEditPlan
	}
	s.editPlans[key] = cloneEditPlan(normalized)
	s.editPlanRequests[requestKey] = normalized.ID
	return cloneEditPlan(normalized), nil
}

func (s *MemoryStore) EditPlan(_ context.Context, owner, projectID, id string) (EditPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.editPlans[editPlanKey(owner, projectID, id)]
	if !ok {
		return EditPlan{}, ErrEditPlanNotFound
	}
	return cloneEditPlan(plan), nil
}

func (s *MemoryStore) EditPlanByRequest(_ context.Context, owner, projectID, requestID string) (EditPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.editPlanRequests[editPlanRequestKey(owner, projectID, requestID)]
	if id == "" {
		return EditPlan{}, ErrEditPlanNotFound
	}
	return cloneEditPlan(s.editPlans[editPlanKey(owner, projectID, id)]), nil
}

func (s *MemoryStore) UpdateEditPlanState(_ context.Context, owner, projectID, id string, state EditPlanState, message string, now time.Time) (EditPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := editPlanKey(owner, projectID, id)
	plan, ok := s.editPlans[key]
	if !ok {
		return EditPlan{}, ErrEditPlanNotFound
	}
	if !validEditPlanStateTransition(plan.State, state) {
		return EditPlan{}, ErrInvalidEditPlan
	}
	plan.State, plan.Error, plan.UpdatedAt = state, strings.TrimSpace(message), now.UTC()
	s.editPlans[key] = cloneEditPlan(plan)
	return cloneEditPlan(plan), nil
}

func enabledEditPlan(plan EditPlan, enabledIDs []string) EditPlan {
	if enabledIDs == nil {
		return plan
	}
	enabled := make(map[string]struct{}, len(enabledIDs))
	for _, id := range enabledIDs {
		enabled[strings.TrimSpace(id)] = struct{}{}
	}
	for index := range plan.Operations {
		_, plan.Operations[index].Enabled = enabled[plan.Operations[index].ID]
	}
	return plan
}

func (s *MemoryStore) ApplyStoredEditPlan(_ context.Context, owner, projectID, id string, enabledIDs []string, now time.Time) (EditPlan, Timeline, EditPlanApplySummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := editPlanKey(owner, projectID, id)
	plan, ok := s.editPlans[key]
	if !ok {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrEditPlanNotFound
	}
	if canonicalASRTimeline(plan.TimelineID) != DefaultTimelineID {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrInvalidEditPlan
	}
	projectKey := aicutProjectKey(owner, projectID)
	item, ok := s.projects[projectKey]
	if !ok || item.project.DeletedAt != nil {
		return EditPlan{}, Timeline{}, EditPlanApplySummary{}, ErrProjectNotFound
	}
	if plan.State == EditPlanApplied {
		return cloneEditPlan(plan), cloneTimeline(item.timeline), s.editPlanSummaries[key], nil
	}
	if plan.State != EditPlanReady {
		return cloneEditPlan(plan), cloneTimeline(item.timeline), EditPlanApplySummary{}, ErrInvalidEditPlan
	}
	if item.timeline.Revision != plan.BaseRevision {
		plan.State, plan.Error, plan.UpdatedAt = EditPlanStale, "草案已过期，请基于最新工程重新生成", now.UTC()
		s.editPlans[key] = cloneEditPlan(plan)
		return cloneEditPlan(plan), cloneTimeline(item.timeline), EditPlanApplySummary{}, ErrRevisionConflict
	}
	selected := enabledEditPlan(cloneEditPlan(plan), enabledIDs)
	if err := refreshMemoryEditPlanAssets(s, &selected); err != nil {
		return plan, cloneTimeline(item.timeline), EditPlanApplySummary{}, err
	}
	prepared, summary, err := ApplyEditPlan(item.timeline, selected)
	if err != nil {
		return cloneEditPlan(plan), cloneTimeline(item.timeline), EditPlanApplySummary{}, err
	}
	prepared.Revision = item.timeline.Revision + 1
	item.timeline = cloneTimeline(prepared)
	if item.historyCursor < 0 || item.historyCursor >= len(item.history) {
		item.history = []Timeline{cloneTimeline(Timeline{Tracks: []Track{}})}
		item.historyCursor = 0
	}
	item.history = append(item.history[:item.historyCursor+1], cloneTimeline(prepared))
	item.historyCursor++
	if summary.AspectRatio != "" {
		item.project.AspectRatio = summary.AspectRatio
	}
	item.project.UpdatedAt = now.UTC()
	s.projects[projectKey] = item
	plan.Operations = selected.Operations
	plan.Summary.EstimatedDurationMS = summary.EstimatedDurationMS
	plan.State, plan.Error, plan.AppliedRevision, plan.UpdatedAt = EditPlanApplied, "", prepared.Revision, now.UTC()
	s.editPlans[key] = cloneEditPlan(plan)
	s.editPlanSummaries[key] = summary
	return cloneEditPlan(plan), cloneTimeline(prepared), summary, nil
}

func editPlanPlacedAssets(plan EditPlan) map[string]bool {
	ids := map[string]bool{}
	for _, op := range plan.Operations {
		if op.Enabled && (op.Type == EditPlaceAsset || op.Type == EditAddBGM) {
			ids[op.AssetID] = true
		}
	}
	return ids
}

// Caller holds the store mutex; metadata cannot change between validation and commit.
func refreshMemoryEditPlanAssets(s *MemoryStore, plan *EditPlan) error {
	for id := range editPlanPlacedAssets(*plan) {
		a, ok := s.assets[aicutAssetKey(plan.Owner, plan.ProjectID, id)]
		if !ok {
			return ErrAssetNotFound
		}
		replaceEditPlanAsset(plan, a)
	}
	return nil
}
func replaceEditPlanAsset(plan *EditPlan, a Asset) {
	value := EditPlanAsset{ID: a.ID, Name: a.Name, Kind: a.Kind, DurationMS: a.DurationMS}
	for i := range plan.Request.SelectedAssets {
		if plan.Request.SelectedAssets[i].ID == a.ID {
			plan.Request.SelectedAssets[i] = value
			return
		}
	}
	plan.Request.SelectedAssets = append(plan.Request.SelectedAssets, value)
}
