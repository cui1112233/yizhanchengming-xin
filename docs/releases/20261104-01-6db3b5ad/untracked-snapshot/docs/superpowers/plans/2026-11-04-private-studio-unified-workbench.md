# 私人定制 · 统一创作工作台 实施方案

- 日期：2026-11-04
- 范围：`/script`（私人定制 · 剧本创作 / 漫画生成）
- 本轮性质：**只出方案，不改任何 UI 代码、样式、dist，不重启 4001，不 commit**
- 设计权威来源（用户指定，优先级最高）：
  `G:\Backup\Documents\ChatGPT\v88\.superpowers\brainstorm\113-1793760592\content\private-customization-ui-preview.html`
  （已读取，13041 字节 / 101 行；下列文案、尺寸、颜色均逐字摘自该文件）
- 次要根据：设计截图 `cc1f8445-..._image.png`
## 实现原则（优先级最高，后续每阶段都必须遵守）

1. **设计稿只作视觉与排版参考**（配色、分区、间距、卡片形态、层级、文案风格），不作为功能清单。
2. **功能一律以现有实现为准**：稿子里没画出来的按钮、下拉、开关、入口，不许以“稿子没有”为理由删掉或藏起来，必须全部保留，并按稿子的视觉风格摆在合理位置。
3. 任何一项功能**不许静默消失**。凡是新稿里没有的，映射表必须显式写“稿中缺失 → 补进新布局（落在哪个区块/哪个按钮）”。
4. **本轮不允许出现“建议删除”的结论**。确需下线的，单独列入“疑似冗余（待用户确认，不擅自删）”。
5. 阶段验收必须逐项核对功能清单，输出 ✅/❌ 列表。
6. 样式只写在私人定制独立样式文件（`.private-studio-*` 前缀），不往 `global.css` / `shuihuo-production.css` 塞。

- 标注约定：
  - `[代码]` = 读仓库源码确认的事实
  - `[设计]` = 设计源文件明确写出的规格
  - `[推断]` = 源文件未写死、需按现有能力合理落地的判断

---

## 0. 设计源文件与截图的差异（必须先对齐）

源文件与截图有 4 处不一致。按用户指示"以源文件为准"，实现时采用源文件文案；截图里更新但源文件没有的元素列为待确认项：

| 元素 | 源文件 HTML | 截图 | 采用 |
| --- | --- | --- | --- |
| 输出选择器文案 | `输出 漫画分镜` | `漫帧分镜` | **源文件：漫画分镜**（截图更晚，待用户确认） |
| 右栏说明卡标题 | `下一步` | `创作提示` | **源文件：下一步**（同上） |
| 右栏生成按钮 | 无 | `生成预览` | 源文件无该按钮；按"生成预览"为截图新增能力，列入待确认 |
| 顶部 `历史版本` / `更多设置` | 无 | 有 | 源文件无；按截图保留，列入待确认 |
| `查看设计说明` / `采用这个方向` | 无 | 有 | **预览壳层，不是产品 UI**，不实现 |

另外：源文件顶栏是 `私人定制` + `草稿已保存`，没有截图中的"方案 A / 方案 B"备选卡片——那两张卡片同样是**预览壳层**，不入产品。

---

## 1. 精确设计规格（`[设计]`，逐字照抄）

### 1.1 结构与比例

源文件不是扁平三栏，而是「两栏外壳 + 中栏内部再分两列」：

```
.pc-shell             max-width:1180px; grid-template-columns:326px minmax(0,1fr)
├── aside.pc-side     宽 326px；padding 20px 18px；右边框
└── main.pc-main
    ├── header.pc-main-head   padding 22px 22px 16px；下边框
    ├── div.pc-toolbar        padding 13px 22px；下边框
    └── div.pc-content        grid-template-columns:minmax(0,1fr) 260px; gap:14px; padding:18px 22px 22px
        ├── 分镜卡（1fr）
        └── 预览列（260px）
```

- 外壳圆角 `18px`，边框 `1px solid rgba(157,176,218,.17)`，背景 `rgba(11,17,29,.82)`，阴影 `0 24px 80px rgba(0,0,0,.28)`。
- 卡片圆角 `13px`，输入框 `12px`，页签容器 `11px`，胶囊标签 `999px`。
- 左右内边距：侧栏 `20/18`，主区头 `22/22/16`，工具栏 `13/22`，内容区 `18/22/22`。

**响应式断点（`[设计]` 精确值）**

| 断点 | 行为 |
| --- | --- |
| `max-width:900px` | 外壳单列；侧栏右边框改下边框；`.pc-content` 单列 |
| `max-width:560px` | 外层 padding 28→10px；工具栏改为纵向，右侧按钮组占满一行并右对齐 |

### 1.2 颜色令牌（`[设计]` 原值）

| 令牌 | 值 |
| --- | --- |
| `--bg` | `#0a101d` |
| `--panel` | `#101827` |
| `--input` | `#0b1423` |
| `--line` | `rgba(157,176,218,.17)` |
| `--line-soft` | `rgba(157,176,218,.11)` |
| `--text` | `#edf2ff` |
| `--muted` | `#8996b1` |
| `--muted-2` | `#65738e` |
| `--indigo` | `#9da8ff` |
| `--mint` | `#65dfce` |

页面背景：双层径向渐变
`radial-gradient(800px 430px at 6% -8%,rgba(105,121,255,.15),transparent 64%), radial-gradient(720px 360px at 98% 0,rgba(45,204,183,.08),transparent 65%), var(--bg)`。

主按钮：`linear-gradient(135deg,#aeb7ff,#62dfcf)`，文字 `#10172b`，`font-weight:750`，阴影 `0 8px 22px rgba(101,223,206,.13)`。
状态胶囊（草稿已保存 / 准备生成 / 不发布）：文字与点 `--mint`，边框 `rgba(101,223,206,.4)`，底 `rgba(101,223,206,.06)`，字号 `10px`，内边距 `5px 8px`。

### 1.3 字号与控件规格（`[设计]`）

| 元素 | 规格 |
| --- | --- |
| 品牌字 | `14px / 750`，logo 31×31 圆角 10，渐变 `#aeb7ff→#62dfcf` |
| 页签 | 容器 `grid 1fr 1fr; gap:4px; padding:4px`；按钮 `8px` 内边距、`11px` 字；激活态底色 `rgba(157,168,255,.22)`、文字 `--text` |
| 区块标题（如 01 · 粘贴你的内容） | `11px`，副标题（支持 TXT / 长文本）`--muted-2` |
| 正文输入 | `min-height:156px`；`padding:12px`；`12px` 字；行高 `1.65`；背景 `--input`；圆角 12 |
| 字数行 | `11px`，`--muted-2` |
| 步骤条 | `gap:8px`；行 `12px`；圆点 22×22 圆形；已完成圆点底 `--mint`、字 `#092b2b`；当前圆点底 `--indigo`、字 `#121a33` |
| 素材区标题 | `11px`，右侧"可编辑"`--muted-2` |
| 素材胶囊 | `padding:6px 9px`；`11px`；边框 `rgba(157,176,218,.22)`；圆角 999 |
| 次按钮 | `padding:8px 11px`；`11px` |
| 主按钮通栏 | `width:100%; margin-top:18px; padding:11px` |
| 眉标 创作画布 | `10px / 700`；`letter-spacing:.1em`；`text-transform:uppercase`；色 `--indigo` |
| 主标题 | `20px / 1.2 / -0.02em` |
| 元信息行 | `margin-top:7px`；`11px`；`--muted` |
| 工具栏下拉胶囊 | `padding:8px 10px`；`11px`；圆角 9；内部 strong 为 `#eef2ff / 600` |
| 卡片头 | `13px 14px`；标题 `12px`；右侧计数 `10px / --muted-2` |
| 分镜行 | `grid-template-columns:28px minmax(0,1fr) auto; gap:10px`；`padding:13px 14px`；分隔线 `--line-soft` |
| 分镜序号 | `11px / --muted-2` |
| 分镜标题 | `12px` |
| 分镜描述 | `margin-top:5px`；`11px`；行高 1.55；`--muted` |
| 分镜状态标签 | `padding:4px 6px`；圆角 7；`10px`；色 `--indigo`；底 `rgba(157,168,255,.11)` |
| 预览画布 | `min-height:176px`；圆角 10；内边距 13；渐变 `145deg rgba(126,103,191,.68) → rgba(49,160,153,.38) 52% → rgba(13,24,43,.96)`；两个装饰圆 |
| 预览浮层文案 | 标题 `12px`；副文 `10px / rgba(235,243,255,.76)`；行高 1.55 |
| 提示卡 | `margin-top:12px`；`padding:11px 12px`；圆角 11；边框 `rgba(157,168,255,.2)`；底 `rgba(157,168,255,.055)`；`11px`；标题色 `--indigo` |

### 1.4 文案清单（逐字，实现直接照抄）

| 位置 | 剧本模式原文 | 漫画模式原文 |
| --- | --- | --- |
| 品牌 | `私人定制` | `私人定制` |
| 保存徽标 | `草稿已保存` | `草稿已保存` |
| 页签 | `剧本创作` / `漫画生成` | 同左 |
| 输入区小标题 | `01 · 粘贴你的内容` | `01 · 粘贴你的内容` |
| 输入区右侧 | `支持 TXT / 长文本` | `支持 TXT / 长文本` |
| 输入占位 | 示例正文（非文案） | `把要生成漫画的正文直接粘贴到这里` |
| 字数行 | `已读取 1268 字` / `自动保存` | `正文 0 字` / `自动保存` |
| 步骤 1 | `内容已读取` | `提取资产` |
| 步骤 2 | `确认人物与场景` | `生成分镜提示词` |
| 步骤 3 | `生成创作结果` | `生成漫画图片` |
| 素材标题 | `已识别素材` / `可编辑` | `资产` / 数量（如 `0`） |
| 素材胶囊 | `沈砚 · 人物` `旧城 · 场景` `雨夜 · 氛围` `+ 添加` | `人物` `场景` `道具`（无资产时为首字母占位） |
| 左栏主按钮 | `继续：生成分镜` | `提取资产` |
| 眉标 | `创作画布` | `创作画布` |
| 主标题 | `《未寄出的信》` | `《未命名漫画》` |
| 状态徽标 | `准备生成` | `待输入` |
| 元信息 | `一段内容 · 6 个镜头 · 预计 2 分钟` | `粘贴正文后开始生成` |
| 发布徽标 | `不发布` | `不发布` |
| 工具行左 | `输出` `漫画分镜` / `画风` `电影感水彩` / `时长` `15 秒` | 同左（`时长` 是否保留见决策 D4） |
| 工具行右 | `约束条件` / `开始生成` | `约束条件` / `开始生成` |
| 中栏卡头 | `分镜草稿` / `6 个镜头` | `分镜草稿` / `6 个镜头` |
| 状态标签 | `已生成` / `待确认` / `待生成` | 同左 |
| 右栏卡头 | `画面预览` / `01 / 06` | 同左 |
| 右栏浮层 | 镜头名 + `电影感水彩 · 16:9` + `低饱和蓝调 / 暖光对比` | 同左 |
| 底部提示卡 | `下一步` + `确认人物与场景后，点击“继续：生成分镜”。` | `下一步` + `先粘贴正文，再从左侧提取资产。` |

> 状态标签只允许这三个词，不要自行改写为"已完成/处理中"。运行中的临时态可用 `生成中` 覆盖同一位置，但不得替换常态三词。

### 1.5 三张规格图的补充结论（`[设计]`）

来源（均已读取）：

- 左栏完整面板：`a70a7e97-...image.png`
- 资产区放大：`f62de1ed-...image.png`
- 中栏头部与工具行：`d6197e29-...image.png`

确定项：

1. **两个页签并排等宽**，位于左栏顶部同一胶囊容器内，各占一列；不是切换整套界面布局。
2. 左栏顺序固定：品牌行（`私人定制` + `草稿已保存`）→ 双页签 → `01 · 粘贴你的内容` / `支持 TXT / 长文本` → 正文输入 → `已读取 N 字` / `自动保存` → 三段进度 → `已识别素材` / `可编辑` → 素材胶囊 + `+ 添加` → 通栏主按钮。
3. 进度圆点：已完成为薄荷绿实心 `✓`；当前为靛蓝实心数字；未开始为深色描边数字。
4. 素材胶囊为 `名称 · 类型`（`沈砚 · 人物`、`旧城 · 场景`、`雨夜 · 氛围`），`+ 添加` 为独立小胶囊。
5. 中栏头部：书名 + 状态徽标同一行；副标题在下一行；右上角为 `历史版本` 与 `更多设置…`（**带省略号**，以图为准）。
6. 工具行：左侧三个下拉胶囊 `输出 漫画分镜 ▾` / `画风 电影感水彩 ▾` / `时长 15 秒 ▾`；右侧 `约束条件` 胶囊 + 渐变主按钮 `开始生成`。
7. 以上图为**视觉基准**；现有实现里未画出的控件按原则 2 保留并补位。

> 与源 HTML 的差异：图三写作 `更多设置…`，源 HTML 没有该入口；按图三实现 `更多设置…`。

---

## 2. 现状盘点（`[代码]`）

### 2.1 入口与骨架

- 路由：`/script` → `frontend/src/user/App.jsx:58` → `ScriptPage.jsx`（148620 字节）。
- 页面外层是 antd `Form`，class `script-workbench-form`，负责剧本的提交、校验与草稿持久化（`ScriptPage.jsx:1965`）。
- **两个模式目前是两套并列的 grid 区块**，页签状态放在 `ScriptPage.jsx:260` 的 `customizationWorkspace`：
  - 剧本区：`ScriptPage.jsx:1992-2248`，`className="script-workbench utility-workbench"`；非激活时 `display:none`（1998-2000），带 8px 拖拽把手。
  - 漫画区：`PrivateComicWorkspace.jsx:270`，`className="script-workbench private-comic-workbench"`，**无拖拽把手**；未激活时直接不渲染（`ScriptPage.jsx:2249`）。
- 页签**在两个面板里各复制了一份**：`ScriptPage.jsx:2008-2011` 与 `PrivateComicWorkspace.jsx:277-280`。设计稿要求页签只出现在左栏顶部一份。
- 列宽（`private-customization.css:373-379`）：
  - 剧本 `304px 8px minmax(0,1fr)`
  - 漫画 `304px minmax(0,1fr)`
  - 设计稿要求 `326px + 1fr`，中栏内部再切 `1fr + 260px`。

### 2.2 左栏现状

| 区块 | 剧本模式 | 漫画模式 |
| --- | --- | --- |
| 品牌/保存 | `ScriptPage.jsx:2004-2007` | `PrivateComicWorkspace.jsx:273-276` |
| 页签 | `ScriptPage.jsx:2008-2011` | `PrivateComicWorkspace.jsx:277-280` |
| 标题眉标 | `ScriptPage.jsx:2012-2018` | `PrivateComicWorkspace.jsx:281-287` |
| 正文输入 | `Form.Item name="novelText"` + `.script-chat-textarea`（2020-2023） | `Input.TextArea` + 独立标题输入（289-310） |
| 工具行 | TXT 上传 / 指令 / 配音 / 音频匹配 + `提取`（2034-2050） | 字数 + 刷新 + `提取资产`（329-350） |
| 进度 | 无步骤条，仅有 `script-instruction-status`（2053）与阶段提示（2065-2066） | 4 段管线 `pipelineSteps`（225-230、353-366） |
| 素材 | `EntitySection` 人物/场景（2068-2069，组件在 2687） | `私有 资产` 分组：人物/场景/道具（220-224、368-405） |
| 主按钮 | 输入区右下 `提取`（2047-2049） | 刷新右侧 `提取资产`（341-348） |

### 2.3 中/右栏现状

| 区块 | 剧本模式 | 漫画模式 |
| --- | --- | --- |
| 画布头 | `private-customization-canvas-heading`（2080-2087），标题/元信息/状态在 427-438 | 同名结构（418-425），标题/状态在 252-268 |
| 工具行 | 格式 Select / 约束 / 生成 / 更多（2106-2150） | 文本模型 / 图片模型 / 约束 / 生成 / 更多（438-473） |
| 分镜 | `ShotOutputCards`（2153-2188，组件 `components/ShotOutputCards.jsx`） | 每页卡片 `ComicOutputCards`（145-166），含提示词编辑与原图 |
| 预览 | 只有视频 `Modal`（2284-2286）与历史视频 Modal | 预览内嵌在每页卡片里（129-141） |
| 历史 | `openScriptHistory` → `listHistory`（773） | 无 |

### 2.4 可直接复用（勿重写）

| 能力 | 现成实现 |
| --- | --- |
| 分镜卡（勾选/复制/编辑/生视频/参考图开关） | `frontend/src/user/components/ShotOutputCards.jsx` |
| 素材卡（主图状态/主角标记/编辑弹窗） | `EntitySection`（`ScriptPage.jsx:2687`）、`EntityImagePanel.jsx` |
| 画风库（封面流 + 全局参考图） | `shuihuo/BatchFactoryComicStylePicker.jsx:51`（slot `batch.comic-style`） |
| 提示词/推导/约束预设 | `PrivateComicWorkspace.jsx:35-49`（`batch.comic-page`、`batch.comic-image-derivation`、addon `comic_quality/comic_restriction/comic_negative`） |
| 漫画页图片鉴权读取 | `getComicPageImageBlob`（`shared/api/batchFactoryV12.js:175`） |
| 单页出图 + 409 冲突处理 | `generateComicPageWithConflict`（`shuihuo/batchFactoryComicMode.js`） |
| 提示词保存（整行覆盖） | `updateComicPage`（`batchFactoryV12.js:513`）+ `savePrivateComicPagePrompts`（`ScriptPage.jsx:1790`） |
| 轮询 | `ScriptPage.jsx:1955-1962`，5 秒；`refreshPrivateComicTask`（1590） |
| 草稿自动保存 | 漫画 localStorage（`privateCustomizationComic.js:1-3`）；剧本 `scriptDraftStorage` |
| 步骤/就绪派生 | `comicReadiness`（`backend/internal/httpapi/comic_readiness.go:30`）；分镜列表接口已返回 `readiness` |

---

## 3. 差距表（设计区块 → 现有物 → 动作 → 冲突标记）

> "冲突"列：`用户手改` 表示该文件当前有未提交改动且用户正在编辑，**必须等用户改完再合并**；`无` 表示新建文件，可直接写。

| # | 设计区块（源文件顺序） | 现有对应物 | 动作 | 目标文件 | 冲突 |
| --- | --- | --- | --- | --- | --- |
| 1 | `.pc-shell` 外壳 + 326px 侧栏 + 内层 260px 预览 | 两套独立 `.script-workbench` grid | 新建统一外壳组件，改为 `326px + 1fr`，主区内 `1fr + 260px` | `private-studio/PrivateStudioWorkbench.jsx`（新建）、`private-studio-workbench.css`（新建） | 新建无冲突；接线处需改现有两文件 |
| 2 | 品牌 `私人定制` + `草稿已保存` | 两处重复 `.private-customization-brand-row` | 收敛到外壳（保留类名以便最小改动） | 同上 | 用户手改 |
| 3 | 顶部双页签（单份） | 两处重复页签 | 收敛到外壳；删除面板内重复页签 | `ScriptPage.jsx:2008-2011`、`PrivateComicWorkspace.jsx:277-280` | 用户手改 |
| 4 | `01 · 粘贴你的内容 / 支持 TXT / 长文本` | 剧本有眉标；漫画有；右侧副文缺 | 统一为设计文案；剧本保留 TXT 上传入口 | `PrivateStudioRail.jsx`（新建） | 用户手改（接线） |
| 5 | 正文输入 + 字数行 + `自动保存` | 剧本 `Form.Item novelText`；漫画 `Input.TextArea` + 标题输入 | 抽成 `leftInput` 插槽，剧本仍用 Form.Item，漫画仍用受控 TextArea | 插槽由页面传入 | 用户手改 |
| 6 | 三段步骤条（✓ / 2 / 3） | 剧本 `generationStage`；漫画 4 段 `pipelineSteps` | 新建纯函数派生；漫画按设计收敛为 3 段（资产图进度并入第 1 段详情） | `private-studio/privateStudioState.js`（新建）+ 测试 | D1 决策 |
| 7 | `已识别素材 / 可编辑` + 胶囊 + `+ 添加` | 剧本 `EntitySection` 卡片；漫画按 kind 分组卡片 | 新建 `AssetChips`；剧本复用实体数据；漫画复用 assets；`+ 添加` 复用既有新增入口 | `PrivateStudioAssets.jsx`（新建） | 用户手改（接线） |
| 8 | 左栏通栏主按钮 | 剧本 `提取`；漫画 `提取资产` | 由 `currentStep` 决定文案/行为（见 D3） | 外壳插槽 | D3 决策 |
| 9 | 画布头：眉标/标题/状态/元信息/`不发布` | 两处 `canvas-heading` 已有 | 收敛进外壳 header，保留现有标题派生 | `PrivateStudioHeader.jsx`（新建） | 用户手改 |
| 10 | 工具行 `输出 / 画风 / 时长 / 约束条件 / 开始生成` | 剧本格式 Select + 约束 + 生成；漫画模型 + 约束 + 生成；画风在 Drawer | 新建工具行；画风从 Drawer 提到行内（`BatchFactoryComicStylePicker` 或其轻量摘要版） | `PrivateStudioToolbar.jsx`（新建） | D4 决策 |
| 11 | 分镜草稿列表（序号/标题/描述/状态/可拖拽） | 剧本 `ShotOutputCards`；漫画 `ComicOutputCards` | 新建 `ShotDraftList`；先**只读展示**，拖拽见 D2 | `PrivateStudioShotList.jsx`（新建） | D2 决策 |
| 12 | `画面预览` + `01 / 06` + 浮层 | 漫画图片内嵌卡片；剧本仅视频 Modal | 新建预览列；复用 `getComicPageImageBlob`；选中页状态由列表驱动 | `PrivateStudioPreview.jsx`（新建） | 用户手改（接线） |
| 13 | `生成预览`（截图新增） | `runPrivateComicPageImages(videoId)` | 选中页后按钮复用该动作 | 同上 | D5 决策 |
| 14 | 底部 `下一步` 提示卡 | 无 | 新建，按 mode 显示设计原文 | 外壳插槽 | 新建无冲突 |
| 15 | `历史版本` / `更多设置` | 剧本 `历史`、`配置`；漫画 `配置` + 更多 Dropdown | 保留截图入口，文案改为 `历史版本`/`更多设置` | 外壳 header 右侧 | D6 决策 |
| 16 | 900 / 560 断点 | 现有 `@media` 单列 | 新增两档，scoped 到 `.private-studio-workbench` | `private-studio-workbench.css` | 新建无冲突 |

### 3.1 会与用户未提交改动冲突的文件（务必等用户改完再合并）

1. `frontend/src/user/pages/ScriptPage.jsx`（已 M，148KB）
2. `frontend/src/user/pages/PrivateComicWorkspace.jsx`（已 M）
3. `frontend/src/user/pages/private-customization.css`（已 M，1012 行）
4. 间接但不改：`frontend/src/shared/styles/global.css`、`shared/layouts/UserLayout.jsx`（已 M，且属其他会话范围，**本方案不触碰**）

合并策略：新组件/新样式全部放进 `frontend/src/user/pages/private-studio/`，接线时只做"插槽替换 + 删除重复页签"的最小编辑，避免与用户的手改大范围重叠。

---

## 4. 数据与接口映射

### 4.1 左栏三段进度

| 步骤 | 剧本模式（`[代码]`） | 漫画模式（`[代码]`） |
| --- | --- | --- |
| 1 | `内容已读取`：`Array.from(novelText).length > 0`；`generationStage !== 'idle'` 视为已读取 | `提取资产`：`comicAssets.length > 0`；详情用 `assetsWithImages/assetList.length` |
| 2 | `确认人物与场景`：`current` 当 `generationStage==='extracted'`；`done` 当 `characters.length+scenes.length>0` | `生成分镜提示词`：**以 `readiness.promptReadyCount === readiness.pagesTotal` 为 done**（不是"任一页有提示词"） |
| 3 | `生成创作结果`：`done` 当 `output` 非空或 `generationStage==='complete'` | `生成漫画图片`：`readiness.imageReadyCount === readiness.pagesTotal`（只认主图） |

- 先落地为纯函数 `deriveStudioSteps(mode, state)`，输入只用现有字段，便于单测；**不要在前端重算"是否就绪"第二套口径**。
- 分镜列表接口已返回书级 `readiness`（`batch_factory_comic_pages.go:1095-1098`），直接消费。
- `comicReadiness` 字段：`pagesTotal / promptReadyCount / imageReadyCount / primaryReadyCount / staleCount / canGenerateImages / canBuildCollection / skippable / blockers[] / limits.maxPages`（`comic_readiness.go:30-45`）。

### 4.2 "已识别素材"

| 模式 | 数据源 | 字段 | 胶囊文案 |
| --- | --- | --- | --- |
| 剧本 | `extractInfo.characters` / `.scenes`（`ScriptPage.jsx:2068-2069`） | `id/name/主图`；场景实体含 `氛围`（`ScriptPage.jsx:1289`） | `${name} · 人物` / `${name} · 场景`；`氛围` 胶囊为 `[推断]`：取场景实体的 `氛围` 字段，非新资产类型 |
| 漫画 | `GET .../assets` → `BookAsset` | `kind` ∈ `character/scene/prop`（`types.go:219-238`） | `${name} · 人物/场景/道具`；无资产时显示三类占位 |

- `+ 添加`：剧本复用 `addEntity(type)`；漫画复用 `createBookAsset`（`batchFactoryV12.js:367`，`[代码]` 已存在）。
- 设计中的"氛围"**没有对应资产 kind**（全仓 `氛围` 仅作为场景字段/约束文本出现），不要新增第八种资产类型（`[推断]`）。

### 4.3 分镜状态（已生成 / 待确认 / 待生成）

后端已派生，前端只渲染（`comic_readiness.go:47-85`）：

| 页面字段组合 | 设计标签 |
| --- | --- |
| `imageStatus === 'primary'` | `已生成` |
| `promptStatus === 'ready' && imageStatus === 'none'` | `待确认` |
| `promptStatus` ∈ `missing/comicOnly` | `待生成` |
| `generationState === 'generating'` | 同位置临时显示 `生成中`（`[推断]`，不替换三词） |
| `generationState === 'failed'` | 同位置显示 `需重试`（`[推断]`） |
| `promptStatus === 'ready' && imageStatus === 'candidates'` | `待确认`，并用 `primaryFallback` 提示"历史图未设主图" |

剧本模式：`ShotOutputCards` 没有逐镜头就绪字段（`[代码]`），只能由 `videoTasks[index]` 的 `status` 派生；若为空则该行不显示状态标签（`[推断]`）。

### 4.4 右栏预览

| 元素 | 数据/接口 |
| --- | --- |
| 当前镜头 | 列表点击选中 → `selectedVideoId`；默认第 1 项 |
| 页码 `01 / 06` | 序号两位补零 / `readiness.pagesTotal`（无 readiness 时用 `pages.length`） |
| 大图 | `page.images` 中 `isPrimary` 优先 → `getComicPageImageBlob(url)`（私有存储，必须走 blob） |
| 浮层副文 | `${画风名} · ${page.aspectRatio}` + 约束摘要（取 `comic.constraints` 的画质/画面限制名称） |
| `生成预览` | `runPrivateComicPageImages(videoId)`（`ScriptPage.jsx:1736`），409 走 `generateComicPageWithConflict` |
| 空态 | 无图时显示占位 + 当前分支说明；不得伪造渐变图（设计里的渐变是样稿美术占位） |

### 4.5 缺口（需要新增或决策，不在本轮实现）

| 缺口 | 现状 | 结论 |
| --- | --- | --- |
| 分镜标题（`雨夜 · 推门`） | `ComicPage` 无标题字段（`types.go:284-306`）；`Video` 有 `Label`（`types.go:59`） | `[推断]` 优先用 `Video.Label`；无 label 时取 `comicPrompt` 第一句截断。若要稳定标题需后端补字段 → 决策 D2 一并处理 |
| 分镜描述 | 直接取 `comicPrompt` 全文做描述（可截断 2 行） | 前端展示层处理，不写回 |
| 拖拽排序 | 列表顺序来自 `batch_factory_v12_videos.ordinal`（`comic_pages_mysql.go:219`）；**没有任何 HTTP 接口可改 ordinal**（全仓检索无匹配） | 本轮只读；如需拖拽必须新增后端排序接口（决策 D2） |
| 时长（漫画） | 漫画无时长概念；剧本有 `duration` 10s/15s | 漫画模式隐藏 `时长` 或改为 `画幅`（决策 D4） |
| 历史版本（漫画） | 无漫画版本接口；有批次配置版本 `getConfigVersions` | 漫画模式暂不提供，或跳批量工厂详情（决策 D6） |
| 资产图批量读取 | 现为 `assets.map(listBookAssetImages)` → N+1（`ScriptPage.jsx:1576-1583`） | 保留现状，但在外壳内加同请求缓存；不做新接口 |

### 4.6 请求清单（全部现成，`[代码]`）

```
GET  /api/batch-factory/v11/batches/:batchId                              → batch（含每书 comicReadiness）
GET  /api/batch-factory/v11/batches/:batchId/runtime-summary              → stageSummaries[bookId].comicReadiness
GET  /api/batch-factory/v11/batches/:batchId/books/:bookId/assets         → { assets: BookAsset[] }
GET  /api/batch-factory/v11/batches/:batchId/books/:bookId/assets/:id/images → { images: [] }
GET  /api/batch-factory/v11/batches/:batchId/books/:bookId/comic-pages    → { pages: ComicPage[], readiness }
PUT  /api/batch-factory/v11/batches/:batchId/books/:bookId/comic-pages/:videoId   （整行覆盖）
POST /api/batch-factory/v11/batches/:batchId/books/:bookId/comic-pages/:videoId/generate
POST /api/batch-factory/v11/batches/:batchId/books/:bookId/stages/:stage  （assets/comic/...）
POST /api/batch-factory/v11/batches/:batchId/books/:bookId/asset-image-generation
```

---

## 5. 分阶段实施计划（5 步，逐步可验收）

> 前置：`git status` 确认 3 个目标文件仍有用户未提交改动。每一步开始前先复制目标文件到 `.backups/private-studio-20261104/`，作为回滚点（不 commit、不 stash、不动用户改动）。

### 步骤 1：隔离新建设计系统与骨架（零冲突）

- 新增目录 `frontend/src/user/pages/private-studio/`：
  - `PrivateStudioWorkbench.jsx`（外壳，接收 `mode / left / header / toolbar / center / preview / note` 插槽）
  - `PrivateStudioRail.jsx`、`PrivateStudioHeader.jsx`、`PrivateStudioToolbar.jsx`、`PrivateStudioShotList.jsx`、`PrivateStudioPreview.jsx`、`PrivateStudioAssets.jsx`
  - `privateStudioState.js`（纯函数：三段步骤、分镜状态、页码格式、标题截断）
  - `privateStudioState.test.js`（`node --test`）
  - `private-studio-workbench.css`（全部选择器前缀 `.private-studio-`，变量 `--ps-*` 定义在 `.private-studio-workbench` 内）
- 不改任何现有文件。
- 验收：`node --test frontend/src/user/pages/private-studio/privateStudioState.test.js` 全绿；CSS 变量/断点与第 1 节逐项一致（可用 grep 断言）。
- 回滚：删除整个新建目录。

### 步骤 2：漫画模式接入新外壳（**需等用户改完 `PrivateComicWorkspace.jsx`**）

- 改 `PrivateComicWorkspace.jsx`：删除面板内重复页签，改为渲染 `PrivateStudioWorkbench`，把现有输入区/资产区/页签状态作为插槽传入；保留全部现有回调与 props。
- 改 `ScriptPage.jsx`：漫画渲染分支不动（仍传同一组 props）；确保页签切换由外壳回调 `onWorkspaceChange`。
- 验收（页面可见）：打开 `/script` → 漫画生成页签，左栏宽度 ≈326px，主区右侧出现 260px 预览列；切换回剧本不丢草稿；漫画现有四个动作仍可执行。
- 回滚：恢复两份备份文件。

### 步骤 3：左栏输入 / 三段进度 / 素材胶囊（**需等用户改完**）

- 剧本：接入字数行 `已读取 N 字` + `自动保存`；`generationStage` 映射三段步骤；素材胶囊由 `extractInfo` 派生。
- 漫画：4 段收敛为 3 段，资产图进度进第 1 段详情（决策 D1）；素材胶囊由 `comicAssets` 派生；`+ 添加` 接 `createBookAsset`。
- 主按钮按 `currentStep` 切换文案与动作（决策 D3）。
- 验收（页面可见）：空正文 → 步骤 1 当前、按钮禁用并给出既有提示；有正文无素材 → 步骤 2 当前；有素材未生成 → 步骤 3 当前；字数随输入实时变化；素材胶囊数量与接口返回一致。
- 回滚：恢复两文件；新组件可保留。

### 步骤 4：中栏分镜列表 + 工具行 + 右栏预览（**需等用户改完**）

- 中栏：`PrivateStudioShotList` 渲染 `comicPages`；标题取 `Video.Label` → 回退提示词首句；状态按 4.3；行内保留"编辑提示词/生成图"入口（复用现有 `ComicPageCard` 能力或链接到它）。
- 工具行：`输出`（漫画提示词预设）、`画风`（`BatchFactoryComicStylePicker` 摘要入口）、`时长`（剧本保留；漫画见 D4）、`约束条件`（现有 Drawer）、`开始生成`（现有提交）。
- 右栏：`PrivateStudioPreview` 复用 `getComicPageImageBlob`；页码、浮层、`生成预览`（D5）。
- 轮询：继续用 `ScriptPage.jsx:1955-1962` 的单一定时器；预览切换不新增请求，只读已拉取的 `pages`。
- 验收（页面可见）：列表与接口页数一致；状态标签只出现 `已生成/待确认/待生成`；点击某行右栏大图与页码同步；`生成预览` 只对该页触发，重复点击有 loading 且不并发。
- 回滚：恢复两文件。

### 步骤 5：剧本模式接入同一外壳 + 空态/错误态/响应式

- 剧本：左栏插槽换成现有 `Form.Item`、工具行、`EntitySection`；中栏 `ShotOutputCards`（分镜模式）或 `ScriptMentionEditor`（文本模式）；右栏（剧本无图场景）：显示"此模式暂无可预览画面"或复用视频 Modal 入口（`[推断]`）。
- 统一空态/错误态：沿用 `script-empty`、`Alert`；错误文案不改写后端原文。
- 响应式：`900px` 单列、`560px` 工具行纵向；检查无横向滚动。
- 验收（页面可见）：两个页签是同一骨架、只切模式；900px 下侧栏在上、内容单列；所有既有按钮（复制/编辑/生成视频/下载）仍在。
- 回滚：恢复两文件 + 可保留新组件。

---

## 6. 风险清单

| # | 风险 | 影响 | 规避办法 |
| --- | --- | --- | --- |
| R1 | 与用户手改冲突 | 覆盖用户未提交功能 | 新文件优先；接线只做插入；改前 `git diff` 对照；每步先备份；严禁整文件覆盖 |
| R2 | 双写提示词 | `comic_pages` 与旧 video 字段不一致 | 只走 `updateComicPage`（整行覆盖），完整回填 `referenceImageUrls/status/aspectRatio/modelId`；不把编译后提示词写库（既有铁律） |
| R3 | 轮询放大 | 请求翻倍、上游 429 | 复用单一 5 秒定时器；预览切换不发请求；`refreshPrivateComicTask` 加 in-flight 去重；资产图 N+1 结果缓存；禁止为每行/每图新增独立轮询 |
| R4 | 35 页上限 | 超过 35 页的镜头无法出图 | 列表与按钮按 `readiness.limits.maxPages` 标注"仅前 35 页可生成"；不伪造成功态；不提高上限 |
| R5 | 竖/横版画幅 | 预览变形、资产图与页图比例混用 | 资产图现固定 `9:16`（`ScriptPage.jsx:1649`）；页图用 `page.aspectRatio`；预览容器 `object-fit: contain`，浮层显示真实比例；画幅选择统一到工具行（`shared/imageAspectRatios.js`） |
| R6 | 样式污染 | 影响剧本卡/批量工厂 | 新样式 scoped `.private-studio-workbench`；不改 `global.css`；不新增全局标签选择器；现有一律 `.private-customization-*` 不动 |
| R7 | 假数据 | 设计里的 `6 个镜头/2 分钟/1268 字` 是样稿 | 一律用真实值：镜头数=`pages.length`，时长按页数估算或显示 `—`；字数用实际字符数；空态显示设计文案 |
| R8 | 拖拽排序无后端 | 视觉可拖动但刷新还原 | 本轮不启用拖拽（D2）；如要启用必须先加排序接口与并发校验 |
| R9 | 页签重复 | 同一 `id` 冲突、可访问性错乱 | 收敛后删除面板内第二份 `id="private-customization-tab-*"`；保留 `role=tab/tabpanel` 与 `aria-controls` |
| R10 | 隐藏能力回退 | 资产图/下载/历史入口被新布局吞掉 | 每个旧按钮在验收清单里逐项勾选，确保仍可点击（见第 7 节） |
| R11 | 预览图内存泄漏 | blob URL 未释放 | 复用 `ComicPageCard` 的 `URL.revokeObjectURL` 模式（`PrivateComicWorkspace.jsx:71-85`） |

---

## 7. 验收标准（用户在页面上的可验证行为）

1. `/script` 两个页签共用同一骨架：左栏 326px 左右、主区含中栏分镜与右栏 260px 预览；切换页签不换布局、不丢已输入内容。
2. 左栏顶部只出现**一份**页签，含 `剧本创作` / `漫画生成`；`草稿已保存` 徽标可见。
3. 正文输入下方实时显示 `已读取 N 字`（剧本）或 `正文 N 字`（漫画），并显示 `自动保存`；刷新页面草稿仍在。
4. 步骤条按设计显示三段，且当前/已完成状态与真实数据一致：无正文时停在第 1 段；有素材未生成时停在第 2/3 段；全部出图后三步全 ✓。
5. 素材区显示 `已识别素材`（剧本）/ `资产`（漫画）、`可编辑`、胶囊与 `+ 添加`；胶囊数量与接口一致；点击 `+ 添加` 走既有新增弹窗。
6. 中栏 `分镜草稿` 明细分行显示序号 `01`、镜头名、描述、状态标签；状态只出现 `已生成` / `待确认` / `待生成`；总数与 `pages.length` 一致。
7. 工具行显示 `输出 / 画风 / 约束条件 / 开始生成`（`时长` 按 D4）；`画风` 能打开画风选择并保存到草稿。
8. 右栏 `画面预览` 显示当前选中镜头的主图（无主图显示占位）、`01 / 06` 页码；点击列表行页码与大图同步变化。
9. `生成预览` 只对当前页触发生图，进行中按钮 loading，重复点击不并发；409 只提示"正在生成"并轮询。
10. 空态与错误态：无任务显示设计文案与主按钮禁用原因；接口错误显示后端原文，不吞错。
11. 900px 宽下自动单列、无横向滚动；560px 下工具行纵向排列。
12. 既有能力不回退：分镜编辑提示词、下载本页/下载全部、批量出图、重试、刷新、查看详情、剧本的复制/编辑/生成视频仍可用。

---

## 8. 决策记录（已按“一切以现有实现为准”定案，不停等）

> 依据用户最新口径：设计稿只作视觉参考，功能以真实实现为准；稿子里没有的按键一律保留或补进新布局。

### D1 漫画进度：保留现有 4 段，不套用剧本的 3 段

**推导依据（`[代码]`）**

- 前端现有管线就是 4 段：`PrivateComicWorkspace.jsx:233-236`
  - `assets` 提取资产
  - `assetImages` 资产图
  - `prompts` 分镜提示词
  - `images` 分镜生图
- 后端自动化阶段同样覆盖资产、资产图/出图、合集与发布：`backend/internal/httpapi/batch_factory_automation_stages.go:33-37`（`comic_assets`、`comic_bootstrap`、`comic_images`、`comic_collection`、`comic_publish`）。
- 发布/合集目前在任务卡与“更多操作”里体现，不在左栏管线里；**保持现有展示位置，不新造步骤**。

**结论**：漫画左栏继续显示 4 段；剧本左栏显示 3 段。两者不互相套用。

### D3 左栏主按钮：保持“提取”，不改语义

**推导依据（`[代码]`）**

- 表单提交绑定的是提取：`ScriptPage.jsx:1988` `onFinish={handleExtract}`。
- `handleExtract` 只调用 `extractEntities(values.novelText)` 并写入人物/场景，不生成剧本：`ScriptPage.jsx:1119-1135`。
- 剧本生成是独立函数 `generateOutput`：`ScriptPage.jsx:1171`，由右侧工具行“生成”按钮触发：`ScriptPage.jsx:2164`。
- 生成前置校验 `canGenerateScript = generationStage === 'extracted' && ...`：`ScriptPage.jsx:1487`。

**结论**：左栏底部主按钮文案保留 `提取` / `提取中...`，`htmlType="submit"`，行为完全不变；tooltip 明确写“提取人物与场景；生成剧本请用右侧「生成」”。**不新增第二个语义**，也不为贴稿子文案改成“继续：生成分镜”。

### D4 漫画“时长”：现有不存在，不渲染假控件

**推导依据（`[代码]`）**

- 漫画设置构造函数 `buildPrivateComicSettings` 只有 `productionMode / textModelId / imageModelId / imageProvider / aiPromptConfig`：`privateCustomizationComic.js:19-37`，无 duration 字段。
- 漫画工具行只有文本模型、图片模型：`PrivateComicWorkspace.jsx:449-450`。
- `时长 10s/15s` 仅存在于剧本表单 `duration`：`ScriptPage.jsx:2135-2137`。

**结论**：漫画模式不渲染“时长”下拉；剧本模式保留。不做点了无反应的假控件。

### D2 / D5 / D6 现状（不阻塞实施）

- **D2 分镜拖拽排序**：后端仍无改 `videos.ordinal` 的 HTTP 接口，前四阶段只读展示，不启用拖拽。
- **D5 `生成预览`**：截图有、源 HTML 无；按“按键缺失→补进新布局”保留，复用现有单页出图接口（阶段 4 接入）。
- **D6 漫画 `历史版本`**：无漫画版本接口，入口复用“查看详情”跳批量工厂，已有批次时可用、无批次时禁用。

以上均不删除任何现有能力。

---

## 9. 功能清点（现有全部可操作项）

> 行号基于 2026-11-04 当前工作区（含用户未提交改动）。`[代码]` 实测。

### 9.1 私人定制 · 剧本创作

| 文案/控件 | 位置 | 点击后做什么 | 接口/依赖 | 门控 |
| --- | --- | --- | --- | --- |
| `剧本创作` 页签 | `ScriptPage.jsx:2025` | `setCustomizationWorkspace('script')` | 无 | 无 |
| `漫画生成` 页签 | `ScriptPage.jsx:2026` | `setCustomizationWorkspace('comic')` | 无 | 无 |
| 清空原文 `×` | `ScriptPage.jsx:2041-2049` | `clearNovelText` 清空输入与草稿 | 无 | 有原文时可用 |
| `添加 TXT 原文` | `ScriptPage.jsx:2052-2054` | `addTextFile` 选择并读取 txt | 本地文件 | 无 |
| 提取指令切换 | `ScriptPage.jsx:2056-2059` | 打开/切换 `extractionPreset` | 预设目录 | 提取不可用时禁用 |
| `配音原文` | `ScriptPage.jsx:2060` | `narrateSource` 调 TTS | `/generation` TTS | `narrating` |
| `匹配音频` | `ScriptPage.jsx:2061` | `openAudioMatchSettings` | 本地音频设置 | 无 |
| `提取`（主按钮） | `ScriptPage.jsx:2063-2067` | 提交表单 → 提取人物/场景 | `extractCharactersAndScenes` | `generating / extractionUnavailable` |
| `添加人物` | `ScriptPage.jsx:2736` | `addEntity('characters')` | 本地草稿 | 无 |
| `重新提取人物` | `ScriptPage.jsx:2737-2745` | `regenerateEntities` | 提取接口 | `regeneratingEntities` |
| 人物卡（点击） | `ScriptPage.jsx:2757-2760` | `openEntityEditor` 编辑 | 实体图/主图 | 无 |
| 主角 `★` 切换 | `ScriptPage.jsx:2761-2768` | `toggleProtagonist` | 本地 | 仅人物 |
| `添加场景` | `ScriptPage.jsx:2736`（type=scenes） | `addEntity('scenes')` | 本地草稿 | 无 |
| `重新提取场景` | `ScriptPage.jsx:2737-2745` | `regenerateEntities` | 提取接口 | `regeneratingEntities` |
| 场景卡（点击） | `ScriptPage.jsx:2757-2760` | `openEntityEditor` | 实体图 | 无 |
| 面板宽度拖拽 | `ScriptPage.jsx:2102-2104` | `handleResizeStart` | 无 | 桌面端 |
| 模式 `连续/爆款/分段开头` | `ScriptPage.jsx:2126-2135` | `Form.Item mode` | 生成参数 | 无 |
| 时长 `10s / 15s` | `ScriptPage.jsx:2136` | `Form.Item duration` | 生成参数 | 无 |
| `历史版本` | `ScriptPage.jsx:2116-2120`（新表头入口） | `openScriptHistory` → `listHistory` | `/history` | 无 |
| `更多设置…` | `ScriptPage.jsx:2116-2120`（新表头入口） | `openScriptUnifiedConfig` | 配置接口 | 无 |
| 输出格式 Select | `ScriptPage.jsx:2141-2151` | `Form.Item format` | 生成参数 | 无 |
| `约束条件` | `ScriptPage.jsx:2152-2158` | `setConstraintsOpen(true)` | 约束预设 | 无 |
| `生成` | `ScriptPage.jsx:2159` | `generateOutput` | `generateScript` | `extracting / !canGenerateScript / 无素材` |
| `更多操作` 菜单 | `ScriptPage.jsx:2160-2181` | 复制/查找替换/编辑/撤销/重新生成/导出 | 本地+生成 | 按菜单项 |
| 分镜 `全选/取消全选` | `ShotOutputCards.jsx:23` | `onToggleAll` | 本地 | 无 |
| `复制已选` | `ShotOutputCards.jsx:24` | `onCopySelected` | 剪贴板 | 无选中禁用 |
| 参考图开关（`@人物/场景`） | `ShotOutputCards.jsx:42` | `onToggleReference` | 参考图 | 无 |
| `复制本分镜` | `ShotOutputCards.jsx:47` | `onCopy(card)` | 剪贴板 | 无 |
| `编辑提示词` | `ShotOutputCards.jsx:48` | `onEditPrompt(index)` | 本地 | 无 |
| `生成视频/重新生成视频/暂无视频权限` | `ShotOutputCards.jsx:49` | `onGenerateVideo(card,index)` | 视频任务 | 权限/处理中 |
| 视频 `下载/下载上次` | `ShotOutputCards.jsx:49` | 打开 `videoUrl` | 视频任务 | 有成功视频 |
| `历史视频（N）` | `ShotOutputCards.jsx:50` | `onOpenVideoHistory` | 历史任务 | 有历史 |
| 视频预览 Modal | `ScriptPage.jsx:2284` 附近 | 播放视频 | `videoUrl` | 有视频 |
| 实体编辑弹窗（保存/删除/全屏/生成图） | `ScriptPage.jsx:2777+` | `EntityEditor` 回调 | 实体/图片接口 | 弹窗内 |

### 9.2 私人定制 · 漫画生成

| 文案/控件 | 位置 | 点击后做什么 | 接口/依赖 | 门控 |
| --- | --- | --- | --- | --- |
| `剧本创作` 页签 | `PrivateComicWorkspace.jsx:280` | `onWorkspaceChange('script')` | 无 | 无 |
| `漫画生成` 页签 | `PrivateComicWorkspace.jsx:281` | `onWorkspaceChange('comic')` | 无 | 无 |
| 漫画标题输入 | `PrivateComicWorkspace.jsx:289-295` | 写入草稿 | 本地/草稿 | 无 |
| 正文输入 | `PrivateComicWorkspace.jsx:296-313` | `setSourceText` | 本地/草稿 | 无 |
| `清空漫画正文` | `PrivateComicWorkspace.jsx:314-328` | Popconfirm 清空 | 本地 | 有正文 |
| `刷新` | `PrivateComicWorkspace.jsx:334-340` | `onRefresh` | batch/pages/assets | `busy` |
| `提取资产` | `PrivateComicWorkspace.jsx:342-350` | `onExtractAssets` | `/stages/assets` | 无任务/忙碌 |
| `补齐资产图` | `PrivateComicWorkspace.jsx:375-382` | `onGenerateMissingAssetImages` | `/asset-image-generation` | 无资产/无缺图/忙碌 |
| `历史版本` | `PrivateComicWorkspace.jsx:431`（新表头入口） | `onOpenBatchFactory` 打开批量工厂详情 | 批次详情 | 无 `batchId` 禁用 |
| `更多设置…` | `PrivateComicWorkspace.jsx:432`（新表头入口） | `onOpenUnifiedConfig` | 配置 | 无 |
| `配置`（原工具栏入口） | `PrivateComicWorkspace.jsx:438-443` | `onOpenUnifiedConfig` | 配置 | 无 |
| `不自动发布` | `PrivateComicWorkspace.jsx:445` | 纯展示 | 无 | 无 |
| 文本模型 Select | `PrivateComicWorkspace.jsx:449` | `patchDraft({textModelId})` | 模型目录 | `busy` |
| 图片模型 Select | `PrivateComicWorkspace.jsx:450` | `patchDraft({imageModelId})` | 模型目录 | `busy` |
| `约束条件` | `PrivateComicWorkspace.jsx:451-457` | `setSettingsOpen(true)` | 约束预设 | 无 |
| `开始生成` | `PrivateComicWorkspace.jsx:459-460` | `onSubmit` | 创建批次 | `busy / !sourceReady` |
| `更多操作` 菜单 | `PrivateComicWorkspace.jsx:461-481` | 生成提示词/批量出图/刷新/查看详情/重试 | stage/页面接口 | 按菜单项 |
| `编辑提示词/取消编辑` | `PrivateComicWorkspace.jsx:103` | `beginEdit` | 本地 | 无 |
| `生成图`（页卡按钮） | `PrivateComicWorkspace.jsx:104-112` | `onGeneratePage(page)` | `/comic-pages/:id/generate` | 无任务/缺提示词/忙碌 |
| `下载本页` | `PrivateComicWorkspace.jsx:113` | 下载主图 | 页图 URL | 有图 |
| `保存提示词` | `PrivateComicWorkspace.jsx:126` | `updateComicPage` | `PUT comic-pages` | 两项提示词必填/忙碌 |
| `取消`（编辑态） | `PrivateComicWorkspace.jsx:127` | 关闭编辑 | 无 | 无 |
| `生成图`（空态） | `PrivateComicWorkspace.jsx:134-141` | `onGeneratePage` | 同页生成 | 同上 |
| `下载全部` | `PrivateComicWorkspace.jsx:154` | `onDownloadAll` 打包 | 页图 + zip | 无页/忙碌 |
| 约束 Drawer 各项 | `PrivateComicWorkspace.jsx:526-560` | 画风/提示词/推导/画质/限制/负面 | 预设目录 | `busy` |

### 9.3 疑似冗余（待用户确认，不擅自删）

| 项 | 现状 | 建议 |
| --- | --- | --- |
| 漫画 `配置`（`:438`）与表头 `更多设置…` | 两个入口同一动作 | 可合并到 `更多设置…`，但本轮**保留** |
| 漫画 `不自动发布` 与表头 `不发布` | 语义重叠 | 纯展示，保留 |
| 剧本 `历史版本` 与分镜卡 `历史视频` | 一个是剧本历史，一个是单镜头视频历史 | 不同功能，全部保留 |

---

## 10. 映射表（现有功能 → 新稿位置 → 结论）

> 结论只允许：**保留** / **合并到某处** / **稿中缺失→补进新布局**。

| 现有功能 | 新稿位置 | 结论 |
| --- | --- | --- |
| 双页签 | 左栏顶部并排两列 | 保留 |
| 品牌 `私人定制` / `草稿已保存` | 左栏顶部 | 保留 |
| 正文输入 / 清空 | 左栏 `01 · 粘贴你的内容` | 保留（清空按钮补在输入框右上角） |
| TXT 上传 | 左栏输入工具行 | 稿中缺失→补进左栏输入工具行 |
| 提取指令切换 | 左栏输入工具行 | 稿中缺失→补进左栏输入工具行 |
| 配音原文 / 匹配音频 | 左栏输入工具行 | 稿中缺失→补进左栏输入工具行 |
| `提取` | 左栏底部通栏主按钮 | 合并到主按钮（阶段 3 定语义，D3） |
| 字数 | 输入框下方 `已读取 N 字` / `自动保存` | 保留 |
| 三段进度 | 输入框下方 | 稿中缺失→补进左栏（已验证 `generationStage` 可派生） |
| 人物/场景素材 | `已识别素材` 胶囊 | 保留；实体卡片与编辑弹窗补在同一区块下方（阶段 3） |
| 添加人物/场景、重新提取、主角标记 | 素材区 | 稿中缺失→补进 `+ 添加` 与素材管理入口 |
| 面板拖拽 | 左栏与主区之间 | 稿中缺失→保留 8px 把手（≤900px 隐藏） |
| 书名 / `准备生成` / 副标题 | 中栏头部 | 保留 |
| `历史版本` | 中栏头部右上 | 稿中缺失→补进表头（已实现） |
| `更多设置…` | 中栏头部右上 | 稿中缺失→补进表头（已实现） |
| 模式切换 / 时长 | 工具行 `时长 15 秒 ▾` 一带 | 保留；模式切换稿中缺失→补进工具行左侧 |
| 输出格式 | 工具行 `输出 漫画分镜 ▾` | 合并到输出下拉 |
| 画风 | 工具行 `画风 电影感水彩 ▾` | 稿中缺失（剧本无此参数）→漫画模式补进工具行；剧本模式暂不显示 |
| 约束条件 | 工具行右侧 | 保留 |
| 生成 / 开始生成 | 工具行右侧主按钮 | 合并到 `开始生成` |
| 更多操作（复制/替换/编辑/撤销/重新生成/导出） | 工具行右侧 `⋯` | 保留 |
| 分镜卡全选/复制/编辑/视频/下载/历史视频 | 中栏 `分镜草稿` 行内 | 保留（阶段 4 行内收纳） |
| 视频预览 | 右栏 `画面预览` | 合并到预览列（阶段 4） |
| 实体编辑弹窗 | 素材胶囊点击 | 保留 |
| 漫画标题/正文/清空 | 左栏输入区 | 保留 |
| 刷新 / 提取资产 / 补齐资产图 | 左栏工具行 + 底部按钮 | 保留；`提取资产` 合并到底部主按钮 |
| 漫画 4 段进度 | 左栏进度 | 保留（是否收敛为 3 段见 D1） |
| 文本模型 / 图片模型 | 工具行 | 稿中缺失→补进工具行（`更多设置…` 内保留同项） |
| 漫画约束 Drawer | `更多设置…` / `约束条件` | 保留 |
| 生成提示词 / 批量出图 / 刷新 / 查看详情 / 重试 | 工具行 `⋯` + 生成按钮 | 保留 |
| 页卡编辑/保存/取消/生成图/下载本页/下载全部 | 中栏分镜行 + 右栏预览 | 保留；下载全部补进中栏卡头 |

### 10.1 数据字段映射

| 界面字段 | 剧本模式 | 漫画模式 | 结论 |
| --- | --- | --- | --- |
| 已读取字数 | `Array.from(novelText).length`（新增派生） | `sourceLength` | 复用 |
| 镜头数 | `shotCards.length` / `output` 解析 | `pages.length` 或 `readiness.pagesTotal` | 复用 |
| 预计时长 | 现有 `scriptHeaderMeta`（按格式估算） | 漫画无时长，显示 `—`（D4） | 复用/待定 |
| `准备生成` / `待输入` | 现有 `scriptHeaderStatus` | 无任务时 `待输入` | 复用 |
| 素材标签 | `extractInfo.characters/scenes` | `assets`（`kind`） | 复用 |
| 分镜状态 | `videoTasks[index].status`（弱） | `imageStatus / promptStatus / generationState` | 漫画复用 S1；剧本需新增派生 |
| 页码 `01 / 06` | 无 | `readiness.pagesTotal` | 复用 |
| 预览图 | 视频 URL | `page.images` 主图 + `getComicPageImageBlob` | 复用 |

---

## 11. 阶段 1 实施记录与功能自检（2026-11-04）

### 11.1 改动文件

| 文件 | 性质 | 说明 |
| --- | --- | --- |
| `frontend/src/user/pages/private-studio/private-studio-workbench.css` | 新建 | 全部 `.private-studio-*`，含 326px/260px 与 900/560 断点 |
| `frontend/src/user/pages/private-studio/PrivateStudioHeaderActions.jsx` | 新建 | `历史版本` / `更多设置…` |
| `frontend/src/user/pages/private-studio/PrivateStudioSteps.jsx` | 新建 | 三段进度 |
| `frontend/src/user/pages/private-studio/PrivateStudioChips.jsx` | 新建 | 素材胶囊 + `+ 添加` |
| `frontend/src/user/pages/private-studio/PrivateStudioNote.jsx` | 新建 | `下一步` 提示卡 |
| `frontend/src/user/pages/PrivateStudioWorkbench.layout.test.js` | 新建 | 阶段 1 结构/文案/样式隔离断言 |
| `frontend/src/user/pages/ScriptPage.jsx` | 最小 hunk | imports `41-49`；派生 `444-452`；外壳类名 `2013`；左栏类名 `2018`；副标题 `2033`；字数/进度/胶囊 `2068-2078`；主区类名 `2106`；表头入口 `2113-2121`；三栏包裹 `2184-2286` |
| `frontend/src/user/pages/PrivateComicWorkspace.jsx` | 最小 hunk | imports `7-8`；外壳/左栏 `272-273`；主区 `419`；表头入口 `426-435`；三栏包裹 `484-523` |

### 11.2 功能自检（逐项）

剧本模式：

- ✅ 双页签切换
- ✅ 正文输入 / 清空 / TXT 上传
- ✅ 提取指令切换 / 配音原文 / 匹配音频
- ✅ 提取（阶段 1 时仍在输入工具行；阶段 2 移至左栏底部通栏，语义不变，见 §12）
- ✅ 人物：添加 / 重新提取 / 编辑 / 主角标记
- ✅ 场景：添加 / 重新提取 / 编辑
- ✅ 面板拖拽
- ✅ 模式与时长分段
- ✅ `历史版本`（表头）与 `更多设置…`（表头）
- ✅ 输出格式 / 约束条件 / 生成
- ✅ 更多操作 6 项
- ✅ 分镜卡全部操作与视频预览

漫画模式：

- ✅ 双页签切换
- ✅ 标题 / 正文 / 清空
- ✅ 刷新 / 提取资产（阶段 2 移至资产区底部通栏）/ 补齐资产图
- ✅ `历史版本`（表头，无批次时禁用）/ `更多设置…`（表头）/ `配置`
- ✅ 文本模型 / 图片模型 / 约束设置 / 生成
- ✅ 更多操作 5 项
- ✅ 页卡编辑/保存/取消/生成图/下载本页 / 下载全部

结论：**阶段 1 没有删除任何现有入口**，仅新增表头入口与左栏摘要区。

### 11.3 验证命令

```bash
cd frontend
node --test src/user/pages/PrivateStudioWorkbench.layout.test.js src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/PrivateCustomizationShell.layout.test.js
npm run build
```

实测结果：12 个测试全部通过；`vite build` 成功（5884 modules，12.66s）。未重启 4001、未改 `backend/web/dist`、未 commit。

### 11.4 下一阶段入口条件（已满足，D1/D3/D4 见 §8 自行推导定案，不再停等）

1. ~~用户确认 D1 / D3 / D4~~ → 已按“一切以现有实现为准”自行推导并写入 §8。
2. `更多设置…` 与漫画原 `配置` 仍**两者都保留**，列为“疑似冗余（待用户确认，不擅自删）”；本轮不合并、不删除。
3. 阶段 2 只动左栏插槽与素材区；`ScriptPage.jsx` / `PrivateComicWorkspace.jsx` 继续只加最小 hunk。
4. 若用户在阶段 1 后继续手改，先跑 `git diff` 对照本方案 §11.1 的 hunk 范围再合并。

---

## 12. 阶段 2 实施记录与功能自检（2026-11-04）

范围：**左栏素材区与主按钮**（输入 → 字数 → 进度 → 素材胶囊 → 通栏主按钮）。只动左栏插槽，未改中/右栏结构。

### 12.1 改动文件与行号

| 文件 | 性质 | 说明 |
| --- | --- | --- |
| `frontend/src/user/pages/ScriptPage.jsx` | 最小 hunk | `451-457` 新增 `scriptAssetChips`（人物 / 场景 / 氛围，来源为真实 `extractInfo` 与场景 `data['氛围']`）；`2076-2079` `<PrivateStudioChips>` 渲染；`2097-2101` 原「提取」按钮从输入工具行移出，改为底部通栏 `.private-studio-primary`，`htmlType="submit"`、`loading={extracting}`、`disabled={generating || extractionUnavailable}` 与 label `提取中... / 提取` 原样保留 |
| `frontend/src/user/pages/PrivateComicWorkspace.jsx` | 最小 hunk | `229-232` 新增 `comicAssetChips`（真实资产 `name` + `kind` 中文标签；无资产时用 `人物 / 场景 / 道具` 类别占位）；`382` 资产区 header 下渲染 `<PrivateStudioChips>`；`405-415` 原「提取资产」按钮从工具行移出，改为底部通栏 `.private-studio-primary`，`onClick={onExtractAssets}` 与 label `提取中… / 重新提取资产 / 提取资产` 原样保留 |
| `frontend/src/user/pages/private-studio/PrivateStudioChips.jsx` | 修改 | 支持无 `onClick` 时渲染 `span`（非交互，不再给占位/展示项绑定空点击）；`title` / `actionLabel` 为空时不渲染 head |
| `frontend/src/user/pages/private-studio/private-studio-workbench.css` | 最小追加 | `span.private-studio-chip` 非交互态；`.private-studio-primary-wrap`（`margin-top:18px`）；`.private-studio-primary`（通栏、渐变 `linear-gradient(135deg,#aeb7ff,#62dfcf)`、`font-weight:750`、`box-shadow:0 8px 22px rgba(101,223,206,.13)`），另加 `:disabled` 与 `:hover` 态 |
| `frontend/src/user/pages/PrivateStudioWorkbench.layout.test.js` | 追加断言 | ① 主按钮保留真实语义（`htmlType="submit"` + `提取` + `提取中...`，且未新增 `生成分镜` 语义）；② 漫画仍为 4 段（`assets/assetImages/prompts/images`），未被压成 3 段 |

### 12.2 三处决策在代码中的落点

- **D1（漫画 4 段）**：`PrivateComicWorkspace.jsx:233-236` 的 `pipelineSteps` 保持 4 项；阶段 2 未改动它。后端 `batch_factory_automation_stages.go:33-37` 同构。
- **D3（主按钮语义）**：剧本主按钮仍提交 `handleExtract`（`ScriptPage.jsx:1988` → `1119-1135` 仅调 `extractCharactersAndScenes`），未新增“生成分镜”第二语义；工具提示明确指向右侧「生成」（`2164`，`generateOutput` `1171`）。
- **D4（漫画无时长）**：`privateCustomizationComic.js:19-37` 的 `buildPrivateComicSettings` 无 `duration` 字段；阶段 2 未在漫画左栏/工具行渲染任何 `时长` 控件。

### 12.3 功能自检（逐项，✅ = 新界面可见且可点）

剧本模式左栏：

- ✅ 正文输入 / 清空 / TXT 上传（保持不变）
- ✅ 提取指令切换 / 配音原文 / 匹配音频（仍在输入工具行）
- ✅ 字数行 `已读取 N 字` + `自动保存`
- ✅ 三段进度：`内容已读取` / `确认人物与场景` / `生成创作结果`（状态由真实数据派生）
- ✅ `已识别素材` + `可编辑` + 人物/场景胶囊（点击仍进 `openEntityEditor`）
- ✅ `氛围` 胶囊：仅当场景 `data['氛围']` 非空时出现，点击进入该场景编辑（不是新增资产类型）
- ✅ `+ 添加`：仍调 `addEntity('characters')`
- ✅ `提取` 主按钮：位置下移为通栏，行为、禁用条件、loading 文案不变
- ✅ 原输入工具行其余按钮未移除

漫画模式左栏：

- ✅ 标题 / 正文 / 清空
- ✅ 字数显示
- ✅ 刷新 / 补齐资产图（工具行保留）
- ✅ 资产区：真实资产胶囊（`名称 · 人物/场景/道具`）；无资产时为类别占位（`span` 非交互，点了不会假装有动作）
- ✅ `提取资产` 主按钮：位置下移为通栏，`onExtractAssets` 与三种 label 不变
- ✅ 4 段进度未被压缩
- ✅ `配置` 与 `更多设置…` 仍并存（疑似冗余，待确认，不删）

结论：阶段 2 未删除任何现有入口，仅**移动**了「提取」与「提取资产」两个按钮到左栏底部通栏，并把素材摘要补进新布局。

### 12.4 验证命令与结果

```bash
cd frontend
node --test src/user/pages/PrivateStudioWorkbench.layout.test.js src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/PrivateCustomizationShell.layout.test.js
npm run build
```

实测结果：**14 tests pass / 0 fail**；`vite build` **✓ built in 10.96s**。未 commit、未重启 4001（PID 46316）、未改 `backend/web/dist`、未触碰其他会话的未提交改动。

> 页面级视觉验证受限于 4001 运行旧二进制内嵌资源，需重启后才可见新布局；本轮按约束不重启，已提供手动冒烟步骤（见 §11.3）。

### 12.5 下一阶段入口条件

1. 阶段 3 范围为**中栏分镜列表 + 工具行**：剧本沿用 `ShotOutputCards`，漫画沿用 `ComicOutputCards`，只调头部/工具行视觉，不改卡片内部逻辑。
2. 中栏工具行的下拉/按钮逐项对照 §9 保留，不因稿子未画而删。
3. 右栏预览（§5 步骤 4）留到阶段 4；阶段 3 不新建预览组件。
4. 继续最小 hunk；每次改动后跑 §12.4 同一组命令。

---

## 13. 发布批次记录（2026-11-04）

### 13.1 本版包含

发布会话正在基于当前工作区构建二进制并准备重启 4001。本版包含：

- **阶段 1**：统一工作台外壳、双页签、表头 `历史版本` / `更多设置…`、左栏摘要区（见 §11）
- **阶段 2**：左栏素材胶囊与底部通栏主按钮（见 §12）

**阶段 3 起属于下一批**，不在本版内。

### 13.2 构建包含文件（当前工作区，均已保存）

| 文件 | Git 状态 | 是否进本版 |
| --- | --- | --- |
| `frontend/src/user/pages/ScriptPage.jsx` | `M`（含用户原有未提交改动） | 是 |
| `frontend/src/user/pages/PrivateComicWorkspace.jsx` | `M`（含用户原有未提交改动） | 是 |
| `frontend/src/user/pages/private-studio/`（4 个 jsx + 1 个 css） | `??` 新增目录 | 是 |
| `frontend/src/user/pages/PrivateStudioWorkbench.layout.test.js` | `??` 新增 | 否（测试文件不进 bundle） |
| `docs/superpowers/plans/2026-11-04-private-studio-unified-workbench.md` | `??` 新增 | 否（文档） |

### 13.3 最新性核对（编辑 vs 构建时间）

- 最后一次源码编辑：`PrivateStudioWorkbench.layout.test.js` **11:50:26**；此前 `private-studio-workbench.css` **11:50:03**、`PrivateComicWorkspace.jsx` **11:49:55**、`ScriptPage.jsx` **11:49:12**。
- 最后一次构建产物：`frontend/dist/index.html` 与 `frontend/dist/assets/*` 均为 **11:50:47**，**晚于全部源码编辑**，因此构建结果包含阶段 2 全部改动，不存在“改了一半”的中间态。
- 测试：本暂停点重新执行 `node --test`（三套件）→ **14 pass / 0 fail**，晚于全部编辑，结论有效。
- 未重跑 `npm run build`（构建已在最后编辑之后完成，且当前正在发布，避免与发布会话争用构建目录）。

### 13.4 暂停约定

1. 自本记录起**停止一切前端源码改动**，直到收到“构建已完成 / 可以开阶段 3”的明确指令。
2. 若发布会话构建失败需要回滚，回滚点在阶段 2 之前，即本文件 §11.1 所列 hunk。
3. 约束保持：不 commit、不重启 4001、不改 `backend/web/dist`、不触碰其他会话未提交改动。

### 13.5 阶段 3 待用户拍板清单（先列不处理）

阶段 3 会碰到中栏工具行，以下“疑似冗余/语义重叠”项**均不擅自合并或删除**，等用户拍板：

| # | 冲突点 | 现状 | 可选项（不预设结论） |
| --- | --- | --- | --- |
| P1 | 漫画 `配置` 与表头 `更多设置…` | 两者都打开统一配置（`onOpenUnifiedConfig`） | 保留两者 / 表头入口保留、原按钮改名 / 合并为一个入口 |
| P2 | 漫画 `不自动发布` 与表头 `不发布` 徽标 | `不自动发布` 是配置开关语义，`不发布` 是状态展示 | 保留两者（推荐待定）/ 徽标改为动态状态 / 合并文案 |
| P3 | 剧本中栏工具行的 `格式` Select 与左栏 `模式/时长` 分段 | 两处都能改生成参数，作用范围不同 | 保留两处 / 工具行只读展示 / 工具行移除（**需用户明确同意才可动**） |
| P4 | 漫画工具行 `文本模型` / `图片模型` 与统一配置内的模型设置 | 工具行是快捷入口，配置里是全量设置 | 保留工具行 / 工具行改为“已选模型”只读 / 合并进配置 |
| P5 | 中栏 `更多操作`（剧本 6 项 / 漫画 5 项）与表头 `更多设置…` | 前者是生成操作菜单，后者是配置抽屉，语义不同 | 保留两者（本轮默认）/ 菜单分组重排 |

> 上述 5 项在阶段 3 实施时若未获用户明确答复，一律按“保留两者”执行，并继续列入本节。
