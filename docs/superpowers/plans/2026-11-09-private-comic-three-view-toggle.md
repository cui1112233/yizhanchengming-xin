# 私人定制人物三视图开关实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在私人定制漫画的人物资产编辑框增加默认开启的三视图开关，并把选择传给现有资产图生成接口。

**Architecture:** `PrivateComicWorkspace` 负责当前人物资产的开关状态和可见控件，`ScriptPage` 负责把布尔值写入现有 `generateBookAssetImages` 请求。后端接口和统一资产编辑外壳保持不变。

**Tech Stack:** React、Ant Design、Node.js 内置测试、Vite、Go embed 本地运行时

## Global Constraints

- 开关文案固定为“人物三视图”，默认开启。
- 开关只在 `character` 资产编辑框出现。
- 单项人物资产生成使用用户当前选择。
- 场景、道具单项生成传入 `false`。
- 批量补齐资产图继续默认使用三视图参数，后端只对人物应用人物模板。
- 不新增后端接口或数据库字段。
- 只更新本地 4001，不部署线上。
- 不新建分支，不推送远端。

---

### Task 1: 接通私人定制三视图选择链路

**Files:**
- Create: `frontend/src/user/pages/private-studio/privateComicThreeView.js`
- Create: `frontend/src/user/pages/private-studio/privateComicThreeView.test.js`
- Modify: `frontend/src/user/pages/PrivateComicWorkspace.jsx`
- Modify: `frontend/src/user/pages/ScriptPage.jsx`

**Interfaces:**
- Consumes: `onGenerateAssetImage(asset, useThreeView)` 和现有 `generateBookAssetImages(..., { useThreeView })`。
- Produces: `privateComicAssetUseThreeView(assetKind, enabled): boolean` 和 `runPrivateComicAssetImages(assetId = '', useThreeView = true)`；单项生成按资产类型与用户选择决定，批量调用省略第二个参数时默认开启。

- [ ] **Step 1: 写失败行为测试**

新增 `privateComicThreeView.test.js`，用字面值验证人物、场景和道具的生成参数：

```js
import assert from 'node:assert/strict';
import test from 'node:test';
import { privateComicAssetUseThreeView } from './privateComicThreeView.js';

test('私人定制只有人物资产按开关使用三视图', () => {
  assert.equal(privateComicAssetUseThreeView('character', true), true);
  assert.equal(privateComicAssetUseThreeView('character', false), false);
  assert.equal(privateComicAssetUseThreeView('scene', true), false);
  assert.equal(privateComicAssetUseThreeView('prop', true), false);
});
```

- [ ] **Step 2: 运行测试并确认因行为模块缺失失败**

Run: `node --test src/user/pages/private-studio/privateComicThreeView.test.js`

Expected: FAIL，找不到 `privateComicThreeView.js`。

- [ ] **Step 3: 实现资产类型判断并在私人定制资产框加入开关**

新增 `privateComicThreeView.js`：

```js
export function privateComicAssetUseThreeView(assetKind, enabled) {
  return assetKind === 'character' && enabled === true;
}
```

在 `PrivateComicWorkspace.jsx`：

```jsx
const [assetImageThreeView, setAssetImageThreeView] = useState(true);
```

人物资产右侧操作区加入与批量工厂一致的 `Tooltip + Switch + 人物三视图`，并把生成调用改为：

```jsx
onGenerateAssetImage?.(editingAsset, privateComicAssetUseThreeView(editingAsset?.kind, assetImageThreeView))
```

- [ ] **Step 4: 在 ScriptPage 接通布尔参数**

把生成函数签名改为：

```js
async function runPrivateComicAssetImages(assetId = '', useThreeView = true)
```

请求参数改为：

```js
useThreeView: Boolean(useThreeView)
```

单项回调改为：

```jsx
onGenerateAssetImage={(asset, useThreeView) => void runPrivateComicAssetImages(asset?.id || '', useThreeView)}
```

批量补齐调用保持 `runPrivateComicAssetImages()`，由默认参数继续开启三视图。

- [ ] **Step 5: 运行相关测试和前端构建**

```powershell
node --test src/user/pages/private-studio/privateComicThreeView.test.js src/user/pages/private-studio/PrivateStudioAssetEditor.source.test.js src/user/pages/PrivateComicWorkspace.source.test.js
npm run build
```

Expected: 所有测试通过，Vite 构建退出码为 0。

- [ ] **Step 6: 提交实现**

```powershell
git add -- frontend/src/user/pages/private-studio/privateComicThreeView.js frontend/src/user/pages/private-studio/privateComicThreeView.test.js frontend/src/user/pages/PrivateComicWorkspace.jsx frontend/src/user/pages/ScriptPage.jsx
git commit -m "feat: add private comic three-view toggle"
```

### Task 2: 重建并验收本地 4001

**Files:**
- Build output: `backend/qiantie-local.exe`
- Runtime assets: `backend/web/dist`

**Interfaces:**
- Consumes: Task 1 的前端实现和已存在的 Go `useThreeView` 处理。
- Produces: 本地 `/script` 人物资产框可选择三视图，场景与道具不显示该开关。

- [ ] **Step 1: 核对并停止 4001**

确认监听进程路径为 `F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`；路径不符则停止操作，路径一致时只停止该 PID。

- [ ] **Step 2: 重建并启动本地二进制**

```powershell
.\scripts\build-go-local.ps1
.\scripts\start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached
```

- [ ] **Step 3: 验证运行身份与接口**

确认 `/health`、`/script` 返回 200，`/api/runtime-build-info` 的 Git SHA 等于当前 HEAD，4001 进程路径正确。

- [ ] **Step 4: 用内置浏览器验收**

刷新 `/script`，进入漫画生成，点击人物资产，确认“人物三视图”默认开启；关闭弹窗后点击场景或道具资产，确认不显示开关。

- [ ] **Step 5: 检查工作树**

Run: `git status --short`

Expected: 无输出；不执行线上发布命令。
