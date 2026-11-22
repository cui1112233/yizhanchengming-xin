import assert from 'node:assert/strict';
import test from 'node:test';

import { executionViewModel } from './executionStatus.js';

const expected = {
  running: '执行中', reconciling: '核对中', retry_wait: '等待重试',
  needs_attention: '需要处理', completed: '已完成', cancelled: '已取消'
};

for (const [status, label] of Object.entries(expected)) {
  test(`execution status ${status} has a Chinese label`, () => {
    const view = executionViewModel({ id: 'execution-1', status });
    assert.equal(view.label, label);
    assert.equal(view.id, 'execution-1');
    assert.equal(typeof view.terminal, 'boolean');
  });
}
