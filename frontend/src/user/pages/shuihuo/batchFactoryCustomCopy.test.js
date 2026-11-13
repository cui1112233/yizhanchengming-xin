import test from 'node:test';
import assert from 'node:assert/strict';
import { copyInputIDs, selectedCopyGroups } from './batchFactoryCustomCopy.js';
test('large IDs stay strings and invalid input is not silently skipped', () => {
  assert.deepEqual(copyInputIDs('2100242549395006289\n2100242549395006289'), ['2100242549395006289']);
  assert.throws(() => copyInputIDs('2100242549395006289\nabc'));
});
test('requires explicit choice and imports exactly selected version under real platform', () => {
  const bookId = '2100242549395006289';
  const copies = [{ id: '280894', bookId, title: '青霄赴', platform: '15', content: '第一份' }, { id: '280893', bookId, title: '青霄赴', platform: '15', content: '第二份\n完整结尾' }];
  assert.throws(() => selectedCopyGroups([], [{ bookId, copies }], {}));
  const groups = selectedCopyGroups([], [{ bookId, copies }], { [bookId]: '280893' });
  assert.equal(groups[0].platformId, '15');
  assert.equal(groups[0].sourceTextByBookId[bookId], '第二份\n完整结尾');
  assert.equal(groups[0].sourceMetadataByBookId[bookId].customCopyRecordId, '280893');
  assert.throws(() => selectedCopyGroups([], [{ bookId, copies: [] }], { [bookId]: '280894' }));
});
