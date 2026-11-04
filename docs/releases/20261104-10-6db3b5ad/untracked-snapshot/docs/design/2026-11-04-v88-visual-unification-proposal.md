# V88 全站视觉统一提案（2026-11-04）

> 本轮只交付规范草案与盘点，不落地换肤代码。代码层面只做了「顶栏重复标题全站收口 + 假状态文案修复」。
> 相关代码改动证据：`frontend/src/shared/layouts/UserLayout.jsx`、`frontend/src/shared/notifications/globalTaskCenter.js`、`frontend/src/user/pages/NovelFetchPage.jsx`。
>
> **2026-11-04 决策更新（以代码为准）**：主色改为青绿 `#63dfc0`，珊瑚 `#f07167` 降为强调/警示，靛蓝 `#9da8ff` 只做信息/链接。实现见 `frontend/src/shared/styles/v88-tokens.css`，本文第三章的 `--v88-color-primary: #f07167` 已作废。

## 一、六个一级页面外壳盘点

| 页面 | 容器宽度策略 | 标题出现次数与位置 | 顶栏状态文案来源 | 空态实现 | 与目标差异 |
| --- | --- | --- | --- | --- | --- |
| 首页 `/` | 全宽；`.home-page` 左右 `24px`；hero 正文 `width:min(780px,100%)`、标题 `max-width:820px` | 顶栏 0（首页走独立 `home-shell`，不渲染顶栏）；内容 1（hero `<h1>` 营销标题）；hero 内另有 7 个导航链接 | 无顶栏，无状态文案 | 无空态 | 营销页允许保留独立视觉；不纳入工作台容器统一 |
| 私人定制 `/script` | 全宽；`.script-workbench-form` 左右 `18px`；网格 `326px + 8px + 1fr`，`max-width:none` | 顶栏 0（已在上一轮对 `/script` 收口）；内容 1（左栏品牌行「私人定制」）；侧栏 1（导航高亮） | 仅全局任务/接口错误事件；无业务页常量文案 | `.script-empty` + `.script-empty-card`；虚线边框，圆角 `8px` | 容器已是目标形态；素材编辑统一、按钮体系仍待统一 |
| 小说获取 `/novel-fetch` | 全宽；`.novel-fetch-workbench` `width:100%`、`max-width:none`；外层左右 `24px` + 面板 `18px`，实际约 `42px` 留白；内层左右两栏 | 顶栏 1（本轮移除）；内容 0（过程页无页面级标题）；侧栏 1（导航高亮） | 原为 `processing` 本地标志派生的「正在处理任务。／暂无正在处理的任务。」，**与右侧真实队列「排队 1」矛盾**；本轮改为真实队列计数 | 任务列表用 AntD `Empty`；面板为空时无统一空态组件 | 左右留白 `42px`，比私人定制 `18px` 明显更窄；内容区缺页面标题 |
| 水货生产 `/batch-factory` | 全宽；`.shuihuo-project-library` `padding:28px 24px 64px`，无 `max-width`；项目网格 `auto-fill minmax(220px,266px)` | 顶栏 1（本轮移除）；内容 1（`<h1>`「漫剧解说」，与导航名「水货生产」不一致）；侧栏 1 | 全局任务/接口错误事件 | `.shuihuo-empty` / `.shuihuo-workbench-empty`：虚线边框、`min-height:180px`、无圆角、`grid place-items:center` | 页面标题命名与导航不一致；空态视觉与私人定制不同 |
| Agent 工作区 `/agent` | 全宽；`.agent-page.utility-workbench` `width:100%;height:100%;padding:0`，单列网格；头部 `padding:22px 24px 16px` | 顶栏 1（本轮移除）；内容 1（`<h3>`「CM Agent」或当前任务标题，与导航名「Agent 工作区」不一致）；侧栏 1 | 全局任务/接口错误事件 | `.agent-task-empty` / `.agent-workbench-empty`：虚线 `7px` 圆角 / 纯居中文字 | 页面标题命名与导航不一致；空态两套写法 |
| 配音 `/tts` | 全宽；`.tts-workbench.utility-workbench`；工具栏 `padding:12px 20px`；卡片网格 `padding:20px`、`auto-fill minmax(360px,1fr)` | 顶栏 1（本轮移除）；内容 0（工具栏只有控件，没有页面标题）；侧栏 1 | 全局任务/接口错误事件 | `.tts-empty`：虚线 `8px` 圆角、`min-height:260px`、居中图标 + 两行文案 | 移除顶栏标题后内容区无页面标题；等待动画与按钮形态未统一 |

### 统一规则（可直接逐条对照上表）

1. **顶栏只承担两件事**：全局导航（实际在左侧栏）+ 真实全局状态。顶栏不再渲染页面标题；没有真实状态时整条顶栏不渲染。
2. **页面标题只在内容区出现一次**，命名与侧栏导航一致（如「水货生产」「Agent 工作区」「配音」）；不得在顶栏、面包屑、内容区重复三次。
3. **状态文案必须来自真实数据**。禁止用页面本地表单标志派生「正在处理 / 暂无任务」这类全局口径文案；无真实状态时显示空而不是占位句。
4. **工作台容器统一为全宽**：外层左右 gutter 统一 `24px`，内部卡片内边距统一 `18px`；不再出现 `1100px / 1180px` 这类把工作台锁在中间的上限。
5. **空态统一一套组件**：虚线边框、`8px` 圆角、`min-height:180px`、图标 + 主文案 + 副文案；不得一页 `Empty`、一页纯文字、一页无圆角。
6. **按钮统一一套令牌与变体**（见第三节），不允许同一动作在不同页出现 `999px` 胶囊、`11px`、`6px`、`14px` 四种圆角。
7. **等待态统一分级**（见第四节）：超过 `300ms` 出现，超过 `2s` 升级为进度或骨架。
8. **动效统一时长与缓动**（见第五节），并尊重 `prefers-reduced-motion`。

## 二、现状盘点（具体值）

### 2.1 色系来源

| 来源文件 | 背景 | 卡片 / 面板 | 边框 | 主色 | 成功 / 警告 / 错误 | 文字 |
| --- | --- | --- | --- | --- | --- | --- |
| `frontend/src/shared/styles/global.css` | `#070b11` | `#0e1621` | `rgba(151,168,190,.22)` | `#f07167` | `#75ddb4` / `#ffd08a` / `#ff9b9d` | `#f5f7fb` / `#c7d0dd` / `#a7b4c5` |
| `frontend/src/user/pages/shuihuo-production.css` | `#070b11` | `#0e1621` / `#111b28` | `rgba(151,168,190,.16)` | `#63dfc0`（青绿） | `#f08a8a`（仅错误） | `#f5f7fb` / `#8391a4` |
| `frontend/src/user/pages/private-studio/private-studio-workbench.css` | `#0a101d` | `#101827` | `rgba(157,176,218,.17)` | `#9da8ff`（靛蓝）+ `#65dfce`（薄荷） | 无独立令牌 | `#edf2ff` / `#8996b1` / `#65738e` |
| `frontend/src/shared/styles/agent-workspace.css` | 复用 `--surface-base` | 复用 `--surface-panel` | 复用 `--border-default` | `#f07167`（珊瑚） | `#5dba8a` | 复用三级文字 |
| `frontend/src/user/pages/novel-fetch.css` | 复用 `--legacy-*` | 复用 `--legacy-card` | 复用 `--legacy-border` | 复用 `--legacy-accent` | 复用全局 | 复用全局 |
| 首页 `home.css/global.css` | 视频 hero + `--legacy-*` | `--legacy-card` | `--legacy-border` | `#f5976c`（橙） | 复用全局 | `#fff` / `rgba(232,232,240,.78)` |

**结论**：主色至少有 4 套（珊瑚 `#f07167`、青绿 `#63dfc0`、靛蓝 `#9da8ff`、橙 `#f5976c`），背景至少 3 档（`#070b11` / `#0a101d` / `#0e1621`），同页内也常同时出现两套（私人定制 = 靛蓝 + 薄荷）。

### 2.2 按钮现状

- AntD `Button`：675 处；其中 `type="primary"` 161、`type="text"` 95、`type="link"` 36、带 `danger` 54。
- 原生 `<button>`：246 处（大量分布在自定义胶囊、列表行、工具栏）。
- 自定义按钮 class：32 处（`private-studio-primary`、`shuihuo-create-project`、`GradientButton` 等）。
- 圆角实测四种并存：`GradientButton` `999px` 胶囊、私人定制主按钮 `11px`、AntD 默认 `6px`、项目卡/入口卡 `14px`。
- 危险操作：AntD `danger` 有 54 处，但水货生产与私人定制大量使用自定义 `--sh-danger` / `--ps-*`，没有统一红色令牌。
- 四态：`loading` 属性 313 处；仍存在只有文字变化、没有 disabled 防重复提交的原生按钮。

### 2.3 等待态现状

| 等待点 | 现状 | 证据 |
| --- | --- | --- |
| 进入作品 | **没有 loading**：`openProject` 只 `await getProject()`，期间无遮罩/骨架/按钮态 | `ShuihuoProductionPage.jsx` `openProject` |
| 项目库首次加载 | 有 `<Spin>` | `ShuihuoProductionPage.jsx` `shuihuo-loading` |
| 批量工程首页 | 有 `<Spin>` + 文案 | `BatchFactoryWorkbenchPage.jsx` |
| 出图（资产图 / 漫画页） | 只有 `message.loading` 气泡，没有页面级进度 | `BatchFactoryNovelList.jsx`、`AssetsView.jsx` |
| 提取素材 | 有按钮 `loading` + 禁用 | `ScriptPage.jsx` `private-studio-extract` |
| 同步（批量后台配置 / 执行器） | 只有 `message.loading` 或轮询状态文字 | `BatchFactoryNovelList.jsx` |
| 保存（草稿 / 配置 / 设置） | 草稿有真实三态；配置页多为按钮 `loading` | `ScriptPage.jsx` `draftSaveState`；`SettingsPage.jsx` |
| 发布 | 无统一 `发布中` 按钮态，依赖任务状态轮询 | `agent-comic-deliverables.js` 只有状态标签 `发布中` |
| 上传 | 15 处 `<Upload>`；部分按钮上传只有 `uploading` 布尔，无进度 | `uploading` 变量 13 处 |

统计：`<Spin>` 15 处、`<Skeleton>` 8 处、`<Progress>` 11 处、`message.loading` 9 处、按钮 `loading` 313 处。**同一个等待动作在不同页可能分别用气泡、按钮转圈、文字或什么都不用。**

### 2.4 动效现状

- `transition` 声明约 218 个时间值，主流是 `120ms / 140ms / 160ms / 180ms / 200ms / 220ms ease`，但混用 `s` 与 `ms`。
- 存在 `0.35s`、`9999s` 等例外；同一元素常同时 transition 4 个属性，但曲线不统一（`ease`、`ease-out` 混用）。
- `@keyframes` 31 个；8 个 CSS 文件已有 `prefers-reduced-motion` 分支，其余动画（进度条、宠物、渐变）没有统一降级策略。

## 三、令牌草案（可直接落地，做映射而不是推倒重来）

命名前缀 `--v88-*`，先在 `:root` 定义，再让现有变量指向它：

```css
:root {
  /* 颜色：暗色基准 */
  --v88-color-bg: #070b11;
  --v88-color-surface: #0e1621;
  --v88-color-surface-raised: #111b28;
  --v88-color-border: rgba(151, 168, 190, .22);
  --v88-color-border-strong: rgba(176, 195, 219, .36);
  --v88-color-primary: #f07167;
  --v88-color-primary-hover: #f08a7d;
  --v88-color-success: #75ddb4;
  --v88-color-warning: #ffd08a;
  --v88-color-danger: #ff9b9d;
  --v88-color-text: #f5f7fb;
  --v88-color-text-secondary: #c7d0dd;
  --v88-color-text-tertiary: #a7b4c5;

  /* 圆角与间距 */
  --v88-radius-sm: 6px;
  --v88-radius-md: 8px;
  --v88-radius-lg: 12px;
  --v88-radius-xl: 18px;
  --v88-space-page: 24px;
  --v88-space-card: 18px;
  --v88-space-gap: 16px;

  /* 字号与阴影 */
  --v88-font-page-title: 24px;
  --v88-font-section-title: 16px;
  --v88-font-body: 13px;
  --v88-font-caption: 12px;
  --v88-shadow-card: 0 8px 32px rgba(0, 0, 0, .22);

  /* 动效 */
  --v88-motion-fast: 120ms;
  --v88-motion-base: 160ms;
  --v88-motion-slow: 200ms;
  --v88-ease-standard: cubic-bezier(.2, 0, .2, 1);
}
```

**共存映射（不推倒重来）**：

- `global.css`：`--legacy-bg → var(--v88-color-bg)`，`--legacy-panel/--legacy-card → var(--v88-color-surface)`，`--legacy-border → var(--v88-color-border)`，`--legacy-accent → var(--v88-color-primary)`。
- `shuihuo-production.css`：`--sh-bg → var(--v88-color-bg)`，`--sh-panel → var(--v88-color-surface)`，`--sh-accent → var(--v88-color-primary)`。青绿作为“生产态强调色”可以保留，但只用于状态点，不用于主按钮。
- `private-studio-workbench.css`：`--ps-bg → var(--v88-color-bg)`，`--ps-panel → var(--v88-color-surface)`，`--ps-line → var(--v88-color-border)`，`--ps-indigo/--ps-mint` 降级为局部装饰色，不再承担主按钮色。
- `agent-workspace.css` 已经把 `--agent-*` 指向 `--surface-*`，只需把 `--surface-*` 指到 `--v88-*`。
- 不动 `.private-studio-*` 类名与结构，只替换变量值，保证另一会话的布局测试不被破坏。

## 四、按钮体系

五种变体 × 四种状态：

| 变体 | 默认 | 悬停 | 禁用 | 加载 |
| --- | --- | --- | --- | --- |
| Primary | 主色底、白字、`radius-md` | 主色 hover、上浮 `translateY(-1px)` | 主色 30% 透明、文字 50% | 左侧 `Spin`，宽度锁定，禁用点击 |
| Secondary | 透明底、边框 `border`、正文色 | 边框加强、表面色 | 边框 50%、文字 50% | `Spin` + 文案，禁用 |
| Ghost | 无底无边、次级文字 | 表面色底、文字提亮 | 文字 40% | `Spin` 替换图标 |
| Danger | 危险色底或危险边框 | 危险色 hover | 危险色 30% | `Spin`，禁用 |
| Icon | 方形/圆形、`radius-md` | 表面色底 | 40% 透明 | `Spin` 居中，尺寸不跳 |

硬规则：

1. 加载态必须出现转圈，且 `disabled`，禁止重复提交。
2. 同一动作全站只能有一个变体（保存 = Primary，取消 = Secondary，删除 = Danger，工具栏 = Ghost/Icon）。
3. 胶囊 `999px` 只允许出现在首页营销 `GradientButton`，工作台内禁止。

## 五、等待动画规范

| 时长 | 表现 | 场景 |
| --- | --- | --- |
| `< 300ms` | 不显示加载态，避免闪烁 | 本地状态切换、轻量接口 |
| `300ms–2s` | 按钮转圈或 `Spin` | 保存、提取、单页出图 |
| `> 2s` | 进度条 / 骨架屏 / 阶段文案 | 进作品、批量出图、整本漫画、批量发布 |
| 无法估算 | 骨架屏 + 阶段文案 | 项目工作台首屏、列表首屏 |

必须补齐的等待点：**进入作品**（本轮盘点发现完全无 loading）、发布、同步。

## 六、动效规范

- 时长：`120ms`（颜色/边框）、`160ms`（位移/透明度）、`200ms`（展开/收起）；禁止超过 `250ms` 的常规交互。
- 缓动：统一 `cubic-bezier(.2, 0, .2, 1)`；进入用 `ease-out`，退出用 `ease-in` 的旧写法逐步替换。
- 允许动的元素：按钮、卡片 hover、抽屉/弹窗、进度条、状态点。
- 禁止动：正文排版、表格行高、输入框尺寸（避免布局抖动）。
- `prefers-reduced-motion: reduce` 时：`animation: none`、`transition-duration: 1ms`，保留最终态。

## 七、分批替换计划

| 批次 | 页面 | 改动点 | 回滚方式 | 验收标准 |
| --- | --- | --- | --- | --- |
| 0（本轮） | 全站外壳 | 顶栏标题收口、假状态改真实队列 | 单文件 `UserLayout.jsx` / `NovelFetchPage.jsx` 回退 | 顶栏无页面标题；无状态时不渲染；队列文案与右侧一致 |
| 1（样板页） | **配音 `/tts`** | 令牌映射、按钮五变体、等待态、空态、动效 | 新增 `v88-tokens.css` + 页面局部覆盖，可整文件回退 | 1440 / 1280 下无横向滚动；按钮四态一致；空态统一 |
| 2 | 小说获取 `/novel-fetch` | gutter `42px → 24+18` 统一、补内容区标题、等待态 | 页面 CSS + 标题节点，可单独回退 | 与样板页同 gutter、同按钮、同空态 |
| 3 | 水货生产 `/batch-factory` | 标题改名「水货生产」、空态统一、进作品 loading | 页面 CSS + 标题节点 | 标题与导航一致；进作品有骨架 |
| 4 | Agent `/agent` | 标题改名「Agent 工作区」、空态统一、按钮统一 | 页面 CSS + 标题节点 | 命名一致；两套空态合并 |
| 5 | 私人定制 `/script` | 只替换变量值，不改结构 | 只回退变量映射 | 现有布局测试全过；视觉与样板页一致 |
| 6 | 首页 `/` | 仅统一按钮与动效，保留营销视觉 | 首页 CSS 回退 | 营销感保留，交互时长统一 |

**为什么第一批选配音**：它是六页里结构最简单、没有私人定制会话占用、没有复杂业务弹窗的页面；改动能完整覆盖「容器 / 按钮 / 空态 / 等待 / 动效」五类规范，回滚面最小，适合当样板给用户看。

## 八、风险

- 涉及文件：全站样式约 20+ 个 CSS 文件，按钮/空态散落在 100+ 个 JSX 文件；第一批务必限制在令牌 + 单页覆盖。
- 自定义样式覆盖：`shuihuo-production.css` 使用独立 `--sh-*` 且带亮色主题；`private-studio-workbench.css` 被另一会话占用，只能改变量。
- 按钮被挤没：统一高度/内边距后，工具栏密集区（小说获取、水货生产）可能出现横向溢出；验收必须加 1280 / 1024 断点检查。
- `global.css` 是全站公共文件，任何改动都要逐条说明影响面；第一批不要直接改它，先新增 `v88-tokens.css` 承载令牌。
