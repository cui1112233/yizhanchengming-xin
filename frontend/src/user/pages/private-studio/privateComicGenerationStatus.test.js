import assert from 'node:assert/strict';
import test from 'node:test';

import { privateComicGenerationStatus } from './privateComicGenerationStatus.js';

test('没有提示词时分别显示未生成和等待提示词', () => {
  assert.deepEqual(privateComicGenerationStatus({ assetCount: 5 }), {
    prompts: {
      key: 'prompts',
      label: '未生成',
      detail: '0 页',
      tone: 'idle',
      running: false,
      canRetry: false
    },
    images: {
      key: 'images',
      label: '等待提示词',
      detail: '提示词完成后可生成',
      tone: 'waiting',
      running: false,
      canRetry: false
    }
  });
});

test('提示词区分生成中、部分完成、失败和已完成', () => {
  const base = { assetCount: 5, pagesTotal: 20, promptReadyCount: 8 };

  assert.deepEqual(privateComicGenerationStatus({ ...base, action: 'prompts' }).prompts, {
    key: 'prompts', label: '生成中', detail: '已完成 8/20 页', tone: 'running', running: true, canRetry: false
  });
  assert.deepEqual(privateComicGenerationStatus(base).prompts, {
    key: 'prompts', label: '部分完成', detail: '已完成 8/20 页', tone: 'partial', running: false, canRetry: true
  });
  assert.deepEqual(privateComicGenerationStatus({ ...base, promptFailed: true }).prompts, {
    key: 'prompts', label: '生成失败', detail: '已完成 8/20 页', tone: 'failed', running: false, canRetry: true
  });
  assert.deepEqual(privateComicGenerationStatus({ ...base, promptReadyCount: 20 }).prompts, {
    key: 'prompts', label: '已完成', detail: '已完成 20/20 页', tone: 'complete', running: false, canRetry: false
  });
});

test('图片区分未生成、生成中、部分完成、失败和已完成', () => {
  const ready = { assetCount: 5, pagesTotal: 20, promptReadyCount: 20 };

  assert.deepEqual(privateComicGenerationStatus(ready).images, {
    key: 'images', label: '未生成', detail: '已完成 0/20 页', tone: 'idle', running: false, canRetry: false
  });
  assert.deepEqual(privateComicGenerationStatus({ ...ready, imageReadyCount: 6, action: 'pages' }).images, {
    key: 'images', label: '生成中', detail: '已完成 6/20 页', tone: 'running', running: true, canRetry: false
  });
  assert.deepEqual(privateComicGenerationStatus({ ...ready, imageReadyCount: 6 }).images, {
    key: 'images', label: '部分完成', detail: '已完成 6/20 页', tone: 'partial', running: false, canRetry: true
  });
  assert.deepEqual(privateComicGenerationStatus({ ...ready, imageReadyCount: 6, imageFailed: true }).images, {
    key: 'images', label: '生成失败', detail: '已完成 6/20 页', tone: 'failed', running: false, canRetry: true
  });
  assert.deepEqual(privateComicGenerationStatus({ ...ready, imageReadyCount: 20, imageFailed: true }).images, {
    key: 'images', label: '已完成', detail: '已完成 20/20 页', tone: 'complete', running: false, canRetry: false
  });
});

test('任务失败会归到当前尚未完成的阶段而不覆盖已完成阶段', () => {
  const promptFailure = privateComicGenerationStatus({ pagesTotal: 10, promptReadyCount: 3, taskFailed: true });
  assert.equal(promptFailure.prompts.label, '生成失败');
  assert.equal(promptFailure.images.label, '等待提示词');

  const failureWithoutAssets = privateComicGenerationStatus({ taskFailed: true });
  assert.equal(failureWithoutAssets.prompts.canRetry, false);

  const imageFailure = privateComicGenerationStatus({ pagesTotal: 10, promptReadyCount: 10, imageReadyCount: 4, taskFailed: true });
  assert.equal(imageFailure.prompts.label, '已完成');
  assert.equal(imageFailure.images.label, '生成失败');
});
