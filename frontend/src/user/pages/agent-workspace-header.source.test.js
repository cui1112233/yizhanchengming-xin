import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));

function read(relativePath) {
  return fs.readFileSync(path.resolve(dirname, relativePath), 'utf8');
}

test('共享头部提供工作流和故事板视图', () => {
  const header = read('../components/AgentWorkspaceHeader.jsx');
  assert.match(header, /工作流/);
  assert.match(header, /故事板/);
  assert.match(header, /onViewChange/);
  assert.match(header, /onToggleDrawer/);
  assert.match(header, /onToggleConversation/);
});

test('切换视图保留任务编号和其他地址参数', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /new URLSearchParams\(window\.location\.search\)/);
  assert.match(page, /searchParams\.set\('view', view\)/);
  assert.match(page, /searchParams\.set\('taskId', taskId\)/);
  assert.match(page, /searchParams\.delete\('taskId'\)/);
});

test('共享头部状态从画布读取并进入现有保存队列', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /normalizeWorkspaceView\(loadedCanvas\.viewState\)/);
  assert.match(page, /viewState:\s*nextViewState/);
  assert.match(page, /canvasSaveQueueRef\.current\?\.enqueue\(nextCanvas\)/);
});
