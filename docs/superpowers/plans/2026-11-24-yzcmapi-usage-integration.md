# YZCM API 用量与制作 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** 接通管理员共享 YZCM API 账单，并准确记录组员在本站的调用归属。

**Architecture:** Go 的 `memberusage` 服务负责上游只读查询、MySQL 调用账本和去重；个人中心接口负责身份范围与加密绑定，React 保留现有主题展示真实结果。共享上游账单只展示一次，组员数据按实际操作者区分。

**Tech Stack:** Go、MySQL、现有 AES-GCM 配置密钥、React/Antd；复用现有网络调度，不新增部署技术。

## Global Constraints

- 管理员配置和使用 API，组员调用所属管理员的 API。
- 历史无法分人的数据保留未归属，不平摊、不复制管理员总额给组员。
- 普通成员只读本人，主管只读所属组员，开发者沿用现有管理范围。
- 密钥仅保存在后端密文中，不进入前端返回、日志或仓库。
- 使用北京时间区间；读取上游换算配置，缺数据及接口失败不伪装成零。
- 主开发分支工作，不创建新分支；保留其他任务未提交改动。
- 本地 4001 验证后再同步远端和线上；不产生付费生成测试。

## Task 1: 上游只读客户端和调用账本

**Files:** `backend/internal/memberusage/{types,client,repository,summary}.go` 及对应测试。

- [ ] 定义绑定、调用事件、汇总和上游结果接口，覆盖成功/失败/免费/未知费用。
- [ ] 先测上游鉴权、时区、金额换算、错误和分页，再实现固定站点只读查询。
- [ ] MySQL 保存加密绑定和调用事件，按请求编号去重，按付费管理员与实际用户隔离。
- [ ] 覆盖数据库聚合及多用户隔离测试。

## Task 2: 调用归属采集

**Files:** `backend/internal/memberusage/{context,transport}.go`、`backend/internal/imagecapacity/http.go` 及测试。

- [ ] 仅观察 YZCM API 的实际模型请求，不采集普通网页、轮询、统计接口。
- [ ] 复用现有调用身份，HTTP 请求与后台任务都保留真实操作者。
- [ ] 非流式与流式读取均保留原响应，不保存提示词、正文或密钥。
- [ ] 实际调用和重试分别计量；重复保存事件幂等，未知数据明确标记。

## Task 3: 个人中心接口与配置接线

**Files:** `backend/internal/httpapi/member_usage*.go`、`member_profile.go`、`member_team_read_compat.go`、`router.go`、`backend/internal/app/app.go`、`backend/internal/storage/member_usage_schema.go`。

- [ ] 先测本人/主管/开发者访问边界及凭证不回传，再实现 GET/PUT/DELETE `/api/member/usage/connection`。
- [ ] 绑定前只读验证管理员上游身份，复用配置加密密钥，错误时保留原有效绑定。
- [ ] `/api/member/me` 和组员用量返回真实汇总和状态；组员不能指定他人或主管身份。
- [ ] 接入现有出站调度装饰器，保留容量、超时、SSRF 和重定向保护。
- [ ] 核对最大迁移号后追加新表迁移；本地启动验证实际表结构。

## Task 4: 页面展示与绑定入口

**Files:** `frontend/src/user/pages/{UsageStatsPage,UsageConnectionPanel}.jsx`、`frontend/src/shared/api/member.js` 及必要样式/测试。

- [ ] 管理员可绑定上游统计账号，组员只看到继承关系。
- [ ] 展示真实账单来源、更新时间、未归属数据及异常，费用未知时不用零替代。
- [ ] 组员排行只累加本人成员数据，制作量与 API 调用次数分开。
- [ ] 浏览器验证绑定、取消、失败与重试，保持现有个人中心主题。

## Task 5: 联调与交付

- [ ] 审查每个子任务的实现及权限、凭证、去重等交叉边界。
- [ ] 运行相关 Go/前端检查及干净构建，更新本地 4001。
- [ ] 若已有授权会话或统计凭证，使用只读真实数据核对；若缺少凭证，明确展示待绑定而不是伪造成功。
- [ ] 记录验证结果并按本地、远端、线上顺序交付。
