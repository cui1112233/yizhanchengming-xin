# 小程序“我的空间”身份卡氛围增强 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保留真实用户头像和真实任务数据的前提下，为“我的空间”身份卡增加薰衣草晚霞景深、头像光环和缓慢漂浮的花瓣粒子。

**Architecture:** 继续使用现有 `ProfilePage` 和真实数据流，只给 `profile-space-identity` 增加独立的背景、粒子和光环表现层。粒子由一个小尺寸透明 WebP 资产和固定的七个渲染槽组成，动画只使用小程序支持的 CSS 关键帧；“降低透明度”和系统减少动画设置负责关闭动效。

**Tech Stack:** Taro 4、React 18、SCSS、微信小程序 CSS 动画、Vitest、内置 Image Gen。

## Global Constraints

- 左侧圆形内容始终是当前用户的真实头像；没有头像时继续使用姓名首字母。
- 名称、角色、团队、进行中数量、已完成数量继续来自真实接口，不写入示例数字。
- 只优化身份卡，不调整创作额度、入口列表、底部导航或其他页面。
- 花瓣必须使用透明图片资产，不用 emoji、文字字符或手绘 SVG 代替。
- 开启“降低透明度”或系统减少动画时停止花瓣和头像呼吸动画。
- 不新增接口，不上传、不部署、不新增分支。

---

### Task 1: 花瓣粒子结构与真实头像保护

**Files:**
- Create: `miniprogram/src/assets/profile/profile-petal-v1.webp`
- Modify: `miniprogram/src/pages/profile/index.tsx`
- Modify: `miniprogram/src/pages/profile/index.test.tsx`

**Interfaces:**
- Consumes: `reducedGlass: boolean`、现有 `facts.avatarUrl`、`taskMetrics`。
- Produces: `.profile-space-particles` 容器、七个 `.profile-space-petal` 图片和 `.is-reduced-motion` 卡片状态类。

- [ ] **Step 1: 写失败测试，锁定真实头像和七个粒子槽**

在 `index.test.tsx` 增加测试：加载带真实 `avatarUrl` 的账号，断言身份卡仍渲染该头像；查找 `.profile-space-petal`，断言数量为 7，所有粒子使用同一个本地 WebP；开启降低透明度时身份卡包含 `is-reduced-motion`。

- [ ] **Step 2: 运行测试确认失败**

Run: `npm test -- --run src/pages/profile/index.test.tsx`

Expected: FAIL，提示缺少 `.profile-space-petal` 或数量不是 7。

- [ ] **Step 3: 生成并保存透明花瓣资产**

使用内置 Image Gen 生成单枚粉紫色半透明花瓣：软 3D 插画质感、青紫边缘光、透明背景、无文字、无水印。检查后转为 160×160 以内的无损 WebP，保存为 `profile-petal-v1.webp`。

- [ ] **Step 4: 添加最小粒子结构**

在 `index.tsx` 中导入花瓣资产，定义固定七项：

```ts
const PROFILE_PETALS = ["one", "two", "three", "four", "five", "six", "seven"] as const;
```

身份卡内增加：

```tsx
<View className="profile-space-particles" ariaHidden>
  {PROFILE_PETALS.map(name => <Image key={name} className={`profile-space-petal is-${name}`} src={petalIcon} mode="aspectFit" />)}
</View>
```

身份卡类名在 `reducedGlass` 为真时增加 `is-reduced-motion`。真实头像和首字母逻辑保持不变。

- [ ] **Step 5: 运行测试确认通过**

Run: `npm test -- --run src/pages/profile/index.test.tsx`

Expected: 现有测试和新增粒子测试全部 PASS。

- [ ] **Step 6: 提交**

```bash
git add miniprogram/src/assets/profile/profile-petal-v1.webp miniprogram/src/pages/profile/index.tsx miniprogram/src/pages/profile/index.test.tsx
git commit -m "feat(miniprogram): add profile card petals"
```

### Task 2: 晚霞花田、头像光环和动画降级

**Files:**
- Modify: `miniprogram/src/pages/profile/parity.scss`
- Modify: `miniprogram/src/ui/liquidGlassPages.test.ts`

**Interfaces:**
- Consumes: Task 1 的 `.profile-space-particles`、`.profile-space-petal`、`.is-reduced-motion`。
- Produces: 三条花瓣轨迹、头像呼吸光、背景景深和静态降级样式。

- [ ] **Step 1: 写失败测试，锁定分层和降级选择器**

在 `liquidGlassPages.test.ts` 的 profile 测试中读取 `parity.scss`，断言包含 `@keyframes profile-petal-drift-a`、`.profile-space-identity.is-reduced-motion .profile-space-petal` 和 `@media (prefers-reduced-motion: reduce)`。

- [ ] **Step 2: 运行测试确认失败**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts`

Expected: FAIL，提示缺少粒子动画或降级规则。

- [ ] **Step 3: 实现视觉分层**

在 `parity.scss` 中：

- 身份卡使用项目现有 `wallpaper-dark.webp` 作为背景图并把花田定位在卡片底部；浅色主题改用 `wallpaper-light.webp`。
- 保留半透明深紫遮罩和底部粉紫雾光，保证正文对比度。
- 头像光环只改变阴影强度，不缩放头像。
- 七枚花瓣使用 9–16 秒周期、三条轨迹、不同大小和负延迟；粒子层 `pointer-events:none`、圆角内裁切。
- `.is-reduced-motion` 和 `prefers-reduced-motion: reduce` 将动画设为 `none`，保留静态花瓣构图。

- [ ] **Step 4: 运行视觉规则测试和微信样式兼容检查**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts src/ui/wxssCompatibility.test.ts`

Expected: 两个测试文件全部 PASS。

- [ ] **Step 5: 提交**

```bash
git add miniprogram/src/pages/profile/parity.scss miniprogram/src/ui/liquidGlassPages.test.ts
git commit -m "style(miniprogram): enrich profile identity atmosphere"
```

### Task 3: 构建与微信开发者工具视觉验收

**Files:**
- Modify: `miniprogram/.qa/profile-space-design-qa.md`
- Create: `miniprogram/.qa/profile-identity-atmosphere-phone.png`

**Interfaces:**
- Consumes: Task 1 和 Task 2 的最终身份卡。
- Produces: 可复查的测试、构建和手机截图证据。

- [ ] **Step 1: 运行全量小程序测试**

Run: `npm test`

Expected: 0 个失败；已有明确跳过项保持不变。

- [ ] **Step 2: 构建并清理微信样式兼容项**

Run: `npm run build:weapp:sandbox`

Expected: `Compiled successfully`，没有缺失导出警告。

Run: `node scripts/sanitize-weapp-wxss.mjs`

Expected: 脚本正常完成。

- [ ] **Step 3: 在微信开发者工具重新编译并打开“我的”**

清理编译缓存但保留登录和存储数据，重新编译后切换到“我的”。确认真实头像、真实任务数字和粒子层同时出现。

- [ ] **Step 4: 截图并进行同视口比较**

保存 iPhone 15 模拟器截图到 `miniprogram/.qa/profile-identity-atmosphere-phone.png`。对照用户参考图检查：背景景深、头像光环、花瓣层次、文字可读性、统计数字、粒子越界、底部导航。

- [ ] **Step 5: 更新验收记录并提交**

在 `profile-space-design-qa.md` 写明本轮截图、实际任务数字、动效降级结果和最终状态。

```bash
git add miniprogram/.qa/profile-space-design-qa.md miniprogram/.qa/profile-identity-atmosphere-phone.png
git commit -m "test(miniprogram): verify profile card atmosphere"
```
