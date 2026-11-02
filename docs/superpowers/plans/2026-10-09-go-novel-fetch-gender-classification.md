# Go 小说正文获取与男女频判断迁移实施计划

> **给执行代理：**按任务逐项执行；每项都必须先写失败测试，再写最小实现，再运行验证。不要修改公网，不要读取或提交本地密钥。

**目标：**把小说正文获取和男女频判断统一迁移到 Go：用户手工填写优先，121 的 `category/genre` 次之，AI 最后兜底；同时保留完整原文和按 `max_txt` 处理的正文。

**架构：**Go 后端负责请求 `https://txt.121w.com/api.php`、解析上游返回、执行男女频判断优先级、保存任务正文与判断来源。Node 暂时只作为前端兼容层，通过现有本地桥接调用 Go，不在 React 中直接请求 121。Worker 不参与小说正文获取或男女频判断。

**技术栈：**Go、MySQL、现有 `backend/internal/novelfetchworkshop`、现有 HTTP Bridge、React 前端只改调用接线（本计划不包含 UI 重做）。

## 全局约束

- 维护分支只能是 `v88`；不创建新分支，不推送远端。
- Git `v88` 是唯一源码；不修改公网运行副本。
- 不读取、不输出、不提交 `.env.v88-review`、`.local-review/`、API Key、Cookie、Token。
- 用户手工填写男女频优先级最高，不能被 121 或 AI 覆盖。
- `category` 或 `genre` 明确包含“男生、男频、男性、男向”映射为男频；明确包含“女生、女频、女性、女向”映射为女频。
- `category` 与 `genre` 冲突或无法明确判断时，才调用 AI；AI 仍无法判断时返回“待确认”，不得猜测。
- 保存 `genderSource`：`manual`、`bookinfo`、`ai`、`unknown`。
- 保存完整 `originalRaw`，并单独保存按 `max_txt` 截取后的处理正文；不得用截断正文覆盖完整原文。
- 页面字数显示必须保持“实际处理字数/原文总字数”，如 `2000/27831`。
- 外部接口地址只存在 Go 配置/服务端，不硬编码到 React。

---

### 任务 1：定义并测试 121 上游响应与男女频分类器

**文件：**
- 创建：`backend/internal/novelfetchworkshop/gender.go`
- 创建：`backend/internal/novelfetchworkshop/gender_test.go`
- 修改：`backend/internal/novelfetchworkshop` 中与正文获取结果模型相邻的类型文件（按现有结构选择）

**接口：**
- `type GenderSource string`
- `const GenderManual, GenderBookInfo, GenderAI, GenderUnknown GenderSource`
- `type BookInfo struct { Category string; Genre string }`
- `type GenderDecision struct { Gender string; Source GenderSource; Reason string }`
- `func DecideGender(manual string, info BookInfo, aiGender string) GenderDecision`

**步骤：**

- [ ] 写失败测试：手工填写“女频”时，无论书籍信息和 AI 输入是什么，都返回 `Gender="女频"`、`Source=GenderManual`。
- [ ] 写失败测试：无手工值且 `category="男频"` 返回男频和 `bookinfo` 来源。
- [ ] 写失败测试：无手工值且 `genre="女性向"` 返回女频和 `bookinfo` 来源。
- [ ] 写失败测试：`category="男频"`、`genre="女频"` 冲突时不采用 bookinfo，使用合法 AI 结果；AI 空值时返回待确认和 `unknown`。
- [ ] 运行：`go test ./internal/novelfetchworkshop -run Gender`；预期先失败，因为分类器尚不存在。
- [ ] 实现最小关键词归一化和分类优先级；关键词匹配必须基于 `category + genre`，冲突不能静默选一边。
- [ ] 运行同一命令，预期全部通过。

### 任务 2：迁移 121 正文 HTTP 客户端

**文件：**
- 创建或修改：`backend/internal/novelfetchworkshop/upstream.go`
- 创建：`backend/internal/novelfetchworkshop/upstream_test.go`
- 修改：`backend/internal/config/config.go`（仅增加可配置的 121 正文地址和超时，保持默认地址）

**接口：**
- `type UpstreamFetcher struct { Client *http.Client; Endpoint string }`
- `type FetchResult struct { Code int; Message string; OriginalRaw string; BookInfo BookInfo }`
- `func (f UpstreamFetcher) Fetch(ctx context.Context, bookID, platformID string, maxTxt int) (FetchResult, error)`

**步骤：**

- [ ] 写失败测试：测试服务器收到 `bookid`、`platform`、`max_txt` 三个查询参数。
- [ ] 写失败测试：`{"code":200,"data":"完整正文","bookinfo":{"category":"男频","genre":"都市"}}` 能解析正文和书籍分类。
- [ ] 写失败测试：非 200 业务码、非 JSON、超时分别返回明确错误，不把错误响应当正文保存。
- [ ] 运行 `go test ./internal/novelfetchworkshop -run Upstream`，确认先失败。
- [ ] 实现 URL 构造、超时、JSON 解析和上游错误处理；默认地址为 `https://txt.121w.com/api.php`。
- [ ] 运行定向测试确认通过。

### 任务 3：实现完整原文与 max_txt 处理结果

**文件：**
- 修改：`backend/internal/novelfetchworkshop` 的文档/正文模型和服务文件
- 创建或修改：`backend/internal/novelfetchworkshop/processing_test.go`
- 修改：`backend/internal/storage/novel_fetch_workshop_schema.go`（如现有 JSON 文档字段已能承载，则不新增表）

**接口：**
- `type ProcessedBody struct { OriginalRaw string; ProcessedText string; OriginalRawChars int; ProcessedChars int; MaxTxt int }`
- `func ProcessRawText(originalRaw string, maxTxt int) ProcessedBody`

**步骤：**

- [ ] 写失败测试：原文超过 `maxTxt` 时，`OriginalRaw` 保持完整，`ProcessedText` 只取前 `maxTxt` 个字符。
- [ ] 写失败测试：短文本不被错误截断，两个字数统计准确。
- [ ] 写失败测试：空文本和非法 `maxTxt` 返回稳定结果，不产生负数统计。
- [ ] 运行定向测试确认先失败。
- [ ] 实现最小处理函数；保存文档时同时写入完整原文、处理正文和计数字段。
- [ ] 运行定向测试及现有小说工作台测试。

### 任务 4：把分类优先级接入 Go 小说工作台流程

**文件：**
- 修改：`backend/internal/novelfetchworkshop` 服务流程
- 修改：`backend/internal/httpapi/novel_fetch_workshop.go`
- 创建或修改：`backend/internal/httpapi/novel_fetch_gender_test.go`

**接口：**
- 请求输入保留 `gender` 或 `manualGender` 字段，兼容现有前端字段。
- 返回任务文档包含 `gender`、`genderSource`、`genderReason`、`bookInfo.category`、`bookInfo.genre`。

**步骤：**

- [ ] 写失败测试：手工男女频进入任务后，不能被 bookinfo 或 AI 覆盖。
- [ ] 写失败测试：无手工值时先使用 bookinfo，只有冲突/未知才调用 AI；用可注入的 AI 判断函数计数调用次数。
- [ ] 写失败测试：AI 无法判断时返回“待确认”，任务不显示为已确定男频/女频。
- [ ] 运行 HTTP/服务定向测试确认先失败。
- [ ] 实现最小接线，保留现有任务状态和审计字段。
- [ ] 运行：`go test ./internal/httpapi ./internal/novelfetchworkshop`。

### 任务 5：接入 Node 兼容桥，保持前端单端口

**文件：**
- 修改：`app.js` 或现有 Node 小说工作台桥接路由
- 修改：Node 侧对应测试
- 不修改：`frontend/src` 页面布局和 UI 样式

**步骤：**

- [ ] 写失败测试：Node 本地请求小说获取时必须通过 `QIANTIE_GO_BASE_URL` 指向本地 Go，不得请求公网项目服务器。
- [ ] 写失败测试：桥接请求包含现有 `QIANTIE_BRIDGE_SECRET` 签名/认证。
- [ ] 运行 Node 定向测试确认先失败。
- [ ] 实现兼容桥，将 Go 返回的字段原样传给现有前端；不在 Node 中重复分类判断。
- [ ] 运行 Node 相关测试和 `go test ./...`。

### 任务 6：Worker 解耦与本地验证

**文件：**
- 修改：`docker-compose.v88-review.yml`
- 修改：本地启动说明配置（不写密钥）
- 不删除：`services/121-browser-worker`

**步骤：**

- [ ] 写配置测试：基础模式启动不依赖 `novel-fetch-121-worker`；完整模式仍可显式启动 Worker。
- [ ] 运行配置测试确认先失败。
- [ ] 将 Worker 从基础平台强制依赖中拆出；Go 小说正文接口不依赖 Worker。
- [ ] 启动基础模式并验证：MySQL 健康、Redis 健康、Go `/health` 返回 200、Node 桥接能读取小说正文。
- [ ] 启动完整模式时单独验证 Worker，不把 Worker 失败误报成 Go 失败。

### 任务 7：回归与提交边界检查

**文件：**
- 只包含本次迁移涉及的 Go、Node 桥接、配置测试文件

**步骤：**

- [ ] 运行 `go test ./...`。
- [ ] 运行相关 Node 测试。
- [ ] 使用测试服务器验证 121 请求不会把 API Key、Cookie 或本地密钥写入日志/响应。
- [ ] 检查 `git diff --stat` 和 `git status --short`，确认不包含 `.env.v88-review`、`.local-review/`、`frontend/dist`、构建日志。
- [ ] 记录测试结果和仍依赖外部 121 服务的边界。
- [ ] 用户确认后再单独提交本次迁移；未获确认前不提交、不推送。
