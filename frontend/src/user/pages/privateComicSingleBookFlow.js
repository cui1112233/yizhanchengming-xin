import {
  buildPrivateComicSettings,
  privateComicMinSourceLength,
  privateComicTaskSnapshot,
  privateComicTitle,
  shouldReusePrivateComicTask
} from './privateCustomizationComic.js';

export function canExtractPrivateComicAssets({ sourceText = '', busy = false, action = '' } = {}) {
  return Array.from(String(sourceText || '').trim()).length >= privateComicMinSourceLength
    && !busy
    && !action;
}

export async function ensurePrivateComicTask({ task = null, draft = {}, createTask } = {}) {
  const initialBatchSettings = buildPrivateComicSettings({
    sourceText: draft.sourceText,
    comic: draft.comic,
    textModelId: draft.textModelId,
    imageModelId: draft.imageModelId,
    imageProvider: draft.imageProvider
  });
  if (shouldReusePrivateComicTask(task, draft.sourceText)) {
    return { task, created: false, settings: initialBatchSettings };
  }
  if (typeof createTask !== 'function') throw new Error('漫画任务创建方法不可用');
  const created = await createTask({
    title: privateComicTitle(draft.sourceText, draft.title),
    sourceText: draft.sourceText,
    initialBatchSettings
  });
  const snapshot = privateComicTaskSnapshot(created, draft.sourceText);
  if (!snapshot.batchId || !snapshot.bookId) throw new Error('漫画任务已创建，但没有返回完整任务编号');
  return { task: snapshot, created: true, settings: initialBatchSettings };
}

export async function preparePrivateComicTaskSettings({ ensured, saveSettings } = {}) {
  if (!ensured?.task?.batchId) throw new Error('漫画任务编号不完整');
  if (!ensured.created) {
    if (typeof saveSettings !== 'function') throw new Error('漫画设置保存方法不可用');
    await saveSettings(ensured.task.batchId, ensured.settings);
  }
  return ensured.task;
}
