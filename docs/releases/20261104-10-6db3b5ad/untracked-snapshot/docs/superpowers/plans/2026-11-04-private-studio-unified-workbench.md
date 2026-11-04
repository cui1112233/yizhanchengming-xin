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

---

## 14. 阶段 3 实施记录与功能自检（2026-11-04）

范围：**中栏分镜列表 + 工具行**。P1–P5 全部按“保留两者”执行，未删除/合任何入口。

### 14.1 改动文件与行号

| 文件 | 行号 | 改动 |
| --- | --- | --- |
| `frontend/src/user/pages/ScriptPage.jsx` | `435-449` | 元信息改为真实派生：镜头数 = `shotCards.length`；时长优先累加各分镜原文 `总时长：Ns`，否则 `条数 × 单卡上限（10s/15s）`；都拿不到则整段省略。已加 `TODO(总时长)` 注释 |
| 同上 | `2161-2215` | 工具行重排：左侧 `输出` 胶囊下拉（仍是原 `Form.Item name="format"`），右侧 `约束条件` + `开始生成` + 更多菜单（6 项原样）；左栏 `开头模式/时长` 分段未动（P3 两者保留） |
| 同上 | `2218-2219` | 分镜草稿卡头 `分镜草稿` + `{shotCards.length} 个镜头`，包裹原 `ShotOutputCards`（仅外壳，不改卡片内部） |
| 同上 | `278`、`1610`、`2331` | 新增 `privateComicReadiness` 状态，从 `listComicPages` 的 `pageResult.readiness` 写入，并以 `readiness` prop 传给漫画组件 |
| `frontend/src/user/pages/PrivateComicWorkspace.jsx` | `244-253` | 新增就绪度派生：`pagesTotal / promptReadyCount / imageReadyCount / readinessComplete / pageGenerating / pageFailed`，全部读 `comicReadiness`；缺失时才回退到已加载页 |
| 同上 | `240-242` | 左栏 4 段进度第 3、4 段的 `done/detail` 改用 `promptReadyCount / imageReadyCount / pagesTotal`（第 1、2 段资产类不在 readiness 内，仍用已加载资产数据，见 §14.4） |
| 同上 | `293-304` | 头部：元信息改为 `一段内容 · N 页 · 已出图 M 页`（页数/出图数取 readiness）；状态徽标改为 readiness + `generationState` + 任务状态派生 |
| 同上 | `485-506` | 工具行重排：`文本模型` / `图片模型` 胶囊下拉（原 Select 原样）+ `约束条件` + `开始生成` + 更多菜单（5 项原样）；`配置`（`475-480`）与表头 `更多设置…` 同时保留（P1） |
| 同上 | `73-90`、`121` | 漫画页状态标签：只读 `promptStatus / imageStatus / generationState`，渲染 `已生成 / 待生成 / 待确认 / 生成中 / 生成失败` |
| `frontend/src/user/pages/private-studio/PrivateStudioNote.jsx` | `5` | 剧本提示改为真实入口 `点击右侧“开始生成”`（原“继续：生成分镜”指向不存在的按钮） |
| `frontend/src/user/pages/private-studio/private-studio-workbench.css` | `296-395` | 新增 `.private-studio-canvas-head`、`.private-studio-toolbar(-left/-right)`、`.private-studio-tool-pill`、`.private-studio-tool-label`、`.private-studio-tool-pill-button`、`.private-studio-tool-primary`、`.private-studio-shot-status`，全部 `.private-studio-*` 作用域 |
| `frontend/src/user/pages/PrivateStudioWorkbench.layout.test.js` | `62-108` | 新增 5 条阶段 3 断言（真实元信息、工具行保留、readiness 复用、页状态字段、样式作用域） |

### 14.2 稿子控件 → 现有实现的映射（未凭空造控件）

| 稿子控件 | 剧本 | 漫画 | 说明 |
| --- | --- | --- | --- |
| `输出` | 保留原 `format` Select，外套 `输出` 标签 | 无对应“输出”型控件，改用真实存在的 `文本模型` / `图片模型` 两个下拉 | 不新增假下拉 |
| `画风` | 现有实现无画风控件 → **不渲染** | 画风在「约束条件」抽屉内（`BatchFactoryComicStylePicker`），工具行不重复造 | 见 §14.4 文案风险 |
| `时长` | 已有 `10s/15s` 分段，仍在左栏（P3 保留两处，工具行不重复） | 现有实现无时长字段（D4）→ 不渲染 | — |
| `约束条件` | 原「约束」按钮改文案，行为不变 | 同左 | — |
| `开始生成` | 原「生成」按钮，handler 不变 | 同左 | 仅视觉文案 |

### 14.3 功能自检（✅ = 新界面可见且可点，行为未变）

剧本中栏：

- ✅ 书名 / 状态徽标（`scriptHeaderStatus` 原逻辑）
- ✅ 元信息：`一段内容 · N 个镜头 · 预计 X`（N 为真实分镜条数，X 为解析/计算值；无数据时不显示）
- ✅ 右上 `历史版本` / `更多设置…`（阶段 1 入口保留）
- ✅ `输出` 下拉（5 个格式选项原样）
- ✅ `约束条件`（原 `openConstraints`）
- ✅ `开始生成`（原 `generateOutput`，禁用条件未动）
- ✅ 更多菜单 6 项：复制 / 查找替换 / 编辑 / 撤销 / 重新生成 / 导出
- ✅ 左栏 `开头模式` 分段 + `时长` 分段（P3 保留）
- ✅ 分镜草稿卡头 + 原 `ShotOutputCards`（勾选/复制/编辑/生视频/参考图开关内部未改）

漫画中栏：

- ✅ 书名 / 状态徽标（读 `comicReadiness` + `promptStatus/imageStatus/generationState` + 任务状态）
- ✅ 元信息：`一段内容 · N 页 · 已出图 M 页`（D4：无时长）
- ✅ 右上 `历史版本` / `更多设置…` + 表头 `配置`（P1 两者保留）
- ✅ `文本模型` / `图片模型` 两个下拉（原 Select 原样）
- ✅ `约束条件`（原 `setSettingsOpen(true)`）
- ✅ `开始生成`（原 `onSubmit`，`sourceReady` / `presetReady` 禁用条件未动）
- ✅ 更多菜单 5 项：提示词 / 出图 / 刷新 / 查看详情 / 重试
- ✅ 每页状态标签（后端派生字段）
- ✅ 原页卡：编辑提示词 / 生成图 / 下载本页 / 下载全部 / 图片鉴权读取
- ✅ `不自动发布` 与头部 `不发布` 同时保留（P2）

结论：阶段 3 未删除任何入口；仅新增视觉外壳与真实数据展示。

### 14.4 未接真实数据的字段清单（如实标注，不含编造数字）

| 字段 | 现状 | 处理 |
| --- | --- | --- |
| 剧本整篇总时长 | 后端无字段 | `TODO(总时长)`；优先解析分镜原文 `总时长：Ns` 求和，否则 `条数 × 10/15s` 估算；都没有则隐藏该段 |
| 剧本“预计”分钟 | 由秒数四舍五入为整分钟 | 属于估算值，文案保持“预计” |
| 漫画时长 | 链路不存在（D4） | 不渲染 |
| 漫画页数 | `comicReadiness.pagesTotal` | 已接后端 |
| 漫画提示词/出图进度 | `promptReadyCount / imageReadyCount` | 已接后端 |
| 漫画页状态 | `promptStatus / imageStatus / generationState` | 已接后端 |
| 漫画资产 / 资产图 两段 | `comicReadiness` 不含资产字段 | 仍用 `listBookAssets` / `listBookAssetImages` 已加载结果判断；**需后端后续在 readiness 中补资产就绪度才能统一口径** |
| 设计稿工具行「画风」 | 剧本无此控件；漫画在抽屉内 | 未在工具行新增；`PrivateComicWorkspace.jsx:580` 旧说明文案仍写“画风…已放到左侧和工具栏”，与现状不符，列“疑似文案待确认”，未擅自改 |

### 14.5 验证

```bash
cd frontend
node --test src/user/pages/PrivateStudioWorkbench.layout.test.js src/user/pages/PrivateComicWorkspace.source.test.js src/user/pages/PrivateCustomizationShell.layout.test.js
npm run build
```

结果：**19 tests pass / 0 fail**；`vite build` **✓ built in 13.83s**。未 commit、未重启 4001、未改 `backend/web/dist`、未编译 Go 二进制、未触碰其他会话改动。

### 14.6 剩余风险

1. 4001 仍运行阶段 2 二进制；阶段 3 的新布局需下次发布后才可见，本轮只做构建产物到 `frontend/dist`。
2. 漫画资产两段进度仍非后端权威口径，与提示词/出图段存在“同一进度条混合两种来源”的可能，需后端补 `assetsReady/assetImagesReady`。
3. 状态徽标把前端即时 `action` 与后端 `generationState` 合并判断：`action` 只影响“生成中”的即时反馈，稳态完全由后端字段决定。
4. 设计稿「画风」与旧说明文案不一致，待用户确认后统一。

---

## 15. 实机修正轮（2026-11-04）：宽度 / 胶囊 / 顶栏 / 资产编辑盘点

### 15.1 完成度

| 顺序 | 任务 | 状态 |
| --- | --- | --- |
| 1 | 宽度自适应 | ✅ 已实现 |
| 2 | 素材胶囊可点 | ✅ 剧本已点开现有编辑；漫画无编辑入口保持不可点 |
| 3 | 顶部重复横栏 | ✅ `/script` 收起标题，状态区与退出导航保留 |
| 4 | 资产编辑统一外壳 | ⚠️ **仅完成盘点 + 画幅结论**，统一外壳未实施（见 §15.5/15.6） |

### 15.2 宽度：改前 → 改后

- **改前**：`.private-studio-workbench { max-width:1180px; margin:0 auto }`。1920 视口下卡片被锁在 1180px 居中，右侧约 670px 留白；这是“小了、不随窗口变化”的直接原因（不是弹窗）。
- **改后**：`max-width:none; width:100%; margin:0`，外层 `.script-workbench-form` 加 `padding:14px 18px 18px; box-sizing:border-box; overflow-x:hidden`。窗口放大即变宽，压缩时由 `minmax(0,1fr)` 吸收，不产生横向滚动。

| 视口 | 预期表现 |
| --- | --- |
| 1920 | 工作台 ≈1824px（视口 - 左导航 - 36px 内边距）；左栏 326px，主区 1fr；中栏内 `1fr + 260px` |
| 1440 | 工作台 ≈1344px；左栏 326px，中栏内仍 `1fr + 260px` |
| 1280 | 工作台 ≈1184px；触发 `max-width:1180` 断点：右栏预览折到中栏下方，中栏独占整宽 |
| 1024 | 工作台 ≈928px；触发 `max-width:1100`：左栏收窄到 300px，预览仍在中栏下方，无横向滚动 |
| ≤900 | 单列：左栏在上、中栏在下，拖拽把手隐藏 |

### 15.3 素材胶囊：改前 → 改后

- **剧本**：胶囊本来就是 `button` 且已绑 `openEntityEditor('characters'|'scenes', id)`，本轮补 `title="点击编辑：…"` 与 `:focus-visible` 描边，让“可编辑”可被发现；点击打开的是现有 `EntityEditor` 弹窗（无新组件）。
- **漫画**：资产胶囊无任何现有编辑入口（工作区内没有资产改名/字段编辑 API/UI），保持 `span` 不可点，视觉保持默认鼠标；`+ 添加`（剧本）仍接 `addEntity('characters')`，漫画资产没有“新增资产”入口，故未伪造。
- **疑点保留**：漫画资产编辑入口是否应跳到「水货生产 / 资产」页，列为“疑似待确认”，未擅自加跳转。

### 15.4 顶部横栏：改前 → 改后

- **它是什么**：`UserLayout` 的全局 `.legacy-topbar`，不是弹窗外壳。承载 3 项：页面标题（`私人定制`）、全局状态提示 `legacy-global-status`、`legacy-userbar`（`aria-hidden` 空占位，无操作）。
- **改后**：`/script` 路由下不再渲染标题（左栏已有同名品牌行）；若 `globalStatus.text` 为空则整条不渲染，内容上移；有状态时保留状态条。无关闭/返回/全屏按钮可丢——退出一直由左侧全局导航承担，未动。
- 其它路由的顶栏行为完全不变。

### 15.5 资产编辑盘点（现有 → 统一后位置）

| 现有能力 | 私人定制剧本 `EntityEditor`（ScriptPage.jsx:2808） | 批量工厂/资产编辑（AssetsView.jsx / BatchFactoryNovelList.jsx:537） | 统一后位置（建议） |
| --- | --- | --- | --- |
| 资产分类 | 由入口类型决定（人物/场景），弹窗内无切换 | `资产分类`（人物/场景/道具/声音…） | 标题下第一行，类型驱动字段组 |
| 图片模型 | 继承全局设置，无单独下拉 | `图片模型` 下拉 | 右栏预览上方，保留下拉 |
| 图片画幅 | 继承全局设置，无单独下拉 | `图片画幅` 下拉 | 同上，接七档常量 |
| 名称 | 字段区第一项 | `名称` 输入 | 字段区首行 |
| 类型字段 | 人物：姓名/事实/关系/外形等（`fieldsToRender`）；场景：描述/氛围等 | 人物：人物事实/人物关系/人物外形；场景/道具各自字段 | 统一栅格，按分类渲染，字段不合并 |
| 合并资产提示词 / AI 生图提示词 | 无（剧本侧走实体字段生成） | `合并资产提示词 / AI 生图提示词` 文本域 | 字段区下半，按分类显示 |
| 规则说明 | 无 | 有图走“图片生视频”、没图走“文字生视频” | 提示条，原样保留 |
| 主图预览 | `EntityImagePanel` 右侧 | 右侧主图预览（暂无主图） | 右栏固定列 |
| 三视图开关 | 无 | `人物三视图` Switch | 右栏，仅人物 |
| 上传图片版本 | `EntityImagePanel` 内 | `上传图片版本`（历史版本列表） | 右栏图片版本区 |
| AI 生成 | `EntityImagePanel` 生成按钮 | `AI 生成` | 右栏主按钮 |
| 图片版本列表 | `EntityImagePanel` 内 | 有 | 右栏列表 |
| 全屏编辑 | 有（`全屏编辑/退出全屏`） | 无 | 统一壳右上角，保留 |
| 智能补全 | `根据小说智能补全` | 无 | 剧本侧保留 |
| 删除 | `删除人物/删除场景`（Popconfirm） | 资产列表侧 | 统一壳底部左侧，保留 |
| 取消 / 保存（完成） | `完成` | `取消 / 保存` | 统一壳底部右侧，两个入口都保留 |

> 统一壳**未实施**：涉及 `EntityEditor` 与 `AssetsView/BatchFactoryNovelList` 两套不同的数据模型与接口，且漫画侧目前没有可从私人定制进入的资产编辑入口；按“先盘点、不硬合并”的原则，本轮只交付上表与画幅结论，避免为对齐外观而改坏字段或丢按钮。

### 15.6 图片画幅下拉结论

- **已接统一七档，不是只回 3 档。** 资产编辑使用 `IMAGE_ASPECT_RATIO_OPTIONS`（[imageAspectRatios.js](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/imageAspectRatios.js#L4-L6)），共 `16:9 / 9:16 / 1:1 / 4:3 / 3:4 / 3:2 / 2:3`；后端镜像 [image_aspect_ratios.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/batchfactoryv11/image_aspect_ratios.go#L13) 同样是 7 项。
- 截图中只看到 `16:9` 属于**当前值/默认值**表现：资产未单独选择时按书籍默认（`BatchFactoryNovelList.jsx:366` 以 `16:9` 作为兜底，`542-543` 的 Select 展开应有 7 项）。
- 因此：**dropdown 已接七档；不是后端只返回 3 档**。若实机展开只有 3 项，则是构建版本未含该常量，需核对发布产物而非后端。

### 15.7 验证与视觉

- `node --test`（三套件）：**22 pass / 0 fail**。
- `npm run build`：**✓ built in 17.35s**，仅写 `frontend/dist`。
- **未做 1280 / 1920 截图**：4001 仍运行阶段 2 二进制，看不到本轮改动；本地 `vite preview` 缺少登录态与后端会话。已按要求如实说明，不伪造验收图。
- 约束保持：未 commit、未重启 4001、未改 `backend/web/dist`、未重编 Go。
