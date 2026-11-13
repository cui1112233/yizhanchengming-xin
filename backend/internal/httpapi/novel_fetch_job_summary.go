package httpapi

// Queue polling needs status only. Completed job results may contain every
// book body and must remain on the job detail endpoint, not repeated per poll.
func novelFetchJobStatusSummary(job map[string]any) map[string]any {
	result := map[string]any{}
	for _, key := range []string{"id", "status", "queue_state", "stage", "progress", "total", "completed", "failed", "createdAt", "updatedAt", "created_at", "updated_at", "started_at", "finished_at", "error", "message", "batch_id"} {
		if value, ok := job[key]; ok {
			result[key] = value
		}
	}
	return result
}
