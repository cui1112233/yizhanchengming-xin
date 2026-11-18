# CORE-02：长篇、短篇与资料文件的保存合同（待实施）

这是CORE-01之后的规格草案，尚未写入产品，不能当成功能已实现。配置继续用我们的模型、MySQL、Redis及TOS；系统提示词只放后端prompt模块。现有长篇章节版本和未知结果保护必须保留。

## 原版已经看到的外部证据

原文件统一位于 `_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/`，来源固定bundle SHA256=5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a。

- NovelStudioPage.readable.js:297：同一会话分别恢复book/bookId与shortStory/shortStoryId，不能把短篇只当成长篇的一章。
- NovelStudioShortPanel.readable.js:7-11、41-61：短篇有独立编号、chapters、artifacts、outlinePreview、salesPreview、coverImagePath、status。
- 同文件:76-102、112-136：按storyId及文件key读取和保存内容，成功后发saved，失败提示保存失败。:150-159通过outline/前缀定位完整大纲。:189-245分别显示章节结构、大纲、成品文件。
- NovelStudioBookPanel.readable.js:42-58：长篇分别读取foundationFiles、coverImages、chapters、characters、worldviewSummary/storyFiles.story_frame。:77-125读取章节和核心文件，:145-193分开保存，:257-341呈现人物、封面及完整设定。
- NovelStudioPage.readable.js:541-583：检查点保存tool/checkpoint/storyId/bookId/sessionId/message；outline检查点停止在大纲审修之后，正文步骤仍待执行。
- 同文件:587-637：继续使用检查点身份和当前模型；若再次awaitingConfirmation仍显示新检查点，只有明确完成才清掉。此为前端可见合同，尚未恢复原native内部算法。

对应已有Craft证据：完整目录清单ev_b40de5d2d1ad021b0a7c1518c8fb9ad8a818619be967ebe13cde0e9ef8e4879d；原继续函数Ve探查在既有function-probes.json中。以上文件/行号是恢复分析文件，不能冒称原工程源码。

## 独立实现的四项规则

| 对象 | 输入 | 输出 | 边界 | 失败提示 |
|---|---|---|---|---|
| 长篇book | 账号、作品编号、已确认设定 | 原有作品及章节，另外带人物/核心文件/世界观的真实保存记录 | 不复制旧章；已有章节引用和版本编号不变 | 不存在/无权限一致返回找不到作品 |
| 短篇story | 账号、专用创作参数、来源会话、固定创建编号 | 独立storyId、标题、状态、实际字数、章节结构及文件摘要 | 不隐式建立长篇；重复创建返回同一个story；请求编号换参数拒绝 | 字数/来源/模型配置问题在调用前说明；上游未知保持待核对 |
| 资料文件 | 所属对象类型和编号、允许的文件key、内容、期望版本 | 服务端版本、内容、更新时间、摘要 | key由后端目录映射，禁止路径穿越；账号和对象隔离；版本冲突不覆盖 | 失败保留输入；明确“资料已变化，请核对后保存” |
| 内容检查点 | 原run、对象编号、检查点编号、已保存版本及当前模型 | 当前待确认状态或同run继续结果 | 保存资料不自动调用模型；继续只消费已保存版本；重复继续不重复派发 | 版本过期提示重新读取；稍后保留检查点；结果不明只查询 |

以上接口及存储字段是我们的设计合同，具体SQL表/API路径待实现前结合已有Store确定；不宣称这些是原版私有接口。图片/封面先只读取已有资源，不用空按钮伪造生成入口。

## 执行顺序与验收

1. 长短篇对象和文件目录：MySQL追加迁移、账号隔离、只读重开及版本冲突测试，保证已有长篇不回归。
2. 短篇专用参数快照和产物：落实独立story，不用ChapterVersion冒充完整短篇；空值/重复/错误/取消先免费测试。
3. 大纲及内容检查点：先落盘→允许编辑保存→明确继续→读取该版本；结果未知禁止自动补发。
4. 原版资料栏分区/图标/点击进入编辑器，对相同输入输出状态截图；不能只比较颜色。
5. 专用表单和后端字段完全连通后，才按新增明确额度验证真实模型成品；当前预算0。

尚待证据：原native文件写入、完整私有提示词、原算法及实际收费次序；不能用我们自编的实现来证明原版内部相同。
