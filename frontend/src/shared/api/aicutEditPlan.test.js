import test from 'node:test';
import assert from 'node:assert/strict';
import { applyAICutEditPlan, cancelAICutEditPlan, createAICutEditPlan, getAICutEditPlan, previewAICutEditPlan } from './aicutEditPlan.js';

test('AI自动剪辑发送可重试请求编号和完整审核参数', async () => {
  let received;
  const controller = new AbortController();
  const plan = await createAICutEditPlan('cut-1', {
    clientRequestId: 'request-1', modelId: 'text-main', expectedRevision: 3,
    assetIds: ['asset-1'], goal: '剪成一分钟竖屏视频', targetDurationMs: 60000,
    aspectRatio: '9:16', includeCaptions: true, includeTransitions: true,
    includeFilters: false, includeBgm: true, extraRequirement: '保留人物对白', signal: controller.signal
  }, async (path, options) => { received = { path, options }; return { plan: { id: 'plan-1', state: 'ready' } }; });
  assert.equal(plan.id, 'plan-1');
  assert.equal(received.path, '/api/ai-cut/projects/cut-1/edit-plans');
  assert.equal(received.options.signal, controller.signal);
  assert.equal(received.options.timeoutMs, 0);
  assert.deepEqual(JSON.parse(received.options.body), {
    clientRequestId: 'request-1', modelId: 'text-main', expectedRevision: 3, assetIds: ['asset-1'], goal: '剪成一分钟竖屏视频', targetDurationMs: 60000, aspectRatio: '9:16', includeCaptions: true, includeTransitions: true, includeFilters: false, includeBgm: true, extraRequirement: '保留人物对白'
  });
});

test('AI自动剪辑校验项目、模型、素材、目标、时长和比例', async () => {
  const request = async () => { throw new Error('不应发请求'); };
  const valid = { clientRequestId: 'request-1', modelId: 'text-main', assetIds: ['asset-1'], goal: '自动剪辑', targetDurationMs: 1000, aspectRatio: '9:16' };
  await assert.rejects(() => createAICutEditPlan('../cut', valid, request), /项目无效/);
  await assert.rejects(() => createAICutEditPlan('cut-1', { ...valid, modelId: '' }, request), /文本模型/);
  await assert.rejects(() => createAICutEditPlan('cut-1', { ...valid, assetIds: [] }, request), /素材/);
  await assert.rejects(() => createAICutEditPlan('cut-1', { ...valid, goal: '' }, request), /剪辑目标/);
  await assert.rejects(() => createAICutEditPlan('cut-1', { ...valid, targetDurationMs: 7200001 }, request), /目标时长/);
  await assert.rejects(() => createAICutEditPlan('cut-1', { ...valid, aspectRatio: '4:3' }, request), /画面比例/);
});

test('AI自动剪辑查询、取消和应用使用同一草案地址', async () => {
  const calls = [];
  const request = async (path, options = {}) => { calls.push({ path, options }); return path.endsWith('/apply') ? { timeline: { revision: 4 }, summary: { applied: 2 } } : { plan: { id: 'plan-1' } }; };
  await getAICutEditPlan('cut-1', 'plan-1', request);
  await cancelAICutEditPlan('cut-1', 'plan-1', request);
  const applied = await applyAICutEditPlan('cut-1', 'plan-1', ['op-1', 'op-2'], request);
  assert.deepEqual(calls.map(item => item.path), ['/api/ai-cut/projects/cut-1/edit-plans/plan-1', '/api/ai-cut/projects/cut-1/edit-plans/plan-1/cancel', '/api/ai-cut/projects/cut-1/edit-plans/plan-1/apply']);
  assert.deepEqual(JSON.parse(calls[2].options.body), { enabledOperationIds: ['op-1', 'op-2'] });
  assert.equal(applied.timeline.revision, 4);
});

test('审核按当前勾选请求后端试算，不使用模型时长', async () => {
 let called;
 const summary=await previewAICutEditPlan('p','plan',['move'],async (path,options)=>{called={path,options};return {summary:{estimatedDurationMs:1234}}});
 assert.equal(summary.estimatedDurationMs,1234);
 assert.equal(called.path,'/api/ai-cut/projects/p/edit-plans/plan/preview');
 assert.deepEqual(JSON.parse(called.options.body),{enabledOperationIds:['move']});
 await assert.rejects(previewAICutEditPlan('p','plan',['../bad'],async()=>{}),/无效/);
});
