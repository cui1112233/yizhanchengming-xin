import test from 'node:test';
import assert from 'node:assert/strict';
import { createCreativeCard, creativeCardTypes, creativeAssetSelection, creativeScriptText, defaultCreativeAssetSelection, persistCreativeCanvasDraft, transferableCreativeAssets, transferableCreativeScriptNodes } from './aicutCreativeCanvas.js';
import { createCanvasSaveQueue } from './agent-canvas-state.js';
import { isDraftAtLeastAsFresh, loadCanvasDraft } from './agent-canvas-draft-storage.js';

test('shortcuts produce persisted script/text nodes with valid editable configs and caller IDs', () => {
  for (const [index, template] of creativeCardTypes.entries()) {
    const node = createCreativeCard(template.key, { id: `creative-${index}`, position: { x: 120, y: 80 } });
    assert.equal(node.id, `creative-${index}`);
    assert.equal(node.type, template.kind);
    assert.equal(node.title, template.label);
    assert.deepEqual(node.position, { x: 120, y: 80 });
    assert.equal(node.status, 'draft');
    assert.deepEqual(node.sourceNodeIds, []);
    assert.equal(node.config[template.kind === 'script' ? 'sourceText' : 'operation'], template.kind === 'script' ? '' : 'generate');
  }
  assert.throws(() => createCreativeCard('unknown', { id: 'x' }));
  assert.throws(() => createCreativeCard('story'));
});

const assets = [
  { id: 'image', kind: 'image', url: '/real-image' },
  { id: 'video', kind: 'video', url: '/real-video' },
  { id: 'audio', kind: 'audio', url: '/real-audio' },
  { id: 'text', kind: 'text', url: '/text' },
  { id: 'missing', kind: 'image' },
  { id: 'image', kind: 'image', url: '/duplicate' }
];
test('transfer selection excludes unavailable media, duplicates and stale selection', () => {
  assert.deepEqual(transferableCreativeAssets(assets).map(asset => asset.id), ['image', 'video', 'audio']);
  assert.deepEqual(creativeAssetSelection(['audio', 'audio', 'missing', 'text', 'gone'], assets), ['audio']);
  assert.deepEqual(defaultCreativeAssetSelection(assets), ['image', 'video', 'audio']);
  assert.deepEqual(defaultCreativeAssetSelection(assets, [{ id: 'node', assetRef: { id: 'video' } }], ['node']), ['video']);
  assert.deepEqual(defaultCreativeAssetSelection([]), []);
});
test('transfer selection stays within the server batch limit', () => {
  const many = Array.from({ length: 40 }, (_, i) => ({ id: String(i), kind: 'audio', url: '/real' }));
  assert.equal(defaultCreativeAssetSelection(many).length, 30);
  assert.equal(creativeAssetSelection(many.map(asset => asset.id), many).length, 30);
});

test('script picker exposes actual nonempty saved text and script content only', () => {
  const nodes = [{ id: 's', type: 'script', config: { sourceText: ' 故事正文 ' }, summary: '摘要' }, { id: 't', type: 'text', config: { prompt: '角色详情' } }, { id: 'r', type: 'text', summary: '分镜描述' }, { id: 'empty', type: 'script', config: { sourceText: ' ' } }, { id: 'media', type: 'image', summary: '图片' }];
  assert.deepEqual(transferableCreativeScriptNodes(nodes).map(node => node.id), ['s', 't', 'r']);
  assert.equal(creativeScriptText(nodes[0]), '故事正文');
  assert.equal(creativeScriptText(nodes[4]), '');
});

test('newer edits remain recoverable after an older in-flight save succeeds and their save fails offline', async () => {
  const values = new Map();
  const storage = { getItem: key => values.get(key), setItem: (key, value) => values.set(key, value) };
  const first = { version: 1, revision: 0, nodes: [{ id: 'story', type: 'text', summary: 'A' }], edges: [] };
  const second = { ...first, nodes: [{ ...first.nodes[0], summary: 'B' }] };
  const session = { username: 'creator', taskId: 'creative-project', canvas: first, queue: null };
  let finishFirst;
  const firstResponse = new Promise(resolve => { finishFirst = resolve; });
  let attempts = 0;
  session.queue = createCanvasSaveQueue({
    delay: 60000,
    save: async () => {
      if (++attempts !== 1) throw new Error('offline');
      const result = await firstResponse;
      persistCreativeCanvasDraft(storage, session, result.canvas.revision);
      return result;
    },
    onStatus: status => { if (status === 'saving') persistCreativeCanvasDraft(storage, session); }
  });
  session.queue.enqueue(first);
  const savingFirst = session.queue.flush();
  session.canvas = second;
  persistCreativeCanvasDraft(storage, session);
  session.queue.enqueue(second);
  const saved = { ...first, revision: 1 };
  finishFirst({ canvas: saved });
  await savingFirst;
  const betweenRequests = loadCanvasDraft(storage, session.username, session.taskId);
  assert.equal(betweenRequests.canvas.nodes[0].summary, 'B');
  assert.equal(isDraftAtLeastAsFresh(betweenRequests, saved), true);
  await assert.rejects(session.queue.flushRequired(), /offline/);
  const draft = loadCanvasDraft(storage, session.username, session.taskId);
  assert.equal(draft.canvas.nodes[0].summary, 'B');
  assert.equal(draft.baseRevision, 1);
  assert.equal(isDraftAtLeastAsFresh(draft, saved), true);
  session.queue.dispose();
});
