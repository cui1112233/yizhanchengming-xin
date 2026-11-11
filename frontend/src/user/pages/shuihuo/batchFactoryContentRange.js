export function contentRangeLinesForBook(book) {
  const value = Number(book?.sourceMetadata?.contentRangeLines);
  return Number.isInteger(value) && value > 0 ? Math.min(value, 500) : 5;
}

export function contentCaptureCharactersForBook(book) {
  const value = Number(book?.sourceMetadata?.contentCaptureCharacters);
  return Number.isInteger(value) && value >= 100 && value <= 100000 ? value : 4000;
}

export function inheritedContentSettingsForNewBook(books, fallbackRangeLines = 5, fallbackCaptureCharacters = 4000) {
  const existing = Array.isArray(books) ? books.find(book => book && typeof book === 'object') : null;
  const fallbackBook = {
    sourceMetadata: {
      contentRangeLines: fallbackRangeLines,
      contentCaptureCharacters: fallbackCaptureCharacters
    }
  };
  const source = existing || fallbackBook;
  return {
    contentRangeLines: contentRangeLinesForBook(source),
    contentCaptureCharacters: contentCaptureCharactersForBook(source)
  };
}

function comicAnalysisTargetCharactersForBook(book) {
  const value = Number(book?.sourceMetadata?.comicAnalysisCharacters);
  return Number.isInteger(value) && value >= 100 && value <= 100000 ? value : 2000;
}

export function comicEstimatedPageRange(characterCount) {
  const characters = Math.max(1, Number(characterCount) || 1);
  return {
    min: Math.min(35, Math.ceil(characters / 120)),
    max: Math.min(35, Math.ceil(characters / 90))
  };
}

function normalizedSourceLine(value) {
  return String(value || '')
    .replace(/&nbsp;|\u00a0|　/g, ' ')
    .replace(/<[^>]*>/g, '')
    .trim();
}

function isLeadingPageStateLine(value) {
  return /^(?:修改中|加载中|正文加载中|请稍候)$/.test(value);
}

function isPunctuationOnlyLine(value) {
  return /^[，。！？、；：…·—～~,.!?;:()（）【】\[\]{}「」『』“”"'\-]+$/.test(value);
}

export function comicAnalysisCharactersForBook(book) {
  const source = Array.from(batchFactoryCleanSourceText(book?.sourceText || ''));
  const target = comicAnalysisTargetCharactersForBook(book);
  if (source.length <= target) return source.length;
  const isBoundary = value => '。！？!?；;\n'.includes(value);
  let next = -1;
  for (let index = target; index < source.length; index += 1) {
    if (isBoundary(source[index])) { next = index; break; }
  }
  let previous = -1;
  for (let index = target - 1; index >= 0; index -= 1) {
    if (isBoundary(source[index])) { previous = index; break; }
  }
  if (previous >= 0 && (next < 0 || target - previous <= next - target)) return previous + 1;
  return next >= 0 ? next + 1 : target;
}

export function batchFactorySourceText(value) {
  if (typeof value !== 'string') return '';
  const text = value.trim();
  return text === '[object Object]' ? '' : text;
}

function usableSourceLines(sourceText) {
  const lines = batchFactorySourceText(sourceText).split(/\r?\n/).map(normalizedSourceLine).filter(Boolean);
  let firstContent = 0;
  while (firstContent < lines.length && isLeadingPageStateLine(lines[firstContent])) firstContent += 1;
  return lines.slice(firstContent).filter(line => !isPunctuationOnlyLine(line));
}

// Detail dialogs, previews and production must use the same cleaned text.
// The raw upstream response is retained separately in source metadata.
export function batchFactoryCleanSourceText(sourceText) {
  return usableSourceLines(sourceText).join('\n');
}

export function batchFactoryPreviewText(sourceText, book) {
  return usableSourceLines(sourceText)
    .slice(0, contentRangeLinesForBook(book))
    .join('\n');
}

export function batchFactoryProductionText(sourceText, workingFrontContent, book) {
  const working = batchFactorySourceText(workingFrontContent);
  return working || batchFactoryPreviewText(sourceText, book);
}

// A user-approved viral rewrite becomes the book's new complete source. The
// workbench still decides what to show from the saved source through
// batchFactoryPreviewText and the book's selected contentRangeLines.
export function sourceTextAfterViralAdoption(viralCandidate) {
  return batchFactorySourceText(viralCandidate);
}


export function publishContentWithWorkingFront(sourceText, workingFrontContent, book) {
  const normalizedSource = batchFactorySourceText(sourceText);
  const sourceLines = normalizedSource.split(/\r?\n/);
  const editedLines = batchFactorySourceText(workingFrontContent).split(/\r?\n/).map(line => line.trim()).filter(Boolean);
  if (!editedLines.length) return normalizedSource;
  const limit = contentRangeLinesForBook(book);
  let replacement = 0;
  let seen = 0;
  return sourceLines.map(line => {
    if (!line.trim() || seen >= limit) return line;
    seen += 1;
    if (replacement >= editedLines.length) return line;
    const next = editedLines[replacement];
    replacement += 1;
    return next;
  }).join('\n');
}
