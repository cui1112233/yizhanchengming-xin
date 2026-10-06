import { expect, test } from './support/acceptance.js'
import { defaultUser } from './support/fixtures.js'

test.describe('@auth multi-tab session acceptance', () => {
  test('@auth a refresh failure in one tab does not force another authenticated tab to Login', async ({ page, context }) => {
    await page.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: defaultUser }) })
    })
    await page.route('**/api/v1/batch-projects', async (route) => {
      await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ projects: [] }) })
    })

    await page.goto('/batch-factory')
    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()

    const secondPage = await context.newPage()
    let refreshCalls = 0
    await secondPage.route('**/api/auth/current-user', async (route) => {
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
    })
    await secondPage.route('**/api/auth/refresh', async (route) => {
      refreshCalls += 1
      await route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ code: 'SESSION_EXPIRED' }) })
    })

    await secondPage.goto('/batch-factory')

    await expect(secondPage.getByText('一战晟铭登录')).toBeVisible()
    await expect(secondPage.getByText('当前登录状态已过期，请重新登录。')).toBeVisible()
    expect(refreshCalls).toBe(1)

    await expect(page.getByRole('heading', { name: '批量工厂' })).toBeVisible()
    await expect(page.getByText('一战晟铭登录')).toHaveCount(0)

    await secondPage.close()
  })
})
