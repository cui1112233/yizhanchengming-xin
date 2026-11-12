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

export function buildPrivateComicSettings({ sourceText = '', comic = {}, textModelId = '', imageModelId = '', imageProvider = '', assetImages = true } = {}) {
  const normalizedSource = text(sourceText);
  if (!normalizedSource) throw new Error('请先粘贴漫画正文');
  if (Array.from(normalizedSource).length < privateComicMinSourceLength) throw new Error(`漫画正文至少需要 ${privateComicMinSourceLength} 个字`);
  const normalizedComic = comic && typeof comic === 'object' ? { ...comic } : {};
  return {
    productionMode: 'comic',
    // 私人定制默认必须生成资产图（角色/场景参考图），否则每一页都没有参考图，角色一致性无从保证。
    // 仅当调用方显式传 assetImages: false 时才关闭。
    assetImagesEnabled: assetImages !== false,
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
