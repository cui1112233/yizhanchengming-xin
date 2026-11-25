package novelcreation

import (
	"context"
	"testing"
)

func TestDashboardIncludesStatusCountsAndWordTrend(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, NewMemoryRuntime(), nil)
	ctx := context.Background()
	project, err := service.CreateProject(ctx, "alice", CreateProjectInput{Title: "看板书", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	for _, chapter := range []Chapter{
		{ID: "c2", ProjectID: project.ID, Number: 2, Title: "第二章", Content: "字字字", WordCount: 1500, AuditStatus: "warning", Version: 1},
		{ID: "c1", ProjectID: project.ID, Number: 1, Title: "第一章", Content: "字字", WordCount: 1000, AuditStatus: "passed", Version: 1},
		{ID: "c3", ProjectID: project.ID, Number: 3, Title: "第三章", Content: "字", WordCount: 500, AuditStatus: "passed", Version: 1},
	} {
		if err = store.SaveChapterBundle(ctx, "alice", ChapterBundle{ProjectID: project.ID, Chapter: chapter}); err != nil {
			t.Fatal(err)
		}
	}
	dashboard, err := service.GetDashboard(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.ChapterCount != 3 || dashboard.TotalWords != 3000 {
		t.Fatalf("totals=%+v", dashboard)
	}
	if dashboard.AvgWords != 1000 {
		t.Fatalf("avg=%v want 1000", dashboard.AvgWords)
	}
	if dashboard.StatusCounts["passed"] != 2 || dashboard.StatusCounts["warning"] != 1 {
		t.Fatalf("status=%v", dashboard.StatusCounts)
	}
	if len(dashboard.WordTrend) != 3 || dashboard.WordTrend[0].Number != 1 || dashboard.WordTrend[2].Number != 3 {
		t.Fatalf("trend not sorted: %+v", dashboard.WordTrend)
	}
	if dashboard.WordTrend[1].WordCount != 1500 || dashboard.WordTrend[0].Status != "passed" {
		t.Fatalf("trend=%+v", dashboard.WordTrend)
	}
}

func TestDashboardEmptyProjectHasEmptyTrend(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, NewMemoryRuntime(), nil)
	ctx := context.Background()
	project, err := service.CreateProject(ctx, "alice", CreateProjectInput{Title: "空书", Genre: "悬疑", ModelID: "text-1"})
	if err != nil {
		t.Fatal(err)
	}
	dashboard, err := service.GetDashboard(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.AvgWords != 0 || dashboard.WordTrend == nil || len(dashboard.WordTrend) != 0 || len(dashboard.StatusCounts) != 0 {
		t.Fatalf("empty dashboard=%+v", dashboard)
	}
}
