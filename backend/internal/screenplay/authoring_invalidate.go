package screenplay

import (
	"context"
	"errors"
)

// InvalidateAuthoringContent preserves edited artifacts and pauses downstream work until the user resumes production.
func InvalidateAuthoringContent(ctx context.Context, s Store, o, p, kind string) error {
	if repo, ok := s.(ProductionRepository); ok {
		if err := repo.SetProductionState(ctx, o, p, "paused", ""); err != nil {
			return err
		}
	}
	from := 2
	switch kind {
	case "source":
		from = 1
	case "asset":
		from = 4
	case "storyboard":
		from = 6
	}
	_, e := s.MutateBatch(ctx, o, p, func(b *Batch) error {
		keys := []string{}
		for i := from; i < len(StepDefinitions); i++ {
			keys = append(keys, StepDefinitions[i].Key)
		}
		MarkStale(b, keys)
		for _, key := range keys {
			step := b.Step(key)
			if step == nil {
				continue
			}
			step.Error = ""
			step.StartedAt = nil
			step.Artifacts = nil
			for _, ep := range step.Episodes {
				ep.Status = StepStale
				ep.Reason = ""
				ep.UpdatedAt = nowUTC()
			}
		}
		b.Status = BatchPaused
		b.LastError = ""
		b.CurrentEpisode = 0
		return nil
	})
	if errors.Is(e, ErrNotFound) {
		return nil
	}
	return e
}

// PauseArchivedProduction cancels local scheduling, without submitting or cancelling an upstream task.
func PauseArchivedProduction(ctx context.Context, s Store, o, p string) error {
	if repo, ok := s.(ProductionRepository); ok {
		job, e := repo.GetProductionJob(ctx, o, p)
		if e != nil && !errors.Is(e, ErrNotFound) {
			return e
		}
		if e == nil && job.Status != "completed" {
			if e = repo.SetProductionState(ctx, o, p, "paused", ""); e != nil {
				return e
			}
		}
	}
	_, e := s.MutateBatch(ctx, o, p, func(b *Batch) error {
		if b.Status != BatchCompleted {
			b.Status = BatchPaused
			b.CurrentEpisode = 0
		}
		return nil
	})
	if errors.Is(e, ErrNotFound) {
		return nil
	}
	return e
}
