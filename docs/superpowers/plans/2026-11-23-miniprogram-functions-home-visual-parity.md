# 微信小程序功能首页视觉对齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将微信小程序“功能”首页重构为与 `docs/design-references/miniprogram-functions-first/01-functions-home.png` 同结构、同视觉比例的手机页面，同时保留七个真实功能入口和四项真实导航。

**Architecture:** 页面继续使用 Taro React 和现有路由配置，但把功能首页拆成清晰的顶部封面、继续创作、内容创作、生产成片、AI 工具五个视觉区块。独立 3D 插画以透明 WebP 放入 `src/assets/functions/parity/`，页面只负责排版；底部导航保留现有自定义 TabBar 逻辑，仅替换图标素材和尺寸。

**Tech Stack:** Taro 4.2.1、React 18.3.1、TypeScript、SCSS、Vitest、微信小程序原生自定义 TabBar、WebP 图片素材。

## Global Constraints

- 唯一视觉基准：`docs/design-references/miniprogram-functions-first/01-functions-home.png`。
- 保留现有七个 `moduleRoute` 功能入口和最后使用功能状态。
- 底部导航顺序固定为“功能、工作台、任务、我的”。
- 功能首页不得启用全屏原生动态 Canvas。
- 不修改功能内部页面、后端接口、提示词、数据库或任务执行逻辑。
- 不上传、不部署、不发布线上版本。
- 只在现有 `v88` 分支工作，不创建新分支；提交时只包含本计划涉及文件。

---

### Task 1: 建立高清插画资源合同

**Files:**
- Create: `miniprogram/src/assets/functions/parity/hero-dragons.webp`
- Create: `miniprogram/src/assets/functions/parity/private-book.webp`
- Create: `miniprogram/src/assets/functions/parity/novel-book.webp`
- Create: `miniprogram/src/assets/functions/parity/batch-books.webp`
- Create: `miniprogram/src/assets/functions/parity/video-clapper.webp`
- Create: `miniprogram/src/assets/functions/parity/ai-cut.webp`
- Create: `miniprogram/src/assets/functions/parity/comic-film.webp`
- Create: `miniprogram/src/assets/functions/parity/tools-orb.webp`
- Create: `miniprogram/src/pages/functions/functionVisuals.ts`
- Create: `miniprogram/src/pages/functions/functionVisuals.test.ts`

**Interfaces:**
- Consumes: `ModuleKey` from `src/config/modules.ts`.
- Produces: `FUNCTION_VISUALS: Record<ModuleKey, { asset: string; label: string; description: string }>` and `HERO_DRAGONS`.

- [ ] **Step 1: Write the failing asset-contract test**

```ts
import { expect, test } from "vitest";
import { FUNCTION_VISUALS, HERO_DRAGONS } from "./functionVisuals";

test("provides one dedicated high-resolution illustration for every function entry", () => {
  expect(Object.keys(FUNCTION_VISUALS)).toEqual(["script", "fetch", "batch", "shuihuo", "agent", "comic", "toolkit"]);
  expect(new Set(Object.values(FUNCTION_VISUALS).map(item => item.asset)).size).toBe(7);
  expect(HERO_DRAGONS).toMatch(/hero-dragons\.webp$/);
});
```

- [ ] **Step 2: Run the test and verify RED**

Run: `npm test -- --run src/pages/functions/functionVisuals.test.ts`

Expected: FAIL because `functionVisuals.ts` does not exist.

- [ ] **Step 3: Generate and inspect the eight transparent WebP assets**

Generate each asset from the approved reference with the same cyan/blue/purple/pink 3D material. Use a transparent background, centered subject, no text except the `AI` badge on `private-book.webp`, and enough empty edge space to avoid clipping. Inspect every generated image before copying it into `src/assets/functions/parity/`.

- [ ] **Step 4: Add the visual manifest**

```ts
import type { ModuleKey } from "../../config/modules";
import heroDragons from "../../assets/functions/parity/hero-dragons.webp";
import privateBook from "../../assets/functions/parity/private-book.webp";
import novelBook from "../../assets/functions/parity/novel-book.webp";
import batchBooks from "../../assets/functions/parity/batch-books.webp";
import videoClapper from "../../assets/functions/parity/video-clapper.webp";
import aiCut from "../../assets/functions/parity/ai-cut.webp";
import comicFilm from "../../assets/functions/parity/comic-film.webp";
import toolsOrb from "../../assets/functions/parity/tools-orb.webp";

export const HERO_DRAGONS = heroDragons;
export const FUNCTION_VISUALS: Record<ModuleKey, { asset: string; label: string; description: string }> = {
  script: { asset: privateBook, label: "私人定制", description: "按你的想法，生成专属内容" },
  fetch: { asset: novelBook, label: "小说创作", description: "获取小说与改文" },
  batch: { asset: batchBooks, label: "批量测书", description: "批量获取与生产" },
  shuihuo: { asset: videoClapper, label: "水货生产", description: "漫剧项目与素材生产" },
  agent: { asset: aiCut, label: "AI-CUT", description: "Agent 工作区" },
  comic: { asset: comicFilm, label: "漫剧创作", description: "小说一键漫剧，角色生动开演" },
  toolkit: { asset: toolsOrb, label: "会员与工具", description: "会员、用量与独立配音" },
};
```

- [ ] **Step 5: Run the test and verify GREEN**

Run: `npm test -- --run src/pages/functions/functionVisuals.test.ts`

Expected: PASS, one unique asset per module.

### Task 2: 重建顶部沉浸封面和功能卡结构

**Files:**
- Modify: `miniprogram/src/pages/functions/index.tsx`
- Modify: `miniprogram/src/pages/functions/functions.interaction.test.tsx`

**Interfaces:**
- Consumes: `FUNCTION_VISUALS` and `HERO_DRAGONS` from Task 1; `moduleRoute(key)` from `src/config/modules.ts`.
- Produces: `.functions-hero`, `.functions-resume`, `.function-layout`, `.function-card` DOM structure used by Task 3.

- [ ] **Step 1: Write failing structure tests**

Add assertions that the real page renders one immersive hero, one resume card, exactly seven `.function-card` buttons, one large card in each main section, and passes `dynamicGlass={false}` to `ThemedPage`.

```ts
const tree = render();
expect(tree.filter(n => n.props?.className?.includes("functions-hero"))).toHaveLength(1);
expect(tree.filter(n => n.props?.className?.includes("functions-resume"))).toHaveLength(1);
expect(tree.filter(n => n.type === "Button" && n.props?.className?.includes("function-card"))).toHaveLength(7);
expect(tree.find(n => n.type === "ThemedPage").props.dynamicGlass).toBe(false);
```

- [ ] **Step 2: Run the test and verify RED**

Run: `npm test -- --run src/pages/functions/functions.interaction.test.tsx`

Expected: FAIL because the new class structure is absent.

- [ ] **Step 3: Replace the compact header with the approved structure**

Remove `BrandHeader` from this page. Render the brand logo lockup, caption, title and subtitle inside `.functions-hero`; render the resume content and `HERO_DRAGONS` inside `.functions-resume`. Use `FUNCTION_VISUALS` for every card and keep `selectModule` unchanged.

- [ ] **Step 4: Run the test and verify GREEN**

Run: `npm test -- --run src/pages/functions/functions.interaction.test.tsx`

Expected: all functions interaction tests PASS and every route still uses `Taro.navigateTo({ url: moduleRoute(key) })`.

### Task 3: 对齐手机尺寸、卡片比例和背景层次

**Files:**
- Replace: `miniprogram/src/pages/functions/gallery.scss`
- Modify: `miniprogram/src/ui/liquidGlassPages.test.ts`
- Modify: `miniprogram/src/ui/wxssCompatibility.test.ts` only if the new page introduces a real incompatibility.

**Interfaces:**
- Consumes: DOM class names from Task 2.
- Produces: responsive 390px baseline layout with the exact section and card proportions from the approved spec.

- [ ] **Step 1: Add failing page-style assertions**

Assert the function page stylesheet contains the required baseline values and excludes the compressed legacy heights:

```ts
expect(styles).toContain(".functions-resume");
expect(styles).toContain("height:190PX");
expect(styles).toContain("height:232PX");
expect(styles).toContain("height:110PX");
expect(styles).not.toContain("height:118PX");
expect(styles).not.toContain("height:140PX");
```

- [ ] **Step 2: Run the test and verify RED**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts`

Expected: FAIL because the current page still uses 118px/140px compressed cards.

- [ ] **Step 3: Implement the responsive SCSS**

Use a 390px baseline with 16px horizontal padding, 300px hero, 190px resume card, 232px large cards, 110px small cards, 10px card gaps and 38px section gaps. Set the upper background to a vivid sunset image with a dark lower fade; keep the lower canvas deep navy. Use only class selectors supported by WeChat WXSS.

- [ ] **Step 4: Verify focused styles**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts src/ui/wxssCompatibility.test.ts`

Expected: function-page assertions PASS; no new incompatible selector originates from `pages/functions/index.wxss`.

### Task 4: 重绘四项底部导航图标和选中态

**Files:**
- Create: `miniprogram/src/assets/icons/parity/tab-functions-selected.webp`
- Create: `miniprogram/src/assets/icons/parity/tab-functions-normal.webp`
- Create: `miniprogram/src/assets/icons/parity/tab-workspace-normal.webp`
- Create: `miniprogram/src/assets/icons/parity/tab-tasks-normal.webp`
- Create: `miniprogram/src/assets/icons/parity/tab-profile-normal.webp`
- Modify: `miniprogram/src/custom-tab-bar/index.tsx`
- Modify: `miniprogram/src/custom-tab-bar/index.scss`
- Modify: `miniprogram/src/ui/studioNavigation.test.ts`

**Interfaces:**
- Consumes: existing `STUDIO_NAV_ITEMS` and real tab routes.
- Produces: image-backed bottom navigation icons with selected cyan clover and unselected white-gray outlines.

- [ ] **Step 1: Write the failing navigation visual test**

Assert the custom TabBar maps all four routes to dedicated image assets, keeps the exact order `functions, workspace, tasks, profile`, and renders `.studio-tab-active-line` only for the selected item.

- [ ] **Step 2: Run the test and verify RED**

Run: `npm test -- --run src/ui/studioNavigation.test.ts`

Expected: FAIL because current navigation still renders generic `AppIcon` shapes.

- [ ] **Step 3: Generate and wire the five navigation assets**

Generate the cyan selected functions icon and four white-gray outline icons on transparent backgrounds. Replace generic `AppIcon` rendering with `Image`, preserve `Taro.switchTab`, and keep text labels unchanged.

- [ ] **Step 4: Implement the navigation dimensions**

Use a 72px navigation body plus safe area, 27–30px icons, 12px labels and a 30px by 2px cyan active line. Keep the material opaque enough to remain readable over the page.

- [ ] **Step 5: Run navigation tests and verify GREEN**

Run: `npm test -- --run src/ui/studioNavigation.test.ts src/app.config.test.ts`

Expected: PASS with four correct routes and the requested order.

### Task 5: Build, open and perform visual QA

**Files:**
- Create: `miniprogram/design-qa.md`
- Modify only if QA finds a reference mismatch: `miniprogram/src/pages/functions/index.tsx`, `miniprogram/src/pages/functions/gallery.scss`, or the new parity assets.

**Interfaces:**
- Consumes: completed Tasks 1–4 and the approved reference PNG.
- Produces: a fresh WeChat build, implementation screenshot and blocking QA report.

- [ ] **Step 1: Run focused tests**

Run:

```powershell
npm test -- --run src/pages/functions/functionVisuals.test.ts src/pages/functions/functions.interaction.test.tsx src/theme.test.ts src/ui/liquidGlassPages.test.ts src/ui/studioNavigation.test.ts src/app.config.test.ts
```

Expected: all focused tests PASS.

- [ ] **Step 2: Close the WeChat project before building**

Run `cli.bat close --project F:\脚本测试\一战晟铭-v88\miniprogram` and confirm no Taro watch process is writing to `dist`.

- [ ] **Step 3: Build exactly once**

Run: `npm run build:weapp:sandbox`

Expected: exit 0, `dist/app.json` exists, main package remains below 2,097,152 bytes, and `dist/pages/functions/index.wxss` contains no unsupported pseudo selectors.

- [ ] **Step 4: Open the WeChat project and capture the same viewport**

Open the project with the official CLI. Capture the functional page at the same 390px viewport used for the reference comparison. Do not treat `webview page ready` as visual proof.

- [ ] **Step 5: Write blocking design QA**

In `miniprogram/design-qa.md`, record reference path, implementation screenshot path, viewport, hero height, resume height, large/small card heights, icon differences, navigation differences and console state. Set `final result: passed` only when P0/P1/P2 issues are fixed; otherwise set `final result: blocked`.

- [ ] **Step 6: Commit only the function-home implementation**

Stage the new parity assets, function-page source/tests, custom TabBar source/tests and `design-qa.md`. Verify `git diff --cached --name-only` contains no unrelated files, then commit with `feat: align miniprogram functions home with reference`.
