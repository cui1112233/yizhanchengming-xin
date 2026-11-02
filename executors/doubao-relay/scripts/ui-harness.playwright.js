// Phase 1 账号页 UI 验收脚本(Task 1.10)。用法:先 `cargo tauri dev`(起 vite),再把本文件
// 交给 playwright MCP 的 browser_run_code_unsafe(filename),它会在普通 Chromium 里加载 vite(http://localhost:1477),
// 注入假的 window.__TAURI_INTERNALS__(invoke / 事件)模拟后端各状态,逐项截图。
// 账号/状态 payload 的形状与真后端(Task 1.9 to_camel 输出)一致。
async (page) => {
  const OUT = '/Users/james/github/bwy/doubao-relay/docs/analysis/screenshots/phase1';
  const results = [];
  const shot = async (name) => {
    await page.waitForTimeout(150);
    await page.screenshot({ path: `${OUT}/${name}.png` });
    results.push(name);
  };

  await page.addInitScript(() => {
    const status = (patch = {}) => ({
      running: true,
      cdpConnected: true,
      cdpPort: 9707,
      doubaoVersionDetected: '2.32.5',
      uiCompatCheck: 'ok',
      uiCompatDetails: { status: 'ok', accountButtonFound: true, inputFound: true, sendButtonFound: true, checkedAt: Date.now() },
      chatTargets: 1,
      reconnecting: false,
      lastCheckAt: Date.now(),
      ...patch,
    });
    const acc = (name, subscription, extra = {}) => ({
      name,
      subscription,
      label: `${name}\n${subscription ?? ''}`.trim(),
      isCurrent: false,
      isAmbiguous: false,
      ...extra,
    });
    const mock = {
      status: status(),
      list: {
        current: '她是翻哥',
        accounts: [
          acc('她是翻哥', '体验套餐', { isCurrent: true }),
          acc('李四', '高级套餐'),
          acc('王五', '标准套餐', { isAmbiguous: true }),
          acc('王五', null, { isAmbiguous: true }),
        ],
        readAt: Date.now(),
      },
      listError: null,
      listDelayMs: Number(sessionStorage.getItem('mockListDelay') ?? 0),
      reconnect: { action: 'already_connected' },
      reconnectError: null,
      switchDelayMs: 1200,
      calls: [],
      listeners: {},
      callbacks: {},
      next: 1,
      emit(event, payload) {
        for (const id of this.listeners[event] ?? []) {
          const cb = this.callbacks[id];
          if (cb) cb({ event, id, payload });
        }
      },
    };
    window.__mock = mock;
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
    window.__TAURI_EVENT_PLUGIN_INTERNALS__ = {
      unregisterListener(event, id) {
        mock.listeners[event] = (mock.listeners[event] ?? []).filter((x) => x !== id);
      },
    };
    window.__TAURI_INTERNALS__ = {
      transformCallback(cb) {
        const id = mock.next++;
        mock.callbacks[id] = cb;
        return id;
      },
      unregisterCallback(id) {
        delete mock.callbacks[id];
      },
      async invoke(cmd, args) {
        mock.calls.push({ cmd, args });
        switch (cmd) {
          case 'plugin:event|listen': {
            (mock.listeners[args.event] ??= []).push(args.handler);
            return args.handler;
          }
          case 'plugin:event|unlisten':
            return null;
          case 'doubao_status':
            return mock.status;
          case 'account_list':
          case 'account_refresh':
            if (mock.listDelayMs) await sleep(mock.listDelayMs);
            if (mock.listError) throw mock.listError;
            return mock.list;
          case 'account_switch': {
            await sleep(mock.switchDelayMs);
            const target = args.target;
            const found = mock.list.accounts.filter((a) => a.name === target);
            if (found.length === 0) throw { code: 'account_not_found', message: '指定账号在豆包中未找到,请刷新账号列表后重试', details: { requested: target, available: mock.list.accounts.map((a) => a.name) } };
            if (found.length > 1 || found[0].isAmbiguous) throw { code: 'account_ambiguous', message: '账号名在豆包中有多个,请在豆包里改名后重试', details: { requested: target } };
            if (mock.switchError) throw mock.switchError;
            const previous = mock.list.current;
            mock.list = {
              ...mock.list,
              current: target,
              accounts: mock.list.accounts.map((a) => ({ ...a, isCurrent: a.name === target && !a.isAmbiguous })),
              readAt: Date.now(),
            };
            const switchedAt = Date.now();
            mock.emit('account_switched', { previous, current: target, switchedAt });
            mock.emit('account_list_changed', mock.list);
            return { ok: true, previous, current: target, switchedAt };
          }
          case 'doubao_reconnect': {
            await sleep(600);
            if (mock.reconnectError) throw mock.reconnectError;
            mock.status = status();
            mock.emit('doubao_status_changed', mock.status);
            return { ...mock.reconnect, cdpReadyAt: Date.now() };
          }
          default:
            throw { code: 'system_internal_error', message: `mock 未实现命令 ${cmd}` };
        }
      },
    };
  });

  await page.setViewportSize({ width: 1000, height: 680 });

  // --- 场景 1:多账号(UC-03),含同名 ⚠ ---
  await page.goto('http://localhost:1477/');
  await page.evaluate(() => sessionStorage.clear()); // 上一轮骨架屏场景留下的延迟标记
  await page.reload();
  await page.getByText('共 4 个已登录账号').waitFor();
  await shot('uc03-multi-accounts');
  const rowsText = await page.locator('.accounts-table tbody tr').allInnerTexts();

  // --- 场景 2:切换(UC-04):确认框 → 切换中 → 新当前账号 + toast ---
  const switchButtons = page.locator('.accounts-table tbody tr', { hasText: '李四' }).getByRole('button', { name: '切换' });
  await switchButtons.click();
  await page.getByRole('dialog').waitFor();
  await shot('uc04-confirm-dialog');
  const dialogText = await page.getByRole('dialog').innerText();
  await page.getByRole('dialog').getByRole('button', { name: '切换' }).click();
  await page.getByText('切换中').waitFor();
  await shot('uc04-switching');
  await page.getByText('已切换到账号“李四”').waitFor({ timeout: 5000 });
  await shot('uc04-switched');
  const currentLine = await page.locator('.current-account').innerText();
  const currentRow = await page.locator('.accounts-table tbody tr.row-current').innerText();
  // 同名账号没有切换按钮
  const ambiguousRowText = await page.locator('.accounts-table tbody tr[data-ambiguous]').first().innerText();
  const ambiguousHasButton = await page.locator('.accounts-table tbody tr[data-ambiguous] button').count();

  // --- 场景 2b:切换失败(后端报错)→ 错误 toast 常驻 ---
  await page.evaluate(() => {
    window.__mock.switchError = { code: 'account_switch_timeout', message: '账号切换超时,请检查豆包窗口后重试', details: { elapsedMs: 20000 } };
    window.__mock.switchDelayMs = 200;
  });
  await page.locator('.accounts-table tbody tr', { hasText: '她是翻哥' }).getByRole('button', { name: '切换' }).click();
  await page.getByRole('dialog').getByRole('button', { name: '切换' }).click();
  await page.getByText('切换账号失败:账号切换超时').waitFor({ timeout: 5000 });
  await page.waitForTimeout(3500); // 成功 toast 已自动消失,错误 toast 仍在
  const toastsAfter = await page.locator('.toast').allInnerTexts();
  await shot('uc04-switch-error-toast');
  await page.locator('.toast-close').first().click();

  // --- 场景 3:未登录(UC-05):account_list 抛 doubao_not_logged_in ---
  await page.evaluate(() => {
    window.__mock.listError = { code: 'doubao_not_logged_in', message: '豆包未登录,请先在豆包中登录账号' };
  });
  await page.getByRole('button', { name: '刷新' }).click();
  await page.getByText('豆包尚未登录任何账号').waitFor();
  await shot('uc05-not-logged-in');

  // --- 场景 4:豆包退出(UC-06):status_changed running=false → 状态栏红 + 空状态 + 错误 toast ---
  await page.evaluate(() => {
    window.__mock.listError = null;
    window.__mock.status = { ...window.__mock.status, running: false, cdpConnected: false, reconnecting: false, chatTargets: 0 };
    window.__mock.emit('doubao_status_changed', window.__mock.status);
  });
  await page.locator('.empty-state[data-kind="not_connected"]').waitFor();
  await page.getByText('豆包已退出,等待重启。点击“重连豆包”').waitFor();
  await shot('uc06-doubao-exited');
  const statusBar = await page.locator('.status-bar').innerText();
  const dotColor = await page.locator('.status-item').first().getAttribute('data-color');
  // 点"重连豆包"(空状态里的那个)→ mock 返回 started → 恢复已连接并重读账号
  await page.evaluate(() => {
    window.__mock.reconnect = { action: 'started' };
  });
  await page.locator('.empty-state').getByRole('button', { name: '重连豆包' }).click();
  await page.getByText('连接中…').first().waitFor();
  await page.getByText('已启动豆包并连接').waitFor({ timeout: 5000 });
  await page.getByText('共 4 个已登录账号').waitFor();
  await shot('uc06-recovered');
  const statusBarAfter = await page.locator('.status-bar').innerText();

  // --- 场景 5:重连中(黄)与 CDP 未连(连接豆包按钮)---
  await page.evaluate(() => {
    window.__mock.status = { ...window.__mock.status, running: true, cdpConnected: false, reconnecting: true };
    window.__mock.emit('doubao_status_changed', window.__mock.status);
  });
  await page.getByText('正在尝试重新连接豆包').waitFor();
  const yellow = await page.locator('.status-item').first().getAttribute('data-color');
  await shot('uc06-reconnecting');

  // --- 场景 6:首屏加载骨架屏 ---
  await page.evaluate(() => {
    window.__mock.status = { ...window.__mock.status, running: true, cdpConnected: true, reconnecting: false };
    sessionStorage.setItem('mockListDelay', '100000');
  });
  await page.reload();
  await page.locator('.skeleton').waitFor();
  await shot('loading-skeleton');

  return { results, rowsText, dialogText, currentLine, currentRow, ambiguousRowText, ambiguousHasButton, toastsAfter, statusBar, dotColor, statusBarAfter, yellow };
}
