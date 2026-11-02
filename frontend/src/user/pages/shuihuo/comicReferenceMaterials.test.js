import assert from 'node:assert/strict';
import test from 'node:test';

import { referenceSourceLabel, removablePageReferences } from './comicReferenceMaterials.js';

test('shows a plain Chinese source label for every comic reference source', () => {
  assert.equal(referenceSourceLabel('manual-upload'), '手动上传');
  assert.equal(referenceSourceLabel('scene-library'), '场景图库');
  assert.equal(referenceSourceLabel('book-asset'), '漫画资产');
  assert.equal(referenceSourceLabel('global-reference'), '全局参考图');
});

test('removing a chip only removes the selected manual page reference', () => {
  const references = [
    { url: '/manual/a', source: 'manual-upload', removable: true },
    { url: '/asset/b', source: 'book-asset', removable: false },
    { url: '/manual/c', source: 'manual-upload', removable: true }
  ];
  assert.deepEqual(removablePageReferences(references, '/manual/a'), ['/manual/c']);
  assert.deepEqual(removablePageReferences(references, '/asset/b'), ['/manual/a', '/manual/c']);
});
