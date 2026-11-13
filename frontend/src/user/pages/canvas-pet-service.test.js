import test from 'node:test';
import assert from 'node:assert/strict';
import { createCanvasWriteGate, lockCanvasPetInput, runCanvasPetRequest } from './canvas-pet-service.js';

test('pending uploads and script writes prevent apply; applying prevents new writers', () => {
  const gate = createCanvasWriteGate(); const upload = gate.beginWrite(), split = gate.beginWrite();
  assert.throws(() => gate.beginApply(), /仍在处理/);
  upload(); upload(); assert.throws(() => gate.beginApply(), /仍在处理/);
  split(); const release = gate.beginApply();
  assert.throws(() => gate.beginWrite(), /正在应用/);
  release(); const next = gate.beginWrite(); next();
  const nextRelease = gate.beginApply(); release();
  assert.throws(() => gate.beginWrite(), /正在应用/);
  nextRelease();
});

test('approval write lock blocks input until persistence completes and releases afterwards', () => {
  const target = new EventTarget(); const unlock = lockCanvasPetInput(target);
  assert.equal(target.dispatchEvent(new Event('click', { cancelable: true })), false);
  assert.equal(target.dispatchEvent(new Event('keydown', { cancelable: true })), false);
  unlock();
  assert.equal(target.dispatchEvent(new Event('click', { cancelable: true })), true);
});

test('saves first, sends structured selected IDs and never invents system prompts', async () => {
  const order = [];
  const result = await runCanvasPetRequest({
    taskId: 'one', getTaskId: () => 'one', getSelectedIds: () => ['n1'],
    flush: async () => { order.push('save'); return { revision: 7 }; },
    ask: async input => { order.push('ask'); assert.equal(input.context.canvasPet.revision, 7); assert.deepEqual(input.context.canvasPet.selectedNodeIds, ['n1']); assert.equal(input.context.director.mode, 'supervised'); assert.equal(input.prompt, '整理故事'); return { task: { id: 'one' } }; }
  }, { prompt: '整理故事' });
  assert.equal(result.task.id, 'one'); assert.deepEqual(order, ['save', 'ask']);
});

test('save failure and scope change prevent requests', async () => {
  let requests = 0;
  const base = { taskId: 'one', getTaskId: () => 'one', getSelectedIds: () => [], ask: async () => { requests++; } };
  await assert.rejects(runCanvasPetRequest({ ...base, flush: async () => { throw Error('save failed'); } }, { prompt: 'x' }), /save failed/);
  let id = 'one';
  await assert.rejects(runCanvasPetRequest({ ...base, getTaskId: () => id, flush: async () => { id = 'two'; return { revision: 1 }; } }, { prompt: 'x' }), /项目已切换/);
  assert.equal(requests, 0);
});

test('late response from old project is rejected and approval uses durable revision', async () => {
  let id = 'one';
  const base = { taskId: 'one', getTaskId: () => id, getSelectedIds: () => [], flush: async () => ({ revision: 4 }) };
  await assert.rejects(runCanvasPetRequest({ ...base, ask: async () => { id = 'two'; return {}; } }, { prompt: 'x' }), /项目已切换/);
  id = 'one';
  const result = await runCanvasPetRequest({ ...base, approve: async (task, plan, revision) => ({ task, plan, revision }) }, { planId: 'plan-1' });
  assert.deepEqual(result, { task: 'one', plan: 'plan-1', revision: 4 });
});
