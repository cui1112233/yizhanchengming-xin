# 创作对话短篇分流修复与执行登记

继续执行用户已批准的原版对齐范围；模型使用本系统配置。2026-11-20。

## 规则与依据

原版 `source/ida/0x1801fd690.c:3177–3668`：short_run 从已保存待确认动作读取 direction / instruction / summary / title，优先取非空创作方向，调用独立短篇生产入口。当前CM仅支持create_book，这属于明确缺口。使用原版流程语义，适配CM现有独立短篇实体及任务接口，不宣称二者内部全流程已经一致。

输入为服务端保存的short_run提案及用户选择的自有模型；输出为同一次确认创建的短篇作品、会话、任务及story。重复确认返回原结果；取消后不可启动；不允许用客户端替换待确认正文。空方向、超长输入、未知动作应明确拒绝，不建作品。保留原create_book及冻结讨论输入、账号隔离、事务、防重复保护。

前端卡片区分长篇和短篇，显示实际方向与参数。确认后按返回session.mode导航，短篇显示短篇结果/进度；不得硬编码long_original。提案提示词在后端prompt模块，后台可管理。

Agent设置区分加载中、成功及失败；加载中不显示失败或0/0，读取失败保持禁止保存并提供重读。切换范围后旧请求不得覆盖新范围。

## 执行计划

- [x] 固定原版分支证据，确认复用现有短篇任务与事务。
- [x] 增加短篇提案、确认、重复、取消、越权及非法输入回归，先确认旧实现失败。
- [x] 修改服务端协议、后端提示词、确认分流及前端导航；不新增模型通道。
- [x] 修复Agent加载误报，验证慢加载、失败、重读及成功保存。
- [x] 定向回归与构建，登记运行状态及仍未核验的差异。

不运行原版、不消耗模型额度、不发布公网。历史18项前端失败尚未全部诊断，本次不得将其登记为已全部解决。

## 已完成的验证

- 后端旧实现新增short_run回归明确失败（无待确认动作）；修复后通过。direction / instruction / summary / title按原生优先级回退，默认13000字/1章；原版全部短篇字段尚未穷举，不宣称完全对齐。
- 复用短篇事务主体，确认记录写入失败时整个短篇创建回滚；4路同时确认仅产生同一个任务。账号隔离、取消、重复确认覆盖。
- Studio与审稿维度两个浏览器测试文件12项通过；整页长短篇导航与刷新恢复2项通过。使用模拟接口，不消耗模型额度，不代表真实上游出文。
- Agent慢加载测试只检查root真实界面文本，排除脚本内文字造成的假判断。旧源码通过esbuild只读替换重新验证，确实误报；新源码慢加载、失败、重读测试通过。旧源码位于本地归档novel-routing-fix-20261120/baseline，未覆盖产品文件。
- 新协议使用后端预设ID novel-creation.studio-planning-v2；不覆盖管理员已保存旧词，保证新建请求真正消费新协议。旧任务仍沿用冻结词。

## 尚未关闭

write_next / audit_chapter / revise_chapter等其他Studio动作仍未接通。原版全部生产内部阶段、全部提示词消费、像素级还原及真实收费模型全程输出未完成。历史18项前端失败保持开放，不能把本次定向通过当作全站全流程通过。

## 本地运行验收

修复提交：`33e6dac1`。后端小说及prompt两包、HTTP小说入口、5项源合同测试通过；前端构建通过。新增程序链路测试用模拟模型依次走提案、大纲、大纲审稿、正文、正文审稿、推荐文案，得到已持久化正文及至少3个成品文件；重复运行不新增调用。这仍不替代真实模型及MySQL全流程验收。

4001实际返回git_sha/runtime_manifest_sha=33e6dac1，入口`/assets/user-CScrVE-5.js`，小说页资源`NovelCreationPage-BGmnM6ye.js`。health=true；后台确有新预设ID且包含short_run。具体资源指纹见`../../verification/novel-contract-test-20261120/repair-runtime.json`。

旧二进制文件被占用，复制替换没有成功，故未继续把旧文件标成新版；已恢复旧文件配套清单，改用已构建的新候选二进制在原4001运行：

- 程序：`F:/脚本测试/一战晟铭-v88/.local-review/qiantie-novel-routing-candidate.exe`
- 专用清单：`.local-review/novel-routing-runtime-manifest.json`
- 进程：61220（仅为本次验收时的值）
- SHA256：5826760B186629F0D2FF247EBB1C3934236F04C19C043EFAC02AB76461FF46BE
- 重启方式：`scripts/start-local-go-review.ps1 -UseYcmMysql -SkipBuild -Detached -BinaryPath .local-review/qiantie-novel-routing-candidate.exe -ManifestPath .local-review/novel-routing-runtime-manifest.json`

已通过CUA刷新实际登录页面并读取Agent设置：37项/29项选中，无目录失败提示。没有保存设置，没有恢复或重试旧任务；返回后仍暂停5%/0章。未通过真实模型在页面新建短篇，收费调用0；没有公网部署。
