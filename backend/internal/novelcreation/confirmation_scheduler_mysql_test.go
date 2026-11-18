package novelcreation

import (
	"context"
	"testing"
)

func TestConfirmationSchedulerRealMySQLRoundTrip(t *testing.T) {
	db := studioOwnedDB(t)
	ctx := context.Background()
	owner := novelID("scheduler-sql")
	st := NewMySQLStore(db)
	s := NewService(st, NewMemoryRuntime(), nil)
	project, e := s.CreateProject(ctx, owner, CreateProjectInput{Title: "调度持久化", Genre: "悬疑", ModelID: "fake-only", TargetChapterWords: 2000})
	if e != nil {
		t.Fatal(e)
	}
	session, e := s.CreateSession(ctx, owner, CreateSessionInput{ProjectID: project.ID, Mode: "long_original", Title: "调度"})
	if e != nil {
		t.Fatal(e)
	}
	settings, e := s.GetProjectSettings(ctx, owner, project.ID)
	if e != nil {
		t.Fatal(e)
	}
	settings.Agent.PauseAfterBlueprint = boolPointer(false)
	settings.Agent.RequireSecondaryConfirm = true
	if _, e = s.UpdateProjectSettings(ctx, owner, project.ID, settings); e != nil {
		t.Fatal(e)
	}
	reopened := NewService(NewMySQLStore(db), NewMemoryRuntime(), nil)
	saved, e := reopened.GetProjectSettings(ctx, owner, project.ID)
	if e != nil || saved.Agent.PauseAfterBlueprint == nil || *saved.Agent.PauseAfterBlueprint || !saved.Agent.RequireSecondaryConfirm {
		t.Fatal("independent settings lost in SQL", saved, e)
	}
	input := SchedulerStartInput{IdempotencyKey: "sql-start", SessionID: session.ID, ModelID: "fake-only", IntervalSeconds: 1800, DailyChapterCap: 10, MaxConsecutiveFailures: 3}
	plan, e := reopened.StartScheduler(ctx, owner, project.ID, input)
	if e != nil {
		t.Fatal(e)
	}
	next := NewService(NewMySQLStore(db), NewMemoryRuntime(), nil)
	again, e := next.StartScheduler(ctx, owner, project.ID, input)
	if e != nil || again.ID != plan.ID {
		t.Fatal("SQL start replay", e)
	}
	for _, action := range []string{"pause", "resume", "stop"} {
		v, e := next.ControlSchedule(ctx, owner, project.ID, plan.ID, action)
		if e != nil {
			t.Fatal(action, e)
		}
		read, e := NewMySQLStore(db).GetSchedule(ctx, owner, plan.ID)
		if e != nil || v.Status != read.Status {
			t.Fatal("control not durable", action, e)
		}
	}
	runs, e := st.ListRuns(ctx, owner, project.ID)
	if e != nil || len(runs) != 0 {
		t.Fatal("start generated or corrupted dispatch", e)
	}
	t.Log("MySQL confirms independent settings, scheduler receipt replay, pause/resume/stop and zero model runs")
}
