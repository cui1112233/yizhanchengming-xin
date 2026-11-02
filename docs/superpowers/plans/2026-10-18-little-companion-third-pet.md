# 小家伙第三只宠物实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把用户原创的“小家伙”作为 V88 第三只可选桌面宠物，保留八套动作并复用现有拖拽、AI 对话、任务气泡和设置保存能力。

**Architecture:** `CmPenguinCompanion` 继续作为共同交互外壳；`petCatalog` 增加显示类型和小家伙目录项；新的纯动作模块负责状态、动作和帧时间线；新的 React 显示组件只负责绘制多张 PNG 图集。CM 和貔貅继续走原有单图集路径，后端配置格式和默认宠物保持不变。

**Tech Stack:** React 18、Vite 6、Node 内置测试运行器、CSS、Go embed 静态资源。

## Global Constraints

- 工作分支必须保持为 `v88`，不得创建新分支。
- 只修改本地源码和正式静态资源，不发布公网、不推送远端。
- 当前工作区已有大量用户改动；不得还原、覆盖或顺带格式化无关文件。
- 不复制压缩包中的 Taro 页面代码，不引入 Playwright、Sharp 或新的运行依赖。
- 默认宠物保持 CM；未知宠物标识继续回退 CM。
- 小家伙保存标识固定为 `little-companion`，用户可见名称固定为“小家伙”。
- 静态资源正式来源是根目录 `pets/`；`frontend/dist` 和 `backend/web/dist` 仅为构建产物，不作为实现源码。
- 用户没有明确授权提交；每项完成后只检查差异，不执行 `git commit` 或 `git push`。

---

## 文件结构

- Create `pets/little-companion/pet.json`：第三只宠物的可追溯资源说明。
- Create `pets/little-companion/{idle,walk,sleep,think,candy,dance,peace,peek}.png`：运行时八套动作图集。
- Create `frontend/src/shared/pet/littleCompanionMotion.js`：动作元数据、状态时间线、帧位置计算。
- Create `frontend/src/shared/pet/littleCompanionMotion.test.js`：纯动作逻辑回归测试。
- Create `frontend/src/shared/pet/LittleCompanionSprite.jsx`：多动作图集显示器。
- Modify `frontend/src/shared/pet/petCatalog.js`：注册第三只宠物及其显示类型。
- Modify `frontend/src/shared/pet/petCatalog.test.js`：验证目录、回退和切换事件。
- Modify `frontend/src/shared/pet/CmPenguinCompanion.jsx`：按显示类型选择旧图集或小家伙显示器。
- Modify `frontend/src/shared/pet/cm-penguin-companion.css`：小家伙裁切、占位和减少动态效果。
- Modify `frontend/src/user/pages/SettingsPage.jsx`：设置页预览使用相同显示器。
- Modify `frontend/src/shared/styles/global.css`：设置页两种显示类型的稳定尺寸。
- Verify only `scripts/build-go-local.ps1`、`backend/web/dist/pets/little-companion/*`：确认 Go embed 构建会包含新资源。

---

### Task 1: 注册第三只宠物和正式资源

**Files:**
- Create: `pets/little-companion/pet.json`
- Create: `pets/little-companion/idle.png`
- Create: `pets/little-companion/walk.png`
- Create: `pets/little-companion/sleep.png`
- Create: `pets/little-companion/think.png`
- Create: `pets/little-companion/candy.png`
- Create: `pets/little-companion/dance.png`
- Create: `pets/little-companion/peace.png`
- Create: `pets/little-companion/peek.png`
- Modify: `frontend/src/shared/pet/petCatalog.test.js`
- Modify: `frontend/src/shared/pet/petCatalog.js`

**Interfaces:**
- Consumes: 压缩包 `assets/manifest.json` 中的八套图集元数据。
- Produces: `getPetDefinition('little-companion')`，其中 `renderer === 'multi-sheet'`、`assetBasePath === '/pets/little-companion'`、`previewPose === 'idle'`。

- [ ] **Step 1: 写目录失败测试**

在 `petCatalog.test.js` 新增行为测试，明确捕获“第三只宠物未进入设置选项”和“事件发错宠物”的回归：

```js
import { getPetDefinition, getPetOptions, petAnimationDelay, previewPetSelection } from './petCatalog.js';

test('offers the original Little Companion as the third selectable pet', () => {
  assert.deepEqual(getPetOptions(), [
    { label: 'CM', value: 'stacky' },
    { label: '貔貅', value: 'pixiu' },
    { label: '小家伙', value: 'little-companion' }
  ]);
  assert.deepEqual(
    (({ id, displayName, renderer, assetBasePath, previewPose }) => ({ id, displayName, renderer, assetBasePath, previewPose }))(
      getPetDefinition('little-companion')
    ),
    {
      id: 'little-companion',
      displayName: '小家伙',
      renderer: 'multi-sheet',
      assetBasePath: '/pets/little-companion',
      previewPose: 'idle'
    }
  );
});
```

- [ ] **Step 2: 运行测试并确认按预期失败**

Run: `node --test frontend/src/shared/pet/petCatalog.test.js`

Expected: FAIL，第三个选项不存在，或 `little-companion` 被回退为 `stacky`。

- [ ] **Step 3: 复制经过核对的运行时图片并写资源说明**

从解压后的源码包只复制八张动作 PNG 到 `pets/little-companion/`。不要复制 `home-art*.png`、`card-head.png`、Taro 源码和导出工具。写入：

```json
{
  "id": "little-companion",
  "displayName": "小家伙",
  "description": "小家伙，陪你创作的原创桌面宠物。",
  "renderer": "multi-sheet",
  "assetBasePath": "/pets/little-companion",
  "source": "User original design, source package 2026-10-18"
}
```

复制后逐个核对文件尺寸和 SHA-256，确保与压缩包内原件一致。

- [ ] **Step 4: 写最小目录实现**

在 `petCatalog.js` 增加冻结对象并放到数组第三项：

```js
const littleCompanionPet = Object.freeze({
  id: 'little-companion',
  displayName: '小家伙',
  description: '小家伙，陪你创作的原创桌面宠物。',
  renderer: 'multi-sheet',
  assetBasePath: '/pets/little-companion',
  previewPose: 'idle',
  behaviorProfile: 'cm-v1',
  speechProfile: 'cm-v1',
  renderMode: 'smooth'
});
```

不要改变 `DEFAULT_PET_ID` 和现有两只宠物的字段。

- [ ] **Step 5: 运行目录测试并检查本任务差异**

Run: `node --test frontend/src/shared/pet/petCatalog.test.js`

Expected: PASS，现有 CM/貔貅节奏测试仍通过。

Run: `git diff --check -- frontend/src/shared/pet/petCatalog.js frontend/src/shared/pet/petCatalog.test.js pets/little-companion/pet.json`

Expected: exit 0。只记录结果，不提交。

---

### Task 2: 用纯逻辑定义八套动作和状态优先级

**Files:**
- Create: `frontend/src/shared/pet/littleCompanionMotion.test.js`
- Create: `frontend/src/shared/pet/littleCompanionMotion.js`

**Interfaces:**
- Produces: `LITTLE_COMPANION_POSES`、`littleCompanionTimeline(state)`、`littleCompanionFrame(pose, frame)`。
- `littleCompanionTimeline(state)` 返回冻结的 `{ pose, durationMs }[]`。
- `littleCompanionFrame(pose, frame)` 返回 `{ pose, src, column, row, columns, rows, frameCount, frameDelayMs }`。

- [ ] **Step 1: 写动作逻辑失败测试**

```js
import { littleCompanionFrame, littleCompanionTimeline } from './littleCompanionMotion.js';

test('cycles every approved idle pose in order', () => {
  assert.deepEqual(littleCompanionTimeline('idle').map(item => item.pose), [
    'idle', 'walk', 'candy', 'dance', 'sleep', 'peek'
  ]);
});

test('task states interrupt the idle timeline', () => {
  assert.deepEqual(littleCompanionTimeline('working').map(item => item.pose), ['think']);
  assert.deepEqual(littleCompanionTimeline('success').map(item => item.pose), ['peace', 'dance']);
  assert.deepEqual(littleCompanionTimeline('error').map(item => item.pose), ['idle']);
});

test('calculates a wrapped frame from hand-checked sprite metadata', () => {
  assert.deepEqual(littleCompanionFrame('walk', 13), {
    pose: 'walk',
    src: '/pets/little-companion/walk.png',
    column: 1,
    row: 0,
    columns: 8,
    rows: 2,
    frameCount: 12,
    frameDelayMs: 60
  });
});

test('falls back to idle for an unknown pose', () => {
  assert.equal(littleCompanionFrame('missing', 0).pose, 'idle');
});
```

这些测试会在模块不存在时失败，且预期值由素材清单人工换算，不调用生产代码计算。

- [ ] **Step 2: 运行测试并确认失败原因正确**

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js`

Expected: FAIL with module not found。

- [ ] **Step 3: 写最小动作模块**

动作元数据使用源码包的准确值：`idle 24/3600/8x4`、`walk 12/720/8x2`、`sleep 48/4600/8x7`、`think 64/12000/8x9`、`candy 40/4800/8x6`、`dance 12/610/8x2`、`peace 24/2000/8x4`、`peek 24/4000/8x4`。

实现规则：

```js
const IDLE_TIMELINE = ['idle', 'walk', 'candy', 'dance', 'sleep', 'peek'];
const TASK_TIMELINES = {
  working: ['think'],
  success: ['peace', 'dance'],
  error: ['idle']
};
```

`frameDelayMs` 使用 `Math.round(durationMs / frames)`；帧号先对 `frameCount` 做非负取模，再按 8 列计算行列；无效动作回退 `idle`，无效状态回退空闲时间线。

- [ ] **Step 4: 运行测试并检查边界**

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js`

Expected: PASS，全部用例 0 失败。

补充一个负帧号用例，断言 `littleCompanionFrame('idle', -1)` 返回最后一帧，而不是负坐标；先观察失败，再补最小归一化代码并重新运行。

- [ ] **Step 5: 检查本任务差异**

Run: `git diff --check -- frontend/src/shared/pet/littleCompanionMotion.js frontend/src/shared/pet/littleCompanionMotion.test.js`

Expected: exit 0。只记录结果，不提交。

---

### Task 3: 接入共同宠物外壳和设置预览

**Files:**
- Create: `frontend/src/shared/pet/LittleCompanionSprite.jsx`
- Modify: `frontend/src/shared/pet/CmPenguinCompanion.jsx`
- Modify: `frontend/src/shared/pet/cm-penguin-companion.css`
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Modify: `frontend/src/shared/styles/global.css`

**Interfaces:**
- Consumes: `littleCompanionTimeline(state)` 和 `littleCompanionFrame(pose, frame)`。
- Produces: `<LittleCompanionSprite state="idle" preview={false} label="小家伙" />`。
- `preview=true` 固定使用 `idle` 姿态，但仍播放待机帧；不会自动切换到其他姿态。

- [ ] **Step 1: 先给时间线推进增加可测试接口**

在 Task 2 测试中先增加：

```js
import { nextLittleCompanionStep } from './littleCompanionMotion.js';

test('advances within a timeline and wraps without leaking the previous task state', () => {
  assert.equal(nextLittleCompanionStep('idle', 5), 0);
  assert.equal(nextLittleCompanionStep('success', 0), 1);
  assert.equal(nextLittleCompanionStep('working', 4), 0);
});
```

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js`

Expected: FAIL，因为导出尚不存在。然后实现非负整数归一化和按当前时间线长度取模，再确认 PASS。

- [ ] **Step 2: 实现独立显示组件**

组件只维护 `stepIndex` 和 `frameIndex`。状态改变时两者归零；一个计时器按当前姿态 `frameDelayMs` 推进帧，另一个计时器按时间线项 `durationMs` 推进姿态。每个 effect 必须清理自己的 timer。

渲染结构固定为：

```jsx
<span className={`little-companion-sprite${failed ? ' is-failed' : ''}`} aria-hidden="true">
  {failed ? <span className="little-companion-fallback">{label}</span> : (
    <img src={frame.src} alt="" draggable="false" onError={() => setFailed(true)} style={{
      '--little-columns': frame.columns,
      '--little-rows': frame.rows,
      '--little-column': frame.column,
      '--little-row': frame.row
    }} />
  )}
</span>
```

图片地址或姿态变化时清除 `failed`。`prefers-reduced-motion` 由 CSS 暂停画面，并在组件中通过 `matchMedia` 停止两个推进计时器；监听设置变化并在卸载时移除监听。

- [ ] **Step 3: 在共同外壳选择显示方式**

保留当前 `spriteRow`、`spriteFrame` 和旧 `<img>` 逻辑。仅将角色画面替换为条件分支：

```jsx
{pet.renderer === 'multi-sheet'
  ? <LittleCompanionSprite state={state} label={pet.displayName} />
  : <img src={pet.spritesheetPath} alt="" draggable="false" style={{ imageRendering: pet.renderMode === 'smooth' ? 'auto' : undefined }} />}
```

不要移动或修改对话、拖拽、点击、键盘和任务状态处理代码。

- [ ] **Step 4: 让设置预览复用同一显示器**

`PetPreview` 中按同一个 `pet.renderer` 分支渲染。小家伙使用：

```jsx
<LittleCompanionSprite state="idle" preview label={pet.displayName} />
```

CM 和貔貅继续渲染原 `<img>`，确保现有预览裁切不变。

- [ ] **Step 5: 添加稳定尺寸和裁切样式**

`.little-companion-sprite` 占满父级并以 120 x 120 单元格居中；内部图片宽高分别为列数和行数乘以单元格尺寸，位移为列号和行号乘以负单元格尺寸。设置预览通过父级 CSS 变量把单元格改为 96px。占位文字不得撑大 120 x 130 的外壳。

现有 `.cm-penguin-pet > img` 规则要收窄到旧图集专用类，避免把 960 x 1430 尺寸套到小家伙图片上。为旧 `<img>` 增加 `className="cm-penguin-atlas"`。

- [ ] **Step 6: 运行逻辑测试和前端构建**

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js`

Expected: PASS。

Run: `npm --prefix frontend run build`

Expected: exit 0，无 JSX、导入或 CSS 构建错误。构建会更新已有生成目录；不要把生成文件当源码实现，也不要清理用户原有生成差异。

- [ ] **Step 7: 检查本任务差异**

Run: `git diff --check -- frontend/src/shared/pet/LittleCompanionSprite.jsx frontend/src/shared/pet/CmPenguinCompanion.jsx frontend/src/shared/pet/cm-penguin-companion.css frontend/src/user/pages/SettingsPage.jsx frontend/src/shared/styles/global.css`

Expected: exit 0。只记录结果，不提交。

---

### Task 4: Go embed 构建和真实页面验收

**Files:**
- Verify: `scripts/build-go-local.ps1`
- Generated verification only: `backend/web/dist/pets/little-companion/*.png`
- Verify: 当前本地 Go 服务和浏览器页面

**Interfaces:**
- Consumes: 根目录 `pets/little-companion/*` 和前端构建结果。
- Produces: Go embed 中可通过 `/pets/little-companion/<pose>.png` 读取的八张图片。

- [ ] **Step 1: 运行完整本地 Go embed 构建**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

Expected: 前端构建和 Go 构建成功；脚本把根 `pets/` 完整复制到 `backend/web/dist/pets/`。

- [ ] **Step 2: 核对嵌入资源完整性**

检查 `backend/web/dist/pets/little-companion/` 恰好包含 `pet.json` 与八张动作 PNG；逐张比较源文件和生成文件 SHA-256，结果必须一致。缺失、额外运行依赖或哈希不一致均视为失败。

- [ ] **Step 3: 启动本地服务并检查静态响应**

使用仓库现有本地 Go 启动脚本和当前批准的本地配置，不修改数据库迁移记录。确认八个 `/pets/little-companion/*.png` 请求返回 200 且内容类型为图片。

- [ ] **Step 4: 在已登录浏览器完成真实交互**

依次验证：

1. 设置页宠物选择共有 CM、貔貅、小家伙三项；
2. 选择小家伙后立即显示待机动画，保存后刷新仍是小家伙；
3. 等待完整空闲轮播，六种空闲姿态均出现且容器不跳动；
4. 拖拽位置、点击打开 AI 对话、关闭对话均正常；
5. 触发一个真实任务，执行中显示思考，成功显示胜利/跳舞；失败路径用可控失败请求检查抖动和恢复；
6. 切换回 CM 和貔貅，旧图集尺寸、帧速和拖拽不回归；
7. 桌面和窄屏各检查一次，无文字、气泡和角色重叠；
8. 开启系统“减少动态效果”后，角色停止自动轮播。

保存桌面和窄屏截图作为本次本地验收证据，但不写入仓库。

- [ ] **Step 5: 最终验证**

Run: `node --test frontend/src/shared/pet/littleCompanionMotion.test.js frontend/src/shared/pet/petCatalog.test.js`

Run: `npm --prefix frontend test`

Run: `npm --prefix frontend run build`

Run from `backend`: `go test ./internal/web ./internal/httpapi -count=1`

Run: `git diff --check -- docs/superpowers/specs/2026-10-18-little-companion-third-pet-design.md docs/superpowers/plans/2026-10-18-little-companion-third-pet.md frontend/src/shared/pet frontend/src/user/pages/SettingsPage.jsx frontend/src/shared/styles/global.css pets/little-companion`

Expected: 所有命令 exit 0。记录确切测试数量、构建结果、浏览器检查结果、当前 `v88` HEAD 和本次修改路径；不声称公网已发布。

- [ ] **Step 6: 提交边界**

列出只属于本功能的文件并向用户汇报。由于未获得明确提交授权，停止在已验证的本地差异，不执行提交或推送。

