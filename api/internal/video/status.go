package video

import (
	"context"
	"fmt"
)

type ProjectVideoStatusStore interface {
	ListLatestProductionJobsByProject(context.Context, int64) ([]ProductionJob, error)
	ListProductionTasksByJobIDs(context.Context, []int64) (map[int64][]ProductionTask, error)
}

type ProjectVideoStatus struct {
	BatchProjectID int64             `json:"batchProjectId"`
	Books          []BookVideoStatus `json:"books"`
}

type BookVideoStatus struct {
	BookID       int64              `json:"bookId"`
	JobID        int64              `json:"jobId"`
	Provider     string             `json:"provider"`
	Model        string             `json:"model"`
	Status       JobStatus          `json:"status"`
	Attempts     []VideoAttemptView `json:"attempts"`
	ErrorMessage string             `json:"errorMessage,omitempty"`
	OutputURL    string             `json:"outputUrl,omitempty"`
}

type VideoAttemptView struct {
	ID           int64      `json:"id"`
	Attempt      int        `json:"attempt"`
	Status       TaskStatus `json:"status"`
	ErrorCode    ErrorCode  `json:"errorCode,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
	OutputURL    string     `json:"outputUrl,omitempty"`
}

func (s *Service) ProjectStatus(ctx context.Context, projectID int64) (ProjectVideoStatus, error) {
	if projectID <= 0 {
		return ProjectVideoStatus{}, fmt.Errorf("video: invalid batch project id")
	}
	statusStore, ok := s.store.(ProjectVideoStatusStore)
	if !ok {
		return ProjectVideoStatus{}, providerError(ErrorProviderUnavailable, "video status store unavailable", nil)
	}
	jobs, err := statusStore.ListLatestProductionJobsByProject(ctx, projectID)
	if err != nil {
		return ProjectVideoStatus{}, err
	}
	jobIDs := make([]int64, 0, len(jobs))
	for _, job := range jobs {
		jobIDs = append(jobIDs, job.ID)
	}
	tasksByJob, err := statusStore.ListProductionTasksByJobIDs(ctx, jobIDs)
	if err != nil {
		return ProjectVideoStatus{}, err
	}
	result := ProjectVideoStatus{BatchProjectID: projectID, Books: make([]BookVideoStatus, 0, len(jobs))}
	for _, job := range jobs {
		book := BookVideoStatus{
			BookID: job.BookID, JobID: job.ID, Provider: job.Provider, Model: job.Model,
			Status: job.Status, ErrorMessage: job.ErrorMessage,
			Attempts: make([]VideoAttemptView, 0, len(tasksByJob[job.ID])),
		}
		for _, task := range tasksByJob[job.ID] {
			view := VideoAttemptView{ID: task.ID, Attempt: task.Attempt, Status: task.Status, ErrorCode: task.ErrorCode, ErrorMessage: task.ErrorMessage, OutputURL: task.OutputURL}
			book.Attempts = append(book.Attempts, view)
			if task.OutputURL != "" {
				book.OutputURL = task.OutputURL
			}
			if task.ErrorMessage != "" {
				book.ErrorMessage = task.ErrorMessage
			}
		}
		result.Books = append(result.Books, book)
	}
	return result, nil
}
