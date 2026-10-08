import { theme as antdTheme } from 'antd'
import { themeTokens, sharedComponents } from '../../ui/theme-contract.js'

export {
  typography, baseToken, semanticColors, shellMetrics, controlRoles, radiusRoles,
  sharedComponents, themeTokens, themeCssVariables, applyThemeVariables,
} from '../../ui/theme-contract.js'

export function themeConfig(mode = 'light') {
  return {
    algorithm: mode === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: themeTokens(mode),
    components: Object.fromEntries(Object.entries(sharedComponents)
      .map(([key, value]) => [key, { ...value }])),
  }
}
