export function normalizedCharacters(raw) {
  const characters = [];
  const ends = [];
  let offset = 0;
  for (const character of raw) {
    offset += character.length;
    if (/[\p{L}\p{N}]/u.test(character)) { characters.push(character); ends.push(offset); }
  }
  return { characters, ends };
}

export function overlapBoundary(previous, next) {
  const a = normalizedCharacters(previous).characters.slice(-600);
  const { characters: b, ends } = normalizedCharacters(next);
  const width = Math.min(b.length, 600);
  let row = Array.from({ length: width + 1 }, (_, j) => j);
  for (let i = 1; i <= a.length; i += 1) {
    const current = [0];
    for (let j = 1; j <= width; j += 1) {
      current[j] = Math.min(row[j] + 1, current[j - 1] + 1, row[j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1));
    }
    row = current;
  }
  let best = 0;
  let score = 19;
  for (let j = 20; j <= width; j += 1) {
    const candidate = j - 5 * row[j];
    if (row[j] / j <= 0.12 && candidate > score) { score = candidate; best = j; }
  }
  return best ? { offset: ends[best - 1], matched: best, edits: row[best] } : null;
}

export function appendFrame(body, previous, next) {
  const current = next.trim();
  if (!current) return { body, boundary: null, skipped: true };
  if (!body) return { body: current, boundary: null, skipped: false };
  const normalized = normalizedCharacters(current).characters.join('');
  const tail = normalizedCharacters(body.slice(-2500)).characters.join('');
  if (normalized && tail.includes(normalized)) return { body, boundary: { matched: normalized.length, edits: 0 }, skipped: true };
  const boundary = overlapBoundary(previous, current);
  if (!boundary) return { body: `${body}\n\n${current}`, boundary: null, skipped: false };
  const extra = current.slice(boundary.offset);
  if (!/[\p{L}\p{N}]/u.test(extra)) return { body, boundary, skipped: true };
  const bodyEnds = normalizedCharacters(body).ends;
  return { body: body.slice(0, bodyEnds[bodyEnds.length - 1]) + extra, boundary, skipped: false };
}
