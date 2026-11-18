# 小说创作完整差异审计与执行登记（2026-11-18）

结论：不是一个弹窗没做好，而是六种创作被压进同一条“作品→会话→通用章节生产”流程，导致入口、参数、动作确认、资料形态和输出都系统性偏离。当前不能宣称已一比一复刻。

## 本轮实际完成与边界

实际通过Craft Anything 0.5.0 CLI，对原版23个Novel命名组件和当前10个对应React组件逐文件运行`js-app-map`，另做14个重点函数的`js-function-explain`。恢复后的原包中23个Novel组件与本轮目录逐一对应，遗漏的Novel命名组件为0。这不覆盖全软件的其他模块、原生后端全部函数或动态生成组件。

原包可读恢复证据：`ev_6360e90b82e1fa629455cdfcc39d51638dda68618f15c5db900334c9af0cfaf8`；原bundle SHA256为`5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a`。组件图和重点函数证据见下表、`components.json`与归档`function-probes.json`。原23组件是带原hash/UTF16范围的检查材料，不是完整开发源码。Go代码直接阅读并记录摘要，不把JS解析器当Go逆向器。

本轮只做静态审计、来源核查与规格登记：模型调用0、原/CM逐按钮实际执行0、同状态原版截图0、产品修改0、运行包重启0。不将已枚举点击或代码里有函数算成已运行通过。原版native、私有系统提示词、模型调用真实次序与费用仍有未知。

## 检查结果（计数是当前检查清单，不是全部软件缺陷数）

| 状态 | 项数 | 含义 |
|---|---:|---|
| 合同不同 | 18 | 已有相似实现，但输入、默认值、动作、作用域或输出不同 |
| 缺失 | 28 | 在已检查的对应小说UI/链路中没有同等入口或产物 |
| 部分实现 | 14 | 真实能力有一部分，原入口与状态仍不齐 |
| 已有实现待对照 | 10 | 已有实际实现/既往测试；本轮不冒称原版同状态验收通过 |
| 尚未验证 | 5 | 原native或运行/像素证据不足，不能补成已完成 |

共75项检查。另枚举原版133处、当前64处`onClick`声明，见`controls.json`；这些是声明位置，不是运行按钮数量，也不是缺失按钮数。循环渲染、动态属性及非点击交互仍需逐状态确认。

## 根因与纠偏

1. **工作流对象错配。** 原版聊天可先无书，建议动作待确认；长篇书籍与短篇story分别保存。当前先强制作品，再让六模式走同一章链。缺的是对象和状态合同，单加表单不能解决。
2. **参数接收面过窄。** 仿写/番外专用字段没有DTO，`NormalizeModeInputs`只保留少数文本键；即使添加按钮，参数仍可能被丢弃。
3. **产物读写面不足。** 当前章正文/版本是真功能，但没有原角色、核心文件、世界观、封面、短篇大纲/成品/推荐等独立读写合同。
4. **两套确认混淆。** 建议动作确认、内容里程碑确认与逐模型请求批准不是一个东西；可靠性token可以保留，用户含义不能替换。
5. **提示词与模型分工尚未配对。** 原17配置槽不能直接等同CM8个生产步骤；原native所有提示词和调用次数未恢复，不得写成“已全部接进来”。自家模型选择必须保留。
6. **验收方式不足。** 自己实现的单元/浏览器测试通过，只证明当前规则工作；需要原版同页面/同状态/同输入的对照，才能证明复刻。

## 已完成能力不能重复判缺

左下六工具位置、主区工具页返回、导入弹窗、独立备忘录弹层已经修过；续写正式导入与幂等恢复存在；章节编辑/版本、持久派发、未知结果保护也是真实现。上一批自定义审稿/分类写作要求、本地后写检查和>45%长度门槛已补。旧报告中这些“全部没有”的说法不再作为当前结论。

## 逐项登记

### 六种创作

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| E01 | 长篇原创首次发送与建书确认 | 合同不同 | 原版允许先建无书会话，聊天返回创建作品建议并确认；CM先要求新建作品，随后直接进入通用章链。 | 原 [NovelStudioPage.readable.js:420](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioPage.readable.js:420)；CM [NovelCreationPage.jsx:95](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:95)；证据 `ev_ab1ec87d7b639d8eed18d18b284304d936f4768662c24fb4262eae196ac6a389` | 空作品状态发创作意图→建议动作→确认建书→会话归入作品；取消不得生成。 |
| E02 | 短篇原创专用产物流程 | 合同不同 | 原版短篇大纲/正文/审核/打包与独立storyId；CM短篇结果是一个ChapterVersion。 | 原 [NovelConfirmCard.readable.js:37](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelConfirmCard.readable.js:37)；CM [exports.go:101](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/exports.go:101)；证据 `ev_102cdd3e022c3d3bcfaa326b4584ea6a3abb5ec4981e5b4c2f0ea3245355dcf6` | 以独立短篇结构保存大纲、正文和成品；重新打开与导出仍是同一结果。 |
| E03 | 仿写长篇专用导入窗 | 缺失 | 原版进入专用仿写表单；CM只切模式和通用参考文本框。 | 原 [NovelImportPage.readable.js:130](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:130)；CM [NovelCreationPage.jsx:266](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:266)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 点击仿写长篇打开对应表单，取消返回原草稿；控件顺序和图标对照截图。 |
| E04 | 长仿写参数与参考规则 | 合同不同 | 原版每章字数、目标章数、storyIdea、偏好、选定方向、补充信息等；CM白名单仅referenceText。 | 原 [NovelImportPage.readable.js:375](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:375)；CM [validation.go:83](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/validation.go:83)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 逐字段保存并传入正确步骤；空/99/100字边界与原文分块不能被静默截断。 |
| E05 | 仿写短篇专用弹窗 | 缺失 | 专用弹窗缺失；不能用Agent设置页或通用参考框代替。 | 原 [NovelImitationShortDialog.readable.js:181](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImitationShortDialog.readable.js:181)；CM [NovelCreationPage.jsx:266](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:266)；证据 `ev_b426eb57fcb7d99a020011ddc75881f8cc0983a7ad3063cf29a21d41f1d42395` | 标题、上传、字数快捷值、偏好、标签、备忘录、关联作品及取消均有真实动作。 |
| E06 | 短仿写总字数和篇章数 | 合同不同 | 原版总字数13000默认、篇章1；CM默认每章2000、章字数最多8000，两者单位与范围不同。 | 原 [NovelImitationShortDialog.readable.js:153](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImitationShortDialog.readable.js:153)；CM [validation.go:10](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/validation.go:10)；证据 `ev_b426eb57fcb7d99a020011ddc75881f8cc0983a7ad3063cf29a21d41f1d42395` | 分别定义总字数/每章字数，不通过放大单章上限伪装专用短篇合同。 |
| E07 | 番外正传关联与四类型 | 缺失 | 原版有parentBookId和sequel/prequel/parallel/if分支；CM番外标签实际用fanfiction和一段existingSetting。 | 原 [NovelImportPage.readable.js:437](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:437)；CM [NovelCreationPage.jsx:283](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:283)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 可选正传并保留关联ID；每种番外类型加载相应规则，不仅拼一段世界设定。 |
| E08 | 同人与番外语义区分 | 合同不同 | 原导入器有fanfic和spinoff两个分支；CM把番外入口映射为fanfiction，模式合同仍叫同人。 | 原 [NovelImportPage.readable.js:116](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:116)；CM [contracts.js:3](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/contracts.js:3)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 入口名称、请求模式、上下文和输出归属一致，既有作品数据不能错归类。 |
| E09 | 续写导入正式章节 | 部分实现 | 已存在专用续写导入、目标书/新书、分章、多文风选择和正式保存；不能继续列为完全缺失。原版文件选择和形态仍未同状态验收。 | 原 [NovelImportPage.readable.js:461](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:461)；CM [ContinuationImportDialog.jsx:36](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ContinuationImportDialog.jsx:36)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 粘贴/上传→预览分章→目标书→一次保存→关闭重开；重复请求不重复入库。 |
| E10 | 无书会话与正式作品的关系 | 合同不同 | 原版存在独立会话、建书后归档到书籍；CM普通会话和工具普遍要求project。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [service.go:78](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/service.go:78)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 无书讨论与有书生产分别建模，创建失败或取消不得生成空作品。 |

### 布局与导航

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| L01 | 六入口排列与模式图标 | 合同不同 | 原版截图两列带模式图标；CM单列文字按钮，未使用原模式图标组件。 | 原 [NovelStudioSidebar.readable.js:250](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:250)；CM [NovelCreationPage.jsx:266](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:266)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 在约定主题下逐入口对照位置、图标语义、尺寸、选中态、悬停及键盘焦点。 |
| L02 | 作品类型筛选 | 缺失 | 原版有全部/类型筛选并保留选择；CM只区分归档与未归档。 | 原 [NovelStudioSidebar.readable.js:50](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:50)；CM [NovelCreationPage.jsx:259](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:259)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 切换筛选后作品和会话归属正确，刷新保留筛选且不创建任务。 |
| L03 | 每本书展开嵌套会话 | 合同不同 | 原版作品展开显示其会话并能在书下新建；CM扁平作品列表+当前模式会话。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [NovelCreationPage.jsx:269](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:269)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 逐书展开/折叠/切会话，不互相覆盖，保留未发草稿。 |
| L04 | 会话相对时间与运行标记 | 部分实现 | CM有当前运行标记，但无原列表相对更新时间样式与完整组别。 | 原 [NovelStudioSidebar.readable.js:351](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:351)；CM [NovelCreationPage.jsx:269](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:269)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 时间、运行中、选中态与书籍关系能在刷新后正确恢复。 |
| L05 | 作品与会话重命名 | 已有实现待对照 | 当前有真实重命名入口；交互位置/弹窗及空值边界尚未与原版同状态比较。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [NovelCreationPage.jsx:269](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:269)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 空名拒绝、取消不改名、失败保留输入、重复点击不重复提交。 |
| L06 | 作品快捷导出 | 部分实现 | CM有正文导出，但没有原侧栏同位置动作和全部产物合同。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [ChaptersPanel.jsx:28](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:28)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 侧栏导出应指向选中作品，格式与保存内容一致；空作品明确提示。 |
| L07 | 删除与可恢复归档 | 合同不同 | CM采用可恢复归档，原版有删除动作。不得擅自永久删除来追求界面一致。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [NovelCreationPage.jsx:268](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:268)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 明确保留/删除策略、文案与恢复动作；归档后自动任务保持暂停。 |
| L08 | 六工具进入位置与返回 | 已有实现待对照 | 已修成左下工具、主区工具页、导入弹窗与独立备忘录弹层；旧报告统一Drawer结论过时。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [NovelCreationPage.jsx:251](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:251)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 逐工具进出、返回、切作品、关弹窗后保留原会话草稿；再做原版同状态截图。 |

### 聊天与动作

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| C01 | 聊天发送与快捷键 | 已有实现待对照 | CM有真实发送/Enter/Shift+Enter和中文输入保护；不代表原版会话/动作解析相同。 | 原 [NovelStudioChat.readable.js:248](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:248)；CM [NovelCreationPage.jsx:284](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:284)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 中文组合输入不误发、连续点击不重复提交、失败保留草稿。 |
| C02 | 可调上下文长度 | 缺失 | 原版有上下文长度控件；CM会话预算写死120000，没有用户调整入口。 | 原 [NovelStudioChat.readable.js:253](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:253)；CM [NovelCreationPage.jsx:168](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:168)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 按模型能力限定预算，修改值真正进入请求及冻结快照。 |
| C03 | 清除上下文与Ctrl+K | 缺失 | 当前聊天区没有原清上下文按钮/动作；不能以新建会话代替。 | 原 [NovelStudioChat.readable.js:48](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:48)；CM [NovelCreationPage.jsx:281](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:281)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 清除只影响约定范围，正文作品不丢；取消/失败不清会话。 |
| C04 | 回复复制 | 缺失 | 当前消息只有角色和文本，无复制控件。 | 原 [NovelStudioChat.readable.js:286](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:286)；CM [NovelCreationPage.jsx:278](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:278)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 复制完整对应回复；失败提示且不改正文。 |
| C05 | 消息存档 | 缺失 | 原消息可存档；当前只有作品/会话归档，不是同一种动作。 | 原 [NovelStudioChat.readable.js:48](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:48)；CM [NovelCreationPage.jsx:278](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:278)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 保存选定消息及来源，不能误归档整本或创建生成任务。 |
| C06 | 单条消息及配对回复删除 | 缺失 | 原区分用户消息及对应回复/单AI回复；CM消息没有删除入口。 | 原 [NovelStudioChat.readable.js:48](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:48)；CM [NovelCreationPage.jsx:278](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:278)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 必须先明确可恢复策略；不能混淆删除对话与删除正式章节。 |
| C07 | 流式正文展示 | 缺失 | 当前只渲染snapshot.messages及RunCard，没有原delta正文流显示合同。 | 原 [NovelStudioPage.readable.js:926](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioPage.readable.js:926)；CM [NovelCreationPage.jsx:278](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:278)；证据 `ev_ab1ec87d7b639d8eed18d18b284304d936f4768662c24fb4262eae196ac6a389` | 分会话接流、重连补齐、切会话不串文；保存结果与最后文本一致。 |
| C08 | 思考块展开与耗时 | 缺失 | 参数详情不是原思考过程块，当前没有该折叠/耗时显示。 | 原 [NovelThinkingBlock.readable.js:49](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelThinkingBlock.readable.js:49)；CM [RunCard.jsx:22](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/RunCard.jsx:22)；证据 `ev_fd0f8e9cec3f4907363e4066e0676f2c8b9742aacaea5ebb840d9449605e8c1b` | 只显示提供方真实返回的思考字段；未提供时不捏造内容或耗时。 |
| C09 | 模型与思考强度选择 | 部分实现 | 已有自己的模型目录与思考选项；这是用户要求保留的适配，原搜索/分组/能力控制仍待对照。 | 原 [NovelStudioChat.readable.js:231](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioChat.readable.js:231)；CM [NovelCreationPage.jsx:286](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:286)；证据 `ev_c8f870165cdb6b93d6ea769b6f375f89f2bc463ef2bca39ad8483107ba6c7b33` | 所有选项来自自家已配置模型；不显示不可用的原渠道或无效思考参数。 |
| C10 | 建议动作确认卡 | 缺失 | 原聊天返回action，确认后才建书/写章等；当前运行蓝图确认不是这套待执行动作卡。 | 原 [NovelConfirmCard.readable.js:33](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelConfirmCard.readable.js:33)；CM [RunCard.jsx:18](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/RunCard.jsx:18)；证据 `ev_102cdd3e022c3d3bcfaa326b4584ea6a3abb5ec4981e5b4c2f0ea3245355dcf6` | 建议→显示参数→确认或取消；刷新恢复同一待执行动作，防重复。 |
| C11 | 快捷写下一章 | 部分实现 | 当前可通过普通发送写章，但缺少原同位置快捷动作及相同确认语义。 | 原 [NovelQuickActions.readable.js:11](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelQuickActions.readable.js:11)；CM [NovelCreationPage.jsx:266](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:266)；证据 `ev_abaf9c7dc782cc4887b11db79979e7841e5710eeadbc53d0ac7f746eada9deca` | 对当前作品最后章节写下一章，空作品和运行中状态正确处理。 |
| C12 | 独立手动审稿 | 缺失 | 自动生产中审核存在；当前维护只开放文风分析/事实刷新，未见对应手动审稿入口。 | 原 [NovelQuickActions.readable.js:11](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelQuickActions.readable.js:11)；CM [maintenance.go:48](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/maintenance.go:48)；证据 `ev_abaf9c7dc782cc4887b11db79979e7841e5710eeadbc53d0ac7f746eada9deca` | 选定已保存章节→独立审核→保存报告，不隐式生成新正文。 |
| C13 | 独立手动修订 | 缺失 | 失败重试不是对已保存章节主动修订；缺同位置选择/执行/版本结果链。 | 原 [NovelQuickActions.readable.js:11](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelQuickActions.readable.js:11)；CM [RunCard.jsx:19](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/RunCard.jsx:19)；证据 `ev_abaf9c7dc782cc4887b11db79979e7841e5710eeadbc53d0ac7f746eada9deca` | 指定章节和审核问题修订，保留旧版本，取消不覆盖正文。 |
| C14 | 原工具阶段与日志 | 部分实现 | CM有持久run/百分比/错误，原按tool动态阶段与日志卡没有完全对应。 | 原 [NovelToolSteps.readable.js:7](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelToolSteps.readable.js:7)；CM [RunCard.jsx:1](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/RunCard.jsx:1)；证据 `ev_d37ad7f377c0dda693b51bae7c51460cdc1fe988cd4c10975fb03a085b9dfe96` | 每个模式显示实际执行阶段，不填虚构步骤或把11阶段当11次外呼。 |
| C15 | 停止、失败重试、结果未知 | 已有实现待对照 | 当前有停止后续步骤、原任务恢复、结果未知不自动重试；这是已有真实保护，原native细节尚未知。 | 原 [NovelToolSteps.readable.js:203](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelToolSteps.readable.js:203)；CM [RunCard.jsx:14](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/RunCard.jsx:14)；证据 `ev_d37ad7f377c0dda693b51bae7c51460cdc1fe988cd4c10975fb03a085b9dfe96` | 运行中停止、已发请求、丢回包、保存失败分别验收；禁止重试造成重复收费。 |

### 资料与产物

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| V01 | 已保存章节选择/编辑/版本 | 已有实现待对照 | 当前能编辑正文、版本历史与恢复；不应列为无功能，列表/编辑形态仍不同。 | 原 [NovelStudioBookPanel.readable.js:225](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioBookPanel.readable.js:225)；CM [ChaptersPanel.jsx:32](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:32)；证据 `ev_57fcade461d98616e319bdeaf303708e9edb8790a5cec2879bd6dfaeb5b50d77` | 保存新版本→刷新读回→恢复旧内容生成新版本，冲突保留草稿。 |
| V02 | 角色档案列表 | 缺失 | 当前右栏仅正文列表，缺原角色主次分类及档案浏览。 | 原 [NovelStudioBookPanel.readable.js:261](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioBookPanel.readable.js:261)；CM [ChaptersPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25)；证据 `ev_57fcade461d98616e319bdeaf303708e9edb8790a5cec2879bd6dfaeb5b50d77` | 从真实生成/保存的人物数据读取，空态明确，不用示例角色充数。 |
| V03 | 核心文件与文件浏览 | 缺失 | 缺核心文件树/文件详情，现有蓝图弹窗不能覆盖全部文件。 | 原 [NovelStudioBookPanel.readable.js:287](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioBookPanel.readable.js:287)；CM [ChaptersPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25)；证据 `ev_57fcade461d98616e319bdeaf303708e9edb8790a5cec2879bd6dfaeb5b50d77` | 列目录→读取→编辑→保存→重新打开，路径与所属作品校验。 |
| V04 | 世界观摘要与完整设定 | 缺失 | 当前项目创建表单的world字段不等于原右栏摘要和完整设定文件。 | 原 [NovelStudioBookPanel.readable.js:58](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioBookPanel.readable.js:58)；CM [ChaptersPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25)；证据 `ev_57fcade461d98616e319bdeaf303708e9edb8790a5cec2879bd6dfaeb5b50d77` | 摘要由实际设定读取，完整文件可编辑并供下一阶段使用。 |
| V05 | 封面资源与预览 | 缺失 | 当前小说资料区未提供原封面资源列表/预览与加载失败状态。 | 原 [NovelCoverPreview.readable.js:100](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelCoverPreview.readable.js:100)；CM [ChaptersPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25)；证据 `ev_211c796a4afc9bff054886de1e5c5d45fcd6097d3c4da13cc7db49058205eca2` | 只显示当前作品真实封面；无图/坏图/多图可辨识。 |
| V06 | 生成封面专用动作 | 缺失 | 原有封面确认动作和独立图像模型；当前只列文本步骤，未接对应功能。 | 原 [NovelConfirmCard.readable.js:39](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelConfirmCard.readable.js:39)；CM [AgentPanel.jsx:34](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AgentPanel.jsx:34)；证据 `ev_102cdd3e022c3d3bcfaa326b4584ea6a3abb5ec4981e5b4c2f0ea3245355dcf6` | 自家图像模型选择→确认费用→生成→归档封面；无预算本轮不调用。 |
| V07 | 短篇章节结构 | 缺失 | 当前短篇只是通用章列表的一次结果，没有独立短篇章节结构。 | 原 [NovelStudioShortPanel.readable.js:193](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioShortPanel.readable.js:193)；CM [ChaptersPanel.jsx:7](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:7)；证据 `ev_2f63a1a6c767459a08e9d2564878542c7079a7af0b80d1f45a185411ac6f59ba` | 短篇多章结构、编号和字数来自同一story实体。 |
| V08 | 短篇大纲查看/编辑 | 缺失 | 长篇卷章蓝图不是短篇完整大纲文件；缺专用输出/编辑/继续。 | 原 [NovelStudioShortPanel.readable.js:153](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioShortPanel.readable.js:153)；CM [BlueprintEditor.jsx:13](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/BlueprintEditor.jsx:13)；证据 `ev_2f63a1a6c767459a08e9d2564878542c7079a7af0b80d1f45a185411ac6f59ba` | 大纲修改成功后继续必须读取保存版本，不用旧快照。 |
| V09 | 短篇成品文件 | 缺失 | 原artifacts列表与文件编辑缺失，现有一个ChapterVersion无法承载。 | 原 [NovelStudioShortPanel.readable.js:223](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioShortPanel.readable.js:223)；CM [exports.go:101](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/exports.go:101)；证据 `ev_2f63a1a6c767459a08e9d2564878542c7079a7af0b80d1f45a185411ac6f59ba` | 正文、大纲等成品分文件读取与导出，空结果不能标完成。 |
| V10 | 短篇推荐文案 | 缺失 | 缺推荐预览、完整推荐与保存字段。 | 原 [NovelStudioShortPanel.readable.js:53](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioShortPanel.readable.js:53)；CM [ChaptersPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25)；证据 `ev_2f63a1a6c767459a08e9d2564878542c7079a7af0b80d1f45a185411ac6f59ba` | 推荐和正文分开保存/查看，失败不得以空文案成功。 |
| V11 | 通用资料详情编辑器 | 缺失 | CM只有结构化蓝图和章节编辑，不能打开原任意业务资料文件。 | 原 [NovelPanelArtifactView.readable.js:63](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelPanelArtifactView.readable.js:63)；CM [BlueprintEditor.jsx:6](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/BlueprintEditor.jsx:6)；证据 `ev_dfb2b0f2e292cdea966ed54eab5f50caaef87632a64dc864d21c94652b297573` | 读取/编辑/取消/保存失败留草稿，保存后可重新打开验证。 |
| V12 | 编辑章节后真相同步与回放 | 合同不同 | CM修改后标记dirty再显式刷新事实记忆；原保存payload含syncTruth/replayFrom，原算法未知。 | 原 [NovelStudioBookPanel.readable.js:173](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioBookPanel.readable.js:173)；CM [ChaptersPanel.jsx:27](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:27)；证据 `ev_57fcade461d98616e319bdeaf303708e9edb8790a5cec2879bd6dfaeb5b50d77` | 明确自动/手动同步策略，后续创作不能读脏记忆；不宣称算法已复原。 |
| V13 | 资料栏显隐/宽度/拖动 | 部分实现 | 当前有可拖宽资料栏与按页显隐，不能据此说内容、默认宽度和像素全一致。 | 原 [NovelStudioPage.readable.js:137](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioPage.readable.js:137)；CM [NovelCreationPage.jsx:262](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:262)；证据 `ev_ab1ec87d7b639d8eed18d18b284304d936f4768662c24fb4262eae196ac6a389` | 同尺寸原/CM量测最小最大宽度、空态、工具页返回和主题适配。 |

### 文风与资料

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| S01 | 独立文风新增编辑表单 | 合同不同 | 原直接名称/原文/指南编辑；当前复用资料解析与单文风选择，流程不同。 | 原 [NovelStylePanel.readable.js:49](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStylePanel.readable.js:49)；CM [MaterialsPanel.jsx:7](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:7)；证据 `ev_a87188231852d6980f63d1260589f95f793326febff8d266061a54d60575ffaa` | 新建/编辑/取消/保存分别处理原文和可编辑指南，保存失败不清输入。 |
| S02 | 文风多选与应用到对话 | 合同不同 | 原数组选择并应用；CM只有单styleProfileId，不能以普通资料多选冒充。 | 原 [NovelStylePanel.readable.js:11](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStylePanel.readable.js:11)；CM [MaterialsPanel.jsx:7](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:7)；证据 `ev_a87188231852d6980f63d1260589f95f793326febff8d266061a54d60575ffaa` | 多文风选择、取消选择和应用真实进入请求，并保留顺序/来源。 |
| S03 | 指南为空时分析策略 | 合同不同 | 原UI声明空指南保存时自动分析；CM保存统计，AI分析单独点击。原native费用与算法未知。 | 原 [NovelStylePanel.readable.js:231](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStylePanel.readable.js:231)；CM [styles.go:49](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/styles.go:49)；证据 `ev_a87188231852d6980f63d1260589f95f793326febff8d266061a54d60575ffaa` | 明确是否调用模型及失败提示；未取得预算不偷偷自动分析。 |
| S04 | 原文风八维/统计指纹 | 合同不同 | 原Tln/qln静态合同已恢复，CM仍是runes、单换行分段与简单句长，不能称等价。 | 原 [NovelImportPage.readable.js:379](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:379)；CM [styles.go:29](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/styles.go:29)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 同一含emoji/空行/对话样本，两侧统计与八维字段逐值比较。 |
| S05 | 上传/粘贴资料真实解析 | 已有实现待对照 | CM有真实TXT/DOCX/PDF上传、解析状态与章节范围；原文件选择/追加和资料用途不同。 | 原 [NovelImportPage.readable.js:225](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:225)；CM [MaterialsPanel.jsx:31](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:31)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 空文件/坏文件/重复/部分失败/取消上传都验收；现有能力保留。 |
| S06 | 长参考分块与当前截取上限 | 合同不同 | 原仿写参考分块合同与CM每份选文最多12000字不同，不能静默丢后文。 | 原 [NovelImportPage.readable.js:380](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js:380)；CM [MaterialsPanel.jsx:43](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:43)；证据 `ev_44444b0ca9c841035b5ef993ec6135bee1ff91566401f9746aaa8115c6e9cc5e` | 明确长文本分块、来源、顺序和总上下文限制；超限有提示。 |
| S07 | 文风删除/编辑完整管理 | 部分实现 | CM选用和建立文风存在；没有同一完整管理列表中的全部编辑/删除/多选动作。 | 原 [NovelStylePanel.readable.js:191](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStylePanel.readable.js:191)；CM [MaterialsPanel.jsx:52](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:52)；证据 `ev_a87188231852d6980f63d1260589f95f793326febff8d266061a54d60575ffaa` | 关联中的文风删除后有明确失效提示，不能静默换模型或随机换风格。 |
| S08 | 备忘录弹层与选择 | 部分实现 | 独立弹层与CRUD已实现；原共享笔记/会话与当前按作品noteIds、正文拼接语义仍不同。 | 原 [NovelStudioSidebar.readable.js:41](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioSidebar.readable.js:41)；CM [NotesPanel.jsx:5](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/NotesPanel.jsx:5)；证据 `ev_77141d2f306fc5e57f52588a93a44ec8ad08f582af128b37f95e0d4312a96c16` | 切作品、选多笔记、编辑、取消、重复保存和来源追溯一致。 |

### Agent与提示词

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| A01 | 原17项模型分工 | 合同不同 | 原17配置包含聊天/编排/校验/卷压缩/互动/封面；CM仅8个实际收费步骤。不能添加空选项宣称接通。 | 原 [NovelAgentSettings.readable.js:212](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:212)；CM [AgentPanel.jsx:35](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AgentPanel.jsx:35)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 每项建立真实执行节点映射及输入输出；仍使用自家模型候选。 |
| A02 | 37项审核维度 | 缺失 | 自定义补充已补不等于内置37维、不可改说明、全选/全不选及勾选注入。 | 原 [NovelAgentSettings.readable.js:354](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:354)；CM [AgentRequirements.jsx:9](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AgentRequirements.jsx:9)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 原目录逐项有来源，启停真实影响审核；不捏造37条凑数。 |
| A03 | Agent保存范围 | 合同不同 | 原无作品/会话限定参数，CM按owner+project；原profile隔离物理实现仍未知。 | 原 [NovelAgentSettings.readable.js:463](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:463)；CM [service.go:658](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/service.go:658)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 自家账号隔离下定义共享默认/作品覆盖；先核对范围再迁移，不串用户。 |
| A04 | 补充审稿新增停用删除与注入 | 已有实现待对照 | 本批已真实补入reviewPrompts、校验、冻结和审核/修订注入，不再列为全缺。 | 原 [NovelAgentSettings.readable.js:172](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:172)；CM [agent_requirements.go:11](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/agent_requirements.go:11)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 新增/关/删/空/重复/保存失败/刷新，确认实际请求内容对应。 |
| A05 | 三类写作要求 | 已有实现待对照 | 已补分类与写作/修订注入；六模式本身不同仍会影响产物，不能以此证明模式完整。 | 原 [NovelAgentSettings.readable.js:127](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:127)；CM [agent_requirements.go:57](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/agent_requirements.go:57)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 长篇/短篇/仿写不串用、停用不传、格式修复继承同样要求。 |
| A06 | 审稿循环与升温算法 | 尚未验证 | UI数值与零值已支持；首轮偏移/同分回退/否决项/温度上限未证实与原native一致。 | 原 [NovelAgentSettings.readable.js:29](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:29)；CM [pipeline.go:529](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/pipeline.go:529)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 恢复或观察原规则后逐边界对照；不把当前有界循环写成原算法。 |
| A07 | 后写校验与跨章输入 | 部分实现 | 15规则本地实现已补，跨章完整正文/标题/人称等生产输入仍不齐；显式未检查不是已通过。 | 原 [NovelAgentSettings.readable.js:342](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:342)；CM [local_checks.go:30](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/local_checks.go:30)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 正常、重复、错标题、人称及跨章样本；不调用模型、不改正文。 |
| A08 | AI味检测 | 尚未验证 | 当前是明确标注CM自有统计，不是已恢复的原分类算法。 | 原 [NovelAgentSettings.readable.js:346](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:346)；CM [local_checks.go:310](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/local_checks.go:310)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 原算法/词表/边界未恢复前，保留真实指标但不得标原版一致。 |
| A09 | 长度归一化 | 部分实现 | 已补严格超过45%才触发模型；原计字、路由和扩缩策略还未全部恢复。 | 原 [NovelAgentSettings.readable.js:336](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:336)；CM [local_checks.go:112](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/local_checks.go:112)；证据 `ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311` | 44.9/45/45.1%、中英文emoji、扩写与压缩，核对是否触发和实际字数。 |
| A10 | 生产二次确认与文件重读 | 合同不同 | 原是大纲/蓝图/设定后暂停；CM还有逐模型请求确认，含义不同。 | 原 [NovelCheckpointCard.readable.js:22](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelCheckpointCard.readable.js:22)；CM [model_confirmation.go:17](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/model_confirmation.go:17)；证据 `ev_c04ce66b900162f7d9151b85fbc203b6e1612680f9ac491bb9a38305ee00a4f2` | 保存资料→继续读新版本；稍后、取消、重复继续各有明确结果。 |
| A11 | 原系统提示词与路由调用次序 | 尚未验证 | 当前后端提示词是本项目独立合同，原native完整提示词/实际角色次序没有全部恢复。不能因标签同名就称已配对。 | 原 [NovelStudioPage.readable.js:307](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioPage.readable.js:307)；CM [definitions.go:8](F:/脚本测试/一战晟铭-v88/backend/internal/prompt/novelcreation/definitions.go:8)；证据 `ev_ab1ec87d7b639d8eed18d18b284304d936f4768662c24fb4262eae196ac6a389` | 逐阶段登记原来源、输入变量和输出schema；缺证保留未知，不声称全部复制。 |
| A12 | 互动世界相关建议与模型配置 | 缺失 | 原建议动作和Agent枚举包含互动世界；当前六入口以外也没有对应完整玩法执行。 | 原 [NovelConfirmCard.readable.js:38](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelConfirmCard.readable.js:38)；CM [AgentPanel.jsx:34](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AgentPanel.jsx:34)；证据 `ev_102cdd3e022c3d3bcfaa326b4584ea6a3abb5ec4981e5b4c2f0ea3245355dcf6` | 先明确本模块是否纳入互动，再接真实状态推进；不可摆放无功能按钮。 |

### 自动写作

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| T01 | 定时写作默认与启停 | 部分实现 | 30分钟/每日10/连续失败3的默认方向相符，CM有完整持久任务；原全局调度器状态与CM多计划UI不同。 | 原 [NovelSchedulerPanel.readable.js:21](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelSchedulerPanel.readable.js:21)；CM [AutomationPanel.jsx:19](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AutomationPanel.jsx:19)；证据 `ev_e211619bb2d8329389490dfb96eaabcf5c12ad4c62eaae69f218763086441f78` | 启动/暂停/恢复/停止/日限/失败限/重启后状态逐项比较，保存不得误立即生成。 |
| T02 | 批量写作默认/进度/取消 | 部分实现 | 默认5章失败3次相符；CM是统一自动任务表单与持久队列，原简洁进度弹窗不同。 | 原 [NovelBatchDialog.readable.js:20](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelBatchDialog.readable.js:20)；CM [AutomationPanel.jsx:19](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AutomationPanel.jsx:19)；证据 `ev_33e257223bb78fc6627f919558deacc0208716236a835fbadf616346ed5b0ea2` | 逐章保存计数，取消后不多启动一章，失败重试不重复已完成章。 |
| T03 | 自动任务引用与冻结 | 已有实现待对照 | CM已有模型/引用/模式快照、原编号恢复和状态控制；这是可靠性实现，不是原入口形态等价。 | 原 [NovelSchedulerPanel.readable.js:8](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelSchedulerPanel.readable.js:8)；CM [AutomationPanel.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AutomationPanel.jsx:25)；证据 `ev_e211619bb2d8329389490dfb96eaabcf5c12ad4c62eaae69f218763086441f78` | 丢回包恢复、修改计划、切会话、资料失效时禁止串模式或重建重复任务。 |

### 看板与伏笔

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| D01 | 总章数/总字数/平均字数 | 部分实现 | CM只有章数/字数/任务数/成功率，缺原平均字数和相同统计口径。 | 原 [NovelDashboard.readable.js:46](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelDashboard.readable.js:46)；CM [NovelCreationPage.jsx:274](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:274)；证据 `ev_be8f13955c4b618aa7216c588c662fc9a06da71d9d71bbf08485785b45affeca` | 以已保存章节为基准对照，空/零字数/失败章状态清楚。 |
| D02 | 章节状态分布与字数趋势图 | 缺失 | CM没有原按章节状态分布条与逐章字数趋势。 | 原 [NovelDashboard.readable.js:54](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelDashboard.readable.js:54)；CM [NovelCreationPage.jsx:274](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:274)；证据 `ev_be8f13955c4b618aa7216c588c662fc9a06da71d9d71bbf08485785b45affeca` | 真实章节状态和字数驱动，空数据不显示虚构图形。 |
| D03 | 伏笔管理与活跃/已解决筛选 | 缺失 | 原看板实际挂载伏笔组件，显示起始/推进/解决；CM看板无该列表和状态合同。 | 原 [NovelHooksPanel.readable.js:40](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelHooksPanel.readable.js:40)；CM [NovelCreationPage.jsx:274](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:274)；证据 `ev_302e52d1c91af8de1b99d7d632e09980e779dd1033fc540ab699d0bba83fd240` | 真实伏笔记录可筛选并追溯章节；没有数据时明确空态。 |

### 验收与可靠性

| 编号 | 检查项 | 状态 | 原版与当前差异 | 来源 | 修复后必须验证 |
|---|---|---|---|---|---|
| Q01 | 已保存数据与失败恢复保护 | 已有实现待对照 | 持久派发、租约、恢复保存和未知结果保护已做，必须保留；原native同等实现尚未证明。 | 原 [NovelToolSteps.readable.js:203](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelToolSteps.readable.js:203)；CM [worker_store.go:381](F:/脚本测试/一战晟铭-v88/backend/internal/novelcreation/worker_store.go:381)；证据 `ev_d37ad7f377c0dda693b51bae7c51460cdc1fe988cd4c10975fb03a085b9dfe96` | 数据库落盘、刷新、重复操作、断网、取消后回包、重启恢复逐状态测试。 |
| Q02 | 主题/图标/同状态截图 | 尚未验证 | 当前沿用用户要求的玻璃主题；原模式图标/控件几何与全部弹窗没有完成同状态对照。 | 原 [NovelCreateIcon.readable.js:5](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelCreateIcon.readable.js:5)；CM [AgentPanel.css:3](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/novel-creation/AgentPanel.css:3)；证据 `ev_ac128bab24fd40b42fd4d98ed478d7987ab33e53ad59fc9a629be4d7c6a15b5e` | 约定保留自家主题，逐页对齐布局、图标、顺序、状态、对比度；不以主题差异掩盖缺控件。 |
| Q03 | 每个按钮真实效果 | 尚未验证 | 本轮枚举声明和静态调用，不是原应用每个按钮都运行过；自动测试通过也不等于原/CM完整对照。 | 原 [NovelStudioPage.readable.js:5](F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelStudioPage.readable.js:5)；CM [NovelCreationPage.jsx:25](F:/脚本测试/一战晟铭-v88/frontend/src/user/pages/NovelCreationPage.jsx:25)；证据 `ev_ab1ec87d7b639d8eed18d18b284304d936f4768662c24fb4262eae196ac6a389` | 每控件建立页面/前置状态/输入/请求/数据变化/结果/失败/取消/刷新证据，完成后才标已对齐。 |

## 先后顺序与完成门槛

每一步登记代码提交、来源证据、测试结果、运行版本、页面截图；没有结果就保留未完成。先处理合同与数据对象，再做UI逐像素验收。

| 步骤 | 范围 | 依赖 | 当前登记 | 完成门槛 |
|---|---|---|---|---|
| 0 | 全组件目录、点击声明、75项差异 | 原包固定hash | 本轮静态完成 | 来源锚点全部可定位；不等同原生全分析 |
| 1 | 补齐原版未恢复的native入口、提示词和存储合同 | 0 | 待执行 | 未知项逐个有函数/接口证据，不能用猜测填完 |
| 2 | 分开无书会话、建议动作、长篇book、短篇story与资料文件 | 0、已知合同 | 待执行 | DTO、MySQL读写、所有权、幂等、刷新恢复一致；不影响已有作品 |
| 3 | 六种创作入口、专用表单与字段贯通 | 2 | 待执行 | 每模式输入→保存→实际请求→对应输出，不能只做壳 |
| 4 | 内容检查点、资料修改后继续、短篇独立阶段 | 2、3 | 待执行 | 取消/稍后/继续/重复/失败状态可重现，读到保存的新文件 |
| 5 | 长短篇资料栏、文件编辑、封面、推荐、伏笔 | 2、4 | 待执行 | 每个可见资料都有真实存储/读取与错误状态 |
| 6 | 文风、Agent17项、37维、定时/批量、看板 | 1、2 | 部分已有，待逐项对照 | 保留自己的模型与配置；每个选项真实影响对应步骤 |
| 7 | 聊天动作、导航列表、图标、尺寸和状态细节 | 2—6 | 待执行 | 每页每控件同状态截图和键盘/鼠标行为对照 |
| 8 | 无费异常回归与获授权后的真实模型成品验证 | 3—7 | 未完成 | 正常/空/重复/错误/取消/刷新；没有落盘成品不能报完成 |

## 后续规格必须逐项填写的四件事

- **输入**：字段、来源、类型、默认值、计量单位、可选/必填、模型和资料归属。
- **输出**：实体/文件/版本、状态、下一页面、保存位置、重新打开如何读取。
- **边界**：空值、长度、emoji计数、重复提交、跨会话、失效资料、取消和迟到回包。
- **失败提示**：用户能理解的提示、保留哪些草稿、哪些可重试、未知结果如何避免重复收费。

## 证据目录

归档根目录：`F:/脚本测试/一战晟铭-v88/_local-archive-20261028/novel-full-parity-audit-20261118`。完整Craft调查为`full-novel.cany/`，33个组件图、14个重点函数结果和参数逐文件保留。`craft-full-audit-evidence.json`可导入Craft核对；文档只总结可证明事实，不复制任何凭据。

机器可读清单：`features.json`（75项）、`controls.json`（点击声明）、`components.json`（组件来源）、`summary.json`（计数与边界）。

本审计不提供“已完成百分比”：75是本次分解出来的检查项，有依赖和重叠，不能拿同名按钮数量或组件数当功能完成率。
