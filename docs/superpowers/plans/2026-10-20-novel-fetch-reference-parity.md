# 小说获取参考功能闭环补齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (\`- [ ]\`) syntax for tracking.

**Goal:** 保持当前五个小说获取页面和地址，补齐输入、任务追踪、敏感词 AI、配置档、网站提交和日志的真实业务闭环。

**Architecture:** Go 工作台接口先扩展真实任务阶段和配置字段，React 只读取这些字段。Redis 继续负责队列、租约、去重和取消；MySQL 保存任务、计划、配置、提交和日志；处理规则、操作日志仍分别位于知识库、提交与记录的二级工作页。

**Tech Stack:** Go、MySQL、Redis、React 18、Ant Design 5、Node test runner、Go embed、Vite。

## Global Constraints

- 只保留处理、任务、知识库、配置、提交与记录五个一级页面。
- Go 后端继续位于 \`backend\`；React 源码继续位于 \`frontend/src\`；不得用编译产物覆盖源码。
- 系统提示词和密钥只由后端保存；公开配置只能返回掩码和“已配置”状态。
- 配置和任务按账号隔离；接口也必须执行管理员/开发者的数据范围校验。
- 任务“完成”不等于“已提交”；版本未提交、失败、未知结果、可重试时间必须分别可见。
- 每项任务只提交列出的文件，不带入现有脏工作区的其他修改。

---

### Task 1: 输入格式、解析预览和计划元数据

**Files:**
- Modify: \`backend/internal/httpapi/novel_fetch_workshop.go:1113-1255\`
- Modify: \`backend/internal/httpapi/novel_fetch_legacy_workflow.go:199-230,422-459\`
- Modify: \`backend/internal/httpapi/novel_fetch_workshop_test.go\`
- Modify: \`frontend/src/shared/api/novelFetchWorkshop.js\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:930-1050\`
- Create: \`frontend/src/user/pages/novel-fetch-input-format.js\`
- Create: \`frontend/src/user/pages/novel-fetch-input-format.test.js\`

**Interfaces:**
- Consumes: \`POST /api/novel-fetch-workshop/process/preview\`、\`POST /api/novel-fetch-workshop/process/start\`。
- Produces: \`input_format\`、\`column_preset\`、\`column_order\`、逐行 \`parse_rows\`；计划记录 \`selected_versions\`、\`text_model_id\`、\`auto_submit\`。

- [ ] **Step 1: 写失败的输入归一化测试**

\`\`\`js
test('固定列预设保留 ID、书名、男女频和理由', () => {
  assert.deepEqual(normalizeNovelFetchInput('2094511573566019532\t鼓励式老公\t女频\t闺蜜背刺', { format: 'paid_id_title_gender_reason' }).rows[0], {
    bookId: '2094511573566019532', bookName: '鼓励式老公', gender: '女频', reason: '闺蜜背刺', error: ''
  });
});
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`node --test frontend/src/user/pages/novel-fetch-input-format.test.js\`

Expected: FAIL，因为输入预设模块不存在。

- [ ] **Step 3: 实现最小输入和计划契约**

在前端模块定义智能识别、表头、多表头、固定列、自定义列和书城 ID 格式；返回有效行、重复 ID、空 ID 和逐行错误。处理页在启动前调用预览并展示结果。

在 Go 请求结构和计划记录中保存格式、列顺序、版本、模型和自动提交；定时任务继续默认自动提交，但已运行计划只读。

- [ ] **Step 4: 写 Go 契约测试并运行**

\`\`\`go
func TestWorkshopPreviewKeepsInputFormatMetadata(t *testing.T) {}
func TestWorkshopSchedulePersistsSelectedVersionsAndAutoSubmit(t *testing.T) {}
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-input-format.test.js; go test ./backend/internal/httpapi -run 'TestWorkshopPreviewKeepsInputFormatMetadata|TestWorkshopSchedulePersistsSelectedVersionsAndAutoSubmit'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 1**

Run: \`git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/novel_fetch_legacy_workflow.go backend/internal/httpapi/novel_fetch_workshop_test.go frontend/src/shared/api/novelFetchWorkshop.js frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-input-format.js frontend/src/user/pages/novel-fetch-input-format.test.js; git commit -m "feat(novel-fetch): add input presets and schedule metadata"\`

### Task 2: 任务阶段和逐版本提交状态

**Files:**
- Modify: \`backend/internal/httpapi/novel_fetch_workshop.go:522-576,1348-1500,2523-2617\`
- Modify: \`backend/internal/novelfetchworkshop/status.go\`
- Modify: \`backend/internal/novelfetchworkshop/status_test.go\`
- Modify: \`frontend/src/user/pages/novel-fetch-status.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-task-detail.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-task-detail.test.js\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:364-453,1180-1500\`

**Interfaces:**
- Consumes: \`Document.Meta\`、版本正文、网站提交日志。
- Produces: \`stages[classify|original|rules|sensitive|rewrite|submit]\` 与 \`version_submit_status\`。

- [ ] **Step 1: 写失败的阶段状态测试**

\`\`\`go
func TestDeriveTaskStatusKeepsPendingSubmitWhenProcessingIsDone(t *testing.T) {
  doc := Document{OriginalRawChars: 20, Versions: map[string]string{"ai1": "正文"}, Meta: map[string]any{"status": "completed", "siteSubmitStatus": "pending_submit"}}
  if got := DeriveTaskStatus(doc); got != TaskStatusPendingSubmit { t.Fatalf("status=%s", got) }
}
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`go test ./backend/internal/novelfetchworkshop -run TestDeriveTaskStatusKeepsPendingSubmitWhenProcessingIsDone\`

Expected: FAIL，因为当前没有逐版本提交状态的统一派生。

- [ ] **Step 3: 实现稳定状态输出**

每阶段保存状态、起止时间、尝试次数、错误码、可读说明和下次重试时间。任务列表和详情返回解析来源、推送时间、抓取重试、敏感词命中/修复、版本提交状态和下次重试；React 明确渲染“处理完成，待提交”。

- [ ] **Step 4: 写前端显示测试并运行**

\`\`\`js
test('已改文但未提交的版本显示为待提交', () => {
  assert.equal(novelFetchSubmissionLabel({ aiStatus: 'done', aiGeneratedCount: 1, siteSubmitStatus: 'pending_submit' }), '待提交');
});
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-status.test.js frontend/src/user/pages/novel-fetch-task-detail.test.js; go test ./backend/internal/novelfetchworkshop ./backend/internal/httpapi -run 'TestDeriveTaskStatusKeepsPendingSubmitWhenProcessingIsDone|TestAppendWorkshopTaskLogPersistsTaskOperationLog'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 2**

Run: \`git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/novelfetchworkshop/status.go backend/internal/novelfetchworkshop/status_test.go frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-status.js frontend/src/user/pages/novel-fetch-task-detail.js frontend/src/user/pages/novel-fetch-task-detail.test.js; git commit -m "feat(novel-fetch): expose task stages and version submissions"\`

### Task 3: 批量任务操作和自动队列状态

**Files:**
- Modify: \`backend/internal/httpapi/novel_fetch_workshop.go:1337-1499,2071-2157\`
- Modify: \`backend/internal/httpapi/novel_fetch_legacy_workflow.go:462-570\`
- Modify: \`backend/internal/httpapi/novel_fetch_workshop_test.go\`
- Modify: \`frontend/src/shared/api/novelFetchWorkshop.js\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:1060-1710\`
- Modify: \`frontend/src/user/pages/novel-fetch-status.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-status.test.js\`

**Interfaces:**
- Consumes: 选中书籍 ID、规则范围、AI 数量。
- Produces: 每个批量接口统一返回 \`{ requested, updated, skipped, failed, task_ids }\`。

- [ ] **Step 1: 写失败的批量调整测试**

\`\`\`go
func TestBatchAdjustReturnsAffectedTaskIDs(t *testing.T) {}
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`go test ./backend/internal/httpapi -run TestBatchAdjustReturnsAffectedTaskIDs\`

Expected: FAIL，因为批量调整没有统一影响结果。

- [ ] **Step 3: 实现批量结果与页面确认**

补齐调整选中、调整待处理、重试选中、规则处理选中/全部、删除选中/失败；拒绝空选中，单本失败不阻断其余任务。操作前显示影响数量，成功后只刷新任务、实时状态和当前批次。

队列摘要继续读取 Redis/运行注册表的真实字段，显示运行中、排队、等待重试、最近任务和更新时间；页面隐藏时停止轮询，前台实时状态 2 秒、队列 5 秒、任务表至少 10 秒。

- [ ] **Step 4: 写状态测试并运行**

\`\`\`js
test('队列摘要显示等待重试和最近任务', () => {
  const s = novelFetchQueueSummary({ enabled: true, counts: { running: 1, queued: 2, waiting_retry: 3 }, latestJob: { id: 'job-7' } });
  assert.equal(s.waitingRetry, 3);
  assert.equal(s.latestJob.id, 'job-7');
});
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-status.test.js; go test ./backend/internal/httpapi -run 'TestBatchAdjustReturnsAffectedTaskIDs|TestNovelFetchScheduleDefaultsToAutomaticUpload'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 3**

Run: \`git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/novel_fetch_legacy_workflow.go backend/internal/httpapi/novel_fetch_workshop_test.go frontend/src/shared/api/novelFetchWorkshop.js frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-status.js frontend/src/user/pages/novel-fetch-status.test.js; git commit -m "feat(novel-fetch): complete task batch controls"\`

### Task 4: 敏感词 AI 参数和规则追踪

**Files:**
- Modify: \`backend/internal/httpapi/novel_fetch_workshop.go:821-951\`
- Modify: \`backend/internal/novelfetchworkshop/sensitive.go\`
- Modify: \`backend/internal/novelfetchworkshop/sensitive_scope_test.go\`
- Modify: \`backend/internal/httpapi/novel_fetch_rewrite_test.go\`
- Modify: \`frontend/src/user/pages/novel-fetch-config-defaults.js\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:153-228,255-363\`
- Modify: \`frontend/src/user/pages/novel-fetch-task-detail.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-task-detail.test.js\`

**Interfaces:**
- Consumes: \`sensitive_ai.mode\`、\`concurrency\`、\`max_hits\`、\`retries\`、\`context_chars\`、\`temperature\`、\`prompt\`、\`enabled\`。
- Produces: 每个命中词的模式、来源、模型、尝试次数、前后文本和错误记录。

- [ ] **Step 1: 写失败的敏感词模式测试**

\`\`\`go
func TestSensitiveAIRepairRespectsHitLimitAndPerLineMode(t *testing.T) {}
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`go test ./backend/internal/httpapi -run TestSensitiveAIRepairRespectsHitLimitAndPerLineMode\`

Expected: FAIL，因为现有实现没有命中上限和模式切换。

- [ ] **Step 3: 实现整组和逐条策略**

\`group\` 模式将限定命中和上下文一次发送；\`per_line\` 模式按配置并发逐条处理。两种模式都应用保存的温度、重试、提示词和上下文长度。AI 关闭或失败时保留确定性替换结果，并将完整记录写入 \`rulesTrace\`、\`sensitiveHits\` 和 \`sensitiveFixed\`。

规则二级页只编辑和保存这些 Go 配置字段；详情页和日志页读取同一份追踪数据。

- [ ] **Step 4: 写默认值测试并运行**

\`\`\`js
test('敏感词 AI 默认配置可持久化', () => {
  const cfg = withNovelFetchConfigDefaults({});
  assert.equal(cfg.sensitive_ai.mode, 'group');
  assert.equal(cfg.sensitive_ai.max_hits, 50);
});
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-config-defaults.test.js frontend/src/user/pages/novel-fetch-task-detail.test.js; go test ./backend/internal/novelfetchworkshop ./backend/internal/httpapi -run 'TestSensitiveAIRepairRespectsHitLimitAndPerLineMode|TestSensitiveRuleGroupsHonorEnabledAndScope'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 4**

Run: \`git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/novelfetchworkshop/sensitive.go backend/internal/novelfetchworkshop/sensitive_scope_test.go backend/internal/httpapi/novel_fetch_rewrite_test.go frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-config-defaults.js frontend/src/user/pages/novel-fetch-config-defaults.test.js frontend/src/user/pages/novel-fetch-task-detail.js frontend/src/user/pages/novel-fetch-task-detail.test.js; git commit -m "feat(novel-fetch): add sensitive AI execution controls"\`

### Task 5: AI 预设、模型职责和 AI1-AI5 配置档

**Files:**
- Modify: \`backend/internal/httpapi/novel_fetch_workshop.go:2188-2255,2471-2735\`
- Modify: \`backend/internal/httpapi/batch_rewrite_compat.go:860-1030\`
- Modify: \`backend/internal/httpapi/novel_fetch_workshop_config_test.go\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:255-363,820-930\`
- Modify: \`frontend/src/user/pages/novel-fetch-config-defaults.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-config-defaults.test.js\`

**Interfaces:**
- Consumes: \`appConfig.ai_presets\`、\`model_roles\`、\`web_submit.default_profile\`、\`web_submit.version_profiles\`。
- Produces: 掩码后的预设、模型职责和原文/AI1-AI5 的有效配置档。

- [ ] **Step 1: 写失败的版本映射测试**

\`\`\`go
func TestNovelFetchConfigKeepsAllVersionProfileBindings(t *testing.T) {}
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`go test ./backend/internal/httpapi -run TestNovelFetchConfigKeepsAllVersionProfileBindings\`

Expected: FAIL，因为当前页面没有完整展示 AI2-AI5 映射。

- [ ] **Step 3: 实现校验和有效值展示**

保存原文和 AI1-AI5 的默认或指定配置档；非空配置档必须存在于已同步目录。继续掩码密钥和密码。配置页展示选择值及实际生效值；模型预设测试只走受控 Go 接口。

- [ ] **Step 4: 写前端默认值测试并运行**

\`\`\`js
test('版本配置档默认让全部版本跟随默认配置档', () => {
  assert.deepEqual(withNovelFetchConfigDefaults({}).web_submit.version_profiles, { original: '', ai1: '', ai2: '', ai3: '', ai4: '', ai5: '' });
});
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-config-defaults.test.js; go test ./backend/internal/httpapi -run 'TestNovelFetchConfigKeepsAllVersionProfileBindings|TestNovelFetchWorkshopConfigHidesCredentialsAndPreservesOnEdit'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 5**

Run: \`git add backend/internal/httpapi/novel_fetch_workshop.go backend/internal/httpapi/batch_rewrite_compat.go backend/internal/httpapi/novel_fetch_workshop_config_test.go frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-config-defaults.js frontend/src/user/pages/novel-fetch-config-defaults.test.js; git commit -m "feat(novel-fetch): complete config presets and version profiles"\`

### Task 6: 提交环境、逐版本结果和操作日志

**Files:**
- Modify: \`backend/internal/httpapi/web_submit_compat.go\`
- Modify: \`backend/internal/httpapi/web_submit_preview_compat.go\`
- Modify: \`backend/internal/httpapi/web_submit_submit_compat.go\`
- Modify: \`backend/internal/httpapi/web_submit_service.go\`
- Modify: \`backend/internal/httpapi/web_submit_submit_compat_test.go\`
- Modify: \`frontend/src/shared/api/novelFetch.js\`
- Modify: \`frontend/src/user/pages/NovelFetchPage.jsx:1510-1900\`
- Modify: \`frontend/src/user/pages/novel-fetch-submit.js\`
- Modify: \`frontend/src/user/pages/novel-fetch-submit.test.js\`

**Interfaces:**
- Consumes: \`web_submit.execution_mode\`、\`python_path\`、\`browser_path\`、\`concurrency\`、\`min_file_bytes\`、\`retries\`、\`skip_submitted\`、\`version_profiles\`。
- Produces: 每个版本的 \`operation_id\`、\`group_id\`、\`profile\`、\`status\`、\`retryable\`、\`error\`、\`readback_status\`。

- [ ] **Step 1: 写失败的逐版本提交测试**

\`\`\`go
func TestWebSubmitStoresPerVersionProfileAndReadbackState(t *testing.T) {}
\`\`\`

- [ ] **Step 2: 运行失败测试**

Run: \`go test ./backend/internal/httpapi -run TestWebSubmitStoresPerVersionProfileAndReadbackState\`

Expected: FAIL，因为提交结果没有完整暴露逐版本配置档和回读状态。

- [ ] **Step 3: 实现真实提交审计**

验证执行模式、路径、并发、字节阈值和重试范围。预览只读并返回任务/版本/配置档/分组；提交写入运行 ID、实际配置档、上传组、可重试标记和回读结果。未知结果只回读，确认失败后才可重试，确认已提交的版本必须跳过。

提交二级页按运行 ID、任务 ID、书籍 ID、阶段和时间显示准备、结果、历史和日志。

- [ ] **Step 4: 写前端重试测试并运行**

\`\`\`js
test('单版本重试携带运行 ID', () => {
  assert.deepEqual(singleVersionRetryPayload({ book_id: 'book-9', version: 'ai3', operation_id: 'op-1' }), { taskIds: ['book-9'], versions: ['ai3'], operationId: 'op-1' });
});
\`\`\`

Run: \`node --test frontend/src/user/pages/novel-fetch-submit.test.js; go test ./backend/internal/httpapi -run 'TestWebSubmitStoresPerVersionProfileAndReadbackState|TestWebSubmitUncertainOutcomeOnlyReadsBackAndNeverUploadsTwice'\`

Expected: PASS。

- [ ] **Step 5: 提交任务 6**

Run: \`git add backend/internal/httpapi/web_submit_compat.go backend/internal/httpapi/web_submit_preview_compat.go backend/internal/httpapi/web_submit_submit_compat.go backend/internal/httpapi/web_submit_service.go backend/internal/httpapi/web_submit_submit_compat_test.go frontend/src/shared/api/novelFetch.js frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-submit.js frontend/src/user/pages/novel-fetch-submit.test.js; git commit -m "feat(novel-fetch): complete submission audit trail"\`

### Task 7: Go embed 构建和已登录浏览器验收

**Files:**
- Verify: \`frontend/src/user/pages/NovelFetchPage.jsx\`
- Verify: \`frontend/src/user/pages/novel-fetch.css\`
- Verify: \`frontend/src/user/pages/novel-fetch-subpages.css\`
- Verify: \`scripts/build-go-local.ps1\`
- Verify: \`scripts/start-local-go-review.ps1\`

**Interfaces:**
- Consumes: 前六项的 Go API、React 页面和 Go embed 构建脚本。
- Produces: 本机 \`http://localhost:4001/novel-fetch\` 的已登录可操作页面。

- [ ] **Step 1: 运行全部定向测试**

Run: \`node --test frontend/src/user/pages/novel-fetch-*.test.js frontend/src/shared/api/novelFetchWorkshop.test.js; go test ./backend/internal/novelfetchworkshop ./backend/internal/httpapi\`

Expected: PASS；任何无关既有失败单独记录，不能跳过。

- [ ] **Step 2: 构建嵌入式运行包**

Run: \`powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1\`

Expected: React 和 Go embed 构建成功。

- [ ] **Step 3: 启动本机审阅服务**

Run: \`powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql\`

Expected: \`4001\` 监听，五个页面和二级页都能加载。

- [ ] **Step 4: 已登录浏览器验收**

验证混合输入预览、计划创建/取消、自动队列、单书阶段追踪、敏感词两种模式、AI1-AI5 映射、提交预览、失败或未知提交回读，以及浅色/深色主题下的可读性。只以真实接口和外部提交回读为准。

## Plan Self-Review

- 设计中的输入、计划、阶段状态、批量操作、敏感词、配置档、提交与日志分别由任务 1 至 6 覆盖。
- React 只读取 Go 返回的真实状态，不模拟队列或提交成功。
- 密钥掩码、账号隔离、空知识库保护和“处理完成不等于已提交”均有实现与验收步骤。
- 全文没有未完成占位措辞。
