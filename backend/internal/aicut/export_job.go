package aicut

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrExportJobNotFound = errors.New("ai-cut export job not found")
	ErrExportQueueEmpty  = errors.New("ai-cut export queue empty")
	ErrInvalidExportJob  = errors.New("invalid ai-cut export job")
)

type ExportJobState string

const (
	ExportQueued          ExportJobState = "queued"
	ExportRunning         ExportJobState = "running"
	ExportCancelRequested ExportJobState = "cancel_requested"
	ExportSucceeded       ExportJobState = "succeeded"
	ExportFailed          ExportJobState = "failed"
	ExportCanceled        ExportJobState = "canceled"
)

type ExportOptions struct {
	TimelineID           string `json:"timelineId,omitempty"`
	Resolution           string `json:"resolution,omitempty"`
	FrameRate            int    `json:"frameRate,omitempty"`
	Quality              string `json:"quality,omitempty"`
	Codec                string `json:"codec,omitempty"`
	BitDepth             int    `json:"bitDepth,omitempty"`
	BitrateMbps          int    `json:"bitrateMbps,omitempty"`
	Alpha                bool   `json:"alpha,omitempty"`
	HardwareAcceleration string `json:"hardwareAcceleration,omitempty"`
	BackgroundColor      string `json:"-"`
	ResolvedVideoEncoder string `json:"-"`
}

type ExportJob struct {
	ID               string         `json:"id"`
	Owner            string         `json:"-"`
	ProjectID        string         `json:"projectId"`
	Format           string         `json:"format"`
	ExpectedRevision int64          `json:"expectedRevision"`
	State            ExportJobState `json:"state"`
	Progress         int            `json:"progress"`
	Error            string         `json:"error,omitempty"`
	OutputPath       string         `json:"-"`
	Options          ExportOptions  `json:"options,omitempty"`
	ProjectSnapshot  Project        `json:"-"`
	TimelineSnapshot Timeline       `json:"-"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

func validExportState(state ExportJobState) bool {
	switch state {
	case ExportQueued, ExportRunning, ExportCancelRequested, ExportSucceeded, ExportFailed, ExportCanceled:
		return true
	default:
		return false
	}
}

func validateExportJob(job ExportJob) error {
	job.ID = strings.TrimSpace(job.ID)
	job.Owner = strings.TrimSpace(job.Owner)
	job.ProjectID = strings.TrimSpace(job.ProjectID)
	job.Format = strings.ToLower(strings.TrimSpace(job.Format))
	if job.ID == "" || job.Owner == "" || job.ProjectID == "" || job.ProjectSnapshot.ID != job.ProjectID || job.ExpectedRevision < 0 || job.TimelineSnapshot.Revision != job.ExpectedRevision {
		return ErrInvalidExportJob
	}
	if job.Format != "mp4" && job.Format != "webm" && job.Format != "mov" && job.Format != "mp3" {
		return ErrInvalidExportJob
	}
	if !validExportState(job.State) || job.Progress < 0 || job.Progress > 100 || job.CreatedAt.IsZero() || job.UpdatedAt.IsZero() {
		return ErrInvalidExportJob
	}
	return nil
}

func (job ExportJob) WithState(state ExportJobState, progress int, message, output string, now time.Time) ExportJob {
	job.State = state
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	job.Progress = progress
	job.Error = strings.TrimSpace(message)
	if strings.TrimSpace(output) != "" {
		job.OutputPath = strings.TrimSpace(output)
	}
	job.UpdatedAt = now.UTC()
	return job
}
