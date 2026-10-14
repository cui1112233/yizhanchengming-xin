package httpapi

import (
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
