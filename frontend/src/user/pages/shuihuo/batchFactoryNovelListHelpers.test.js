import test from 'node:test';
import assert from 'node:assert/strict';

import { bookPlatformName, sensitiveReviewPresentation } from './batchFactoryBookPresentation.js';

test('HY internal platform code is displayed as 黑岩付费', () => {
  assert.equal(bookPlatformName({ platform: 'HY', sourceMetadata: { platformCode: 'HY' } }), '黑岩付费');
  assert.equal(bookPlatformName({ platform: '黑岩', sourceMetadata: {} }), '黑岩付费');
});

test('comic sensitive hits are presented as non-blocking automatic masking', () => {
  assert.deepEqual(
    sensitiveReviewPresentation({ sensitiveReviewCount: 4 }, 'comic'),
    {
      title: '有 4 处敏感词已自动打码',
      description: '漫画自动化会继续执行，不需要人工确认。',
    },
  );
});

test('video sensitive hits also continue automation after masking', () => {
  assert.deepEqual(
    sensitiveReviewPresentation({ sensitiveReviewCount: 5 }, 'video'),
    {
      title: '有 5 处敏感词已自动打码',
      description: '自动化会继续执行，不需要人工确认。',
    },
  );
});
