import assert from 'node:assert/strict';
import test from 'node:test';

let performanceLogic = {};
try {
  performanceLogic = await import('./batchFactoryPerformance.js');
} catch {
  // 首次红灯时模块尚不存在；下面用行为断言给出清楚的失败原因。
}

test('comic row stays loaded after it first enters the nearby viewport', () => {
  assert.equal(typeof performanceLogic.nextViewportLoaded, 'function');
  assert.equal(performanceLogic.nextViewportLoaded(false, [{ isIntersecting: false }]), false);
  assert.equal(performanceLogic.nextViewportLoaded(false, [{ isIntersecting: true }]), true);
  assert.equal(performanceLogic.nextViewportLoaded(true, [{ isIntersecting: false }]), true);
});

test('runtime polling is fast only while automation is active', () => {
  assert.equal(typeof performanceLogic.batchFactoryRuntimePollDelay, 'function');
  assert.equal(performanceLogic.batchFactoryRuntimePollDelay('running', 0), 8000);
  assert.equal(performanceLogic.batchFactoryRuntimePollDelay('scheduled', 0), 8000);
  assert.equal(performanceLogic.batchFactoryRuntimePollDelay('idle', 0), 60000);
  assert.equal(performanceLogic.batchFactoryRuntimePollDelay('stopped', 0), 60000);
  assert.equal(performanceLogic.batchFactoryRuntimePollDelay('idle', 2), 15000);
});

test('comic rows poll every second only while an unfinished image is generating', () => {
  assert.equal(typeof performanceLogic.comicRowResultPollDelay, 'function');
  assert.equal(performanceLogic.comicRowResultPollDelay([
    { generationState: 'generating', images: [] },
  ]), 1000);
  assert.equal(performanceLogic.comicRowResultPollDelay([
    { generationState: 'generating', images: [{ isPrimary: true }] },
  ]), 0);
  assert.equal(performanceLogic.comicRowResultPollDelay([
    { generationState: 'idle', images: [] },
  ]), 0);
});

test('completed comic page is revealed without waiting for the heavy batch refresh', async () => {
  assert.equal(typeof performanceLogic.finishComicPageGeneration, 'function');
  const events = [];
  let releaseBatchRefresh;
  const batchRefresh = new Promise(resolve => { releaseBatchRefresh = resolve; });

  const completed = performanceLogic.finishComicPageGeneration({
    waitForResult: async () => { events.push('result-ready'); return { id: 'page-1' }; },
    revealResult: page => { events.push(`revealed:${page.id}`); },
    refreshBatch: async () => { events.push('batch-refresh-started'); await batchRefresh; },
  });

  assert.deepEqual(await completed, { id: 'page-1' });
  assert.deepEqual(events, ['result-ready', 'revealed:page-1', 'batch-refresh-started']);
  releaseBatchRefresh();
});
