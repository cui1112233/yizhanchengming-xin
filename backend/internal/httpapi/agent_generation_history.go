package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func RegisterAgentGenerationHistoryRoutes(mux *http.ServeMux, auth BridgeAuth, store agentworkspace.Store, assets agentworkspace.AssetStore, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	mux.Handle("GET /api/agent/generation-history", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		tasks, err := store.ListOwnerGenerationTasks(req.Context(), identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "生成历史读取失败"})
			return
		}
		kind := strings.TrimSpace(req.URL.Query().Get("kind"))
		modelID := strings.TrimSpace(req.URL.Query().Get("modelId"))
		state := agentworkspace.GenerationState(strings.TrimSpace(req.URL.Query().Get("state")))
		query := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("q")))
		items := make([]agentworkspace.GenerationTask, 0, len(tasks))
		for _, task := range tasks {
			if len(task.Results) == 0 || (kind != "" && task.Kind != kind) || (modelID != "" && task.ModelID != modelID) || (state != "" && task.State != state) || !agentGenerationHistoryMatches(task, query) {
				continue
			}
			items = append(items, task)
		}
		writeJSON(w, http.StatusOK, map[string]any{"tasks": items})
	})))
	mux.Handle("POST /api/agent/tasks/{id}/generation-history/{taskID}/results/{resultID}/insert", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if assets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		owner := identity.Username
		targetProjectID := strings.TrimSpace(req.PathValue("id"))
		sourceTaskID := strings.TrimSpace(req.PathValue("taskID"))
		sourceResultID := strings.TrimSpace(req.PathValue("resultID"))
		sourceTask, sourceResult, err := agentGenerationHistoryResult(req.Context(), store, owner, sourceTaskID, sourceResultID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "生成历史结果不存在"})
			return
		}
		if sourceResult.Kind != "image" && sourceResult.Kind != "video" && sourceResult.Kind != "audio" {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "该历史结果不能作为媒体节点插入"})
			return
		}
		sourceAsset, err := assets.Asset(req.Context(), owner, sourceTask.ProjectID, sourceResult.AssetID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "历史素材不存在"})
			return
		}
		canvas, err := store.Canvas(req.Context(), owner, targetProjectID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "当前项目不存在"})
			return
		}
		createdAt := time.Now().UTC()
		if now != nil {
			createdAt = now().UTC()
		}
		assetID := newAgentID("history-asset")
		importedAsset := agentworkspace.Asset{
			ID: assetID, ProjectID: targetProjectID, Kind: sourceAsset.Kind, Name: sourceAsset.Name, ContentType: sourceAsset.ContentType,
			ObjectKey: sourceAsset.ObjectKey, ByteSize: sourceAsset.ByteSize, CreatedAt: createdAt,
		}
		if err := assets.CreateAsset(req.Context(), owner, importedAsset); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "历史素材写入失败"})
			return
		}
		node := agentworkspace.Node{
			ID: newAgentID("history-node"), Type: sourceResult.Kind, Title: "历史素材：" + sourceAsset.Name, Version: 1, Status: "confirmed",
			Position: agentGenerationHistoryNodePosition(canvas),
			AssetRef: map[string]any{
				"id": assetID, "assetId": assetID, "kind": sourceResult.Kind, "name": sourceAsset.Name,
				"url": agentAssetContentURL(targetProjectID, assetID), "originProjectId": sourceTask.ProjectID,
				"generationTaskId": sourceTask.ID, "generationResultId": sourceResult.ID,
			},
		}
		canvas.Nodes = append(canvas.Nodes, node)
		updated, err := store.SaveCanvas(req.Context(), owner, targetProjectID, canvas, canvas.Revision, createdAt)
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前项目已更新，请刷新后重试"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"canvas": updated, "node": node, "asset": importedAsset})
	})))
}

func agentGenerationHistoryResult(ctx context.Context, store agentworkspace.GenerationHistoryStore, owner, taskID, resultID string) (agentworkspace.GenerationTask, agentworkspace.GenerationResult, error) {
	tasks, err := store.ListOwnerGenerationTasks(ctx, owner)
	if err != nil {
		return agentworkspace.GenerationTask{}, agentworkspace.GenerationResult{}, err
	}
	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}
		for _, result := range task.Results {
			if result.ID == resultID {
				return task, result, nil
			}
		}
	}
	return agentworkspace.GenerationTask{}, agentworkspace.GenerationResult{}, agentworkspace.ErrGenerationNotFound
}

func agentGenerationHistoryNodePosition(canvas agentworkspace.Canvas) agentworkspace.Position {
	maxY := 0.0
	for _, node := range canvas.Nodes {
		if node.Position.Y > maxY {
			maxY = node.Position.Y
		}
	}
	return agentworkspace.Position{X: 80, Y: maxY + 240}
}

func agentGenerationHistoryMatches(task agentworkspace.GenerationTask, query string) bool {
	if query == "" {
		return true
	}
	fields := []string{task.ID, task.ProjectID, task.Kind, task.Provider, task.ModelID, task.Mode}
	for _, result := range task.Results {
		fields = append(fields, result.ID, result.AssetID)
		for _, key := range []string{"name", "title"} {
			if value, ok := result.Asset[key].(string); ok {
				fields = append(fields, value)
			}
		}
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	return false
}
