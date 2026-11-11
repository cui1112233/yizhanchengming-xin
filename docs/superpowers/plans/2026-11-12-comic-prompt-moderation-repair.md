# Comic Prompt Moderation Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 当漫画提示词因成人内容或其他内容审核被拒绝时，自动生成安全替代请求，保留原文对白用于页面文字，并只补跑未完成页面。

**Architecture:** 后端提示词模块负责把审核失败请求转换为安全重试请求，并保存对白占位符与原文的映射。漫画提示词服务最多执行两轮安全重试；成功解析后只把原文对白恢复到 `screen_prompt`，绝不恢复进交给图片模型的 `image_prompt`。所有已完成页面继续沿用现有缺页补跑机制，不重新调用。

**Tech Stack:** Go、现有 `batchfactoryv11` 服务、后端 `prompt` 模块、Go 单元测试、Go embed 单二进制发布。

## Global Constraints

- 不改小说原文和已经生成完成的漫画页。
- 只有内容审核错误触发安全重试；网络、鉴权、地址错误沿用原有错误处理。
- 原始调用加最多两轮安全重试，避免无限扣费。
- 对白只恢复到 `screen_prompt`；`image_prompt` 始终保持安全版本。
- 先本地测试和打包，再部署线上；线上只续跑《爱到此处停止》的失败页。
- 不创建分支、不清理或覆盖工作树中的无关改动。

---

### Task 1: 安全请求与对白映射

**Files:**
- Modify: `backend/internal/prompt/comic_visual_safety.go`
- Create: `backend/internal/prompt/comic_visual_safety_test.go`

**Interfaces:**
- Produces: `ComicVisualSafetyRetryPlan(systemPrompt, userPrompt string, round int) ComicVisualSafetyPlan`
- Produces: `ComicVisualSafetyPlan.RestoreScreenPrompt(screenPrompt string) string`
- `ComicVisualSafetyPlan` 包含 `SystemPrompt string`、`UserPrompt string` 和内部对白映射。

- [ ] **Step 1: Write the failing tests**

  覆盖三点：重试请求不再包含原始成人对白；使用稳定的 `{对白1}` 占位；恢复函数只恢复导演说明中的对白。

- [ ] **Step 2: Run tests to verify they fail**

  Run: `go test ./internal/prompt -run 'TestComicVisualSafetyRetryPlan|TestComicVisualSafetyRestore' -count=1`

  Expected: FAIL，因为新接口尚不存在。

- [ ] **Step 3: Write minimal implementation**

  在提示词模块增加安全重试计划：抽取中文引号中的对白并替换为占位符；清理暴力、血腥、成人暗示和身体接触画面词；第二轮追加“日常、无接触、空镜优先”的更严格规则。恢复时替换 `screen_prompt` 中的占位符；若模型漏掉占位符，则把原文对白作为“页面文字”附加到 `screen_prompt`，但不提供任何恢复 `image_prompt` 的接口。

- [ ] **Step 4: Run tests to verify they pass**

  Run: `go test ./internal/prompt -run 'TestComicVisualSafetyRetryPlan|TestComicVisualSafetyRestore' -count=1`

  Expected: PASS。

### Task 2: 两轮内容审核重试

**Files:**
- Modify: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Modify: `backend/internal/batchfactoryv11/comic_prompt_service_test.go`

**Interfaces:**
- Consumes: `prompt.ComicVisualSafetyRetryPlan`
- Produces: 原调用加最多两轮安全重试；安全调用成功后恢复 `comicPromptResult.ScreenPrompt`。

- [ ] **Step 1: Write the failing service test**

  新测试提供者让前两次调用返回 `PROHIBITED_CONTENT/成人内容拒绝`，第三次只有在原始对白已被占位、严格安全规则存在时才返回 JSON。断言总调用数为 3、结果成功、`screen_prompt` 恢复原文对白、`image_prompt` 不含原始成人对白。

- [ ] **Step 2: Run test to verify it fails**

  Run: `go test ./internal/batchfactoryv11 -run TestRunComicPromptExtractionRetriesAdultPolicyTwiceAndRestoresDialogueOnlyToScreen -count=1`

  Expected: FAIL，因为现有代码只有一轮重试且保留原始对白。

- [ ] **Step 3: Implement the retry loop**

  原始调用失败且属于内容审核时，依次执行 round 1、round 2；任一轮成功即停止。解析成功后调用该轮计划的 `RestoreScreenPrompt`，保持 `ImagePrompt` 原样。非内容审核错误不进入循环。

- [ ] **Step 4: Run focused and regression tests**

  Run: `go test ./internal/batchfactoryv11 -run 'TestRunComicPromptExtractionRetries|TestRunMissingComicPromptExtractionPreservesCompletedPages' -count=1`

  Expected: PASS；旧测试调整为断言对白在发送时被占位、结果中仍可恢复。

### Task 3: 本地验证与单二进制发布

**Files:**
- Create: `backend/internal/releasenotes/notes/<release-id>.json`
- Modify: `backend/current-release.json`（由发布脚本生成）

**Interfaces:**
- Produces: 带嵌入前端资源的 Linux 单二进制和对应发布清单。

- [ ] **Step 1: Run backend verification**

  Run: `go test ./internal/prompt ./internal/batchfactoryv11 -count=1`

  Expected: PASS。

- [ ] **Step 2: Build and locally preflight the release**

  Run: `powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\release-v88.ps1 -ReleaseId <release-id> -SkipFrontendBuild -AllowMissingEnv`

  Expected: Windows/Linux 构建成功，本地 `/health`、`/api/runtime-build-info` 和发布公告校验成功。

- [ ] **Step 3: Deploy the Linux binary**

  使用现有 `_local-tmp/deploy_qiantie_release.py` 上传到 `/opt/www/qiantie`，切换 `qiantie.service` 后验证服务、二进制 SHA256、公开健康检查和更新公告。

### Task 4: 只续跑失败作品并验收

**Files:**
- No source file changes.

**Interfaces:**
- Consumes: batch `batch_cd021a9d701e19cca2aeb9bc810c89d2`
- Consumes: book `book_6f1b7c8636d901dbb757092f1ca7b5bd`（《爱到此处停止》）

- [ ] **Step 1: Retry only the failed book**

  调用单本重试脚本，不触碰已完成的《碎在凛冬的谎言与真心》。

- [ ] **Step 2: Verify the missing prompt recovers**

  验证原失败页“漫画页 08”不再停在 `comic_prompt_recovery`，日志中出现安全重试并成功保存提示词。

- [ ] **Step 3: Verify real image return**

  观察该书从现有 `1/27` 继续增长，确认至少一张新图真实写入并在页面可见；若上游仍拒绝，则停止重复付费调用并保留单页明确失败原因。

