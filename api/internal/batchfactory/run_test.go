package batchfactory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type fakeRunIntakeCreator struct {
	input  CreateIntakeInput
	result CreateIntakeResult
	err    error
}

func (f *fakeRunIntakeCreator) Create(_ context.Context, input CreateIntakeInput) (CreateIntakeResult, error) {
	f.input = input
	return f.result, f.err
}

type fakeRunStarter struct {
	input StartInput
	job   pipeline.Job
	err   error
}

func (f *fakeRunStarter) Start(_ context.Context, input StartInput) (pipeline.Job, error) {
	f.input = input
	return f.job, f.err
}

func TestRunServiceCreatesIntakeThenStartsSamePipeline(t *testing.T) {
	runAt := time.Date(2026, 10, 4, 9, 30, 0, 0, time.UTC)
	intakes := &fakeRunIntakeCreator{result: CreateIntakeResult{IntakeID: "intake-1", GroupCount: 2, BookCount: 3}}
	starter := &fakeRunStarter{job: pipeline.Job{ID: "job-1", IntakeID: "intake-1", RunAt: runAt, Status: "queued"}}
	service := RunService{Intakes: intakes, Starter: starter}

	got, err := service.CreateAndStart(context.Background(), CreateRunInput{
		Intake: CreateIntakeInput{Owner: "user-1", Groups: []IntakeGroupInput{
			{PlatformID: "zhihu", PlatformName: "知乎", Books: []IntakeBookInput{{BookID: "z1"}, {BookID: "z2"}}},
			{PlatformID: "dianzhong", PlatformName: "点众", Books: []IntakeBookInput{{BookID: "d1"}}},
		}},
		RunAt: runAt,
	})
	if err != nil { t.Fatal(err) }
	if starter.input.IntakeID != "intake-1" || !starter.input.RunAt.Equal(runAt) {
		t.Fatalf("start input=%+v", starter.input)
	}
	if got.IntakeID != "intake-1" || got.JobID != "job-1" || got.GroupCount != 2 || got.BookCount != 3 || got.Status != "queued" {
		t.Fatalf("result=%+v", got)
	}
}

func TestRunServiceReturnsCreatedIntakeWhenQueueStartFails(t *testing.T) {
	intakes := &fakeRunIntakeCreator{result: CreateIntakeResult{IntakeID: "intake-1", GroupCount: 1, BookCount: 1}}
	starter := &fakeRunStarter{err: errors.New("redis unavailable")}
	service := RunService{Intakes: intakes, Starter: starter}

	got, err := service.CreateAndStart(context.Background(), CreateRunInput{Intake: CreateIntakeInput{Owner: "user-1", Groups: []IntakeGroupInput{{PlatformID:"2", PlatformName:"番茄", Books:[]IntakeBookInput{{BookID:"b1"}}}}}})
	if err == nil { t.Fatal("expected start error") }
	if got.IntakeID != "intake-1" || got.BookCount != 1 {
		t.Fatalf("created intake should remain recoverable: %+v", got)
	}
}
