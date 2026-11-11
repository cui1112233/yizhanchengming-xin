# 「自定义预设」与「添加自定义模型」入口拆分（2026-11-11）

范围：`API 配置` 页三个模型分组（文本 / 图片 / 视频）的渠道区。
结论：两个入口已拆开，职责单一；勾选结果写入服务端 `channelSelection`，表单只增改模型。
状态：纯逻辑与渲染已验证；**点击级 UI 交互本轮未验证**（本机浏览器桥接 CDP 超时，见第六节）。

## 一、两个入口的职责

| 入口 | 形态 | 只做什么 | 不做什么 |
| --- | --- | --- | --- |
| 自定义预设 | 选择弹窗（760px，左「全部预置模型」带复选框 / 右「已选预置模型」可单项移除） | 决定该分组**展示哪些预置模型卡片** | 不填 Key、不改地址、不影响模型下拉 |
| 添加自定义模型 / 编辑自定义模型 | 表单弹窗 | 新增或修改一条用户自建模型 | 不管理预置模型展示清单 |

弹窗底部按钮：预设弹窗为「初始化/恢复默认」「取消」「确定」；表单弹窗为「重置」「添加模型 / 保存」。

## 二、改动文件与行号

| 文件 | 行号 | 改动 |
| --- | --- | --- |
| `frontend/src/user/pages/apiPresetSelection.js` | 新增，全文件 | 纯逻辑：`explicitPresetIds` / `selectedPresetIds` / `filterPresetsBySelection` / `buildPresetSelectionPayload` |
| `frontend/src/user/pages/ApiPresetSelectionModal.jsx` | 新增，全文件（112 行） | 勾选弹窗：`toggle`(28-33)、`restoreDefault`(35-37)、footer(48-56)、左栏(62-83)、右栏(84-109) |
| 同上 | 16-23 | 修复：父组件每次渲染传新数组导致 `useEffect` 在保存状态变化时重置用户勾选；改为按内容指纹（`selectedKey`）同步 |
| `frontend/src/user/pages/ApiIntegrationSettings.jsx` | 5 | 引入 `ApiPresetSelectionModal` |
| 同上 | 7 | 引入预设选择纯逻辑 |
| 同上 | 194 | 新增 `presetModal` / `presetSaving` 状态 |
| 同上 | 409-431 | `savePresetSelection`：写 `channelSelection` → POST → 回填 → 关闭弹窗；失败 `message.error('预置模型选择保存失败')` |
| 同上 | 692 | `buildGroupMerge(group, def, presets = def.presets)`，卡片列表改用传入的预置集合 |
| 同上 | 722-725 | `renderChannelGroup` 先按勾选过滤预置再合并 |
| 同上 | 735-741 | 组内按钮区：`自定义预设`（`.api-preset-open`）+ `添加自定义模型`（`.api-custom-open`） |
| 同上 | 791-794 | 分组头部计数用同一份过滤结果，避免「渠道 9 / 实际 8 张卡」 |
| 同上 | 894-906 | 挂载 `ApiPresetSelectionModal` |
| `frontend/src/shared/styles/api-integrations.css` | 218-224 | `.api-enabled-actions` |
| 同上 | 225-330 | `.api-preset-picker*`（两栏、列高 280-420、head/list/meta/remove/empty/footer） |
| `frontend/src/user/pages/apiPresetSelection.test.js` | 新增，68 行 | 五场景单测 |
| `frontend/src/user/previewApiSettings.jsx` | 33、234-236 | 预览 mock 改为 `channelSelection = {}`，新增 `?dialog=preset` |
| `frontend/package.json` | 7 | `test` 脚本纳入 `apiPresetSelection.test.js` |

## 三、数据落点（勾选与保存）

1. 前端：`buildPresetSelectionPayload(channelSelection, groupId, ids)` 生成 `{ [groupId]: { enabled: [...ids] } }`，仅覆盖本次分组。
2. 请求：`saveApiIntegrations({ channelSelection })` → `POST /api/settings/api-integrations`（`frontend/src/shared/api/apiIntegrations.js:7-12`，只发 `channelSelection`）。
3. 后端合并：`backend/internal/httpapi/api_integrations.go:508-524`；显式 `{enabled: []}` 会被保留为空数组（`isExplicitChannelSelection` 判定，`api_integrations.go:99-116`），未提交的分组保持不变。
4. 持久化键：配置项 `apiIntegrationsSettingKey` 下的 `channelSelection`（`api_integrations.go:525`）。
5. 读取生效：`GET /api/settings/api-integrations` 回显（`api_integrations.go:245`）；模型目录按同一份 `channelSelection` 归属渠道（`backend/internal/httpapi/model_catalog.go:607`、`677`）。
6. 自定义模型表单：`submitCustomModel`（`ApiIntegrationSettings.jsx:433` 起）→ `customChannels` 整表提交，与 `channelSelection` 互不影响。

## 四、截图

目录：`docs/verification/2026-11-11-api-preset-split/`

| 文件 | 内容 |
| --- | --- |
| `v88-api-group-entry-buttons-dark.png` | 分组内两个独立入口按钮 + 卡片区 |
| `v88-api-preset-picker-dark.png` | 「自定义预设 · 文本模型」弹窗（7 条预置、左右两栏、初始化/取消/确定） |
| `v88-api-preset-picker-light.png` | 同上，浅色主题 |
| `v88-api-custom-model-form-dark.png` | 「添加自定义模型 · 文本」纯表单（API 格式 / 请求地址 + 完整 URL / 模型 ID / 展示名称 / API Key / 高级配置） |
| `v88-api-custom-model-form-light.png` | 同上，浅色主题 |

截图来源：`frontend/src/user/previewApiSettings.jsx`（Vite dev server，mock 假数据）。**不是 4001 真实实例**。

## 五、五场景实测

| 场景 | 已执行的验证 | 结论 | 状态 |
| --- | --- | --- | --- |
| 正常 | 单测「未记录时展示全部，勾选后只展示勾选项」；预览弹窗渲染 7 条预置、已选 7 条 | 通过 | 已验证（逻辑 + 渲染） |
| 空值 | 单测「空对象、坏结构回落全部；显式空数组表示全部取消」 | 通过 | 已验证（逻辑） |
| 重复 | 单测「重复勾选不产生重复记录，连续保存结果一致」 | 通过 | 已验证（逻辑） |
| 错误 | 单测「未知 id 被丢弃、数字 id 归一化、不污染其他分组」；UI 失败分支为 `message.error('预置模型选择保存失败')` | 通过 | 逻辑已验证；**UI 失败提示未触发验证** |
| 取消 | 单测「纯函数不修改入参」；`onCancel` 只 `setPresetModal(null)`，不发请求 | 通过 | 逻辑已验证；**点击取消未实操验证** |

单测执行：`node --test src/user/pages/apiPresetSelection.test.js` → 5 tests / 5 pass / 0 fail。

## 六、未验证项与阻塞

1. 本机浏览器桥接不可用：`browser_navigate` / `browser_evaluate` 均返回 `browser bridge response timed out`（CDP 超时，连续 3 次）。因此以下未能执行：
   - 预览里点击复选框 → 「已选预置模型」7→6 → 确定 → 卡片区少一条 → 重开保持；
   - 4001 真实实例上的同一链路（含服务端 `channelSelection` 实际写入复核）。
2. 未跑 `npm run build`，未改 `frontend/dist`，未重启 4001。
3. 本轮修掉一个**交互级 bug**（`ApiPresetSelectionModal.jsx:16-23`）：父组件每次渲染都传新数组，原来的 `useEffect([open, selectedIds])` 会在自动保存状态变化时把用户刚取消的勾选重置回去。修复后仅按内容指纹同步。该修复**只做了代码级确认，没有点击级复测**（受第 1 条阻塞）。
4. 4001 上当前运行的是拆分前构建，需要重新构建+重启才能看到这两个入口（重启前会先询问）。

## 七、洛水参考情况

- 「自定义预设」（预置模型展示清单、双栏勾选）：**洛水无对应界面**。洛水 bundle 中无「预置模型 / 渠道分组 / 自定义预设」设置界面文案（`预置模型` 唯一命中在 `index-DszfAkKD.js:87292` 的 MCP 工具说明里，不是设置页）。V88 为自设计。
- 「添加自定义模型」：洛水只有**在模型下拉里输入名称追加一条**的做法（`index-DszfAkKD.js:380193-380207`：输入为空提示「请输入自定义模型名称」，与内置列表重复提示「该模型已在内置列表中」，成功提示「已添加自定义模型：xxx」）；没有 API 格式 / 请求地址 / 完整 URL / 模型 ID / 展示名称 / 并发数表单。V88 的表单为自设计，字段按本轮需求定义。
