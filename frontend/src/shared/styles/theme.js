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
