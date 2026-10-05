# ECS / v88 Batch Factory 能力审计（Task 9.1.5）

本文件用于把旧公网/ECS 能力区分为：**存在 / 可用 / 已坏 / 未验证**。

> 重要：`源码存在` 不等于 `公网可用`。本审计把源码/路由证据和运行证据分开记录。

审计日期：2026-10-05

参考资料：

- `docs/migration/v88-batch-factory-v11-frontend-inventory.md`
- `docs/migration/v88-batch-factory-v11-backend-api-inventory.md`
- `docs/migration/v88-batch-factory-v11-status-model-inventory.md`
- `docs/migration/v88-batch-factory-video-provider-inventory.md`
- 旧公网历史实测记录

## 1. 状态定义

| 状态 | 定义 |
|---|---|
| `存在` | v88 源码、路由、组件或 Go handler 明确存在，但没有足够运行证据证明当前可用 |
| `可用` | 有明确成功运行/接口实测证据 |
| `已坏` | 有明确失败实测证据，且尚无后续成功证据证明已经修复 |
| `未验证` | 需要当前登录态、账号凭据、在线执行器、ECS shell 或公网实机访问，本轮无法安全确认 |

同一能力可能同时是“源码存在 + 当前运行未验证”。本表的 `运行分类` 以运行证据为准。

## 2. 本轮实时探测限制

本轮尝试直接读取：

- `http://115.190.156.223:3000/`
- `/batch-factory`
- `/api/build-info`
- `/api/runtime-build-info`
- `/api/batch-factory/v12/capabilities`
- `/api/batch-factory/v11/video-provider/status?provider=personal_api`

当前外部网页抓取通道无法访问这个 **裸 IP + HTTP** 目标。

这只代表当前审计工具无法直连，不代表 ECS 宕机。因此：

- 不把“当前抓取不到”标成 `已坏`
- 只有以前真实复现过的失败才标成 `已坏`
- 其它需要实时运行验证的项标为 `未验证`

最终“当前公网可用”仍必须在 Task 16 用真实 ECS/浏览器回归确认。

---

# 3. 核心页面与入口

| 能力 | 源码/路由 | 源码存在 | 最近运行证据 | 运行分类 | 后续动作 |
|---|---|---:|---|---|---|
| 公网首页 | `/` | ✅ | 历史公网可打开并进入业务页面；本轮无法直连裸 IP | 未验证 | Task 16 重新实测 |
| 水货生产 | `/shuihuo-production` | ✅ | 历史公网页面存在并被多次实际使用；本轮未复测完整链 | 未验证 | Task 11 迁移；Task 16 实测 |
| Batch Factory 主入口 | `/batch-factory` | ✅ | 历史公网存在；曾进入批量工厂排查功能 | 未验证 | Task 9 实现后 Task 16 对标 |
| Batch Factory 预览入口 | `/batch-factory-preview` | ✅ | 源码存在，无当前运行证据 | 未验证 | 迁移时判断是否保留兼容入口 |
| 小说获取 | `/novel-fetch` | ✅ | 历史页面实际使用过，但存在字段缺失/已选0个问题 | 已坏 | Task 8 新主线已重写基础链；Task 16 对标 |

---

# 4. 小说获取 / 121 链路

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| 121 正文直取上游 `txt.121w.com/api.php` | ✅ | Task 4 已验证接口契约和成功响应 | 可用 | 这是上游接口能力，不等于旧 ECS UI 全流程可用 |
| ECS 小说获取路由 | ✅ | 历史公网实际执行过 | 存在 | 当前实时未复测 |
| 121 bookinfo 补书名/category/genre | ✅ | 源码明确实现；历史页面曾出现书名/平台等不显示 | 未验证 | 上游能力与页面显示需分开 |
| 男女频补全 | ✅ | 源码存在；历史公网曾出现男女频不显示 | 未验证 | 不能把源码存在等同页面可用 |
| 风格补全 | ✅ | 源码存在；历史公网曾出现风格不显示 | 未验证 | 后续新主线以 MySQL 字段为事实源 |
| 小说获取结果展示 ID/书名/平台/男女频/风格 | ✅ | 历史实测出现“不显示” | 已坏 | 新 Task 8 已在新仓库重做，但旧 ECS 仍不算修复 |
| 提交后选中数量 | ✅ | 历史实测提交后仍显示“已选 0 个” | 已坏 | 属于旧公网交互缺陷 |
| “解析输入”旧入口 | ✅ | 历史公网仍存在，而需求已要求移除 | 已坏 | Task 10 由“版本对应配置档”替换 |

---

# 5. Batch Factory V11/V12 核心服务

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| V12 浏览器公共 API | ✅ | v88 当前前端 BASE 已指向 `/api/batch-factory/v12` | 存在 | 当前公网实时 API 未复测 |
| Node V12 → V11 rewrite | ✅ | 源码确认 | 存在 | 当前桥接健康未验证 |
| V11 Go Batch/Book/Settings | ✅ | Go handler/Store 明确存在 | 未验证 | 当前 ECS Go 服务健康未知 |
| V11 Go Director/Hook/Stages | ✅ | Go handler 明确存在 | 未验证 | 需带真实文本模型运行 |
| V11 Prompt Compiler | ✅ | Go handler 明确存在 | 未验证 | 需真实 Batch/Book/VIDEO 数据 |
| V11 Production | ✅ | Go ProductionService 明确存在 | 未验证 | provider/Go bridge 曾出现故障 |
| Runtime summary | ✅ | V12 Node 聚合实现存在 | 未验证 | 依赖 V11 runtime-index、production、merge、stage |
| Automation controller | ✅ | Node 实现存在 | 未验证 | 旧自动化事实源与新 Run.run_at 后续需收敛 |
| Schedule Router | ✅ | Node `/v11/schedules` 存在 | 未验证 | 后续不应与新 MySQL Run.run_at 双轨长期存在 |

---

# 6. personal_api / 视频 Provider

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| `personal_api` provider | ✅ | 历史实际调用 status 时返回 503 | 已坏 | 至少历史公网该路径不可用；当前是否恢复未验证 |
| `yd2.0-mini` | ✅ | provider/model 映射明确存在 | 未验证 | 需要账号真实 API Key 与生产请求 |
| `GET .../video-provider/status?provider=personal_api` | ✅ | 历史实测 `503 Batch Factory V11 Go service unavailable` | 已坏 | Go handler 未配置本应 200 configured=false，因此不是单纯缺 Key |
| Provider config sync | ✅ | Node 代码存在 | 未验证 | 依赖 Node→Go bridge 与账号凭据 |
| VIDEO submit | ✅ | ProductionService 实现存在 | 未验证 | 需真实 provider 请求 |
| VIDEO polling | ✅ | Yadi/AutoDL/YFAI/Local adapters 均有 Poll | 未验证 | 需真实 providerTaskId |
| VIDEO cancel | ✅ | 生产服务有取消语义 | 未验证 | 只有 provider 真接受才可 cancelled |
| `doubao_local_executor` | ✅ | Go/Node/前端接口存在 | 未验证 | 必须有当前账号已配对且在线执行器 |
| `autodl_comfyui` / H3 | ✅ | Adapter + V12 H3 kernel 存在 | 未验证 | 需要 H3 Key/真实 workflow |
| `yfai_seedance` | ✅ | Adapter/模型映射存在 | 未验证 | 需要账号 credential |
| Local executor artifact | ✅ | 受保护媒体路由存在 | 未验证 | 依赖在线 executor 和 artifact store |

### personal_api 503 结论

旧源码确认：

- Go 的 provider status 对“未配置”本应返回 HTTP 200 + `configured:false`
- Node 对非 4xx 上游故障会统一显示 `Batch Factory V11 Go service unavailable`

因此历史 503 更可能来自：

- Go V11 服务不可达
- Bridge secret/签名或路由问题
- provider config 同步失败
- production slice/service 未启用

Task 9.1.5 只记录故障事实，不在没有 ECS 实机证据时臆断唯一根因。

---

# 7. Director / Hook / 最终提示词

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| Hook 生成 | ✅ | Node+Go API 存在 | 未验证 | 需要真实文本模型和书籍 |
| Hook approve | ✅ | Go API 存在 | 未验证 | 需真实 hookId |
| Director 单书 | ✅ | Node+Go API 存在 | 未验证 | 需文本模型 |
| Director 批量 | ✅ | Node+Go API 存在 | 未验证 | 需批次数据 |
| opening variants | ✅ | Go + Node 后处理存在 | 未验证 | 需换开头配置 |
| H3 Director | ✅ | 原生 V12 Go 路由存在 | 未验证 | 需 H3 输入/模型 |
| Final Prompt | ✅ | PromptCompilerService + HTTP API 存在 | 未验证 | 需有效 VIDEO |
| Effective Settings | ✅ | Go API 存在 | 未验证 | 需有效 Batch/Book/VIDEO |

---

# 8. 配置 / 提示词 / 权限

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| `/api/presets` 系统预设 | ✅ | 公网页面/后台曾实际读取 | 存在 | 当前完整运行未复测 |
| `/api/script-constraint-prompts?category=prefix` | ✅ | 历史先出现 404；修路由后又出现 Unauthorized invalid/expired token | 已坏 | 需 Task 15 统一认证，不应简单全局白名单 |
| Batch settings | ✅ | V12→V11 保存链存在 | 未验证 | 需真实保存刷新回归 |
| Book settings | ✅ | V12→V11 保存链存在 | 未验证 | 需真实保存刷新回归 |
| Video override | ✅ | V11 API 存在 | 未验证 | 需实际 VIDEO |
| Config versions | ✅ | V11 API 存在 | 未验证 | Task 10 会重新定义“版本对应配置档” |
| 生产统一设置 Drawer | ✅ | v88 有组件 | 未验证 | 新需求要求不跳页、右侧宽 Drawer |
| 发布统一设置 Drawer | ✅ | v88 有组件 | 未验证 | 新需求要求不跳页、右侧宽 Drawer |

---

# 9. 登录 / Auth

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| 登录 API | ✅ | 历史公网可登录，但后续出现状态过期跳转 | 存在 | 登录本身曾成功，不代表会话保持健康 |
| Token/会话保持 | ✅ | 历史实测出现“当前登录状态已过期，已切换到登录页面”闪退 | 已坏 | Task 15 专门处理 |
| Node apiAuth | ✅ | 大量路由依赖 | 存在 | 当前 token 兼容未验证 |
| Node→Go BridgeAuth | ✅ | Go Router 全面依赖 | 未验证 | personal_api 503 表明至少历史上桥接链存在故障可能 |
| 用户/团队权限 | ✅ | memberStore/权限代码存在 | 未验证 | 需多账号实际回归 |

---

# 10. Merge / 存储 / 发布

| 能力 | 源码存在 | 最近运行证据 | 运行分类 | 说明 |
|---|---:|---|---|---|
| MergeService | ✅ | Go 实现存在 | 未验证 | 需真实视频文件 |
| 本地 ffmpeg 合并 | ✅ | Go merge/local artifact 代码存在 | 未验证 | 需 ECS ffmpeg/runtime |
| TOS merged output | ✅ | TOS adapter 路径存在 | 未验证 | 需真实 TOS credential |
| Merge cover | ✅ | ffmpeg 截帧实现存在 | 未验证 | 需成功 merge artifact |
| 121 发布 | ✅ | Node 121 session/publisher + Go external 能力存在 | 未验证 | 需真实 121 登录会话 |
| External publish intent/audit | ✅ | Go API 存在 | 未验证 | 需 provider credential |
| 发布权限校验 | ✅ | 历史逻辑存在 | 未验证 | Task 15 重做统一权限 |

---

# 11. 当前能力汇总

## 11.1 历史明确 `已坏`

1. `personal_api` provider status：历史 503 `Batch Factory V11 Go service unavailable`
2. `/api/script-constraint-prompts?category=prefix`：历史 404，随后 Unauthorized / token 问题
3. 登录会话保持：历史“登录状态已过期”闪退
4. 小说获取结果字段：历史 ID/书名/平台/风格/男女频显示异常
5. 小说提交选择计数：历史“已选 0 个”
6. “解析输入”仍残留，与已批准新交互冲突

这些都不能因为源码里存在对应功能就标为可用。

## 11.2 源码明确 `存在`，但当前运行 `未验证`

- Batch Factory V12/V11 主链
- Batch/Book/Settings
- Hook/Director/Stages
- Final Prompt
- Automation/Schedule
- Merge
- TOS
- 121 publish
- Doubao local executor
- H3 / AutoDL
- YFAI Seedance
- 用户/团队权限

## 11.3 有明确成功证据的能力

- 121 正文上游直取 API：已在 Task 4 验证成功契约
- 旧公网登录曾可完成登录动作，但会话保持存在后续故障，因此本表不把整个登录体系标为“可用”

---

# 12. 迁移优先级影响

本审计对后续任务的直接约束：

### Task 9

先迁：

- BatchProject 列表
- 真实 MySQL 项目/书籍
- 项目详情
- Run/Book/Stage 细状态

不要先把旧 Node 多层兼容全部搬入新仓库。

### Task 10

必须解决：

- 移除“解析输入”
- 新增“版本对应配置档”
- 统一设置 Drawer
- 配置事实源统一

### Task 14

必须解决：

- provider config/status 不再依赖易失内存 registry
- personal_api status 未配置 != 503
- provider/model 强绑定
- local executor 在线状态
- VIDEO submit/poll/retry/cancel

### Task 15

必须解决：

- 登录状态过期闪退
- token 生命周期
- `/api/script-constraint-prompts` 认证兼容
- 发布权限

### Task 16

必须重新验证全部“未验证”和“历史已坏”项，并用真实公网结果更新本审计。

---

# 13. Task 9.1.5 验收结论

本任务完成的是**能力现状分类**，不是修复这些历史故障。

已完成：

- 区分源码存在和运行可用
- 标记历史明确失败项
- 标记需要当前 ECS/登录/执行器/凭据才能验证的项
- 不把无法直连裸 IP 的工具限制误判为 ECS 宕机
- 建立后续 Task 9/10/14/15/16 的验证优先级

下一步 Task 9.1.6 将基于 9.1.1–9.1.5 的盘点结果，建立 **旧 v88 模块 -> 新 Go/React 模块** 的迁移映射表。