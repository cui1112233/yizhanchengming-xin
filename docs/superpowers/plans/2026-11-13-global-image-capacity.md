# 全站图片并发 Implementation Plan

> **For agentic workers:** Execute inline with superpowers:executing-plans, task by task. No new branch or parallel agents; preserve unrelated working changes.

**Goal:** 全站按上游账号和模型共享图片并发，已验证请求初始 12、硬上限 20，未知请求维持 4。

**Architecture:** imagecapacity 独立模块保存 MySQL 策略，通过现有 Redis 传输执行原子领取、续租、释放、轮转等待和自适应。safehttp 的应用级出站包装统一拦截图片生成，httpapi 图片操作上下文使名额持续到完整结果处理完毕；后台任务同样经过包装。用户只看状态，管理员修改策略。

**Tech Stack:** Go、MySQL、Redis Lua、React/Antd；Go embed。

## Global Constraints

- 50 次实测预算已用完，禁止额外真实模型测试。
- 仅已验证的账号/模型/尺寸/无参考图方式启用 20；未知方式维持 4。
- 0.019 与 0.034 共用账号总池；普通用户不能通过别名、配置或并发选项抬高上限。
- Redis 不可用不能直连；响应头到达不提前释放；未知在途保守保留占用。
- 只提交本次文件/行，前端现有未提交修改必须保留。本地 4001 验证后才能发布。

### Task 1: 策略与 Redis 控制器

Files: backend/internal/imagecapacity/{policy.go,mysql.go,redis.go,http.go,*_test.go}; backend/internal/storage/image_capacity_schema.go。

- [ ] 先写失败测试：两个控制器共用 Redis 的账号/模型峰值、取消等待、重复归还、租约保留、不同模型共用账号、相同凭据别名不拆池；策略匹配尺寸和参考图；普通失败不触发升降，429 冷却降速，稳定成功按步恢复。
- [ ] 接口：Request{Endpoint,Credential,Model,Size,References,Owner,Book}; PolicyRepository.Resolve(ctx,Request) / Save(ctx,Policy)；Controller.Acquire(ctx,Request) 返回 Lease；Lease.Complete(Outcome)。
- [ ] Lua 使用一个账号键槽下的在用租约、模型计数、等待队列和运行值；每个请求唯一 token，轮转 owner/book，释放幂等。过期租约先隔离到供应商安全期限，不能无限重发。
- [ ] MySQL 保存账号分组及请求方式策略，凭据只保存不可逆标识。查询返回限额及是否实测，不返回凭据。
- [ ] 用真实隔离 Redis 与模拟时钟/上游验证，禁止清空共享数据库。

### Task 2: 所有图片出站链路

Files: backend/internal/safehttp/client.go；backend/internal/httpapi/image_capacity.go；shuihuo_projects.go、integration_channel_exec.go、batch_factory_v12_image.go、reference_assets_image.go；backend/internal/app/app.go。

- [ ] 先写失败测试：提前回响应头但正文阻塞时仍占用；多入口同时调用不超限；Redis 故障无上游调用；上下文取消排队无副作用；不改变文本/视频请求。
- [ ] safehttp 提供 SetOutboundDecorator；应用启动注入 Controller 包装。只拦截已标记图片操作或标准图片生成/编辑路径；解析可重读 JSON/multipart，模型别名按实际 model/credential 归组。
- [ ] httpapi 图片操作注册生命周期，defer 完成后归还；异步/超时未知结果保留占用，不按 HTTP 200 头释放。
- [ ] 统一池启用时绕过旧进程 4 路与主机 4 路闸门，单本工作队列允许请求共享池最大 20；未知模型由新池的 4 路策略保护。

### Task 3: 管理与用户状态

Files: backend/internal/httpapi/image_capacity.go；frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx；frontend/src/user/pages/ImageCapacitySettings.jsx；frontend/src/shared/api/imageCapacity.js。

- [ ] GET /api/config/image-capacity 根据已配置通道解析真实模型/凭据并只返回策略、占用、排队、冷却。
- [ ] 管理员 PUT 同一路由保存全站策略/分组；借用现有管理员鉴权，普通用户只能查询自己的通道状态。
- [ ] 统一配置改为自动分配状态，显示账号/模型当前运行值、硬上限、占用和排队，未知方式注明未实测；管理员入口可调总量。保留他人已有前端改动。
- [ ] 显示等待/冷却而非生成失败；不添加不会生效的按钮。

### Task 4: 验收与交付

- [ ] Redis 真集成、模拟上游、鉴权和前端相关测试通过；整体构建通过。
- [ ] 启动 Go embed 本地 4001，核对版本和登录后真实界面。检查没有生产任务后再替换本地程序。
- [ ] 从实际保存的已测试通道通过管理员接口建立对应 20/12 策略；不得给未知账号或模型套用。
- [ ] 提交本次源码，冻结发布包。沿用已核实 qiantie.service 的版本链接/回退备份流程，不推送 Git，不改用户凭据。
- [ ] 公网只验证策略、版本、占用和页面，不真实生产/重试。记录截图、版本和剩余限制。
