const text = value => String(value || '').trim();

export function normalizeBatchAutoAudioConfig(config = {}) {
  const next = JSON.parse(JSON.stringify(config && typeof config === 'object' ? config : {}));
  const publish = next.comicPublishSettings && typeof next.comicPublishSettings === 'object'
    ? next.comicPublishSettings
    : {};
  next.comicPublishSettings = {
    ...publish,
    audioCategorySelectionMode: 'auto',
    audioCategoryId: null,
    audioCategoryName: '',
    audioCandidates: [],
    specifiedAudioId: '',
    specifiedAudioName: ''
  };
  return next;
}

export function audioMatchingModeOptions(singleBook = false) {
  const options = [{ value: 'auto', label: '按风格自动匹配' }];
  return singleBook ? [...options, { value: 'manual', label: '手动指定' }] : options;
}

export function mappedAudioCategoryName(metadata = {}) {
  if (text(metadata.classifyStatus || metadata.classify_status).toLowerCase() !== 'classified') return '';
  const gender = text(metadata.gender || metadata.gender_name);
  const style = text(metadata.style || metadata.styleName || metadata.style_name);
  const explicit = {
    现代悬疑: '现代悬疑', 男频都市: '男频都市',
    古风通用: '古风通用', 玄幻: '古风通用', 历史: '古风通用',
    古风虐文: '古风虐文', 古风甜文: '古风甜文',
    现代虐文: '现代虐文', 年代虐文: '现代虐文', 家庭伤感: '现代虐文',
    现代甜文: '现代通用', 现代通用: '现代通用', 年代甜文: '现代通用',
    年代通用: '现代通用', 现代女主: '现代通用', 家庭奇葩: '现代通用'
  }[style];
  if (explicit) return explicit;
  if (gender === '男频' || gender === '男') return '男频都市';
  if (gender === '女频' || gender === '女') return '现代通用';
  return '';
}

export function inferBookN8AudioCategory(book = {}, n8Categories = []) {
  const metadata = book?.sourceMetadata || {};
  if (text(metadata.classifyStatus || metadata.classify_status).toLowerCase() !== 'classified') return { category: null, categoryName: '', reason: 'unclassified' };
  const categoryName = mappedAudioCategoryName(metadata);
  if (!categoryName) return { category: null, categoryName: '', reason: 'unclassified' };
  const matches = (Array.isArray(n8Categories) ? n8Categories : []).filter(category => text(category?.status).toUpperCase() === 'ENABLE' && text(category?.name) === categoryName);
  return matches.length === 1
    ? { category: matches[0], categoryName, reason: 'matched' }
    : { category: null, categoryName, reason: matches.length ? 'ambiguous' : 'unavailable' };
}

export function inferBatchN8AudioCategory(books = [], n8Categories = []) {
  const source = Array.isArray(books) ? books : [];
  if (!source.length) return { category: null, reason: 'empty' };
  const inferred = source.map(book => inferBookN8AudioCategory(book, n8Categories));
  if (inferred.some(item => item.reason === 'unclassified')) return { category: null, reason: 'unclassified' };
  if (new Set(inferred.map(item => item.categoryName)).size !== 1) return { category: null, reason: 'mixed' };
  const first = inferred[0];
  return first.reason === 'matched' ? { category: first.category, reason: 'matched' } : { category: null, reason: first.reason };
}
