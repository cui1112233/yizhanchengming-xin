import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const page = fs.readFileSync(new URL('./AICutPage.jsx', import.meta.url), 'utf8');
const panel = fs.readFileSync(new URL('./aicut/ExportQueuePanel.jsx', import.meta.url), 'utf8');

test('批量导出使用持久队列列表而不是单个展示进度', () => {
  assert.match(page, /listAICutExportJobs/);
  assert.match(page, /setExportJobs/);
  assert.match(page, /ExportQueuePanel/);
  assert.match(page, /成片 \+ MP3 加入队列/);
  assert.doesNotMatch(page, /setExportJob\(/);
});

test('队列任务提供真实取消、下载和失败重试操作', () => {
  assert.match(panel, /onCancel\?\.\(job\)/);
  assert.match(panel, /onDownload\?\.\(job\)/);
  assert.match(panel, /onRetry\?\.\(job\)/);
  assert.match(panel, /MySQL 保存任务与时间线快照/);
  assert.match(panel, /Redis 按顺序唤醒服务器渲染/);
});
