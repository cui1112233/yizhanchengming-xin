export const privateComicDraftKey = 'private-customization-comic-draft';
export const privateComicTaskKey = 'private-customization-comic-task';
export const privateComicMinSourceLength = 20;

function text(value) {
  return String(value ?? '').trim();
}

export function privateComicTitle(sourceText = '', title = '') {
  const explicit = text(title);
  if (explicit) return explicit;
  const firstLine = String(sourceText ?? '')
    .split(/\r?\n/)
    .map(line => line.trim())
    .find(Boolean);
  return firstLine ? firstLine.slice(0, 80) : '未命名漫画';
}

export function buildPrivateComicSettings({ sourceText = '', comic = {}, textModelId = '', imageModelId = '', imageProvider = '' } = {}) {
  const normalizedSource = text(sourceText);
  if (!normalizedSource) throw new Error('请先粘贴漫画正文');
  if (Array.from(normalizedSource).length < privateComicMinSourceLength) throw new Error(`漫画正文至少需要 ${privateComicMinSourceLength} 个字`);
  const normalizedComic = comic && typeof comic === 'object' ? { ...comic } : {};
  return {
    productionMode: 'comic',
    // 单本精致制作始终允许只用资产文字提示词继续；资产图仅作为可选参考图。
    assetImagesEnabled: false,
    textModelId: text(textModelId),
    imageModelId: text(imageModelId),
    imageProvider: text(imageProvider),
    aiPromptConfig: {
      comic: {
        ...normalizedComic,
        enabled: true,
        scope: 'private_customization'
      }
    }
  };
}

export function mergePrivateComicReference(settings = {}, reference = {}) {
  const currentComic = settings?.aiPromptConfig?.comic && typeof settings.aiPromptConfig.comic === 'object'
    ? settings.aiPromptConfig.comic
    : {};
  const currentReference = currentComic.globalReferenceImage && typeof currentComic.globalReferenceImage === 'object'
    ? currentComic.globalReferenceImage
    : {};
  const uploadedReference = reference && typeof reference === 'object' ? reference : {};
  return {
    ...settings,
    aiPromptConfig: {
      ...(settings.aiPromptConfig || {}),
      comic: {
        ...currentComic,
        globalReferenceImage: {
          ...currentReference,
          ...uploadedReference,
          source: 'manual-upload',
          enabled: currentReference.enabled !== false,
          applyToAssets: currentReference.applyToAssets !== false,
          applyToPages: currentReference.applyToPages !== false
        }
      }
    }
  };
}

export function shouldReusePrivateComicTask(task, sourceText) {
  if (!task?.batchId || !task?.bookId) return false;
  const previous = text(task?.sourceText);
  if (!previous) return true;
  return previous === text(sourceText);
}

export function privateComicTaskSnapshot(result = {}, sourceText = '') {
  const snapshot = {
    batchId: text(result?.batch?.id || result?.batchId),
    bookId: text(result?.book?.id || result?.bookId),
    status: 'created'
  };
  const source = text(sourceText);
  if (source) snapshot.sourceText = source;
  return snapshot;
}
