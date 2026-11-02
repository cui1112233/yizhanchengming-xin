function createId() {
  if (typeof crypto?.randomUUID === 'function') return crypto.randomUUID();
  return `entity-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

export function createEntity(data, id = createId()) {
  return { id, data };
}

function isEntityRecord(value) {
  return value && typeof value === 'object'
    && typeof value.id === 'string'
    && Object.hasOwn(value, 'data');
}

function normalizeEntityList(value) {
  const items = Array.isArray(value) ? value : [];
  const ids = new Set();
  return items.map(item => {
    const record = isEntityRecord(item) && !ids.has(item.id) ? item : createEntity(entityData(item));
    ids.add(record.id);
    return record;
  });
}

export function normalizeExtractInfo(value) {
  const characters = normalizeEntityList(value?.characters);
  const scenes = normalizeEntityList(value?.scenes);
  const characterIds = new Set(characters.map(item => item.id));
  return {
    characters,
    scenes,
    // 提取流程会返回一条由题材决定的统一风格。它是画面前缀的
    // 动态来源，不能在剧本生成时丢弃，否则只能使用静态的“真人实拍”。
    visualStyle: typeof value?.visualStyle === 'string' ? value.visualStyle.trim() : '',
    protagonistIds: Array.isArray(value?.protagonistIds)
      ? value.protagonistIds.filter(id => characterIds.has(id))
      : []
  };
}

export function entityData(record) {
  return isEntityRecord(record) ? record.data : record;
}

export function collectProtagonists(extractInfo) {
  const normalized = normalizeExtractInfo(extractInfo);
  const protagonistIds = new Set(normalized.protagonistIds);
  return normalized.characters
    .filter(item => protagonistIds.has(item.id))
    .map(entityData);
}

function entityName(record) {
  const data = entityData(record);
  if (typeof data === 'string') return data.trim();
  return String(data?.角色名称 || data?.名称 || data?.name || '').trim();
}

function occurrenceCount(text, value) {
  if (value.length < 2) return 0;
  return String(text || '').split(value).length - 1;
}

// 提取结果没有明确主角字段时，至少选择原文中最常出现的人物。
// 若人物资料已写明“主角 / 男主 / 女主 / 视角角色”，则优先保留所有这些明确主角。
export function selectDefaultProtagonistIds(extractInfo, novelText = '') {
  const normalized = normalizeExtractInfo(extractInfo);
  if (normalized.protagonistIds.length || !normalized.characters.length) return normalized.protagonistIds;
  const explicit = normalized.characters.filter(item => {
    const data = entityData(item);
    const role = typeof data === 'object' ? JSON.stringify(data) : String(data || '');
    return /(主角|男主|女主|主人公|第一视角|protagonist)/i.test(role);
  }).map(item => item.id);
  if (explicit.length) return explicit;
  const selected = normalized.characters.reduce((best, item) => {
    const score = occurrenceCount(novelText, entityName(item));
    return score > best.score ? { id: item.id, score } : best;
  }, { id: normalized.characters[0].id, score: occurrenceCount(novelText, entityName(normalized.characters[0])) });
  return selected.id ? [selected.id] : [];
}

export function toGenerationEntities(extractInfo) {
  const normalized = normalizeExtractInfo(extractInfo);
  return {
    characters: normalized.characters.map(entityData),
    scenes: normalized.scenes.map(entityData),
    visualStyle: normalized.visualStyle,
    protagonists: collectProtagonists(normalized)
  };
}
