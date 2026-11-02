# 14 · 验收标准

## 验收等级

| 等级 | 必须满足 |
|---|---|
| **MVP (Phase 1)** | 账号读取 + UI 展示 |
| **Alpha (Phase 2)** | 单账号单任务生图,HTTP API 可用 |
| **Beta (Phase 3+4)** | 生视频 + 多轮确认 + 多任务并发 |
| **RC (Phase 5)** | macOS 双端打通 |
| **1.0 (Phase 6)** | 稳定性 + 打包签名 |

## 测试矩阵

每条验收用例标记:
- `[P0]` 阻塞发布
- `[P1]` 高优,SHOULD 过
- `[P2]` 中优,MAY 延后
- `[P3]` 低优,nice to have

## Phase 1 · MVP 验收

### UC-01 [P0] · 启动豆包连接
**前置**:豆包未开 / 开着但 CDP 没开 / 开着且 CDP 已开(三种状态)
**操作**:启动 doubao-relay,点"连接豆包"
**预期**:
- 若豆包未开 → 自动找到路径启动,20 秒内 CDP 就绪
- 若豆包开着无 CDP → 杀掉重启
- 若 CDP 已开 → 直接复用
- 任何场景失败都有明确中文错误消息

### UC-02 [P0] · 账号列表读取
**前置**:豆包已连 + 已登录 1 个账号
**操作**:主界面展示账号
**预期**:
- 账号名正确
- 套餐名正确(体验/标准/高级 等)
- "当前账号"标记正确

### UC-03 [P0] · 多账号读取
**前置**:豆包已登录 2+ 账号
**操作**:同 UC-02
**预期**:全部账号展示;顺序一致;"更多账号"子菜单的也读到

### UC-04 [P0] · 账号切换
**前置**:豆包有 2+ 账号
**操作**:点非当前账号的"切换"
**预期**:
- 弹确认框
- 确认后 20 秒内切换成功
- UI 刷新显示新的"当前账号"
- 切换过程中 UI 显示"切换中"loading

### UC-05 [P1] · 未登录处理
**前置**:豆包已连但未登录任何账号
**操作**:主界面
**预期**:显示空状态"豆包尚未登录",附操作建议

### UC-06 [P1] · 豆包崩溃处理
**前置**:正常运行中
**操作**:手动杀 Doubao.exe
**预期**:
- UI 15 秒内检测到并显示"豆包已退出,等待重启"
- 底部状态栏变红
- 用户点"重连"按钮能恢复

### UC-07 [P2] · 平台切换(同一台机器试双端)
**操作**:在 Mac 上跑一遍 UC-01..06,再在 Windows 上跑一遍
**预期**:行为一致,UI 一致

## Phase 2 · Alpha 验收

### UC-10 [P0] · 单次生图(API)
**前置**:账号已登录,配额充足
**操作**:
```bash
curl -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"account":"张三","type":"image","model_name":"Seedream 5.0 Pro",
       "prompt":"一只蓝色小鸟","ratio":"16:9"}' \
  $BASE/api/v1/generate
```
**预期**:
- 立即返回 task_id
- 轮询 `/tasks/{id}` 看到 stage 变化
- 3 分钟内 status=completed
- `GET /tasks/{id}/download/0` 返回有效 PNG

### UC-11 [P0] · 账号不存在
**操作**:POST /generate 带不存在的 account
**预期**:404 + `account_not_found`

### UC-12 [P0] · 账号未登录
**操作**:POST /generate 用一个 /accounts 里但 is_current=false 的账号
**预期**:自动切换,然后开始生成

### UC-13 [P0] · 参数校验
**操作**:
- prompt 空 → 400 `request_missing_field`
- model 不在白名单 → 400 `request_invalid_enum`
- ratio 不在白名单 → 400 `request_invalid_enum`
- prompt 超 10000 字 → 400 `request_prompt_too_long`
- 参考图 > 10 张 → 400 `request_too_many_references`
**预期**:全部按错误码返回,不入队

### UC-14 [P0] · 任务查询
**前置**:有正在跑的任务
**操作**:GET /tasks/{id}
**预期**:
- 返回当前 status + stage
- stage_history 完整
- 轮询时字段变化

### UC-15 [P1] · 任务取消
**操作**:提交 1 个任务,立即取消(pending 状态下)
**预期**:200 + cancelled

### UC-16 [P1] · 任务排队
**操作**:同账号连续提交 5 个任务
**预期**:
- 全部成功入队
- queue_position 递增
- 串行执行,一次只跑 1 个
- 全部完成后结果都正确

### UC-17 [P1] · 鉴权
**操作**:无 token / 错 token 调 API
**预期**:401

### UC-18 [P1] · 结果下载
**操作**:GET /tasks/{id}/download/0
**预期**:
- Content-Type 正确
- Content-Length 匹配
- 下载的文件可用

### UC-19 [P2] · 结果过期
**前置**:等 24+ 小时
**操作**:下载旧任务结果
**预期**:404 `result_expired`,元数据仍在

## Phase 3 · Beta 验收(视频 + 多轮确认)

### UC-20 [P0] · 单次生视频
**操作**:POST /generate `type=video, model=Seedance 2.0, duration=5`
**预期**:15 分钟内 completed,下载 mp4 可播放

### UC-21 [P0] · 视频无水印三条路径
**手动测试**:
- 场景 A:豆包返回带 lr=unwatermarked 的直链 → `watermark_source=lr_unwatermarked`
- 场景 B:需要点"下载"按钮 → `watermark_source=lr_unwatermarked`
- 场景 C:返回加密 token → 本地 AES 解密 → `watermark_source=fallback_decrypted`

**预期**:三种场景的 mp4 都无水印

### UC-22 [P0] · 多轮确认(自动)
**前置**:选个会触发"是否使用会员配额"确认的模型
**操作**:提交任务,`confirmation_mode=auto`
**预期**:
- stage 流经 `checking` → `quota_checking` → `generating`
- 自动代回复"继续生成"
- 最终 completed

### UC-23 [P1] · 多轮确认(手动)
**操作**:`confirmation_mode=manual`
**预期**:
- stage 到 `waiting_user_confirm` 停住
- `POST /tasks/{id}/confirm` 后继续
- 1 小时未 confirm → failed `confirm_timeout`

### UC-24 [P1] · 多任务并发(全局)
**操作**:3 个不同账号各提交 1 个任务
**预期**:3 个并发跑,全局并发 3 的硬限

### UC-25 [P1] · 超额并发
**操作**:短时提交 51 个 pending
**预期**:第 51 个返回 429 `queue_full`

### UC-26 [P1] · Webhook 回调
**前置**:启本地 HTTP server 接收
**操作**:提交带 callback_url 的任务
**预期**:任务完成时收到 POST,签名正确

### UC-27 [P2] · 参考图上传与引用
**操作**:POST /uploads 上传 PNG → 用 upload_id 发 /generate
**预期**:生成结果采用了参考图

### UC-28 [P2] · 用户干扰检测
**前置**:任务 generating 中
**操作**:用户在豆包里新建对话(新开 chat 页)
**预期**:
- 任务检测到并 failed `multiple_chat_pages`
- 不污染结果

## Phase 4 · 冲突/异常验收

### UC-30 [P0] · 账号配额用完
**前置**:账号已用完当日配额
**操作**:提交任务
**预期**:任务 failed `account_quota_exhausted`

### UC-31 [P0] · 豆包页 reload
**操作**:切换账号(会 reload) → 立即提交任务
**预期**:
- 切账号成功
- 新任务 CDP 重连后正常跑

### UC-32 [P1] · 豆包异常退出
**操作**:生成中手动杀豆包
**预期**:
- 正在跑的任务 failed `doubao_crashed`
- 新 /generate 返回 503

### UC-33 [P1] · relay 重启后任务状态
**前置**:有正在跑的任务
**操作**:杀 doubao-relay,重启
**预期**:原任务 failed `relay_restart`,新任务可继续提交

### UC-34 [P2] · UI 兼容性自检失败
**模拟**:注入一段 JS 让 compat check 返回 fail
**预期**:API 返回 503 `doubao_ui_changed`,UI 红色 banner

## Phase 5 · RC 验收(双端)

### UC-40 [P0] · 跨平台一致性
**操作**:同样的任务在 Windows 和 macOS 上各跑一次
**预期**:
- API 行为一致
- UI 行为一致
- 结果等效(可能文件字节不完全相同,但 MIME/尺寸/内容应一致)

### UC-41 [P0] · 平台特定路径
**操作**:启动时读 app_data_dir
**预期**:
- Windows: `%APPDATA%\doubao-relay`
- macOS: `~/Library/Application Support/doubao-relay`
- 不走错位置

### UC-42 [P1] · macOS Gatekeeper
**前置**:未签名的 app
**操作**:首次打开
**预期**:系统弹"未知开发者"提示;用户授权后能开(或者签了名就无此提示)

## Phase 6 · 1.0 验收

### UC-50 [P0] · 24 小时稳定运行
**操作**:启动后挂机 24 小时
**预期**:
- 内存不泄露(RSS 不超过 300 MB)
- CPU 空闲时 < 2%
- 日志无 ERROR 级别

### UC-51 [P0] · 大量任务
**操作**:1 小时内提交 100 个任务
**预期**:
- 全部成功处理
- SQLite 查询仍快(< 500ms)
- 完成后清理机制启动

### UC-52 [P0] · 升级
**操作**:v0.9 → v1.0
**预期**:
- 配置保留
- SQLite 自动 migrate
- 老任务数据可查

### UC-53 [P1] · 日志清理
**前置**:日志超过 10 MB
**操作**:观察自动滚动
**预期**:新文件创建,老文件保留 7 天后清

### UC-54 [P1] · 任务结果清理
**前置**:有超过 24 小时的任务结果
**操作**:触发 GC
**预期**:文件被删,SQLite 元数据保留

## 非功能性验收

### NF-01 [P0] · 响应时间
GET /tasks/{id} < 50ms(P50),< 500ms(P99)

### NF-02 [P0] · 资源占用
见 02-non-functional.md N1.3,不超预算

### NF-03 [P0] · 启动时间
从双击到 UI < 3 秒;从 UI 到豆包就绪 < 20 秒

### NF-04 [P0] · 安全
- 外部端口只 loopback
- 无 token 访问全部 401
- 日志无 bearer_token 出现(info 级别)

### NF-05 [P1] · 可观测
- 每条 API 请求有一条 info 日志
- 每次任务状态变化有一条 info 日志
- debug 级别有 CDP 命令日志

## 回归测试矩阵

每次发版前必跑:

| 场景 | Windows | macOS |
|---|---|---|
| UC-01 启动豆包 | ✓ | ✓ |
| UC-02 账号列表 | ✓ | ✓ |
| UC-04 账号切换 | ✓ | ✓ |
| UC-10 单次生图 | ✓ | ✓ |
| UC-13 参数校验 | ✓ | ✓ |
| UC-20 单次生视频 | ✓ | ✓ |
| UC-22 多轮确认(自动) | ✓ | ✓ |
| UC-30 配额用完 | 一端跑即可 | — |
| UC-40 跨平台一致性 | ✓ | ✓ |

## 回归测试工具

- Postman / HTTPie 收藏夹:覆盖所有 API 端点
- 手动用例脚本:`tests/manual/UC-*.md`
- 自动化(可选):Playwright 跑 WebView UI,配合 mock CDP server

## 验收通过的判据

一个 Phase MUST 通过时,**所有 P0 用例全绿**。P1 允许有已知问题但 MUST 有 workaround。P2 可延后。

每个 release 的 CHANGELOG MUST 记录:
- 通过了哪些用例
- 已知问题列表
- 建议的使用边界
