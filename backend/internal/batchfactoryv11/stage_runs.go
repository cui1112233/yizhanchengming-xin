package batchfactoryv11

import (
	"context"
	"sort"
	"strings"
	"time"
)

type BookStage string

const (
	BookStageAssets   BookStage = "assets"
	BookStageDirector BookStage = "director"
	BookStageOpening  BookStage = "opening"
	BookStageVisual   BookStage = "visual"
	BookStageComic    BookStage = "comic"
	BookStageImage    BookStage = "image"
	BookStageVideo    BookStage = "video"
)

type StageMode string

const (
	StageModeMissing   StageMode = "missing"
	StageModeForce     StageMode = "force"
	StageModeBootstrap StageMode = "bootstrap"
)

type BookStageRun struct {
	ID            string          `json:"id"`
	Owner         string          `json:"-"`
	BatchID       string          `json:"batchId"`
	BookID        string          `json:"bookId"`
	Stage         BookStage       `json:"stage"`
	Status        ProductionState `json:"status"`
	Attempt       int             `json:"attempt"`
	RequestID     string          `json:"requestId,omitempty"`
	InputRevision string          `json:"inputRevision,omitempty"`
	ErrorMessage  string          `json:"errorMessage,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

type BookStageRunRepository interface {
	CreateBookStageRun(context.Context, BookStageRun) (BookStageRun, error)
	UpdateBookStageRun(context.Context, string, string, BookStageRun) (BookStageRun, error)
	ListBookStageRuns(context.Context, string, string, string) ([]BookStageRun, error)
}

func validBookStage(value BookStage) bool {
	switch value {
	case BookStageAssets, BookStageDirector, BookStageOpening, BookStageVisual, BookStageComic, BookStageImage, BookStageVideo:
		return true
	default:
		return false
	}
}

func validStageMode(stage BookStage, mode StageMode) bool {
	switch mode {
	case StageModeMissing, StageModeForce:
		return true
	case StageModeBootstrap:
		return stage == BookStageComic
	default:
		return false
	}
}

func normalizeBookStageRun(value BookStageRun) (BookStageRun, error) {
	value.BatchID = strings.TrimSpace(value.BatchID)
	value.BookID = strings.TrimSpace(value.BookID)
	value.Owner = strings.TrimSpace(value.Owner)
	value.RequestID = strings.TrimSpace(value.RequestID)
	value.InputRevision = strings.TrimSpace(value.InputRevision)
	value.ErrorMessage = productionError(errText(value.ErrorMessage))
	if value.Owner == "" || value.BatchID == "" || value.BookID == "" || !validBookStage(value.Stage) || value.Attempt < 1 {
		return BookStageRun{}, ErrInvalid
	}
	value.Status = normalizeProductionState(value.Status)
	return value, nil
}

func errText(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return stageRunError(value)
}

type stageRunError string

func (e stageRunError) Error() string { return string(e) }

func LatestFailedBookStageRun(values []BookStageRun) *BookStageRun {
	// A failed attempt is no longer actionable once a later attempt for that
	// same stage has succeeded. Retry must select an unresolved failure, not a
	// historical one; otherwise a book can keep retrying an already recovered
	// asset stage forever.
	latestByStage := make(map[BookStage]BookStageRun, len(values))
	for _, value := range values {
		current, exists := latestByStage[value.Stage]
		if !exists || value.UpdatedAt.After(current.UpdatedAt) || (value.UpdatedAt.Equal(current.UpdatedAt) && value.ID > current.ID) {
			latestByStage[value.Stage] = value
		}
	}
	failed := make([]BookStageRun, 0, len(latestByStage))
	for _, value := range latestByStage {
		// A missing-mode director run that arrives after a revision already
		// exists is a stale duplicate, not a user-actionable failure. Older
		// records may still contain it, so keep it out of the retry state.
		if value.Status == ProductionFailed && !(value.Stage == BookStageDirector && strings.Contains(value.ErrorMessage, "导演分镜已存在")) {
			failed = append(failed, value)
		}
	}
	if len(failed) == 0 {
		return nil
	}
	sort.SliceStable(failed, func(i, j int) bool {
		if failed[i].UpdatedAt.Equal(failed[j].UpdatedAt) {
			return failed[i].ID > failed[j].ID
		}
		return failed[i].UpdatedAt.After(failed[j].UpdatedAt)
	})
	value := failed[0]
	return &value
}
