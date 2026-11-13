# 剧本创作画风与提示词链路重建 Implementation Plan

> Use superpowers:subagent-driven-development for independent investigations and implementation tasks, then review findings and run integration acceptance. User requests evidence, specifications, independent implementation and five classes of comparison tests. Continue without repeated approval questions.

**Goal:** 让画风、剧本提示词、分镜提示词及资产设计提示词的选择、保存和真实使用对应起来，按原版证据规定行为；未证实链路不能标为完成。

**Architecture:** 先固定原版文件与当前样例，恢复前端桥接参数、提示词类型、配置快照、文件和请求边界；形成输入/输出/边界/失败规格后，复用现有Go/React/MySQL/Redis/TOS服务独立实现。系统提示词和拼接规则放后端prompt模块，前端只选择ID及传用户输入。

**Tech Stack:** REA、Craft Anything、Go、React/Antd、MySQL、Redis、TOS、Go embed。

## Global Constraints

- 当前主线v88直接维护，不建立分支或工作树；保留其他聊天已有改动，不混入提交。
- 用户本轮已授权最多4次文本请求，图片/视频0次；截至当前实际付费请求0次，不沿用此前图片预算。
- 原版只读分析及非付费交互，不修改其安装包、凭据或权限。
- 用户明确要求提示词和调用链可接入；以其已可访问内容和本地数据为证据，不能复制凭据或绕过账号权限。
- 原始压缩文件引用真实行号、列和偏移；反格式化文件明确标为派生证据，不伪造原始源码行号。
- 本地优先，修改提交并构建embed。4001被其他工作切换后保留，改在4004和独立MySQL/Redis验证；不推送或发布公网。

### Task 1: 原版证据与需求规格

- [x] REA及Craft记录目标身份、可用能力和证据；核对原包摘要。
- [x] 查入口、画风来源、四类提示词、配置保存、桥接参数、取消和错误分支；结论附定位及证据。
- [x] 核对选中配置及SQLite四份正文；桥接参数为静态证据，原模型HTTP与生产文件变化明确未验证。
- [x] 写输入/输出/边界/失败规格和五类验收表。

### Task 2: 本项目差异与独立实现

- [x] 独立用户画风库与四类提示词目录，保留旧模板。
- [x] 实现项目冻结选择、正文解析、生成参数配对、预览、错误与取消保护。
- [x] 五类模拟边界与真实MySQL测试通过；前端49项通过。
- [x] 独立评审修复图片模型被首项覆盖、V7/V3/V9协议错配、批次漏读V9。

### Task 3: 本地验收与交付

- [x] 重新构建React/Go embed，在空闲4004启动，保留4001进程。
- [ ] 登录页面核对目录、预览、选择、保存和刷新；真实API核对持久化与错误。
- [x] 原版AX读取确认1212/文本模型/3D国漫1；原版输入几何与截图失败，不能执行真实生成/取消对照。
- [ ] 提交证据、规格、差异与验收报告，更新Craft发现、未知与交接。
