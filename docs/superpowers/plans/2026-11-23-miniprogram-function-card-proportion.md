# 微信小程序功能卡片比例还原 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将小程序功能首页的三组“一大两小”卡片按用户提供的 418px 参考截图等比例还原到 390px 手机宽度，同时保留九个真实入口。

**Architecture:** 保持 `FunctionsPage` 的九功能数据、路由和点击逻辑不变，只调整 `gallery.scss` 的 Flex 几何、字号、间距和插画盒。使用源码级样式合同测试防止尺寸回退，再用微信开发者工具同尺寸截图完成视觉验收。

**Tech Stack:** Taro 4.2.1、React 18、SCSS、Vitest、微信开发者工具。

## Global Constraints

- 仅修改功能首页卡片区域，不修改顶部主视觉、底部导航、后端或功能内部页面。
- 390px 基准：左右边距 15px，主卡约 220px，小卡列约 132px，列间距 9–10px。
- 主卡高 136–140px；小卡高 59–61px；小卡间距 12–15px。
- 分区标题约 19px，主卡标题 20–21px，小卡标题 13–14px。
- 保留 Flex，不使用 CSS Grid；保留 `aspectFit`，不拉伸插画。
- 九个入口、最近使用状态和“继续创作”逻辑必须保持不变。
- 不上传、不发布、不部署。

---

### Task 1: 锁定卡片比例合同

**Files:**
- Modify: `miniprogram/src/pages/functions/functions.interaction.test.tsx`
- Modify: `miniprogram/src/pages/functions/gallery.scss`

**Interfaces:**
- Consumes: `FunctionsPage` 当前输出的 `.gallery-feature-layout`、`.gallery-feature`、`.gallery-compact-stack` 和 `.gallery-compact` 类名。
- Produces: 三组统一的 Flex 卡片几何；现有点击处理函数与路由接口不变。

- [ ] **Step 1: 写入会失败的尺寸合同测试**

在现有 `uses WeChat-safe sizing instead of auto-height images or CSS grid` 测试中补充以下断言：

```ts
expect(galleryStyles).toContain("padding:0 15px")
expect(galleryStyles).toContain("gap:10px")
expect(galleryStyles).toContain("height:138px")
expect(galleryStyles).toContain("height:60px")
expect(galleryStyles).toContain("font-size:20px")
expect(galleryStyles).toContain("font-size:13px")
```

- [ ] **Step 2: 运行测试并确认旧尺寸导致失败**

Run: `npm test -- --run src/pages/functions/functions.interaction.test.tsx`

Expected: FAIL，至少缺少 `padding:0 15px`、`height:138px` 或 `height:60px`。

- [ ] **Step 3: 最小化修改卡片几何和字号**

在 `gallery.scss` 中采用以下基准值：

```scss
.gallery-body { padding: 0 15PX; }
.gallery-feature-layout.gallery-mosaic { gap: 10PX; }
.gallery-mosaic > .is-feature { flex: 1.61 1 0; }
.gallery-compact-stack { flex: .97 1 0; gap: 14PX; }
.gallery-feature.tools-entry { height: 138PX; }
.gallery-compact.tools-entry { height: 62PX; }
.gallery-section-title { font-size: 19PX; }
.gallery-feature-name { font-size: 20PX; }
.gallery-compact-title { font-size: 13PX; }
.gallery-feature-art { height: 112PX; }
.gallery-compact-art { width: 46PX; height: 46PX; flex: 0 0 46PX; }
```

小卡高度使用 62px 是为了让 `62 + 14 + 62 = 138`，保持两列严格等高；相对参考目标 59–61px 的偏差约 1px，仍在 5% 验收范围内。

- [ ] **Step 4: 运行聚焦测试确认通过**

Run: `npm test -- --run src/pages/functions/functions.interaction.test.tsx`

Expected: 9 tests PASS。

- [ ] **Step 5: 本地提交比例调整**

```bash
git add miniprogram/src/pages/functions/functions.interaction.test.tsx miniprogram/src/pages/functions/gallery.scss
git commit -m "fix: align miniprogram function card proportions"
```

### Task 2: 编译和微信视觉验收

**Files:**
- Modify: `miniprogram/design-qa.md`

**Interfaces:**
- Consumes: Task 1 生成的卡片样式和现有九张 WebP 插画。
- Produces: 可在微信开发者工具直接查看的 `miniprogram/dist` 与带有截图结论的 QA 记录。

- [ ] **Step 1: 运行完整测试**

Run: `npm test`

Expected: 所有非跳过测试通过；已有 React DOM 警告可以记录但不得出现新失败。

- [ ] **Step 2: 重新构建微信生产包**

Run: `npm run build:weapp`

Expected: `Compiled successfully`，WXSS 兼容处理完成。

- [ ] **Step 3: 重新载入开发者工具并截取手机区域**

Run: `F:\软件\微信web开发者工具\cli.bat auto --project F:\脚本测试\一战晟铭-v88\miniprogram --trust-project --lang zh`

Expected: IDE 成功载入 `pages/functions/index`，控制台无业务错误，页面不黑屏。

- [ ] **Step 4: 同尺寸视觉比较**

将开发者工具截图与 `C:\Users\ADMINI~1\AppData\Local\Temp\codex-clipboard-005be6d9-da88-476b-8515-cac7c45e3001.png` 对比，逐项记录：左右边距、两列宽度、主卡高度、小卡高度、标题间距、插画可视尺寸和底栏遮挡。

Expected: 几何偏差不超过约 5%，九个入口完整，插画不糊、不裁切。

- [ ] **Step 5: 更新 QA 结论并提交**

在 `miniprogram/design-qa.md` 写明参考图、实现截图、测试结果和 `final result: passed`；若无法获得同尺寸截图则必须写 `final result: blocked`。

```bash
git add miniprogram/design-qa.md
git commit -m "docs: verify miniprogram function card parity"
```
