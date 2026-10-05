import { defineConfig } from '@playwright/test'

const publicBaseURL = process.env.PUBLIC_BASE_URL?.trim()
const localBaseURL = 'http://127.0.0.1:4173'
const publicSmoke = Boolean(publicBaseURL)

export default defineConfig({
  testDir: './e2e',
  outputDir: 'test-results',
  timeout: 45_000,
  expect: { timeout: 8_000 },
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: publicSmoke
    ? [['line']]
    : (process.env.CI
      ? [['line'], ['html', { outputFolder: 'playwright-report', open: 'never' }]]
      : [['list'], ['html', { outputFolder: 'playwright-report', open: 'never' }]]),
  use: {
    baseURL: publicBaseURL || localBaseURL,
    // Public smoke uses real account credentials and session cookies. Do not
    // retain browser artifacts that can serialize typed secrets or headers.
    trace: publicSmoke ? 'off' : 'retain-on-failure',
    screenshot: publicSmoke ? 'off' : 'only-on-failure',
    video: publicSmoke ? 'off' : 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 20_000,
  },
  webServer: publicBaseURL
    ? undefined
    : {
        command: 'npm run dev -- --port 4173',
        url: localBaseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      },
})
