const COMIC_PAGE_RESOURCE = 'comic_page';
const COMIC_COLLECTION_RESOURCE = 'comic_collection';
const MERGE_RESOURCE = 'merge';
const UNSYNCED = '未同步';

const PAGE_STATE_LABELS = {
  draft: '草稿',
  queued: '排队中',
  running: '生成中…',
  succeeded: '已生成',
  failed: '生成失败',
  cancelled: '已取消',
  stale: '配置已过期'
};

// 121/YD 发布状态来自只读外部发布记录；N8 状态来自合集发布草稿。
const PUBLISH_STATUS_LABELS = {
  published: '已发布',
  uploaded: '已上传',
  submitted: '已提交',
  confirmed: '已确认',
  pending: '待发布',
  running: '发布中',
  failed: '发布失败',
  succeeded: '已成功',
  success: '已成功'
};

function readNodeConfig(node) {
  const config = node?.config ?? node?.data?.config;
  return config && typeof config === 'object' ? config : {};
}

function readNodeField(node, key) {
  const direct = node?.[key];
  if (direct !== undefined && direct !== null) return direct;
  return node?.data?.[key];
}

function readNodeGroupId(node) {
  return String(readNodeField(node, 'groupId') || '').trim();
}

function readNodeTitle(node) {
  return String(readNodeField(node, 'title') || '').trim();
}

function readPositiveNumber(value) {
  const number = Number(value);
  return Number.isFinite(number) && number > 0 ? number : 0;
}

function readNodeState(node, config) {
  const state = String(config.mappedState || readNodeField(node, 'status') || '').trim().toLowerCase();
  return PAGE_STATE_LABELS[state] ? state : 'queued';
}

export function comicPublishStatusLabel(value) {
  const status = String(value || '').trim();
  if (!status) return UNSYNCED;
  return PUBLISH_STATUS_LABELS[status.toLowerCase()] || status;
}

export function formatComicTimestamp(value) {
  const raw = String(value || '').trim();
  if (!raw) return '';
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN', { hour12: false });
}

export function comicPagePresentation(node) {
  const config = readNodeConfig(node);
  const assetRef = readNodeField(node, 'assetRef') || {};
  const stale = config.configStale === true;
  const state = stale ? 'stale' : readNodeState(node, config);
  const url = String(assetRef.url || '').trim();
  return {
    id: String(readNodeField(node, 'id') || config.resourceId || '').trim(),
    title: readNodeTitle(node) || '漫画页',
    state,
    statusLabel: PAGE_STATE_LABELS[state] || '待处理',
    stale,
    hasImage: Boolean(url),
    url
  };
}

function comicCollectionPresentation(node, index) {
  const config = readNodeConfig(node);
  const assetRef = readNodeField(node, 'assetRef') || {};
  const pageCount = readPositiveNumber(config.pageCount);
  const n8Status = String(config.n8PublishStatus || '').trim();
  return {
    id: String(readNodeField(node, 'id') || config.resourceId || '').trim(),
    title: readNodeTitle(node) || `漫画合集 ${index + 1}`,
    pageCount,
    pageCountLabel: pageCount > 0 ? `${pageCount} 页` : UNSYNCED,
    statusLabel: comicPublishStatusLabel(n8Status),
    n8Status,
    url: String(assetRef.url || '').trim()
  };
}

function publishEntry(status, submittedAt) {
  return {
    status: String(status || '').trim(),
    label: comicPublishStatusLabel(status),
    submittedAt: formatComicTimestamp(submittedAt)
  };
}

function n8PublishEntry(collection) {
  if (!collection || !collection.n8Status) {
    return { status: '', label: UNSYNCED, publishedAt: '', carouselId: 0, workIds: [], detail: '' };
  }
  const source = collection.source || {};
  const carouselId = readPositiveNumber(source.n8PublishCarouselId);
  const workIds = Array.isArray(source.n8PublishWorkIds)
    ? source.n8PublishWorkIds.map(Number).filter(value => Number.isFinite(value) && value > 0)
    : [];
  const publishedAt = formatComicTimestamp(source.n8PublishedAt);
  const details = [];
  if (carouselId) details.push(`轮播 ${carouselId}`);
  if (workIds.length) details.push(`作品 ${workIds.length}`);
  if (publishedAt) details.push(publishedAt);
  return {
    status: collection.n8Status,
    label: comicPublishStatusLabel(collection.n8Status),
    publishedAt,
    carouselId,
    workIds,
    detail: details.join(' · ')
  };
}

function buildProgress(pages) {
  if (!pages.length) return { total: 0, generated: 0, running: 0, failed: 0, stale: 0, label: `进度：${UNSYNCED}` };
  const generated = pages.filter(page => page.hasImage).length;
  const running = pages.filter(page => page.state === 'running').length;
  const failed = pages.filter(page => page.state === 'failed').length;
  const stale = pages.filter(page => page.stale).length;
  const parts = [`已生成 ${generated}/${pages.length}`];
  if (running) parts.push(`生成中 ${running}`);
  if (failed) parts.push(`失败 ${failed}`);
  if (stale) parts.push(`需更新 ${stale}`);
  return { total: pages.length, generated, running, failed, stale, label: `进度：${parts.join(' · ')}` };
}

export function buildComicDeliverables(nodes = []) {
  const list = Array.isArray(nodes) ? nodes : [];
  const books = new Map();
  const ensureBook = (bookId, groupId) => {
    const key = bookId || groupId || 'comic';
    if (!books.has(key)) {
      books.set(key, {
        key,
        bookId,
        groupId,
        bookLabel: '',
        hasComic: false,
        pages: [],
        collections: [],
        mergeConfig: null
      });
    }
    return books.get(key);
  };

  for (const node of list) {
    const config = readNodeConfig(node);
    if (String(config.source || '').trim() !== 'batch_factory') continue;
    const resourceType = String(config.resourceType || '').trim();
    const bookId = String(config.bookId || '').trim();
    const groupId = readNodeGroupId(node);
    if (resourceType === 'book_script') {
      const book = ensureBook(bookId, groupId);
      if (!book.bookLabel) book.bookLabel = readNodeTitle(node);
      continue;
    }
    if (resourceType === COMIC_PAGE_RESOURCE) {
      const book = ensureBook(bookId, groupId);
      book.hasComic = true;
      book.pages.push(comicPagePresentation(node));
      continue;
    }
    if (resourceType === COMIC_COLLECTION_RESOURCE) {
      const book = ensureBook(bookId, groupId);
      book.hasComic = true;
      const collection = comicCollectionPresentation(node, book.collections.length);
      collection.source = config;
      book.collections.push(collection);
      continue;
    }
    if (resourceType === MERGE_RESOURCE) {
      const book = ensureBook(bookId, groupId);
      book.mergeConfig = config;
    }
  }

  const result = [];
  for (const book of books.values()) {
    if (!book.hasComic) continue;
    const mergeConfig = book.mergeConfig || {};
    const latestN8Collection = book.collections.length
      ? book.collections.reduce((latest, collection) => (collection.n8Status ? collection : latest), null)
      : null;
    result.push({
      key: book.key,
      bookId: book.bookId,
      bookLabel: book.bookLabel || (book.bookId ? `书籍 ${book.bookId}` : '漫画交付物'),
      pages: book.pages,
      collections: book.collections.map(({ source, ...collection }) => ({
        ...collection,
        n8Status: String(collection.n8Status || '').trim()
      })),
      progress: buildProgress(book.pages),
      publish: {
        n8: n8PublishEntry(latestN8Collection),
        provider121: publishEntry(mergeConfig.publish121Status, mergeConfig.publish121SubmittedAt),
        yd: publishEntry(mergeConfig.publishYadiStatus, mergeConfig.publishYadiSubmittedAt)
      }
    });
  }
  return result;
}
