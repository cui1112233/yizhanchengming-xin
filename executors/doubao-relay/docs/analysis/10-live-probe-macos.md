# 10 · 真豆包探测结论(macOS)

> 各 Phase 在真豆包上实测得到的事实,按 Task 追加。只记录**观察到的结果**与对后续 Task 的影响,不改 spec。
>
> 环境:macOS(Darwin 25.2.0)· 豆包 2.32.5 · `--remote-debugging-port=9707` · Node 25.6.1 · 探针工具 `scripts/cdp-probe.mjs`。

## Task 1.2 · 账号注入脚本 `accounts.js`(2026-10-16)

### CDP target 列表(`GET /json/list`)

| type | url | 处理 |
|---|---|---|
| `page` | `doubao://doubao-chat/chat` | chat session(账号/生成) |
| `page` | `doubao://doubao-background/` | background session |
| `iframe` | `https://www.doubao.com/drive-iframe/drive/home/...` | 过滤 |
| `other` | `doubao://doubao-chat/cross-site-support/` | 过滤 |

页面内 `location.href` 实际是 `chrome://doubao-chat/chat`(scheme 为 `chrome`),而 `/json/list` 里报的是 `doubao://`——白名单必须同时接受两种 scheme(spec 已如此规定)。

### 逐方法验证结果(侧栏展开、已登录、2 个账号时)

| 方法 | 返回 | 结论 |
|---|---|---|
| `probe()` | `{buttonFound:true, buttonText:"她是翻哥\n体验套餐", expanded:"false", loginVisible:false, hasInput:true, readyState:"complete", visibility:"visible", innerWidth:1512, innerHeight:879}` | 通过 |
| `current()` | `{status:"logged_in", name:"她是翻哥", subscription:"体验套餐", raw:"她是翻哥\n体验套餐"}` | 通过;macOS 两行文本解析正确 |
| `openMenu()` | `{action:"pointer_events", expanded:"true"}` | 通过;`pointerdown/pointerup` 同步把 `aria-expanded` 翻成 `true` |
| `hoverMore()` | `{action:"none", count:0}` | 2 个账号时**没有"更多账号"子菜单**(所有账号平铺在一级菜单);≥3 账号的情况本机无法验证 |
| `readEntries()` | 11 个可见 menuitem,见下表 | 通过,但账号行判定要改(见"差异") |
| `closeMenu()` | `{action:"escape", stillOpen:false}` | 修正后通过(见"差异") |
| `switchTo("用户225205")` | `{action:"clicked", name:"用户225205"}` | 点击成功,但**账号没有切换**,见"切换实测" |
| `expandSidebar()` | 折叠态 `{action:"temporary_expand", via:"testid"}`;展开态 `{action:"none", count:0}` | 修正后通过(见"差异") |

`readEntries()` 的 11 条(顺序即 DOM 顺序):

| lines | hasAvatar | hasPopup | data-testid |
|---|---|---|---|
| 设置 | false | false | `profile_menu_setting` |
| 外观设置 | false | **true**(子菜单) | — |
| 收藏夹 | false | false | — |
| 豆包官网 | false | false | `doubao_product_website` |
| 检查更新 | false | false | — |
| 帮助与反馈 | false | false | `profile_menu_feedback` |
| 当前时段额度用量 / 0% | false | false | — |
| 她是翻哥 / 体验套餐 | **true** | false | `avatar_menu_account_row` |
| 用户225205 | **false** | false | `avatar_menu_account_row` |
| 添加账号 | false | false | `avatar_menu_add_account` |
| 退出登录 | false | false | `setup_logout` |

菜单根节点:`[role="menu"][data-testid="chat_header_menu"]`,`aria-labelledby` = 账号按钮 id,`data-side="top"`。菜单里另有一个 `width=0,height=0` 的空 menuitem(DOM 第 1 项),`visible` 过滤后不会出现。

### 与计划的差异(已改进 `accounts.js`)

1. **账号行不能靠 `<img>` 判定**。没有自定义头像的账号(`用户225205`)菜单行里**没有 `<img>`**,只有占位 `div`;有头像的账号行里 `<img alt="avatar">` 存在(macOS 菜单项的 alt 不为空,但账号按钮上的 alt 仍为空,保持不依赖 alt)。改为:优先 `[role="menuitem"][data-testid="avatar_menu_account_row"]`,豆包去掉 testid 时再退回"无 `aria-haspopup` + 有 `<img>`"启发式。`readEntries()` 的每条增加 `isAccountRow`、`testId` 两个字段,原字段不变。`switchTo()` 改为只在账号行里按解析出的 `name` 匹配。
2. **当前账号行的标记**:当前账号那一行多一个勾选 `svg`(`class="size-14 ..."`),非当前行没有 svg。Task 1.7 若需要"当前账号"也可以从这里读,但 `current()` 读左下角按钮已足够。
3. **`closeMenu()` 的 `stillOpen` 不能看 `document.querySelector('[role="menu"]')`**:页面常驻一个 0 尺寸、0 子项的空 `[role="menu"]`(`id=radix-:r46v:`,`data-state="open"`),导致永远 `stillOpen=true`。改为:有账号按钮就只看它的 `aria-expanded`;找不到按钮才看可见账号行数。
4. **`closeMenu()` 改为 async**。合成 `KeyboardEvent('keydown', Escape)` 派发到 `document` 即可关闭(Radix DismissableLayer 监听 document),但 `aria-expanded` 要等 React 提交后约 50ms 才翻 `false`,账号行在退出动画中再多留约 200ms(实测时间线:0ms `true/2行`,55ms `false/2行`,209ms `false/0行`)。同步读会误判并触发 pointer 兜底(再点一次按钮 = 重新打开)。现在:Escape → 等 150ms → 仍开着才 pointer 兜底 → 再等 150ms。连续两轮 open/close 均返回 `stillOpen:false`。调用方必须 `awaitPromise:true`。
   - 三种关闭方式都验证有效:合成 Escape keydown、CDP `Input.dispatchKeyEvent`(keyDown+keyUp,`windowsVirtualKeyCode=27`,探针加了 `--key Escape`)、再次 pointer 点击账号按钮。Rust 侧用脚本里的合成 Escape 即可,不必走 `Input.dispatchKeyEvent`。
5. **`expandSidebar()` 的位置启发式不命中**。折叠态的"显示侧栏"按钮在 `(93,15) 26x26`,`data-testid="siderbar_closed_status_btn"`,`aria-label="显示侧栏"`;计划里 `left<50` 匹配 0 个。改为优先 testid/aria-label,启发式放宽到 `left<120`。展开态的"收起侧栏"按钮是 `data-testid="siderbar_close_btn"`,`aria-label="收起侧栏"`,位于 `(296,15)`。
6. **折叠侧栏时账号按钮仍在 DOM**,rect 为 `(-299,823,231,48)`(整个侧栏平移到屏外),`visible` 判定仍为 true;`findAccountButton()` 走"唯一候选"兜底能找到,`current()`/`probe()` 仍正确。但 pointer 事件坐标在屏外,`openMenu()` 在折叠态未验证——Task 1.7 在 `openMenu` 前若按钮 `left<0` 应先 `expandSidebar()`。

### 账号按钮 DOM 特征(供 Task 1.7)

- `button[aria-haspopup="menu"][data-slot="dropdown-menu-trigger"]`,`id=radix-:r2n:`,`aria-controls=radix-:r2o:`(菜单 id),**没有** `data-testid` / `aria-label`。
- rect `(9,823,231,48)`,`innerHeight=879`:满足计划的位置过滤(`left<80`、`top>innerHeight-120`、`bottom<=innerHeight+2`、宽 35–360、高 20–85),且是唯一命中。
- 页面里 `aria-haspopup="menu"` 的按钮有 27–31 个(对话列表每行的"更多"按钮、上传按钮等,数量随对话数变化),**只有账号按钮含 `<img>`**;但位置过滤本身已经唯一命中(底部高度带里只有它),所以 `<img>` 不是必要条件——见下文"审查后修正 1"。
- 按钮里的头像是裸 `<img class="rounded-full object-cover mr-8 h-36 w-36">` 直接放在 div 里(不是带 fallback 的 Avatar 组件),与菜单账号行 `avatarUrl ? <img/> : <div/>` 极可能是同一条渲染路径——即**当前账号为默认头像时按钮里很可能没有 `<img>`**。
- 输入框:`[contenteditable="true"]` 存在(`hasInput=true`)。

### 切换账号实测(UC-04 相关,重要)

- 菜单里点击 `用户225205`(`.click()`,返回 `clicked`)后:
  - 0–12.9s:chat target 不变(`2216B0F8`),WebSocket 不断,`current()` 一直是 `她是翻哥`,页面没有 reload。
  - +13.2s:`/json/list` 新增一个 `type="page"` 的 target:`https://www.doubao.com/client-login-modal?enterMethod=add_account&entry=app_main_chat&feature=add_account&...`,随后 302 到 `https://www.doubao.com/flow-account/client-login/client-login-modal?...`——即豆包弹出了**登录窗口**(独立 BrowserWindow)。
  - 之后该 target 自行消失(可能被后续在主窗口派发的 pointer 事件关掉),`用户225205` **从菜单里消失**,菜单不再有任何 `avatar_menu_account_row`(只剩 `添加账号` / `退出登录`),当前账号仍是 `她是翻哥`。
- 解读:`用户225205` 的本地登录态已失效,豆包的切换流程是"校验目标账号 token → 失效则弹登录窗并把它从列表移除"。整个过程**没有 reload 主页面、没有换 target**。
- 本机现在只剩 1 个有效账号,**成功切换(页面 reload、target 失效、重连)这条路径无法在本机验证**;Task 1.8 的重连逻辑仍按 03-account-module 的 Windows 观察(切换会 reload)实现,并用 mock WS 覆盖。
- 对 Task 1.7/1.8 的要求:
  1. 切换验证循环期间要轮询 `/json/list`,**出现 host 为 `www.doubao.com` 且 path 含 `client-login-modal` 的 page target 即判定 `account_switch_failed`**(目标账号需要重新登录),不要等 20s 超时;同时不要向该 target 建 session。
  2. 只有 1 个账号时菜单**没有账号行**,`account_list` 应返回 `[current]`(来自 `current()`),而不是空列表。
  3. 验证循环里 `current()` 读的是按钮文本,按钮文本在切换成功前不会变,可以直接用作"是否已切换"的判据。

### 窗口隐藏 / 最小化时的页面状态

| 操作(`osascript` System Events) | `document.visibilityState` | `document.hidden` | `document.hasFocus()` | DOM 查询 |
|---|---|---|---|---|
| 正常前台 | `visible` | false | true | 正常 |
| `set visible of process "Doubao" to false`(隐藏 App) | **`hidden`** | true | true | `probe()` 全部正常 |
| `AXMinimized = true`(最小化窗口) | **`hidden`** | true | true | `probe()` 全部正常 |
| 恢复 | `visible` | false | true | 正常 |

结论:macOS 上最小化/隐藏后 `visibilityState` 会变成 `hidden`(与 DECISIONS.md Phase 0"可能可选"的猜测不同),但 `getBoundingClientRect`、`innerWidth/innerHeight` 与 DOM 读取不受影响,**账号读取不需要** `Page.setWebLifecycleState` / `Emulation.setFocusEmulationEnabled`。生成/捕获链路(定时器节流、SSE)在 `hidden` 下是否受影响待 Task 2.1 验证,届时再决定是否启用这两条命令。

### 探针工具

`scripts/cdp-probe.mjs`:`node scripts/cdp-probe.mjs [--port 9707] [--target chat|background] [--file path.js] [--key Escape] [--timeout 30000] '<expr>'`。`--file` 先注入脚本再求值,`--key` 用 `Input.dispatchKeyEvent` 按键。整体看门狗默认 30s(超时打印 `timeout after Nms`、exit 2),单条 CDP 命令 15s 超时(与 spec 09 一致),WS 出错/关闭时挂起的命令全部 reject;实测 `--timeout 2000 'new Promise(()=>{})'` 2.04s 退出。注意:`accounts.js` 的幂等守卫按 `version` 判断,改脚本时要同时改 `VERSION`,否则同版本重复注入是 no-op(或先 `node scripts/cdp-probe.mjs 'delete window.__doubaoRelayAccounts'`)。

### 审查后修正(accounts.js 1.1.0,2026-10-16)

1. **`findAccountButton()` 不再硬性要求 `<img>`**。新策略:候选 = `aria-haspopup=menu` 且不在 `[role=menu]` 内的可见按钮 → 左下角位置带唯一命中即返回(不看 img)→ 位置带内多个时依次用"含 `<img>`"、"文本能解析出套餐"收窄 → 侧栏折叠时(按钮 `left<0`,实测 `(-299,823)`)用 `left -400–0` 的同一高度带再找 → 最后退回全页唯一含 `<img>`/含套餐文本的候选。真机验证:haspopup 候选 28 个,位置带内只有账号按钮 1 个;**在真机 DOM 里临时摘掉按钮内 `<img>` 后 `probe().buttonFound=true`、`current().status='logged_in'`**,放回后恢复。
   - **未验证风险(Task 1.7)**:当前登录账号为默认头像时按钮里是否真的没有 `<img>`,本机无法切到这类账号证实(仅剩 1 个有效账号且有自定义头像),属推断。Task 1.7 的 `live_` 测试在 `current().status !== 'logged_in'` 时必须打印 `probe()` 全量——`probe()` 新增 `buttonHasImg`、`haspopupCount`、`candidates[]`(位置带内候选的 text/hasImg/hasSub/rect)便于定位。
2. **版本守卫**:对象带 `version: '1.1.0'`,守卫改为 `window.__doubaoRelayAccounts?.version === VERSION` 才跳过,不同版本整体覆盖(relay 升级后豆包页面不 reload 也能换新脚本)。Task 1.7 注入后断言 `version` 与 Rust 侧常量一致。
3. **`menuOpen()` 无按钮时的兜底**改为"菜单根 `[role=menu][data-testid=chat_header_menu]` 可见 或 可见 menuitem 里有`退出登录`"(两者任何账号数下都存在;单账号时没有账号行,原来按账号行数兜底会把打开的菜单判成关闭)。真机:打开时菜单根 rect 非 0、`退出登录` 可见 1 项;关闭后菜单根从 DOM 移除、可见 `退出登录` 0 项。无按钮路径下 `closeMenu()` 在 Escape 后多等一轮 150ms(退出动画约 200ms)。
4. **`switchTo()`** 0 匹配返回 `{action:'not_found',count:0}`,≥2 匹配才是 `ambiguous`(对应 11-error-codes 的两个不同 code)。真机 `switchTo('__nope__')` → `not_found`。
5. **`expandSidebar()`** 去掉 `aria-label === '显示侧栏'` 分支(Global Constraints:不依赖 aria-label),只留 testid → 位置启发式。真机折叠态仍经 testid 命中。
6. 测试:`frontend/src/accounts-script.test.ts`(vitest + happy-dom,`?raw` 读取 `src-tauri/src/doubao/scripts/accounts.js`,手工 stub `getBoundingClientRect`)覆盖以上 1–5;vite 配置 `server.fs.allow` 加了 `../src-tauri/src/doubao/scripts`。注意 happy-dom 的 `innerText` 与浏览器一致:块级 div 之间才产生换行,fixture 必须按真豆包结构用两个 div 表示"名字 / 套餐"。

## Task 1.4 · CDP 客户端 `cdp::session` / `cdp::discovery`(2026-10-16)

`cargo test live_ -- --ignored --nocapture` 在本机真豆包(2.32.5,端口 9707)上的观察:

- `GET /json/version`:`Browser: Chrome/147.0.7727.149`,`Protocol-Version: 1.3`,带 `webSocketDebuggerUrl: ws://127.0.0.1:9707/devtools/browser/<uuid>`(browser 级 target,本项目不连它;白名单只认 `/devtools/page/`)。
- `GET /json/list` 4 条:`page doubao://doubao-chat/chat`、`iframe https://www.doubao.com/drive-iframe/...`、`other doubao://doubao-chat/cross-site-support/`、`page doubao://doubao-background/`;`select_pages(…, 9707)` 得 chat 1 个、background 1 个,与 Task 1.2 一致。target id 自 Task 1.2 以来未变(chat `2216B0F8…`),说明期间主页面没有 reload。
- 从 **page session** 发 `Target.getTargets` 可用(`targetInfos` 4 条),可作心跳命令(Task 1.8)。
- `Runtime.evaluate` + `returnByValue` + `awaitPromise`:返回 Promise 的表达式正常解析为值;页面内 `location.href` 为 `chrome://doubao-chat/chat`(与 `/json/list` 的 `doubao://` 不同,白名单已同时接受)。`throw new Error('探针异常')` 的 `exceptionDetails.exception.description` 为 `Error: 探针异常\n    at <anonymous>:1:…`,`extract_js_exception` 取第一行去 `Error: ` 前缀得 `探针异常`。
- 注意:`Runtime.evaluate` 不带 `replMode` 时表达式里不能用顶层 `await`(语法错误),异步要写成 `Promise.then(...)` 或 async IIFE。
- 连接/关闭:`CdpSession::connect` 握手 <10ms;`close()` 发 Close 帧后豆包立即回 Close,读循环正常退出,不需要 abort。

## Task 1.5 · 平台适配层 `platform::macos`(2026-10-16)

`cargo test platform::macos::tests::macos_only::live_ -- --ignored --nocapture` 在本机的观察:

- 豆包进程树(`ps -axo pid,comm`):主进程 `/Applications/Doubao.app/Contents/MacOS/Doubao`;浏览器内核在 `Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser`,其 Helper / Renderer / Accessory / crashpad 都在 `Doubao Browser.app/Contents/Frameworks/Doubao Browser Framework.framework/Versions/147.0.7727.149/Helpers/` 下。`Contents/Helpers/` 里另有 `豆包.app`、`relaunch_helper`、`saman_updater`、`libmcp_helper.dylib`。所有这些路径都含 `/Doubao.app/`,`is_doubao_exe_path` 按此判定即可覆盖全部子进程。
- `lsof -nP -iTCP:9707 -sTCP:LISTEN -Fp` 输出 `p74576\nf57`(只有 `p` 行是 pid);pid 74576 经 sysinfo 0.38 `exe()` 得 `/Applications/Doubao.app/Contents/Helpers/Doubao Browser.app/Contents/MacOS/Doubao Browser`——`live_listening_process_exe` 通过,证实 Global Constraints 里"监听者是 Doubao Browser 子进程"的判断。
- `find_doubao_exe` 走"已运行进程 → 回推 `Doubao.app` 包 → `Contents/MacOS/Doubao`"返回 `/Applications/Doubao.app/Contents/MacOS/Doubao`。
- `defaults read /Applications/Doubao.app/Contents/Info.plist CFBundleShortVersionString` → `2.32.5`。
- sysinfo `System::new()` + `refresh_processes_specifics(…, ProcessRefreshKind::nothing().with_exe(UpdateKind::Always))` 在本机对所有进程枚举 <50ms;`listening_process_exe` 非 live 测试(绑一个临时端口后查自己)也通过,可作回归。
- `launch_doubao_with_cdp` / `kill_doubao` 本 Task 未在真机执行(禁止无故杀豆包),留给 Task 1.10 的 UC-06 验收。
- Windows 交叉编译:`cargo check --target x86_64-pc-windows-msvc` 整个 crate 失败在 `ring`(reqwest rustls 依赖)的 build.rs:`cc --target=x86_64-pc-windows-msvc` 找不到 `assert.h`(本机没有 MSVC 头文件),与本 Task 代码无关;`cargo check` / `cargo clippy --all-targets -- -D warnings` 用 **`x86_64-pc-windows-gnu`**(homebrew mingw-w64 可编 ring)整个 crate 通过;另用只含 `error.rs` + `platform/*` 的临时 crate 对 `x86_64-pc-windows-msvc` 跑 check + clippy 也通过。
- (Task 1.5 审查修复)`ps -axo pid,ppid,comm` 复核进程树:除主进程 `Doubao`(ppid=1)与 `Doubao Browser`(ppid=主进程)外,还有两条 **launchd 直属(ppid=1)** 的豆包进程:`Doubao Browser.app/Contents/XPCServices/AgentInfraService.xpc/Contents/MacOS/AgentInfraService` 和 `.../Helpers/chrome_crashpad_handler`。它们可能比主进程晚退出,因此 `is_doubao_running` / `wait_until_gone` 改用 `doubao_core_processes()`(排除 `/XPCServices/` 与文件名含 `crashpad` 的进程),`signal_all` 仍对全部豆包进程发信号。`kill_doubao` 的 `osascript quit` 加 5s 超时(`status_with_timeout`,`kill_on_drop`),最坏链路 5s+5s+1s+1s+1s=13s,在"重连 20s 失败升级重启"预算内;非 live 测试用 `sh -c 'exec sleep 60'` 验证超时返回且子进程不残留。`is_doubao_exe_path` 改为按路径分量判定:`Doubao.app` 不区分大小写,或 `*豆包*.app`(与 `find_doubao_exe` 第 3 级一致);真机 `Doubao.app/Contents/Helpers/豆包.app` 嵌套包由 `app_bundle_from_exe` 取最外层回推到 `Doubao.app`。live_ 5/5 复跑通过(路径/版本结论不变)。

## Task 1.6 · 豆包启动 `doubao::bootstrap` / 兼容性自检 `doubao::compat`(2026-10-16)

`cargo test live_ -- --ignored --nocapture` 在本机真豆包(2.32.5,端口 9707)上的观察:

- `ensure_doubao_started(platform, 9707)` → `AlreadyConnected`:`/json/version` 通,`listening_process_exe(9707)` 为 `Doubao.app/Contents/Helpers/Doubao Browser.app/.../Doubao Browser`,`is_doubao_exe_path` 判为豆包。启动/重启分支(kill → launch → wait_for_cdp)本 Task 未在真机执行(禁止无故杀豆包),用假 `Platform` + 本地 HTTP mock 覆盖,真机验收留给 Task 1.10 的 UC-06。
- **发送按钮 selector 与计划不符**。计划写 `button[aria-label*="发送"],button[type="submit"]`,真机命中 **0** 个。实测:
  - 发送按钮是 `button[data-testid="chat_input_send_button"]`,包在 `div.send-btn-wrapper` 里,`aria-label=""`(空)、无 `type` 属性、`aria-disabled="false"`,rect `(1350,817,36,36)`(输入区右端)。
  - **只有输入框有文字时才渲染**。空输入时同一槽位是一个麦克风 `div`(`cursor-pointer`,`aria-busy="false"`,无 testid / role),整页没有任何 `[data-testid*="send"]`。
  - 输入区结构:编辑器 `div.tiptap.ProseMirror[contenteditable="true"]`(无 testid,空时 `innerText === "\n"`),外层 `[data-testid="chat_input_input"]` → `.guidance-input-editor-viewport` → `.guidance-input-surface`;动作栏 `.guidance-input-actions` 里按钮:`upload_file_button`、`chat_input_action_mode`("对话")、`guidance-skill-bar`(`skill_bar_button_3` 图像生成 / `skill_bar_button_17` 视频生成 / `skill_bar_button_5000` PPT / `skill_bar_button_1005` 录音转写 / `skill_bar_button_5003` AI 表格 / `skill_bar_button_2` 帮我写作 / `skill_bar_button_more`)、`chat_input_action_model`("豆包 快速")。这些 testid 对 Task 2.1/2.3 定位"图像生成 / 视频生成"入口与模型选择有用。
  - `compat.rs` 的做法:`SEND_BUTTON_SELECTOR = [data-testid="chat_input_send_button"],.send-btn-wrapper button,button[aria-label*="发送"],button[type="submit"]`;直接命中即 `direct`;没命中且输入框为空时 `ed.focus(); execCommand('insertText', false, 'x')` → 等 250ms → 再找(`transient_text`)→ `execCommand('selectAll'); execCommand('delete')` 还原(原先无焦点则 `blur`)。真机:`before="\n"`、`direct=false`、`transient=true`、还原后 `after="\n"` 且发送按钮随即消失。输入框**非空**(用户有草稿)时不动用户内容,此时若 selector 没命中则 `send_button_found=false` → `status=warn`。
  - `live_compat` → `UiCompat { status: "ok", account_button_found: true, input_found: true, send_button_found: true }`。
- `tiptap` 编辑器接受 `document.execCommand('insertText')`(ProseMirror 监听 beforeinput),Task 2.3 填 prompt 可沿用;`selectAll + delete` 能清空且不留残余段落。

## Task 1.7 · 账号读取 `doubao::accounts`(2026-10-16)

`cargo test live_read_accounts -- --ignored --nocapture` 与 `scripts/cdp-probe.mjs` 在本机真豆包(2.32.5,端口 9707,仅 1 个有效账号)上的观察:

- `read_accounts` 全流程(注入 → probe → current → openMenu → 650ms → hoverMore → 650ms → 4 轮 hoverMore/400ms/readEntries → closeMenu)耗时 **3106ms**,返回 `AccountList { current: Some("她是翻哥"), accounts: [她是翻哥 / 体验套餐, is_current=true], read_at }`;读取后再调 `closeMenu()` 得 `{action:"none", stillOpen:false}`,菜单确实已关闭,脚本内合成 Escape 足够,`Input.dispatchKeyEvent` 兜底未触发。
- **单账号时 `readEntries()` 共 9 条可见 menuitem,没有任何 `avatar_menu_account_row`**(设置 / 外观设置[子菜单] / 收藏夹 / 豆包官网 / 检查更新 / 帮助与反馈 / `当前时段额度用量\n0%` / 添加账号 / 退出登录),`hoverMore()` → `{action:"none", count:0}`。`parse_entries` 过滤后为空,按计划退化为 `[current]`(`label` 取 `current().raw` 即按钮两行文本 `她是翻哥\n体验套餐`)。
- **`parse_entries` 不能把 `hasAvatar` 当必要条件**(计划照抄的沈浪规则是 `!avatar → 跳过`):Task 1.2 已实测无自定义头像的账号行没有 `<img>`(`用户225205` `hasAvatar=false`),若按计划过滤会丢掉这类账号。改为"`isAccountRow`(页面按 `data-testid="avatar_menu_account_row"` 判定,去掉 testid 时退回无子菜单+有 `<img>`)**或** `hasAvatar`"二者之一即视为账号行,其余过滤(子菜单 / 白名单 / 额度前缀 / name 长度 1..=120)不变。`RawEntry` 相应增加 `#[serde(default)] is_account_row` / `test_id` 两个字段(老结构缺字段时默认 false/空,不影响反序列化)。
- **侧栏折叠时不需要先展开**:用 `siderbar_close_btn` 收起侧栏后,`probe().buttonFound=true`(候选 rect `(-299,823,231,48)`,屏外),`openMenu()` → `{action:"pointer_events", expanded:"true"}`,`readEntries()` 得 9 条 menuitem、`logoutVisible=true`,`closeMenu()` → `{action:"escape", stillOpen:false}`——合成 pointer 事件不看坐标是否在视口内。所以 Rust 侧只在 `probe().buttonFound=false` 时才 `expandSidebar()`(Task 1.2 第 6 条的顾虑不成立)。展开后的侧栏不自动收回(与沈浪 `temporary_expand` 一致)。
- `current().status === 'logged_in'`,按钮含 `<img>`(`buttonHasImg=true`),"默认头像账号的按钮是否无 `<img>`"仍无法在本机证实;`live_read_accounts` 在 `status !== 'logged_in'` 时会打印 `probe()` 全量。
- 注入方式:`(() => { if (!window.__doubaoRelayAccounts || .version !== "1.1.0") { <accounts.js> } return window.__doubaoRelayAccounts.version; })()`,一次 `Runtime.evaluate` 完成注入 + 版本断言;脚本自身的版本守卫让重复注入为 no-op。

## Task 1.8 · 连接管理 `doubao::connection`(2026-10-16)

`cargo test doubao::connection::tests::live_ -- --ignored --nocapture` 在本机真豆包(2.32.5,端口 9707,仅 1 个有效账号)上的观察:

- `DoubaoConnection::connect(false)` → `AlreadyConnected`,**397ms**(`/json/version` + `lsof` + `/json/list` + 连 chat/background 两个 session + `Network.enable` + UI 兼容性自检 + `defaults read` 版本)。`status()`:`running=true, cdp_connected=true, doubao_version_detected="2.32.5", ui_compat_check="ok", chat_targets=1, reconnecting=false`。
- `Network.enable`(`maxTotalBufferSize=16777216, maxResourceBufferSize=8388608`,spec 09 的握手参数)在 chat page session 上返回 `{"id":1,"result":{}}`;它的响应比同时发出的 `Target.getTargets` 更晚回来(Node 直连实测 id 2 先于 id 1),Rust 侧按 id 匹配不受影响。
- 心跳 `Target.getTargets`(page session)真机 **1ms** 返回;5s 超时裕量充足。
- `accounts(false)` 首次 3111ms(与 Task 1.7 的 `read_accounts` 一致),2s 内再次调用直接返回缓存(`read_at` 相同)。
- `live_switch_roundtrip`:本机只有 1 个有效账号,**自动 SKIP**(打印 `SKIP: 本机只有 1 个有效账号`)。"点击 → 页面 reload → 老 session 关闭 → target id 变化 → 重连 → `current()` 变为目标账号"这条路径用 mock 豆包覆盖(`switch_account_succeeds_through_reload_and_reconnect`:同一端口同时服务 `/json/*` 与 WebSocket,`switchTo` 返回 `clicked` 时服务端关掉所有 socket 并把 target id 换掉)。
- Task 1.2 发现的"切换到登录态失效账号会弹 `client-login-modal` page target"已落实为 `is_login_modal_target()`:验证循环每轮先查 `/json/list`,出现 `www.doubao.com` + 路径含 `client-login-modal` 的 page 立即返回 `account_switch_failed(reason=login_required, login_modal_url)`,不等 20s,也不向该 target 建 session(mock 覆盖:`switch_account_fails_fast_when_login_modal_appears`)。本机没有失效账号可再次触发,未在真机复测。
- `Page.setWebLifecycleState` / `Emulation.setFocusEmulationEnabled` **没有发**(Task 1.2 结论:macOS 最小化/隐藏不影响 DOM 读取);生成链路是否需要留 Task 2.1 决定。

### 审查后修正(Task 1.8,2026-10-16)

- **`live_` 并行互相干扰**:`cargo test live_ -- --ignored`(默认多线程)里 `accounts::live_read_accounts`、`connection::live_connect_status_and_heartbeat`、`connection::live_switch_roundtrip` 三个都要开/关同一个账号菜单,审查者一次复现 `doubao_ui_changed { failed_selectors: ["[role=\"menuitem\"]"], open_action: "already_open" }`。修复:`doubao::LIVE_MENU_LOCK`(`#[cfg(test)]` 的进程级 `tokio::sync::Mutex<()>`),三个测试先 `lock().await` 再跑;计划里的 live 命令不变。
- **关菜单后立刻再开会被豆包忽略(真机新发现,影响生产路径)**:串行化之后仍有 1/3 概率失败,但形态变成 `open_action: "pointer_events"` + 0 个 menuitem(菜单根本没打开)。用 `scripts/cdp-probe.mjs` 注入 `accounts.js` 后在页面里循环 `openMenu → closeMenu → 等 gap → openMenu → 700ms 后 readEntries().menuItemCount`:gap=0 → **4/4 为 0**(`aria-expanded` 仍 `false`);gap=150ms → 3/3 为 9;gap=300ms → 3/3 为 9。即 `closeMenu()`(Escape + 150ms)返回后菜单退出动画尚未结束,这段时间里的 pointer 事件被丢弃。`switch_account` = `accounts_locked(true)`(读完关菜单)→ 立刻 `click_account`(开菜单)正是这条路径,真机有 2 个账号时会稳定 `account_switch_failed(account_row_not_clickable)`。修复:`ReadTiming` 增 `after_close`(默认 300ms),`read_accounts_with` 在 `closeMenu` 之后等它再返回;`read_accounts` 真机耗时从 ~3.1s 变为 ~3.4s。修复后 `cargo test live_ -- --ignored --nocapture` 连续 5 次 10/10 通过。
- **心跳 `running` 字段**:进程退出后原实现要到 `crash_after`(20s)之后的那次心跳才查进程,期间 `running=true, reconnecting=true`,UC-06 "15s 内显示豆包已退出"无法满足。现在心跳失败分支每轮 `is_doubao_running()`(macOS `spawn_blocking` 进程枚举,10s 一次),`running=false` 时 `reconnecting=false` 且跳过 `/json/list` 重连;`Crashed` 仍按 ≥20s 且 `!running` 只发一次。本机未 pkill 豆包(留给 Task 1.10 的 UC-06 验收),mock 用例 `heartbeat_marks_not_running_before_crash_after_without_crashed` 覆盖。
- `accounts.js` 1.1.1:`switchTo` 的 `ambiguous` 返回带 `labels`(同名各行原文),Rust `account_ambiguous.details.matches` 透传;本机单账号未触发,DOM 测试(happy-dom)覆盖。

## Phase 1 验收 · Task 1.10 前端账号页 + 联调(2026-10-16)

环境同上(豆包 2.32.5,仅 1 个有效账号 `她是翻哥 / 体验套餐`)。`cargo tauri dev` 后台运行,relay 窗口为 Tauri WKWebView。

### 验证手段(screencapture 不可用的替代)

- **真 relay 窗口**:macOS 未授予屏幕录制权限(`screencapture -x` 报 `could not create image from display`),但 **Accessibility 已授权**,`osascript` System Events 能读 WKWebView 的无障碍树(`entire contents of window 1 of process "doubao-relay"` → `AXStaticText`/`AXButton` 的 name/value/enabled)并 `click` 按钮。真机各项以 AX 文本为证据,配合 relay 日志(`~/Library/Logs/doubao-relay/doubao-relay.2026-10-16.log`)。
- **截图**:用 Playwright(普通 Chromium)打开 `cargo tauri dev` 起的 vite `http://localhost:1420`,`addInitScript` 注入假的 `window.__TAURI_INTERNALS__`(`invoke` / 事件回调),payload 形状与 Task 1.9 的 camelCase 输出一致,驱动多账号 / 切换 / 未登录 / 退出 / 重连中 / 骨架屏等状态逐项截图。脚本 `scripts/ui-harness.playwright.js`(传给 playwright MCP 的 `browser_run_code_unsafe`),截图在 `docs/analysis/screenshots/phase1/`。

### 逐项结果

| UC | 手段 | 结果 |
|---|---|---|
| UC-01 启动豆包连接(CDP 已开 → 复用) | 真机:relay 启动 `auto_connect_on_startup` → 日志 `豆包已连接 action=AlreadyConnected`;点"重连豆包"→ 日志再出一条 `AlreadyConnected`,状态栏保持"● 豆包 已连接" | 通过 |
| UC-01(豆包未开 → 自动启动) | 真机(作为 UC-06 的恢复步骤):`pkill -x Doubao` 后点"重连豆包"→ 日志 `豆包未运行,准备启动 exe=/Applications/Doubao.app/Contents/MacOS/Doubao` → `豆包已启动,CDP 就绪 elapsed_ms=859` → `chat session 已建立` → `豆包已连接 action=Started`;UI 4.6s 内变回"已连接" | 通过(但见下文"启动后页面未渲染") |
| UC-02 账号列表(1 账号) | 真机 AX:`当前账号: 她是翻哥 (体验套餐)`;表格行 `★ 她是翻哥 / 体验套餐 / 当前`;`共 1 个已登录账号。`;点"刷新"→ 日志新增一条 `账号列表读取完成 count=1` | 通过 |
| UC-03 多账号 + 同名 ⚠ | mock(本机只有 1 个账号):4 行,`★ 她是翻哥 / 体验套餐 / 当前`、`李四 / 高级套餐 / [切换]`、两行 `⚠ 王五 … 同名,不支持通过 API 调用`(无切换按钮) → `uc03-multi-accounts.png` | 通过(mock) |
| UC-04 账号切换 | mock:点"切换"→ 确认框(`uc04-confirm-dialog.png`)→ 确认后行内"切换中…"转圈(`uc04-switching.png`)→ `account_switched` + `account_list_changed` 事件后当前账号变为 `李四 (高级套餐)`,toast"已切换到账号“李四”"(`uc04-switched.png`);后端报 `account_switch_timeout` 时错误 toast 常驻、成功 toast 3s 自动消失(`uc04-switch-error-toast.png`)。真机只有 1 个账号,无法走真实切换 | 通过(mock);真机 SKIP |
| UC-05 未登录 | mock:`account_list` 抛 `doubao_not_logged_in` → 空状态"豆包尚未登录任何账号 / 请先在豆包中登录,然后点击刷新"+[刷新](`uc05-not-logged-in.png`) | 通过(mock) |
| UC-06 豆包崩溃 | **真机**(本 Task 唯一一次 `pkill -x Doubao`,15:10:10):心跳在 +6s 发现(`心跳失败,标记豆包连接断开` / `豆包进程不在,等待其重新启动`),AX 在 **+9.7s** 读到状态栏"豆包已退出,等待重启"、空状态同文案 + [重连豆包]、"刷新"按钮禁用、错误 toast"豆包已退出,等待重启。点击“重连豆包”可由 relay 重新启动";点"重连豆包"后 relay 拉起豆包(见 UC-01 第二行),4.6s 后状态栏"已连接"。mock 截图 `uc06-doubao-exited.png`(状态栏红)/ `uc06-recovered.png` / `uc06-reconnecting.png`(重连中黄) | 通过 |
| 骨架屏 | mock:`account_list` 挂起时首屏骨架屏(`loading-skeleton.png`) | 通过 |

### 真机新发现:relay 拉起豆包后页面要再等几秒才可操作

`connect()` 在 `Started` 路径里 `wait_for_cdp` 返回(启动后 859ms `/json/version` 可用)时,`doubao-chat` page target 已存在、session 能建、`Runtime.evaluate` 能跑,但 **DOM 还是空的**:紧接着的 UI 自检得 `status=error`(三项全 false),前端随后的 `account_list` 报 `doubao_ui_changed`(`账号按钮未找到且无登录按钮,判定豆包界面变化 haspopup_count=0`),账号页落到"还没有读到账号列表"空状态;约 2s 后再点"刷新"即正常(`probe()` 全部命中)。修复:`DoubaoConnection::connect` 在 `Started`/`Restarted` 后先轮询 `PAGE_READY_EXPRESSION`(`readyState==='complete'` 且有 `[contenteditable="true"]` 或页面文本含"登录"),每 500ms 一次、最多 15s(`ConnTiming.page_ready_timeout/page_ready_poll`),就绪或超时后再自检;`AlreadyConnected`/`Reconnected` 不等。mock 用例 `connect_after_launch_waits_for_page_ready_before_compat_check` / `wait_for_page_ready_times_out_without_error` 覆盖;因只允许杀一次豆包,修复后的 Started 路径**未在真机复跑**,留 Task 6.1 断线重连硬化时一并验证。

### 其它观察

- Tauri WKWebView 的 AX 树里 `<table>` 会被 `entire contents` 重复枚举(同一行出现 3 次),只影响读取方式不影响 UI。
- `confirm-dialog` 的 `autofocus` 在 Chromium 下报 "Autofocus processing was blocked because a document already has a focused element"(点击"切换"按钮后焦点仍在按钮上),仅 console 提示,Enter/Escape 快捷键由 `keydown` 监听处理,不依赖焦点。
- vite 首次被 Tauri 窗口加载时 `optimized dependencies changed. reloading` 会自动刷新一次 WebView,启动后 `doubao_status` / `account_list` 各调用一次,无重复读取。
## Task 5.1 · 设置页 / 日志页 / 调试面板(2026-10-16)

环境同上(豆包 2.32.5 运行中、1 个有效账号),`cargo tauri dev` 起的真 relay 窗口(WKWebView)用 `osascript` System Events 驱动,Playwright(普通 Chromium + vite + 假 `__TAURI_INTERNALS__`,脚本 `scripts/ui-harness-phase5.playwright.js`)截图到 `docs/analysis/screenshots/phase5/`。本 Task 没有执行任何消耗豆包配额的操作。

### 真 relay 窗口(真后端 + 真插件)

| 项目 | 手段 / 证据 | 结果 |
|---|---|---|
| 设置页六个分组按 spec 06 渲染 | AX 树:`HTTP API 服务器`(端口 8787 / `aab9…0d26` 打码 token / 显示·复制·重新生成 / curl 示例)、`并发`(3 / 1)、`结果保留`(1 天 / 7 天 + 立即清理)、`日志`(info + CDP 复选框)、`高级`(9707 + 两个开关)、`关于`(0.1.0 · macos · Tauri 2.12.4 · rustc 1.94.0,三个按钮禁用) | 通过 |
| 复制 token(clipboard-manager 插件 + capability) | 点"复制"后 `pbpaste` 为 64 位 hex,前缀 / 后缀与打码显示一致;**60s 后 `pbpaste` 为空**(16:02:20 复制,16:04:03 已清空) | 通过(spec 12 的 60s 清除) |
| 开启 CDP 日志热生效 | 勾选 + "保存日志" → relay 日志 `CDP 调试日志已开启 file=~/Library/Logs/doubao-relay/cdp-debug.log` + `配置已保存 needs_restart=false`;文件立刻创建,一个心跳周期(12s)后 23 行(`Target.getTargets` 收发 + `Network.*` 事件);relay 被 `cargo tauri dev` 重编译重启后(配置里已是 `true`)启动路径继续写同一文件 | 通过 |
| 关闭 CDP 日志热生效 | 取消勾选 + "保存日志" → `CDP 调试日志已关闭`;文件 2823 行,再等 13s 仍 2823 行;配置回到 `cdp_debug_log=false`(与本机原值一致) | 通过 |
| 日志页读真日志 | AX 树里出现真后端条目(`doubao-relay 启动 cdpDebugLog=true …`、`CDP 调试日志已开启 file=…`、`豆包已在运行且 CDP 端口可连`),工具栏 级别(全部)/ 清空 / 导出 | 通过 |
| 导出(dialog 插件 + capability + `log_export`) | 点"导出"→ 出现原生 NSSavePanel sheet(Where / Save As / Tags / Cancel / Save)→ 填名点 Save → relay 日志 `日志已导出 count=8`,文件 8 行 ndjson(首行即 `doubao-relay 启动`) | 通过(文件落在 sheet 当时的 Where 目录,自动化把绝对路径填进了文件名栏,是脚本问题不是产品问题;已删除) |
| 日志级别热更新 | 未在真机改级别(改了要再改回,只有单测 + `logging::set_level` 的 reload 句柄覆盖);mock 走通 | 单测覆盖 |

### 真机新发现

1. **CDP 日志量**:`Network.enable` 开着,豆包空闲时 `cdp-debug.log` 约 230 行 / 分钟(12 分钟 2823 行,大部分是 `Network.webSocketFrame*` / `Network.dataReceived`),文件不滚动、不限大小;内存缓冲 2000 条、面板 500 条不受影响。spec 09 定位为"豆包改版诊断用,默认关闭",设置页文案已提示"平时关闭";若之后要长开,需要给文件加滚动。
2. **WKWebView AX 树里同名按钮无法区分**:五个分组的"保存"按钮在 `entire contents` 里同名,而且滚出视口的元素 `position` 全部回报同一个夹住的坐标(y=485),按序号点击命中了别的分组(配置保存成功但 CDP 开关没变,草稿被重读覆盖)。给每个保存按钮加 `aria-label="保存<组名>"` 后按名字点击一次命中。对 Task 2.9 等后续 UI 验收的提示:同名动作按钮都给 `aria-label`。
3. **AX `click` 复选框有效但要给时间**:AXPress 能切换 `<input type=checkbox>` 并触发 `change`(草稿在切页后仍保留),紧接着的 `click`(<0.5s)偶尔读到旧值;两步之间等 1s 稳定。
4. **AX dump 在大表格上极慢**:CDP 面板 500 行 × `<details>` 时 `entire contents` 要数分钟,验收脚本要么先"清空"要么只查小页面。
5. 日志页 tab 标题的"(未开启)"来自 `cdpDebugEnabled` 信号,只在设置页读过配置后才正确;已改为进日志页时先 `debug_cdp_log(limit=1)` 同步 `enabled`。
6. `cargo tauri dev` 下改任何 `src-tauri/src/*.rs` 都会重编译并**重启 relay 窗口**(WebView 重新加载、路由回到账号页),真机验收期间不要动 Rust 文件;改前端文件走 HMR(prefresh),页面模块级信号会被重置(草稿清空)。

### Playwright(mock 后端)截图

`settings-overview` / `settings-token-copied` / `settings-regenerate-confirm` / `settings-saved-needs-restart` / `settings-invalid-port` / `settings-gc-done` / `logs-all` / `logs-filter-warn` / `logs-exported` / `logs-empty` / `logs-cdp-panel` / `logs-cdp-expanded` / `logs-cdp-disabled`。脚本同时断言:token 打码 `0123…cdef`、curl 示例随端口 / token 变化、复制后 mock 剪贴板为完整 token、重新生成走确认框、端口 0 → inline 红字且不调后端、级别 + CDP 开关一次保存、`log_entry` 事件实时追加(8 → 9 行)、warn 过滤只剩 WARN/ERROR 两行、导出把当前级别传给 `log_export`、`cdp_log_entry` 追加(5 → 6 行)、`Cmd+,` / `Cmd+L` / `Cmd+R` 切到 设置 / 日志 / 账号。

- (Task 1.10 审查修复后补充)上一条的"不依赖焦点"是缺陷:全局 Enter 一律当"确定"会把键盘焦点在"取消"上的 Enter 变成切换。现在确认框打开时显式 `focus()` 确认按钮(程序化 focus 不受 autofocus 限制),Enter 落在对话框内按钮上交给按钮自己的 click。Playwright(普通 Chromium,vite + 假 `__TAURI_INTERNALS__`)复验:打开后 `document.activeElement` 即确认按钮;Shift+Tab → "取消" → Enter:对话框关闭、`account_switch` 0 次;把焦点强制挪回行内"切换"按钮再按 Enter:`account_switch` 恰好 1 次、无第二个确认框、行内"切换中…"→ 当前账号变为李四。另外注意 preact `useEffect` 在 rAF 后才跑,自动化里对话框出现后立刻按键会漏掉第一个按键(人手不会),已改 `useLayoutEffect`。`account_list` 抛 `doubao_ui_changed` 时现在有常驻错误 toast 与空状态原因(含 `failed_selectors`),不再只显示"还没有读到账号列表"。

## Task 2.1 · 生成页探测(对话模式 vs 工作模式,2026-10-16)

环境同上(豆包 2.32.5,端口 9707,仅 1 个有效账号 `她是翻哥 / 体验套餐`,窗口 1512×879)。所有探测只跑到 `checkSend`,**没有执行 `send`**,没有消耗配额。探针:scratchpad 里基于 `scripts/cdp-probe.mjs` 同一套 CDP 调用的 Node 脚本(需要监听 `Page.fileChooserOpened` 事件,`cdp-probe.mjs` 不支持事件,未改动它);注入方式 `if (typeof shenlangDoubaoGeneration === 'undefined') { <doubao-generation-page.js> }`,每步 `Promise.resolve(shenlangDoubaoGeneration.run("<step>", <req>))` + `awaitPromise`。

### 先读脚本:`submissionMode` 其实有三条分支,不是两条

把 `doubao-generation-page.js` 的字符串表还原(只在 scratchpad 做,仓库里的脚本原封不动)后,`run('prepare', req)` 的分派是:

```
prepare: submissionMode === 'work'  → 工作模式分支(飞书/豆包"工作"输入区,返回 {mode:'work'})
         else type === 'video'      → 视频分支:
              submissionMode === 'conversation' → 只允许 Seedance 2.0 Mini 且 11 ≤ duration ≤ 15,
                                                   确保处于"普通对话"(不在图像/视频技能里),返回 {mode:'conversation'};
                                                   不选模型、不设比例、不碰时长
              其它任何值(含 undefined)        → 进入"视频生成"技能:点模型菜单选 modelName → 比例/时长弹层;
                                                   比例或时长与请求不符 → {needsDurationInput:true},
                                                   相符 → {mode:'video', model, ratio, duration}
         else(image)                → 进入"图像生成"技能:选模型 → 选比例 → {mode:'image', model, ratio}
                                      (submissionMode 不参与,'conversation' 与其它值行为一致)
```

`run` 支持的 step:`prepare / fileInput / uploadReady / fill / checkSend / send / durationDragPlan / durationState / durationCommitPoint / verifyVideoSettings / videoSettingsDiagnostic / fileInputDiagnostic / feedbackDiagnostic / acceptFeedback / videoFollowup / videoFeedback / terminalFeedback / videoResult / requestVideoDownload`;导出的工具函数:`buildTextPrompt / conversationTab / modelAccessNotice / parseModel / parseVideoModel / press / normalize / selectReferenceFileInput / durationSliderScale / horizontalDurationPoints / feedbackKind / confirmationMatches / terminalFeedbackKind / collapseAttachmentImages`。

也就是说 Review Focus 1 的猜测("其它模型/时长必须走 work 模式")**不成立**:非 Mini / 非 11–15s 的视频走的是第三条"技能"分支(`submissionMode` 取 `'conversation'`/`'work'` 以外的任何值,下文记作 `skill`),analysis/04 只写了两种取值是漏记。`checkSend`/`send` 的核对逻辑同样按这三条分支:`work` 核对输入区与 prompt;`video+conversation` 核对"不在技能里 + Mini + 11–15s + prompt";其它核对"在对应技能里 + 模型 + 比例(+时长)+ prompt"。

### 页面入口:没有独立的"AI 创作"页,技能都在同一个 chat 页面的输入栏里

- 普通对话态(`chrome://doubao-chat/chat`)输入栏动作区:`upload_file_button`、`chat_input_action_mode`("对话"下拉)、`guidance-skill-bar` 里 `skill_bar_button_3` 图像生成 / `skill_bar_button_17` 视频生成 / `skill_bar_button_5000` PPT 生成 / `skill_bar_button_1005` 录音转写 / `skill_bar_button_5003` AI 表格 / `skill_bar_button_2` 帮我写作 / `skill_bar_button_more`、`chat_input_action_model`("豆包 快速")。另有 `[data-testid="conversation-mode-switch"]` 分段开关,两个 `<button>` "对话" / "工作"(父元素文本 `对话 工作`,正是脚本 `_0x40a354` 找的结构;`shenlangDoubaoGeneration.conversationTab()` 在普通对话态返回这个"对话"按钮,在技能态返回 `null`)。
- 脚本进入技能靠**按文本点 `图像生成` / `视频生成`**(`skill_bar_button_3/17`),不需要换 URL;进入后 `href` 仍是 `chrome://doubao-chat/chat`,输入栏被技能控件替换,并出现 `[data-testid="skill_input_exit_button"]`(退出技能的 ×,`innerText` 为技能名)和 `flow_chat_guidance_page`。
- 图像技能控件(均为 `button[aria-haspopup="menu"]`,无 testid):`模型 Seedream 5.0 Pro`、`比例 1:1`、`风格`、`模板`(`expand_btn_wrapper`)。视频技能控件:`模型 Seedance 2.5`、`16:9 · 10s`(比例·时长合一的摘要按钮,脚本正则 `^(自动|\d+:\d+)\s*[·•.]?\s*\d+\s*s$`)。
- 发送按钮两种技能态下都是 `button#flow-end-msg-send[data-testid="chat_input_send_button"]`(`aria-disabled="false"`),脚本首选 `#flow-end-msg-send` 命中;与 Task 1.6 一样**只有输入框有文字时才渲染**。
- 文件上传:`fileInput` 在两种技能态下都直接找到输入区里**已存在**的 `input[type="file"]`(`accept=".jpg,.png,.jpeg,.webp,.apng"`,`multiple=true`),给它打上 `data-shenlang-reference-upload="true"` 并返回 `{multiple:true}`,**不点任何按钮**,所以 `Page.setInterceptFileChooserDialog` 开着也**没有 `Page.fileChooserOpened` 事件**。Rust 侧(Task 3.2)走 analysis/04 的路径 2:`DOM.getDocument` → `DOM.querySelector('input[data-shenlang-reference-upload="true"]')` → `DOM.setFileInputFiles {nodeId, files}`;`fileChooserOpened` 分支只作兜底。视频技能的 accept 列表里**没有视频格式**,"Seedance 2.0 Fast 参考视频"在本版本 UI 上没有入口,`reference_files` 一律按图片处理。**(Task 3.2 复审更正:进入视频技能后 input 接受 mp4,参考视频能上传,见文末 Task 3.2 "参考视频与视频草稿残留"。)**

### 逐步结果

| 请求 | step | 结果 | 耗时 |
|---|---|---|---|
| image / Seedream 5.0 Pro / 1:1 / `conversation` / parallelTracking | `prepare` | `{mode:"image", model:"Seedream 5.0 Pro", ratio:"1:1"}`;点了"新对话"(href 从 `/chat/3844…` 变成 `/chat`)→ 点"图像生成" → 模型菜单选 5.0 Pro → 比例菜单选 1:1 | <1s |
| 同上 | `fill` / `checkSend` | `{filled:true}` / `{ready:true}`,发送按钮出现 | 32ms / 7ms |
| 同上,`prompt:''` | `fill` | **抛 `豆包输入框已变化，未覆盖已有内容。`**——计划里"`fill` 后用 `run('fill', {...req, prompt:''})` 清空"不可行(`fill` 先检查输入框必须为空)。清空用 Task 1.6 的办法:`ed.focus(); execCommand('selectAll'); execCommand('delete')` → `innerText` 回到 `"\n"`,发送按钮消失 | — |
| image / 5.0 Flash / 16:9;4.5 / 自动 | `prepare` | 都成功,模型/比例按请求切换 | 246ms / 119ms |
| image / 5.0 Pro / **21:9** | `prepare` | **抛 `未找到所选生成比例。`**(图像比例菜单没有 21:9,见下) | 6.2s |
| video / Seedance 2.0 Mini / 12s / 16:9 / `conversation`,页面在图像技能里 | `prepare` | **抛 `新对话尚未准备好，任务未发送`**(12s):脚本点"新对话"想回到普通对话,但技能态下"新对话"仍留在技能里 | 12.3s |
| 同上,先点 `skill_input_exit_button` 回到普通对话 | `prepare` | `{mode:"conversation"}` | 3ms |
| 同上,`prompt:"请使用 Seedance 2.0 Mini 生成12秒视频，比例16:9。\n测试"` | `fill` / `checkSend` | `{filled:true}` / `{ready:true}`(conversation 分支的 `checkSend` 要求输入框文本 == `req.prompt`,所以**前缀必须由 Rust 拼进 `prompt` 再传给 JS**) | 32ms / 7ms |
| video / Seedance 2.5 / 10s / `conversation` | `prepare` | **抛 `对话模式仅支持Mini的11～15秒任务`**(1ms,纯参数检查,不碰 DOM) | 1ms |
| video / Seedance 2.5 / 10s / `work`,普通对话态 | `prepare` | **抛 `未识别飞书工作模式输入区，任务未发送`**(12s):脚本等输入区出现 `本地电脑|云电脑` 文本 | 12.1s |
| video / Seedance 2.5 / 10s / 16:9 / **`skill`**,普通对话态 | `prepare` | `{needsDurationInput:true}`:进入"视频生成"技能 → 模型菜单选 2.5 → 读"比例 · 时长"摘要按钮,**首次进技能的这一瞬间摘要尚未渲染(读到 null)**,与请求不一致 → 点摘要选 16:9 → 展开时长滑块并**把弹层留在打开状态、滑块 focus** → 返回 needsDurationInput。脚本并非一律返回这个形状:还原后的视频分支只在 `summary?.ratio !== req.ratio \|\| summary?.duration !== req.duration` 时走这条;摘要已匹配时 focus 输入框后直接返回 `{mode:"video",model,ratio,duration}`,不开任何弹层(见下文四个模型那一行) | 647ms |
| 同上 | `verifyVideoSettings` | `{mode:"video", model:"Seedance 2.5", ratio:"16:9", duration:10}` | 12ms |
| 同上 | `durationDragPlan`(目标 10) | `{matched:true, duration:10}`(滑块未展开时直接读摘要) | 4ms |
| 同上 | `durationDragPlan`(目标 12 / 30) | `{startX:744.7, endX:769.0, y:767.4, matched:false, duration:10, endpointKey:null, fineTuneKey:"ArrowRight"}` / `{…endX:1013.4, endpointKey:"End"…}`(会点开摘要弹层并 focus 滑块,未拖动) | 107ms / 1ms |
| 同上 | `durationState` | `{duration:6}`——这里弹层被上一步 `prepare` 留在打开状态,滑块可见,**返回的是滑块 `aria-valuenow` 原值(刻度索引 6 = 10s),不是秒**。还原后的 `durationState` 有两条分支:① 没有可见滑块且摘要的 ratio/duration 已与请求一致 → 不开弹层,直接返回 `{duration:<摘要里的秒数>, fromSummary:true}`;② 否则(滑块已可见,或摘要不一致时先点开弹层等滑块出现)→ 返回滑块原始 `aria-valuenow`(索引)。`durationDragPlan` 内部用 `durationSliderScale` 加 offset 后才是秒,Task 3.2 判断"是否到位"只能看 `durationDragPlan.matched` / `verifyVideoSettings`,不要把 `duration` 字段当成统一单位 | 0ms |
| 同上 | `durationCommitPoint`(12) | `{x:744.4, y:767.2}` | 1ms |
| 同上 | `fill` / `checkSend` | `{filled:true}` / `{ready:true}` | 30ms / 3ms |
| video / 2.0、2.0 Fast、2.0 Mini、2.5 / 10s / 16:9 / `skill`(已在视频技能里,摘要 `16:9 · 10s` 与请求一致) | `prepare` | 四个都 `{mode:"video", model:<名>, ratio:"16:9", duration:10}`(**不是** needsDurationInput,没有打开比例/时长弹层);模型菜单是 **DOM `[role=menuitem]`**(不是原生菜单),菜单项文本 `Seedance 2.5 旗舰视频标杆 • 5 倍消耗` / `Seedance 2.0 进阶画面表现 • 2 倍消耗` / `Seedance 2.0 Fast 快速出片选择` / `Seedance 2.0 Mini 日常生成使用`,没有"升级"标记(体验套餐四个都能用);切换模型保留比例/时长 | 约 430ms 各 |
| image → video(`skill`)不退出技能 | `prepare` | **抛 `豆包控件未唯一匹配，已停止发送。`**(12s) | 12.0s |
| video → image 不退出技能 | `prepare` | **抛 `豆包控件未唯一匹配，已停止发送。`**(12s) | 12.1s |
| video 技能里 → Mini `conversation` | `prepare` | **抛 `新对话尚未准备好，任务未发送`**(12s) | 12.3s |
| 先点 `skill_input_exit_button` 再 `prepare`(任一方向) | `prepare` | 成功(图像 331ms / 视频 416ms) | — |

### 菜单与取值范围(真机 DOM)

- 图像模型菜单(`[role=menu]` + `[role=menuitem]`):`Seedream 5.0 Flash 专业快速`、`Seedream 5.0 Pro 专业出图・4 倍消耗`、`Seedream 4.5 日常生成`、`Seedream 4.0 基础生图`(4.0 不在白名单,脚本 `parseModel` 也不识别)。体验套餐可选 5.0 Pro,`modelAccessNotice` 未触发"豆包订阅/立即升级"弹窗。
- **图像比例**(弹层里 `role=menu` 下的 span,不是 menuitem):`自动 9:16 2:3 3:4 1:1 4:3 3:2 16:9`——**没有 21:9**。
- **视频比例**:`自动 3:4 4:3 9:16 16:9 1:1 21:9`——**没有 3:2 / 2:3**。
- 视频时长:弹层里 `[role=slider]`(`<span>`,18px 宽的圆点),`aria-valuemin=0`,`aria-valuemax` = Seedance 2.5 **26**(标签 `4s … 30s`),2.0 / 2.0 Fast / 2.0 Mini **11**(标签 `4s … 15s`);`aria-valuenow` 是索引,秒 = 索引 + 4(`durationSliderScale(0,26)→{min:4,max:30,offset:4}`,`(0,11)→{min:4,max:15,offset:4}`,刻度不在 `{6,11,26}`/`{10,15,30}` 内会抛 `未识别视频时长滑块刻度`)。默认(本机)`16:9 · 10s`,切模型不重置。
- 图像"风格"菜单 32 项(人像摄影 … 绘本),relay 不用。
- 视频技能的参考文件 input:`accept=".jpg,.png,.jpeg,.webp,.apng"`(同图像),multiple。

### 工作模式(`work`)实测

点 `conversation-mode-switch` 的"工作"后:输入栏变成 `本地电脑`(`chat_input_action_mode` 下拉项为 `对话 / 工作任务·本地电脑 / 工作任务·云电脑`)、`项目`、`按需确认`、`general-agent-skill-button`(技能)、`general-agent-plugin-button`(插件)、`自动 / 低`,没有图像/视频技能条,`/json/list` 不新增 target。此时 `prepare(work)` 立即返回 `{mode:"work"}`;`buildTextPrompt` 把参数拼进自然语言:`测试\n请生成视频。使用模型：Seedance 2.5；比例：16:9；时长：10秒。`。即"工作模式"= 豆包的通用 agent(操作本地电脑/云电脑),模型/比例/时长只是**文字提示**,由 agent 自行决定怎么做,没有 DOM 层的参数保证,`checkSend` 也只核对 prompt。点"对话"可切回普通对话态。**relay 不使用 `work` 模式**。

### 窗口隐藏时的生成链路

`osascript` 把豆包 `visible` 置 false(`document.visibilityState === 'hidden'`)后:页面 `setTimeout(100)` × 5 实测 505ms(**没有被节流**),`prepare`(含点新对话、进技能、选模型、选 3:2)344ms、`fill` 13ms、`checkSend` 4ms,全程 869ms,与前台一致。结论:提交链路(prepare → fill → checkSend)在隐藏态不需要 `Page.setWebLifecycleState active` / `Emulation.setFocusEmulationEnabled`;Task 2.3 的"focus 伪装"可省略(Chromium 对隐藏页的 intensive throttling 在隐藏 5 分钟后才启动,捕获阶段的长轮询若受影响留 Task 2.4 观察)。

### 结论表(供 Task 2.3 `SubmissionPlan::for_request`)

| type | model | submissionMode 传给 JS | prompt 前缀 | prepare 前置 | prepare 返回 |
|---|---|---|---|---|---|
| image | Seedream 5.0 Flash / 5.0 Pro / 4.5 | `conversation`(image 分支不看它;沿用沈浪默认值) | 无,原样 | 若有 `skill_input_exit_button` 先点它退出当前技能并等技能条出现;其余交给脚本(新对话 → 图像生成 → 模型 → 比例) | `{mode:"image",model,ratio}` |
| video | Seedance 2.5 / 2.0 / 2.0 Fast / 2.0 Mini(任意 4–30 / 4–15 秒) | **`skill`**(任何非 `conversation`/`work` 的值) | 无,原样 | 同上退出技能;脚本自己进视频生成 → 模型 → 读摘要 → (仅不一致时)比例 → 展开时长 | **两种形状,Task 2.3 都要接受**:(a) 摘要在那一瞬间与请求不一致(含首次进技能摘要还是 null)→ `{needsDurationInput:true}`,此时比例已选好、比例/时长弹层留在打开状态且滑块 focus → Task 3.2 `durationDragPlan` 循环 → `verifyVideoSettings` 得 `{mode:"video",model,ratio,duration}`;时长本来就匹配时 `durationDragPlan` 第一轮即 `matched:true`。(b) 摘要已匹配(豆包记住上次的比例/时长,真实使用中是常态)→ 直接 `{mode:"video",model,ratio,duration}`,没有弹层打开,**不要**再调 `set_duration` |
| video | Seedance 2.0 Mini,11–15 s | `conversation`(可选的另一条路) | `请使用 Seedance 2.0 Mini 生成{duration}秒视频，比例{ratio}。\n{prompt}`(沈浪 PS 规则,由 Rust 拼好传入) | 退出技能回到普通对话 | `{mode:"conversation"}`;不选模型,由豆包按文字理解,参数无 DOM 保证 |
| 任意 | 任意 | `work` | `buildTextPrompt` | 切到"工作" | 不采用 |

决定:**v0.1 图像用 image 分支、视频一律用 `skill` 分支**(同一套 DOM 控件、参数可核对、四个模型全部可用),`conversation` 视频路径不启用(只覆盖 Mini 11–15s 且无参数保证),`work` 不启用。`SubmissionPlan.submission_mode` 取值因此是 `"conversation"`(image)/ `"skill"`(video),计划里写的 `conversation|work` 需按此调整;prompt 不加任何前缀。

### 对后续 Task 的要求

1. **Task 2.3**:`submit` 在 `prepare` 之前加"退出残留技能"步:`[data-testid="skill_input_exit_button"]` 可见就 `press` 并等 `guidance-skill-bar` 出现(≤3s),否则跨技能(图像↔视频)和回普通对话都会 12s 后失败。`prepare` 超时 60s 足够(真机 <1s;失败分支固定 12s)。视频 `prepare` 返回 `{needsDurationInput:true}` 或 `{mode:"video",…}` 两种形状都是成功,只在前者时调 `duration::set_duration`(Phase 3 之前前者返回 Err),后者直接进 `fill`。异常映射按 11-error-codes 的语义拆开:`prepare`/`fill` 抛出的"任务未发送"类异常(`豆包控件未唯一匹配`、`未找到所选生成比例`、`新对话尚未准备好`、`豆包输入框已变化，未覆盖已有内容` 等)→ `generation_preparing_failed`;`checkSend`/`send` 抛出的异常(`豆包发送按钮尚未就绪，未发送。`、发送时的 `视频提示词已变化` / `豆包模式、模型、比例或提示词已变化` 等)→ `generation_submit_failed`;两者 `details.js_exception` 都带原文。
2. **Task 2.3**:清空输入框不能用 `fill('')`,用 `selectAll+delete`(与 `compat.rs` 相同)。
3. **Task 2.2 / 2.8**:比例白名单按类型拆分——图像 `自动 1:1 16:9 9:16 4:3 3:4 3:2 2:3`(无 21:9),视频 `自动 1:1 16:9 9:16 4:3 3:4 21:9`(无 3:2 / 2:3);spec 01 F3.1 的统一列表在真机上图像 21:9 / 视频 3:2、2:3 会在 `prepare` 阶段以 `未找到所选生成比例` 失败。建议 `validate_request` 直接拒绝(`request_invalid_enum`,`allowed` 按类型给),而不是入队后失败。
4. **Task 3.2**:滑块是 `[role=slider]` span,`aria-valuenow` 为索引;循环终止条件看 `durationDragPlan.matched`,最后 `verifyVideoSettings`。`durationState.duration` **单位不统一**:滑块不可见且摘要已匹配时返回摘要里的秒数并带 `fromSummary:true`,其余情况(滑块可见,或脚本刚点开弹层)返回滑块原始索引(秒 = 索引 + 4),不要把它当成统一的索引或秒来比较。目标时长等于当前值时 `durationDragPlan` 不展开弹层直接 `matched:true`。
5. **Task 3.2**:参考图走 `input[data-shenlang-reference-upload="true"]` + `DOM.setFileInputFiles`;`Page.fileChooserOpened` 只作兜底;~~视频无参考视频入口~~(复审更正:视频技能接受 mp4 参考视频,卡片为 `attachment-video-card`)。
6. 脚本把"新对话"按钮当作"回到普通对话"的手段,但技能态下"新对话"不退出技能——这是 2.32.5 的行为。沈浪脚本**并非没有**退出技能的逻辑:图像入口函数里有一段启发式,找可见的 `/^(关闭|移除|取消)(?:\s*(视频生成|PPT生成|…))?$/` 按钮,或文本等于技能名(`视频生成` 等)且含两个可见 svg 的元素并点第二个 svg;但它永远走不到——在它之前先执行的"进技能"步在另一个技能里会落到按文本找 `技能` 按钮的分支,12s 后以 `豆包控件未唯一匹配，已停止发送。` 抛出。所以退出残留技能必须由 Rust 侧在 `prepare` 前完成(要求 1);`exit` 按钮的 testid `skill_input_exit_button` 是本版本观察值,Task 2.3 应把它放进 `compat` 自检或失败时的 `failed_selectors`。

## Task 2.3 · 生成管线 `doubao::generation`(2026-10-16)

环境同上(豆包 2.32.5,端口 9707,仅 1 个有效账号)。本 Task **没有执行 `send`**(不消耗配额);`submit` 的 send → `Network.requestWillBeSent` → `submissionIdentity` 这段只有 mock 覆盖,真机留给 UC-10 验收(Task 2.7/2.9)一并验证。

### 退出残留技能(`scripts/cdp-probe.mjs`,只读 + 一次点击)

- Task 2.1 结束后页面仍停在"图像生成"技能:`[data-testid="skill_input_exit_button"]` 命中 **2 个**嵌套元素(外层 `div` 106×32 `cursor-pointer bg-g-exit-skill-btn-bg`,`data-value="3"`,`contenteditable="false"`,内含图标 `<img>` + 文本 `图像生成` + 关闭 svg;内层 `div` 70×22 同 testid),`querySelector` 取到外层即可。此时 `[data-testid="guidance-skill-bar"]`、`skill_bar_button_3/17`、`#flow-end-msg-send` 都为 0。
- 对外层元素派发 pointerdown/mousedown/pointerup/mouseup/click 五事件链后 **157ms** 技能条出现:`[data-testid="guidance-skill-bar"]`(`div`,709×32,文本 `图像生成\n视频生成\nPPT 生成\n录音转写…`)与 `skill_bar_button_3`(`button` 92×32)同时可见,退出按钮从 DOM 移除,`href` 不变。所以技能条 selector 是 **`[data-testid="guidance-skill-bar"]`**(testid,不是 class),`generation.rs` 的 `SKILL_BAR_SELECTOR` / `SKILL_EXIT_SELECTOR` 按此取值,等待上限 3s。

### `live_submit_dry_run`(Rust 路径,跑到 `checkSend` 为止,然后 selectAll+delete 清空)

`cargo test live_submit_dry_run -- --ignored --nocapture`,图像 / Seedream 5.0 Pro / 1:1:

| 起始状态 | 退出技能 → 注入 → prepare → fill → checkSend | 结果 |
|---|---|---|
| 普通对话态(刚退出技能) | **2958ms**(脚本点"新对话" → 进图像生成 → 选模型 → 选比例) | `Ok`,阶段 `准备豆包生成页面 / 模型与参数已就绪 / 提示词已填写`;清空后 `innerText="\n"`,页面留在图像技能(`exit=true`,与沈浪行为一致) |
| 残留在图像技能 | **194ms**(`exit_residual_skill` 点退出 → 技能条出现 → 脚本再进图像生成) | `Ok`,同上 |

即 Rust 侧的"退出残留技能"步修复了 Task 2.1 发现的跨技能 12s 失败;`prepare` 60s 超时裕量充足。

### `submissionIdentity` 契约(Node 里对混淆后的 `doubao-result-reader.js` 实测)

- 签名 `submissionIdentity(request, prompt, options = {})`,`request = {url, method, postData}`;URL 必须 `https:`、host 以 `doubao.com` / `snssdk.com` 结尾、`POST`、pathname 含 `completion|generate|/im/message/send` 且**不含 `history|recommend|config`**(注意:reader 的负向词表没有 `list`,capture 的 PS 规则多一个 `list`;Rust `is_generation_request` 按 PS 的四个词过滤)。
- `pathname === '/chat/completion'` 且 body 有 `messages[]` 时走专用分支:取恰好一条 `local_message_id` 匹配 `^[a-zA-Z0-9_-]{1,160}$` 且 `messageText()` 非空的消息,`options.verifiedSend` 只在这条分支上放宽 prompt 核对;`conversationId` 来自 `client_meta.conversation_id`。其它路径(含 `/samantha/chat/completion`)走通用递归扫描:收集 `local_message_id / message_id / conversation_id` 与 `text / content / prompt` 字段,prompt 必须被拼接文本包含,`local_message_id` 恰 1 个(或 0 个 + `message_id` 恰 1 个)、`conversation_id` ≤1 个,返回 `{localMessageId, messageId, conversationId}`;否则 `null`。
- **`localMessageId` 可以为 null**(只有 `message_id` 的请求),所以 Rust `SubmissionIdentity.local_message_id` 是 `Option<String>`(计划写的是 `String`),`parse_identity` 要求 `localMessageId` / `messageId` 至少一个。
- reader 对 `postData` 做 `JSON.parse`,裸的 18 位 `conversation_id` 会丢精度(实测 `123456789012345678` → `"123456789012345680"`);Rust 在拼 `identity_expression` 前先对 `postData` 跑 `json_fix::fix_big_integers`,把这些字段包成字符串后再交给页面。
- 真豆包 2.32.5 的发送请求实际打到哪个 endpoint、body 形状如何,本 Task 未观察(需要真实 `send`),留 UC-10 验收时用 `cdp_debug_log` 记录。

## Task 2.4 · 捕获管线 `doubao::capture` / `doubao::feedback`(2026-10-16)

环境同上(豆包 2.32.5,端口 9707,仅 1 个有效账号)。本 Task **没有执行 `send`**(不消耗配额);完整的"send → 捕获 → 原图"链路留 UC-10(Task 2.7/2.9)。

### 脚本契约(Node `vm` 里加载混淆脚本实测,DOM 用 stub)

`doubao-result-reader.js`(`shenlangDoubaoResultReader`,导出 `streamSubmissionBinding / nativeShadowSnapshot / followupSubmissionIds / submissionIdentity / collectSubmission / collectFeedback / messageText / fallbackAPI / matchingDownload / unwatermarkedURL / resolveVideos / contentBlocks / collectVideos / videoURL / videoURLs / read / analyze / collect / validateRequest / conversationId / messages`):

| 函数 | 实测 |
|---|---|
| `analyze(body, request, options)` | 返回 `{ok:true, nativeShadow:null, images:[], feedback:null, binding:null}`(无匹配时);`binding` 形状 `{conversationId, submissionId, createdAt}`(`collectSubmission` 源码),`feedback` 形状 **`{message, messageId, submissionId, createdAt, scope}`(`collectFeedback` 源码,没有 `stage`**;第二轮审查更正,原记录 `{stage, message}` 有误)——要拿到 stage 必须交给 `run('acceptFeedback', {reply: <该对象>})` 分类 |
| `read(request, options)` | 同上,内部 `fetch` 失败 → `{ok:false, reason:"read_failed", images:[]}`;HTTP 非 2xx → `{ok:false, status}`;>8MB → `reason:"response_too_large"` |
| `validateRequest(request)` | URL 必须 `https:` + host `www.doubao.com` + path `^/im/(chain/single\|message/chain_by_block)/?$` + `POST`;**postData 里 `conversation_id` 必须恰好一个**,否则抛;返回 `{url, conversationId}` |
| `messages(body)` | 读 `downlink_body.<chain>.messages` 或 `downlink_body.pull_message_chain_by_block_downlink_body.messages`(先 `fixBigIntegers` + `JSON.parse`),其它形状一律 `[]`——所以 Rust 传 body **原样**即可,不必也不应预处理 |
| `collect(body, {boundSubmissionId,…})` | **`boundSubmissionId` 为空时不按消息链过滤**,会收下 `since` 之后整个对话的图 → Rust 侧只在已绑定 submission 后采信 `images` |
| `analyze` 的 `requireBinding` | `options.boundConversationId !== conversationId` 直接返回空;`boundConversationId` 未知(新对话)时若仍传 `requireBinding:true` 永远绑不上 → Rust 只在已知对话时传 true,`requireIdentity:true` 不变(靠 `expectedLocalMessageId` 认领) |
| `streamSubmissionBinding(text, {expectedLocalMessageId, expectedMessageId})` | 两个 expected 都空 → `null`;`text` 可为 JSON 或 SSE(`data:` 行);用于从生成响应流里绑定新对话 |

`doubao-generation-page.js`(用 `Proxy` 记录 `run(step, req)` 访问的请求字段):

| step | 读取的字段 | 返回 |
|---|---|---|
| `terminalFeedback` | `type`, `prompt` | `{stage:"unknown"}`(无提交状态);有状态时 `{stage, message?}` |
| `videoFeedback` | `prompt`, `type` | `{stage, message, model}` |
| `videoFollowup` | `type`(+ `userConfirmed` / `confirmationMode` / `autoUseMembership`) | `{sent:false, stage:"unknown"}`;识别到 `checking`/`quota_checking` 且允许自动时自己 `fill('继续生成'或补发文案)` + 点 `#flow-end-msg-send` 返回 `{sent:true, stage}`;auto 但页面上没有可自动确认的提示 → `{sent:false, stage:"checking", manualRequired:true, message}`。**注意 `quota_checking` 分支只看 `userConfirmed \|\| autoUseMembership===true`**,`userConfirmed` 为 true 就发送 |
| `acceptFeedback` | `boundSubmissionId`, **`reply`**, `prompt`, `modelName`, `type` | **不是代回复,是历史反馈的分类钩子**(两轮审查纠正):PS(doubao-capture.ps1:341-345)在已绑定、reader 结果无图时把 reader 的 `feedback` 消息对象**原样**作为 `reply` 传入;`feedbackKind` 分类后返回 `{stage, message, createdAt, model, source:"current_conversation_message", messageId, submissionId}`——`该操作将消耗会员额度，是否继续生成？` → `quota_checking`、`今日生成额度已用完…` → `quota_exhausted`、"视频参数确认…确认后开始生成视频" → `checking`、拒绝类文案("抱歉…请重试")→ `needs_check`,其它 `unknown`;`reply` 为字符串一律 `unknown`;不 fill、不点发送。relay 用它做**分类**(`capture::accept_feedback_expression`),"继续生成"的发送仍由 `accept_reply_expression` 完成 |
| `feedbackDiagnostic` | — | `{reason:"submission_state_missing", submittedLength:0, baselineCount:null, matchingCount:0, anchorConnected:false, networkStage:null}`——feedback 依赖 `send` 时写进 trackingKey 全局对象的提交状态,**capture 不能重新注入覆盖它**,只能 `typeof <key> === 'undefined'` 守卫 |

### `live_capture_expressions_dry_run`(真机,只读)

`cargo test live_capture_expressions_dry_run -- --ignored --nocapture`,注入 reader 与一次性 `shenlangDoubaoTask_live<ms>` 全局对象(结束后 `delete`),各表达式一次求值,全程 160ms,无异常:

| 表达式 | 返回 |
|---|---|
| `streamSubmissionBinding("data: {}", identity)` | `null` |
| `read({url:/im/chain/single, postData:{conversation_id:"1"}}, options)` | `{"binding":null,"feedback":null,"images":[],"nativeShadow":null,"ok":true}`(页面真的发出了 fetch;对话不存在但接口 2xx,reader 解析出空消息链) |
| `analyze("{}", request, options)` | 同上 |
| `terminalFeedback`(带守卫注入) | `{"stage":"unknown"}`(没有 `missing`,说明注入成功) |
| `videoFollowup` | `{"sent":false,"stage":"unknown"}` |
| `acceptFeedback`(第二轮审查后补,reader 形状消息对象) | 会员额度文案 → `{"stage":"quota_checking","source":"current_conversation_message","messageId":"1002","submissionId":"1001",…}`;额度用完文案 → `stage:"quota_exhausted"` |
| 输入框脏检查 | `{"found":true,"text":""}` |
| `/json/list` chat page 数 | 1 |

### 审查后修正(Task 2.4,2026-10-16)

1. **图像代回复原先是空操作**。原实现调 `run('acceptFeedback', {...payload, reply:"继续生成"})`,在 Node vm 里实跑证实它只返回 `{stage:"unknown"}`、不发送(见上表更正);auto 模式和用户确认后都不会回复,任务会停在 `checking` 直到 5 分钟超时。现在 Rust 注入自己的表达式 `capture::accept_reply_expression`(复刻 `videoFollowup` 的发送路径):输入框为空 → `run('fill', {...payload, prompt:"继续生成"})`(沈浪自己的 fill + 核对)→ 等唯一且可用的发送按钮(`#flow-end-msg-send,[data-testid="chat_input_send_button"],.send-btn-wrapper button`,≤1.5s)→ `.click()` → 等输入框清空(≤2s)→ `{sent:true, cleared, waitedMs}`;输入框已是"继续生成"则只点发送;输入区有其它内容 → 抛 `豆包输入区已有内容，未发送继续生成`(沈浪同文案,走降级链);发送按钮没出现 / 不唯一 → `{sent:false, pending:true, reason}`;trackingKey 全局对象不存在 → `{sent:false, missing:true}`(下次轮询先重注入)。Rust 只在返回 `sent:true` 时记"代回复已发出";`sent!=true` 连续 3 次(`ACCEPT_MAX_ATTEMPTS`)→ 降级链 / 再问用户。沈浪脚本内部的 `_0xedde11()`(把反馈锚点重定到"继续生成"消息)不可从外部调用,图像的后续阶段主要靠 reader 的 `collectFeedback`(它按 `normalize(text)==='继续生成'` 认领代回复)与 `images`,`terminalFeedback` 本来也只识别终态拒绝。
   - 真机验证(只读):把生成的表达式用不存在的 trackingKey 在真豆包页面求值 → `{"sent":false,"missing":true}`,语法在 Chromium 147 下合法、不碰 DOM;Node vm + stub DOM 覆盖 8 条分支(key 缺失 / 无输入框 / 有草稿 / 空→fill+click / 已是继续生成→只 click / 无发送按钮 / 按钮不唯一 / fill 抛异常)。真实发送留 UC-10。
2. **用户确认改为一次性消费**。原来 `user_confirmed` 一旦为 true 永不回落:用户为 `checking` 确认一次后,`quota_checking` 会被图像的 Rust 分支和视频的 `videoFollowup`(`userConfirmed` 为 true 就发)自动代回复,未经同意消耗会员配额。现在:代回复真正发出(图像 `sent:true` / 视频 feedback 的 `sent`)后 `user_confirmed`/`needs_confirm_emitted` 回落,下一个确认 stage 重新按 `should_auto_confirm` 判断、再发一次 `NeedsUserConfirm`;`feedback_payload.userConfirmed` 随之回落,视频在 `quota_checking` 回到 `videoFeedback`(只识别)直到用户再次确认。`confirmed()` 改为边沿触发(只在 `watch` 收到新值且为 true 时返回;runner 每次 `POST /confirm` 都 `send(true)`,`watch::Sender::send` 不比较旧值仍会通知),降级把 `user_confirmed` 置回 false 后不会被旧电平立刻重新置 true。mock 覆盖:`image_confirmation_is_consumed_per_stage` / `video_confirmation_is_consumed_per_stage`(manual + `auto_use_membership=false`,checking 确认后出现 quota_checking → 必须再发 `NeedsUserConfirm`,且确认前 acceptReply / videoFollowup 不发送)。
3. **视频路径 1 原先是死代码**:`responseReceived` 只匹配 `generation_requests`(POST 生成请求),CDN 媒体是独立 GET,永远不在集合里。现在视频任务在 `requestWillBeSent` 把 `wallTime >= send_since` 的**所有** requestId 记进 `video_requests`(PS `videoRequests` 语义),`responseReceived` 按它匹配 `video/*` + `is_video_cdn_host`(mock:`video_cdn_response_after_send_is_a_candidate`)。
4. 其它:`videoFollowup` 返回 `manualRequired:true`(auto 但页面无可自动确认的提示)→ 降级 manual 并发 `NeedsUserConfirm`(`FeedbackState.manual_required`,false 时不序列化);`Runtime.evaluate` 超时(`CdpError::Timeout`)与 JS 异常一样走降级链;`/json/list` 里 0 个 chat page 不再报 `multiple_chat_pages`,而是与 `/json/list` 失败一样计数、连续 3 次 → `doubao_cdp_disconnected{reason:"chat_page_missing"}`;chat 重连时一并重取 background session(重新订阅 + `Network.enable`),只有 background 断开时在下一次 target 检查重新接上;`cancel_returns_within_50ms` 改为从取消时刻起算 <50ms。

### 第二轮审查后修正(Task 2.4,2026-10-17)

1. **历史反馈从未被分类**。reader `read`/`analyze` 的 `feedback` 是消息对象(没有 `stage`,见上表更正),原实现直接 `parse_feedback` 得 `unknown` 后丢弃;而 `terminalFeedback` 的 `terminalFeedbackKind` 对 `该操作将消耗会员额度，是否继续生成？` 只回 `unknown`(Node vm 实测),所以图像的 `quota_checking` / `NeedsUserConfirm` / 代回复 / 降级链真实场景下全部走不到,任务会等到 5 分钟超时。修复:`process_analysis` 在 feedback 无 `stage`、有 `messageId`、已绑定 submission、结果无图时,对 chat 页求值 `typeof <key>==='undefined' ? {stage:'unknown',missing:true} : Promise.resolve(<key>.run('acceptFeedback', {boundSubmissionId, reply:<feedback 对象>, prompt, modelName, type}))`(新 `Job::Classify`,每个 messageId 只分类一次;`missing` / CDP 断线 / 超时会重试),结果按 feedback 处理(source `history`)。分类出的确认 stage 记为 `history_confirm`:用户确认 / auto 允许时在 feedback 节拍上触发 `accept_reply_expression`(`terminalFeedback` 不会再报这个 stage),代回复 `sent:true` 后清除。只出现在历史里的 `quota_exhausted` 由此映射为 `account_quota_exhausted`。mock 全部改为 reader 真实形状 + `acceptFeedback` 分类 mock:`history_feedback_is_classified_via_accept_feedback_with_reader_object`(断言 `reply` 是对象、带 `boundSubmissionId/modelName`,同一消息只分类一次)、`quota_exhausted_in_history_fails_with_account_quota_exhausted`、`unbound_history_feedback_is_not_classified`、`manual_mode_emits_needs_user_confirm_then_replies_after_confirm` / `image_confirmation_is_consumed_per_stage`(改走历史分类)。真机只读:`live_capture_expressions_dry_run` 新增 `acceptFeedback` 两条文案,返回 `quota_checking` / `quota_exhausted`。`CaptureOptions` 因此多 `model_name: Option<String>`(runner 传 `request.model_name`)。
2. **capture 订阅晚于生成请求**。生成请求的 `requestWillBeSent` 在 `submit` 的 identity 窗口里就被消费,capture 订阅时已错过,该请求后续的 SSE(`streamSubmissionBinding` 绑定)与 `loadingFinished`(直接 `extract_image_originals`)全被丢弃。修复:`generation::ObservedRequest {target_id, request_id}`;`SubmissionReport.generation_requests` 收集窗口内见到的全部生成请求(含携带 identity 的那条),runner 原样放进 `CaptureOptions.generation_requests`,capture 以 `(target_id, request_id)` 预置 `generation_requests`(原来只按 requestId,现在区分 chat / background)。identity 返回到 capture 订阅之间的 SSE 片段仍可能漏,但 `loadingFinished` 后 `getResponseBody` 取的是完整响应体,可兜回。mock:`preloaded_generation_request_receives_later_sse_and_body`(不推 `requestWillBeSent`,只推之后的 SSE + `loadingFinished`,断言 `Bound` 与原图)与对照 `generation_request_not_preloaded_is_ignored`;`submit_runs_steps_in_order_and_returns_identity` 断言 `generation_requests`。
3. (minor)**多图只交付一部分**:原来第一次拿到原图立即返回。现在拿到原图后等 `CaptureTiming.result_settle`(默认 1.5s)无新 URL 才返回,从第一张起最多 `result_settle_max`(6s);静默窗口内撞上总 deadline 时按已捕获结果完成而不是超时;返回前若还没发过 `Stage(Result)` 补发一次(reader 有图时 feedback 为 null,不会自己报 `result`)。mock:`multi_image_results_wait_for_settle_window`。

### 对后续 Task 的要求

1. **Task 2.7(runner)**:`capture` 的 `account_quota_exhausted` details 只有 `{reason, doubao_message}`,写入 `task.error` 时补 `account`;`CaptureEvent::NeedsUserConfirm` → `waiting_user_confirm`,**每次**用户 `confirm` 都对 `watch::Sender<bool>` 调一次 `send(true)`(capture 边沿触发:一次确认只消费一次,`checking` 与 `quota_checking` 各自需要一次确认,同一 stage 的重复确认无副作用);`Bound` → 写 `conversation_id/submission_id`;`Stage` → `stage/stage_message`;`Originals` 之后 `capture` 再等静默窗口(`result_settle` 1.5s,上限 6s)才返回,下载走 Task 2.5。
2. **UC-10**:首次真实 `send` 时打开 `cdp_debug_log`,记录生成请求 endpoint、历史请求(`/im/chain/single` 的 postData 形状、`headers` 里哪些 `x-*` 被 reader 透传)、`binding.createdAt` 单位、`terminalFeedback` 在 `checking`/`result` 时的真实返回,以及图像代回复 `accept_reply_expression` 的真实返回(`cleared` 是否为 true、发送按钮在确认提示出现时是 `#flow-end-msg-send` 还是 `chat_input_send_button`),补到本节。
3. **Task 3.3**:`VideoCandidates` / `Fallbacks` 事件已就位:`responseReceived` 在 `task_type == Video` 时按 `video_requests`(`send_since` 之后的所有请求,含 CDN 媒体 GET)匹配 `video/*` + 视频 CDN 白名单;视频 `read` 返回的 `videos` / `fallbacks` 字段名按 PS(`$value.fallbacks[].vid/api`)取,真机未验证。

## Task 2.7 · 调度器 `tasks::scheduler` / 运行器 `tasks::runner`(2026-10-16)

本 Task **没有碰真豆包**:计划的 Task 2.7 没有 live 步骤,真实 `send → capture → 下载` 的整链路属于 UC-10(Task 2.9 Step 3,会消耗一次配额),本机也只有 1 个有效账号。全部行为由 mock `GenerationBackend`(`scheduler.rs` 测试模块)覆盖:成功路径 / capture 失败 / submit 失败 / preparing 超时 / generating 超时 / 确认超时 / 等待确认不消耗生成预算 / pending 取消 / 运行中取消 / 同账号串行 / 跨账号并发 ≤ global_max / 队列满 51 / manual 两次确认 / 排队超时扫描 / `fail_all_running` / 提交校验 / 豆包未连接 503。

### 与前面真机事实的衔接(实现时据此取舍)

| 真机事实(来源) | 对 2.7 的影响 |
|---|---|
| 切换账号会整页 reload,所有 CdpSession 作废(Task 1.2 / 1.8) | runner 的 `ensure_account` 先看 `accounts(false)` 缓存的 `current`,一致就**不碰菜单**(真机读列表 ~3.4s);不一致才 `switch_account`。切换与 `submit` 都算在 preparing 60s 内(切换验证上限 20s)。 |
| 提交状态写在页面 trackingKey 全局对象里,reload 后丢失(Task 2.4) | **跨账号并发的隐患**:若另一账号的任务此时切换账号,正在 capture 的任务会丢提交状态并对错误账号重放历史——这是从上述事实推断的,未真机验证(只有 1 个账号);计划"已知问题"(Task 2.7)⑦ 已交给 Task 4.2 处理,本 Task 按 spec 08 允许跨账号并发。 |
| `capture` 的 deadline 在等待用户确认期间照样走(Task 2.4 的 mock 用例 `auto_mode_replies_once_and_degrades_to_manual_on_exception` 依赖它) | 没有改 `capture.rs`;runner 用 `PhaseClock` 自己计"生成预算(不含等待)"与"确认等待 1h",传给 `capture` 的 deadline 是宽松上限(预算 + 2h)。 |
| `account_quota_exhausted` details 只有 `{reason, doubao_message}`(Task 2.4) | runner 写 `task.error` 时补 `account`(`task_error_for`)。 |
| 用户确认一次性消费、`confirmed()` 边沿触发(Task 2.4 审查修复) | `Scheduler::confirm` 每次都 `confirm_tx.send(true)`;runner 另订阅一份 `watch` 边沿把 `waiting → generating` 独立落库;`checking` 与 `quota_checking` 各需一次确认(mock 用例 `manual_mode_waits_for_confirm_then_continues` 走完两次)。 |
| `download_image` 写 `image_{index+1}.{ext}`,扩展名按内容定(Task 2.5) | `ResultFile.filename` 取实际落盘文件名,`relative_path = results/{task_id}/{filename}`。 |

### 留给 UC-10(Task 2.9)在真机上核对的点

1. `DoubaoBackend::submit`(`chat_session` + `background_session` → `generation::submit`)与 `capture` 串起来后,`SubmissionReport.identity` 是否总是 `Some`(runner 把 `None` 映射为 `generation_submit_identity_missing`)。
2. 真实 feedback 节奏下 `Stage` 事件与 `NeedsUserConfirm` 的先后、`stage_history` 是否出现 `sent → checking → rendering → result → downloading → completed` 这条主线。
3. 下载阶段 `generation_download_failed` 时任务是否正确落 `failed`(mock 下载永远成功)。

## Task 2.8 · HTTP API 服务器 `http`(2026-10-16)

环境同上(豆包 2.32.5,端口 9707,仅 1 个有效账号)。本 Task **只做只读操作**:没有 `send`、没有切换账号、没有 pkill 豆包;`POST /generate` 的成功路径只在 mock(`FakeBackend`,调度循环不启动或 submit 挂起)下验证,真实"提交 → 生图 → 下载"仍留 UC-10(Task 2.9)。

### `live_http_status_and_accounts`(`cargo test live_http_status_and_accounts -- --ignored --nocapture`)

真 `DoubaoConnection(9707)` + 临时目录 `Storage` + `DoubaoBackend` 组成 `AppState`,经 `http::router` 的 `oneshot`:

| 请求 | 返回 | 耗时 |
|---|---|---|
| `connect(false)` | `AlreadyConnected` | — |
| `GET /api/v1/doubao/status` | `{running:true, cdp_connected:true, cdp_port:9707, doubao_version_detected:"2.32.5", ui_compat_check:"ok", ui_compat_details:{…}, chat_targets:1, reconnecting:false, last_check_at}`(snake_case,比 spec 04 多 `reconnecting` / `ui_compat_details`) | 0ms |
| `GET /api/v1/accounts` | `{current:"她是翻哥", accounts:[{name:"她是翻哥", subscription:"体验套餐", label:"她是翻哥\n体验套餐", is_current:true, is_ambiguous:false}], read_at}` | **3580ms**(读账号菜单,与 Task 1.8 的 ~3.4s 一致;2s 内再读走缓存) |
| `GET /api/v1/health` | `{status:"ok", version:"0.1.0", uptime_seconds}` | 0ms |

注意 `label` 里是真机按钮的两行文本(`\n`),与 IPC 一致,HTTP 调用方按 `name` 用即可。

### `cargo tauri dev` 冒烟(真 relay 进程,生产数据目录)

启动时间线(`~/Library/Logs/doubao-relay/doubao-relay.2026-10-16.log`):`doubao-relay 启动` → +15ms `数据目录已就绪 data_dir=~/Library/Application Support/doubao-relay`(SQLite 迁移)→ +550ms `HTTP API 已监听 127.0.0.1:8787` → +560ms 首轮 `定时清理完成`(0/0/0)→ +1.15s `豆包已连接 action=AlreadyConnected`。即 **HTTP 先于豆包连接可用**,这段时间 `/accounts` 等会返回 503 `doubao_*`,调用方应以 `/health` 探活、以 `/doubao/status.cdp_connected` 判断可用。启动恢复 `fail_non_terminal("relay_restart")` 命中 0 条(不打日志)。

`curl` 逐项(`TOKEN` 取自 config.toml):

| 请求 | 状态 | 体 |
|---|---|---|
| `GET /accounts` 无 header | 401 | `auth_missing_token` |
| `GET /accounts` `Bearer nope` | 401 | `auth_invalid_token` |
| `GET /doubao/status` | 200 | 同上表,`last_check_at` 为心跳时间 |
| `GET /accounts` | 200 | 同上表(真机 ~3.5s) |
| `POST /generate` `{"type":"image"}` | 400 | `request_missing_field{field:"account"}` |
| `GET /tasks?limit=5` | 200 | `{tasks:[], next_cursor:null}` |
| `POST /uploads` `-F file=@1x1.png`(68 字节) | 200 | `{upload_id:<ULID>, mime:"image/png", bytes:68, expires_at:+1h}`;落盘 `uploads/<ULID>.png` |
| `GET /uploads/{id}` | 200 | 同上 |
| `GET /tasks/nope` | 404 | `task_not_found{task_id:"nope"}` |
| `GET /nope` | 404 | `request_malformed`("接口不存在") |

每个请求在日志里一条 `HTTP 请求 method= path= status= duration_ms=`(不含 header / body)。验收后只 kill 了 relay 与 vite(`pkill -f target/debug/doubao-relay`),豆包进程未动;那条 68 字节的测试上传留在生产数据目录,1h 后由 GC 清理。

### 对后续 Task 的要求

1. **Task 2.9**:`AppState.api_listening`(`AtomicBool`)已由 `http::serve` 维护,`app.status`/状态栏 "● API" 直接读它;`task_detail_json(&Task)`(`http/routes/tasks.rs`)是 HTTP 详情体,IPC `task_get` 可 `to_camel` 复用,保证两端字段一致;`FakeBackend` / `test_state_with` 在 `commands::testing`,命令测试可直接用。
2. **UC-10**:用 `scripts/uc-10.sh` 走 `POST /uploads`(可选)→ `POST /generate` → 轮询 `GET /tasks/{id}` → `GET /tasks/{id}/download/0`;`download_url` 已在详情体 `result.files[].download_url` 给出,下载响应 `Content-Type` 取库里 mime、`Content-Disposition` 文件名 `task-{id}-{index}.{ext}`。
3. `POST /accounts/switch` **与 Tauri `account_switch`** 在任何任务运行中都 409 / `account_busy`(守卫在 `AppState::switch_account` → `Scheduler::ensure_switchable`,见计划"已知问题" Task 2.8 ④ 与"Task 2.8 审查修复");Task 4.2 做跨账号串行化时可把它放宽为"只看当前账号 + 目标账号",并顺带用调度器锁盖住"检查 → 切换"之间的派发窗口。

### 审查后修正(2026-10-16,无真机动作)

审查用真 relay(生产数据目录)探到的三处与 spec 不符已修,均只有 mock / 单测覆盖、未再起真 relay:`Bearer  <tok>`(双空格)与 `Bearer <tok> `(尾随空格)之前 200,现 401 `auth_invalid_token`;`POST /api/v1/health` 之前裸 405 空体,现 405 + `{error:{code:"request_malformed"}}`;50MB+1 文件的 `upload_too_large.bytes` 之前是整个 multipart body 的 `Content-Length`(52429004),现是文件字节数(52428801)。

## Task 2.9 · 任务命令 + 前端任务页 / 详情页 / 新建弹窗 + UC-10 ~ UC-18(2026-10-17)

环境同上(豆包 2.32.5,端口 9707,仅 1 个有效账号 `她是翻哥 / 体验套餐`)。`cargo tauri dev`(真 relay,生产数据目录)后台运行,用 `scripts/uc-10.sh` 走 HTTP;**只真实生成了 1 张图**(计划允许的 UC-10),没有切换账号、没有 pkill 豆包。

### `scripts/uc-10.sh`(不带参数只跑不耗配额的检查;`--generate` 才真实生成)

| UC | 结果 |
|---|---|
| UC-17 鉴权 | 无 token 401、错 token 401 — 通过 |
| UC-11 账号不存在 | 404 `account_not_found` — 通过 |
| UC-13 参数校验 | prompt 空 `request_missing_field`、模型 / 比例非白名单 `request_invalid_enum`、**图像 `21:9` 也是 `request_invalid_enum`**(比例白名单已按类型拆分)、10001 字 `request_prompt_too_long`、11 张参考图 `request_too_many_references`,全部 400 且任务数不变 — 通过 |
| UC-10 单次生图 | `Seedream 5.0 Pro` / `16:9` / 19 字 prompt:提交即返回 `task_id`,**49s** `completed`(preparing 622ms — 豆包停在图像技能态、记住了上次设置;sent → 原图 43s;下载 3s)— 通过 |
| UC-14 任务查询 | 轮询看到 `preparing/preparing → generating/sent → generating/downloading → completed`;`stage_history` = `pending → preparing → sent → result → downloading → completed`(6 条)— 通过 |
| UC-15 pending 取消 | 第一个任务运行中再提交同账号任务 → `queue_position=1`,立即 `DELETE` → 200 `cancelled_from_status=pending`,状态 `cancelled`,从未派发 — 通过 |
| UC-16 排队 | 只验证了"同账号第二个任务排队(per_account_max=1)";5 个任务串行跑完会耗 5 次配额,未做(mock 覆盖)— 部分 |
| UC-18 结果下载 | `download/0` 200,`Content-Type: image/png` = `result.files[0].mime`,`Content-Length` 5267759 = 实际字节 = `bytes`,`Content-Disposition: attachment; filename="task-<id>-0.png"`;Range 206;index 越界 404 `result_index_out_of_range` — 通过 |
| UC-12 自动切换账号 | 本机只有 1 个账号,SKIP |

### 真机事实(UC-10 首次真实 send,补 Task 2.4 / 2.7 留下的核对点)

- `SubmissionReport.identity` 为 `Some`:`local_message_id` 是 UUID 形式(`ae10a060-c983-11f1-9fa5-f5f463feda03`),`message_id` 为空;capture 开始后 **647ms** 绑定 `conversation_id=38447732767746818`(17 位)/ `submission_id=58769960792465410`,大整数以字符串保存,精度无损。
- 原图来源 `source="generation_response"`(生成请求的 SSE / 响应直接给出 `image_ori_raw`,不需要历史重放);静默窗口 1.5s 后返回,`originals=1`。
- **这次没有出现 `checking` / `quota_checking` / `rendering` 阶段**:体验套餐 + Seedream 5.0 Pro 单图,豆包直接出图、没有确认提示,所以 `stage_history` 主线是 `sent → result → downloading → completed`,不是 Task 2.7 推测的 `sent → checking → rendering → result → …`;代回复 `accept_reply_expression` 因此仍未在真机触发过(留 UC-22)。
- 结果:16:9 实际 **2848×1600** PNG(5.0MB),`watermark_source=conversation_original`,落盘 `~/Library/Application Support/doubao-relay/results/<id>/image_1.png`。
- `generation_download_failed` 路径真机未触发(下载一次成功)。

### UI 验收(`scripts/ui-harness-tasks.playwright.js`,截图 `docs/analysis/screenshots/phase2/`)

真 relay 窗口这次 AX 树取不到窗口(`count of windows = 0`,Window 菜单里有 `doubao-relay` 但点了也不出现,疑似在别的 Space),改用 Task 1.10 的办法:Playwright 打开 vite `localhost:1420` 注入假 `__TAURI_INTERNALS__`。**数据是真的**:UC-10 那个任务的 HTTP 详情体 → camelCase + `result.files[].localPath`(与 IPC `task_get` 同形,写到 gitignore 的 `target/ui-fixture-tasks.js`),`convertFileSrc` mock 成 `http://asset.localhost/<path>` 并由 `page.route` 用真实结果文件应答。

- `tasks-list.png`:真实的完成 / 已取消两行 + 演示行(待确认 / 视频生成中 / 失败 / 排队),状态图标 ✓绿 ⏳蓝 ⚠红 ⏸灰 🚫灰,hover `title` 为 prompt 前 60 字(真实行 = `一只蓝色小鸟站在开花的树枝上,水彩风格`);注入 `task_updated` 后"排队"行原地变"⏳生成中"。
- `task-detail-completed.png`:真实缩略图加载成功(naturalWidth/Height = 2848×1600),时间轴 6 行;`task-detail-raw-request.png` 原始请求折叠展开。
- `task-detail-waiting-confirm.png`:`waiting_user_confirm` 时显示"手动确认"+"取消"。
- `new-task-dialog*.png`:默认当前账号 / 图像 / Seedream 5.0 Pro / 1:1;切视频后比例菜单 = `自动 1:1 16:9 9:16 4:3 3:4 21:9`、时长提示 `秒(4–30)`;空 prompt 提交 → `request_missing_field` 内联红字。
- 真实 WKWebView 里 `convertFileSrc`(`asset://localhost/…`)能否加载缩略图**未能直接截图证实**(无屏幕录制、AX 取不到窗口);`assetProtocol.scope` 用的是 `$DATA/doubao-relay/results/**`(Tauri 的 `$APPDATA` = `$DATA/<identifier>`,与 relay 数据目录 `dirs::data_dir()/doubao-relay` 不同,见计划"已知问题" Task 2.9),`cargo tauri dev` 编译与启动正常(`protocol-asset` feature 与配置一致,否则 tauri-build 会报错)。

### 对后续 Task 的要求

1. **Task 5.1**:状态栏"● API"仍固定"未启动"(`AppState.api_listening` 已有,但 spec 05 没有读取它的命令);设置页 / 状态栏实现时加命令或事件暴露它。
2. **UC-22**:真实的 `checking` / `quota_checking` 代回复至今没有触发过,手动确认路径(`task_confirm`)真机未验证。

## Task 6.2 · 打包 / 签名脚本本机验证(2026-10-17)

环境同上(macOS arm64,Rust 1.94,tauri-cli 2.12.1,豆包 2.32.5 @9707,仅 1 个有效账号)。本 Task **没有任何豆包发送 / 切换 / 重启动作**,只做打包与启动冒烟。

### `cargo tauri build --debug`

- 冷构建(独立 worktree target 目录)Rust 编译约 41s + 前端 `pnpm build`,产出 2 个 bundle:`target/debug/bundle/macos/doubao-relay.app`(67.5 MiB,debug 未优化)与 `target/debug/bundle/dmg/doubao-relay_0.1.0_aarch64.dmg`(18.7 MiB)。
- **DMG 文件名按架构命名**:本机构建是 `_aarch64.dmg`,spec 13 写的 `_x64.dmg` 只是示例;`_universal.dmg` 只在 `--target universal-apple-darwin`(release.yml)时出现。
- `--debug` 构建也启用了 `custom-protocol`(前端打进包里,不连 vite :1420),DMG 卷里是 `doubao-relay.app` + `Applications` 链接 + 卷图标。
- 包内可执行文件只有链接器 ad-hoc 签名(`flags=0x20002(adhoc,linker-signed)`)、无 hardened runtime——未签名发布包在别的机器上首次打开会被 Gatekeeper 拦(README 已写"右键 → 打开")。

### 安装后启动(从 DMG 拷出 .app 到临时目录,`open -n`)

| 检查 | 结果 |
|---|---|
| `GET /api/v1/health` | `open` 返回后 1s 内 200 `{status:"ok",version:"0.1.0"}` |
| 无 token `GET /accounts` | 401 |
| `GET /doubao/status` | `running/cdp_connected=true, doubao_version_detected:"2.32.5", ui_compat_check:"ok"`(AlreadyConnected,未重启豆包) |
| 窗口(System Events AX 树递归读文本) | 侧栏 `账号/任务/日志/设置`,状态栏 `豆包 已连接`、`API 未启动`,账号表 `她是翻哥 体验套餐 当前`,`共 1 个已登录账号` |
| 退出 relay(`kill <pid>`) | 8787 端口释放,豆包进程仍在(spec"关闭 relay 不杀豆包") |

状态栏 `API 未启动` 是已知现象:`state.apiListening` 在 Task 2.9 接 `app_status.api_listening` 之前固定为 false(见计划"已知问题" Task 1.10 ③),HTTP 实际已在监听。

### `scripts/sign-macos.sh` 真跑(ad-hoc 身份 `-`)

对上面 bundle 的拷贝执行 `APPLE_SIGNING_IDENTITY=- scripts/sign-macos.sh <bundle>`:`codesign --force --deep --timestamp --options runtime` 成功(`--timestamp` 对 ad-hoc 身份不报错)、`codesign --verify --deep --strict` 通过;暂存目录(`ditto` .app + `/Applications` 链接)→ `hdiutil create -format UDZO` 重建 DMG 并 `codesign` 通过。**hardened runtime(`flags=0x10002(adhoc,runtime)`)下不加任何 entitlements,app 照常启动**:HTTP 200、CDP 连上豆包、窗口渲染正常——WKWebView 不需要 JIT entitlement,监听 127.0.0.1 / 连 9707 也不需要(非沙盒)。真 Developer ID 签名与 `notarytool` 公证本机没有证书,未验证,只有 `DRY_RUN` 测试(`scripts/tests/sign-scripts.test.sh`)。

## Task 3.2 · 视频时长 `doubao::duration` + 参考图上传 `doubao::upload`(2026-10-17)

环境同上(豆包 2.32.5 @9707,窗口隐藏 `visibilityState=hidden`,仅 1 个有效账号)。全部是**干跑**:`prepare_and_fill` 跑到 `checkSend` 为止,然后 selectAll+delete 清空输入框,**没有执行 `send`**,不消耗配额。命令:`DOUBAO_LIVE_DURATION=<秒> DOUBAO_LIVE_MODEL=<模型> cargo test live_video_duration_dry_run -- --ignored --nocapture`、`DOUBAO_LIVE_REF=<a.png,b.png> [DOUBAO_LIVE_TYPE=video] cargo test live_reference_upload_dry_run -- --ignored --nocapture`(参考图是 scratchpad 里生成的 512×512 PNG)。

### 时长滑块(`set_duration`)

| 起始(豆包记住的值) | 请求 | prepare | 调整过程(`durationDragPlan.duration` 为**秒**) | 耗时 |
|---|---|---|---|---|
| 10s | Mini 7s | needsDurationInput | 拖拽 → 匹配 | 5.5s 全链路(含首次进技能) |
| 7s | Mini 13s | needsDurationInput | `Drag{784→958}` 1 次 → 13 | 675ms |
| 13s | Mini 14s | needsDurationInput | `ArrowRight` 1 次 → 14 | 206ms |
| 14s | 2.5 30s | needsDurationInput | 第 0 轮 `endpointKey:"End"` → 30 | 287ms |
| 30s | 2.5 4s | needsDurationInput | 第 0 轮 `endpointKey:"Home"` → 4 | 286ms |
| 4s | Fast 5s | needsDurationInput | `ArrowRight` → 5 | 209ms |
| 2.5 @30s | Mini 12s | needsDurationInput | 切到 Mini 后滑块被重置到 **10s**(不是夹到 15s),`ArrowRight` ×2 → 12 | 407ms |
| Mini @12s | Mini 12s | **直接形状** `{mode:"video",model,ratio:"16:9",duration:12}`,不调 `set_duration` | — | 420ms |
| Mini @12s | Fast 5s | needsDurationInput | `Drag{933→730}` 1 次 → 5 | 656ms |

- **窗口隐藏时 `Input.dispatchMouseEvent` / `Input.dispatchKeyEvent` 照样生效**(拖拽一次即准,键盘从未被判无效)。`durationDragPlan` 每次都会 `focus()` 滑块,只有滑块拿到焦点才给 `endpointKey` / `fineTuneKey`,所以按键总落在滑块上。
- 脚本只在"目标是端点"时给 `endpointKey`(`End`=最大、`Home`=4s),相差较大时只给拖拽坐标(`startX` = 当前圆点中心、`endX` = 目标刻度位置、`y` = 滑块中线)。
- `set_duration` 结束后时长弹层仍开着(`[role=slider]` 仍在 DOM),`fill` / `checkSend` 照常通过,无需手动关弹层。
- 结束时豆包停在"视频生成 / Seedance 2.0 Fast / 16:9 / 5s",Task 3.3 的 UC-20 用这一组参数时 `prepare` 直接返回匹配形状。

### 参考图(`upload_references`)

- 图像技能 2 张:`fileInput` → `DOM.getDocument` + `DOM.querySelector('input[data-shenlang-reference-upload="true"]')`(两次 nodeId 不同:54 / 109,每张图后 input 被重新渲染)→ `DOM.setFileInputFiles{nodeId,files:[path]}` → `uploadReady`,**每张约 0.9s**,2 张共 1.9s;`Page.setInterceptFileChooserDialog` 开着也**没有** `Page.fileChooserOpened`(与 Task 2.1 一致,主路径是 DOM 标记 input)。上传后输入区出现 `[data-testid="attachment_area"]` 与每张一个 `[data-testid="attachment-image-card"]`(`role=button`,`aria-label` = 文件名),`fill` / `checkSend`(请求带 `uploadedCount=2`)通过。
- **视频技能也能上传参考图**(Seedance 2.0 Fast / 16:9 / 5s + 1 张,1.3s,`checkSend` 通过)。Task 2.1 记的"视频技能没有参考**视频**入口"只针对视频文件;Task 2.9 新建弹窗"参考图只对图像开放"是误读,前端可以对视频也开放参考图(本 Task 未改前端)。

### 真机新发现:残留参考图属于技能草稿,会卡住下一次 `prepare`

- 参考图卡片随**当前技能的草稿**保存:退出技能(`skill_input_exit_button`)后卡片仍显示在普通对话输入区;此时图像 `prepare` 在 12s 后抛 `新图片对话未准备好，任务未发送`。
- 在普通对话态点卡片里的删除按钮(`[data-testid="attachment-delete-btn"]`,`svg`,只在悬停时有尺寸,派发 pointerdown/mousedown/pointerup/mouseup/click 即可)卡片会消失,但**再进图像技能又出现**,`prepare` 立即(142ms)抛 `豆包输入框已有参考图片，请先移除后重试。`;在技能里删除才真正生效(退出再进入技能后仍为 0)。
- 图像技能的草稿不影响视频技能:普通对话态删掉卡片后 `prepare(video)` 正常,而随后再进图像技能卡片又出现(反方向未验证)。
- relay 处理(`generation.rs`):① `prepare` 前 `upload::remove_attachments` 删掉当前可见的卡片;② `prepare` 抛含 `已有参考图` 的异常 → 在技能里再删一次并重试 `prepare` 一次(真机:`removed=1` → 重试 240ms 成功);③ 上传开始后任何失败(上传 / `fill` / `checkSend` / `send` 失败、`send` 前取消)都删掉本次上传的卡片再返回错误。成功 `send` 后卡片随消息发出,不需要清理(UC-20/UC-11 真机时核对)。

### 复审修复:参考视频与视频草稿残留(2026-10-17)

- **视频技能接受 mp4 参考视频**:Seedance 2.0 Fast / 16:9 / 5s + 1 段 2s mp4(11KB),`fileInput` → `DOM.querySelector` → `DOM.setFileInputFiles` → `uploadReady` 约 1.3s,`fill` / `checkSend` 通过;卡片是 `[data-testid="attachment-video-card"]`(在 `video-attachment-scroll-container` 里),删除按钮同为 `[data-testid="attachment-delete-btn"]`。mp4 + png 混传(2 个)2.7s 通过。上面"只缺参考视频"的说法作废。
- **视频技能的残留异常文案不同**:视频草稿里有残留卡片时,视频 `prepare` 立即抛 `视频输入区已有内容，任务未发送`(不是图像技能的 `已有参考图`)。普通对话态看不到视频草稿里的卡片,`prepare` 前的删除帮不上。
- relay 修复:① `remove_attachments` 的卡片选择器覆盖 `attachment-image-card` 与 `attachment-video-card`;② `prepare` 异常含 `已有参考图` **或** `输入区已有内容` → 在技能里删卡片 + 清空输入框后重试一次;③ 失败后的清理(删卡片、关文件选择框拦截、清空输入框)每步最多等 2s,能放进 runner 取消后的 5s 宽限期(超时只是不再等,页面里的删除脚本照样跑完)。
- 真机复现(干跑,未发送):`DOUBAO_LIVE_KEEP_REFS=1 DOUBAO_LIVE_TYPE=video DOUBAO_LIVE_REF=ref.mp4 live_reference_upload_dry_run` 在视频草稿留下 1 张 `attachment-video-card` → `live_video_duration_dry_run`(Fast 5s):`prepare` 抛 `视频输入区已有内容，任务未发送` → `removed=1` → 重试成功,总 877ms。`live_dry_run` 默认在技能里删掉本次卡片(mp4 + png:`removed=2`),结束时视频与图像草稿都为空。

## Task 3.3 · 视频捕获三条路径 + UC-20 真机生成(2026-10-17)

环境同上(豆包 2.32.5 @9707,窗口隐藏,仅 1 个有效账号 `她是翻哥 / 体验套餐`)。debug relay(`target/debug/doubao-relay`,自限时 25 分钟)走 HTTP,**真实生成了 1 段视频**(UC-20,计划允许的唯一一次):`POST /generate {type:video, model_name:"Seedance 2.0 Fast", ratio:"16:9", duration:5, confirmation_mode:auto, auto_use_membership:false}`,prompt 18 字(橘猫窗台伸懒腰)。另外只做了只读探测(reader `read` / `acceptFeedback` 分类 / `feedbackDiagnostic` / Network 旁听,一次 fallback API GET)。

### 时间线(任务 `01M52YMRBJD9X9T3T8ZA5EKDDP`)

| 时刻 | 事件 |
|---|---|
| +0.0s | 入队 → preparing;退出残留技能 66ms,`prepare` 422ms **直接形状**(豆包记住 Fast/16:9/5s,不调 `set_duration`) |
| +0.7s | send → identity(`local_message_id` UUID 形,147ms)→ generating/sent |
| +1.3s | capture 绑定 `conversation_id=38447520992069890 / submission_id=58762089631066370`(17 位,字符串无损) |
| +2.6s 起 | 豆包回复 **"视频生成参数确认"**(模型 / 时长 / 比例 / 创作方向,"确认后我再开始生成视频")——**视频任务先停在 `checking`**,体验套餐 + Fast 也不例外 |
| +2.6s ~ +12min | relay **一直没识别出 `checking`**,任务停在 generating/sent(两个缺陷,见下) |
| +12.2min | 探测时我在页面里调了一次 `acceptFeedback`(把脚本内部的网络阶段置为 checking)→ relay 的 `videoFollowup` 轮询报 `checking` + `manualRequired`(缺陷 2)→ 降级 manual → waiting_user_confirm |
| +14.5min | 手动用沈浪脚本自己的 `run('videoFollowup', {...payload, modelName, duration:5, ratio:"16:9", referenceFiles:[], userConfirmed:true})` 发出补发确认(`sent:true`);豆包回"本次使用 Seedance 2.0 Fast 生成,大约需要 1-3 分钟……本次生成将消耗每日免费额度" |
| +16.6min | 历史分类出 `result`("你的视频生成好了。")→ 同一次 `read` 带 `fallbacks[{vid, api}]` → **路径 3**:Rust GET API + 本地解密 **341ms** 拿到原片 → 静默 1.5s → downloading → **completed**(下载 0.5s) |

结果:`watermark_source=fallback_decrypted`,`video_1.mp4` 732,582 字节,`duration_ms=5057`(`mvhd` 解析;API 给的 `video_duration=5.057`)。Spotlight 元数据 `H.264 + MPEG-4 AAC`、1280×720、5.05s;`qlmanage -t` 渲染出画面帧(scratchpad `t33/uc20.mp4.png`),画面右下角**无豆包水印**。stage_history:`pending → preparing → sent → checking → waiting_user_confirm → result → downloading → completed`。

### 三条路径的真机事实

- **路径 3(fallback + 解密)是本版本唯一出现的路径**:reader `read`(type=video)对已完成的视频返回 `{ok:true, images:[], fallbacks:[{vid, api}]}`——`resolveVideos` 只要消息里的视频有 `fallback_api` 就**不再**调 `/creativity/resource/get_without_watermark`。`api` 形如 `https://vas-lf-x.snssdk.com/video/fplay/1/<hash>/<vid>?aid=482431&codec_type=0&…&key_seed=<base64>&logo_type=unwatermarked&…&channel=no`(reader `fallbackAPI` 已改写 `channel=no` / `codec_type=0` / `logo_type=unwatermarked`、删 `force_fids`),**无 cookie 的 GET 即 200**(3,133 字节 JSON),响应 `video_info.data.{key_seed, video_list.video_1.{main_url, backup_url_1}}`,`main_url` 是 `qAAB…` token。Rust `video_decoder::decode` 解出 `https://v11-vdl.doubao.com/<…>/?a=482431&…&lr=unwatermarked&…`(**本身已带 `lr=unwatermarked`**,不需要追加),与 Node 跑原版 `doubao-video-decoder.js` 的结果逐字相同。样本已存 `src-tauri/tests/fixtures/video_token_1.json`,Task 3.1 的 `fixture_video_token_1_decodes` 去掉 `#[ignore]` 并通过。
- **路径 1 没有出现**:窗口隐藏时页面不加载 `<video>`,整个生成期间 Network 旁听(5 分钟)没有任何视频 CDN 媒体请求;只有 `mark_conv_read`、`mcs.doubao.com/list`(埋点 `video_generation_result_show`)、`conversation/abstract`、`chain_by_block`。
- **路径 2 没有触发**:`result` 与 `fallbacks` 在同一次 `read` 里到达,原片已有,不再点"下载"按钮(这是预期的优先级)。
- reader 的 `images[{url, unwatermarked:true}]`(`get_without_watermark` 分支)本次没有走到,保留 mock 覆盖。

### 真机发现的缺陷(已修)

1. **流式回复被提前分类、之后不再分类**:capture 每秒 `read` 一次;bot 的"视频生成参数确认"在流式输出期间被读到时只有半句,`acceptFeedback` 分类为 `unknown`,而原实现按 **messageId** 去重,文本长全后不再送分类,`checking` 永远识别不出(12 分钟无进展)。修复:去重键改为 `messageId + 消息文本`(`capture::classify_key`);mock `streaming_feedback_is_reclassified_and_video_auto_confirm_carries_params`(半句 3 次 → 全文)。同一页面上直接对全文调 `acceptFeedback` 得 `checking`、对"你的视频生成好了。"得 `result`(relay 实际也是这样拿到 result 的)。
2. **视频 auto 代回复缺参数**:`videoFollowup` 在 auto + `checking` 时先用 `confirmationMatches` 把确认文案里的 `模型/时长/比例` 与请求的 `modelName/duration/ratio` 逐项比对,再补发 `生成视频：<prompt>\n使用模型：<modelName>；时长：<duration>秒；比例：<ratio>。<参考素材句>以上参数已由用户核对……`;原 `feedback_payload` 没有这几个字段 → `manualRequired:true` → 降级 manual。修复:`CaptureOptions` 增 `ratio / duration / reference_count`,payload 带 `modelName / ratio / duration(数字)/ referenceFiles`(脚本只用 `.length`,给占位名);runner 从请求填。真机已验证带全参数的 `videoFollowup` 发送成功并开始生成。
3. **reader 自家重放认不出**:真机 `requestWillBeSent.initiator.stack` 只有一帧 `chrome://doubao-chat/slardar/fe/sdk-web/`(slardar 包装了 `fetch`),看不到 `shenlangReadBoundConversation`,`is_reader_initiated` 失效,relay 把自己每秒的重放当成新历史请求再取响应体 analyze 一遍。修复:`record_history` 遇到 postData 与当前模板**完全相同**(重放复用模板的 `sequence_id`;页面自己的请求每次 `sequence_id` 不同)直接跳过;mock `replay_with_identical_post_data_is_not_reanalyzed`。

### 沈浪脚本 DOM 反馈在视频技能下失效(未修,靠历史分类兜底)

`feedbackDiagnostic` = `{reason:"new_user_message_not_located", submittedLength:18, baselineCount:1, matchingCount:1, anchorConnected:false}`:`send` 时脚本统计"输入区之外含本次 prompt 的元素"作基线,视频技能态下输入编辑器不在 `#input-engine-container` 里,编辑器自身被算进基线(=1);发送后页面只有 1 条用户消息(豆包改写为 `生成视频：<prompt>，16:9，5s`),`matches[1]` 不存在,锚点永远定位不到,`videoFeedback` / `videoFollowup` 的 DOM 识别一直 `unknown`(`videoFeedback` 返回 `{stage:"unknown", message:"我"}`)。视频的阶段识别因此**完全依赖历史 `read` + `acceptFeedback` 分类**(分类会顺带写入脚本内部的网络阶段,`videoFollowup` 的 `_0x386fff` 优先用它,所以分类之后 auto 代回复能工作);补发确认后脚本 `_0xedde11()` 以补发文本重建锚点(普通对话态,基线 0),之后 DOM 识别恢复。PS 原版对视频也有同样的 60s"尚未识别到本次任务反馈"兜底提示。

### 对后续 Task 的要求

1. **Task 4.1(多轮确认)**:视频**每次**都会先出"视频生成参数确认"(`checking`),auto 模式靠缺陷 1/2 修复后的历史分类 + `videoFollowup` 补发;修复后的 auto 链路真机**未完整跑过**(本次补发是手动调用同一脚本函数完成的),UC-22 时核对 stage 流 `checking → generating(rendering)→ result`。补发文案里时长是 `5秒`,豆包回显 `5 秒`。
2. **本次没有出现 `generating` 阶段事件**:补发后的"本次使用 Seedance 2.0 Fast 生成,大约需要 1-3 分钟"是新的 bot 消息,但 stage_history 从 waiting_user_confirm 直接到 result(历史分类出的是最新一条消息;DOM 轮询在 manual 降级后只走 `videoFeedback`)。
3. UC-21 场景 A/B(路径 1/2)本版本隐藏窗口下未出现,只有 mock 覆盖;UC-27(参考图生成)未做(只允许 1 次生成,选了无参考图的最便宜组合)。

## Task 4.1b · 多轮确认收尾 + 终态 webhook 接入(2026-10-17)

**本 Task 没有真机动作**:auto / manual 确认、降级、`videoFollowup` 补发都只能靠一次真实发送触发,而执行约束禁止除 UC-20 外任何消耗配额的发送;豆包 2.32.5 @9707 未被触碰。结论全部来自 mock(`doubao::capture` / `tasks::scheduler` 单测),真机事实沿用 Task 3.3 UC-20:

- 视频任务**每次**先出"视频生成参数确认"(`checking`);修复后的 auto 链路(历史分类 → `videoFollowup` 带全参数补发)仍**未在真机完整跑过**,UC-22 时需核对。
- 本 Task 新增的可观测点:代回复真正发出(图像 `acceptReply` 返回 `sent:true` / 视频 `videoFollowup` 返回 `sent:true`)时 capture 发 `Stage(Sent, "已自动点继续")`(auto)或 `Stage(Sent, "已按用户确认继续")`(用户确认后),stage_history 多一条 `sent`;UC-20 那种"waiting_user_confirm 直接跳到 result"的时间线,在 relay 自己补发时会变成 `waiting_user_confirm → generating → sent(已按用户确认继续)→ … → result`。同一确认 stage 重复报 `sent:true` 只记一次;`checking` 不会因为豆包提示尚未刷新而被重复广播。
- 终态 webhook:runner 的终态落库后与调度器的 `finish_queued`(pending 取消 / 排队超时 / 崩溃 / panic)都会投递;用 `hook.test` 域名 + `reqwest::ClientBuilder::resolve` 指到本地 axum mock 验证(`callback_url` 校验拒绝 loopback,真机回调需要非 loopback 地址)。

## Task 4.2 · 并发队列硬化 + 用户干扰检测 + 崩溃处理(2026-10-17)

**真机动作只有只读探测 + 一次开关空白 chat 窗口**(不发送、不切账号、不杀豆包;本机仅 1 个有效账号,崩溃处理全部 mock)。

### UC-28 前置:第二个 chat page 在 `/json/list` 里的形状(`Target.createTarget` 后立即 `closeTarget`)

- 浏览器端点 `Target.createTarget{url:"doubao://doubao-chat/chat", newWindow:true, background:true}` 成功;2.5s 后 `/json/list` 有**两个** `type=page` 的 doubao-chat:原页 `doubao://doubao-chat/chat/38447520992069890` + 新页 `doubao://doubao-chat/chat`(注意本机 chat 页的 scheme 当前是 `doubao://`,不是早期探测记录的 `chrome://`;白名单两者都接受)。`select_pages` 会把两者都算作 chat → capture 的 2s target 检查判 `generation_user_interrupted{interruption:"multiple_chat_pages", chat_targets:2}`(mock `multiple_chat_pages_interrupts_task` 覆盖)。`Target.closeTarget` 返回 `success:true`,1.5s 后回到 1 个。用户在豆包里"新开窗口"是否与此等价未能验证(无屏幕录制权限,不能操作豆包 UI);`/json/list` 是唯一信号源,形状一致即可。
- 输入框检查表达式真机只读:`[contenteditable="true"]` 恰好 1 个,空闲时 `{found:true, text:""}`。

### `live_ensure_ready_and_process_probe`(`cargo test live_ensure_ready_and_process_probe -- --ignored --nocapture`)

- 新建 `DoubaoConnection` 未连接时 `ensure_ready` → `doubao_not_ready`(503);`connect_sessions` 9ms 后放行,`last_seen_at` 为当前毫秒;`is_doubao_running`(lsof/ps)8ms。

### 设计取舍(与 spec 08 / UC-24 的冲突)

1. **跨账号任务串行**:豆包只有一个"当前账号",切账号整页 reload(Task 1.2 实测);若按 spec 08 让 3 个不同账号的任务同时跑,第 2 个任务的 `ensure_account` 会把第 1 个任务的页面切走(其 capture 丢失页面内提交状态、对错误账号重放历史)。调度器改为:队头任务的账号与任一运行中任务不同 → 等它们全部结束。叠加 `per_account_max=1`,单个豆包实例上**实际并发恒为 1**,`global_max` 仍作为上限生效(热改为 0 即停止派发)。UC-24"3 个不同账号并发 3"在单实例上不可达,只能做到"按 FIFO 依次切换账号执行"。
2. **手动切账号暂停派发**:`Scheduler::begin_switch` 在 `running` 锁内检查"无非终态任务"并置位 `switching`,守卫存活期间不派发,drop 后唤醒(消除 Task 2.8 留下的 check-then-act 窗口);切换中提交的任务正常入队,切换结束才 preparing(UC-31 mock)。
3. **新请求 503**:`Scheduler::submit` 在校验后先 `backend.ready()` → `DoubaoConnection::ensure_ready`:`cdp_connected` 或"进程在且心跳正在重连"(切账号 reload 的短暂断线)放行,其余 `doubao_not_ready{running, cdp_connected, reconnecting}`。原先只靠 `accounts(false)`,2s 缓存命中时崩溃后仍会入队。
4. **崩溃判定提前**:真杀豆包时 capture 约 6s(3 次 chat page 缺失)就以 `doubao_cdp_disconnected` 结束,早于连接层 20s 的 `Crashed` 事件,任务会落成 `doubao_cdp_disconnected` 而不是 UC-32 期望的 `doubao_crashed`。`DoubaoBackend` 在 submit / capture 返回断线类错误时查一次进程,进程不在 → `doubao_crashed{last_seen_at, cause}`;`Crashed` 事件仍兜底 `fail_all_running_with("doubao_crashed", …, {last_seen_at})`(运行中 + 排队中)。真机未杀豆包验证(执行约束)。
5. **输入框被改**:正常 = 空 / 等于本任务 prompt / relay 代回复的"继续生成"(代回复进行中不检查);其它内容 → warn(只记 `text_length`)+ `CaptureEvent::InputModified` → runner 把 `stage_message` 改为"检测到输入框被修改"并发 `task_updated`(不追加阶段历史、不终止),同一内容只报一次。

## Phase 6 稳定性 · Task 6.1 断线重连硬化 + 稳定性压测(2026-10-17)

### 重连行为(实现 = spec 09"断线检测与重连")

- 心跳失败 → 关老 socket → **等 2s**(`ConnTiming.reconnect_delay`,只在本次断线的第一次)→ 重新握手并复测;之后后台心跳在"断线中且豆包进程在"时每 2s 重试(`retry_soon`),不等 10s 周期。
- 自首次失败起 **20s**(`restart_after`)仍连不上且进程还在 → 升级为 `connect(true)`(杀掉并带 `--remote-debugging-port` 重启),成功即 `StatusChanged(cdp_connected)` → `doubao_status_changed`。**每次断线只升级一次**(失败后回到 10s 常规心跳,不会每 2s 杀一次豆包);断线恢复后再断重新计时。
- 不升级的两种情况:进程已退出(UC-06 要求"豆包已退出,等待重启",由用户点"重连豆包";`Crashed` 照旧在 20s 后发)、重连失败原因为 `multiple_chat_pages`(用户多开 chat 窗口,CDP 本身是通的,杀豆包会毁掉用户窗口)。
- 捕获中途 chat session 断开:`DoubaoConnection::await_chat_session(cancel)` 每 2s 重试 `chat_session()`,最长 20s,进程不在立即放弃,取消立即返回;原实现单次 `chat_session()` 失败(reload 时 chat page 暂时不在 `/json/list`)就以 `doubao_cdp_disconnected` 结束任务。
- `CdpSession::send_timeout` 加 drop 守卫:调用方外层 `tokio::time::timeout` / `select!` 丢弃 future 时 pending 条目同步删除(原实现只在内部超时时删,外层超时会在长期存活的 session 里永久残留一条)。

### `live_heartbeat_reconnects_after_socket_drop`(真豆包,只断 relay 一侧 socket,不杀豆包、不发送)

3 轮"relay 主动关 chat/background socket → `heartbeat()`":每轮 2016–2032ms 重连成功(= 2s 等待 + ~20ms 握手),target id 不变、`ui_compat=ok`、`reconnecting=false`;`await_chat_session` 在 session 已关时 3ms 接上。**20s 升级重启路径与"真杀豆包 → 自动恢复"未在真机验证**(任务约束禁止 pkill 豆包),只有 mock(`heartbeat_escalates_to_restart_after_restart_after`、`restart_is_attempted_once_per_outage`、`no_restart_when_doubao_exited_or_multiple_chat_pages`)。Task 1.10 留下的"Started 路径页面就绪等待真机复验"同理未做。

### `tests/soak.sh`(mock 后端,不耗配额)

`src-tauri/examples/soak_server.rs` = 真实 Storage / Scheduler / runner / HTTP / 日志 / GC + 假豆包后端(capture 200ms、下载写 64×64 PNG);`SOAK_CDP_PORT=9707` 时同时连真豆包跑 10s 心跳(生成仍走假后端)。release 构建,本机 M 系列:

| 指标 | 结果 | 阈值 |
|---|---|---|
| 100 个任务 | 100 completed / 0 failed,约 12s 跑完(串行,`queue_full` 429 退避 16 次) | UC-51 全部成功 |
| GET /tasks/{id} ×1000 | P50 0.43ms,P99 0.68ms,max 8.7ms | NF-01 P50<50ms / P99<500ms |
| RSS | 启动 24.2MB → 跑完任务 27.0MB(峰值 27.5MB)→ 连真豆包空闲 10 分钟后 15.3MB,无上涨趋势 | UC-50 <300MB,N1.3 空闲 <100MB |
| CPU(空闲末) | 0.0% | UC-50 <2% |
| 日志 ERROR | 0 条(2524 条 INFO) | UC-50 无 ERROR |

**未做**:24h 挂机(本次只有 10 分钟空闲 + 100 任务,RSS 曲线平稳下降后持平);真豆包 10 个真实任务(任务约束禁止消耗配额;脚本外部模式默认只采样不提交,`SOAK_SUBMIT_REAL=1` 才提交且封顶 10 个)。完整 Tauri 窗口进程(WKWebView)的 RSS 不在此数字内。

## Task 5.2 · 隐私声明 / 托盘 / 关闭流程(2026-10-17)

环境:豆包 2.32.5 运行中(1 个有效账号)。没有执行任何消耗豆包配额的操作,没有杀豆包。

### 窗口自动化前提

- 直接从后台 shell 跑 `target/debug/doubao-relay`(+ vite 1420)时,System Events 读到菜单栏(`doubao-relay / 编辑 / 窗口`)与托盘(`menu bar 2` 的 status menu,`help` = 托盘提示文字),但 **`count of windows = 0`**(与 Task 2.9 同一现象)。改为 `cargo tauri build --debug --bundles app` 后 `open -n <bundle>`,窗口与 WKWebView AX 树都可读、可点(Task 6.2 也是这样)。**后续 UI 真机验收用 `open -n` 打开 debug bundle。**
- 托盘图标的下拉菜单不进 AX 树(`click menu bar item 1 of menu bar 2` 后 `menu 1` 无效索引),托盘"显示主窗口 / 退出"两项真机只能人工点;同一处理函数由应用菜单"退出 doubao-relay"覆盖。

### 结果

| 检查 | 证据 | 结论 |
|---|---|---|
| 首次启动显示全屏声明 | 生产数据目录 `metadata` 无 `privacy_accepted`;AX:`使用前请阅读`、spec 12 四条承诺 + 两条警告原文、复选框 `我已阅读并理解上述说明与风险`(0)、`不同意并退出`(可点)、`同意并继续`(**enabled=false**) | 通过 |
| 同意前不启动 HTTP / 不自动连豆包 | 日志 `等待用户同意隐私声明,同意前不启动 HTTP API、不自动连接豆包`;`lsof -iTCP:8787` 无监听 | 通过 |
| 勾选 → 同意并继续 | `metadata.privacy_accepted = 1792172498178`;同一毫秒 `HTTP API 已监听 127.0.0.1:8787`,0.4s 后 `启动时自动连接豆包完成 action=AlreadyConnected` | 通过 |
| 关闭按钮(默认 `minimize_to_tray_on_close=true`) | 点 `AXCloseButton` 后 `count of windows = 0`、进程在、`GET /health` 200;日志 `主窗口已隐藏到托盘,HTTP 服务继续运行` | 通过 |
| Dock / 再次 `open`(`RunEvent::Reopen`) | 窗口数 0 → 1 | 通过 |
| 托盘 | `menu bar 2` 有 status item,提示文字 `doubao-relay · 空闲` | 通过(菜单项见上) |
| 设置页关掉"点击关闭按钮时隐藏到托盘而不是退出"后点关闭 | `config.toml` 写入 `false`(热生效,无需重启);点关闭后进程退出、8787 释放;日志依次 `开始关闭:停止接受新请求` → `HTTP API 已停止` → `关闭流程完成(豆包保持运行) cancelled=0 forced=0` → `数据库已关闭,日志 flush 后退出`(最后一行确实写进了日志文件,`flush` 生效);豆包进程仍在 | 通过 |
| 应用菜单 `退出 doubao-relay`(Cmd+Q) | 同上关闭流程日志,`exit=0`,豆包仍在 | 通过 |
| 声明页 `不同意并退出` | 走同一关闭流程退出,豆包仍在 | 通过 |

验收后已把生产配置的 `minimize_to_tray_on_close` 改回 `true`,并删除 `metadata.privacy_accepted`(用户下次打开会看到声明、自己确认)。

### 真机事实

- **macOS 默认菜单的 Quit(Cmd+Q)不可拦截**:muda 的 `PredefinedMenuItem::quit` 调 `NSApp terminate:`,tao 在 `applicationWillTerminate` 里直接发 `LoopDestroyed` → `RunEvent::Exit`,不经过 `ExitRequested`,`prevent_exit` 无从调用,关闭流程会被跳过。因此 macOS 上换成自定义应用菜单:`退出 doubao-relay`(`CmdOrCtrl+Q`,走关闭流程)+ 系统的 关于 / 隐藏 / 隐藏其他 / 全部显示,另保留"编辑"(撤销 / 重做 / 剪切 / 复制 / 粘贴 / 全选,WKWebView 输入框的快捷键依赖它)与"窗口"(最小化 / 关闭窗口)。`RunEvent::Exit` 仍兜底:关闭流程没跑过(例如系统注销)时用短时限(宽限 2s)同步跑一遍。
- 关闭流程中 generating 任务的 10s 宽限、超时强制 `failed/system_shutdown` 只有 mock 覆盖(真机跑需要真实生成,消耗配额)。

## 并发方案 C:专用窗口 验证(2026-10-17,豆包 2.32.5 macOS,部分完成)

只做了窗口创建 / 收起 / 内存与代码影响分析;第 3–6 步(新窗口账号、prepare→fill→checkSend、主窗口隔离、两窗口并行)**未执行**,探测中途停止。全程未发送、未读凭证、未切账号,创建的 target 已全部关闭(`Target.getTargets` 复核只剩原 chat 页)。

**窗口创建(browser 级 `Target.createTarget`,每次约 2.5s 出现在 `/json/list`)**
- `{url:"doubao://doubao-chat/chat"}` 与 `chrome://doubao-chat/chat` 都成功,`/json/list` 里 url 均为 `doubao://doubao-chat/chat`;不带 `newWindow` 时作为标签页加进同一个"豆包浏览器"窗口(windowId 相同)。
- 新窗口属于 `Doubao Browser` 进程,标题 `豆包 - 豆包浏览器 - <当前账号名>`(标题显示同一账号"烤鱼"),1280×800;用户主窗口"豆包"属于 `Doubao` 进程,两者分开。
- `background:true` 不阻止窗口弹出。
- `{newWindow:true, windowState:"minimized"}`:窗口创建即最小化(AXMinimized=true),不在屏幕上出现,只在程序坞里。**这是不打扰用户、且用户仍能看到的做法。**
- 事后 `Browser.setWindowBounds{windowState:"minimized"}` 也有效;移到屏外(`left:-4000`)被系统拉回原位,无效。
- `hidden:true` 能创建无窗口 target(不在 `/json/list`、`getWindowForTarget` 报 not found),未继续评估。

**内存**:同站点页面共用渲染进程。5 个额外页面时主渲染进程从 119 MB 涨到 863 MB(另一个从 107→220、54→270),关闭后回落到 347 MB(GC 滞后)。粗估每个额外页面 **150–250 MB**,高于此前估的 100–150 MB。

**relay 现有"chat 页 ≠ 1"判定(改成按 target id 认领专用窗口要动的地方)**
- `cdp/discovery.rs::select_pages` —— 只按 host 分类,需区分"用户窗口"与"relay 创建的窗口"(按 relay 记录的 target id 集合)。
- `doubao/connection.rs::connect_sessions_inner`(≈L452)—— chat ≠ 1 即 `doubao_not_ready`;改为:用户 chat 页按"非 relay 创建"挑选,专用页另建 session 池。
- `doubao/connection.rs::submit_session`(≈L638)—— >1 判 `multiple_chat_pages`;改为从池里取空闲专用页。
- `doubao/capture.rs`(≈L2116)每 2s 的 chat 数检查 —— 改为只检查"本任务认领的 target 还在";`heartbeat` 里 `multiple_chat_pages` 分支(≈L788)同理。
- 新增:专用页的创建(`newWindow + windowState:minimized`)、空闲/占用状态、关闭、崩溃后重建;`scheduler` 的 `per_account_max`/跨账号串行规则不变。

**账号切换的影响(只读推断,未实测)**:窗口标题带账号名、cookie 存储为进程级共享,切账号会让所有 doubao-chat 页一起变成新账号并 reload(Task 1.2/1.8 实测主页切换会 reload)。专用页上进行中的提交会中断、session 失效。因此专用窗口只解决"同账号并发 + 不碰用户窗口",跨账号仍须串行,切账号前所有专用页任务必须结束(现有 `begin_switch` 已保证无运行中任务)。

## 2026-10-19 · 豆包 2.32.6:发送后页面不再主动拉历史

- 现象:烤鱼 / 用户973735 / 用户741487 生图任务提交、绑定都成功,豆包里已出图,relay 却一直 `generating` 到 5 分钟超时。
  2026-10-16 成功的那次(她是翻哥,体验套餐,2.32.5)原图来自生成请求的 SSE 响应体(`generation_response`)。
- 原因:这些任务的生成响应里没有 `image_ori_raw`(结果改走推送),而页面在发送后不再调用 `/im/chain/single`,
  capture 一直等不到可重放的历史查询模板。
- 实测(只读,刷新对话):`POST /im/chain/single` 仍可用、返回含 `image_ori_raw`;请求体为
  `{"cmd":3100,"uplink_body":{"pull_singe_chain_uplink_body":{"conversation_id":"<字符串>","anchor_index":9007199254740991,"conversation_type":3,"direction":1,"limit":20,...}},"sequence_id":"<uuid>","channel":2,"version":"1"}`,
  头 `Agw-Js-Conv: str`、`Content-Type: application/json; encoding=utf-8`,query 与其它 `www.doubao.com` 请求相同(`aid` / `device_id` …,凭证在 cookie)。
  页面内 `fetch` 合成请求 → 200,读到原图 `…~tplv-…-image_raw.png?…x-expires=…`。
- 处理:capture 绑定后 3 秒(`SYNTH_HISTORY_AFTER`)仍无模板 → 合成上述模板交给 reader 重放;页面之后发了真的会替换。

### 2026-10-19 补充:合成请求方案撤回

- 合成的 `/im/chain/single`(无 `a_bogus`/`msToken`)第一次返回 200,后续同样请求被网关 403;豆包大部分接口带
  按接口+请求体计算的反爬签名。持续发无签名请求有被风控标记的风险,**不伪造签名**,撤回合成方案(`0e0c9d5` 的逻辑已删除)。
- 页面**自己**发的 `/im/chain/single`(重新打开对话时)不带签名、原样重放返回 200 且含 `image_ori_raw`;
  reader `analyze()` 对新拉数据能正确识别本次任务的图片与绑定。
- 当天"她是翻哥"任务中断的真实原因:`cargo tauri dev` 监视源码,排查时改代码触发自动重编译重启,打断进行中的任务
  (`relay_restart` / curl `Empty reply from server`)。之后以 `cargo tauri dev --no-watch` 运行。
- 同日实测:她是翻哥 Seedream 4.5 经 `/v1/images/generations` 10 秒完成,原图来自生成响应(`generation_response`),
  CDN 地址为 `…~tplv-…-image_raw.png`。烤鱼 / 用户741487 后续被豆包文字拒绝("我目前不能直接生成图片"),属账号侧限制。

### 2026-10-19 · 豆包 2.32.6:参考图上传入口改为按需生成

- 2.32.5 时技能输入区常驻一个隐藏 `input[type=file]`,沈浪 `fileInput` 只标记它;2.32.6 起页面上没有这个 input,
  `fileInput` 报"未找到唯一参考图上传入口"。
- 新入口:"+" 按钮 `[data-testid=upload_file_button][aria-haspopup=menu]` → 菜单项"添加图片"(图像技能)/
  "上传素材"(视频技能),点后才生成 `input[type=file]`(accept `.jpg,.png,.jpeg,.webp,.apng`, multiple)。
  点菜单项前须已开启 `Page.setInterceptFileChooserDialog`,否则会弹出系统文件对话框。
- 页面刷新后 JS 合成 pointer 事件打不开这个菜单;CDP `Input.dispatchMouseEvent` 可信点击后再补发 pointer 事件可以打开
  (窗口不在前台时第一下点击只激活窗口)。`upload::open_upload_entry` 按此实现,每张参考图的 `fileInput` 之前执行。
- 刷新页面后真机干跑(上传 → fill → checkSend,不发送):图像 Seedream 5.0 Pro、视频 Seedance 2.0 Fast 各通过。

### 2026-10-19 · 视频技能的"安全确认"弹窗与参数确认话术

- 首次发送带素材的生成时,点"发送"后弹 `[data-testid=safety_authorization_dialog]`(授权声明,按钮
  `safety_authorization_confirm_btn` / `safety_authorization_reject_btn`),确认后才发出生成请求。产品决定:默认自动点确认,
  不提供开关(`generation::accept_safety_dialog`,发送后最多等 800ms)。真机:自动确认后 356ms 识别到提交身份。
- 视频技能(Seedance 2.0 Mini + 参考图)豆包先回"我将基于参考图生成……请先确认参数:……确认后我再开始生成视频。"
  沈浪 `videoFeedback`/`videoFollowup` 对这句只返回 `{stage:"unknown", message:"我"}`,reader `analyze` 的 feedback 为 null,
  所以不会自动回复。消息 `content` 为空,文字在 `content_block[0].content.text_block.text` / `brief` / `tts_content`。
- 处理:视频 feedback 轮询时 relay 另读绑定对话最后一条消息(`[data-testid=receive_message]` vs `send_message`),
  命中 `确认参数|参数确认|确认后…开始生成|是否继续生成` 就按 checking 走 relay 自己的代回复(`accept_reply_expression`,回"继续生成")。

## 2026-10-24 · Windows 当前豆包：`chat_input` 生成工具栏

环境:Windows 本机已登录豆包，CDP `127.0.0.1:9707`，只读 DOM 探测后执行一次 `live_submit_dry_run`
（准备到 `checkSend`，不发送、不消耗额度）。

- 当前页面 `chrome://doubao-chat/chat` 的标题是已有会话；旧选择器
  `[data-testid="guidance-skill-bar"]`、`[data-testid="chat_input_action_mode"]`、
  `#input-engine-container` 都不存在。
- 图像工具栏已经在 `[data-testid="chat_input"]` 内，实际可见文本包含
  `图像生成 模型 Seedream 4.5 比例 自动 风格 模板`。这不是普通聊天页，不能因旧选择器缺失而
  触发“新对话”快捷键。
- `live_submit_dry_run` 使用 `Seedream 4.5 / 自动` 在 639ms 内完成
  `prepare → fill → checkSend`；结束后输入框已清空，没有执行 `send`，没有生成图片。
- 兼容规则:输入区选择器同时接受旧 `input-engine-container` 与新版 `chat_input`；当目标模型与参数
  控件已经稳定显示时，直接复用当前工具栏，不重新打开“对话”或重复点击技能。

## 2026-10-25 · Windows 图片工具栏切换时短暂消失

环境:Windows 本机豆包，CDP `127.0.0.1:9707`，`Seedream 4.5 / 16:9` 真机干跑（不发送）。

- 真实失败任务的 `prepare` 诊断为 `model=""`、`settings=[]`、`expectedRatio="16:9"`；相同请求此前可成功，
  排除模型、比例、提示词与分辨率差异，原因是模式切换后输入工具栏发生二次挂载。
- 修复后的干跑首次仍复现相同诊断；relay 在发送前重新打开一次图片生成工具栏，第二次探针读到
  `Seedream 4.5 / 16:9`，随后 `prepare → fill → checkSend` 通过并清空输入框，没有执行 `send`。
- 恢复范围只匹配新版图片诊断里模型为空且比例控件列表为空的情况；普通参数不支持、发送失败与发送后阶段
  均不自动重试，避免重复生成或重复消耗额度。
