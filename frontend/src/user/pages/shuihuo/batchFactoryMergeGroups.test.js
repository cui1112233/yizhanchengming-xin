import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';
import { currentMergeJobsByOpening, publishMergeFiles } from './batchFactoryMergeGroups.js';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

test('shows one current complete video per opening while retry attempts stay in history', () => {
  const jobs = [
    { id: 'opening-2-old', openingVariantIndex: 2, status: 'succeeded', outputUrl: '/2-old.mp4', updatedAt: '2026-10-17T10:00:00Z' },
    { id: 'opening-0', openingVariantIndex: 0, status: 'succeeded', outputUrl: '/0.mp4', updatedAt: '2026-10-17T10:01:00Z' },
    { id: 'opening-3', openingVariantIndex: 3, status: 'succeeded', outputUrl: '/3.mp4', updatedAt: '2026-10-17T10:03:00Z' },
    { id: 'opening-1', openingVariantIndex: 1, status: 'succeeded', outputUrl: '/1.mp4', updatedAt: '2026-10-17T10:02:00Z' },
    { id: 'opening-2-new', openingVariantIndex: 2, status: 'succeeded', outputUrl: '/2-new.mp4', updatedAt: '2026-10-17T10:04:00Z' }
  ];

  const result = currentMergeJobsByOpening(jobs);

  assert.deepEqual(result.current.map(job => job.id), ['opening-0', 'opening-1', 'opening-2-new', 'opening-3']);
  assert.equal(result.historyCount, 1);
  assert.deepEqual(result.current.map(job => job.openingLabel), ['原开头', '换开头 1', '换开头 2', '换开头 3']);
  assert.match(source, /currentMergeJobsByOpening\(mergeJobs\)/);
  assert.match(source, /job\.openingLabel/);
  assert.match(source, /上传组成/);
});

test('keeps the newest completed video visible when a later retry fails', () => {
  const result = currentMergeJobsByOpening([
    { id: 'completed', openingVariantIndex: 1, status: 'succeeded', outputUrl: '/ready.mp4', updatedAt: '2026-10-17T10:00:00Z' },
    { id: 'failed-retry', openingVariantIndex: 1, status: 'failed', updatedAt: '2026-10-17T10:01:00Z' }
  ]);

  assert.equal(result.current[0].id, 'completed');
  assert.equal(result.historyCount, 1);
});

test('requires four uniquely named complete videos for a four-opening book', () => {
  const jobs = [0, 1, 2, 3].map(index => ({
    id: `merge-${index}`,
    openingVariantIndex: index,
    status: 'succeeded',
    outputUrl: `/${index}.mp4`,
    updatedAt: `2026-10-17T10:0${index}:00Z`
  }));

  const complete = publishMergeFiles('123456', jobs, { openingEnabled: true, openingCount: 4 }, 2);
  assert.equal(complete.ready, true);
  assert.deepEqual(complete.files, [
    '123456-opening-0.mp4',
    '123456-opening-1.mp4',
    '123456-opening-2.mp4',
    '123456-opening-3.mp4'
  ]);

  const incomplete = publishMergeFiles('123456', jobs.filter(job => job.openingVariantIndex !== 2), { openingEnabled: true, openingCount: 4 }, 2);
  assert.equal(incomplete.ready, false);
  assert.deepEqual(incomplete.missingOpeningIndexes, [2]);
});
