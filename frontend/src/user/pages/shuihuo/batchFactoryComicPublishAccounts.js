const COMIC_PUBLISH_DIRECTORY_CACHE_KEY = 'batch-factory.n8-comic-publish-directory';

// N8 后台当前已同步的启用分类。接口暂时未返回时先保证下拉框可用，
// 接口返回新目录后会立即替换这组兜底数据。
export const DEFAULT_AUDIO_CATEGORIES = [
  { id: 12, name: '古风-爽文（打脸）', status: 'ENABLE' },
  { id: 11, name: '古风-男悲伤', status: 'ENABLE' },
  { id: 10, name: '古风-女悲伤', status: 'ENABLE' },
  { id: 9, name: '女-爽文（打脸）', status: 'ENABLE' },
  { id: 8, name: '女，悲伤', status: 'ENABLE' }
];

export function producerAccountOptions(accounts = [], savedAccount = '', organizationId = '', accountOrganizations = {}) {
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
    values.push({ value: saved, label: isInDirectory ? (belongsToOrganization(saved) ? `${saved}（已保存）` : `${saved}（不属于当前组织）`) : `${saved}（已保存·目录未同步）` });
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

export function chooseN8Audio(audios = [], savedID = '') {
  const enabled = (audios || []).filter(item => String(item?.status || 'ENABLE').toUpperCase() === 'ENABLE');
  const saved = enabled.find(item => String(item?.id) === String(savedID));
  if (saved) return saved;
  return enabled.sort((a, b) => {
    const updated = new Date(b?.updatedAt || b?.updated_at || 0) - new Date(a?.updatedAt || a?.updated_at || 0);
    return updated || Number(b?.id || 0) - Number(a?.id || 0);
  })[0] || null;
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
