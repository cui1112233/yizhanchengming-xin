# 03 六种创作入口及素材导入规格

本规格描述原版；不把CM当前实现反过来当原版。依据是当前bundle `7c252056…` 的恢复片段。原版桥接混淆表达式尚未全部重新解码，相关名称在05中只有常量级依据时，必须标“绑定待核验”。下面的来源行号均指本目录 `source/frontend/`。

## C01 工作区和入口

- 输入：当前模式、作品/会话列表、当前作品和会话ID。
- 输出：六个入口：长篇原创、短篇原创、仿写长篇、仿写短篇、番外创作、续写创作；辅助入口为写作风格、导入素材、Agent设置、定时写作、数据看板、备忘录。
- 边界：作品和会话是不同对象；当前模式会话不能凭空作为作品。作品过滤器保存于`novel_book_filter`。运行状态按会话读取，切会话不应把旧会话进度显示给新会话。
- 失败：列表为空显示空态；不能把空列表补成假作品或假资料。
- 来源：`NovelStudioSidebar.readable.js:6–115`；父组件事件绑定`NovelStudioPage.readable.js:1520–1585`。完整原生归属与过滤工具函数仍待绑定。

## C02 长篇原创

- 输入：对话中的题材、世界观、人物、核心冲突；选择的模型、文风资料ID、思考设置和可选上下文长度。
- 输出：会话回复及可能的`action`；经确认后创建/加载作品，逐章生产。不是点击入口便生成整本书。
- 边界：先确保会话存在；请求中的`sessionId`取实际创建结果；当前模型由父页面提供。前端`thinkingMode`在发送载荷中为true，同时给出Codex/Claude的reasoning字段；不能据此断言每个上游模型实际支持这些参数。
- 失败：取消不按普通“发送失败”弹错；其他失败使用返回message或“发送失败”。请求结束重新加载会话和关联作品。
- 来源：`NovelStudioPage.readable.js:442–492`；`NovelStudioChat.readable.js:47–68,229,331`。
- 待补：原创的全部工具路由、动作确认如何落成架构/大纲/章稿的完整原生调用图。前端入口有证据，不等于完整原创引擎已恢复。

## C03 短篇原创

- 输入：短篇创作会话的自然语言要求、模型和文风选择；与长篇保留不同模式。
- 输出：短篇实体和文件面板，可展示大纲、章结构、成品、推荐。不能把长篇的第1章冒称短篇成品。
- 边界：会话可包含`shortStoryId`；轮询和资料加载会检查会话/短篇ID，防止串作品。右侧空态必须按实际文件存在情况显示。
- 失败：空文件、保存失败和生成失败分别显示；不能只完成进度标签便声称有成品。
- 来源：`NovelStudioPage.readable.js:836–851`；`NovelStudioShortPanel.readable.js:55–62,145–272`。
- 待补：短篇原创如何从对话动作进入`ShortFictionRunner`，与仿写短篇复用到哪一步。未以仿写路径代替原创路径证明。

## C04 仿写长篇表单

| 字段/选项 | 输入及默认 | 输出/处理 | 边界与错误 | 当前源码 |
|---|---|---|---|---|
| 暂定标题 | 可空 | trim；供创建作品 | 留空自动生成是界面说明，后端实际生成待核验 | `NovelImportPage.readable.js:487–490,315–322` |
| 参考原文 | 粘贴或txt/docx导入 | trim，追加选中备忘录，形成referenceText | 原文非空且去空白至少100；先校验原文，不能靠备忘录凑数 | `NovelImportPage.readable.js:301–321` |
| 文件导入 | 多文件 | 文本读入上方，可继续编辑；文件记录和原文不同 | 不支持文件选择、读取异常、空内容分别处理；取消不算导入成功 | `NovelImportPage.readable.js:228–267` |
| 参考章节 | 最近3/5/10章、手选、全选、清空、确认/取消 | 从解析章节生成引用文本 | 弹窗先维护临时选择，确认才生效；取消不覆盖已选项 | `NovelImportPage.readable.js:267–299,522–548,779–823` |
| 文风统计指纹 | 参考文本 | 平均句长、波动、段长、词汇多样性、句首/修辞 | 页面称纯算法不收费；8维深度拆解是后端另一阶段，不可混为一次AI | `NovelImportPage.readable.js:548–562` |
| 每章字数 | 默认5000；3000/4000/5000/6000/8000/10000 | `targetWordCount` | 当前传`value || undefined`；具体后端裁剪待补，不能套短篇上下限 | `NovelImportPage.readable.js:54–82,364–370,562–585` |
| 目标章节 | 默认100；50/100/150/200/300 | `targetChapters` | 原输入框范围见02；实际总字数不能误读为每章字数 | `NovelImportPage.readable.js:55,84–102,585–605` |
| 读者偏好 | 自由填写、标签追加 | `readerPreferences` | 默认空，不能自动补用户未选题材 | `NovelImportPage.readable.js:605–623` |
| 原创故事想法 | 可空 | `storyIdea` | 与原文、仿写方向分开 | `NovelImportPage.readable.js:623–630,315–322` |
| 仿写方向 | 可空 | `selectedDirection` | 界面说明留空默认方向一；实际原生默认分支须由函数证实 | `NovelImportPage.readable.js:630–635,364–370` |
| 补充信息 | 可空 | `supplementaryInfo` | trim后空值不传 | `NovelImportPage.readable.js:637–642,364–370` |
| 文风设置 | 多选资料ID | `knowledgeBookIds` | 选中ID不是原文；不得凭空填“当前资料” | `NovelImportPage.readable.js:644–653,315–322` |
| 备忘录 | 多选 | 在referenceText末尾按选择次序追加标题与plain | 不存在或无plain的条目跳过；不改原文框内容 | `NovelImportPage.readable.js:310–314,655–664` |

提交输出：`creationMode='styleImitation'`、`styleProfile`、`styleDimensions`及上表字段；文本按50000个JS码元分片，多于1片才传`referenceChunks`。表单发出`submit-start`后由父组件补当前会话/模型。来源：`NovelImportPage.readable.js:324–378`、`NovelStudioPage.readable.js:501–568`。混淆helper的当前版本字符串解码及统计全部细则登记为待补，不复用旧bundle下标。

## C05 仿写短篇表单（与长篇分开）

| 字段 | 输入/默认/选项 | 提交结果 | 边界和失败 | 当前源码 |
|---|---|---|---|---|
| 暂定标题 | 可空 | trim或undefined | 空标题不阻止提交 | `NovelImitationShortDialog.readable.js:150–164,192–195` |
| 参考原文 | 粘贴或多文件 | trim后追加选中备忘录 | 去空白少于100禁用开始；提交再校验 | `NovelImitationShortDialog.readable.js:80–81,143–149` |
| 文件 | docx/txt | 自动填入参考框；多个文件有分隔标题 | path相同跳过；移除文件标签是否同时删文本须单独测试，不能想当然 | `NovelImitationShortDialog.readable.js:83–140,211–226` |
| 目标字数 | 默认13000；10000/11000/12000/13000/15000/20000 | 有限且≥1000则四舍五入，上限500000；否则13000 | 不是限制只能选预设；不把总字数当每章字数 | `NovelImitationShortDialog.readable.js:56–75,105–112` |
| 目标章节 | 默认1 | 有限且≥1则四舍五入，上限50；否则1 | `targetChapters`和`chapters`均为这个数，不是章节数组 | `NovelImitationShortDialog.readable.js:109–112,150–164` |
| 读者偏好 | 自由写；轻松搞笑/热血爽文/智斗权谋/悬疑惊悚/甜宠治愈 | `readerPreferences`；原版`styleName`也取同一个值 | 标签追加规则按点击代码；不能拆成两个无关字段 | `NovelImitationShortDialog.readable.js:56,150–164,249–266` |
| 补充信息 | 可空 | trim或undefined | 不可丢失人设/禁忌/必须桥段 | `NovelImitationShortDialog.readable.js:150–164,279–286` |
| 灵感标签 | 22项；含女频/男频/BE/HE等 | 选中数组以顿号join | 再点取消该标签；不能自动选全 | `NovelImitationShortDialog.readable.js:76,286–297` |
| 文风设置 | 多选 | `knowledgeBookIds` | 无可用文风显示空态 | `NovelImitationShortDialog.readable.js:297–308` |
| 关联备忘录 | 多选 | referenceText追加素材 | 不满足原文100字时仍不可提交 | `NovelImitationShortDialog.readable.js:143–149,308–317` |
| 关联存档 | 可空，已有作品列表 | `linkedSaveId` | 不等于自动覆盖该作品；原生实际写入关系待补 | `NovelImitationShortDialog.readable.js:77–79,150–164,317–324` |

开始提交后原表单清空并关闭；父页面检查/创建会话、补模型，然后进入短篇生产。这里的关闭只是表单关闭，不是取消已提交任务。来源：`NovelImitationShortDialog.readable.js:143–165`、`NovelStudioPage.readable.js:854–932`。

## C06 番外创作

- 输入：书名、参考文本/文风、正传书籍ID、故事方向；类型为后传(sequel，默认)、前传、平行线、If线。
- 输出：携带`parentBookId`和`spinoffType`的建书结果；成功显示“书籍已创建”，发出created事件。
- 边界：必须选择正传；同一表单的通用字段和模式字段分开。此分支没有像长篇仿写父处理那样明确补`model`，不能套同一载荷。
- 失败：缺正传提示“请选择正传书籍”；返回失败显示message或“创建失败”；finally解除忙碌状态。
- 来源：`NovelImportPage.readable.js:39–43,303–322,393–420,439–444,666–703`。
- 待补：后台如何读正传事实、角色和时间线，正传被删/同名/无章节时的真实失败分支。

## C07 续写创作

- 输入：目标书籍或新书名、章节正文、参考文风和备忘录。
- 输出：将章节数组导入现有或新建作品，显示“章节已导入”，使用`data.book || data`作为创建结果；后续写作需另走写章链路。
- 边界：不能将“导入成功”等同“续写生成成功”。解析使用原章节框文本，不是已追加备忘录的referenceText。章节按“第N章”切分；非空无章标题时保留为第1章；wordCount按去空白JS码元统计。
- 失败：“请选择目标书籍或填写新书书名”“请粘贴续写章节正文”“未能解析出章节，请检查格式”；服务失败显示message或“创建失败”。
- 来源：`NovelImportPage.readable.js:196–218,303–322,420–444,672–681,726–731`。
- 待补：归属权限、同章重复导入、覆盖策略、状态重建和取消中的数据库事务。

## C08 素材导入中的同人入口

- 输入：书名、原作参考、同人方向、文风；模式canon/AU/OOC/CP。
- 输出：通用载荷加`fanficMode`，独立创建结果。
- 边界：同人不是六个主入口之一，但导入素材仍有该选项，规格不能漏。书名与原作参考均受前端校验。
- 失败：缺书名/原文分别提示；创建失败不显示成功。
- 来源：`NovelImportPage.readable.js:113–146,303–322,383–404,704–726`。
- 待补：同人专用审稿维度与正文prompt的完整原生分派。

## C09 每个入口的验收步骤

所有入口均测：①正常最短有效输入；②全空/仅空白；③连续点击/重复文件/重复导入；④错误格式/已失效关联对象/模型失败；⑤弹窗取消、文件选择取消、生成取消。记录提交载荷、真实返回、文件/数据库前后变化；未运行时登记“待测”，不能仅按按钮是否存在判通过。具体测试矩阵见07。
