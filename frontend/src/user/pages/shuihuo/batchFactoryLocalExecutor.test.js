import assert from 'node:assert/strict';
import test from 'node:test';
import { batchExecutorKinds, localExecutorTaskToast, usesYizhanshengmingExecutor } from './batchFactoryLocalExecutor.js';

test('recognizes the Yizhanshengming executor for comic images and videos', () => {
  assert.equal(usesYizhanshengmingExecutor({ imageProvider: 'doubao_local_executor' }), true);
  assert.equal(usesYizhanshengmingExecutor({ videoProvider: 'doubao_local_executor' }), true);
  assert.equal(usesYizhanshengmingExecutor({ imageProvider: 'other', videoProvider: 'other' }), false);
});

test('finds the model kinds using the executor across batch and book settings', () => {
  assert.deepEqual(batchExecutorKinds({ settingsState: { patch: { imageProvider: 'doubao_local_executor' } }, books: [] }), ['image']);
  assert.deepEqual(batchExecutorKinds({ settingsState: { patch: {} }, books: [{ settingsState: { patch: { videoProvider: 'doubao_local_executor' } } }] }), ['video']);
  assert.deepEqual(batchExecutorKinds({ settingsState: { patch: { imageProvider: 'doubao_local_executor' } }, books: [{ settingsState: { patch: { videoProvider: 'doubao_local_executor' } } }] }), ['image', 'video']);
});

test('maps local executor task changes to visible generation notices', () => {
  assert.deepEqual(localExecutorTaskToast({ status: 'queued' }), { type: 'loading', text: '一战晟铭执行器：任务已排队' });
  assert.deepEqual(localExecutorTaskToast({ status: 'running' }), { type: 'loading', text: '一战晟铭执行器：正在生成' });
  assert.deepEqual(localExecutorTaskToast({ status: 'succeeded' }), { type: 'success', text: '一战晟铭执行器：生成完成' });
  assert.deepEqual(localExecutorTaskToast({ status: 'failed', errorMessage: '服务无响应' }), { type: 'error', text: '一战晟铭执行器：生成失败：服务无响应' });
  assert.equal(localExecutorTaskToast({ status: 'unknown' }), null);
});
