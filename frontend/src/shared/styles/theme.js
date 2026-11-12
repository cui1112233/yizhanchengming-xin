import { theme as antdTheme } from 'antd';
import {
  composeThemeId,
  listThemeModes,
  listThemeSchemes,
  normalizeThemeId,
  parseThemeId,
  resolveSchemeId,
  resolveThemeId,
  schemeDefinition,
  themeDefinition
} from './theme-catalog.js';

export {
  composeThemeId,
  listThemeModes,
  listThemeSchemes,
  normalizeThemeId,
  parseThemeId,
  resolveSchemeId,
  resolveThemeId,
  schemeDefinition,
  themeDefinition
};

const fontFamily = '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif';

export function createAntTheme(mode, systemMode = 'dark') {
  const { tokens } = themeDefinition(mode, systemMode);
  const isDark = resolveThemeId(mode, systemMode) === 'dark';
  // 浮层派生色：统一从语义令牌取，避免 antd 用浅色算法推导深色组件。
  const overlaySurface = tokens.surfaceElevated;
  const subtleSurface = tokens.surface2;
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
      // 主色实心上的文字由风格决定（珊瑚用深墨、青/蓝/紫用白），不再一律用近白字。
      colorTextLightSolid: tokens.textOnAccent,
      colorBorder: tokens.border,
      colorBorderSecondary: tokens.borderStrong,
      colorSplit: tokens.border,
      colorError: tokens.danger,
      colorLink: tokens.info,
      colorSuccess: tokens.success,
      colorWarning: tokens.warning,
      controlItemBgHover: tokens.surfaceHover,
      controlItemBgActive: tokens.surfaceActive,
      controlItemBgActiveHover: tokens.surfaceActive,
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
        optionActiveBg: tokens.surfaceHover,
        optionSelectedBg: tokens.surfaceActive
      },
      Table: {
        headerBg: subtleSurface,
        headerColor: tokens.textSecondary,
        rowHoverBg: tokens.surfaceHover,
        rowSelectedBg: tokens.surfaceActive,
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

// V88 页面只补充自己的圆角与控件高度；颜色、深浅算法、组件皮肤全部继承外层
// ConfigProvider（UserLayout 的 createAntTheme）。风格切换因此不需要逐页改 JSX，
// 也不会因为页面 useMemo 依赖没变而拿到过期色板。
export function createV88AntTheme() {
  return {
    token: {
      borderRadius: 9,
      controlHeight: 36
    }
  };
}
