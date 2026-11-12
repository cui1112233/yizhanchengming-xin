package aicut

import (
	"context"
	"time"
)

// ScopedStore routes timeline commands to exactly one selected timeline. Named
// timelines reject asynchronous jobs until job persistence supports timeline IDs.
func (s *TimelineCollectionStore) ScopedStore(ctx context.Context, o, p, id string) (Store, error) {
	if _, e := s.Timeline(ctx, o, p, id); e != nil {
		return nil, e
	}
	if id == DefaultTimelineID {
		return s.base, nil
	}
	return &timelineScopedStore{Store: s.base, collection: s, owner: o, project: p, id: id}, nil
}

type timelineScopedStore struct {
	Store
	collection         *TimelineCollectionStore
	owner, project, id string
}

func (s *timelineScopedStore) scope(o, p string) error {
	if o != s.owner || p != s.project {
		return ErrProjectNotFound
	}
	return nil
}
func (s *timelineScopedStore) Timeline(ctx context.Context, o, p string) (Timeline, error) {
	if e := s.scope(o, p); e != nil {
		return Timeline{}, e
	}
	return s.collection.Timeline(ctx, o, p, s.id)
}
func (s *timelineScopedStore) SaveTimeline(ctx context.Context, o, p string, v Timeline, r int64, n time.Time) (Timeline, error) {
	if e := s.scope(o, p); e != nil {
		return Timeline{}, e
	}
	return s.collection.Save(ctx, o, p, s.id, v, r, n)
}
func (s *timelineScopedStore) TimelineHistory(ctx context.Context, o, p string) (TimelineHistoryState, error) {
	if e := s.scope(o, p); e != nil {
		return TimelineHistoryState{}, e
	}
	return s.collection.History(ctx, o, p, s.id)
}
func (s *timelineScopedStore) UndoTimeline(ctx context.Context, o, p string, r int64, n time.Time) (Timeline, TimelineHistoryState, error) {
	if e := s.scope(o, p); e != nil {
		return Timeline{}, TimelineHistoryState{}, e
	}
	return s.collection.Undo(ctx, o, p, s.id, r, n)
}
func (s *timelineScopedStore) RedoTimeline(ctx context.Context, o, p string, r int64, n time.Time) (Timeline, TimelineHistoryState, error) {
	if e := s.scope(o, p); e != nil {
		return Timeline{}, TimelineHistoryState{}, e
	}
	return s.collection.Redo(ctx, o, p, s.id, r, n)
}
func (s *timelineScopedStore) CreateASRJob(context.Context, string, ASRJob) error {
	return ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) CreateExportJob(context.Context, string, ExportJob) error {
	return ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ASRJob(context.Context, string, string, string) (ASRJob, error) {
	return ASRJob{}, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ListASRJobs(context.Context, string, string) ([]ASRJob, error) {
	return nil, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ClaimNextASRJob(context.Context, time.Time) (ASRJob, error) {
	return ASRJob{}, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) UpdateASRJob(context.Context, ASRJob) error {
	return ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) RequestASRCancel(context.Context, string, string, string, time.Time) (ASRJob, error) {
	return ASRJob{}, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ExportJob(context.Context, string, string, string) (ExportJob, error) {
	return ExportJob{}, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ListExportJobs(context.Context, string, string) ([]ExportJob, error) {
	return nil, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) ClaimNextExportJob(context.Context, time.Time) (ExportJob, error) {
	return ExportJob{}, ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) UpdateExportJob(context.Context, ExportJob) error {
	return ErrTimelineJobUnsupported
}
func (s *timelineScopedStore) RequestExportCancel(context.Context, string, string, string, time.Time) (ExportJob, error) {
	return ExportJob{}, ErrTimelineJobUnsupported
}
