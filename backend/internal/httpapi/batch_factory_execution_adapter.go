package httpapi

import (
	"context"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryexecution"
)

type batchExecutionLegacyGateway struct {
	coordinator *BatchAutomationCoordinator
	reader      AutomationJobReader
}

func newBatchExecutionLegacyGateway(coordinator *BatchAutomationCoordinator, store AutomationStateStore) batchfactoryexecution.LegacyAutomationGateway {
	reader, _ := store.(AutomationJobReader)
	if coordinator == nil || reader == nil {
		return nil
	}
	return &batchExecutionLegacyGateway{coordinator: coordinator, reader: reader}
}

func (g *batchExecutionLegacyGateway) PeekActive(ctx context.Context, owner, batchID string) (batchfactoryexecution.LegacyTask, bool, error) {
	job, found, err := g.reader.FindActiveJob(ctx, owner, batchID)
	if err != nil || !found {
		return batchfactoryexecution.LegacyTask{}, found, err
	}
	return legacyTaskFromAutomationJob(job), true, nil
}

func (g *batchExecutionLegacyGateway) Start(ctx context.Context, owner, batchID string, input batchfactoryexecution.LegacyStartInput) (batchfactoryexecution.LegacyTask, error) {
	payload := map[string]any{"bookIds": input.BookIDs, "productionMode": input.Mode, "target": input.Target, "clientRequestId": input.ClientRequestID, "queueIfActive": true}
	for key, value := range input.Config {
		payload[key] = value
	}
	job, err := g.coordinator.Start(ctx, owner, batchID, payload)
	if err != nil {
		return batchfactoryexecution.LegacyTask{}, err
	}
	return legacyTaskFromAutomationJob(job), nil
}

func (g *batchExecutionLegacyGateway) Read(ctx context.Context, _ string, jobID string) (batchfactoryexecution.LegacyTask, bool, error) {
	job, found, err := g.reader.FindJob(ctx, jobID)
	if err != nil || !found {
		return batchfactoryexecution.LegacyTask{}, found, err
	}
	return legacyTaskFromAutomationJob(job), true, nil
}

func legacyTaskFromAutomationJob(job map[string]any) batchfactoryexecution.LegacyTask {
	task := batchfactoryexecution.LegacyTask{JobID: strings.TrimSpace(automationStringValue(job["id"], "")), State: strings.TrimSpace(automationStringValue(job["state"], "")), Since: automationStoredJobTime(job["createdAt"]), Books: map[string]batchfactoryexecution.LegacyBook{}}
	if task.Since.IsZero() {
		task.Since = time.Now().UTC()
	}
	books, _ := job["books"].(map[string]any)
	for key, raw := range books {
		entry, _ := raw.(map[string]any)
		if entry == nil {
			continue
		}
		id := strings.TrimSpace(automationStringValue(entry["bookId"], key))
		task.Books[id] = batchfactoryexecution.LegacyBook{BookID: id, State: strings.TrimSpace(automationStringValue(entry["status"], automationStringValue(entry["state"], ""))), Message: strings.TrimSpace(automationStringValue(entry["message"], automationStringValue(entry["error"], "")))}
	}
	return task
}
