package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
)

type fakeStore struct {
	intakes  map[int64]intake.Intake
	projects []intake.BatchProject
	runs     []intake.Run
}

func (s *fakeStore) GetIntake(_ context.Context, id int64) (intake.Intake, error) {
	value, ok := s.intakes[id]
	if !ok {
		return intake.Intake{}, errors.New("not found")
	}
	return value, nil
}

func (s *fakeStore) CreateBatchProject(_ context.Context, project intake.BatchProject) (intake.BatchProject, error) {
	project.ID = int64(len(s.projects) + 1)
	s.projects = append(s.projects, project)
	return project, nil
}

func (s *fakeStore) CreateRun(_ context.Context, run intake.Run) (intake.Run, error) {
	run.ID = int64(len(s.runs) + 1)
	s.runs = append(s.runs, run)
	return run, nil
}

func TestCreateImmediateAndScheduledUseSamePipeline(t *testing.T) {
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	store := &fakeStore{intakes: map[int64]intake.Intake{
		11: {ID: 11, Name: "知乎+黑岩", Status: intake.StatusCompleted},
		12: {ID: 12, Name: "番茄", Status: intake.StatusCompleted},
	}}
	service := NewService(store, func() time.Time { return now })

	immediate, err := service.Create(context.Background(), CreateRequest{IntakeID: 11, Name: "立即执行项目"})
	if err != nil {
		t.Fatalf("immediate Create: %v", err)
	}
	if !immediate.Run.RunAt.Equal(now) {
		t.Fatalf("immediate run_at = %v, want %v", immediate.Run.RunAt, now)
	}
	if immediate.Run.Status != intake.RunStatusPending {
		t.Fatalf("immediate status = %q", immediate.Run.Status)
	}

	scheduledAt := now.Add(6 * time.Hour)
	scheduled, err := service.Create(context.Background(), CreateRequest{IntakeID: 12, Name: "自动化项目", RunAt: scheduledAt})
	if err != nil {
		t.Fatalf("scheduled Create: %v", err)
	}
	if !scheduled.Run.RunAt.Equal(scheduledAt) {
		t.Fatalf("scheduled run_at = %v, want %v", scheduled.Run.RunAt, scheduledAt)
	}

	if len(store.projects) != 2 || len(store.runs) != 2 {
		t.Fatalf("projects/runs = %d/%d, want 2/2", len(store.projects), len(store.runs))
	}
	if store.projects[0].IntakeID != 11 || store.projects[1].IntakeID != 12 {
		t.Fatalf("projects = %+v", store.projects)
	}
}

func TestCreateDefaultsProjectNameFromIntake(t *testing.T) {
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	store := &fakeStore{intakes: map[int64]intake.Intake{
		21: {ID: 21, Name: "知乎+点众", Status: intake.StatusCompleted},
	}}
	service := NewService(store, func() time.Time { return now })

	created, err := service.Create(context.Background(), CreateRequest{IntakeID: 21})
	if err != nil {
		t.Fatal(err)
	}
	if created.Project.Name != "知乎+点众" {
		t.Fatalf("project name = %q", created.Project.Name)
	}
}

func TestCreateRejectsNonCompletedIntake(t *testing.T) {
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	statuses := []intake.Status{intake.StatusPending, intake.StatusRunning, intake.StatusPartial, intake.StatusFailed}
	for i, status := range statuses {
		store := &fakeStore{intakes: map[int64]intake.Intake{
			31: {ID: 31, Name: "未完成", Status: status},
		}}
		service := NewService(store, func() time.Time { return now })
		_, err := service.Create(context.Background(), CreateRequest{IntakeID: 31})
		if err == nil {
			t.Fatalf("case %d status %q: expected error", i, status)
		}
		if len(store.projects) != 0 || len(store.runs) != 0 {
			t.Fatalf("status %q created project/run unexpectedly", status)
		}
	}
}

func TestCreateRejectsScheduledTimeInPast(t *testing.T) {
	now := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	store := &fakeStore{intakes: map[int64]intake.Intake{
		41: {ID: 41, Name: "完成批次", Status: intake.StatusCompleted},
	}}
	service := NewService(store, func() time.Time { return now })

	_, err := service.Create(context.Background(), CreateRequest{IntakeID: 41, RunAt: now.Add(-time.Minute)})
	if err == nil {
		t.Fatal("expected past run_at error")
	}
	if len(store.projects) != 0 || len(store.runs) != 0 {
		t.Fatalf("created project/run for invalid run_at")
	}
}
