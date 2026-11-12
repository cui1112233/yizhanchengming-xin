import { cinematicWallpapers, windowVideoDefinition } from './cinematic-wallpapers.js';
export { windowVideoDefinition } from './cinematic-wallpapers.js';
// 浅色/深色各一套底，避免只依赖 --surface-* 时被深色变量盖住。
const lightBase = 'linear-gradient(165deg, color-mix(in srgb, var(--theme-primary, #f07167) 4%, #f6f7fb) 0%, #ffffff 62%, color-mix(in srgb, var(--theme-primary, #f07167) 3%, #f6f7fb) 100%)';
const darkBase = 'linear-gradient(165deg, color-mix(in srgb, var(--theme-primary, #f07167) 12%, #070b11) 0%, #0e1621 62%, #070b11 100%)';
const grain = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='320' height='320'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.25' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='320' height='320' filter='url(%23n)' opacity='0.6'/%3E%3C/svg%3E")`;

// 每个预设都带 light/dark 两套配方：preview 是浅色，previewDark 是深色。
// 颜色只引用语义变量或明确的浅/深底色，随风格切换仍会重新混色。
export const windowBackgrounds = [
  { id: 'none', label: '不使用', preview: 'none' },
  {
    id: 'default',
    label: '双角光',
    preview: [
      'radial-gradient(58% 52% at 10% 6%, color-mix(in srgb, var(--theme-primary, #f07167) 40%, transparent) 0%, transparent 68%)',
      'radial-gradient(52% 48% at 92% 96%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 34%, transparent) 0%, transparent 70%)',
      lightBase
    ].join(', '),
    previewDark: [
      'radial-gradient(58% 52% at 10% 6%, color-mix(in srgb, var(--theme-primary, #f07167) 56%, transparent) 0%, transparent 68%)',
      'radial-gradient(52% 48% at 92% 96%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 46%, transparent) 0%, transparent 70%)',
      darkBase
    ].join(', ')
  },
  {
    id: 'preset2',
    label: '顶部晕染',
    preview: [
      'radial-gradient(120% 62% at 50% -12%, color-mix(in srgb, var(--theme-primary, #f07167) 42%, transparent) 0%, transparent 72%)',
      'radial-gradient(60% 40% at 84% 30%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 24%, transparent) 0%, transparent 70%)',
      'linear-gradient(180deg, #ffffff, #f6f7fb)'
    ].join(', '),
    previewDark: [
      'radial-gradient(120% 62% at 50% -12%, color-mix(in srgb, var(--theme-primary, #f07167) 56%, transparent) 0%, transparent 72%)',
      'radial-gradient(60% 40% at 84% 30%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 40%, transparent) 0%, transparent 70%)',
      'linear-gradient(180deg, #0e1621, #070b11)'
    ].join(', ')
  },
  {
    id: 'preset3',
    label: '斜向光带',
    preview: [
      'linear-gradient(116deg, transparent 10%, color-mix(in srgb, var(--theme-primary, #f07167) 38%, transparent) 32%, transparent 60%)',
      'radial-gradient(70% 60% at 88% 4%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 28%, transparent) 0%, transparent 72%)',
      'linear-gradient(200deg, #f6f7fb, #ffffff)'
    ].join(', '),
    previewDark: [
      'linear-gradient(116deg, transparent 10%, color-mix(in srgb, var(--theme-primary, #f07167) 54%, transparent) 32%, transparent 60%)',
      'radial-gradient(70% 60% at 88% 4%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 44%, transparent) 0%, transparent 72%)',
      'linear-gradient(200deg, #070b11, #0e1621)'
    ].join(', ')
  },
  {
    id: 'preset4',
    label: '细网格',
    preview: [
      'repeating-linear-gradient(0deg, color-mix(in srgb, #334155 7%, transparent) 0 1px, transparent 1px 30px)',
      'repeating-linear-gradient(90deg, color-mix(in srgb, #334155 7%, transparent) 0 1px, transparent 1px 30px)',
      'radial-gradient(70% 60% at 14% -6%, color-mix(in srgb, var(--theme-primary, #f07167) 34%, transparent) 0%, transparent 70%)',
      'linear-gradient(170deg, #f6f7fb, #ffffff)'
    ].join(', '),
    previewDark: [
      'repeating-linear-gradient(0deg, color-mix(in srgb, #cbd5e1 8%, transparent) 0 1px, transparent 1px 30px)',
      'repeating-linear-gradient(90deg, color-mix(in srgb, #cbd5e1 8%, transparent) 0 1px, transparent 1px 30px)',
      'radial-gradient(70% 60% at 14% -6%, color-mix(in srgb, var(--theme-primary, #f07167) 44%, transparent) 0%, transparent 70%)',
      'linear-gradient(170deg, #070b11, #0e1621)'
    ].join(', ')
  },
  {
    id: 'preset5',
    label: '极光幕',
    preview: [
      'conic-gradient(from 210deg at 72% 18%, transparent 0deg, color-mix(in srgb, var(--theme-primary, #f07167) 34%, transparent) 70deg, transparent 150deg, color-mix(in srgb, var(--status-info-text, #3b4bc8) 28%, transparent) 230deg, transparent 330deg)',
      'linear-gradient(170deg, #f6f7fb, #ffffff)'
    ].join(', '),
    previewDark: [
      'conic-gradient(from 210deg at 72% 18%, transparent 0deg, color-mix(in srgb, var(--theme-primary, #f07167) 54%, transparent) 70deg, transparent 150deg, color-mix(in srgb, var(--status-info-text, #3b4bc8) 44%, transparent) 230deg, transparent 330deg)',
      'linear-gradient(170deg, #070b11, #0e1621)'
    ].join(', ')
  },
  // 图像感预设：仍全部是纯 CSS / 内联 SVG，不引入图片文件；
  // 浅色用浅底 + 浅色光斑，深色单独增强一档，避免发灰。
  {
    id: 'preset6',
    label: '极光帘幕',
    preview: [
      'repeating-linear-gradient(100deg, transparent 0 7%, color-mix(in srgb, var(--theme-primary, #f07167) 40%, transparent) 9%, transparent 13%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 48%, transparent) 16%, transparent 22%)',
      'radial-gradient(80% 70% at 18% 12%, color-mix(in srgb, var(--theme-primary, #f07167) 30%, transparent) 0%, transparent 68%)',
      'radial-gradient(65% 60% at 88% 88%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 40%, transparent) 0%, transparent 70%)',
      lightBase
    ].join(', '),
    previewDark: [
      'repeating-linear-gradient(100deg, transparent 0 7%, color-mix(in srgb, var(--theme-primary, #f07167) 80%, transparent) 9%, transparent 13%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 92%, transparent) 16%, transparent 22%)',
      'radial-gradient(80% 70% at 18% 12%, color-mix(in srgb, var(--theme-primary, #f07167) 66%, transparent) 0%, transparent 68%)',
      'radial-gradient(65% 60% at 88% 88%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 80%, transparent) 0%, transparent 70%)',
      darkBase
    ].join(', ')
  },
  {
    id: 'preset7',
    label: '细颗粒噪点',
    preview: [
      grain,
      'radial-gradient(70% 60% at 16% 10%, color-mix(in srgb, var(--theme-primary, #f07167) 30%, transparent) 0%, transparent 70%)',
      'radial-gradient(60% 55% at 86% 92%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 40%, transparent) 0%, transparent 72%)',
      'linear-gradient(165deg, color-mix(in srgb, var(--theme-primary, #f07167) 6%, #f6f7fb), #ffffff)'
    ].join(', '),
    previewDark: [
      grain,
      'radial-gradient(70% 60% at 16% 10%, color-mix(in srgb, var(--theme-primary, #f07167) 66%, transparent) 0%, transparent 70%)',
      'radial-gradient(60% 55% at 86% 92%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 80%, transparent) 0%, transparent 72%)',
      'linear-gradient(165deg, color-mix(in srgb, var(--theme-primary, #f07167) 18%, #070b11), #0e1621)'
    ].join(', '),
    size: '320px 320px, cover, cover, cover'
  },
  {
    id: 'preset8',
    label: '像素马赛克',
    preview: [
      'repeating-conic-gradient(from 0deg at 50% 50%, color-mix(in srgb, var(--theme-primary, #f07167) 30%, transparent) 0 25%, transparent 0 50%)',
      'radial-gradient(75% 80% at 88% 82%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 48%, transparent) 0%, transparent 74%)',
      'linear-gradient(150deg, transparent 24%, rgba(100, 116, 139, 0.10) 100%)',
      lightBase
    ].join(', '),
    previewDark: [
      'repeating-conic-gradient(from 0deg at 50% 50%, color-mix(in srgb, var(--theme-primary, #f07167) 66%, transparent) 0 25%, transparent 0 50%)',
      'radial-gradient(75% 80% at 88% 82%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 92%, transparent) 0%, transparent 74%)',
      'linear-gradient(150deg, transparent 24%, rgba(7, 11, 17, 0.45) 100%)',
      darkBase
    ].join(', '),
    size: '18px 18px, cover, cover, cover'
  },
  {
    id: 'preset9',
    label: '斜向光带·层叠',
    preview: [
      'linear-gradient(112deg, transparent 6%, color-mix(in srgb, var(--theme-primary, #f07167) 62%, transparent) 16% 22%, transparent 30%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 54%, transparent) 44% 48%, transparent 56%, color-mix(in srgb, var(--text-success, #75ddb4) 54%, transparent) 70% 74%, transparent 84%)',
      'radial-gradient(90% 65% at 10% 6%, color-mix(in srgb, var(--theme-primary, #f07167) 30%, transparent) 0%, transparent 72%)',
      'radial-gradient(60% 50% at 92% 96%, color-mix(in srgb, var(--text-success, #75ddb4) 30%, transparent) 0%, transparent 74%)',
      lightBase
    ].join(', '),
    previewDark: [
      'linear-gradient(112deg, transparent 6%, color-mix(in srgb, var(--theme-primary, #f07167) 100%, transparent) 16% 22%, transparent 30%, color-mix(in srgb, var(--status-info-text, #3b4bc8) 92%, transparent) 44% 48%, transparent 56%, color-mix(in srgb, var(--text-success, #75ddb4) 92%, transparent) 70% 74%, transparent 84%)',
      'radial-gradient(90% 65% at 10% 6%, color-mix(in srgb, var(--theme-primary, #f07167) 44%, transparent) 0%, transparent 72%)',
      'radial-gradient(60% 50% at 92% 96%, color-mix(in srgb, var(--text-success, #75ddb4) 44%, transparent) 0%, transparent 74%)',
      darkBase
    ].join(', ')
  }
  , ...[{ id: 'cinema-auto', label: '映窗·四景轮播', poster: cinematicWallpapers[0].poster }, ...cinematicWallpapers].map(item => ({ ...item, video: true, preview: `url("${item.poster}")` }))
];

// Retain legacy recipes only for reading existing account preferences.
export const windowBackgroundChoices = windowBackgrounds.filter(item => item.video);

export const windowBackgroundMotions = [
  { id: 'static', label: '静态' },
  { id: 'soft', label: '柔和' },
  { id: 'strong', label: '明显' }
];

const MOTION_KEY = 'v88-window-bg-motion';
const MOTION_IDS = new Set(windowBackgroundMotions.map(item => item.id));

const supported = new Set([...windowBackgrounds.map(item => item.id), 'custom']);

export function normalizeWindowBackgroundId(value, customUrl = '') {
  const id = typeof value === 'string' ? value.trim() : '';
  if (id === 'custom' && !customUrl) return 'none';
  return supported.has(id) ? id : 'none';
}

// 主题模式必须显式传入：切换主题时 html[data-theme] 可能还没更新，
// 只读 DOM 会让背景图滞后一拍，于是浅色主题仍沿用深色配方。
export function resolveWindowBackgroundMode(mode) {
  if (mode === 'light' || mode === 'dark') return mode;
  if (typeof document !== 'undefined') {
    return document.documentElement?.dataset?.theme === 'light' ? 'light' : 'dark';
  }
  return 'dark';
}

export function windowBackgroundPreview(item, mode) {
  if (!item || item.id === 'none') return 'none';
  return resolveWindowBackgroundMode(mode) === 'dark'
    ? (item.previewDark || item.preview)
    : (item.preview || item.previewDark);
}

export function windowBackgroundStyle(id, customUrl = '', mode = '') {
  const normalized = normalizeWindowBackgroundId(id, customUrl);
  if (normalized === 'none') return {};
  if (normalized === 'custom') return { backgroundImage: `url(${JSON.stringify(customUrl)})` };
  const preset = windowBackgrounds.find(item => item.id === normalized);
  const style = { backgroundImage: windowBackgroundPreview(preset, mode) };
  // 噪点/马赛克需要按图层指定尺寸，其余预设仍交给 CSS 的 cover 处理。
  if (preset?.size) style.backgroundSize = preset.size;
  return style;
}

export function normalizeWindowBackgroundMotion(value) {
  const id = typeof value === 'string' ? value.trim() : '';
  return MOTION_IDS.has(id) ? id : 'soft';
}

// 自定义图片只允许柔和动效，明显档会退化为柔和。
export function effectiveWindowBackgroundMotion(backgroundId, motion) {
  const normalized = normalizeWindowBackgroundMotion(motion);
  return backgroundId === 'custom' && normalized === 'strong' ? 'soft' : normalized;
}

export function readWindowBackgroundMotion() {
  if (typeof window === 'undefined') return 'soft';
  try {
    return normalizeWindowBackgroundMotion(window.localStorage.getItem(MOTION_KEY));
  } catch (error) {
    return 'soft';
  }
}

export function applyWindowBackgroundMotion(motion) {
  if (typeof document === 'undefined') return 'soft';
  const normalized = normalizeWindowBackgroundMotion(motion);
  document.documentElement.dataset.windowBgMotion = normalized;
  return normalized;
}

export function saveWindowBackgroundMotion(motion) {
  const normalized = applyWindowBackgroundMotion(motion);
  if (typeof window !== 'undefined') {
    try { window.localStorage.setItem(MOTION_KEY, normalized); } catch (error) { /* 隐私模式下忽略 */ }
  }
  return normalized;
}

// 首屏就按本地偏好生效，避免只在设置页打开时才看得到动效。
applyWindowBackgroundMotion(readWindowBackgroundMotion());
