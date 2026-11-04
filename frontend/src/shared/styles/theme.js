const common = {
  colorPrimary: '#f07167',
  borderRadius: 8,
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif'
};

const dark = {
  colorText: '#f5f7fb',
  colorTextSecondary: '#c7d0dd',
  colorTextTertiary: '#a7b4c5',
  colorTextPlaceholder: '#a7b4c5',
  colorTextDisabled: 'rgba(199, 208, 221, 0.62)',
  colorBgBase: '#070b11',
  colorBgContainer: '#0e1621',
  colorBgElevated: '#0e1621',
  colorBgLayout: '#070b11',
  colorBgContainerDisabled: '#0b111a',
  colorBorder: 'rgba(151, 168, 190, 0.22)',
  colorBorderSecondary: 'rgba(176, 195, 219, 0.36)'
};

const light = {
  colorText: '#172033',
  colorTextSecondary: '#3f4e63',
  colorTextTertiary: '#596a80',
  colorTextPlaceholder: '#596a80',
  colorTextDisabled: '#748196',
  colorBgBase: '#f4f6fa',
  colorBgContainer: '#ffffff',
  colorBgElevated: '#ffffff',
  colorBgLayout: '#f4f6fa',
  colorBgContainerDisabled: '#ffffff',
  colorBorder: '#c8d2df',
  colorBorderSecondary: '#9daabd'
};

export function createAntTheme(mode) {
  return { token: { ...common, ...(mode === 'light' ? light : dark) } };
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
      colorPrimary: v88Palette.primary,
      colorPrimaryHover: v88Palette.primaryHover,
      colorPrimaryActive: v88Palette.primaryActive,
      colorError: v88Palette.danger,
      colorLink: v88Palette.info,
      colorSuccess: v88Palette.success,
      colorWarning: v88Palette.warning,
      borderRadius: 9,
      controlHeight: 36
    }
  };
}
