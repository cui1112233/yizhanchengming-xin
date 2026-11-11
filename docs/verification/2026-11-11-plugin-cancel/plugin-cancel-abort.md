# 插件面板请求取消（AbortController）· 2026-11-11

结论：四个提示词工具、AI改文、封面生成（design→出图两阶段 + 改图）、资产参考都已改为**真实 abort**；请求层不再把用户取消当网络失败上报，也不弹全局错误。逻辑层与请求层已用单测覆盖，浏览器点击级未验证（构建冻结 + 浏览器桥接 CDP 超时）。

## 一、机制

1. 每个面板持有一个 gate（同一时刻只允许一个生成）：`createGenerationGate()` 返回 `begin()/end()/abort()`，`begin()` 在已有运行时返回 `null`，用于挡重复提交。
2. 发起请求时用 `apiRequest(path, { ...options, signal: controller.signal })`，而不是先请求再决定是否使用结果。
3. 卸载 / 关闭面板 / 切路由：`useEffect` 的清理函数调用 `gate.abort()`；资产参考的「取消」按钮先 `abort()` 断请求，再调服务端取消排队任务。
4. `apiRequest` 识别 caller abort：抛 `AbortError`（`aborted: true`），**不写客户端错误上报、不弹全局错误提示**；自有超时仍按超时处理。
5. 取消后调用方不写结果、不弹错误；资产参考提示「已取消本次操作」。

## 二、改动文件与行号

| 文件 | 行号 | 改动 |
| --- | --- | --- |
| `frontend/src/user/pages/pluginGeneration.js` | 新增，全文件（91 行） | `isAbortError`(3-5)、`cancelledError`(7-12)、`createGenerationGate`(15-38)、`validateGenerationInput`(40-45)、`executeGeneration`(48-90，取消返回 `{ok:false,cancelled:true,text:''}`) |
| `frontend/src/user/pages/pluginGeneration.test.js` | 新增，全文件（102 行） | 五场景单测 |
| `frontend/src/shared/api/client.js` | 88-96 | 新增 `createCancelledError({path,method})` |
| 同上 | 151-155 | fetch 阶段：caller abort → 抛 `AbortError`，不再走 `reportClientError` / `notifyApiFailure` |
| 同上 | 206-210 | 读 body 阶段同样按取消处理，只有超时才报超时 |
| `frontend/src/user/pages/PluginToolPage.jsx` | 9 | 引入 gate / executeGeneration / validate |
| 同上 | 65-68 | gate ref + 卸载 `gateRef.current.abort()` |
| 同上 | 116-129 | `generatePart(prompt, request)`：取消不重试，其他失败重试 1 次 |
| 同上 | 131-168 | `generate()`：`gate.begin()` 挡重复；`requestWithSignal`(141)；取消直接 return(157)；`gate.end`(165) |
| `frontend/src/user/pages/PluginRewritePage.jsx` | 8、28-31 | 同上：gate + 卸载 abort |
| 同上 | 72-95 | `generate()` 改为 `executeGeneration` + 带 signal 请求；取消不写结果、不报错 |
| `frontend/src/user/components/CoverGenerationWorkbench.jsx` | 16-18 | 引入 `apiRequest`、`isAbortError` |
| 同上 | 130-136 | `abortRef` + 卸载 abort |
| 同上 | 241-262 | `beginRun/endRun/scopedRequest/pollGeneration(request)` |
| 同上 | 264-360 | 两阶段生成：design 与每张出图都带 signal；多张共用一个 controller，取消一次性中止未完成的，已完成的 `setResults` 保留 |
| 同上 | 374-420 | 改图（image_to_image）同样带 signal，取消不弹错误 |
| `frontend/src/user/components/AssetReferenceWorkbench.jsx` | 86-90 | 新增 `abortRef` |
| 同上 | 105-110 | 卸载：置取消标记 + abort |
| 同上 | 191-196 | 生成时创建 controller 与 `requestWithSignal` |
| 同上 | 206、215 | 所有 create/poll 请求带 signal |
| 同上 | 237-241 | finally 清理 controller |
| 同上 | 244-250 | 「取消」按钮先 abort 再走服务端取消 |
| `frontend/src/user/pages/pluginCatalog.layout.test.js` | 138-145 | 静态断言改为匹配抽离后的校验函数（原断言指向已删除的行内校验） |
| `frontend/package.json` | 7 | test 脚本纳入 `pluginGeneration.test.js` |
| `docs/acceptance/plugin-batch1-prompt-tools-2026-11-11.md` | 14/156/167/217/260 | 把「取消：不通过」改为本轮实测结论 |

## 三、五场景结论

| 场景 | 实测 | 结论 |
| --- | --- | --- |
| 正常 | 两段串行 → `结果一\n\n结果二`，进度 `1/2 → 2/2` | 通过（单测） |
| 空值 | 空白输入被 `validateGenerationInput` 拦下；空分段返回 `没有可生成的内容` 且不发请求 | 通过（单测） |
| 重复点击 | `gate.begin()` 第二次返回 `null`；结束后可再次 `begin()` | 通过（单测）；页面 `running` 分支保留 |
| 真实失败 | `runPart` 抛 `模型调用失败` → `{ok:false,cancelled:false}`，页面弹 error | 通过（单测） |
| 用户取消 | 取消后 `controller.signal.aborted === true`，`outcome.cancelled === true`，`results === []`、`text === ''` | 通过（单测） |

执行：`node --test src/user/pages/pluginGeneration.test.js` → 5/5；与 `pluginCatalog.layout.test.js`、`pluginToolLogic.test.js` 合并跑 → 21/21。

请求层：`frontend/src/shared/api/client.test.js` 既有断言仍通过——caller abort 抛 `AbortError`、不判超时、不触发失败上报。

## 四、覆盖范围

| 面板 | 请求 | 取消入口 |
| --- | --- | --- |
| 小说转剧 / 分镜大师 / 爆款开头 / 角色修改 | `/api/plugins/{tool}/run` | 关闭面板、返回、切路由（卸载）、重复提交拦截 |
| AI改文 | 同上 | 关闭面板、切路由 |
| 封面生成 | `designCoverPrompt` → `createAgentImageGeneration` → 轮询 `getAgentGeneration` | 离开工作台（卸载） |
| 封面改图 | `createAgentImageGeneration`(image_to_image) → 轮询 | 同上 |
| 资产参考 | 多条 `createAgentImageGeneration` + 轮询 | 「取消」按钮（abort + 服务端取消）、卸载 |
| 格式化文本 / 字幕合并 / 文档合并 | 纯本地，无网络请求 | 不涉及 |

## 五、只能静态验证 / 未验证

1. **浏览器点击级未验证**：`browser_navigate` / `browser_evaluate` 仍返回 `browser bridge response timed out`（CDP 超时），无法在 4001 或 dev server 上点「关闭面板」看 Network 是否 `canceled`。
2. 未跑 `npm run build`，未改 `frontend/dist`，未重启 4001；4001 仍是旧构建，本轮改动尚未部署。
3. 未消耗额度发起真实生成，因此「取消发生在真实模型请求中途」只在单测的假请求上验证。
4. 已知窗口期：封面/资产参考的轮询在 `wait(2500)` 期间不响应 abort，最多 2.5 秒后才发下一次请求并被 abort；不会写结果，但请求终止不是瞬时。
5. 「已取消」提示：提示词工具与 AI改文只在面板仍然挂载时提示（`mountedRef`），关闭面板后不再补弹；资产参考用「已取消本次操作」；封面生成只有离开工作台这一种取消路径，因此不弹。
6. 以上取消链路全部是代码 + 单测级证据；浏览器里「Network 显示 canceled / UI 回到可点击」尚未实测。
