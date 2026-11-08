package batchfactoryv11

import (
	"encoding/json"
	"testing"
	"time"
)

func TestStoryboardTaskCandidatesKeepThePreviousVersion(t *testing.T) {
	book := Book{
		ID:     "book-1",
		Videos: []Video{{ID: "video-active", Ordinal: 1}},
		VideoHistory: []Video{
			{ID: "video-old", Ordinal: 1, CompatibilityState: "orphaned"},
			{ID: "video-other", Ordinal: 0, CompatibilityState: "orphaned"},
		},
	}
	activeTask := ProductionTask{ID: "task-active", VideoID: "video-active", Status: ProductionRunning, UpdatedAt: time.Date(2026, 11, 5, 1, 0, 0, 0, time.UTC)}
	oldTask := ProductionTask{ID: "task-old", VideoID: "video-old", Status: ProductionSucceeded, MediaURL: "https://cdn/old.mp4", UpdatedAt: time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)}
	otherOrdinalTask := ProductionTask{ID: "task-other", VideoID: "video-other", Status: ProductionSucceeded, MediaURL: "https://cdn/other.mp4", UpdatedAt: time.Date(2026, 11, 5, 0, 30, 0, 0, time.UTC)}

	current := productionTaskSelections([]ProductionJob{{
		ID: "job-current", BookID: "book-1", DirectorRevisionID: "director-2", Tasks: []ProductionTask{activeTask},
	}}, "book-1", "director-2")
	all := productionTaskSelectionsAllRevisions([]ProductionJob{
		{ID: "job-current", BookID: "book-1", DirectorRevisionID: "director-2", Tasks: []ProductionTask{activeTask}},
		{ID: "job-old", BookID: "book-1", DirectorRevisionID: "director-1", Tasks: []ProductionTask{oldTask, otherOrdinalTask}},
	}, "book-1")

	candidates := storyboardTaskCandidates(book.Videos[0], book, current, all)
	if len(candidates) != 2 {
		t.Fatalf("candidates = %+v, want active task + previous version of the same ordinal only", candidates)
	}
	if candidates[0].Task.ID != "task-old" || candidates[1].Task.ID != "task-active" {
		t.Fatalf("candidates must be ordered by update time: %+v", candidates)
	}
}

func TestSelectedProductionTaskAcceptsAPrimaryFromThePreviousVersion(t *testing.T) {
	video := Video{
		ID:      "video-active",
		Ordinal: 1,
		SettingsState: SettingsState{Patch: SettingsPatch{
			"primaryMediaTaskId": json.RawMessage(`"task-old"`),
		}},
	}
	old := productionTaskSelection{JobID: "job-old", Task: ProductionTask{
		ID: "task-old", VideoID: "video-old", Status: ProductionSucceeded, MediaURL: "https://cdn/old.mp4",
	}}
	running := productionTaskSelection{JobID: "job-current", Task: ProductionTask{
		ID: "task-active", VideoID: "video-active", Status: ProductionRunning,
	}}
	selected, ok := selectedProductionTask(video, []productionTaskSelection{running, old})
	if !ok || selected.Task.ID != "task-old" {
		t.Fatalf("selected = %+v ok=%v, want the pinned previous version", selected, ok)
	}
}

// 槽位级保留：未变化槽位跨修订沿用同一 video 身份，成功任务挂在旧修订的
// job 上。候选查询必须并入同 ID 的全修订任务，否则合成会认为保留槽位没有
// 已生成媒体。
func TestStoryboardTaskCandidatesIncludeSameIDTasksFromOlderRevisions(t *testing.T) {
	book := Book{
		ID:     "book-1",
		Videos: []Video{{ID: "video-kept", Ordinal: 0}},
	}
	keptTask := ProductionTask{ID: "task-kept", VideoID: "video-kept", Status: ProductionSucceeded, MediaURL: "https://cdn/kept.mp4", UpdatedAt: time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)}
	current := productionTaskSelections([]ProductionJob{{
		ID: "job-new", BookID: "book-1", DirectorRevisionID: "director-2", Tasks: []ProductionTask{},
	}}, "book-1", "director-2")
	all := productionTaskSelectionsAllRevisions([]ProductionJob{
		{ID: "job-old", BookID: "book-1", DirectorRevisionID: "director-1", Tasks: []ProductionTask{keptTask}},
	}, "book-1")

	candidates := storyboardTaskCandidates(book.Videos[0], book, current, all)
	if len(candidates) != 1 || candidates[0].Task.ID != "task-kept" {
		t.Fatalf("candidates = %+v, want the kept slot's old-revision success", candidates)
	}
	selected, ok := selectedProductionTask(book.Videos[0], candidates)
	if !ok || selected.Task.ID != "task-kept" {
		t.Fatalf("selected = %+v ok=%v, the kept slot must be merge-eligible", selected, ok)
	}
}
