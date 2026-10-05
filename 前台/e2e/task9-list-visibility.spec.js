import { expect, test } from './support/acceptance.js'
import { routeAuthenticated } from './support/fixtures.js'

const SOURCE_INPUT = 'input[placeholder="例如：阳光、常读、知乎"]'
const PLATFORM_INPUT = 'input[placeholder="例如：4"]'
const BOOK_IDS_INPUT = 'textarea:not([aria-hidden="true"])'

test.describe('@batch Task 9 project-list acceptance', () => {
  test.beforeEach(async ({ page }) => {
    await routeAuthenticated(page)
  })

  test('@batch Task8 completed intake project appears immediately in Batch Factory list', async ({ page }) => {
    const book = { id: 11, source: '知乎', platformId: '6', bookId: '9001', title: 'Task8新书', gender: '女频', style: '情感', status: 'fetched', errorMessage: '' }
    await page.route('**/api/v1/intakes', async (route) => route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify({ intake: { id: 51, name: 'Task8完成批次' }, books: [] }) }))
    await page.route('**/api/v1/intakes/51/execute', async (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ intakeId: 51, status: 'completed', fetched: 1, failed: 0 }) }))
    await page.route('**/api/v1/intakes/51/books', async (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ books: [book] }) }))
    await page.route('**/api/v1/intakes/51/batch-projects', async (route) => route.fulfill({ status: 201, contentType: 'application/json', body: JSON.stringify({ project: { id: 123, intakeId: 51, name: 'Task8完成批次' }, run: { id: 901, batchProjectId: 123, status: 'pending' } }) }))
    await page.route('**/api/v1/batch-projects', async (route) => route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [{ id: 123, intakeId: 51, name: 'Task8完成批次', sources: ['知乎'], bookCount: 1, genders: ['女频'], styles: ['情感'], runStatus: 'pending' }] }) }))

    await page.goto('/novel-fetch')
    await page.locator(SOURCE_INPUT).fill('知乎')
    await page.locator(PLATFORM_INPUT).fill('6')
    await page.locator(BOOK_IDS_INPUT).fill('9001')
    await page.getByRole('button', { name: '添加书城' }).click()
    await page.getByRole('button', { name: '立即执行' }).click()
    await expect(page.getByText('立即执行任务已创建。')).toBeVisible()

    await page.getByRole('button', { name: '批量工厂' }).click()
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Task8完成批次' })).toBeVisible()
    await expect(page.getByText('知乎')).toBeVisible()
    await expect(page.getByText('女频')).toBeVisible()
    await expect(page.getByText('情感')).toBeVisible()
    await expect(page.getByText('待执行')).toBeVisible()
  })
})
