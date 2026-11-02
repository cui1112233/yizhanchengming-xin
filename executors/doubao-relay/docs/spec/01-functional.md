# 01 · 功能需求

## 需求编号约定

- `F<n>` 表示功能需求(Functional)
- `F<n>.<m>` 表示子项
- MUST / SHOULD / MAY 用 RFC 2119 语义

## F1 · 豆包生命周期管理

### F1.1 · 自动发现豆包安装路径
MUST 支持从以下来源发现 Doubao.exe / Doubao.app:
- **Windows**: 已运行进程 `Path`、HKCU/HKLM `App Paths`、HKCU/HKLM `Uninstall` + DisplayName 包含"豆包|Doubao"、`%LOCALAPPDATA%\Doubao\app\`、`%LOCALAPPDATA%\Programs\Doubao\`、各盘根 `\Doubao\app\`
- **macOS**: `/Applications/*Doubao*.app`、`~/Applications/*Doubao*.app`、`mdfind "kMDItemFSName == 'Doubao.app'"`

### F1.2 · 启动豆包带调试端口
MUST 用以下命令启动:
- **Windows**: `Start-Process Doubao.exe --remote-debugging-port=9707 --remote-debugging-address=127.0.0.1 --force-renderer-accessibility=complete`
- **macOS**: `open -a Doubao --args --remote-debugging-port=9707 --remote-debugging-address=127.0.0.1`

### F1.3 · 杀掉已运行的豆包
在 F1.2 启动前,如果已有豆包进程但 9707 端口未开 → MUST 优雅关闭(CloseMainWindow + 5 秒超时 + Kill)。

### F1.4 · 探测 CDP 可用性
启动后 MUST 在 20 秒内轮询 `http://127.0.0.1:9707/json/version` 直到成功。超时则报告 `startup_timeout`。

### F1.5 · 持续健康检查
服务运行期间 MUST 每 30 秒探测一次 CDP 连通性,断开后 MUST 尝试重连(F1.1-F1.4 全流程)。

## F2 · 账号管理

### F2.1 · 列出所有已登录账号
MUST 返回豆包左下角账号菜单中可见的所有账号。每个账号包含:
- `name`: 账号显示名
- `subscription`: 套餐(体验/标准/加强/高级/专业/基础/null)
- `label`: 完整标签(用于内部匹配)
- `is_current`: bool

### F2.2 · 识别当前账号
MUST 返回当前前台激活的账号 name。

### F2.3 · 切换账号
MUST 支持指定 target account 切换。切换行为:
- MUST 在豆包 UI 上模拟点击,不绕过
- MUST 验证切换成功(20 秒窗口内目标账号 label 匹配)
- MUST 处理切换导致的页面 reload + CDP target 重连

### F2.4 · 账号同名处理
当两个账号 name 相同时 MUST 在列表里标注 `is_ambiguous: true`,API 调用用 name 引用时返回 `account_ambiguous`。

### F2.5 · 刷新账号列表
用户在豆包里添加/删除账号后,MUST 在下次 `list_accounts` 调用时反映。SHOULD 支持手动"刷新"按钮触发。

## F3 · 任务提交

### F3.1 · 提交生图任务
MUST 支持以下参数:
- `account`(必填): 账号 name
- `type`: `"image"`
- `prompt`: 1-10000 字符
- `model_name`: 白名单内 — Seedream 5.0 Flash / Pro, Seedream 4.5
- `ratio`: 白名单内 — 自动, 1:1, 16:9, 9:16, 4:3, 3:4, 3:2, 2:3, 21:9
- `reference_files`: 0-10 个图片资源 ID(F6.1 预上传返回)
- `confirmation_mode`: `"auto"` / `"manual"`,默认 `"auto"`
- `auto_use_membership`: bool,默认 `true`
- `callback_url`(可选): 任务完成后 POST 的 webhook URL

### F3.2 · 提交生视频任务
在 F3.1 基础上额外:
- `type`: `"video"`
- `model_name`: Seedance 2.5 / 2.0 / 2.0 Fast / 2.0 Mini
- `duration`: 4-15 秒(Seedance 2.5 支持 4-30)
- `reference_files`: 0-10 个(Fast 支持 0-11)

### F3.3 · 任务 ID
MUST 返回唯一 `task_id`(ULID),调用方用这个 ID 查询状态。

### F3.4 · 入队
MUST 按账号串行、全局最多 3 个并发(可配置)。超过时任务进入 pending 队列。

### F3.5 · 请求校验
MUST 在接受请求前校验:
- 必填字段齐全
- 枚举值合法
- 账号存在且已登录
- `reference_files` ID 真实且文件未过期
- callback_url(如有)是合法 http(s)

校验失败 MUST 立即返回 400 + 错误码,**不入队**。

## F4 · 任务生命周期

### F4.1 · 状态机
MUST 按以下状态流转:
```
pending → preparing → generating → completed
                   ↘ failed
                   ↘ cancelled
```
详细见 `08-task-lifecycle.md`。

### F4.2 · 阶段细化
`generating` 阶段 MUST 再细分为:
- `checking` — 豆包进入参数确认
- `quota_checking` — 豆包进入会员配额确认
- `rendering` — 真正生成中
- `rendering_unverified` — 豆包只回"正在渲染"但未真开始

### F4.3 · 任务查询
MUST 支持 `GET /tasks/{id}` 返回当前状态 + 阶段 + 消息 + 结果(如完成)。

### F4.4 · 任务取消
MUST 支持 `DELETE /tasks/{id}` 取消 pending 或 preparing 状态任务。已 generating 的任务取消时 MUST 最大努力中断,但不保证能停下豆包的生成。

### F4.5 · 多轮确认处理
豆包返回"是否继续生成?"类提示时:
- `confirmation_mode=auto`: MUST 自动代回复"继续生成"
- `confirmation_mode=manual`: MUST 把任务置于 `waiting_user_confirm` 状态,等待外部 `POST /tasks/{id}/confirm` 触发

### F4.6 · 超时
任务 MUST 有总超时:
- 图像:5 分钟
- 视频:20 分钟
超时 → 状态 `failed`,错误码 `task_timeout`。

## F5 · 结果获取

### F5.1 · 结果元数据
`GET /tasks/{id}` 完成态 MUST 返回:
```json
{
  "result": {
    "files": [
      { "index": 0, "mime": "image/png", "width": 1920, "height": 1080, "bytes": 2456321, "download_url": "/api/v1/tasks/.../download/0" }
    ],
    "watermark_source": "conversation_original" | "lr_unwatermarked" | "fallback_decrypted"
  }
}
```

### F5.2 · 结果下载
MUST 支持 `GET /tasks/{id}/download/{index}` 流式返回文件内容。
- `Content-Type` 对应 MIME
- `Content-Disposition: attachment; filename="..."`
- 支持 `Range` 请求(视频断点续传)

### F5.3 · 结果保留期
生成文件 MUST 本地保留至少 24 小时。超期后 MAY 清理,API 返回 `result_expired`。

### F5.4 · 结果元数据保留期
任务元数据(status / stage / error 等) MUST 保留至少 7 天。

## F6 · 参考图上传

### F6.1 · 预上传
MUST 支持 `POST /uploads` multipart/form-data 上传参考图,返回 `{ upload_id, expires_at }`。
- 文件大小 ≤ 50 MB
- MIME ∈ { image/png, image/jpeg, image/webp, video/mp4(视频参考)}
- `expires_at` 1 小时

### F6.2 · 上传后使用
`POST /generate` 的 `reference_files` 字段用 upload_id 引用。使用后 upload_id MAY 被清理(或任务完成后清理)。

## F7 · UI 展示

### F7.1 · 账号列表页
MUST 展示所有已登录账号 + 当前账号标记 + 套餐。

### F7.2 · 任务列表页
MUST 展示进行中和近期完成任务。

### F7.3 · 任务详情页
MUST 展示任务的 stage 变化历史、豆包反馈消息、生成结果(如完成时的缩略图)。

### F7.4 · 设置页
MUST 支持配置:
- HTTP API 监听端口(默认 8787)
- Bearer Token(自动生成,可重新生成)
- 全局并发数(默认 3)
- 日志级别

### F7.5 · 系统托盘
MAY 在系统托盘有图标,显示当前运行中任务数。关闭主窗口 SHOULD 变成最小化到托盘(保持 HTTP 服务运行)。

## F8 · 可观测

### F8.1 · 日志
MUST 输出结构化日志到:
- 控制台(开发模式)
- 日志文件 `~/Library/.../logs/` 或 `%APPDATA%\...\logs\`
- 日志滚动(单文件 10 MB,保留 7 天)

### F8.2 · 任务审计
每个任务 MUST 记录完整的阶段变化历史(stage + ts + message)。

### F8.3 · 调试面板
MAY 在 UI 里有"调试"页,显示 CDP 命令发送/接收历史(便于排查豆包改版问题)。

## 功能优先级

| 优先级 | 需求 | 阶段 |
|---|---|---|
| P0 | F1 启动豆包 + F2.1-F2.3 账号读取 | Phase 1 |
| P0 | F3.1 生图提交 + F4 任务生命周期 + F5 结果获取 | Phase 2 |
| P1 | F3.2 生视频 + F6 参考图上传 | Phase 3 |
| P1 | F4.5 多轮确认 + F3.5 请求校验 | Phase 4 |
| P2 | F7 UI 全套 | Phase 1-5 |
| P2 | F2.4 同名 + F2.5 刷新 + F8 可观测 | Phase 5-6 |
