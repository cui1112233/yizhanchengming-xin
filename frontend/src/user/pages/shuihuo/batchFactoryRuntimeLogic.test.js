import assert from 'node:assert/strict';
import test from 'node:test';
import { batchMediaCounts, bookVideoVersions, resolveBookEditableContent, resolveBookProductionText } from './batchFactoryRuntimeLogic.js';

test('resolves working production text before the configured source preview', () => {
  assert.equal(resolveBookProductionText({
    sourceText: '第一行\n第二行\n第三行',
    sourceMetadata: { contentRangeLines: 2 },
    workingFrontContent: '人工改写正文'
  }), '人工改写正文');
  assert.equal(resolveBookProductionText({
    sourceText: '第一行\n第二行\n第三行',
    sourceMetadata: { contentRangeLines: 2 }
  }), '第一行\n第二行');
});

test('comic editing uses the complete captured source instead of the video line range', () => {
  const book = {
    sourceText: '第一行\n第二行\n第三行',
    sourceMetadata: { contentRangeLines: 1 }
  };
  assert.equal(resolveBookEditableContent(book, 'video'), '第一行');
  assert.equal(resolveBookEditableContent(book, 'comic'), '第一行\n第二行\n第三行');
});

test('counts only playable VIDEO tasks and completed merge outputs', () => {
  const counts = batchMediaCounts(
    [{ videos: [{ id: 'v1' }, { id: 'v2' }] }],
    { jobs: [{ tasks: [
      { id: 't1', videoId: 'v1', status: 'succeeded', mediaUrl: 'https://video/1.mp4' },
      { id: 't2', videoId: 'v2', status: 'failed', mediaUrl: 'https://video/2.mp4' }
    ] }] },
    { jobs: [{ id: 'm1', status: 'succeeded', outputUrl: 'https://video/final.mp4' }] }
  );
  assert.deepEqual(counts, { storyboards: 2, playable: 1, merges: 1 });
});

test('history tasks from an orphaned video record never inflate the playable count', () => {
  const counts = batchMediaCounts(
    [{ videos: [{ id: 'v-active' }] }],
    { jobs: [
      { tasks: [
        { id: 'old1', videoId: 'v-orphaned', status: 'succeeded', mediaUrl: 'https://video/old1.mp4' },
        { id: 'old2', videoId: 'v-orphaned', status: 'succeeded', mediaUrl: 'https://video/old2.mp4' }
      ] },
      { tasks: [
        { id: 'new1', videoId: 'v-active', status: 'succeeded', mediaUrl: 'https://video/new1.mp4' },
        { id: 'new2', videoId: 'v-active', status: 'running' }
      ] }
    ] },
    {}
  );
  // 旧轮 2 条 + 新轮 1 条成功，但只有 active 记录下的 1 条算“可播放”。
  assert.equal(counts.playable, 1);
  assert.equal(counts.storyboards, 1);
});

test('a previous version keeps the storyboard playable when the active record has no media yet', () => {
  const counts = batchMediaCounts(
    [{
      id: 'book-1',
      videos: [{ id: 'v-active', ordinal: 1 }],
      videoHistory: [{ id: 'v-old', ordinal: 1, compatibilityState: 'orphaned' }]
    }],
    { jobs: [{ tasks: [
      { id: 'old', videoId: 'v-old', status: 'succeeded', mediaUrl: 'https://video/old.mp4' },
      { id: 'new', videoId: 'v-active', status: 'running' }
    ] }] },
    {}
  );
  // 同一个分镜（ordinal 1）的上一版成功视频仍算这个分镜可播放，而不是「暂无可选视频」。
  assert.equal(counts.storyboards, 1);
  assert.equal(counts.playable, 1);
});

test('a previous version stays selectable when the active record has no media', () => {
  const book = {
    id: 'book-1',
    videos: [{ id: 'v-active', ordinal: 1 }],
    videoHistory: [{ id: 'v-old', ordinal: 1, compatibilityState: 'orphaned' }]
  };
  const status = { jobs: [
    { id: 'job-old', tasks: [
      { id: 'old', videoId: 'v-old', status: 'succeeded', mediaUrl: 'https://video/old.mp4', updatedAt: '2026-11-04T16:21:00Z' }
    ] },
    { id: 'job-current', tasks: [
      { id: 'new', videoId: 'v-active', status: 'running', updatedAt: '2026-11-04T16:38:00Z' }
    ] }
  ] };
  const versions = bookVideoVersions(book, status).get('v-active');
  assert.equal(versions.length, 1);
  assert.equal(versions[0].id, 'old');
  assert.equal(versions[0].isHistory, true);
  assert.equal(versions[0].mediaUrl, 'https://video/old.mp4');
});

test('the active version and the previous version are listed together', () => {
  const book = {
    id: 'book-1',
    videos: [{ id: 'v-active', ordinal: 0 }],
    videoHistory: [
      { id: 'v-old', ordinal: 0, compatibilityState: 'orphaned' },
      { id: 'v-other', ordinal: 1, compatibilityState: 'orphaned' }
    ]
  };
  const status = { jobs: [
    { id: 'job-old', tasks: [
      { id: 'old', videoId: 'v-old', status: 'succeeded', mediaUrl: 'https://video/old.mp4', updatedAt: '2026-11-04T15:00:00Z' },
      { id: 'other', videoId: 'v-other', status: 'succeeded', mediaUrl: 'https://video/other.mp4', updatedAt: '2026-11-04T15:00:00Z' }
    ] },
    { id: 'job-current', tasks: [
      { id: 'new', videoId: 'v-active', status: 'succeeded', mediaUrl: 'https://video/new.mp4', updatedAt: '2026-11-04T16:30:00Z' }
    ] }
  ] };
  const versions = bookVideoVersions(book, status).get('v-active');
  assert.deepEqual(versions.map(version => version.id), ['old', 'new']);
  assert.equal(versions[0].isHistory, true);
  assert.equal(versions[1].isHistory, false);
});
