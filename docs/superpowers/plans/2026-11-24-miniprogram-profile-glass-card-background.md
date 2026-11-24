# 小程序身份卡玻璃背景 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为“我的”页面身份卡生成独立的紫蓝晚霞花海背景，并用多层实时样式还原参考图的玻璃高光。

**Architecture:** 新背景图只承载晚霞、花海和环境光，不包含头像、文字或数字。`profile/index.tsx` 显式导入该素材，`profile/parity.scss` 分离背景、玻璃染色、边缘高光、折射和动态粒子层，继续复用真实账号与任务数据。

**Tech Stack:** Taro 4、React、TypeScript、SCSS、Vitest、微信小程序、内置 imagegen。

## Global Constraints

- 只修改“我的”页面顶部身份卡，不改变额度卡、设置列表、接口或四项底部导航。
- 背景不得包含头像、人物、双龙、按钮、图标、文字、数字、Logo 或水印。
- 真实头像、名称、团队、进行中和已完成数据必须继续由现有接口渲染。
- 保持现有身份卡宽度、圆角、内边距和页面占比，不扩大卡片。
- 开启“降低透明度”或系统减少动态时停止花瓣与头像动画。
- 不上传、不部署、不推送远端，只直接提交到当前 `v88` 分支。

---

### Task 1: 生成身份卡专用背景素材

**Files:**
- Create: `miniprogram/src/assets/profile/profile-identity-sunset-v2.webp`
- Test: `miniprogram/src/pages/profile/index.test.tsx`

**Interfaces:**
- Consumes: 用户参考图的紫蓝晚霞、底部花海和玻璃氛围。
- Produces: `profileIdentityBackground`，供身份卡页面的 `<Image>` 背景层使用。

- [ ] **Step 1: 写入失败测试**

在 `index.test.tsx` 的身份卡测试中断言独立背景层存在，并且背景层不是头像：

```ts
const background = findByClass(identity, "profile-space-identity-background");
expect(background.props.src).toContain("profile-identity-sunset-v2");
expect(background.props.src).not.toContain("avatar");
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- --run src/pages/profile/index.test.tsx`

Expected: FAIL，因为 `profile-space-identity-background` 尚不存在。

- [ ] **Step 3: 生成并整理素材**

使用内置 imagegen 生成横向纯背景：紫蓝暮色、远处花田、底部粉紫花海、少量花瓣、中央负空间；禁止任何人物、头像、UI、文字和水印。检查成图后裁切到约 `2.5:1`，转换为无损 WebP，保存为 `profile-identity-sunset-v2.webp`。

- [ ] **Step 4: 在页面中显式导入背景层**

在 `index.tsx` 中加入：

```tsx
import profileIdentityBackground from "../../assets/profile/profile-identity-sunset-v2.webp";

<Image
  className="profile-space-identity-background"
  src={profileIdentityBackground}
  mode="aspectFill"
/>
```

背景 Image 放在粒子层和真实内容层之前。

- [ ] **Step 5: 运行测试并确认通过**

Run: `npm test -- --run src/pages/profile/index.test.tsx`

Expected: 该文件全部测试通过。

- [ ] **Step 6: 提交素材与页面结构**

```bash
git add miniprogram/src/assets/profile/profile-identity-sunset-v2.webp miniprogram/src/pages/profile/index.tsx miniprogram/src/pages/profile/index.test.tsx
git commit -m "feat(miniprogram): add profile sunset card background"
```

### Task 2: 实现多层玻璃高光

**Files:**
- Modify: `miniprogram/src/pages/profile/parity.scss`
- Test: `miniprogram/src/ui/liquidGlassPages.test.ts`

**Interfaces:**
- Consumes: Task 1 的 `.profile-space-identity-background`。
- Produces: 背景层、玻璃染色层、连续渐变亮边、顶部反光和底部辉光。

- [ ] **Step 1: 写入失败样式测试**

在 `liquidGlassPages.test.ts` 中加入：

```ts
expect(style).toContain('.profile-space-identity-background');
expect(style).toContain('.profile-space-identity-glass');
expect(style).toContain('linear-gradient(115deg');
expect(style).toContain('inset 0 1PX 0 rgba(255,255,255');
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts`

Expected: FAIL，因为新背景层和玻璃覆盖层尚未定义。

- [ ] **Step 3: 加入玻璃覆盖层结构**

在身份卡背景 Image 之后加入：

```tsx
<View className="profile-space-identity-glass" ariaHidden />
```

- [ ] **Step 4: 实现背景与玻璃样式**

在 `parity.scss` 中：

```scss
.profile-space-identity-background { position:absolute; z-index:0; inset:0; width:100%; height:100%; }
.profile-space-identity-glass { position:absolute; z-index:1; inset:0; background:linear-gradient(115deg, rgba(116,91,188,.42), rgba(38,70,130,.28) 55%, rgba(227,102,217,.2)); }
```

将卡片边缘改为粉紫到冰蓝的渐变边框表现；使用伪元素加入顶部弧形反光和底部粉紫辉光。内容层保持在 `z-index: 4`，花瓣层保持在内容下方。减少动态模式下提高玻璃染色层不透明度并停止动画。

- [ ] **Step 5: 运行聚焦测试与 WXSS 检查**

Run: `npm test -- --run src/ui/liquidGlassPages.test.ts src/ui/wxssCompatibility.test.ts src/pages/profile/index.test.tsx`

Expected: 全部通过。

- [ ] **Step 6: 提交玻璃样式**

```bash
git add miniprogram/src/pages/profile/index.tsx miniprogram/src/pages/profile/parity.scss miniprogram/src/ui/liquidGlassPages.test.ts
git commit -m "style(miniprogram): deepen profile card glass"
```

### Task 3: 微信开发者工具视觉验收

**Files:**
- Create: `miniprogram/.qa/profile-glass-card-phone.png`
- Modify: `miniprogram/.qa/profile-space-design-qa.md`

**Interfaces:**
- Consumes: Task 1 的背景素材和 Task 2 的玻璃层。
- Produces: 最终手机截图与可复查的验收记录。

- [ ] **Step 1: 完整测试与开发构建**

Run: `npm test`

Expected: 0 个失败。

Run: `npm run build:weapp:sandbox && node scripts/sanitize-weapp-wxss.mjs`

Expected: 编译成功，WXSS 兼容处理成功。

- [ ] **Step 2: 在微信开发者工具重新编译并进入“我的”**

确认页面没有黑屏，底部“我的”选中，真实头像和真实任务数字已经返回。

- [ ] **Step 3: 截图并逐项核对**

保存 `profile-glass-card-phone.png`，核对：不再出现青色几何条纹；可见紫蓝晚霞和底部花海；粉紫/冰蓝边缘连续；顶部反光和底部辉光可辨认；头像、文字、数字没有被遮挡；卡片大小没有增加。

- [ ] **Step 4: 更新验收记录**

在 `profile-space-design-qa.md` 中记录最终截图路径、当前真实任务值、动画与降低透明度行为、测试和构建结果。

- [ ] **Step 5: 提交验收证据**

```bash
git add miniprogram/.qa/profile-glass-card-phone.png miniprogram/.qa/profile-space-design-qa.md
git commit -m "test(miniprogram): verify profile glass card"
```
