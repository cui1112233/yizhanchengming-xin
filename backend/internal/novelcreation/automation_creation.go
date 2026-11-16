package novelcreation

import (
	"context"
	"strings"
)

// Creation describes the immutable original request, independently of subsequent
// edits and execution progress. It lives in the plan's existing input JSON.
type AutomationCreation struct {
	IdempotencyKey string `json:"idempotencyKey"`
	Fingerprint    string `json:"fingerprint"`
}

func automationCreation(owner, project, kind, key string, input any) (string, *AutomationCreation, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil, nil
	}
	if len(key) > 256 {
		return "", nil, inputError("创建编号过长")
	}
	return kind + "-intent-" + fingerprint([]string{strings.TrimSpace(owner), strings.TrimSpace(project), kind, key})[:40], &AutomationCreation{key, fingerprint(input)}, nil
}

func creationMatches(existing, requested *AutomationCreation) error {
	if existing == nil || requested == nil || existing.IdempotencyKey != requested.IdempotencyKey || existing.Fingerprint != requested.Fingerprint {
		return &DomainError{Code: ErrIdempotencyConflict, Message: "创建编号已用于不同的原始请求，请先核对原任务"}
	}
	return nil
}

func (s *Service) findScheduleCreation(ctx context.Context, owner, project string, in ScheduleInput) (Schedule, bool, error) {
	if s == nil || s.store == nil {
		return Schedule{}, false, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	id, creation, err := automationCreation(owner, project, "schedule", in.IdempotencyKey, in)
	if err != nil || creation == nil {
		return Schedule{}, false, err
	}
	p, err := s.store.GetSchedule(ctx, strings.TrimSpace(owner), id)
	if IsCode(err, ErrProjectNotFound) {
		return Schedule{}, false, nil
	}
	if err != nil {
		return p, false, err
	}
	if p.ProjectID != strings.TrimSpace(project) {
		return Schedule{}, false, &DomainError{Code: ErrProjectNotFound, Message: "定时任务不存在"}
	}
	return p, true, creationMatches(p.Creation, creation)
}

func (s *Service) findBatchCreation(ctx context.Context, owner, project string, in BatchInput) (Batch, bool, error) {
	if s == nil || s.store == nil {
		return Batch{}, false, &DomainError{Code: ErrPersistenceFailed, Message: "novel store is unavailable"}
	}
	id, creation, err := automationCreation(owner, project, "batch", in.IdempotencyKey, in)
	if err != nil || creation == nil {
		return Batch{}, false, err
	}
	p, err := s.store.GetBatch(ctx, strings.TrimSpace(owner), id)
	if IsCode(err, ErrProjectNotFound) {
		return Batch{}, false, nil
	}
	if err != nil {
		return p, false, err
	}
	if p.ProjectID != strings.TrimSpace(project) {
		return Batch{}, false, &DomainError{Code: ErrProjectNotFound, Message: "批量任务不存在"}
	}
	return p, true, creationMatches(p.Creation, creation)
}
