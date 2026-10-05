package video

import (
	"context"
	"testing"
)

type projectVideoStatusStore struct {
	*redStore
	jobs  []ProductionJob
	tasks map[int64][]ProductionTask
}

func (s *projectVideoStatusStore) ListLatestProductionJobsByProject(context.Context, int64) ([]ProductionJob, error) {
	return append([]ProductionJob(nil), s.jobs...), nil
}
func (s *projectVideoStatusStore) ListProductionTasksByJobIDs(context.Context, []int64) (map[int64][]ProductionTask, error) {
	out := map[int64][]ProductionTask{}
	for jobID, tasks := range s.tasks {
		out[jobID] = append([]ProductionTask(nil), tasks...)
	}
	return out, nil
}

func TestProjectVideoStatusReturnsProviderModelAttemptsErrorAndOutput(t *testing.T) {
	store := &projectVideoStatusStore{
		redStore: newRedStore(ProviderConfig{}),
		jobs: []ProductionJob{{ID: 30, BatchProjectID: 7, BookID: 9, Provider: ProviderYFAISeedance, Model: ModelSeedance20Official, Status: JobFailed}},
		tasks: map[int64][]ProductionTask{
			30: {
				{ID: 31, ProductionJobID: 30, Attempt: 1, Status: TaskFailed, ErrorMessage: "first failed"},
				{ID: 32, ProductionJobID: 30, Attempt: 2, Status: TaskSucceeded, OutputURL: "https://tos.example/final.mp4"},
			},
		},
	}
	service := NewService(store, nil, nil, nil, nil)
	status, err := service.ProjectStatus(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Books) != 1 {
		t.Fatalf("status = %+v", status)
	}
	book := status.Books[0]
	if book.BookID != 9 || book.Provider != ProviderYFAISeedance || book.Model != ModelSeedance20Official {
		t.Fatalf("book status = %+v", book)
	}
	if len(book.Attempts) != 2 || book.Attempts[0].ErrorMessage != "first failed" || book.Attempts[1].OutputURL != "https://tos.example/final.mp4" {
		t.Fatalf("attempt history = %+v", book.Attempts)
	}
}
