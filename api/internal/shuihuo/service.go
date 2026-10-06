package shuihuo

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	store Store
	smart SmartSegmenter
}

func NewService(store Store, smart SmartSegmenter) *Service {
	return &Service{store: store, smart: smart}
}

func (s *Service) CreateProject(ctx context.Context, actor Actor, input CreateProjectInput) (ReadModel, error) {
	if s == nil || s.store == nil {
		return ReadModel{}, ErrInvalid
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return ReadModel{}, ErrInvalid
	}
	if !actor.BypassOwnership && actor.UserID <= 0 {
		return ReadModel{}, ErrForbidden
	}
	mode := strings.TrimSpace(input.ProductionMode)
	if mode == "" {
		mode = "commentary"
	}
	project, err := s.store.CreateProject(ctx, Project{OwnerUserID: actor.UserID, TeamID: actor.TeamID, Name: name, ProductionMode: mode, SegmentationStatus: SegmentationDraft})
	if err != nil {
		return ReadModel{}, err
	}
	return ReadModel{Project: project, Segments: []Segment{}}, nil
}

func (s *Service) ListProjects(ctx context.Context, actor Actor) ([]Project, error) {
	if s == nil || s.store == nil {
		return nil, ErrInvalid
	}
	return s.store.ListProjects(ctx, actor)
}

func (s *Service) GetProject(ctx context.Context, actor Actor, id int64) (ReadModel, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return ReadModel{}, err
	}
	segments, err := s.store.ListSegments(ctx, id)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return ReadModel{}, err
	}
	if segments == nil {
		segments = []Segment{}
	}
	return ReadModel{Project: project, Segments: segments}, nil
}

func (s *Service) ReplaceSource(ctx context.Context, actor Actor, id int64, source string) (ReadModel, error) {
	if _, err := s.projectForActor(ctx, actor, id); err != nil {
		return ReadModel{}, err
	}
	source = strings.TrimSpace(source)
	if source == "" {
		return ReadModel{}, ErrInvalid
	}
	project, err := s.store.SaveSourceAndReset(ctx, id, source)
	if err != nil {
		return ReadModel{}, err
	}
	return ReadModel{Project: project, Segments: []Segment{}}, nil
}

func (s *Service) DeleteProject(ctx context.Context, actor Actor, id int64) error {
	if _, err := s.projectForActor(ctx, actor, id); err != nil {
		return err
	}
	return s.store.DeleteProject(ctx, id)
}

func (s *Service) ParagraphSegmentation(ctx context.Context, actor Actor, id int64, input SegmentationInput) ([]Candidate, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	rows := paragraphCandidates(sourceOr(input.Text, project.SourceText))
	if len(rows) == 0 {
		return nil, ErrInvalid
	}
	return rows, nil
}

func (s *Service) FixedSegmentation(ctx context.Context, actor Actor, id int64, input FixedSegmentationInput) ([]Candidate, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if input.LinesPerSegment <= 0 {
		return nil, ErrInvalid
	}
	rows := fixedCandidates(sourceOr(input.Text, project.SourceText), input.LinesPerSegment)
	if len(rows) == 0 {
		return nil, ErrInvalid
	}
	return rows, nil
}

func (s *Service) ImportSegmentation(ctx context.Context, actor Actor, id int64, input SegmentationInput) ([]Candidate, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	rows := importedCandidates(sourceOr(input.Text, project.SourceText))
	if len(rows) == 0 {
		return nil, ErrInvalid
	}
	return rows, nil
}

func (s *Service) SmartSegmentation(ctx context.Context, actor Actor, id int64, input SegmentationInput) ([]Candidate, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if s.smart == nil {
		return nil, ErrSmartUnavailable
	}
	text := strings.TrimSpace(sourceOr(input.Text, project.SourceText))
	if text == "" {
		return nil, ErrInvalid
	}
	rows, err := s.smart.Segment(ctx, text)
	if err != nil {
		return nil, err
	}
	return normalizeCandidates(rows)
}

func (s *Service) ConfirmSegmentation(ctx context.Context, actor Actor, id int64, candidates []Candidate) (ReadModel, error) {
	project, err := s.projectForActor(ctx, actor, id)
	if err != nil {
		return ReadModel{}, err
	}
	rows, err := normalizeCandidates(candidates)
	if err != nil {
		return ReadModel{}, err
	}
	segments, err := s.store.ReplaceSegments(ctx, id, rows)
	if err != nil {
		return ReadModel{}, err
	}
	project.SegmentationStatus = SegmentationConfirmed
	return ReadModel{Project: project, Segments: segments}, nil
}

func (s *Service) CreateSegment(ctx context.Context, actor Actor, projectID int64, input SegmentInput) (Segment, error) {
	if _, err := s.projectForActor(ctx, actor, projectID); err != nil {
		return Segment{}, err
	}
	input = normalizeSegmentInput(input)
	if input.SourceText == "" {
		return Segment{}, ErrInvalid
	}
	return s.store.CreateSegment(ctx, projectID, input)
}

func (s *Service) UpdateSegment(ctx context.Context, actor Actor, segmentID int64, input SegmentInput) (Segment, error) {
	seg, err := s.store.GetSegment(ctx, segmentID)
	if err != nil {
		return Segment{}, err
	}
	if _, err := s.projectForActor(ctx, actor, seg.ProjectID); err != nil {
		return Segment{}, err
	}
	input = normalizeSegmentInput(input)
	if input.SourceText == "" {
		return Segment{}, ErrInvalid
	}
	return s.store.UpdateSegment(ctx, segmentID, input)
}

func (s *Service) DeleteSegment(ctx context.Context, actor Actor, segmentID int64) error {
	seg, err := s.store.GetSegment(ctx, segmentID)
	if err != nil {
		return err
	}
	if _, err := s.projectForActor(ctx, actor, seg.ProjectID); err != nil {
		return err
	}
	return s.store.DeleteSegment(ctx, segmentID)
}

func (s *Service) ReorderSegments(ctx context.Context, actor Actor, projectID int64, ids []int64) (ReadModel, error) {
	project, err := s.projectForActor(ctx, actor, projectID)
	if err != nil {
		return ReadModel{}, err
	}
	if len(ids) == 0 {
		return ReadModel{}, ErrInvalid
	}
	rows, err := s.store.ReorderSegments(ctx, projectID, ids)
	if err != nil {
		return ReadModel{}, err
	}
	return ReadModel{Project: project, Segments: rows}, nil
}

func (s *Service) projectForActor(ctx context.Context, actor Actor, id int64) (Project, error) {
	if s == nil || s.store == nil || id <= 0 {
		return Project{}, ErrInvalid
	}
	p, err := s.store.GetProject(ctx, id)
	if err != nil {
		return Project{}, err
	}
	if !CanAccess(actor, p) {
		return Project{}, ErrForbidden
	}
	return p, nil
}

func sourceOr(candidate, fallback string) string {
	if strings.TrimSpace(candidate) != "" {
		return candidate
	}
	return fallback
}

func normalizeSegmentInput(input SegmentInput) SegmentInput {
	input.SourceText = strings.TrimSpace(input.SourceText)
	input.SubtitleText = strings.TrimSpace(input.SubtitleText)
	if input.SubtitleText == "" {
		input.SubtitleText = input.SourceText
	}
	input.Speaker = normalizeSpeaker(input.Speaker)
	return input
}
