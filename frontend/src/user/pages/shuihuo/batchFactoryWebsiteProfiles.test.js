import test from 'node:test';
import assert from 'node:assert/strict';
import { reconcileSyncedWebsiteProfiles } from './batchFactoryWebsiteProfiles.js';

test('同步结果不含旧配置档时清除失效选择', () => {
  const current = {
    websiteProfileId: '121-420',
    versionProfile: '8个解压',
    organization: '',
    websiteProfiles: [{ id: '121-420', name: '8个解压' }]
  };
  const synced = [
    { id: '121-417', name: '4滚屏', organization: '1' },
    { id: '121-379', name: '2滚2解', organization: '1' }
  ];

  assert.deepEqual(reconcileSyncedWebsiteProfiles(current, synced), {
    publishSettings: {
      websiteProfileId: '',
      versionProfile: '',
      organization: '',
      websiteProfiles: synced
    },
    selectionExpired: true
  });
});

test('同步结果仍含当前配置档时刷新它的组织归属', () => {
  const current = {
    websiteProfileId: '121-417',
    versionProfile: '4滚屏',
    organization: '',
    websiteProfiles: []
  };
  const synced = [{ id: '121-417', name: '4滚屏', organization: 1 }];

  assert.deepEqual(reconcileSyncedWebsiteProfiles(current, synced), {
    publishSettings: {
      websiteProfileId: '121-417',
      versionProfile: '4滚屏',
      organization: '1',
      websiteProfiles: synced
    },
    selectionExpired: false
  });
});
