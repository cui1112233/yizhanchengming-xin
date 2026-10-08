// Framework-neutral public shell primitives. AntD algorithms belong to each app.
function freezeRecord(record) {
  for (const value of Object.values(record)) {
    if (value && typeof value === 'object') freezeRecord(value)
  }
  return Object.freeze(record)
}

export const typography = freezeRecord({
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif',
  fontSize: 14,
})
export const shellMetrics = freezeRecord({
  sidebarExpanded: 220, sidebarCollapsed: 60, sidebarRail: 64,
  topbar: 56, gutter: 24, contentMax: 1480, railBreakpoint: 900,
})
export const controlRoles = freezeRecord({ small: 24, default: 32, homePill: 42 })
export const radiusRoles = freezeRecord({ form: 8, nav: 10, card: 14 })
export const semanticColors = freezeRecord({
  dark: {
    bg: '#070b11', panel: 'rgba(11,17,26,.96)', card: 'rgba(14,22,33,.96)',
    cardHover: 'rgba(20,31,45,.98)', input: '#0b111a', text: '#f5f7fb',
    secondary: '#8391a4', dim: '#586679', border: 'rgba(151,168,190,.12)',
    borderStrong: 'rgba(176,195,219,.2)', shadowHover: '0 8px 32px rgba(0,0,0,.22)',
  },
  light: {
    bg: '#f4f6fa', panel: 'rgba(255,255,255,.9)', card: '#fff', cardHover: '#f5edf0',
    input: '#fff', text: '#202330', secondary: '#667085', dim: '#98a2b3',
    border: '#e4e7ee', borderStrong: '#cfd5e1', shadowHover: '0 8px 24px rgba(31,41,55,.08)',
  },
})
export const baseToken = freezeRecord({
  colorPrimary: '#f07167', colorPrimaryHover: '#f08a7d',
  borderRadius: radiusRoles.form, controlHeight: controlRoles.default,
  controlHeightSM: controlRoles.small, ...typography,
})
// Existing local compatibility values; these were not measured in public admin.
export const sharedComponents = freezeRecord({
  Card: { headerHeight: 52 },
  Drawer: { footerPaddingBlock: 16, footerPaddingInline: 24 },
  Table: { cellPaddingBlock: 14 },
})

function colorsFor(mode) {
  return semanticColors[mode === 'dark' ? 'dark' : 'light']
}

export function themeTokens(mode = 'light') {
  const colors = colorsFor(mode)
  return {
    ...baseToken,
    colorBgLayout: colors.bg, colorBgContainer: colors.input, colorBgElevated: colors.card,
    colorText: colors.text, colorTextSecondary: colors.secondary, colorTextTertiary: colors.dim,
    colorBorder: colors.border, colorBorderSecondary: colors.borderStrong,
  }
}

const kebab = (key) => key.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)

export function themeCssVariables(mode = 'light') {
  const colors = colorsFor(mode)
  return {
    ...Object.fromEntries(Object.entries(colors).map(([key, value]) => [`--shell-${kebab(key)}`, value])),
    '--shell-muted': colors.secondary,
    '--shell-shadow': colors.shadowHover,
    '--shell-accent': baseToken.colorPrimary,
    '--shell-accent-hover': baseToken.colorPrimaryHover,
    '--shell-accent-tint': 'rgba(240,113,103,.14)',
    '--app-font-family': typography.fontFamily,
    ...Object.fromEntries(Object.entries(shellMetrics)
      .filter(([key]) => key !== 'railBreakpoint')
      .map(([key, value]) => [`--app-${kebab(key)}`, `${value}px`])),
    ...Object.fromEntries(Object.entries(controlRoles).map(([key, value]) => [`--app-control-${kebab(key)}`, `${value}px`])),
    ...Object.fromEntries(Object.entries(radiusRoles).map(([key, value]) => [`--app-radius-${key}`, `${value}px`])),
  }
}

export function applyThemeVariables(target, mode = 'light') {
  if (typeof target?.style?.setProperty !== 'function') return
  for (const [key, value] of Object.entries(themeCssVariables(mode))) target.style.setProperty(key, value)
}
