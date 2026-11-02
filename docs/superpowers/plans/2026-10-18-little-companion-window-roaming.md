# 小家伙窗口活动与边缘探头实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让“小家伙”在窗口底部安全区域内真实走动、自动转身并在左右窗口边缘探头，同时正确响应窗口缩放和暂停条件。

**Architecture:** 新增无 DOM 依赖的纯位置模块，集中计算活动范围、走路目标、探头位置和插值坐标。`LittleCompanionSprite` 只报告当前姿态并接受朝向；`CmPenguinCompanion` 继续拥有外壳坐标，通过 `requestAnimationFrame` 推进实际位置，并在对话、任务、拖拽、页面隐藏或减少动态效果时停止推进。

**Tech Stack:** React 18、浏览器 `requestAnimationFrame`、Node 内置测试运行器、Vite 6、Go embed。

## Global Constraints

- 直接在当前 `v88` 工作，不新建分支。
- 保留当前工作区全部已有改动，不还原、不覆盖、不顺带格式化。
- 自主活动只适用于 `little-companion`；CM 和貔貅保持原行为。
- 正常活动区域限制在窗口底部 220 像素，水平安全边距为 8 像素。
- 左右探头至少保留 46 像素可见和可点击区域。
- 只有手动拖拽结束才保存位置；自主活动不得写入位置存储。
- 不增加依赖，不修改后端接口、数据库或提示词。
- 只在本地构建和验证；不提交、不推送、不发布公网。

---

## 文件结构

- Create `frontend/src/shared/pet/littleCompanionRoaming.js`：活动范围、走路目标、探头位置、位置插值和暂停判断。
- Create `frontend/src/shared/pet/littleCompanionRoaming.test.js`：固定输入的纯位置回归测试。
- Modify `frontend/src/shared/pet/LittleCompanionSprite.jsx`：向父组件报告姿态并接受左右朝向。
- Modify `frontend/src/shared/pet/CmPenguinCompanion.jsx`：管理自主活动状态、动画帧、窗口变化和暂停条件。
- Modify `frontend/src/shared/pet/cm-penguin-companion.css`：小家伙朝向翻转和运动中的稳定显示。
- Verify `scripts/build-go-local.ps1`、本地 `4001` 进程和已登录设置页。

---

### Task 1: 建立可测试的窗口活动几何规则

**Files:**
- Create: `frontend/src/shared/pet/littleCompanionRoaming.test.js`
- Create: `frontend/src/shared/pet/littleCompanionRoaming.js`

**Interfaces:**
- Produces: `littleCompanionActivityBounds(viewport)`。
- Produces: `clampLittleCompanionToViewport(position, viewport)`。
- Produces: `littleCompanionWalkTarget(position, viewport, ratio)`。
- Produces: `littleCompanionPeekTarget(position, viewport)`。
- Produces: `interpolateLittleCompanionPosition(from, to, progress)`。
- Produces: `canLittleCompanionRoam(context)`。

- [ ] **Step 1: 写几何规则失败测试**

```js
test('uses the bottom 220 pixels as the desktop activity area', () => {
  assert.deepEqual(littleCompanionActivityBounds({ width: 1440, height: 900 }), {
    minLeft: 8, maxLeft: 1312, minTop: 680, maxTop: 762
  });
});

test('walks toward the opposite half and returns an upper drop to the bottom band', () => {
  assert.deepEqual(littleCompanionWalkTarget({ left: 100, top: 120 }, { width: 1000, height: 800 }, 0.5), {
    left: 656, top: 662, direction: 1
  });
});

test('keeps exactly 46 pixels visible while peeking at the nearest edge', () => {
  assert.deepEqual(littleCompanionPeekTarget({ left: 80, top: 650 }, { width: 1000, height: 800 }), {
    side: 'left', left: -74, top: 650
  });
  assert.deepEqual(littleCompanionPeekTarget({ left: 850, top: 650 }, { width: 1000, height: 800 }), {
    side: 'right', left: 954, top: 650
  });
});

test('clamps resize changes without pulling a manual upper position into the bottom band', () => {
  assert.deepEqual(clampLittleCompanionToViewport({ left: 900, top: 120 }, { width: 700, height: 600 }), {
    left: 572, top: 120
  });
});
```

补充表驱动测试，覆盖窄于 136 像素的窗口、无效数字、`progress` 小于 0 和大于 1，以及全部暂停条件。

- [ ] **Step 2: 运行测试确认模块缺失而失败**

Run: `node --test frontend/src/shared/pet/littleCompanionRoaming.test.js`

Expected: FAIL with module not found。

- [ ] **Step 3: 写最小纯函数实现**

固定常量：宠物 `120 x 130`、安全边距 `8`、顶部最低值 `58`、底部活动区高度 `220`、探头可见宽度 `46`。走路目标按当前位置选择水平对侧的一半，`ratio` 限制到 `0..1` 后决定对侧半区内的目标。垂直目标必须进入底部活动区。

`interpolateLittleCompanionPosition` 使用线性插值，并把进度限制到 `0..1`。`canLittleCompanionRoam` 只有在宠物 ID 正确、状态为 `idle`、对话关闭、没有拖拽、页面可见、未减少动态、图片未失败时返回 `true`。

- [ ] **Step 4: 运行测试并检查差异**

Run: `node --test frontend/src/shared/pet/littleCompanionRoaming.test.js`

Expected: 全部 PASS。

Run: `git diff --check -- frontend/src/shared/pet/littleCompanionRoaming.js frontend/src/shared/pet/littleCompanionRoaming.test.js`

Expected: exit 0；不提交。

---

### Task 2: 让图集显示器报告姿态并支持朝向

**Files:**
- Modify: `frontend/src/shared/pet/littleCompanionMotion.test.js`
- Modify: `frontend/src/shared/pet/littleCompanionMotion.js`
- Modify: `frontend/src/shared/pet/LittleCompanionSprite.jsx`
- Modify: `frontend/src/shared/pet/cm-penguin-companion.css`

**Interfaces:**
- Produces: `littleCompanionPoseStep(state, stepIndex, preview)`，返回当前 `{ pose, durationMs }`。
- Extends: `<LittleCompanionSprite direction={1|-1} onPoseChange={(step) => {}} />`。

- [ ] **Step 1: 先写姿态选择失败测试**

```js
test('exposes the current pose step for the roaming controller', () => {
  assert.deepEqual(littleCompanionPoseStep('idle', 1, false), { pose: 'walk', durationMs: 720 });
  assert.deepEqual(littleCompanionPoseStep('idle', 4, true), { pose: 'idle', durationMs: 3600 });
  assert.deepEqual(littleCompanionPoseStep('success', 1, false), { pose: 'dance', durationMs: 610 });
});
```

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js`

Expected: FAIL，因为导出不存在。

- [ ] **Step 2: 实现姿态选择并接入组件回调**

`littleCompanionPoseStep` 必须复用已验证的时间线并安全循环索引。`LittleCompanionSprite` 使用它获得 `step`，每次 `pose` 或 `durationMs` 改变时调用 `onPoseChange(step)`。`preview` 永远报告 `idle`。

外层 `.little-companion-sprite` 根据 `direction` 添加 `is-facing-left`，CSS 只翻转角色画面，不改变 120 x 120 裁切框和外壳坐标。

- [ ] **Step 3: 验证动作逻辑和前端构建**

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/littleCompanionRoaming.test.js`

Run: `npm --prefix frontend run build`

Expected: 两个命令 exit 0；不提交。

---

### Task 3: 在共同外壳中驱动真实位置

**Files:**
- Modify: `frontend/src/shared/pet/CmPenguinCompanion.jsx`
- Modify: `frontend/src/shared/pet/LittleCompanionSprite.jsx`

**Interfaces:**
- Consumes: Task 1 的五个纯函数和 Task 2 的姿态回调。
- Produces: 小家伙在 `walk` 时产生位移、在 `peek` 时进入最近边缘、在暂停条件出现时停在当前坐标。

- [ ] **Step 1: 加入活动状态而不改变现有拖拽保存规则**

在共同外壳中维护：`littlePose`、`littleDirection`、`isDragging`、`pageVisible`、`reducedMotion`、`spriteFailed` 和当前坐标 ref。现有 `setPosition` 同步更新 ref；`savePosition` 仍只在 `endDrag` 调用。

组件挂载时监听 `visibilitychange`、`matchMedia('(prefers-reduced-motion: reduce)')` 和 `resize`，卸载时移除。`resize` 只调用 `clampLittleCompanionToViewport`，因此上方手动位置不会突然跳到底部。

- [ ] **Step 2: 用动画帧推进走路和探头位移**

姿态变为 `walk` 且 `canLittleCompanionRoam` 为真时，调用 `littleCompanionWalkTarget(positionRef.current, viewport, Math.random())`，在 720 毫秒内用 `requestAnimationFrame` 和线性插值更新位置及朝向。

姿态变为 `peek` 时，调用 `littleCompanionPeekTarget`，在不超过 320 毫秒内移动到边缘。暂停条件改变时取消当前动画帧；取消前最后一次已写入的坐标保持不变。姿态离开 `walk/peek` 后不再更新坐标。

- [ ] **Step 3: 保持点击、拖拽和任务优先级**

`beginDrag` 先取消自主动画并设置 `isDragging=true`；`endDrag` 使用现有限制规则和保存逻辑后设置 `isDragging=false`。对话打开或状态离开 `idle` 时取消自主动画。切换到 CM/貔貅时清理姿态和动画，不启动位置计算。

将 `direction`、`onPoseChange` 和 `onLoadError` 传给 `LittleCompanionSprite`。图片失败时停止自主移动并保留现有占位。

- [ ] **Step 4: 运行回归和构建**

Run: `node --test frontend/src/shared/pet/littleCompanionRoaming.test.js frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js frontend/src/shared/pet/companion-position.test.js`

Run: `npm --prefix frontend test`

Run: `npm --prefix frontend run build`

Expected: 全部 exit 0。

---

### Task 4: 重建 4001 并完成真实页面验收

**Files:**
- Verify: `scripts/build-go-local.ps1`
- Verify: `scripts/start-local-go-review.ps1`
- Verify: 本地已登录 `http://localhost:4001/settings`

- [ ] **Step 1: 重建并重启本地 Go 进程**

运行 `scripts/build-go-local.ps1`，停止且仅停止监听 `4001` 的本地 `backend/qiantie-local.exe`，再用 `scripts/start-local-go-review.ps1 -UseYcmMysql` 启动。不得修改数据库迁移记录。

- [ ] **Step 2: 验证运行入口**

运行 `scripts/check-local-go.ps1`，并确认 `/api/runtime-build-info` 的 `frontend_entry` 与 `.local-review/runtime-manifest.json` 一致；返回的入口脚本包含 `little-companion`。

- [ ] **Step 3: 在真实页面检查桌面和窄窗口**

登录后选择“小家伙”，观察至少一轮 `idle -> walk -> candy -> dance -> sleep -> peek`：走路必须产生水平位移和正确转身；探头必须在左右边缘至少保留 46 像素；打开对话、拖拽、触发任务时移动停止；调整窗口宽高后宠物保持可见。再切换 CM、貔貅确认它们不自主移动。

- [ ] **Step 4: 最终验证**

Run: `node --test frontend/src/shared/pet/littleCompanionRoaming.test.js frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js frontend/src/shared/pet/companion-position.test.js`

Run: `npm --prefix frontend test`

Run from `backend`: `go test ./internal/web ./internal/httpapi -count=1`

Run: `git diff --check -- docs/superpowers/specs/2026-10-18-little-companion-window-roaming-design.md docs/superpowers/plans/2026-10-18-little-companion-window-roaming.md frontend/src/shared/pet`

Expected: 全部 exit 0。记录浏览器实测范围和剩余缺口；不提交、不推送、不发布。
