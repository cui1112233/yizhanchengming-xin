# 小家伙自然生活节奏与边缘探头实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让“小家伙”大部分时间安静待机、偶尔散步、活动数分钟后才睡觉或思考，并让手动或自主靠近左右窗口边缘时触发方向正确且不越界的探头。

**Architecture:** 新增无 DOM 依赖的纯行为模块，集中选择待机、散步、互动和疲劳休息；扩展现有纯位置模块，判断左右边缘、计算完整可见的边缘位置和向内退出目标。`CmPenguinCompanion` 负责浏览器计时和坐标动画，`LittleCompanionSprite` 只显示父组件给出的空闲姿态，同时继续自行处理任务状态动画。

**Tech Stack:** React 18、浏览器 `setTimeout` / `requestAnimationFrame`、Node 内置测试运行器、Vite 6、Go embed。

## Global Constraints

- 直接在当前 `v88` 工作，不新建分支。
- 保留当前工作区全部已有改动，不还原、不覆盖、不顺带格式化。
- 自然行为与边缘探头只适用于 `little-companion`；CM 和貔貅保持原行为。
- 待机 20 到 45 秒；疲劳阈值 2 到 4 分钟。
- 轻量行为权重固定为：45% 继续待机、35% 普通散步、8% 吃糖、5% 跳舞、7% 边缘散步。
- 疲劳后固定为：60% 睡觉 20 到 35 秒、40% 思考 10 到 18 秒。
- 普通散步距离 160 到 420 像素，速度每秒 90 到 140 像素。
- 左右边缘触发阈值为 32 像素；宠物外壳始终完整保留在 8 像素安全边距内。
- 探头持续 3 到 6 秒，结束后向窗口内部移动 120 到 240 像素。
- 自主活动不得写入拖拽位置存储；只有手动拖拽结束保存正常可见位置。
- 不增加依赖，不修改后端接口、数据库、迁移或提示词。
- 只在本地构建和验证；不提交、不推送、不发布公网。

---

## 文件结构

- Create `frontend/src/shared/pet/littleCompanionBehavior.js`：自然时间范围、行为权重和疲劳休息选择。
- Create `frontend/src/shared/pet/littleCompanionBehavior.test.js`：使用固定随机输入验证全部边界。
- Modify `frontend/src/shared/pet/littleCompanionRoaming.js`：短距离散步、走路时长、边缘判断、完整可见的探头位置和向内退出目标。
- Modify `frontend/src/shared/pet/littleCompanionRoaming.test.js`：覆盖左右边缘和窄窗口。
- Modify `frontend/src/shared/pet/LittleCompanionSprite.jsx`：接受父组件提供的空闲姿态，不再固定轮播空闲动作。
- Modify `frontend/src/shared/pet/littleCompanionMotion.js`：空闲默认只保持 `idle`，任务状态时间线保持不变。
- Modify `frontend/src/shared/pet/littleCompanionMotion.test.js`：验证父组件空闲姿态与任务动画边界。
- Modify `frontend/src/shared/pet/CmPenguinCompanion.jsx`：自然行为调度、暂停计时、边缘探头、拖拽触发和动画衔接。
- Modify `frontend/src/shared/pet/cm-penguin-companion.css`：左右探头朝向类。

---

### Task 1: 建立可测试的自然行为选择器

**Files:**
- Create: `frontend/src/shared/pet/littleCompanionBehavior.js`
- Create: `frontend/src/shared/pet/littleCompanionBehavior.test.js`

**Interfaces:**
- Produces: `littleCompanionIdleDelay(randomValue): number`。
- Produces: `littleCompanionFatigueDelay(randomValue): number`。
- Produces: `littleCompanionChooseAction({ activeMs, fatigueMs, actionRoll, durationRoll }): { kind, pose, durationMs, resetFatigue }`。
- Produces: `littleCompanionPeekDuration(randomValue): number`。
- Produces: `littleCompanionInwardDistance(randomValue): number`。

- [ ] **Step 1: 写时间范围和行为权重失败测试**

```js
test('keeps idle and fatigue delays inside the approved natural ranges', () => {
  assert.equal(littleCompanionIdleDelay(0), 20_000);
  assert.equal(littleCompanionIdleDelay(1), 45_000);
  assert.equal(littleCompanionFatigueDelay(0), 120_000);
  assert.equal(littleCompanionFatigueDelay(1), 240_000);
});

test('uses the approved lightweight action weights before fatigue', () => {
  const input = { activeMs: 60_000, fatigueMs: 180_000, durationRoll: 0 };
  assert.equal(littleCompanionChooseAction({ ...input, actionRoll: 0.20 }).kind, 'idle');
  assert.equal(littleCompanionChooseAction({ ...input, actionRoll: 0.60 }).kind, 'walk');
  assert.equal(littleCompanionChooseAction({ ...input, actionRoll: 0.84 }).kind, 'candy');
  assert.equal(littleCompanionChooseAction({ ...input, actionRoll: 0.90 }).kind, 'dance');
  assert.equal(littleCompanionChooseAction({ ...input, actionRoll: 0.97 }).kind, 'edge-walk');
});

test('rests only after fatigue and resets the fatigue clock', () => {
  assert.deepEqual(littleCompanionChooseAction({ activeMs: 180_000, fatigueMs: 180_000, actionRoll: 0.2, durationRoll: 0 }), {
    kind: 'sleep', pose: 'sleep', durationMs: 20_000, resetFatigue: true
  });
  assert.deepEqual(littleCompanionChooseAction({ activeMs: 180_000, fatigueMs: 180_000, actionRoll: 0.8, durationRoll: 1 }), {
    kind: 'think', pose: 'think', durationMs: 18_000, resetFatigue: true
  });
});
```

- [ ] **Step 2: 运行测试并确认因模块缺失而失败**

Run: `node --test frontend/src/shared/pet/littleCompanionBehavior.test.js`

Expected: FAIL with module not found。

- [ ] **Step 3: 实现最小纯函数**

随机输入统一限制到 `0..1`。所有时长使用线性插值并四舍五入；达到疲劳阈值时先走休息分支，未达到时按 `0.45 / 0.80 / 0.88 / 0.93 / 1.00` 分段。

- [ ] **Step 4: 运行测试和差异检查**

Run: `node --test frontend/src/shared/pet/littleCompanionBehavior.test.js`

Run: `git diff --check -- frontend/src/shared/pet/littleCompanionBehavior.js frontend/src/shared/pet/littleCompanionBehavior.test.js`

Expected: 全部 PASS；不提交。

---

### Task 2: 改造走路与左右边缘几何规则

**Files:**
- Modify: `frontend/src/shared/pet/littleCompanionRoaming.js`
- Modify: `frontend/src/shared/pet/littleCompanionRoaming.test.js`

**Interfaces:**
- Changes: `littleCompanionWalkTarget(position, viewport, distanceRoll, directionRoll)` 返回短距离正常目标。
- Produces: `littleCompanionWalkDuration(from, to, speedRoll): number`。
- Produces: `littleCompanionEdgeSide(position, viewport): 'left' | 'right' | null`。
- Produces: `littleCompanionEdgeTarget(position, viewport, side): { left, top, side }`。
- Produces: `littleCompanionInwardTarget(position, viewport, side, distance): { left, top, direction }`。
- Changes: `littleCompanionPeekTarget` 不再返回窗口外坐标，改为兼容调用 `littleCompanionEdgeTarget`。

- [ ] **Step 1: 写短距离散步和边缘失败测试**

```js
test('walks a bounded local distance instead of crossing the whole viewport', () => {
  assert.deepEqual(littleCompanionWalkTarget({ left: 500, top: 650 }, { width: 1200, height: 800 }, 0, 1), {
    left: 660, top: 662, direction: 1
  });
  assert.equal(littleCompanionWalkDuration({ left: 500, top: 662 }, { left: 660, top: 662 }, 1), 1143);
});

test('detects only the correct nearby edge', () => {
  assert.equal(littleCompanionEdgeSide({ left: 20, top: 600 }, { width: 1000, height: 800 }), 'left');
  assert.equal(littleCompanionEdgeSide({ left: 850, top: 600 }, { width: 1000, height: 800 }), 'right');
  assert.equal(littleCompanionEdgeSide({ left: 440, top: 600 }, { width: 1000, height: 800 }), null);
});

test('keeps the full shell visible while peeking and exits inward', () => {
  assert.deepEqual(littleCompanionEdgeTarget({ left: 20, top: 650 }, { width: 1000, height: 800 }, 'left'), { left: 8, top: 650, side: 'left' });
  assert.deepEqual(littleCompanionEdgeTarget({ left: 850, top: 650 }, { width: 1000, height: 800 }, 'right'), { left: 872, top: 650, side: 'right' });
  assert.deepEqual(littleCompanionInwardTarget({ left: 8, top: 650 }, { width: 1000, height: 800 }, 'left', 160), { left: 168, top: 650, direction: 1 });
});
```

- [ ] **Step 2: 运行测试并确认旧实现返回跨屏或越界坐标**

Run: `node --test frontend/src/shared/pet/littleCompanionRoaming.test.js`

Expected: FAIL because the current walk target crosses the opposite half and peek returns `-74` / `width - 46`。

- [ ] **Step 3: 实现最小几何逻辑并保留暂停判断**

普通散步先根据 `directionRoll < 0.5 ? -1 : 1` 选择方向，再用 `160 + distanceRoll * 260` 计算距离，并限制到正常活动范围。边缘目标始终使用 `minLeft` 或 `maxLeft`。窗口宽度不足 `200` 像素时 `littleCompanionEdgeSide` 返回 `null`。

- [ ] **Step 4: 运行行为与几何测试**

Run: `node --test frontend/src/shared/pet/littleCompanionBehavior.test.js frontend/src/shared/pet/littleCompanionRoaming.test.js`

Expected: 全部 PASS；不提交。

---

### Task 3: 用自然调度替换固定空闲轮播

**Files:**
- Modify: `frontend/src/shared/pet/littleCompanionMotion.js`
- Modify: `frontend/src/shared/pet/littleCompanionMotion.test.js`
- Modify: `frontend/src/shared/pet/LittleCompanionSprite.jsx`
- Modify: `frontend/src/shared/pet/CmPenguinCompanion.jsx`
- Modify: `frontend/src/shared/pet/cm-penguin-companion.css`

**Interfaces:**
- Extends: `<LittleCompanionSprite idleStep={{ pose, durationMs }} peekSide={'left'|'right'|null} />`。
- Consumes: Task 1 行为函数和 Task 2 几何函数。
- Produces: 父组件控制的 `littleAction`，结构为 `{ kind, pose, durationMs, peekSide?, resetFatigue? }`。

- [ ] **Step 1: 先修改动作测试，要求空闲不再固定轮播**

```js
test('keeps idle stable until the parent behavior scheduler chooses an action', () => {
  assert.deepEqual(littleCompanionTimeline('idle'), [{ pose: 'idle', durationMs: 3600 }]);
});

test('task states still own their interrupting timelines', () => {
  assert.deepEqual(littleCompanionTimeline('working').map(item => item.pose), ['think']);
  assert.deepEqual(littleCompanionTimeline('success').map(item => item.pose), ['peace', 'dance']);
});
```

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js`

Expected: FAIL because idle still returns six fixed poses。

- [ ] **Step 2: 让显示组件接受父级空闲姿态**

当 `state === 'idle'` 且提供 `idleStep` 时直接显示该姿态；任务状态继续使用 `littleCompanionPoseStep`。移除空闲状态内部的自动 `nextLittleCompanionStep`，保留任务成功的 `peace -> dance`。

- [ ] **Step 3: 在共同外壳中接入自然调度器**

新增 `littleAction`、`littleActionTimerRef`、`littleActiveMsRef`、`littleFatigueMsRef` 和 `littlePeekSide`。符合 `canLittleCompanionRoam` 时按当前动作持续时间安排下一动作；暂停时取消计时并记录已经过的可活动时间，恢复后从 `idle` 重新等待，不能补播。

`walk` 使用短距离目标；`edge-walk` 选择较近边缘并走到完整可见的边缘目标。走路结束后调用 `littleCompanionEdgeSide`：接近边缘则进入 `peek`，否则进入新一轮 `idle`。`peek` 结束后进入 `inward-walk`，走到向内目标后再待机。

- [ ] **Step 4: 接入拖动触发和左右视觉方向**

`endDrag` 保存正常位置后检查边缘；接近边缘则把位置吸附到对应安全坐标并设置 `peek`。拖动开始立即取消行为计时和动画。

CSS 使用 `.is-peeking-left` 和 `.is-peeking-right`，保证身体贴边、头朝窗口内部。实际映射必须在浏览器左右各验证一次；如果素材默认方向与变量名称相反，只调整类映射，不修改几何含义。

- [ ] **Step 5: 运行宠物回归和前端构建**

Run: `node --test frontend/src/shared/pet/littleCompanionBehavior.test.js frontend/src/shared/pet/littleCompanionRoaming.test.js frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js frontend/src/shared/pet/companion-position.test.js`

Run: `npm --prefix frontend test`

Run: `npm --prefix frontend run build`

Expected: 全部 exit 0；不提交。

---

### Task 4: 重建 4001 并完成真实节奏与左右探头验收

**Files:**
- Verify: `scripts/build-go-local.ps1`
- Verify: `scripts/start-local-go-review.ps1`
- Verify: `http://localhost:4001/settings`

- [ ] **Step 1: 重建并只重启已确认的本地 Go 进程**

运行 `scripts/build-go-local.ps1`。解析监听 `4001` 的进程，只有其可执行文件严格等于当前仓库 `backend/qiantie-local.exe` 时才停止，再用 `scripts/start-local-go-review.ps1 -UseYcmMysql` 启动。

- [ ] **Step 2: 验证运行入口**

Run: `scripts/check-local-go.ps1`

确认 `/api/runtime-build-info` 的 `frontend_entry` 与 `.local-review/runtime-manifest.json` 一致，并确认入口脚本包含小家伙资源路径。

- [ ] **Step 3: 在真实设置页验证自然节奏**

选择“小家伙”但不保存用户原配置。连续观察至少 50 秒：前 20 秒内不得自行切换空闲动作；待机时间达到范围后最多发生一次轻量动作，不能快速轮播睡觉、思考、吃糖、跳舞和探头。

测试时允许仅在浏览器页面内临时注入固定随机输入或缩短调度时间，用于覆盖 2 到 4 分钟疲劳分支；这类临时验证不得写入源码、存储或构建产物。确认疲劳前无睡眠/空闲思考，疲劳后只进入睡眠或思考，结束后重新待机。

- [ ] **Step 4: 在真实页面验证左右探头**

分别拖到左、右 32 像素边缘区，确认松手后外壳坐标仍为 `8` 或 `viewport.width - 128`，整块点击区域在窗口内，素材头朝窗口内部。拖到中央不得探头。验证一次自主 `edge-walk` 抵达边缘后进入相同探头流程，并确认结束后先向内走。

- [ ] **Step 5: 最终验证**

Run: `node --test frontend/src/shared/pet/littleCompanionBehavior.test.js frontend/src/shared/pet/littleCompanionRoaming.test.js frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js frontend/src/shared/pet/companion-position.test.js`

Run: `npm --prefix frontend test`

Run from `backend`: `go test ./internal/web ./internal/httpapi -count=1`

Run: `git diff --check -- docs/superpowers/specs/2026-10-18-little-companion-natural-rhythm-and-edge-peek-design.md docs/superpowers/plans/2026-10-18-little-companion-natural-rhythm-and-edge-peek.md frontend/src/shared/pet`

Expected: 全部 exit 0。记录真实浏览器观察到的待机时间、左右边缘坐标、方向、暂停行为和剩余缺口；不提交、不推送、不发布。
