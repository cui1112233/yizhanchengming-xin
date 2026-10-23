package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

type agentBatchFactoryBindingResponse struct {
	ProjectID string `json:"projectId"`
	BatchID   string `json:"batchId"`
	BookID    string `json:"bookId,omitempty"`
	Source    string `json:"source"`
	Version   int64  `json:"version"`
	State     string `json:"state"`
}

type agentBatchFactoryBridgeStore interface {
	ListBatches(context.Context, string) ([]batchfactoryv11.Batch, error)
	agentworkspace.BatchFactoryProjectionReader
}

// RegisterAgentBatchFactoryBridge exposes only owner-scoped references between
// Batch Factory and Agent. It intentionally has no Batch Factory delete port.
func RegisterAgentBatchFactoryBridge(root *http.ServeMux, middleware BridgeAuth, projects agentworkspace.Store, batches agentBatchFactoryBridgeStore, projection *agentworkspace.BatchFactoryProjectionService, clock func() time.Time) {
	if root == nil || projects == nil || batches == nil {
		return
	}
	projectionStore, ok := projects.(agentworkspace.BatchFactoryProjectionStore)
	if !ok {
		return
	}
	if clock == nil {
		clock = time.Now
	}
	if projection == nil {
		projection = &agentworkspace.BatchFactoryProjectionService{CanvasStore: projectionStore, BatchFactory: batches, Now: clock}
	}
	if projection.CanvasStore == nil || projection.BatchFactory == nil {
		return
	}

	ownerFor := func(w http.ResponseWriter, req *http.Request) (string, bool) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return "", false
		}
		return strings.TrimSpace(identity.Username), true
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if owner, ok := ownerFor(w, req); ok {
				fn(w, req, owner)
			}
		})))
	}

	handle("POST /api/batch-factory/v11/batches/{batchId}/agent-project", func(w http.ResponseWriter, req *http.Request, owner string) {
		batch, err := batches.GetBatch(req.Context(), owner, req.PathValue("batchId"))
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		binding, err := projects.BatchFactoryBinding(req.Context(), owner, batch.ID)
		if errors.Is(err, agentworkspace.ErrBatchFactoryBindingNotFound) {
			now := clock().UTC()
			projectID := agentBatchFactoryProjectID(owner, batch.ID)
			project := agentworkspace.Project{ID: projectID, Title: "批量工厂项目", TitleMode: "automatic", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
			if createErr := projects.CreateProject(req.Context(), owner, project); createErr != nil && !errors.Is(createErr, agentworkspace.ErrProjectExists) {
				writeAgentStoreError(w, createErr)
				return
			}
			binding, err = projects.BindBatchFactoryProject(req.Context(), agentworkspace.BatchFactoryBinding{Owner: owner, BatchID: batch.ID, ProjectID: projectID, CreatedAt: now, UpdatedAt: now})
		}
		if errors.Is(err, agentworkspace.ErrBatchFactoryBindingConflict) {
			// The MySQL unique key can win a concurrent insert after the first
			// lookup. The deterministic project ID makes an exact readback safe.
			if existing, readErr := projects.BatchFactoryBinding(req.Context(), owner, batch.ID); readErr == nil && existing.ProjectID == agentBatchFactoryProjectID(owner, batch.ID) {
				binding, err = existing, nil
			}
		}
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bindingResponse(binding, "", batch.Revision, "bound"))
	})

	handle("GET /api/agent/tasks/{id}/batch-factory", func(w http.ResponseWriter, req *http.Request, owner string) {
		if _, err := projects.Project(req.Context(), owner, req.PathValue("id")); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		binding, batch, err := agentBatchFactoryBindingForProject(req.Context(), projects, batches, owner, req.PathValue("id"))
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bindingResponse(binding, "", batch.Revision, "bound"))
	})

	handle("POST /api/agent/tasks/{id}/batch-factory/books/{bookId}/sync", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := req.PathValue("id")
		if _, err := projects.Project(req.Context(), owner, projectID); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		binding, batch, err := agentBatchFactoryBindingForProject(req.Context(), projects, batches, owner, projectID)
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		bookID := strings.TrimSpace(req.PathValue("bookId"))
		if !batchFactoryBookExists(batch, bookID) {
			writeAgentBatchFactoryError(w, batchfactoryv11.ErrNotFound)
			return
		}
		canvas, err := projection.SyncBatchFactoryBook(req.Context(), owner, binding, bookID)
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bindingResponse(binding, bookID, canvas.Revision, "synchronized"))
	})

	handle("DELETE /api/agent/tasks/{id}/batch-factory/nodes/{nodeId}", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := req.PathValue("id")
		nodeID := strings.TrimSpace(req.PathValue("nodeId"))
		if nodeID == "" {
			writeAgentBatchFactoryError(w, agentworkspace.ErrBatchFactoryNodeOriginInvalid)
			return
		}
		if _, err := projects.Project(req.Context(), owner, projectID); err != nil {
			writeAgentStoreError(w, err)
			return
		}
		binding, batch, err := agentBatchFactoryBindingForProject(req.Context(), projects, batches, owner, projectID)
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		bookID, origins, found, err := agentBatchFactoryNodeOrigins(req.Context(), projects, owner, binding, batch, nodeID)
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		if !found {
			writeAgentBatchFactoryError(w, agentworkspace.ErrBatchFactoryBindingNotFound)
			return
		}
		canvas, err := projects.Canvas(req.Context(), owner, projectID)
		if err != nil {
			writeAgentStoreError(w, err)
			return
		}
		next := removeAgentBatchFactoryCanvasNode(canvas, nodeID)
		remaining := make([]agentworkspace.BatchFactoryNodeOrigin, 0, len(origins)-1)
		for _, origin := range origins {
			if origin.NodeID != nodeID {
				remaining = append(remaining, origin)
			}
		}
		saved, err := projection.CanvasStore.SaveBatchFactoryProjection(req.Context(), owner, binding, bookID, next, canvas.Revision, remaining, clock().UTC())
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bindingResponse(binding, bookID, saved.Revision, "canvas_reference_deleted"))
	})
}

func bindingResponse(binding agentworkspace.BatchFactoryBinding, bookID string, version int64, state string) agentBatchFactoryBindingResponse {
	return agentBatchFactoryBindingResponse{ProjectID: binding.ProjectID, BatchID: binding.BatchID, BookID: strings.TrimSpace(bookID), Source: "batch_factory", Version: version, State: state}
}

func agentBatchFactoryProjectID(owner, batchID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(batchID)))
	return "agent-bf-" + hex.EncodeToString(digest[:12])
}

func agentBatchFactoryBindingForProject(ctx context.Context, projects agentworkspace.Store, batches agentBatchFactoryBridgeStore, owner, projectID string) (agentworkspace.BatchFactoryBinding, batchfactoryv11.Batch, error) {
	items, err := batches.ListBatches(ctx, owner)
	if err != nil {
		return agentworkspace.BatchFactoryBinding{}, batchfactoryv11.Batch{}, err
	}
	for _, item := range items {
		binding, bindingErr := projects.BatchFactoryBinding(ctx, owner, item.ID)
		if errors.Is(bindingErr, agentworkspace.ErrBatchFactoryBindingNotFound) || binding.ProjectID != strings.TrimSpace(projectID) {
			continue
		}
		if bindingErr != nil {
			return agentworkspace.BatchFactoryBinding{}, batchfactoryv11.Batch{}, bindingErr
		}
		batch, getErr := batches.GetBatch(ctx, owner, item.ID)
		if getErr != nil {
			return agentworkspace.BatchFactoryBinding{}, batchfactoryv11.Batch{}, getErr
		}
		return binding, batch, nil
	}
	return agentworkspace.BatchFactoryBinding{}, batchfactoryv11.Batch{}, agentworkspace.ErrBatchFactoryBindingNotFound
}

func batchFactoryBookExists(batch batchfactoryv11.Batch, bookID string) bool {
	for _, book := range batch.Books {
		if book.ID == strings.TrimSpace(bookID) {
			return true
		}
	}
	return false
}

func agentBatchFactoryNodeOrigins(ctx context.Context, projects agentworkspace.Store, owner string, binding agentworkspace.BatchFactoryBinding, batch batchfactoryv11.Batch, nodeID string) (string, []agentworkspace.BatchFactoryNodeOrigin, bool, error) {
	nodeID = strings.TrimSpace(nodeID)
	for _, book := range batch.Books {
		origins, err := projects.BatchFactoryNodeOrigins(ctx, owner, binding.ProjectID, book.ID)
		if err != nil {
			return "", nil, false, err
		}
		for _, origin := range origins {
			if origin.NodeID == nodeID && origin.BatchID == binding.BatchID && origin.ProjectID == binding.ProjectID {
				return book.ID, origins, true, nil
			}
		}
	}
	return "", nil, false, nil
}

func removeAgentBatchFactoryCanvasNode(canvas agentworkspace.Canvas, nodeID string) agentworkspace.Canvas {
	nodeID = strings.TrimSpace(nodeID)
	nodes := make([]agentworkspace.Node, 0, len(canvas.Nodes))
	for _, node := range canvas.Nodes {
		if node.ID != nodeID {
			nodes = append(nodes, node)
		}
	}
	edges := make([]agentworkspace.Edge, 0, len(canvas.Edges))
	for _, edge := range canvas.Edges {
		if edge.Source != nodeID && edge.Target != nodeID {
			edges = append(edges, edge)
		}
	}
	canvas.Nodes, canvas.Edges = nodes, edges
	return canvas
}

func writeAgentBatchFactoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, batchfactoryv11.ErrNotFound), errors.Is(err, agentworkspace.ErrBatchFactoryBindingNotFound), errors.Is(err, agentworkspace.ErrProjectNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "资源不存在"})
	case errors.Is(err, agentworkspace.ErrBatchFactoryBindingConflict), errors.Is(err, agentworkspace.ErrRevisionConflict):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "资源已更新，请重新加载"})
	case errors.Is(err, agentworkspace.ErrBatchFactoryBindingInvalid), errors.Is(err, agentworkspace.ErrBatchFactoryNodeOriginInvalid), errors.Is(err, agentworkspace.ErrCanvasInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求数据无效"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Agent 数据保存失败"})
	}
}
