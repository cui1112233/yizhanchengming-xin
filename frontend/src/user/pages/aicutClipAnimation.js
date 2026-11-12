export const clipAnimationPresets = {
  entrance: [{ value: 'none', label: '无' }, { value: 'fade', label: '淡入' }, { value: 'slide', label: '从左滑入' }, { value: 'zoom', label: '缩放进入' }],
  exit: [{ value: 'none', label: '无' }, { value: 'fade', label: '淡出' }, { value: 'slide', label: '向右滑出' }, { value: 'zoom', label: '缩放退出' }],
  loop: [{ value: 'none', label: '无' }, { value: 'pulse', label: '轻微呼吸' }, { value: 'float', label: '轻微浮动' }],
};
const clamp = (value, low, high) => Math.min(high, Math.max(low, value));
const finite = (value, fallback) => Number.isFinite(Number(value)) ? Number(value) : fallback;

// Command writes must reject invalid inputs; tolerant normalization is only for reading.
export function validateClipAnimation(value) {
  if (value === null || value === undefined) return null;
  if (typeof value !== 'object' || Array.isArray(value)) throw new Error('动画设置格式不正确');
  for (const [phase, options] of Object.entries(clipAnimationPresets)) {
    const setting = value[phase];
    if (setting == null) continue;
    if (typeof setting !== 'object' || Array.isArray(setting)) throw new Error('动画设置格式不正确');
    if (setting.preset !== undefined && !options.some(option => option.value === setting.preset)) throw new Error('暂不支持这个动画');
    if (setting.durationMs !== undefined && (typeof setting.durationMs !== 'number' || !Number.isFinite(setting.durationMs) || !Number.isInteger(setting.durationMs) || setting.durationMs < 50 || setting.durationMs > 10000)) throw new Error('动画时长应为 0.05 到 10 秒');
  }
  return normalizeClipAnimation(value);
}

export function normalizeClipAnimation(value) {
  return Object.fromEntries(Object.entries(clipAnimationPresets).map(([phase, options]) => [phase, {
    preset: options.some(option => option.value === value?.[phase]?.preset) ? value[phase].preset : 'none',
    durationMs: clamp(Math.round(finite(value?.[phase]?.durationMs, phase === 'loop' ? 2000 : 500)), 50, 10000),
  }]));
}

export function clipAnimationDurations(value, durationMs) {
  const settings = normalizeClipAnimation(value);
  let entranceMs = settings.entrance.preset === 'none' ? 0 : settings.entrance.durationMs;
  let exitMs = settings.exit.preset === 'none' ? 0 : settings.exit.durationMs;
  const ratio = Math.min(1, Math.max(0, finite(durationMs, 0)) / Math.max(1, entranceMs + exitMs));
  entranceMs *= ratio;
  exitMs *= ratio;
  return { entranceMs, exitMs };
}

// All phases use output time, including reversed, frozen and retimed clips.
export function clipAnimationFrame(value, durationMs, offsetMs) {
  const settings = normalizeClipAnimation(value);
  const duration = Math.max(0, finite(durationMs, 0));
  const time = clamp(finite(offsetMs, 0), 0, duration);
  const { entranceMs, exitMs } = clipAnimationDurations(settings, duration);
  const frame = { x: 0, y: 0, scale: 1, opacity: 1 };
  for (const [phase, amount] of [
    ['entrance', entranceMs > 0 ? 1 - clamp(time / entranceMs, 0, 1) : 0],
    ['exit', exitMs > 0 ? clamp((time - duration + exitMs) / exitMs, 0, 1) : 0],
  ]) {
    const preset = settings[phase].preset;
    if (preset === 'fade') frame.opacity *= 1 - amount;
    if (preset === 'slide') frame.x += (phase === 'entrance' ? -1 : 1) * amount;
    if (preset === 'zoom') frame.scale *= 1 - .2 * amount;
  }
  const angle = time / settings.loop.durationMs * Math.PI * 2;
  if (settings.loop.preset === 'pulse') frame.scale *= 1 + .025 * (1 - Math.cos(angle));
  if (settings.loop.preset === 'float') frame.y += .015 * Math.sin(angle);
  return frame;
}

export function clipAnimationPreviewStyle(value, durationMs, offsetMs, baseStyle = {}) {
  const frame = clipAnimationFrame(value, durationMs, offsetMs);
  return {
    ...baseStyle,
    transform: `translate(${frame.x * 100}%, ${frame.y * 100}%) scale(${frame.scale}) ${baseStyle.transform || ''}`.trim(),
    transformOrigin: '50% 50%',
    opacity: finite(baseStyle.opacity, 1) * frame.opacity,
  };
}
