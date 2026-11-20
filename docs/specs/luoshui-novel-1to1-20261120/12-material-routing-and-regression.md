# 素材分流与回归排查

2026-11-20，继续用户批准的洛水对标范围。模型使用本系统配置，不调用收费模型，不执行原版模块，不发布公网。

## 原版依据与真实缺口

- `source/frontend/NovelStudioPage.readable.js:706–709`：assistedAction按targetRoute打开材料入口，提示补材料；不把它当建书或生成完成。
- `source/ida/0x1801fd690.c:2242–2285`：保存会话后返回assistedAction/targetRoute。
- 原版同一函数`:7383–7437`：audit_chapter读取bookId/chapterNumber后调用独立审稿；`:8367–8433`：独立修订。
- CM现有写章循环的内部审稿/修订不能等同于独立章节动作。目前没有完整对外独立审稿/修订入口，仍未接通，禁止假装已对齐或接成写下一章。

## 本轮实施规则

1. `style_imitation`确认后返回固定imitation入口，打开已有仿写长篇表单；`continuation_import`返回固定chapters入口，打开已有续写章节导入表单。其他素材动作尚未纳入本次实现。
2. 路由由后端固定，不能执行模型返回的任意URL。确认必须使用已保存的待确认动作，支持原回执重放；取消后不可确认，账号之间不能读取或确认。
3. 打开表单不创建作品、不生成正文、不调用模型；无模型配置也可打开。表单本身仍沿用原有输入验证和明确提交。
4. 素材动作显示“素材入口已打开”，不能显示“作品已创建”；打开/取消表单后讨论仍可输入。确认后的回执可再次打开同一个入口。
5. 提案协议放后端prompt模块，新增v3预设，不覆盖管理员旧预设。老任务继续使用已冻结的提示词。
6. 已发现的真实漏洞：短篇提案direction/instruction/summary/title/readerPreferences没有完整计入上下文字数。修复后与旧长篇提案一起计入；超限在调用模型前拒绝，保留原待确认建议及消息记录。

## 登记

- [x] 原版材料分流和独立章节动作证据核对。
- [x] 短篇上下文上限旧实现失败，新实现通过；字段覆盖及无额外模型调用验证。
- [x] 两种素材动作原实现拒绝；新实现确认、重放、取消、账号隔离、不建作品、不调用模型通过。
- [x] 两个真实页面弹窗的打开与取消回归通过；两个卡片无模型配置仍可打开且不锁讨论。
- [x] 历史失败的现行合同迁移及定向复测：见下表。
- [x] 后端/前端相关回归、构建、本地4001验收与运行身份登记。

## 历史失败的分类

以下5条使用现行合同替换并通过。这是测试合同迁移，不是5个业务漏洞修复；历史18失败中其余13条没有在本轮获得关闭证据。

| 历史失败 | 原因及现行验证 |
|---|---|
| C5a四工具切换 | 寻找已移除的“创建定时任务”按钮；改验证当前三参数调度器，返回原会话草稿不丢、不生成。 |
| C5a R1定时内部补建会话 | 旧用例要求短篇走旧定时生成；当前短篇使用独立成品流程。改验证长篇无会话时创建真实会话，启动中的三参数保持，返回对话草稿保留；接口为scheduler/start。 |
| C5a无作品工具导航 | 旧用例要求所有工具必须先有作品；账号默认Agent及无书可查看调度器已经支持。验证账号默认可保存、作品范围禁用、导航不创建作品。 |
| 内容暂停与模型调用确认 | 模拟设置接口漏掉当前真实返回的reviewDimensionCatalog，补用后端同一37项目录后，两个开关独立保存重开验证通过。 |
| 旧设置缺少内容暂停字段 | 同上，补齐真实目录；保存其他设置不偷偷写入缺失pauseAfterBlueprint字段。 |

另有账号无作品保存默认设置用例通过。定时测试最初在启动成功后仍找已收起的输入框，已改为延迟返回来验证启动中的参数，再核验成功控制及返回草稿；没有为了过测试修改产品行为。

新增素材卡片测试最初重新挂载了一个不恢复会话编号的测试Host，导致自身超时；已改为只切换可用模型状态，验证当前已保存动作无模型仍可确认。后端旧实现拒绝的失败证据保持独立，不将测试Host超时误当业务根因。

## 仍未完成

独立写下一章/审稿/修订的Studio提案和执行、fanfic/spinoff材料分流、完整原版生产阶段、全部提示词消费及像素还原仍开放。本轮没有真实上游出文或真实模型全流程验收。

## 相关回归及REA补证

后端novelcreation与prompt两包通过，HTTP小说入口通过；Studio/审稿目录/调度器3文件21项浏览器用例通过，整页8项通过，5项源合同通过。总计29项浏览器定向验证，无跳过，不代表历史全套用例通过。

本轮实际调用REA静态分析`NovelStudioPage.readable.js`与`NovelConfirmCard.readable.js`：2个文件解析成功，parse_failures=0。父证据`ev_f0e187d6e388d20a51d7fff0764f21a5a47e21229aa58b245e767bcd1947add9`，摘要证据`ev_c0e27039c65838a6e46d4a7b5839658c080d3c70da430e13fe28b26eb19f0ed3`，落盘`source/rea-intake-components-summary.json`。

输入为已恢复的只读文件，不是新执行原程序；22文件未选，语义图partial/truncated、3208未知，不能把parse成功当作全链路正确或全量复刻证据。具体分流判断仍按前述文件行号人工核对。

## 本地4001验收

修复提交`016d1dc3`，构建与本地重启完成；未发布公网。实际`git_sha`与`runtime_manifest_sha`均为016d1dc3，health=true，用户入口`/assets/user-CwMHU1M2.js`，小说资源`/assets/NovelCreationPage-DrcAvVGw.js`。后台已创建`novel-creation.studio-planning-v3`，实际保存内容同时包含style_imitation和continuation_import。

- 进程43524（验收时值），程序`.local-review/qiantie-novel-intake-016d1dc3.exe`。
- 清单`.local-review/novel-intake-runtime-manifest.json`。
- 重启：`scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached -BinaryPath .local-review/qiantie-novel-intake-016d1dc3.exe -ManifestPath .local-review/novel-intake-runtime-manifest.json`。
- 程序SHA256：4b6717400df0d8ee274200a0ea32a4df8e0a023f892f563c81e988c2f57ecf54。
- 资源及后台预设验证：`../../verification/novel-contract-test-20261120/material-routing-runtime.json`。

CUA刷新真实登录页面，作品仍3本，原anti2任务仍暂停5%、0章，未保存设置、恢复、重试或生成。实际材料动作完整交互由隔离浏览器模拟接口回归验证，未在用户数据中伪造提案；本轮收费调用0。原版独立审稿/修订及其他13条历史失败继续保留，不以本次结果宣布全部对齐。
