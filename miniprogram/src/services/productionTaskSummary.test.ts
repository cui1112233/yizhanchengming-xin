import { expect, test } from 'vitest';
import { loadTaskSummary } from './taskSummary';
test('task center distinguishes fetch, real Agent tasks, batch summaries and per-project Shuihuo tasks', async () => {
  const payloads: Record<string, unknown> = {
    '/api/history': { entries: [] }, '/api/studio/runs': { runs: [{ id: 'a', title: 'Agent' }] },
    '/api/batch-rewrite/tasks': { tasks: [{ id: 'f', title: 'Fetch' }] },
    '/api/batch-factory/v12/batches/summary': { batches: [{ id: 'b', title: 'Batch' }] },
    '/api/batch-factory/v12/batches/b/runtime-summary': {
      automation: { state: 'idle', counts: { total: 0 } },
      production: { jobs: [{ id: 'job-b', bookId: 'book-b', status: 'running', tasks: [{ id: 'video-b', videoId: 'shot-b', status: 'running', createdAt: '2026-11-24T10:00:00Z' }] }] },
      merge: { jobs: [] }, stageSummaries: {},
    },
    '/api/shuihuo-production/projects': { projects: [{ id: 7, name: 'Water' }] },
    '/api/shuihuo-production/projects/7/tasks': { tasks: [{ id: 9, kind: 'video', status: 'queued' }] },
  };
  const result = await loadTaskSummary(async path => { if (!(path in payloads)) throw new Error('Unexpected route'); return payloads[path]; });
  expect(result.sources.every(source => source.state === 'ready')).toBe(true);
  expect(result.entries.map(entry => entry.source)).toEqual(['batch', 'agent', 'fetch', 'shuihuo']);
  expect(result.entries.find(entry => entry.source === 'batch')).toMatchObject({ id: 'video-b', status: 'running', moduleKey: 'batch', kind: 'video' });
  expect(result.entries.some(entry => entry.id === 'b')).toBe(false);
  expect(result.entries.find(entry => entry.source === 'shuihuo')).toMatchObject({ projectId: '7', status: 'running' });
});

test('one slow batch runtime cannot erase the real batch list', async () => {
  const payloads: Record<string, unknown | Error> = {
    '/api/history': { entries: [] },
    '/api/studio/runs': { runs: [] },
    '/api/batch-rewrite/tasks': { tasks: [] },
    '/api/batch-factory/v12/batches/summary': { batches: [
      { id: 'slow', title: '十月老书', productionMode: 'comic', updatedAt: '2026-11-22T01:04:00Z' },
      { id: 'active', title: '正在生产', productionMode: 'video', updatedAt: '2026-11-24T10:00:00Z' },
    ] },
    '/api/batch-factory/v12/batches/slow/runtime-summary': new Error('请求超时，请检查网络后重试'),
    '/api/batch-factory/v12/batches/active/runtime-summary': {
      automation: { state: 'running', counts: { ready: 3, total: 10 } },
      production: { jobs: [] }, merge: { jobs: [] }, stageSummaries: {},
    },
    '/api/shuihuo-production/projects': { projects: [] },
  };
  const result = await loadTaskSummary(async path => {
    const value = payloads[path];
    if (value instanceof Error) throw value;
    if (value === undefined) throw new Error(`Unexpected route: ${path}`);
    return value;
  });

  expect(result.sources.find(item => item.source === 'batch')).toMatchObject({ state: 'ready' });
  expect(result.entries.filter(entry => entry.source === 'batch')).toEqual(expect.arrayContaining([
    expect.objectContaining({ id: 'slow', title: '十月老书', status: 'unknown', kind: 'comic' }),
    expect.objectContaining({ id: 'active-automation', title: '正在生产 · 自动生产', status: 'running', progress: { completed: 3, total: 10 } }),
  ]));
});

test('an idle batch is still a real task record instead of disappearing', async () => {
  const payloads: Record<string, unknown> = {
    '/api/history': { entries: [] }, '/api/studio/runs': { runs: [] }, '/api/batch-rewrite/tasks': { tasks: [] },
    '/api/batch-factory/v12/batches/summary': { batches: [{ id: 'idle', title: '已建立的批次', productionMode: 'text' }] },
    '/api/batch-factory/v12/batches/idle/runtime-summary': {
      automation: { state: 'idle', counts: { ready: 0, total: 0 } }, production: { jobs: [] }, merge: { jobs: [] }, stageSummaries: {},
    },
    '/api/shuihuo-production/projects': { projects: [] },
  };
  const result = await loadTaskSummary(async path => payloads[path]);
  expect(result.entries.find(entry => entry.source === 'batch')).toMatchObject({ id: 'idle', title: '已建立的批次', status: 'unknown', kind: 'text' });
});
