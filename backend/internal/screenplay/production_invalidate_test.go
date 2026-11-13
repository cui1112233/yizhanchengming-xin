package screenplay

import (
	"context"
	"testing"
)

type combinedProductionFixture struct {
	*MemoryStore
	*MemoryProductionRepository
}

func TestProductionMediaEditInvalidatesOnlyAffectedStages(t *testing.T) {
	for _, kind := range []string{"references", "edit", "voice"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s := combinedProductionFixture{NewMemoryStore(), NewMemoryProductionRepository()}
			_, _ = s.CreateProject(ctx, Project{ID: "p", OwnerUsername: "a"})
			b := NewBatch("b", "p", "a", "", PermissionBypass)
			for _, st := range b.Steps {
				st.Status = StepDone
			}
			b.Status = BatchCompleted
			_ = s.CreateBatch(ctx, b)
			_ = s.SetProductionState(ctx, "a", "p", "completed", "")
			if e := InvalidateMediaContent(ctx, s, "a", "p", kind); e != nil {
				t.Fatal(e)
			}
			b, _ = s.GetBatch(ctx, "a", "p")
			if b.Status != BatchPaused || b.Step("5_storyboard").Status != StepDone || b.Step("8_finalize").Status != StepStale {
				t.Fatal(b)
			}
			if (b.Step("7_video").Status == StepStale) != (kind == "references") {
				t.Fatal("wrong video invalidation")
			}
			j, _ := s.GetProductionJob(ctx, "a", "p")
			if j.Status != "paused" {
				t.Fatal(j)
			}
		})
	}
}
