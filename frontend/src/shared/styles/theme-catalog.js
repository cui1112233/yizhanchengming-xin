const sharedStatus = Object.freeze({
  success: '#21835f',
  warning: '#a76510',
  danger: '#c4424a',
  info: '#5366cf'
});

function defineTheme(id, label, colorScheme, tokens) {
  return Object.freeze({
    id,
    label,
    colorScheme,
    tokens: Object.freeze({ ...sharedStatus, ...tokens })
  });
}

const darkTokens = {
  background: '#070b11',
  surface: '#0e1621',
  surfaceElevated: '#141f2d',
  surfaceDisabled: '#0b111a',
  textPrimary: '#f5f7fb',
  textSecondary: '#c7d0dd',
  textMuted: '#a7b4c5',
  textDisabled: 'rgba(199, 208, 221, 0.62)',
  border: 'rgba(151, 168, 190, 0.22)',
  borderStrong: 'rgba(176, 195, 219, 0.36)',
  primary: '#f07167',
  primaryHover: '#f58a82',
  primaryActive: '#d85f57',
  success: '#75ddb4',
  warning: '#ffd08a',
  danger: '#ff9b9d',
  info: '#9da8ff',
  canvasBackground: '#090f17',
  canvasGrid: '#263446',
  nodeSurface: '#111c29',
  edgeColor: '#778394',
  viewportBackground: '#05080d',
  timelineBackground: '#0a111b'
};

const lightTokens = {
  background: '#f4f6fa',
  surface: '#ffffff',
  surfaceElevated: '#ffffff',
  surfaceDisabled: '#ffffff',
  textPrimary: '#172033',
  textSecondary: '#3f4e63',
  textMuted: '#596a80',
  textDisabled: '#748196',
  border: '#c8d2df',
  borderStrong: '#9daabd',
  primary: '#f07167',
  primaryHover: '#e45f56',
  primaryActive: '#c94d45',
  canvasBackground: '#eef2f7',
  canvasGrid: '#c7d0dc',
  nodeSurface: '#ffffff',
  edgeColor: '#69788b',
  viewportBackground: '#e8edf4',
  timelineBackground: '#edf1f6'
};

const definitions = Object.freeze([
  defineTheme('system', '跟随系统', 'system', darkTokens),
  defineTheme('dark', '深夜黑', 'dark', darkTokens),
  defineTheme('light', '明亮白', 'light', lightTokens)
]);

const byId = new Map(definitions.map(definition => [definition.id, definition]));

export function listThemeDefinitions() {
  return definitions.map(definition => ({ ...definition, tokens: { ...definition.tokens } }));
}

export function normalizeThemeId(value) {
  const id = typeof value === 'string' ? value.trim().toLowerCase() : '';
  // 账号/本地缓存里可能残留已下线的色板 id（graywhite、mint、pureblack…），统一回落到跟随系统。
  return byId.has(id) ? id : 'system';
}

export function resolveThemeId(value, systemMode = 'dark') {
  const id = normalizeThemeId(value);
  if (id !== 'system') return id;
  return systemMode === 'light' ? 'light' : 'dark';
}

export function themeDefinition(value, systemMode = 'dark') {
  return byId.get(resolveThemeId(value, systemMode)) || byId.get('dark');
}

