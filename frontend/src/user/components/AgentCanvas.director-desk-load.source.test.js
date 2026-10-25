import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));

test('导演台动态模块失败时显示错误和重新加载入口', () => {
  const canvas = fs.readFileSync(path.join(dirname, 'AgentCanvas.jsx'), 'utf8');
  const boundary = fs.readFileSync(path.join(dirname, 'director-desk/DirectorDeskLoadBoundary.jsx'), 'utf8');
  assert.match(canvas, /DirectorDeskLoadBoundary/);
  assert.match(boundary, /导演台资源加载失败/);
  assert.match(boundary, /重新加载页面/);
  assert.match(boundary, /window\.location\.reload/);
});
