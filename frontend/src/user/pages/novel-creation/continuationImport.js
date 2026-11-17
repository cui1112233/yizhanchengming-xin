// Independent implementation of the recovered import form's local chapter contract.
// Count JavaScript UTF16 units, excluding whitespace; the server recounts on save.
export function splitContinuationChapters(value) {
  const text = String(value || '').trim();
  if (!text) return [];
  const starts = [...text.matchAll(/第[0-9一二三四五六七八九十百千]+章/g)].map(match => match.index);
  const boundaries = [...new Set([0, ...starts, text.length])];
  const chapters = [];
  for (let i = 0; i < boundaries.length - 1; i++) {
    const segment = text.slice(boundaries[i], boundaries[i + 1]).trim();
    if (!segment) continue;
    const heading = segment.match(/^第[0-9一二三四五六七八九十百千]+章[^\n]*/)?.[0];
    const content = (heading ? segment.slice(heading.length) : segment).trim();
    if (content) chapters.push({ title: heading?.trim() || `第${chapters.length + 1}章`, content, wordCount: content.replace(/\s/g, '').length });
  }
  return chapters.length ? chapters : [{ title: '第1章', content: text, wordCount: text.replace(/\s/g, '').length }];
}
