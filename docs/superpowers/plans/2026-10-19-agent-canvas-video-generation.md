# Agent 画布视频生成闭环 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Agent 画布的图片参考、视频生成、私有素材保存和人工确认回填形成可刷新恢复的真实闭环。

**Architecture:** 在 agentworkspace 增加项目范围的视频任务事实记录和原子确认回填；在 httpapi 提供受登录保护的 Agent 视频 API，并复用现有视频提供方和本地执行器能力。参考图由后端生成短时 TOS 读取地址，结果由后端保存为项目私有视频素材；画布只接收项目素材引用。

**Tech Stack:** Go、MySQL 8.4、Redis、TOS、React、Ant Design、React Flow、@xyflow/react。

## Global Constraints

- 只在本机 V88 实施；不得发布公网。
- Go 后端只放入 backend；React 前端只放入 frontend；打包仍走 Go embed。
- MySQL 已执行迁移不可修改；仅追加版本大于 8800005 的幂等迁移。
- Redis 仅用于任务锁和短时进度，不作为任务或素材的事实记录。
- TOS 密钥只保留在后端；浏览器只能得到项目私有内容地址，模型只得到短时预签名读取地址。
- 不复制参考产品的源码、品牌、账号、云同步、收费或远端接口。
- 当前工作区已经有用户改动；不重置、清理、覆盖或提交无关文件。未获得明确提交授权前，只保留本地修改和测试记录。

---

## File Structure

| 文件 | 责任 |
|---|---|
| backend/internal/agentworkspace/video_jobs.go | 视频任务状态、参数快照、输入校验和存储接口。 |
| backend/internal/agentworkspace/mysql_video_jobs.go | MySQL 视频任务读写、幂等提交和确认回填事务。 |
| backend/internal/agentworkspace/memory_video_jobs.go | 路由测试用内存实现。 |
| backend/internal/agentworkspace/tos_assets.go | 项目私有素材的短时预签名读取地址。 |
| backend/internal/storage/agent_workspace_schema.go | 追加 agent_video_jobs 迁移。 |
| backend/internal/httpapi/agent_video_jobs.go | Agent 视频任务 API、提交/轮询/保存/确认协调。 |
| backend/internal/httpapi/script_video_service.go | 从兼容路由抽出的共享视频提供方提交与查询服务。 |
| backend/internal/httpapi/router.go | 注入并注册 Agent 视频任务路由。 |
| frontend/src/shared/api/agent.js | Agent 视频任务 API 客户端。 |
| frontend/src/user/pages/agent-canvas-video-jobs.js | 任务状态转换、参考图收集与轮询纯函数。 |
| frontend/src/user/components/AgentCanvas.jsx | 图片/视频预览、参数面板和操作按钮。 |

### Task 1: 追加视频任务存储和迁移

**Files:**
- Create: backend/internal/agentworkspace/video_jobs.go
- Create: backend/internal/agentworkspace/memory_video_jobs.go
- Create: backend/internal/agentworkspace/mysql_video_jobs.go
- Create: backend/internal/agentworkspace/video_jobs_test.go
- Modify: backend/internal/agentworkspace/store.go
- Modify: backend/internal/agentworkspace/mysql_store.go
- Modify: backend/internal/storage/agent_workspace_schema.go
- Modify: backend/internal/storage/agent_workspace_schema_test.go
- Modify: backend/internal/storage/migrations_mysql84_test.go

**Interfaces:**
- Consumes: Project, Canvas, Node, Asset 和既有 Store。
- Produces: VideoJobStore 的 CreateVideoJob、VideoJob、ListVideoJobs、UpdateVideoJob、ConfirmVideoJob 和 RetryVideoJob。

- [ ] **Step 1: 写失败测试，锁定状态和项目隔离**

    func TestVideoJobStoreRejectsCrossProjectAndDuplicateIdempotency(t *testing.T) {
        store := NewMemoryStore()
        mustCreateProject(t, store, "alice", "agent-a")
        mustCreateProject(t, store, "alice", "agent-b")
        _, err := store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{
            ProjectID: "agent-a", NodeID: "video-1", NodeVersion: 1, IdempotencyKey: "idem-1",
        })
        if err != nil { t.Fatal(err) }
        _, err = store.CreateVideoJob(ctx, "alice", CreateVideoJobInput{
            ProjectID: "agent-b", NodeID: "video-1", NodeVersion: 1, IdempotencyKey: "idem-1",
        })
        if !errors.Is(err, ErrVideoJobExists) { t.Fatalf("err=%v", err) }
    }

- [ ] **Step 2: 运行测试，确认缺少视频任务接口时失败**

Run (from backend): go test ./internal/agentworkspace -run TestVideoJobStoreRejectsCrossProjectAndDuplicateIdempotency -count=1  
Expected: FAIL，提示 CreateVideoJob 或相关类型不存在。

- [ ] **Step 3: 写最小领域类型和内存存储**

    type VideoJobState string
    const (
        VideoJobPreparing VideoJobState = "preparing"
        VideoJobQueued VideoJobState = "queued"
        VideoJobRunning VideoJobState = "running"
        VideoJobResultReady VideoJobState = "result_ready"
        VideoJobAwaitingConfirmation VideoJobState = "awaiting_confirmation"
        VideoJobApplied VideoJobState = "applied"
        VideoJobFailed VideoJobState = "failed"
        VideoJobCanceled VideoJobState = "canceled"
        VideoJobAbandoned VideoJobState = "abandoned"
    )

    type VideoJob struct {
        ID, ProjectID, OwnerUsername, NodeID, ModelKey, ProviderTaskID string
        NodeVersion int64
        State VideoJobState
        Settings map[string]any
        ReferenceAssetIDs []string
        ResultAssetID, ErrorMessage, IdempotencyKey string
        CreatedAt, UpdatedAt time.Time
    }

    type VideoJobStore interface {
        CreateVideoJob(context.Context, string, CreateVideoJobInput) (VideoJob, error)
        VideoJob(context.Context, string, string, string) (VideoJob, error)
        ListVideoJobs(context.Context, string, string) ([]VideoJob, error)
        UpdateVideoJob(context.Context, string, VideoJob, time.Time) (VideoJob, error)
        ConfirmVideoJob(context.Context, string, string, string, time.Time) (VideoJob, error)
        RetryVideoJob(context.Context, string, string, string, time.Time) (VideoJob, error)
    }

CreateVideoJob 只接受非空项目、视频节点、正版本和幂等键；同一账号的幂等键只能创建一次。内存实现与 MySQL 实现都返回深拷贝，避免调用方改写存储中的 map 或切片。

- [ ] **Step 4: 追加迁移 8800006 和 MySQL 实现**

    CREATE TABLE IF NOT EXISTS agent_video_jobs (
      id VARCHAR(96) NOT NULL PRIMARY KEY,
      project_id VARCHAR(96) NOT NULL,
      owner_username VARCHAR(191) NOT NULL,
      node_id VARCHAR(96) NOT NULL,
      node_version BIGINT NOT NULL,
      state VARCHAR(32) NOT NULL,
      model_key VARCHAR(96) NOT NULL,
      provider_task_id VARCHAR(191) NULL,
      settings_json JSON NOT NULL,
      reference_asset_ids_json JSON NOT NULL,
      result_asset_id VARCHAR(96) NULL,
      error_message TEXT NULL,
      idempotency_key VARCHAR(191) NOT NULL,
      created_at DATETIME(6) NOT NULL,
      updated_at DATETIME(6) NOT NULL,
      UNIQUE KEY uq_agent_video_jobs_owner_idempotency (owner_username, idempotency_key),
      KEY idx_agent_video_jobs_project_node_updated (owner_username, project_id, node_id, updated_at)
    ) ENGINE=InnoDB

ConfirmVideoJob 在一个事务内锁定任务和画布节点，验证任务为 awaiting_confirmation、节点版本仍等于提交版本，再更新节点 asset_ref_json、状态、版本和 agent_node_versions，最后将任务改为 applied。冲突时返回 ErrVideoJobNodeChanged，绝不覆盖用户后续编辑。

- [ ] **Step 5: 运行存储和迁移测试**

Run (from backend): go test ./internal/agentworkspace ./internal/storage -count=1  
Expected: PASS。若配置了 QIANTIE_TEST_MYSQL84_DSN，再运行 go test -tags mysql84 ./internal/storage -run AgentWorkspace -count=1，确认空库和重复迁移均通过。

- [ ] **Step 6: 检查范围，不提交**

Run: git diff --check -- backend/internal/agentworkspace backend/internal/storage  
Expected: 无输出。保留本地改动，不执行 git commit。

### Task 2: 为私有素材提供安全参考地址和结果入库能力

**Files:**
- Modify: backend/internal/agentworkspace/assets.go
- Modify: backend/internal/agentworkspace/tos_assets.go
- Modify: backend/internal/agentworkspace/asset_store_test.go
- Create: backend/internal/agentworkspace/tos_assets_test.go
- Create: backend/internal/httpapi/agent_video_asset_ingest.go
- Create: backend/internal/httpapi/agent_video_asset_ingest_test.go

**Interfaces:**
- Consumes: agentworkspace.AssetStore、AssetObjectStore、localexecutor.Service、localartifact.Store。
- Produces: ReferenceURLProvider.ReferenceURL 和 agentVideoAssetIngestor.SaveVideo。

- [ ] **Step 1: 写失败测试，明确不把私有 API 地址交给远端模型**

    func TestReferenceURLProviderProducesShortLivedHTTPSURL(t *testing.T) {
        store := fakeReferenceURLProvider{url: "https://tos.example/video.png?X-Tos-Signature=x"}
        url, err := store.ReferenceURL(context.Background(), asset, 15*time.Minute)
        if err != nil { t.Fatal(err) }
        if !strings.HasPrefix(url, "https://") || strings.Contains(url, "/api/agent/") {
            t.Fatalf("url=%q", url)
        }
    }

- [ ] **Step 2: 运行测试，确认接口尚不存在时失败**

Run (from backend): go test ./internal/agentworkspace ./internal/httpapi -run 'TestReferenceURLProviderProducesShortLivedHTTPSURL|TestAgentVideoAssetIngest' -count=1  
Expected: FAIL，提示 ReferenceURL 或 agentVideoAssetIngestor 未定义。

- [ ] **Step 3: 实现短时预签名读取地址**

    type ReferenceURLProvider interface {
        ReferenceURL(context.Context, Asset, time.Duration) (string, error)
    }

TOSAssetObjectStore.ReferenceURL 使用 TOS PreSignedURL、HTTP GET 和 15 分钟有效期。非 TOS 的内存对象存储不假装生成 URL，而是返回 ErrReferenceURLUnavailable；路由将其转换为“当前素材存储无法提供模型参考图”的明确错误。

- [ ] **Step 4: 实现后端结果保存器**

    type agentVideoAssetIngestor struct {
        assets agentworkspace.AssetStore
        objects agentworkspace.AssetObjectStore
        fetch func(context.Context, string) (io.ReadCloser, string, error)
        now func() time.Time
    }

保存器只接受 video/mp4，生成新的 asset-* ID 和 AgentAssetObjectKey，先写对象存储，再创建 agent_assets 记录；数据库创建失败时删除刚写入的对象。远端结果 URL 只允许 HTTPS；本地执行器结果使用 localexecutor.Service.GetArtifact 加 localartifact.Store.Open 读取，不能让后端伪造浏览器 Cookie。

- [ ] **Step 5: 运行素材安全测试**

Run (from backend): go test ./internal/agentworkspace ./internal/httpapi -run 'TestReferenceURL|TestAgentVideoAssetIngest' -count=1  
Expected: PASS，且测试覆盖对象写入失败、数据库写入失败回滚、非 HTTPS 远端结果拒绝和本地执行器结果读取。

- [ ] **Step 6: 检查范围，不提交**

Run: git diff --check -- backend/internal/agentworkspace backend/internal/httpapi  
Expected: 无输出。保留本地改动，不执行 git commit。

### Task 3: 提取共享视频提供方服务并实现 Agent 任务协调器

**Files:**
- Create: backend/internal/httpapi/script_video_service.go
- Create: backend/internal/httpapi/agent_video_jobs.go
- Create: backend/internal/httpapi/agent_video_jobs_test.go
- Modify: backend/internal/httpapi/script_video_routes.go
- Modify: backend/internal/httpapi/router.go

**Interfaces:**
- Consumes: Task 1 的 VideoJobStore、Task 2 的 ReferenceURLProvider/保存器、scriptVideoModelSpec、localexecutor.Service。
- Produces: RegisterAgentVideoJobRoutes 和 agentVideoCoordinator.Submit/Poll/Confirm/Cancel/Retry。

- [ ] **Step 1: 写失败路由测试，覆盖提交、状态和确认**

    func TestAgentVideoJobRoutesRequireOwnedVideoNodeAndConfirmResult(t *testing.T) {
        api := newAgentVideoJobTestRouter(t)
        created := signedJSONRequest(t, api, now, "alice", http.MethodPost,
            "/api/agent/tasks/agent-1/video-jobs", map[string]any{
                "nodeId": "video-1", "idempotencyKey": "key-1",
            })
        if created.Code != http.StatusAccepted {
            t.Fatalf("status=%d body=%s", created.Code, created.Body.String())
        }
        confirmed := signedJSONRequest(t, api, now, "alice", http.MethodPost,
            "/api/agent/tasks/agent-1/video-jobs/job-1/confirm", nil)
        if confirmed.Code != http.StatusOK { t.Fatalf("status=%d", confirmed.Code) }
    }

- [ ] **Step 2: 运行测试，确认新路由尚未注册时失败**

Run (from backend): go test ./internal/httpapi -run TestAgentVideoJobRoutesRequireOwnedVideoNodeAndConfirmResult -count=1  
Expected: FAIL，状态为 404 或缺少测试辅助对象。

- [ ] **Step 3: 从旧兼容路由抽出共享服务**

    type scriptVideoService struct {
        resolver scriptVideoConfigResolver
        client *http.Client
        localExecutor *localexecutor.Service
    }

    func (s *scriptVideoService) Submit(ctx context.Context, owner string, input scriptVideoLocalInput) (scriptVideoSubmission, error)
    func (s *scriptVideoService) Poll(ctx context.Context, owner, taskID string) (scriptVideoStatus, error)
    func (s *scriptVideoService) Cancel(ctx context.Context, owner, taskID string) error

保留 /api/script-video 的 JSON 合约和已有测试。Submit 必须使用既有 scriptVideoModelSpec，在模型不支持参考图时返回可识别的 UNSUPPORTED_REFERENCE_IMAGES 领域错误，而不是直接删掉图片。

- [ ] **Step 4: 实现 Agent 协调器和受保护路由**

    func RegisterAgentVideoJobRoutes(root *http.ServeMux, auth BridgeAuth, coordinator *agentVideoCoordinator)

提交时读取当前项目画布，验证 nodeId 是视频节点；只收集直接指向它、且带有当前项目图片素材的连线。协调器保存节点版本和参数快照，生成预签名图片 URL 后调用共享视频服务。轮询成功后调用 Task 2 保存器，状态顺序只能前进到 awaiting_confirmation。确认和重试使用 Task 1 的事务接口；取消同时尝试取消本地执行器任务，远端不支持取消时只把 Agent 任务标为 abandoned 并保留原因。

- [ ] **Step 5: 运行后端路由回归**

Run (from backend): go test ./internal/httpapi -run 'TestAgentVideoJob|TestScriptVideoRoutes' -count=1  
Expected: PASS，旧 /api/script-video 测试继续通过；新测试覆盖跨账号、非视频节点、没有参考图权限、参考图不支持确认、轮询成功保存、确认冲突、取消和重试。

- [ ] **Step 6: 检查范围，不提交**

Run: git diff --check -- backend/internal/httpapi  
Expected: 无输出。保留本地改动，不执行 git commit。

### Task 4: 前端 API、任务状态和视频节点交互

**Files:**
- Modify: frontend/src/shared/api/agent.js
- Create: frontend/src/user/pages/agent-canvas-video-jobs.js
- Create: frontend/src/user/pages/agent-canvas-video-jobs.test.js
- Modify: frontend/src/user/components/AgentCanvas.jsx
- Modify: frontend/src/user/pages/agent-canvas-context-menu.js
- Modify: frontend/src/user/pages/agent-canvas-context-menu.test.js

**Interfaces:**
- Consumes: Task 3 的 API 返回 { job }，其中 job.state、job.resultAsset、job.errorMessage 和 job.nodeId 为前端事实来源。
- Produces: collectVideoReferences、videoNodeViewState、watchVideoJob 和节点参数面板。

- [ ] **Step 1: 写失败的纯函数测试**

    test("collectVideoReferences keeps only directly connected confirmed image assets", () => {
      const refs = collectVideoReferences({ nodes, edges, videoNodeId: "video-1" });
      assert.deepEqual(refs.map(item => item.id), ["image-1"]);
    })

    test("videoNodeViewState exposes confirmation instead of auto-applying result", () => {
      assert.equal(videoNodeViewState({ state: "awaiting_confirmation" }).canConfirm, true);
      assert.equal(videoNodeViewState({ state: "awaiting_confirmation" }).canGenerate, false);
    })

- [ ] **Step 2: 运行测试，确认模块不存在时失败**

Run: node --test frontend/src/user/pages/agent-canvas-video-jobs.test.js  
Expected: FAIL，提示模块或导出不存在。

- [ ] **Step 3: 增加 Agent 视频 API 客户端和状态模块**

    export const createAgentVideoJob = (taskId, payload) =>
      apiRequest("/api/agent/tasks/" + encodeURIComponent(taskId) + "/video-jobs", {
        method: "POST", body: JSON.stringify(payload)
      })
    export const getAgentVideoJob = (taskId, jobId) =>
      apiRequest("/api/agent/tasks/" + encodeURIComponent(taskId) + "/video-jobs/" + encodeURIComponent(jobId))
    export const confirmAgentVideoJob = (taskId, jobId) =>
      apiRequest("/api/agent/tasks/" + encodeURIComponent(taskId) + "/video-jobs/" + encodeURIComponent(jobId) + "/confirm", {
        method: "POST"
      })

collectVideoReferences 必须按边的创建顺序读取直接图片输入，过滤未确认素材、错误项目素材和重复素材。轮询只针对 preparing、queued、running、result_ready 状态；组件卸载时必须停止定时器。

- [ ] **Step 4: 实现媒体节点和参数面板**

图片节点用 assetRef.url 显示 img 缩略图；视频节点在有 assetRef.url 时显示 video preload=metadata。选中视频节点时在现有手动编辑面板显示提示词、模型、比例、清晰度、时长和参考方式，并显示生成、取消、重试、确认回填中与任务状态相符的操作。

提交返回 UNSUPPORTED_REFERENCE_IMAGES 时使用 Ant Design Modal.confirm，明确显示当前模型不支持参考图；用户选择继续才以 ignoreReferences: true 再次提交，取消不改节点、不创建任务。awaiting_confirmation 的视频只允许预览和确认，不能自动写入 assetRef。

- [ ] **Step 5: 运行前端单元测试和构建**

Run: node --test frontend/src/user/pages/agent-canvas-video-jobs.test.js frontend/src/user/pages/agent-canvas-context-menu.test.js  
Expected: PASS。

Run: npm --prefix frontend run build  
Expected: PASS，且 frontend/dist 只作为构建输出，不作为实现来源。

- [ ] **Step 6: 检查范围，不提交**

Run: git diff --check -- frontend/src/shared/api/agent.js frontend/src/user/components/AgentCanvas.jsx frontend/src/user/pages  
Expected: 无输出。保留本地改动，不执行 git commit。

### Task 5: 本机 Go embed 运行与真实交互验收

**Files:**
- Modify: 仅在测试发现缺口时修改对应的实现或测试文件；不得修改已执行迁移。
- Test: backend/cmd/qiantie/main_test.go
- Test: backend/internal/app/app_test.go

**Interfaces:**
- Consumes: 前四项已通过的 Go API 和 React 画布。
- Produces: 本机嵌入前端后的可访问页面和可恢复的视频任务界面。

- [ ] **Step 1: 写或补充应用装配测试**

    func TestApplicationRegistersAgentVideoJobRoutes(t *testing.T) {
        handler := newTestApplication(t)
        response := signedJSONRequest(t, handler, now, "alice", http.MethodPost,
            "/api/agent/tasks/agent-1/video-jobs", map[string]any{
                "nodeId": "video-1", "idempotencyKey": "idempotency-1",
            })
        if response.Code == http.StatusNotFound {
            t.Fatal("agent video routes were not registered")
        }
    }

- [ ] **Step 2: 运行装配测试，确认失败原因是明确的**

Run (from backend): go test ./cmd/qiantie ./internal/app ./internal/httpapi ./internal/agentworkspace ./internal/storage -count=1  
Expected: PASS；如果无可用模型或无 TOS 预签名能力，测试必须验证明确错误，不能接受假成功。

- [ ] **Step 3: 构建 Go embed 本机运行包**

Run: powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1  
Expected: PASS，前端重新构建后嵌入 Go 二进制。

- [ ] **Step 4: 启动或重启本机 4001 并做浏览器验收**

Run: powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql  
Expected: 本机 http://localhost:4001/agent 可访问。

在已登录浏览器中完成：新建或打开 Agent 项目、上传一张图片、放入图片节点、创建并连接视频节点、填写参数、点击生成。至少验证提交后的任务状态能在刷新后恢复；若本机没有模型或 TOS 配置，页面显示明确配置错误。只有在真实可用模型和对象存储都已配置时，才继续验证“生成结果进项目素材 -> 用户确认 -> 视频节点回填”。

- [ ] **Step 5: 记录验收边界，不提交或发布**

Run: git status --short  
Expected: 只报告本任务文件与用户已有改动的并存状态；不清理、不回退、不发布公网，也不执行 git commit。

## Plan Self-Review

- 规格覆盖：任务 1 覆盖 MySQL 事实记录、幂等和节点版本；任务 2 覆盖私有参考图和结果素材；任务 3 覆盖 Go API、状态、取消与重试；任务 4 覆盖节点预览和用户操作；任务 5 覆盖 Go embed、本机运行和真实浏览器验收。
- 迁移安全：只追加 8800006，不触碰 8800001 到 8800005 及其记录。
- 类型一致性：前端只使用 API 的 job；后端以 VideoJob 为任务事实，确认操作由 ConfirmVideoJob 原子完成。
- 已知运行前提：远端参考图要求 Agent TOS 存储可生成 HTTPS 预签名地址；没有该能力时，系统必须失败并提示，不允许拿私有浏览器地址冒充参考图。
