import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'App.jsx'), 'utf8');

test('用户端保留公网 V78 的历史和问题日志页面路由', () => {
  assert.match(source, /HistoryPage/);
  assert.match(source, /IssueLogPage/);
  assert.match(source, /['"]\/history['"]/);
  assert.match(source, /['"]\/issues['"]/);
});
