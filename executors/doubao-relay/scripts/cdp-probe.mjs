// 开发工具:在真豆包页面里执行 JS,打印 JSON 结果。
// 用法: node scripts/cdp-probe.mjs [--port 9707] [--target chat|background] [--file path.js] [--key Escape] [--timeout 30000] '<expression>'
//   --file    先把文件内容注入(evaluate),再 evaluate 表达式。
//   --key     在 evaluate 表达式之前用 Input.dispatchKeyEvent 按一次该键(目前支持 Escape)。
//   --timeout 整体看门狗毫秒数(默认 30000),到时打印 timeout 并以 exit code 2 退出;单条 CDP 命令另有 15s 超时(与 spec 09 一致)。
// 需要 Node 22+(自带全局 WebSocket / fetch)。
const args = process.argv.slice(2);
const opt = (k, d) => {
  const i = args.indexOf(k);
  return i >= 0 ? args.splice(i, 2)[1] : d;
};
const port = opt('--port', '9707');
const target = opt('--target', 'chat');
const file = opt('--file', null);
const key = opt('--key', null);
const totalTimeout = Number(opt('--timeout', '30000'));
const expr = args[0] ?? 'document.title';
const CMD_TIMEOUT_MS = 15_000;

// 整体看门狗:豆包 WS 不响应 / 页面正在 reload / 表达式返回永不 resolve 的 Promise 时,脚本不能挂死。
const watchdog = setTimeout(() => {
  console.error(`timeout after ${totalTimeout}ms`);
  process.exit(2);
}, totalTimeout);

const withTimeout = (promise, ms, label) =>
  Promise.race([
    promise,
    new Promise((_, rej) => setTimeout(() => rej(new Error(`${label} timed out after ${ms}ms`)), ms).unref?.()),
  ]);

const list = await (await fetch(`http://127.0.0.1:${port}/json/list`, { signal: AbortSignal.timeout(5000) })).json();
const page = list.find((t) => t.type === 'page' && new URL(t.url).host === `doubao-${target}`);
if (!page) {
  console.error('target not found', list.map((t) => t.url));
  process.exit(1);
}

const ws = new WebSocket(page.webSocketDebuggerUrl);
let id = 0;
const pending = new Map();
const send = (method, params) =>
  withTimeout(
    new Promise((res, rej) => {
      const i = ++id;
      pending.set(i, { res, rej });
      ws.send(JSON.stringify({ id: i, method, params }));
    }),
    CMD_TIMEOUT_MS,
    method,
  );
ws.onmessage = (e) => {
  const m = JSON.parse(e.data);
  if (m.id && pending.has(m.id)) {
    const p = pending.get(m.id);
    pending.delete(m.id);
    m.error ? p.rej(m.error) : p.res(m.result);
  }
};
const failAll = (why) => {
  for (const p of pending.values()) p.rej(new Error(why));
  pending.clear();
};
ws.onerror = () => failAll('websocket error');
ws.onclose = () => failAll('websocket closed');
await withTimeout(
  new Promise((res, rej) => {
    ws.onopen = res;
    ws.onerror = (e) => rej(new Error(`websocket error: ${e?.message ?? 'connect failed'}`));
  }),
  CMD_TIMEOUT_MS,
  'websocket open',
);
ws.onerror = () => failAll('websocket error');

const evaluate = async (expression) => {
  const r = await send('Runtime.evaluate', { expression, returnByValue: true, awaitPromise: true });
  if (r.exceptionDetails) throw new Error(r.exceptionDetails.exception?.description ?? 'js exception');
  return r.result.value;
};

const KEYS = { Escape: { key: 'Escape', code: 'Escape', windowsVirtualKeyCode: 27, nativeVirtualKeyCode: 27 } };
const pressKey = async (name) => {
  const k = KEYS[name];
  if (!k) throw new Error(`unsupported key: ${name}`);
  await send('Input.dispatchKeyEvent', { type: 'keyDown', ...k });
  await send('Input.dispatchKeyEvent', { type: 'keyUp', ...k });
};

try {
  if (file) await evaluate(await (await import('node:fs/promises')).readFile(file, 'utf8'));
  if (key) await pressKey(key);
  console.log(JSON.stringify(await evaluate(expr), null, 2));
} finally {
  clearTimeout(watchdog);
  ws.close();
}
