import test from 'node:test';
import assert from 'node:assert/strict';

import { cancelUpscaleJob, getUpscaleConfig, getUpscaleJob, submitUpscaleJob } from './upscale.js';

test('读取超分配置时缺字段回落到未配置', async () => {
  const config = await getUpscaleConfig(async () => ({}));
  assert.deepEqual(config, { configured: false, resolutions: [] });
  const configured = await getUpscaleConfig(async () => ({ configured: true, resolutions: [{ value: '2k', label: '2K' }] }));
  assert.equal(configured.configured, true);
  assert.deepEqual(configured.resolutions, [{ value: '2k', label: '2K' }]);
});

test('提交超分任务使用 multipart 并带上文件与分辨率', async () => {
  let captured = null;
  const request = async (path, options) => {
    captured = { path, options };
    return { job: { id: 'upscale-1', status: 'running' } };
  };
  const file = new File(['video-bytes'], 'demo.mp4', { type: 'video/mp4' });
  const job = await submitUpscaleJob({ file, resolution: '2k' }, request);

  assert.equal(job.id, 'upscale-1');
  assert.equal(captured.path, '/api/ai-upscale/jobs');
  assert.equal(captured.options.method, 'POST');
  assert.equal(captured.options.timeoutMs, 0);
  assert.equal(captured.options.body.get('resolution'), '2k');
  const uploaded = captured.options.body.get('file');
  assert.equal(uploaded.name, 'demo.mp4');
  assert.equal(uploaded.type, 'video/mp4');
});

test('提交超分任务缺少文件或后端未返回任务时明确报错', async () => {
  await assert.rejects(() => submitUpscaleJob({ file: null, resolution: '1080p' }, async () => ({})), /请选择要超分的视频文件/);
  const file = new File(['x'], 'demo.mp4', { type: 'video/mp4' });
  await assert.rejects(() => submitUpscaleJob({ file, resolution: '1080p' }, async () => ({})), /提交超分任务失败/);
});

test('查询与取消超分任务返回后端任务状态', async () => {
  const job = await getUpscaleJob('upscale-1', async () => ({ job: { id: 'upscale-1', status: 'completed' }, warning: '' }));
  assert.equal(job.job.status, 'completed');
  const cancelled = await cancelUpscaleJob('upscale-1', async (path, options) => {
    assert.equal(path, '/api/ai-upscale/jobs/upscale-1/cancel');
    assert.equal(options.method, 'POST');
    return { job: { id: 'upscale-1', status: 'cancelled' } };
  });
  assert.equal(cancelled.status, 'cancelled');
  await assert.rejects(() => cancelUpscaleJob('upscale-1', async () => ({})), /取消超分任务失败/);
});
