package screenplay

import (
	"context"
	"errors"
)

// InvalidateMediaContent pauses delivery affected by an explicit media edit.
// Previously submitted provider tasks remain in the durable request history.
func InvalidateMediaContent(ctx context.Context, store Store, owner, id, kind string) error {
	keys := []string{"8_finalize"}
	if kind == "references" {
		keys = []string{"7_video", "8_finalize"}
	}
	if repo, ok := store.(ProductionRepository); ok {
		if e := repo.SetProductionState(ctx, owner, id, "paused", ""); e != nil {
			return e
		}
	}
	_, e := store.MutateBatch(ctx, owner, id, func(b *Batch) error {
		MarkStale(b, keys)
		for _, key := range keys {
			st := b.Step(key)
			if st == nil {
				continue
			}
			st.Artifacts = nil
			st.Error = ""
			for _, ep := range st.Episodes {
				ep.Status = StepStale
				ep.Artifact = ""
				ep.Reason = ""
			}
		}
		b.Status = BatchPaused
		b.LastError = ""
		return nil
	})
	if errors.Is(e, ErrNotFound) {
		return nil
	}
	return e
}
