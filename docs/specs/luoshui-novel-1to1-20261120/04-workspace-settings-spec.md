# 04 工作区、设置和辅助页面

来源均为当前bundle的`source/frontend/*.readable.js`，行号指恢复片段。每项将“UI行为”和“原生执行”分开。

| 编号/功能 | 输入 | 输出 | 边界 | 失败/取消 | 原码证据 |
|---|---|---|---|---|---|
| W01 两列创作入口及作品/会话区 | books、modeSessions、projectSessions、过滤项 | 六入口、作品树、会话、数量、运行标记 | 作品与会话分开；新会话不等于新作品；列表滚动与底部工具区分开 | 空列表不补假项；删除前确认 | `NovelStudioSidebar.readable.js:6–115,285–444` |
| W02 写作工具区 | activeTool/模式 | 风格/导入/Agent/调度/看板/备忘录导航 | 是工具导航，不是六个生成入口；底部布局由CSS决定 | 返回对话保留何种状态须实测 | `NovelStudioSidebar.readable.js:95–115`、`NovelStudioPage.readable.js:1537–1556` |
| W03 紧凑对话输入区 | message、selectedModel、上下文、thinking | send事件；发送/生成时取消入口 | Enter发送、Shift+Enter换行；无效输入禁用；不能在输入区额外塞进原版弹窗字段 | 当前回复未完提示等待；复制失败单独报错 | `NovelStudioChat.readable.js:47–83,229,267–289,650–804` |
| W04 模型选择 | modelOptions、搜索词、当前模型 | 选择模型/渠道、思考强度相关值 | 模型来自配置；原版模型名不能硬塞CM；无匹配显示空态 | 模型失效不能假选成功；实际上游参数支持待核验 | `NovelStudioChat.readable.js:195–229,735–804` |
| W05 上下文长度 | 100k/200k/300k/400k/1m/无限档，默认300k | contextLength随发送载荷提供 | 页面称无限不等于上游无限；清除上下文和删除作品不同 | 清空动作/本地会话与服务端消息一致性待测 | `NovelStudioChat.readable.js:63–82,662–712`、`NovelStudioPage.readable.js:439–464` |
| W06 回复和任务卡 | 消息、reasoning、toolSteps、action | 思考展开、工具步骤、复制、存档、删除、确认/关闭、重试 | 普通回复、操作确认、生产检查点是不同卡；不能统一当“下一步” | 删除有确认；取消不能显示任务完成 | `NovelStudioChat.readable.js:331–638`、`NovelConfirmCard.readable.js:1`、`NovelCheckpointCard.readable.js:1` |
| W07 写作风格库 | 名称、参考文本、文件、可选风格指南 | 保存/更新风格，多选应用到对话 | 名称必填；至少文本或文件；指南留空说明为自动分析，实际收费/算法调用待核验 | 保存失败/导入失败/删除失败分别提示；删除需原确认 | `NovelStylePanel.readable.js:60–94,130–190,209–280` |
| W08 文风清除 | 当前选中风格ID | 清除参考，刷新选择 | 不删风格库原件 | 界面提示“已清除写作风格参考”；持久化待测 | `NovelStudioPage.readable.js:205–222` |
| W09 章节与核心资料面板 | book、chapters、角色、文件、世界观 | 展开、查看、编辑、保存；部分文件保存同步truth | 正文、角色、世界观和规则不混成一份文件；无资料显示空态 | “保存失败”；成功分“已保存”和“已保存并同步真相” | `NovelStudioBookPanel.readable.js:145–190,207–337` |
| W10 短篇资料 | shortStory、文件、章节、大纲、推荐 | 文件编辑和保存，成品查看 | 草稿/进行中/已完成按对象状态；不能凭推荐文本判成品完成 | 保存失败有提示；编辑器另有saving结束即退出编辑的分支，不能据此判保存成功 | `NovelStudioShortPanel.readable.js:55–62,140–272` |
| W11 单章快捷操作 | bookId、sessionId、model、章节 | 写下一章、审稿、修订 | 审稿应消费实际章节，不是只展示维度；生成中禁用互斥操作 | 写章/审稿/修订各自message回退 | `NovelQuickActions.readable.js:1`、`NovelStudioPage.readable.js:1070–1305` |
| W12 批量写章 | 默认5章、书籍/会话/模型 | batch_progress事件、current/total、章号、stoppedReason | 事件需匹配bookId；结束清理监听；不能把局部进度当整批完成 | stoppedReason显示；错误不能伪造100% | `NovelBatchDialog.readable.js:14–99` |
| W13 作品/会话重命名 | 对象ID、新标题 | 保存后重新加载 | 重命名对象类型明确，不能改错会话/作品 | 分别“重命名失败” | `NovelStudioPage.readable.js:1400–1438,1698–1705` |
| W14 删除作品/会话 | 已确认的对象ID | 删除并刷新列表 | 确认/取消分支；当前对象删除后的选择恢复待测 | 删除失败保留原列表；不可静默跳过确认 | `NovelStudioPage.readable.js:1373–1394,1439–1460` |
| W15 导出 | 作品ID；md/txt/epub | 桥接结果content、filename，浏览器Blob下载 | 格式mime不同；EPUB字节/内容编码是否正确待实测，不能用txt改扩展名 | message或“导出失败”；取消格式选择不调用 | `NovelStudioPage.readable.js:1318–1363` |
| W16 存档到备忘录 | 会话/回复文字 | 备忘录对象、存档成功提示 | 不等于新作品、不自动成为参考资料；下次需显式选择 | 失败保留原消息 | `NovelStudioPage.readable.js:1244–1255` |
| W17 定时写作 | bookId/sessionId/model；默认间隔30分钟、每日10章、连续失败3次 | running/paused/stopped；累计/今日/失败次数 | intervalSeconds=分钟×60；10秒状态轮询；离开清计时器；暂停和停止不同动作 | 启动失败有提示；暂停/停止失败路径不能猜成功 | `NovelSchedulerPanel.readable.js:12–40,62–180,185–235` |
| W18 数据看板 | 实际章节集合 | 总章数、总字数、均字数、状态分布、趋势 | draft/通过/待审/未过/降级映射；空章集合显示暂无数据 | 未见独立后端请求；不能用静态示例替真实数据 | `NovelDashboard.readable.js:18–63` |
| W19 伏笔 | hooks、currentChapter | 活跃/已解决/全部筛选；卡片 | planted/advancing为活跃；距上次推进≥8警告、≥12错误色；已解决为成功色 | 空集合“暂无伏笔记录”；此组件未提供新增/删除事件，不添加假按钮 | `NovelHooksPanel.readable.js:17–63` |
| W20 封面预览 | 封面图片/提示词对象 | 查看资源或生成入口 | 封面图片与LLM文本模型不同路由 | 未实跑图像，不登记生成通过 | `NovelCoverPreview.readable.js:1`、`NovelStudioPage.readable.js:1041–1054` |

## S01 Agent设置：真实可保存字段

来源：`NovelAgentSettings.readable.js:19–45,138–179,250–269,280–475`。

| 字段 | 默认/范围 | 保存输出 | 必须区分的行为 |
|---|---|---|---|
| modelRoutes | 各节点可留空 | 节点key→模型标识 | 空值跟随对话默认；封面配图有独立路由；节点key完整动态列表待解码 |
| maxRetries | 3；0–5 | reviewCycle.maxRetries | 保存表达式是`Number(value) || 3`再裁剪：输入0会变回3，属于原版前端缺陷，不能声称0成功禁用循环 |
| passScore | 80；0–100 | reviewCycle.passScore | 同样`||80`导致0不能原样保存 |
| temperatureStep | .15；0–.5 | reviewCycle.temperatureStep | `||.15`导致0回默认；升温只在实际循环使用才有意义 |
| baseTemperature | .65；0–2 | reviewCycle.baseTemperature | `||.65`导致0回默认；不是所有模型都支持温度 |
| enableLengthNorm | true | 布尔值 | 界面说明偏离目标>45%扩写/压缩；真实调用/阈值以原生函数为准 |
| enablePostWrite | true | 布尔值 | 确定性检查，不是另一次收费审稿 |
| enableAiTells | true | 布尔值 | 段落等长、套话、公式化转折检查；不同于审稿37维 |
| requireSecondaryConfirm | false | 布尔值 | 大纲/蓝图/基础设定检查点暂停；不是每次模型请求弹一次确认 |
| reviewPrompts | 内置37项及自定义审稿 | map整理后的条目 | 内置名称说明不可改；可全选/全不选；勾选应进入审稿载荷 |
| customWritePrompts | 长篇正文/短篇小说/仿写创作分类 | id/category/title/content/enabled | 只注入匹配类别，不等于自动追加到所有请求 |

失败行为：加载失败“加载 Agent 设置失败”；保存失败返回message或“保存失败”；成功用返回配置重新回填并提示“已保存”。设置页有保存按钮，不能只保存在前端临时状态。

## S02 37维审稿完整清单

见[审稿维度原文](source/review-dimensions.md)。源为原生已恢复目录`docs/reverse-engineering/novel-review-dimensions-20261119/catalog.json`，其provenance单列。它证明37项条目，不单独证明当前前端加载/模型实际消费已成功。实际闭环需检查：勾选→保存→重载→审稿系统提示→审稿结果→修订消费。

## S03 UI和icon不能靠文案对齐

原图标SVG分支见`NovelCreateIcon.readable.js`；不能用任意相似图标替换后标成1:1。当前CSS原始文件另固定哈希，提取小说相关规则到`source/novel-styles.css`。颜色可适配CM统一主题，但布局、尺寸、入口层级、弹窗字段、控件状态必须逐项保留。静态CSS不证明实际渲染一致，像素对照与真实点击仍在07的未验收项。

## W21 文件详情编辑器的精确状态

输入：title/content/loading/saving/readOnly/autoEdit；输出back和save(草稿)。开始编辑拷贝content到草稿；取消重新拷贝原content并退出；保存只向父组件发草稿。`content===null`显示文件不存在，空字符串显示空文件。readOnly/loading/null内容时不提供编辑。saving=true时保存禁用。

原版边界：监听saving从true变false就退出编辑，**不检查保存是否成功**；父页面失败提示和编辑状态可能脱节，不能把它描述成“失败一定保留编辑态”。自动编辑还检查autoEdit/loading/readOnly。来源：`NovelPanelArtifactView.readable.js:6–45,53–102`。这属于原版可见静态分支，是否实际表现为丢草稿仍需失败请求实测。

## W22 检查点卡与导入壳

- 检查点输入checkpoint/loading；blueprint标题“确认创作蓝图”，foundation为“确认基础设定”，其他“确认继续生成”；确认发confirm，“稍后”发dismiss，loading时两者禁用。dismiss不等于取消后台任务。来源：`NovelCheckpointCard.readable.js:6–41`。
- 导入弹窗输入show/tab/sessionId/styleSources/books/isDarkMode，透传给导入页；created后关闭；mask-closable=true，auto-focus=false。遮罩关闭时是否存在已提交任务由父层处理，不能用关闭行为代替取消。来源：`NovelImportModal.readable.js:6–61`。

## W23 每个组件的参数与事件

23组件逐项输入属性、输出事件、禁用表达式和失败文案原位置见[组件合同附录](08-component-contracts.md)。动态属性和间接调用保留原表达式，不擅自翻译成未经证实的功能。
