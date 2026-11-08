import assert from 'node:assert/strict';
import test from 'node:test';

import { chooseN8Audio, producerAccountOptions } from './batchFactoryComicPublishAccounts.js';

test('only keeps a saved producer account when it belongs to the server-managed valid directory', () => {
  assert.deepEqual(producerAccountOptions(['有效账号', '唐璐', '有效账号'], '有效账号'), [
    { value: '有效账号', label: '有效账号（已保存）' },
    { value: '唐璐', label: '唐璐' }
  ]);
  assert.deepEqual(producerAccountOptions(['有效账号'], '崔家铭2'), [
    { value: '崔家铭2', label: '崔家铭2（已保存·目录未同步）' },
    { value: '有效账号', label: '有效账号' }
  ]);
});

test('keeps the saved audio when it remains among matching candidates', () => {
  const audios = [
    { id: 2, name: '较新音频', status: 'ENABLE', updatedAt: '2026-11-02T10:00:00Z' },
    { id: 1, name: '已保存音频', status: 'ENABLE', updatedAt: '2026-11-01T10:00:00Z' }
  ];
  assert.equal(chooseN8Audio(audios, 1)?.name, '已保存音频');
});

test('randomly chooses only from enabled matching audios when nothing was saved', () => {
  const audios = [
    { id: 1, name: '第一首', status: 'ENABLE' },
    { id: 2, name: '已停用', status: 'DISABLE' },
    { id: 3, name: '最后一首', status: 'ENABLE' }
  ];
  assert.equal(chooseN8Audio(audios, '', () => 0.99)?.name, '最后一首');
  assert.equal(chooseN8Audio(audios, '', () => 0)?.name, '第一首');
});
