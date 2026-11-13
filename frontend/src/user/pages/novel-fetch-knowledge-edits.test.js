import test from 'node:test';
import assert from 'node:assert/strict';
import { parseOpeningDraft } from './novel-fetch-knowledge-edits.js';

test('opening save uses edited text and accepts a single edited item', () => {
  assert.deepEqual(parseOpeningDraft('[{"text":"修改后的开头"}]'), [{ text: '修改后的开头' }]);
  assert.deepEqual(parseOpeningDraft('{"content":"手工修订"}'), [{ content: '手工修订' }]);
  assert.deepEqual(parseOpeningDraft('{"id":"opening_001","text":"新分析"}'), [{ text: '新分析' }]);
});

test('opening save rejects malformed, empty and partial invalid results before any request', () => {
  for (const value of ['not json', '[]', 'null', '[{"text":"有效"},{"text":""}]']) {
    assert.throws(() => parseOpeningDraft(value));
  }
});
