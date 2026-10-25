import test from 'node:test';
import assert from 'node:assert/strict';
import { createDirectorSaveQueue } from './director-save-queue.js';

test('同一轮待保存内容合并为最新状态', async () => {
  const writes = [];
  const queue = createDirectorSaveQueue(value => writes.push(value));
  queue.schedule({ revision: 1 });
  queue.schedule({ revision: 2 });
  await queue.flush();
  assert.deepEqual(writes, [{ revision: 2 }]);
  assert.equal(queue.getStatus(), 'saved');
});

test('flush 等待正在写入和随后更新的状态，写入不并发', async () => {
  const writes = [];
  let finishFirst;
  let signalStarted;
  const started = new Promise(resolve => { signalStarted = resolve; });
  const queue = createDirectorSaveQueue(async value => {
    writes.push(value.revision);
    if (value.revision === 1) await new Promise(resolve => { finishFirst = resolve; signalStarted(); });
  });
  queue.schedule({ revision: 1 });
  await started;
  queue.schedule({ revision: 2 });
  const flushed = queue.flush();
  assert.deepEqual(writes, [1]);
  finishFirst();
  await flushed;
  assert.deepEqual(writes, [1, 2]);
  assert.equal(queue.getStatus(), 'saved');
});

test('失败后 flush 重试失败内容', async () => {
  let fail = true;
  const writes = [];
  const queue = createDirectorSaveQueue(async value => {
    writes.push(value.revision);
    if (fail) throw new Error('保存失败');
  });
  queue.schedule({ revision: 1 });
  await assert.rejects(queue.flush(), /保存失败/);
  assert.equal(queue.getStatus(), 'error');
  fail = false;
  await queue.flush();
  assert.deepEqual(writes, [1, 1]);
  assert.equal(queue.getStatus(), 'saved');
  assert.equal(queue.getError(), null);
});

test('旧写入失败不能覆盖较新的待保存状态，flush 可重试最新状态', async () => {
  const writes = [];
  let rejectFirst;
  let signalStarted;
  const started = new Promise(resolve => { signalStarted = resolve; });
  const queue = createDirectorSaveQueue(async value => {
    writes.push(value.revision);
    if (value.revision === 1) await new Promise((resolve, reject) => { rejectFirst = reject; signalStarted(); });
  });
  queue.schedule({ revision: 1 });
  await started;
  queue.schedule({ revision: 2 });
  const firstFlush = queue.flush();
  rejectFirst(new Error('旧内容保存失败'));
  await assert.rejects(firstFlush, /旧内容保存失败/);
  assert.equal(queue.getStatus(), 'error');
  await queue.flush();
  assert.deepEqual(writes, [1, 2]);
  assert.equal(queue.getStatus(), 'saved');
});

test('错误状态回调中安排的新内容会自动写入，旧失败内容不会自动重试', async () => {
  const writes = [];
  let queue;
  queue = createDirectorSaveQueue(async value => {
    writes.push(value.revision);
    if (value.revision === 1) throw new Error('保存失败');
  }, { onStatus: status => {
    if (status === 'error') queue.schedule({ revision: 2 });
  } });
  queue.schedule({ revision: 1 });
  await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(writes, [1, 2]);
  assert.equal(queue.getStatus(), 'saved');
});

test('没有新内容时失败不会自动重复写入', async () => {
  let attempts = 0;
  const queue = createDirectorSaveQueue(async () => {
    attempts += 1;
    throw new Error('保存失败');
  });
  queue.schedule({ revision: 1 });
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(attempts, 1);
  assert.equal(queue.getStatus(), 'error');
});
