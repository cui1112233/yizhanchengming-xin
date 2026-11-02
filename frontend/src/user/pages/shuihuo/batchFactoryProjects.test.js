import test from 'node:test';
import assert from 'node:assert/strict';
import {
  batchFactoryBatchFromResponse,
  batchFactoryCoverFrom,
  batchFactoryFallbackCoverFromProductionStatus,
  createCoverHydrationScheduler,
  enrichBatchProjectCovers,
  batchFactoryProjectsFrom,
  isBatchFactoryV11Project
} from './batchFactoryProjects.js';

test('V11 batches retain their upstream ID when shown in the shared personal works list', () => {
  const [project] = batchFactoryProjectsFrom([{ id: 'batch-42', title: '九月批量', books: [], coverMedia: { kind: 'image', url: '/cover.png' } }]);

  assert.deepEqual(project, {
    id: 'batch:batch-42',
    batchId: 'batch-42',
    name: '九月批量',
    productionMode: 'batch_factory',
    factoryMode: 'video',
    source: 'batch_factory_v11',
    createdAt: undefined,
    updatedAt: undefined,
    coverMedia: { kind: 'image', url: '/cover.png' },
    batch: { id: 'batch-42', title: '九月批量', books: [], coverMedia: { kind: 'image', url: '/cover.png' } }
  });
  assert.equal(isBatchFactoryV11Project(project), true);
});

test('a legacy Shuihuo project marked batch_factory never calls the V11 batch endpoint', () => {
  assert.equal(isBatchFactoryV11Project({ id: 'water-1', productionMode: 'batch_factory' }), false);
});

test('malformed V11 list entries are omitted instead of producing a /batches/ request', () => {
  assert.deepEqual(batchFactoryProjectsFrom([{ title: '缺少编号' }, null]), []);
});

test('a V11 batch detail response is unwrapped before the novel list reads its books', () => {
  const batch = batchFactoryBatchFromResponse({ batch: { id: 'batch-42', title: '九月批量', books: [{ id: 'book-1' }] } });

  assert.equal(batch.id, 'batch-42');
  assert.equal(batch.books.length, 1);
});


test('uses the first book and storyboard with a successful video before falling back to its image cover', () => {
  const batch = { books: [{ id: 'book-1', videos: [{ id: 'video-1' }, { id: 'video-2' }] }, { id: 'book-2', videos: [{ id: 'video-3' }] }] };
  assert.deepEqual(batchFactoryCoverFrom(batch, { jobs: [{ tasks: [{ videoId: 'video-3', status: 'succeeded', mediaUrl: '/third.mp4' }, { videoId: 'video-1', status: 'succeeded', mediaUrl: '/first.mp4' }] }] }, '/fallback.png'), { kind: 'video', url: '/first.mp4' });
  assert.deepEqual(batchFactoryCoverFrom(batch, { jobs: [] }, '/fallback.png'), { kind: 'image', url: '/fallback.png' });
  assert.equal(batchFactoryCoverFrom(batch, { jobs: [] }), null);
});

test('uses the selected first-shot upload video as the project cover', () => {
  const batch = { books: [{
    id: 'book-1',
    videos: [{ id: 'video-1' }],
    settingsState: { patch: { primaryUploadSource: { kind: 'video', videoId: 'video-1', taskId: 'selected-task' } } }
  }] };
  const status = { jobs: [{ tasks: [
    { id: 'newer-task', videoId: 'video-1', status: 'succeeded', mediaUrl: '/newer.mp4' },
    { id: 'selected-task', videoId: 'video-1', status: 'succeeded', mediaUrl: '/selected.mp4' }
  ] }] };

  assert.deepEqual(batchFactoryCoverFrom(batch, status), { kind: 'video', url: '/selected.mp4' });
});

test('uses a completed production video only when a legacy merge cover is unavailable', () => {
  assert.deepEqual(batchFactoryFallbackCoverFromProductionStatus({ jobs: [{ tasks: [
    { id: 'queued', status: 'queued', mediaUrl: '/queued.mp4' },
    { id: 'complete', status: 'succeeded', mediaUrl: 'https://media.example/first.mp4' }
  ] }] }), { kind: 'video', url: 'https://media.example/first.mp4' });
  assert.equal(batchFactoryFallbackCoverFromProductionStatus({ jobs: [{ tasks: [{ status: 'failed', mediaUrl: '/failed.mp4' }] }] }), null);
});

test('hydrates project covers with bounded request concurrency and streams each result', async () => {
  const projects = Array.from({ length: 5 }, (_, index) => ({
    id: `batch:${index}`,
    batchId: `batch-${index}`,
    batch: { books: [{ videos: [{ id: `video-${index}` }] }] }
  }));
  const updates = [];
  let active = 0;
  let maximumActive = 0;
  const request = value => new Promise(resolve => {
    active += 1;
    maximumActive = Math.max(maximumActive, active);
    setTimeout(() => {
      active -= 1;
      resolve(value);
    }, 5);
  });

  await enrichBatchProjectCovers(projects, {
    concurrency: 2,
    loadProduction: project => request({ jobs: [{ tasks: [{ videoId: `video-${project.batchId.split('-')[1]}`, status: 'succeeded', mediaUrl: `/${project.batchId}.mp4` }] }] }),
    loadMerge: () => request({ jobs: [] }),
    coverFrom: (project, productionStatus) => batchFactoryCoverFrom(project.batch, productionStatus),
    onUpdate: project => updates.push(project)
  });

  assert.ok(maximumActive <= 2, `expected at most 2 active requests, saw ${maximumActive}`);
  assert.equal(updates.length, 5);
  assert.deepEqual(updates.map(project => project.coverMedia?.url), projects.map(project => `/${project.batchId}.mp4`));
});

test('serializes superseding cover hydration runs so their combined request concurrency stays bounded', async () => {
  const schedule = createCoverHydrationScheduler();
  const projects = Array.from({ length: 2 }, (_, index) => ({ id: `batch:${index}`, batchId: `batch-${index}` }));
  let active = 0;
  let maximumActive = 0;
  const request = () => new Promise(resolve => {
    active += 1;
    maximumActive = Math.max(maximumActive, active);
    setTimeout(() => {
      active -= 1;
      resolve({});
    }, 5);
  });
  const hydrate = () => enrichBatchProjectCovers(projects, {
    concurrency: 2,
    loadProduction: request,
    loadMerge: request,
    coverFrom: () => null,
    onUpdate: () => {}
  });

  await Promise.all([schedule(hydrate), schedule(hydrate)]);

  assert.ok(maximumActive <= 2, `expected at most 2 active requests across runs, saw ${maximumActive}`);
});

test('reports hydration request failures so callers do not cache a transient empty result', async () => {
  const updates = [];

  await enrichBatchProjectCovers([{ id: 'batch:1', batchId: 'batch-1' }], {
    loadProduction: async () => { throw new Error('temporary outage'); },
    loadMerge: async () => ({}),
    coverFrom: () => null,
    onUpdate: (project, metadata) => updates.push({ project, metadata })
  });

  assert.equal(updates.length, 1);
  assert.equal(updates[0].metadata.hadRequestFailure, true);
});
