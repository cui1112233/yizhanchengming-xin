export const LONG_IMITATION_REFERENCE_LIMIT = 2_000_000;
export const longImitationCharacters = text => String(text || '').replace(/\s/g, '').length;
export const emptyLongImitationDraft = () => ({
  title: '', referenceText: '', targetWordCount: 5000, targetChapters: 100,
  readerPreferences: '', storyIdea: '', selectedDirection: '', supplementaryInfo: '',
  knowledgeBookIds: [], memoKeys: [], files: [], chapterKeys: []
});
export const imitationChapters = files => files.flatMap(file => file.chapters.map(chapter => ({ ...chapter, key: `${file.key}:${chapter.id}`, filename: file.filename })));
export function imitationChapterCharacters(chapter) {
  const content = chapter.content.trim();
  return longImitationCharacters(chapter.title && content.startsWith(chapter.title) ? content.slice(chapter.title.length) : content);
}
export function selectedImitationText(chapters, keys) {
  const selected = new Set(keys);
  return chapters.filter(chapter => selected.has(chapter.key)).map(chapter => chapter.content.trim()).join('\n\n');
}
export function recentImitationChapters(chapters, count) {
  return chapters.slice(-count).map(chapter => chapter.key);
}
export function validateLongImitationDraft(draft, modelId) {
  if (!modelId) return '请选择当前可用的文本模型';
  if (longImitationCharacters(draft.referenceText) < 100) return '参考原文至少需要 100 字（不计空白）';
  if (draft.referenceText.length > LONG_IMITATION_REFERENCE_LIMIT) return '参考原文最多支持 200 万字元，请缩小选择范围';
  if (!Number.isInteger(draft.targetWordCount) || draft.targetWordCount < 1000 || draft.targetWordCount > 50000) return '每章字数应为 1000 至 50000 的整数';
  if (!Number.isInteger(draft.targetChapters) || draft.targetChapters < 1 || draft.targetChapters > 500) return '目标章节数应为 1 至 500 的整数';
  return '';
}
