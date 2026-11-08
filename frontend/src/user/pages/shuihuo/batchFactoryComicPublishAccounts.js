const COMIC_PUBLISH_DIRECTORY_CACHE_KEY = 'batch-factory.n8-comic-publish-directory';

// N8 后台当前已同步的启用分类。接口暂时未返回时先保证下拉框可用，
// 接口返回新目录后会立即替换这组兜底数据。
export const DEFAULT_AUDIO_CATEGORIES = [
  { id: 14, name: '现代悬疑', status: 'ENABLE' },
  { id: 13, name: '男频都市', status: 'ENABLE' },
  { id: 12, name: '古风通用', status: 'ENABLE' },
  { id: 11, name: '古风虐文', status: 'ENABLE' },
  { id: 10, name: '古风甜文', status: 'ENABLE' },
  { id: 9, name: '现代通用', status: 'ENABLE' },
  { id: 8, name: '现代虐文', status: 'ENABLE' }
];

export function producerAccountOptions(accounts = [], savedAccount = '', organizationId = '', accountOrganizations = {}, loading = false) {
  const saved = String(savedAccount || '').trim();
  const org = String(organizationId || '').trim();
  const hasDirectoryMapping = Object.keys(accountOrganizations || {}).length > 0;
  const belongsToOrganization = value => {
    if (!hasDirectoryMapping || !org) return true;
    const mapped = accountOrganizations?.[value];
    return !Array.isArray(mapped) || mapped.length === 0 || mapped.includes(org);
  };
  const values = [];
  const seen = new Set();
  if (saved) {
    seen.add(saved);
    const isInDirectory = accounts.some(item => String(item || '').trim() === saved);
    const savedLabel = isInDirectory
      ? (belongsToOrganization(saved) ? `${saved}（已保存）` : `${saved}（不属于当前组织）`)
      // 目录还在读取时不能显示成“目录未同步”，否则看起来像账号缺失。
      : (loading ? `${saved}（已保存·目录读取中…）` : `${saved}（已保存·目录未同步）`);
    values.push({ value: saved, label: savedLabel });
  }
  for (const item of accounts || []) {
    const value = String(item || '').trim();
    if (!value || seen.has(value) || !belongsToOrganization(value)) continue;
    seen.add(value);
    values.push({ value, label: value });
  }
  return values;
}

export function readComicPublishDirectoryCache() {
  try {
    const raw = window.localStorage.getItem(COMIC_PUBLISH_DIRECTORY_CACHE_KEY);
    const parsed = raw ? JSON.parse(raw) : null;
    if (!parsed || typeof parsed !== 'object') return null;
    return {
      producerAccounts: Array.isArray(parsed.producerAccounts) ? parsed.producerAccounts : [],
      organizations: Array.isArray(parsed.organizations) ? parsed.organizations : [],
      audioCategories: Array.isArray(parsed.audioCategories) ? parsed.audioCategories : [],
      producerAccountOrganizations: parsed.producerAccountOrganizations && typeof parsed.producerAccountOrganizations === 'object' ? parsed.producerAccountOrganizations : {}
    };
  } catch {
    return null;
  }
}

export function chooseN8Audio(audios = [], savedID = '', random = Math.random) {
  const enabled = (audios || []).filter(item => String(item?.status || 'ENABLE').toUpperCase() === 'ENABLE');
  const saved = enabled.find(item => String(item?.id) === String(savedID));
  if (saved) return saved;
  if (!enabled.length) return null;
  const index = Math.min(enabled.length - 1, Math.max(0, Math.floor(Number(random()) * enabled.length)));
  return enabled[index] || null;
}

export function writeComicPublishDirectoryCache(directory) {
  try {
    window.localStorage.setItem(COMIC_PUBLISH_DIRECTORY_CACHE_KEY, JSON.stringify({
      producerAccounts: Array.isArray(directory?.producerAccounts) ? directory.producerAccounts : [],
      organizations: Array.isArray(directory?.organizations) ? directory.organizations : [],
      audioCategories: Array.isArray(directory?.audioCategories) ? directory.audioCategories : [],
      producerAccountOrganizations: directory?.producerAccountOrganizations && typeof directory.producerAccountOrganizations === 'object' ? directory.producerAccountOrganizations : {}
    }));
  } catch {
    // 本地缓存不可用时不影响 N8 实时读取。
  }
}
