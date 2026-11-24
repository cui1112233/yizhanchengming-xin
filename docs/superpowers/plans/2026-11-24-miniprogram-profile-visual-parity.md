# 小程序“我的空间”视觉还原 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将小程序第四个底部页面按 `docs/design-references/miniprogram-functions-first/04-profile.png` 还原，并让身份、任务统计和额度全部显示真实数据。

**Architecture:** 保留现有账号会话、会员资料接口和详情页面。新增一个纯数据转换模块负责把会员资料和任务汇总转换成首页可展示的数据；页面分别读取账号资料和任务统计，任一来源失败都不阻止另一部分显示。布局继续使用现有 `ThemedPage`、四项自定义导航和液态玻璃主题。

**Tech Stack:** Taro 4、React 18、TypeScript、SCSS、Vitest、微信小程序开发者工具。

## Global Constraints

- 唯一视觉基线：`docs/design-references/miniprogram-functions-first/04-profile.png`。
- 只修改“我的”首页及直接需要的数据、样式、测试和图标资源。
- 不写死示例头像、角色、任务数量或额度。
- 进行中和已完成必须复用任务中心的真实汇总规则。
- 所有入口要么进入真实功能，要么明确显示“暂未开放”。
- 兼容 320px 至 430px 手机宽度。
- 直接提交到当前 `v88`，不新建分支、不推送、不上传。

---

### Task 1: 个人中心首页数据模型

**Files:**
- Create: `miniprogram/src/services/profileOverview.ts`
- Create: `miniprogram/src/services/profileOverview.test.ts`
- Reuse: `miniprogram/src/services/taskSummary.ts`
- Reuse: `miniprogram/src/config/profileData.ts`

**Interfaces:**
- Consumes: `TaskSummary`、`PublicMemberCenter`、`profileFacts(center)`。
- Produces: `profileTaskMetrics(summary): { running?: number; completed?: number }`。
- Produces: `profileQuota(center): { value: string; caption: string }`。

- [ ] **Step 1: Write failing tests for real task counts and quota fallbacks**

```ts
expect(profileTaskMetrics(summary)).toEqual({ running: 2, completed: 28 });
expect(profileTaskMetrics(null)).toEqual({ running: undefined, completed: undefined });
expect(profileQuota(unlimitedCenter)).toEqual({ value: "不限额", caption: "本月剩余" });
expect(profileQuota(missingCenter)).toEqual({ value: "暂未返回", caption: "创作额度" });
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `npx vitest run src/services/profileOverview.test.ts`

Expected: FAIL because `profileOverview.ts` does not exist.

- [ ] **Step 3: Implement pure conversion helpers**

```ts
export function profileTaskMetrics(summary: TaskSummary | null) {
  if (!summary) return { running: undefined, completed: undefined };
  const counts = taskCounts(summary.entries);
  return { running: counts.running, completed: counts.completed };
}
```

额度转换必须区分真实的 `0`、不限额和字段缺失，不能用 `||` 把 0 当成空值。

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `npx vitest run src/services/profileOverview.test.ts src/services/taskSummary.test.ts`

Expected: PASS.

- [ ] **Step 5: Commit the data model**

```bash
git add miniprogram/src/services/profileOverview.ts miniprogram/src/services/profileOverview.test.ts
git commit -m "feat(miniprogram): add real profile overview data"
```

### Task 2: 页面结构与真实入口

**Files:**
- Modify: `miniprogram/src/pages/profile/index.tsx`
- Modify: `miniprogram/src/pages/profile/index.test.tsx`
- Modify: `miniprogram/src/pages/profile/profile-iconography.test.ts`

**Interfaces:**
- Consumes: `loadTaskSummary(apiRequest)`、`profileTaskMetrics(summary)`、`profileQuota(center)`。
- Keeps: existing profile loading, avatar upload, logout, appearance sheet and session-revision protection.
- Produces: the ordered sections `顶部场景 → 创作者身份卡 → 创作额度卡 → 功能列表 → 降低透明度`。

- [ ] **Step 1: Write failing render and interaction tests**

```tsx
expect(source).toContain("我的空间");
expect(source).not.toContain('studio-page-heading"><Text>我的</Text>');
expect(source).toContain("我的资产");
expect(source).toContain("创作偏好");
expect(source).toContain("会员与用量");
expect(source).toContain("账号与安全");
expect(source).toContain("帮助与反馈");
```

交互测试验证：我的资产切换到 `/pages/workspace/index`；创作偏好打开现有设置面板；会员与用量打开 `memberCenter`；账号与安全打开 `security`；帮助与反馈只提示暂未开放。

- [ ] **Step 2: Run focused tests and verify RED**

Run: `npx vitest run src/pages/profile/index.test.tsx src/pages/profile/profile-iconography.test.ts`

Expected: FAIL because the old page still contains duplicate heading and old section structure.

- [ ] **Step 3: Recompose the page without changing account contracts**

```tsx
<View className="profile-space-hero">
  <Text className="profile-space-title">我的空间</Text>
  <Text className="profile-space-subtitle">账号、资产与创作偏好</Text>
</View>
```

任务统计作为独立异步状态读取；任务读取失败时显示“暂未读到”，会员资料继续显示。额度卡使用 `profileQuota`。功能入口必须调用现有导航和面板，不创建空白详情页。

- [ ] **Step 4: Run focused tests and verify GREEN**

Run: `npx vitest run src/pages/profile/index.test.tsx src/pages/profile/profile-iconography.test.ts src/services/profileOverview.test.ts`

Expected: PASS.

- [ ] **Step 5: Commit page structure**

```bash
git add miniprogram/src/pages/profile/index.tsx miniprogram/src/pages/profile/index.test.tsx miniprogram/src/pages/profile/profile-iconography.test.ts
git commit -m "feat(miniprogram): rebuild profile space layout"
```

### Task 3: 视觉样式和匹配图标

**Files:**
- Modify: `miniprogram/src/pages/profile/parity.scss`
- Add only if required: `miniprogram/src/assets/profile/*.webp`
- Modify: `miniprogram/src/pages/profile/profile-palette.test.tsx`

**Interfaces:**
- Consumes the class names introduced in Task 2.
- Reuses existing brand logo, late-sunset wallpaper and available 3D assets where their meaning matches.
- Adds dedicated WebP profile icons only for missing meanings; generated assets contain no text or numbers.

- [ ] **Step 1: Write failing visual-contract tests**

Assert that the stylesheet contains compact mobile tokens for 12–14px page gutters, identity-card aspect ratio, single-line overflow handling, bottom safe-area spacing, selected navigation color, and reduced-transparency variants.

- [ ] **Step 2: Run visual-contract tests and verify RED**

Run: `npx vitest run src/pages/profile/profile-palette.test.tsx src/pages/profile/profile-iconography.test.ts`

Expected: FAIL for missing new profile-space classes.

- [ ] **Step 3: Implement the reference-aligned SCSS**

The SCSS must preserve the visual order and proportions of `04-profile.png`, but scale the large concept art down to a 375px phone. Use `clamp()` only where the WeChat compiler supports the emitted WXSS; otherwise use rpx/px rules already established in this project.

- [ ] **Step 4: Run focused tests and build**

Run: `npx vitest run src/pages/profile/profile-palette.test.tsx src/pages/profile/profile-iconography.test.ts`

Run: `npm run build:weapp:sandbox`

Run: `node scripts/sanitize-weapp-wxss.mjs`

Expected: tests pass, build exits 0, unsupported pseudo selectors are removed.

- [ ] **Step 5: Commit visual parity**

```bash
git add miniprogram/src/pages/profile/parity.scss miniprogram/src/pages/profile/profile-palette.test.tsx miniprogram/src/pages/profile/profile-iconography.test.ts miniprogram/src/assets/profile
git commit -m "style(miniprogram): match profile visual reference"
```

### Task 4: 微信开发者工具视觉验收

**Files:**
- Create: `miniprogram/.qa/profile-space-final-window.png`
- Create: `miniprogram/.qa/profile-space-final-phone.png`
- Create: `miniprogram/.qa/profile-space-design-qa.md`

**Interfaces:**
- Consumes the final WeChat build from Task 3.
- Produces a phone screenshot and an acceptance record against `04-profile.png`.

- [ ] **Step 1: Run the full suite**

Run: `npm test`

Expected: all non-skipped tests pass.

- [ ] **Step 2: Compile and open the project in WeChat DevTools**

Run: `npm run build:weapp:sandbox`

Run: `node scripts/sanitize-weapp-wxss.mjs`

Run: `F:\软件\微信web开发者工具\cli.bat open --project F:\脚本测试\一战晟铭-v88\miniprogram --trust-project`

- [ ] **Step 3: Capture and compare the real profile page**

Verify: no black screen, no console runtime errors, no duplicate title, no text overlap, the five rows are visible by scrolling, and all displayed values are real or explicitly unavailable.

- [ ] **Step 4: Record QA result**

`profile-space-design-qa.md` must list the reference, phone size, data state, console result, visual differences, and final `pass` or `blocked` result.

- [ ] **Step 5: Commit only source and intentional QA evidence**

Do not stage unrelated dirty files, generated build output, credentials, private project configuration or the browser companion directory.
