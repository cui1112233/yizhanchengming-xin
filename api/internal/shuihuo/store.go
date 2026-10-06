package shuihuo

import "context"

type Store interface {
	CreateProject(context.Context, Project) (Project, error)
	ListProjects(context.Context, Actor) ([]Project, error)
	GetProject(context.Context, int64) (Project, error)
	SaveSourceAndReset(context.Context, int64, string) (Project, error)
	DeleteProject(context.Context, int64) error
	ReplaceSegments(context.Context, int64, []Candidate) ([]Segment, error)
	ListSegments(context.Context, int64) ([]Segment, error)
	GetSegment(context.Context, int64) (Segment, error)
	CreateSegment(context.Context, int64, SegmentInput) (Segment, error)
	UpdateSegment(context.Context, int64, SegmentInput) (Segment, error)
	DeleteSegment(context.Context, int64) error
	ReorderSegments(context.Context, int64, []int64) ([]Segment, error)
}

type SmartSegmenter interface {
	Segment(context.Context, string) ([]Candidate, error)
}
