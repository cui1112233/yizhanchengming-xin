import test from 'node:test';
import assert from 'node:assert/strict';
import { artworkFrameTimes, artworkPeaks, createArtworkQueue } from './mediaArtwork.js';

test('frame samples follow trimmed source, speed, reverse and frozen frames', () => {
  const asset = { durationMs: 20000 };
  const clip = { sourceStartMs: 2000, durationMs: 4000, playbackRate: 2 };
  assert.deepEqual(artworkFrameTimes(asset, clip, true), [2000, 6000, 9999]);
  assert.deepEqual(artworkFrameTimes(asset, { ...clip, reversed: true }, true), [9999, 5999, 2000]);
  assert.deepEqual(artworkFrameTimes(asset, { ...clip, freezeSourceMs: 3500 }, true), [3500, 3500, 3500]);
  assert.deepEqual(artworkFrameTimes(asset, { ...clip, freezeSourceMs: 500 }, true), [500, 500, 500]);
  assert.deepEqual(artworkFrameTimes(asset, { sourceStartMs: 0, durationMs: 4000, speedCurve: [{ position: 0, rate: 1 }, { position: 1, rate: 2 }] }, true), [0, 2500, 5999]);
});

test('waveform uses source peaks and preserves silence, trim and reverse', () => {
  const analysis = { durationMs: 8000, peaks: [0, .1, .2, .3, .4, .5, .6, .7] };
  assert.deepEqual(artworkPeaks(analysis, { sourceStartMs: 2000, durationMs: 4000 }, 4), [.2, .3, .4, .5]);
  assert.deepEqual(artworkPeaks(analysis, { sourceStartMs: 2000, durationMs: 4000, reversed: true }, 4), [.5, .4, .3, .2]);
  assert.deepEqual(artworkPeaks({ durationMs: 1000, peaks: [0, 0] }, null, 2), [0, 0]);
  assert.deepEqual(artworkPeaks({ durationMs: 0, peaks: [] }), []);
});

test('queue deduplicates work, bounds concurrency and evicts cached results', async () => {
  const queue = createArtworkQueue({ concurrency: 1, maxCache: 1, maxPending: 2 });
  let finish;
  let starts = 0;
  const first = queue.acquire('a', () => { starts++; return new Promise(resolve => { finish = resolve; }); });
  const duplicate = queue.acquire('a', () => { throw new Error('duplicate started'); });
  const second = queue.acquire('b', () => { starts++; return 'B'; });
  await Promise.resolve();
  assert.equal(starts, 1);
  finish('A');
  assert.equal(await first.promise, 'A');
  assert.equal(await duplicate.promise, 'A');
  assert.equal(await second.promise, 'B');
  assert.equal(starts, 2);
  assert.equal(await queue.acquire('a', () => 'new A').promise, 'new A');
});

test('queue cancels unseen pending work and rejects excess without loading it', async () => {
  const queue = createArtworkQueue({ concurrency: 1, maxPending: 1 });
  let finish;
  const first = queue.acquire('a', () => new Promise(resolve => { finish = resolve; }));
  const second = queue.acquire('b', () => { throw new Error('unseen work started'); });
  const excess = queue.acquire('c', () => 'C');
  await assert.rejects(excess.promise, /队列/);
  second.release();
  await assert.rejects(second.promise, { name: 'AbortError' });
  await Promise.resolve();
  finish('A');
  await first.promise;
});

test('failed media is cached instead of repeatedly decoding, while cancelled work can retry', async () => {
  const queue = createArtworkQueue();
  let calls = 0;
  const fail = () => { calls++; throw new Error('unreadable'); };
  await assert.rejects(queue.acquire('bad', fail).promise, /unreadable/);
  await assert.rejects(queue.acquire('bad', fail).promise, /unreadable/);
  assert.equal(calls, 1);
  const work = queue.acquire('cancel', signal => new Promise((resolve, reject) => {
    if (signal.aborted) reject(Object.assign(new Error('cancelled'), { name: 'AbortError' }));
    else signal.addEventListener('abort', () => reject(Object.assign(new Error('cancelled'), { name: 'AbortError' })));
  }));
  work.release();
  await assert.rejects(work.promise, { name: 'AbortError' });
  assert.equal(await queue.acquire('cancel', () => 'retried').promise, 'retried');
});
