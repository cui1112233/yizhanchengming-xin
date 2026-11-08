import assert from 'node:assert/strict';
import test from 'node:test';
import {
  PRODUCT_UPDATE_SEEN_KEY,
  markProductUpdateSeen,
  normalizeProductUpdate,
  readSeenReleaseId,
  resolveProductUpdateStorage,
  shouldShowProductUpdate
} from './productUpdate.js';

function memoryStorage(initial = {}) {
  const data = new Map(Object.entries(initial));
  return {
    getItem: key => (data.has(key) ? data.get(key) : null),
    setItem: (key, value) => data.set(key, String(value)),
    removeItem: key => data.delete(key)
  };
}

const payload = {
  releaseId: 'qiantie-test',
  version: '17',
  releasedAt: '2026-11-06',
  title: '产品更新 · 2026-11-06',
  summary: '一句话总结',
  sections: [
    { type: 'feature', label: '新功能', items: [{ title: 'A', desc: 'a', tags: ['标签'], link: '/x' }] },
    { type: 'improvement', label: '改进', items: [{ title: 'B', desc: 'b' }] },
    { type: 'fix', label: '修复', items: [{ title: 'C', desc: 'c' }] },
    { type: 'other', label: '其它', items: [{ title: 'D' }] }
  ]
};

test('normalizeProductUpdate keeps the three known groups in payload order', () => {
  const note = normalizeProductUpdate(payload);

  assert.ok(note);
  assert.equal(note.releaseId, 'qiantie-test');
  assert.deepEqual(note.sections.map(section => section.type), ['feature', 'improvement', 'fix']);
  assert.equal(note.sections[0].items[0].title, 'A');
  assert.deepEqual(note.sections[0].items[0].tags, ['标签']);
  assert.equal(note.sections[0].items[0].link, '/x');
});

test('normalizeProductUpdate leaves link empty when the server omits it', () => {
  const note = normalizeProductUpdate(payload);

  assert.equal(note.sections[1].items[0].link, '');
});

test('normalizeProductUpdate is silent for a 404 payload or empty content', () => {
  assert.equal(normalizeProductUpdate({ error: '暂无更新说明' }), null);
  assert.equal(normalizeProductUpdate(null), null);
  assert.equal(normalizeProductUpdate({ releaseId: 'x', title: '', sections: [] }), null);
  assert.equal(normalizeProductUpdate({ releaseId: 'x', title: '产品更新', sections: [] }), null);
});

test('a release only shows once per browser', () => {
  const storage = memoryStorage();

  assert.equal(shouldShowProductUpdate('qiantie-test', storage), true);
  assert.equal(markProductUpdateSeen(storage, 'qiantie-test'), true);
  assert.equal(shouldShowProductUpdate('qiantie-test', storage), false);
  assert.equal(shouldShowProductUpdate('qiantie-next', storage), true);
  assert.equal(readSeenReleaseId(storage), 'qiantie-test');
  assert.equal(storage.getItem(PRODUCT_UPDATE_SEEN_KEY), 'qiantie-test');
});

test('unavailable localStorage degrades to not showing the modal', () => {
  const blocked = {
    getItem() { throw new Error('blocked'); },
    setItem() { throw new Error('blocked'); }
  };

  assert.equal(shouldShowProductUpdate('qiantie-test', blocked), false);
  assert.equal(shouldShowProductUpdate('qiantie-test', null), false);
  assert.equal(markProductUpdateSeen(blocked, 'qiantie-test'), false);
  assert.equal(readSeenReleaseId(blocked), '');
});

test('an empty releaseId never shows', () => {
  assert.equal(shouldShowProductUpdate('', memoryStorage()), false);
  assert.equal(markProductUpdateSeen(memoryStorage(), ''), false);
});

test('resolveProductUpdateStorage returns null without a browser window', () => {
  assert.equal(resolveProductUpdateStorage(), null);
});
