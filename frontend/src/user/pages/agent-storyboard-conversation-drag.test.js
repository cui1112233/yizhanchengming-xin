import assert from 'node:assert/strict';
import test from 'node:test';
import { hasStoryboardAssetDrop, readStoryboardAssetDrop, writeStoryboardAssetDrag } from './agent-storyboard-conversation-drag.js';

function transfer() {
  const values = new Map();
  return {
    effectAllowed: '',
    types: [],
    setData(type, value) {
      values.set(type, value);
      this.types = [...values.keys()];
    },
    getData(type) { return values.get(type) || ''; }
  };
}

test('storyboard asset drag round-trips only a current project asset', () => {
  const dataTransfer = transfer();
  const poster = { id: 'asset-poster', name: '人物主视觉', kind: 'image' };
  assert.equal(writeStoryboardAssetDrag(dataTransfer, poster), true);
  assert.equal(dataTransfer.effectAllowed, 'copy');
  assert.equal(hasStoryboardAssetDrop(dataTransfer), true);
  assert.equal(readStoryboardAssetDrop(dataTransfer, [poster]), poster);
  assert.equal(readStoryboardAssetDrop(dataTransfer, [{ id: 'other' }]), null);
});

test('storyboard drop ignores malformed or unrelated drag data', () => {
  const dataTransfer = transfer();
  dataTransfer.setData('text/plain', 'asset-poster');
  assert.equal(hasStoryboardAssetDrop(dataTransfer), false);
  assert.equal(readStoryboardAssetDrop(dataTransfer, [{ id: 'asset-poster' }]), null);
});
