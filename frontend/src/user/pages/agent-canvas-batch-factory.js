const BOOK_GROUP_PREFIX = 'book:';

export function bookGroupId(bookId) {
  const value = String(bookId || '').trim();
  return value ? `${BOOK_GROUP_PREFIX}${value}` : '';
}

export function bookIdFromGroupId(groupId) {
  const value = String(groupId || '').trim();
  return value.startsWith(BOOK_GROUP_PREFIX) ? value.slice(BOOK_GROUP_PREFIX.length) : '';
}

export function agentCanvasURL({ projectId, batchId, bookId } = {}) {
  const query = new URLSearchParams();
  query.set('taskId', String(projectId || '').trim());
  query.set('batchId', String(batchId || '').trim());
  query.set('bookId', String(bookId || '').trim());
  query.set('view', 'workflow');
  return `/agent/canvas?${query.toString()}`;
}

export function isBatchFactoryNode(node) {
  return Boolean(bookIdFromGroupId(node?.groupId ?? node?.data?.groupId));
}

export function projectedNodePresentation(node) {
  const groupId = String(node?.groupId ?? node?.data?.groupId ?? '').trim();
  if (!bookIdFromGroupId(groupId)) return { groupId, groupLabel: '', sourceLabel: '' };
  const groupLabel = {
    script: '书籍正文',
    text: '书籍文本',
    image: '书籍素材',
    video: '书籍视频',
    audio: '书籍音频'
  }[String(node?.type ?? node?.data?.type ?? '').trim()] || '书籍内容';
  return { groupId, groupLabel, sourceLabel: '批量工厂' };
}

export function batchFactoryGroupLabel(groupId) {
  const bookId = bookIdFromGroupId(groupId);
  return bookId ? `书籍分组 ${bookId}` : String(groupId || '').trim();
}
