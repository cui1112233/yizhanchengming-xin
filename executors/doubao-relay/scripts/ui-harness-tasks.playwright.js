// Phase 2 任务页 / 详情页 / 新建弹窗 UI 验收脚本(Task 2.9)。用法同 ui-harness.playwright.js:
// 先 `cargo tauri dev`(起 vite 1420),跑完 `scripts/uc-10.sh --generate` 后把真实任务详情(HTTP 体 → camelCase
// + result.files[].localPath,与 IPC `task_get` 同形)写成 `target/ui-fixture-tasks.js`(`window.__FIXTURE = {detail, list}`),
// 再把本文件交给 playwright MCP 的 browser_run_code_unsafe。`convertFileSrc` 被 mock 成
// http://asset.localhost/<encoded path>,由 page.route 用真实结果文件应答,所以缩略图是真机生成的那张图。
async (page) => {
  const ROOT = '/Users/james/github/bwy/doubao-relay';
  const OUT = `${ROOT}/docs/analysis/screenshots/phase2`;
  // FIXTURE 放在 target/(gitignore),里面是本机真实任务详情与结果文件绝对路径,不入库。
  const FIXTURE = `${ROOT}/target/ui-fixture-tasks.js`;
  const results = [];
  const shot = async (name) => {
    await page.waitForTimeout(200);
    await page.screenshot({ path: `${OUT}/${name}.png` });
    results.push(name);
  };

  await page.route('http://asset.localhost/**', (route) => {
    const p = decodeURIComponent(new URL(route.request().url()).pathname.slice(1));
    return route.fulfill({ path: p });
  });
  await page.addInitScript({ path: FIXTURE });

  await page.addInitScript(() => {
    const F = window.__FIXTURE;
    const now = Date.now();
    const status = {
      running: true,
      cdpConnected: true,
      cdpPort: 9707,
      doubaoVersionDetected: '2.32.5',
      uiCompatCheck: 'ok',
      uiCompatDetails: { status: 'ok', accountButtonFound: true, inputFound: true, sendButtonFound: true, checkedAt: now },
      chatTargets: 1,
      reconnecting: false,
      lastCheckAt: now,
    };
    const extra = (id, patch) => ({
      taskId: id,
      account: '她是翻哥',
      type: 'image',
      status: 'pending',
      stage: null,
      createdAt: now - 60_000,
      completedAt: null,
      promptPreview: '示例任务',
      ...patch,
    });
    // 真实任务 + 几条演示行(生成中 / 失败 / 排队 / 待确认 / 视频)
    const list = [
      extra('01M5ZZDEMOWAITCONFIRM00001', { status: 'waiting_user_confirm', stage: 'checking', createdAt: now - 20_000, promptPreview: '手动确认演示:一座雪山下的木屋' }),
      extra('01M5ZZDEMOGENERATING000001', { type: 'video', status: 'generating', stage: 'rendering', createdAt: now - 2 * 60_000, promptPreview: '视频演示:海浪拍打礁石,慢镜头' }),
      ...F.list.tasks,
      extra('01M5ZZDEMOFAILED0000000001', { status: 'failed', stage: 'refused', createdAt: now - 5 * 60_000, completedAt: now - 4 * 60_000, promptPreview: '失败演示' }),
      extra('01M5ZZDEMOPENDING000000001', { status: 'pending', createdAt: now - 5 * 60_000, promptPreview: '排队演示' }),
    ];
    const waitDetail = {
      ...F.detail,
      taskId: '01M5ZZDEMOWAITCONFIRM00001',
      status: 'waiting_user_confirm',
      stage: 'checking',
      message: '豆包确认生成参数,等待手动确认',
      request: { ...F.detail.request, prompt: '手动确认演示:一座雪山下的木屋', confirmationMode: 'manual', autoUseMembership: false, ratio: '4:3' },
      result: undefined,
      resultFiles: null,
      completedAt: null,
      stageHistory: [
        { stage: 'pending', ts: now - 20_000 },
        { stage: 'preparing', ts: now - 19_000 },
        { stage: 'sent', ts: now - 18_000 },
        { stage: 'checking', ts: now - 15_000, message: '豆包确认生成参数' },
        { stage: 'waiting_user_confirm', ts: now - 15_000, message: '豆包确认生成参数,等待手动确认' },
      ],
    };
    const mock = {
      list,
      details: { [F.detail.taskId]: F.detail, [waitDetail.taskId]: waitDetail },
      calls: [],
      listeners: {},
      callbacks: {},
      next: 1,
      emit(event, payload) {
        for (const id of this.listeners[event] ?? []) this.callbacks[id]?.({ event, id, payload });
      },
    };
    window.__mock = mock;
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
      convertFileSrc(p) {
        return `http://asset.localhost/${encodeURIComponent(p)}`;
      },
      async invoke(cmd, args) {
        mock.calls.push({ cmd, args });
        switch (cmd) {
          case 'plugin:event|listen':
            (mock.listeners[args.event] ??= []).push(args.handler);
            return args.handler;
          case 'plugin:event|unlisten':
            return null;
          case 'doubao_status':
            return status;
          case 'account_list':
          case 'account_refresh':
            return {
              current: '她是翻哥',
              accounts: [{ name: '她是翻哥', subscription: '体验套餐', label: '她是翻哥\n体验套餐', isCurrent: true, isAmbiguous: false }],
              readAt: Date.now(),
            };
          case 'task_list': {
            const f = args.filter ?? {};
            const rows = mock.list.filter((t) => !f.status || f.status.includes(t.status));
            return { tasks: rows, nextCursor: null };
          }
          case 'task_get': {
            const d = mock.details[args.taskId];
            if (!d) throw { code: 'task_not_found', message: '任务不存在,请检查 task_id', details: { task_id: args.taskId } };
            return d;
          }
          case 'task_submit': {
            const r = args.request;
            if (!r.prompt.trim()) throw { code: 'request_missing_field', message: '缺少必填字段 prompt,请补全后重试', details: { field: 'prompt' } };
            throw { code: 'queue_full', message: '演示环境不真正提交' };
          }
          case 'task_confirm':
            return { ok: true, newStage: 'generating' };
          case 'task_cancel':
            return { ok: true, cancelledFromStatus: 'pending' };
          case 'task_open_result_folder':
            return { ok: true, path: '/tmp' };
          default:
            throw { code: 'system_internal_error', message: `mock 未实现命令 ${cmd}` };
        }
      },
    };
  });

  await page.setViewportSize({ width: 1000, height: 680 });
  await page.goto('http://localhost:1477/');
  await page.getByRole('button', { name: '任务' }).click();
  await page.locator('.tasks-table').waitFor();
  await shot('tasks-list');
  const listRows = await page.locator('.tasks-table tbody tr').allInnerTexts();
  const realRowTitle = await page.locator(`.tasks-table tr[data-task-id="${await page.evaluate(() => window.__FIXTURE.detail.taskId)}"]`).getAttribute('title');

  // 实时更新:task_updated 把"排队演示"改成生成中
  await page.evaluate(() =>
    window.__mock.emit('task_updated', { taskId: '01M5ZZDEMOPENDING000000001', status: 'generating', stage: 'sent', message: null, progress: null, updatedAt: Date.now() }),
  );
  await page.locator('tr[data-task-id="01M5ZZDEMOPENDING000000001"]', { hasText: '生成中' }).waitFor();
  const liveRow = await page.locator('tr[data-task-id="01M5ZZDEMOPENDING000000001"]').innerText();

  // 详情:真实完成任务 + 缩略图
  await page.evaluate(() => document.querySelector(`tr[data-task-id="${window.__FIXTURE.detail.taskId}"]`).click());
  await page.locator('.result-media').waitFor();
  await page.waitForFunction(() => {
    const img = document.querySelector('img.result-media');
    return img && img.complete && img.naturalWidth > 0;
  });
  const imgSize = await page.evaluate(() => {
    const img = document.querySelector('img.result-media');
    return [img.naturalWidth, img.naturalHeight];
  });
  await shot('task-detail-completed');
  await page.locator('.raw-request summary').click();
  await page.locator('.main').evaluate((el) => el.scrollTo(0, el.scrollHeight));
  await shot('task-detail-raw-request');
  const timeline = await page.locator('.timeline').innerText();

  // 详情:等待手动确认
  await page.getByRole('button', { name: '← 返回' }).click();
  await page.locator('tr[data-task-id="01M5ZZDEMOWAITCONFIRM00001"]').click();
  await page.getByRole('button', { name: '手动确认' }).waitFor();
  await shot('task-detail-waiting-confirm');
  const waitButtons = await page.locator('.page-actions button').allInnerTexts();

  // 新建弹窗:空 prompt 提交 → 内联红字
  await page.getByRole('button', { name: '← 返回' }).click();
  await page.getByRole('button', { name: '新建' }).click();
  await page.getByRole('dialog').waitFor();
  await shot('new-task-dialog');
  await page.getByRole('dialog').getByRole('button', { name: '提交' }).click();
  await page.locator('.field-error[data-field="prompt"]').waitFor();
  await page.getByRole('dialog').getByLabel('视频').check();
  await page.locator('.field-error').first().waitFor({ state: 'detached' }).catch(() => {});
  await page.getByRole('dialog').getByRole('button', { name: '提交' }).click();
  await page.locator('.field-error[data-field="prompt"]').waitFor();
  await shot('new-task-dialog-video-field-error');
  const videoRatios = await page.locator('#nt-ratio option').allInnerTexts();

  return { results, listRows, realRowTitle, liveRow, imgSize, timeline, waitButtons, videoRatios };
}
