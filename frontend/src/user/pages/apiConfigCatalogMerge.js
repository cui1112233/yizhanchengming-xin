// 「API 配置」把模型目录条目与渠道条目合进同一组展示时的去重口径。
// 后端 /api/models 合成下拉时也按「地址 + 模型名」判断同一个真实模型，
// 这里保持同一条规则，避免页面上合并了、下拉里却还出现两条。

export function normalizeEndpoint(value) {
  const text = String(value ?? '').trim().toLowerCase();
  if (!text) return '';
  return text.replace(/\/+$/, '');
}

export function normalizeModelName(value) {
  return String(value ?? '').trim().toLowerCase();
}

// 任一为空时返回空串：无法证明是同一个模型，宁可都保留也不误合并。
export function modelIdentity(baseUrl, model) {
  const endpoint = normalizeEndpoint(baseUrl);
  const name = normalizeModelName(model);
  if (!endpoint || !name) return '';
  return `${endpoint}|${name}`;
}

// 已下线模型的命名约定：[off] 前缀。这类条目即使还留有凭证也不进界面与下拉。
export function isRetiredModel(model) {
  return ['displayName', 'modelId', 'name'].some(key => (
    String(model?.[key] ?? '').trim().toLowerCase().startsWith('[off]')
  ));
}

// 本地执行器不依赖 API Key（靠配对心跳），没有凭证也应算可用。
export function modelNeedsNoCredential(model) {
  const marker = [model?.provider, model?.adapterKind]
    .map(value => String(value ?? '').trim().toLowerCase())
    .join('|');
  if (marker.includes('local_executor') || marker.includes('local-executor')) return true;
  return String(model?.id ?? '').trim().toLowerCase().startsWith('local-');
}

// 目录模型可用性：服务端下发 available 时以它为准（页面与 /api/models 同口径）；
// 没有该字段（本地预览假后端）时按同一条规则在前端推导。
export function isCatalogModelAvailable(model) {
  if (!model) return false;
  if (typeof model.available === 'boolean') return model.available;
  if (isRetiredModel(model)) return false;
  if (model.enabled === false) return false;
  if (model.hasCredential === true) return true;
  return modelNeedsNoCredential(model);
}

// 平台托管的模型：预设面板里的条目（按面板 id 动态判断，新增预设自动生效）
// 与本地执行器（在引擎设置里维护）。这类模型不重复出现在文本/视频/图片三组。
export function isPlatformManagedModel(model, platformPresetIds) {
  if (!model) return false;
  const ids = platformPresetIds instanceof Set ? platformPresetIds : new Set(platformPresetIds || []);
  return ids.has(model.id) || modelNeedsNoCredential(model);
}

// 用户自建的目录模型：可以打开「自定义模型」表单改地址 / 模型名 / 展示名 / Key。
// 判定全部依赖数据字段，不写死 id：
// - 平台预设面板里的条目、本地执行器 → 由平台托管，不在组里出现；
// - 带 adapterKind 的适配器模型 → 由管理端维护，个人中心不提供通用表单；
// - 没有地址或模型名的条目 → 通用表单无处可写。
export function isUserManagedCatalogModel(model, platformPresetIds) {
  if (!model) return false;
  if (isPlatformManagedModel(model, platformPresetIds)) return false;
  if (model.builtIn === true) return false;
  if (String(model.adapterKind ?? '').trim() !== '') return false;
  if (!String(model.baseUrl ?? '').trim() || !String(model.modelId ?? '').trim()) return false;
  return true;
}

// catalogModels: /api/config/models 返回的条目（含 enabled / hasCredential）
// channelEntries: [{ id, name, baseUrl, model, entry }]
// 返回顺序固定为「目录模型在前、渠道在后」，渠道只有在与目录模型同源时才被合并掉。
export function mergeGroupEntries({ catalogModels = [], channelEntries = [] } = {}) {
  const catalog = [];
  const mergedChannels = [];
  const channels = [];
  const seen = new Map();
  let catalogDuplicates = 0;

  for (const model of Array.isArray(catalogModels) ? catalogModels : []) {
    if (!model) continue;
    const identity = modelIdentity(model.baseUrl, model.modelId);
    if (identity && seen.has(identity)) {
      catalogDuplicates += 1;
      continue;
    }
    if (identity) seen.set(identity, { kind: 'catalog', id: model.id });
    catalog.push({ ...model, source: 'catalog', identity });
  }

  for (const channel of Array.isArray(channelEntries) ? channelEntries : []) {
    if (!channel) continue;
    const identity = modelIdentity(channel.baseUrl, channel.model);
    if (identity && seen.has(identity)) {
      mergedChannels.push({ ...channel, identity });
      continue;
    }
    if (identity) seen.set(identity, { kind: 'channel', id: channel.id });
    channels.push({ ...channel, source: 'channel', identity });
  }

  return { catalog, channels, mergedChannels, catalogDuplicates };
}
