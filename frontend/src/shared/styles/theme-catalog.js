// 主题目录：风格（scheme）× 深浅（mode）两个维度。
// 色值唯一来源：docs/design/2026-11-12-ui-direction-research.md 第 3 章、
// docs/design/2026-11-12-theme-palette-spec.md 第 14.1 章、2026-11-12-ui-direction-preview.html。
// 文档未给值的角色（禁用字、画布、滚动条、阴影）沿用原深浅基础值，见 theme-schemes.css 顶部说明。

export const DEFAULT_SCHEME_ID = 'coral';

// 文档没有给出「禁用字 / 画布 / 画布网格」等角色的分风格取值，这里保留原有的深浅基础值。
const modeBase = Object.freeze({
  dark: Object.freeze({
    surfaceDisabled: '#0b111a',
    textDisabled: 'rgba(199, 208, 221, 0.62)',
    canvasBackground: '#090f17',
    canvasGrid: '#263446',
    nodeSurface: '#111c29',
    edgeColor: '#778394',
    viewportBackground: '#05080d',
    timelineBackground: '#0a111b'
  }),
  light: Object.freeze({
    surfaceDisabled: '#ffffff',
    textDisabled: '#748196',
    canvasBackground: '#eef2f7',
    canvasGrid: '#c7d0dc',
    nodeSurface: '#ffffff',
    edgeColor: '#69788b',
    viewportBackground: '#e8edf4',
    timelineBackground: '#edf1f6'
  })
});

const basePalettes = Object.freeze({
  coral: Object.freeze({
    label: '珊瑚·暖墨',
    preview: Object.freeze({
      light: Object.freeze({ page: '#F6F4F1', surface: '#FFFFFF', primary: '#D95B52' }),
      dark: Object.freeze({ page: '#12100E', surface: '#1A1714', primary: '#F07167' })
    }),
    light: Object.freeze({
      background: '#F6F4F1',
      surface: '#FFFFFF',
      surface2: '#F1EEE9',
      surfaceHover: '#EAE5DE',
      surfaceActive: '#E1DBD2',
      surfaceElevated: '#FFFFFF',
      surfaceInput: '#FFFFFF',
      well: '#EEEBE6',
      border: '#D8D1C7',
      borderStrong: '#8C8478',
      textPrimary: '#1C1917',
      textSecondary: '#44403C',
      textMuted: '#6B645C',
      primary: '#D95B52',
      primaryHover: '#F07167',
      primaryActive: '#B8463F',
      primarySoft: '#FFF0ED',
      primaryBorder: '#FFCBC4',
      textOnAccent: '#2B0F0C',
      focusRing: '#B8463F',
      success: '#1F7A4D',
      successBg: '#E9F7EF',
      successBorder: '#B7DFC7',
      warning: '#8A5300',
      warningBg: '#FDF4E0',
      warningBorder: '#E8CE97',
      danger: '#B32633',
      dangerBg: '#FBE9EA',
      dangerBorder: '#EFB9BD',
      info: '#2F5FA8',
      infoBg: '#EAF0FB',
      infoBorder: '#BDD0EE'
    }),
    dark: Object.freeze({
      background: '#12100E',
      surface: '#1A1714',
      surface2: '#221E1A',
      surfaceHover: '#2A251F',
      surfaceActive: '#332C25',
      surfaceElevated: '#2A251F',
      surfaceInput: '#161310',
      well: '#0B0A08',
      border: '#4A4239',
      borderStrong: '#8A7C6A',
      textPrimary: '#F5F0E8',
      textSecondary: '#D6CEC2',
      textMuted: '#A79C8C',
      primary: '#F07167',
      primaryHover: '#F58A7D',
      primaryActive: '#D95B52',
      primarySoft: 'rgba(240, 113, 103, 0.16)',
      primaryBorder: 'rgba(240, 113, 103, 0.42)',
      textOnAccent: '#2B0F0C',
      focusRing: '#F07167',
      success: '#7BE0A8',
      successBg: 'rgba(123, 224, 168, 0.14)',
      successBorder: 'rgba(123, 224, 168, 0.36)',
      warning: '#F0BE63',
      warningBg: 'rgba(240, 190, 99, 0.16)',
      warningBorder: 'rgba(240, 190, 99, 0.38)',
      danger: '#FF7B85',
      dangerBg: 'rgba(255, 123, 133, 0.14)',
      dangerBorder: 'rgba(255, 123, 133, 0.36)',
      info: '#9DB4FF',
      infoBg: 'rgba(157, 180, 255, 0.14)',
      infoBorder: 'rgba(157, 180, 255, 0.36)'
    })
  }),
  ocean: Object.freeze({
    label: '深海青',
    preview: Object.freeze({
      light: Object.freeze({ page: '#F3F7F6', surface: '#FFFFFF', primary: '#0F766E' }),
      dark: Object.freeze({ page: '#07110F', surface: '#0C1917', primary: '#2DD4BF' })
    }),
    light: Object.freeze({
      background: '#F3F7F6',
      surface: '#FFFFFF',
      surface2: '#EDF4F2',
      surfaceHover: '#E3EEEC',
      surfaceActive: '#D8E8E4',
      surfaceElevated: '#FFFFFF',
      surfaceInput: '#FFFFFF',
      well: '#EAF0EF',
      border: '#CBDCD8',
      borderStrong: '#6F8F89',
      textPrimary: '#0F1F1C',
      textSecondary: '#334C47',
      textMuted: '#5A746F',
      primary: '#0F766E',
      primaryHover: '#0D9488',
      primaryActive: '#115E59',
      primarySoft: '#E6F5F2',
      primaryBorder: '#A9D9D1',
      textOnAccent: '#FFFFFF',
      focusRing: '#0F766E',
      success: '#15803D',
      successBg: '#E8F7EE',
      successBorder: '#B7DFC7',
      warning: '#8A5300',
      warningBg: '#FBF3DC',
      warningBorder: '#E8D08F',
      danger: '#B91C1C',
      dangerBg: '#FBE9EA',
      dangerBorder: '#EFB9BD',
      info: '#1D4ED8',
      infoBg: '#E8F0FE',
      infoBorder: '#B9CDF5'
    }),
    dark: Object.freeze({
      background: '#07110F',
      surface: '#0C1917',
      surface2: '#12221F',
      surfaceHover: '#182B27',
      surfaceActive: '#1F3631',
      surfaceElevated: '#182B27',
      surfaceInput: '#081412',
      well: '#040B0A',
      border: '#314641',
      borderStrong: '#5C7C76',
      textPrimary: '#EAF4F1',
      textSecondary: '#C2D6D1',
      textMuted: '#8FA9A3',
      primary: '#2DD4BF',
      primaryHover: '#5EEAD4',
      primaryActive: '#14B8A6',
      primarySoft: 'rgba(45, 212, 191, 0.14)',
      primaryBorder: 'rgba(45, 212, 191, 0.38)',
      textOnAccent: '#06201D',
      focusRing: '#2DD4BF',
      success: '#4ADE80',
      successBg: 'rgba(74, 222, 128, 0.14)',
      successBorder: 'rgba(74, 222, 128, 0.36)',
      warning: '#FACC15',
      warningBg: 'rgba(250, 204, 21, 0.16)',
      warningBorder: 'rgba(250, 204, 21, 0.38)',
      danger: '#FB7185',
      dangerBg: 'rgba(251, 113, 133, 0.14)',
      dangerBorder: 'rgba(251, 113, 133, 0.36)',
      info: '#93C5FD',
      infoBg: 'rgba(147, 197, 253, 0.14)',
      infoBorder: 'rgba(147, 197, 253, 0.36)'
    })
  }),
  graphite: Object.freeze({
    label: '石墨·电蓝',
    preview: Object.freeze({
      light: Object.freeze({ page: '#FAFAFA', surface: '#FFFFFF', primary: '#2F5BEA' }),
      dark: Object.freeze({ page: '#08090A', surface: '#101113', primary: '#6E93FF' })
    }),
    light: Object.freeze({
      background: '#FAFAFA',
      surface: '#FFFFFF',
      surface2: '#F4F4F5',
      surfaceHover: '#EDEDEF',
      surfaceActive: '#E4E4E7',
      surfaceElevated: '#FFFFFF',
      surfaceInput: '#FFFFFF',
      well: '#F1F1F2',
      border: '#E0E0E3',
      borderStrong: '#8A8A93',
      textPrimary: '#0A0A0B',
      textSecondary: '#3F3F46',
      textMuted: '#6B6B74',
      primary: '#2F5BEA',
      primaryHover: '#4F7BFF',
      primaryActive: '#2447C4',
      primarySoft: '#EAF0FF',
      primaryBorder: '#BFD2FF',
      textOnAccent: '#FFFFFF',
      focusRing: '#2F5BEA',
      success: '#16794C',
      successBg: '#E8F6EE',
      successBorder: '#B7DFC7',
      warning: '#8A5300',
      warningBg: '#FBF3DB',
      warningBorder: '#E6CD8A',
      danger: '#C42B34',
      dangerBg: '#FBE9EB',
      dangerBorder: '#EFB9BD',
      info: '#0E7490',
      infoBg: '#E7F5F8',
      infoBorder: '#B3DCE5'
    }),
    dark: Object.freeze({
      background: '#08090A',
      surface: '#101113',
      surface2: '#17181B',
      surfaceHover: '#1D1F23',
      surfaceActive: '#24262B',
      surfaceElevated: '#1D1F23',
      surfaceInput: '#0B0C0E',
      well: '#050506',
      border: '#2E3036',
      borderStrong: '#6B6F78',
      textPrimary: '#F7F8F8',
      textSecondary: '#B4BBC8',
      textMuted: '#8A8F98',
      primary: '#6E93FF',
      primaryHover: '#93B4FF',
      primaryActive: '#4F7BFF',
      primarySoft: 'rgba(110, 147, 255, 0.14)',
      primaryBorder: 'rgba(110, 147, 255, 0.38)',
      textOnAccent: '#05070F',
      focusRing: '#6E93FF',
      success: '#5EE0A0',
      successBg: 'rgba(94, 224, 160, 0.14)',
      successBorder: 'rgba(94, 224, 160, 0.36)',
      warning: '#F5C451',
      warningBg: 'rgba(245, 196, 81, 0.16)',
      warningBorder: 'rgba(245, 196, 81, 0.38)',
      danger: '#FF7B85',
      dangerBg: 'rgba(255, 123, 133, 0.14)',
      dangerBorder: 'rgba(255, 123, 133, 0.36)',
      info: '#67E8F9',
      infoBg: 'rgba(103, 232, 249, 0.14)',
      infoBorder: 'rgba(103, 232, 249, 0.36)'
    })
  }),
  violet: Object.freeze({
    label: '墨紫',
    preview: Object.freeze({
      light: Object.freeze({ page: '#F8F6FB', surface: '#FFFFFF', primary: '#6D4AE0' }),
      dark: Object.freeze({ page: '#0B0912', surface: '#141020', primary: '#A78BFA' })
    }),
    light: Object.freeze({
      background: '#F8F6FB',
      surface: '#FFFFFF',
      surface2: '#F2EEF8',
      surfaceHover: '#EAE4F4',
      surfaceActive: '#E0D8EF',
      surfaceElevated: '#FFFFFF',
      surfaceInput: '#FFFFFF',
      well: '#F0EBF6',
      border: '#DAD2E6',
      borderStrong: '#8B7FA0',
      textPrimary: '#1B1626',
      textSecondary: '#453A5A',
      textMuted: '#6B5F80',
      primary: '#6D4AE0',
      primaryHover: '#8B6CF0',
      primaryActive: '#5B3BC7',
      primarySoft: '#F2EDFF',
      primaryBorder: '#D5C6FF',
      textOnAccent: '#FFFFFF',
      focusRing: '#6D4AE0',
      success: '#167245',
      successBg: '#E9F6EE',
      successBorder: '#B7DFC7',
      warning: '#8A5300',
      warningBg: '#FBF3DC',
      warningBorder: '#E8D08F',
      danger: '#B32633',
      dangerBg: '#FBE9EA',
      dangerBorder: '#EFB9BD',
      info: '#3F55C5',
      infoBg: '#ECEFFE',
      infoBorder: '#C3CBF6'
    }),
    dark: Object.freeze({
      background: '#0B0912',
      surface: '#141020',
      surface2: '#1B162B',
      surfaceHover: '#241D36',
      surfaceActive: '#2E2544',
      surfaceElevated: '#241D36',
      surfaceInput: '#100C19',
      well: '#07050C',
      border: '#3A3050',
      borderStrong: '#756890',
      textPrimary: '#F3EFFA',
      textSecondary: '#CFC5E2',
      textMuted: '#9C90B5',
      primary: '#A78BFA',
      primaryHover: '#BCA4FF',
      primaryActive: '#8B6CF0',
      primarySoft: 'rgba(167, 139, 250, 0.14)',
      primaryBorder: 'rgba(167, 139, 250, 0.38)',
      textOnAccent: '#160F26',
      focusRing: '#A78BFA',
      success: '#6FE3AC',
      successBg: 'rgba(111, 227, 172, 0.14)',
      successBorder: 'rgba(111, 227, 172, 0.36)',
      warning: '#F5C268',
      warningBg: 'rgba(245, 194, 104, 0.16)',
      warningBorder: 'rgba(245, 194, 104, 0.38)',
      danger: '#FF7B85',
      dangerBg: 'rgba(255, 123, 133, 0.14)',
      dangerBorder: 'rgba(255, 123, 133, 0.36)',
      info: '#A8B2FF',
      infoBg: 'rgba(168, 178, 255, 0.14)',
      infoBorder: 'rgba(168, 178, 255, 0.36)'
    })
  })
});

const palettes = Object.freeze({
  ...basePalettes,
  glass: Object.freeze({
    label: '映窗·玻璃',
    preview: { light: { page: '#EAF0F7', surface: '#F6F9FD', primary: '#486B94' }, dark: { page: '#10131D', surface: '#1A2030', primary: '#B7CCF3' } },
    light: Object.freeze({ ...basePalettes.graphite.light, background: '#EAF0F7', surface: '#F6F9FD', surface2: '#E2E9F3', surfaceInput: '#F6F9FD', surfaceElevated: '#F6F9FD', surfaceHover: '#DEE7F4', surfaceActive: '#CEDCF0', textPrimary: '#192B42', textSecondary: '#3B526F', textMuted: '#526781', border: '#B4C2D5', borderStrong: '#748BA6', primary: '#486B94', primaryHover: '#597FA9', primaryActive: '#365471', textOnAccent: '#FFFFFF' }),
    dark: Object.freeze({ ...basePalettes.graphite.dark, background: '#10131D', surface: '#1A2030', surface2: '#222B3D', surfaceInput: '#141C2B', surfaceElevated: '#242D40', surfaceHover: '#2D3950', surfaceActive: '#35445E', textPrimary: '#FFFFFF', textSecondary: '#F5F7FB', textMuted: '#E4EAF3', border: '#45516B', borderStrong: '#7C8DAB', primary: '#B7CCF3', primaryHover: '#D6E3FC', primaryActive: '#92B1E5', textOnAccent: '#111D30' })
  })
});

export const THEME_MODES = Object.freeze([
  Object.freeze({ id: 'light', label: '浅色' }),
  Object.freeze({ id: 'dark', label: '深色' }),
  Object.freeze({ id: 'system', label: '跟随系统' })
]);

const modeIds = new Set(THEME_MODES.map(mode => mode.id));
const schemeIds = new Set(Object.keys(palettes));

// 历史 themeId（洛水下线的多套色板）按「原本的深浅倾向」映射到珊瑚方案的对应档位；
// 无法识别的值走跟随系统，保证升级后既不报错也不会突然变亮/变暗。
const legacyModeById = Object.freeze({
  light: 'light',
  dark: 'dark',
  system: 'system',
  pureblack: 'dark',
  purple: 'dark',
  graywhite: 'light',
  mint: 'light',
  snow: 'light',
  vanilla: 'light',
  sakura: 'light',
  willow: 'light',
  ivory: 'light',
  lotus: 'light'
});

function normalizeToken(value) {
  return typeof value === 'string' ? value.trim().toLowerCase() : '';
}

export function normalizeSchemeId(value) {
  const schemeId = normalizeToken(value);
  return schemeIds.has(schemeId) ? schemeId : DEFAULT_SCHEME_ID;
}

export function normalizeThemeMode(value) {
  const modeId = normalizeToken(value);
  return modeIds.has(modeId) ? modeId : 'system';
}

export function composeThemeId(schemeId, modeId) {
  return `${normalizeSchemeId(schemeId)}-${normalizeThemeMode(modeId)}`;
}

// 解析任意历史值：'dark' / 'light' / 'system' / 'mint' / 'ocean-dark' 全部收敛成 { scheme, mode }。
export function parseThemeId(value) {
  const raw = normalizeToken(value);
  if (legacyModeById[raw]) {
    return { scheme: DEFAULT_SCHEME_ID, mode: legacyModeById[raw], id: `${DEFAULT_SCHEME_ID}-${legacyModeById[raw]}` };
  }
  if (schemeIds.has(raw)) {
    return { scheme: raw, mode: 'system', id: `${raw}-system` };
  }
  const [schemePart, modePart] = raw.split('-');
  if (schemeIds.has(schemePart) && modeIds.has(modePart)) {
    return { scheme: schemePart, mode: modePart, id: `${schemePart}-${modePart}` };
  }
  return { scheme: DEFAULT_SCHEME_ID, mode: 'system', id: `${DEFAULT_SCHEME_ID}-system` };
}

export function normalizeThemeId(value) {
  return parseThemeId(value).id;
}

export function resolveSchemeId(value) {
  return parseThemeId(value).scheme;
}

// 保留原签名：返回 'light' | 'dark'（真正落到 html[data-theme] 的值）。
export function resolveThemeId(value, systemMode = 'dark') {
  const { mode } = parseThemeId(value);
  if (mode !== 'system') return mode;
  return systemMode === 'light' ? 'light' : 'dark';
}

export function listThemeSchemes() {
  return ['glass'].map(id => ({
    id,
    label: palettes[id].label,
    preview: {
      light: { ...palettes[id].preview.light },
      dark: { ...palettes[id].preview.dark }
    }
  }));
}

export function listThemeModes() {
  return THEME_MODES.map(mode => ({ ...mode }));
}

export function schemeDefinition(value) {
  const schemeId = normalizeSchemeId(value);
  return { id: schemeId, ...palettes[schemeId], preview: { ...palettes[schemeId].preview } };
}

export function themeDefinition(value, systemMode = 'dark') {
  const { scheme, mode, id } = parseThemeId(value);
  const colorScheme = mode === 'system' ? (systemMode === 'light' ? 'light' : 'dark') : mode;
  return {
    id,
    scheme,
    label: palettes[scheme].label,
    colorScheme,
    tokens: { ...modeBase[colorScheme], ...palettes[scheme][colorScheme] }
  };
}
