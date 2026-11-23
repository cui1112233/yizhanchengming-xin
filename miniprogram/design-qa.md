# 功能首页卡片区域 Design QA

- source visual truth: `C:/Users/ADMINI~1/AppData/Local/Temp/codex-clipboard-b72f9e5f-5398-4d68-b31c-7f5917fdc764.png`
- implementation screenshot: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-functions-card-cache-cleared.png`
- normalized implementation region: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-functions-card-implementation-region.png`
- side-by-side comparison: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-functions-card-comparison.png`
- viewport: WeChat Developer Tools, iPhone 12/13 simulator, about `390px` CSS width
- source pixels: `425 x 466`
- normalized implementation pixels: `425 x 466`, scaled from the simulator's `365 x 400` desktop-stage capture
- state: 功能首页、深色主题、顶部以下的“内容创作/生产成片/AI 与工具”区域

## Findings

本轮最终对比没有剩余 P0、P1 或 P2 差异。

- 两列宽度、主卡高度、小卡高度和组间节奏与参考图接近，归一化后的主要几何偏差在约 5% 内。
- “私人定制”已换成参考方向的金色羽毛、魔法书、环形青光与 AI 晶体独立透明插画。
- “漫剧创作”已换成参考方向的耳机人物、晚霞画面、播放键与环绕胶片独立透明插画。
- 九个入口仍是独立按钮，未用整张截图替代交互页面。

## Full-view comparison evidence

`qa-functions-card-comparison.png` 左侧为用户参考图，右侧为微信模拟器实现区域。对比覆盖两个完整分组和第三组标题，能直接判断页面边距、两列比例、卡片高度、标题间距以及两张主插画的可视占比。

## Focused region comparison evidence

- `private-custom-reference-3d.webp` 与 `comic-production-reference-3d.webp` 均为 `384 x 384`、带 Alpha 的 WebP。
- 微信模拟器最终截图能看到私人定制的 AI 晶体，也能看到漫剧创作画面中的耳机人物，证明运行时没有继续使用旧缓存素材。
- 不再需要额外局部截图；两张主卡在归一化对比图中已足够清晰。

## Required fidelity surfaces

- Fonts and typography: 分区标题约 19px、主卡标题约 20px、小卡标题约 13px；层级、粗细和单行截断与参考一致。实现字体略粗属于可接受的系统字体差异。
- Spacing and layout rhythm: 390px 基准左右边距 15px，列间距 10px，主卡 138px，高度由两张 62px 小卡和 14px 间距严格对齐；分区间距 24px。
- Colors and visual tokens: 深海军蓝背景、蓝绿玻璃卡片、青色边框与参考方向一致；没有增加遮字的大面积炫光。
- Image quality and asset fidelity: 两张新主图来自参考图定向生成，使用 384px 透明 WebP；微信中显示清晰，没有旧 160px 素材的明显模糊。
- Copy and content: 私人定制、小说创作、批量测书、水货生产、漫剧创作、AI-CUT、CM 画布、配音、插件均保留；分区文案不变。

## Comparison history

1. P1 黑屏/高度塌陷：移除功能区 CSS Grid 与首图自动高度依赖，改为明确比例和 Flex；修复后页面可见。
2. P2 卡片偏小：将主卡调整为 138px，小卡调整为 62px，列间距 10px，重新截图后几何接近参考。
3. P2 主插画偏小：增大主卡插画可视区域；第二轮发现漫剧图压字后停止使用旧图定位方案。
4. 用户更新素材要求：生成并接入两张新透明素材；首次微信截图仍命中旧编译缓存。
5. P1 旧资源缓存：使用新文件名、清理微信编译缓存并重新打开项目；最终截图显示两张新素材，控制台 0 错误。

## Primary interactions and runtime checks

- 组件测试验证九张卡片均保存当前入口并通过 `moduleRoute` 打开真实页面。
- 微信开发者工具最终页为 `pages/functions/index`，没有业务代码错误。
- 完整 Vitest：`120 passed | 1 skipped`，`762 passed | 3 skipped`。
- 生产构建：`Compiled successfully`；WXSS 兼容处理成功。

## Follow-up polish

- P3：实现图的主插画青色轮廓光略亮于参考图；不影响文字可读性、结构或操作，本轮保留。

final result: passed

---

# 任务中心 Design QA

- source visual truth: `F:/脚本测试/一战晟铭-v88/docs/design-references/miniprogram-functions-first/03-tasks.png`
- implementation screenshot: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-tasks-implementation.png`
- normalized reference: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-tasks-reference.png`
- side-by-side comparison: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-tasks-comparison.png`
- viewport: WeChat Developer Tools, iPhone 15 simulator, `363 x 783` desktop-stage crop
- state: 已登录、批量工厂返回 6 条真实记录、Agent 来源未同步、三类统计均为服务端真实的 0

## Findings

本轮没有剩余 P0、P1 或 P2 差异。

- 顶部、统计栏、筛选栏和任务行均按 852px 参考稿折算到约 390px 手机宽度，不再直接照搬大画布固定尺寸。
- 一屏可见 6 条真实任务，任务行 76px、筛选栏 42px、统计栏 72px，内容密度接近参考稿。
- 双龙插画使用独立透明素材，文字、按钮和任务列表仍是可交互原生组件，不是整张截图。
- 当前 6 条记录都来自批量工厂，因此使用同一批量测书图标；没有为了模仿参考图伪造成不同任务类型。
- 服务端未返回可识别状态时显示“状态同步中”，没有伪造参考稿里的进度、完成或失败结果。

## Runtime and regression history

1. P0 黑屏：动态玻璃画布的层级覆盖了任务内容；任务头部和主体增加明确层级后，微信运行时恢复显示。
2. P2 尺寸偏大：统计栏从 106px 缩到 72px，筛选栏从 52px 缩到 42px，任务行从 102px 缩到 76px。
3. P2 统计图标像白色药丸：把渐变底座与图标图片拆成两个组件，恢复参考稿的独立发光图标结构。
4. P2 顶部缺插画：接入双龙透明插画并用裁切容器控制手机端可视占比。
5. 微信开发者工具最终页为 `pages/tasks/index`，页面显示真实任务与真实部分来源异常，没有业务代码错误。

## Required fidelity surfaces

- Typography: 主标题 27px、任务标题 14px、辅助信息 10px；没有再次放大成桌面比例。
- Spacing: 页面左右 13px，主体间距 10px，列表采用单一连续玻璃面板。
- Status: 三项计数、四项筛选、进度条和状态均由真实字段驱动；无效 total 不显示进度。
- Navigation: 底部保持“功能、工作台、任务、我的”四项，任务为第三项高亮。

## Follow-up polish

- P3：真实运行态多了一张 Agent 未同步提示卡，参考稿没有该异常状态；这是有用的真实错误信息，不隐藏也不伪装成功。

final result: passed

---

# 工作台执行中心 Design QA

- source visual truth: `F:/脚本测试/一战晟铭-v88/docs/design-references/miniprogram-functions-first/02-workbench.png`
- implementation screenshot: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-workbench-implementation.png`
- normalized reference: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-workbench-reference.png`
- side-by-side comparison: `F:/脚本测试/一战晟铭-v88/miniprogram/qa-workbench-comparison.png`
- viewport: WeChat Developer Tools, iPhone 12/13 simulator, `363 x 786` desktop-stage crop
- state: 已登录、当前选择“漫剧创作”、没有可继续的真实任务、Agent 数据源暂未同步

## Findings

本轮没有剩余 P0、P1 或 P2 差异。

- 页面已从功能入口集合改为单一执行中心，不再重复展示功能卡片。
- 当前工作区、主操作、项目进度和本次执行沿用参考图的信息层级。
- 真实账号当前没有可继续任务，因此实现图显示空状态；没有伪造参考图中的项目名、68%、13 页、模型或画幅。
- 顶部品牌栏和微信安全区来自真实小程序运行环境，参考稿未完整表达这部分，属于可接受的平台差异。

## Required fidelity surfaces

- Typography: 标题、当前功能、状态、按钮和说明形成五级字号；小屏没有大字挤压或截断。
- Spacing: 390px 基准左右 15px，卡片间距 16px，底部为四项固定导航并保留安全区。
- Cards: 当前工作区 128px，真实项目卡 205px；空状态只在没有真实任务时出现。
- Colors and material: 深海军蓝背景、青色描边、克制的玻璃内高光和单一青色主按钮与参考一致。
- Data integrity: 进度只在服务端提供有效 total 时显示；当前功能只匹配自己的任务，不跨功能串记录。

## Runtime and regression history

1. P0 黑屏：确认是微信开发者工具仍引用旧页面分包，而磁盘已是新运行时代码；完整退出、重新构建并打开后恢复。
2. P1 跨功能记录：建立功能与真实任务来源的明确映射，未映射的漫剧功能不会误显示剧本任务。
3. P1 示例数据：移除固定 68%、示例项目名和示例模型；无真实数据时进入空状态。
4. P2 重复入口：删除工作台里的功能目录和选择弹层，只保留一个主操作。
5. 微信开发者工具最终页面为 `pages/workspace/index`，控制台 0 个业务错误；仅有基础库弃用和本地证书类警告。

## Follow-up polish

- P3：空状态比参考稿的真实运行态留白更多；这是数据状态差异，待用户产生任务后会自动替换为项目进度和本次执行卡，不用假数据补满。

final result: passed
