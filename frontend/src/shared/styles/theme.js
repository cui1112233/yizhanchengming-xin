import { theme as antdTheme } from 'antd';
import { listThemeDefinitions, normalizeThemeId, resolveThemeId, themeDefinition } from './theme-catalog.js';

export { listThemeDefinitions, normalizeThemeId, resolveThemeId, themeDefinition };

const fontFamily = '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif';

export function createAntTheme(mode, systemMode = 'dark') {
  const { tokens } = themeDefinition(mode, systemMode);
  const isDark = resolveThemeId(mode, systemMode) === 'dark';
  // 浮层派生色：统一从语义令牌取，避免 antd 用浅色算法推导深色组件。
  const overlaySurface = tokens.surfaceElevated;
  const subtleSurface = isDark ? tokens.surfaceElevated : tokens.background;
  const overlayShadow = isDark
    ? '0 12px 32px rgba(0, 0, 0, 0.55)'
    : '0 12px 32px rgba(23, 32, 51, 0.14)';
  return {
    // 关键：深浅色算法交给 antd，原生组件的 hover/禁用/派生色不再按单一色板反推。
    algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: tokens.primary,
      colorPrimaryHover: tokens.primaryHover,
      colorPrimaryActive: tokens.primaryActive,
      colorText: tokens.textPrimary,
      colorTextSecondary: tokens.textSecondary,
      colorTextTertiary: tokens.textMuted,
      colorTextPlaceholder: tokens.textMuted,
      colorTextDisabled: tokens.textDisabled,
      colorBgBase: tokens.background,
      colorBgContainer: tokens.surface,
      colorBgElevated: tokens.surfaceElevated,
      colorBgLayout: tokens.background,
      colorBgContainerDisabled: tokens.surfaceDisabled,
      colorBgSpotlight: isDark ? tokens.surfaceElevated : tokens.textPrimary,
      colorTextLightSolid: isDark ? tokens.textPrimary : tokens.surface,
      colorBorder: tokens.border,
      colorBorderSecondary: tokens.borderStrong,
      colorSplit: tokens.border,
      colorError: tokens.danger,
      colorLink: tokens.info,
      colorSuccess: tokens.success,
      colorWarning: tokens.warning,
      controlItemBgHover: subtleSurface,
      controlItemBgActive: subtleSurface,
      controlItemBgActiveHover: subtleSurface,
      boxShadow: overlayShadow,
      boxShadowSecondary: overlayShadow,
      boxShadowTertiary: overlayShadow,
      borderRadius: 8,
      fontFamily
    },
    // 浮层类组件：底色、阴影走同一套语义色，深色下不再出现浅色面板。
    components: {
      Modal: {
        contentBg: overlaySurface,
        headerBg: overlaySurface,
        titleColor: tokens.textPrimary,
        footerBg: overlaySurface
      },
      Popover: {
        popoverBg: overlaySurface,
        popoverColor: tokens.textPrimary
      },
      Select: {
        selectorBg: tokens.surface,
        optionActiveBg: subtleSurface,
        optionSelectedBg: tokens.nodeSurface
      },
      Table: {
        headerBg: subtleSurface,
        headerColor: tokens.textSecondary,
        rowHoverBg: subtleSurface,
        borderColor: tokens.border,
        headerSplitColor: tokens.border
      },
      Tabs: {
        itemColor: tokens.textSecondary,
        itemHoverColor: tokens.textPrimary,
        itemSelectedColor: tokens.primary,
        inkBarColor: tokens.primary
      },
      Segmented: {
        trackBg: tokens.background,
        itemSelectedBg: isDark ? tokens.surfaceElevated : tokens.surface,
        itemColor: tokens.textSecondary,
        itemSelectedColor: tokens.textPrimary,
        itemHoverBg: subtleSurface,
        itemHoverColor: tokens.textPrimary
      },
      Pagination: {
        itemBg: 'transparent',
        itemLinkBg: 'transparent',
        itemActiveBg: tokens.surfaceElevated,
        itemActiveBgDisabled: tokens.surfaceDisabled
      }
    }
  };
}

// V88 视觉统一：主色改为青绿，珊瑚降级为强调/警示，靛蓝只给信息与链接。
// 只在启用页面（第一批 /tts）通过 ConfigProvider 局部生效，不改变全站默认主题。
export const v88Palette = {
  primary: '#63dfc0',
  primaryHover: '#7de8cf',
  primaryActive: '#4fcbb0',
  emphasis: '#f07167',
  danger: '#ff9b9d',
  success: '#75ddb4',
  warning: '#ffd08a',
  info: '#9da8ff'
};

export function createV88AntTheme(mode = 'dark') {
  const base = createAntTheme(mode);
  return {
    ...base,
    token: {
      ...base.token,
      borderRadius: 9,
      controlHeight: 36
    }
  };
}
