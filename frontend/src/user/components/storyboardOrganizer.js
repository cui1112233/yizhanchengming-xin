// 分镜整理：把批量导入的分镜文档解析成镜头列表，再按 15 秒 / 30 秒分组导出 JSON。
// 纯浏览器本地处理，不调用任何模型。
import { parseSubtitle, parseSubtitleTime } from './localToolLogic.js';

const FIELD_ALIASES = {
  duration: ['时长', '时间', 'duration', 'seconds', '秒数'],
  content: ['画面', '内容', '描述', '镜头内容', 'scene', 'picture', 'visual'],
  dialogue: ['台词', '对白', '旁白', '字幕', 'dialogue', 'line'],
  camera: ['运镜', '镜头', '景别', '机位', 'camera', 'shot'],
  audio: ['音效', '配乐', '音乐', 'audio', 'sound'],
  note: ['备注', '说明', 'note']
};

const SHOT_MARKER = /^(?:#{1,6}\s*)?(?:镜头|分镜|镜号|shot|镜)\s*[#第]?\s*(\d{1,4})/i;
const NUMBERED_MARKER = /^(\d{1,4})\s*[.、，)]\s*/;

function normalize(text) {
  return String(text || '').replace(/\r\n?/g, '\n').replace(/^\uFEFF/, '');
}

function splitField(line) {
  const matched = line.match(/^([^：:]{1,12})[：:]\s*(.*)$/);
  if (!matched) return null;
  const label = matched[1].trim().toLowerCase();
  for (const [key, aliases] of Object.entries(FIELD_ALIASES)) {
    if (aliases.some(alias => label === alias.toLowerCase())) return { key, value: matched[2].trim() };
  }
  return null;
}

function extractDuration(text, fallback) {
  const source = String(text || '');
  const seconds = source.match(/(\d+(?:\.\d+)?)\s*(?:秒|s\b|sec\b)/i);
  if (seconds) return Number(seconds[1]);
  const clock = source.match(/(\d{1,2}):(\d{2})(?::(\d{2}))?/);
  if (clock) {
    const first = Number(clock[1]);
    const second = Number(clock[2]);
    const third = clock[3] === undefined ? null : Number(clock[3]);
    return third === null ? first * 60 + second : first * 3600 + second * 60 + third;
  }
  const numeric = source.match(/^\s*(\d+(?:\.\d+)?)\s*$/);
  if (numeric) return Number(numeric[1]);
  return fallback;
}

function buildShot({ index, raw, fields = {}, content = '', duration, source, start, end }) {
  return {
    index,
    source,
    duration: Number.isFinite(duration) ? Math.max(0, duration) : 0,
    start: start ?? null,
    end: end ?? null,
    content: String(content || '').trim(),
    dialogue: String(fields.dialogue || '').trim(),
    camera: String(fields.camera || '').trim(),
    audio: String(fields.audio || '').trim(),
    note: String(fields.note || '').trim(),
    raw: String(raw || '').trim()
  };
}

function parseTimed(content, fileName) {
  const parsed = parseSubtitle(content, { fileName });
  return parsed.cues.map((cue, index) => buildShot({
    index: index + 1,
    raw: cue.text,
    content: cue.text,
    duration: (cue.end - cue.start) / 1000,
    start: cue.start / 1000,
    end: cue.end / 1000,
    source: fileName
  }));
}

function splitTableRow(line) {
  if (line.includes('|')) return line.split('|').map(cell => cell.trim());
  if (line.includes('｜')) return line.split('｜').map(cell => cell.trim());
  if (line.includes('\t')) return line.split('\t').map(cell => cell.trim());
  return null;
}

function parseTable(content, fileName, defaultShotSeconds) {
  const lines = normalize(content).split('\n');
  const rows = [];
  for (const line of lines) {
    const cells = splitTableRow(line);
    if (!cells || cells.length < 2) {
      if (rows.length) break;
      continue;
    }
    if (cells.every(cell => !cell || /^[-: ]+$/.test(cell))) continue;
    rows.push(cells);
  }
  if (rows.length < 2) return [];

  const header = rows[0].map(cell => cell.toLowerCase());
  const findColumn = key => header.findIndex(cell => FIELD_ALIASES[key].some(alias => cell.includes(alias.toLowerCase())));
  const index = {
    duration: findColumn('duration'),
    content: findColumn('content'),
    dialogue: findColumn('dialogue'),
    camera: findColumn('camera'),
    audio: findColumn('audio'),
    note: findColumn('note')
  };
  if (index.duration === -1 && index.content === -1) return [];

  const shots = [];
  for (const cells of rows.slice(1)) {
    const fields = {};
    for (const key of ['dialogue', 'camera', 'audio', 'note']) {
      if (index[key] >= 0) fields[key] = cells[index[key]] || '';
    }
    const durationCell = index.duration >= 0 ? cells[index.duration] : '';
    const shotContent = index.content >= 0 ? cells[index.content] : cells.filter(Boolean).join(' ');
    if (!String(shotContent || '').trim() && !fields.dialogue) continue;
    shots.push(buildShot({
      index: shots.length + 1,
      raw: cells.join(' | '),
      fields,
      content: shotContent,
      duration: extractDuration(durationCell, defaultShotSeconds),
      source: fileName
    }));
  }
  return shots;
}

function parseBlocks(content, fileName, defaultShotSeconds) {
  const lines = normalize(content).split('\n');
  const blocks = [];
  let current = null;

  const startBlock = (marker, durationText, seed = '') => {
    current = { marker, lines: [], header: durationText || '', seed };
    blocks.push(current);
  };

  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!line) {
      if (current) current.lines.push('');
      continue;
    }
    const shotMatched = line.match(SHOT_MARKER);
    if (shotMatched) {
      const seed = line
        .replace(SHOT_MARKER, '')
        .replace(/^[）)】\]\s]*/, '')
        .replace(/[（(]?\s*\d+(?:\.\d+)?\s*(?:秒|s)\s*[)）]?/i, '')
        .trim();
      startBlock(shotMatched[1], line, seed);
      continue;
    }
    const numbered = line.match(NUMBERED_MARKER);
    if (numbered && !splitField(line)) {
      startBlock(numbered[1], line, line.replace(NUMBERED_MARKER, '').trim());
      continue;
    }
    if (!current) startBlock('', '');
    current.lines.push(line);
  }

  const shots = [];
  for (const block of blocks) {
    const fields = {};
    const rest = block.seed ? [block.seed] : [];
    let duration = extractDuration(block.header, null);
    for (const line of block.lines) {
      if (!line) continue;
      if (line.startsWith('画面') || line.startsWith('内容') || line.startsWith('描述')) {
        const field = splitField(line);
        if (field) {
          rest.push(field.value);
          continue;
        }
      }
      const field = splitField(line);
      if (field) {
        if (field.key === 'duration') duration = extractDuration(field.value, duration);
        else fields[field.key] = fields[field.key] ? `${fields[field.key]}\n${field.value}` : field.value;
        continue;
      }
      if (duration === null) {
        const matched = line.match(/[（(]?\s*(\d+(?:\.\d+)?)\s*(?:秒|s)\s*[)）]?/i);
        if (matched) {
          duration = Number(matched[1]);
          continue;
        }
      }
      rest.push(line);
    }
    const contentText = rest.join('\n').trim();
    if (!contentText && !fields.dialogue) continue;
    shots.push(buildShot({
      index: shots.length + 1,
      raw: [block.header, ...block.lines].join('\n'),
      fields,
      content: contentText,
      duration: duration === null ? defaultShotSeconds : duration,
      source: fileName
    }));
  }
  return shots;
}

export function parseStoryboardDocument(content, { fileName = '', defaultShotSeconds = 3 } = {}) {
  const source = normalize(content);
  if (!source.trim()) return [];
  if (source.includes('-->')) return parseTimed(source, fileName);
  const table = parseTable(source, fileName, defaultShotSeconds);
  if (table.length) return table;
  return parseBlocks(source, fileName, defaultShotSeconds);
}

export function parseStoryboardFiles(files, { defaultShotSeconds = 3 } = {}) {
  const shots = [];
  for (const file of Array.isArray(files) ? files : []) {
    const parsed = parseStoryboardDocument(file.content, { fileName: file.name, defaultShotSeconds });
    for (const shot of parsed) {
      shots.push({ ...shot, index: shots.length + 1 });
    }
  }
  return shots;
}

/**
 * 按目标时长分组：加入下一镜会超过目标时切组；单镜超过目标时独占一组。
 */
export function groupStoryboard(shots, targetSeconds = 15) {
  const target = Number(targetSeconds) > 0 ? Number(targetSeconds) : 15;
  const groups = [];
  let current = null;
  let cursor = 0;

  for (const shot of shots) {
    const duration = Math.max(0, Number(shot.duration) || 0);
    if (!current || (current.duration > 0 && current.duration + duration > target + 1e-6)) {
      current = { index: groups.length + 1, start: cursor, duration: 0, shots: [] };
      groups.push(current);
    }
    const placed = { ...shot, start: cursor, end: cursor + duration };
    cursor += duration;
    current.shots.push(placed);
    current.duration = Number((current.duration + duration).toFixed(3));
    current.end = cursor;
  }

  return groups;
}

export function buildStoryboardJson(shots, { targetSeconds = 15 } = {}) {
  const groups = groupStoryboard(shots, targetSeconds);
  const totalSeconds = Number(shots.reduce((sum, shot) => sum + (Number(shot.duration) || 0), 0).toFixed(3));
  return {
    targetSeconds: Number(targetSeconds),
    totalShots: shots.length,
    totalSeconds,
    groupCount: groups.length,
    groups: groups.map(group => ({
      index: group.index,
      startSeconds: Number(group.start.toFixed(3)),
      endSeconds: Number(Number(group.end || group.start).toFixed(3)),
      durationSeconds: group.duration,
      shots: group.shots.map(shot => ({
        index: shot.index,
        source: shot.source || '',
        durationSeconds: shot.duration,
        startSeconds: Number(shot.start.toFixed(3)),
        endSeconds: Number(shot.end.toFixed(3)),
        content: shot.content,
        dialogue: shot.dialogue,
        camera: shot.camera,
        audio: shot.audio,
        note: shot.note
      }))
    }))
  };
}

export function buildStoryboardPreview(shots, { targetSeconds = 15 } = {}) {
  const groups = groupStoryboard(shots, targetSeconds);
  return groups.map(group => {
    const head = `【第 ${group.index} 组】0:00 - ${formatClock(group.end || group.start)}（${group.duration} 秒，${group.shots.length} 镜）`;
    const body = group.shots.map(shot => {
      const parts = [`${shot.index}. ${shot.duration}秒`];
      if (shot.camera) parts.push(`[${shot.camera}]`);
      if (shot.content) parts.push(shot.content.replace(/\n/g, ' '));
      if (shot.dialogue) parts.push(`台词：${shot.dialogue.replace(/\n/g, ' ')}`);
      return `  ${parts.join(' ')}`;
    });
    return `${head}\n${body.join('\n')}`;
  }).join('\n\n');
}

function formatClock(totalSeconds) {
  const safe = Math.max(0, Math.round(Number(totalSeconds) || 0));
  const minutes = Math.floor(safe / 60);
  const seconds = safe % 60;
  return `${minutes}:${String(seconds).padStart(2, '0')}`;
}

export { parseSubtitleTime };
