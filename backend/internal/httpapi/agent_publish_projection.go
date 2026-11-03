package httpapi

import (
	"context"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11/external"
)

// batchFactoryPublishResults adapts the external publish service into the
// Agent canvas' read-only projection port. It exposes status only; payloads and
// credentials stay inside the publish system.
type batchFactoryPublishResults struct {
	service *external.Service
}

func (r batchFactoryPublishResults) BatchFactoryPublishResults(ctx context.Context, owner, batchID, bookID string) ([]agentworkspace.BatchFactoryPublishResult, error) {
	results, err := r.service.PublishResults(ctx, owner, batchID, bookID)
	if err != nil {
		return nil, err
	}
	out := make([]agentworkspace.BatchFactoryPublishResult, 0, len(results))
	for _, result := range results {
		out = append(out, agentworkspace.BatchFactoryPublishResult{
			Provider: string(result.Provider), Status: result.Status, SubmittedAt: result.SubmittedAt,
		})
	}
	return out, nil
}

func newBatchFactoryPublishResults(service *external.Service) agentworkspace.BatchFactoryPublishResultReader {
	if service == nil {
		return nil
	}
	return batchFactoryPublishResults{service: service}
}
