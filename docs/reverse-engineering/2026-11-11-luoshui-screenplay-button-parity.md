# 洛水 v2.9.0「剧本创作」按钮与链路逐项对照表

日期：2026-11-11  
一战晟铭入口：`/comic-screenplay`，产品名显示为「漫剧创作」  
证据基线：`2026-11-11-luoshui-screenplay-evidence.md`、`前端桥接方法全集.md`、洛水已登录页面截图  

## 判定规则

- **已打通**：按钮调用真实 API，数据写入 Store/MySQL，刷新后仍存在，并有正常/空值/错误测试。
- **部分打通**：有真实模型或数据调用，但缺少洛水同类的编辑、取消、历史或导出链路。
- **未实现**：界面不存在，或只有状态/图标而没有对应业务执行器。
- **不直接照搬**：洛水私有服务、密钥、供应商账户和桌面专属文件能力不会复制；改接一战晟铭自己的模型、TOS、Redis、MySQL。

## 1. 项目库与新建

| 洛水操作 | 输入 | 应产生的真实变化 | 一战晟铭状态 | 对应接口/说明 |
| --- | --- | --- | --- | --- |
| 个人作品 | 当前账号 | 读取本人项目 | 已打通 | `GET /api/comic-screenplays/projects` |
| 团队作品 | 团队身份 | 读取团队共享项目及权限 | 未实现 | 当前只显示空态，不能算完成 |
| 搜索作品 | 关键词 | 过滤项目列表 | 已打通 | 前端本地过滤已加载列表 |
| 时间升/降序 | 排序方向 | 项目顺序变化 | 已打通 | 前端按 `updated_at` 排序 |
| 网格/列表 | 视图模式 | 布局变化 | 已打通 | 只改显示，不改数据 |
| 历史 | 历史范围 | 打开删除/归档/旧项目记录 | 未实现 | 当前按钮没有链路 |
| 新版/旧版 | 编辑器版本 | 切换对应工作台 | 未实现 | 当前“新版”禁用，“旧版”仅显示 |
| 创建剧本 | 名称、模型、提示词、画风、比例、解析档、TXT/DOCX | 创建项目并解析文件 | 已打通 | `POST /projects` + `POST /source/parse` |
| 取消创建 | 无 | 关闭弹窗且不写库 | 已打通 | 弹窗本地取消；洛水取消前后数据库哈希不变 |
| 删除项目 | 项目 ID、确认 | 软删/删除项目并从列表消失 | 已打通 | `DELETE /projects/{id}`；一战晟铭当前为删除项目数据 |
| 打开项目 | 项目 ID | 读取项目、版本、场次、资产、分镜、批次 | 已打通 | `GET /projects/{id}/workspace` |

## 2. 剧本与章节

| 洛水操作/能力 | 输入 | 输出/数据变化 | 一战晟铭状态 | 证据接口 |
| --- | --- | --- | --- | --- |
| 上传 TXT/DOCX | 文件 | 解析为原文；空文件/损坏文件报错 | 已打通 | `select_script_file`；本项目 `/source/parse` |
| 普通剧本推理 | 原文、模型、提示词 | 新剧本版本、分集和场次 | 已打通 | `screenplay_ai_inference`；本项目 `/generate` |
| 大师剧本推理 | 原文、模型、资产解析档 | 结构化剧本及更细资产字段 | 部分打通 | `screenplay_master_inference`；当前用资产解析档控制资产提示词，但未完整复刻大师字段 |
| 大师分镜推理 | 剧本、风格、模型 | 分镜结构 | 部分打通 | `screenplay_master_storyboard_inference`；当前有 CM 分镜生成 |
| 取消推理 | 运行中任务 | 中止模型调用、状态变取消 | 部分打通 | `cancel_screenplay_ai_inference`；前端 AbortController 可取消 HTTP，缺少持久化任务取消 |
| 章节加载/保存 | 章节数组 | 章节持久化 | 未实现 | `load_screenplay_chapters` / `save_screenplay_chapters` |
| 章节计划 | 集数、字数、范围 | 可编辑章节计划 | 未实现 | 洛水章节计划接口族 |
| 保存剧本 | 当前文本/结构 | 新版本或覆盖草稿 | 部分打通 | 生成时保存版本；没有手动编辑保存按钮 |
| 版本历史 | 项目 ID | 列出并恢复旧版本 | 部分打通 | 后端保存多版本并返回列表；前端尚无恢复按钮 |
| AI 配置 | 模型/参数 | 项目级配置持久化 | 部分打通 | `load/save_screenplay_ai_config`；创建时保存模型，工作台设置尚未保存 |
| 显示配置 | 面板/字号/布局 | 下次打开恢复 | 未实现 | `load/save_screenplay_display_config` |

## 3. 场次/镜头按钮

| 按钮 | 输入 | 输出/边界 | 一战晟铭状态 | 对应接口 |
| --- | --- | --- | --- | --- |
| 新增场次 | 集数、标题、内容、角色、场景、道具 | 插入到该集末尾；标题和内容不能同时为空 | 已打通 | `create_screenplay_shot`；`POST /projects/{id}/shots` |
| 编辑场次 | 同上 | 覆盖该镜头字段，刷新后保留 | 已打通 | `update_screenplay_shot`；`PATCH /projects/{id}/shots/{shotId}` |
| 上移 | 镜头 ID | 与同集上一条交换顺序；首条提示边界 | 已打通 | `move_screenplay_shot`；`POST .../move {direction:up}` |
| 下移 | 镜头 ID | 与同集下一条交换顺序；末条提示边界 | 已打通 | `move_screenplay_shot`；`POST .../move {direction:down}` |
| 删除 | 镜头 ID、二次确认 | 删除后刷新消失；重复删除返回未找到 | 已打通 | `delete_screenplay_shot`；`DELETE .../shots/{shotId}` |
| 指定位置插入 | 前一镜头 ID | 插入到指定镜头之后并重排 | 未实现 | `insert_screenplay_shot_after` |
| 取消编辑 | 无 | 关闭弹窗，不写数据 | 已打通 | 前端 Modal 取消 |

## 4. 角色、场景、道具和音色

| 操作 | 输入/输出 | 一战晟铭状态 | 洛水证据接口 |
| --- | --- | --- | --- |
| 批量提取角色 | 剧本、解析档、模型 -> 角色资产 | 已打通 | 资产解析链；本项目 `/assets/extract` |
| 批量提取场景 | 剧本、解析档、模型 -> 场景资产 | 已打通 | 同上 |
| 批量提取道具 | 剧本、解析档、模型 -> 道具资产 | 已打通 | 同上 |
| 新增/编辑/删除角色 | 角色资料、描述、提示词、音色、出现集数 | 已打通（参考图待补） | 独立实现 `POST/PATCH/DELETE /projects/{id}/assets` |
| 新增/编辑/删除场景 | 场景资料、描述、提示词、出现集数 | 已打通（参考图待补） | 独立实现 `POST/PATCH/DELETE /projects/{id}/assets` |
| 新增/编辑/删除道具 | 道具资料、描述、提示词、出现集数 | 已打通（参考图待补） | 独立实现 `POST/PATCH/DELETE /projects/{id}/assets` |
| 保存资产图片 | 资产 ID、图片 | 未实现 | `save_screenplay_item_image` |
| 读取/绑定音色 | 角色、音色 ID、设置 | 未实现 | `get_screenplay_voices` / `update_storyboard_voice` |
| Agent 资产加载/保存 | 会话资产 | 未实现 | `screenplay_load_agent_assets` / `screenplay_save_agent_assets` |

## 5. 分镜编辑

| 操作 | 输入/输出 | 一战晟铭状态 | 洛水证据接口 |
| --- | --- | --- | --- |
| 生成分镜 | 剧本、集数、风格、模型 -> 分镜组 | 已打通 | `storyboard_ai_inference`；本项目 `/storyboards/generate` |
| 取消生成 | 任务 ID -> cancelled | 部分打通 | `storyboard_ai_cancel`；当前只取消前端请求 |
| 查看进度 | 任务 ID -> 进度/错误 | 未实现 | `storyboard_ai_get_progress` |
| 新增分镜 | 前一项、分镜数据 | 未实现 | `add_storyboard_after` / `add_storyboard_after_by_path` |
| 编辑分镜 | 分镜字段 | 未实现 | `update_storyboard` / `update_storyboard_by_path` |
| 删除分镜 | 分镜 ID | 未实现 | `delete_storyboard` / `delete_storyboard_by_path` |
| 拆分分镜 | 分镜 ID、拆分点 | 未实现 | `split_storyboard` / `split_storyboard_by_path` |
| 合并分镜 | 多个分镜 ID | 未实现 | `merge_storyboards` / `merge_storyboards_by_path` |
| 批量调整 | 调整规则 | 未实现 | `batch_update_storyboards_by_path` / `apply_storyboard_adjustments_by_path` |
| 补全 Agent | 缺失字段/分镜 -> 补全结果 | 未实现 | `run_storyboard_completion_agent` |

## 6. AI 对话与 Agent

| 操作 | 输入/输出 | 一战晟铭状态 | 洛水证据接口 |
| --- | --- | --- | --- |
| AI 对话 | 消息、上下文、模型 -> 流式回复 | 未实现 | `screenplay_ai_chat` |
| 取消对话 | 会话/任务 ID | 未实现 | `screenplay_ai_chat_cancel` |
| 对话历史 | 项目 ID -> 历史；保存历史 | 未实现 | `load/save_screenplay_chat_history` |
| Agent 执行 | 目标、技能、项目上下文 -> 操作记录 | 未实现 | `screenplay_agents_run` |
| Agent 取消 | 任务 ID | 未实现 | `screenplay_agents_cancel` |
| AI 结果读取/解析/修改 | 结果 ID/JSON | 未实现 | `get/load/parse/update_screenplay_ai_result*` |

## 7. 图片、视频、字幕和旁白

| 操作 | 输入/输出 | 一战晟铭状态 | 洛水证据接口 |
| --- | --- | --- | --- |
| 单张分镜图 | 分镜提示词、模型 -> 图片 | 未实现 | `generate_storyboard_image` |
| 批量分镜图 | 分镜列表、模型 -> 图片任务 | 未实现 | `generate_storyboard_images_batch` |
| 即梦单张/批量 | 分镜、即梦配置 -> 图片/任务 | 不直接照搬 | `screenplay_jimeng_generate_single/batch`；应改接自己的图片模型目录 |
| 取消图片任务 | 任务 ID | 未实现 | `screenplay_jimeng_cancel` |
| 宫格图拆分 | 图片/路径 -> 单格图 | 未实现 | `split/ai_split_screenplay_collection_image*` |
| 分镜图复制/编辑/删除 | 图片 ID、编辑参数 | 未实现 | `duplicate/edit/delete_storyboard_image*` |
| 视频结果保存/删除/排序 | 镜头视频、顺序 | 未实现 | `update_shot_video_result` / `delete_screenplay_video_by_path` / `update_screenplay_video_sequence` |
| 旁白/配音 | 文本、音色 -> 音频 | 未实现 | `add_voice_to_storyboard` / `update_storyboard_voice` |
| 合成字幕 | 剧本/旁白时间 -> 字幕 | 未实现 | `screenplay_generate_combined_subtitle` |

## 8. 任务、生产和导出

| 操作 | 输入/输出 | 一战晟铭状态 | 说明 |
| --- | --- | --- | --- |
| 启动生产 | 项目 ID -> 批次状态 | 部分打通 | 当前只持久化九步状态，没有真正执行图片/视频/旁白 Worker |
| 暂停/继续 | 批次 ID -> paused/running | 已打通状态 | 状态真实保存，但执行器尚未完整接入 |
| 定时任务读取/保存/取消 | 计划、任务 ID | 未实现 | `load/save/cancel_screenplay_scheduled_tasks` |
| 任务状态 | 任务 ID -> 进度/失败原因 | 部分打通 | 当前仅批次状态与日志 |
| 导出项目包 | 项目 ID -> JSON | 已打通 | 一战晟铭独立格式 |
| 导出剪映 | 分镜、媒体、字幕 -> 剪映草稿 | 未实现 | `export_storyboard_to_jianying`；网页端需生成可下载草稿包，不能调用用户本机剪映 |

## 9. 当前可验证调用链

```text
漫剧创作
  -> 项目库 GET /projects
  -> 新建弹窗 GET /prompts + 一战晟铭模型目录
  -> 创建 POST /projects
  -> 文件解析 POST /source/parse
  -> 剧本生成 POST /generate
  -> MySQL 保存 Document + Shot + Character + Scene
  -> 工作台 GET /workspace
  -> 场次新增/编辑/移动/删除 API
  -> 资产提取 POST /assets/extract
  -> 分镜生成 POST /storyboards/generate
  -> 批次状态 POST/GET /batch
  -> JSON 导出 GET /export
```

## 10. 不能宣称完成的范围

目前不能宣称“洛水剧本创作一比一全部完成”。未完成的关键用户链路是：团队作品、历史/新版、章节编辑与恢复、资产参考图、分镜 CRUD/拆分/合并、AI Chat/Agent、图片/视频/旁白/字幕生产、真实 Redis Worker、任务进度与取消、剪映草稿导出。后续实现必须继续以本表为验收清单，不能用图标或空弹窗代替。

## 11. 2026-11-12 增量核查

- 项目库“历史”已改为真实最近创作列表；它不是删除/归档回收站。
- 剧本正文已可手工编辑并保存为新版本，版本历史可恢复；恢复不会覆盖旧版本，场次、人物、场景会一起复制。
- 项目设置已可保存到后端，带 revision 冲突保护。
- 生产批次已接入逐步执行器和在线视频生成通道；仍未补齐图片、旁白、字幕以及 Redis Worker，所以不能写成“全部生产完成”。
- 团队作品与新版编辑器没有真实数据层，已改为明确不可用，不再用空列表或可点击空壳冒充。
- 最新逐按钮验收表见 `docs/acceptance/comic-screenplay-button-audit-2026-11-12.md`。
