import { expect, test } from './support/acceptance.js'
import { routeAuthenticated } from './support/fixtures.js'

const SOURCE_INPUT = 'input[placeholder="例如：阳光、常读、知乎"]'
const PLATFORM_INPUT = 'input[placeholder="例如：4"]'
const BOOK_IDS_INPUT = 'textarea:not([aria-hidden="true"])'
const isPublic = Boolean(process.env.PUBLIC_BASE_URL?.trim())

async function addGroup(page, source, platformId, ids) {
  await page.locator(SOURCE_INPUT).fill(source)
  await page.locator(PLATFORM_INPUT).fill(String(platformId))
  await page.locator(BOOK_IDS_INPUT).fill(ids)
  await page.getByRole('button', { name: '添加书城' }).click()
}

async function loginPublic(page) {
  const username = process.env.E2E_USERNAME?.trim()
  const password = process.env.E2E_PASSWORD?.trim()
  test.skip(!username || !password, 'Public smoke requires E2E_USERNAME and E2E_PASSWORD secrets')
  await page.goto('/novel-fetch')
  if (await page.getByText('一战晟铭登录').count()) {
    await page.getByLabel('用户名').fill(username)
    await page.getByLabel('密码').fill(password)
    await page.locator('button[type="submit"]').click()
  }
  await expect(page.getByRole('heading', { name: '小说获取工作台' })).toBeVisible()
}

async function routeIntakeWorkflow(page, { executionStatus = 'completed', failed = 0, books } = {}) {
  const returnedBooks = books || [
    { id: 1, source: '阳光', platformId: '4', bookId: '101', title: '受控测试小说', category: '男频', genre: '都市', gender: '男频', style: '现实', status: executionStatus === 'completed' ? 'fetched' : 'retryable_failed', errorMessage: executionStatus === 'completed' ? '' : 'fixture upstream failure' },
  ]
  let batchProjectCalls = 0
  let batchProjectBody = null

  await page.route('**/api/v1/intakes/51/batch-projects', async (route) => {
    batchProjectCalls += 1
    batchProjectBody = route.request().postDataJSON()
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        project: { id: 123, name: batchProjectBody?.name || 'Task16 Batch' },
        run: { id: 901, status: 'pending', runAt: batchProjectBody?.runAt || '2030-01-01T00:00:00Z' },
      }),
    })
  })
  await page.route('**/api/v1/intakes/51/execute', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ intakeId: 51, status: executionStatus, fetched: returnedBooks.length - failed, failed }),
    })
  })
  await page.route('**/api/v1/intakes/51/books', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ books: returnedBooks }) })
  })
  await page.route('**/api/v1/intakes', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ intake: { id: 51, name: 'Task16 Intake' }, books: [] }),
    })
  })

  return {
    getBatchProjectCalls: () => batchProjectCalls,
    getBatchProjectBody: () => batchProjectBody,
  }
}

test.describe('@novel novel fetch acceptance', () => {
  test.beforeEach(async ({ page }) => {
    if (!isPublic) await routeAuthenticated(page)
  })

  test('@novel page opens and supports multi-store grouping, dedupe, clear input and N-book tags', async ({ page }) => {
    await page.goto('/novel-fetch')
    await expect(page.getByRole('heading', { name: '小说获取工作台' })).toBeVisible()

    await addGroup(page, '阳光', 4, '101 101 102')
    await expect(page.getByText('阳光 2本 · P4')).toBeVisible()
    await expect(page.locator(BOOK_IDS_INPUT)).toHaveValue('')

    await addGroup(page, '阳光', 4, '102 103')
    await expect(page.getByText('阳光 3本 · P4')).toBeVisible()

    await addGroup(page, '常读', 2, '201, 202')
    await expect(page.getByText('常读 2本 · P2')).toBeVisible()
    await expect(page.getByText('已选书城')).toBeVisible()
    await expect(page.getByText('已选小说')).toBeVisible()
  })

  test('@novel duplicate IDs already present in the same store are rejected without changing count', async ({ page }) => {
    await page.goto('/novel-fetch')
    await addGroup(page, '知乎', 6, '301 302')
    await addGroup(page, '知乎', 6, '301 302')

    await expect(page.getByText('知乎 2本 · P6')).toBeVisible()
    await expect(page.getByText('这些 Book ID 已经添加到同一书城，无需重复添加。')).toBeVisible()
  })

  test('@novel partial failure is visible and blocks BatchProject creation', async ({ page }) => {
    const fixture = await routeIntakeWorkflow(page, {
      executionStatus: 'partial_failed',
      failed: 1,
      books: [
        { id: 1, source: '阳光', platformId: '4', bookId: '401', title: '成功小说', category: '男频', genre: '都市', gender: '男频', style: '现实', status: 'fetched', errorMessage: '' },
        { id: 2, source: '阳光', platformId: '4', bookId: '402', title: '失败小说', category: '', genre: '', gender: '', style: '', status: 'retryable_failed', errorMessage: 'fixture upstream failure' },
      ],
    })

    await page.goto('/novel-fetch')
    await addGroup(page, '阳光', 4, '401 402')
    await page.getByRole('button', { name: '立即执行' }).click()

    await expect(page.getByText(/本批次未全部成功/)).toBeVisible()
    await expect(page.getByText('失败小说')).toBeVisible()
    await expect(page.getByText('fixture upstream failure')).toBeVisible()
    expect(fixture.getBatchProjectCalls()).toBe(0)
  })

  test('@novel immediate execution creates a BatchProject only after completed intake', async ({ page }) => {
    const fixture = await routeIntakeWorkflow(page)

    await page.goto('/novel-fetch')
    await addGroup(page, '阳光', 4, '101')
    await page.getByRole('button', { name: '立即执行' }).click()

    await expect(page.getByText('立即执行任务已创建。')).toBeVisible()
    await expect(page.getByText(/#123/)).toBeVisible()
    expect(fixture.getBatchProjectCalls()).toBe(1)
    expect(fixture.getBatchProjectBody()?.runAt).toBeUndefined()
  })

  test('@novel automation submits a future runAt and creates the scheduled Run', async ({ page }) => {
    const fixture = await routeIntakeWorkflow(page)

    await page.goto('/novel-fetch')
    await addGroup(page, '常读', 2, '501')
    await page.getByRole('button', { name: '自动化' }).click()
    await expect(page.getByText('自动化执行')).toBeVisible()
    await page.locator('input[type="datetime-local"]').fill('2030-01-01T12:00')
    await page.getByRole('button', { name: '确认创建' }).click()

    await expect(page.getByText('自动化任务已创建。')).toBeVisible()
    expect(fixture.getBatchProjectCalls()).toBe(1)
    expect(new Date(fixture.getBatchProjectBody()?.runAt).getTime()).toBeGreaterThan(Date.now())
  })

  test('@novel @public-smoke real 121 smoke is opt-in only', async ({ page }) => {
    test.skip(process.env.E2E_REAL_121 !== '1', 'Real 121 calls are enabled only with E2E_REAL_121=1')
    test.skip(!process.env.PUBLIC_BASE_URL, 'Real 121 smoke requires PUBLIC_BASE_URL')
    test.skip(!process.env.E2E_121_BOOK_ID || !process.env.E2E_121_SOURCE || !process.env.E2E_121_PLATFORM_ID, 'Real 121 smoke requires E2E_121_BOOK_ID/E2E_121_SOURCE/E2E_121_PLATFORM_ID')

    await loginPublic(page)
    await addGroup(page, process.env.E2E_121_SOURCE, process.env.E2E_121_PLATFORM_ID, process.env.E2E_121_BOOK_ID)
    await page.getByRole('button', { name: '立即执行' }).click()

    await expect(page.getByText(/任务已创建|本批次未全部成功/)).toBeVisible({ timeout: 120_000 })
  })
})
