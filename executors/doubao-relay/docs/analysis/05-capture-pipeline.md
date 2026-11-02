# 05 · 结果捕获管线

## 模块目标

**从 `doubao-generate.ps1` 点发送之后,到最终拿到无水印原图/原视频文件之间,整个"等待 + 识别 + 下载"的全过程**。核心文件 `doubao-capture.ps1`(500 行)+ `doubao-result-reader.js`(页面注入)。

## 整体循环结构

`doubao-capture.ps1` 本质是一个 15 分钟的 while 大循环:

```powershell
$deadline = (Get-Date).AddMinutes(15)
while (($isVideo -or (Get-Date) -lt $deadline) -and
       !$script:originals.Count -and                 # 还没拿到结果
       !$generationRefused) {                        # 没被豆包拒绝

    # ① 定期刷新 chat 页连接(账号切换/页面 reload 时需要重连)
    if ((Get-Date) -ge $nextPageCheck) {
        Refresh-ChatConnection
        $nextPageCheck = (Get-Date).AddSeconds(2)
    }

    # ② 处理所有 session 的消息队列
    foreach ($s in $sessions) {
        while ($s.socket.Messages.TryDequeue([ref]$message)) {
            # 根据 event.method 分派处理 ...
        }
    }

    # ③ 定期触发 feedback 轮询(识别"是否继续生成"类提示)
    if ((Get-Date) -ge $nextFeedbackPoll) {
        # 调 shenlangDoubaoGeneration.run('terminalFeedback'/'videoFeedback')
        $nextFeedbackPoll = (Get-Date).AddMilliseconds(500)
    }

    # ④ 定期触发历史重放(主动拉对话快照)
    if ($sendSince -and !$script:originals.Count -and (Get-Date) -ge $nextHistoryPoll) {
        # 用 Runtime.evaluate 执行 shenlangDoubaoResultReader.read(historyRequest, options)
        $nextHistoryPoll = (Get-Date).AddSeconds(1)
    }

    Start-Sleep -Milliseconds 120
}
```

每 120 ms 一轮,四个子任务按时间轮换触发。

## 监听的四种 CDP 事件

启动阶段 `Network.enable` 后,CDP 会推送:

| 事件 method | 时机 | 本工具用途 |
|---|---|---|
| `Network.requestWillBeSent` | 请求即将发出 | 识别"这是一次生成请求"或"这是一次历史查询" |
| `Network.responseReceived` | 响应头到达 | 对视频任务:记下 `video/*` MIME 的直链 URL |
| `Network.loadingFinished` | 响应体加载完 | 触发 `Network.getResponseBody` 读取 body |
| `Network.eventSourceMessageReceived` | SSE 流推送一条消息 | 实时拿到流式生成的中间内容 |

## 事件 1:requestWillBeSent 分类(doubao-capture.ps1:235-264)

```powershell
if ($event.method -eq 'Network.requestWillBeSent') {
    $request = $event.params.request

    # 过滤 1: 跳过自己注入的 shenlangReadBoundConversation 发的请求(避免递归)
    if ($parallelTracking -and
        @($event.params.initiator.stack.callFrames |
          Where-Object { $_.functionName -eq 'shenlangReadBoundConversation' }).Count) {
        continue
    }

    # 过滤 2: 视频任务 - 记下所有在 sendSince 之后的 requestId,用于后续 videoNetworkCandidates
    if ($isVideo -and $sendSince -and
        ([double]$event.params.wallTime * 1000) -ge $sendSince) {
        $s.videoRequests[[string]$event.params.requestId] = $true
    }

    # 分类 A: 这是一次"历史查询"请求 → 记下来作为重放模板
    $uri = $null
    if ($sendSince -and $request.method -eq 'POST' -and
        [uri]::TryCreate($request.url, [UriKind]::Absolute, [ref]$uri) -and
        $uri.Scheme -eq 'https' -and
        $uri.Host -eq 'www.doubao.com' -and
        $uri.AbsolutePath -match '^/im/(chain/single|message/chain_by_block)/?$' -and
        $request.postData) {

        $requestConversationId = Request-ConversationId $request
        # 优先用 /im/chain/single (全消息链),次要 /im/message/chain_by_block (单条块)
        if (!$s.historyRequest -or $uri.AbsolutePath -match '^/im/chain/single/?$') {
            $s.historyRequest = @{
                url = [string]$request.url
                method = 'POST'
                postData = [string]$request.postData
                headers = $request.headers
            }
        }
        $s.historyPending[[string]$event.params.requestId] = $s.historyRequest
    }

    # 分类 B: 这是一次"生成"请求 → 记录 requestId
    $uri = $null
    if ([uri]::TryCreate($request.url, [UriKind]::Absolute, [ref]$uri) -and
        $uri.Scheme -eq 'https' -and
        $uri.Host -match '(^|\.)(doubao\.com|snssdk\.com)$' -and
        $uri.AbsolutePath -match '(completion|generate|/im/message/send)' -and
        $uri.AbsolutePath -notmatch '(history|list|recommend|config)' -and
        $request.method -eq 'POST') {

        # 非并行模式: 不允许同一 session 并发多次生成
        if (!$parallelTracking -and $s.requests.Count -and
            !$s.requests.ContainsKey($event.params.requestId) -and !$isVideo) {
            throw '同时发现多次生成,请一次只生成一张图片。'
        }

        # 并行模式: 没锁定 boundConversationId 时,扫 POST body 识别 identity
        if ($parallelTracking -and !$boundConversationId -and !$submission.submissionIdentity -and
            $request.postData -and ([double]$event.params.wallTime * 1000) -ge $sendSince) {
            $identityRequest = @{
                url = [string]$request.url
                method = 'POST'
                postData = [string]$request.postData
            } | ConvertTo-Json -Depth 12 -Compress
            # 调 submissionIdentity() 在页面侧提取 local_message_id
            $expression = $readerSource +
                ';shenlangDoubaoResultReader.submissionIdentity(' + $identityRequest + ',' +
                ($feedbackRequest.prompt | ConvertTo-Json -Compress) + ')'
            $identityId = Command $s 'Runtime.evaluate' @{
                expression = $expression; returnByValue = $true
            }
            $identityPending[$s.targetId + ':' + [string]$identityId] = $s.page
        }

        $s.requests[$event.params.requestId] = $true
        $report.generationRequests++
    }
}
```

### 关键 URL 规则

**历史查询识别**:
- Host: `www.doubao.com`
- Path: `/im/chain/single` 或 `/im/message/chain_by_block`
- Method: POST
- 带 postData

**生成请求识别**:
- Host: `*.doubao.com` 或 `*.snssdk.com`
- Path: 包含 `completion` / `generate` / `/im/message/send`
- Path: 不包含 `history` / `list` / `recommend` / `config`
- Method: POST

**为什么要双 Host 识别**:豆包的 AI 推理走的是字节的 `snssdk.com` 域,而对话历史、账号管理走 `doubao.com`。工具两个都监听。

## 事件 2:eventSourceMessageReceived(SSE 流,doubao-capture.ps1:268-272)

```powershell
elseif ($event.method -eq 'Network.eventSourceMessageReceived' -and
        $s.requests.ContainsKey($event.params.requestId)) {

    if (!$parallelTracking) {
        # 简单模式:直接扫 image_ori_raw
        Parse-Body $event.params.data
    }
    elseif (!$isVideo -and !$boundConversationId -and
            $streamBindingBacklog.Count -lt 64 -and
            ($streamBindingBytes + ([string]$event.params.data).Length) -le 1048576) {
        # 并行模式:还没绑定 submissionId 之前,先缓冲最近 1MB 的 SSE 数据
        # 等 submissionIdentity 识别出来后,用这个缓冲去做 streamSubmissionBinding
        $streamBindingBacklog.Enqueue([string]$event.params.data)
        $streamBindingBytes += ([string]$event.params.data).Length
    }
}
```

### Parse-Originals / Parse-Body(简单模式的核心)

```powershell
function Parse-Originals($value, [int]$depth = 0) {
    if ($isVideo) { return }         # 视频走另一条路径
    if ($null -eq $value -or $depth -gt 50) { return }

    if ($value -is [string]) {
        # 字符串:尝试当 JSON parse 后递归
        if ($value.TrimStart().StartsWith('{') -or $value.TrimStart().StartsWith('[')) {
            try {
                $parsed = ConvertFrom-Json -InputObject $value
                Parse-Originals $parsed ($depth + 1)
            } catch {}
        }
        return
    }

    if ($value -is [System.Collections.IEnumerable] -and $value -isnot [pscustomobject]) {
        foreach ($child in $value) { Parse-Originals $child ($depth + 1) }
        return
    }

    foreach ($prop in $value.PSObject.Properties) {
        # ★ 核心:遇到 image_ori_raw 字段,取 url
        if ($prop.Name -eq 'image_ori_raw' -and $prop.Value.url) {
            $url = [Net.WebUtility]::HtmlDecode([string]$prop.Value.url)
            $uri = $null
            if ([uri]::TryCreate($url, [UriKind]::Absolute, [ref]$uri) -and
                $uri.Scheme -eq 'https' -and
                $uri.Host -match '(^|\.)(byteimg\.com|volces\.com)$') {
                $script:originals[$url] = $true
            }
        }
        Parse-Originals $prop.Value ($depth + 1)
    }
}

function Parse-Body([string]$body) {
    if (!$body -or $body.Length -gt 16777216) { return }
    try {
        $data = $body | ConvertFrom-Json
        Parse-Originals $data; return
    } catch {}
    # SSE 格式:"data: {...}\n" per line
    foreach ($line in ($body -split "`n")) {
        if ($line -match '^data:\s*(.+)') {
            try {
                $data = $Matches[1] | ConvertFrom-Json
                Parse-Originals $data
            } catch {}
        }
    }
}
```

**核心字段 `image_ori_raw.url`**:
- 这是豆包后端返回给前端的 JSON 字段,包含**无水印原图**的 CDN URL。
- 前端 UI 默认显示带水印版本,但响应里始终带原始版本的字段 —— 工具直接从网络层抢。
- 允许的 CDN 域:`byteimg.com`(字节图片 CDN)、`volces.com`(火山云)。

**不是接口返回的 bug,是服务端设计**:前端做任何修改(加水印、低分辨率预览)都需要原图作为基础,所以 JSON 里一定有。

## 事件 3:loadingFinished(doubao-capture.ps1:273-280)

```powershell
elseif ($event.method -eq 'Network.loadingFinished' -and
        $s.historyPending.ContainsKey([string]$event.params.requestId)) {

    $requestForReply = $s.historyPending[[string]$event.params.requestId]
    [void]$s.historyPending.Remove([string]$event.params.requestId)

    # 发命令读响应体
    $id = Command $s 'Network.getResponseBody' @{ requestId = $event.params.requestId }
    $s.pending[[string]$id] = @{ kind = 'history'; request = $requestForReply }
}
elseif ($event.method -eq 'Network.loadingFinished' -and
        $s.requests.ContainsKey($event.params.requestId)) {

    $id = Command $s 'Network.getResponseBody' @{ requestId = $event.params.requestId }
    $s.pending[[string]$id] = $true
}
```

CDP 不会自动推送响应 body,要主动调 `Network.getResponseBody`。这个请求本身也要走 request-id 对应回来。

### 响应 body 处理(doubao-capture.ps1:367-381)

```powershell
elseif ($event.id -and $s.pending.ContainsKey([string]$event.id)) {
    $pending = $s.pending[[string]$event.id]
    [void]$s.pending.Remove([string]$event.id)

    if (!$event.error -and $event.result.body) {
        $body = $event.result.body
        if ($event.result.base64Encoded) {
            $body = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($body))
        }

        if ($pending -is [hashtable] -and $pending.kind -eq 'history') {
            # 历史查询响应 → 调 result-reader.js 的 analyze() 做严格匹配
            $payload = $pending.request | ConvertTo-Json -Depth 12 -Compress
            $options = @{
                nativeShadow = $nativeShadowEnabled
                followupSubmissionIds = $followupSubmissionIds
                requireBinding = $parallelTracking
                requireIdentity = $parallelTracking
                expectedLocalMessageId = $submission.submissionIdentity.localMessageId
                expectedMessageId = $submission.submissionIdentity.messageId
                expectedConversationId = $submission.submissionIdentity.conversationId
                boundSubmissionId = $boundSubmissionId
                boundSubmissionCreatedAt = $boundSubmissionCreatedAt
                boundConversationId = $boundConversationId
                prompt = $feedbackRequest.prompt
                since = $sendSince
                type = $(if ($isVideo) { 'video' } else { 'image' })
            } | ConvertTo-Json -Compress
            $expression = $readerSource +
                ';shenlangDoubaoResultReader.analyze(' +
                ($body | ConvertTo-Json -Compress) + ',' + $payload + ',' + $options + ')'
            $id = Command $s 'Runtime.evaluate' @{
                expression = $expression; returnByValue = $true; awaitPromise = $true
            }
            $s.readerPending[[string]$id] = $pending.request
        }
        elseif (!$parallelTracking) {
            # 普通生成响应 → 直接扫 image_ori_raw
            Parse-Body $body
        }
    }
}
```

**parallel 模式 vs 非 parallel 模式的差异**:
- **非 parallel**:响应 body 直接 `Parse-Body` 挖原图
- **parallel**:响应 body 必须经过 `result-reader.js` 的 `analyze()` 做严格的 identity 匹配

## 历史重放(doubao-capture.ps1:432-449)

每秒主动重放一次历史查询:

```powershell
if ($sendSince -and !$script:originals.Count -and (Get-Date) -ge $nextHistoryPoll) {
    foreach ($s in $sessions) {
        if ($s.page -ne 'doubao-chat' -or $s.closed -or !$s.ready) { continue }
        if (!$s.historyRequest) { $report.lastReadReason = 'no_conversation_request'; continue }
        if ($s.readerPending.Count) { $report.lastReadReason = 'waiting_for_reader'; continue }

        # 原样重放先前观察到的 /im/chain/single 请求 — 保留 postData/headers
        # 这是一次"只读"请求 (POST 查询),不会触发新的生成
        $payload = $s.historyRequest | ConvertTo-Json -Depth 12 -Compress
        $options = @{ ... } | ConvertTo-Json -Compress
        $expression = $readerSource +
            ';shenlangDoubaoResultReader.read(' + $payload + ',' + $options + ')'

        try {
            $id = Command $s 'Runtime.evaluate' @{
                expression = $expression; returnByValue = $true; awaitPromise = $true
            }
            $s.readerPending[[string]$id] = $s.historyRequest
            $report.historyPolls++
        } catch {
            $s.closed = $true
            $report.lastReadReason = 'chat_transport_closed'
        }
    }
    $nextHistoryPoll = (Get-Date).AddSeconds(1)
}
```

**为什么要主动重放**:
- 豆包页面本身只在用户滚动或切换对话时才会主动查历史。
- 工具自己每秒重放一次,就能拿到最新的消息快照,即使用户看都没看。
- 重放的 fetch 发在豆包页面上下文内,用的是豆包的 cookie + 签名 —— 等效于用户自己又看了一眼。

### shenlangReadBoundConversation(result-reader.js 导出)

页面侧的重放实现:

```js
async function shenlangReadBoundConversation(request, options) {
    const info = validateRequest(request);     // 校验 URL 合法性
    const headers = {};
    for (const [name, value] of Object.entries(request.headers || {})) {
        if (/^(content-type|accept|authorization|agw-js-conv|x-[\w-]+)$/i.test(name)) {
            headers[name] = value;
        }
    }
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), 10000);

    try {
        const response = await fetch(info.url, {
            method: 'POST',
            body: request.postData,
            headers,
            credentials: 'include',
            cache: 'no-store',
            signal: controller.signal
        });
        if (!response.ok) return { ok: false, status: response.status, images: [] };
        const text = await response.text();
        if (text.length > 8388608) return { ok: false, reason: 'response_too_large', images: [] };
        const data = fixBigIntegers(text);         // ★ 下面详述
        return await analyze(data, request, options);
    } catch {
        return { ok: false, reason: 'read_failed', images: [] };
    } finally {
        clearTimeout(timer);
    }
}
```

**大整数修复 trick**(`_0x582c53`):
```js
JSON.parse(text.replace(
    /"(conversation_id|local_message_id|message_id)"\s*:\s*(-?\d{16,})(?=\s*[,}])/g,
    '"$1":"$2"'
));
```
豆包的 ID 都是 16+ 位整数,超过 JS Number 精度(2^53 ≈ 15 位)。直接 `JSON.parse` 会丢精度。工具用正则把这几个字段的值前后加引号,变成字符串保存,保留完整数字。**这是对 Rust 侧很重要的提醒**:Rust `serde_json` 默认也会丢精度(除非用 `serde_json::Value` 存 String),一定要在读之前预处理。

## analyze() 的完整流程

`doubao-result-reader.js` 的 `analyze(body, request, options)` 是整个过滤的核心:

```js
async function analyze(body, request, options) {
    const info = validateRequest(request);
    const conversationId = info.conversationId;

    // 1. 并行模式下必须 boundConversationId 一致
    if (options?.requireBinding && options.boundConversationId !== conversationId) {
        return { ok: true, images: [], feedback: null, binding: null };
    }

    // 2. 识别本次任务的 submissionId (对应前面的 submissionIdentity)
    const binding = collectSubmission(body, {
        ...options,
        conversationId
    });
    if (options?.requireBinding && !options.boundSubmissionId && !binding) {
        return { ok: true, images: [], feedback: null, binding: null,
                 reason: 'submission_not_matched' };
    }

    // 3. 识别"确认链":所有属于本次任务的消息 ID
    const followupSubmissionIds = followup(body, options, conversationId, binding);

    // 4. 识别反馈(是否继续生成、会员配额、拒绝等)
    const feedback = collectFeedback(body, {
        ...options,
        boundSubmissionId: binding?.submissionId || options?.boundSubmissionId,
        boundSubmissionCreatedAt: binding?.createdAt || options?.boundSubmissionCreatedAt,
        followupSubmissionIds,
        conversationId
    });

    // 5. 按类型 (图/视频) 挖结果
    if (options?.type === 'video') {
        return resolveVideos(body, request, {
            ...options,
            followupSubmissionIds,
            boundSubmissionId: binding?.submissionId || options?.boundSubmissionId,
            boundSubmissionCreatedAt: binding?.createdAt || options?.boundSubmissionCreatedAt
        }).then(result => ({
            ...result,
            feedback,
            binding,
            followupSubmissionIds,
            nativeShadow: nativeShadowSnapshot(body, options, binding, feedback, followupSubmissionIds)
        }));
    }

    // 图片
    return {
        ok: true,
        nativeShadow: nativeShadowSnapshot(body, options, binding, feedback, followupSubmissionIds),
        images: collect(body, {
            ...options,
            boundSubmissionId: binding?.submissionId || options?.boundSubmissionId,
            boundSubmissionCreatedAt: binding?.createdAt || options?.boundSubmissionCreatedAt,
            conversationId
        }),
        feedback,
        binding
    };
}
```

## feedback 系统 — 处理多轮确认

### 识别"是否继续生成" (collectFeedback)

```js
// 遍历消息,找属于 submission 链的 bot 回复
for (const msg of messages) {
    if (Number(msg.user_type) === 2 &&                     // bot 消息
        chain.has(String(msg.bot_reply_message_id)) &&     // 属于本次任务链
        /是否继续生成/.test(text) &&                       // 关键词 1
        /会员|付费|套餐/.test(text)) {                     // 关键词 2
        pendingConfirm = true;
    }
    // user 消息,如果是"继续生成"或"以上参数已由画布自动核对"
    if (Number(msg.user_type) === 1 && msg.id !== originalSubmissionId &&
        ((prompt && normalize(text).includes(prompt) &&
          /以上参数已由(?:用户|画布自动)核对/.test(text)) ||
         (pendingConfirm && normalize(text) === '继续生成'))) {
        chain.add(msg.id);                                 // ← 把这条认领进任务链
        pendingConfirm = false;
    }
}
```

**7 种 stage**:

| stage | 含义 |
|---|---|
| `unknown` | 还没识别到状态 |
| `checking` | 豆包在参数确认(等用户/工具按"确认") |
| `quota_checking` | 豆包在会员配额确认 |
| `generating` | 真正在生成 |
| `result` | 结果已返回 |
| `needs_check` | 不确定的回答,需要人介入 |
| `rendering_unverified` | 只回了"正在渲染",还没真生成 |
| `refused` / `failed` / `quota_exhausted` / `not_executed` | 终态拒绝 |

### 自动代回复(capture.ps1:411-430)

```powershell
if ($feedbackRequest -and (Get-Date) -ge $nextFeedbackPoll) {
    foreach ($s in $sessions) {
        if ($s.page -ne 'doubao-chat' -or $s.closed -or !$s.ready) { continue }

        # 后台追踪模式:当前会话不是绑定的会话,等用户切回来
        if ($backgroundTracking) {
            if ((Test-Path $confirmPath) -and $activeConversationId -eq $boundConversationId) {
                $backgroundTracking = $false; $handoffAnnounced = $false
            } else {
                if (Test-Path $confirmPath) {
                    [Console]::Out.WriteLine('DOUBAO_NEEDS_PAGE')
                }
                continue
            }
        }

        # 根据模式选择 run 分支
        $feedbackStep = if (!$isVideo) { 'terminalFeedback' }
                        elseif ($confirmed -or
                                $feedbackRequest.confirmationMode -eq 'auto' -or
                                $feedbackRequest.autoUseMembership -eq $true) { 'videoFollowup' }
                        else { 'videoFeedback' }

        $expression = $feedbackBootstrap +
            'typeof shenlangDoubaoGeneration === "undefined" ? ({stage:"unknown"}) : ' +
            'Promise.resolve(shenlangDoubaoGeneration.run(' + ($feedbackStep | ConvertTo-Json -Compress) +
            ',' + $payload + ')).then(value => ({...value, diagnostic: ...}))'

        $id = Command $s 'Runtime.evaluate' @{
            expression = $expression; returnByValue = $true; awaitPromise = $true
        }
        $feedbackPending[$s.targetId + ':' + [string]$id] = $true
    }
    $nextFeedbackPoll = (Get-Date).AddMilliseconds(500)
}
```

**三个入口的差异**(JS 侧):
- `terminalFeedback`(图像):只做识别,不发消息
- `videoFeedback`(视频,未自动确认):只识别,等用户介入
- `videoFollowup`(视频,自动确认开启):识别到"是否继续生成"时自动调 `fill`+`send` 发出"继续生成"文本

### 降级链(capture.ps1:300-304)

```powershell
# 自动确认失败 → 降级手动
if (($event.error -or $event.result.exceptionDetails) -and
    $feedbackRequest.confirmationMode -eq 'auto' -and
    $report['feedbackStage'] -eq 'checking') {

    $feedbackRequest | Add-Member -NotePropertyName confirmationMode -NotePropertyValue 'manual' -Force
    [Console]::Out.WriteLine('DOUBAO_AUTO_MANUAL')      # 父进程弹窗要用户点
}

# 会员额度自动失败 → 降级手动
if (($event.error -or $event.result.exceptionDetails) -and
    $feedbackRequest.autoUseMembership -eq $true -and
    $report['feedbackStage'] -eq 'quota_checking') {

    $feedbackRequest | Add-Member -NotePropertyName autoUseMembership -NotePropertyValue $false -Force
    [Console]::Out.WriteLine('DOUBAO_MEMBERSHIP_MANUAL')
}
```

### "正在渲染"假开始检测(capture.ps1:306-313)

```powershell
if ($feedback.stage -eq 'rendering_unverified' -and
    !$script:originals.Count -and !$resultSeenAt -and
    $report['feedbackStage'] -notin @('generating','result')) {

    if (!$unverifiedRenderingSince) { $unverifiedRenderingSince = Get-Date }
    if (((Get-Date) - $unverifiedRenderingSince).TotalSeconds -ge 30) {
        $feedback.stage = 'not_executed'
        $feedback.message = '豆包仅返回"正在渲染,请稍候",未确认有效的视频生成任务,请核对账号权限后重试。'
    }
} elseif ($feedback.stage -ne 'unknown') {
    $unverifiedRenderingSince = $null
}
```

**30 秒内不进入 generating/result** → 判定为"假开始",切 `not_executed`,防止卡 15 分钟。

### 手动介入的信号(confirm-video.json)

用户在沈浪前端点"确认继续"时,前端会在 OutputDirectory 下写一个 `confirm-video.json` 文件。下一轮轮询读到就当作用户确认:

```powershell
$confirmed = Test-Path -LiteralPath $confirmPath
if ($confirmed) {
    $feedbackRequest | Add-Member -NotePropertyName userConfirmed -NotePropertyValue $true -Force
    $payload = $feedbackRequest | ConvertTo-Json -Depth 12 -Compress
}
```

发完 followup 后,这个文件被删掉:
```powershell
if ($feedback.sent) {
    Remove-Item -LiteralPath $confirmPath -Force -ErrorAction SilentlyContinue
    $submission | Add-Member -NotePropertyName followupSent -NotePropertyValue $true -Force
    [Console]::Out.WriteLine('DOUBAO_STAGE:resent')
}
```

**Tauri 的替代实现**:用 Rust 内部的 channel 或 `tokio::sync::watch` 替代文件 IPC。

## 视频三条无水印获取路径

视频比图片麻烦得多,因为豆包不直接给无水印 URL,要间接拿。

### 路径 1:直链 lr=unwatermarked(简单模式,capture.ps1:265-267)

```powershell
elseif ($isVideo -and $sendSince -and
        $event.method -eq 'Network.responseReceived' -and
        $event.params.response.mimeType -like 'video/*' -and
        $s.videoRequests.ContainsKey([string]$event.params.requestId)) {

    $uri = $null
    $mediaURL = [string]$event.params.response.url
    if ([uri]::TryCreate($mediaURL, [UriKind]::Absolute, [ref]$uri) -and
        $uri.Scheme -eq 'https' -and
        $uri.Host -match '(^|\.)(doubaovideo\.com|bytevcloudcn\.com|bytevcloud\.com|volces\.com|volcvideo\.com|bytevideo\.com|ixigua\.com|v[0-9]+-vdl\.doubao\.com)$') {
        $videoNetworkCandidates[$mediaURL] = $true
    }
}
```

**豆包视频 CDN 域**:
- `doubaovideo.com`
- `bytevcloud(cn)?.com`
- `volces.com` / `volcvideo.com` / `bytevideo.com`
- `ixigua.com`
- `v<n>-vdl.doubao.com`

有的 URL 直接带 `?lr=unwatermarked`,这就是无水印版本。

### 路径 2:点"下载"按钮触发(capture.ps1:421-424)

```powershell
if ($report['feedbackStage'] -eq 'result' -and
    !$downloadButtonTriggered -and
    (Get-Date) -ge $nextDownloadAttempt -and
    !$downloadPending.Count) {

    $expr = 'shenlangDoubaoGeneration.run("requestVideoDownload",' + $payload + ')'
    $id = Command $s 'Runtime.evaluate' @{ expression = $expr; returnByValue = $true }
    $downloadPending[$s.targetId + ':' + [string]$id] = $true
    $nextDownloadAttempt = (Get-Date).AddSeconds(5)
}
```

JS 侧找到视频下方的下载按钮,hover 显示 → click。豆包会发请求拿无水印版本,工具旁听到这个请求。

### 路径 3:fallback API + 本地 AES-CBC 解密

这是**最硬核**的一条,见 [06-video-decryption.md](./06-video-decryption.md)。流程:

1. result-reader.js 的 `collectVideos` 从消息里找到 `fallback_url` 字段(加密 base64 token)
2. 对应的 `api` 字段是 `https://vas-<id>.snssdk.com/<path>` 的接口
3. capture.ps1 调 `Invoke-WebRequest` 请求这个 API,拿到带 token 的 JSON
4. 把 JSON 塞给 `shenlangDoubaoVideoDecoder.decode()` 本地 AES-128-CBC 解密出真实 URL

```powershell
if ($isVideo -and $value.fallbacks.Count) {
    foreach ($item in $value.fallbacks) {
        $vid = [string]$item.vid
        if ($fallbackAttempts.ContainsKey($vid) -and
            ((Get-Date) - $fallbackAttempts[$vid]).TotalSeconds -lt 10) { continue }
        $fallbackAttempts[$vid] = Get-Date

        try {
            $uri = [uri]$item.api
            if ($uri.Scheme -ne 'https' -or
                $uri.Host -notmatch '^vas-[a-z0-9-]+\.snssdk\.com$' -or
                $uri.UserInfo) { throw 'Invalid media API' }

            $response = Invoke-WebRequest -UseBasicParsing -Uri $uri.AbsoluteUri `
                -Headers @{ Accept = 'application/json,text/plain,*/*' } `
                -TimeoutSec 10 -MaximumRedirection 0
            $body = [string]$response.Content
            if ($body.Length -gt 8388608) { throw 'Media API response too large' }

            # 把 JSON 塞给页面解密
            $expression = $decoderSource +
                ';shenlangDoubaoVideoDecoder.decode(JSON.parse(' +
                ($body | ConvertTo-Json -Compress) + '),' +
                ([string]$item.api | ConvertTo-Json -Compress) + ')' +
                '.then(result => ({...result, url: shenlangDoubaoResultReader.unwatermarkedURL(result.url)}))'

            $id = Command $s 'Runtime.evaluate' @{
                expression = $expression; returnByValue = $true; awaitPromise = $true
            }
            $fallbackPending[$s.targetId + ':' + [string]$id] = $vid
        } catch {
            $report['fallbackReadReason'] = 'fallback_request_failed'
        }
    }
}
```

**注意**:`vas-<id>.snssdk.com` 是字节的媒体资源 API,API host 的具体 id 是动态的,所以用正则匹配。

## 下载落盘(capture.ps1:463-498)

拿到无水印 URL 后:

```powershell
if ($script:originals.Count) {
    [Console]::Out.WriteLine('DOUBAO_STAGE:downloading')

    if ($isVideo) {
        # 视频
        foreach ($candidate in $script:originals.Keys) {
            if ($candidate -notmatch '[?&]lr=unwatermarked(?:&|$)') {
                throw '未收到无水印视频地址。'
            }
        }
        $folder = Join-Path $OutputDirectory 'videos'
        [void](New-Item -ItemType Directory -Path $folder -Force)
        $file = Join-Path $folder 'video_1.mp4'
        $url = if ($script:videoPrimaryURL) { $script:videoPrimaryURL } else { @($script:originals.Keys)[0] }

        Invoke-WebRequest -UseBasicParsing -Uri $url `
            -Headers @{ Referer = 'https://www.doubao.com/' } `
            -OutFile $file -TimeoutSec 180 -MaximumRedirection 0

        # 校验: 不能太小/太大, 必须是 MP4
        if ((Get-Item $file).Length -le 12 -or (Get-Item $file).Length -gt 1073741824) {
            Remove-Item $file; throw '视频文件大小无效。'
        }
        $stream = [IO.File]::OpenRead($file)
        try {
            $header = New-Object byte[] 12
            [void]$stream.Read($header, 0, 12)
            # MP4 的 ftyp box 魔术字节在 offset 4-7
            if ([Text.Encoding]::ASCII.GetString($header, 4, 4) -ne 'ftyp') {
                throw '结果不是有效MP4视频。'
            }
        } finally { $stream.Dispose() }

        $report.status = 'video_downloaded'
    } else {
        # 图片 - 可能有多张(1-20)
        $folder = Join-Path $OutputDirectory 'images'
        [void](New-Item -ItemType Directory -Path $folder)
        $i = 0
        foreach ($url in ($script:originals.Keys | Select-Object -First 20)) {
            $i++
            $file = Join-Path $folder ('image_' + $i + '.png')
            Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $file `
                -TimeoutSec 40 -MaximumRedirection 0

            if ((Get-Item $file).Length -gt 104857600) {   # >100MB
                Remove-Item $file; throw '原图超过保存上限。'
            }

            # 用 System.Drawing.Image 识别格式并改扩展名
            $im = [Drawing.Image]::FromFile($file)
            try {
                $extension = if ($im.RawFormat.Guid -eq [Drawing.Imaging.ImageFormat]::Jpeg.Guid) { '.jpg' }
                           elseif ($im.RawFormat.Guid -eq [Drawing.Imaging.ImageFormat]::Png.Guid) { '.png' }
                           else { throw '原图格式暂不支持。' }
            } finally { $im.Dispose() }
            # 改名
            $correct = Join-Path $folder ('image_' + $i + $extension)
            if ($correct -ne $file) { Move-Item $file $correct; $file = $correct }
        }
        $report.status = 'original_downloaded_needs_visual_check'
    }
}
```

**下载时的 Referer 伪装**:视频 URL 可能校验 Referer,所以带上 `https://www.doubao.com/`。

**图片格式识别**:URL 可能不含扩展名,下载后用 .NET 的 `System.Drawing.Image.RawFormat.Guid` 识别真实格式。Rust 侧可以用 `image` crate 的 `image::guess_format` 做同样的事。

## 后台追踪模式(capture.ps1:457-460)

```powershell
if ($parallelTracking -and !$handoffAnnounced -and
    $boundConversationId -and $boundSubmissionId -and
    (!$isVideo -or $report['feedbackStage'] -eq 'generating')) {

    $backgroundTracking = $true
    $handoffAnnounced = $true
    $report['backgroundTracking'] = $true
    [Console]::Out.WriteLine('DOUBAO_BACKGROUND:' +
        (@{ conversationId = $boundConversationId; submissionId = $boundSubmissionId } |
         ConvertTo-Json -Compress))
}
```

**什么时候切后台**:并行模式 + 已经绑定了 conversationId + submissionId + 任务已进入 generating(视频) 或任意阶段(图片)。

**后台模式下做什么**:仍然每秒重放 historyRequest,但不再在豆包 chat 页上发任何操作(用户可以切别的对话、切别的账号)。工具安静地等结果。

**切回前台的条件**:
```powershell
if ($backgroundTracking) {
    if ((Test-Path -LiteralPath $confirmPath) -and $activeConversationId -eq $boundConversationId) {
        $backgroundTracking = $false; $handoffAnnounced = $false
    } else {
        if (Test-Path -LiteralPath $confirmPath) {
            [Console]::Out.WriteLine('DOUBAO_NEEDS_PAGE')   # 需要用户切回来
        }
        continue
    }
}
```

**对 Tauri 的参考**:生成任务不一定要独占 UI,用户可以切别的任务,中转程序在后台持续轮询。

## 捕获报告(doubao-result-capture-report.json)

写回 OutputDirectory:

```json
{
  "schemaVersion": 1,
  "status": "video_downloaded",
  "sessions": 2,
  "readySessions": 2,
  "generationRequests": 3,
  "responseBodies": 5,
  "originalImages": 1,
  "files": [
    { "file": "video_1.mp4", "bytes": 2457632 }
  ],
  "watermarkVerified": false,
  "historyRequests": 2,
  "chatConnections": 1,
  "pageReconnects": 0,
  "pageDiscoveryFailures": 0,
  "lastReadReason": "",
  "historyPolls": 23,
  "historyReadFailures": 0,
  "resultSource": "current_video_without_watermark",
  "message": "本次视频的无水印地址文件已下载。",
  "boundConversationId": "7501234567890",
  "boundSubmissionId": "7501234567891",
  "feedbackStage": "result"
}
```

## 给 Tauri 的实现要点

```rust
// 核心循环
async fn capture_loop(session: CdpSession, request: CaptureRequest) -> Result<CaptureReport> {
    let deadline = Instant::now() + Duration::from_secs(15 * 60);
    let mut originals = HashSet::<String>::new();
    let mut generation_refused = false;

    let mut next_feedback_poll = Instant::now();
    let mut next_history_poll = Instant::now();

    let (event_tx, mut event_rx) = mpsc::channel(1024);
    session.on_event(move |ev| { let _ = event_tx.try_send(ev); });

    loop {
        tokio::select! {
            _ = tokio::time::sleep_until(deadline.into()) => break,
            Some(ev) = event_rx.recv() => handle_event(ev, ...).await,
            _ = tokio::time::sleep(Duration::from_millis(500)), if Instant::now() >= next_feedback_poll => {
                feedback_poll(...).await;
                next_feedback_poll = Instant::now() + Duration::from_millis(500);
            }
            _ = tokio::time::sleep(Duration::from_secs(1)), if Instant::now() >= next_history_poll => {
                history_poll(...).await;
                next_history_poll = Instant::now() + Duration::from_secs(1);
            }
        }

        if !originals.is_empty() || generation_refused { break; }
    }

    download_originals(originals, ...).await
}
```

**关键库**:
- `serde_json` + 自己的大整数预处理
- `reqwest` 做下载
- `image` 识别格式
- `regex` 匹配 URL / 消息

**JS 复用**:三个 `doubao-*.js` 文件作为 `include_str!` 常量嵌入 Rust 二进制,`Runtime.evaluate` 时拼接发给豆包页面。

## 下一步

[06-video-decryption.md](./06-video-decryption.md) — 深入看 `doubao-video-decoder.js` 的本地 AES-CBC 解密算法。
