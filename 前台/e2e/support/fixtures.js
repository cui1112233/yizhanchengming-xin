export const defaultUser = { id: 7, username: 'e2e-user', displayName: 'E2E User' }

export async function routeAuthenticated(page, user = defaultUser) {
  await page.route('**/api/auth/current-user', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user }) })
  })
}

export async function routeAnonymous(page, { refreshStatus = 401 } = {}) {
  await page.route('**/api/auth/current-user', async (route) => {
    await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'UNAUTHENTICATED', message: 'authentication required' }) })
  })
  await page.route('**/api/auth/refresh', async (route) => {
    await route.fulfill({ status: refreshStatus, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_MISSING', message: 'no refresh session' }) })
  })
}

export const projectFixture = {
  id: 123,
  name: 'Task16 Fixture Project',
  sources: ['知乎', '常读'],
  bookCount: 2,
  genders: ['女频', '男频'],
  styles: ['现实情感', '都市'],
  runStatus: 'pending',
}

export const projectBooksFixture = [
  { id: 1, bookId: '1001', title: '林子深处有声音', source: '知乎', platformId: '6', gender: '女频', style: '现实情感', status: 'fetched', errorMessage: '' },
  { id: 2, bookId: '2001', title: '第二本测试小说', source: '常读', platformId: '2', gender: '男频', style: '都市', status: 'retryable_failed', errorMessage: 'fixture upstream failure' },
]

export function settingsFixture() {
  return {
    project: {
      production: { productionMode: 'viral', aiCopyEnabled: true, aiCopyCount: 2 },
      publishing: { uploadVideoType: 'merged', materialReuse: false, versionProfile: 'task16-fixture' },
    },
    profile: {
      name: 'Task16 Fixture Profile',
      version: 'v1',
      settings: {
        processingRulePromptRef: 'processing-rule-v3',
        knowledgePromptRef: 'knowledge-v7',
        styleTypes: { styles: ['现实情感'], genres: ['都市'], genders: ['女频'] },
      },
    },
  }
}

export async function routeBatchProjectFixtures(page, { generationSummary, audioDurationMs = 28000 } = {}) {
  await page.route('**/api/v1/batch-projects', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [projectFixture] }) })
  })
  await page.route('**/api/v1/batch-projects/123', async (route) => {
    if (new URL(route.request().url()).pathname !== '/api/v1/batch-projects/123') return route.fallback()
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ project: projectFixture, books: projectBooksFixture }) })
  })
  await page.route('**/api/v1/batch-projects/123/settings', async (route) => {
    const payload = settingsFixture()
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(payload) })
  })
  await page.route('**/api/v1/batch-projects/123/settings/production', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
  })
  await page.route('**/api/v1/batch-projects/123/settings/publishing', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
  })
  await page.route('**/api/v1/batch-projects/123/version-profile', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
  })
  await page.route('**/api/v1/batch-projects/123/version-profile/sync-121', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
  })
  await page.route('**/api/v1/batch-projects/123/version-profile/sync-style-types', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
  })

  if (generationSummary) {
    await page.route('**/api/v1/batch-projects/123/generation', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(generationSummary()) })
        return
      }
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ status: 'accepted' }) })
    })
    await page.route('**/api/v1/batch-projects/123/books/*/audio-measurement', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ durationMs: audioDurationMs, source: 'fixture.wav' }) })
    })
  }
}

export function makeGenerationSummary(overrides = {}) {
  const base = {
    completed: 1,
    running: 0,
    failed: 0,
    pending: 0,
    books: [
      {
        bookId: '1001',
        title: '林子深处有声音',
        stages: {
          SCRIPT: { status: 'completed', outputText: 'script' },
          HOOK: { status: 'completed', outputText: 'hook' },
          DIRECTOR: { status: 'completed', outputText: 'director' },
          FINAL_PROMPT: { status: 'completed', outputText: 'final' },
        },
      },
    ],
  }
  return { ...base, ...overrides }
}
