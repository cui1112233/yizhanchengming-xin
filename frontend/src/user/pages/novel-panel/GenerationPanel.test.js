import test from 'node:test';
import assert from 'node:assert/strict';
import { beginGeneration, cancelGeneration, createGenerationState, finishGeneration } from './generationState.js';

test('does not start a second generation while one is running', () => {
  const running = beginGeneration(createGenerationState());
  assert.equal(beginGeneration(running), running);
  assert.equal(running.status, 'running');
});

test('cancelled generation ignores a late completion', () => {
  const running = beginGeneration(createGenerationState());
  const cancelled = cancelGeneration(running);
  assert.equal(finishGeneration(cancelled, { video: 'late' }).status, 'cancelled');
});

test('failed generation remains retryable with its error', () => {
  const failed = finishGeneration(beginGeneration(createGenerationState()), null, { message: '生成失败', retryable: true });
  assert.equal(failed.status, 'failed');
  assert.equal(failed.retryable, true);
  assert.equal(failed.error.message, '生成失败');
});
