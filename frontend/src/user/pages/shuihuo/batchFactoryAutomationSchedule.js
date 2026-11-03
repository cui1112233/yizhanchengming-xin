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
