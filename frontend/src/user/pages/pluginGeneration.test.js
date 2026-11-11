import assert from 'node:assert/strict';
import test from 'node:test';
import {
  createGenerationGate,
  executeGeneration,
  isAbortError,
  validateGenerationInput
} from './pluginGeneration.js';

function abortableHold(signal, payload = 'x') {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      const error = new Error('已取消');
      error.name = 'AbortError';
      reject(error);
      return;
    }
    signal?.addEventListener('abort', () => {
      const error = new Error('已取消');
      error.name = 'AbortError';
      reject(error);
    }, { once: true });
    // 永不主动 resolve，只有取消才会结束。
    void payload;
    void resolve;
  });
}

test('正常：多段按顺序拼接并汇报进度', async () => {
  const progress = [];
  const outcome = await executeGeneration({
    parts: ['一', '二'],
    concurrent: false,
    runPart: async part => `结果${part}`,
    onProgress: (current, total) => progress.push(`${current}/${total}`)
  });
  assert.equal(outcome.ok, true);
  assert.equal(outcome.cancelled, false);
  assert.equal(outcome.text, '结果一\n\n结果二');
  assert.deepEqual(outcome.results, ['结果一', '结果二']);
  assert.deepEqual(progress, ['1/2', '2/2']);
});

test('空值：输入为空被拦截，空分段不发起生成', async () => {
  const blocked = validateGenerationInput({ input: '   ', modelId: 'm', promptId: 'p' });
  assert.equal(blocked.ok, false);
  assert.equal(blocked.message, '请输入或上传内容');
  const noModel = validateGenerationInput({ input: '正文', modelId: '', promptId: 'p' });
  assert.equal(noModel.message, '请选择模型');

  const empty = await executeGeneration({ parts: [], runPart: async () => 'x' });
  assert.equal(empty.ok, false);
  assert.equal(empty.cancelled, false);
  assert.equal(empty.error.message, '没有可生成的内容');
});

test('重复点击：同一时刻只允许一个生成，结束后可再次提交', async () => {
  const gate = createGenerationGate();
  const first = gate.begin();
  assert.ok(first instanceof AbortController);
  assert.equal(gate.begin(), null);
  assert.equal(gate.running, true);
  gate.end(first);
  assert.equal(gate.running, false);
  const second = gate.begin();
  assert.ok(second instanceof AbortController);
  assert.notEqual(second, first);
  gate.end(second);
});

test('串行部分失败：保留已完成段落和失败序号，供界面写回', async () => {
  const outcome = await executeGeneration({
    parts: ['一', '二', '三'],
    concurrent: false,
    runPart: async part => {
      if (part === '三') throw new Error('模型调用失败');
      return `结果${part}`;
    }
  });
  assert.equal(outcome.ok, false);
  assert.equal(outcome.cancelled, false);
  assert.deepEqual(outcome.results, ['结果一', '结果二']);
  assert.equal(outcome.text, '结果一\n\n结果二');
  assert.equal(outcome.failedIndex, 2);
});

test('并发生成部分失败：保留已成功分段，失败序号指向第一段失败', async () => {
  const outcome = await executeGeneration({
    parts: ['一', '二', '三'],
    concurrent: true,
    runPart: async (part, index) => {
      if (index === 1) throw new Error('第二段失败');
      return `结果${part}`;
    }
  });
  assert.equal(outcome.ok, false);
  assert.equal(outcome.cancelled, false);
  assert.deepEqual(outcome.results, ['结果一', '结果三']);
  assert.equal(outcome.failedIndex, 1);
});

test('真实失败：抛出错误，不标记为取消', async () => {
  const outcome = await executeGeneration({
    parts: ['一'],
    runPart: async () => { throw new Error('模型调用失败'); }
  });
  assert.equal(outcome.ok, false);
  assert.equal(outcome.cancelled, false);
  assert.equal(outcome.error.message, '模型调用失败');
  assert.equal(isAbortError(outcome.error), false);
});

test('用户取消：signal 真的 aborted，且不返回任何结果', async () => {
  const gate = createGenerationGate();
  const controller = gate.begin();
  const pending = executeGeneration({
    parts: ['一', '二'],
    concurrent: true,
    runPart: () => abortableHold(controller.signal),
    signal: controller.signal
  });
  // 让所有分段进入等待态后再取消。
  await Promise.resolve();
  const aborted = gate.abort();
  const outcome = await pending;
  assert.equal(aborted, true);
  assert.equal(controller.signal.aborted, true);
  assert.equal(gate.running, false);
  assert.equal(outcome.ok, false);
  assert.equal(outcome.cancelled, true);
  assert.deepEqual(outcome.results, []);
  assert.equal(outcome.text, '');
});
