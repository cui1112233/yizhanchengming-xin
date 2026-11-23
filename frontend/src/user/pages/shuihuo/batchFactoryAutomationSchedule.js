const BEIJING_OFFSET_HOURS = 8;
const DATETIME_LOCAL = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/;

function pad(value) {
  return String(value).padStart(2, '0');
}

export function parseBeijingDatetimeLocal(value) {
  const matched = String(value || '').trim().match(DATETIME_LOCAL);
  if (!matched) return '';
  const [, year, month, day, hour, minute] = matched;
  const instant = new Date(Date.UTC(Number(year), Number(month) - 1, Number(day), Number(hour) - BEIJING_OFFSET_HOURS, Number(minute)));
  if (!Number.isFinite(instant.getTime())) return '';
  return instant.toISOString();
}

export function formatBeijingDatetimeLocal(value) {
  const instant = new Date(value);
  if (!Number.isFinite(instant.getTime())) return '';
  const beijing = new Date(instant.getTime() + BEIJING_OFFSET_HOURS * 60 * 60 * 1000);
  return `${beijing.getUTCFullYear()}-${pad(beijing.getUTCMonth() + 1)}-${pad(beijing.getUTCDate())}T${pad(beijing.getUTCHours())}:${pad(beijing.getUTCMinutes())}`;
}

export function normalizeAutomationConcurrency(value, fallback = 4) {
  const number = Number(value);
  return [1, 2, 4].includes(number) ? number : fallback;
}

export function automationPresetSnapshot(presets, presetId) {
  const selected = (Array.isArray(presets) ? presets : []).find(item => String(item?.id || '') === String(presetId || ''));
  const config = selected?.config && typeof selected.config === 'object' && !Array.isArray(selected.config) ? selected.config : null;
  if (!selected || !config || !Object.keys(config).length) return null;
  return {
    id: String(selected.id),
    name: String(selected.name || ''),
    version: Number(selected.version || 0),
    config: JSON.parse(JSON.stringify(config))
  };
}

// 历史预设没有生产模式，三个模式都可选择。
const presetModeLabels = new Map([['video', '视频'], ['comic', '漫画'], ['text', '文本']]);

export function automationPresetMode(preset) {
  const config = preset?.config && typeof preset.config === 'object' && !Array.isArray(preset.config) ? preset.config : {};
  const mode = String(config.productionMode || '').trim().toLowerCase();
  return presetModeLabels.has(mode) ? mode : '';
}

export function automationPresetModeLabel(mode) {
  return presetModeLabels.get(mode) || '通用';
}

// 只保留与当前批次模式匹配（或通用）的预设；mode 为空（无模式上下文）返回全部。
export function automationPresetsForMode(presets, mode) {
  if (!presetModeLabels.has(mode)) return Array.isArray(presets) ? [...presets] : [];
  return (Array.isArray(presets) ? presets : []).filter(preset => !automationPresetMode(preset) || automationPresetMode(preset) === mode);
}

// 把当前批次生产模式写进预设 config；mode 为空时保持原 config 不动。
export function withAutomationPresetMode(config, mode) {
  const next = config && typeof config === 'object' && !Array.isArray(config) ? { ...config } : {};
  if (presetModeLabels.has(mode)) next.productionMode = mode;
  return next;
}
