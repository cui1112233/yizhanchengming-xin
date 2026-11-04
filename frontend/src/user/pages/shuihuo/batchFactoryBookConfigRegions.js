export const BOOK_CONFIG_REGIONS = [
  { key: 'engine', label: '引擎配置' },
  { key: 'assets', label: '资产设置' },
  { key: 'constraints', label: '约束设置' },
  { key: 'video', label: '视频设置' },
  { key: 'visual', label: '画面设置' }
];

const ENGINE_KEYS = ['textModelId', 'imageModelId', 'videoModelId', 'videoProvider', 'aspectRatio', 'imageAspectRatio', 'videoAspectRatio', 'videoResolution', 'productionMode', 'storyboardDurationLimit', 'maxVideoDuration', 'audioSource', 'audioPlanningEnabled', 'audioMergeEnabled', 'audioDurationSeconds', 'publishSettings', 'publishRewriteEnabled'];
const AI_MODULE_BY_REGION = {
  assets: 'assets',
  constraints: 'constraints',
  video: 'video',
  visual: 'visual'
};

const sameJSON = (left, right) => JSON.stringify(left ?? null) === JSON.stringify(right ?? null);

// 徽标跟着运行时合并链走：运行时按“单书显式值 > 批次统一配置”取值，只有单书
// patch 里真的写了键才算本书自有值；值与批次一致时说明来自最近一次“保存统一
// 配置”的批量推送，因此标成“来自统一配置”而不是“已单书覆盖”。
export function bookConfigRegionStatus(book, region, batchPatch = {}) {
  const patch = book?.settingsState?.patch || {};
  const pushedKeys = keys => keys.every(key => sameJSON(patch[key], batchPatch?.[key]));
  const ownStatus = (keys, label = '已单书覆盖') => pushedKeys(keys)
    ? { label: '来自统一配置', tone: 'pushed' }
    : { label, tone: 'overridden' };
  if (region === 'engine') {
    const keys = ENGINE_KEYS.filter(key => Object.hasOwn(patch, key));
    if (keys.length) return ownStatus(keys);
    const modules = ['video', 'comic'].filter(key => Object.hasOwn(patch?.aiPromptConfig || {}, key));
    if (modules.length) {
      const pushed = modules.every(key => sameJSON(patch.aiPromptConfig?.[key], batchPatch?.aiPromptConfig?.[key]));
      return pushed ? { label: '来自统一配置', tone: 'pushed' } : { label: '已单书覆盖', tone: 'overridden' };
    }
    return { label: '继承作品配置', tone: 'inherited' };
  }
  const module = patch?.aiPromptConfig?.[AI_MODULE_BY_REGION[region]];
  if (region === 'assets' && Object.hasOwn(patch, 'starredCharacterNames')) return ownStatus(['starredCharacterNames']);
  if (!module) return { label: '继承作品配置', tone: 'inherited' };
  if (module.enabled === false) return { label: '本书未启用', tone: 'disabled' };
  const pushedModule = sameJSON(module, batchPatch?.aiPromptConfig?.[AI_MODULE_BY_REGION[region]]);
  if (region === 'constraints') {
    const count = Array.isArray(module.selections) ? module.selections.filter(Boolean).length : 0;
    if (count > 0) return pushedModule
      ? { label: `${count} 条约束 · 来自统一配置`, tone: 'pushed' }
      : { label: `${count} 条约束已启用`, tone: 'overridden' };
  }
  return pushedModule ? { label: '来自统一配置', tone: 'pushed' } : { label: '已单书覆盖', tone: 'overridden' };
}

export function bookAssetSummary(book) {
  const assets = book?.assets || {};
  return `人物 ${assets.characters?.length || 0} · 场景 ${assets.scenes?.length || 0} · 道具 ${assets.props?.length || 0}`;
}
