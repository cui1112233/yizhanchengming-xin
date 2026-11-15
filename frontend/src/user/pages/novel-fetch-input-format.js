const columnPresets = {
  paid_id_title_gender_reason: ['bookId', 'bookName', 'gender', 'reason'],
  paid_name_reason: ['paidBookId', 'bookName', 'reason'],
  paid_name_gender_reason: ['paidBookId', 'bookName', 'gender', 'reason'],
  free_paid_name_gender_reason: ['freeBookId', 'paidBookId', 'bookName', 'gender', 'reason'],
  sample_input: ['bookId', 'bookName', 'reason', 'gender', 'tags', 'rating'],
  full_11: ['ignore', 'freeBookId', 'paidBookId', 'bookName', 'rating', 'ignore', 'ignore', 'ignore', 'ignore', 'gender', 'reason']
};

// 与后端 parse_mode / column_presets 一一对应，避免前后端可选格式不一致。
export const NOVEL_FETCH_INPUT_FORMATS = [
  { value: 'smart', label: '智能识别', description: '每行输入 ID 和书名即可', columnPreset: '智能识别', columnOrder: [] },
  { value: 'header', label: '表头识别', description: '自动识别表头列名，支持书籍ID/书名/男女频/理由', columnPreset: '表头识别', columnOrder: [] },
  { value: 'multi_header', label: '多表头识别', description: '兼容两行或多行表头的表格', columnPreset: '多表头识别', columnOrder: [] },
  { value: 'paid_id_title_gender_reason', label: '付费 ID、书名、男女频、理由', description: '按制表符分列', columnPreset: '付费 ID + 书名 + 男女频 + 理由', columnOrder: ['bookId', 'bookName', 'gender', 'reason'] },
  { value: 'fixed_full_11', label: '固定顺序：完整 11 列', description: '忽略/免费ID/付费ID/书名/评级/4个忽略列/男女频/理由', columnPreset: '完整11列', columnOrder: columnPresets.full_11 },
  { value: 'fixed_from_b', label: '固定顺序：从 B 列开始', description: '免费ID/付费ID/书名/评级/4个忽略列/男女频/理由', columnPreset: '完整11列（从B列）', columnOrder: columnPresets.full_11.slice(1) },
  { value: 'fixed_paid_basic', label: '固定顺序：付费ID/书名/男女频/理由', description: '四列固定顺序，无需表头', columnPreset: '付费ID/书名/男女频/推荐理由', columnOrder: ['paidBookId', 'bookName', 'gender', 'reason'] },
  { value: 'custom', label: '自定义列顺序', description: '选择每一列对应的信息', columnPreset: '自定义列顺序', columnOrder: ['bookId', 'bookName', 'gender', 'reason'] },
  { value: 'bookid_only', label: '书城 ID 抓取', description: '从任意文本或链接中提取书籍 ID', columnPreset: '书城ID抓取', columnOrder: [] }
];

const fixedColumnOrders = {
  fixed_full_11: columnPresets.full_11,
  fixed_from_b: columnPresets.full_11.slice(1),
  fixed_paid_basic: ['paidBookId', 'bookName', 'gender', 'reason']
};

const headerAliases = {
  bookId: ['id', '书籍id', '小说id', '作品id', 'bookid', '书号', '书籍编号', '小说编号', '作品编号', '付费id', '免费id', '付费书籍id', '免费书籍id', 'paidbookid', 'freebookid'],
  bookName: ['书名', '作品名称', '小说名称', '书籍名称', '名称', 'bookname', 'title'],
  gender: ['男女频', '性别', '频道', '频道类型', 'gender'],
  reason: ['推荐理由', '推荐原因', '入选理由', '理由', '简介', 'reason']
};

function splitInputLine(line) {
  if (line.includes('\t')) return line.split('\t');
  if (line.includes('|')) return line.split('|');
  if (/\s{2,}/.test(line)) return line.split(/\s{2,}/);
  if (line.includes(',')) return line.split(',');
  return [line];
}

const bookIdPattern = /\d{5,25}/;
const scopedBookIdPattern = /u_[A-Za-z0-9_-]+_\d{5,25}/;
const bookIdUrlPattern = /(?:bookid|book_id|bookId)=(\d{5,25})/i;

function extractBookId(text) {
  const value = String(text || '');
  const scoped = value.match(scopedBookIdPattern);
  if (scoped) return scoped[0];
  const fromUrl = value.match(bookIdUrlPattern);
  if (fromUrl) return fromUrl[1];
  const numeric = value.match(bookIdPattern);
  return numeric ? numeric[0] : '';
}

function rowFromFixedOrders(line, columns) {
  const values = splitInputLine(line).map(value => value.trim());
  const row = { bookId: '', bookName: '', gender: '', reason: '', error: '' };
  let paidBookId = '';
  let freeBookId = '';
  columns.forEach((column, index) => {
    if (!column || column === 'ignore') return;
    const value = values[index] || '';
    if (column === 'reason' || column === 'tags') {
      row.reason = row.reason ? `${row.reason} ${value}` : value;
      return;
    }
    if (column === 'paidBookId') {
      paidBookId = extractBookId(value) || value;
      return;
    }
    if (column === 'freeBookId') {
      freeBookId = extractBookId(value) || value;
      return;
    }
    row[column] = value;
  });
  // 同时存在免费/付费 ID 时与后端一致，优先使用付费 ID。
  row.bookId = paidBookId || freeBookId || row.bookId;
  return row;
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

function detectHeader(lines, maxScan = lines.length) {
  const limit = Math.min(lines.length, Math.max(1, maxScan));
  for (let rowIndex = 0; rowIndex < limit; rowIndex += 1) {
    const values = splitInputLine(lines[rowIndex]).map(value => value.trim());
    const mapping = {};
    values.forEach((value, index) => {
      const key = headerKey(value);
      if (key && mapping[key] === undefined) mapping[key] = index;
    });
    if (mapping.bookId !== undefined && Object.keys(mapping).length >= 2) {
      let rowsToSkip = rowIndex + 1;
      // 多行表头：继续吞掉后续仍能识别出字段名的行，并合并列映射。
      while (rowsToSkip < limit) {
        const nextValues = splitInputLine(lines[rowsToSkip]).map(value => value.trim());
        const nextMapping = {};
        nextValues.forEach((value, index) => {
          const key = headerKey(value);
          if (key && nextMapping[key] === undefined) nextMapping[key] = index;
        });
        if (Object.keys(nextMapping).length < 2) break;
        Object.entries(nextMapping).forEach(([key, index]) => {
          if (mapping[key] === undefined) mapping[key] = index;
        });
        rowsToSkip += 1;
      }
      return { mapping, rowsToSkip };
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
  const headerFormat = format === 'smart' || format === 'header' || format === 'multi_header';
  const detectedHeader = headerFormat
    ? detectHeader(lines, format === 'multi_header' ? 8 : lines.length)
    : null;
  const explicitMapping = columnMapping && typeof columnMapping === 'object'
    ? columnMapping
    : detectedHeader?.mapping;
  const fixedColumns = fixedColumnOrders[format];
  const dataLines = detectedHeader ? lines.slice(detectedHeader.rowsToSkip) : lines;
  const rows = dataLines.map(line => {
      if (format === 'bookid_only') {
        const bookId = extractBookId(line);
        return { bookId, bookName: '', gender: '', reason: '', error: bookId ? '' : '缺少书籍 ID' };
      }
      if (fixedColumns) return rowFromFixedOrders(line, fixedColumns);
      if (explicitMapping) return rowFromMapping(line, explicitMapping);
      if (format === 'smart') return normalizeSmartRow(line);
      const values = splitInputLine(line).map(value => value.trim());
      const fixedRow = { bookId: '', bookName: '', gender: '', reason: '', error: '' };
      const mapping = columnMapping || Object.fromEntries(columns.map((column, index) => [column, index]));
      Object.entries(mapping).forEach(([column, index]) => {
        fixedRow[column] = values[index] || '';
      });
      return fixedRow;
    });

  rows.forEach(row => {
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
  });

  return { rows, validRows, duplicateIds, emptyIdRows };
}

export function serializeNovelFetchRows(rows = []) {
  return rows
    .filter(row => row && !row.error && String(row.bookId || '').trim())
    .map(row => ['bookId', 'bookName', 'gender', 'reason'].map(key => String(row[key] || '').replace(/[\t\r\n]+/g, ' ').trim()).join('\t'))
    .join('\n');
}
