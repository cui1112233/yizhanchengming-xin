import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));
const novelList = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

// 只截取 pollComicPromptBook 的函数体做结构断言：函数外的同名调用不参与判定。
function pollFunctionSource() {
  const found = novelList.match(/const pollComicPromptBook = async \(\) => \{[\s\S]*?\n    \};/);
  assert.ok(found, 'BatchFactoryNovelList.jsx 必须包含 pollComicPromptBook');
  return found[0];
}

const source = pollFunctionSource();

test('页面隐藏时不发起任何漫画轮询请求', () => {
  const hiddenGuard = "if (typeof document !== 'undefined' && document.hidden) return;";
  const hiddenIndex = source.indexOf(hiddenGuard);
  assert.notEqual(hiddenIndex, -1, '缺少 document.hidden 提前返回：隐藏标签页仍会拉数据');
  // 守卫必须排在所有网络请求之前：隐藏页面既不拉 runtime-summary，也不拉整批。
  for (const call of ['getBatchRuntimeSummary(', 'getBatch(']) {
    const callIndex = source.indexOf(call);
    assert.notEqual(callIndex, -1, `pollComicPromptBook 缺少 ${call}`);
    assert.ok(hiddenIndex < callIndex, `${call} 必须排在 document.hidden 守卫之后，否则隐藏页面仍会发起请求`);
  }
});

test('comic run 的 status|时间戳未变化时不拉整批，变化时才拉一次并更新 lastComicRunStamp', () => {
  const stampLine = "const stamp = comicRun ? `${comicRun.status}|${comicRun.updatedAt || comicRun.createdAt || ''}` : '';";
  assert.ok(source.includes(stampLine), 'stamp 必须是 status|updatedAt(或 createdAt)，否则比对没有意义');

  const gate = 'if (!stamp || stamp === lastComicRunStamp) return;';
  assert.notEqual(source.indexOf(gate), -1, '缺少 stamp === lastComicRunStamp 提前返回：每次轮询都会拉整批');
  assert.equal((source.match(/lastComicRunStamp = stamp;/g) || []).length, 1, '变化时 lastComicRunStamp 必须且只能更新一次');
  assert.equal((source.match(/await getBatch\(/g) || []).length, 1, '整批 getBatch 每次轮询最多只能拉一次');

  // 顺序即行为：算 stamp -> 一致则 return -> 更新 stamp -> 才允许拉整批。
  const stampIndex = source.indexOf('const stamp =');
  const gateIndex = source.indexOf(gate);
  const assignIndex = source.indexOf('lastComicRunStamp = stamp;');
  const batchIndex = source.indexOf('const response = await getBatch(');
  assert.ok(stampIndex < gateIndex, 'stamp 必须先算出来再比对');
  assert.ok(gateIndex < assignIndex, '必须在一致性检查之后再更新 lastComicRunStamp');
  assert.ok(assignIndex < batchIndex, '只有 stamp 变化后才允许拉整批');
});

test('轮询间隔固定为 4000ms，禁止回落到 1500ms', () => {
  // setInterval 在函数体之外，这里对整份源码断言。
  assert.match(novelList, /const timer = setInterval\(pollComicPromptBook, 4000\);/, '轮询间隔必须是 4000ms');
  assert.doesNotMatch(novelList, /setInterval\(pollComicPromptBook, 1500\)/, '不允许恢复 1500ms 整批轮询');
});
