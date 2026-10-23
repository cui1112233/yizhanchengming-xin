const CATEGORY_OPTIONS = Object.freeze({
  text: Object.freeze([
    { value: 'all', label: '全部' },
    { value: 'script', label: '剧本' },
    { value: 'other_text', label: '其他' }
  ]),
  audio: Object.freeze([
    { value: 'all', label: '全部' },
    { value: 'voice', label: '配音' },
    { value: 'music', label: '音乐' },
    { value: 'effect', label: '音效' }
  ]),
  image: Object.freeze([
    { value: 'all', label: '全部' },
    { value: 'character', label: '角色' },
    { value: 'scene', label: '场景' },
    { value: 'object', label: '物品' },
    { value: 'other_image', label: '其他' }
  ]),
  video: Object.freeze([
    { value: 'all', label: '全部' },
    { value: 'clip', label: '视频片段' },
    { value: 'final_video', label: '成片' }
  ])
});

const DEFAULT_CATEGORY = Object.freeze({
  text: 'other_text',
  audio: 'voice',
  image: 'other_image',
  video: 'clip'
});

export const STORYBOARD_LIBRARY_DEFAULT_WIDTH = 340;
export const STORYBOARD_LIBRARY_MIN_WIDTH = 280;
export const STORYBOARD_LIBRARY_MAX_WIDTH = 520;
export const STORYBOARD_SECTION_DEFAULT_HEIGHTS = Object.freeze({ text: 220, audio: 240, image: 320 });
export const STORYBOARD_SECTION_MIN_HEIGHT = 120;
export const STORYBOARD_SECTION_MAX_HEIGHT = 520;

export function storyboardLibraryWidth(value, maxWidth = STORYBOARD_LIBRARY_MAX_WIDTH) {
  const numeric = Number(value);
  const candidate = Number.isFinite(numeric) ? numeric : STORYBOARD_LIBRARY_DEFAULT_WIDTH;
  const safeMaximum = Math.max(
    STORYBOARD_LIBRARY_MIN_WIDTH,
    Math.min(STORYBOARD_LIBRARY_MAX_WIDTH, Number(maxWidth) || STORYBOARD_LIBRARY_MAX_WIDTH)
  );
  return Math.round(Math.max(STORYBOARD_LIBRARY_MIN_WIDTH, Math.min(safeMaximum, candidate)));
}

export function storyboardSectionHeight(kind, value) {
  const fallback = STORYBOARD_SECTION_DEFAULT_HEIGHTS[kind] || STORYBOARD_SECTION_DEFAULT_HEIGHTS.text;
  const numeric = Number(value);
  const candidate = Number.isFinite(numeric) ? numeric : fallback;
  return Math.round(Math.max(STORYBOARD_SECTION_MIN_HEIGHT, Math.min(STORYBOARD_SECTION_MAX_HEIGHT, candidate)));
}

export const DEFAULT_STORYBOARD_STATE = Object.freeze({
  directorMode: 'supervised',
  selectedKey: '',
  filters: Object.freeze({ text: 'all', audio: 'all', image: 'all', video: 'all' }),
  categories: Object.freeze({}),
  order: Object.freeze([]),
  libraryWidth: STORYBOARD_LIBRARY_DEFAULT_WIDTH,
  sectionHeights: STORYBOARD_SECTION_DEFAULT_HEIGHTS
});

function safeObject(value) {
  return value && typeof value === 'object' && !Array.isArray(value) ? value : {};
}

function allowedCategory(kind, value, includeAll = false) {
  const options = CATEGORY_OPTIONS[kind] || CATEGORY_OPTIONS.text;
  return options.some(option => option.value === value && (includeAll || option.value !== 'all'));
}

export function storyboardCategoryOptions(kind, { includeAll = true } = {}) {
  const options = CATEGORY_OPTIONS[kind] || CATEGORY_OPTIONS.text;
  return options.filter(option => includeAll || option.value !== 'all').map(option => ({ ...option }));
}

export function normalizeStoryboardState(value = {}) {
  const source = safeObject(value);
  const sourceFilters = safeObject(source.filters);
  const filters = {};
  for (const kind of Object.keys(CATEGORY_OPTIONS)) {
    const candidate = typeof sourceFilters[kind] === 'string' ? sourceFilters[kind] : 'all';
    filters[kind] = allowedCategory(kind, candidate, true) ? candidate : 'all';
  }

  const categories = {};
  for (const [key, category] of Object.entries(safeObject(source.categories))) {
    if (!key || typeof category !== 'string') continue;
    if (Object.values(CATEGORY_OPTIONS).some(options => options.some(option => option.value === category && category !== 'all'))) {
      categories[key] = category;
    }
  }

  const seen = new Set();
  const order = [];
  for (const key of Array.isArray(source.order) ? source.order : []) {
    if (typeof key !== 'string' || !key || seen.has(key)) continue;
    seen.add(key);
    order.push(key);
  }

  const sourceSectionHeights = safeObject(source.sectionHeights);
  const sectionHeights = {};
  for (const kind of Object.keys(STORYBOARD_SECTION_DEFAULT_HEIGHTS)) {
    sectionHeights[kind] = storyboardSectionHeight(kind, sourceSectionHeights[kind]);
  }

  return {
    directorMode: source.directorMode === 'automatic' ? 'automatic' : 'supervised',
    selectedKey: typeof source.selectedKey === 'string' ? source.selectedKey : '',
    filters,
    categories,
    order,
    libraryWidth: storyboardLibraryWidth(source.libraryWidth),
    sectionHeights
  };
}

export function patchStoryboardState(current = {}, patch = {}) {
  const normalized = normalizeStoryboardState(current);
  const nextPatch = safeObject(patch);
  return normalizeStoryboardState({
    ...normalized,
    ...nextPatch,
    filters: { ...normalized.filters, ...safeObject(nextPatch.filters) },
    categories: { ...normalized.categories, ...safeObject(nextPatch.categories) },
    sectionHeights: { ...normalized.sectionHeights, ...safeObject(nextPatch.sectionHeights) }
  });
}

export function toggleStoryboardBatchSelection(selectedKeys = [], key = '') {
  const normalizedKey = typeof key === 'string' ? key.trim() : '';
  const current = [];
  const seen = new Set();
  for (const value of Array.isArray(selectedKeys) ? selectedKeys : []) {
    const normalized = typeof value === 'string' ? value.trim() : '';
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    current.push(normalized);
  }
  if (!normalizedKey) return current;
  if (seen.has(normalizedKey)) return current.filter(value => value !== normalizedKey);
  return [...current, normalizedKey];
}

export function storyboardBatchCategoryPatch(items = [], selectedKeys = [], category = '') {
  const selected = new Set((Array.isArray(selectedKeys) ? selectedKeys : []).filter(value => typeof value === 'string' && value));
  const targets = (Array.isArray(items) ? items : []).filter(item => selected.has(item?.key));
  if (!targets.length) return null;
  const kinds = new Set(targets.map(item => item?.kind));
  if (kinds.size !== 1) return null;
  const [kind] = kinds;
  if (!allowedCategory(kind, category)) return null;
  return Object.fromEntries(targets.map(item => [item.key, category]));
}

export function storyboardBatchAssets(items = [], selectedKeys = []) {
  const selected = new Set((Array.isArray(selectedKeys) ? selectedKeys : []).filter(value => typeof value === 'string' && value));
  const seen = new Set();
  const assets = [];
  for (const item of Array.isArray(items) ? items : []) {
    if (!selected.has(item?.key)) continue;
    const asset = item?.asset;
    const id = String(asset?.id || asset?.assetId || '').trim();
    if (!id || seen.has(id)) continue;
    seen.add(id);
    assets.push(asset);
  }
  return assets;
}

export function storyboardItemCategory(item, state = {}) {
  const kind = Object.hasOwn(CATEGORY_OPTIONS, item?.kind) ? item.kind : 'text';
  const normalized = normalizeStoryboardState(state);
  const override = normalized.categories[item?.key];
  if (allowedCategory(kind, override)) return override;

  const stored = item?.asset?.category || item?.config?.storyboardCategory;
  if (allowedCategory(kind, stored)) return stored;
  if (kind === 'audio') {
    const operation = item?.config?.parameters?.operation || item?.config?.mode;
    if (allowedCategory('audio', operation)) return operation;
    if (operation === 'speech') return 'voice';
  }
  return DEFAULT_CATEGORY[kind];
}

export function filterStoryboardLibrary(library = [], state = {}, query = '') {
  const normalized = normalizeStoryboardState(state);
  const orderIndex = new Map(normalized.order.map((key, index) => [key, index]));
  const needle = String(query || '').trim().toLocaleLowerCase();
  return (Array.isArray(library) ? library : []).map(section => {
    const filter = normalized.filters[section.kind] || 'all';
    const sourceItems = Array.isArray(section.items) ? section.items : [];
    const naturalIndex = new Map(sourceItems.map((item, index) => [item.key, index]));
    const items = sourceItems
      .map(item => ({ ...item, category: storyboardItemCategory(item, normalized) }))
      .filter(item => filter === 'all' || item.category === filter)
      .filter(item => !needle || [item.title, item.summary, item.asset?.name].some(value => String(value || '').toLocaleLowerCase().includes(needle)))
      .sort((left, right) => {
        const leftOrder = orderIndex.has(left.key) ? orderIndex.get(left.key) : Number.MAX_SAFE_INTEGER;
        const rightOrder = orderIndex.has(right.key) ? orderIndex.get(right.key) : Number.MAX_SAFE_INTEGER;
        return leftOrder - rightOrder || naturalIndex.get(left.key) - naturalIndex.get(right.key);
      });
    return { ...section, items };
  });
}

export function moveStoryboardMaterial(state = {}, key = '', direction = 0, availableKeys = []) {
  const normalized = normalizeStoryboardState(state);
  const keys = [];
  const seen = new Set();
  for (const value of [...normalized.order, ...(Array.isArray(availableKeys) ? availableKeys : [])]) {
    if (typeof value !== 'string' || !value || seen.has(value)) continue;
    seen.add(value);
    keys.push(value);
  }
  const available = new Set((Array.isArray(availableKeys) ? availableKeys : []).filter(value => typeof value === 'string' && value));
  const visible = keys.filter(value => available.has(value));
  const index = visible.indexOf(key);
  const target = index + Math.sign(Number(direction) || 0);
  if (index < 0 || target < 0 || target >= visible.length) return normalized;
  [visible[index], visible[target]] = [visible[target], visible[index]];
  let visibleIndex = 0;
  const order = keys.map(value => available.has(value) ? visible[visibleIndex++] : value);
  return normalizeStoryboardState({ ...normalized, order });
}

export function reorderStoryboardMaterial(state = {}, sourceKey = '', targetKey = '', availableKeys = []) {
  const normalized = normalizeStoryboardState(state);
  const available = [];
  const availableSet = new Set();
  for (const key of Array.isArray(availableKeys) ? availableKeys : []) {
    if (typeof key !== 'string' || !key || availableSet.has(key)) continue;
    availableSet.add(key);
    available.push(key);
  }
  if (sourceKey === targetKey || !availableSet.has(sourceKey) || !availableSet.has(targetKey)) return normalized;

  const keys = [];
  const seen = new Set();
  for (const key of [...normalized.order, ...available]) {
    if (typeof key !== 'string' || !key || seen.has(key)) continue;
    seen.add(key);
    keys.push(key);
  }
  const visible = keys.filter(key => availableSet.has(key) && key !== sourceKey);
  const targetIndex = visible.indexOf(targetKey);
  if (targetIndex < 0) return normalized;
  visible.splice(targetIndex, 0, sourceKey);
  let visibleIndex = 0;
  const order = keys.map(key => availableSet.has(key) ? visible[visibleIndex++] : key);
  return normalizeStoryboardState({ ...normalized, order });
}

export function storyboardPointerDropTarget(sourceKey = '', sourceKind = '', candidates = []) {
  for (const candidate of Array.isArray(candidates) ? candidates : []) {
    const key = typeof candidate?.key === 'string' ? candidate.key : '';
    const kind = typeof candidate?.kind === 'string' ? candidate.kind : '';
    if (key && key !== sourceKey && kind === sourceKind) return key;
  }
  return '';
}
