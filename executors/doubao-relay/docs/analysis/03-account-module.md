# 03 · 账号模块

## 模块目标

**读取用户当前豆包桌面端上已登录的所有账号 + 当前正在用的账号 + 可以在它们之间切换**。整个过程**不要求用户输入密码**,全部靠已登录状态做自动化。

## 入口分发(doubao-accounts.ps1:4-12)

新代码里 `accounts.ps1` 只做分发,真正的实现在 `background.ps1`:

```powershell
if ($request.operation -in @('read','connect','switch')) {
    try {
        & (Join-Path $PSScriptRoot 'doubao-accounts-background.ps1') `
            -InputFile $InputFile -OutputFile $OutputFile
    } catch {
        @{
            ok = $false
            accounts = @()
            currentLabel = ''
            syncMode = 'background'
            message = '后台同步组件执行失败,请查看同步诊断后重试。'
            diagnostics = @{ status = 'component_error'; errorType = $_.Exception.GetType().FullName }
        } | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $OutputFile -Encoding UTF8
    }
    exit 0
}
# 否则走下面的老路径(基本死代码,新前端不会走到)
```

剩下 `accounts.ps1` 老代码(100 多行)用 `AccountMouse.SetCursorPos` + `mouse_event` 调 WinAPI 真实移动鼠标点击 —— 这是会抢物理鼠标的,但 operation=read/connect/switch 覆盖了所有场景,所以**老代码实际跑不到**。

### 支持的 operation

| operation | 作用 |
|---|---|
| `connect` | 只确保豆包启动 + CDP 就绪,不读账号 |
| `read` | 读账号列表 + 当前账号 |
| `switch` | 读完后点击目标账号完成切换 |

### 请求 JSON 结构

```json
{
  "operation": "switch",
  "name": "张三",
  "label": "avatar 张三 标准套餐",
  "returnWindowHandle": "123456"
}
```

- `name` 用于匹配 → clean 规则:去掉 "avatar " 前缀 + 去掉 "标准套餐/体验套餐/..." 后缀。
- `label` 是完整带前缀后缀的字符串。
- `returnWindowHandle` 是沈浪自己的窗口 handle,操作完后要把焦点还给它(调 `SetForegroundWindow`)。

## 读账号列表 — 双通道设计

**两条路同时跑,互为备份**。为什么?因为豆包的 UI 可能有两种实现:

| 通道 | 原理 | 什么时候成功 |
|---|---|---|
| **DOM 通道** | 在豆包页面 Runtime.evaluate 一段 JS,扫 `aria-label^="avatar"` 的元素 | 账号菜单是普通 HTML/CSS 的时候 |
| **UIA 通道** | 用 .NET UIAutomation API 遍历 Doubao.exe 主窗口的 MenuItem 控件 | 账号菜单是 Electron 的原生菜单(窗口菜单 / 系统菜单)的时候 |

**沈浪两条都跑**,最后合并去重。

### DOM 通道 — 定位账号按钮(accounts-background.ps1:240-245)

第一步是找到左下角那个"账号按钮"(就是你当前头像的那个)。它注入这段 JS:

```js
(() => {
  const nodes = [...document.querySelectorAll(
    '[aria-label^="avatar "],[role="menuitem"],button'
  )];
  const accounts = [];
  for (const el of nodes) {
    const aria = el.getAttribute('aria-label') || '';
    const text = (el.textContent || '').trim();
    const label = /^avatar\s+/.test(aria) ? aria :
                  /^avatar\s+/.test(text) ? text : '';
    if (label && label.length <= 300 && !accounts.some(a => a.label === label)) {
      accounts.push({label});
    }
  }
  const ids = __ACCOUNT_IDS__;    // ← 从 UIA 通道传来的候选 id,PS 代码会 Replace 掉
  const accountButtons = ids
    .map(id => document.getElementById(id))
    .filter(Boolean)
    .map(el => ({
      tag: el.tagName,
      role: el.getAttribute('role'),
      hasPopup: el.getAttribute('aria-haspopup'),
      expanded: el.getAttribute('aria-expanded')
    }));
  return {
    accountButtons,
    frameCount: document.querySelectorAll('iframe').length,
    accounts: accounts.slice(0, 100),
    readyState: document.readyState,
    visibility: document.visibilityState
  };
})()
```

关键点:
- 用 `aria-label` 开头是 "avatar " 来识别账号项 → 这是豆包页面的无障碍规约。
- 账号按钮本身是 `button[aria-haspopup="menu"]`,位置在左下角。
- 带回 `document.readyState` 和 `visibilityState` 用于诊断。

### DOM 通道 — 展开侧边栏(accounts-background.ps1:256-262)

如果找不到候选按钮,可能是侧边栏折叠了。它跑这段 JS 查找并点开:

```js
(() => {
  const matches = [...document.querySelectorAll('button,[role="button"]')].filter(el => {
    const r = el.getBoundingClientRect();
    // 左上角 15-60 像素见方,带 SVG 的小按钮 → 大概率是侧栏切换
    return r.width >= 15 && r.width <= 60 &&
           r.height >= 15 && r.height <= 60 &&
           r.left >= 0 && r.left < 50 &&
           r.top >= 0 && r.top < 80 &&
           !!el.querySelector('svg');
  });
  if (matches.length !== 1) return { action: 'none', count: matches.length, ... };
  const el = matches[0];
  el.setAttribute('data-canvas-sidebar-probe', 'v10');   // ← 自己标记,避免重复触发
  el.click();
  return { action: 'temporary_expand', count: 1 };
})()
```

展开后给它 250 ms × 6 轮重试,每轮再调一次 locateExpression。

### DOM 通道 — 展开账号菜单(accounts-background.ps1:271-275)

找到按钮后,**不能用 `.click()`**,豆包识别不到。要派发 PointerEvent:

```js
(() => {
  const ids = __ACCOUNT_IDS__;
  const matches = ids.map(id => document.getElementById(id)).filter(el =>
    el && el.tagName === 'BUTTON' && el.getAttribute('aria-haspopup') === 'menu'
  );
  if (matches.length !== 1) return { action: 'ambiguous', count: matches.length };
  const el = matches[0];
  if (el.getAttribute('aria-expanded') === 'true') return { action: 'already_open' };

  el.dispatchEvent(new PointerEvent('pointerdown', {
    bubbles: true, cancelable: true,
    pointerId: 1, pointerType: 'mouse',
    button: 0, buttons: 1, isPrimary: true
  }));
  el.dispatchEvent(new PointerEvent('pointerup', {
    bubbles: true, cancelable: true,
    pointerId: 1, pointerType: 'mouse',
    button: 0, buttons: 0, isPrimary: true
  }));
  return { action: 'page_pointer_events', expanded: el.getAttribute('aria-expanded') };
})()
```

### DOM 通道 — 展开"更多账号"子菜单(accounts-background.ps1:279-282)

豆包菜单第一级只显示 2 个常用账号,其余藏在"更多账号"子菜单里。需要 pointermove 悬停触发(hover):

```js
(() => {
  const entries = [...document.querySelectorAll('[role="menuitem"]')];
  const more = entries.filter(el => (el.textContent || '').trim() === '更多账号');
  if (more.length !== 1) return { action: 'none', count: more.length };
  const el = more[0];
  el.dispatchEvent(new PointerEvent('pointermove', {
    bubbles: true, pointerType: 'mouse', pointerId: 1
  }));
  return { action: 'submenu_hover', expanded: el.getAttribute('aria-expanded') };
})()
```

### DOM 通道 — 读账号列表(accounts-background.ps1:286-288)

```js
(() => {
  return {
    entries: [...document.querySelectorAll('[role="menuitem"]')]
      .filter(el => {
        const r = el.getBoundingClientRect();
        return r.width > 0 && r.height > 0;
      })
      .map(el => ({
        label: (() => {
          const aria = el.getAttribute('aria-label') || '';
          if (/^avatar\s+/.test(aria)) return aria.trim().slice(0, 300);
          // 兜底:从 innerText 拼
          const full = (el.innerText || el.textContent || '').trim();
          const badges = [...el.querySelectorAll('span,div,small')].filter(n =>
            !n.children.length &&
            /^[一-鿿A-Za-z0-9]{1,24}套餐$/.test((n.textContent || '').trim())
          );
          const badge = badges.length === 1 ? badges[0].textContent.trim() : '';
          return badge && full.endsWith(badge)
            ? (full.slice(0, -badge.length).trim() + ' ' + badge).slice(0, 300)
            : full.slice(0, 300);
        })(),
        hasPopup: el.getAttribute('aria-haspopup'),
        avatar: !!el.querySelector('img[alt="avatar"]'),
        imageAlts: [...el.querySelectorAll('img')].map(im => im.getAttribute('alt') || '').slice(0, 3)
      }))
      .slice(0, 100)
  };
})()
```

### DOM 通道 — 稳定性等待(accounts-background.ps1:289-302)

账号菜单会异步渲染。先循环读直到两次读到的数量相同,才认为渲染稳定:

```powershell
$menuValue = $null; $previousCount = -1; $stableRounds = 0
$menuLimit = if ($request.operation -eq 'switch') { 40 } else { 8 }

for ($menuAttempt = 0; $menuAttempt -lt $menuLimit; $menuAttempt++) {
    [void](Evaluate $socket $moreExpression)       # 再次 hover 更多账号(防菜单关掉)
    Start-Sleep -Milliseconds $(if ($request.operation -eq 'switch') { 100 } else { 400 })
    $menuValue = Evaluate $socket $menuExpression
    $entryCount = @($menuValue.entries).Count
    if ($entryCount -eq $previousCount) { $stableRounds++ } else { $stableRounds = 0 }
    $previousCount = $entryCount
    if ($menuAttempt -ge 3 -and $stableRounds -ge 2) { break }
}
```

- `operation=read` 时慢节奏(400ms/轮,8 轮),给页面充分渲染时间。
- `operation=switch` 时快节奏(100ms/轮,40 轮),抢时间立即找到目标点击。

### DOM 通道 — 过滤垃圾条目

菜单里不只是账号,还有"豆包订阅/免费领额度/设置..."等 UI 项。白名单过滤:

```powershell
$excluded = @(
    '设置','外观设置','收藏夹','豆包官网','检查更新','帮助与反馈',
    '更多账号','添加账号','退出登录',
    '豆包订阅','豆包订阅升级','免费领额度','免费领取额度'
)

foreach ($entry in @($menuValue.entries)) {
    $label = [string]$entry.label
    if (!$entry.avatar -or !$label -or $entry.hasPopup -or
        $label -in $excluded -or $label -match '^当前时段额度用量') { continue }

    $label = (($label -replace '^avatar\s+','') -replace '\s+',' ').Trim()
    if ($label -in $excluded) { continue }

    $subscription = ''
    if ($label -match '(体验套餐|标准套餐|加强套餐|高级套餐|专业套餐|基础套餐|...套餐)$') {
        $subscription = $Matches[1]
        $label = $label.Substring(0, $label.Length - $subscription.Length).Trim()
    }

    if ($label.Length -gt 0 -and $label.Length -le 120) {
        $parsed += @{ name = $label; subscription = $subscription; label = $entry.label; avatar = $entry.avatar }
    }
}
```

**判别规则**:
- 必须有头像(`avatar: !!el.querySelector('img[alt="avatar"]')`)
- 不能有子菜单(`!hasPopup`,即不是"更多账号"这种)
- 名字 1-120 字符
- 不在白名单屏蔽列表

### DOM 通道 — 识别"当前账号"(accounts-background.ps1:319-337)

额外跑一次识别当前账号的 JS:

```js
(() => {
  const visible = el => { const r = el.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').trim();
  const clean = s => s.replace(/^avatar\s*/, '')
                      .replace(/(体验套餐|...|...套餐)\s*$/, '')
                      .trim();

  // 左下角的账号按钮
  const buttons = [...document.querySelectorAll('button,[role="button"]')].filter(el => {
    const r = el.getBoundingClientRect();
    return visible(el) &&
           r.left >= 0 && r.left < 80 &&
           r.top > innerHeight - 110 && r.bottom <= innerHeight + 2 &&
           r.width >= 100 && r.width <= 360 &&
           r.height >= 20 && r.height <= 85;
  });

  // 菜单项(此时菜单可能展开)
  const menus = [...document.querySelectorAll('[role="menuitem"]')].filter(visible);

  const login = buttons.some(el => /^(登录|立即登录|登录豆包)$/.test(label(el)));
  const logout = menus.some(el => /^退出登录$/.test(label(el)));
  const identities = buttons.filter(el =>
    el.getAttribute('aria-haspopup') === 'menu' && !/^(登录|...)$/.test(label(el))
  );

  // 关键判断:有"退出登录"菜单项 + 唯一的账号按钮 → 已登录
  if (logout && identities.length === 1) {
    const el = identities[0];
    const full = (el.innerText || el.textContent || '').trim();
    const subscription = full.match(/(体验套餐|...|...套餐)\s*$/)?.[1] || '';
    const name = clean(full);
    // ... (处理公司账号等多行情况)
    return { status: 'logged_in', name, subscription, evidence: 'account_button_and_logout' };
  }
  return {
    status: login && !logout ? 'not_logged_in' : 'unknown',
    evidence: login ? 'login_button' : 'none'
  };
})()
```

**判定逻辑**:
- 看到"登录"按钮 → 未登录
- 看到"退出登录"菜单项 + 唯一的账号按钮 → 已登录 + 这个按钮的文字就是当前账号名
- 否则 → 未知

## UIA 通道 — 原生菜单读取(accounts-background.ps1:193-217)

备用:当 DOM 通道失效(菜单是 Electron 原生实现)时:

```powershell
Add-Type -AssemblyName UIAutomationClient
Add-Type -AssemblyName UIAutomationTypes

$windows = @(Get-Process -Name Doubao -ErrorAction SilentlyContinue |
             Where-Object { $_.MainWindowHandle -ne 0 })
foreach ($p in $windows) {
    $w = [Windows.Automation.AutomationElement]::FromHandle($p.MainWindowHandle)
    $buttons = $w.FindAll(
        [Windows.Automation.TreeScope]::Descendants,
        (New-Object Windows.Automation.PropertyCondition(
            [Windows.Automation.AutomationElement]::ControlTypeProperty,
            [Windows.Automation.ControlType]::Button))
    )

    # 找左下角位置的按钮(账号区)
    $rect = $w.Current.BoundingRectangle
    $region = @($buttons | Where-Object {
        $c = $_.Current; $b = $c.BoundingRectangle
        !$c.IsPassword -and !$b.IsEmpty -and
        $b.Height -ge 20 -and $b.Height -le 85 -and
        $b.Width -ge 100 -and $b.Width -le 360 -and
        $b.Left -ge ($rect.Left - 2) -and $b.Left -lt ($rect.Left + 80) -and
        $b.Top -gt ($rect.Bottom - 100) -and $b.Bottom -le ($rect.Bottom + 2)
    })

    # 枚举 MenuItem 控件,匹配 avatar 开头的 Name
    $items = @()
    foreach ($e in $w.FindAll([Windows.Automation.TreeScope]::Descendants,
                              (New-Object Windows.Automation.PropertyCondition(
                                  [Windows.Automation.AutomationElement]::ControlTypeProperty,
                                  [Windows.Automation.ControlType]::MenuItem)))) {
        if (!$e.Current.IsPassword -and $e.Current.Name -match '^avatar\s+') {
            $items += @{ label = $e.Current.Name }
        }
    }
    $count = Add-Accounts $items 'uia'
}
```

**UIA 的妙处**:它读的是 Windows 的无障碍层(accessibility tree),这个层是 **Chromium + Electron 自动生成的**,即使元素被 `display:none` 隐藏、或者是原生菜单项,UIA 都能看到。因此与 CSS 布局无关。

**UIA 的限制**:它比 DOM 慢(枚举要几百 ms)、无法派发 PointerEvent(只能用 InvokePattern/ExpandCollapsePattern,豆包的现代按钮可能不响应)。所以作读取主力好、作操作主力差。

## 账号切换 — DOM 操作(accounts-background.ps1:347-429)

### 构造切换表达式

```js
(() => {
  const wanted = __TARGET_NAME__;    // 要切到的账号名
  const clean = s => s.replace(/^avatar\s*/, '')
                      .replace(/(...|...套餐)\s*$/, '')
                      .trim();

  const items = [...document.querySelectorAll('[role="menuitem"]')].filter(el => {
    const r = el.getBoundingClientRect();
    return r.width > 0 && r.height > 0 &&
           !el.getAttribute('aria-haspopup') &&
           el.querySelector('img[alt="avatar"]') &&
           clean((el.getAttribute('aria-label') || el.innerText || el.textContent || '').trim()) === wanted;
  });

  if (items.length !== 1) return { action: 'ambiguous', count: items.length };
  const el = items[0];
  if (el.getAttribute('aria-disabled') === 'true') return { action: 'disabled' };

  const name = clean((el.getAttribute('aria-label') || el.textContent || '').trim());
  el.click();     // ← 这里用 .click() 是可以的,因为是菜单项点击
  return { action: 'clicked_element', name };
})()
```

**为什么这里能用 `.click()`?** 账号按钮的切换触发器是菜单项,豆包用的是普通的 click handler;而主输入按钮(发送、模型切换)用的是 PointerEvent 识别。所以要分场景。

### 切换后验证(accounts-background.ps1:361-427)

切完要**反复验证**切换是否生效。核心表达式:

```js
(() => {
  const clean = s => s.replace(/^avatar\s*/, '')
                      .replace(/(...|...套餐)\s*$/, '')
                      .replace(/\s*[›〉»>]\s*$/, '')
                      .replace(/\s+/g, ' ')
                      .trim();
  const known = __KNOWN_NAMES__;      // 已知所有账号名
  const wanted = __VERIFY_TARGET__;   // 目标账号名

  const all = [...document.querySelectorAll('button,[role="button"]')]
    .filter(el => el.getAttribute('aria-haspopup') === 'menu' && !el.closest('[role="menu"]'));

  const namesFor = el => {
    const full = (el.innerText || el.textContent || '').trim();
    const lines = full.split(/\r?\n/).map(clean).filter(Boolean);
    const exact = known.filter(name =>
      clean(lines.join(' ')) === name ||
      clean(full) === name ||
      clean(el.getAttribute('aria-label') || '') === name
    );
    return exact.length ? exact : known.filter(name => lines.includes(name));
  };

  // 两条判定路径:
  // 1. 布局路径:左下角那个唯一的按钮
  const positioned = all.filter(el => {
    const r = el.getBoundingClientRect();
    return r.width >= 35 && r.width <= 360 && r.height >= 20 && r.height <= 85 &&
           r.left >= 0 && r.left < 80 &&
           r.top > innerHeight - 110 && r.bottom <= innerHeight + 2;
  });
  // 2. 语义路径:对 hidden/minimized 窗口 fallback —— 带 avatar img 且名字唯一匹配
  const semantic = all.filter(el =>
    el.querySelector('img[alt="avatar"]') && namesFor(el).length === 1
  );
  const buttons = positioned.length === 1 ? positioned : semantic;

  if (buttons.length !== 1) {
    return { name: '', targetName: wanted, targetMatched: false,
             candidateCount: buttons.length, reason: 'identity_not_unique', ... };
  }
  const el = buttons[0];
  const names = namesFor(el);
  return {
    name: names.length === 1 ? names[0] : '',
    targetName: wanted,
    targetMatched: names.length === 1 && names[0] === wanted,
    source: positioned.length === 1 ? 'layout' : 'avatar_identity',
    ...
  };
})()
```

**验证的两条路径**:
- 布局路径:新账号加载完,按钮会出现在左下角原位置。
- 语义路径:窗口被最小化时左下角按钮没布局,就靠 avatar 图片 + 名字匹配 known 列表。

### 切换时的 reload 处理(accounts-background.ps1:395-410)

**关键细节**:账号切换会让豆包 reload 整个页面,老的 CDP target 会失效。沈浪的处理方式:

```powershell
$verified = $false
$verifyDeadline = [DateTime]::UtcNow.AddSeconds(20)
$reconnectNeeded = $false

do {
    Start-Sleep -Milliseconds 100
    try {
        if ($reconnectNeeded) {
            # 老 socket 作废,重新发现目标
            if ($socket) { $socket.Dispose() }
            $freshTargets = Invoke-RestMethod 'http://127.0.0.1:9707/json/list' -TimeoutSec 1
            $freshPages = @($freshTargets | Where-Object { ... Host -eq 'doubao-chat' })
            if ($freshPages.Count -ne 1) { continue }
            # 重连新 target
            $socket = New-Object AccountReadSocket
            $socket.Connect($freshPages[0].webSocketDebuggerUrl)
            $pageFocusEnabled = Set-PageFocus $socket $true
            Resume-Page $socket
            $reconnectNeeded = $false
        }
        $actual = Evaluate $socket $verifyExpression
        if ($actual.targetMatched -eq $true -and $actual.targetName -ceq $targetName) {
            $verified = $true; break
        }
    } catch {
        $reconnectNeeded = $true    # ← 任何异常 → 下轮强制重连
    }
} while ([DateTime]::UtcNow -lt $verifyDeadline)
```

**20 秒窗口 + 异常触发重连**。这段逻辑对 Tauri 中转尤其重要:**切账号后你连 target 的 WebSocket 一定会断**,要有自动重连策略。

## Report 结构(写回 OutputFile)

成功:
```json
{
  "ok": true,
  "accounts": [
    { "name": "张三", "label": "avatar 张三 标准套餐" },
    { "name": "李四 字节科技", "label": "avatar 李四 字节科技" },
    ...
  ],
  "currentName": "张三",
  "currentLabel": "avatar 张三 标准套餐",
  "message": "账号已同步。",
  "diagnostics": { ... }
}
```

失败:
```json
{
  "ok": false,
  "accounts": [],
  "currentLabel": "",
  "message": "豆包尚未登录,请登录后再次同步。",
  "diagnostics": {
    "startup": { "action": "already_connected", "error": "" },
    "status": "not_logged_in",
    "connection": "connected",
    "accountCount": 0,
    "stages": [ ... ]
  }
}
```

## 给 Tauri 的 API 设计建议

```rust
// Tauri command
#[tauri::command]
async fn list_doubao_accounts() -> Result<AccountListResult, String>

#[tauri::command]
async fn switch_doubao_account(name: String) -> Result<AccountListResult, String>

struct AccountListResult {
    ok: bool,
    accounts: Vec<DoubaoAccount>,
    current_name: Option<String>,
    message: String,
    diagnostics: Value,
}

struct DoubaoAccount {
    name: String,
    label: String,           // 带套餐后缀的完整标签
    subscription: Option<String>,  // "标准套餐" / "高级套餐" / None
}
```

**不建议**在 Rust 里重新造 UIA 调用(Windows UIAutomation Rust 绑定不成熟)。更务实的做法:**只做 DOM 通道,UIA 通道通过嵌入一个小 PowerShell/C# fallback 来实现**,或者先接受"只支持 DOM 可读的版本"。

## 风险提示

- 账号切换会 reload 页面,正在跑的生成任务会被打断。中转程序应**排队调度**:同一时刻只在一个账号上跑任务。
- 多账号自动轮询会被豆包风控系统识别为异常,**建议手动指定账号**而不是自动轮转。
- 当前账号与"目标账号已经登录"是两回事。如果用户只登录了 1 个号,这个模块只能读到 1 个,无法"切换"到没登录的号。
