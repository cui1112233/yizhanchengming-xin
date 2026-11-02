import test from 'node:test';
import assert from 'node:assert/strict';
import { giantMaterialOcrStatus } from './giantMaterialOcrStatus.js';

test('queued OCR work with no executor is shown as waiting, not reading', () => {
  assert.deepEqual(giantMaterialOcrStatus({ state: 'queued', waitingForExecutor: true }), {
    title: '等待 OCR 执行器上线',
    detail: '正文尚未开始读取。执行器上线后会自动领取任务。',
    progressStatus: 'normal'
  });
});

test('claimed OCR work is shown as active reading progress', () => {
  assert.deepEqual(giantMaterialOcrStatus({ state: 'running', progress: { completed: 12, total: 30, percent: 40 } }), {
    title: '滚屏 OCR',
    detail: '正在读取正文 40% · 视频 12/30 秒',
    progressStatus: 'active'
  });
});
