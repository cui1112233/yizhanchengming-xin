import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const start = source.indexOf('async function runSelectedProduction()');
const selected = source.slice(start, source.indexOf('function mergePayload(', start));

test('selected execution keeps one client request id for the request and polls canonical status', () => {
  assert.match(selected, /crypto\.randomUUID\(\)/);
  assert.match(selected, /clientRequestId/);
  assert.match(selected, /getBatchExecution\(executionId\)/);
  assert.match(selected, /completed.*needs_attention.*cancelled/s);
});

test('selected execution submits only preflight runnable books and reports blocked books', () => {
  assert.match(selected, /preflightBatchAutomation/);
  assert.match(selected, /runnableBookIds/);
  assert.match(selected, /blockedBooks/);
});
