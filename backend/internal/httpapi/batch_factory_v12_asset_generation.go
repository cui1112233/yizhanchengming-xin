package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

const assetImageHTTPFastPathWait = 500 * time.Millisecond
const assetImageDetachedGenerationTimeout = 15 * time.Minute

type assetImageGenerationTask struct {
	ID         string
	Owner      string
	BatchID    string
	BookID     string
	AssetIDs   []string
	Status     string
	Error      string
	ImageCount int
	UpdatedAt  time.Time
}

// assetImageGenerationTasks makes detached provider work observable to the
// browser. Before this registry existed, the POST request returned after the
// fast-path window while a later provider failure was only written to stdout;
// the UI therefore kept polling empty image lists until its local timeout.
type assetImageGenerationTasks struct {
	mu    sync.RWMutex
	tasks map[string]assetImageGenerationTask
	path  string
}

const (
	assetImageTaskStoreDefaultPath = "data/system/asset-image-generation-tasks.json"
	assetImageTaskStoreRetention   = 200
)

func newAssetImageGenerationTasks() *assetImageGenerationTasks {
	path := strings.TrimSpace(os.Getenv("QIANTIE_ASSET_IMAGE_TASKS_PATH"))
	if path == "" {
		path = assetImageTaskStoreDefaultPath
	}
	tasks := &assetImageGenerationTasks{tasks: map[string]assetImageGenerationTask{}, path: path}
	tasks.load()
	return tasks
}

// load 在进程启动时恢复任务记录。服务重启时仍在生成中的任务已经不可能继续，
// 标记为失败并给出明确原因，避免前端只能拿到 404 而一直空轮询。
func (tasks *assetImageGenerationTasks) load() {
	content, err := os.ReadFile(tasks.path)
	if err != nil {
		return
	}
	var stored struct {
		Tasks []assetImageGenerationTask `json:"tasks"`
	}
	if err := json.Unmarshal(content, &stored); err != nil {
		log.Printf("load asset image generation tasks failed path=%s: %v", tasks.path, err)
		return
	}
	changed := false
	for _, task := range stored.Tasks {
		if task.ID == "" {
			continue
		}
		if task.Status == "processing" {
			task.Status = "failed"
			task.Error = "服务重启导致任务中断，请重新生成"
			task.UpdatedAt = time.Now().UTC()
			changed = true
		}
		task.AssetIDs = append([]string(nil), task.AssetIDs...)
		tasks.tasks[task.ID] = task
	}
	if changed {
		tasks.save()
	}
}

func (tasks *assetImageGenerationTasks) put(task assetImageGenerationTask) {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	task.AssetIDs = append([]string(nil), task.AssetIDs...)
	tasks.tasks[task.ID] = task
	tasks.pruneLocked()
	tasks.saveLocked()
}

// pruneLocked 只保留最近的任务记录，避免状态文件无限增长。
func (tasks *assetImageGenerationTasks) pruneLocked() {
	if len(tasks.tasks) <= assetImageTaskStoreRetention {
		return
	}
	oldest := ""
	var oldestAt time.Time
	for id, task := range tasks.tasks {
		if oldest == "" || task.UpdatedAt.Before(oldestAt) {
			oldest, oldestAt = id, task.UpdatedAt
		}
	}
	if oldest != "" {
		delete(tasks.tasks, oldest)
	}
}

func (tasks *assetImageGenerationTasks) save() {
	tasks.mu.RLock()
	defer tasks.mu.RUnlock()
	tasks.saveLocked()
}

func (tasks *assetImageGenerationTasks) saveLocked() {
	if strings.TrimSpace(tasks.path) == "" {
		return
	}
	list := make([]assetImageGenerationTask, 0, len(tasks.tasks))
	for _, task := range tasks.tasks {
		list = append(list, task)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].UpdatedAt.After(list[j].UpdatedAt) })
	if len(list) > assetImageTaskStoreRetention {
		list = list[:assetImageTaskStoreRetention]
	}
	content, err := json.MarshalIndent(map[string]any{"tasks": list}, "", "  ")
	if err != nil {
		log.Printf("encode asset image generation tasks failed: %v", err)
		return
	}
	if dir := filepath.Dir(tasks.path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("create asset image task dir failed path=%s: %v", dir, err)
			return
		}
	}
	tempPath := tasks.path + ".tmp"
	if err := os.WriteFile(tempPath, content, 0o644); err != nil {
		log.Printf("write asset image generation tasks failed path=%s: %v", tasks.path, err)
		return
	}
	if err := os.Rename(tempPath, tasks.path); err != nil {
		log.Printf("commit asset image generation tasks failed path=%s: %v", tasks.path, err)
	}
}

func (tasks *assetImageGenerationTasks) get(id string) (assetImageGenerationTask, bool) {
	tasks.mu.RLock()
	defer tasks.mu.RUnlock()
	task, ok := tasks.tasks[id]
	task.AssetIDs = append([]string(nil), task.AssetIDs...)
	return task, ok
}

func (tasks *assetImageGenerationTasks) finish(id string, images []batchfactoryv11.BookAssetImage, generationErr error) {
	tasks.mu.Lock()
	defer tasks.mu.Unlock()
	task, ok := tasks.tasks[id]
	if !ok {
		return
	}
	task.ImageCount = len(images)
	task.UpdatedAt = time.Now().UTC()
	if len(images) == 0 {
		task.Status = "failed"
		if generationErr == nil {
			task.Error = "资产图片生成结束，但没有返回可保存的图片"
		} else {
			task.Error = generationErr.Error()
		}
	} else {
		task.Status = "succeeded"
		if generationErr != nil {
			task.Error = generationErr.Error()
		}
	}
	tasks.tasks[id] = task
	tasks.saveLocked()
}

func assetImageGenerationTaskResponse(task assetImageGenerationTask) map[string]any {
	return map[string]any{
		"taskId":     task.ID,
		"status":     task.Status,
		"batchId":    task.BatchID,
		"bookId":     task.BookID,
		"assetIds":   task.AssetIDs,
		"imageCount": task.ImageCount,
		"error":      task.Error,
		"updatedAt":  task.UpdatedAt,
	}
}

func assetImageGenerationContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), assetImageDetachedGenerationTimeout)
}

// registerV12AssetGenerationRoute keeps the asset image action on the same Go
// store/provider boundary as the Shuihuo project page. The old Node handler
// used to receive this one action even when every other batch operation was on
// Go, which made the asset step the last hidden runtime split.
func registerV12AssetGenerationRoute(mux *http.ServeMux, store batchfactoryv11.Store, files *localartifact.Store, localExecutors *localexecutor.Service, presetCatalog *prompt.Catalog, catalogStore novelfetchworkshop.Store, rewrite comicPromptRewriter) {
	if mux == nil || store == nil || files == nil {
		return
	}
	tasks := newAssetImageGenerationTasks()
	mux.HandleFunc("GET /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/images/generate/{taskId}", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		task, found := tasks.get(r.PathValue("taskId"))
		if !found || task.Owner != owner || task.BatchID != r.PathValue("batchId") || task.BookID != r.PathValue("bookId") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "资产图片生成任务不存在"})
			return
		}
		writeJSON(w, http.StatusOK, assetImageGenerationTaskResponse(task))
	})
	mux.HandleFunc("POST /api/batch-factory/v12/batches/{batchId}/books/{bookId}/assets/images/generate", func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId"))
		if err != nil {
			writeStoreError(w, err)
			return
		}
		bookID := r.PathValue("bookId")
		var selected *batchfactoryv11.Book
		for index := range batch.Books {
			if batch.Books[index].ID == bookID {
				selected = &batch.Books[index]
				break
			}
		}
		if selected == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "小说不存在"})
			return
		}
		var input struct {
			AssetIDs      []string `json:"assetIds"`
			ModelID       string   `json:"modelId"`
			Provider      string   `json:"provider"`
			AspectRatio   string   `json:"imageAspectRatio"`
			UseThreeView  *bool    `json:"useThreeView"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		selectedBatch := batch
		selectedBatch.Books = []batchfactoryv11.Book{*selected}
		batchID := r.PathValue("batchId")
		assetIDs := append([]string(nil), input.AssetIDs...)
		modelID, aspectRatio, provider := input.ModelID, input.AspectRatio, input.Provider
		// 未传时默认开启三视图，保持既有漫画资产出图行为。
		threeView := true
		if input.UseThreeView != nil {
			threeView = *input.UseThreeView
		}
		taskID, idErr := localexecutor.NewArtifactID()
		if idErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "无法创建资产图片生成任务"})
			return
		}
		tasks.put(assetImageGenerationTask{ID: taskID, Owner: owner, BatchID: batchID, BookID: bookID, AssetIDs: assetIDs, Status: "processing", UpdatedAt: time.Now().UTC()})
		resultCh := make(chan struct {
			images []batchfactoryv11.BookAssetImage
			err    error
		}, 1)
		go func() {
			generationCtx, cancel := assetImageGenerationContext(context.Background())
			defer cancel()
			generationCtx = withShuihuoAssetThreeView(generationCtx, threeView)
			images, generationErr := generateShuihuoAssetImagesWithRepair(generationCtx, store, files, owner, selectedBatch, assetIDs, modelID, aspectRatio, provider, localExecutors, presetCatalog, rewrite, catalogStore)
			if generationErr != nil {
				log.Printf("batch factory asset image generation failed batch=%s book=%s assets=%v model=%s provider=%s: %v", batchID, bookID, assetIDs, modelID, provider, generationErr)
			}
			tasks.finish(taskID, images, generationErr)
			resultCh <- struct {
				images []batchfactoryv11.BookAssetImage
				err    error
			}{images: images, err: generationErr}
		}()
		select {
		case result := <-resultCh:
			for index := range result.images {
				result.images[index] = bookAssetImageResponse(result.images[index], batch.ID, bookID, result.images[index].AssetID)
			}
			if result.err != nil && len(result.images) == 0 {
				writeStoreError(w, result.err)
				return
			}
			status := http.StatusCreated
			response := map[string]any{"images": result.images}
			if result.err != nil {
				response["warning"] = result.err.Error()
			}
			writeJSON(w, status, response)
		case <-time.After(assetImageHTTPFastPathWait):
			task, _ := tasks.get(taskID)
			response := assetImageGenerationTaskResponse(task)
			response["message"] = "资产图已提交，生成完成后会自动回填。"
			writeJSON(w, http.StatusAccepted, response)
		}
	})
}
