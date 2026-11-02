# 漫画与视频双模块链路对齐计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (recommended) to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保持视频和漫画为两个独立生产模块、保持现有横向漫画页交互的前提下，核对并修正漫画资产、原文分页、分镜提示词、单页/整本生成和合集整理之间的实际链路。

**Architecture:** 两个模块共用批量工厂外壳、书籍、原文和统一配置交互，但各自拥有独立的资产来源、提示词字段、生产状态和结果集合。漫画页可以继续复用现有分镜编号作为内部关联，但漫画页的分页数量、提示词和合集校验不能被视频模块的状态或资产误导。

**Tech Stack:** Go 后端、React/Antd 前端、MySQL 持久化、现有 Go embed 本地运行路径、Node 前端测试、Go 单元测试。

**Spec:** `docs/superpowers/specs/2026-10-25-batch-factory-comic-production-design.md`、`docs/superpowers/specs/2026-10-27-batch-factory-comic-source-segmentation-design.md`、本地 REA 取证资料 `G:/Backup/Documents/ChatGPT/漫画/雪漫画_反编译源码/REA_CHAIN_AND_PROMPTS_REPORT.md`。

## Global Constraints

- 视频和漫画必须保持两个独立生产模块；统一配置只负责模式切换和共用交互外壳。
- 漫画页继续使用当前横向翻页展示，不改成雪漫画的纵向展示方式。
- 漫画资产只能读取 `source=comic` 的人物、场景和道具资产。
- 视频资产、视频提示词、视频任务和视频合成不能被漫画流程读取或覆盖。
- 本计划不处理图片模型供应商、图片接口、额度、重试传输和生图质量。
- 不调用真实图片生成，不消耗额度，不发布，不修改无关的脏工作区文件。
- 系统提示词继续由后端 prompt 模块提供，前端不硬编码提示词正文。

## Review Focus

- 已存在视频资产但不存在漫画资产时，漫画提取必须仍然执行；对应测试放在 `backend/internal/httpapi/batch_factory_automation_test.go`。
- 已有漫画分页时，漫画页数量必须由漫画原文分段决定，不能因视频分镜数量或旧状态被静默复用；对应测试放在 `backend/internal/batchfactoryv11/comic_source_segmenter_test.go` 或现有分页测试文件。
- 漫画页横向切换时，页码、原文段、提示词和候选图必须保持同一页绑定；对应测试放在 `frontend/src/user/pages/shuihuo/BatchFactoryNovelList` 相关测试。
- 漫画提示词只绑定本页真实出场角色和漫画资产，不能用普通文本包含关系误绑资产；对应测试放在 `backend/internal/httpapi/batch_factory_comic_pages_test.go`。
- 原文、提示词或配置变化后，旧主图不能直接进入新的漫画合集；对应测试放在 `backend/internal/batchfactoryv11/comic_pages_test.go` 或合集接口测试。

### Task 1: 建立单书与批量自动化的链路基线

**Files:**
- Inspect: `backend/internal/batchfactoryv11/stage_service.go`
- Inspect: `backend/internal/httpapi/batch_factory_automation.go`
- Inspect: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Test: `backend/internal/httpapi/batch_factory_automation_test.go`

**Interfaces:**
- Consumes: 当前单书按钮、单书阶段接口、批量自动化入口和 `productionMode`。
- Produces: 一张明确的行为矩阵，记录每个按钮读取什么、写什么、是否跳过、是否只补缺失内容。

- [ ] **Step 1: 记录单书操作链路**

  明确“漫画资产提取”“漫画获取/提示词提取”“生成本页漫画”“生成整本漫画”“整理漫画合集”分别调用的接口、阶段和存储对象。

- [ ] **Step 2: 记录批量自动化链路**

  明确自动化如何判断 `bootstrap`、补齐提示词、复用现有结果以及进入漫画图片阶段。

- [ ] **Step 3: 写下单书与批量链路差异**

  只保留必要差异；如果两条入口对同一状态的判断不同，先记录为问题，不在本任务直接修改。

- [ ] **Step 4: 用现有测试和只读浏览器状态核对矩阵**

  浏览器只查看按钮和状态，不点击提取、生成、重试或发布。

### Task 2: 修正漫画资产与视频资产的边界判断

**Files:**
- Modify: `backend/internal/httpapi/batch_factory_automation.go`
- Inspect/possibly modify: `backend/internal/batchfactoryv11/comic_bootstrap_service.go`
- Test: `backend/internal/httpapi/batch_factory_automation_test.go`
- Test: `backend/internal/batchfactoryv11/comic_reuse_test.go`

**Interfaces:**
- Consumes: `Book.AssetRecords`、资产 `Source`、原文摘要和漫画分页/提示词状态。
- Produces: 批量自动化和单书阶段使用同一套漫画资产复用判断。

- [ ] **Step 1: 写失败测试**

  覆盖“只有视频资产、没有漫画资产”的书，断言批量自动化选择漫画初始化；覆盖“已有完整漫画资产”的书，断言普通运行复用而不是重复提取。

- [ ] **Step 2: 运行目标测试确认当前差异**

  运行 `go test ./backend/internal/httpapi -run 'Test.*Comic.*(Preparation|Bootstrap|Automation)'` 和相关 `batchfactoryv11` 测试，记录失败原因。

- [ ] **Step 3: 统一判断条件**

  批量自动化不得使用任意资产数量作为漫画资产存在依据，必须与单书逻辑一样只认有效的 `source=comic` 资产，并同时考虑原文摘要是否对应当前原文。

- [ ] **Step 4: 运行目标测试和全部 Go 测试**

  目标测试通过后运行 `go test ./...`，确认视频资产流程没有回归。

### Task 3: 确认漫画分页与漫画页的真实绑定关系

**Files:**
- Inspect: `backend/internal/batchfactoryv11/comic_source_segmenter.go`
- Inspect: `backend/internal/batchfactoryv11/comic_source_segments_mysql.go`
- Inspect: `backend/internal/batchfactoryv11/comic_pages_mysql.go`
- Inspect: `backend/internal/batchfactoryv11/types.go`
- Test: `backend/internal/httpapi/batch_factory_comic_source_segments_test.go`
- Test: `backend/internal/batchfactoryv11/comic_pages_test.go`

**Interfaces:**
- Consumes: `ComicSourceSegment`、`ComicPage`、稳定页 ID、页码顺序和现有 `VideoID` 关联。
- Produces: 结论明确的内部绑定规则；只有在确认旧 Video 数量会影响漫画分页时，才提出最小修正。

- [ ] **Step 1: 验证分页来源**

  证明漫画分页是否来自连续原文分段、是否完整覆盖原文、是否可能直接复用旧视频分镜数量。

- [ ] **Step 2: 验证漫画页存储**

  确认 `ComicPage` 是否独立保存提示词、参考图、候选图和主图；确认 `VideoID` 只是内部关联还是实际控制页数。

- [ ] **Step 3: 验证横向展示绑定**

  不改变展示方向，确认左右切页时页码、原文段、提示词和图片来自同一稳定页。

- [ ] **Step 4: 只修实际影响页数或顺序的问题**

  如果只是内部复用而不影响漫画分页，不做结构拆分；如果确实受视频状态影响，增加最小的漫画页顺序门禁和回归测试。

### Task 4: 固定漫画“分页后再取提示词”的顺序

**Files:**
- Inspect/modify if confirmed: `backend/internal/batchfactoryv11/comic_bootstrap_service.go`
- Inspect/modify if confirmed: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Inspect: `backend/internal/batchfactoryv11/stage_service.go`
- Test: `backend/internal/batchfactoryv11/comic_prompt_service_test.go`

**Interfaces:**
- Consumes: 已保存的漫画原文段、前页/当前页/后页上下文、漫画资产和漫画提示词预设。
- Produces: 漫画提示词只针对已确定的漫画页，不改变用户看到的独立按钮流程。

- [ ] **Step 1: 验证手动按钮顺序**

  确认用户点击漫画获取时，是先得到原文分段，再为分段生成提示词，而不是把视频分镜当作漫画分页来源。

- [ ] **Step 2: 验证自动化补齐规则**

  确认自动化只补没有提示词的漫画页，不覆盖用户已经修改或确认的漫画页。

- [ ] **Step 3: 验证原文变化规则**

  原文变化后，旧分页和旧提示词必须标记过期；未受影响的页不能被无理由重算。

- [ ] **Step 4: 增加分页、提示词、资产三者分离的回归测试**

  测试必须证明：先提取资产不会自动生成漫画页；先有分页不会自动生成图片；只补缺失提示词不会覆盖已有页。

### Task 5: 修正漫画页资产引用的结构化绑定

**Files:**
- Modify if confirmed: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Inspect: `backend/internal/batchfactoryv11/comic_prompt_service.go`
- Test: `backend/internal/httpapi/batch_factory_comic_pages_test.go`

**Interfaces:**
- Consumes: 每页的 `roles`、`role_stage`、`scene_presets` 和 `source=comic` 资产。
- Produces: 只给真实出场角色绑定对应漫画资产，不靠普通字符串包含关系猜测角色。

- [ ] **Step 1: 写失败测试**

  覆盖资产名为“我”、简称和正式名称存在包含关系、角色只出现在对白但不在画面中的情况。

- [ ] **Step 2: 运行目标测试确认误绑路径**

  确认当前参考图查找是否仍由 `strings.Contains` 决定。

- [ ] **Step 3: 使用结构化角色绑定**

  优先使用 `role_stage` 和 `roles` 对应正式资产；无法确定时不自动绑定，而不是扩大匹配范围。

- [ ] **Step 4: 验证视频资产不参与漫画页引用**

  运行漫画页接口测试，并确认视频资产不会出现在漫画页参考图列表。

### Task 6: 校验主图、过期状态和漫画合集门禁

**Files:**
- Inspect/modify if confirmed: `backend/internal/batchfactoryv11/comic_pages_mysql.go`
- Inspect/modify if confirmed: `backend/internal/httpapi/batch_factory_comic_pages.go`
- Test: `backend/internal/httpapi/batch_factory_comic_pages_test.go`
- Test: `backend/internal/httpapi/batch_factory_n8_comic_publish_test.go`

**Interfaces:**
- Consumes: 每页主图、页面快照、配置快照、原文分段版本和漫画页状态。
- Produces: 只有当前有效漫画页的主图才能进入漫画合集；缺页、过期页或无主图时返回明确清单。

- [ ] **Step 1: 验证当前合集校验**

  确认合集现在检查的是所有漫画页主图，而不是只检查对应 Video 是否有主图。

- [ ] **Step 2: 增加过期页场景测试**

  覆盖原文分段变化、提示词变化、画风变化和参考图变化后，旧主图不能静默进入新合集。

- [ ] **Step 3: 保留单页生成与整本生成的独立入口**

  单页只处理当前漫画页；整本只补缺失页并在全部主图有效后整理合集。

- [ ] **Step 4: 运行合集接口和发布前门禁测试**

  本任务不调用真实图片模型，只使用已有测试图片或内存存储验证状态。

### Task 7: 本地运行验收与交付记录

**Files:**
- Inspect: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- Inspect: `frontend/src/user/pages/shuihuo/BatchFactoryProductionModeVisibility.source.test.js`
- Inspect: `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.source.test.js`
- Runtime: `/health`、`/api/runtime-build-info`

**Interfaces:**
- Consumes: 前六项已确认的链路规则和测试结果。
- Produces: 可在浏览器中核验的漫画/视频模式边界、按钮职责和状态文案。

- [ ] **Step 1: 运行前端目标测试和构建**

  运行漫画模式、生产模式显示和内容范围相关测试，随后运行 `npm run build`。

- [ ] **Step 2: 构建并启动 Go embed 本地运行路径**

  只在确认有代码改动时执行构建和重启；核对运行时构建 SHA 与当前源码一致。

- [ ] **Step 3: 浏览器只读验收**

  验证漫画模式显示漫画资产、漫画分页、漫画提示词和漫画合集；验证视频模式仍显示视频资产、视频分镜和视频合成；不点击任何生图按钮。

- [ ] **Step 4: 记录剩余边界**

  明确哪些已经通过真实页面和接口验证，哪些只能通过代码和测试验证，不能把“生图供应商未验证”混入漫画链路结论。

## 完成标准

- 视频和漫画可以在同一套批量工厂交互中切换，但彼此不读取、不覆盖对方生产数据。
- 漫画资产提取、漫画原文分页、漫画提示词提取、单页生成、整本生成和合集整理的职责清楚且可独立重试。
- 已有视频资产不会阻止漫画资产提取；已有漫画资产不会被普通复用流程重复覆盖。
- 漫画页数量由漫画原文分页决定；横向展示不改变页码、原文段和提示词绑定。
- 漫画页提示词只引用漫画资产，合集不会收录过期或缺少主图的页面。
- 目标 Go 测试、前端目标测试、前端构建和本地运行时验收均有记录。
