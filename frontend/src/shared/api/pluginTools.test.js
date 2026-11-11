import test from 'node:test';
import assert from 'node:assert/strict';
import { analyzeAISegment, exportAISegment, getPluginToolConfig, runPluginTool } from './pluginTools.js';

test('网页文字插件只提交工具ID和用户内容给后端', async () => {
  let received;
  const result = await runPluginTool({ toolId: 'ai-rewrite', input: '原始文本', modelId: 'text-1', promptId: 'default' }, async (path, options) => {
    received = { path, options };
    return { result: '改写结果' };
  });
  assert.equal(received.path, '/api/plugins/ai-rewrite/run');
  assert.deepEqual(JSON.parse(received.options.body), { input: '原始文本', modelId: 'text-1', promptId: 'default' });
  assert.equal(result, '改写结果');
});

test('网页文字插件加载后端提示词名称但不读取提示词正文', async () => {
  const prompts = await getPluginToolConfig('ai-rewrite', async path => {
    assert.equal(path, '/api/plugins/ai-rewrite/config');
    return { prompts: [{ id: 'default', name: '短篇/长篇改文' }] };
  });
  assert.deepEqual(prompts, [{ id: 'default', name: '短篇/长篇改文' }]);
});

test('网页文字插件拒绝不合法工具ID', async () => {
  await assert.rejects(() => runPluginTool({ toolId: '../bad', input: '文本' }), /插件工具无效/);
});

test('AI分段分析只提交轨道规则和用户内容', async () => {
  let received;
  const result = await analyzeAISegment({ mode: 'ai', modelId: 'text-1', script: '剧本', clips: [{ index: 1, startUs: 0, endUs: 60000000 }] }, async (path, options) => {
    received = { path, options };
    return { episodes: [{ index: 1, startClip: 1, endClip: 1 }] };
  });
  assert.equal(received.path, '/api/plugins/ai-segment/analyze');
  assert.equal(JSON.parse(received.options.body).script, '剧本');
  assert.equal(result.episodes[0].endClip, 1);
  assert.doesNotMatch(received.options.body, /SystemPrompt|你负责/);
});

test('AI分段导出上传视频和已确认分集并返回压缩包', async () => {
  let received;
  const video = new Blob(['video'], { type: 'video/mp4' });
  const result = await exportAISegment({ video, episodes: [{ startSeconds: 0, endSeconds: 10 }], resolution: '1080p', fps: 30 }, async (path, options) => {
    received = { path, options };
    return new Blob(['zip'], { type: 'application/zip' });
  });
  assert.equal(received.path, '/api/plugins/ai-segment/export');
  assert.ok(received.options.body instanceof FormData);
  assert.equal(received.options.responseType, 'blob');
  assert.equal(JSON.parse(received.options.body.get('episodes'))[0].endSeconds, 10);
  assert.equal(result.type, 'application/zip');
});
