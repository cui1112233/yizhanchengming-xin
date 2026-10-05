import { configDefaults, defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    setupFiles: ['./src/testSetup.js'],
    exclude: [...configDefaults.exclude, 'e2e/**'],
  },
})
