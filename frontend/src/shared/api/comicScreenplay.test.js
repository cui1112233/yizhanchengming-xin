import assert from 'node:assert/strict';
import test from 'node:test';
import { cancelComicScreenplayBatch, createComicScreenplay, createComicScreenplayAsset, createComicScreenplayShot, createComicStoryboardShot, deleteComicScreenplayAsset, deleteComicScreenplayShot, deleteComicStoryboardShot, exportComicScreenplay, generateComicScreenplay, listComicScreenplayPrompts, mergeComicStoryboardShot, moveComicScreenplayShot, restoreComicScreenplayDocument, runComicScreenplayBatchNext, saveComicScreenplayDocument, saveComicScreenplaySource, splitComicStoryboardShot, startComicScreenplayBatch, updateComicScreenplay, updateComicScreenplayAsset, updateComicScreenplayShot, updateComicStoryboardShot } from './comicScreenplay.js';

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

test('生产任务启动、执行下一步、暂停和导出都调用项目专属接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => { calls.push([path, options.method || 'GET']); return {}; };
  await startComicScreenplayBatch('comic-1', request);
  await runComicScreenplayBatchNext('comic-1', undefined, request);
  await cancelComicScreenplayBatch('comic-1', request);
  await exportComicScreenplay('comic-1', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/batch', 'POST'],
    ['/api/comic-screenplays/projects/comic-1/batch/run-next', 'POST'],
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

test('剧本手工保存和版本恢复调用真实版本接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push([path, options.method || 'GET', options.body ? JSON.parse(options.body) : null]);
    return { document: { id: 'doc-2', version: 2 } };
  };
  await saveComicScreenplayDocument('comic-1', '修订后的剧本', request);
  await restoreComicScreenplayDocument('comic-1', 'doc-1', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/document', 'PUT', { content: '修订后的剧本' }],
    ['/api/comic-screenplays/projects/comic-1/documents/doc-1/restore', 'POST', null],
  ]);
});

test('剧本手工保存拒绝空内容', async () => {
  await assert.rejects(() => saveComicScreenplayDocument('comic-1', '  ', async () => ({})), /剧本内容不能为空/);
});

test('分镜条目新增、编辑、拆分、合并和删除调用真实接口', async () => {
  const calls = [];
  const request = async (path, options = {}) => {
    calls.push([path, options.method || 'GET', options.body ? JSON.parse(options.body) : null]);
    return { storyboard: { id: 'storyboard-1', shots: [] } };
  };
  const input = { afterShotId: 'sb-shot-1', durationSec: 3, characters: '小Y', scene: '门口', props: '钥匙', videoDesc: '新增分镜' };
  await createComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, input, request);
  await updateComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, 'sb-shot-2', { ...input, videoDesc: '编辑分镜' }, request);
  await splitComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, 'sb-shot-2', request);
  await mergeComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, 'sb-shot-2', request);
  await deleteComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, 'sb-shot-2', request);
  assert.deepEqual(calls, [
    ['/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/1/shots', 'POST', input],
    ['/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/1/shots/sb-shot-2', 'PATCH', { ...input, videoDesc: '编辑分镜' }],
    ['/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/1/shots/sb-shot-2/split', 'POST', {}],
    ['/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/1/shots/sb-shot-2/merge-next', 'POST', {}],
    ['/api/comic-screenplays/projects/comic-1/storyboards/xiaoluo_v7_15s/1/shots/sb-shot-2', 'DELETE', null],
  ]);
});

test('分镜条目拒绝空内容、无效时长和无效路径参数', async () => {
  await assert.rejects(() => createComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, { durationSec: 0, videoDesc: '内容' }, async () => ({})), /时长/);
  await assert.rejects(() => createComicStoryboardShot('comic-1', 'xiaoluo_v7_15s', 1, { durationSec: 1, videoDesc: ' ' }, async () => ({})), /内容不能为空/);
  await assert.rejects(() => splitComicStoryboardShot('comic-1', 'bad engine', 1, 'sb-shot-1', async () => ({})), /分镜引擎无效/);
});
