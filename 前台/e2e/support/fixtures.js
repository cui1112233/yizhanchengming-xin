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

export function makeVideoStatus({
  status = 'succeeded',
  provider = 'personal_api',
  model = 'yd2.0-mini',
  errorMessage = '',
  outputURL = 'https://example.invalid/task16/video.mp4',
  attemptID = 701,
  attempts = 1,
} = {}) {
  const values = []
  for (let index = 1; index <= attempts; index += 1) {
    const isLatest = index === attempts
    values.push({
      id: isLatest ? attemptID : attemptID - (attempts - index),
      attempt: index,
      status: isLatest ? status : 'failed',
      errorCode: isLatest && errorMessage ? 'provider_request_failed' : '',
      errorMessage: isLatest ? errorMessage : 'previous fixture failure',
      outputUrl: isLatest ? outputURL : '',
    })
  }
  return {
    batchProjectId: 123,
    books: [{
      bookId: 1001,
      jobId: 801,
      provider,
      model,
      status,
      attempts: values,
      errorMessage,
      outputUrl: outputURL,
    }],
  }
}

export async function routeBatchProjectFixtures(page, { generationSummary, audioDurationMs = 28000, videoStatus } = {}) {
  await page.route('**/api/v1/batch-projects', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [projectFixture] }) })
  })
  await page.route('**/api/v1/batch-projects/123', async (route) => {
    if (new URL(route.request().url()).pathname !== '/api/v1/batch-projects/123') return route.fallback()
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ project: projectFixture, books: projectBooksFixture }) })
  })
  await page.route('**/api/v1/batch-projects/123/settings', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(settingsFixture()) })
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
  await page.route('**/api/v1/batch-projects/123/video', async (route) => {
    const payload = typeof videoStatus === 'function'
      ? videoStatus()
      : (videoStatus || { batchProjectId: 123, books: [] })
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(payload) })
  })

  if (generationSummary) {
    await page.route('**/api/v1/batch-projects/123/generation/runs/*', async (route) => {
      const runId = Number(new URL(route.request().url()).pathname.split('/').at(-1))
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ runId, batchProjectId: 123, status: 'completed', terminal: true, counts: { total: 1, pending: 0, running: 0, completed: 1, failed: 0 }, tasks: [] }) })
    })
    await page.route('**/api/v1/batch-projects/123/generation', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(generationSummary()) })
        return
      }
      await route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ runId: 900, taskIds: [501], status: 'queued', dispatch: 'queued', pollUrl: '/api/v1/batch-projects/123/generation/runs/900' }) })
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
        run: { id: 501, runId: 900, status: 'completed' },
        stages: {
          SCRIPT: { status: 'completed', bookRunId: 501, outputText: 'script' },
          HOOK: { status: 'completed', bookRunId: 501, outputText: 'hook' },
          DIRECTOR: { status: 'completed', bookRunId: 501, outputText: 'director' },
          FINAL_PROMPT: { status: 'completed', bookRunId: 501, outputText: 'final' },
        },
      },
    ],
  }
  return { ...base, ...overrides }
}
