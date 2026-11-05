import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryPreflightModal.jsx'), 'utf8');

test('preflight panel shows passed/blocked counts and never submits when every book is blocked', () => {
  assert.match(source, /质检完成：通过 \$\{passedCount\} 本 \/ 阻断 \$\{blockedCount\} 本/);
  assert.match(source, /const allBlocked = books\.length > 0 && passedBooks\.length === 0/);
  assert.match(source, /const canRun = !batchIssues\.length && passedBooks\.length > 0/);
  assert.match(source, /disabled=\{loading \|\| !canRun\}/);
});

test('partial blockage defaults to running only passed books and names the skipped books', () => {
  assert.match(source, /只跑通过的 \$\{passedCount\} 本（\$\{blockedCount\} 本不会提交）/);
  assert.match(source, /这 \$\{blockedCount\} 本不会提交：\$\{blockedBooks\.map\(book => book\.title \|\| book\.bookId\)\.join\('、'\)\}/);
});

test('all-blocked state renders a repair checklist and an explicit cancel-everything action', () => {
  assert.match(source, /先修这几项：/);
  assert.match(source, /全部取消，先修配置/);
  assert.match(source, /const fixList = \[\.\.\.batchIssues, \.\.\.blockedBooks\.flatMap\(book => book\.blocking \|\| \[\]\)\]/);
});

test('batch-level issues disable submit with an explicit label', () => {
  assert.match(source, /batchIssues\.length\s*\n?\s*\? '先修复整批配置问题'/);
  assert.match(source, /整批配置问题/);
});

test('single-book overrides warn that batch settings will not affect that book', () => {
  assert.match(source, /有单书覆盖/);
  assert.match(source, /改统一配置对这本无效，请到「单书设置」修改/);
});

test('an active batch job is shown on open and disables submit', () => {
  assert.match(source, /const taskState = result\?\.taskState \|\| ''/);
  assert.match(source, /const taskActive = Boolean\(taskState\)/);
  assert.match(source, /当前批次已有任务在进行（\$\{taskStateLabel[^}]*\}）/);
  assert.match(source, /const canRun = !batchIssues\.length && passedBooks\.length > 0 && !taskActive/);
  assert.match(source, /disabled=\{loading \|\| !canRun\}/);
  assert.match(source, /停止当前任务/);
});

test('panel states that preflight is read-only and offers a jump per blocking item', () => {
  assert.match(source, /预检只读取配置，不会创建任务、不会修改批次或书籍设置/);
  assert.match(source, /去修复：\{fixLabel\(issue\)\}/);
});
