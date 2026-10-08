package workshop

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

var ErrNotFound = errors.New("workshop: not found")

type Store interface {
	Load(context.Context, int64) (json.RawMessage, error)
	Save(context.Context, int64, json.RawMessage) (json.RawMessage, error)
}

type IntakeReader interface {
	GetIntake(context.Context, int64) (intake.Intake, error)
	ListBooks(context.Context, int64) ([]intake.Book, error)
}

type PromptReader interface {
	ListPrompts(context.Context) ([]generation.Prompt, error)
}

type Snapshot struct {
	Intake   intake.Intake       `json:"intake"`
	Books    []intake.Book       `json:"books"`
	Settings json.RawMessage     `json:"settings"`
	Prompts  []generation.Prompt `json:"prompts"`
}

type Service struct {
	store   Store
	intakes IntakeReader
	prompts PromptReader
}

func NewService(store Store, intakes IntakeReader, prompts PromptReader) *Service {
	return &Service{store: store, intakes: intakes, prompts: prompts}
}

func normalize(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(`{}`), nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, fmt.Errorf("workshop settings must be a JSON object")
	}
	return raw, nil
}

func (s *Service) Snapshot(ctx context.Context, intakeID int64) (Snapshot, error) {
	if s == nil || s.store == nil || s.intakes == nil || s.prompts == nil {
		return Snapshot{}, errors.New("workshop unavailable")
	}
	item, err := s.intakes.GetIntake(ctx, intakeID)
	if err != nil {
		return Snapshot{}, err
	}
	books, err := s.intakes.ListBooks(ctx, intakeID)
	if err != nil {
		return Snapshot{}, err
	}
	settings, err := s.store.Load(ctx, intakeID)
	if err != nil {
		return Snapshot{}, err
	}
	settings, err = normalize(settings)
	if err != nil {
		return Snapshot{}, err
	}
	prompts, err := s.prompts.ListPrompts(ctx)
	if err != nil {
		return Snapshot{}, err
	}
	for i := range prompts {
		prompts[i].Content = ""
	}
	return Snapshot{Intake: item, Books: books, Settings: settings, Prompts: prompts}, nil
}

func (s *Service) Save(ctx context.Context, intakeID int64, settings json.RawMessage) (json.RawMessage, error) {
	if s == nil || s.store == nil || s.intakes == nil {
		return nil, errors.New("workshop unavailable")
	}
	if _, err := s.intakes.GetIntake(ctx, intakeID); err != nil {
		return nil, err
	}
	settings, err := normalize(settings)
	if err != nil {
		return nil, err
	}
	return s.store.Save(ctx, intakeID, settings)
}
