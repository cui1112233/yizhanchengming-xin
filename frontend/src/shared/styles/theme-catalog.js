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
  defineTheme('light', '明亮白', 'light', lightTokens),
  defineTheme('graywhite', '雾灰白', 'light', {
    ...lightTokens,
    background: '#eceff2', surface: '#f8f9fa', surfaceElevated: '#ffffff', surfaceDisabled: '#f0f2f4',
    textPrimary: '#20262e', textSecondary: '#4b5663', textMuted: '#687482', textDisabled: '#87919d',
    border: '#c7cdd4', borderStrong: '#9fa8b2', primary: '#526a82', primaryHover: '#445d75', primaryActive: '#384f65',
    canvasBackground: '#e7ebef', canvasGrid: '#c0c7cf', nodeSurface: '#f9fafb', edgeColor: '#72808d',
    viewportBackground: '#e1e6eb', timelineBackground: '#e8ecef'
  }),
  defineTheme('mint', '薄荷绿', 'light', {
    ...lightTokens,
    background: '#eef8f2', surface: '#f9fdfb', surfaceElevated: '#ffffff', surfaceDisabled: '#f0f8f4',
    textPrimary: '#17372d', textSecondary: '#355f52', textMuted: '#587b70', textDisabled: '#77978d',
    border: '#bdd9cc', borderStrong: '#8ebaaa', primary: '#16866c', primaryHover: '#11745d', primaryActive: '#0d604e',
    canvasBackground: '#e8f4ed', canvasGrid: '#b7d7c8', nodeSurface: '#fbfefc', edgeColor: '#4f8978',
    viewportBackground: '#e2f0e8', timelineBackground: '#e7f3ec'
  }),
  defineTheme('snow', '雪山蓝', 'light', {
    ...lightTokens,
    background: '#eef5fb', surface: '#f8fbfe', surfaceElevated: '#ffffff', surfaceDisabled: '#eff5fa',
    textPrimary: '#183247', textSecondary: '#3d6079', textMuted: '#628098', textDisabled: '#8099ab',
    border: '#bfd4e4', borderStrong: '#91b3cb', primary: '#347fa8', primaryHover: '#286e95', primaryActive: '#205b7d',
    canvasBackground: '#e7f1f8', canvasGrid: '#b8d0e1', nodeSurface: '#fbfdff', edgeColor: '#5f839b',
    viewportBackground: '#e0edf6', timelineBackground: '#e5f0f7'
  }),
  defineTheme('vanilla', '香草黄', 'light', {
    ...lightTokens,
    background: '#fbf6e8', surface: '#fffdf7', surfaceElevated: '#ffffff', surfaceDisabled: '#f8f2e3',
    textPrimary: '#3b3018', textSecondary: '#665637', textMuted: '#827354', textDisabled: '#9b8c70',
    border: '#e2d3aa', borderStrong: '#c7af72', primary: '#a56c16', primaryHover: '#8e5c11', primaryActive: '#754b0d',
    canvasBackground: '#f8f0dc', canvasGrid: '#ddcca3', nodeSurface: '#fffdf8', edgeColor: '#937b51',
    viewportBackground: '#f4ead2', timelineBackground: '#f7eedc'
  }),
  defineTheme('sakura', '樱花粉', 'light', {
    ...lightTokens,
    background: '#fbf0f3', surface: '#fff9fb', surfaceElevated: '#ffffff', surfaceDisabled: '#f9f0f3',
    textPrimary: '#492630', textSecondary: '#714653', textMuted: '#8c6671', textDisabled: '#a3848d',
    border: '#e4c3cc', borderStrong: '#c994a4', primary: '#b94f72', primaryHover: '#a54162', primaryActive: '#89344f',
    canvasBackground: '#f8e9ee', canvasGrid: '#dfbdc7', nodeSurface: '#fffafd', edgeColor: '#986879',
    viewportBackground: '#f4e2e8', timelineBackground: '#f7e8ed'
  }),
  defineTheme('pureblack', '纯黑', 'dark', {
    ...darkTokens,
    background: '#000000', surface: '#080808', surfaceElevated: '#111111', surfaceDisabled: '#070707',
    textPrimary: '#f7f8fb', textSecondary: '#d2d6dd', textMuted: '#aeb5c0', textDisabled: '#747b85',
    border: '#2a2d31', borderStrong: '#474c53', primary: '#63dfc0', primaryHover: '#7de8cf', primaryActive: '#4fcbb0',
    canvasBackground: '#000000', canvasGrid: '#24282d', nodeSurface: '#0b0d0f', edgeColor: '#8a929d',
    viewportBackground: '#000000', timelineBackground: '#050505'
  }),
  defineTheme('purple', '鸢尾紫', 'light', {
    ...lightTokens,
    background: '#f4f0fb', surface: '#fbf9ff', surfaceElevated: '#ffffff', surfaceDisabled: '#f3eff9',
    textPrimary: '#302346', textSecondary: '#584775', textMuted: '#75658f', textDisabled: '#9185a4',
    border: '#d2c4e5', borderStrong: '#ad98cc', primary: '#7253a6', primaryHover: '#624493', primaryActive: '#51377d',
    canvasBackground: '#eee8f7', canvasGrid: '#c9badd', nodeSurface: '#fcfaff', edgeColor: '#806b9d',
    viewportBackground: '#e9e1f4', timelineBackground: '#ede7f6'
  }),
  defineTheme('willow', '柳芽黄', 'light', {
    ...lightTokens,
    background: '#f7f7e8', surface: '#fdfdf5', surfaceElevated: '#ffffff', surfaceDisabled: '#f5f5e8',
    textPrimary: '#353719', textSecondary: '#5d6137', textMuted: '#787c54', textDisabled: '#949778',
    border: '#d7d9ad', borderStrong: '#b7bb77', primary: '#747f20', primaryHover: '#626c1a', primaryActive: '#515916',
    canvasBackground: '#f2f3df', canvasGrid: '#cfd29f', nodeSurface: '#fefef8', edgeColor: '#858a54',
    viewportBackground: '#edefd7', timelineBackground: '#f1f2df'
  }),
  defineTheme('ivory', '象牙白', 'light', {
    ...lightTokens,
    background: '#f8f5ec', surface: '#fdfbf5', surfaceElevated: '#ffffff', surfaceDisabled: '#f5f1e8',
    textPrimary: '#343027', textSecondary: '#5f594d', textMuted: '#7b7467', textDisabled: '#989084',
    border: '#ddd5c4', borderStrong: '#bdb29e', primary: '#7b6748', primaryHover: '#69583e', primaryActive: '#574934',
    canvasBackground: '#f3efe5', canvasGrid: '#d2c9b8', nodeSurface: '#fffdf8', edgeColor: '#887d6b',
    viewportBackground: '#eee9dd', timelineBackground: '#f2ede3'
  }),
  defineTheme('lotus', '荷花白', 'light', {
    ...lightTokens,
    background: '#f8f6f4', surface: '#fffdfc', surfaceElevated: '#ffffff', surfaceDisabled: '#f5f2f0',
    textPrimary: '#3a2e2d', textSecondary: '#655251', textMuted: '#806d6b', textDisabled: '#9a8987',
    border: '#dfd0ce', borderStrong: '#c1aaa7', primary: '#9a5f65', primaryHover: '#854f55', primaryActive: '#6f4146',
    canvasBackground: '#f3eeec', canvasGrid: '#d6c5c2', nodeSurface: '#fffdfc', edgeColor: '#8e7472',
    viewportBackground: '#eee7e5', timelineBackground: '#f2ecea'
  })
]);

const byId = new Map(definitions.map(definition => [definition.id, definition]));

export function listThemeDefinitions() {
  return definitions.map(definition => ({ ...definition, tokens: { ...definition.tokens } }));
}

export function normalizeThemeId(value) {
  const id = typeof value === 'string' ? value.trim().toLowerCase() : '';
  return byId.has(id) ? id : 'dark';
}

export function resolveThemeId(value, systemMode = 'dark') {
  const id = normalizeThemeId(value);
  if (id !== 'system') return id;
  return systemMode === 'light' ? 'light' : 'dark';
}

export function themeDefinition(value, systemMode = 'dark') {
  return byId.get(resolveThemeId(value, systemMode)) || byId.get('dark');
}

