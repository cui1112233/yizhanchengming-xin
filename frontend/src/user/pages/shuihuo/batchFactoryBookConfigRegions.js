export const BOOK_CONFIG_REGIONS = [
  { key: 'engine', label: '引擎配置' },
  { key: 'assets', label: '资产设置' },
  { key: 'constraints', label: '约束设置' },
  { key: 'video', label: '视频设置' },
  { key: 'visual', label: '画面设置' }
];

const ENGINE_KEYS = ['textModelId', 'imageModelId', 'videoModelId', 'videoProvider', 'aspectRatio', 'imageAspectRatio', 'videoAspectRatio', 'videoResolution', 'productionMode', 'storyboardDurationLimit', 'maxVideoDuration', 'fixedSingleVideo', 'audioSource', 'audioPlanningEnabled', 'audioMergeEnabled', 'audioDurationSeconds', 'publishSettings', 'publishRewriteEnabled'];
const AI_MODULE_BY_REGION = {
  assets: 'assets',
  constraints: 'constraints',
  video: 'video',
  visual: 'visual'
};

export function bookConfigRegionStatus(book, region) {
  const patch = book?.settingsState?.patch || {};
  if (region === 'engine') {
    return ENGINE_KEYS.some(key => Object.hasOwn(patch, key))
      ? { label: '已单书覆盖', tone: 'overridden' }
      : { label: '继承作品配置', tone: 'inherited' };
  }
  const module = patch?.aiPromptConfig?.[AI_MODULE_BY_REGION[region]];
  if (region === 'assets' && Object.hasOwn(patch, 'starredCharacterNames')) return { label: '已单书覆盖', tone: 'overridden' };
  if (!module) return { label: '继承作品配置', tone: 'inherited' };
  if (module.enabled === false) return { label: '本书未启用', tone: 'disabled' };
  if (region === 'constraints') {
    const count = Array.isArray(module.selections) ? module.selections.filter(Boolean).length : 0;
    if (count > 0) return { label: `${count} 条约束已启用`, tone: 'overridden' };
  }
  return { label: '已单书覆盖', tone: 'overridden' };
}

export function bookAssetSummary(book) {
  const assets = book?.assets || {};
  return `人物 ${assets.characters?.length || 0} · 场景 ${assets.scenes?.length || 0} · 道具 ${assets.props?.length || 0}`;
}
