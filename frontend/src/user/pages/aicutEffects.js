import { catalogForTab } from './aicut/editorCatalog.js';

export const transitionCatalog = [
  { id: 'cross-dissolve', label: '交叉溶解', description: '前后画面自然叠化' },
  { id: 'fade', label: '淡入淡出', description: '柔和切换画面' },
  { id: 'flash-black', label: '黑场过渡', description: '经过短暂黑场' },
  { id: 'flash-white', label: '白闪', description: '经过短暂白场' },
  { id: 'slide-left', label: '向左推拉', description: '下一镜头从右侧进入' },
  { id: 'slide-right', label: '向右推拉', description: '下一镜头从左侧进入' },
  { id: 'slide-up', label: '向上推拉', description: '下一镜头从下方进入' },
  { id: 'slide-down', label: '向下推拉', description: '下一镜头从上方进入' },
  { id: 'zoom', label: '缩放过渡', description: '通过缩放连接镜头' },
  { id: 'blur', label: '模糊过渡', description: '模糊后切换到下一镜头' },
];

export const visualEffectCatalog = [
  { id: 'original', label: '原片' }, { id: 'bright', label: '明亮' },
  { id: 'cool-cinema', label: '冷色调' }, { id: 'warm', label: '暖色调' },
  { id: 'grayscale', label: '黑白' }, { id: 'soft-focus', label: '画面模糊' },
  { id: 'vignette', label: '暗角' }, { id: 'glow', label: '发光' },
  { id: 'shake', label: '镜头抖动' }, { id: 'breathe', label: '呼吸缩放' },
  { id: 'flash', label: '闪烁' }, { id: 'mosaic', label: '马赛克' },
];

export const textStyleCatalog = [
  { id: 'plain', label: '普通文字' }, { id: 'title', label: '标题' },
  { id: 'subtitle', label: '副标题' }, { id: 'outline', label: '描边文字' },
  { id: 'shadow', label: '阴影文字' }, { id: 'box', label: '背景框文字' },
  { id: 'gradient', label: '渐变文字' }, { id: 'vertical', label: '竖排文字' },
];

export const subtitleStyleCatalog = [
  { id: 'classic-outline', label: '经典白字黑边' }, { id: 'yellow-emphasis', label: '黄色强调字幕' },
  { id: 'translucent-box', label: '白字半透明底' }, { id: 'short-drama', label: '短剧大字字幕' },
  { id: 'dual-color', label: '上下双色字幕' }, { id: 'word-highlight', label: '逐字高亮' },
];

export const subtitleAnimationCatalog = [
  { id: 'none', label: '无动效' }, { id: 'fade', label: '淡入淡出' },
  { id: 'pop', label: '弹入' }, { id: 'float-up', label: '上浮' },
  { id: 'typewriter', label: '打字机' }, { id: 'word-highlight', label: '逐字高亮' },
];

export const visualEffectPresets = visualEffectCatalog.map(item => [item.id, item.label]);

export const editorFilterCatalog = catalogForTab('filter');
export const editorEffectCatalog = catalogForTab('effect');
export const editorTransitionCatalog = catalogForTab('transition');
export const editorTextCatalog = catalogForTab('text');

const presets = {
  original: { preset: 'original', brightness: 0, contrast: 1, saturation: 1, blur: 0, grayscale: false },
  bright: { preset: 'bright', brightness: .12, contrast: 1.05, saturation: 1.08, blur: 0, grayscale: false },
  'cool-cinema': { preset: 'cool-cinema', brightness: -.04, contrast: 1.18, saturation: .82, blur: 0, grayscale: false },
  warm: { preset: 'warm', brightness: .06, contrast: 1.08, saturation: 1.25, blur: 0, grayscale: false },
  grayscale: { preset: 'grayscale', brightness: 0, contrast: 1.05, saturation: 0, blur: 0, grayscale: true },
  'soft-focus': { preset: 'soft-focus', brightness: .05, contrast: .95, saturation: 1.05, blur: 2, grayscale: false },
  vignette: { preset: 'vignette', brightness: 0, contrast: 1.05, saturation: 1, blur: 0, grayscale: false, intensity: .55, speed: 1 },
  glow: { preset: 'glow', brightness: .08, contrast: 1.05, saturation: 1.08, blur: 0, grayscale: false, intensity: .5, speed: 1 },
  shake: { preset: 'shake', brightness: 0, contrast: 1, saturation: 1, blur: 0, grayscale: false, intensity: .45, speed: 1 },
  breathe: { preset: 'breathe', brightness: 0, contrast: 1, saturation: 1, blur: 0, grayscale: false, intensity: .35, speed: 1 },
  flash: { preset: 'flash', brightness: .08, contrast: 1, saturation: 1, blur: 0, grayscale: false, intensity: .45, speed: 1 },
  mosaic: { preset: 'mosaic', brightness: 0, contrast: 1, saturation: 1, blur: 0, grayscale: false, intensity: .5, speed: 1 },
};

function finite(value, fallback) {
  const number = Number(value);
  return Number.isFinite(number) ? number : fallback;
}

function clamp(value, minimum, maximum) {
  return Math.min(maximum, Math.max(minimum, value));
}

export function visualEffectPreset(id) {
  return { ...(presets[id] || presets.original) };
}

export function normalizeVisualEffect(value) {
  if (!value) return visualEffectPreset('original');
  const preset = presets[value.preset] ? value.preset : 'original';
  return {
    preset: value.preset === 'custom' ? 'custom' : preset,
    brightness: clamp(finite(value.brightness, 0), -1, 1),
    contrast: clamp(finite(value.contrast, 1), .5, 2),
    saturation: clamp(finite(value.saturation, 1), 0, 3),
    blur: clamp(finite(value.blur, 0), 0, 20),
    grayscale: Boolean(value.grayscale),
    intensity: clamp(finite(value.intensity, presets[preset]?.intensity ?? .5), 0, 1),
    speed: clamp(finite(value.speed, presets[preset]?.speed ?? 1), .1, 5),
  };
}

export function visualEffectCSS(value) {
  const effect = normalizeVisualEffect(value);
  const grayscale = effect.grayscale ? ' grayscale(1)' : '';
  const glow = effect.preset === 'glow' ? ` drop-shadow(0 0 ${Math.round(18 * effect.intensity)}px rgba(255,255,255,.85))` : '';
  return `brightness(${1 + effect.brightness}) contrast(${effect.contrast}) saturate(${effect.saturation}) blur(${effect.blur}px)${grayscale}${glow}`;
}

export function visualEffectPreviewStyle(value, elapsedMs = 0) {
  const effect = normalizeVisualEffect(value);
  const phase = Number(elapsedMs || 0) / 1000 * effect.speed * Math.PI * 2;
  const style = { filter: visualEffectCSS(effect) };
  if (effect.preset === 'vignette') style.boxShadow = `inset 0 0 ${Math.round(140 * effect.intensity)}px rgba(0,0,0,.9)`;
  if (effect.preset === 'shake') style.transform = `translate(${Math.sin(phase * 3) * 8 * effect.intensity}px, ${Math.cos(phase * 4) * 6 * effect.intensity}px) scale(1.02)`;
  if (effect.preset === 'breathe') style.transform = `scale(${1 + (Math.sin(phase) + 1) * .025 * effect.intensity})`;
  if (effect.preset === 'flash') style.opacity = Math.max(.2, 1 - (Math.sin(phase) + 1) * .35 * effect.intensity);
  if (effect.preset === 'mosaic') {
    style.imageRendering = 'pixelated';
    style.filter = `${style.filter} blur(${Math.max(1, Math.round(effect.intensity * 4))}px)`;
  }
  return style;
}
