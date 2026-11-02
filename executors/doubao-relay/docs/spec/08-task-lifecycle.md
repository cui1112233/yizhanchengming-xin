# 08 · 任务生命周期与并发模型

## 任务状态机

```
                    POST /generate
                           │
                           ▼
                      ┌─────────┐
          cancel ◄──  │ pending │
                      └────┬────┘
                           │  调度器挑中
                           ▼
                      ┌───────────┐
          cancel ◄──  │ preparing │
                      └────┬──────┘
             失败           │   成功到达"等生成"
          ┌──────────┐     ▼
          │  failed  │ ◄─ ┌─────────────┐
          └──────────┘    │ generating  │  ◄──┐
          ▲               └─────┬───────┘     │
          │                     │             │
          │                     │  manual     │
          │                     │  模式       │
          │                     ▼             │
          │               ┌──────────────┐    │
          │               │ waiting_user │    │
          │               │   _confirm   │────┘ (confirm 后回到 generating)
          │               └──────┬───────┘
          │                      │
          │             豆包返回结果   豆包拒绝/超时
          │                      │           │
          │                      ▼           ▼
          │               ┌───────────┐   ┌────────┐
          │               │ completed │   │ failed │
          │               └───────────┘   └────────┘
          │                      ▲
          │                      │  cancel
          │               ┌───────────┐
          └────────────── │ cancelled │
                          └───────────┘
```

## 状态与子阶段

### `pending`

- 入口:`POST /generate` 校验通过后
- 退出:调度器 pick 到(→ `preparing`)或 cancel(→ `cancelled`)
- 字段:只有 `queue_position` 是动态的

### `preparing`

- 入口:调度器 pick 到
- 行为:
  1. 若当前账号 ≠ 任务指定账号 → 调 `account.switch`
  2. 若需要上传参考图 → 调 `fileInput` + `uploadReady`
  3. 调 `prepare`(选模型/比例)
  4. 若视频:调 `durationDragPlan`
  5. 调 `fill`(填 prompt)
  6. 调 `checkSend`
  7. 调 `send`
- 退出:
  - 全部成功 → `generating` (stage=`sent`)
  - 任何步骤 throw → `failed`
  - cancel → `cancelled`

**超时**:整个 preparing 阶段 60 秒超时。

### `generating`

- 入口:preparing 成功
- 行为:启动 capture 循环,每秒重放 history 请求 + 每 500ms 轮询 feedback
- 退出:
  - 识别到结果 → `completed`
  - 豆包拒绝/配额用完 → `failed`
  - 超时(图 5 min,视频 20 min)→ `failed`
  - 用户 cancel → `cancelled`
  - manual 模式遇到确认 → `waiting_user_confirm`

**子阶段**(`stage` 字段):
- `sent` → 刚发送,等豆包回应
- `checking` → 豆包"是否继续生成?"参数确认
- `quota_checking` → 豆包"是否使用会员配额?"
- `rendering` → 豆包真正开始生成
- `rendering_unverified` → 豆包只回"正在渲染"(30 秒内未进入 rendering 则判 `not_executed`)
- `downloading` → 结果已返回,正在下载
- `result` → 结果已识别,等入 completed

### `waiting_user_confirm`

- 入口:`generating/checking` 或 `generating/quota_checking` + `confirmation_mode=manual`
- 行为:暂停自动回复,等待 `POST /tasks/{id}/confirm`
- 退出:
  - confirm → 回到 `generating`,代回复"继续生成"
  - cancel → `cancelled`
  - 1 小时无 confirm → `failed` (超时)

### `completed`

- 终态
- 字段:`completed_at` + `result.files` + `watermark_source`

### `failed`

- 终态
- 字段:`failed_at` + `error.code` + `error.message` + `error.details`
- 可能的 error.code:见 `11-error-codes.md`

### `cancelled`

- 终态
- 字段:`cancelled_at` + `cancelled_from_status`

## 并发模型

### 全局约束

| 约束 | 默认 | 配置项 |
|---|---|---|
| 全局最大并发(preparing + generating) | 3 | `concurrency.global_max` |
| 单账号并发 | 1 | `concurrency.per_account_max` |
| 队列最大长度(pending + waiting_user_confirm) | 50 | `concurrency.queue_depth_max` |

**说明**:
- `per_account_max=1` 是硬约束(豆包同一账号只能同一时间跑一个任务,UI 层限制)
- `global_max=3` 是软约束(电脑性能和豆包风控考虑)
- 队列满 → `POST /generate` 返回 429 `queue_full`

### 调度算法

**最简单的策略(v0.1)**:FIFO,严格按入队顺序。

```
伪代码:
for each new pending task:
    if running.count >= global_max: wait
    if account_lock[task.account].locked: wait
    acquire account_lock
    acquire global_semaphore
    spawn task
    on finish: release both locks
```

**不做**:优先级队列、公平调度、短任务优先。保留扩展位。

### 账号互斥

**每个账号一把锁**。切换账号时额外 take 一把"切账号锁",阻塞所有新任务:

```rust
struct AccountMutex {
    inner: HashMap<String, Arc<Mutex<()>>>,
    switch_lock: Arc<Mutex<()>>,
}

async fn with_account(&self, account: &str, f: impl Future) {
    let _switch = self.switch_lock.lock().await;  // 切账号期间阻塞
    let lock = self.inner.entry(account.into()).or_default();
    let _guard = lock.lock().await;
    f.await
}
```

### 任务冲突检测

调度器在 preparing 阶段开始前,主动探测:
1. doubao-chat 页数量 ≠ 1 → `multiple_chat_pages`,终止任务
2. 当前账号 ≠ 任务指定账号 → 自动切换
3. 账号切换失败 → `account_switch_failed`,终止任务

在 generating 阶段每 10 秒再探测:
1. doubao-chat 页数量 ≠ 1 → `user_created_multiple_chats`,终止任务并标记 `user_interrupted`
2. 当前输入框内容 ≠ 预期的 prompt → `user_modified_input`,**warn 但不终止**(可能是用户看了几眼)

## 事件广播

每个状态变化 MUST 触发:

1. **tracing 日志**:info 级别,结构化字段 `{task_id, from_status, to_status, stage, duration_ms}`
2. **Tauri 事件**:`task_updated` 推送给 WebView
3. **数据库写**:update tasks + insert task_stage_history
4. **Webhook**(终态时):POST callback_url

**顺序**:db write 先完成,再广播。保证订阅者看到的 db 状态和事件一致。

## 恢复策略

**app 重启时**:
1. 读 SQLite 所有非终态任务(`status IN ('pending','preparing','generating','waiting_user_confirm')`)
2. **全部置为 `failed`**(error.code = `relay_restart`)
3. 原因:中转程序重启后,豆包的状态已经不可信,强行恢复 generating 中的任务是危险的(可能重复识别别人的结果)

**豆包崩溃时**(CDP 断开):
1. 当前所有非终态任务置为 `failed`(error.code = `doubao_crashed`)
2. 新请求直接返回 503
3. 重连成功后恢复接受

## 取消的最大努力

### 从 pending 取消
简单,从队列删除,立即 `cancelled`。

### 从 preparing 取消
需要能中断正在跑的 CDP 命令。实现:preparing 整个流程包在 `CancellationToken` 下,检查点在每个 CDP 调用之间。
- 风险:prompt 已经填进去了,点"发送"之前取消 → 用户豆包里会看到残留文字
- 补偿:取消时 MAY 调 `fill` with empty string 清空(best-effort)

### 从 generating 取消
最复杂。豆包的生成**无法从外部取消**(没有 API)。实现:
- 立刻停止 capture 循环
- 不下载后续结果
- 任务标记 `cancelled`
- 用户的豆包账号配额**仍然被消耗**(警告用户)

## 超时处理

| 阶段 | 超时时长 | 超时后 |
|---|---|---|
| pending(未被 pick) | 1 小时 | `failed/queue_timeout` |
| preparing | 60 秒 | `failed/preparing_timeout` |
| generating(图像) | 5 分钟 | `failed/task_timeout` |
| generating(视频) | 20 分钟 | `failed/task_timeout` |
| waiting_user_confirm | 1 小时 | `failed/confirm_timeout` |

超时的检测由后台定时器每 10 秒跑一次。

## 任务 ID 生成

```rust
use ulid::Ulid;
let task_id = Ulid::new().to_string();  // 26 字符,时序可排
```

ULID 的好处:
- 排序:按生成时间自然有序,数据库索引友好
- 唯一:同一毫秒多个任务也不重复
- 可读:比 UUID 短
- 跨进程安全

## 流程图(典型成功路径)

```
T0:     POST /generate              → task 入 pending, task_id 返回给调用方
T0+1s:  调度器 pick → preparing
T0+2s:  检查账号 → 已是目标账号,跳过切换
T0+2s:  prepare(选模型 Seedream 5.0 Pro) ✓
T0+3s:  (不是视频,跳过时长设置)
T0+3s:  (无参考图,跳过上传)
T0+3s:  fill(prompt 填入)                ✓
T0+3s:  checkSend                         ✓
T0+3s:  send                              ✓ → stage=sent
T0+4s:  capture 循环启动,subscribe Network
T0+5s:  收到 submissionIdentity:local_message_id=xxx
T0+7s:  第一次历史重放,boundConversationId 识别
T0+15s: feedback stage=checking           → stage=checking
T0+16s: 自动回复"继续生成"
T0+17s: feedback stage=generating         → stage=rendering
T0+65s: feedback stage=result            → stage=downloading
T0+66s: image_ori_raw.url 识别到
T0+68s: 下载完成,写文件,insert result_files → completed
T0+68s: Tauri event + Webhook POST + Save DB
T0+68s: 调用方下一次 GET /tasks/{id} 收到 status=completed
```

## 异常路径示例

### 豆包配额用完

```
T0:     提交 → pending → preparing → send ✓
T0+5s:  feedback stage=refused, message="当前账号额度不足"
        → status=failed, error.code=account_quota_exhausted
```

### 用户同时在豆包里打字

```
T0:     提交 → preparing → send ✓
T0+10s: 调度器下一轮 check:当前输入框内容 ≠ 预期
        → warn 日志,status 不变,继续等结果
T0+65s: 结果识别到(如果没被用户污染)→ completed
```

### 账号 reload 导致 CDP 断

```
T0:     提交 → preparing(切账号过程中)
T0+5s:  切换目标账号 → 豆包页面 reload
T0+6s:  CDP target id 变化,capture socket 断开
T0+6s:  重连逻辑接管:fresh /json/list → 重连新 target
T0+8s:  重连成功 → 继续 preparing
...
```
