import { listThemeDefinitions, normalizeThemeId, resolveThemeId, themeDefinition } from './theme-catalog.js';

export { listThemeDefinitions, normalizeThemeId, resolveThemeId, themeDefinition };

const fontFamily = '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif';

export function createAntTheme(mode, systemMode = 'dark') {
  const { tokens } = themeDefinition(mode, systemMode);
  return {
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
      colorBorder: tokens.border,
      colorBorderSecondary: tokens.borderStrong,
      colorError: tokens.danger,
      colorLink: tokens.info,
      colorSuccess: tokens.success,
      colorWarning: tokens.warning,
      borderRadius: 8,
      fontFamily
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
