package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
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
		if err == nil {
			binding, err = ensureLiveAgentBatchFactoryProject(req.Context(), projects, binding, clock)
		}
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
		response, err := syncAgentBatchFactoryBookProjection(req.Context(), projects, batches, projection, owner, projectID, req.PathValue("bookId"))
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, response)
	})

	handle("GET /api/agent/tasks/{id}/batch-factory/nodes/{nodeId}/media", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID, nodeID := strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("nodeId"))
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
		node, found := agentCanvasNodeByID(canvas, nodeID)
		if !found {
			writeAgentBatchFactoryError(w, batchfactoryv11.ErrNotFound)
			return
		}
		origin, found := agentBatchFactoryNodeOrigin(origins, nodeID)
		if !found || origin.BookID != bookID || origin.BatchID != binding.BatchID {
			writeAgentBatchFactoryError(w, agentworkspace.ErrBatchFactoryBindingNotFound)
			return
		}
		mediaURL, err := agentBatchFactoryProjectionMediaURL(req.Context(), batches, owner, binding.BatchID, bookID, origin, node)
		if err != nil {
			writeAgentBatchFactoryError(w, err)
			return
		}
		http.Redirect(w, req, mediaURL, http.StatusFound)
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

func ensureLiveAgentBatchFactoryProject(ctx context.Context, projects agentworkspace.Store, binding agentworkspace.BatchFactoryBinding, clock func() time.Time) (agentworkspace.BatchFactoryBinding, error) {
	if _, err := projects.Project(ctx, binding.Owner, binding.ProjectID); err == nil {
		return binding, nil
	} else if !errors.Is(err, agentworkspace.ErrProjectNotFound) {
		return agentworkspace.BatchFactoryBinding{}, err
	}
	current := binding
	for range 16 {
		projectID := agentBatchFactoryReplacementProjectID(current.Owner, current.BatchID, current.ProjectID)
		now := clock().UTC()
		project := agentworkspace.Project{ID: projectID, Title: "批量工厂项目", TitleMode: "automatic", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
		createErr := projects.CreateProject(ctx, current.Owner, project)
		if errors.Is(createErr, agentworkspace.ErrProjectExists) {
			if _, liveErr := projects.Project(ctx, current.Owner, projectID); errors.Is(liveErr, agentworkspace.ErrProjectNotFound) {
				current.ProjectID = projectID
				continue
			} else if liveErr != nil {
				return agentworkspace.BatchFactoryBinding{}, liveErr
			}
		} else if createErr != nil {
			return agentworkspace.BatchFactoryBinding{}, createErr
		}
		rebound, rebindErr := projects.RebindBatchFactoryProject(ctx, current.Owner, current.BatchID, binding.ProjectID, projectID, now)
		if rebindErr == nil {
			return rebound, nil
		}
		if !errors.Is(rebindErr, agentworkspace.ErrBatchFactoryBindingConflict) {
			return agentworkspace.BatchFactoryBinding{}, rebindErr
		}
		latest, readErr := projects.BatchFactoryBinding(ctx, current.Owner, current.BatchID)
		if readErr != nil {
			return agentworkspace.BatchFactoryBinding{}, readErr
		}
		if _, liveErr := projects.Project(ctx, current.Owner, latest.ProjectID); liveErr == nil {
			return latest, nil
		}
		binding, current = latest, latest
	}
	return agentworkspace.BatchFactoryBinding{}, agentworkspace.ErrBatchFactoryBindingConflict
}

// syncAgentBatchFactoryBookProjection is the only Agent-side projection write.
// Media-producing actions remain on Batch Factory endpoints; after they finish
// the canvas receives the authoritative records through this one path.
func syncAgentBatchFactoryBookProjection(ctx context.Context, projects agentworkspace.Store, batches agentBatchFactoryBridgeStore, projection *agentworkspace.BatchFactoryProjectionService, owner, projectID, bookID string) (agentBatchFactoryBindingResponse, error) {
	binding, batch, err := agentBatchFactoryBindingForProject(ctx, projects, batches, owner, projectID)
	if err != nil {
		return agentBatchFactoryBindingResponse{}, err
	}
	bookID = strings.TrimSpace(bookID)
	if !batchFactoryBookExists(batch, bookID) {
		return agentBatchFactoryBindingResponse{}, batchfactoryv11.ErrNotFound
	}
	canvas, err := projection.SyncBatchFactoryBook(ctx, owner, binding, bookID)
	if err != nil {
		return agentBatchFactoryBindingResponse{}, err
	}
	return bindingResponse(binding, bookID, canvas.Revision, "synchronized"), nil
}

func bindingResponse(binding agentworkspace.BatchFactoryBinding, bookID string, version int64, state string) agentBatchFactoryBindingResponse {
	return agentBatchFactoryBindingResponse{ProjectID: binding.ProjectID, BatchID: binding.BatchID, BookID: strings.TrimSpace(bookID), Source: "batch_factory", Version: version, State: state}
}

func agentBatchFactoryProjectID(owner, batchID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(batchID)))
	return "agent-bf-" + hex.EncodeToString(digest[:12])
}

func agentBatchFactoryReplacementProjectID(owner, batchID, previousProjectID string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(batchID) + "\x00" + strings.TrimSpace(previousProjectID)))
	return agentBatchFactoryProjectID(owner, batchID) + "-" + hex.EncodeToString(digest[:6])
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

func agentBatchFactoryNodeOrigin(origins []agentworkspace.BatchFactoryNodeOrigin, nodeID string) (agentworkspace.BatchFactoryNodeOrigin, bool) {
	for _, origin := range origins {
		if origin.NodeID == nodeID {
			return origin, true
		}
	}
	return agentworkspace.BatchFactoryNodeOrigin{}, false
}

func agentBatchFactoryProjectionMediaURL(ctx context.Context, batches agentBatchFactoryBridgeStore, owner, batchID, bookID string, origin agentworkspace.BatchFactoryNodeOrigin, node agentworkspace.Node) (string, error) {
	mediaID := strings.TrimSpace(agentNodeAssetRefID(node.AssetRef))
	switch origin.ResourceType {
	case "book_asset":
		images, err := batches.ListBookAssetImages(ctx, owner, batchID, bookID, origin.ResourceID)
		if err != nil {
			return "", err
		}
		for _, image := range images {
			if image.ID != mediaID {
				continue
			}
			if mediaURL := strings.TrimSpace(image.URL); mediaURL != "" {
				return mediaURL, nil
			}
			if mediaURL, ok := agentBatchFactoryLocalImageArtifactURL(image); ok {
				return mediaURL, nil
			}
		}
	case "video":
		jobs, err := batches.ListProductionJobs(ctx, owner, batchID)
		if err != nil {
			return "", err
		}
		for _, job := range jobs {
			if job.BookID != bookID {
				continue
			}
			for _, task := range job.Tasks {
				if task.VideoID == origin.ResourceID && task.ID == mediaID && strings.TrimSpace(task.MediaURL) != "" {
					return strings.TrimSpace(task.MediaURL), nil
				}
			}
		}
	case "merge":
		jobs, err := batches.ListMergeJobs(ctx, owner, batchID)
		if err != nil {
			return "", err
		}
		for _, job := range jobs {
			if job.BookID == bookID && job.ID == mediaID && strings.TrimSpace(job.OutputURL) != "" {
				return strings.TrimSpace(job.OutputURL), nil
			}
		}
	}
	return "", batchfactoryv11.ErrNotFound
}

// agentBatchFactoryLocalImageArtifactURL keeps locally generated Batch Factory
// images behind the established owner-scoped artifact reader. The canvas never
// exposes the storage reference itself, and the target reader rechecks owner.
func agentBatchFactoryLocalImageArtifactURL(image batchfactoryv11.BookAssetImage) (string, bool) {
	if image.Source != batchfactoryv11.VideoProviderDoubaoLocal {
		return "", false
	}
	storageRef := strings.TrimSpace(image.StorageRef)
	if storageRef == "" || filepath.Base(storageRef) != storageRef {
		return "", false
	}
	artifactID := strings.TrimSuffix(storageRef, filepath.Ext(storageRef))
	if artifactID == "" {
		return "", false
	}
	return "/api/shuihuo-production/local-executor-artifacts/" + url.PathEscape(artifactID), true
}

func agentNodeAssetRefID(ref map[string]any) string {
	value, _ := ref["id"].(string)
	return value
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
