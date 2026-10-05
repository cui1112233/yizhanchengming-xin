import { expect, test } from './support/acceptance.js'
import { makeGenerationSummary, routeAuthenticated, routeBatchProjectFixtures } from './support/fixtures.js'
import { validateTimeline } from './support/timeline.js'

function stageSummary(stageOverrides = {}) {
  const summary = makeGenerationSummary()
  summary.books[0].stages = { ...summary.books[0].stages, ...stageOverrides }
  return summary
}

test.describe('@generation @audio generation and matchAudio acceptance', () => {
  test.beforeEach(async ({ page }) => {
    await routeAuthenticated(page)
  })

  test('@generation Script / Hook enabled / Director / Final Prompt stages render stable status', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary() })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()

    await expect(page.getByText('已检测音频：28.00 秒')).toBeVisible()
    await expect(page.getByText('完成', { exact: true })).toHaveCount(4)
    await expect(page.getByText('Script')).toBeVisible()
    await expect(page.getByText('Hook')).toBeVisible()
    await expect(page.getByText('Director')).toBeVisible()
    await expect(page.getByText('Final Prompt')).toBeVisible()
  })

  test('@generation Hook disabled is represented as skipped, not failed', async ({ page }) => {
    await routeBatchProjectFixtures(page, {
      generationSummary: () => stageSummary({ HOOK: { status: 'skipped', outputText: '' } }),
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()

    await expect(page.getByText('已跳过')).toBeVisible()
    await expect(page.getByRole('button', { name: '重试 Hook' })).toHaveCount(0)
  })

  test('@generation Retry executes only the failed Stage and keeps successful upstream stages', async ({ page }) => {
    let retried = false
    await routeBatchProjectFixtures(page, {
      generationSummary: () => retried
        ? stageSummary({ SCRIPT: { status: 'completed', outputText: 'script-after-retry' } })
        : stageSummary({ SCRIPT: { status: 'failed', errorMessage: 'fixture script failure' } }),
    })
    let retryPath = ''
    await page.route('**/api/v1/batch-projects/123/books/1001/generation/stages/SCRIPT/retry', async (route) => {
      retryPath = new URL(route.request().url()).pathname
      retried = true
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ status: 'accepted' }) })
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()
    await expect(page.getByText('fixture script failure')).toBeVisible()
    await page.getByRole('button', { name: '重试 Script' }).click()

    await expect(page.getByText('fixture script failure')).toHaveCount(0)
    expect(retryPath).toBe('/api/v1/batch-projects/123/books/1001/generation/stages/SCRIPT/retry')
    await expect(page.getByText('完成', { exact: true })).toHaveCount(4)
  })

  test('@generation H3 Director API accepts explicit h3 mode without replacing the real app', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary() })
    let payload = null
    await page.route('**/api/v1/batch-projects/123/books/1001/generation', async (route) => {
      payload = route.request().postDataJSON()
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ status: 'accepted', directorMode: 'h3' }) })
    })

    await page.goto('/batch-factory')
    const status = await page.evaluate(async () => {
      const response = await fetch('/api/v1/batch-projects/123/books/1001/generation', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ hookEnabled: true, directorMode: 'h3', matchAudio: false, requestId: 'task16-h3-fixture' }),
      })
      return response.status
    })

    expect(status).toBe(200)
    expect(payload?.directorMode).toBe('h3')
    expect(payload?.hookEnabled).toBe(true)
  })

  test('@audio matchAudio switch uses measured 28.00s and ≤15s shot limit in execution payload', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary(), audioDurationMs: 28000 })
    let payload = null
    await page.route('**/api/v1/batch-projects/123/books/1001/generation', async (route) => {
      payload = route.request().postDataJSON()
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ status: 'accepted' }) })
    })

    await page.goto('/batch-factory')
    await page.getByRole('button', { name: '生成状态' }).click()
    await expect(page.getByText('已检测音频：28.00 秒')).toBeVisible()
    await page.getByRole('switch', { name: '匹配音频 Book 1001' }).click()
    await expect(page.getByText('最终分镜总时长将严格匹配音频时长')).toBeVisible()
    await page.getByRole('button', { name: '单本执行' }).click()

    expect(payload?.matchAudio).toBe(true)
    expect(payload?.audioDurationSec).toBe(28)
    expect(payload?.shotDurationLimitSec).toBe(15)
  })

  test('@audio timeline validator enforces start=0 end=28.00 no gap/overlap and each shot ≤15s', async () => {
    const valid = [
      { start: 0, end: 10 },
      { start: 10, end: 20 },
      { start: 20, end: 28 },
    ]
    expect(validateTimeline(valid, 28, 15)).toBe(true)
    expect(() => validateTimeline([{ start: 0, end: 16 }, { start: 16, end: 28 }], 28, 15)).toThrow(/exceeds 15s/)
    expect(() => validateTimeline([{ start: 0, end: 10 }, { start: 11, end: 28 }], 28, 15)).toThrow(/gap\/overlap/)
    expect(() => validateTimeline([{ start: 0, end: 14 }, { start: 14, end: 27.99 }], 28, 15)).toThrow(/expected 28/)
  })

  test('@audio missing ffprobe is an explicit API error and does not crash the page', async ({ page }) => {
    await routeBatchProjectFixtures(page, { generationSummary: () => makeGenerationSummary() })
    await page.route('**/api/v1/batch-projects/123/books/1001/audio-measurement', async (route) => {
      await route.fulfill({
        status: 503,
        contentType: 'application/json',
        body: JSON.stringify({ code: 'FFPROBE_UNAVAILABLE', message: 'ffprobe is not installed on this runtime' }),
      })
    })

    await page.goto('/batch-factory')
    const result = await page.evaluate(async () => {
      const response = await fetch('/api/v1/batch-projects/123/books/1001/audio-measurement', { credentials: 'include' })
      return { status: response.status, body: await response.json() }
    })

    expect(result.status).toBe(503)
    expect(result.body.code).toBe('FFPROBE_UNAVAILABLE')
    expect(result.body.message).toMatch(/ffprobe/i)
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
  })
})
