import assert from 'node:assert/strict';
import test from 'node:test';
import { batchMediaCounts, resolveBookEditableContent, resolveBookProductionText } from './batchFactoryRuntimeLogic.js';

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
