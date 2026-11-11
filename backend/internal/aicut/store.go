package aicut

import (
	"context"
	"time"
)

type Store interface {
	ListProjects(context.Context, string) ([]Project, error)
	CreateProject(context.Context, string, Project) error
	Project(context.Context, string, string) (Project, error)
	RenameProject(context.Context, string, string, string, time.Time) (Project, error)
	DeleteProject(context.Context, string, string, time.Time) error
	Timeline(context.Context, string, string) (Timeline, error)
	SaveTimeline(context.Context, string, string, Timeline, int64, time.Time) (Timeline, error)
	TimelineHistory(context.Context, string, string) (TimelineHistoryState, error)
	UndoTimeline(context.Context, string, string, int64, time.Time) (Timeline, TimelineHistoryState, error)
	RedoTimeline(context.Context, string, string, int64, time.Time) (Timeline, TimelineHistoryState, error)
	ListAssets(context.Context, string, string) ([]Asset, error)
	CreateAsset(context.Context, string, Asset) error
	Asset(context.Context, string, string, string) (Asset, error)
	CreateASRJob(context.Context, string, ASRJob) error
	ASRJob(context.Context, string, string, string) (ASRJob, error)
	ListASRJobs(context.Context, string, string) ([]ASRJob, error)
	ClaimNextASRJob(context.Context, time.Time) (ASRJob, error)
	UpdateASRJob(context.Context, ASRJob) error
	RequestASRCancel(context.Context, string, string, string, time.Time) (ASRJob, error)
	CreateExportJob(context.Context, string, ExportJob) error
	ExportJob(context.Context, string, string, string) (ExportJob, error)
	ListExportJobs(context.Context, string, string) ([]ExportJob, error)
	ClaimNextExportJob(context.Context, time.Time) (ExportJob, error)
	UpdateExportJob(context.Context, ExportJob) error
	RequestExportCancel(context.Context, string, string, string, time.Time) (ExportJob, error)
}
