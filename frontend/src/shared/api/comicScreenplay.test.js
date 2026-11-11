import assert from 'node:assert/strict';
import test from 'node:test';
import { cancelComicScreenplayBatch, createComicScreenplay, createComicScreenplayAsset, createComicScreenplayShot, deleteComicScreenplayAsset, deleteComicScreenplayShot, exportComicScreenplay, generateComicScreenplay, listComicScreenplayPrompts, moveComicScreenplayShot, saveComicScreenplaySource, startComicScreenplayBatch, updateComicScreenplay, updateComicScreenplayAsset, updateComicScreenplayShot } from './comicScreenplay.js';

test('项目设置和原文都调用已有项目接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push([path, options.method || 'GET', options.body ? JSON.parse(options.body) : null]);
    return { project: { id: 'comic-1' } };
  };
  await updateComicScreenplay('comic-1', { modelId: 'text-model', expectedRevision: 3 }, request);
  await saveComicScreenplaySource('comic-1', '原文正文', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1', 'PATCH', { modelId: 'text-model', expectedRevision: 3 }],
    ['/api/comic-screenplays/projects/comic-1/source', 'PUT', { content: '原文正文' }],
  ]);
});

test('漫剧创作项目拒绝空名称', async () => {
  await assert.rejects(() => createComicScreenplay({ title: ' ' }, async () => ({})), /作品名称不能为空/);
});

test('漫剧生成提交真实内容并强制显式确认', async () => {
  let captured;
  await generateComicScreenplay('comic-1', { mode: 'a1', content: '原文' }, async (path, options) => { captured = { path, body: JSON.parse(options.body) }; return {}; });
  assert.equal(captured.path, '/api/comic-screenplays/projects/comic-1/generate');
  assert.equal(captured.body.content, '原文');
  assert.equal(captured.body.confirmed, true);
});

test('生产任务启动、暂停和导出都调用项目专属接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => { calls.push([path, options.method || 'GET']); return {}; };
  await startComicScreenplayBatch('comic-1', request);
  await cancelComicScreenplayBatch('comic-1', request);
  await exportComicScreenplay('comic-1', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/batch', 'POST'],
    ['/api/comic-screenplays/projects/comic-1/batch/cancel', 'POST'],
    ['/api/comic-screenplays/projects/comic-1/export', 'GET'],
  ]);
});

test('新建弹窗提示词目录来自后端且不包含提示词正文', async () => {
  const rows = await listComicScreenplayPrompts(async path => {
    assert.equal(path, '/api/comic-screenplays/prompts');
    return { prompts: [{ id: 'b_director_screenplay', name: '专业剧本', description: '导演级结构' }] };
  });
  assert.equal(rows[0].name, '专业剧本');
  assert.equal(Object.hasOwn(rows[0], 'body'), false);
});

test('镜头新增、编辑、移动和删除都调用真实项目接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push([path, options.method || 'GET', options.body ? JSON.parse(options.body) : null]);
    return { shot: { id: 'shot-1' }, shots: [] };
  };
  await createComicScreenplayShot('comic-1', { episodeNo: 1, heading: '1-2', scriptText: '新增镜头' }, request);
  await updateComicScreenplayShot('comic-1', 'shot-1', { episodeNo: 1, heading: '1-2', scriptText: '编辑镜头' }, request);
  await moveComicScreenplayShot('comic-1', 'shot-1', 'up', request);
  await deleteComicScreenplayShot('comic-1', 'shot-1', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/shots', 'POST', { episodeNo: 1, heading: '1-2', scriptText: '新增镜头' }],
    ['/api/comic-screenplays/projects/comic-1/shots/shot-1', 'PATCH', { episodeNo: 1, heading: '1-2', scriptText: '编辑镜头' }],
    ['/api/comic-screenplays/projects/comic-1/shots/shot-1/move', 'POST', { direction: 'up' }],
    ['/api/comic-screenplays/projects/comic-1/shots/shot-1', 'DELETE', null],
  ]);
});

test('镜头拒绝空内容和无效移动方向', async () => {
  await assert.rejects(() => createComicScreenplayShot('comic-1', { episodeNo: 1, heading: ' ', scriptText: ' ' }, async () => ({})), /不能同时为空/);
  await assert.rejects(() => moveComicScreenplayShot('comic-1', 'shot-1', 'left', async () => ({})), /移动方向/);
});

test('资产新增、编辑和删除调用真实项目接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push([path, options.method || 'GET', options.body ? JSON.parse(options.body) : null]);
    return { asset: { id: 'asset-1', revision: 1 } };
  };
  await createComicScreenplayAsset('comic-1', { type: 'character', name: '小Y' }, request);
  await updateComicScreenplayAsset('comic-1', 'asset-1', { type: 'character', name: '小Y', expectedRevision: 1 }, request);
  await deleteComicScreenplayAsset('comic-1', 'asset-1', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/assets', 'POST', { type: 'character', name: '小Y' }],
    ['/api/comic-screenplays/projects/comic-1/assets/asset-1', 'PATCH', { type: 'character', name: '小Y', expectedRevision: 1 }],
    ['/api/comic-screenplays/projects/comic-1/assets/asset-1', 'DELETE', null],
  ]);
});

test('资产拒绝空名称和无效类型', async () => {
  await assert.rejects(() => createComicScreenplayAsset('comic-1', { type: 'scene', name: ' ' }, async () => ({})), /名称不能为空/);
  await assert.rejects(() => createComicScreenplayAsset('comic-1', { type: 'video', name: '片段' }, async () => ({})), /类型无效/);
});
