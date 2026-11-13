import { manualBookIDsFromInput, upsertPlatformGroup } from './batchFactoryManualFetch.js';

export const CUSTOM_COPY_OPTION = { value: 'custom_copy', label: '文案获取' };

export function selectedCopyGroups(groups, results, selections, options = []) {
  let next = groups;
  for (const result of results) {
    const copy = result.copies?.find(item => item.id === selections[result.bookId]);
    if (!copy || copy.bookId !== result.bookId || !copy.content?.trim() || !/^\d+$/.test(copy.platform)) {
      throw new Error(`请为 ${result.bookId} 选择一份有效文案`);
    }
    const platformName = options.find(item => item.value === copy.platform)?.label || `书城 ${copy.platform}`;
    next = upsertPlatformGroup(next, { platformId: copy.platform, platformName, inputText: copy.bookId });
    next = next.map(group => group.platformId !== copy.platform ? group : {
      ...group,
      sourceTextByBookId: { ...group.sourceTextByBookId, [copy.bookId]: copy.content },
      sourceMetadataByBookId: { ...group.sourceMetadataByBookId, [copy.bookId]: {
        bookName: copy.title, sourceMode: 'custom_copy', customCopyRecordId: copy.id,
        customCopyCreatedAt: copy.createdAt, customCopyPlatform: copy.platform
      } }
    });
  }
  return next;
}

export function copyInputIDs(input) {
  const lines = String(input).split(/\r?\n/).filter(line => line.trim());
  if (!lines.length || lines.some(line => !/^\s*\d{6,25}\s*$/.test(line))) throw new Error('每行只填写一个书籍 ID');
  const ids = manualBookIDsFromInput(input);
  if (ids.length > 50) throw new Error('一次最多获取 50 本书');
  return ids;
}
