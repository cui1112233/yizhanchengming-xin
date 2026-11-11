# 插件界面一致性整改收尾结论（2026-11-11）

对照依据：[plugin-ui-consistency-audit-2026-11-11.md](./plugin-ui-consistency-audit-2026-11-11.md)
洛水证据目录：`F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\shots\`
洛水 bundle：`F:\脚本测试\work\洛水2026-v2.9.0-反编译分析\recovered\frontend\index-DszfAkKD.js`
V88 整改后截图：`docs/acceptance/ui-consistency-2026-11-11/after-2026-11-11/`

本轮只改源码、本地构建、重启 4001 验证；**未 commit、未部署、未调用任何生成模型**。

## 一、逐条改前 / 改后

| # | 条目 | 改前 | 改后 | 依据 |
| --- | --- | --- | --- | --- |
| 1 | AI-Subtitle 形态 | 插件中心内 720px 居中单栏 | 整页双栏：左参数列（`minmax(280px,340px)`）+ 右「AI 分析结果」列 | 洛水截图 `plugin-ai-subtitle-luoshui.png`；[DraftToolWorkbench.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/DraftToolWorkbench.jsx#L379-L543) |
| 2 | AI-Filler 形态 | 720px 单栏 | 整页双栏，右列「要删除的语气词」 | 洛水截图 `plugin-ai-filler-luoshui.png`；同一组件 |
| 3 | AI-Upscale 形态 | 720px 单栏，结果在底部状态条 | 整页双栏，右侧「进度和日志」+ 空态「等待开始 / 开始后，每一条视频的结果会在这里。」 | 洛水截图 `plugin-ai-upscale-luoshui.png`；[UpscaleWorkbench.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/UpscaleWorkbench.jsx#L203-L361) |
| 4 | AI-分段 形态 | 1080px 遮罩弹窗（`createPortal` + `.ai-segment-overlay/.ai-segment-modal`） | 独立整页双栏：左参数列 + 右「处理日志」，顶部「返回插件 / AI-分段 / 已授权 / 关闭」 | 洛水截图 `plugin-ai-segment-luoshui.png`；[AISegmentWorkbench.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/AISegmentWorkbench.jsx#L205-L300)；[ai-segment-workbench.css](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/ai-segment-workbench.css) |
| 5 | AI-分段 模式标签 | 无 Tab | 三个模式：按剪映主轨道 / 按剪映文本轨道 / AI分集（沿用洛水小按钮名） | 洛水截图 `plugin-ai-segment-luoshui.png` |
| 6 | AI-Subtitle 缺字段 | 只有 草稿 → 统计条 → 模型 → 校对剧本 | 补齐：选择模型 → 选择剪映草稿+刷新 → 时间线 → 字幕轨道+读取字幕 → 已读提示 → 剧本文本+选择文件/上传 → 已读字幕 → 开始分析 → 黄色提示；右列底部「确认修改（写回剪映字幕轨）」 | 洛水截图 `plugin-ai-subtitle-luoshui.png` 元素清单 |
| 7 | AI-Filler 缺字段 | 同上单栏 | 同 6 的字段顺序，右列标题「要删除的语气词」，底部「删除这些字词（写回剪映字幕轨）」 | 洛水截图 `plugin-ai-filler-luoshui.png` |
| 8 | AI-Subtitle / AI-Filler 只分析不写回 | 单次调用直接改动 | 两阶段：`开始分析` 只取 preview（不落库）→ `确认修改` 才 apply 写回剪映字幕轨 | 洛水截图；后端 [aicut_ai_routes.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/aicut_ai_routes.go) 的 `Preview`/`Corrections`/`Fillers` 分支 |
| 9 | AI-Upscale 视频输入 | 单文件 Upload | 「选择文件夹」+「多选视频」两个入口，队列展示 | 洛水截图 `plugin-ai-upscale-luoshui.png` |
| 10 | AI-Upscale 保存文件夹 | 无 | 「保存文件夹」+ 选择按钮；写入受跨域限制时提示改用结果链接 | 洛水截图 `plugin-ai-upscale-luoshui.png` |
| 11 | 分镜大师 / 爆款开头 / 角色修改 并发生成 | 三个面板都渲染「并发生成」开关 | 三条已删除；**小说转剧保留**（`concurrent: true` 未动） | 洛水三张截图均无此行；[pluginCatalog.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/pluginCatalog.js#L55-L133) |
| 12 | 角色修改结果区标题 | 已为「修改结果」 | 保持「修改结果」 | 洛水截图 `plugin-character-edit-luoshui.jpg` |
| 13 | 爆款开头输入占位 | 「请输入小说原文，或上传txt/word文档...」 | **未改**（见第三节） | 洛水 bundle 原文为「小说原文」 |
| 14 | 主按钮配色 | 统一珊瑚红 | 跟随各插件主题色：小说转剧 `#0ea795`、分镜大师/爆款开头/角色修改 `#0eb780`、AI-Subtitle `#22c55e`、AI-Filler `#fb7185`、AI-Upscale `#1a1a1a`、AI-分段 `#8b5cf6` | 洛水截图取色 + 小说转剧 bundle 主题色 |

配色通过 [PluginPanelShell.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/PluginPanelShell.jsx) 的 `primaryColor` 注入 `--pp-primary`，由 [PluginToolPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/PluginToolPage.jsx) 传入 `panel.buttonColor`，不再使用统一珊瑚红。

## 二、构建 / 重启 / 一致性校验（最后一次结果）

1. `npm run build`：成功（`frontend/dist`，入口 `assets/user-DFNpDdXb.js`）。
2. `scripts/build-go-local.ps1`：先把 `frontend/dist` 同步到 `backend/web/dist`，再 `go build`，产物 `backend/qiantie-local.exe`（13:11:53）。
3. 停 `V88-Local-4001` 计划任务 → 结束监听 4001 的进程 → 确认 `PORT_4001_FREE`、`qiantie-local` 进程数 0 → 重新启动任务。
4. 校验：
   - `GET /` → `200`
   - 服务端 index 入口 = 本地 `frontend/dist/index.html` 入口 = `/assets/user-DFNpDdXb.js`（`ENTRY_MATCH=True`）
   - 服务端下载该入口文件的 SHA256 与本地 dist 文件 SHA256 **一致**（`SHA_MATCH=True`）
   - 服务端返回的 `PluginCenterPage-*.js` chunk → `200`
   - 进程：`qiantie-local` PID 8412，13:12:02 启动

另：中途出现过一次「重启后仍返回旧 chunk」的现象，根因是磁盘上的 exe 已更新、但 4001 上跑的还是更早启动的进程镜像（旧 dist 的 `user-DDx5uzr2.js` / `PluginCenterPage-cKEWYWXU.js`）。结束旧进程并重新启动计划任务后消失，最终校验以第二节结果为准。

## 三、未改 / 未补项（如实记录）

1. **爆款开头输入占位**：洛水截图里读到的是「小原文」这类疑似截断文本；bundle 里能找到的原文是「小说原文」，无法据此判断洛水完整占位句。**洛水原文疑似截断，未改**，保持 V88 现有「请输入小说原文，或上传txt/word文档...」。
2. **AI-分段「开始编号」**：审计表 3.8 提到洛水参数含「开始编号」，本轮未补该字段（现有参数为 前 N 集、每集 分/秒、后续每集时长、分辨率、帧率）。
3. **AI-Subtitle / AI-Filler「时间线」**：洛水是输入框，V88 目前是随草稿自动带出的只读下拉（`主时间线 · rN`），语义一致但控件形态不同。
4. **AI-Subtitle / AI-Filler 字幕读取**：选择字幕轨道后自动读取，同时保留「读取字幕」按钮；洛水为手动点击触发，交互略有差异。
5. **AI-Upscale 保存文件夹**：浏览器跨域限制下无法直接写入目录，已提示改用结果链接下载（洛水为桌面端写入，能力边界不同）。
6. 资产参考、AI-CUT 不在本轮 P0/P1 范围，未改。

## 四、4001 真实实例截图

| 条目 | 截图 |
| --- | --- |
| 插件市场首页 | `after-2026-11-11/plugins-list-v88.jpg` |
| AI-Subtitle（整页双栏 + 全部字段） | `after-2026-11-11/ai-subtitle-v88.jpg` |
| AI-Filler（整页双栏 + 右列「要删除的语气词」） | `after-2026-11-11/ai-filler-v88.jpg` |
| AI-Upscale（整页双栏 + 保存文件夹 + 进度和日志） | `after-2026-11-11/ai-upscale-v88.jpg` |
| AI-分段（整页双栏 + 三个模式标签 + 处理日志） | `after-2026-11-11/ai-segment-v88.jpg` |
| 小说转剧（并发生成保留） | `after-2026-11-11/novel-to-drama-v88.jpg` |
| 分镜大师（无并发生成） | `after-2026-11-11/storyboard-master-v88.jpg` |
| 爆款开头（无并发生成） | `after-2026-11-11/viral-opening-v88.jpg` |
| 角色修改（无并发生成 + 「修改结果」） | `after-2026-11-11/character-edit-v88.jpg` |

截图全部取自 `http://127.0.0.1:4001`，由用户已登录的 Chrome 通过浏览器桥接拍摄，未用源码推断代替。AI-Subtitle / AI-Filler / AI-Upscale / 小说转剧 / 分镜大师 / 爆款开头 / 角色修改 拍自 12:58 构建；最终 13:11 构建只再改了 AI-分段头部按钮对齐，因此 AI-分段 已重拍，其余页面代码未变。

## 五、验证命令

```powershell
cd frontend; npm run build
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
schtasks /End /TN "V88-Local-4001"
schtasks /Run  /TN "V88-Local-4001"
```

测试：`node --test src/user/pages/pluginCatalog.layout.test.js src/user/pages/ai-segment-workbench.layout.test.js src/user/components/ai-segment-workbench.source.test.js` → 14/14 通过；后端 `go test ./internal/httpapi -run 'AICutCaption|AICutFiller' -count=1` 通过。

## 六、API 设置页异步回包覆盖本地输入（useEffect 重置）修复

本轮只改源码与单测，**未 commit、未部署、未调用生成模型**。4001 上只验证了上一轮已完成的「自定义预设」内容指纹修复；本轮新增的渠道 / Key / 封面守卫因需要重建二进制并重启 4001 才能生效，**仅做静态验证与单测**，未重启。

### 6.1 4001 真实页面验证（仅预设弹窗）

环境：`http://127.0.0.1:4001/api-config`，当时服务端入口为 `assets/user-DlBiVJ74.js`。

| 操作 | 实际观察 | 结论 |
| --- | --- | --- |
| 打开「自定义预设」 | 初始「已选预置模型」7 条 | 基线 |
| 取消勾选「深度探索 DeepSeek」 | 数字 7 → 6，复选框变为未勾选 | 生效 |
| 原地等待 5 秒（自动保存状态文案变化） | 仍为 6 条，该项未被自动勾回 | **PASS** |
| 对同一项快速连点 6 次 | 数字按 6/7/6/7/6/7 交替，复选框同步，无自行回弹、无报错 | **PASS** |
| 快速开关弹窗 3 次 | 浏览器桥接点击「取消 / Esc / 右上角 X」均未能关闭弹窗（modal 仍 `visibility: visible`），循环未完成 | **BLOCKED，未记为已验证** |

结论：上一轮「按内容指纹同步」的修复在真实页面有效，取消勾选后 5 秒观察窗内没有被重置。快速开关弹窗受浏览器自动化限制未验证，静态看 `onCancel` 只执行父级关闭，不再写回勾选。

### 6.2 同类问题扫描与修法

统一写法：新增 [hydrationGuard.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/hydrationGuard.js)，
「用户动过 → `canHydrate()` 为 false → 异步回包不落本地状态」；
请求飞行中可能再次本地修改的，用 `createRevisionTracker()` 只接受最新版本回包；
回包清理输入框草稿前用 `sameText` 比对，不清用户新输入。

| # | 文件与行号 | 现象 / 触发条件 | 改法 | 验证 |
| --- | --- | --- | --- | --- |
| 1 | [ApiPresetSelectionModal.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ApiPresetSelectionModal.jsx#L18-L46) | 父组件每次渲染传入新 `selectedIds`；旧 `useEffect` 把用户刚取消的勾选重置回旧值。触发：自动保存状态变化或父组件重渲染 | 保留内容指纹依赖，增加 `guard`：用户 toggle / 恢复默认后不再回灌；关闭弹窗时 `reset()` | **4001 实测 PASS**（见 6.1） |
| 2 | [ApiIntegrationSettings.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ApiIntegrationSettings.jsx#L358-L377) | 渠道整表保存回包无条件 `setChannels`。触发：保存请求飞行中又删除/编辑了渠道，旧回包把已删渠道复活或回滚编辑 | `createRevisionTracker()`；每次本地渠道变更 `bump()`，回包版本不是最新就丢弃 | 单测 + 静态 |
| 3 | [ApiIntegrationSettings.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ApiIntegrationSettings.jsx#L516-L533) | 目录模型 Key 保存成功后无条件清空草稿。触发：保存请求飞行中用户继续输入新 Key，回包把新输入清掉 | 用 `sameText(currentDraft, submittedDraft)` 判断，只清刚提交的那份 | 单测 + 静态 |
| 4 | [ApiCustomModelModal.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ApiCustomModelModal.jsx#L65-L88) | `useEffect` 依赖 `initial` 对象；父组件若在打开期间替换对象引用，会整表重填覆盖编辑中的表单 | 用户任意编辑 / 重置后 `guard.touch()`，打开期间不再回灌；关闭时 `reset()` | 仅静态验证（当前父级持有的 `initial` 引用稳定，未在 4001 复现） |
| 5 | [ApiConfigPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/ApiConfigPage.jsx#L90-L112) | 模型目录刷新会回写 H3 工作流 / Fanke 模型选择 | 刷新路径原有 `dirtyRef` 守卫；本轮把首次加载的两处回写也加上同一守卫 | 仅静态验证（刷新路径原本已正确） |
| 6 | [CoverGenerationWorkbench.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/components/CoverGenerationWorkbench.jsx#L130-L156) | 模型列表异步返回后无条件选中第一项，会覆盖用户已选模型 | 用户 `onChange` 时 `touch()`，异步回包只在未动过时设置默认模型 | 仅静态验证 |
| 7 | [PluginToolPage.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/PluginToolPage.jsx#L75-L86) | 配置加载只写 `models` / `prompts` 选项数组，不写 `modelId` / `promptId` | 无回灌点，不改 | 仅静态验证 |

### 6.3 测试

新增 [hydrationGuard.test.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/user/pages/hydrationGuard.test.js)，覆盖：

1. 用户在输入中，异步回包到达，值不被覆盖；
2. 取消勾选后，异步回包不会把勾选重置；
3. 请求飞行中用户又改过，旧回包按版本丢弃；
4. 异步保存只清理刚提交的草稿，不清用户新输入。

执行：

```powershell
cd frontend
node --test src/user/pages/hydrationGuard.test.js src/user/pages/apiPresetSelection.test.js src/user/pages/ApiConfigPage.source.test.js src/user/pages/api-config-preset-state.test.js src/user/pages/apiConfigCatalogMerge.test.js
```

结果：`tests 31 / pass 31 / fail 0`。`npm run build` 成功。

### 6.4 重建二进制、重启 4001 与实测（用户批准后）

流程：`npm run build` → 停 `V88-Local-4001` 并结束旧监听进程（确认 4001 无监听、`qiantie-local` 进程数 0）→ `scripts/build-go-local.ps1` → `schtasks /Run`。

本次启动时核对：

| 项 | 值 |
| --- | --- |
| 新 PID | 52708 |
| 进程启动时间 | 2026-11-11 14:04:40 |
| 二进制 mtime | 2026-11-11 14:04:34（启动时间晚于构建，确认不是旧进程镜像） |
| 服务端入口 | `/assets/user-DWpFSzhL.js`，与本地 `frontend/dist/index.html` 一致 |
| 入口 SHA256 | 服务端与本地一致 |
| `/` | 200 |

**并发会话冲突（新问题）**：16:00 复查时，4001 的监听者已不是本次启动的 `qiantie-local`，而是 `qiantie-local.aicut-lasso-scroll-candidate.exe`（PID 51376，启动 15:52:46）；计划任务 `V88-Local-4001` 的最后运行时间仍是 14:04:38。同期 `backend/qiantie-local.exe` 的 mtime 变为 15:29:25，说明另一个并行会话在此期间重建并替换了运行镜像。当前服务端入口为 `/assets/user-hMuuPTBB.js`，仍与当前本地 dist 一致，`/` 200、SHA256 一致（`6F90B19D…53D5860`），但该进程/构建不由本次任务控制。未擅自结束该进程以免打断并行会话，需用户决定是否收回 4001。

实测结果（临时数据，测试后已清理）：

| # | 场景 | 结果 | 说明 |
| --- | --- | --- | --- |
| 1 | 删除自定义模型后是否复活 | **PASS（部分场景）** | 创建 `V88-TEMP-A` → 删除 → 出现删除成功提示 → 等 5 秒，DOM 中不再出现 A，其他渠道未受影响。**并发编辑未构造**：浏览器桥接无法稳定打开第二条目的编辑弹窗 |
| 2 | 保存飞行中再改一次，最终保留最后一次 | **未验证** | 编辑弹窗在浏览器桥接里多次无法打开/提交；本地回包过快，也无法构造真实飞行窗口。守卫逻辑由单测覆盖 |
| 3 | Key 框保存触发后立刻输入字符不被清空 | **未验证** | 页面没有可安全改动的「目录 Key」测试数据（平台目录 Key 是真实凭证）；自定义模型 Key 走表单保存后弹窗即关闭，不存在飞行中输入窗口 |
| 4 | 快速开关「自定义预设」弹窗 3 次 | **未验证** | 弹窗能打开；antd 将「取消」渲染为「取 消」，按文本匹配失败；桥接层点右上角关闭与 Esc 也未能关闭。未记为通过 |

测试数据：`V88-TEMP-A`、`V88-TEMP-B` 均已删除，页面恢复测试前条目。


