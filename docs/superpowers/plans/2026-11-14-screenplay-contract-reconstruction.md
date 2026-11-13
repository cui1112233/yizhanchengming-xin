# 剧本创作画风与提示词链路重建 Implementation Plan

> Use superpowers:subagent-driven-development for independent investigations and implementation tasks, then review findings and run integration acceptance. User requests evidence, specifications, independent implementation and five classes of comparison tests. Continue without repeated approval questions.

**Goal:** 让画风、剧本提示词、分镜提示词及资产设计提示词的选择、保存和真实使用对应起来，按原版证据规定行为；未证实链路不能标为完成。

**Architecture:** 先固定原版文件与当前样例，恢复前端桥接参数、提示词类型、配置快照、文件和请求边界；形成输入/输出/边界/失败规格后，复用现有Go/React/MySQL/Redis/TOS服务独立实现。系统提示词和拼接规则放后端prompt模块，前端只选择ID及传用户输入。

**Tech Stack:** REA、Craft Anything、Go、React/Antd、MySQL、Redis、TOS、Go embed。

## Global Constraints

- 当前主线v88直接维护，不建立分支或工作树；保留其他聊天已有改动，不混入提交。
- 不调用付费模型、不触发原版生产任务；之前预算已用完，不自行增加。
- 原版只读分析及非付费交互，不修改其安装包、凭据或权限。
- 用户明确要求提示词和调用链可接入；以其已可访问内容和本地数据为证据，不能复制凭据或绕过账号权限。
- 原始压缩文件引用真实行号、列和偏移；反格式化文件明确标为派生证据，不伪造原始源码行号。
- 本地优先，修改需提交、构建embed并验收4001；不推送或发布公网。

### Task 1: 原版证据与需求规格

- [ ] REA及Craft记录目标身份、可用能力和证据；复用既有证据但核对当前摘要。
- [ ] 查入口、画风来源、四类提示词、配置保存、生成参数、取消和错误分支；每结论附文件定位及证据编号。
- [ ] 追踪本地文件/SQLite变化和网络请求构造，静态候选与运行观察分开。
- [ ] 写需求规格和正常、空值、重复、错误、取消验收表，列出不确定项。

### Task 2: 本项目差异与独立实现

- [ ] 按规格确定画风库及各提示词目录来源，禁止用批量工厂画风或固定引擎假装配对。
- [ ] 实现项目级选择和快照、后端解析与模型参数传递、预览、错误提示及取消。
- [ ] 正常/空值/重复/错误/取消测试，使用真实适配器和模拟供应商观察请求，不消耗额度。
- [ ] 独立评审权限、串配、覆盖、重复生成、取消及旧项目兼容。

### Task 3: 本地验收与交付

- [ ] 重新构建React/Go embed，只替换已确认空闲的本地进程。
- [ ] 登录页面核对目录、预览、选择、保存和刷新；真实API核对持久化与错误。
- [ ] 按原版场景测试可执行的非付费行为；无法证实的生成/网络细节保留未验证。
- [ ] 提交证据、规格、差异与验收报告，更新Craft发现、未知与交接。
