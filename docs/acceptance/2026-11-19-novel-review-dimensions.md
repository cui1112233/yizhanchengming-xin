# 37项审稿维度交付

## 补充：维度实际作用与修订链路核查

2026-11-19用户要求确保有效，继续核查实际模型请求。发现原审核请求缺少明确的style_profile/mode_inputs，原修订请求缺少规则、备忘、素材和基础设定，虽有问题列表但核对依据不完整。新增withReviewDimensionContext，针对含维度设置的新任务，将已冻结的这些资料同时传给审核和修订，不读取任务之后的新资料，不改变旧快照无维度字段的请求体。

TestReviewDimensionsContextAndRevisionEffect先失败于审核请求缺style_profile，修复后通过：37项各自完整名称和说明均进入首次审核、修订和复审，10项关键上下文字段齐全，具体issue id/description传到修订。TestReviewDimensionFindingReachesSavedChapter验证选中6/32/33贯穿循环、未选37不泄漏，修订后的指纹动作写进最终正文而非只留在中间结果。

TestReviewDimensionFindingRealMySQL在隔离MySQL8.4通过：测试模型8次响应，首次审稿70分→修订→复审92分，重新建Store读取最终正文已变更，数据库落下2条novel_audits、1条novel_revisions。整包novelcreation/prompt及HTTP相关测试通过。测试使用受控假模型，证明请求、修订、复审和持久化接通；不等于真实模型判断准确率或原版native质量已验证。没有真实收费调用。

产品提交：`e4fe05a1`。本地4001运行同提交标记，入口`/assets/user-B66IWgQt.js`，二进制位于`_local-archive-20261028/novel-review-dimensions-20261119/qiantie-review-dimensions.exe`。未发布公网。

## 原版依据

- 原文件：`F:/软件/洛水2026/洛水2026/modules/novels.pyd`，SHA256 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。
- `scripts/extract-novel-review-dimensions.py`只读提取Nuitka常量区，不导入或运行原版代码。37条顺序、中文名称、中文说明和每条字节偏移见`docs/reverse-engineering/novel-review-dimensions-20261119/`。相邻集合记录1–27、32、33，共29项默认启用；用户截图29/37与此一致。
- 原版界面合同：`_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelAgentSettings.readable.js:354`起。只读说明、逐项开关、全选/全不选、已选计数、自定义项追加在后。
- Craft Anything记录：`ev_0421deb495b1c7aa4121307f559a45f041da7536398365e061445d704d8d92a9`，归档目录及来源文件SHA可复核。这是目录恢复证据，不能冒充原版完整审稿算法。

## 输入、输出、边界和失败

- `GET/PUT /api/novel-creation/projects/{projectID}/settings`返回后端`reviewDimensionCatalog`；设置只保存`agent.reviewDimensionIds`。内置文案在`backend/internal/prompt/novelcreation/review_dimensions.json`，前端不硬编码提示词。
- 默认29项；显式`[]`始终表示全不选。未知、重复或越界编号返回400及“审稿维度必须是1至37之间不重复的编号”。其他账号返回404。
- 新任务冻结有效选择，修改设置不改变已排队任务；历史快照无此字段时保持旧提示行为。旧客户端省略新字段不会清空已有选择。
- 选中项标题及说明进入章节审稿、短篇审稿及相关修订请求，自定义审稿仍追加在后。不注入规划、正文写作、结算、索引。全不选取消额外维度，不取消基础审核和JSON输出合同。
- 缺少正典、设定、前文或memo时，提示词要求注明无法核验，不能虚构或声称通过。维度并不自动创建这些资料；原版完整资料检索和native私有审稿算法仍未对齐。
- 前端保存失败保留修改，不假报成功；目录缺失显示重新打开设置提示，避免伪造37条。
- 按既有项目范围保存，非原版账户全局配置迁移。短篇新建任务使用默认维度，未在本轮补齐所有模式的全局Agent设置继承。

## 验证

- Go：`go test ./internal/novelcreation ./internal/prompt/novelcreation -count=1`通过。
- HTTP：小说创作、调度、短篇、导入、章节及审稿维度相关测试通过；验证37目录、单项/全不选保存、越界拒绝、跨账号隔离和零生成任务。
- `TestReviewDimensionsActualProviderRequestAndLegacyOmission`检查实际假模型请求：仅审核调用含所选1/32项，不含未选37项；调用数仍6，无新增付费阶段。
- `TestReviewDimensionsRealMySQLRoundTrip`在隔离MySQL8.4库通过：新Store重新读取单项选择和显式空列表。首次测试因测试夹具漏题材失败，补齐夹具后通过。
- 新浏览器2/2通过：37行、29默认、固定说明、全选/全不选、保存重开、保留自定义项、保存失败保留编辑。已有调度器7/7通过。
- Vite构建、Go嵌入构建通过；仅已有包体大小警告。
- 实际登录页面：打开Agent设置，显示37项和29/37；保存默认选择，刷新重开仍正确。截图`_local-archive-20261028/novel-review-dimensions-20261119/review-dimensions-live.png`及同目录AX已保存。测试书“本地检查验收 · 不调用模型”，不生成正文。
- 真实模型调用0；原版与CM付费产出质量未比较。全软件一比一对齐仍未完成。
