import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.dirname(fileURLToPath(import.meta.url));
const layout = fs.readFileSync(path.join(root, 'UserLayout.jsx'), 'utf8');
const novelFetch = fs.readFileSync(path.join(root, '..', '..', 'user', 'pages', 'NovelFetchPage.jsx'), 'utf8');
const statusCenter = fs.readFileSync(path.join(root, '..', 'notifications', 'globalTaskCenter.js'), 'utf8');

test('顶栏不再渲染页面标题，只保留全局状态区', () => {
  assert.doesNotMatch(layout, /legacy-page-title/);
  assert.match(layout, /legacy-global-status/);
  assert.match(layout, /className="legacy-userbar"/);
});

test('没有真实全局状态时不渲染顶栏', () => {
  assert.match(layout, /!isAgentCanvasRoute && globalStatus\.text \? <header/);
  assert.doesNotMatch(layout, /!isScriptRoute \|\| globalStatus\.text/);
});

test('空状态文本可以主动清空顶栏，避免保留过期队列文案', () => {
  assert.match(layout, /if \(!detail\.text\) \{/);
  assert.match(layout, /setGlobalStatus\(\{ text: '', tone: 'idle', expiresAt: 0 \}\)/);
  assert.match(statusCenter, /durationMs/);
});

test('小说获取顶栏状态来自真实队列而不是本地表单标志', () => {
  assert.doesNotMatch(novelFetch, /暂无正在处理的任务/);
  assert.match(novelFetch, /const running = Number\(queueSummary\.running \|\| 0\)/);
  assert.match(novelFetch, /排队 \$\{queued\} 个/);
  assert.match(novelFetch, /\[queueSummary\.running, queueSummary\.queued\]/);
});
