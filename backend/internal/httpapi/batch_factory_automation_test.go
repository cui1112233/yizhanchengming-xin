package httpapi

import (
	"encoding/json"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestProductionLinesHonorsConfiguredContentRange(t *testing.T) {
	book := batchfactoryv11.Book{SourceText: "第一行\n\n第二行\r\n第三行\n第四行", SourceMetadata: map[string]any{"contentRangeLines": 3}}
	got := productionLines(book)
	if len(got) != 3 || got[0] != "第一行" || got[2] != "第三行" {
		t.Fatalf("productionLines() = %#v, want first three non-empty lines", got)
	}
}

func TestExistingAudioPlanIsReusableForTheSameProductionLines(t *testing.T) {
	book := batchfactoryv11.Book{
		SourceText:     "第一行\n第二行\n第三行",
		SourceMetadata: map[string]any{"contentRangeLines": 3},
		SettingsState: batchfactoryv11.SettingsState{Patch: batchfactoryv11.SettingsPatch{
			"audioPlanningEnabled": json.RawMessage("true"),
			"audioLineDurations":   json.RawMessage(`[1.1,2.2,3.3]`),
			"audioDurationSeconds": json.RawMessage(`6.6`),
		}},
	}
	if !hasReusableAudioPlan(book) {
		t.Fatal("an existing valid line-duration plan should be reusable")
	}
}

func TestProductionJobForRequestDoesNotReuseAnOlderCompletedJob(t *testing.T) {
	jobs := []batchfactoryv11.ProductionJob{
		{BookID: "book-1", RequestID: "old-video", Status: batchfactoryv11.ProductionSucceeded, Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded, MediaURL: "old.mp4"}}},
		{BookID: "book-1", RequestID: "current-video", Status: batchfactoryv11.ProductionRunning, Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionRunning}}},
	}
	job, ok := productionJobForRequest(jobs, "book-1", "current-video")
	if !ok || job.RequestID != "current-video" {
		t.Fatalf("productionJobForRequest() = %#v, %v; want current job", job, ok)
	}
}

func TestProductionJobCompleteRequiresSuccessfulMedia(t *testing.T) {
	job := batchfactoryv11.ProductionJob{Tasks: []batchfactoryv11.ProductionTask{{Status: batchfactoryv11.ProductionSucceeded}}}
	if productionJobComplete(job) {
		t.Fatal("productionJobComplete accepted a task without media URL")
	}
	job.Tasks[0].MediaURL = "/api/batch-factory/v12/media/video-1"
	if !productionJobComplete(job) {
		t.Fatal("productionJobComplete rejected a successful task with media URL")
	}
}
