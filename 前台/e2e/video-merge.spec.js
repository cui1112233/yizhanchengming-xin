import { expect, test } from './support/acceptance.js'
import { makeGenerationSummary, makeVideoStatus, routeAuthenticated, routeBatchProjectFixtures } from './support/fixtures.js'

async function appRequest(page, path, options = {}) {
  return page.evaluate(async ({ requestPath, requestOptions }) => {
    const api = await import('/src/api.js')
    try {
      const payload = await api.requestJSON(requestPath, requestOptions)
      return { ok: true, payload }
    } catch (error) {
      return { ok: false, status: error?.status || 0, code: error?.code || '', message: error?.message || '' }
    }
  }, { requestPath: path, requestOptions: options })
}

test.describe('@video Task 14 VIDEO / Merge acceptance', () => {
  test.beforeEach(async ({ page }) => {
    await routeAuthenticated(page)
  })

  test('@video provider status supports unconfigured/available/unavailable/auth_failed and exposes no secret values', async ({ page }) => {
    const statuses = ['unconfigured', 'available', 'unavailable', 'auth_failed']
    let index = 0
    await page.route('**/api/v1/video-providers/personal_api/models/yd2.0-mini/status', async (route) => {
      const status = statuses[index++]
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ providerKey: 'personal_api', model: 'yd2.0-mini', configured: status !== 'unconfigured', enabled: true, status, message: `fixture ${status}` }),
      })
    })
    await page.goto('/batch-factory')

    for (const expectedStatus of statuses) {
      const result = await appRequest(page, '/api/v1/video-providers/personal_api/models/yd2.0-mini/status')
      expect(result.ok).toBe(true)
      expect(result.payload.status).toBe(expectedStatus)
      const serialized = JSON.stringify(result.payload).toLowerCase()
      expect(serialized).not.toMatch(/access[_-]?token|refresh[_-]?token|api[_-]?secret|credential_ciphertext|aes[_-]?nonce|encryptedsecret|secretnonce|authorization|password/)
    }
  })

  test('@video project VIDEO UI shows provider/model/status/attempts/error', async ({ page }) => {
    await routeBatchProjectFixtures(page, {
      generationSummary: () => makeGenerationSummary(),
      videoStatus: () => makeVideoStatus({ status: 'failed', errorMessage: 'provider fixture failure', outputURL: '', attempts: 2 }),
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()

    const table = page.locator('.ant-table-tbody')
    await expect(table.getByText('personal_api', { exact: true })).toBeVisible()
    await expect(table.getByText('yd2.0-mini', { exact: true })).toBeVisible()
    await expect(table.getByText('失败', { exact: true })).toBeVisible()
    await expect(table.getByText('尝试 2 次', { exact: true })).toBeVisible()
    await expect(table.getByText('provider fixture failure', { exact: true })).toBeVisible()
  })

  test('@video project VIDEO outputUrl from Go response renders 查看视频 link', async ({ page }) => {
    test.fail(true, 'TASK16-VIDEO-001 Project VIDEO status uses outputUrl but frontend reads outputURL')
    await routeBatchProjectFixtures(page, {
      generationSummary: () => makeGenerationSummary(),
      videoStatus: () => makeVideoStatus({ status: 'succeeded', outputURL: 'https://example.invalid/task16/video.mp4' }),
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()

    await expect(page.locator('.ant-table-tbody').getByRole('link', { name: '查看视频' })).toHaveAttribute('href', 'https://example.invalid/task16/video.mp4')
  })

  test('@video Start and Poll preserve Task14 provider/model/status fields', async ({ page }) => {
    let startPayload = null
    let pollCalls = 0
    await page.route('**/api/v1/batch-projects/123/books/1001/video', async (route) => {
      startPayload = route.request().postDataJSON()
      await route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ job: { id: 801, provider: 'personal_api', model: 'yd2.0-mini', status: 'running' }, task: { id: 701, attempt: 1, provider: 'personal_api', model: 'yd2.0-mini', status: 'running' } }) })
    })
    await page.route('**/api/v1/video-tasks/701/poll', async (route) => {
      pollCalls += 1
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ id: 701, attempt: 1, provider: 'personal_api', model: 'yd2.0-mini', status: 'succeeded', outputURL: 'https://example.invalid/task16/video.mp4' }) })
    })
    await page.goto('/batch-factory')

    const started = await appRequest(page, '/api/v1/batch-projects/123/books/1001/video', { method: 'POST', body: JSON.stringify({ provider: 'personal_api', model: 'yd2.0-mini', requestId: 'task16-video-start' }) })
    const polled = await appRequest(page, '/api/v1/video-tasks/701/poll', { method: 'POST' })

    expect(started.ok).toBe(true)
    expect(startPayload).toMatchObject({ provider: 'personal_api', model: 'yd2.0-mini' })
    expect(started.payload.task.status).toBe('running')
    expect(polled.ok).toBe(true)
    expect(polled.payload.status).toBe('succeeded')
    expect(polled.payload.outputURL).toContain('video.mp4')
    expect(pollCalls).toBe(1)
  })

  test('@video Retry from failed VIDEO refreshes status and does not rerun generation stages', async ({ page }) => {
    let retried = false
    let generationMutationCalls = 0
    let retryCalls = 0
    await routeBatchProjectFixtures(page, {
      generationSummary: () => makeGenerationSummary(),
      videoStatus: () => retried
        ? makeVideoStatus({ status: 'running', attemptID: 702, attempts: 2, outputURL: '' })
        : makeVideoStatus({ status: 'failed', attemptID: 701, errorMessage: 'provider fixture failure', outputURL: '' }),
    })
    await page.route('**/api/v1/batch-projects/123/generation', async (route) => {
      if (route.request().method() === 'POST') generationMutationCalls += 1
      await route.fallback()
    })
    await page.route('**/api/v1/video-tasks/701/retry', async (route) => {
      retryCalls += 1
      retried = true
      await route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ job: { id: 801, status: 'running' }, task: { id: 702, attempt: 2, status: 'running' } }) })
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()
    await page.getByRole('button', { name: '重试 VIDEO' }).click()

    await expect(page.locator('.ant-table-tbody').getByText('执行中', { exact: true })).toBeVisible()
    expect(retryCalls).toBe(1)
    expect(generationMutationCalls).toBe(0)
  })

  test('@video Cancel from queued/running VIDEO calls only cancel and refreshes cancelled state', async ({ page }) => {
    let cancelled = false
    let cancelCalls = 0
    await routeBatchProjectFixtures(page, {
      generationSummary: () => makeGenerationSummary(),
      videoStatus: () => cancelled
        ? makeVideoStatus({ status: 'cancelled', outputURL: '' })
        : makeVideoStatus({ status: 'running', outputURL: '' }),
    })
    await page.route('**/api/v1/video-tasks/701/cancel', async (route) => {
      cancelCalls += 1
      cancelled = true
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ id: 701, status: 'cancelled' }) })
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()
    await page.getByRole('button', { name: '取消 VIDEO' }).click()

    await expect(page.locator('.ant-table-tbody').getByText('已取消', { exact: true })).toBeVisible()
    expect(cancelCalls).toBe(1)
  })

  test('@video unconfigured provider remains a provider-level status and Batch Factory stays usable', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary() })
    await page.route('**/api/v1/video-providers/personal_api/models/yd2.0-mini/status', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ providerKey: 'personal_api', model: 'yd2.0-mini', configured: false, enabled: false, status: 'unconfigured', message: 'provider is not configured' }) })
    })
    await page.goto('/batch-factory')

    const status = await appRequest(page, '/api/v1/video-providers/personal_api/models/yd2.0-mini/status')
    expect(status.payload.status).toBe('unconfigured')
    await page.getByRole('button', { name: '生成状态' }).click()
    await expect(page.locator('.ant-drawer-title').filter({ hasText: '剧本与 VIDEO 流水线' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'VIDEO' })).toBeVisible()
  })

  test('@video Merge Start/Get moves queued→running→succeeded without starting VIDEO again', async ({ page }) => {
    let mergeRead = 0
    let videoStartCalls = 0
    await page.route('**/api/v1/batch-projects/123/books/1001/video', async (route) => {
      videoStartCalls += 1
      await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' })
    })
    await page.route('**/api/v1/batch-projects/123/books/1001/merge', async (route) => {
      await route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ job: { id: 901, status: 'queued' }, attempt: { id: 902, status: 'queued' } }) })
    })
    await page.route('**/api/v1/video-merge-jobs/901', async (route) => {
      mergeRead += 1
      const status = mergeRead === 1 ? 'running' : 'succeeded'
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ job: { id: 901, status, outputURL: status === 'succeeded' ? 'https://example.invalid/task16/merged.mp4' : '' }, attempts: [{ id: 902, status }] }) })
    })
    await page.goto('/batch-factory')

    const started = await appRequest(page, '/api/v1/batch-projects/123/books/1001/merge', { method: 'POST', body: JSON.stringify({ productionTaskIds: [701], aspectRatio: '9:16', speed: 1 }) })
    const running = await appRequest(page, '/api/v1/video-merge-jobs/901')
    const succeeded = await appRequest(page, '/api/v1/video-merge-jobs/901')

    expect(started.payload.job.status).toBe('queued')
    expect(running.payload.job.status).toBe('running')
    expect(succeeded.payload.job.status).toBe('succeeded')
    expect(succeeded.payload.job.outputURL).toContain('merged.mp4')
    expect(videoStartCalls).toBe(0)
  })

  test('@video Merge Retry retries the failed merge attempt without regenerating succeeded VIDEO', async ({ page }) => {
    let retryCalls = 0
    let videoStartCalls = 0
    await page.route('**/api/v1/batch-projects/123/books/1001/video', async (route) => {
      videoStartCalls += 1
      await route.fulfill({ status: 500, contentType: 'application/json', body: '{}' })
    })
    await page.route('**/api/v1/video-merge-attempts/902/retry', async (route) => {
      retryCalls += 1
      await route.fulfill({ status: 202, contentType: 'application/json', body: JSON.stringify({ job: { id: 901, status: 'running' }, attempt: { id: 903, status: 'running' } }) })
    })
    await page.goto('/batch-factory')

    const result = await appRequest(page, '/api/v1/video-merge-attempts/902/retry', { method: 'POST' })
    expect(result.ok).toBe(true)
    expect(result.payload.attempt.status).toBe('running')
    expect(retryCalls).toBe(1)
    expect(videoStartCalls).toBe(0)
  })

  test('@video missing ffmpeg is an explicit Merge failure and does not crash Batch Factory', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary() })
    await page.route('**/api/v1/video-merge-jobs/901', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ job: { id: 901, status: 'failed', errorMessage: 'ffmpeg executable not found' }, attempts: [{ id: 902, status: 'failed', errorMessage: 'ffmpeg executable not found' }] }) })
    })
    await page.goto('/batch-factory')

    const merge = await appRequest(page, '/api/v1/video-merge-jobs/901')
    expect(merge.ok).toBe(true)
    expect(merge.payload.job.status).toBe('failed')
    expect(merge.payload.job.errorMessage).toMatch(/ffmpeg/i)
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
  })

  test('@video @public-smoke Tier 2 real provider smoke is explicit opt-in only', async () => {
    test.skip(process.env.E2E_REAL_VIDEO_PROVIDER !== '1', 'Real paid/provider VIDEO smoke requires E2E_REAL_VIDEO_PROVIDER=1')
    test.skip(!process.env.PUBLIC_BASE_URL, 'Real VIDEO provider smoke requires PUBLIC_BASE_URL')
  })
})