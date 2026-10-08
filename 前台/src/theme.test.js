import { describe, expect, it, vi } from 'vitest'
import { theme as antdTheme } from 'antd'
import * as local from './theme.js'
import * as other from '../../后台/src/theme.js'

const font = '-apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", sans-serif'
const colors = {
  dark: { bg: '#070b11', panel: 'rgba(11,17,26,.96)', card: 'rgba(14,22,33,.96)', cardHover: 'rgba(20,31,45,.98)', input: '#0b111a', text: '#f5f7fb', secondary: '#8391a4', dim: '#586679', border: 'rgba(151,168,190,.12)', borderStrong: 'rgba(176,195,219,.2)', shadowHover: '0 8px 32px rgba(0,0,0,.22)' },
  light: { bg: '#f4f6fa', panel: 'rgba(255,255,255,.9)', card: '#fff', cardHover: '#f5edf0', input: '#fff', text: '#202330', secondary: '#667085', dim: '#98a2b3', border: '#e4e7ee', borderStrong: '#cfd5e1', shadowHover: '0 8px 24px rgba(31,41,55,.08)' },
}
const components = { Card: { headerHeight: 52 }, Drawer: { footerPaddingBlock: 16, footerPaddingInline: 24 }, Table: { cellPaddingBlock: 14 } }
const sharedNames = ['typography', 'baseToken', 'semanticColors', 'shellMetrics', 'controlRoles', 'radiusRoles', 'sharedComponents']

function expectFrozen(record) {
  expect(Object.isFrozen(record)).toBe(true)
  for (const value of Object.values(record)) if (value && typeof value === 'object') expectFrozen(value)
}

describe('shared public theme contract', () => {
  it('shares recursively immutable audited primitives across apps', () => {
    for (const name of sharedNames) {
      expect(local[name]).toBe(other[name])
      expectFrozen(local[name])
    }
    expect(local.typography).toEqual({ fontFamily: font, fontSize: 14 })
    expect(local.typography.fontFamily).not.toContain('Inter')
    expect(local.semanticColors).toEqual(colors)
    expect(local.shellMetrics).toEqual({ sidebarExpanded: 220, sidebarCollapsed: 60, sidebarRail: 64, topbar: 56, gutter: 24, contentMax: 1480, railBreakpoint: 900 })
    expect(local.controlRoles).toEqual({ small: 24, default: 32, homePill: 42 })
    expect(local.radiusRoles).toEqual({ form: 8, nav: 10, card: 14 })
    expect(local.sharedComponents).toEqual(components)
  })

  it.each(['light', 'dark'])('maps %s semantics into identical app configurations', (mode) => {
    const c = colors[mode]
    const expected = {
      colorPrimary: '#f07167', colorPrimaryHover: '#f08a7d', borderRadius: 8,
      controlHeight: 32, controlHeightSM: 24, fontSize: 14, fontFamily: font,
      colorBgLayout: c.bg, colorBgContainer: c.input, colorBgElevated: c.card,
      colorText: c.text, colorTextSecondary: c.secondary, colorTextTertiary: c.dim,
      colorBorder: c.border, colorBorderSecondary: c.borderStrong,
    }
    expect(local.themeTokens(mode)).toEqual(expected)
    expect(other.themeTokens(mode)).toEqual(expected)
    expect(local.themeConfig(mode).token).toEqual(expected)
    expect(other.themeConfig(mode).token).toEqual(expected)
    expect(local.themeConfig(mode).components).toEqual(components)
    expect(other.themeConfig(mode).components).toEqual(components)
    expect(local.themeConfig(mode).algorithm).toBe(mode === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm)
    expect(local.themeConfig(mode).token).not.toHaveProperty('controlHeightLG')
  })

  it('normalizes unknown modes to light and isolates mutable framework consumers', () => {
    expect(local.themeConfig('unexpected')).toEqual(local.themeConfig('light'))
    expect(local.themeTokens()).toEqual(local.themeTokens('light'))
    expect(local.themeCssVariables('unexpected')).toEqual(local.themeCssVariables('light'))
    const first = local.themeConfig('dark')
    const next = local.themeConfig('dark')
    expect(first).not.toBe(next)
    expect(first.token).not.toBe(next.token)
    expect(first.components).not.toBe(next.components)
    for (const name of Object.keys(components)) expect(first.components[name]).not.toBe(next.components[name])
    first.components.Card.headerHeight = 1
    first.token.colorPrimary = 'invalid'
    expect(local.themeConfig('dark').components.Card.headerHeight).toBe(52)
    expect(other.themeConfig('dark').components.Card.headerHeight).toBe(52)
    expect(local.sharedComponents.Card.headerHeight).toBe(52)
    expect(local.themeTokens('dark').colorPrimary).toBe('#f07167')
  })

  it.each(['light', 'dark'])('installs the complete %s CSS inheritance contract without preferences', (mode) => {
    const c = colors[mode]
    const expected = {
      '--shell-bg': c.bg, '--shell-panel': c.panel, '--shell-card': c.card,
      '--shell-card-hover': c.cardHover, '--shell-input': c.input, '--shell-text': c.text,
      '--shell-secondary': c.secondary, '--shell-muted': c.secondary, '--shell-dim': c.dim,
      '--shell-border': c.border, '--shell-border-strong': c.borderStrong,
      '--shell-shadow-hover': c.shadowHover, '--shell-shadow': c.shadowHover,
      '--shell-accent': '#f07167', '--shell-accent-hover': '#f08a7d', '--shell-accent-tint': 'rgba(240,113,103,.14)',
      '--app-font-family': font, '--app-sidebar-expanded': '220px', '--app-sidebar-collapsed': '60px',
      '--app-sidebar-rail': '64px', '--app-topbar': '56px', '--app-gutter': '24px', '--app-content-max': '1480px',
      '--app-control-small': '24px', '--app-control-default': '32px', '--app-control-home-pill': '42px',
      '--app-radius-form': '8px', '--app-radius-nav': '10px', '--app-radius-card': '14px',
    }
    expect(local.themeCssVariables(mode)).toEqual(expected)
    expect(other.themeCssVariables(mode)).toEqual(expected)
    expect(local.themeCssVariables(mode)).not.toBe(local.themeCssVariables(mode))
    const setProperty = vi.fn()
    const target = { style: { setProperty }, dataset: Object.freeze({}) }
    const request = vi.spyOn(globalThis, 'fetch')
    local.applyThemeVariables(target, mode)
    expect(setProperty.mock.calls).toHaveLength(Object.keys(expected).length)
    expect(Object.fromEntries(setProperty.mock.calls)).toEqual(expected)
    expect(target.dataset).toEqual({})
    expect(Object.keys(target)).toEqual(['style', 'dataset'])
    expect(request).not.toHaveBeenCalled()
    request.mockRestore()
    expect(() => local.applyThemeVariables(null)).not.toThrow()
    expect(() => local.applyThemeVariables({ style: {} })).not.toThrow()
  })
})
