# React 页面迁移设计：小说获取、小说面板与 Go 统一链路

## 目标

在不丢失公网页面已有功能的前提下，把小说获取、小说面板从 `public` 静态工作台/iframe 迁移为正式 React 页面，并统一调用 Go 后端；迁移完成后由 Go `embed` 发布 React 构建产物，旧 Node/旧静态接口不再作为正式入口。

## 已确认的基线

- 远端维护基线是 `v88`，不能按历史分支名称选择源码。
- 公网 `http://10.0.10.115:8081/` 是行为参考，当前公开版本的任务阶段包括原文抓取、清洗、配音时长、智能统一、人物资产、导演分镜、故事板、视频生成、下载和合成。
- 本地 Go 正式接口以 `/api/batch-factory/v12/*` 和 `/api/novel-fetch-workshop/*` 为准。
- `public/batch-rewrite` 与 `public/novel-panel/workbench` 目前仍是旧工作台资源；React 页面仍通过 iframe 使用它们。

## 统一规则

1. React 页面是唯一正式页面入口；不得再由页面直接依赖 Node 服务。
2. Go 是唯一业务接口实现；前端不能硬编码提示词，提示词由 Go 返回。
3. 公网页面只作为交互和阶段行为参考，不直接复制其 `/api/tasks` 接口层。
4. 所有阶段必须保留可追踪状态、错误原因、重试入口和原始数据。
5. 迁移期间可以保留兼容转发，但不能让新页面同时调用两套业务接口。
6. 只有 React 页面和 Go 接口完成验收后，才能删除旧 `public` 工作台资源。

## 页面迁移范围

### 小说获取

- 当前入口：`frontend/src/user/pages/NovelFetchPage.jsx` 的 `iframe /batch-rewrite/v2`。
- 目标入口：同一 React 路由直接渲染任务、配置、知识库、处理规则、提交和日志页面。
- 主要 Go 接口：`/api/novel-fetch-workshop/*`，必要的批量接入使用 `/api/batch-factory/v12/intakes/novel-fetch`。
- 必须覆盖：书籍 ID 添加、原文抓取、内容范围、平台/书城、风格与男女频判断、原文/AI 版本、敏感词规则、知识库、网站提交和失败状态。

### 小说面板

- 当前入口：`frontend/src/user/pages/NovelPanelPage.jsx` 的 `iframe /novel-panel/workbench`。
- 目标入口：React 页面直接承载项目、人物/场景/道具资产、剧本、分镜和导出功能。
- 主要 Go 接口：`/api/novel-panel/*`，资产文件使用 `/api/novel-panel/reference-assets/*`。
- 必须覆盖：资产缩略图、`@` 资产选择、角色/场景/道具分组、AI 分析、编辑保存、历史和导出。

## 执行顺序

1. 固化公网功能与本地 Go 接口对应表。
2. 先切换小说获取 React 正式页面，保留旧 iframe 资源但不再作为默认入口。
3. 验证小说获取全流程后，再迁移小说面板。
4. 将 `public` 资源从正式构建输入中移除，只保留确实仍被静态引用的品牌资源。
5. 重新构建 Go `embed` 二进制并做本地和公网等价验收。

## 验收标准

- `/novel-fetch` 和 `/novel-panel` 页面 DOM 中不再出现旧工作台 iframe。
- 页面请求只命中 Go 正式接口，不命中 Node 端口或 `/batch-rewrite` 旧业务接口。
- 小说获取能够完成：添加书籍 → 抓取原文 → 判断平台/风格/男女频 → 处理规则 → 选择版本 → 进入批量工厂。
- 小说面板能够完成：载入项目 → 资产显示 → `@` 选择带图/无图资产 → 保存 → 导出。
- Go 测试、前端测试和单个 Go 二进制启动检查全部通过。
