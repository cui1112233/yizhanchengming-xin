import test from 'node:test';
import assert from 'node:assert/strict';
import { buildComicDeliverables, comicPagePresentation, comicPublishStatusLabel, formatComicTimestamp } from './agent-comic-deliverables.js';

function comicPageNode(overrides = {}) {
  return {
    id: 'page-1',
    type: 'image',
    title: '漫画页 01',
    status: 'succeeded',
    groupId: 'book:book-1',
    assetRef: { kind: 'comic_page', id: 'img-1', url: 'https://cdn.example.com/page-1.png' },
    config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', resourceType: 'comic_page', resourceId: 'page-1', mappedState: 'succeeded' },
    ...overrides
  };
}

function collectionNode(config = {}, overrides = {}) {
  return {
    id: 'collection-1',
    type: 'image',
    title: '漫画合集',
    status: 'succeeded',
    groupId: 'book:book-1',
    assetRef: { kind: 'comic_collection', url: 'https://cdn.example.com/page-1.png' },
    config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', resourceType: 'comic_collection', resourceId: 'collection-1', pageCount: 1, ...config },
    ...overrides
  };
}

function mergeNode(config = {}) {
  return {
    id: 'merge-1',
    type: 'video',
    title: '最终合成视频',
    groupId: 'book:book-1',
    config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', resourceType: 'merge', resourceId: 'book-1', ...config }
  };
}

test('ignores canvases without comic nodes', () => {
  assert.deepEqual(buildComicDeliverables(), []);
  assert.deepEqual(buildComicDeliverables([{ id: 'video-1', config: { source: 'batch_factory', resourceType: 'video' } }]), []);
});

test('projects comic pages with thumbnails and per-page progress', () => {
  const [book] = buildComicDeliverables([comicPageNode()]);
  assert.equal(book.bookLabel, '书籍 book-1');
  assert.equal(book.pages.length, 1);
  assert.equal(book.pages[0].url, 'https://cdn.example.com/page-1.png');
  assert.equal(book.pages[0].statusLabel, '已生成');
  assert.equal(book.progress.generated, 1);
  assert.equal(book.progress.total, 1);
  assert.equal(book.progress.label, '进度：已生成 1/1');
});

test('keeps a page readable when its thumbnail is not synced yet', () => {
  const node = comicPageNode({ assetRef: { kind: 'comic_page' }, config: { ...comicPageNode().config, mappedState: 'running' } });
  const page = comicPagePresentation(node);
  assert.equal(page.hasImage, false);
  assert.equal(page.url, '');
  assert.equal(page.statusLabel, '生成中…');
  const [book] = buildComicDeliverables([node]);
  assert.equal(book.progress.generated, 0);
  assert.equal(book.progress.running, 1);
  assert.equal(book.progress.label, '进度：已生成 0/1 · 生成中 1');
});

test('flags pages whose prompt configuration is stale', () => {
  const node = comicPageNode({ config: { ...comicPageNode().config, configStale: true, mappedState: 'queued' } });
  const [book] = buildComicDeliverables([node]);
  assert.equal(book.pages[0].state, 'stale');
  assert.equal(book.pages[0].statusLabel, '配置已过期');
  assert.equal(book.progress.stale, 1);
  assert.equal(book.progress.label, '进度：已生成 1/1 · 需更新 1');
});

test('reads the N8 comic publish record from the collection projection', () => {
  const [book] = buildComicDeliverables([
    collectionNode({ pageCount: 3, n8PublishStatus: 'published', n8PublishCarouselId: 42, n8PublishWorkIds: [7, 8], n8PublishedAt: '2026-10-23T10:05:00Z' })
  ]);
  assert.equal(book.collections[0].pageCountLabel, '3 页');
  assert.equal(book.publish.n8.label, '已发布');
  assert.equal(book.publish.n8.carouselId, 42);
  assert.deepEqual(book.publish.n8.workIds, [7, 8]);
  assert.equal(book.publish.n8.detail, `轮播 42 · 作品 2 · ${formatComicTimestamp('2026-10-23T10:05:00Z')}`);
});

test('shows 未同步 for missing collection and publish fields instead of failing', () => {
  const [book] = buildComicDeliverables([comicPageNode()]);
  assert.deepEqual(book.collections, []);
  assert.deepEqual(book.publish, {
    n8: { status: '', label: '未同步', publishedAt: '', carouselId: 0, workIds: [], detail: '' },
    provider121: { status: '', label: '未同步', submittedAt: '' },
    yd: { status: '', label: '未同步', submittedAt: '' }
  });
});

test('projects 121 and YD publish status from the merge node', () => {
  const [book] = buildComicDeliverables([
    comicPageNode(),
    mergeNode({ publish121Status: 'submitted', publish121SubmittedAt: '2026-10-23T10:06:00Z', publishYadiStatus: 'confirmed' })
  ]);
  assert.equal(book.publish.provider121.label, '已提交');
  assert.equal(book.publish.provider121.submittedAt, formatComicTimestamp('2026-10-23T10:06:00Z'));
  assert.equal(book.publish.yd.label, '已确认');
  assert.equal(book.publish.yd.submittedAt, '');
});

test('splits comic deliverables per book group', () => {
  const secondPage = comicPageNode({
    id: 'page-2',
    groupId: 'book:book-2',
    config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-2', resourceType: 'comic_page', resourceId: 'page-2', mappedState: 'queued' }
  });
  const groups = buildComicDeliverables([comicPageNode(), secondPage]);
  assert.equal(groups.length, 2);
  assert.deepEqual(groups.map(group => group.bookId), ['book-1', 'book-2']);
  assert.equal(groups[1].progress.label, '进度：已生成 1/1');
});

test('accepts React Flow node shapes and malformed configs without throwing', () => {
  const flowNode = {
    id: 'page-flow',
    type: 'agent',
    data: {
      type: 'image',
      title: '漫画页 02',
      status: 'succeeded',
      groupId: 'book:book-1',
      assetRef: { url: 'https://cdn.example.com/page-2.png' },
      config: { source: 'batch_factory', bookId: 'book-1', resourceType: 'comic_page' }
    }
  };
  assert.doesNotThrow(() => buildComicDeliverables([flowNode, { id: 'broken' }, null]));
  const [book] = buildComicDeliverables([flowNode]);
  assert.equal(book.pages[0].title, '漫画页 02');
  assert.equal(book.pages[0].url, 'https://cdn.example.com/page-2.png');
});

test('labels unknown publish statuses and timestamps safely', () => {
  assert.equal(comicPublishStatusLabel(''), '未同步');
  assert.equal(comicPublishStatusLabel('weird'), 'weird');
  assert.equal(formatComicTimestamp('not-a-date'), '');
  assert.equal(formatComicTimestamp(''), '');
});
