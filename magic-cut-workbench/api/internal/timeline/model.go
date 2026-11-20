package timeline

import (
	"errors"
	"fmt"
)

type Timeline struct {
	Revision   uint64  `json:"revision"`
	DurationMs int64   `json:"durationMs"`
	Tracks     []Track `json:"tracks"`
}

type Track struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Clips []Clip `json:"clips"`
}

type Clip struct {
	ID            string  `json:"id"`
	MaterialID    string  `json:"materialId"`
	StartMs       int64   `json:"startMs"`
	DurationMs    int64   `json:"durationMs"`
	SourceStartMs int64   `json:"sourceStartMs"`
	Speed         float64 `json:"speed"`
}

func Validate(value Timeline) error {
	if value.DurationMs < 0 {
		return errors.New("timeline duration must be nonnegative")
	}
	trackIDs := map[string]bool{}
	clipIDs := map[string]bool{}
	for _, track := range value.Tracks {
		if track.ID == "" || trackIDs[track.ID] {
			return errors.New("track id must be unique")
		}
		trackIDs[track.ID] = true
		for _, clip := range track.Clips {
			if clip.ID == "" || clipIDs[clip.ID] {
				return errors.New("clip id must be unique")
			}
			if clip.MaterialID == "" {
				return errors.New("clip material is required")
			}
			if clip.StartMs < 0 || clip.SourceStartMs < 0 {
				return errors.New("clip bounds must be nonnegative")
			}
			if clip.DurationMs <= 0 || clip.Speed <= 0 {
				return errors.New("clip duration and speed must be positive")
			}
			clipIDs[clip.ID] = true
		}
	}
	return nil
}

type RevisionConflictError struct{ CurrentRevision uint64 }

func (e *RevisionConflictError) Error() string {
	return fmt.Sprintf("timeline revision conflict: current=%d", e.CurrentRevision)
}
