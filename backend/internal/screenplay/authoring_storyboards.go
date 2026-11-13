package screenplay

import (
	"context"
	"fmt"
	"math"
)

type StoryboardBatchEdit struct {
	ExpectedRevision int64            `json:"expected_revision"`
	Operation        string           `json:"operation"`
	ShotIDs          []string         `json:"shot_ids"`
	Shots            []StoryboardShot `json:"shots"`
	Order            []string         `json:"order"`
	Duration         float64          `json:"duration"`
}

func EditStoryboardBatch(ctx context.Context, s Store, o, p string, episode int, engine StoryboardEngine, in StoryboardBatchEdit) (Storyboard, error) {
	if episode <= 0 {
		return Storyboard{}, ErrInvalidArgument
	}
	if _, _, e := resolveStoryboardEngine(engine); e != nil {
		return Storyboard{}, e
	}
	if _, e := s.GetProject(ctx, o, p); e != nil {
		return Storyboard{}, e
	}
	b, e := s.GetStoryboard(ctx, o, p, episode, engine)
	if e != nil {
		return b, e
	}
	if b.Revision != in.ExpectedRevision {
		return b, ErrRevisionConflict
	}
	selected := map[string]bool{}
	for _, id := range in.ShotIDs {
		if selected[id] || storyboardShotIndex(b.Shots, id) < 0 {
			return b, ErrInvalidArgument
		}
		selected[id] = true
	}
	switch in.Operation {
	case "delete", "duplicate", "duration":
		if len(selected) == 0 {
			return b, ErrInvalidArgument
		}
		out := []StoryboardShot{}
		for _, sh := range b.Shots {
			if selected[sh.ID] {
				switch in.Operation {
				case "delete":
					continue
				case "duration":
					sh.Duration = in.Duration
				case "duplicate":
					out = append(out, sh)
					sh.ID = newScreenplayID("sb_shot")
				}
			}
			out = append(out, sh)
		}
		b.Shots = out
	case "reorder":
		if len(in.Order) != len(b.Shots) {
			return b, ErrInvalidArgument
		}
		out := []StoryboardShot{}
		seen := map[string]bool{}
		for _, id := range in.Order {
			i := storyboardShotIndex(b.Shots, id)
			if i < 0 || seen[id] {
				return b, ErrInvalidArgument
			}
			seen[id] = true
			out = append(out, b.Shots[i])
		}
		b.Shots = out
	case "update":
		if len(in.Shots) == 0 {
			return b, ErrInvalidArgument
		}
		seen := map[string]bool{}
		for _, sh := range in.Shots {
			i := storyboardShotIndex(b.Shots, sh.ID)
			if i < 0 || seen[sh.ID] || sh.OwnerUsername != "" && sh.OwnerUsername != o || sh.ProjectID != "" && sh.ProjectID != p || sh.StoryboardID != "" && sh.StoryboardID != b.ID {
				return b, ErrInvalidArgument
			}
			seen[sh.ID] = true
			b.Shots[i] = sh
		}
	default:
		return b, ErrInvalidArgument
	}
	for i := range b.Shots {
		sh := &b.Shots[i]
		if math.IsNaN(sh.Duration) || math.IsInf(sh.Duration, 0) || sh.Duration > 600 {
			return b, ErrInvalidArgument
		}
		if e = validateStoryboardShotInput(*sh); e != nil {
			return b, e
		}
		sh.Seq = i + 1
	}
	if len(b.Shots) == 0 {
		return b, fmt.Errorf("%w: 不可删除全部镜头", ErrInvalidArgument)
	}
	cas, ok := s.(interface {
		SaveStoryboardRevision(context.Context, Storyboard, int64) (Storyboard, error)
	})
	if !ok {
		return b, ErrInvalidState
	}
	return cas.SaveStoryboardRevision(ctx, b, in.ExpectedRevision)
}
func PublishChapterScripts(ctx context.Context, s Store, o, p, content string) (Document, error) {
	r, e := (&ScreenplayService{Store: s}).persist(ctx, o, p, "chapters", content, 1)
	return r.Document, e
}
