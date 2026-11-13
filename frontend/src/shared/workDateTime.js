const workDateFormatter = new Intl.DateTimeFormat('en-GB', {
  timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit',
  hour: '2-digit', minute: '2-digit', hourCycle: 'h23'
});

// Display only: keep the original timestamps for sorting, storage and API calls.
export function formatWorkDateTime(value, fallback = '—') {
  if (value === null || value === undefined || value === '') return fallback;
  const date = value instanceof Date ? value : new Date(value);
  if (!Number.isFinite(date.getTime()) || date.getUTCFullYear() < 1900) return fallback;
  const parts = Object.fromEntries(workDateFormatter.formatToParts(date).map(part => [part.type, part.value]));
  return `${parts.year.slice(-2)}-${parts.month}-${parts.day}|${parts.hour}：${parts.minute}`;
}
