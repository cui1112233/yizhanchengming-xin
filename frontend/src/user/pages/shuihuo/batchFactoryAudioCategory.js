const text = value => String(value || '').trim();

function mappedCategoryName(metadata = {}) {
  if (text(metadata.classifyStatus).toLowerCase() !== 'classified') return '';
  const gender = text(metadata.gender || metadata.gender_name);
  const style = text(metadata.style || metadata.styleName || metadata.style_name);
  const isFemale = gender === '女频' || gender === '女';
  const isMale = gender === '男频' || gender === '男';
  const isRevenge = /爽文|打脸/.test(style);
  const isSad = /悲伤/.test(style);
  if (isFemale && isRevenge) return '女-爽文（打脸）';
  if (isFemale && isSad) return '古风-女悲伤';
  if (isMale && isSad) return '古风-男悲伤';
  if (isMale && isRevenge) return '古风-爽文（打脸）';
  return '';
}

export function inferBatchN8AudioCategory(books = [], n8Categories = []) {
  const source = Array.isArray(books) ? books : [];
  if (!source.length) return { category: null, reason: 'empty' };
  const names = source.map(book => mappedCategoryName(book?.sourceMetadata || {}));
  if (source.some(book => text(book?.sourceMetadata?.classifyStatus).toLowerCase() !== 'classified')) return { category: null, reason: 'unclassified' };
  if (names.some(name => !name)) return { category: null, reason: 'unsupported' };
  if (new Set(names).size !== 1) return { category: null, reason: 'mixed' };
  const matches = (Array.isArray(n8Categories) ? n8Categories : []).filter(category => text(category?.status).toUpperCase() === 'ENABLE' && text(category?.name) === names[0]);
  return matches.length === 1 ? { category: matches[0], reason: 'matched' } : { category: null, reason: 'unavailable' };
}
