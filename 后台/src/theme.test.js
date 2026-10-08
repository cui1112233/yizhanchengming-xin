import { describe, expect, it, vi } from 'vitest'
import { theme as antdTheme } from 'antd'
import * as local from './theme.js'
import * as other from '../../前台/src/theme.js'
import { registerThemeContractTests } from '../../ui/theme-contract-test-suite.js'

registerThemeContractTests({
  local, other, describe, expect, it, vi,
  algorithms: { light: antdTheme.defaultAlgorithm, dark: antdTheme.darkAlgorithm },
})
