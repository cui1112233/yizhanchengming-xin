// 洛水三条纯本地工具的实际处理逻辑：格式化文本、字幕合并、文档合并。
// 全部在浏览器内完成，不调用任何接口、不消耗模型额度。

/* ------------------------------------------------------------------ */
/* 格式化文本                                                          */
/* ------------------------------------------------------------------ */

const CJK = /[\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff]/;
const HALF_TO_FULL = {
  ',': '，',
  '.': '。',
  '?': '？',
  '!': '！',
  ':': '：',
  ';': '；',
  '(': '（',
  ')': '）'
};

function isCjkChar(ch) {
  return Boolean(ch) && (CJK.test(ch) || /[\u3000-\u303f\uff00-\uffef]/.test(ch));
}

// 只改“紧挨中文”的半角标点：英文句子、URL、小数（3.5）、千分位（1,000）都不动。
function normalizeCjkPunctuation(line) {
  if (!CJK.test(line)) return line;
  let source = line;
  // 成对的半角引号先按配对替换，避免逐字符替换后无法判断开合。
  source = source.replace(/"([^"]*)"/g, '“$1”');
  source = source.replace(/\.{3,}/g, '……');
  let out = '';
  for (let index = 0; index < source.length; index += 1) {
    const ch = source[index];
    const prev = source[index - 1] || '';
    const next = source[index + 1] || '';
    const mapped = HALF_TO_FULL[ch];
    const betweenDigits = /[0-9]/.test(prev) && /[0-9]/.test(next);
    if (mapped && !betweenDigits && (isCjkChar(prev) || isCjkChar(next))) {
      out += mapped;
    } else {
      out += ch;
    }
  }
  // 连续重复的同类标点收成一个（省略号、破折号不在表内，不受影响）。
  return out.replace(/([，。！？：；])\1+/g, '$1');
}

function removeCjkSpaces(line) {
  return line.replace(/(?<=[\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff])[ \t]+(?=[\u3400-\u4dbf\u4e00-\u9fff\uf900-\ufaff])/g, '');
}

const TEXT_FORMAT_DEFAULTS = {
  trimLines: true,
  collapseBlank: true,
  removeBlank: false,
  normalizePunctuation: true,
  removeCjkSpaces: true,
  indentParagraphs: false
};

export function formatText(input, options = {}) {
  const config = { ...TEXT_FORMAT_DEFAULTS, ...options };
  const source = String(input || '').replace(/\r\n?/g, '\n');
  const rawLines = source.split('\n');
  const lines = [];
  let changedLines = 0;

  for (const raw of rawLines) {
    let line = config.trimLines ? raw.trim() : raw;
    if (config.normalizePunctuation) line = normalizeCjkPunctuation(line);
    if (config.removeCjkSpaces) line = removeCjkSpaces(line);
    if (line !== raw) changedLines += 1;
    lines.push(line);
  }

  let output = lines;
  if (config.removeBlank) {
    output = output.filter(line => line !== '');
  } else if (config.collapseBlank) {
    const collapsed = [];
    let blank = 0;
    for (const line of output) {
      if (line === '') {
        blank += 1;
        if (blank > 1) continue;
      } else {
        blank = 0;
      }
      collapsed.push(line);
    }
    output = collapsed;
  }

  if (config.indentParagraphs) {
    output = output.map(line => (line ? `　　${line}` : line));
  }

  const text = output.join('\n').replace(/^\n+/, '').replace(/\n+$/, '');
  return {
    text,
    stats: {
      charsBefore: source.length,
      charsAfter: text.length,
      linesBefore: rawLines.length,
      linesAfter: text.split('\n').filter(line => line !== '').length,
      changedLines
    }
  };
}

/* ------------------------------------------------------------------ */
/* 字幕解析 / 合并                                                     */
/* ------------------------------------------------------------------ */

const TIME_PATTERN = /(\d{1,2}):(\d{2}):(\d{2})[,.](\d{1,3})/;

function toMs(hour, minute, second, fraction) {
  const ms = Number(`${fraction}00`.slice(0, 3));
  return ((Number(hour) * 60 + Number(minute)) * 60 + Number(second)) * 1000 + ms;
}

export function parseSubtitleTime(value) {
  const matched = String(value || '').match(TIME_PATTERN);
  if (!matched) return null;
  return toMs(matched[1], matched[2], matched[3], matched[4]);
}

export function formatSrtTime(totalMs) {
  const safe = Math.max(0, Math.round(Number(totalMs) || 0));
  const hours = Math.floor(safe / 3600000);
  const minutes = Math.floor((safe % 3600000) / 60000);
  const seconds = Math.floor((safe % 60000) / 1000);
  const millis = safe % 1000;
  const pad = (value, size) => String(value).padStart(size, '0');
  return `${pad(hours, 2)}:${pad(minutes, 2)}:${pad(seconds, 2)},${pad(millis, 3)}`;
}

function stripAssTags(text) {
  return text
    .replace(/\{[^}]*\}/g, '')
    .replace(/\\N/gi, '\n')
    .replace(/\\n/gi, '\n')
    .replace(/\\h/gi, ' ');
}

function stripVttTags(text) {
  return text.replace(/<[^>]+>/g, '').replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>');
}

function normalizeCueText(text) {
  return String(text || '')
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)
    .join('\n');
}

function parseSrtOrVtt(content, formatHint) {
  const source = String(content || '').replace(/\r\n?/g, '\n').replace(/^\uFEFF/, '');
  const isVtt = formatHint === 'vtt' || /^WEBVTT/i.test(source);
  const lines = source.split('\n');
  const cues = [];
  let current = null;

  const flush = () => {
    if (!current) return;
    const body = current.lines.join('\n');
    const text = normalizeCueText(isVtt ? stripVttTags(body) : body);
    if (text && current.start !== null && current.end !== null) {
      cues.push({ start: current.start, end: current.end, text });
    }
    current = null;
  };

  // 下一非空行是不是时间行：用于识别“没有空行分隔”的 SRT 序号行。
  const nextNonEmptyIsArrow = index => {
    for (let cursor = index + 1; cursor < lines.length; cursor += 1) {
      const trimmed = lines[cursor].trim();
      if (!trimmed) continue;
      return trimmed.includes('-->') && TIME_PATTERN.test(trimmed);
    }
    return false;
  };

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index].trim();
    if (!line) {
      flush();
      continue;
    }
    if (line.includes('-->') && TIME_PATTERN.test(line)) {
      flush();
      const [left, right] = line.split('-->');
      current = { start: parseSubtitleTime(left), end: parseSubtitleTime(right), lines: [] };
      continue;
    }
    if (!current) continue;
    if (/^\d+$/.test(line) && nextNonEmptyIsArrow(index)) {
      flush();
      continue;
    }
    current.lines.push(line);
  }
  flush();

  return { format: isVtt ? 'vtt' : 'srt', cues };
}

function parseAss(content) {
  const source = String(content || '').replace(/\r\n?/g, '\n').replace(/^\uFEFF/, '');
  const lines = source.split('\n');
  let inEvents = false;
  let fields = null;
  const cues = [];
  for (const raw of lines) {
    const line = raw.trim();
    if (/^\[.*\]$/.test(line)) {
      inEvents = /^\[events\]$/i.test(line);
      fields = null;
      continue;
    }
    if (!inEvents) continue;
    if (/^format\s*:/i.test(line)) {
      fields = line
        .slice(line.indexOf(':') + 1)
        .split(',')
        .map(item => item.trim().toLowerCase());
      continue;
    }
    if (!/^dialogue\s*:/i.test(line)) continue;
    const body = line.slice(line.indexOf(':') + 1);
    const layout = fields || ['layer', 'start', 'end', 'style', 'name', 'marginl', 'marginr', 'marginv', 'effect', 'text'];
    const parts = [];
    let rest = body;
    for (let index = 0; index < layout.length - 1; index += 1) {
      const comma = rest.indexOf(',');
      if (comma === -1) break;
      parts.push(rest.slice(0, comma));
      rest = rest.slice(comma + 1);
    }
    parts.push(rest);
    const startIndex = layout.indexOf('start');
    const endIndex = layout.indexOf('end');
    const textIndex = layout.indexOf('text');
    if (startIndex === -1 || endIndex === -1 || textIndex === -1) continue;
    const start = parseSubtitleTime(parts[startIndex]);
    const end = parseSubtitleTime(parts[endIndex]);
    if (start === null || end === null) continue;
    const text = normalizeCueText(stripAssTags(parts[textIndex] || ''));
    if (!text) continue;
    cues.push({ start, end, text });
  }
  return { format: 'ass', cues };
}

export function detectSubtitleFormat(fileName, content) {
  const name = String(fileName || '').toLowerCase();
  if (name.endsWith('.ass') || name.endsWith('.ssa')) return 'ass';
  if (name.endsWith('.vtt')) return 'vtt';
  if (name.endsWith('.srt')) return 'srt';
  if (/\[events\]/i.test(content) && /^dialogue\s*:/im.test(content)) return 'ass';
  if (/^WEBVTT/i.test(String(content || '').trimStart())) return 'vtt';
  return 'srt';
}

export function parseSubtitle(content, { fileName = '', format } = {}) {
  const resolved = format || detectSubtitleFormat(fileName, content);
  if (resolved === 'ass') return parseAss(content);
  return parseSrtOrVtt(content, resolved);
}

function overlapRatio(a, b) {
  const start = Math.max(a.start, b.start);
  const end = Math.min(a.end, b.end);
  const overlap = Math.max(0, end - start);
  const shortest = Math.min(a.end - a.start, b.end - b.start) || 1;
  return overlap / shortest;
}

function cueKey(text) {
  return String(text || '').replace(/\s+/g, '').toLowerCase();
}

/**
 * files: [{ name, content, offsetMs?, order? }]
 * mode: 'sort'   多份文件落在同一条时间轴上，合并后按开始时间重排
 *       'concat' 按文件顺序首尾相接（分集字幕合并），gapMs 为集与集之间的间隔
 */
export function mergeSubtitles(files, { mode = 'sort', gapMs = 0, dedupe = true } = {}) {
  const list = Array.isArray(files) ? files.filter(item => item && String(item.content || '').trim()) : [];
  const parsedFiles = list.map((file, index) => {
    const parsed = parseSubtitle(file.content, { fileName: file.name });
    const offset = Number(file.offsetMs) || 0;
    return {
      name: file.name || `文件${index + 1}`,
      format: parsed.format,
      offsetMs: offset,
      cues: parsed.cues.map(cue => ({ ...cue, start: cue.start + offset, end: cue.end + offset }))
    };
  });

  const cues = [];
  if (mode === 'concat') {
    let cursor = 0;
    for (const file of parsedFiles) {
      if (!file.cues.length) continue;
      const first = Math.min(...file.cues.map(cue => cue.start));
      const last = Math.max(...file.cues.map(cue => cue.end));
      const delta = cursor === 0 ? 0 : cursor + Math.max(0, Number(gapMs) || 0) - first;
      for (const cue of file.cues) {
        cues.push({ start: cue.start + delta, end: cue.end + delta, text: cue.text, source: file.name });
      }
      cursor = last + delta;
    }
  } else {
    for (const file of parsedFiles) {
      for (const cue of file.cues) cues.push({ ...cue, source: file.name });
    }
  }

  let dropped = 0;
  const valid = [];
  for (const cue of cues) {
    if (!(cue.end > cue.start)) {
      dropped += 1;
      continue;
    }
    valid.push(cue);
  }

  valid.sort((a, b) => a.start - b.start || a.end - b.end || a.text.localeCompare(b.text, 'zh-CN'));

  const merged = [];
  let duplicates = 0;
  for (const cue of valid) {
    if (dedupe) {
      const duplicated = merged.some(existing =>
        existing.key === cueKey(cue.text) && overlapRatio(existing, cue) >= 0.8);
      if (duplicated) {
        duplicates += 1;
        continue;
      }
    }
    merged.push({ ...cue, key: cueKey(cue.text) });
  }

  const text = buildSrt(merged);
  const durationMs = merged.reduce((max, cue) => Math.max(max, cue.end), 0);
  return {
    text,
    cues: merged.map((cue, index) => ({ index: index + 1, ...cue })),
    stats: {
      files: parsedFiles.length,
      cues: merged.length,
      dropped,
      duplicates,
      durationMs,
      perFile: parsedFiles.map(file => ({ name: file.name, format: file.format, cues: file.cues.length }))
    }
  };
}

export function buildSrt(cues) {
  if (!cues.length) return '';
  const blocks = cues.map((cue, index) => [
    String(index + 1),
    `${formatSrtTime(cue.start)} --> ${formatSrtTime(cue.end)}`,
    cue.text
  ].join('\n'));
  return `${blocks.join('\n\n')}\n`;
}

/* ------------------------------------------------------------------ */
/* 文档合并                                                            */
/* ------------------------------------------------------------------ */

const CN_DIGITS = { 零: 0, 〇: 0, 一: 1, 二: 2, 两: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9 };
const CN_UNITS = { 十: 10, 百: 100, 千: 1000, 万: 10000 };

export function parseChineseNumber(value) {
  const source = String(value || '').trim();
  if (!source) return null;
  if (/^\d+$/.test(source)) return Number(source);
  let total = 0;
  let section = 0;
  let number = 0;
  for (const ch of source) {
    if (ch in CN_DIGITS) {
      number = CN_DIGITS[ch];
      continue;
    }
    if (ch in CN_UNITS) {
      const unit = CN_UNITS[ch];
      if (unit === 10000) {
        section = (section + number) * unit;
        total += section;
        section = 0;
      } else {
        section += (number || 1) * unit;
      }
      number = 0;
      continue;
    }
    return null;
  }
  return total + section + number;
}

const CHAPTER_PATTERNS = [
  /第\s*(\d{1,5})\s*[章节回卷集篇]/,
  /第\s*([零〇一二两三四五六七八九十百千万]{1,8})\s*[章节回卷集篇]/,
  /chapter\s*(\d{1,5})/i,
  /(?:^|[^\d])(\d{1,5})(?=[^\d]|$)/
];

export function extractChapterNumber(text) {
  const source = String(text || '');
  for (const pattern of CHAPTER_PATTERNS) {
    const matched = source.match(pattern);
    if (!matched) continue;
    const value = /^\d+$/.test(matched[1]) ? Number(matched[1]) : parseChineseNumber(matched[1]);
    if (Number.isFinite(value) && value > 0) return value;
  }
  return null;
}

function naturalCompare(a, b) {
  return String(a).localeCompare(String(b), 'zh-CN', { numeric: true, sensitivity: 'base' });
}

function stripExtension(name) {
  return String(name || '').replace(/\.(txt|md|markdown|text)$/i, '');
}

const DOCUMENT_SEPARATORS = {
  title: name => `=== ${name} ===`,
  blank: () => '',
  none: null
};

export function mergeDocuments(files, {
  order = 'natural',
  separator = 'title',
  collapseBlank = true,
  trimLines = true,
  dedupeParagraphs = false
} = {}) {
  const list = (Array.isArray(files) ? files : [])
    .filter(file => file && String(file.content || '').trim())
    .map((file, index) => {
      const name = file.name || `文档${index + 1}`;
      const title = stripExtension(name);
      const chapter = extractChapterNumber(title) ?? extractChapterNumber(String(file.content || '').slice(0, 200));
      return { name, title, chapter, content: file.content, index };
    });

  if (order === 'chapter') {
    list.sort((a, b) => {
      if (a.chapter === null && b.chapter === null) return naturalCompare(a.name, b.name);
      if (a.chapter === null) return 1;
      if (b.chapter === null) return -1;
      return a.chapter - b.chapter || naturalCompare(a.name, b.name);
    });
  } else if (order === 'natural') {
    list.sort((a, b) => naturalCompare(a.name, b.name));
  }

  const makeSeparator = Object.prototype.hasOwnProperty.call(DOCUMENT_SEPARATORS, separator)
    ? DOCUMENT_SEPARATORS[separator]
    : DOCUMENT_SEPARATORS.title;
  const blocks = [];
  for (const file of list) {
    let body = String(file.content || '').replace(/\r\n?/g, '\n').replace(/^\uFEFF/, '');
    let lines = body.split('\n');
    if (trimLines) lines = lines.map(line => line.trimEnd());
    if (collapseBlank) {
      const collapsed = [];
      let blank = 0;
      for (const line of lines) {
        if (line.trim() === '') {
          blank += 1;
          if (blank > 1) continue;
        } else {
          blank = 0;
        }
        collapsed.push(line);
      }
      lines = collapsed;
    }
    body = lines.join('\n').replace(/^\n+/, '').replace(/\n+$/, '');
    if (dedupeParagraphs) {
      const seen = new Set();
      body = body
        .split(/\n{2,}/)
        .filter(paragraph => {
          const key = paragraph.replace(/\s+/g, '');
          if (!key) return false;
          if (seen.has(key)) return false;
          seen.add(key);
          return true;
        })
        .join('\n\n');
    }
    const header = makeSeparator ? makeSeparator(file.title) : null;
    blocks.push(header ? `${header}\n${body}` : body);
  }

  const text = `${blocks.filter(Boolean).join('\n\n')}\n`.replace(/^\n+/, '');
  return {
    text,
    stats: {
      files: list.length,
      chars: text.length,
      perFile: list.map(file => ({ name: file.name, title: file.title, chapter: file.chapter, chars: String(file.content || '').length }))
    }
  };
}

/* ------------------------------------------------------------------ */
/* 文件读取                                                            */
/* ------------------------------------------------------------------ */

// 中文 TXT/SRT 常见 GBK 编码，File.text() 只会按 UTF-8 解，会得到乱码。
// 这里先按 BOM 判断 UTF-16，再严格试 UTF-8，失败后退到 GBK。
export async function readTextFile(file) {
  const buffer = await file.arrayBuffer();
  const bytes = new Uint8Array(buffer);
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    return new TextDecoder('utf-16le').decode(buffer).replace(/^\uFEFF/, '');
  }
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    return new TextDecoder('utf-16be').decode(buffer).replace(/^\uFEFF/, '');
  }
  try {
    return new TextDecoder('utf-8', { fatal: true }).decode(buffer).replace(/^\uFEFF/, '');
  } catch {
    try {
      return new TextDecoder('gbk').decode(buffer).replace(/^\uFEFF/, '');
    } catch {
      return new TextDecoder('utf-8').decode(buffer).replace(/^\uFEFF/, '');
    }
  }
}

export function downloadTextFile(name, value, type = 'text/plain;charset=utf-8', withBom = false) {
  const payload = withBom ? `\uFEFF${value}` : value;
  const url = URL.createObjectURL(new Blob([payload], { type }));
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = name;
  anchor.click();
  URL.revokeObjectURL(url);
}
