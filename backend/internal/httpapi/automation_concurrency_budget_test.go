package httpapi

import (
	"strings"
	"testing"
)

func TestAutomationConcurrencyBudgetReportsGatewayColdStart(t *testing.T) {
	budget := automationConcurrencyBudget(map[string]any{"concurrency": 6, "imageGlobalConcurrency": 8})
	if got := automationNumberValue(budget["bookWorkers"], 0); got != automationBookWorkerCeiling {
		t.Fatalf("bookWorkers=%d, want ceiling %d", got, automationBookWorkerCeiling)
	}
	if got := automationNumberValue(budget["effectiveImageConcurrency"], 0); got != shuihuoAdaptiveInitialLimit {
		t.Fatalf("effectiveImageConcurrency=%d, want cold start %d", got, shuihuoAdaptiveInitialLimit)
	}
	if got := toString(budget["limitingLayer"]); got != "gateway_cold_start" {
		t.Fatalf("limitingLayer=%q, want gateway_cold_start", got)
	}
	if got := automationNumberValue(budget["gatewayCeiling"], 0); got != shuihuoAdaptiveMaxLimit {
		t.Fatalf("gatewayCeiling=%d, want %d", got, shuihuoAdaptiveMaxLimit)
	}
}

func TestAutomationConcurrencyBudgetReportsTaskLimit(t *testing.T) {
	budget := automationConcurrencyBudget(map[string]any{"concurrency": 2, "imageGlobalConcurrency": 1})
	if got := automationNumberValue(budget["bookWorkers"], 0); got != 2 {
		t.Fatalf("bookWorkers=%d, want 2", got)
	}
	if got := automationNumberValue(budget["effectiveImageConcurrency"], 0); got != 1 {
		t.Fatalf("effectiveImageConcurrency=%d, want 1", got)
	}
	if got := toString(budget["limitingLayer"]); got != "task_image_concurrency" {
		t.Fatalf("limitingLayer=%q, want task_image_concurrency", got)
	}
	if got := automationNumberValue(budget["comicBookSlots"], 0); got != comicBookGenerationConcurrency() {
		t.Fatalf("comicBookSlots=%d, want %d", got, comicBookGenerationConcurrency())
	}
}

func TestAutomationConcurrencyBudgetLineExplainsLayers(t *testing.T) {
	line := automationConcurrencyBudgetLine(map[string]any{"concurrency": 3, "imageGlobalConcurrency": 8})
	for _, want := range []string{"书 3/4 路", "本任务出图 8 路", "网关冷启动 2 路", "上限 30 路"} {
		if !strings.Contains(line, want) {
			t.Fatalf("budget line=%q, want %q", line, want)
		}
	}
}

func TestPublicAutomationJobExposesConcurrencyBudget(t *testing.T) {
	view := publicAutomationJob(map[string]any{
		"id": "job-1", "batchId": "batch-1", "state": "running",
		"concurrency": 3, "imageGlobalConcurrency": 4,
	})
	budget, ok := view["concurrencyBudget"].(map[string]any)
	if !ok {
		t.Fatalf("concurrencyBudget missing from public job: %v", view["concurrencyBudget"])
	}
	if got := automationNumberValue(budget["taskImageConcurrency"], 0); got != 4 {
		t.Fatalf("taskImageConcurrency=%d, want 4", got)
	}
}
