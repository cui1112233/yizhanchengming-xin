import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    setupFiles: ['./src/testSetup.js'],
    include: ['src/**/*.{test,spec}.{js,jsx,ts,tsx}'],
  },
})
