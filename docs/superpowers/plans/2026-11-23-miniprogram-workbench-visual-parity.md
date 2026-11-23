# 小程序工作台原稿对齐实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把第二个主导航“工作台”改成当前功能的真实执行中心，并在手机尺寸下对齐已确认的工作台原稿。

**Architecture:** 保留现有账号门禁、模块配置、任务聚合和真实路由。新增一个无界面的工作台状态辅助文件，负责模块与任务来源映射、当前任务选择、状态文案和进度计算；页面只负责渲染，样式文件负责原稿布局。所有展示值来自 `AggregatedTask`，缺失字段整行隐藏。

**Tech Stack:** Taro 4.2、React 18、TypeScript、SCSS、Vitest、微信小程序构建。

## Global Constraints

- 只优化 `pages/workspace`，不修改功能页、任务页、我的页面和后端接口。
- 不伪造项目名、百分比、页数、模型或更新时间。
- 系统提示词不在前端硬编码。
- 导航顺序保持“功能、工作台、任务、我的”。
- 在 `390 × 844` 手机视口验收，底部导航不得遮挡内容。
- 不新建分支、不上传、不部署、不提交无关脏文件。

---

### Task 1: 建立真实工作台状态模型

**Files:**
- Create: `miniprogram/src/pages/workspace/workbenchState.ts`
- Create: `miniprogram/src/pages/workspace/workbenchState.test.ts`

**Interfaces:**
- Consumes: `ModuleKey`、`AggregatedTask`。
- Produces: `taskSourceForModule(key: ModuleKey)`、`selectCurrentTask(key: ModuleKey, entries: AggregatedTask[])`、`workbenchStatus(task?: AggregatedTask)`、`workbenchProgress(task?: AggregatedTask)`。

- [ ] **Step 1: 写失败测试**

```ts
test("selects only the current module task", () => {
  const entries = [task("fetch", "other"), task("script", "current")];
  expect(selectCurrentTask("script", entries)?.id).toBe("current");
});

test("never invents progress", () => {
  expect(workbenchProgress(task("script", "a"))).toBeUndefined();
  expect(workbenchProgress({ ...task("script", "b"), progress: { completed: 3, total: 4 } }))
    .toEqual({ completed: 3, total: 4, percent: 75 });
});
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- src/pages/workspace/workbenchState.test.ts`

Expected: FAIL，提示 `workbenchState` 或导出函数不存在。

- [ ] **Step 3: 实现最小状态模型**

```ts
const MODULE_SOURCES: Partial<Record<ModuleKey, TaskSource>> = {
  script: "script", fetch: "fetch", batch: "batch", shuihuo: "shuihuo", agent: "agent",
};

export function selectCurrentTask(key: ModuleKey, entries: AggregatedTask[]) {
  const source = MODULE_SOURCES[key];
  return source ? entries.find(entry => entry.source === source) : undefined;
}

export function workbenchProgress(task?: AggregatedTask) {
  const progress = task?.progress;
  if (!progress || progress.total <= 0) return undefined;
  return { ...progress, percent: Math.round(progress.completed / progress.total * 100) };
}
```

状态文案严格由 `running / failed / completed / unknown` 转换，不读取不到的数据。

- [ ] **Step 4: 运行测试并确认通过**

Run: `npm test -- src/pages/workspace/workbenchState.test.ts`

Expected: PASS。

- [ ] **Step 5: 提交状态模型**

```powershell
git add -- miniprogram/src/pages/workspace/workbenchState.ts miniprogram/src/pages/workspace/workbenchState.test.ts
git commit -m "feat: model current miniprogram workbench task"
```

### Task 2: 把工作台改成单一执行中心

**Files:**
- Modify: `miniprogram/src/pages/workspace/index.tsx`
- Modify: `miniprogram/src/pages/workspace/studio.interaction.test.tsx`
- Reuse: `miniprogram/src/pages/functions/functionVisuals.ts`

**Interfaces:**
- Consumes: Task 1 的状态函数、`FUNCTION_VISUALS`、`moduleRoute`、`taskDestination`。
- Produces: `.workbench-current-card`、`.workbench-project-card`、`.workbench-detail-card`、`.workbench-empty` 页面结构。

- [ ] **Step 1: 改写交互测试并确认旧页面失败**

测试必须断言：

```ts
expect(find("studio-entry-grid")).toBeUndefined();
expect(find("workbench-current-card")).toBeDefined();
expect(find("workbench-primary-action").props.children).toContain("开始配置");
```

再增加有任务、失败任务、数据源失败和无任务四种状态；有任务时按钮恢复真实 ID，无任务时进入 `moduleRoute(activeModule.key)`。

Run: `npm test -- src/pages/workspace/studio.interaction.test.tsx`

Expected: FAIL，旧页面仍存在功能入口网格。

- [ ] **Step 2: 移除重复功能入口和私人定制弹层**

删除 `workspaceShortcuts`、`privateOpen`、`studio-entry-grid` 渲染、私人定制选择弹层及其隐藏导航逻辑。保留登录、会话读取和重新加载能力。

- [ ] **Step 3: 读取当前模块并筛选当前任务**

在每次 `useDidShow` 时重新读取 `yzcm.mobile.activeModule`，加载任务摘要后使用 `selectCurrentTask` 选择当前功能任务。页面不得把其他来源的最近任务当作当前项目。

- [ ] **Step 4: 渲染原稿层级**

```tsx
<View className="workbench-current-card">
  <Image className="workbench-current-art" src={visual.asset} mode="aspectFit" />
  <Text className="workbench-current-label">当前工作区</Text>
  <Text className="workbench-current-title">{visual.label}</Text>
  <Button className="workbench-primary-action">{task ? "继续执行" : "开始配置"}</Button>
</View>
```

有真实任务时渲染项目卡；无任务时渲染“当前功能还没有可继续的任务”。详情卡只渲染存在的 `progress` 和 `createdAt`，不添加虚构模型或画幅。

- [ ] **Step 5: 运行交互测试并确认通过**

Run: `npm test -- src/pages/workspace/studio.interaction.test.tsx src/pages/workspace/workbenchState.test.ts`

Expected: PASS。

- [ ] **Step 6: 提交页面结构**

```powershell
git add -- miniprogram/src/pages/workspace/index.tsx miniprogram/src/pages/workspace/studio.interaction.test.tsx
git commit -m "feat: turn miniprogram workspace into execution center"
```

### Task 3: 对齐工作台原稿的手机视觉

**Files:**
- Modify: `miniprogram/src/pages/workspace/studio.scss`
- Test: `miniprogram/src/pages/workspace/studio.interaction.test.tsx`

**Interfaces:**
- Consumes: Task 2 的工作台类名。
- Produces: 390px 基准的晚霞背景、黑青玻璃卡片、青色状态、单一主按钮和安全区布局。

- [ ] **Step 1: 增加样式合同断言**

读取 `studio.scss` 并断言存在：

```ts
expect(styles).toContain(".workbench-current-card");
expect(styles).toContain(".workbench-project-card");
expect(styles).toContain("padding-bottom:calc(96PX + env(safe-area-inset-bottom))");
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- src/pages/workspace/studio.interaction.test.tsx`

Expected: FAIL，新工作台样式尚不存在。

- [ ] **Step 3: 实现手机布局**

按 390px 视口设置：页面左右留白 15px；当前工作区卡约 128px；项目卡最小 210px；模块标题 20px；正文 13–15px；按钮高度 48px。玻璃边框使用现有 `--liquid-edge`，关键状态使用现有青色变量，不新增另一套主题。

项目卡进度和阶段只在真实数据存在时显示；没有进度时卡片收缩，不留空白占位。页面底部使用 `calc(96PX + env(safe-area-inset-bottom))`。

- [ ] **Step 4: 运行工作台测试**

Run: `npm test -- src/pages/workspace/studio.interaction.test.tsx src/pages/workspace/workbenchState.test.ts`

Expected: PASS。

- [ ] **Step 5: 提交视觉实现**

```powershell
git add -- miniprogram/src/pages/workspace/studio.scss miniprogram/src/pages/workspace/studio.interaction.test.tsx
git commit -m "style: align miniprogram workbench with reference"
```

### Task 4: 微信端视觉验收和完整验证

**Files:**
- Modify: `miniprogram/design-qa.md`
- Create: `miniprogram/qa-workbench-reference.png`
- Create: `miniprogram/qa-workbench-implementation.png`
- Create: `miniprogram/qa-workbench-comparison.png`

**Interfaces:**
- Consumes: `docs/design-references/miniprogram-functions-first/02-workbench.png` 与微信开发者工具实际画面。
- Produces: 同尺寸对比图和可复查验收结论。

- [ ] **Step 1: 运行工作台聚焦测试**

Run: `npm test -- src/pages/workspace/workbenchState.test.ts src/pages/workspace/studio.interaction.test.tsx src/pages/primary-navigation.test.ts`

Expected: PASS，0 failure。

- [ ] **Step 2: 运行全量测试**

Run: `npm test`

Expected: 所有非跳过测试通过。

- [ ] **Step 3: 构建正式微信包**

Run: `npm run build:weapp`

Expected: `Compiled successfully`，无 WXSS 编译错误。

- [ ] **Step 4: 在微信开发者工具验收**

打开第二个“工作台”，分别检查有任务与无任务状态。保存 390 × 844 画面，与原稿裁成同尺寸并排比较；检查页面可见、主卡比例、字体层级、插画清晰度、滚动范围、底部导航和控制台错误。

- [ ] **Step 5: 记录差异并修复 P0–P2**

在 `design-qa.md` 记录每轮差异。黑屏、错路由、假数据、遮挡为 P0/P1；明显比例、字体、间距和图像错误为 P2；只有不影响使用的轻微光效差异允许保留为 P3。

- [ ] **Step 6: 提交验收证据**

```powershell
git add -- miniprogram/design-qa.md miniprogram/qa-workbench-reference.png miniprogram/qa-workbench-implementation.png miniprogram/qa-workbench-comparison.png
git commit -m "test: verify miniprogram workbench visual parity"
```
