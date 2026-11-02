// Phase 5 设置页 / 日志页 UI 验收脚本(Task 5.1)。用法同 ui-harness.playwright.js:先 `cargo tauri dev`(起 vite),
// 再把本文件交给 playwright MCP 的 browser_run_code_unsafe,它在普通 Chromium 里加载 http://localhost:1477,
// 注入假的 window.__TAURI_INTERNALS__(invoke / 事件 / 剪贴板 / 对话框插件)逐项截图。
async (page) => {
  const OUT = '/Users/james/github/bwy/doubao-relay/.claude/worktrees/wf_5bc9c88a-747-4/docs/analysis/screenshots/phase5';
  const results = [];
  const shot = async (name) => {
    await page.waitForTimeout(150);
    await page.screenshot({ path: `${OUT}/${name}.png` });
    results.push(name);
  };

  await page.addInitScript(() => {
    const TOKEN = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef';
    const status = () => ({
      running: true,
      cdpConnected: true,
      cdpPort: 9707,
      doubaoVersionDetected: '2.32.5',
      uiCompatCheck: 'ok',
      uiCompatDetails: { status: 'ok', accountButtonFound: true, inputFound: true, sendButtonFound: true, checkedAt: Date.now() },
      chatTargets: 1,
      reconnecting: false,
      lastCheckAt: Date.now(),
    });
    const mock = {
      status: status(),
      config: {
        server: { host: '127.0.0.1', port: 8787, bearerToken: TOKEN },
        concurrency: { globalMax: 3, perAccountMax: 1, queueDepthMax: 50 },
        storage: { resultRetentionHours: 24, metadataRetentionDays: 7, uploadExpiryHours: 1 },
        logging: { level: 'info', fileMaxSizeMb: 10, fileRetentionDays: 7, cdpDebugLog: false },
        doubao: { cdpPort: 9707, cdpAddress: '127.0.0.1', autoConnectOnStartup: true, uiCompatCheckOnStartup: true },
        ui: { minimizeToTrayOnClose: true, startMinimized: false },
      },
      logs: [],
      cdp: [],
      clipboard: '',
      savePath: '/Users/james/Downloads/doubao-relay-logs.ndjson',
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
    const t0 = Date.now() - 60_000;
    const L = (i, level, target, message, fields = {}) => ({ ts: t0 + i * 1000, level, target, message, fields });
    mock.logs = [
      L(1, 'info', 'doubao_relay_lib', 'doubao-relay 启动', { platform: 'macos', cdp_port: 9707 }),
      L(2, 'info', 'doubao_relay_lib::doubao::connection', 'chat session 已建立', { target_id: '9E9F63BE' }),
      L(3, 'info', 'doubao_relay_lib::doubao::compat', 'UI 兼容性自检完成', { status: 'ok' }),
      L(4, 'debug', 'cdp', 'CDP 命令完成', { method: 'Target.getTargets', elapsed_ms: 1 }),
      L(5, 'info', 'doubao_relay_lib::doubao::accounts', '账号列表读取完成', { count: 1, current: '她是翻哥' }),
      L(6, 'warn', 'doubao_relay_lib::doubao::connection', '心跳失败,标记豆包连接断开', { error: 'Closed' }),
      L(7, 'error', 'doubao_relay_lib::doubao::accounts', '账号按钮未找到,判定豆包界面变化', { haspopup_count: 0 }),
      L(8, 'info', 'doubao_relay_lib::commands::app', '配置已保存', { needs_restart: false }),
    ];
    mock.cdp = [
      { ts: t0 + 1000, direction: 'send', method: 'Target.getTargets', params: {} },
      { ts: t0 + 1001, direction: 'recv', method: 'Target.getTargets', result: { targetInfos: [{ type: 'page', url: 'doubao://doubao-chat/chat' }] } },
      { ts: t0 + 2000, direction: 'send', method: 'Runtime.evaluate', params: { expression: 'window.__doubaoRelayAccounts.current()', returnByValue: true, awaitPromise: true } },
      { ts: t0 + 2010, direction: 'recv', method: 'Runtime.evaluate', result: { result: { type: 'object', value: { status: 'logged_in', name: '她是翻哥' } } } },
      { ts: t0 + 3000, direction: 'recv', method: 'Network.requestWillBeSent', event: { requestId: '1', request: { url: 'https://www.doubao.com/samantha/chat/completion' } } },
    ];
    window.__mock = mock;
    window.__TAURI_EVENT_PLUGIN_INTERNALS__ = {
      unregisterListener(event, id) {
        mock.listeners[event] = (mock.listeners[event] ?? []).filter((x) => x !== id);
      },
    };
    const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
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
          case 'plugin:event|listen':
            (mock.listeners[args.event] ??= []).push(args.handler);
            return args.handler;
          case 'plugin:event|unlisten':
            return null;
          case 'plugin:clipboard-manager|write_text':
            mock.clipboard = args.data ?? args.text ?? JSON.stringify(args);
            return null;
          case 'plugin:clipboard-manager|read_text':
            return mock.clipboard;
          case 'plugin:dialog|save':
            return mock.savePath;
          case 'doubao_status':
            return mock.status;
          case 'account_list':
          case 'account_refresh':
            return { current: '她是翻哥', accounts: [{ name: '她是翻哥', subscription: '体验套餐', label: '她是翻哥\n体验套餐', isCurrent: true, isAmbiguous: false }], readAt: Date.now() };
          case 'app_get_version':
            return { version: '0.1.0', buildTarget: 'macos', tauriVersion: '2.12.4', rustcVersion: '1.94.0' };
          case 'app_get_config':
            return JSON.parse(JSON.stringify(mock.config));
          case 'app_save_config': {
            await sleep(200);
            const patch = args.patch;
            if (patch.server && patch.server.port === 0) throw { code: 'request_malformed', message: '配置取值不合法,请按提示修改后重试', details: { reason: 'server.port 不能为 0' } };
            const needsRestart = (patch.server && patch.server.port !== undefined && patch.server.port !== mock.config.server.port) || (patch.doubao && patch.doubao.cdpPort !== undefined && patch.doubao.cdpPort !== mock.config.doubao.cdpPort);
            for (const [g, v] of Object.entries(patch)) mock.config[g] = { ...mock.config[g], ...v };
            return { ok: true, needsRestart: !!needsRestart };
          }
          case 'app_regenerate_token':
            await sleep(200);
            mock.config.server.bearerToken = 'fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210';
            return { bearerToken: mock.config.server.bearerToken };
          case 'maintenance_gc':
            await sleep(300);
            return { resultFilesRemoved: 3, tasksRemoved: 2, uploadsRemoved: 1 };
          case 'log_recent':
            return { entries: mock.logs.slice(-(args.limit ?? 200)) };
          case 'log_export':
            return { ok: true, path: args.path, count: mock.logs.length };
          case 'debug_cdp_log':
            return { entries: mock.config.logging.cdpDebugLog ? mock.cdp : [], enabled: mock.config.logging.cdpDebugLog };
          default:
            throw { code: 'system_internal_error', message: `mock 未实现命令 ${cmd}` };
        }
      },
    };
  });

  await page.setViewportSize({ width: 1000, height: 680 });
  await page.goto('http://localhost:1477/');

  // --- 设置页:全部分组 ---
  await page.getByRole('button', { name: '设置', exact: true }).click();
  await page.getByText('HTTP API 服务器').waitFor();
  await shot('settings-overview');
  const groups = await page.locator('.settings-group-title').allInnerTexts();
  const maskedToken = await page.locator('.token').innerText();
  const curl = await page.locator('.curl-example').innerText();

  // 复制 token → 剪贴板里是完整 token;toast 提示 60s 清除
  await page.getByRole('button', { name: '复制' }).click();
  await page.getByText('已复制 token').waitFor();
  const clipboardAfterCopy = await page.evaluate(() => window.__mock.clipboard);
  await shot('settings-token-copied');

  // 显示 token
  await page.getByRole('button', { name: '显示' }).click();
  const shownToken = await page.locator('.token').innerText();

  // 重新生成 → 确认框 → 新 token
  await page.getByRole('button', { name: '重新生成' }).click();
  await page.getByRole('dialog').waitFor();
  await shot('settings-regenerate-confirm');
  await page.getByRole('dialog').getByRole('button', { name: '重新生成' }).click();
  await page.getByText('已重新生成 token').waitFor();
  const regeneratedToken = await page.locator('.token').innerText();

  // 端口改 9000 并保存 → needsRestart 提示
  const portInput = page.locator('.settings-group[data-group="server"] input[type="number"]');
  await portInput.fill('9000');
  await page.locator('.settings-group[data-group="server"]').getByRole('button', { name: '保存' }).click();
  await page.locator('.toast', { hasText: '需要重启' }).waitFor();
  await shot('settings-saved-needs-restart');
  const curlAfterPort = await page.locator('.curl-example').innerText();

  // 非法端口 → inline 红字 + 不调后端
  const callsBefore = await page.evaluate(() => window.__mock.calls.filter((c) => c.cmd === 'app_save_config').length);
  await portInput.fill('0');
  await page.locator('.settings-group[data-group="server"]').getByRole('button', { name: '保存' }).click();
  await page.locator('.settings-error').first().waitFor();
  const inlineError = await page.locator('.settings-error').first().innerText();
  const callsAfter = await page.evaluate(() => window.__mock.calls.filter((c) => c.cmd === 'app_save_config').length);
  await shot('settings-invalid-port');
  await portInput.fill('9000');

  // 日志组:级别 debug + 开 CDP 日志 → 保存(热生效,不提示重启)
  await page.locator('.settings-group[data-group="logging"] select').selectOption('debug');
  await page.locator('.settings-group[data-group="logging"] input[type="checkbox"]').check();
  await page.locator('.settings-group[data-group="logging"]').getByRole('button', { name: '保存' }).click();
  await page.locator('.toast.toast-success', { hasText: /^已保存\s*×$/ }).waitFor();
  const loggingToasts = await page.locator('.toast').allInnerTexts();
  const savedLogging = await page.evaluate(() => window.__mock.config.logging);

  // 立即清理 → 汇总 toast
  await page.getByRole('button', { name: '立即清理' }).click();
  await page.getByText('清理完成').waitFor();
  const gcToast = await page.locator('.toast', { hasText: '清理完成' }).innerText();
  await shot('settings-gc-done');
  for (const b of await page.locator('.toast-close').all()) await b.click().catch(() => {});

  // --- 日志页 ---
  await page.getByRole('button', { name: '日志', exact: true }).click();
  await page.locator('.log-table tbody tr').first().waitFor();
  await shot('logs-all');
  const logRowCount = await page.locator('.log-table tbody tr').count();
  // 实时追加
  await page.evaluate(() => {
    window.__mock.emit('log_entry', { ts: Date.now(), level: 'info', target: 'doubao_relay_lib::tasks::scheduler', message: '任务 01HQZ9ABCDEF 完成', fields: { taskId: '01HQZ9ABCDEF' } });
  });
  await page.getByText('任务 01HQZ9ABCDEF 完成').waitFor();
  const logRowCountAfterEvent = await page.locator('.log-table tbody tr').count();
  // 级别过滤 warn
  await page.locator('.page-logs select').selectOption('warn');
  await page.waitForTimeout(100);
  const warnRows = await page.locator('.log-table tbody tr').allInnerTexts();
  await shot('logs-filter-warn');
  // 导出
  await page.getByRole('button', { name: '导出' }).click();
  await page.getByText('已导出').waitFor();
  const exportToast = await page.locator('.toast', { hasText: '已导出' }).innerText();
  const exportCall = await page.evaluate(() => window.__mock.calls.find((c) => c.cmd === 'log_export'));
  await shot('logs-exported');
  for (const b of await page.locator('.toast-close').all()) await b.click().catch(() => {});
  // 清空 → 空状态
  await page.locator('.page-logs select').selectOption('');
  await page.getByRole('button', { name: '清空' }).click();
  await page.locator('.empty-state[data-kind="no_logs"]').waitFor();
  await shot('logs-empty');

  // --- CDP 调试面板(已开启)---
  await page.getByRole('tab', { name: /CDP 调试/ }).click();
  await page.locator('.cdp-table tbody tr').first().waitFor();
  await shot('logs-cdp-panel');
  const cdpRows = await page.locator('.cdp-table tbody tr').count();
  await page.evaluate(() => {
    window.__mock.emit('cdp_log_entry', { ts: Date.now(), direction: 'send', method: 'Target.getTargets', params: {} });
  });
  await page.waitForTimeout(100);
  const cdpRowsAfterEvent = await page.locator('.cdp-table tbody tr').count();
  await page.locator('.cdp-table details').first().locator('summary').click();
  await shot('logs-cdp-expanded');

  // --- CDP 未开启的面板 ---
  await page.evaluate(() => {
    window.__mock.config.logging.cdpDebugLog = false;
  });
  await page.reload();
  await page.getByRole('button', { name: '日志', exact: true }).click();
  await page.getByRole('tab', { name: /CDP 调试/ }).click();
  await page.locator('.empty-state[data-kind="cdp_disabled"]').waitFor();
  await shot('logs-cdp-disabled');

  // --- 快捷键:Cmd+, 设置 / Cmd+L 日志 / Cmd+R 账号 ---
  await page.keyboard.press('Meta+,');
  await page.getByText('HTTP API 服务器').waitFor();
  const afterCmdComma = await page.locator('.page-title').innerText();
  await page.keyboard.press('Meta+l');
  await page.waitForTimeout(100);
  const afterCmdL = await page.locator('.page-title').innerText();
  await page.keyboard.press('Meta+r');
  await page.waitForTimeout(100);
  const afterCmdR = await page.locator('.page-title').innerText();

  return {
    results, groups, maskedToken, curl, clipboardAfterCopy, shownToken, regeneratedToken, curlAfterPort, inlineError,
    saveCallsDelta: callsAfter - callsBefore, loggingToasts, savedLogging, gcToast, logRowCount, logRowCountAfterEvent,
    warnRows, exportToast, exportCall, cdpRows, cdpRowsAfterEvent, afterCmdComma, afterCmdL, afterCmdR,
  };
}
