const columnPresets = {
  paid_id_title_gender_reason: ['bookId', 'bookName', 'gender', 'reason']
};

const headerAliases = {
  bookId: ['id', '书籍id', '小说id', '作品id', 'bookid', '书号', '书籍编号', '小说编号', '作品编号'],
  bookName: ['书名', '作品名称', '小说名称', '书籍名称', '名称', 'bookname', 'title'],
  gender: ['男女频', '性别', '频道', '频道类型', 'gender'],
  reason: ['推荐理由', '推荐原因', '入选理由', '理由', '简介', 'reason']
};

function splitInputLine(line) {
  if (line.includes('\t')) return line.split('\t');
  if (line.includes('|')) return line.split('|');
  if (line.includes(',')) return line.split(',');
  return [line];
}

function normalizeSmartRow(line) {
  const trimmed = line.trim();
  const idMatch = trimmed.match(/^(\d{8,})(?:\s+|[|,])?(.*)$/);
  const row = { bookId: '', bookName: '', gender: '', reason: '', error: '' };
  if (!idMatch) return row;

  row.bookId = idMatch[1];
  const remainder = idMatch[2].trim();
  if (!remainder) return row;
  if (!/[\t|]/.test(remainder)) {
    row.bookName = remainder;
    return row;
  }

  const values = splitInputLine(remainder).map(value => value.trim()).filter(Boolean);
  row.bookName = values.shift() || '';
  const genderIndex = values.findIndex(value => value.includes('男频') || value.includes('女频'));
  if (genderIndex >= 0) {
    row.gender = values[genderIndex];
    values.splice(genderIndex, 1);
  }
  row.reason = values.join(' ');
  return row;
}

function headerKey(value) {
  const normalized = String(value || '').replace(/[\s*_：:（）()【】\[\]·-]/g, '').toLowerCase();
  return Object.entries(headerAliases).find(([, aliases]) => aliases.includes(normalized))?.[0] || '';
}

function detectHeader(lines) {
  for (let rowIndex = 0; rowIndex < lines.length; rowIndex += 1) {
    const values = splitInputLine(lines[rowIndex]).map(value => value.trim());
    const mapping = {};
    values.forEach((value, index) => {
      const key = headerKey(value);
      if (key && mapping[key] === undefined) mapping[key] = index;
    });
    if (mapping.bookId !== undefined && Object.keys(mapping).length >= 2) {
      return { mapping, rowsToSkip: rowIndex + 1 };
    }
  }
  return null;
}

function rowFromMapping(line, mapping) {
  const values = splitInputLine(line).map(value => value.trim());
  return {
    ...Object.fromEntries(['bookId', 'bookName', 'gender', 'reason'].map(key => [key, values[mapping[key]] || ''])),
    error: ''
  };
}

export function normalizeNovelFetchInput(input, { format = 'smart', columnMapping } = {}) {
  const columns = columnPresets[format] || columnPresets.paid_id_title_gender_reason;
  const seenIDs = new Set();
  let validRows = 0;
  let duplicateIds = 0;
  let emptyIdRows = 0;

  const lines = String(input || '')
    .replace(/\r\n?/g, '\n')
    .split('\n')
    .filter(line => line.trim() !== '');
  const detectedHeader = format === 'smart' ? detectHeader(lines) : null;
  const explicitMapping = columnMapping && typeof columnMapping === 'object'
    ? columnMapping
    : detectedHeader?.mapping;
  const dataLines = detectedHeader ? lines.slice(detectedHeader.rowsToSkip) : lines;
  const rows = dataLines.map(line => {
      const row = explicitMapping
        ? rowFromMapping(line, explicitMapping)
        : format === 'smart'
        ? normalizeSmartRow(line)
        : (() => {
            const values = splitInputLine(line).map(value => value.trim());
            const fixedRow = { bookId: '', bookName: '', gender: '', reason: '', error: '' };
            const mapping = columnMapping || Object.fromEntries(columns.map((column, index) => [column, index]));
            Object.entries(mapping).forEach(([column, index]) => {
              fixedRow[column] = values[index] || '';
            });
            return fixedRow;
          })();

      if (!row.bookId) {
        row.error = '缺少书籍 ID';
        emptyIdRows += 1;
      } else if (seenIDs.has(row.bookId)) {
        row.error = '重复书籍 ID';
        duplicateIds += 1;
      } else {
        seenIDs.add(row.bookId);
        validRows += 1;
      }
      return row;
    });

  return { rows, validRows, duplicateIds, emptyIdRows };
}

export function serializeNovelFetchRows(rows = []) {
  return rows
    .filter(row => row && !row.error && String(row.bookId || '').trim())
    .map(row => ['bookId', 'bookName', 'gender', 'reason'].map(key => String(row[key] || '').replace(/[\t\r\n]+/g, ' ').trim()).join('\t'))
    .join('\n');
}
