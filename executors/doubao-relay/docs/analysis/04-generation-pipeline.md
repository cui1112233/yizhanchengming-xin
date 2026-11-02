# 04 · 发任务管线

## 模块目标

**把一次画布生成请求(prompt、模型、比例、时长、参考图)操作到豆包聊天页面上,让豆包以为是用户自己手动点的**。

## 入口:doubao-generate.ps1 的参数

```powershell
param([string]$InputFile, [string]$ReportFile)
```

- `$InputFile`:JSON,包含本次请求的所有参数。
- `$ReportFile`:向父进程回报的 JSON 文件。父进程读这个文件知道提交进度。

## 请求 JSON 结构(由父进程写)

```json
{
  "type": "video",                          // 'image' 或 'video'
  "modelName": "Seedance 2.5",              // 模型名,白名单内
  "prompt": "一只蓝色小鸟...",              // 文本提示词,<=10000 字
  "ratio": "16:9",                          // 画面比例
  "duration": 5,                            // 视频时长(秒),image 时忽略
  "referenceFiles": [                       // 参考图/视频路径数组
    "C:\\Users\\...\\ref1.png",
    "C:\\Users\\...\\ref2.jpg"
  ],
  "submissionMode": "conversation",         // 'conversation' / 'work'
  "parallelTracking": true,                 // 启用多任务绑定
  "trackingKey": "shenlangDoubaoTask_abc123", // 全局变量名空间
  "confirmationMode": "auto",               // 'auto' / 'manual'
  "autoUseMembership": true                 // 遇到会员配额弹窗自动点"继续"
}
```

### 白名单校验(doubao-generate.ps1:110-119)

```powershell
if (!$request.prompt -or $request.prompt.Length -gt 10000) { throw '提示词无效。' }

if ($request.modelName -notin @(
    'Seedream 5.0 Flash','Seedream 5.0 Pro','Seedream 4.5',
    'Seedance 2.5','Seedance 2.0','Seedance 2.0 Fast','Seedance 2.0 Mini'
)) { throw '豆包图片模型无效。' }

if ($request.type -eq 'video') {
    $max = if ($request.modelName -eq 'Seedance 2.5') { 30 } else { 15 }
    if ($request.duration -lt 4 -or $request.duration -gt $max -or
        [int]$request.duration -ne $request.duration) { throw '视频时长无效。' }

    if ($request.submissionMode -eq 'conversation') {
        # 对话模式的特殊 prompt 前缀,告诉豆包要用哪个模型多长时间
        $request.prompt = '请使用 Seedance 2.0 Mini 生成' + $request.duration +
                          '秒视频,比例' + $request.ratio + '。' + "`n" + $request.prompt
    }
}

if ($request.ratio -notin @('自动','1:1','16:9','9:16','4:3','3:4','3:2','2:3','21:9')) {
    throw '豆包生成比例无效。'
}

$files = @($request.referenceFiles)
$maxReferences = if ($request.type -eq 'video' -and
                     $request.modelName -eq 'Seedance 2.0 Fast') { 11 } else { 10 }
if ($files.Count -gt $maxReferences) { throw '参考图片最多10张,Fast参考视频最多1个。' }
foreach ($file in $files) {
    if (!(Test-Path -LiteralPath $file -PathType Leaf)) { throw '参考图片文件不存在。' }
}
```

**支持的模型清单**(这是沈浪 v1.1.8 的最新列表):
- 图像:Seedream 5.0 Flash / 5.0 Pro / 4.5
- 视频:Seedance 2.5 / 2.0 / 2.0 Fast / 2.0 Mini

**支持的比例**:自动、1:1、16:9、9:16、4:3、3:4、3:2、2:3、21:9

**时长限制**:4-15 秒(除了 Seedance 2.5 支持 4-30 秒)。

**参考图/视频数量**:默认 10 张,Seedance 2.0 Fast 的视频参考可到 11。

## 操作步骤总览

```
doubao-generate.ps1
  │
  ├─ 1. 连 doubao-chat 页面 WebSocket
  ├─ 2. setWebLifecycleState=active + setFocusEmulationEnabled=true
  ├─ 3. Runtime.evaluate 注入 doubao-generation-page.js
  │
  ├─ 4. Step 'prepare'           ← 选模型、设比例
  │    失败时 fallback 到 Windows 原生菜单
  │
  ├─ 5. Step 'durationDragPlan'  ← 视频才有,拖滑动条设时长
  │    失败时 fallback 用 Arrow 键微调
  │
  ├─ 6. Page.setInterceptFileChooserDialog = true   ← 开始拦截文件对话框
  ├─ 7. 循环对每个参考图:
  │      Step 'fileInput' → DOM.setFileInputFiles
  │      失败时 fallback 到 Windows 原生对话框 (Select-NativeUploadFile)
  │      Step 'uploadReady'       ← 等上传完成
  │
  ├─ 8. Step 'fill'              ← 填 prompt
  ├─ 9. Step 'checkSend'         ← 确认发送按钮可点
  ├─ 10. Sync-CaptureBeforeSend  ← 同步 capture 子进程订阅
  ├─ 11. Prepare-SubmissionObservers  ← 并行模式:准备身份识别监听
  ├─ 12. Step 'send'             ← 真正点"发送"
  │
  └─ 13. 2 秒内尝试识别本次提交的 local_message_id (身份锚)
```

## Step 封装(doubao-generate.ps1:48-52)

```powershell
function Step([string]$name, [int]$timeout = 20000) {
    $report['operation'] = $name
    $json = $request | ConvertTo-Json -Depth 8 -Compress
    return Evaluate ('shenlangDoubaoGeneration.run(' +
                     ($name | ConvertTo-Json -Compress) + ',' +
                     $json + ')') $timeout
}
```

每个 Step 其实都是在豆包页面里跑 `shenlangDoubaoGeneration.run('<stepName>', request)`。注入的 JS 对象 `shenlangDoubaoGeneration` 内部是个大 switch,根据 stepName 分派不同处理。

### Evaluate 封装

```powershell
function Evaluate([string]$expression, [int]$timeout = 15000) {
    $value = Command 'Runtime.evaluate' @{
        expression = $expression
        returnByValue = $true
        awaitPromise = $true
    } $timeout

    if ($value.exceptionDetails) {
        $message = [string]$value.exceptionDetails.exception.description
        if ($message -match 'Error: ([^\r\n]+)') { $message = $Matches[1] }
        if ($message -match '[\u4e00-\u9fff]') { throw $message }   # 中文错误原样抛
        throw '豆包页面操作失败,请检查当前输入区。'
    }
    return $value.result.value
}
```

关键:**从 JS 异常里挖中文错误消息**。注入的 JS 遇到问题会 `throw new Error('中文描述')`,这里会直接把中文抛到用户眼前。

## Step 1: `prepare` — 选模型 + 设比例

prepare 的 JS 实现(来自 generation-page.js,已还原):

```js
async prepare(request) {
  // 1. 识别页面当前模式(对话 vs 工作)
  if (!inConversationMode()) throw new Error('未在对话输入区,任务未发送');

  // 2. 点"模型"下拉菜单
  const modelDropdown = findModelDropdown();
  clickLikeHuman(modelDropdown);              // PointerEvent 五事件链

  // 3. 等菜单项出现
  await waitFor(() => queryModelMenuItems().length > 0, 3000);

  // 4. 找到目标模型菜单项
  const target = queryModelMenuItems().find(el =>
    normalize(el.textContent) === normalize(request.modelName)
  );
  if (!target) throw new Error('未识别到所选图片模型菜单项');  // ← fallback 触发点
  clickLikeHuman(target);

  // 5. 视频类任务:设置比例
  if (request.type === 'video') {
    const ratioButton = findRatioButton();
    clickLikeHuman(ratioButton);
    const ratioOption = queryRatioOptions().find(el => el.textContent === request.ratio);
    if (!ratioOption) throw new Error('未找到所选生成比例');
    clickLikeHuman(ratioOption);

    // 对视频而言,下一步需要通过 durationDragPlan 设时长
    return { needsDurationInput: true };
  }

  return { needsDurationInput: false };
}
```

### 原生模型菜单 fallback(doubao-generate.ps1:132-139 + 53-74)

有些豆包版本的模型菜单是 Electron 原生菜单(渲染在系统层,DOM 看不到):

```powershell
try { $prepared = Step 'prepare' 60000 }
catch {
    $prepareError = $_.Exception.Message
    if ($prepareError -match '未识别到所选(?:图片|视频)模型菜单项') {
        $nativeSelected = $false
        try { $nativeSelected = Select-NativeModel } catch {}
        if (!$nativeSelected) { throw $prepareError }
        $report['modelMenuSource'] = 'native'
        $prepared = Step 'prepare' 60000      # ← 重试
    } else { throw }
}

function Select-NativeModel {
    Add-Type -AssemblyName UIAutomationClient
    Add-Type -AssemblyName UIAutomationTypes

    $root = [Windows.Automation.AutomationElement]::RootElement
    $condition = New-Object Windows.Automation.PropertyCondition(
        [Windows.Automation.AutomationElement]::ControlTypeProperty,
        [Windows.Automation.ControlType]::MenuItem
    )

    $matches = @()
    $wanted = ([string]$request.modelName -replace '\s','')

    foreach ($el in $root.FindAll([Windows.Automation.TreeScope]::Descendants, $condition)) {
        try {
            if ($el.Current.IsOffscreen -or !$el.Current.IsEnabled) { continue }
            $label = ([string]$el.Current.Name -replace '\s','')
            # 匹配 "Seedance2.5" (无空格), 允许后缀分隔符
            if ($label -notmatch ('^' + [regex]::Escape($wanted) + '(?:[^A-Za-z0-9.]|$)')) { continue }
            # 必须属于 Doubao.exe 进程
            if ((Get-Process -Id $el.Current.ProcessId -ErrorAction Stop).ProcessName -ine 'Doubao') { continue }
            $matches += ,$el
        } catch {}
    }

    if ($matches.Count -ne 1) { return $false }
    $pattern = $null
    # 优先 Invoke(普通按钮点击),否则 SelectionItem(单选菜单项)
    if ($matches[0].TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern, [ref]$pattern)) {
        $pattern.Invoke(); return $true
    }
    if ($matches[0].TryGetCurrentPattern([Windows.Automation.SelectionItemPattern]::Pattern, [ref]$pattern)) {
        $pattern.Select(); return $true
    }
    return $false
}
```

**学习点**:
- UIAutomation 的 `InvokePattern` / `SelectionItemPattern` 对原生控件的操作是从 Accessibility 层发起的,与"真实用户点击"效果等价,但不会抢物理鼠标。
- 匹配时去空格(`Seedance 2.5` → `Seedance2.5`),因为原生菜单文本可能带/不带空格。
- 要校验 ProcessId 属于 Doubao.exe,防止匹配到其他程序的菜单。

## Step 2: `durationDragPlan` — 滑动条拖拽时长

视频独有。豆包的时长选择是一个水平滑动条,拖到不同位置对应 4-30 秒。

### 混合策略:粗拖 + 键盘微调(doubao-generate.ps1:140-182)

```powershell
if ($prepared.needsDurationInput) {
    $matched = $false
    $dragAttempts = 0
    $keyboardDisabled = $false
    $keyboardBefore = $null
    $keyboardKey = $null

    for ($attempt = 0; $attempt -lt 40; $attempt++) {
        $plan = Step 'durationDragPlan'
        if ($plan.matched) { $matched = $true; break }

        # ① 第一次尝试:用 Home/End 键跳到端点
        if ($attempt -eq 0 -and $plan.endpointKey) {
            $code = if ($plan.endpointKey -eq 'End') { 35 } else { 36 }
            [void](Command 'Input.dispatchKeyEvent' @{
                type = 'keyDown'; key = $plan.endpointKey; code = $plan.endpointKey
                windowsVirtualKeyCode = $code; nativeVirtualKeyCode = $code
            })
            [void](Command 'Input.dispatchKeyEvent' @{ type = 'keyUp'; ... })
            Start-Sleep -Milliseconds 250
            continue
        }

        # ② 后续尝试:如果拖得差一两秒,用 Arrow 键微调
        if ($dragAttempts -gt 0 -and !$keyboardDisabled -and $plan.fineTuneKey) {
            $keyboardBefore = [int]$plan.duration
            $keyboardKey = [string]$plan.fineTuneKey
            $code = if ($keyboardKey -eq 'ArrowRight') { 39 } else { 37 }
            [void](Command 'Input.dispatchKeyEvent' @{
                type = 'keyDown'; key = $keyboardKey; code = $keyboardKey
                windowsVirtualKeyCode = $code; nativeVirtualKeyCode = $code
            })
            [void](Command 'Input.dispatchKeyEvent' @{ type = 'keyUp'; ... })
            Start-Sleep -Milliseconds 180
            continue
        }

        # ③ Arrow 键无效(键盘被豆包禁用),纯靠拖
        if ($dragAttempts -ge 8) { break }
        $dragAttempts++

        $startX = [double]$plan.startX
        $endX = [double]$plan.endX
        $y = [double]$plan.y

        # 鼠标按下 → 分 12 步平滑移动 → 释放
        [void](Command 'Input.dispatchMouseEvent' @{ type='mouseMoved'; x=$startX; y=$y })
        [void](Command 'Input.dispatchMouseEvent' @{
            type='mousePressed'; x=$startX; y=$y; button='left'; buttons=1; clickCount=1
        })
        try {
            for ($move = 1; $move -le 12; $move++) {
                $x = $startX + ($endX - $startX) * $move / 12
                [void](Command 'Input.dispatchMouseEvent' @{
                    type='mouseMoved'; x=$x; y=$y; button='left'; buttons=1
                })
                Start-Sleep -Milliseconds 20
            }
        } finally {
            [void](Command 'Input.dispatchMouseEvent' @{
                type='mouseReleased'; x=$endX; y=$y; button='left'; buttons=0; clickCount=1
            })
        }
        Start-Sleep -Milliseconds 300
    }

    $lastPlan = Step 'durationDragPlan'
    if (!$lastPlan.matched) {
        throw ('视频时长未生效:期望' + $request.duration + '秒,实际' + $lastPlan.duration + '秒;任务未发送。')
    }
    [void](Step 'verifyVideoSettings' 10000)
}
```

### durationDragPlan 返回什么

JS 侧的 `durationDragPlan` 根据当前实际时长和目标时长,计算出:
- `matched: true` → 已到目标,不用再调
- `duration: <当前秒数>` → 当前滑块对应的秒数
- `startX/endX/y: <坐标>` → 推荐的拖拽路径
- `endpointKey: 'End'/'Home'` → 推荐先跳端点(距离太远时)
- `fineTuneKey: 'ArrowLeft'/'ArrowRight'` → 推荐用方向键微调(差 1 秒时)

### 为什么这么复杂

- 滑动条的像素位置和秒数不是线性对应(有锚点、步进)。
- 鼠标拖拽终点可能差 1-2 像素,对应时长可能差 1 秒。
- Arrow 键在某些豆包版本上被禁用了,所以要**检测键盘是否生效**:记下按键前的时长(`keyboardBefore`),按键后预期 ±1,如果实际没变化就标记 `keyboardDisabled`,后面只用拖。

**容错逻辑在 attempt 之间会调整策略**。这是整个 PS 代码里最像 "控制系统"的一段。

## Step 3-4: 上传参考图

### 拦截文件对话框(doubao-generate.ps1:184-187)

```powershell
if ($files.Count) {
    [void](Command 'Page.enable' @{})
    [void](Command 'Page.setInterceptFileChooserDialog' @{ enabled = $true })
    $fileChooserIntercepted = $true
```

开启后,豆包页面调 `<input type=file>.click()` 不会真的弹对话框,而是触发 `Page.fileChooserOpened` 事件到 CDP。

### 循环上传(doubao-generate.ps1:188-205)

```powershell
for ($index = 0; $index -lt $files.Count; $index++) {
    [Console]::Out.WriteLine(('DOUBAO_UPLOAD:' + ($index+1) + '/' + $files.Count))
    $script:fileChooserNodeId = $null
    $nativeUploaded = $false

    try {
        [void](Step 'fileInput')       # ← 让 JS 点"添加参考图"按钮
    } catch {
        $uploadError = $_.Exception.Message
        # 如果 CDP 没抓到 fileChooserOpened,说明豆包用的是原生 Open 对话框
        if (!$script:fileChooserNodeId) {
            $nativeUploaded = Select-NativeUploadFile $files[$index]
            if (!$nativeUploaded) { throw $uploadError }
        }
    }

    if (!$nativeUploaded) {
        $dom = Command 'DOM.getDocument' @{ depth = 0 }
        # 优先用 shenlang 自己打了标记的 input
        $node = Command 'DOM.querySelector' @{
            nodeId = $dom.root.nodeId
            selector = 'input[data-shenlang-reference-upload="true"]'
        }
        if ($script:fileChooserNodeId) {
            [void](Command 'DOM.setFileInputFiles' @{
                backendNodeId = $script:fileChooserNodeId
                files = @($files[$index])
            })
        } elseif ($node.nodeId) {
            [void](Command 'DOM.setFileInputFiles' @{
                nodeId = $node.nodeId; files = @($files[$index])
            })
        } else {
            throw '参考图上传入口已变化,未发送。'
        }
    }

    $request | Add-Member -NotePropertyName uploadedCount -NotePropertyValue ($index+1) -Force
    [void](Step 'uploadReady' 70000)    # ← 70 秒等上传完成
}
```

**上传路径优先级**:
1. **CDP 的 fileChooserOpened**:豆包用的是 `<input type=file>` + `.click()` → CDP 拦截
2. **shenlang 自己打标记的 input**:生成-page.js 可能提前给页面注入了一个有 `data-shenlang-reference-upload="true"` 的 input 元素,直接 `DOM.setFileInputFiles` 塞文件
3. **Windows 原生对话框**:豆包调起 Win32 Open 对话框,CDP 看不到 → `Select-NativeUploadFile`

### 原生文件对话框 fallback(doubao-generate.ps1:75-108)

```powershell
function Select-NativeUploadFile([string]$file) {
    Add-Type -AssemblyName UIAutomationClient
    Add-Type -AssemblyName UIAutomationTypes

    $deadline = [DateTime]::UtcNow.AddSeconds(8)
    while ([DateTime]::UtcNow -lt $deadline) {
        $windows = @()
        $root = [Windows.Automation.AutomationElement]::RootElement
        $condition = New-Object Windows.Automation.PropertyCondition(
            [Windows.Automation.AutomationElement]::ControlTypeProperty,
            [Windows.Automation.ControlType]::Window
        )

        foreach ($window in $root.FindAll([Windows.Automation.TreeScope]::Children, $condition)) {
            try {
                # ClassName #32770 是 Win32 对话框的标准类名
                if ($window.Current.ClassName -ne '#32770' -or $window.Current.IsOffscreen) { continue }
                if ((Get-Process -Id $window.Current.ProcessId -ErrorAction Stop).ProcessName -ine 'Doubao') { continue }
                $windows += ,$window
            } catch {}
        }

        if ($windows.Count -gt 1) { throw '豆包有多个文件选择窗口,未选择文件' }
        if ($windows.Count -eq 1) {
            $dialog = $windows[0]
            $edits = @()      # "文件名" 输入框
            $open = @()       # "打开" 按钮

            foreach ($el in $dialog.FindAll([Windows.Automation.TreeScope]::Descendants,
                                            [Windows.Automation.Condition]::TrueCondition)) {
                try {
                    if ($el.Current.IsOffscreen -or !$el.Current.IsEnabled) { continue }
                    # 找 Edit 控件,匹配 "文件名"/"File name"/AutomationId=1148
                    if ($el.Current.ControlType -eq [Windows.Automation.ControlType]::Edit -and
                        ($el.Current.Name -match '^文件名|^File name' -or
                         $el.Current.AutomationId -eq '1148')) {
                        $edits += ,$el
                    }
                    # 找 Button,匹配 "打开"/"Open"/AutomationId=1
                    if ($el.Current.ControlType -eq [Windows.Automation.ControlType]::Button -and
                        ($el.Current.AutomationId -eq '1' -or $el.Current.Name -match '^打开|^Open')) {
                        $open += ,$el
                    }
                } catch {}
            }

            if ($edits.Count -eq 1 -and $open.Count -eq 1) {
                $value = $null; $invoke = $null
                if ($edits[0].TryGetCurrentPattern([Windows.Automation.ValuePattern]::Pattern, [ref]$value) -and
                    $open[0].TryGetCurrentPattern([Windows.Automation.InvokePattern]::Pattern, [ref]$invoke)) {
                    $value.SetValue($file)           # 把文件路径塞进"文件名"框
                    $invoke.Invoke()                 # 按"打开"
                    $report['uploadSource'] = 'doubao_native_file_dialog'
                    return $true
                }
            }
        }
        Start-Sleep -Milliseconds 200
    }
    return $false
}
```

**这段是整个工具里最硬核的"跨界融合"**:
- CDP(浏览器层)→ 到不了原生对话框
- JS(页面层)→ 也到不了
- UIAutomation(Windows 层)→ 可以操作任何 Win32 窗口

**AutomationId 魔数**:
- `1148` → Windows 文件对话框的"文件名"输入框标准 ID
- `1` → Windows 文件对话框的"打开"按钮标准 ID(来自旧版 Win32 CommDlg)

对 Tauri 的启示:Rust 侧可以用 `windows` crate 的 UIAutomation 绑定做同样的事,但很啰嗦。更务实:**用 `powershell` 调一个辅助 PS1 脚本处理原生对话框**。

## Step 5: `fill` — 把 prompt 填进输入框

### 定位输入框

```js
const inputs = [...document.querySelectorAll(
    'textarea,[contenteditable="true"],[role="textbox"]'
)].filter(/* 可见、未禁用 */);
if (inputs.length !== 1) throw new Error('未找到唯一豆包输入区,请检查当前任务页面。');
const input = inputs[0];
```

**三种选择器**兼容豆包不同版本:普通 `<textarea>`、现代富文本 `contenteditable`、ARIA `role="textbox"`。

### 填值(两分支)

```js
async fill(request) {
    const input = findInput();
    input.focus();

    if (input.tagName === 'TEXTAREA' || input.hasAttribute('contenteditable')) {
        // 分支 A: contenteditable - 先全选再 insertText(覆盖)
        const range = document.createRange();
        range.selectNodeContents(input);
        selection.removeAllRanges();
        selection.addRange(range);
        document.execCommand('insertText', false, request.prompt);
        input.dispatchEvent(new Event('input', { bubbles: true }));
    } else {
        // 分支 B: 其他 - 直接 insertText
        document.execCommand('insertText', false, request.prompt);
        input.dispatchEvent(new Event('input', { bubbles: true }));
    }

    return await /* 等 input.value 真的匹配 */;
}
```

### 为什么用 `execCommand('insertText')`

**不是**因为老 API,是因为:

1. **豆包用 React / 类 React 做输入框**。React 维护受控状态,直接 `input.value = text` 会被下次 render 覆盖。
2. **`execCommand('insertText')` 效果等同于用户用输入法输入**:浏览器自己派发完整的 `beforeinput` / `input` 事件链,React 的 `onChange` 能正常触发,状态被更新。
3. **compose 过程**也被模拟,中文输入法行为也兼容。

**替代方案**(写 Tauri 时考虑):
- 直接改 React fiber 的 pendingState 太脆弱,豆包版本一变就挂。
- 调 React 的 `nativeInputValueSetter`(社区常见做法):可能有效但依赖具体 React 版本。
- `execCommand` 虽然 deprecated,**但 Chromium 对它的支持短期内不会去**。

## Step 6: `checkSend` — 检查发送按钮

```js
checkSend(request) {
    const sendButton = findSendButton();
    if (!sendButton) throw new Error('未找到发送按钮');
    if (sendButton.hasAttribute('aria-disabled') &&
        sendButton.getAttribute('aria-disabled') === 'true') {
        throw new Error('发送按钮当前不可点,请检查输入');
    }
    if (sendButton.disabled) throw new Error('...');
    // 一些豆包版本用 CSS pointer-events:none 禁用
    const style = getComputedStyle(sendButton);
    if (style.pointerEvents === 'none') throw new Error('...');
    return { ok: true };
}
```

**三种禁用形式**都要检查:HTML `disabled` 属性、ARIA `aria-disabled`、CSS `pointer-events:none`。

## Step 7: `send` — 真正点发送

```js
async send(request) {
    const sendButton = findSendButton();
    clickLikeHuman(sendButton);         // 完整五事件链
    // 可能需要等豆包识别到"发送中"状态
    await waitFor(() => isSendingStarted(), 2000);
    return { ok: true };
}
```

### clickLikeHuman 五事件链(generation-page.js 位置 2500 附近)

```js
function clickLikeHuman(el) {
    const r = el.getBoundingClientRect();
    const payload = {
        clientX: r.x + r.width / 2,
        clientY: r.y + r.height / 2,
        button: 0, buttons: 1
    };
    if (typeof PointerEvent !== 'undefined') {
        el.dispatchEvent(new PointerEvent('pointerdown', {
            ...payload,
            pointerId: 1, pointerType: 'mouse', isPrimary: true
        }));
    }
    el.dispatchEvent(new MouseEvent('mousedown', payload));

    if (typeof PointerEvent !== 'undefined') {
        el.dispatchEvent(new PointerEvent('pointerup', {
            ...payload,
            buttons: 0, pointerId: 1, pointerType: 'mouse', isPrimary: true
        }));
    }
    el.dispatchEvent(new MouseEvent('mouseup', { ...payload, buttons: 0 }));
    el.dispatchEvent(new MouseEvent('click', { ...payload, buttons: 0 }));
}
```

**五事件链的必要性**:
- 现代 React 组件常常只监听 `onPointerDown`,`.click()` 不会触发它。
- Pointer + Mouse 两种事件都发,兼容新旧组件。
- 没有 `.click()` 调用时,有些组件靠 `click` 事件做可访问性回退,所以最后补一个。

## 身份锚:parallelTracking 模式(doubao-generate.ps1:210-228)

### 为什么需要身份锚

多个画布任务同时跑时,所有任务都在**同一个豆包账号的同一个聊天页**里。要靠什么区分"哪条生成结果是我的"?

答:**提交时的 `local_message_id`**。豆包前端在发请求前,会给这条消息分配一个唯一 ID,这个 ID 出现在:
- 发出的 POST body 里
- 服务端返回的流式响应里
- 最终的历史消息对象里

工具在"点发送"的瞬间开始监听网络,把带这个 `local_message_id` 的那条 POST body 整个保存下来,这就是本次任务的"身份锚"。

### 实现

```powershell
# 1. 开启监听
Prepare-SubmissionObservers        # 连 doubao-background 页也监听
[void](Command 'Network.enable' @{ maxPostDataSize = 16777216 })
$script:submissionRequests.Clear()
$script:submitArmed = $true        # ← 从这一刻起,所有 POST 都被记录

# 2. 执行发送
$report.submitAttempted = $true
$report['sendStartedAt'] = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
[void](Step 'send')
$report.submitted = $true

# 3. 在 2 秒内识别出 submissionIdentity
$identitySource = [IO.File]::ReadAllText('doubao-result-reader.js', ...)
$identityDeadline = [DateTime]::UtcNow.AddSeconds(2)
do {
    [void](Evaluate 'null')        # trick: 空 eval 强制同步事件队列
    foreach ($candidate in $script:submissionRequests) {
        $identity = Evaluate ($identitySource +
            ';shenlangDoubaoResultReader.submissionIdentity(' +
            ($candidate | ConvertTo-Json -Depth 15 -Compress) + ',' +
            ($request.prompt | ConvertTo-Json -Compress) +
            ', {verifiedSend: true})')
        if ($identity) {
            $report['submissionIdentity'] = $identity
            Save
            break
        }
    }
    if ($report['submissionIdentity']) { break }
    Start-Sleep -Milliseconds 50
} while ([DateTime]::UtcNow -lt $identityDeadline)

$script:submitArmed = $false
```

### submissionIdentity 返回什么(result-reader.js 的 `_0x46bb89`)

```js
submissionIdentity(request, prompt, options) {
    // 1. 校验 URL:必须是 www.doubao.com 的 completion/generate/send 路径
    // 2. 解析 request.postData,找里面带 prompt 的消息对象
    // 3. 提取 local_message_id / message_id / conversation_id
    return {
        localMessageId: '...',    // 必有
        messageId: null,          // 或 null(新发送还没分配服务端 ID)
        conversationId: '...'     // 或 null
    };
}
```

这个 identity 对象之后传给 `capture.ps1`,它据此做严格的结果过滤 —— 见第 05 章。

## 工作模式(submissionMode='work')

有一种特殊模式叫"工作模式",prompt 不是用户直接输入,而是用注入 JS 的 `buildTextPrompt` 根据画布节点结构自动组装的:

```powershell
if ($request.submissionMode -eq 'work') {
    $request.prompt = Evaluate ('shenlangDoubaoGeneration.buildTextPrompt(' +
                                ($request | ConvertTo-Json -Depth 12 -Compress) + ')')
}
```

这是画布业务层的功能,对中转程序无关(用户自己传 prompt 即可)。

## 错误诊断落盘

出任何错时,会在用户桌面写诊断 JSON:

```powershell
if ($request.type -eq 'video') {
    $diagnostic = [ordered]@{
        version = '1.1.7'
        model = $request.modelName
        operation = $failedOperation
        expectedRatio = $request.ratio
        expectedDuration = $request.duration
        message = $failureMessage
        controls = $report['videoSettingsDiagnostic']
        upload = $report['uploadDiagnostic']
        nativeFileDialogFound = $report['nativeFileDialogFound']
        nativeFilenameFields = $report['nativeFilenameFields']
        nativeOpenButtons = $report['nativeOpenButtons']
    }
    $diagnostic | ConvertTo-Json -Depth 15 |
        Set-Content -LiteralPath (Join-Path ([Environment]::GetFolderPath('Desktop')) '豆包视频操作诊断.json') -Encoding UTF8
}
```

**用户桌面出现 "豆包视频操作诊断.json"** → 说明任务失败了,这个文件会包含整个 CDP 操作链的诊断信息。中转程序应把这类诊断导到自己的日志目录,不要污染用户桌面。

## 给 Tauri 的 API 设计建议

```rust
#[tauri::command]
async fn submit_doubao_generation(request: GenerationRequest) -> Result<SubmissionReport>

struct GenerationRequest {
    task_type: TaskType,           // Image / Video
    model_name: String,            // 枚举值校验
    prompt: String,                // <=10000
    ratio: AspectRatio,            // 枚举
    duration: Option<u32>,         // Video 必填 4-15/30
    reference_files: Vec<PathBuf>, // <=10
    submission_mode: SubmissionMode,
    parallel_tracking: bool,
    tracking_key: Option<String>,
    confirmation_mode: ConfirmationMode,
    auto_use_membership: bool,
}

struct SubmissionReport {
    submitted: bool,
    submission_identity: Option<SubmissionIdentity>,
    send_started_at: i64,          // Unix ms
    stage: GenerationStage,
    message: String,
}

struct SubmissionIdentity {
    local_message_id: String,
    message_id: Option<String>,
    conversation_id: Option<String>,
}
```

**实现要点**:
- Rust 侧负责:参数校验、CDP 命令封装、重放 PS 的 Step 流程。
- 注入的 JS 原封不动用,从 Rust 的 `include_str!` 常量读入。
- 原生对话框 fallback 推荐:在 Tauri app bundle 里带一个辅助 PS1(或 C# 可执行),走原子外部进程调用。

## 下一步

阅读 [05-capture-pipeline.md](./05-capture-pipeline.md) 看 send 之后 capture 怎么拿到生成结果。
