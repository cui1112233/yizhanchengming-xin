# AI-CUT 创作画布实施计划

约束：直接维护 v88；沿用 backend/frontend；后端系统提示词仍在 prompt；不调用收费服务、不推送、不公开部署。已批准设计见同日期 creative-canvas-design。

## A 服务端绑定与素材送入

新增 `backend/internal/httpapi/aicut_creative_canvas.go` 和对应测试，修改 router 注册。

- `POST /api/ai-cut/projects/{id}/creative-canvas`：鉴权、验证剪辑工程存在，稳定派生创作 project ID，幂等创建/读取已有 AgentWorkspace 项目，返回 `{taskId,canvas}`。绑定工程自带故事、角色、场景、道具、分镜文本节点与引用连线，均使用现有有效 config 结构。
- `POST /api/ai-cut/projects/{id}/creative-canvas/assets`，输入 `{assetIds:[...]}`，1–30项，来源只能是绑定创作项目。先检查全部来源，复制真实内容、探测时长、以稳定目标ID保存到 AI-CUT。重复调用返回已有资源；失败返回具体原因；不写时间线。响应 `{assets:[...]}`。
- `POST /api/ai-cut/projects/{id}/creative-canvas/script`，输入 `{nodeIds:[...],expectedRevision:N}`，1–10项已保存且非空的文本或分镜节点；明确追加到当前时间线剧本，保留原剧本、轨道和标记，版本不符拒绝。独立于素材复制，不会静默覆盖或重复提交。
- 先写路由测试确认失败，再实现；测试绑定重复、跨用户隔离、无效来源、真实对象内容及重复送入。用内存 store 与对象存储，禁用付费服务。

## B 真实画布前端

新增 `frontend/src/user/components/AICutCreativeCanvas.jsx` 与 CSS；新增 `frontend/src/user/pages/aicutCreativeCanvas.js` 纯函数和测试；修改 API 文件与 AICutPage 集成（由主执行者）。

- 组件接口 `{projectId,projectTitle,onOpenEditor,onTransferred}`。绑定 API 返回 taskId 后，使用现有 AgentCanvas/getAgentCanvas/saveAgentCanvas/createCanvasSaveQueue。加载/错误/保存中/冲突明确可见，卸载或工程变化不能污染新工程。
- 可添加故事、角色、场景、道具、分镜卡，复用有效文本/script节点。使用 AgentCanvas 已有生成、引用、连线、缩放、撤销流程。
- 用按钮和素材选择面板调用真实素材送入 API，显示真实名称、种类、选择数量和结果；完成调用 onTransferred，进入专业剪辑。空素材明确指引先上传/生成，不伪造占位素材。
- 纯函数测试覆盖初始卡片结构与ID、选择去重和默认可用媒体选择。React build 验证 JSX；主执行者用实际页面验收拖动、保存刷新、素材送入。

## C 集成与验收

主执行者用懒加载组件替换 AICutPage 四步引导区域，保留专业剪辑 Agent，细化“准备素材/剪已有素材”文案；成功送入后重新读取当前工程素材，不自动修改时间线。分别复查 A/B，再进行整体复查、相关回归、React build、Go embed、4001 runtime-build-info 和实际浏览器验收。提交仅本轮改动和验收文档。
