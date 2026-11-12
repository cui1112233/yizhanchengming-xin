import assert from 'node:assert/strict';
import test from 'node:test';
import * as registry from './pageService.js';

const service = (scope = 'alice/project-1', render = () => null) => ({ id: 'aicut', scope, label: 'AI-CUT 剪辑', render });

test('replacement survives stale page cleanup and current cleanup restores generic service', () => {
  const first = registry.registerPetPageService(service());
  const second = registry.registerPetPageService(service('alice/project-2'));
  first();
  assert.equal(registry.getPetPageService().scope, 'alice/project-2');
  second();
  assert.equal(registry.getPetPageService(), null);
});

test('open only targets the active service and retains intent for a late host', () => {
  const cleanup = registry.registerPetPageService(service());
  assert.equal(registry.openPetPageService('other'), false);
  assert.equal(registry.getPetPageService().openRequest, 0);
  assert.equal(registry.openPetPageService('aicut'), true);
  assert.equal(registry.getPetPageService().openRequest, 1);
  cleanup();
  assert.equal(registry.openPetPageService('aicut'), false);
  const nextCleanup = registry.registerPetPageService(service('bob/project-1'));
  assert.equal(registry.getPetPageService().openRequest, 0);
  nextCleanup();
});

test('same-scope render updates preserve open intent while stale cleanup cannot clear updates', () => {
  const firstCleanup = registry.registerPetPageService(service());
  registry.openPetPageService('aicut');
  const render = () => 'updated timeline';
  const cleanup = registry.registerPetPageService(service(undefined, render));
  firstCleanup();
  assert.equal(registry.getPetPageService().render(), 'updated timeline');
  assert.equal(registry.getPetPageService().openRequest, 1);
  cleanup();
});

test('subscription receives changes and unsubscribe stops notifications', () => {
  const snapshots = [];
  const unsubscribe = registry.subscribePetPageService(() => snapshots.push(registry.getPetPageService()));
  const cleanup = registry.registerPetPageService(service());
  registry.openPetPageService('aicut');
  assert.equal(snapshots.length, 2);
  assert.equal(snapshots[1].openRequest, 1);
  unsubscribe();
  cleanup();
  assert.equal(snapshots.length, 2);
});

test('registration rejects incomplete services and strips unsupported callbacks and secrets', () => {
  for (const invalid of [null, {}, { ...service(), scope: '' }, { ...service(), render: 'url' }]) {
    assert.throws(() => registry.registerPetPageService(invalid), TypeError);
  }
  const cleanup = registry.registerPetPageService({ ...service(), model: 'secret', execute: () => {} });
  assert.deepEqual(Object.keys(registry.getPetPageService()).sort(), ['id', 'label', 'openRequest', 'render', 'scope']);
  cleanup();
});
